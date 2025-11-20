# Agency Publishing and Tagging Architecture

**Version**: 1.0  
**Date**: November 20, 2025  
**Status**: Design Specification  

## Overview

This document defines the architecture for **publishing** and **tagging** agencies in CodeValdCortex. Publishing transforms an agency from design mode to production operation, while tagging creates immutable snapshots for versioning, rollback, and experimentation.

## Table of Contents

1. [Conceptual Model](#conceptual-model)
2. [Agency Lifecycle States](#agency-lifecycle-states)
3. [Tagging System](#tagging-system)
4. [Publishing Workflow](#publishing-workflow)
5. [Data Models](#data-models)
6. [Service Architecture](#service-architecture)
7. [UI/UX Design](#uiux-design)
8. [Implementation Plan](#implementation-plan)

---

## 1. Conceptual Model

### 1.1 Core Concepts

**Agency**: A complete AI agent system designed for a specific use case (e.g., "Water Distribution Network Management")

**Publishing**: The process of activating an agency design for production use
- Validates completeness and correctness
- Spawns agents according to the design
- Starts workflow execution
- Enables production monitoring

**Tagging**: Creating an immutable copy of an agency at a specific point in time
- Preserves exact configuration snapshot
- Enables version comparison
- Supports rollback scenarios
- Facilitates experimentation

### 1.2 Key Distinctions

| Aspect | Tag | Publish |
|--------|-----|---------|
| Purpose | Snapshot/version | Go-live operation |
| Mutability | Immutable | Can be stopped/started |
| Agents | No agents spawned | Agents created and started |
| State | Static copy | Dynamic runtime |
| Use Cases | Backup, rollback, experimentation | Production operation |

---

## 2. Agency Lifecycle States

### 2.1 State Machine

```
┌─────────────┐
│    Draft    │ (Design in progress)
└──────┬──────┘
       │ validate()
       ▼
┌─────────────┐
│  Validated  │ (Ready for publishing)
└──────┬──────┘
       │ publish()
       ▼
┌─────────────┐◄──────────┐
│  Published  │           │ republish_tag()
└──────┬──────┘           │
       │                  │
       ├──activate()──────►┌─────────────┐
       │                   │   Active    │ (Agents running)
       │                   └──────┬──────┘
       │                          │
       │                          │ pause()
       │                   ┌──────▼──────┐
       │                   │   Paused    │ (Agents stopped, config preserved)
       │                   └──────┬──────┘
       │                          │ resume()
       │                          │
       │                   ┌──────▼──────┐
       │                   │  Draining   │ (Completing work, no new tasks)
       │                   └──────┬──────┘
       │                          │
       └───────────────────────────┼──────► Archive/Stop
                                   │
                            ┌──────▼──────┐
                            │   Stopped   │ (No agents, config preserved)
                            └─────────────┘
```

### 2.2 State Definitions

| State | Description | Allowed Operations | Agents Running | Accepts Work |
|-------|-------------|-------------------|----------------|--------------|
| **Draft** | Agency being designed | Edit, validate, tag, delete | No | No |
| **Validated** | Design complete, ready for publish | Edit, publish, tag | No | No |
| **Published** | Published but not yet activated | Activate, edit draft, tag | No | No |
| **Active** | Agents running, accepting work | Pause, drain, stop, tag | Yes | Yes |
| **Paused** | Temporarily suspended | Resume, stop, tag | No | No |
| **Draining** | Completing existing work | Monitor, force-stop | Yes | No |
| **Stopped** | Shut down gracefully | Restart, archive, delete | No | No |
| **Archived** | Historical record | View, tag, restore | No | No |

### 2.3 State Transitions

```typescript
interface AgencyStateTransition {
  from: AgencyState;
  to: AgencyState;
  event: string;
  guards?: Guard[];
  actions: Action[];
}

const transitions: AgencyStateTransition[] = [
  {
    from: 'draft',
    to: 'validated',
    event: 'validate',
    guards: [
      'hasIntroduction',
      'hasGoals',
      'hasRoles',
      'hasWorkItems',
      'hasValidWorkflows',
      'hasRACIMatrix'
    ],
    actions: ['runValidation', 'markAsValidated']
  },
  {
    from: 'validated',
    to: 'published',
    event: 'publish',
    guards: ['isValidated', 'noDuplicatePublication'],
    actions: [
      'createPublishedSnapshot',
      'generateDeploymentManifest',
      'updateState'
    ]
  },
  {
    from: 'published',
    to: 'active',
    event: 'activate',
    guards: ['isPublished', 'resourcesAvailable'],
    actions: [
      'spawnAgents',
      'initializeWorkflows',
      'startMonitoring',
      'updateState'
    ]
  },
  {
    from: 'active',
    to: 'paused',
    event: 'pause',
    actions: [
      'stopAcceptingWork',
      'pauseAgents',
      'updateState'
    ]
  },
  {
    from: 'paused',
    to: 'active',
    event: 'resume',
    actions: [
      'resumeAgents',
      'resumeAcceptingWork',
      'updateState'
    ]
  },
  {
    from: 'active',
    to: 'draining',
    event: 'drain',
    actions: [
      'stopAcceptingNewWork',
      'waitForCompletion',
      'updateState'
    ]
  },
  {
    from: 'draining',
    to: 'stopped',
    event: 'drain_complete',
    actions: [
      'stopAllAgents',
      'cleanupResources',
      'updateState'
    ]
  },
  {
    from: ['active', 'paused', 'draining'],
    to: 'stopped',
    event: 'force_stop',
    actions: [
      'forceStopAgents',
      'cleanupResources',
      'updateState'
    ]
  }
];
```

---

## 3. Tagging System

### 3.1 Tag Concept

A **tag** is an immutable, versioned copy of an agency's complete configuration at a specific point in time.

**Tag Characteristics**:
- Immutable once created
- Includes complete specification (introduction, goals, roles, work items, workflows, RACI matrix)
- Includes AI policy configuration
- Semantic versioning support (v1.0.0, v1.1.0, v2.0.0)
- Metadata: creator, timestamp, description, git-style SHA
- Can be created from any state (draft, published, active)

### 3.2 Tag Types

| Type | Purpose | Naming Convention | Example |
|------|---------|-------------------|---------|
| **Release** | Major versions | `v{major}.{minor}.{patch}` | `v1.0.0`, `v2.3.1` |
| **Snapshot** | Point-in-time saves | `snapshot-{timestamp}` | `snapshot-20251120-143022` |
| **Experimental** | Testing variations | `exp-{name}-{timestamp}` | `exp-new-workflow-20251120` |
| **Checkpoint** | Design milestones | `checkpoint-{name}` | `checkpoint-after-review` |

### 3.3 Tag Operations

```typescript
interface TagOperations {
  // Create tag from current agency state
  createTag(agencyID: string, tag: TagRequest): Promise<AgencyTag>;
  
  // List all tags for an agency
  listTags(agencyID: string, filters?: TagFilters): Promise<AgencyTag[]>;
  
  // Get specific tag
  getTag(agencyID: string, tagName: string): Promise<AgencyTag>;
  
  // Compare two tags (diff)
  compareTags(tagA: string, tagB: string): Promise<TagDiff>;
  
  // Restore agency from tag
  restoreFromTag(agencyID: string, tagName: string): Promise<Agency>;
  
  // Publish a tag (create new published agency from tag)
  publishTag(tagName: string): Promise<Agency>;
  
  // Delete tag (with safety checks)
  deleteTag(agencyID: string, tagName: string): Promise<void>;
}

interface TagRequest {
  name: string;              // Tag name (must be unique per agency)
  version?: string;          // Semantic version (optional)
  description: string;       // What this tag represents
  type: TagType;            // release | snapshot | experimental | checkpoint
  metadata?: TagMetadata;    // Additional context
}

interface AgencyTag {
  id: string;               // Unique tag ID
  agencyID: string;         // Source agency ID
  name: string;             // Tag name
  version?: string;         // Semantic version
  description: string;      // Tag description
  type: TagType;
  sha: string;              // Content hash (git-style)
  snapshot: AgencySnapshot; // Complete agency configuration
  metadata: TagMetadata;
  createdAt: Date;
  createdBy: string;
}

interface AgencySnapshot {
  // Complete agency state at time of tagging
  specification: AgencySpecification;
  aiPolicy: AIPolicy;
  settings: AgencySettings;
  metadata: AgencyMetadata;
}
```

### 3.4 Tag Storage

**ArangoDB Collection**: `agency_tags`

**Schema**:
```json
{
  "_key": "tag_water-network-v1.0.0",
  "_id": "agency_tags/tag_water-network-v1.0.0",
  "agency_id": "UC-INFRA-001",
  "name": "v1.0.0",
  "version": "1.0.0",
  "description": "Initial production release",
  "type": "release",
  "sha": "a3f5c8d9e2b1...",
  "snapshot": {
    "specification": {
      "introduction": "...",
      "goals": [...],
      "roles": [...],
      "work_items": [...],
      "workflows": [...],
      "raci_matrix": {...}
    },
    "ai_policy": {...},
    "settings": {...},
    "metadata": {...}
  },
  "metadata": {
    "git_commit": "abc123...",
    "build_number": "42",
    "environment": "production"
  },
  "created_at": "2025-11-20T14:30:00Z",
  "created_by": "user@example.com"
}
```

**Indexes**:
- `agency_id` (hash index)
- `name` (unique, per agency_id)
- `version` (skip list, for semver sorting)
- `created_at` (skip list, for chronological queries)
- `type` (hash index)

---

## 4. Publishing Workflow

### 4.1 Pre-Publish Validation

Before publishing, the system validates:

```typescript
interface PublishValidation {
  // Required components
  hasIntroduction: boolean;
  hasGoals: boolean;
  hasRoles: boolean;
  hasWorkItems: boolean;
  hasWorkflows: boolean;
  hasRACIMatrix: boolean;
  
  // Completeness checks
  allGoalsHaveOwners: boolean;
  allWorkItemsAssigned: boolean;
  allWorkflowsValid: boolean;
  raciMatrixComplete: boolean;
  
  // Dependency checks
  workItemsHaveDependenciesResolved: boolean;
  workflowStepsHaveWorkItems: boolean;
  
  // Resource checks
  resourceQuotasAvailable: boolean;
  databaseInitialized: boolean;
  
  // Policy checks
  aiPolicyDefined: boolean;
  complianceChecksPass: boolean;
}

interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
  recommendations: string[];
}
```

### 4.2 Publish Process

```
┌─────────────────────────────────────────────────────────┐
│                   PUBLISH WORKFLOW                       │
└─────────────────────────────────────────────────────────┘

1. Pre-Publish Validation
   ├─► Validate specification completeness
   ├─► Check resource availability
   ├─► Verify database readiness
   └─► Run compliance checks
        │
        ▼
2. Create Published Snapshot
   ├─► Generate publication ID
   ├─► Deep copy specification
   ├─► Freeze configuration (immutable)
   └─► Store in agency_publications collection
        │
        ▼
3. Generate Deployment Manifest
   ├─► Compute agent spawn plan
   ├─► Create workflow execution graph
   ├─► Allocate resource quotas
   └─► Generate monitoring config
        │
        ▼
4. Initialize Infrastructure
   ├─► Ensure agency database exists
   ├─► Create required collections
   ├─► Set up indexes
   └─► Initialize event bus subscriptions
        │
        ▼
5. Update Agency State
   ├─► Transition: validated → published
   ├─► Record publication metadata
   ├─► Create audit trail entry
   └─► Notify stakeholders
        │
        ▼
6. (Optional) Auto-Activate
   ├─► Spawn agents
   ├─► Start workflows
   ├─► Begin monitoring
   └─► Transition: published → active
```

### 4.3 Activation Process

When activating a published agency:

```typescript
interface ActivationProcess {
  // Phase 1: Agent Spawning
  spawnAgents(): Promise<AgentSpawnResult> {
    // For each role in specification:
    // 1. Create agent instance
    // 2. Assign to pool
    // 3. Allocate resources
    // 4. Initialize health monitoring
    // 5. Transition to 'starting' state
  }
  
  // Phase 2: Workflow Initialization
  initializeWorkflows(): Promise<WorkflowInitResult> {
    // For each workflow:
    // 1. Parse workflow definition
    // 2. Create workflow engine instance
    // 3. Register event handlers
    // 4. Enable workflow execution
  }
  
  // Phase 3: Work Item Monitoring
  enableWorkItemMonitoring(): Promise<void> {
    // 1. Subscribe to work item change streams
    // 2. Set up milestone-based triggers
    // 3. Configure orchestrator rules
  }
  
  // Phase 4: Health Checks
  startHealthMonitoring(): Promise<void> {
    // 1. Enable agent health probes
    // 2. Configure circuit breakers
    // 3. Set up quarantine monitoring
  }
}
```

---

## 5. Data Models

### 5.1 Enhanced Agency Model

```go
package models

import "time"

// Agency represents a use case with enhanced lifecycle management
type Agency struct {
    // ArangoDB fields
    Key string `json:"_key,omitempty"`
    ID  string `json:"id"`
    Rev string `json:"_rev,omitempty"`
    
    // Core fields
    Name        string         `json:"name"`
    DisplayName string         `json:"display_name"`
    Description string         `json:"description"`
    Category    string         `json:"category"`
    Icon        string         `json:"icon"`
    
    // Lifecycle state (enhanced)
    State      AgencyState    `json:"state"`
    Status     AgencyStatus   `json:"status"` // deprecated, use State
    
    // Publishing metadata
    PublishedAt      *time.Time `json:"published_at,omitempty"`
    PublishedBy      string     `json:"published_by,omitempty"`
    PublicationID    string     `json:"publication_id,omitempty"`
    CurrentTagID     string     `json:"current_tag_id,omitempty"`
    
    // Activation metadata
    ActivatedAt      *time.Time `json:"activated_at,omitempty"`
    ActivatedBy      string     `json:"activated_by,omitempty"`
    ActiveAgentCount int        `json:"active_agent_count"`
    
    // Infrastructure
    Database    string         `json:"database"`
    Metadata    AgencyMetadata `json:"metadata"`
    Settings    AgencySettings `json:"settings"`
    
    // Audit trail
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    CreatedBy   string         `json:"created_by"`
    UpdatedBy   string         `json:"updated_by,omitempty"`
}

// AgencyState represents the lifecycle state
type AgencyState string

const (
    AgencyStateDraft     AgencyState = "draft"
    AgencyStateValidated AgencyState = "validated"
    AgencyStatePublished AgencyState = "published"
    AgencyStateActive    AgencyState = "active"
    AgencyStatePaused    AgencyState = "paused"
    AgencyStateDraining  AgencyState = "draining"
    AgencyStateStopped   AgencyState = "stopped"
    AgencyStateArchived  AgencyState = "archived"
)

// AgencyPublication represents a published version of an agency
type AgencyPublication struct {
    ID              string              `json:"_key"`
    AgencyID        string              `json:"agency_id"`
    Version         string              `json:"version"`
    TagID           string              `json:"tag_id,omitempty"`
    Snapshot        AgencySnapshot      `json:"snapshot"`
    Manifest        DeploymentManifest  `json:"manifest"`
    PublishedAt     time.Time           `json:"published_at"`
    PublishedBy     string              `json:"published_by"`
    ActivatedAt     *time.Time          `json:"activated_at,omitempty"`
    DeactivatedAt   *time.Time          `json:"deactivated_at,omitempty"`
}

// DeploymentManifest contains computed deployment plan
type DeploymentManifest struct {
    AgentSpawnPlan      AgentSpawnPlan      `json:"agent_spawn_plan"`
    WorkflowExecution   WorkflowExecution   `json:"workflow_execution"`
    ResourceAllocation  ResourceAllocation  `json:"resource_allocation"`
    MonitoringConfig    MonitoringConfig    `json:"monitoring_config"`
}

// AgentSpawnPlan defines which agents to spawn
type AgentSpawnPlan struct {
    Agents []AgentDefinition `json:"agents"`
}

type AgentDefinition struct {
    RoleCode       string                 `json:"role_code"`
    Name           string                 `json:"name"`
    Type           string                 `json:"type"`
    AutonomyLevel  string                 `json:"autonomy_level"`
    ResourceLimits ResourceLimits         `json:"resource_limits"`
    Configuration  map[string]interface{} `json:"configuration"`
}
```

### 5.2 Tag Model

```go
// AgencyTag represents an immutable snapshot of an agency
type AgencyTag struct {
    // ArangoDB fields
    Key string `json:"_key"`
    ID  string `json:"_id"`
    
    // Tag metadata
    AgencyID    string    `json:"agency_id"`
    Name        string    `json:"name"`
    Version     string    `json:"version,omitempty"`
    Description string    `json:"description"`
    Type        TagType   `json:"type"`
    SHA         string    `json:"sha"` // Content hash
    
    // Snapshot
    Snapshot AgencySnapshot `json:"snapshot"`
    
    // Additional metadata
    Metadata  TagMetadata `json:"metadata"`
    CreatedAt time.Time   `json:"created_at"`
    CreatedBy string      `json:"created_by"`
}

type TagType string

const (
    TagTypeRelease      TagType = "release"
    TagTypeSnapshot     TagType = "snapshot"
    TagTypeExperimental TagType = "experimental"
    TagTypeCheckpoint   TagType = "checkpoint"
)

type AgencySnapshot struct {
    Specification AgencySpecification `json:"specification"`
    AIPolicy      AIPolicy            `json:"ai_policy"`
    Settings      AgencySettings      `json:"settings"`
    Metadata      AgencyMetadata      `json:"metadata"`
}

type TagMetadata struct {
    GitCommit     string                 `json:"git_commit,omitempty"`
    BuildNumber   string                 `json:"build_number,omitempty"`
    Environment   string                 `json:"environment,omitempty"`
    CustomFields  map[string]interface{} `json:"custom_fields,omitempty"`
}
```

---

## 6. Service Architecture

### 6.1 Service Hierarchy

```
┌──────────────────────────────────────────────┐
│         Agency Lifecycle Services             │
├──────────────────────────────────────────────┤
│                                               │
│  ┌────────────────────────────────────────┐ │
│  │   AgencyPublicationService             │ │
│  │   - ValidateForPublish()               │ │
│  │   - Publish()                          │ │
│  │   - Activate()                         │ │
│  │   - Deactivate()                       │ │
│  │   - GetPublicationHistory()            │ │
│  └─────────────┬──────────────────────────┘ │
│                │                             │
│  ┌─────────────▼──────────────────────────┐ │
│  │   AgencyTagService                     │ │
│  │   - CreateTag()                        │ │
│  │   - ListTags()                         │ │
│  │   - GetTag()                           │ │
│  │   - CompareTags()                      │ │
│  │   - RestoreFromTag()                   │ │
│  │   - PublishTag()                       │ │
│  └─────────────┬──────────────────────────┘ │
│                │                             │
│  ┌─────────────▼──────────────────────────┐ │
│  │   AgencyActivationService              │ │
│  │   - SpawnAgents()                      │ │
│  │   - InitializeWorkflows()              │ │
│  │   - StartMonitoring()                  │ │
│  │   - PauseAgency()                      │ │
│  │   - ResumeAgency()                     │ │
│  │   - DrainAgency()                      │ │
│  │   - StopAgency()                       │ │
│  └────────────────────────────────────────┘ │
│                                               │
└──────────────────────────────────────────────┘
```

### 6.2 AgencyPublicationService

```go
package agency

import (
    "context"
    "time"
)

type PublicationService interface {
    // Validate agency is ready for publishing
    ValidateForPublish(ctx context.Context, agencyID string) (*ValidationResult, error)
    
    // Publish agency (create immutable publication)
    Publish(ctx context.Context, agencyID string, req *PublishRequest) (*AgencyPublication, error)
    
    // Activate published agency (spawn agents, start workflows)
    Activate(ctx context.Context, publicationID string) (*ActivationResult, error)
    
    // Deactivate running agency (stop agents, preserve state)
    Deactivate(ctx context.Context, agencyID string, graceful bool) error
    
    // Get publication history for agency
    GetPublicationHistory(ctx context.Context, agencyID string) ([]*AgencyPublication, error)
    
    // Republish from tag (create new publication from tag)
    RepublishFromTag(ctx context.Context, tagID string) (*AgencyPublication, error)
}

type PublishRequest struct {
    Version     string            `json:"version"`
    Description string            `json:"description"`
    AutoActivate bool             `json:"auto_activate"`
    CreateTag   bool              `json:"create_tag"`
    TagName     string            `json:"tag_name,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type ValidationResult struct {
    Valid         bool                `json:"valid"`
    Errors        []ValidationError   `json:"errors"`
    Warnings      []ValidationWarning `json:"warnings"`
    Recommendations []string          `json:"recommendations"`
}

type ActivationResult struct {
    AgentsSpawned      int       `json:"agents_spawned"`
    WorkflowsInitialized int     `json:"workflows_initialized"`
    MonitoringEnabled  bool      `json:"monitoring_enabled"`
    ActivatedAt        time.Time `json:"activated_at"`
}
```

### 6.3 AgencyTagService

```go
package agency

type TagService interface {
    // Create tag from current agency state
    CreateTag(ctx context.Context, agencyID string, req *CreateTagRequest) (*AgencyTag, error)
    
    // List all tags for agency
    ListTags(ctx context.Context, agencyID string, filters *TagFilters) ([]*AgencyTag, error)
    
    // Get specific tag
    GetTag(ctx context.Context, agencyID string, tagName string) (*AgencyTag, error)
    
    // Compare two tags (generate diff)
    CompareTags(ctx context.Context, tagID1, tagID2 string) (*TagComparison, error)
    
    // Restore agency from tag (overwrite current draft)
    RestoreFromTag(ctx context.Context, agencyID string, tagName string) error
    
    // Publish tag as new agency
    PublishTag(ctx context.Context, tagID string, req *PublishTagRequest) (*AgencyPublication, error)
    
    // Delete tag
    DeleteTag(ctx context.Context, agencyID string, tagName string) error
}

type CreateTagRequest struct {
    Name        string            `json:"name" binding:"required"`
    Version     string            `json:"version,omitempty"`
    Description string            `json:"description" binding:"required"`
    Type        TagType           `json:"type" binding:"required"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type TagFilters struct {
    Type      TagType
    NameLike  string
    FromDate  *time.Time
    ToDate    *time.Time
    Limit     int
    Offset    int
}

type TagComparison struct {
    Tag1         *AgencyTag       `json:"tag1"`
    Tag2         *AgencyTag       `json:"tag2"`
    Differences  []TagDifference  `json:"differences"`
    Summary      string           `json:"summary"`
}

type TagDifference struct {
    Path    string      `json:"path"`
    Type    string      `json:"type"` // added, removed, modified
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}
```

### 6.4 AgencyActivationService

```go
package agency

type ActivationService interface {
    // Spawn agents based on publication manifest
    SpawnAgents(ctx context.Context, publicationID string) (*AgentSpawnResult, error)
    
    // Initialize workflows
    InitializeWorkflows(ctx context.Context, publicationID string) (*WorkflowInitResult, error)
    
    // Start health monitoring
    StartMonitoring(ctx context.Context, agencyID string) error
    
    // Pause agency (stop accepting work, pause agents)
    PauseAgency(ctx context.Context, agencyID string) error
    
    // Resume paused agency
    ResumeAgency(ctx context.Context, agencyID string) error
    
    // Drain agency (complete existing work, stop accepting new)
    DrainAgency(ctx context.Context, agencyID string) error
    
    // Stop agency (force stop all agents)
    StopAgency(ctx context.Context, agencyID string, force bool) error
}

type AgentSpawnResult struct {
    TotalAgents   int                `json:"total_agents"`
    SpawnedAgents []SpawnedAgentInfo `json:"spawned_agents"`
    Failures      []SpawnFailure     `json:"failures,omitempty"`
}

type SpawnedAgentInfo struct {
    AgentID       string    `json:"agent_id"`
    RoleCode      string    `json:"role_code"`
    State         string    `json:"state"`
    SpawnedAt     time.Time `json:"spawned_at"`
}

type WorkflowInitResult struct {
    TotalWorkflows       int      `json:"total_workflows"`
    InitializedWorkflows []string `json:"initialized_workflows"`
    Failures             []string `json:"failures,omitempty"`
}
```

---

## 7. UI/UX Design

### 7.1 Agency Designer Enhancements

Add publishing controls to the Agency Designer interface:

**Top Navigation Bar**:
```
┌──────────────────────────────────────────────────────────────┐
│  [← Back to Agencies]  Water Distribution Network Management  │
│                                                                │
│  State: Draft  │  Last saved: 2 mins ago                       │
│                                                                │
│  [Validate]  [Create Tag]  [Publish] [◉ More Actions ▼]       │
└──────────────────────────────────────────────────────────────┘
```

**Publish Dialog**:
```
┌───────────────────────────────────────────────────────┐
│  Publish Agency                                    [×] │
├───────────────────────────────────────────────────────┤
│                                                        │
│  Version: ┌──────────────┐                            │
│           │ v1.0.0       │ (Semantic versioning)      │
│           └──────────────┘                            │
│                                                        │
│  Description: ┌──────────────────────────────────┐    │
│               │ Initial production release      │    │
│               └──────────────────────────────────┘    │
│                                                        │
│  ☑ Create tag before publishing                       │
│  Tag name: v1.0.0-release                              │
│                                                        │
│  ☑ Auto-activate after publishing                     │
│  ☐ Send notifications to stakeholders                 │
│                                                        │
│  ┌──────────────────────────────────────────────┐    │
│  │ Pre-Publish Validation                        │    │
│  │ ✓ Introduction defined                        │    │
│  │ ✓ Goals specified (5)                         │    │
│  │ ✓ Roles defined (8)                           │    │
│  │ ✓ Work items created (12)                     │    │
│  │ ✓ Workflows defined (3)                       │    │
│  │ ✓ RACI matrix complete                        │    │
│  │ ⚠ AI policy partially configured              │    │
│  └──────────────────────────────────────────────┘    │
│                                                        │
│                      [Cancel]  [Publish]               │
└───────────────────────────────────────────────────────┘
```

**Tagging Dialog**:
```
┌───────────────────────────────────────────────────────┐
│  Create Tag                                        [×] │
├───────────────────────────────────────────────────────┤
│                                                        │
│  Tag Name: ┌────────────────────┐                     │
│            │ checkpoint-mvp-1   │                     │
│            └────────────────────┘                     │
│                                                        │
│  Type: ⦿ Checkpoint  ○ Release  ○ Snapshot  ○ Exp     │
│                                                        │
│  Version: ┌────────────────────┐ (Optional)           │
│           │ v0.9.0             │                      │
│           └────────────────────┘                      │
│                                                        │
│  Description: ┌────────────────────────────────┐      │
│               │ MVP Phase 1 completed          │      │
│               │ - All core workflows tested    │      │
│               │ - RACI assignments reviewed    │      │
│               └────────────────────────────────┘      │
│                                                        │
│  Metadata (Optional):                                  │
│  ┌──────────────────┐  ┌──────────────────┐          │
│  │ git_commit       │  │ abc123def456     │          │
│  └──────────────────┘  └──────────────────┘          │
│                                                        │
│                             [Cancel]  [Create Tag]     │
└───────────────────────────────────────────────────────┘
```

### 7.2 Agency Homepage Enhancements

Update agency cards to show lifecycle state:

```
┌─────────────────────────────────────────────────────┐
│  Water Distribution Network Management              │
│  State: [Active]                                     │
│  Published: Nov 15, 2025 (v1.2.0)                    │
│  8 agents running │ 3 workflows active               │
│                                                       │
│  [Open Designer] [Pause] [◉ Menu ▼]                  │
└─────────────────────────────────────────────────────┘

Agency Menu:
- View Publication History
- Manage Tags
- Monitor Agents
- View Metrics
- Pause Agency
- Stop Agency
- Export Configuration
```

### 7.3 Tag Management Page

New page: `/agencies/:id/tags`

```
┌──────────────────────────────────────────────────────────────┐
│  Tags - Water Distribution Network Management                │
├──────────────────────────────────────────────────────────────┤
│  [Create New Tag]                                [Search...]  │
│                                                                │
│  Release Tags (3)                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ v1.2.0 │ Nov 18, 2025 │ Production release with fixes  │  │
│  │        │ [View] [Compare] [Publish] [Delete]           │  │
│  ├────────────────────────────────────────────────────────┤  │
│  │ v1.1.0 │ Nov 10, 2025 │ Feature enhancement release    │  │
│  │        │ [View] [Compare] [Restore] [Delete]           │  │
│  ├────────────────────────────────────────────────────────┤  │
│  │ v1.0.0 │ Nov 1, 2025  │ Initial production release     │  │
│  │        │ [View] [Compare] [Delete]                     │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                                │
│  Checkpoints (2)                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ checkpoint-after-review │ Oct 28, 2025                  │  │
│  │        │ [View] [Restore]                               │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

---

## 8. Implementation Plan

### Phase 1: State Machine and Data Models (Week 1)

**Tasks**:
- [ ] Update `AgencyState` enum with new states
- [ ] Create `AgencyPublication` model
- [ ] Create `AgencyTag` model
- [ ] Add ArangoDB collections: `agency_publications`, `agency_tags`
- [ ] Implement state transition validation logic
- [ ] Add database migrations

**Files to Create/Modify**:
- `internal/agency/models/agency.go` (update State enum)
- `internal/agency/models/publication.go` (new)
- `internal/agency/models/tag.go` (new)
- `internal/agency/state_machine.go` (new)

### Phase 2: Tag Service (Week 2)

**Tasks**:
- [ ] Implement `TagService` interface
- [ ] Implement `TagRepository` (ArangoDB)
- [ ] Add tag creation endpoint
- [ ] Add tag listing endpoint
- [ ] Add tag comparison logic (diff generation)
- [ ] Add snapshot generation (deep copy of specification)
- [ ] Implement SHA generation for content hashing

**Files to Create**:
- `internal/agency/services/tag_service.go`
- `internal/agency/arangodb/tag_repository.go`
- `internal/handlers/tag_handler.go`

### Phase 3: Publication Service (Week 3)

**Tasks**:
- [ ] Implement `PublicationService` interface
- [ ] Implement validation logic (pre-publish checks)
- [ ] Create publication workflow
- [ ] Generate deployment manifests
- [ ] Store publication snapshots
- [ ] Implement publication history tracking

**Files to Create**:
- `internal/agency/services/publication_service.go`
- `internal/agency/arangodb/publication_repository.go`
- `internal/handlers/publication_handler.go`
- `internal/agency/validation/publisher_validator.go`

### Phase 4: Activation Service (Week 4)

**Tasks**:
- [ ] Implement `ActivationService` interface
- [ ] Create agent spawn orchestration
- [ ] Implement workflow initialization
- [ ] Add monitoring setup
- [ ] Implement pause/resume logic
- [ ] Add graceful drain logic

**Files to Create**:
- `internal/agency/services/activation_service.go`
- `internal/orchestration/agency_activator.go`
- `internal/handlers/activation_handler.go`

### Phase 5: UI Implementation (Week 5)

**Tasks**:
- [ ] Add publish button to Agency Designer
- [ ] Create publish dialog component
- [ ] Create tag creation dialog
- [ ] Build tag management page
- [ ] Add agency state badges to homepage
- [ ] Add publication history view
- [ ] Implement tag comparison UI

**Files to Create**:
- `internal/web/pages/agency_designer/publish_dialog.templ`
- `internal/web/pages/agency_designer/tag_dialog.templ`
- `internal/web/pages/tags/tag_management.templ`
- `static/js/agency-designer/publish.js`
- `static/js/tags.js`

### Phase 6: Integration and Testing (Week 6)

**Tasks**:
- [ ] End-to-end testing of publish workflow
- [ ] Test tag creation and restoration
- [ ] Test activation/deactivation
- [ ] Load testing with multiple agencies
- [ ] Documentation updates
- [ ] Training materials

---

## 9. API Endpoints

### 9.1 Publication Endpoints

```
POST   /api/v1/agencies/:id/validate          - Validate for publishing
POST   /api/v1/agencies/:id/publish           - Publish agency
POST   /api/v1/agencies/:id/activate          - Activate published agency
POST   /api/v1/agencies/:id/deactivate        - Deactivate agency
GET    /api/v1/agencies/:id/publications      - Get publication history
POST   /api/v1/publications/:id/activate      - Activate specific publication
```

### 9.2 Tag Endpoints

```
POST   /api/v1/agencies/:id/tags              - Create tag
GET    /api/v1/agencies/:id/tags              - List tags
GET    /api/v1/agencies/:id/tags/:tagName     - Get tag details
DELETE /api/v1/agencies/:id/tags/:tagName     - Delete tag
POST   /api/v1/agencies/:id/tags/:tagName/restore  - Restore from tag
POST   /api/v1/agencies/:id/tags/:tagName/publish  - Publish from tag
GET    /api/v1/tags/:tag1/compare/:tag2       - Compare two tags
```

### 9.3 Lifecycle Control Endpoints

```
POST   /api/v1/agencies/:id/pause             - Pause active agency
POST   /api/v1/agencies/:id/resume            - Resume paused agency
POST   /api/v1/agencies/:id/drain             - Drain agency (graceful stop)
POST   /api/v1/agencies/:id/stop              - Stop agency (force if needed)
GET    /api/v1/agencies/:id/state             - Get current state
GET    /api/v1/agencies/:id/agents            - List active agents
```

---

## 10. Security and Compliance

### 10.1 Access Control

**Permissions Required**:
- `agency:publish` - Publish agency
- `agency:activate` - Activate/deactivate agency
- `agency:tag:create` - Create tags
- `agency:tag:delete` - Delete tags (protected)
- `agency:tag:publish` - Publish from tag
- `agency:lifecycle:control` - Pause/resume/stop agency

### 10.2 Audit Trail

All lifecycle operations logged:
- Publication events
- Activation/deactivation
- Tag creation/deletion
- State transitions
- Agent spawn/stop events

**Audit Collection**: `agency_lifecycle_audit`

```json
{
  "event_id": "evt_abc123",
  "agency_id": "UC-INFRA-001",
  "event_type": "publish",
  "actor": "user@example.com",
  "timestamp": "2025-11-20T14:30:00Z",
  "details": {
    "publication_id": "pub_xyz789",
    "version": "v1.0.0",
    "auto_activated": true
  }
}
```

---

## 11. Monitoring and Observability

### 11.1 Metrics

**Publishing Metrics**:
- `agency_publications_total` (counter)
- `agency_publish_duration_seconds` (histogram)
- `agency_validation_errors_total` (counter)

**Activation Metrics**:
- `agency_activations_total` (counter)
- `agency_agents_spawned_total` (counter)
- `agency_activation_failures_total` (counter)

**Tag Metrics**:
- `agency_tags_created_total` (counter)
- `agency_tag_restores_total` (counter)

**State Metrics**:
- `agencies_by_state` (gauge) - Count per state

### 11.2 Alerts

**Critical Alerts**:
- Agency activation failure
- Agent spawn failure rate > 20%
- Publication validation always failing

**Warning Alerts**:
- Agency in draining state > 10 minutes
- No tags created in 30 days (production agency)

---

## 12. Future Enhancements

### 12.1 Canary Deployments

Publish new version alongside existing:
- Route small % of traffic to new version
- Monitor metrics
- Auto-rollback on errors
- Gradual traffic shift

### 12.2 Blue-Green Deployments

Maintain two versions:
- Blue (current production)
- Green (new version)
- Instant switchover
- Easy rollback

### 12.3 A/B Testing

Run multiple agency versions simultaneously:
- Split traffic by criteria
- Compare metrics
- Choose winning version

---

## Appendix A: Migration Strategy

### From Current Agency Status to AgencyState

**Mapping**:
```
active   → active (if agents running) OR published (if no agents)
inactive → draft OR stopped
paused   → paused
archived → archived
```

**Migration Script**:
```go
func MigrateAgencyStatuses(ctx context.Context, db driver.Database) error {
    // 1. Add 'state' field to all agencies
    // 2. Populate based on status + active_agent_count
    // 3. Deprecate 'status' field (keep for compatibility)
}
```

---

## Appendix B: Example Workflows

### Example 1: Initial Production Release

```
1. Designer creates agency in Draft state
2. Complete all sections (introduction, goals, roles, etc.)
3. Click "Validate" → validates → state: Validated
4. Click "Publish"
   - Version: v1.0.0
   - Description: "Initial production release"
   - ☑ Create tag before publishing (tag: v1.0.0-release)
   - ☑ Auto-activate after publishing
5. System publishes → creates tag → activates → state: Active
6. Agents spawn, workflows start
```

### Example 2: Create Checkpoint Before Major Changes

```
1. Agency is in Active state
2. Want to make significant changes
3. Click "Create Tag"
   - Name: checkpoint-before-refactor
   - Type: Checkpoint
   - Description: "Stable state before workflow refactor"
4. Tag created (immutable snapshot)
5. Edit agency (state remains Active)
6. If changes cause issues → Restore from tag
```

### Example 3: Versioned Release Cycle

```
1. Agency v1.0.0 running (Active)
2. Create tag "v1.0.0-stable" (snapshot of production)
3. Edit agency for new features (create copy or pause)
4. Test changes
5. Click "Publish"
   - Version: v1.1.0
   - Create tag: v1.1.0-release
6. System creates new publication
7. Activate v1.1.0 (stops v1.0.0 agents, starts v1.1.0 agents)
8. Monitor for issues
9. If problems → Publish tag "v1.0.0-stable" (rollback)
```

---

**End of Document**
