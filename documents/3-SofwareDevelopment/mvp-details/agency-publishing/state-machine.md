# MVP-PUB-001: Agency State Machine & Data Models

**Domain**: Agency Publishing & Tagging  
**Priority**: P0 (Critical)  
**Effort**: Medium (1 week)  
**Dependencies**: MVP-044 ✅ (Agency Designer)  

## Overview

Establish the foundational data models and state machine for agency lifecycle management. This task creates the core types, enums, database schema, and state transition logic that all other publishing/tagging features depend on.

## Objectives

1. **Replace simple Status enum** with comprehensive AgencyState lifecycle model
2. **Create AgencyPublication model** for tracking published versions
3. **Create AgencyTag model** for versioning and snapshots
4. **Add ArangoDB collections** with proper indexes and relationships
5. **Implement state transition validation** with guards and actions
6. **Database migration** from old Status to new State field

## Requirements

### 1. AgencyState Enum Enhancement

**Current State** (`internal/agency/models/agency.go`):
```go
type AgencyStatus string

const (
    AgencyStatusActive   AgencyStatus = "active"
    AgencyStatusInactive AgencyStatus = "inactive"
    AgencyStatusPaused   AgencyStatus = "paused"
    AgencyStatusArchived AgencyStatus = "archived"
)
```

**New State** (add alongside Status for backward compatibility):
```go
type AgencyState string

const (
    AgencyStateDraft     AgencyState = "draft"      // Design in progress
    AgencyStateValidated AgencyState = "validated"  // Ready for publishing
    AgencyStatePublished AgencyState = "published"  // Published but not active
    AgencyStateActive    AgencyState = "active"     // Agents running
    AgencyStatePaused    AgencyState = "paused"     // Temporarily suspended
    AgencyStateDraining  AgencyState = "draining"   // Completing existing work
    AgencyStateStopped   AgencyState = "stopped"    // Shut down gracefully
    AgencyStateArchived  AgencyState = "archived"   // Historical record
)
```

**Update Agency struct**:
```go
type Agency struct {
    // ... existing fields ...
    
    Status AgencyStatus `json:"status"` // DEPRECATED: Keep for backward compatibility
    State  AgencyState  `json:"state"`  // NEW: Use this for lifecycle management
    
    // Publishing metadata
    PublishedAt      *time.Time `json:"published_at,omitempty"`
    PublishedBy      string     `json:"published_by,omitempty"`
    PublicationID    string     `json:"publication_id,omitempty"`
    CurrentTagID     string     `json:"current_tag_id,omitempty"`
    
    // Activation metadata
    ActivatedAt      *time.Time `json:"activated_at,omitempty"`
    ActivatedBy      string     `json:"activated_by,omitempty"`
    ActiveAgentCount int        `json:"active_agent_count"`
    
    // ... existing timestamps ...
}
```

### 2. AgencyPublication Model

Create `internal/agency/models/publication.go`:

```go
package models

import "time"

// AgencyPublication represents a published version of an agency
type AgencyPublication struct {
    // ArangoDB fields
    Key string `json:"_key"`
    ID  string `json:"_id"`
    Rev string `json:"_rev,omitempty"`
    
    // Publication metadata
    AgencyID    string    `json:"agency_id"`
    Version     string    `json:"version"`           // Semantic version (v1.0.0)
    TagID       string    `json:"tag_id,omitempty"`  // Optional tag reference
    Description string    `json:"description"`
    
    // Snapshot of agency configuration at publication time
    Snapshot AgencySnapshot `json:"snapshot"`
    
    // Deployment manifest (computed at publish time)
    Manifest DeploymentManifest `json:"manifest"`
    
    // Lifecycle tracking
    PublishedAt   time.Time  `json:"published_at"`
    PublishedBy   string     `json:"published_by"`
    ActivatedAt   *time.Time `json:"activated_at,omitempty"`
    DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
    
    // Metadata
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgencySnapshot captures complete agency state at publication
type AgencySnapshot struct {
    Specification AgencySpecification `json:"specification"`
    AIPolicy      AIPolicy            `json:"ai_policy,omitempty"`
    Settings      AgencySettings      `json:"settings"`
    Metadata      AgencyMetadata      `json:"metadata"`
}

// DeploymentManifest contains computed deployment plan
type DeploymentManifest struct {
    AgentSpawnPlan     AgentSpawnPlan     `json:"agent_spawn_plan"`
    WorkflowExecution  WorkflowExecution  `json:"workflow_execution"`
    ResourceAllocation ResourceAllocation `json:"resource_allocation"`
    MonitoringConfig   MonitoringConfig   `json:"monitoring_config"`
}

// AgentSpawnPlan defines which agents to spawn
type AgentSpawnPlan struct {
    TotalAgents int               `json:"total_agents"`
    Agents      []AgentDefinition `json:"agents"`
}

type AgentDefinition struct {
    RoleCode       string                 `json:"role_code"`
    Name           string                 `json:"name"`
    Type           string                 `json:"type"`
    AutonomyLevel  string                 `json:"autonomy_level"`
    ResourceLimits ResourceLimits         `json:"resource_limits"`
    Configuration  map[string]interface{} `json:"configuration"`
}

type ResourceLimits struct {
    CPULimit    string `json:"cpu_limit"`     // e.g., "1000m"
    MemoryLimit string `json:"memory_limit"`  // e.g., "512Mi"
    TokenBudget int    `json:"token_budget"`
}

// WorkflowExecution defines workflow initialization
type WorkflowExecution struct {
    Workflows []WorkflowConfig `json:"workflows"`
}

type WorkflowConfig struct {
    WorkflowID   string `json:"workflow_id"`
    Name         string `json:"name"`
    Enabled      bool   `json:"enabled"`
    AutoStart    bool   `json:"auto_start"`
}

// ResourceAllocation defines quotas and limits
type ResourceAllocation struct {
    TotalCPU    string `json:"total_cpu"`
    TotalMemory string `json:"total_memory"`
    MaxAgents   int    `json:"max_agents"`
}

// MonitoringConfig defines monitoring setup
type MonitoringConfig struct {
    Enabled         bool     `json:"enabled"`
    MetricsEndpoint string   `json:"metrics_endpoint"`
    Alerts          []string `json:"alerts"`
}
```

### 3. AgencyTag Model

Create `internal/agency/models/tag.go`:

```go
package models

import "time"

// AgencyTag represents an immutable snapshot of an agency
type AgencyTag struct {
    // ArangoDB fields
    Key string `json:"_key"`
    ID  string `json:"_id"`
    Rev string `json:"_rev,omitempty"`
    
    // Tag metadata
    AgencyID    string    `json:"agency_id"`
    Name        string    `json:"name"`              // Unique per agency
    Version     string    `json:"version,omitempty"` // Semantic version (optional)
    Description string    `json:"description"`
    Type        TagType   `json:"type"`
    SHA         string    `json:"sha"` // Content hash (git-style)
    
    // Complete snapshot
    Snapshot AgencySnapshot `json:"snapshot"`
    
    // Additional metadata
    Metadata  TagMetadata `json:"metadata"`
    CreatedAt time.Time   `json:"created_at"`
    CreatedBy string      `json:"created_by"`
}

// TagType defines the purpose of the tag
type TagType string

const (
    TagTypeRelease      TagType = "release"      // Major release version
    TagTypeSnapshot     TagType = "snapshot"     // Point-in-time save
    TagTypeExperimental TagType = "experimental" // Testing variation
    TagTypeCheckpoint   TagType = "checkpoint"   // Design milestone
)

// TagMetadata for custom fields
type TagMetadata struct {
    GitCommit    string                 `json:"git_commit,omitempty"`
    BuildNumber  string                 `json:"build_number,omitempty"`
    Environment  string                 `json:"environment,omitempty"`
    CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ValidTagTypes returns all valid tag types
func ValidTagTypes() []TagType {
    return []TagType{
        TagTypeRelease,
        TagTypeSnapshot,
        TagTypeExperimental,
        TagTypeCheckpoint,
    }
}

// IsValidTagType checks if tag type is valid
func IsValidTagType(t TagType) bool {
    for _, valid := range ValidTagTypes() {
        if t == valid {
            return true
        }
    }
    return false
}
```

### 4. State Machine Implementation

Create `internal/agency/state_machine.go`:

```go
package agency

import (
    "fmt"
    "internal/agency/models"
)

// StateTransition represents a valid state transition
type StateTransition struct {
    From    models.AgencyState
    To      models.AgencyState
    Event   string
    Guards  []Guard
    Actions []Action
}

// Guard is a function that checks if transition is allowed
type Guard func(*models.Agency) error

// Action is a function executed during transition
type Action func(*models.Agency) error

// AgencyStateMachine manages state transitions
type AgencyStateMachine struct {
    transitions map[string][]StateTransition
}

// NewAgencyStateMachine creates a new state machine
func NewAgencyStateMachine() *AgencyStateMachine {
    sm := &AgencyStateMachine{
        transitions: make(map[string][]StateTransition),
    }
    sm.defineTransitions()
    return sm
}

// defineTransitions sets up all valid state transitions
func (sm *AgencyStateMachine) defineTransitions() {
    transitions := []StateTransition{
        {
            From:  models.AgencyStateDraft,
            To:    models.AgencyStateValidated,
            Event: "validate",
            Guards: []Guard{
                guardHasIntroduction,
                guardHasGoals,
                guardHasRoles,
                guardHasWorkItems,
                guardHasWorkflows,
                guardHasRACIMatrix,
            },
            Actions: []Action{
                actionMarkAsValidated,
            },
        },
        {
            From:  models.AgencyStateValidated,
            To:    models.AgencyStatePublished,
            Event: "publish",
            Guards: []Guard{
                guardIsValidated,
                guardNoDuplicatePublication,
            },
            Actions: []Action{
                actionCreatePublication,
                actionUpdatePublishMetadata,
            },
        },
        {
            From:  models.AgencyStatePublished,
            To:    models.AgencyStateActive,
            Event: "activate",
            Guards: []Guard{
                guardIsPublished,
                guardResourcesAvailable,
            },
            Actions: []Action{
                actionSpawnAgents,
                actionInitializeWorkflows,
                actionStartMonitoring,
                actionUpdateActivationMetadata,
            },
        },
        {
            From:  models.AgencyStateActive,
            To:    models.AgencyStatePaused,
            Event: "pause",
            Actions: []Action{
                actionStopAcceptingWork,
                actionPauseAgents,
            },
        },
        {
            From:  models.AgencyStatePaused,
            To:    models.AgencyStateActive,
            Event: "resume",
            Actions: []Action{
                actionResumeAgents,
                actionResumeAcceptingWork,
            },
        },
        {
            From:  models.AgencyStateActive,
            To:    models.AgencyStateDraining,
            Event: "drain",
            Actions: []Action{
                actionStopAcceptingNewWork,
            },
        },
        {
            From:  models.AgencyStateDraining,
            To:    models.AgencyStateStopped,
            Event: "drain_complete",
            Actions: []Action{
                actionStopAllAgents,
                actionCleanupResources,
            },
        },
    }
    
    for _, t := range transitions {
        key := string(t.From)
        sm.transitions[key] = append(sm.transitions[key], t)
    }
}

// CanTransition checks if transition is allowed
func (sm *AgencyStateMachine) CanTransition(agency *models.Agency, event string) error {
    transitions, ok := sm.transitions[string(agency.State)]
    if !ok {
        return fmt.Errorf("no transitions defined for state: %s", agency.State)
    }
    
    for _, t := range transitions {
        if t.Event == event {
            // Check guards
            for _, guard := range t.Guards {
                if err := guard(agency); err != nil {
                    return fmt.Errorf("guard failed: %w", err)
                }
            }
            return nil
        }
    }
    
    return fmt.Errorf("event '%s' not valid for state '%s'", event, agency.State)
}

// Transition executes a state transition
func (sm *AgencyStateMachine) Transition(agency *models.Agency, event string) error {
    if err := sm.CanTransition(agency, event); err != nil {
        return err
    }
    
    transitions := sm.transitions[string(agency.State)]
    for _, t := range transitions {
        if t.Event == event {
            // Execute actions
            for _, action := range t.Actions {
                if err := action(agency); err != nil {
                    return fmt.Errorf("action failed: %w", err)
                }
            }
            
            // Update state
            agency.State = t.To
            return nil
        }
    }
    
    return fmt.Errorf("transition not found")
}

// Guard implementations (stubs for now - will be implemented in later tasks)
func guardHasIntroduction(a *models.Agency) error {
    // TODO: Check agency has introduction
    return nil
}

func guardHasGoals(a *models.Agency) error {
    // TODO: Check agency has goals
    return nil
}

func guardHasRoles(a *models.Agency) error {
    // TODO: Check agency has roles
    return nil
}

func guardHasWorkItems(a *models.Agency) error {
    // TODO: Check agency has work items
    return nil
}

func guardHasWorkflows(a *models.Agency) error {
    // TODO: Check agency has workflows
    return nil
}

func guardHasRACIMatrix(a *models.Agency) error {
    // TODO: Check agency has RACI matrix
    return nil
}

func guardIsValidated(a *models.Agency) error {
    if a.State != models.AgencyStateValidated {
        return fmt.Errorf("agency must be validated")
    }
    return nil
}

func guardNoDuplicatePublication(a *models.Agency) error {
    // TODO: Check no active publication exists
    return nil
}

func guardIsPublished(a *models.Agency) error {
    if a.State != models.AgencyStatePublished {
        return fmt.Errorf("agency must be published")
    }
    return nil
}

func guardResourcesAvailable(a *models.Agency) error {
    // TODO: Check resource quotas available
    return nil
}

// Action implementations (stubs for now)
func actionMarkAsValidated(a *models.Agency) error {
    // Will be implemented in validation logic
    return nil
}

func actionCreatePublication(a *models.Agency) error {
    // Will be implemented in MVP-PUB-003
    return nil
}

func actionUpdatePublishMetadata(a *models.Agency) error {
    // Will be implemented in MVP-PUB-003
    return nil
}

func actionSpawnAgents(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionInitializeWorkflows(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionStartMonitoring(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionUpdateActivationMetadata(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionStopAcceptingWork(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionPauseAgents(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionResumeAgents(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionResumeAcceptingWork(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionStopAcceptingNewWork(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionStopAllAgents(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}

func actionCleanupResources(a *models.Agency) error {
    // Will be implemented in MVP-PUB-004
    return nil
}
```

### 5. ArangoDB Collections

**Collections to Create**:

1. **`agency_publications`** - Published agency versions
2. **`agency_tags`** - Agency snapshots/tags

**Indexes Required**:

```javascript
// agency_publications
db.agency_publications.ensureIndex({ type: "hash", fields: ["agency_id"] });
db.agency_publications.ensureIndex({ type: "skiplist", fields: ["published_at"] });
db.agency_publications.ensureIndex({ type: "hash", fields: ["version"] });

// agency_tags
db.agency_tags.ensureIndex({ type: "hash", fields: ["agency_id"] });
db.agency_tags.ensureIndex({ type: "hash", fields: ["agency_id", "name"], unique: true });
db.agency_tags.ensureIndex({ type: "skiplist", fields: ["version"] });
db.agency_tags.ensureIndex({ type: "skiplist", fields: ["created_at"] });
db.agency_tags.ensureIndex({ type: "hash", fields: ["type"] });
```

Create `internal/database/migrations/006_agency_publishing.go`:

```go
package migrations

import (
    "context"
    "fmt"
    
    "github.com/arangodb/go-driver"
)

// Migration006AgencyPublishing adds publishing/tagging collections
func Migration006AgencyPublishing(ctx context.Context, db driver.Database) error {
    // Create agency_publications collection
    if err := createPublicationsCollection(ctx, db); err != nil {
        return fmt.Errorf("failed to create publications collection: %w", err)
    }
    
    // Create agency_tags collection
    if err := createTagsCollection(ctx, db); err != nil {
        return fmt.Errorf("failed to create tags collection: %w", err)
    }
    
    return nil
}

func createPublicationsCollection(ctx context.Context, db driver.Database) error {
    // Create collection
    _, err := db.CreateCollection(ctx, "agency_publications", nil)
    if err != nil && !driver.IsConflict(err) {
        return err
    }
    
    col, err := db.Collection(ctx, "agency_publications")
    if err != nil {
        return err
    }
    
    // Create indexes
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
    if err != nil {
        return err
    }
    
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"published_at"}, nil)
    if err != nil {
        return err
    }
    
    _, _, err = col.EnsureHashIndex(ctx, []string{"version"}, nil)
    if err != nil {
        return err
    }
    
    return nil
}

func createTagsCollection(ctx context.Context, db driver.Database) error {
    // Create collection
    _, err := db.CreateCollection(ctx, "agency_tags", nil)
    if err != nil && !driver.IsConflict(err) {
        return err
    }
    
    col, err := db.Collection(ctx, "agency_tags")
    if err != nil {
        return err
    }
    
    // Create indexes
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id"}, nil)
    if err != nil {
        return err
    }
    
    // Unique index on agency_id + name
    _, _, err = col.EnsureHashIndex(ctx, []string{"agency_id", "name"}, &driver.EnsureHashIndexOptions{
        Unique: true,
    })
    if err != nil {
        return err
    }
    
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"version"}, nil)
    if err != nil {
        return err
    }
    
    _, _, err = col.EnsureSkipListIndex(ctx, []string{"created_at"}, nil)
    if err != nil {
        return err
    }
    
    _, _, err = col.EnsureHashIndex(ctx, []string{"type"}, nil)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 6. Data Migration from Status to State

Create `scripts/migrate-agency-status-to-state.go`:

```go
package main

// Migration script to populate State field from Status field
// Run once after deploying new models

func migrateAgencyStatusToState() {
    // Mapping:
    // active   → active (if active_agent_count > 0) OR published (if active_agent_count = 0)
    // inactive → draft OR stopped
    // paused   → paused
    // archived → archived
}
```

## Acceptance Criteria

- [ ] AgencyState enum defined with 8 states
- [ ] Agency model updated with State, publishing, and activation metadata fields
- [ ] AgencyPublication model created with complete schema
- [ ] AgencyTag model created with TagType enum and metadata
- [ ] State machine implemented with transition validation
- [ ] Guard functions defined (stubs for later implementation)
- [ ] Action functions defined (stubs for later tasks)
- [ ] ArangoDB collections created: `agency_publications`, `agency_tags`
- [ ] Proper indexes created on both collections
- [ ] Database migration script created and tested
- [ ] Data migration script created for Status→State conversion
- [ ] All models have proper JSON tags
- [ ] Code follows Go naming conventions
- [ ] Documentation comments added to all exported types

## Files to Create/Modify

**Create**:
- `internal/agency/models/publication.go` (~150 lines)
- `internal/agency/models/tag.go` (~80 lines)
- `internal/agency/state_machine.go` (~300 lines)
- `internal/database/migrations/006_agency_publishing.go` (~120 lines)
- `scripts/migrate-agency-status-to-state.go` (~100 lines)

**Modify**:
- `internal/agency/models/agency.go` (add State field and publishing metadata)

**Total Estimated**: ~750 lines of new/modified code

## Testing Strategy

### Unit Tests

Create `internal/agency/state_machine_test.go`:
- Test valid transitions
- Test invalid transitions
- Test guard evaluation
- Test action execution order

Create `internal/agency/models/tag_test.go`:
- Test TagType validation
- Test tag model JSON marshaling

### Integration Tests

- Test database collection creation
- Test index creation
- Test migration script with sample data
- Verify Status→State mapping logic

## Dependencies

**External Packages**:
- `github.com/arangodb/go-driver` (already in project)

**Internal Dependencies**:
- `internal/agency/models` (existing)
- `internal/database` (existing)

## Migration Strategy

1. **Add new fields** to Agency model (backward compatible)
2. **Deploy new models** (State field optional initially)
3. **Run migration script** to populate State from Status
4. **Update application code** to use State instead of Status
5. **Deprecate Status field** (keep for 1-2 releases for rollback safety)

## Next Steps

After MVP-PUB-001 completion, proceed to:
- **MVP-PUB-002**: Tag Service Implementation (uses Tag model and state machine)
