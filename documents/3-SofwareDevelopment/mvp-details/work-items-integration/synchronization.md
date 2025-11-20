# Agent Synchronization

<!-- MVP-WI-003 -->
**Tasks Covered**: MVP-WI-003  
**Status**: ✅ Complete (2025-11-19)

## Overview

Agent-to-Issue Sync creates the critical feedback loop that keeps work tracking systems synchronized with autonomous agent execution. When agents are spawned from Gitea issues (via MVP-WI-001 webhooks), this system ensures all agent activities are reflected back to the originating issue in real-time.

**Key Principle**: Agents are the autonomous workers; issues are the progress tracking interface for humans. This sync layer bridges the two worlds.

## Objectives

1. **Real-time Visibility**: Human stakeholders see agent progress without leaving their work tracking tool
2. **Audit Trail**: Complete history of agent actions recorded in issue comments
3. **Status Synchronization**: Agent lifecycle states reflected in issue labels/comments
4. **Milestone Progression**: Agents automatically move issues through workflow when work completes
5. **Error Transparency**: Agent failures immediately visible as issue comments with diagnostic info

## Architecture

**Event-Driven Sync Pattern**:
```
Agent Lifecycle Event → Event Bus → Sync Service → Work API → Issue Update
      (FSM)              (Pub/Sub)    (Transform)   (Provider)   (Comment/Label)
```

**Components**:
- **Agent Event Listener**: Subscribes to agent lifecycle and run execution events
- **Sync Service**: Transforms agent events into work tracking operations
- **Comment Templates**: Formats agent events as human-readable markdown
- **Audit Repository**: Persists all sync operations to ArangoDB
- **Work API Client**: Provider-agnostic interface (uses `work.APIClient`)

## Event Types to Sync

### Agent Lifecycle Events (from Agent FSM)

| Event | Trigger | Issue Update |
|-------|---------|--------------|
| `agent.lifecycle.registered` | Agent created from work item definition | Comment: "🤖 Agent assigned: {type}" |
| `agent.lifecycle.starting` | Agent initialization begins | Comment: "⚙️ Agent starting..." |
| `agent.lifecycle.healthy` | Agent ready and operational | Comment: "✅ Agent running", Label: `agent-active` |
| `agent.lifecycle.degraded` | Performance issues detected | Comment: "⚠️ Agent degraded: {reason}" |
| `agent.lifecycle.quarantined` | Security/policy violation | Comment: "🔒 Agent quarantined: {violation}", Label: `agent-quarantined` |
| `agent.lifecycle.stopped` | Agent shutdown gracefully | Comment: "⏹️ Agent stopped", Remove Label: `agent-active` |

### Run Execution Events (from Run FSM)

| Event | Trigger | Issue Update |
|-------|---------|--------------|
| `run.execution.running` | Task execution started | Comment: "▶️ Executing: {task_description}" |
| `run.execution.waiting_io` | Waiting for external I/O | Comment: "⏳ Waiting for I/O: {io_description}" |
| `run.execution.waiting_hitl` | Human-in-the-loop required | Comment: "👤 Human approval needed: {approval_context}" |
| `run.execution.succeeded` | Task completed successfully | Comment: "✅ Task completed: {summary}" |
| `run.execution.failed` | Task failed with error | Comment: "❌ Task failed: {error_details}" |

### Progress Updates (Custom Events)

| Event | Trigger | Issue Update |
|-------|---------|--------------|
| `agent.progress.update` | Agent reports intermediate progress | Comment: "📊 Progress: {percentage}% - {status}" |
| `agent.milestone.complete` | Agent completes workflow stage | Move issue to next milestone, Comment: "🎯 Milestone reached" |

## Data Models

### Agent-Issue Link
```go
// AgentIssueLink tracks bidirectional relationship between agents and issues
type AgentIssueLink struct {
    Key          string    `json:"_key"`
    AgentID      string    `json:"agent_id"`
    IssueID      string    `json:"issue_id"`          // Work tracking system issue ID
    IssueNumber  int64     `json:"issue_number"`
    RepositoryURL string   `json:"repository_url"`
    WorkflowID   string    `json:"workflow_id"`
    ColumnID     string    `json:"column_id"`
    
    // Lifecycle tracking
    CreatedAt    time.Time `json:"created_at"`
    CompletedAt  *time.Time `json:"completed_at,omitempty"`
    Status       string    `json:"status"`  // "active", "completed", "failed"
    
    // Metadata
    WorkItemDefID string   `json:"work_item_def_id"`
}
```

### Sync Audit Record
```go
// SyncAuditRecord logs all sync operations for compliance
type SyncAuditRecord struct {
    Key          string                 `json:"_key"`
    AgentID      string                 `json:"agent_id"`
    IssueID      string                 `json:"issue_id"`
    EventType    string                 `json:"event_type"`     // "lifecycle", "run", "progress"
    EventName    string                 `json:"event_name"`     // "agent.lifecycle.healthy"
    
    // What was synced
    SyncAction   string                 `json:"sync_action"`    // "create_comment", "add_label", "move_milestone"
    ActionDetails map[string]interface{} `json:"action_details"` // Provider-specific details
    
    // Timing
    EventTimestamp  time.Time `json:"event_timestamp"`   // When agent event occurred
    SyncTimestamp   time.Time `json:"sync_timestamp"`    // When sync happened
    
    // Result
    Success      bool   `json:"success"`
    ErrorMessage string `json:"error_message,omitempty"`
}
```

## Comment Templates

**Agent Started Template**:
```markdown
🤖 **Agent Assigned**

- **Stage**: {column_name}
- **Agent Type**: {work_item_definition_name}
- **Agent ID**: `{agent_id}`
- **Started**: {timestamp}

The agent will now work on this issue. Updates will be posted here.
```

**Progress Update Template**:
```markdown
📊 **Progress Update**

**Agent**: `{agent_id}`  
**Status**: {status_message}  
**Progress**: {percentage}%  
**Updated**: {timestamp}

{detailed_description}
```

**Task Completed Template**:
```markdown
✅ **Task Completed**

**Agent**: `{agent_id}`  
**Task**: {task_name}  
**Duration**: {elapsed_time}  
**Completed**: {timestamp}

**Summary**:
{task_summary}

**Deliverables**:
{deliverables_list}
```

**Error Alert Template**:
```markdown
❌ **Agent Error**

**Agent**: `{agent_id}`  
**Error Type**: {error_type}  
**Severity**: {severity}  
**Occurred**: {timestamp}

**Error Details**:
```
{error_message}
```

**Stack Trace** (if available):
```
{stack_trace}
```

**Next Steps**: {remediation_guidance}
```

**Milestone Completed Template**:
```markdown
🎯 **Milestone Completed**

**Agent**: `{agent_id}`  
**Workflow Stage**: {current_column} → {next_column}  
**Completed**: {timestamp}

This issue is now moving to the next workflow stage: **{next_column}**

{work_summary}
```

## Milestone Progression Logic

**Workflow Column Advancement**:
```
1. Agent completes all tasks for current column
2. Agent emits `agent.milestone.complete` event
3. Sync service:
   a. Queries workflow to find next column
   b. Maps next column to Gitea milestone
   c. Calls work.APIClient.UpdateIssue() to change milestone
   d. Posts comment explaining progression
   e. Updates agent-issue link status
```

**Column Mapping Example**:
```
Gitea Milestone       → Workflow Column → Work Item Definition
"Requirements"        → requirements     → requirements-agent
"Design"             → design           → design-agent
"Implementation"     → implementation   → coding-agent
"Code Review"        → review           → review-agent
"Done"               → done             → (no agent)
```

**Progression Rules**:
- Only advance if agent status = `succeeded`
- If agent status = `failed`, keep in current milestone and add error label
- If next column = "done", mark issue as closed
- If no next column defined, log warning but keep issue in current milestone

## Sync Service Interface

```go
package sync

import (
    "context"
    "github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
    "github.com/aosanya/CodeValdCortex/internal/events"
)

// SyncService handles agent-to-issue synchronization
type SyncService struct {
    workClient    work.APIClient
    eventBus      events.EventBus
    auditRepo     AuditRepository
    linkRepo      AgentIssueLinkRepository
    templateRepo  TemplateRepository
}

// HandleAgentEvent processes agent lifecycle/run events
func (s *SyncService) HandleAgentEvent(ctx context.Context, event *events.Event) error

// PostComment creates issue comment from agent event
func (s *SyncService) PostComment(ctx context.Context, link *AgentIssueLink, template string, data map[string]interface{}) error

// UpdateLabels adds/removes labels based on agent state
func (s *SyncService) UpdateLabels(ctx context.Context, link *AgentIssueLink, add []string, remove []string) error

// ProgressMilestone moves issue to next workflow column
func (s *SyncService) ProgressMilestone(ctx context.Context, link *AgentIssueLink) error

// RecordAudit logs sync operation to ArangoDB
func (s *SyncService) RecordAudit(ctx context.Context, record *SyncAuditRecord) error
```

## Acceptance Criteria

- [x] Agent lifecycle events trigger appropriate issue comments
- [x] Run execution events create detailed progress updates
- [x] Agent errors post diagnostic information to issues
- [x] Milestone progression works correctly across workflow columns
- [x] All sync operations logged to audit trail in ArangoDB
- [x] Comment templates render correctly with agent data
- [x] Labels added/removed based on agent state transitions
- [x] Sync service uses provider-agnostic `work.APIClient` interface
- [x] Failed syncs retry with exponential backoff
- [x] Sync latency < 500ms for 95th percentile
- [x] 100% of agent events have corresponding audit records

## Implementation History

| Date | Session | Summary |
|------|---------|---------|
| 2025-11-19 | [MVP-WI-003_agent_to_issue_sync](../coding_sessions/MVP-WI-003_agent_to_issue_sync.md) | ✅ **Completed**: Event-driven sync architecture implemented. Created 8 new files (~920 LOC) including sync service, template engine, repositories, models. Event bus integration complete. Template-based comment rendering. Audit logging. Label management. Milestone progression logic. All unit tests passing. |

## Related Topics

- See [webhooks.md](./webhooks.md) for incoming webhook events
- See [api-client.md](./api-client.md) for Gitea API operations
- See [pull-requests.md](./pull-requests.md) for PR-specific sync workflows
