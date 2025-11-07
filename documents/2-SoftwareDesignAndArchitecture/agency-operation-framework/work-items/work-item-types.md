---
title: Work Item Types
path: /documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/work-item-types.md
---

# Work Item Types & Execution Patterns

Work items are triggered by **Gitea/GitLab issues** with specific labels. Each work item type maps to a specific execution pattern that runs in a goroutine.

## Work Item Executor Interface

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

## Core Work Types

### 1. Document Work
**Labels**: `documentation`, `docs`

**Execution Pattern:**
```
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

**Use Cases:**
- Update architecture documentation
- Create new guides
- Fix documentation typos
- Translate documents

**Auto-merge**: Yes (docs typically don't need review)

---

### 2. Software Work
**Labels**: `feature`, `bug`, `enhancement`, `code`

**Execution Pattern:**
```
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

**Use Cases:**
- Implement new features
- Fix bugs
- Add tests
- Refactor code

**Auto-merge**: Conditional (if CI passes and `automated` label)

---

### 3. Proposal Work
**Labels**: `proposal`, `business`, `rfp`

**Execution Pattern:**
```
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

**Use Cases:**
- Update business proposals
- Create RFP responses
- Modify budget sections
- Update timelines

**Auto-merge**: No (requires stakeholder review)

---

### 4. Analysis Work
**Labels**: `investigation`, `research`, `analysis`

**Execution Pattern:**
```
goroutine:
  1. Gather context from ArangoDB (related commits, code, issues)
  2. LLM performs analysis using graph queries
  3. Generate analysis report
  4. Store findings in ArangoDB
  5. Comment analysis on issue
  6. Link related work items via graph edges
```

**Use Cases:**
- Root cause analysis
- Code dependency investigation
- Impact assessment
- Technical feasibility studies

**Auto-merge**: N/A (posts analysis as comment)

---

## Data Models

### Work Item Metadata

```go
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
```

### Work Item → Commit Link (Graph Edge)

```go
type WorkItemCommit struct {
    From      string `json:"_from"` // work_items/wi_123
    To        string `json:"_to"`   // git_commits/abc123
    Type      string `json:"type"`  // implements, fixes, refactors
    CreatedBy string `json:"created_by"` // agent ID
    Automated bool   `json:"automated"`
}
```

## Classification Logic

```go
func (e *WorkItemExecutor) ClassifyWork(issue *gitea.Issue) WorkType {
    labels := issue.Labels
    
    // Priority-based classification
    if hasAnyLabel(labels, "documentation", "docs") {
        return WorkTypeDocument
    }
    if hasAnyLabel(labels, "feature", "bug", "enhancement", "code") {
        return WorkTypeSoftware
    }
    if hasAnyLabel(labels, "proposal", "business", "rfp") {
        return WorkTypeProposal
    }
    if hasAnyLabel(labels, "investigation", "research", "analysis") {
        return WorkTypeAnalysis
    }
    
    // Default: treat as documentation
    return WorkTypeDocument
}
```

## Extending with New Types

To add a new work item type:

1. **Define label**: Add new label in Gitea (e.g., `infrastructure`)
2. **Create executor**: Implement execution pattern
3. **Update classifier**: Add to `ClassifyWork()` function
4. **Define merge policy**: Set auto-merge rules
5. **Create ArangoDB indices**: For new metadata fields

**Example: Infrastructure Work Type**
```go
case "infrastructure":
    return WorkTypeInfrastructure

func (e *WorkItemExecutor) executeInfrastructureWork(ctx context.Context, issue *gitea.Issue) error {
    // 1. Parse terraform/k8s requirements
    // 2. LLM generates infrastructure code
    // 3. Validate with dry-run
    // 4. Commit to branch
    // 5. Create MR with plan output
}
```

---

**See Also:**
- [GitOps Workflow](./gitops-workflow.md) - Detailed execution flow
- [LLM Integration](./llm-integration.md) - Content generation
- [Data Models](./data-models.md) - Complete schemas
