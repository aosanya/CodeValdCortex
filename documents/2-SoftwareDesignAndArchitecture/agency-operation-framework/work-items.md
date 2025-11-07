---
title: Work Items Specification
path: /documents/2-SoftwareDesignAndArchitecture/work-items.md
---

# Work Items (WI) — Design Specification

This document specifies the Work Items subsystem used by CodeValdCortex agency operations. It complements the existing RACI and Goals documentation by describing:

- Work item types and their required fields
- GitOps-based execution model with Gitea integration
- Graph-based storage in ArangoDB for relationships and analytics
- Canonical lifecycle and SLA/SLO behaviour
- LLM-powered code/document generation via goroutines
- Git operations as primary work execution mechanism
- Multi-dimensional graph queries (code, commits, agents, work items)
- Assignment and routing rules (declarative)
- Concurrency, idempotence and reentrancy contracts
- Knowledge graph for expertise tracking

## Architecture Overview

CodeValdCortex implements a **GitOps-first** work item execution model:

```
GitLab/Gitea Issues (Work Item Triggers)
    ↓ Webhooks
CodeValdCortex Workflow Engine (Goroutines)
    ↓ Git Operations (branch, commit, push, MR)
ArangoDB Multi-Graph Storage
    ├── Git Objects (commits, trees, blobs)
    ├── Work Item Metadata
    ├── Code Dependency Graph
    ├── Knowledge Graph (who knows what)
    └── Workflow Execution State
```

Where applicable the document includes example Go interfaces and implementation notes for storage (ArangoDB multi-model) and runtime enforcement.

## 1. Goals and Acceptance

Acceptance criteria for this specification:

- ✅ GitOps workflow: Issues → Goroutines → Git operations → Merge requests
- ✅ Work Items stored in ArangoDB with full graph relationships
- ✅ Git objects (commits, trees, blobs) stored in ArangoDB for advanced queries
- ✅ LLM agents generate code/documents based on issue requirements
- ✅ Multi-dimensional graph traversal (code deps, commits, work items, agents)
- ✅ Knowledge graph tracks expertise (who knows what code)
- ✅ Lifecycle state machine enforced with audit trail
- ✅ SLA/SLO fields exist and breach actions are actionable
- ✅ Searchable code content with full-text and semantic search
- ✅ Impact analysis via graph queries (what breaks when code changes)
- ✅ Concurrency controls prevent duplicate, conflicting external effects
- ✅ Integration with Gitea for issue tracking and code review UI

## 2. Work Item Types & Contracts

Work items are triggered by **Gitea/GitLab issues** with specific labels. Each work item type maps to a specific execution pattern that runs in a goroutine.

### Work Item Execution Model

```go
type WorkItemExecutor interface {
    // Execute work based on Gitea issue
    Execute(ctx context.Context, issue *gitea.Issue) error
    
    // Classify work type from issue labels/content
    ClassifyWork(issue *gitea.Issue) WorkType
    
    // Perform git operations (branch, commit, push, MR)
    ExecuteGitOps(ctx context.Context, workType WorkType, content []byte) error
}
```

### Core Work Types

Each type represents a different execution pattern in goroutines:

#### 1. **Document Work** (`labels: documentation, docs`)

**Execution Pattern:**
```go
goroutine:
  1. Read Gitea issue requirements
  2. Fetch existing document from ArangoDB git storage
  3. LLM generates/updates document content
  4. Create git branch
  5. Commit document to branch
  6. Push branch to Gitea
  7. Create merge request
  8. Link MR back to issue
```

**ArangoDB Storage:**
```go
type GitBlob struct {
    Key         string `json:"_key"`        // SHA-1 hash
    Type        string `json:"type"`        // "blob"
    Content     []byte `json:"content"`     // Raw content
    ContentText string `json:"content_text"` // Searchable text
    Language    string `json:"language"`    // markdown, restructuredtext
    Path        string `json:"path"`        // docs/architecture/...
    RepoID      string `json:"repo_id"`
    Metadata    map[string]interface{} `json:"metadata"` // Frontmatter, etc.
}
```

#### 2. **Software Work** (`labels: feature, bug, enhancement, code`)

**Execution Pattern:**
```go
goroutine:
  1. Parse code requirements from issue
  2. LLM generates implementation plan
  3. LLM generates code files + tests
  4. Create git branch
  5. Commit all files
  6. Build dependency graph in ArangoDB
  7. Push branch to Gitea
  8. Create merge request with CI pipeline
  9. Auto-merge if CI passes and trusted
```

**ArangoDB Storage:**
```go
// Code file as blob
type CodeBlob struct {
    Key         string   `json:"_key"`
    Content     []byte   `json:"content"`
    ContentText string   `json:"content_text"` // For search
    Language    string   `json:"language"`     // go, javascript, python
    Path        string   `json:"path"`
    Metadata    map[string]interface{} `json:"metadata"` // package, imports, symbols
}

// Code dependencies as edges
type CodeDependency struct {
    From    string   `json:"_from"` // git_objects/file1_hash
    To      string   `json:"_to"`   // git_objects/file2_hash
    Type    string   `json:"type"`  // import, reference, call
    Symbols []string `json:"symbols"`
}
```

#### 3. **Proposal Work** (`labels: proposal, business, rfp`)

**Execution Pattern:**
```go
goroutine:
  1. Load existing proposal from ArangoDB
  2. Parse update requirements from issue
  3. LLM updates proposal sections
  4. Generate PDF version
  5. Create git branch
  6. Commit markdown + PDF
  7. Push and create MR
  8. Notify stakeholders
```

#### 4. **Analysis Work** (`labels: investigation, research, analysis`)

**Execution Pattern:**
```go
goroutine:
  1. Gather context from ArangoDB (related commits, code, issues)
  2. LLM performs analysis using graph queries
  3. Generate analysis report
  4. Store findings in ArangoDB
  5. Comment analysis on issue
  6. Link related work items via graph edges
```

### Work Item Data Model

```go
// Work item metadata in ArangoDB
type WorkItem struct {
    Key              string    `json:"_key"`
    AgencyID         string    `json:"agency_id"`
    GiteaIssueID     int64     `json:"gitea_issue_id"`
    GiteaIssueURL    string    `json:"gitea_issue_url"`
    WorkType         string    `json:"work_type"`        // document, software, proposal, analysis
    Status           string    `json:"status"`           // pending, executing, completed, failed
    ExecutionID      string    `json:"execution_id"`
    LLMAgentID       string    `json:"llm_agent_id"`
    Priority         int       `json:"priority"`
    StartedAt        time.Time `json:"started_at"`
    CompletedAt      *time.Time `json:"completed_at,omitempty"`
    IdempotenceKey   string    `json:"idempotence_key"`  // Prevent duplicate execution
    Metadata         map[string]interface{} `json:"metadata"`
}

// Link work items to commits (graph edge)
type WorkItemCommit struct {
    From      string `json:"_from"` // work_items/wi_123
    To        string `json:"_to"`   // git_commits/abc123
    Type      string `json:"type"`  // implements, fixes, refactors
    CreatedBy string `json:"created_by"` // agent ID
    Automated bool   `json:"automated"`
}
```

## 3. Git Storage in ArangoDB

All git objects are stored in ArangoDB for **searchable, graph-enabled git operations**.

### Git Object Collections

```go
// Git commit
type GitCommit struct {
    Key       string    `json:"_key"`     // SHA-1 hash
    Type      string    `json:"type"`     // "commit"
    Tree      string    `json:"tree"`     // Tree SHA-1
    Parents   []string  `json:"parents"`  // Parent commit SHA-1s
    Author    Signature `json:"author"`
    Committer Signature `json:"committer"`
    Message   string    `json:"message"`
    IssueRefs []int64   `json:"issue_refs"` // Linked Gitea issues
    RepoID    string    `json:"repo_id"`
    CreatedAt time.Time `json:"created_at"`
}

// Git tree (directory snapshot)
type GitTree struct {
    Key     string      `json:"_key"`  // SHA-1 hash
    Type    string      `json:"type"`  // "tree"
    Entries []TreeEntry `json:"entries"`
    RepoID  string      `json:"repo_id"`
}

type TreeEntry struct {
    Mode string `json:"mode"`      // File permissions
    Type string `json:"type"`      // blob or tree
    Hash string `json:"hash"`      // Object SHA-1
    Name string `json:"name"`      // File/dir name
}

// Git blob (searchable file content)
type GitBlob struct {
    Key         string                 `json:"_key"`
    Type        string                 `json:"type"`
    Content     []byte                 `json:"content"`
    ContentText string                 `json:"content_text"` // ✅ SEARCHABLE
    Language    string                 `json:"language"`
    MimeType    string                 `json:"mime_type"`
    Path        string                 `json:"path"`
    RepoID      string                 `json:"repo_id"`
    Metadata    map[string]interface{} `json:"metadata"`
    IndexedAt   time.Time              `json:"indexed_at"`
}

// Git reference (branch/tag)
type GitRef struct {
    Key    string `json:"_key"`     // refs/heads/main
    Name   string `json:"name"`     // main
    Type   string `json:"type"`     // branch or tag
    Target string `json:"target"`   // Commit SHA-1
    RepoID string `json:"repo_id"`
}
```

### Graph Collections (Edges)

```go
// Commit parent relationships
type CommitParent struct {
    From string `json:"_from"` // git_commits/abc123
    To   string `json:"_to"`   // git_commits/parent_hash
    Type string `json:"type"`  // parent, merge-parent
}

// Code dependencies
type CodeDependency struct {
    From     string   `json:"_from"` // git_objects/file1_hash
    To       string   `json:"_to"`   // git_objects/file2_hash
    Type     string   `json:"type"`  // import, reference, call
    Symbols  []string `json:"symbols"`
    Strength float64  `json:"strength"` // Coupling strength
}

// Work item → commit links
type WorkItemCommit struct {
    From      string `json:"_from"` // work_items/wi_123
    To        string `json:"_to"`   // git_commits/abc123
    CreatedBy string `json:"created_by"`
    Automated bool   `json:"automated"`
}

// Agent → code expertise
type AgentCodeExpertise struct {
    From       string    `json:"_from"` // agents/agent_123
    To         string    `json:"_to"`   // git_objects/file_hash
    Commits    []string  `json:"commits"`
    Expertise  float64   `json:"expertise"` // 0-1 score
    LastTouch  time.Time `json:"last_touch"`
}
```

### Named Graphs

ArangoDB organizes these edges into named graphs for efficient traversal:

- **`commit_graph`**: Commit history and ancestry
- **`code_graph`**: Code dependencies and relationships
- **`workflow_graph`**: Work items → commits → agents
- **`knowledge_graph`**: Who knows what code (expertise)

## 4. Goroutine Execution Model

Each work item executes in its own goroutine with full GitOps workflow:

### Execution Flow

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

### Parallel Execution

Multiple work items execute in parallel goroutines:

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

## 5. Graph-Powered Queries

ArangoDB's graph capabilities enable powerful queries across the entire codebase:

### Commit History Queries

```go
// Get full commit ancestry
query := `
    FOR v, e, p IN 0..100 OUTBOUND
        @startCommit
        GRAPH 'commit_graph'
        RETURN {commit: v, path: p.vertices[* RETURN CURRENT._key]}
`

// Find merge base (common ancestor)
query := `
    LET ancestors1 = (FOR v IN 0..100 OUTBOUND @commit1 GRAPH 'commit_graph' RETURN v._key)
    LET ancestors2 = (FOR v IN 0..100 OUTBOUND @commit2 GRAPH 'commit_graph' RETURN v._key)
    LET common = INTERSECTION(ancestors1, ancestors2)
    FOR commit IN git_commits
        FILTER commit._key IN common
        SORT commit.created_at DESC
        LIMIT 1
        RETURN commit._key
`
```

### Code Dependency Queries

```go
// Find all files that depend on this file
query := `
    FOR v, e, p IN 1..5 INBOUND
        @fileHash
        GRAPH 'code_graph'
        RETURN {
            file: v.path,
            dependency_type: e.type,
            symbols_used: e.symbols,
            depth: LENGTH(p.edges)
        }
`

// Find circular dependencies
query := `
    FOR v, e, p IN 1..10 OUTBOUND
        ANY v IN git_objects
        GRAPH 'code_graph'
        FILTER p.vertices[0]._id == p.vertices[-1]._id
        FILTER LENGTH(p.vertices) > 2
        RETURN DISTINCT p.vertices[* RETURN CURRENT.path]
`

// Impact analysis: What breaks when I change this file?
query := `
    LET directDeps = (FOR v IN 1..1 INBOUND @fileHash GRAPH 'code_graph' RETURN v)
    LET transitiveDeps = (FOR v IN 2..5 INBOUND @fileHash GRAPH 'code_graph' RETURN DISTINCT v)
    LET testFiles = (
        FOR v IN transitiveDeps
            FILTER LIKE(v.path, "%_test.%") OR LIKE(v.path, "%.test.%")
            RETURN v
    )
    RETURN {
        direct_dependents: LENGTH(directDeps),
        transitive_dependents: LENGTH(transitiveDeps),
        affected_files: UNION(directDeps, transitiveDeps),
        test_files: testFiles,
        impact_score: LENGTH(directDeps) * 10 + LENGTH(transitiveDeps)
    }
`
```

### Work Item Tracing

```go
// Trace work item to all affected code
query := `
    FOR wi IN work_items
        FILTER wi._key == @workItemID
        
        // Get commits implementing this work item
        LET commits = (FOR c IN 1..1 OUTBOUND wi._id work_item_commits RETURN c)
        
        // Get files changed in those commits
        LET files = (
            FOR commit IN commits
                LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
                FOR entry IN tree.entries
                    RETURN DISTINCT entry
        )
        
        // Get dependencies of changed files
        LET dependencies = (
            FOR file IN files
                FOR dep IN 1..2 OUTBOUND
                    CONCAT('git_objects/', file.hash)
                    GRAPH 'code_graph'
                    RETURN DISTINCT dep.path
        )
        
        RETURN {
            work_item: wi,
            commits: commits,
            files_changed: files,
            dependencies_affected: dependencies,
            total_impact: LENGTH(files) + LENGTH(dependencies)
        }
`

// Find which work items touched this code
query := `
    FOR file IN git_objects
        FILTER file._key == @fileHash
        FOR commit IN git_commits
            LET tree = DOCUMENT(CONCAT('git_objects/', commit.tree))
            FILTER @fileHash IN tree.entries[* RETURN CURRENT.hash]
            FOR wi IN 1..1 INBOUND commit._id work_item_commits
                RETURN DISTINCT wi
`
```

### Knowledge Graph Queries

```go
// Find expert for a file
query := `
    FOR v, e IN 1..1 INBOUND
        @fileHash
        GRAPH 'knowledge_graph'
        SORT e.expertise DESC, e.last_touch DESC
        LIMIT 5
        RETURN {
            agent: v,
            commits: LENGTH(e.commits),
            expertise: e.expertise,
            last_contribution: e.last_touch,
            recency_score: DATE_DIFF(e.last_touch, DATE_NOW(), 'days') / -365
        }
`

// Find what code an agent knows best
query := `
    FOR v, e IN 1..1 OUTBOUND
        @agentID
        GRAPH 'knowledge_graph'
        SORT e.expertise DESC
        LIMIT 20
        RETURN {
            file: v.path,
            language: v.language,
            expertise: e.expertise,
            commits: LENGTH(e.commits),
            last_touch: e.last_touch
        }
`
```

### Full-Text Search

```go
// Search all code content
query := `
    FOR doc IN git_content_search
        SEARCH ANALYZER(
            doc.content_text IN TOKENS(@searchTerm, "text_en_code"),
            "text_en_code"
        )
        FILTER doc.repo_id == @repoID
        LET score = BM25(doc)
        SORT score DESC
        LIMIT 50
        RETURN {
            path: doc.path,
            language: doc.language,
            score: score,
            snippet: SUBSTRING(doc.content_text, 0, 200)
        }
`

// Search with line context
query := `
    FOR doc IN git_content_search
        SEARCH ANALYZER(doc.content_text IN TOKENS(@searchTerm, "text_en_code"), "text_en_code")
        LET lines = SPLIT(doc.content_text, "\n")
        LET matches = (
            FOR i IN 0..LENGTH(lines)-1
                FILTER LIKE(lines[i], CONCAT("%", @searchTerm, "%"))
                RETURN {
                    line_number: i + 1,
                    line: lines[i],
                    context_before: SLICE(lines, MAX([0, i-2]), 2),
                    context_after: SLICE(lines, i+1, 2)
                }
        )
        FILTER LENGTH(matches) > 0
        RETURN {path: doc.path, matches: matches}
`
```

## 6. Concurrency Controls & Idempotence

To prevent duplicate work or conflicting external effects, Work Items must adhere to concurrency contracts.

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

### Mutex Scopes

For operations requiring exclusive access (e.g., modifying shared resources):

```go
type MutexLock struct {
    Key       string    `json:"_key"`      // Resource identifier
    Owner     string    `json:"owner"`     // Work item or agent ID
    AcquiredAt time.Time `json:"acquired_at"`
    TTL       int64     `json:"ttl"`       // Seconds
}

// Acquire lock before git operations on same branch
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

### Reentrancy

Work items can be safely retried:

```go
func (e *Executor) Execute(ctx context.Context, issue *gitea.Issue) error {
    // Check idempotence
    idempotenceKey := generateIdempotenceKey(issue)
    
    query := `
        FOR wi IN work_items
            FILTER wi.idempotence_key == @key
            FILTER wi.status == "completed"
            RETURN wi
    `
    cursor, _ := e.db.Query(ctx, query, map[string]interface{}{
        "key": idempotenceKey,
    })
    
    if cursor.HasMore() {
        // Already completed, return cached result
        var existing WorkItem
        cursor.ReadDocument(ctx, &existing)
        return e.returnCachedResult(existing)
    }
    
    // Execute fresh
    return e.executeWork(ctx, issue, idempotenceKey)
}
```

## 7. Gitea Integration & Merge Automation

### Webhook Processing

```go
func (h *WebhookHandler) HandleIssueEvent(c *gin.Context) {
    var payload gitea.IssuePayload
    c.BindJSON(&payload)
    
    switch payload.Action {
    case "opened":
        // New issue created - trigger work item
        if hasLabel(payload.Issue, "work-item") {
            go h.executor.Execute(c.Request.Context(), payload.Issue)
        }
        
    case "labeled":
        // Label added - check if it's a work-item label
        if payload.Label.Name == "work-item" {
            go h.executor.Execute(c.Request.Context(), payload.Issue)
        }
        
    case "closed":
        // Issue closed - update work item status
        h.updateWorkItemStatus(payload.Issue.Index, "completed")
    }
}

func (h *WebhookHandler) HandlePREvent(c *gin.Context) {
    var payload gitea.PullRequestPayload
    c.BindJSON(&payload)
    
    switch payload.Action {
    case "synchronized":
        // PR updated - re-run validations
        go h.validateAndMerge(payload.PullRequest)
        
    case "labeled":
        if payload.Label.Name == "ready-to-merge" {
            go h.validateAndMerge(payload.PullRequest)
        }
    }
}
```

### Merge Strategies

```go
func (e *Executor) getMergeStrategy(issue *gitea.Issue) MergeStrategy {
    for _, label := range issue.Labels {
        switch label.Name {
        case "documentation", "docs":
            return MergeStrategy{
                Style: "squash",
                AutoMerge: true,
                RequireReviews: 1,
                RequireCI: false,
            }
            
        case "automated", "llm-generated":
            return MergeStrategy{
                Style: "squash",
                AutoMerge: true,
                RequireReviews: 0,
                RequireCI: true,
                RequireChecks: []string{"lint", "test", "security-scan"},
            }
            
        case "feature", "enhancement":
            return MergeStrategy{
                Style: "merge",
                AutoMerge: false,
                RequireReviews: 2,
                RequireCI: true,
            }
        }
    }
    
    // Default: conservative
    return MergeStrategy{
        Style: "merge",
        AutoMerge: false,
        RequireReviews: 2,
        RequireCI: true,
    }
}

func (e *Executor) validateAndMerge(pr *gitea.PullRequest) error {
    // 1. Check if mergeable
    if !pr.Mergeable {
        return fmt.Errorf("has merge conflicts")
    }
    
    // 2. Check CI status
    status := e.gitea.GetCombinedStatus(pr.Head.Sha)
    if status.State != "success" {
        return fmt.Errorf("CI not passing: %s", status.State)
    }
    
    // 3. Check reviews
    reviews := e.gitea.ListPullReviews(pr.Index)
    approvals := countApprovals(reviews)
    
    strategy := e.getMergeStrategy(pr.Issue)
    if approvals < strategy.RequireReviews {
        return fmt.Errorf("not enough approvals: %d/%d", 
            approvals, strategy.RequireReviews)
    }
    
    // 4. Auto-merge if allowed
    if strategy.AutoMerge {
        e.gitea.MergePullRequest(pr.Index, gitea.MergePullRequestOption{
            Style: strategy.Style,
            DeleteBranchAfterMerge: true,
        })
    }
    
    return nil
}
```

### Batch Merge for High Volume

```go
func (e *Executor) BatchMerge(ctx context.Context) error {
    // Get all ready-to-merge PRs
    prs := e.gitea.ListPullRequests(ListOptions{
        State: "open",
        Labels: []string{"ready-to-merge", "automated"},
    })
    
    // Sort by priority
    sort.Slice(prs, func(i, j int) bool {
        return getPriority(prs[i]) > getPriority(prs[j])
    })
    
    // Merge in batches
    for i := 0; i < len(prs); i += 5 {
        batch := prs[i:min(i+5, len(prs))]
        
        for _, pr := range batch {
            if err := e.validateAndMerge(pr); err != nil {
                e.gitea.AddComment(pr.Index, 
                    fmt.Sprintf("❌ Merge failed: %v", err))
            }
            time.Sleep(2 * time.Second)
        }
        
        time.Sleep(10 * time.Second) // Between batches
    }
    
    return nil
}
```

## 8. LLM Integration

### LLM-Powered Content Generation

Each work type uses LLM differently:

```go
type LLMAgent struct {
    client   *openai.Client
    model    string
    context  *AgencyContext
}

// Document generation
func (a *LLMAgent) GenerateDocument(issue *gitea.Issue, existing string) (string, error) {
    prompt := fmt.Sprintf(`
    Update this document based on issue requirements:
    
    File: %s
    Current Content:
    ---
    %s
    ---
    
    Issue #%d: %s
    Requirements:
    %s
    
    Instructions:
    1. Review and update relevant sections
    2. Maintain existing structure
    3. Add new sections if needed
    4. Return COMPLETE updated document
    `, issue.Path, existing, issue.Index, issue.Title, issue.Description)
    
    return a.client.Complete(context.Background(), prompt)
}

// Code generation
func (a *LLMAgent) GenerateCode(issue *gitea.Issue) (*CodeGeneration, error) {
    // Step 1: Implementation plan
    planPrompt := fmt.Sprintf(`
    Create implementation plan for:
    
    Issue #%d: %s
    Requirements: %s
    
    Provide JSON with:
    {
        "files": [{"path": "...", "action": "create|modify", "description": "..."}],
        "tests": [{"path": "...", "description": "..."}],
        "dependencies": ["go get ..."],
        "summary": "..."
    }
    `, issue.Index, issue.Title, issue.Description)
    
    planJSON := a.client.Complete(context.Background(), planPrompt)
    plan := parsePlan(planJSON)
    
    // Step 2: Generate code for each file
    var generatedFiles []GeneratedFile
    for _, file := range plan.Files {
        existing := a.readExisting(file.Path)
        
        codePrompt := fmt.Sprintf(`
        %s file: %s
        
        Existing Code:
        ---
        %s
        ---
        
        Requirements: %s
        
        Generate complete, working Go code with:
        1. Proper error handling
        2. Godoc comments
        3. Unit tests
        4. Following best practices
        `, file.Action, file.Path, existing, file.Description)
        
        code := a.client.Complete(context.Background(), codePrompt)
        generatedFiles = append(generatedFiles, GeneratedFile{
            Path: file.Path,
            Content: code,
        })
    }
    
    return &CodeGeneration{
        Plan: plan,
        Files: generatedFiles,
    }, nil
}
```

### Agency Context Integration

LLM has full access to agency context from ArangoDB:

```go
type AgencyContext struct {
    Agency    *Agency
    Goals     []*Goal
    WorkItems []*WorkItem
    Roles     []*Role
    CodeGraph *CodeGraph
}

func (a *LLMAgent) loadContext(agencyID string) (*AgencyContext, error) {
    query := `
        LET agency = DOCUMENT(CONCAT('agencies/', @agencyID))
        
        LET goals = (
            FOR goal IN goals
                FILTER goal.agency_id == @agencyID
                RETURN goal
        )
        
        LET workItems = (
            FOR wi IN work_items
                FILTER wi.agency_id == @agencyID
                RETURN wi
        )
        
        LET roles = (
            FOR role IN roles
                FILTER role.agency_id == @agencyID
                RETURN role
        )
        
        RETURN {
            agency: agency,
            goals: goals,
            work_items: workItems,
            roles: roles
        }
    `
    
    cursor, _ := a.db.Query(context.Background(), query, map[string]interface{}{
        "agencyID": agencyID,
    })
    
    var ctx AgencyContext
    cursor.ReadDocument(context.Background(), &ctx)
    return &ctx, nil
}
```

## 9. APIs & Implementation

### Storage Collections (ArangoDB)

**Document Collections:**
- `work_items` — Work item metadata
- `git_objects` — Git blobs, trees (searchable)
- `git_commits` — Git commits
- `git_refs` — Git branches and tags
- `agencies` — Agency configurations
- `agents` — LLM agents
- `workflow_executions` — Execution state

**Edge Collections:**
- `commit_parents` — Commit graph
- `code_dependencies` — Code dependency graph
- `work_item_commits` — Work items → commits
- `agent_code_expertise` — Knowledge graph
- `issue_dependencies` — Work item dependencies

**Named Graphs:**
- `commit_graph` — Git history
- `code_graph` — Code dependencies
- `workflow_graph` — Work items, commits, agents
- `knowledge_graph` — Expertise tracking

### API Endpoints

```
# Work Items
GET    /api/v1/agencies/:id/work-items
POST   /api/v1/agencies/:id/work-items          # Trigger from Gitea issue
GET    /api/v1/work-items/:id
PUT    /api/v1/work-items/:id/status

# Git Operations
GET    /api/v1/repos/:id/commits
GET    /api/v1/repos/:id/commits/:hash
GET    /api/v1/repos/:id/tree/:hash
GET    /api/v1/repos/:id/blob/:hash

# Graph Queries
GET    /api/v1/graph/commit-history/:hash
GET    /api/v1/graph/dependencies/:fileHash
GET    /api/v1/graph/impact-analysis/:fileHash
GET    /api/v1/graph/experts/:fileHash
GET    /api/v1/graph/work-item-trace/:id

# Search
GET    /api/v1/search/code?q=term&language=go
GET    /api/v1/search/commits?q=term
GET    /api/v1/search/work-items?q=term

# Webhooks
POST   /api/v1/webhooks/gitea/issues
POST   /api/v1/webhooks/gitea/pull-requests
```

## 10. Observability & Metrics

### Key Metrics

```go
type WorkItemMetrics struct {
    TotalCreated      int64         `json:"total_created"`
    TotalCompleted    int64         `json:"total_completed"`
    TotalFailed       int64         `json:"total_failed"`
    ByType            map[string]int `json:"by_type"`
    AvgExecutionTime  time.Duration `json:"avg_execution_time"`
    SuccessRate       float64       `json:"success_rate"`
    AutoMergeRate     float64       `json:"auto_merge_rate"`
}

// Query metrics from ArangoDB
query := `
    FOR wi IN work_items
        FILTER wi.agency_id == @agencyID
        COLLECT 
            type = wi.work_type,
            status = wi.status
        AGGREGATE count = COUNT()
        RETURN {type, status, count}
`
```

### Execution Traces

Store detailed execution traces in ArangoDB:

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

type LLMCall struct {
    Model     string    `json:"model"`
    Prompt    string    `json:"prompt"`
    Response  string    `json:"response"`
    Tokens    int       `json:"tokens"`
    Cost      float64   `json:"cost"`
    Duration  time.Duration `json:"duration"`
}

type GitOperation struct {
    Type      string    `json:"type"` // branch, commit, push, merge
    Hash      string    `json:"hash,omitempty"`
    Branch    string    `json:"branch,omitempty"`
    MRURL     string    `json:"mr_url,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Graph Analytics

```go
// File importance (PageRank-style)
query := `
    FOR file IN git_objects
        FILTER file.type == "blob"
        LET inbound = LENGTH(FOR v IN 1..1 INBOUND file._id code_dependencies RETURN v)
        LET outbound = LENGTH(FOR v IN 1..1 OUTBOUND file._id code_dependencies RETURN v)
        LET importance = inbound * 2 + outbound * 0.5
        SORT importance DESC
        LIMIT 50
        RETURN {file: file.path, importance, dependents: inbound, dependencies: outbound}
`

// Agent productivity
query := `
    FOR agent IN agents
        LET commits = (FOR c IN git_commits FILTER c.author.email == agent.email RETURN c)
        LET expertise = (FOR e IN 1..1 OUTBOUND agent._id agent_code_expertise RETURN e)
        RETURN {
            agent: agent.name,
            total_commits: LENGTH(commits),
            files_touched: LENGTH(expertise),
            avg_expertise: AVG(expertise[* RETURN CURRENT.expertise])
        }
`
```

## 11. Deployment Architecture

### System Components

```yaml
# docker-compose.yml
version: "3"

services:
  # Gitea - Git operations and issue tracking
  gitea:
    image: gitea/gitea:latest
    ports:
      - "3000:3000"
    environment:
      - GITEA__database__DB_TYPE=postgres
      - GITEA__webhook__ALLOWED_HOST_LIST=*
    volumes:
      - ./gitea:/data
  
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=gitea
      - POSTGRES_USER=gitea
      - POSTGRES_PASSWORD=gitea
    volumes:
      - ./postgres:/var/lib/postgresql/data
  
  # ArangoDB - Graph storage and analytics
  arangodb:
    image: arangodb:latest
    ports:
      - "8529:8529"
    environment:
      - ARANGO_ROOT_PASSWORD=openSesame
    volumes:
      - ./arangodb:/var/lib/arangodb3
  
  # CodeValdCortex - Workflow orchestration
  codevaldcortex:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GITEA_URL=http://gitea:3000
      - GITEA_TOKEN=${GITEA_TOKEN}
      - ARANGODB_URL=http://arangodb:8529
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - gitea
      - arangodb
```

### Resource Requirements

**Gitea + PostgreSQL**: ~2GB RAM  
**ArangoDB**: ~2-4GB RAM (depending on repo size)  
**CodeValdCortex**: ~2GB RAM  
**Total**: ~6-8GB RAM

Can run on $40-60/month VPS or local development machine.

## 12. Examples & Use Cases

### Example 1: Documentation Update

**Gitea Issue #42**: "Update architecture documentation with workflow design"  
**Labels**: `documentation`, `work-item`

**Execution Flow**:
1. Webhook triggers goroutine
2. Classify as "document" work
3. LLM reads existing `docs/architecture/workflow.md` from ArangoDB
4. LLM generates updated content
5. Create branch `issue-42-update-architecture`
6. Commit updated markdown file
7. Push to Gitea
8. Create MR with link to issue
9. Auto-merge (docs don't need review)
10. Close issue #42

**Time**: ~30 seconds

### Example 2: Feature Implementation

**Gitea Issue #87**: "Implement user authentication API endpoint"  
**Labels**: `feature`, `backend`, `work-item`

**Execution Flow**:
1. Webhook triggers goroutine
2. Classify as "software" work
3. LLM generates implementation plan
4. LLM generates code files:
   - `internal/api/handlers/auth_handler.go`
   - `internal/api/middleware/jwt.go`
   - `internal/api/handlers/auth_handler_test.go`
5. Build code dependency graph in ArangoDB
6. Create branch `issue-87-user-auth`
7. Commit all files
8. Push to Gitea
9. Create MR with CI pipeline
10. CI runs tests → passes
11. Auto-merge (trusted code generation)
12. Close issue #87

**Time**: ~2-3 minutes

### Example 3: Impact Analysis Query

**Question**: "What breaks if I change `user_service.go`?"

**Graph Query**:
```go
result := repo.AnalyzeChangeImpact("abc123_user_service_hash")

// Returns:
{
    direct_dependents: 5,           // 5 files import this
    transitive_dependents: 23,      // 23 files indirectly depend on it
    affected_files: [...]           // List of all affected files
    test_files: [...]               // Related test files
    impact_score: 73                // High impact
}
```

**Action**: High impact score → require 2 reviewers before merge

### Example 4: Find Code Expert

**Question**: "Who should review changes to the payment processor?"

**Graph Query**:
```go
experts := repo.FindCodeExpert("payment_processor_hash")

// Returns:
[
    {agent: "alice@example.com", expertise: 0.9, commits: 45, last_touch: "2d ago"},
    {agent: "bob@example.com", expertise: 0.6, commits: 12, last_touch: "1w ago"},
]
```

**Action**: Auto-assign Alice as reviewer

### Example 5: Circular Dependency Detection

**Graph Query**:
```go
cycles := repo.FindCircularDependencies()

// Returns:
[
    ["pkg/auth/service.go", "pkg/user/service.go", "pkg/auth/service.go"],
    ["internal/api/handler.go", "internal/db/repo.go", "internal/api/handler.go"]
]
```

**Action**: Create issues to break circular dependencies

## 13. Role Taxonomy Integration

Work items are executed by LLM agents with well-defined capabilities, autonomy levels, and constraints. The complete roles taxonomy is documented separately.

**See**: [Role Taxonomy](./role-taxonomy.md) for comprehensive documentation including:

- **Role Classifications**: LLM Agents for different work types (Document, Code, Analysis)
- **Skills & Tools Contract**: OpenAI API, Gitea SDK, ArangoDB queries
- **Autonomy Levels (L0-L4)**: From manual to full autonomy with policy-bound action scopes
- **Budgeting**: Token/$ budgets, LLM call quotas, cost tracking
- **Data Boundaries**: Access to git repos, ArangoDB collections, Gitea projects
- **Safety Constraints**: Allowed git operations, merge policies, review requirements

### Agent Type Selection for Work Items

Work item execution considers:

1. **Work Type**: Document → DocumentAgent, Code → CodeAgent
2. **Autonomy Requirements**: Trusted labels → auto-merge, else require review
3. **Budget Constraints**: Track LLM tokens and costs per work item
4. **Data Access Needs**: Access to specific repos and collections
5. **Safety Requirements**: Merge policies based on impact analysis

**Assignment Algorithm**:
```
1. Classify work type from Gitea issue labels
2. Select appropriate LLM agent type
3. Load agency context from ArangoDB
4. Verify budget availability for estimated tokens
5. Execute work in goroutine with git operations
6. Apply merge policy based on work type and impact
```

## 14. Benefits Summary

### GitOps Advantages
✅ **Version Control**: All changes tracked in git  
✅ **Code Review**: MR workflow for quality  
✅ **Rollback**: Easy git revert  
✅ **Audit Trail**: Complete git history  
✅ **Collaboration**: Humans can modify agent-created MRs  

### ArangoDB Graph Advantages
✅ **Rich Queries**: Multi-dimensional graph traversal  
✅ **Impact Analysis**: Know exactly what breaks  
✅ **Knowledge Graph**: Auto-find experts  
✅ **Code Search**: Full-text + semantic search  
✅ **Dependency Tracking**: Visualize code relationships  
✅ **Fast Analytics**: Real-time metrics and insights  

### LLM Integration Advantages
✅ **Automated Execution**: Issues → working code/docs  
✅ **Context-Aware**: Uses full agency context  
✅ **Multi-Modal**: Handles docs, code, proposals  
✅ **High Throughput**: Parallel goroutines  
✅ **Cost Tracking**: Monitor LLM usage  

### Overall System Benefits
✅ **Low Latency**: Issue → merged code in minutes  
✅ **High Volume**: Handle 100s of issues/day  
✅ **Quality**: CI + optional review gates  
✅ **Transparency**: Full visibility in Gitea  
✅ **Scalability**: Lightweight (6-8GB total)  
✅ **Extensibility**: Easy to add new work types  

---

**Last Updated**: November 7, 2025  
**Architecture**: GitOps + ArangoDB Multi-Graph + LLM Agents

## 13. Traceability & Validation

To ensure complete traceability from work items to agent actions to artifacts, the system must maintain explicit linkage documents and validate chains for completeness.

### 13.1 Traceability Schema

```typescript
interface TraceabilityValidation {
  validationId: string;
  timestamp: ISODateTime;
  
  checks: {
    goalToWorkItemLinkage: {
      passed: boolean;
      orphanedGoals: string[];  // Goals with no work items
    };
    
    workItemToActionLinkage: {
      passed: boolean;
      orphanedWorkItems: string[];  // Work items with no agent actions
    };
    
    actionToArtifactLinkage: {
      passed: boolean;
      orphanedActions: string[];  // Actions with no artifacts
    };
    
    deterministicIds: {
      passed: boolean;
      duplicateIds: string[];  // Non-unique IDs detected
      invalidFormats: string[];  // IDs not matching format spec
    };
    
    compensationTraceability: {
      passed: boolean;
      incompleteCompensations: string[]; // Saga runs with missing compensation logs
    };
    
    approvalTraceability: {
      passed: boolean;
      missingApprovals: string[]; // Work items requiring approval but lacking evidence
    };
  };
  
  overallStatus: "complete" | "incomplete" | "broken";
  recommendedActions: string[];
}
```

**Validation Schedule**:
- **Real-time**: On work item creation, transition, and completion
- **Daily**: Nightly batch validation of all active traceability chains
- **Monthly**: Comprehensive audit of archived work items and completed sagas

**Broken Chain Resolution**:
1. **Detection**: Validation check identifies missing link
2. **Notification**: Alert work item owner and agency lead
3. **Remediation**: Manual review to restore link or mark as invalid
4. **Prevention**: Enforce foreign key constraints and required fields in schema

### 13.2 Deterministic ID Generation

All work items, actions, and artifacts must have deterministic, globally unique IDs:

**Work Item ID Format**: `WI-{agencyCode}-{type}-{timestamp}-{hash}`
- Example: `WI-FINSERV-CHANGE-20251030-A3F9B2`

**Agent Action ID Format**: `ACT-{workItemId}-{agentId}-{sequence}`
- Example: `ACT-WI-FINSERV-CHANGE-20251030-A3F9B2-agent-01-001`

**Artifact ID Format**: `ART-{actionId}-{artifactType}-{hash}`
- Example: `ART-ACT-WI-FINSERV-CHANGE-20251030-A3F9B2-agent-01-001-log-5D8E3A`

Benefits:
- **Uniqueness**: Collision-resistant due to timestamp + hash
- **Determinism**: Same inputs always produce same ID
- **Readability**: Human-readable components for debugging
- **Traceability**: IDs encode parent relationships

## 14. Implementation Tasks (Next Steps)

### Phase 1: Core Schema & Registry (MVP-030)
- [ ] Add JSON Schemas for each `work_item_type` to `work_item_types` collection
- [ ] Implement role taxonomy fields in `agent_types` registry
- [ ] Add autonomy level, budget, and safety constraint fields to agent schema
- [ ] Create default roles with example configurations

### Phase 2: Lifecycle & SLA Enforcement (MVP-031)
- [ ] Implement server-side transition validator and `POST /work-items/{id}/transition` endpoint
- [ ] Add SLA timer monitoring service with breach detection
- [ ] Implement breach action handlers (escalation, remediation creation)
- [ ] Add lifecycle audit trail to work item documents

### Phase 3: Assignment & Routing (MVP-032)
- [ ] Build routing engine to evaluate declarative rules
- [ ] Implement skill-based assignment with capacity awareness
- [ ] Add cost budget enforcement to assignment algorithm
- [ ] Create escalation path execution service

### Phase 4: Concurrency & Idempotence (MVP-033)
- [ ] Implement idempotence key deduplication layer
- [ ] Add distributed mutex/lock service (Redis or ArangoDB-based)
- [ ] Enforce reentrancy contracts in task execution
- [ ] Add mutex scope validation to work item creation

### Phase 5: Compensation & Sagas (MVP-034)
- [ ] Implement saga orchestration runner (orchestrator pattern)
- [ ] Add compensation step execution with retry logic
- [ ] Create saga run audit trail and visualization
- [ ] Build rollback testing framework

### Phase 6: Policy Gates & Evidence (MVP-035)
- [ ] Implement policy registry with versioned policies
- [ ] Add policy gate evaluation engine for transitions
- [ ] Build evidence capture UI and storage
- [ ] Integrate with external compliance scanners (optional)

### Phase 7: Templates & Catalog (MVP-036)
- [ ] Create template registry with versioning
- [ ] Implement template parameterization and instantiation
- [ ] Add industry templates (PCI change, HIPAA export, SOC2 review)
- [ ] Build template inheritance and composition system

### Phase 8: Traceability & Validation (MVP-037)
- [ ] Implement deterministic ID generation for all entities
- [ ] Add traceability validation service with scheduled checks
- [ ] Build broken chain detection and notification
- [ ] Create traceability dashboard and reports

---

**Document Version**: 0.2.0 — Draft (Enhanced with Agent Taxonomy & Traceability)  
**Last Updated**: 2025-10-30  
**Owner**: CodeValdCortex Architecture Team  
**Status**: Draft - Ready for Review
