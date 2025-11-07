---
title: GitOps Workflow
path: /documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/gitops-workflow.md
---

# GitOps Workflow & Goroutine Execution

Each work item executes in its own goroutine with full GitOps workflow: **branch → commit → push → merge request**.

## Complete Execution Flow

```go
func (e *WorkItemExecutor) Execute(ctx context.Context, issue *gitea.Issue) error {
    // 1. CREATE BRANCH
    branchName := fmt.Sprintf("issue-%d-%s", issue.Index, slugify(issue.Title))
    
    // 2. CLASSIFY WORK TYPE
    workType := e.ClassifyWork(issue) // document, software, proposal, analysis
    
    // 3. EXECUTE BASED ON TYPE
    var content []byte
    var err error
    
    switch workType {
    case "document":
        content, err = e.executeDocumentWork(ctx, issue)
    case "software":
        content, err = e.executeSoftwareWork(ctx, issue)
    case "proposal":
        content, err = e.executeProposalWork(ctx, issue)
    case "analysis":
        content, err = e.executeAnalysisWork(ctx, issue)
    }
    
    if err != nil {
        return fmt.Errorf("execution failed: %w", err)
    }
    
    // 4. GIT OPERATIONS
    repo := e.openGitRepo()
    worktree, _ := repo.Worktree()
    
    // Checkout branch
    worktree.Checkout(&git.CheckoutOptions{Branch: branchName, Create: true})
    
    // Write content to files
    e.writeFiles(content)
    
    // Git add
    worktree.Add(".")
    
    // 5. COMMIT
    commitHash, _ := worktree.Commit(
        fmt.Sprintf("%s: %s\n\nCloses #%d", 
            getCommitType(workType), issue.Title, issue.Index),
        &git.CommitOptions{Author: llmAgentSignature},
    )
    
    // 6. STORE IN ARANGODB
    e.storeCommitInArangoDB(commitHash, issue)
    e.buildCodeDependencyGraph(commitHash)
    
    // 7. PUSH BRANCH
    repo.Push(&git.PushOptions{RefSpecs: []string{branchName}})
    
    // 8. CREATE MERGE REQUEST
    mr := e.gitea.CreateMergeRequest(issue.ProjectID, &CreateMROptions{
        SourceBranch: branchName,
        TargetBranch: "main",
        Title:        fmt.Sprintf("Resolve: %s", issue.Title),
    })
    
    // 9. LINK MR TO ISSUE (ArangoDB graph edge)
    e.createWorkItemCommitEdge(issue.Index, commitHash)
    
    // 10. AUTO-MERGE IF TRUSTED
    if e.shouldAutoMerge(issue, workType) {
        e.gitea.MergePullRequest(mr.ID, MergeOptions{
            Style: "squash",
            DeleteBranch: true,
        })
    }
    
    return nil
}
```

## Parallel Execution

Multiple work items execute in parallel goroutines with concurrency control:

```go
func (e *WorkflowEngine) ExecuteWorkflow(ctx context.Context) error {
    // Get open issues from Gitea
    issues := e.gitea.ListIssues(ListOptions{
        State: "open",
        Labels: []string{"work-item"},
    })
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Max 10 parallel
    
    for _, issue := range issues {
        wg.Add(1)
        semaphore <- struct{}{} // Acquire
        
        go func(iss *gitea.Issue) {
            defer wg.Done()
            defer func() { <-semaphore }() // Release
            
            // Execute with idempotence check
            if !e.alreadyProcessed(iss.Index) {
                err := e.executor.Execute(ctx, iss)
                if err != nil {
                    e.gitea.AddComment(iss.Index, 
                        fmt.Sprintf("❌ Execution failed: %v", err))
                } else {
                    e.markProcessed(iss.Index)
                }
            }
        }(issue)
        
        time.Sleep(2 * time.Second) // Rate limiting
    }
    
    wg.Wait()
    return nil
}
```

## Concurrency Controls

### Idempotence

```go
// Idempotence key generation
func generateIdempotenceKey(issue *gitea.Issue) string {
    return fmt.Sprintf("%d_%s_%s", 
        issue.ProjectID, 
        issue.Index, 
        hashContent(issue.Description))
}

// Check before execution
func (e *Executor) alreadyProcessed(issueIndex int64) bool {
    query := `
        FOR wi IN work_items
            FILTER wi.gitea_issue_id == @issueIndex
            FILTER wi.status IN ["completed", "executing"]
            RETURN wi
    `
    cursor, _ := e.db.Query(ctx, query, map[string]interface{}{
        "issueIndex": issueIndex,
    })
    return cursor.HasMore()
}
```

### Mutex Locks

For operations requiring exclusive access:

```go
type MutexLock struct {
    Key       string    `json:"_key"`      // Resource identifier
    Owner     string    `json:"owner"`     // Work item or agent ID
    AcquiredAt time.Time `json:"acquired_at"`
    TTL       int64     `json:"ttl"`       // Seconds
}

func (e *Executor) acquireLock(resource string) (*MutexLock, error) {
    lock := &MutexLock{
        Key:        resource,
        Owner:      e.workItemID,
        AcquiredAt: time.Now(),
        TTL:        300, // 5 minutes
    }
    
    // Try to insert (fails if exists)
    _, err := e.db.Collection("mutex_locks").CreateDocument(ctx, lock)
    if err != nil {
        return nil, fmt.Errorf("resource locked by another work item")
    }
    
    // Auto-expire after TTL
    go func() {
        time.Sleep(time.Duration(lock.TTL) * time.Second)
        e.db.Collection("mutex_locks").RemoveDocument(ctx, lock.Key)
    }()
    
    return lock, nil
}
```

## Git Operations Detail

### Branch Creation

```go
func (e *Executor) createBranch(branchName string) error {
    repo, _ := git.PlainOpen(e.repoPath)
    headRef, _ := repo.Head()
    
    branchRef := plumbing.NewHashReference(
        plumbing.NewBranchReferenceName(branchName),
        headRef.Hash(),
    )
    
    return repo.Storer.SetReference(branchRef)
}
```

### Commit Creation

```go
func (e *Executor) createCommit(message string, author Signature) (string, error) {
    repo, _ := git.PlainOpen(e.repoPath)
    worktree, _ := repo.Worktree()
    
    hash, err := worktree.Commit(message, &git.CommitOptions{
        Author: &object.Signature{
            Name:  author.Name,
            Email: author.Email,
            When:  time.Now(),
        },
    })
    
    return hash.String(), err
}
```

### Push to Remote

```go
func (e *Executor) pushBranch(branchName string) error {
    repo, _ := git.PlainOpen(e.repoPath)
    
    return repo.Push(&git.PushOptions{
        RemoteName: "origin",
        RefSpecs: []config.RefSpec{
            config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", 
                branchName, branchName)),
        },
        Auth: &http.BasicAuth{
            Username: "git",
            Password: e.giteaToken,
        },
    })
}
```

## Error Handling & Retry

```go
func (e *Executor) executeWithRetry(ctx context.Context, issue *gitea.Issue) error {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := e.Execute(ctx, issue)
        
        if err == nil {
            return nil // Success
        }
        
        if !isRetryableError(err) {
            return err // Non-retryable error
        }
        
        if attempt < maxRetries {
            e.logger.Warnf("Attempt %d failed: %v. Retrying in %v...", 
                attempt, err, backoff)
            time.Sleep(backoff)
            backoff *= 2 // Exponential backoff
        }
    }
    
    return fmt.Errorf("failed after %d attempts", maxRetries)
}

func isRetryableError(err error) bool {
    // Network errors, temporary failures
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "temporary")
}
```

## Execution State Tracking

```go
type WorkflowExecution struct {
    Key         string          `json:"_key"`
    WorkItemKey string          `json:"work_item_key"`
    Status      string          `json:"status"`
    Steps       []ExecutionStep `json:"steps"`
    LLMCalls    []LLMCall      `json:"llm_calls"`
    GitOps      []GitOperation `json:"git_operations"`
    Metrics     ExecutionMetrics `json:"metrics"`
    Error       string          `json:"error,omitempty"`
}

type ExecutionStep struct {
    Step        string    `json:"step"`
    Status      string    `json:"status"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
    Output      string    `json:"output"`
}

// Store execution trace
func (e *Executor) recordStep(execution *WorkflowExecution, step string, status string) {
    execution.Steps = append(execution.Steps, ExecutionStep{
        Step:      step,
        Status:    status,
        StartedAt: time.Now(),
    })
    
    e.db.Collection("workflow_executions").UpdateDocument(
        ctx, execution.Key, execution)
}
```

---

**See Also:**
- [Gitea Integration](./gitea-integration.md) - Webhooks and merge automation
- [Concurrency Controls](./data-models.md#concurrency) - Detailed locking mechanisms
- [Observability](./observability.md) - Execution monitoring
