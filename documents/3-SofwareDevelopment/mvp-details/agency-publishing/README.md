# Agency Publishing & Tagging Domain

**Priority**: P0 (Critical)  
**Total Tasks**: 7 (MVP-PUB-001 through MVP-PUB-007)  
**Estimated Timeline**: 7 weeks  

## Overview

The Agency Publishing & Tagging system transforms agencies from design artifacts into production-ready deployments. It provides:

1. **Publishing**: Validation, versioning, and activation of agency designs into running systems
2. **Tagging**: Immutable snapshots for version control, rollback, and experimentation
3. **Lifecycle Management**: State machine controlling agency progression through 8 states
4. **Activation**: Orchestrated agent spawning and workflow initialization
5. **Instance Management**: Multi-instance deployment from tags with isolated runtime state

**Research Status**: Complete research session conducted with 11 architectural questions answered. See [instance-research-session.md](instance-research-session.md) for full Q&A documentation.

**Documentation Status**:
- ✅ All specification files updated with research insights
- ✅ Complete UI component templates designed (hybrid views, 5-panel dashboard)
- ✅ Data models enhanced with instance isolation patterns
- ✅ Service layer patterns established
- 🔄 Implementation pending (Go handlers, Templ files, JavaScript)

## Conceptual Model

### Publishing vs. Tagging vs. Instances

**Publishing** is the go-live process:
- Validates agency completeness (introduction, goals, roles, workflows, RACI)
- Creates immutable publication record with deployment manifest
- Optionally activates (spawns agents, starts workflows)
- Enables production monitoring and control

**Tagging** is snapshot creation:
- Creates point-in-time copy of entire agency configuration
- Immutable once created (cannot be modified)
- Supports semantic versioning (v1.0.0, v1.1.0)
- Enables rollback and experimentation

**Instances** are runtime deployments:
- Multiple instances can be started from a single tag
- Each instance has isolated runtime state (separate agent pools, workflows)
- Independent lifecycle (start/stop/restart) per instance
- Enables testing, demos, production, blue/green deployments from same tag

**Key Distinctions**: 
- Tags are static snapshots (version control)
- Publications are validated deployment plans (one-time activation)
- Instances are running agencies (multi-instance execution from tags)

## Architecture

### State Machine (8 States)

```
Draft → Validated → Published → Active → Paused/Draining → Stopped → Archived
                                  ↓
                              (can pause/resume/drain)
```

**State Descriptions**:
- **Draft**: Design in progress, can be edited freely
- **Validated**: Passed validation checks, ready to publish
- **Published**: Immutable publication created, not yet running
- **Active**: Agents spawned and running, accepting work
- **Paused**: Temporarily suspended, agents stopped
- **Draining**: Completing existing work, no new tasks accepted
- **Stopped**: All agents terminated, configuration preserved
- **Archived**: Historical record, read-only

### Service Architecture

```
┌────────────────────────────────────┐
│   Agency Publishing Services       │
├────────────────────────────────────┤
│  PublicationService                │
│  - ValidateForPublish()            │
│  - Publish()                       │
│  - GetPublicationHistory()         │
├────────────────────────────────────┤
│  TagService                        │
│  - CreateTag()                     │
│  - ListTags()                      │
│  - CompareTags()                   │
│  - RestoreFromTag()                │
├────────────────────────────────────┤
│  ActivationService                 │
│  - SpawnAgents()                   │
│  - InitializeWorkflows()           │
│  - PauseAgency()                   │
│  - DrainAgency()                   │
├────────────────────────────────────┤
│  InstanceService                   │
│  - StartInstance()                 │
│  - StopInstance()                  │
│  - ListInstances()                 │
│  - RestartInstance()               │
│  - GetInstanceHealth()             │
└────────────────────────────────────┘
│  - StopAgency()                    │
└────────────────────────────────────┘
```

### Data Models

**AgencyState Enum** (replaces simple Status):
```go
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
```

**AgencyPublication** (ArangoDB collection):
```go
type AgencyPublication struct {
    ID              string              // Publication ID
    AgencyID        string              // Source agency
    Version         string              // Semantic version (v1.0.0)
    TagID           string              // Optional tag reference
    Snapshot        AgencySnapshot      // Complete config snapshot
    Manifest        DeploymentManifest  // Agent spawn plan
    PublishedAt     time.Time
    PublishedBy     string
    ActivatedAt     *time.Time
    DeactivatedAt   *time.Time
}
```

**AgencyTag** (ArangoDB collection):
```go
type AgencyTag struct {
    ID          string         // Tag ID
    AgencyID    string         // Source agency
    Name        string         // Tag name (unique per agency)
    Version     string         // Semantic version (optional)
    Type        TagType        // release/snapshot/experimental/checkpoint
    SHA         string         // Content hash (git-style)
    Snapshot    AgencySnapshot // Complete config snapshot
    Description string
    Metadata    TagMetadata    // Custom fields
    CreatedAt   time.Time
    CreatedBy   string
}
```

## Task Index

| Task | Title | Status | Effort | Description |
|------|-------|--------|--------|-------------|
| [MVP-PUB-001](#mvp-pub-001-state-machine--data-models) | State Machine & Data Models | ✅ Complete | Medium | Update AgencyState enum, create Publication/Tag models, ArangoDB collections, migrations ([models](./state-models.md), [transitions](./state-transitions.md), [database](./state-database.md)) |
| [MVP-PUB-002](tag-service.md) | Tag Service Implementation | ✅ Complete | Medium | Tag CRUD, snapshot generation, SHA hashing, diff logic |
| [MVP-PUB-003](publication-service.md) | Publication Service | ✅ Complete | High | Validation, publication workflow, manifest generation |
| [MVP-PUB-004](activation-service.md) | Activation Service | ✅ Complete | High | Agent spawning, workflow init, pause/resume/drain |
| [MVP-PUB-005](ui-implementation.md) | Publishing UI | ✅ Complete | Medium | Publish/tag dialogs, tag management page, state badges |
| [MVP-PUB-006](integration-testing.md) | Integration & Testing | ✅ Complete | Medium | E2E testing, load testing, documentation |
| [MVP-PUB-007](instance-management.md) | Instance Management | 📋 Not Started | High | Multi-instance deployment from tags with isolated runtime state |

## Dependencies

**External Dependencies**:
- MVP-044 ✅ (Agency Designer - provides base UI)
- MVP-032 (Agent Factory - required for agent spawning in MVP-PUB-004)

**Internal Dependencies** (within domain):
- MVP-PUB-002 depends on MVP-PUB-001 ✅
- MVP-PUB-003 depends on MVP-PUB-002 ✅
- MVP-PUB-004 depends on MVP-PUB-003 and MVP-032 ✅
- MVP-PUB-005 depends on MVP-PUB-004 ✅
- MVP-PUB-006 depends on MVP-PUB-005 ✅
- MVP-PUB-007 depends on MVP-PUB-006 ✅

## Implementation Strategy

### Phase 1: Foundation (Weeks 1-2) ✅
Tasks MVP-PUB-001 and MVP-PUB-002
- Establish data models and state machine
- Build tag service (snapshots, versioning)
- No UI changes yet

### Phase 2: Publishing (Week 3) ✅
Task MVP-PUB-003
- Implement publication validation and workflow
- Generate deployment manifests
- Publication history tracking

### Phase 3: Activation (Week 4) ✅
Task MVP-PUB-004
- Orchestrate agent spawning from publication manifest
- Initialize workflows based on agency design
- Implement lifecycle controls (pause/resume/drain/stop)

### Phase 4: User Experience (Week 5) ✅
Task MVP-PUB-005
- Agency Designer publishing controls
- Tag management interface
- Publication history viewer

### Phase 5: Validation (Week 6) ✅
Task MVP-PUB-006
- End-to-end workflow testing
- Performance and load testing
- Documentation and training materials

### Phase 6: Instance Management (Week 7) 🔄
Task MVP-PUB-007
- Multi-instance deployment from tags
- Instance lifecycle management (start/stop/restart)
- Instance health monitoring and resource tracking
- Tag list UI with instance controls

## API Endpoints

### Publication Endpoints
```
POST   /api/v1/agencies/:id/validate          # Pre-publish validation
POST   /api/v1/agencies/:id/publish           # Publish agency
POST   /api/v1/agencies/:id/activate          # Activate published agency
POST   /api/v1/agencies/:id/deactivate        # Deactivate agency
GET    /api/v1/agencies/:id/publications      # Publication history
```

### Tag Endpoints
```
POST   /api/v1/agencies/:id/tags              # Create tag
GET    /api/v1/agencies/:id/tags              # List tags
GET    /api/v1/agencies/:id/tags/:name        # Get tag details
DELETE /api/v1/agencies/:id/tags/:name        # Delete tag
POST   /api/v1/agencies/:id/tags/:name/restore    # Restore from tag
POST   /api/v1/agencies/:id/tags/:name/publish    # Publish from tag
GET    /api/v1/tags/:tag1/compare/:tag2       # Compare tags
```

### Instance Management Endpoints
```
POST   /api/v1/agencies/:id/tags/:name/instances     # Start instance from tag
GET    /api/v1/agencies/:id/instances                # List all instances
GET    /api/v1/agencies/:id/instances/:instanceId    # Get instance details
DELETE /api/v1/agencies/:id/instances/:instanceId    # Stop and delete instance
POST   /api/v1/agencies/:id/instances/:instanceId/restart  # Restart instance
GET    /api/v1/agencies/:id/instances/:instanceId/health   # Instance health
GET    /api/v1/agencies/:id/instances/:instanceId/agents   # Instance agents
GET    /api/v1/agencies/:id/tags/:name/instances     # Instances for specific tag
```

### Lifecycle Control
```
POST   /api/v1/agencies/:id/pause             # Pause active agency
POST   /api/v1/agencies/:id/resume            # Resume paused agency
POST   /api/v1/agencies/:id/drain             # Graceful drain
POST   /api/v1/agencies/:id/stop              # Force stop
GET    /api/v1/agencies/:id/state             # Current state
GET    /api/v1/agencies/:id/agents            # Active agents
```

## Example Workflows

### Workflow 1: Initial Production Release
```
1. Designer creates agency (Draft)
2. Complete all sections → Validate → (Validated)
3. Click "Publish"
   - Version: v1.0.0
   - Create tag: v1.0.0-release
   - Auto-activate: Yes
4. System: publishes → creates tag → spawns agents → (Active)
```

### Workflow 2: Create Checkpoint Before Changes
```
1. Agency running (Active)
2. Create tag "checkpoint-before-refactor" (Checkpoint)
3. Make changes to agency
4. If issues → Restore from tag
```

### Workflow 3: Version Release Cycle
```
1. Agency v1.0.0 (Active)
2. Create tag "v1.0.0-stable"
3. Test new features
4. Publish v1.1.0 → Activate
5. If problems → Publish tag "v1.0.0-stable" (rollback)
```

## Security & Compliance

**Access Control**:
- `agency:publish` - Publish agency
- `agency:activate` - Activate/deactivate
- `agency:tag:create` - Create tags
- `agency:tag:delete` - Delete tags
- `agency:lifecycle:control` - Pause/resume/stop

**Audit Trail** (collection: `agency_lifecycle_audit`):
- All state transitions logged
- Publication events
- Tag creation/deletion
- Agent spawn/stop events

## Monitoring

**Metrics**:
- `agency_publications_total` (counter)
- `agency_activations_total` (counter)
- `agency_agents_spawned_total` (counter)
- `agencies_by_state` (gauge)
- `agency_tags_created_total` (counter)

**Alerts**:
- Agency activation failure (critical)
- Agent spawn failure rate > 20% (critical)
- Agency in draining state > 10 minutes (warning)

## Reference Documentation

**Complete Architecture**: [`/documents/2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md`](../../../2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md)

**Related Systems**:
- Agency Designer (MVP-044) - Base UI for agency configuration
- Agent Factory (MVP-032) - Agent instantiation mechanism
- Workflow System (MVP-030) - Workflow definition and execution
