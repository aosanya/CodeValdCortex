# Instance Management: Data Models & Database Schema

**Related Task**: MVP-PUB-007  
**Component**: Data Layer  
**Research Reference**: See [instance-research-session.md](instance-research-session.md) for architectural Q&A

---

## Data Models

### AgencyInstance Model

```go
// internal/agency/models/instance.go
type AgencyInstance struct {
    // ArangoDB fields (stored in agency-specific database)
    Key string `json:"_key"`
    ID  string `json:"_id"`
    Rev string `json:"_rev,omitempty"`
    
    // Instance metadata
    AgencyID    string `json:"agency_id"`
    TagID       string `json:"tag_id"`        // Immutable reference to source tag
    TagName     string `json:"tag_name"`      // Cached for display
    InstanceID  string `json:"instance_id"`   // Unique identifier (UUID)
    Name        string `json:"name"`          // User-friendly name (MUST be unique per agency)
    Description string `json:"description"`
    
    // Runtime state
    State       InstanceState `json:"state"`  // running, stopping, stopped, failed
    HealthStatus string       `json:"health_status"` // healthy, degraded, unhealthy (calculated on-demand)
    
    // Deployment info
    DeployedAt  time.Time  `json:"deployed_at"`
    DeployedBy  string     `json:"deployed_by"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    StoppedAt   *time.Time `json:"stopped_at,omitempty"`
    
    // Runtime tracking
    AgentCount       int       `json:"agent_count"`        // Active agent references
    WorkflowCount    int       `json:"workflow_count"`     // Active workflows
    LastHeartbeat    time.Time `json:"last_heartbeat"`
    UptimeSeconds    int64     `json:"uptime_seconds"`     // Computed: current_time - started_at
    AcceptsNewJobs   bool      `json:"accepts_new_jobs"`   // False when stopping
    
    // Resource allocation (from tag snapshot)
    ResourceLimits ResourceAllocation `json:"resource_limits"`
    
    // Metadata and tagging
    Tags     []string               `json:"tags,omitempty"`     // Uses existing tag system
    Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional tracking data
    
    // Soft delete
    IsDeleted bool       `json:"is_deleted"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    DeletedBy string     `json:"deleted_by,omitempty"`
}
```

### InstanceState Enum

```go
type InstanceState string

const (
    InstanceStateRunning  InstanceState = "running"   // Instance active (set immediately on creation)
    InstanceStateStopping InstanceState = "stopping"  // Graceful shutdown (rejects new jobs, completes current work)
    InstanceStateStopped  InstanceState = "stopped"   // All agents stopped
    InstanceStateFailed   InstanceState = "failed"    // Failed to start/crashed
)
```

### Supporting Types

```go
type ResourceAllocation struct {
    MaxAgents     int    `json:"max_agents"`
    MaxWorkflows  int    `json:"max_workflows"`
    MemoryLimitMB int    `json:"memory_limit_mb"`
    CPULimit      string `json:"cpu_limit"`
}

type StartInstanceRequest struct {
    Name        string   `json:"name"`        // Must be unique per agency
    Description string   `json:"description"`
    Tags        []string `json:"tags,omitempty"` // Uses existing tag system
}

type InstanceFilters struct {
    TagID       string         `json:"tag_id,omitempty"`
    State       InstanceState  `json:"state,omitempty"`
    FromDate    *time.Time     `json:"from_date,omitempty"`
    ToDate      *time.Time     `json:"to_date,omitempty"`
    Limit       int            `json:"limit,omitempty"`
    Offset      int            `json:"offset,omitempty"`
}

type InstanceHealth struct {
    InstanceID      string    `json:"instance_id"`
    HealthStatus    string    `json:"health_status"`
    AgentsHealthy   int       `json:"agents_healthy"`
    AgentsDegraded  int       `json:"agents_degraded"`
    AgentsUnhealthy int       `json:"agents_unhealthy"`
    LastCheck       time.Time `json:"last_check"`
}
```

---

## Database Schema

### Collection: `agency_instances`

**Location**: Each agency's own database (e.g., `{agency_uuid}/agency_instances`)

**Indexes**:
- `tag_id` (for querying instances by tag)
- `state` (for filtering by state)
- `deployed_at` (for chronological sorting)
- Unique: `instance_id`
- Unique: `agency_id + name` (enforce unique names per agency)

**Purpose**: Store all instance records for an agency

**Data Isolation**: All instances for an agency stored in agency-specific database

### Collection: `instance_agents`

**Purpose**: Track which agent references belong to which instance

**Schema**:
```json
{
  "instance_id": "inst-uuid-001",
  "agent_role_code": "developer-agent-001",
  "tag_id": "tag-uuid-123",
  "state": "referenced",
  "added_at": "2025-11-24T10:00:00Z"
}
```

**Note**: Agents are references to tag configurations, not physical agent entities. Agents are created when workflows trigger them.

**Indexes**:
- `instance_id` (primary query pattern)
- `tag_id` (for tag-based queries)

---

## Design Decisions

### Instance Storage
- **Decision**: All instances stored in `agency_instances` collection within the agency's database
- **Rationale**: Maintains agency isolation, simplifies querying, leverages existing multi-tenancy architecture

### Agent Model
- **Decision**: Agents are references to tag configurations, not physically created until workflows trigger them
- **Rationale**: Reduces resource overhead, enables lazy instantiation, preserves immutability of tag snapshots

### Naming Uniqueness
- **Decision**: Instance names must be unique per agency (enforced at creation via unique index)
- **Rationale**: Clear identification, prevents confusion, enables unambiguous references

### Soft Delete
- **Decision**: Instances are soft-deleted (marked deleted, preserved in database)
- **Rationale**: Audit trail preservation, compliance requirements, debugging support

### Uptime Calculation
- **Decision**: Real-time calculation (`current_time - started_at`), not stored
- **Rationale**: No storage overhead, always accurate, avoids sync issues

### Health Status
- **Decision**: Calculated on-demand when requested, based on current agent states
- **Rationale**: Reduces background processing overhead, ensures real-time accuracy

### 🔴 CRITICAL: Workflow Definitions vs Executions

**Decision**: Workflow definitions DO NOT have `instance_id` field. Only workflow executions do.

**The Pattern**:

```go
// ❌ WRONG - Workflow is a definition/template
type Workflow struct {
    ID          string
    Name        string
    Steps       []Step
    InstanceID  string  // ❌ NO! Definitions are agency-wide blueprints
}

// ✅ CORRECT - Workflow is just the blueprint
type Workflow struct {
    ID          string
    Name        string
    Steps       []Step
    // No InstanceID - this is a reusable template
}

// ✅ CORRECT - WorkflowExecution tracks runtime state
type WorkflowExecution struct {
    ID          string
    WorkflowID  string      // Links to definition
    InstanceID  string      // Optional - scopes to instance if provided
    Status      string
    CurrentStep int
    StartTime   time.Time
    // ... execution state ...
}
```

**Affected Collections**:
- ✅ **Agents**: Runtime entities spawned from role definitions → HAVE `instance_id`
- ❌ **Workflows** (definitions): Agency-wide blueprints → DO NOT have `instance_id`
- ✅ **WorkflowExecutions** (runtime): Instance-scoped runs → HAVE `instance_id`
- ✅ **TaskExecutions** (runtime): Part of workflow execution → HAVE `instance_id` via parent

**Rationale**:
- Workflow definitions are shared templates across all instances
- Same workflow definition can run on multiple instances simultaneously
- Execution state is separate from definition
- Instance isolation happens at execution level, not definition level

**References**:
- See `internal/orchestration/types.go` for `WorkflowExecution` model (already implemented)
- See `internal/agency/models/workflow.go` for Workflow definition (should NOT have instance_id)

---

## Related Files

- **Service Implementation**: [instance-services.md](instance-services.md)
- **API Endpoints**: [instance-api.md](instance-api.md)
- **UI Components**: [instance-ui.md](instance-ui.md)
- **Task Overview**: [instance-management.md](instance-management.md)
