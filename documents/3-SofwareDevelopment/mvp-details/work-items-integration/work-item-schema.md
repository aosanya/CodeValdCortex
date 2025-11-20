# Work Item Schema

<!-- MVP-030, MVP-031, MVP-032 -->
**Tasks Covered**: MVP-030 (Core Schema & Registry), MVP-031 (Lifecycle & SLA), MVP-032 (Assignment & Routing)  
**Status**: 📋 Planned

## Overview

This document covers the internal work item type system that defines how agents are created and behave when assigned to work. While MVP-WI-001 handles external work tracking (Gitea issues), these tasks define the internal schemas, lifecycles, and orchestration logic.

### Covered Tasks

1. **MVP-030**: Work Item Core Schema & Registry - Type definitions and templates
2. **MVP-031**: Work Item Lifecycle & SLA - State machines and timers
3. **MVP-032**: Work Item Assignment & Routing - Agent selection and orchestration

---

## MVP-030: Work Item Core Schema & Registry

### Problem

When an external issue (from Gitea) arrives, the orchestrator needs to know:
- What kind of agent to create?
- What tools should it have access to?
- What are the success criteria?
- What are the SLA constraints?

Work item type definitions provide these answers through JSON schemas.

### Work Item Types

6 fundamental types representing different kinds of work:

1. **Task**: Discrete, well-defined work with clear completion criteria
2. **Job**: Scheduled or recurring work with execution patterns
3. **Investigation**: Exploratory work requiring research and analysis
4. **Change**: Modifications to existing systems with risk assessment
5. **Remediation**: Corrective actions for incidents or problems
6. **Experiment**: Hypothesis-driven work with success/failure outcomes

### Data Model

```go
type WorkItemType struct {
    ID             string                 `json:"id"`
    Name           string                 `json:"name"`           // "Task", "Job", etc.
    Category       string                 `json:"category"`       // Primary classification
    Schema         map[string]interface{} `json:"schema"`         // JSON schema
    Taxonomy       WorkItemTaxonomy       `json:"taxonomy"`
    DefaultSLA     SLAConfig              `json:"default_sla"`
    RequiredTools  []string               `json:"required_tools"`
    SuccessCriteria map[string]interface{} `json:"success_criteria"`
    Templates      []WorkItemTemplate     `json:"templates"`
}
```

### Example Schema

**Task Type**:
```json
{
  "id": "task",
  "name": "Task",
  "category": "operational",
  "schema": {
    "type": "object",
    "properties": {
      "title": { "type": "string", "minLength": 3 },
      "description": { "type": "string" },
      "priority": { "enum": ["low", "medium", "high", "critical"] },
      "due_date": { "type": "string", "format": "date-time" }
    },
    "required": ["title", "description", "priority"]
  },
  "default_sla": {
    "response_time": "1h",
    "completion_time": "24h"
  },
  "required_tools": ["git", "code-editor", "test-runner"]
}
```

### Acceptance Criteria

- [ ] All 6 work item types defined with complete JSON schemas
- [ ] Agent types extended with taxonomy fields
- [ ] WorkItemTypeRegistry service functional with CRUD operations
- [ ] Schema validation prevents invalid work item creation
- [ ] Default values applied from templates
- [ ] ArangoDB collections created with proper indexes
- [ ] Migration scripts for existing data structures
- [ ] API endpoints return proper error codes
- [ ] Documentation for all schemas and taxonomy

---

## MVP-031: Work Item Lifecycle & SLA

### Problem

Work items need lifecycle management to ensure progress and meet deadlines.

### Lifecycle States

```
Planned → In-Progress → Waiting → Review → Done
                ↓
              Failed → Planned (retry)
                       
Done → Rolled-back (if needed)
```

### State Machine Rules

- `Planned → In-Progress` (agent assignment)
- `In-Progress → Waiting` (blocked on dependency)
- `In-Progress → Review` (work submitted)
- `In-Progress → Failed` (error/timeout)
- `Waiting → In-Progress` (dependency resolved)
- `Review → Done` (approved)
- `Review → In-Progress` (changes requested)
- `Done → Rolled-back` (rollback triggered)
- `Failed → Planned` (retry)

### SLA/SLO Enforcement

```go
type WorkItemSLA struct {
    ResponseTime     time.Duration `json:"response_time"`     // e.g., 1h
    CompletionTime   time.Duration `json:"completion_time"`   // e.g., 24h
    EscalationPolicy string        `json:"escalation_policy"` // Policy ID
    BreachActions    []BreachAction `json:"breach_actions"`
}

type BreachAction struct {
    Threshold  float64 `json:"threshold"`   // 0.8 = 80% of SLA
    Action     string  `json:"action"`      // "escalate", "retry", "notify"
    Parameters map[string]interface{} `json:"parameters"`
}
```

### Timer Service

Background service that:
- Monitors all in-progress work items
- Calculates time elapsed vs SLA limits
- Detects breaches at configured thresholds
- Triggers breach actions automatically

### Breach Actions

When SLA thresholds are crossed:

1. **Auto-Escalate**: Reassign to senior agent or manager role
2. **Auto-Retry**: Restart failed work with different parameters
3. **Create Remediation**: Spawn new remediation work item
4. **Notify Stakeholders**: Send alerts to configured recipients
5. **Increase Resources**: Allocate more agents/budget

### Acceptance Criteria

- [ ] State machine enforces valid transitions only
- [ ] Invalid transitions return 400 with error message
- [ ] SLA timers track response and completion times accurately
- [ ] Breach detection triggers at correct thresholds
- [ ] Escalation policies execute correctly
- [ ] State history is complete and queryable
- [ ] Metrics collected for SLA compliance reporting
- [ ] Breach actions are retryable and idempotent
- [ ] Timer service recovers from crashes without losing state

---

## MVP-032: Work Item Assignment & Routing

### Problem

When a Gitea issue arrives:
- Which agent type should handle it?
- Which specific agent instance?
- Does the agent have the required skills?
- Are there cost/budget constraints?
- Is WIP limit exceeded?

### Architecture

```
ArangoDB Change Streams
    ↓ Emits: new work-issues
┌──────────────────────────┐
│  Orchestrator Service    │
│  - Monitor change streams│
│  - Evaluate routing rules│
│  - Match skills          │
│  - Check constraints     │
│  - Select agent          │
└──────────┬───────────────┘
           ↓ Creates
┌──────────────────────────┐
│    Agent Factory         │
│  - Create from blueprint │
│  - Inject work context   │
│  - Start execution       │
└──────────────────────────┘
```

### Routing Rules Engine

```go
type RoutingRule struct {
    RuleID       string              `json:"rule_id"`
    Name         string              `json:"name"`
    Priority     int                 `json:"priority"`      // Higher = evaluated first
    Conditions   []RuleCondition     `json:"conditions"`
    Selection    SelectionStrategy   `json:"selection"`
    Constraints  []Constraint        `json:"constraints"`
}
```

### Agent Selection Algorithms

**1. Skills-Based Matching**: Match agent capabilities to work item requirements

**2. Cost Optimization**: Select cheapest agent that meets requirements

**3. Load Balancing**: Distribute work evenly across agent pool

### WIP (Work-In-Progress) Limits

Prevent overwhelming agents or violating Kanban constraints:

```go
type WIPLimit struct {
    Scope      string `json:"scope"`       // "agent", "agent-type", "column"
    EntityID   string `json:"entity_id"`
    MaxItems   int    `json:"max_items"`
    CurrentItems int  `json:"current_items"`
}
```

### Change Stream Monitoring

```go
func (o *Orchestrator) MonitorWorkIssues(ctx context.Context) {
    stream := o.db.Collection("work-issues").WatchChanges(ctx)
    
    for {
        select {
        case change := <-stream:
            if change.OperationType == "insert" || change.OperationType == "update" {
                issue := change.Document.(*work.WorkIssue)
                
                // Only process milestoned issues
                if issue.Milestone != "" {
                    go o.handleNewWorkItem(ctx, issue)
                }
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### Orchestration Flow

```
1. Change stream detects new work-issues document
2. Orchestrator extracts work item metadata
3. Query work_item_types for matching definition
4. Evaluate routing rules in priority order
5. Find qualified agents (skill match)
6. Apply constraints (budget, data residency, WIP limits)
7. Select best agent using configured algorithm
8. If WIP limit exceeded → queue work item
9. If no agent available → escalate or retry
10. Create agent instance from work item definition
11. Inject work context (issue details)
12. Start agent execution
13. Link agent ↔ work item in tracking table
14. Update work item state: Planned → In-Progress
```

### Acceptance Criteria

- [ ] Change stream monitoring detects new milestoned issues
- [ ] Routing rules evaluate in priority order
- [ ] Skills matching finds qualified agents
- [ ] Cost budgets are respected
- [ ] Data residency rules enforced
- [ ] Load balancing distributes work evenly
- [ ] WIP limits prevent overload
- [ ] Queued work items are processed when capacity available
- [ ] Agent creation includes work context injection
- [ ] Agent-to-work-item tracking is bidirectional
- [ ] Failed assignments trigger fallback mechanisms
- [ ] Metrics track assignment success rate
- [ ] Orchestrator recovers from crashes without losing work

---

## Complete Integration Flow

Putting MVP-WI-001, MVP-030, MVP-031, MVP-032 together:

```
1. Developer creates Gitea issue "Implement OAuth2"
2. Developer assigns milestone "Authentication" to issue
3. Gitea webhook fires: issue.milestoned
4. MVP-WI-001: Webhook handler validates, transforms, persists to work-issues
5. ArangoDB change stream emits event
6. MVP-032: Orchestrator detects new issue
7. MVP-030: Queries work_item_types for "authentication" type
8. MVP-030: Retrieves work item definition and schema
9. MVP-032: Evaluates routing rules
10. MVP-032: Finds agent with OAuth2 skills
11. MVP-031: Checks WIP limits (not exceeded)
12. MVP-032: Creates agent from definition
13. MVP-031: Work item state: Planned → In-Progress
14. MVP-031: SLA timer starts
15. Agent executes: reads requirements, writes code, creates PR
16. MVP-WI-003: Agent posts progress comments to Gitea issue
17. MVP-WI-004: Agent creates PR, links to issue
18. Agent completes successfully
19. MVP-031: Work item state: In-Progress → Review → Done
```

---

## API Endpoints

### Work Item Types (MVP-030)
```
GET    /api/v1/work-item-types              # List all types
GET    /api/v1/work-item-types/{typeId}     # Get specific type
POST   /api/v1/work-item-types              # Create new type
PUT    /api/v1/work-item-types/{typeId}     # Update type
DELETE /api/v1/work-item-types/{typeId}     # Delete type
```

### Lifecycle (MVP-031)
```
POST   /api/v1/work-items/{id}/transition   # Change state
GET    /api/v1/work-items/{id}/history      # State history
GET    /api/v1/work-items/{id}/sla          # SLA status
GET    /api/v1/sla-breaches                 # List breaches
```

### Routing (MVP-032)
```
GET    /api/v1/routing-rules                # List rules
POST   /api/v1/routing-rules                # Create rule
POST   /api/v1/work-items/{id}/assign       # Manual assignment
GET    /api/v1/work-items/queue             # Queued items
```

---

## Related Topics

- See [webhooks.md](./webhooks.md) for external work item ingestion
- See [api-client.md](./api-client.md) for Gitea API integration
- See [synchronization.md](./synchronization.md) for agent-to-issue updates
- See [pull-requests.md](./pull-requests.md) for PR automation workflows

## References

- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md`
