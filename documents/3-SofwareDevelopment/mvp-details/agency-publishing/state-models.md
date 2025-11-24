# Agency State Models & Data Structures (MVP-PUB-001)

**Domain**: Agency Publishing & Tagging  
**Priority**: P0 (Critical)  
**Status**: ✅ Complete (2025-11-20)

## Overview

This document covers the data models and structures for agency lifecycle management:
- **AgencyState Enum**: 8-state lifecycle model replacing simple Status enum
- **AgencyPublication Model**: Tracking published versions with deployment manifests
- **AgencyTag Model**: Immutable snapshots for version control and experimentation

Related documentation:
- [State Transitions](./state-transitions.md) - State machine logic, guards, and actions
- [State Database](./state-database.md) - ArangoDB collections, indexes, and migration

## 1. AgencyState Enum Enhancement

### Previous State (Deprecated)

**Original Status Enum** (`internal/agency/models/agency.go`):
```go
type AgencyStatus string

const (
    AgencyStatusActive   AgencyStatus = "active"
    AgencyStatusInactive AgencyStatus = "inactive"
    AgencyStatusPaused   AgencyStatus = "paused"
    AgencyStatusArchived AgencyStatus = "archived"
)
```

**Limitation**: Too simplistic - doesn't capture design → validation → publishing → activation flow.

### New State Model

**AgencyState Enum** (8 lifecycle states):
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

**State Descriptions**:

| State | Meaning | Key Characteristics |
|-------|---------|---------------------|
| `draft` | Initial design state | Introduction, goals, roles being defined. Not validated. |
| `validated` | Ready for publishing | All required components present. Passed validation checks. |
| `published` | Deployment plan created | Publication record exists, not yet active. Can be activated. |
| `active` | Agents running | Agents spawned, workflows initialized, processing work. |
| `paused` | Temporarily suspended | Agents exist but not processing. Can resume quickly. |
| `draining` | Completing existing work | No new work accepted. Waiting for in-flight tasks to finish. |
| `stopped` | Shut down gracefully | All agents terminated. Resources released. |
| `archived` | Historical record | Read-only. Preserved for audit/compliance. |

### Updated Agency Model

**Agency struct** with State and publishing metadata:
```go
type Agency struct {
    // ... existing fields (ID, Name, Introduction, etc.) ...
    
    // DEPRECATED: Keep for backward compatibility during migration
    Status AgencyStatus `json:"status"`
    
    // NEW: Primary lifecycle field
    State  AgencyState  `json:"state"`
    
    // Publishing metadata
    PublishedAt      *time.Time `json:"published_at,omitempty"`
    PublishedBy      string     `json:"published_by,omitempty"`
    PublicationID    string     `json:"publication_id,omitempty"`
    CurrentTagID     string     `json:"current_tag_id,omitempty"`
    
    // Activation metadata
    ActivatedAt      *time.Time `json:"activated_at,omitempty"`
    ActivatedBy      string     `json:"activated_by,omitempty"`
    ActiveAgentCount int        `json:"active_agent_count"`
    
    // ... existing timestamps (CreatedAt, UpdatedAt) ...
}
```

**Key Changes**:
- `State` is the authoritative lifecycle field (use this, not `Status`)
- `PublicationID` links to `agency_publications` collection
- `CurrentTagID` tracks which tag agency is running from (if any)
- `ActiveAgentCount` enables "active" vs "published" distinction

## 2. AgencyPublication Model

**Purpose**: Track published versions of agencies with complete deployment manifests.

**File**: `internal/agency/models/publication.go`

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

**Key Characteristics**:
- **Immutable snapshot**: Captures complete agency config at publish time
- **Deployment manifest**: Pre-computed plan for activation (no on-the-fly calculations)
- **Lifecycle tracking**: Records when published, activated, deactivated
- **Optional tag reference**: Can publish from a tag or directly from draft

## 3. AgencyTag Model

**Purpose**: Create immutable snapshots of agencies for version control, rollback, and experimentation.

**File**: `internal/agency/models/tag.go`

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

**Key Characteristics**:
- **Immutable**: Once created, tags cannot be modified
- **Content hash (SHA)**: Git-style hash to detect changes
- **Unique name per agency**: Prevents naming conflicts
- **Type categorization**: Different purposes (release, snapshot, experimental, checkpoint)
- **Semantic versioning**: Optional version field for release tags
- **Custom metadata**: Extensible for CI/CD integration (git commits, build numbers)

**Tag vs Publication Distinction**:
- **Tags** are snapshots - multiple tags can exist, no activation state
- **Publications** are deployments - one active publication at a time, tracks activation lifecycle

## Data Relationships

```
Agency
  |
  |-- (1:N) Publications  (agency_id → agency_publications)
  |     |
  |     └── Snapshot (embedded)
  |     └── Manifest (embedded)
  |
  └-- (1:N) Tags  (agency_id → agency_tags)
        |
        └── Snapshot (embedded)
```

**Key Points**:
- Agency has many publications and many tags
- Current publication: `Agency.PublicationID` → `AgencyPublication._id`
- Current tag: `Agency.CurrentTagID` → `AgencyTag._id`
- Publications can reference tags: `Publication.TagID` → `Tag._id`

## Migration Notes

**Status → State Mapping**:

```
Old Status      → New State (depends on context)
--------------------------------------------------
active          → active (if ActiveAgentCount > 0)
                → published (if ActiveAgentCount = 0)
inactive        → draft (if never published)
                → stopped (if was published)
paused          → paused
archived        → archived
```

**Backward Compatibility**:
- Keep `Status` field during migration period
- Populate both `Status` and `State` in writes
- Application code should read from `State` (authoritative)
- Can remove `Status` after 1-2 releases

## Acceptance Criteria

- [x] AgencyState enum defined with 8 states
- [x] Agency model updated with State, publishing, and activation metadata fields
- [x] AgencyPublication model created with complete schema (snapshot + manifest)
- [x] AgencyTag model created with TagType enum and metadata
- [x] All models have proper JSON tags
- [x] Code follows Go naming conventions
- [x] Documentation comments added to all exported types
- [x] Models support both embedded and referenced relationships
- [x] SHA hash generation logic defined for tags
- [x] Semantic versioning support in tags and publications
