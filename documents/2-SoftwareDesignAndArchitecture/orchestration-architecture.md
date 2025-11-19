# Orchestration Architecture

**Last Updated**: November 19, 2025  
**Status**: Core Architecture  
**Scope**: Orchestrator-Agent interaction model for Kanban workflow automation

---

## Overview

CodeValdCortex implements an **orchestrator-driven agent system** where a central orchestrator manages agent lifecycles based on workflow state, and agents autonomously execute work by picking up and processing Gitea issues through Kanban workflow stages.

### Core Principles

1. **ArangoDB as Central State Store**: All Gitea artifacts (issues, PRs, milestones) persisted in ArangoDB
2. **Orchestrator Monitors Change Streams**: Orchestrator watches ArangoDB for new/updated issues
3. **Webhook Handlers as Persistence Layer**: MVP-WI-001 validates and saves Gitea webhooks to ArangoDB
4. **Agents Query ArangoDB**: Agents read issue details from ArangoDB, not directly from Gitea
5. **Workflow-Driven Progression**: Agents move issues forward (or backward) through workflow stages

---

## System Architecture

```
                    ┌─────────────────────────────┐
                    │      Gitea Instance         │
                    │  Issues, PRs, Milestones    │
                    └──────────┬──────────────────┘
                               │ Webhooks
                               ▼
                    ┌─────────────────────────────┐
                    │   MVP-WI-001: Webhook       │
                    │   Handler (Persistence)     │
                    │  - Validate signatures      │
                    │  - Parse payloads           │
                    │  - Save to ArangoDB         │
                    └──────────┬──────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                          ArangoDB                                 │
│  ┌──────────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │  work-issues     │  │  workflows   │  │  work_item_defs    │ │
│  │  - issue_number  │  │  - columns   │  │  - agent blueprints│ │
│  │  - milestone     │  │  - repo map  │  │  - tools           │ │
│  │  - state         │  │  - WIP limits│  │  - instructions    │ │
│  └──────────────────┘  └──────────────┘  └────────────────────┘ │
└──────────────────┬───────────────────────────────────────────────┘
                   │ Change Streams
                   ▼
        ┌─────────────────────────────────────────┐
        │         Orchestrator (MVP-032)          │
        │  - Monitors ArangoDB change streams     │
        │  - Detects milestoned issues            │
        │  - Queries workflow/column mappings     │
        │  - Enforces WIP limits                  │
        │  - Spawns/terminates agents             │
        └──────────────┬──────────────────────────┘
                       │ spawns/terminates
                       │
          ┌────────────┼────────────┬────────────────┐
          ▼            ▼            ▼                ▼
    ┌─────────┐  ┌─────────┐  ┌─────────┐     ┌─────────┐
    │ Agent 1 │  │ Agent 2 │  │ Agent 3 │ ... │ Agent N │
    │ (Type A)│  │ (Type B)│  │ (Type A)│     │ (Type C)│
    └────┬────┘  └────┬────┘  └────┬────┘     └────┬────┘
         │            │            │                │
         │ Query      │ Query      │ Query          │ Query
         │ ArangoDB   │ ArangoDB   │ ArangoDB       │ ArangoDB
         │ for issue  │ for issue  │ for issue      │ for issue
         │            │            │                │
         └────────────┴────────────┴────────────────┘
                       │
                       ▼
         Agents execute work, update issue state in ArangoDB,
         move issues to next milestone via Gitea API
```

---

## Orchestrator Responsibilities

### 1. ArangoDB Change Stream Monitoring

The orchestrator monitors ArangoDB change streams for the `gitea_issues` collection to detect when Gitea webhooks have persisted new or updated issues.

```go
type Orchestrator struct {
    dbClient         *arangodb.Client
    agentPool        *AgentPool
    wipLimitEnforcer *WIPLimitEnforcer
    changeStreams    *ChangeStreamMonitor
}

func (o *Orchestrator) MonitorGiteaIssues() {
    // Subscribe to ArangoDB change streams for work-issues collection
    stream, err := o.dbClient.Collection("work-issues").WatchChanges(ctx, &WatchOptions{
        FullDocument: "updateLookup",
    })
    
    if err != nil {
        log.Fatal("Failed to watch work-issues collection", "error", err)
    }
    
    for change := range stream {
        switch change.OperationType {
        case "insert", "update":
            // New issue or milestone changed
            o.handleIssueChange(change.FullDocument)
        }
    }
}
```

### 2. Agent Lifecycle Management

**Agent Spawning** (triggered by ArangoDB change stream):
```go
func (o *Orchestrator) handleIssueChange(issueDoc *GiteaIssue) {
    // Skip if no milestone
    if issueDoc.Milestone == "" {
        return
    }
    
    // Query: What workflow does this repo map to?
    workflow, err := o.getWorkflowForRepo(issueDoc.RepoURL)
    if err != nil {
        log.Warn("No workflow found for repo", "repo", issueDoc.RepoURL)
        return
    }
    
    // Query: What column does this milestone map to?
    column, err := o.getColumnForMilestone(workflow.ID, issueDoc.Milestone)
    if err != nil {
        log.Warn("No column mapping for milestone", 
            "milestone", issueDoc.Milestone,
            "workflow", workflow.Name)
        return
    }
    
    // Check WIP limit
    activeAgents := o.agentPool.CountActiveInColumn(column.ID)
    if activeAgents >= column.WIPLimit {
        log.Info("WIP limit reached, queueing issue", 
            "column", column.Name, 
            "limit", column.WIPLimit)
        o.queueIssue(issueDoc, column)
        return
    }
    
    // Get work item definition for this column
    workItemDef, err := o.getWorkItemDefinition(column.WorkItemDefID)
    if err != nil {
        log.Error("Failed to get work item definition", "error", err)
        return
    }
    
    // Spawn agent
    agent := o.agentPool.SpawnAgent(AgentSpawnRequest{
        Type:               workItemDef.Type,
        WorkItemDefinition: workItemDef,
        IssueID:            issueDoc.ID,
        IssueNumber:        issueDoc.IssueNumber,
        Column:             column,
        Workflow:           workflow,
    })
    
    log.Info("Agent spawned from ArangoDB change stream", 
        "agent_id", agent.ID, 
        "type", agent.Type, 
        "issue", issueDoc.IssueNumber,
        "milestone", issueDoc.Milestone)
}
```

**Agent Termination**:
```go
func (o *Orchestrator) handleAgentCompletion(event AgentEvent) {
    agent := event.Agent
    
    // Move issue to next/previous column (based on agent decision)
    if agent.CompletionStatus == "success" {
        o.moveIssueToNextColumn(agent.AssignedIssue, agent.Column)
    } else if agent.CompletionStatus == "needs_rework" {
        o.moveIssueToPreviousColumn(agent.AssignedIssue, agent.Column)
    }
    
    // Terminate agent
    o.agentPool.TerminateAgent(agent.ID)
    
    // Process queued issues if any
    o.processQueuedIssues(agent.Column)
}
```

### 3. WIP Limit Enforcement

```go
type WIPLimitEnforcer struct {
    limits map[string]int // column_id -> max_concurrent
    queues map[string][]Issue // column_id -> queued issues
}

func (w *WIPLimitEnforcer) CanSpawnAgent(columnID string) bool {
    activeCount := agentPool.CountActiveInColumn(columnID)
    limit := w.limits[columnID]
    return activeCount < limit
}

func (w *WIPLimitEnforcer) QueueIssue(issue Issue, columnID string) {
    w.queues[columnID] = append(w.queues[columnID], issue)
    log.Info("Issue queued", "column", columnID, "queue_depth", len(w.queues[columnID]))
}

func (w *WIPLimitEnforcer) ProcessQueue(columnID string) {
    queue := w.queues[columnID]
    if len(queue) == 0 {
        return
    }
    
    for w.CanSpawnAgent(columnID) && len(queue) > 0 {
        issue := queue[0]
        queue = queue[1:]
        orchestrator.SpawnAgentForIssue(issue, columnID)
    }
    
    w.queues[columnID] = queue
}
```

---

## Agent Behavior

### 1. Autonomous Issue Pickup

Agents don't wait for orchestrator assignment - they **actively pick up** issues:

```go
type Agent struct {
    ID               string
    Type             string
    AssignedColumn   WorkflowColumn
    WorkItemDef      WorkItemDefinition
    CurrentIssue     *Issue
    Status           AgentStatus
}

func (a *Agent) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // Pick up next issue from assigned column
            issue := a.pickUpIssue()
            if issue == nil {
                time.Sleep(10 * time.Second)
                continue
            }
            
            // Execute work on the issue
            result := a.executeWork(issue)
            
            // Decide next step based on result
            a.progressIssue(issue, result)
            
            // Mark work complete
            a.CurrentIssue = nil
            a.notifyOrchestrator(AgentCompletionEvent{
                AgentID: a.ID,
                Issue:   issue,
                Result:  result,
            })
        }
    }
}
```

### 2. Issue Selection Strategy

```go
func (a *Agent) pickUpIssue() *Issue {
    // Query Gitea for issues in agent's assigned column
    issues, err := a.giteaClient.ListIssuesInMilestone(
        a.AssignedColumn.Repository,
        a.AssignedColumn.MilestoneName,
        &IssueFilter{
            State:        "open",
            NotAssigned:  true, // Only unassigned issues
            Labels:       a.WorkItemDef.RequiredLabels,
            SortBy:       "priority,created_at",
        },
    )
    
    if err != nil || len(issues) == 0 {
        return nil
    }
    
    // Pick highest priority issue
    issue := issues[0]
    
    // Claim the issue by assigning to self
    a.giteaClient.AssignIssue(issue.Number, a.GitHubUsername())
    a.giteaClient.AddComment(issue.Number, 
        fmt.Sprintf("🤖 Agent %s (%s) started work", a.ID, a.Type))
    
    return issue
}
```

### 3. Work Execution

```go
func (a *Agent) executeWork(issue *Issue) WorkResult {
    // Execute work based on work item definition instructions
    instructions := a.WorkItemDef.Instructions
    
    // Use LLM/tools to perform the work
    result := a.llmExecutor.Execute(LLMTask{
        Prompt:       a.buildPrompt(issue, instructions),
        Tools:        a.WorkItemDef.AllowedTools,
        MaxTokens:    a.WorkItemDef.TokenBudget,
        Temperature:  0.7,
    })
    
    // Post progress updates to issue
    a.giteaClient.AddComment(issue.Number, result.Summary)
    
    // If work produces artifacts (code, documents), create PR
    if result.HasArtifacts {
        pr := a.createPullRequest(issue, result.Artifacts)
        a.giteaClient.AddComment(issue.Number, 
            fmt.Sprintf("📝 Created PR: %s", pr.URL))
    }
    
    return result
}
```

### 4. Issue Progression

Agents decide where issues should go next:

```go
func (a *Agent) progressIssue(issue *Issue, result WorkResult) {
    workflow := a.AssignedColumn.Workflow
    currentColumn := a.AssignedColumn
    
    switch result.Status {
    case "success":
        // Move to next column
        nextColumn := workflow.GetNextColumn(currentColumn.ID)
        if nextColumn != nil {
            a.giteaClient.SetMilestone(issue.Number, nextColumn.MilestoneName)
            a.giteaClient.AddComment(issue.Number, 
                fmt.Sprintf("✅ Work complete. Moving to: %s", nextColumn.Name))
        } else {
            // Last column - close issue
            a.giteaClient.CloseIssue(issue.Number)
            a.giteaClient.AddComment(issue.Number, "✅ All work complete!")
        }
        
    case "needs_rework":
        // Move to previous column
        prevColumn := workflow.GetPreviousColumn(currentColumn.ID)
        if prevColumn != nil {
            a.giteaClient.SetMilestone(issue.Number, prevColumn.MilestoneName)
            a.giteaClient.AddComment(issue.Number, 
                fmt.Sprintf("⚠️ Rework needed. Moving back to: %s\n\nReason: %s", 
                    prevColumn.Name, result.ReworkReason))
        }
        
    case "blocked":
        // Keep in current column but add blocked label
        a.giteaClient.AddLabel(issue.Number, "blocked")
        a.giteaClient.AddComment(issue.Number, 
            fmt.Sprintf("🚫 Blocked: %s", result.BlockReason))
    }
}
```

---

## Workflow Configuration

### Workflow Definition

```yaml
workflow:
  id: "software-dev-workflow"
  name: "Software Development"
  repository: "https://gitea.example.com/myorg/myproject"
  
  columns:
    - id: "requirements"
      name: "Requirements Gathering"
      milestone: "Requirements"
      position: 1
      wip_limit: 3
      work_item_definition: "requirements-gathering-v1"
      
    - id: "design"
      name: "System Design"
      milestone: "Design"
      position: 2
      wip_limit: 2
      work_item_definition: "system-design-v1"
      
    - id: "implementation"
      name: "Implementation"
      milestone: "Development"
      position: 3
      wip_limit: 5
      work_item_definition: "coding-v1"
      
    - id: "review"
      name: "Code Review"
      milestone: "Review"
      position: 4
      wip_limit: 3
      work_item_definition: "code-review-v1"
      
    - id: "done"
      name: "Done"
      milestone: "Completed"
      position: 5
      wip_limit: null  # No limit on done column
      work_item_definition: null  # No agent needed
```

### Work Item Definition

```yaml
work_item_definition:
  id: "requirements-gathering-v1"
  type: "requirements-gathering"
  name: "Requirements Gathering Agent"
  
  instructions: |
    You are a requirements gathering agent. Your job is to:
    1. Read the issue description carefully
    2. Identify stakeholders from the issue or ask for them
    3. Conduct structured interviews (via issue comments)
    4. Document functional and non-functional requirements
    5. Create a requirements document and attach to issue
    6. Get stakeholder approval
    7. Move issue to Design milestone when complete
  
  allowed_tools:
    - "gitea_api"
    - "document_generator"
    - "stakeholder_notifier"
  
  token_budget: 10000
  autonomy_level: "supervised"  # Requires approval for milestone changes
  
  success_criteria:
    - "Requirements document created"
    - "At least 2 stakeholders consulted"
    - "Functional requirements >= 3"
    - "Non-functional requirements >= 2"
```

---

## Event Flow Example

### Complete Flow: Issue Through Workflow

```
1. Issue Created: "Build authentication system"
   └─> Assigned to Milestone: "Requirements"
   └─> Webhook → Orchestrator

2. Orchestrator:
   └─> Check WIP limit for "Requirements" column (3)
   └─> Current agents: 1 (< 3) ✓
   └─> Get work_item_definition: "requirements-gathering-v1"
   └─> Spawn RequirementsGatheringAgent

3. RequirementsGatheringAgent (Agent-001):
   └─> Pick up Issue #42
   └─> Assign to self in Gitea
   └─> Post comment: "🤖 Agent-001 started requirements gathering"
   └─> Execute work:
       ├─> Ask stakeholders questions (via comments)
       ├─> Wait for responses
       ├─> Generate requirements document
       ├─> Create PR with document
       └─> Post: "✅ Requirements complete. Review PR #15"
   └─> Move issue to "Design" milestone
   └─> Notify orchestrator: work complete

4. Orchestrator:
   └─> Terminate Agent-001
   └─> Detect issue now in "Design" milestone
   └─> Check WIP limit for "Design" column (2)
   └─> Current agents: 0 (< 2) ✓
   └─> Get work_item_definition: "system-design-v1"
   └─> Spawn SystemDesignAgent

5. SystemDesignAgent (Agent-002):
   └─> Pick up Issue #42
   └─> Read requirements from PR #15
   └─> Generate architecture diagrams
   └─> Create design document
   └─> Post for review
   └─> Move issue to "Development" milestone
   └─> Notify orchestrator: work complete

6. [Cycle continues through Implementation → Review → Done]
```

---

## Key Benefits

### ✅ **Separation of Concerns**
- **Orchestrator**: Resource management, WIP limits, agent lifecycle
- **Agents**: Autonomous work execution, issue progression

### ✅ **Scalability**
- Agents scale horizontally (multiple agents per column type)
- WIP limits prevent overload
- Queue management handles bursts

### ✅ **Flexibility**
- Agents can move issues forward OR backward
- Dynamic workflow reconfiguration
- Easy to add new agent types

### ✅ **Observability**
- Clear agent → issue → column mapping
- Audit trail in Gitea comments
- Orchestrator provides system-wide view

### ✅ **Resilience**
- Agent failures don't block workflow
- Orchestrator can restart failed agents
- Issues remain in Gitea (source of truth)

---

## Implementation Roadmap

| Phase | Component | MVP Task |
|-------|-----------|----------|
| **Phase 1** | Workflow & Work Item Definitions | MVP-030 |
| **Phase 2** | Repository-Workflow Mapping | MVP-031 |
| **Phase 3** | Agent Factory & Lifecycle | MVP-032 |
| **Phase 4** | Gitea Webhook Integration | MVP-WI-001 |
| **Phase 5** | Orchestrator Core | MVP-WI-002 |
| **Phase 6** | Agent-Issue Sync | MVP-WI-003 |
| **Phase 7** | PR Automation | MVP-WI-004 |

---

## Architecture References

- **Work Items System**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/`
- **Gitea Integration**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items/gitea-integration.md`
- **Agent Lifecycle**: MVP-033 through MVP-036
- **Workflow Designer**: MVP-052 (completed)
