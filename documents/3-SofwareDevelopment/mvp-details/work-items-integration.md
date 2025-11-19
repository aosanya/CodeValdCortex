# Work Items Integration

## Overview

The Work Items Integration system provides comprehensive infrastructure for managing work tracking across external systems (Gitea, GitHub, GitLab, Jira) and internal agent workflows. This domain encompasses webhook ingestion, work item definitions, lifecycle management, routing, and bidirectional synchronization with external platforms.

### Vision

Enable CodeValdCortex agents to be automatically spawned and managed based on work items from external tracking systems, creating a seamless bridge between traditional project management tools and autonomous agent execution. Work items serve as both triggers and blueprints for agent creation, with full bidirectional sync keeping external systems updated on progress.

### Architecture

```
External Systems (Gitea/GitHub/etc.)
    ↓ Webhooks
┌─────────────────────────────────────────┐
│  Work Tracking Abstraction Layer        │
│  (pluggable provider architecture)      │
└──────────────┬──────────────────────────┘
               ↓ Persist
         ┌──────────────┐
         │  ArangoDB    │
         │ work-issues  │
         │ work-prs     │
         └──────┬───────┘
                ↓ Change Streams
         ┌─────────────────┐
         │  Orchestrator   │
         │ (MVP-032)       │
         └──────┬──────────┘
                ↓ Creates
         ┌─────────────────┐
         │    Agents       │
         └──────┬──────────┘
                ↓ Sync Back
    External Systems (Comments, Status)
```

### Key Concepts

- **Work Items**: Abstract representations of work (issues, PRs, tasks, jobs) from any provider
- **Work Item Definitions**: Blueprints for creating agents from work items
- **Providers**: Pluggable integrations (Gitea, GitHub, GitLab, Jira)
- **Lifecycle States**: State machines governing work item progression
- **Routing Rules**: Intelligent assignment of work items to qualified agents
- **Bidirectional Sync**: Keep external systems updated with agent progress

---

## Domain Tasks

This domain covers the following interconnected tasks:

1. **MVP-WI-001**: Gitea Webhook Integration ✅ (webhooks → ArangoDB)
2. **MVP-030**: Work Item Core Schema & Registry (type definitions, templates)
3. **MVP-031**: Work Item Lifecycle & SLA (state machines, timers)
4. **MVP-032**: Work Item Assignment & Routing (agent selection, orchestration)
5. **MVP-WI-002**: Gitea API Client ✅ (bidirectional communication)
6. **MVP-WI-003**: Agent-to-Issue Sync (progress updates, comments)
7. **MVP-WI-004**: Pull Request Automation (PR creation, auto-merge)

---

<!-- MVP-WI-001 -->
## Gitea Webhook Integration (MVP-WI-001)

The webhook integration forms the foundation of our external work tracking integration. It implements a **pluggable architecture** that allows switching between different work tracking providers (Gitea, GitHub, GitLab, Jira, Linear) without changing downstream orchestration logic.

**Priority**: P0  
**Effort**: Medium  
**Skills**: Go, Webhooks, ArangoDB, REST API, Security  
**Dependencies**: None  
**Status**: ✅ Completed 2025-11-19

### Why Pluggable Architecture?

Rather than hard-coding Gitea-specific logic throughout the system, we created an abstraction layer (`work` package) that defines provider-agnostic interfaces and data models. This enables:

- ✅ **Flexibility**: Switch from Gitea to GitHub without changing orchestrator
- ✅ **Multi-Platform**: Support multiple work tracking systems simultaneously
- ✅ **Migration**: Easy migration path from one platform to another
- ✅ **Testability**: Mock providers for testing
- ✅ **Vendor Independence**: Not locked into any single platform

### Architecture Layers

**Interface Layer** (`internal/infrastructure/webhooks/work/`):
- `WorkTrackingProvider` - Main contract for all providers
- `WorkIssue`, `WorkPullRequest`, `WorkMilestone` - Provider-agnostic models
- `Repository` - ArangoDB persistence interface
- `WebhookValidator` - Signature verification

**Implementation Layer** (`internal/infrastructure/webhooks/gitea/`):
- `GiteaHandler` - Implements `WorkTrackingProvider`
- `GiteaIssuePayload` - Gitea-specific webhook types
- `ToWorkIssue()` - Transformation to common models
- HMAC SHA-256 signature validation

**Future Implementations**:
- `github/` - GitHub webhooks (X-Hub-Signature-256)
- `gitlab/` - GitLab webhooks (X-Gitlab-Token)
- `jira/` - Jira webhooks (different event model)
- `linear/` - Linear webhooks (GraphQL-based)

### Data Flow

```
Gitea Instance
  └─ Webhook: issue.milestoned
      ↓
POST /api/v1/work/issues
      ↓
1. Validate X-Gitea-Signature (HMAC SHA-256)
2. Parse JSON payload → GiteaIssuePayload
3. Transform → work.WorkIssue (provider-agnostic)
4. Persist → ArangoDB work-issues collection
5. Return 200 OK
      ↓
ArangoDB Change Streams emit event
      ↓
Orchestrator (MVP-032) detects new milestoned issue
      ↓
Creates agent from work item definition
```

### Why ArangoDB-Centric?

**Decoupling Benefits**:
- Webhook handler doesn't need orchestrator, workflows, or agent factory
- Webhooks persist data even if orchestrator is down (resilience)
- Complete audit history of all external events
- Agents query ArangoDB, not external APIs (performance)
- Change streams provide reactive processing (no polling)

### Provider-Agnostic Data Models

**WorkIssue** (common format across all providers):
```go
type WorkIssue struct {
    Provider       string    // "gitea", "github", "gitlab", "jira"
    IssueID        string    // Unique within provider
    IssueNumber    int64     // Human-readable number
    Title          string
    Body           string
    State          string    // "open", "closed"
    Milestone      string
    MilestoneID    string
    RepoURL        string
    ProjectKey     string    // For Jira
    Labels         []string
    Assignees      []string
    AuthorUsername string
    CreatedAt      time.Time
    UpdatedAt      time.Time
    ProviderMetadata map[string]interface{} // Platform-specific fields
}
```

### ArangoDB Collections

- **work-issues**: All issues across all providers
- **work-prs**: All pull requests/merge requests
- **work-milestones**: All milestones/sprints/releases

Query by provider:
```aql
FOR issue IN work-issues
    FILTER issue.provider == "gitea"
    FILTER issue.milestone == "Requirements"
    RETURN issue
```

### Objectives

- Implement ArangoDB persistence layer for Gitea webhook events
- Validate webhook signatures (X-Gitea-Signature HMAC SHA-256)
- Parse and transform Gitea payloads into provider-agnostic models
- Store issues, PRs, milestones in respective collections
- Provide data foundation for Orchestrator monitoring
- Handle webhook delivery failures with proper HTTP status codes

### Requirements

**Functional**:
1. **Webhook Endpoints**:
   - POST /api/v1/work/issues - Issue events
   - POST /api/v1/work/pull-requests - PR events

2. **Supported Events**:
   - **Issues**: `opened`, `milestoned`, `demilestoned`, `closed`, `edited`, `labeled`, `assigned`
   - **Pull Requests**: `opened`, `synchronized`, `closed`, `merged`, `edited`

3. **ArangoDB Persistence**:
   - Save to `work-issues`, `work-prs`, `work-milestones` collections
   - Upsert logic for updates (idempotent)

**Non-Functional**:
1. **Security**: HMAC SHA-256 signature validation, configurable secret, optional IP allowlist
2. **Performance**: Response time <200ms, asynchronous processing
3. **Reliability**: Idempotent handlers, graceful error recovery
4. **Observability**: Structured logging (MVP-WI-001 prefix), error tracking

### Acceptance Criteria

- [x] POST /api/v1/work/issues endpoint registered
- [x] X-Gitea-Signature HMAC SHA-256 validation working
- [x] GiteaIssuePayload transforms to work.WorkIssue correctly
- [x] work-issues collection persists with all required fields
- [x] Duplicate webhooks handled idempotently
- [x] Invalid signatures return 401 Unauthorized
- [x] Malformed payloads return 400 Bad Request
- [x] Server errors return 500 with error logging
- [x] Unit tests for signature validation, transformation, persistence
- [ ] Integration tests with mock Gitea webhooks (deferred)
- [x] Documentation: provider implementation guide

### Technical Implementation

**Completed Files**:
- `internal/infrastructure/work/interfaces.go` - Provider interfaces (WorkTrackingProvider, Repository)
- `internal/infrastructure/work/models.go` - Provider-agnostic models (WorkIssue, WorkPullRequest, WorkMilestone)
- `internal/infrastructure/work/README.md` - Provider implementation guide
- `internal/infrastructure/gitea/models.go` - Gitea payload types and transformers
- `internal/infrastructure/gitea/handler.go` - HTTP webhook handlers with async processing
- `internal/infrastructure/gitea/validator.go` - HMAC SHA-256 signature validation
- `internal/infrastructure/gitea/repository.go` - ArangoDB persistence with idempotent upsert
- `internal/infrastructure/gitea/validator_test.go` - Unit tests (100% pass)
- `internal/app/app.go` - Handler initialization and route registration
- `internal/config/config.go` - WorkTrackingConfig struct and environment bindings
- `config.yaml` - work_tracking configuration section

**Implementation Highlights**:
- Pluggable provider architecture (abstraction layer + Gitea provider)
- ArangoDB-centric design (webhooks persist → orchestrator monitors change streams)
- Async webhook processing (non-blocking, <200ms response times)
- Security: HMAC SHA-256 with constant-time comparison
- Resource-oriented API design (/api/v1/work/issues not /webhooks/gitea/issues)
- Comprehensive structured logging with MVP-WI-001 prefix

**Validation**:
- ✅ Code compiles successfully
- ✅ All unit tests passing
- ✅ No lint errors
- ✅ Configuration loads without errors
- ✅ Ready for Gitea webhook configuration

### Implementation History

| Date | Session | Summary |
|------|---------|---------|
| 2025-11-19 | [MVP-WI-001_gitea_webhook_integration](../coding_sessions/MVP-WI-001_gitea_webhook_integration.md) | ✅ **Completed**: Pluggable work tracking abstraction layer with full Gitea provider implementation. Created 9 new files (~1,570 LOC) including handler, validator, repository, models, tests. HTTP endpoints registered at /api/v1/work/issues and /api/v1/work/pull-requests. HMAC SHA-256 signature validation, async webhook processing, ArangoDB persistence with idempotent upserts. All unit tests passing. |

<!-- /MVP-WI-001 -->

---

<!-- MVP-030 -->
## Work Item Core Schema & Registry (MVP-030)

While MVP-WI-001 handles external work tracking (Gitea issues), MVP-030 defines the internal work item type system. These are the blueprints that determine how agents should be created and behave when assigned to work.

**Priority**: P0  
**Effort**: Medium  
**Skills**: Go, JSON Schema, ArangoDB  
**Dependencies**: MVP-029 (Goals Module)  
**Status**: 📋 Not Started

### The Work Item Type System

Work items are categorized into 6 fundamental types, each representing a different kind of work that agents perform:

1. **Task**: Discrete, well-defined work with clear completion criteria
2. **Job**: Scheduled or recurring work with execution patterns
3. **Investigation**: Exploratory work requiring research and analysis
4. **Change**: Modifications to existing systems with risk assessment
5. **Remediation**: Corrective actions for incidents or problems
6. **Experiment**: Hypothesis-driven work with success/failure outcomes

### Why Type Schemas Matter

When an external issue (from Gitea) arrives, the orchestrator needs to know:
- What kind of agent to create?
- What tools should it have access to?
- What are the success criteria?
- What are the SLA constraints?

Work item type definitions provide these answers through JSON schemas that define structure, validation rules, and default behaviors.

### Objectives

- Implement work item types registry with CRUD operations
- Define JSON schemas for all 6 work item types
- Extend role/agent type schemas with taxonomy fields
- Create ArangoDB collections: `work_item_types`, `agent_types`, `work_item_templates`
- Build backend services: `WorkItemTypeRegistry`, `AgentTypeRegistry`
- Implement schema validation for work item creation
- Provide template management for common work patterns

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

type WorkItemTaxonomy struct {
    Type        string   `json:"type"`         // Task, Job, Investigation, etc.
    Subtypes    []string `json:"subtypes"`     // e.g., "Bug Fix", "Feature Development"
    Complexity  string   `json:"complexity"`   // Low, Medium, High
    Risk        string   `json:"risk"`         // Low, Medium, High, Critical
}

type WorkItemTemplate struct {
    TemplateID   string                 `json:"template_id"`
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    DefaultValues map[string]interface{} `json:"default_values"`
}
```

### API Endpoints

```
GET    /api/v1/work-item-types              # List all types
GET    /api/v1/work-item-types/{typeId}     # Get specific type
POST   /api/v1/work-item-types              # Create new type
PUT    /api/v1/work-item-types/{typeId}     # Update type
DELETE /api/v1/work-item-types/{typeId}     # Delete type

GET    /api/v1/work-item-templates          # List templates
POST   /api/v1/work-item-templates          # Create template

GET    /api/v1/roles                         # List agent role types
GET    /api/v1/roles/{roleId}                # Get specific role
PUT    /api/v1/roles/{roleId}                # Update role with taxonomy
```

### Example: Task Type Schema

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
      "assignee": { "type": "string" },
      "priority": { "enum": ["low", "medium", "high", "critical"] },
      "due_date": { "type": "string", "format": "date-time" },
      "estimated_effort": { "type": "number", "minimum": 0 }
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

### Integration with MVP-WI-001

When orchestrator (MVP-032) detects a new Gitea issue with milestone "Requirements":

1. Query work_item_types for type matching the milestone/labels
2. Retrieve schema and template for that work item type
3. Validate incoming data against schema
4. Apply defaults from template
5. Create agent with specified tools and constraints

This creates a bridge between external work tracking (Gitea) and internal agent creation (work item types).

**Reference**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md`

<!-- /MVP-030 -->

---

<!-- MVP-031 -->
## Work Item Lifecycle & SLA (MVP-031)

Once work items are defined (MVP-030) and ingested (MVP-WI-001), they need lifecycle management. This task implements state machines, timers, and SLA enforcement to ensure work progresses and deadlines are met.

**Priority**: P0  
**Effort**: Medium  
**Skills**: Go, State Machines, ArangoDB, Timers  
**Dependencies**: MVP-030 (Work Item Core Schema)  
**Status**: 📋 Not Started

### Lifecycle States

Work items progress through these states:

1. **Planned**: Work item created, not yet started
2. **In-Progress**: Agent actively working on it
3. **Waiting**: Blocked on external dependency or HITL approval
4. **Review**: Work complete, awaiting validation
5. **Done**: Successfully completed
6. **Failed**: Could not be completed
7. **Rolled-back**: Completion was undone due to issues

### State Machine Rules

```
Planned → In-Progress (agent assignment)
In-Progress → Waiting (blocked)
In-Progress → Review (work submitted)
In-Progress → Failed (error/timeout)
Waiting → In-Progress (dependency resolved)
Review → Done (approved)
Review → In-Progress (changes requested)
Done → Rolled-back (rollback triggered)
Failed → Planned (retry)
```

### SLA/SLO Enforcement

Each work item has Service Level Agreements that define:
- **Response Time**: How quickly work should start
- **Completion Time**: Maximum duration for completion
- **Escalation Thresholds**: When to escalate

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

```go
type TimerService interface {
    MonitorWorkItem(workItemID string, sla WorkItemSLA)
    CheckBreaches() []SLABreach
    ExecuteBreachAction(breach SLABreach) error
}
```

### Breach Actions

When SLA thresholds are crossed:

1. **Auto-Escalate**: Reassign to senior agent or manager role
2. **Auto-Retry**: Restart failed work with different parameters
3. **Create Remediation**: Spawn new remediation work item
4. **Notify Stakeholders**: Send alerts to configured recipients
5. **Increase Resources**: Allocate more agents/budget

### Objectives

- Implement state machine with transition guards
- Create SLA/SLO models and enforcement
- Build timer service for breach detection
- Implement breach action handlers
- Provide API endpoints for lifecycle management
- Track complete state history for audit

### API Endpoints

```
POST   /api/v1/work-items/{id}/transition   # Change state
GET    /api/v1/work-items/{id}/history      # State history
GET    /api/v1/work-items/{id}/sla          # SLA status
GET    /api/v1/sla-breaches                 # List breaches
POST   /api/v1/sla-breaches/{id}/resolve    # Mark breach resolved
```

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

### Integration with MVP-032

When orchestrator creates an agent from a work item:
1. Work item transitions: Planned → In-Progress
2. SLA timer starts
3. Agent executes work
4. If agent succeeds: In-Progress → Review
5. If agent fails: In-Progress → Failed → (optional auto-retry)
6. If SLA breached: Escalation policy triggers

**Reference**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` Section 3

<!-- /MVP-031 -->

---

<!-- MVP-032 -->
## Work Item Assignment & Routing (MVP-032)

The orchestrator is the brain of the work items system. It monitors ArangoDB for new work items (from MVP-WI-001), matches them to work item definitions (MVP-030), checks lifecycle constraints (MVP-031), and creates agents with the right capabilities.

**Priority**: P0  
**Effort**: High  
**Skills**: Go, ArangoDB Change Streams, Agent Orchestration  
**Dependencies**: MVP-031 (Work Item Lifecycle)  
**Status**: 📋 Not Started

### The Orchestration Challenge

When a Gitea issue arrives:
- Which agent type should handle it?
- Which specific agent instance (if multiple available)?
- Does the agent have the required skills?
- Are there cost/budget constraints?
- Are there data residency requirements?
- Is WIP (Work-In-Progress) limit exceeded?

This task implements the intelligent routing and assignment system that answers these questions.

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

Declarative rules that determine agent selection:

```go
type RoutingRule struct {
    RuleID       string              `json:"rule_id"`
    Name         string              `json:"name"`
    Priority     int                 `json:"priority"`      // Higher = evaluated first
    Conditions   []RuleCondition     `json:"conditions"`
    Selection    SelectionStrategy   `json:"selection"`
    Constraints  []Constraint        `json:"constraints"`
}

type RuleCondition struct {
    Field    string `json:"field"`     // e.g., "labels", "milestone", "priority"
    Operator string `json:"operator"`  // "equals", "contains", "matches"
    Value    string `json:"value"`
}

type SelectionStrategy struct {
    Algorithm  string                 `json:"algorithm"`  // "skills", "cost", "load-balance"
    Parameters map[string]interface{} `json:"parameters"`
}
```

### Agent Selection Algorithms

**1. Skills-Based Matching**:
```go
// Match agent capabilities to work item requirements
func SelectBySkills(workItem WorkItem, availableAgents []Agent) Agent {
    requiredSkills := workItem.RequiredSkills
    scores := make(map[string]float64)
    
    for _, agent := range availableAgents {
        score := CalculateSkillMatch(agent.Skills, requiredSkills)
        scores[agent.ID] = score
    }
    
    return HighestScoreAgent(scores)
}
```

**2. Cost Optimization**:
```go
// Select cheapest agent that meets requirements
func SelectByCost(workItem WorkItem, availableAgents []Agent, budget float64) Agent {
    qualifiedAgents := FilterBySkills(availableAgents, workItem.RequiredSkills)
    
    for _, agent := range qualifiedAgents {
        if agent.Cost <= budget {
            return agent
        }
    }
    
    return nil // No agent within budget
}
```

**3. Load Balancing**:
```go
// Distribute work evenly across agent pool
func SelectByLoad(workItem WorkItem, availableAgents []Agent) Agent {
    loads := QueryCurrentLoads(availableAgents)
    return AgentWithLowestLoad(loads)
}
```

### WIP (Work-In-Progress) Limits

Prevent overwhelming agents or violating Kanban constraints:

```go
type WIPLimit struct {
    Scope      string `json:"scope"`       // "agent", "agent-type", "column"
    EntityID   string `json:"entity_id"`
    MaxItems   int    `json:"max_items"`
    CurrentItems int  `json:"current_items"`
}

func CheckWIPLimit(agentID string, workItemType string) bool {
    limit := GetWIPLimit(agentID, workItemType)
    return limit.CurrentItems < limit.MaxItems
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
2. Orchestrator extracts work item metadata:
   - Milestone → maps to workflow column
   - Labels → maps to work item type
   - Priority → affects agent selection
3. Query work_item_types for matching definition
4. Evaluate routing rules in priority order
5. Find qualified agents (skill match)
6. Apply constraints:
   - Budget limits
   - Data residency
   - WIP limits
7. Select best agent using configured algorithm
8. If WIP limit exceeded → queue work item
9. If no agent available → escalate or retry
10. Create agent instance from work item definition
11. Inject work context (issue details)
12. Start agent execution
13. Link agent ↔ work item in tracking table
14. Update work item state: Planned → In-Progress
```

### Objectives

- Implement change stream monitoring for work-issues collection
- Build routing rules engine with rule evaluation
- Implement agent selection algorithms (skills, cost, load)
- Create skill matching with weighted scoring
- Enforce WIP limits per agent/column/type
- Build queueing system for work items when limits exceeded
- Implement AgentFactory.CreateFromWorkItemDefinition()
- Link agents to work items bidirectionally
- Handle edge cases (no match, no budget, no capacity)

### API Endpoints

```
GET    /api/v1/routing-rules                # List rules
POST   /api/v1/routing-rules                # Create rule
PUT    /api/v1/routing-rules/{id}           # Update rule
DELETE /api/v1/routing-rules/{id}           # Delete rule

POST   /api/v1/work-items/{id}/assign       # Manual assignment
GET    /api/v1/work-items/queue             # Queued items
GET    /api/v1/agents/{id}/workload         # Agent current work
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

### Integration: Complete Flow

Putting MVP-WI-001, MVP-030, MVP-031, MVP-032 together:

```
Developer creates Gitea issue "Implement OAuth2"
    ↓
Developer assigns milestone "Authentication" to issue
    ↓
Gitea webhook fires: issue.milestoned
    ↓
MVP-WI-001: Webhook handler validates, transforms, persists to work-issues
    ↓
ArangoDB change stream emits event
    ↓
MVP-032: Orchestrator detects new issue
    ↓
MVP-030: Queries work_item_types for "authentication" type
    ↓
MVP-030: Retrieves work item definition and schema
    ↓
MVP-032: Evaluates routing rules
    ↓
MVP-032: Finds agent with OAuth2 skills
    ↓
MVP-031: Checks WIP limits (not exceeded)
    ↓
MVP-032: Creates agent from definition
    ↓
MVP-031: Work item state: Planned → In-Progress
    ↓
MVP-031: SLA timer starts
    ↓
Agent executes: reads requirements, writes code, creates PR
    ↓
MVP-WI-003: Agent posts progress comments to Gitea issue
    ↓
MVP-WI-004: Agent creates PR, links to issue
    ↓
Agent completes successfully
    ↓
MVP-031: Work item state: In-Progress → Review → Done
```

**Reference**: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` Section 4

<!-- /MVP-032 -->

---

## Future Tasks (MVP-WI-002, MVP-WI-003, MVP-WI-004)

The following tasks complete the bidirectional sync between agents and external work tracking systems:

<!-- MVP-WI-002 -->
## Gitea API Client (MVP-WI-002)

The Gitea API Client provides bidirectional communication with Gitea, enabling agents to interact with work items programmatically. While MVP-WI-001 handles incoming webhooks (Gitea → CodeValdCortex), this task implements outgoing API calls (CodeValdCortex → Gitea).

**Priority**: P0  
**Effort**: Medium  
**Skills**: Go, REST API, Rate Limiting, Error Handling  
**Dependencies**: MVP-WI-001 (Gitea Webhook Integration)  
**Status**: ✅ Completed 2025-11-19

### Objectives

- Implement comprehensive Gitea API client with full CRUD operations
- Support issues, pull requests, milestones, comments, labels, repositories
- Add configurable rate limiting to prevent API throttling
- Implement proper error handling with context wrapping
- Use interface-based design for mockability and testing
- Configure via environment variables and config.yaml

### Architecture

**Interface-Based Design**:
```go
type Client interface {
    // Issue operations
    CreateIssue(ctx context.Context, owner, repo string, opts CreateIssueOptions) (*gitea.Issue, error)
    UpdateIssue(ctx context.Context, owner, repo string, index int64, opts UpdateIssueOptions) (*gitea.Issue, error)
    GetIssue(ctx context.Context, owner, repo string, index int64) (*gitea.Issue, error)
    ListIssues(ctx context.Context, owner, repo string, opts ListIssueOptions) ([]*gitea.Issue, error)
    
    // Pull request operations
    CreatePullRequest(ctx context.Context, owner, repo string, opts CreatePullRequestOptions) (*gitea.PullRequest, error)
    MergePullRequest(ctx context.Context, owner, repo string, index int64, opts MergePullRequestOptions) error
    
    // Milestone operations
    CreateMilestone(ctx context.Context, owner, repo string, opts CreateMilestoneOptions) (*gitea.Milestone, error)
    UpdateMilestone(ctx context.Context, owner, repo string, id int64, opts UpdateMilestoneOptions) (*gitea.Milestone, error)
    
    // Comment and label operations
    PostComment(ctx context.Context, owner, repo string, index int64, body string) (*gitea.Comment, error)
    AddLabel(ctx context.Context, owner, repo string, index int64, labels []int64) error
}
```

**Implementation Features**:
- **Authentication**: Token-based authentication via `X-Access-Token` header
- **Rate Limiting**: Configurable token bucket algorithm (default: 10 req/s)
- **Error Handling**: Context-aware errors with wrapped messages
- **Configuration**: Viper-based config with environment variable overrides
- **Testing**: Unit tests for validation logic, option structs (integration tests deferred)

### Configuration

```yaml
work_tracking:
  gitea_base_url: "https://gitea.example.com"
  gitea_api_token: "${CVXC_WORK_TRACKING_GITEA_API_TOKEN}"
  gitea_timeout: 30s
  gitea_rate_limit: 10  # requests per second
```

**Environment Variables**:
- `CVXC_WORK_TRACKING_GITEA_BASE_URL` - Gitea server URL
- `CVXC_WORK_TRACKING_GITEA_API_TOKEN` - API authentication token
- `CVXC_WORK_TRACKING_GITEA_TIMEOUT` - HTTP timeout duration
- `CVXC_WORK_TRACKING_GITEA_RATE_LIMIT` - Max requests per second

### Usage Examples

**Creating an Issue**:
```go
client, err := giteawebhook.NewClient(giteawebhook.ClientConfig{
    BaseURL: config.WorkTracking.GiteaBaseURL,
    Token:   config.WorkTracking.GiteaAPIToken,
    Timeout: config.WorkTracking.GiteaTimeout,
    RateLimit: config.WorkTracking.GiteaRateLimit,
})

issue, err := client.CreateIssue(ctx, "myorg", "myrepo", giteawebhook.CreateIssueOptions{
    Title: "Agent encountered error",
    Body: "Details: ...",
    Assignees: []string{"user1"},
    Labels: []int64{1, 2, 3},
})
```

**Merging a Pull Request**:
```go
err := client.MergePullRequest(ctx, "myorg", "myrepo", 42, giteawebhook.MergePullRequestOptions{
    Style: "squash",
    Message: "Squash and merge",
})
```

**Posting a Progress Comment**:
```go
comment, err := client.PostComment(ctx, "myorg", "myrepo", 10, 
    "Agent progress: 3/5 subtasks complete")
```

### API Operations

**Issues**:
- `CreateIssue` - Create new issue with title, body, assignees, labels, milestone
- `UpdateIssue` - Update existing issue (partial updates supported)
- `GetIssue` - Retrieve issue by index
- `ListIssues` - List issues with filtering (state, labels, milestone, pagination)
- `CloseIssue` - Mark issue as closed
- `ReopenIssue` - Reopen closed issue

**Pull Requests**:
- `CreatePullRequest` - Create PR with base/head branches
- `UpdatePullRequest` - Update PR title, body, assignees
- `GetPullRequest` - Get PR details
- `ListPullRequests` - List PRs with state filtering
- `MergePullRequest` - Merge with style (merge/rebase/squash)

**Milestones**:
- `CreateMilestone` - Create milestone with title, description, due date
- `UpdateMilestone` - Update milestone details
- `GetMilestone` - Get milestone by ID
- `ListMilestones` - List milestones with state filtering

**Comments & Labels**:
- `PostComment` - Add comment to issue/PR
- `AddLabel` - Apply labels to issue/PR
- `RemoveLabel` - Remove labels from issue/PR

**Repositories**:
- `GetRepository` - Get repository details
- `ListRepositories` - List repositories for user/org

### Error Handling

All methods return errors wrapped with context:
```go
if err != nil {
    // Error message includes operation context
    // Example: "failed to create issue: HTTP 404: repository not found"
}
```

**HTTP Status Codes**:
- 401 Unauthorized: Invalid or missing token
- 403 Forbidden: Token lacks required permissions
- 404 Not Found: Repository, issue, or milestone not found
- 422 Unprocessable Entity: Invalid request data
- 500 Internal Server Error: Gitea server error

### Rate Limiting

The client implements token bucket rate limiting:
- Default: 10 requests/second (configurable)
- Blocks requests that exceed the limit
- Respects context cancellation during rate limit waits
- Example: At 10 req/s, 100 requests take ~10 seconds

### Technical Implementation

**Completed Files**:
- `internal/infrastructure/gitea/client.go` - Client interface, option structs (140 LOC)
- `internal/infrastructure/gitea/client_impl.go` - Core API methods (330 LOC)
- `internal/infrastructure/gitea/client_pr_milestone.go` - PR and milestone operations (270 LOC)
- `internal/infrastructure/gitea/client_test.go` - Unit tests (135 LOC, 13 tests passing)
- `internal/config/config.go` - WorkTrackingConfig updates
- `config.yaml` - Gitea API configuration section

**Implementation Highlights**:
- Interface-based design enables mocking for tests
- Rate limiting prevents API throttling
- Context support enables cancellation and timeouts
- Gitea SDK (`code.gitea.io/sdk/gitea`) for type safety
- Comprehensive option structs for all operations
- Unit tests for validation and option construction

**Validation**:
- ✅ Code compiles successfully
- ✅ All 13 unit tests passing
- ✅ No lint errors
- ✅ Configuration loads without errors
- ✅ Ready for integration with agent sync (MVP-WI-003)

### Acceptance Criteria

- [x] Client interface with 30+ methods defined
- [x] Token-based authentication implemented
- [x] Rate limiting with configurable limits
- [x] Context support for cancellation/timeouts
- [x] Error handling with wrapped context
- [x] Configuration via environment variables
- [x] Issue CRUD operations
- [x] Pull request create, update, merge
- [x] Milestone create, update, list
- [x] Comment posting
- [x] Label management
- [x] Repository queries
- [x] Unit tests (validation, options, structs)
- [ ] Integration tests with real Gitea instance (deferred)
- [x] Documentation with usage examples

### Implementation History

| Date | Session | Summary |
|------|---------|---------|
| 2025-11-19 | [MVP-WI-002_gitea_api_client](../coding_sessions/MVP-WI-002_gitea_api_client.md) | ✅ **Completed**: Full Gitea API client implementation with interface-based design. Created 3 new files (~740 LOC implementation + 135 LOC tests). Supports issues, PRs, milestones, comments, labels, repositories. Added rate limiting (10 req/s default), token authentication, context support, comprehensive error handling. Configuration via config.yaml and environment variables. All 13 unit tests passing. |
| 2025-11-19 | Infrastructure enhancement | ✅ **Provider-Agnostic Layer**: Added `work.APIClient` interfaces and `APIClientAdapter` for multi-provider support. Created adapter layer (672 LOC) that wraps Gitea client and implements provider-agnostic interfaces. Enables services to work with any provider (Gitea, GitHub, GitLab) through common interface. All type conversions (Gitea ↔ work models) implemented. Ready for MVP-WI-003. |

### Provider-Agnostic Architecture

To enable multi-provider support, we extended MVP-WI-002 with an adapter pattern:

**Provider-Agnostic Interfaces** (`internal/infrastructure/work/interfaces.go`):
```go
type APIClient interface {
    IssueClient
    PullRequestClient  
    MilestoneClient
    CommentClient
    LabelClient
    RepositoryClient
}
```

**Gitea Adapter** (`internal/infrastructure/gitea/adapter.go`):
- Implements all `work.APIClient` interfaces
- Converts between Gitea SDK types and `work` package types
- Handles ID normalization (int64 → string for multi-provider support)
- 672 LOC with full type conversion logic

**Usage Pattern**:
```go
// Services depend on work.APIClient, not Gitea-specific client
type AgentSyncService struct {
    apiClient work.APIClient  // Provider-agnostic!
}

// Works with ANY provider
comment, err := svc.apiClient.PostComment(ctx, owner, repo, issueID, "Progress: 75%")
```

**Benefits**:
- ✅ Easy to add GitHub, GitLab, Jira providers
- ✅ Services testable with mock implementations
- ✅ Provider changes isolated to adapter layer
- ✅ Type-safe conversions at compile time

See [ADAPTER.md](../../../internal/infrastructure/gitea/ADAPTER.md) for usage guide.

<!-- /MVP-WI-002 -->

### MVP-WI-003: Agent-to-Issue Sync
Implement bidirectional sync:
- Agent state changes → Gitea issue comments/labels
- Agent progress → Issue progress bar
- Agent completion → Milestone progression
- Agent errors → Issue alerts
- Audit trail of all agent actions

### MVP-WI-004: Pull Request Automation
Automate PR workflows:
- Agent creates PR from work
- Link PR to originating issue
- Auto-merge on approval (when tests pass)
- Update issue milestone on merge/close
- Track code changes attributable to agents

These tasks are detailed in separate sections or will be added as this domain evolves.

---

## Summary

The Work Items Integration domain provides end-to-end infrastructure for managing work from external tracking systems through autonomous agent execution and back. The pluggable architecture ensures vendor independence while maintaining deep integration with each platform's unique features.

**Key Achievement**: Bridges traditional project management tools (Gitea, Jira, GitHub) with autonomous AI agent execution, creating a seamless development workflow where work items automatically spawn and manage agents.
