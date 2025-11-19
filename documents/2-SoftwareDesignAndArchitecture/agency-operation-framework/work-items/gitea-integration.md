# Gitea Integration

This document describes how Gitea (lightweight Git server) integrates with the work item system to enable **automatic agent instantiation** from published agency work item definitions.

## Conceptual Model

### Work Item Definition vs Agent Instance

**Key Distinction**:
- **Work Item Definition**: A blueprint/template defined in an agency specification that describes a type of work (e.g., "Requirements Gathering", "Code Review", "Security Audit")
- **Agent Instance**: An autonomous agent created from a work item definition to execute specific work triggered by a Gitea issue

### Lifecycle Flow

```
1. Agency Design Phase
   └─> Define Work Item Definitions (blueprints)
       Example: "Conduct stakeholder requirements gathering session"
       - Execute structured interviews and workshops
       - Collect functional and non-functional requirements
       - Document findings in standardized format

2. Agency Published
   └─> Work Item Definitions stored in ArangoDB
       - Available as templates/blueprints
       - Ready for agent instantiation

3. Gitea Issue Created
   └─> "Gather requirements for authentication module"
       - Label: "requirements-gathering"
       - Webhook fires to CodeValdCortex

4. Agent Instantiation (THIS IS THE KEY STEP)
   └─> Match issue → Work Item Definition
   └─> Create Agent Instance from definition
       - Agent Type: Requirements Gathering Agent
       - Agent Task: Authentication module requirements
       - Agent Config: Inherited from work item definition

5. Agent Execution
   └─> Agent autonomously executes work
       - Conducts interviews/workshops
       - Collects requirements
       - Posts progress updates to Gitea issue
       - Marks issue complete when done
```

**Critical Understanding**: The Gitea webhook does NOT create a "work item" (task ticket). It creates an **AGENT** based on a work item definition. The agent is the autonomous entity that performs the work.

## Gitea Overview

**Gitea** is a lightweight, self-hosted Git service written in Go:
- **Size**: ~30MB binary (vs GitLab's 2GB+)
- **Resources**: ~500MB RAM minimum
- **Database**: PostgreSQL, MySQL, SQLite
- **API**: Full REST API compatible with GitHub/GitLab
- **Webhooks**: Extensive webhook support
- **UI**: Clean web interface for code review

## Kanban-Based Work Management

### Workflow as Kanban Board

**Conceptual Model**:
- **Workflow** (in published agency) = **Kanban Board** with columns
- **Gitea Issues** = **Cards** placed in columns
- **Agents** = **Workers** that pick up cards from specific columns
- **Work Item Definitions** = **Agent blueprints** assigned to columns

### Kanban Workflow Structure

```yaml
# Agency Workflow Definition (stored in ArangoDB)
workflow:
  id: "software-dev-workflow-001"
  name: "Software Development Kanban"
  agency_id: "software-dev-agency-001"
  
  columns:
    - id: "requirements"
      name: "Requirements Gathering"
      position: 1
      work_item_definition_id: "wid-requirements-001"  # Links to agent blueprint
      agent_config:
        auto_assign: true          # Auto-create agent when issue enters column
        max_concurrent: 3          # Max 3 agents working this column simultaneously
        
    - id: "design"
      name: "System Design"
      position: 2
      work_item_definition_id: "wid-design-001"
      agent_config:
        auto_assign: true
        max_concurrent: 2
        
    - id: "implementation"
      name: "Implementation"
      position: 3
      work_item_definition_id: "wid-coding-001"
      agent_config:
        auto_assign: true
        max_concurrent: 5
        
    - id: "review"
      name: "Code Review"
      position: 4
      work_item_definition_id: "wid-review-001"
      agent_config:
        auto_assign: true
        max_concurrent: 2
        
    - id: "done"
      name: "Done"
      position: 5
      work_item_definition_id: null  # No agent needed for done column
      agent_config:
        auto_assign: false
```

### Gitea Milestones as Kanban Columns

**Mapping Strategy**: Use Gitea **Milestones** to represent Kanban columns

```
Gitea Repository Setup:
├─ Milestone: "Requirements Gathering"  → Maps to workflow column "requirements"
├─ Milestone: "System Design"           → Maps to workflow column "design"
├─ Milestone: "Implementation"          → Maps to workflow column "implementation"
├─ Milestone: "Code Review"             → Maps to workflow column "review"
└─ Milestone: "Done"                    → Maps to workflow column "done"
```

**Why Milestones?**
- ✅ Built-in Gitea feature (no custom labels needed)
- ✅ Visual board in Gitea UI (Projects view)
- ✅ Issues can only be in ONE milestone (enforces single column)
- ✅ Webhook events fire when milestone changes
- ✅ Easy to track progress and WIP limits

### Integration Flow

```
1. Agency Published
   └─> Workflow defined with Kanban columns
       └─> Each column links to Work Item Definition (agent blueprint)

2. Gitea Repository Configured
   └─> Milestones created matching workflow columns
   └─> Repository webhook configured: http://codevaldcortex:8080/api/v1/work/issues

3. Issue Created & Assigned to Milestone
   └─> Issue: "Gather auth requirements"
   └─> Milestone: "Requirements Gathering"
   └─> Webhook fires: issues.milestoned

4. Webhook Processing
   └─> Extract: Repository, Issue, Milestone
   └─> Find workflow: BY repository_url
   └─> Find column: WHERE milestone_name = "Requirements Gathering"
   └─> Get work_item_definition_id from column config

5. Agent Creation (if auto_assign = true)
   └─> Check: Current agents in column < max_concurrent?
   └─> Create agent from work_item_definition
   └─> Assign agent to issue
   └─> Agent starts work

6. Agent Execution
   └─> Agent performs work (requirements gathering)
   └─> Posts updates to issue comments
   └─> When done: Moves issue to next milestone (Design)
   └─> Webhook fires again → Next agent picks it up

7. Issue Moves Through Workflow
   Requirements → Design → Implementation → Review → Done
   (Each stage triggers a new agent specialized for that work)
```

### Repository-Workflow Mapping

**Configuration in ArangoDB**:
```json
{
  "_key": "repo-workflow-map-001",
  "repository_url": "https://gitea.example.com/myorg/myproject",
  "workflow_id": "software-dev-workflow-001",
  "agency_id": "software-dev-agency-001",
  "milestone_column_mapping": {
    "Requirements Gathering": "requirements",
    "System Design": "design",
    "Implementation": "implementation",
    "Code Review": "review",
    "Done": "done"
  },
  "enabled": true,
  "created_at": "2025-11-19T10:00:00Z"
}
```

**Setup UI in Agency Designer**:
```
Agency Designer → Workflows Tab
  ├─ Create Kanban Workflow
  ├─ Define Columns (drag & drop)
  ├─ Assign Work Item Definition to each column
  ├─ Configure agent settings (auto_assign, max_concurrent)
  └─ Link to Gitea Repository
      └─ Auto-create milestones in Gitea
      └─ Configure webhook
```

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
2. Add webhook URL: `http://codevaldcortex:8080/api/v1/work/issues`
3. Select events: `Issues`, `Pull Requests`
4. Add secret token for validation
5. Set content type: `application/json`

**Via Gitea API**:
```go
func (c *GiteaClient) CreateWebhook(repoOwner, repoName string) error {
    webhook := &gitea.CreateHookOption{
        Type: "gitea",
        Config: map[string]string{
            "url":          "http://codevaldcortex:8080/api/v1/work/issues",
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

## Webhook Processing (MVP-WI-001)

### Issue Event Handler - ArangoDB Persistence Layer

**Core Responsibility**: Receive Gitea webhooks, validate signatures, persist artifacts to ArangoDB

```go
type IssueWebhookHandler struct {
    db              *arangodb.Database   // ArangoDB client
    secretToken     string               // Webhook signature validation
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
    
    // 3. Persist to ArangoDB (orchestrator will detect via change streams)
    switch payload.Action {
    case "opened", "milestoned":
        h.saveIssueToArangoDB(&payload)
    case "closed":
        h.updateIssueState(&payload, "closed")
    case "edited":
        h.updateIssueToArangoDB(&payload)
    }
    
    c.JSON(200, gin.H{"status": "persisted"})
}

func (h *IssueWebhookHandler) saveIssueToArangoDB(payload *gitea.IssuePayload) error {
    issueDoc := &GiteaIssue{
        Key:          fmt.Sprintf("gitea-issue-%d", payload.Issue.ID),
        IssueID:      payload.Issue.ID,
        IssueNumber:  payload.Issue.Number,
        Title:        payload.Issue.Title,
        Body:         payload.Issue.Body,
        State:        payload.Issue.State,
        Milestone:    payload.Issue.Milestone.Title,
        MilestoneID:  payload.Issue.Milestone.ID,
        RepoURL:      payload.Repository.HTMLURL,
        RepoOwner:    payload.Repository.Owner.Login,
        RepoName:     payload.Repository.Name,
        Labels:       extractLabels(payload.Issue.Labels),
        Assignees:    extractAssignees(payload.Issue.Assignees),
        CreatedAt:    payload.Issue.CreatedAt,
        UpdatedAt:    payload.Issue.UpdatedAt,
        SyncedAt:     time.Now(),
    }
    
    // Save to ArangoDB - orchestrator will detect via change stream
    collection := h.db.Collection("work-issues")
    _, err := collection.CreateDocument(context.Background(), issueDoc)
    if err != nil {
        log.Error("Failed to save issue to ArangoDB", "error", err)
        return err
    }
    
    log.Info("Issue persisted to ArangoDB",
        "issue_number", issueDoc.IssueNumber,
        "milestone", issueDoc.Milestone,
        "repo", issueDoc.RepoURL)
    
    return nil
}
```

**Key Points**:
- Webhook handler is a **persistence layer only**
- No orchestrator interaction
- No agent creation
- Just validates and saves to ArangoDB
- Orchestrator monitors ArangoDB change streams independently

### Orchestrator Monitors Change Streams (MVP-032)

**Orchestrator detects changes in ArangoDB and creates agents**:

```go
type Orchestrator struct {
    db              *arangodb.Database
    agentFactory    *AgentFactory
    workItemRepo    *WorkItemRepository
}

func (o *Orchestrator) MonitorGiteaIssues() {
    // Subscribe to ArangoDB change streams for work-issues collection
    collection := o.db.Collection("work-issues")
    stream, err := collection.WatchChanges(context.Background(), &WatchOptions{
        FullDocument: "updateLookup",
    })
    
    if err != nil {
        log.Fatal("Failed to watch work-issues collection", "error", err)
    }
    
    log.Info("Orchestrator monitoring work-issues collection")
    
    for change := range stream {
        switch change.OperationType {
        case "insert", "update":
            // New issue or milestone changed
            issueDoc := change.FullDocument.(*GiteaIssue)
            o.handleIssueChange(issueDoc)
        }
    }
}

func (o *Orchestrator) handleIssueChange(issueDoc *GiteaIssue) {
    // Skip if no milestone
    if issueDoc.Milestone == "" {
        return
    }
    
    // 1. Query: What workflow does this repo map to?
    workflow, err := o.getWorkflowForRepo(issueDoc.RepoURL)
    if err != nil {
        log.Warn("No workflow found for repo", "repo", issueDoc.RepoURL)
        return
    }
    
    // 2. Query: What column does this milestone map to?
    column, err := o.getColumnForMilestone(workflow.ID, issueDoc.Milestone)
    if err != nil {
        log.Warn("No column mapping for milestone", 
            "milestone", issueDoc.Milestone,
            "workflow", workflow.Name)
        return
    }
    
    // 3. Check WIP limit
    activeAgents := o.countActiveAgentsInColumn(column.ID)
    if activeAgents >= column.MaxConcurrent {
        log.Info("WIP limit reached, queueing issue",
            "column", column.Name,
            "limit", column.MaxConcurrent)
        o.queueIssue(issueDoc, column)
        return
    }
    
    // 4. Get work item definition (agent blueprint) for this column
    workItemDef, err := o.workItemRepo.GetByID(column.WorkItemDefID)
    if err != nil {
        log.Error("Failed to get work item definition", "error", err)
        return
    }
    
    // 5. Create agent from work item definition
    agent, err := o.agentFactory.CreateFromWorkItemDefinition(AgentCreateRequest{
        WorkItemDefinition: workItemDef,
        IssueID:            issueDoc.Key,
        IssueNumber:        issueDoc.IssueNumber,
        Column:             column,
        Workflow:           workflow,
    })
    
    if err != nil {
        log.Error("Failed to create agent", "error", err)
        return
    }
    
    log.Info("Agent spawned from ArangoDB change stream",
        "agent_id", agent.ID,
        "type", agent.Type,
        "issue", issueDoc.IssueNumber,
        "milestone", issueDoc.Milestone,
        "column", column.Name)
    
    // Agent will query ArangoDB for issue details and start work
    agent.Start()
}
```

**Architecture Flow**:
```
Gitea Webhook → MVP-WI-001 → ArangoDB → Change Stream → MVP-032 Orchestrator → Creates Agent
```
    columnID, exists := mapping.MilestoneColumnMapping[milestoneName]
    if !exists {
        log.Info("Milestone not mapped to workflow column", "milestone", milestoneName)
        return
    }
    
    // 4. Get workflow and column configuration
    workflow, err := h.getWorkflow(mapping.WorkflowID)
    if err != nil {
        log.Error("Workflow not found", "workflow_id", mapping.WorkflowID, "error", err)
        return
    }
    
    column := workflow.GetColumn(columnID)
    if column == nil {
        log.Error("Column not found in workflow", "column_id", columnID)
        return
    }
    
    // 5. Check if auto-assign is enabled for this column
    if !column.AgentConfig.AutoAssign {
        log.Info("Auto-assign disabled for column", "column", column.Name)
        return
    }
    
    // 6. Check WIP (work-in-progress) limit
    currentAgents, err := h.countActiveAgentsInColumn(workflow.ID, columnID)
    if err != nil {
        log.Error("Failed to count active agents", "error", err)
        return
    }
    
    if currentAgents >= column.AgentConfig.MaxConcurrent {
        log.Warn("WIP limit reached", "column", column.Name, "current", currentAgents, "max", column.AgentConfig.MaxConcurrent)
        h.gitea.CreateIssueComment(
            payload.Repository.Owner.UserName,
            payload.Repository.Name,
            payload.Issue.Index,
            fmt.Sprintf("⏸️ WIP limit reached for '%s' (%d/%d). Issue queued.", column.Name, currentAgents, column.AgentConfig.MaxConcurrent),
        )
        // Add to queue for later processing
        h.queueIssue(workflow.ID, columnID, payload.Issue)
        return
    }
    
    // 7. Get work item definition for this column
    if column.WorkItemDefinitionID == "" {
        log.Info("No work item definition for column", "column", column.Name)
        return
    }
    
    workItemDef, err := h.getWorkItemDefinition(column.WorkItemDefinitionID)
    if err != nil {
        log.Error("Work item definition not found", "id", column.WorkItemDefinitionID, "error", err)
        return
    }
    
    // 8. Create agent instance from work item definition
    go h.createAndStartAgent(payload, workflow, column, workItemDef)
}

func (h *IssueWebhookHandler) createAndStartAgent(
    payload *gitea.IssuePayload,
    workflow *Workflow,
    column *WorkflowColumn,
    workItemDef *WorkItemDefinition,
) {
    ctx := context.Background()
    
    agent, err := h.agentFactory.CreateFromWorkItemDefinition(ctx, &AgentCreationRequest{
        WorkflowID:         workflow.ID,
        ColumnID:           column.ID,
        WorkItemDefinition: workItemDef,
        AgencyID:           workflow.AgencyID,
        TriggerSource:      "gitea_kanban",
        RepositoryURL:      payload.Repository.HTMLURL,
        IssueID:            payload.Issue.Index,
        IssueTitle:         payload.Issue.Title,
        IssueBody:          payload.Issue.Body,
        IssueURL:           payload.Issue.HTMLURL,
        Milestone:          payload.Issue.Milestone.Title,
    })
    
    if err != nil {
        log.Error("Agent creation failed", "error", err)
        h.gitea.CreateIssueComment(
            payload.Repository.Owner.UserName,
            payload.Repository.Name,
            payload.Issue.Index,
            fmt.Sprintf("❌ Failed to create agent: %v", err),
        )
        return
    }
    
    // Start agent execution
    if err := agent.Start(ctx); err != nil {
        log.Error("Agent start failed", "agent_id", agent.ID, "error", err)
        h.gitea.CreateIssueComment(
            payload.Repository.Owner.UserName,
            payload.Repository.Name,
            payload.Issue.Index,
            fmt.Sprintf("❌ Agent failed to start: %v", err),
        )
        return
    }
    
    // Link agent to issue
    h.linkAgentToIssue(agent.ID, payload.Issue.Index, payload.Repository.HTMLURL, workflow.ID, column.ID)
    
    // Post success comment
    h.gitea.CreateIssueComment(
        payload.Repository.Owner.UserName,
        payload.Repository.Name,
        payload.Issue.Index,
        fmt.Sprintf("🤖 **Agent Assigned**\n\n"+
            "- **Stage**: %s\n"+
            "- **Agent**: %s\n"+
            "- **Type**: %s\n"+
            "- **ID**: `%s`\n\n"+
            "The agent will now work on this issue. Updates will be posted here.",
            column.Name, workItemDef.Name, workItemDef.Type, agent.ID),
    )
}

// Find repository-workflow mapping
func (h *IssueWebhookHandler) findRepositoryWorkflowMapping(repoURL string) (*RepositoryWorkflowMapping, error) {
    query := `
        FOR mapping IN repository_workflow_mappings
            FILTER mapping.repository_url == @repoURL
            AND mapping.enabled == true
            LIMIT 1
            RETURN mapping
    `
    
    cursor, err := h.db.Query(context.Background(), query, map[string]interface{}{
        "repoURL": repoURL,
    })
    if err != nil {
        return nil, err
    }
    defer cursor.Close()
    
    var mapping RepositoryWorkflowMapping
    if !cursor.HasMore() {
        return nil, fmt.Errorf("no workflow mapping for repository: %s", repoURL)
    }
    
    cursor.ReadDocument(context.Background(), &mapping)
    return &mapping, nil
}

func (h *IssueWebhookHandler) handleIssueLabeled(payload *gitea.IssuePayload) {
    // Extract work item type from the newly added label
    workItemType := extractWorkItemType([]gitea.Label{*payload.Label})
    if workItemType == "" {
        return // Not a work item type label
    }
    
    // Create agent from work item definition (same as handleIssueOpened)
    h.handleIssueOpened(payload)
}

func (h *IssueWebhookHandler) handleIssueClosed(payload *gitea.IssuePayload) {
    // Find agent instance linked to this issue
    query := `
        FOR agent IN agents
            FILTER agent.trigger_source == "gitea" 
            AND agent.gitea_issue_id == @issueID
            RETURN agent
    `
    
    cursor, err := h.db.Query(context.Background(), query, map[string]interface{}{
        "issueID": payload.Issue.Index,
    })
    if err != nil {
        log.Error("Failed to find agent for issue", "issue", payload.Issue.Index, "error", err)
        return
    }
    defer cursor.Close()
    
    var agent Agent
    if cursor.ReadDocument(context.Background(), &agent) {
        // Stop the agent gracefully
        if err := h.agentRegistry.StopAgent(context.Background(), agent.ID); err != nil {
            log.Error("Failed to stop agent", "agent_id", agent.ID, "error", err)
        }
        
        // Update agent status to completed
        h.db.Collection("agents").UpdateDocument(
            context.Background(),
            agent.ID,
            map[string]interface{}{
                "status":       "completed",
                "completed_at": time.Now(),
                "exit_reason":  "gitea_issue_closed",
            },
        )
    }
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
