# Gitea Integration

This document describes how Gitea (lightweight Git server) integrates with the work item system through webhooks, API operations, and merge automation.

## Gitea Overview

**Gitea** is a lightweight, self-hosted Git service written in Go:
- **Size**: ~30MB binary (vs GitLab's 2GB+)
- **Resources**: ~500MB RAM minimum
- **Database**: PostgreSQL, MySQL, SQLite
- **API**: Full REST API compatible with GitHub/GitLab
- **Webhooks**: Extensive webhook support
- **UI**: Clean web interface for code review

## Webhook Configuration

### Webhook Types

Gitea supports multiple webhook events:
- **push** - Code pushed to repository
- **create** - Branch or tag created
- **delete** - Branch or tag deleted
- **issues** - Issue opened, closed, edited, labeled
- **issue_comment** - Comments on issues
- **pull_request** - PR opened, closed, synchronized, edited
- **pull_request_review** - PR reviewed, approved

### Setting Up Webhooks

**Via Gitea UI**:
1. Navigate to repository → Settings → Webhooks
2. Add webhook URL: `http://codevaldcortex:8080/api/v1/webhooks/gitea/issues`
3. Select events: `Issues`, `Pull Requests`
4. Add secret token for validation
5. Set content type: `application/json`

**Via Gitea API**:
```go
func (c *GiteaClient) CreateWebhook(repoOwner, repoName string) error {
    webhook := &gitea.CreateHookOption{
        Type: "gitea",
        Config: map[string]string{
            "url":          "http://codevaldcortex:8080/api/v1/webhooks/gitea/issues",
            "content_type": "json",
            "secret":       c.webhookSecret,
        },
        Events: []string{"issues", "pull_request"},
        Active: true,
    }
    
    _, _, err := c.client.CreateRepoHook(repoOwner, repoName, *webhook)
    return err
}
```

## Webhook Processing

### Issue Event Handler

```go
type IssueWebhookHandler struct {
    executor    *WorkItemExecutor
    db          *arangodb.Database
    secretToken string
}

func (h *IssueWebhookHandler) HandleIssueEvent(c *gin.Context) {
    // 1. Validate webhook signature
    signature := c.GetHeader("X-Gitea-Signature")
    if !h.validateSignature(c.Request.Body, signature) {
        c.JSON(401, gin.H{"error": "invalid signature"})
        return
    }
    
    // 2. Parse payload
    var payload gitea.IssuePayload
    if err := c.BindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": "invalid payload"})
        return
    }
    
    // 3. Process event
    switch payload.Action {
    case "opened":
        h.handleIssueOpened(&payload)
    case "labeled":
        h.handleIssueLabeled(&payload)
    case "closed":
        h.handleIssueClosed(&payload)
    case "edited":
        h.handleIssueEdited(&payload)
    }
    
    c.JSON(200, gin.H{"status": "processed"})
}

func (h *IssueWebhookHandler) handleIssueOpened(payload *gitea.IssuePayload) {
    // Check if issue has work-item label
    if !hasLabel(payload.Issue, "work-item") {
        return
    }
    
    // Trigger work item execution in goroutine
    go func() {
        ctx := context.Background()
        if err := h.executor.Execute(ctx, payload.Issue); err != nil {
            log.Printf("Work item execution failed: %v", err)
            
            // Add error comment to issue
            h.executor.gitea.CreateIssueComment(
                payload.Repository.Owner.UserName,
                payload.Repository.Name,
                payload.Issue.Index,
                fmt.Sprintf("❌ Work item execution failed: %v", err),
            )
        }
    }()
}

func (h *IssueWebhookHandler) handleIssueLabeled(payload *gitea.IssuePayload) {
    // If work-item label was just added, trigger execution
    if payload.Label.Name == "work-item" {
        go h.executor.Execute(context.Background(), payload.Issue)
    }
}

func (h *IssueWebhookHandler) handleIssueClosed(payload *gitea.IssuePayload) {
    // Update work item status in ArangoDB
    query := `
        FOR wi IN work_items
            FILTER wi.gitea_issue_id == @issueID
            UPDATE wi WITH {status: "completed", completed_at: @now} IN work_items
    `
    
    h.db.Query(context.Background(), query, map[string]interface{}{
        "issueID": payload.Issue.Index,
        "now":     time.Now(),
    })
}
```

### Pull Request Event Handler

```go
func (h *PRWebhookHandler) HandlePREvent(c *gin.Context) {
    var payload gitea.PullRequestPayload
    c.BindJSON(&payload)
    
    switch payload.Action {
    case "opened":
        h.handlePROpened(&payload)
    case "synchronized":
        // PR updated with new commits
        h.handlePRUpdated(&payload)
    case "labeled":
        if payload.Label.Name == "ready-to-merge" {
            go h.attemptAutoMerge(&payload)
        }
    case "closed":
        if payload.PullRequest.Merged {
            h.handlePRMerged(&payload)
        }
    }
    
    c.JSON(200, gin.H{"status": "processed"})
}

func (h *PRWebhookHandler) handlePROpened(payload *gitea.PullRequestPayload) {
    pr := payload.PullRequest
    
    // Check if automated PR (from work item)
    if strings.HasPrefix(pr.Title, "[Work Item]") {
        // Extract work item ID from PR body
        workItemID := extractWorkItemID(pr.Body)
        
        // Link PR to work item in ArangoDB
        h.db.Collection("work_items").UpdateDocument(
            context.Background(),
            workItemID,
            map[string]interface{}{
                "merge_request_url": pr.HTMLURL,
                "status":            "review",
            },
        )
    }
}
```

## Gitea API Operations

### Client Setup

```go
type GiteaClient struct {
    client      *gitea.Client
    baseURL     string
    token       string
    repoOwner   string
    repoName    string
}

func NewGiteaClient(baseURL, token, owner, repo string) *GiteaClient {
    client, _ := gitea.NewClient(baseURL, gitea.SetToken(token))
    
    return &GiteaClient{
        client:    client,
        baseURL:   baseURL,
        token:     token,
        repoOwner: owner,
        repoName:  repo,
    }
}
```

### Branch Operations

```go
func (c *GiteaClient) CreateBranch(branchName, fromRef string) error {
    // Get base commit
    ref, _, err := c.client.GetRepoRef(c.repoOwner, c.repoName, fromRef)
    if err != nil {
        return fmt.Errorf("failed to get base ref: %w", err)
    }
    
    // Create new branch
    _, _, err = c.client.CreateBranch(c.repoOwner, c.repoName, gitea.CreateBranchOption{
        BranchName:    branchName,
        OldBranchName: fromRef,
    })
    
    return err
}

func (c *GiteaClient) DeleteBranch(branchName string) error {
    _, err := c.client.DeleteRepoBranch(c.repoOwner, c.repoName, branchName)
    return err
}
```

### File Operations

```go
func (c *GiteaClient) CreateOrUpdateFile(branch, path, content, message string) error {
    // Check if file exists
    existingFile, resp, _ := c.client.GetContents(
        c.repoOwner,
        c.repoName,
        branch,
        path,
    )
    
    option := gitea.CreateFileOptions{
        FileOptions: gitea.FileOptions{
            Message: message,
            BranchOpt: gitea.BranchOpt{
                NewBranch: branch,
            },
        },
        Content: base64.StdEncoding.EncodeToString([]byte(content)),
    }
    
    if resp != nil && resp.StatusCode == 200 {
        // File exists, update it
        option.SHA = existingFile.SHA
        _, _, err := c.client.UpdateFile(c.repoOwner, c.repoName, path, option)
        return err
    }
    
    // File doesn't exist, create it
    _, _, err := c.client.CreateFile(c.repoOwner, c.repoName, path, option)
    return err
}

func (c *GiteaClient) GetFileContent(branch, path string) (string, error) {
    contents, _, err := c.client.GetContents(c.repoOwner, c.repoName, branch, path)
    if err != nil {
        return "", err
    }
    
    decoded, err := base64.StdEncoding.DecodeString(*contents.Content)
    return string(decoded), err
}
```

### Pull Request Operations

```go
func (c *GiteaClient) CreatePullRequest(head, base, title, body string) (*gitea.PullRequest, error) {
    pr, _, err := c.client.CreatePullRequest(c.repoOwner, c.repoName, gitea.CreatePullRequestOption{
        Head:  head,
        Base:  base,
        Title: title,
        Body:  body,
    })
    
    return pr, err
}

func (c *GiteaClient) MergePullRequest(index int64, style string) error {
    _, _, err := c.client.MergePullRequest(
        c.repoOwner,
        c.repoName,
        index,
        gitea.MergePullRequestOption{
            Style: gitea.MergeStyle(style), // "merge", "squash", "rebase"
            DeleteBranchAfterMerge: true,
        },
    )
    
    return err
}

func (c *GiteaClient) GetPullRequest(index int64) (*gitea.PullRequest, error) {
    pr, _, err := c.client.GetPullRequest(c.repoOwner, c.repoName, index)
    return pr, err
}

func (c *GiteaClient) ListPullRequests(state string) ([]*gitea.PullRequest, error) {
    prs, _, err := c.client.ListRepoPullRequests(c.repoOwner, c.repoName, gitea.ListPullRequestsOptions{
        State: gitea.StateType(state), // "open", "closed", "all"
    })
    
    return prs, err
}
```

### Issue Operations

```go
func (c *GiteaClient) CreateIssue(title, body string, labels []string) (*gitea.Issue, error) {
    issue, _, err := c.client.CreateIssue(c.repoOwner, c.repoName, gitea.CreateIssueOption{
        Title:  title,
        Body:   body,
        Labels: labels,
    })
    
    return issue, err
}

func (c *GiteaClient) CreateIssueComment(index int64, comment string) error {
    _, _, err := c.client.CreateIssueComment(c.repoOwner, c.repoName, index, gitea.CreateIssueCommentOption{
        Body: comment,
    })
    
    return err
}

func (c *GiteaClient) CloseIssue(index int64) error {
    state := gitea.StateClosed
    _, _, err := c.client.EditIssue(c.repoOwner, c.repoName, index, gitea.EditIssueOption{
        State: &state,
    })
    
    return err
}
```

## Merge Automation

### Merge Strategy Selection

```go
type MergeStrategy struct {
    Style          string   // "merge", "squash", "rebase"
    AutoMerge      bool
    RequireReviews int
    RequireCI      bool
    RequireChecks  []string
}

func (e *Executor) getMergeStrategy(issue *gitea.Issue) MergeStrategy {
    for _, label := range issue.Labels {
        switch label.Name {
        case "documentation", "docs":
            return MergeStrategy{
                Style:          "squash",
                AutoMerge:      true,
                RequireReviews: 0,
                RequireCI:      false,
            }
            
        case "automated", "llm-generated":
            return MergeStrategy{
                Style:          "squash",
                AutoMerge:      true,
                RequireReviews: 0,
                RequireCI:      true,
                RequireChecks:  []string{"lint", "test"},
            }
            
        case "feature", "enhancement":
            return MergeStrategy{
                Style:          "merge",
                AutoMerge:      false,
                RequireReviews: 2,
                RequireCI:      true,
            }
            
        case "hotfix", "urgent":
            return MergeStrategy{
                Style:          "squash",
                AutoMerge:      true,
                RequireReviews: 1,
                RequireCI:      true,
            }
        }
    }
    
    // Default: conservative
    return MergeStrategy{
        Style:          "merge",
        AutoMerge:      false,
        RequireReviews: 2,
        RequireCI:      true,
    }
}
```

### Auto-Merge Logic

```go
func (e *Executor) attemptAutoMerge(pr *gitea.PullRequest) error {
    // 1. Get merge strategy from original issue
    issue := e.getIssueForPR(pr.Index)
    strategy := e.getMergeStrategy(issue)
    
    if !strategy.AutoMerge {
        return fmt.Errorf("auto-merge not enabled for this PR")
    }
    
    // 2. Check if mergeable
    if !pr.Mergeable {
        return fmt.Errorf("PR has merge conflicts")
    }
    
    // 3. Check CI status
    if strategy.RequireCI {
        status, err := e.gitea.GetCombinedStatus(pr.Head.Sha)
        if err != nil {
            return fmt.Errorf("failed to get CI status: %w", err)
        }
        
        if status.State != gitea.StatusSuccess {
            return fmt.Errorf("CI not passing: %s", status.State)
        }
    }
    
    // 4. Check required checks
    if len(strategy.RequireChecks) > 0 {
        statuses, err := e.gitea.ListStatuses(pr.Head.Sha)
        if err != nil {
            return fmt.Errorf("failed to get check statuses: %w", err)
        }
        
        for _, requiredCheck := range strategy.RequireChecks {
            found := false
            for _, status := range statuses {
                if status.Context == requiredCheck && status.State == gitea.StatusSuccess {
                    found = true
                    break
                }
            }
            
            if !found {
                return fmt.Errorf("required check '%s' not passing", requiredCheck)
            }
        }
    }
    
    // 5. Check reviews
    if strategy.RequireReviews > 0 {
        reviews, err := e.gitea.ListPullReviews(pr.Index)
        if err != nil {
            return fmt.Errorf("failed to get reviews: %w", err)
        }
        
        approvals := 0
        for _, review := range reviews {
            if review.State == gitea.ReviewStateApproved {
                approvals++
            }
        }
        
        if approvals < strategy.RequireReviews {
            return fmt.Errorf("not enough approvals: %d/%d", approvals, strategy.RequireReviews)
        }
    }
    
    // 6. Merge!
    err := e.gitea.MergePullRequest(pr.Index, strategy.Style)
    if err != nil {
        return fmt.Errorf("merge failed: %w", err)
    }
    
    // 7. Update work item
    e.updateWorkItemStatus(issue.Index, "merged")
    
    // 8. Close issue
    e.gitea.CloseIssue(issue.Index)
    
    return nil
}
```

### Batch Merge for High Volume

```go
func (e *Executor) BatchMergeReadyPRs(ctx context.Context) error {
    // Get all PRs with "ready-to-merge" label
    prs, err := e.gitea.ListPullRequests("open")
    if err != nil {
        return err
    }
    
    var readyPRs []*gitea.PullRequest
    for _, pr := range prs {
        if hasLabel(pr.Labels, "ready-to-merge") {
            readyPRs = append(readyPRs, pr)
        }
    }
    
    // Sort by priority (custom label)
    sort.Slice(readyPRs, func(i, j int) bool {
        return getPriority(readyPRs[i]) > getPriority(readyPRs[j])
    })
    
    // Merge in batches of 5 with delays
    for i := 0; i < len(readyPRs); i += 5 {
        batch := readyPRs[i:min(i+5, len(readyPRs))]
        
        for _, pr := range batch {
            if err := e.attemptAutoMerge(pr); err != nil {
                // Log error but continue
                log.Printf("Failed to merge PR #%d: %v", pr.Index, err)
                
                // Add comment explaining failure
                e.gitea.CreateIssueComment(pr.Index, 
                    fmt.Sprintf("❌ Auto-merge failed: %v\n\nManual intervention required.", err))
            } else {
                log.Printf("✅ Merged PR #%d", pr.Index)
            }
            
            time.Sleep(2 * time.Second) // Rate limiting
        }
        
        time.Sleep(10 * time.Second) // Between batches
    }
    
    return nil
}
```

## Status Reporting

### Commit Status API

```go
func (e *Executor) reportStatus(commitSHA, state, context, description string) error {
    status := gitea.CreateStatusOption{
        State:       gitea.StatusState(state), // "pending", "success", "error", "failure"
        Context:     context,
        Description: description,
        TargetURL:   fmt.Sprintf("https://codevaldcortex.ai/work-items/%s", e.workItemID),
    }
    
    _, _, err := e.gitea.client.CreateStatus(
        e.gitea.repoOwner,
        e.gitea.repoName,
        commitSHA,
        status,
    )
    
    return err
}

// Usage during work item execution
func (e *Executor) Execute(ctx context.Context, issue *gitea.Issue) error {
    // Create branch and commit
    commitSHA := e.createCommit(...)
    
    // Report pending
    e.reportStatus(commitSHA, "pending", "work-item/execution", "Work item executing...")
    
    // Execute work
    if err := e.executeWork(ctx, issue); err != nil {
        e.reportStatus(commitSHA, "failure", "work-item/execution", "Execution failed")
        return err
    }
    
    // Report success
    e.reportStatus(commitSHA, "success", "work-item/execution", "Work item completed")
    return nil
}
```

---

**See Also**:
- [GitOps Workflow](./gitops-workflow.md) - Complete execution flow
- [Work Item Types](./work-item-types.md) - Label-based work classification
