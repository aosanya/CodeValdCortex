# MVP-PUB-007: Agency Instance Management System

**Task ID**: MVP-PUB-007  
**Priority**: P0 (Critical)  
**Effort**: High  
**Skills**: Go, ArangoDB, Templ, Frontend Dev  
**Dependencies**: MVP-PUB-006 (Agency-Specific Tag Storage) ✅

---

## Overview

Implement a comprehensive instance management system that enables running multiple independent instances of an agency from any tag snapshot. Each instance operates with isolated runtime state while maintaining an immutable reference to its source tag configuration.

## Objective

Enable users to:
1. Browse all tags for an agency in a dedicated UI
2. Start multiple independent instances from any tag
3. Monitor running instances (status, health, uptime)
4. Stop/restart instances independently
5. View instance history and audit trail

## Architecture Concept

```
Tag (Immutable Snapshot)
    ↓
    ├─> Instance 1 (Running)
    ├─> Instance 2 (Running)
    └─> Instance 3 (Stopped)

Each instance:
- References immutable tag configuration
- Has isolated runtime state (agents, tasks, workflows)
- Independent lifecycle (start/stop/restart)
- Unique instance ID
- Deployment timestamp and metadata
```

## Key Deliverables

### 1. Data Models

#### AgencyInstance Model
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
    Name        string `json:"name"`          // User-friendly name (e.g., "Production Instance", "Test Run #3")
    Description string `json:"description"`
    
    // Runtime state
    State       InstanceState `json:"state"`  // pending, starting, running, stopping, stopped, failed
    HealthStatus string       `json:"health_status"` // healthy, degraded, unhealthy
    
    // Deployment info
    DeployedAt  time.Time  `json:"deployed_at"`
    DeployedBy  string     `json:"deployed_by"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    StoppedAt   *time.Time `json:"stopped_at,omitempty"`
    
    // Runtime tracking
    AgentCount       int       `json:"agent_count"`        // Active agents
    WorkflowCount    int       `json:"workflow_count"`     // Active workflows
    LastHeartbeat    time.Time `json:"last_heartbeat"`
    UptimeSeconds    int64     `json:"uptime_seconds"`
    
    // Resource allocation (from tag snapshot)
    ResourceLimits ResourceAllocation `json:"resource_limits"`
    
    // Metadata
    Labels   map[string]string      `json:"labels,omitempty"`   // user-defined labels
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type InstanceState string

const (
    InstanceStatePending  InstanceState = "pending"   // Deployment queued
    InstanceStateStarting InstanceState = "starting"  // Agents spawning
    InstanceStateRunning  InstanceState = "running"   // Fully operational
    InstanceStateStopping InstanceState = "stopping"  // Graceful shutdown
    InstanceStateStopped  InstanceState = "stopped"   // All agents stopped
    InstanceStateFailed   InstanceState = "failed"    // Failed to start/crashed
)
```

### 2. Database Schema

#### Collection: `agency_instances`
- **Location**: Each agency's own database (e.g., `{agency_uuid}/agency_instances`)
- **Indexes**:
  - `tag_id` (for querying instances by tag)
  - `state` (for filtering by state)
  - `deployed_at` (for chronological sorting)
  - Unique: `instance_id`

#### Collection: `instance_agents`
- **Purpose**: Track which agents belong to which instance
- **Schema**:
  ```json
  {
    "instance_id": "inst-uuid-001",
    "agent_id": "agent-uuid-123",
    "role_code": "developer-agent-001",
    "state": "running",
    "spawned_at": "2025-11-24T10:00:00Z"
  }
  ```

### 3. Instance Service

#### Interface
```go
// internal/agency/services/instance_service.go
type InstanceService interface {
    // StartInstance creates and starts a new instance from a tag
    StartInstance(ctx context.Context, agencyID, tagName string, req *StartInstanceRequest) (*models.AgencyInstance, error)
    
    // StopInstance gracefully stops a running instance
    StopInstance(ctx context.Context, agencyID, instanceID string) error
    
    // RestartInstance stops and restarts an instance
    RestartInstance(ctx context.Context, agencyID, instanceID string) error
    
    // GetInstance retrieves instance details
    GetInstance(ctx context.Context, agencyID, instanceID string) (*models.AgencyInstance, error)
    
    // ListInstances lists instances with filtering
    ListInstances(ctx context.Context, agencyID string, filters *InstanceFilters) ([]*models.AgencyInstance, error)
    
    // GetInstanceHealth retrieves health status
    GetInstanceHealth(ctx context.Context, agencyID, instanceID string) (*InstanceHealth, error)
    
    // DeleteInstance permanently removes a stopped instance
    DeleteInstance(ctx context.Context, agencyID, instanceID string) error
}

type StartInstanceRequest struct {
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Labels      map[string]string `json:"labels,omitempty"`
    AutoStart   bool              `json:"auto_start"` // Start immediately vs deploy-only
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

#### Implementation Flow
```go
func (s *instanceService) StartInstance(ctx, agencyID, tagName, req) (*models.AgencyInstance, error) {
    // 1. Retrieve tag from agency-specific database
    tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDB)
    
    // 2. Create instance record
    instance := &models.AgencyInstance{
        AgencyID:   agencyID,
        TagID:      tag.ID,
        TagName:    tag.Name,
        InstanceID: generateInstanceID(),
        Name:       req.Name,
        State:      models.InstanceStatePending,
        DeployedAt: time.Now(),
        DeployedBy: req.DeployedBy,
    }
    
    // 3. Save instance to agency_instances collection
    err = s.instanceRepo.Create(ctx, instance, agencyID, agencyDB)
    
    if req.AutoStart {
        // 4. Spawn agents from tag snapshot
        go s.spawnAgentsForInstance(ctx, instance, tag)
        
        // 5. Initialize workflows
        go s.initializeWorkflowsForInstance(ctx, instance, tag)
    }
    
    return instance, nil
}
```

### 4. Instance Repository

```go
// internal/agency/arangodb/instance_repository.go
type InstanceRepository interface {
    Create(ctx context.Context, instance *models.AgencyInstance, agencyID, agencyDB string) error
    GetByID(ctx context.Context, agencyID, instanceID, agencyDB string) (*models.AgencyInstance, error)
    GetByInstanceID(ctx context.Context, instanceID, agencyDB string) (*models.AgencyInstance, error)
    List(ctx context.Context, agencyID, agencyDB string, filters *services.InstanceFilters) ([]*models.AgencyInstance, error)
    Update(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error
    Delete(ctx context.Context, agencyID, instanceID, agencyDB string) error
    
    // Agent tracking
    LinkAgent(ctx context.Context, instanceID, agentID, roleCode, agencyDB string) error
    GetInstanceAgents(ctx context.Context, instanceID, agencyDB string) ([]string, error)
}
```

### 5. HTTP API Endpoints

```go
// internal/handlers/instance_handler.go
POST   /api/v1/agencies/:id/instances                   // Start new instance from tag
GET    /api/v1/agencies/:id/instances                   // List instances
GET    /api/v1/agencies/:id/instances/:instance_id      // Get instance details
DELETE /api/v1/agencies/:id/instances/:instance_id      // Delete stopped instance
POST   /api/v1/agencies/:id/instances/:instance_id/stop // Stop instance
POST   /api/v1/agencies/:id/instances/:instance_id/restart // Restart instance
GET    /api/v1/agencies/:id/instances/:instance_id/health // Get health status
GET    /api/v1/agencies/:id/instances/:instance_id/agents // List instance agents
```

### 6. Tag List UI with Instance Controls

#### Template: `internal/web/pages/agency_designer/tags_list.templ`

```templ
templ TagsList(tags []*models.AgencyTag, instances map[string][]*models.AgencyInstance) {
    <div class="tags-list-container">
        <div class="level">
            <div class="level-left">
                <div class="level-item">
                    <h2 class="title is-4">Agency Tags & Instances</h2>
                </div>
            </div>
            <div class="level-right">
                <div class="level-item">
                    <button class="button is-primary" onclick="openTagDialog()">
                        <span class="icon"><i class="fas fa-tag"></i></span>
                        <span>Create Tag</span>
                    </button>
                </div>
            </div>
        </div>
        
        if len(tags) == 0 {
            <div class="notification is-info is-light">
                <p>No tags created yet. Create a tag to save a snapshot of your agency.</p>
            </div>
        } else {
            <div class="tags-grid">
                for _, tag := range tags {
                    <div class="card tag-card mb-4">
                        <header class="card-header">
                            <p class="card-header-title">
                                <span class="icon mr-2">
                                    <i class="fas fa-tag"></i>
                                </span>
                                { tag.Name }
                                <span class="tag is-info ml-2">{ string(tag.Type) }</span>
                            </p>
                        </header>
                        <div class="card-content">
                            <div class="content">
                                <p class="is-size-7 has-text-grey">
                                    Created { tag.CreatedAt.Format("Jan 2, 2006 15:04") } by { tag.CreatedBy }
                                </p>
                                <p>{ tag.Description }</p>
                                if tag.Version != "" {
                                    <p class="is-size-7"><strong>Version:</strong> { tag.Version }</p>
                                }
                                <p class="is-size-7"><strong>SHA:</strong> <code>{ tag.SHA[:8] }</code></p>
                            </div>
                            
                            <!-- Instance List -->
                            @instancesList(tag, instances[tag.Name])
                        </div>
                        <footer class="card-footer">
                            <a class="card-footer-item" onclick={ startInstanceDialog(tag.Name) }>
                                <span class="icon"><i class="fas fa-play"></i></span>
                                <span>Start Instance</span>
                            </a>
                            <a class="card-footer-item" onclick={ viewTagDetails(tag.Name) }>
                                <span class="icon"><i class="fas fa-info-circle"></i></span>
                                <span>Details</span>
                            </a>
                            <a class="card-footer-item" onclick={ restoreFromTag(tag.Name) }>
                                <span class="icon"><i class="fas fa-undo"></i></span>
                                <span>Restore</span>
                            </a>
                        </footer>
                    </div>
                }
            </div>
        }
    </div>
}

templ instancesList(tag *models.AgencyTag, instances []*models.AgencyInstance) {
    if len(instances) > 0 {
        <div class="instances-section mt-3">
            <p class="is-size-7 has-text-weight-bold mb-2">Running Instances ({ strconv.Itoa(len(instances)) })</p>
            <div class="instances-list">
                for _, inst := range instances {
                    <div class="box is-size-7 p-3 mb-2">
                        <div class="level is-mobile">
                            <div class="level-left">
                                <div class="level-item">
                                    <span class={ "tag", getStateClass(inst.State) }>
                                        { string(inst.State) }
                                    </span>
                                </div>
                                <div class="level-item">
                                    <strong>{ inst.Name }</strong>
                                </div>
                            </div>
                            <div class="level-right">
                                <div class="level-item">
                                    <div class="buttons are-small">
                                        if inst.State == models.InstanceStateRunning {
                                            <button class="button is-warning" onclick={ stopInstance(inst.InstanceID) }>
                                                <span class="icon is-small"><i class="fas fa-stop"></i></span>
                                            </button>
                                        }
                                        <button class="button is-info" onclick={ viewInstance(inst.InstanceID) }>
                                            <span class="icon is-small"><i class="fas fa-eye"></i></span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <p class="is-size-7 has-text-grey mt-1">
                            Agents: { strconv.Itoa(inst.AgentCount) } | 
                            Uptime: { formatDuration(inst.UptimeSeconds) }
                        </p>
                    </div>
                }
            </div>
        </div>
    }
}
```

#### Start Instance Dialog: `start_instance_dialog.templ`

```templ
templ StartInstanceDialog() {
    <div class="modal" id="start-instance-dialog">
        <div class="modal-background" onclick="closeStartInstanceDialog()"></div>
        <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">Start New Instance</p>
                <button class="delete" onclick="closeStartInstanceDialog()"></button>
            </header>
            <section class="modal-card-body">
                <input type="hidden" id="start-instance-tag-name" />
                
                <div class="field">
                    <label class="label">Instance Name</label>
                    <div class="control">
                        <input class="input" type="text" id="instance-name" 
                               placeholder="Production Instance, Test Run, etc." required />
                    </div>
                </div>
                
                <div class="field">
                    <label class="label">Description</label>
                    <div class="control">
                        <textarea class="textarea" id="instance-description" 
                                  placeholder="Purpose of this instance..."></textarea>
                    </div>
                </div>
                
                <div class="field">
                    <label class="checkbox">
                        <input type="checkbox" id="instance-auto-start" checked />
                        Start immediately (spawn agents and workflows)
                    </label>
                </div>
                
                <div class="notification is-info is-light is-size-7">
                    <p><strong>Note:</strong> This will create a new independent instance from the selected tag snapshot.</p>
                </div>
            </section>
            <footer class="modal-card-foot">
                <button class="button is-success" id="confirm-start-instance-btn" onclick="handleStartInstance()">
                    <span class="icon"><i class="fas fa-play"></i></span>
                    <span>Start Instance</span>
                </button>
                <button class="button" onclick="closeStartInstanceDialog()">Cancel</button>
            </footer>
        </div>
    </div>
}
```

### 7. JavaScript Implementation

```javascript
// static/js/agency-designer/instances.js

/**
 * Open start instance dialog for a tag
 */
function startInstanceDialog(tagName) {
    document.getElementById('start-instance-tag-name').value = tagName;
    document.getElementById('instance-name').value = `${tagName} - ${new Date().toISOString().split('T')[0]}`;
    document.getElementById('start-instance-dialog').classList.add('is-active');
}

function closeStartInstanceDialog() {
    document.getElementById('start-instance-dialog').classList.remove('is-active');
}

/**
 * Start new instance
 */
async function handleStartInstance() {
    const tagName = document.getElementById('start-instance-tag-name').value;
    const instanceName = document.getElementById('instance-name').value.trim();
    const description = document.getElementById('instance-description').value.trim();
    const autoStart = document.getElementById('instance-auto-start').checked;
    
    if (!instanceName) {
        alert('Instance name is required');
        return;
    }
    
    const btn = document.getElementById('confirm-start-instance-btn');
    btn.classList.add('is-loading');
    
    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/instances`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                tag_name: tagName,
                name: instanceName,
                description: description,
                auto_start: autoStart
            })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            closeStartInstanceDialog();
            showNotification('Success', `Instance "${instanceName}" started successfully`, 'success');
            setTimeout(() => window.location.reload(), 1500);
        } else {
            throw new Error(result.error || 'Failed to start instance');
        }
    } catch (error) {
        alert(`Failed to start instance: ${error.message}`);
    } finally {
        btn.classList.remove('is-loading');
    }
}

/**
 * Stop instance
 */
async function stopInstance(instanceID) {
    if (!confirm('Stop this instance? All agents will be gracefully stopped.')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/instances/${instanceID}/stop`, {
            method: 'POST'
        });
        
        if (response.ok) {
            showNotification('Success', 'Instance stopping...', 'success');
            setTimeout(() => window.location.reload(), 2000);
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to stop instance');
        }
    } catch (error) {
        alert(`Failed to stop instance: ${error.message}`);
    }
}

/**
 * View instance details
 */
function viewInstance(instanceID) {
    window.location.href = `/agencies/${currentAgencyId}/instances/${instanceID}`;
}
```

### 8. Integration with Lifecycle Manager

```go
// internal/lifecycle/instance_manager.go
type InstanceManager struct {
    instances map[string]*InstanceRuntime // instanceID -> runtime
    mu        sync.RWMutex
}

type InstanceRuntime struct {
    InstanceID string
    AgencyID   string
    TagID      string
    Agents     []string // Agent IDs
    State      models.InstanceState
    StartedAt  time.Time
}

func (m *InstanceManager) RegisterInstance(instanceID, agencyID, tagID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.instances[instanceID] = &InstanceRuntime{
        InstanceID: instanceID,
        AgencyID:   agencyID,
        TagID:      tagID,
        Agents:     []string{},
        State:      models.InstanceStateStarting,
        StartedAt:  time.Now(),
    }
}

func (m *InstanceManager) AddAgentToInstance(instanceID, agentID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if runtime, exists := m.instances[instanceID]; exists {
        runtime.Agents = append(runtime.Agents, agentID)
    }
}
```

## Acceptance Criteria

### Data Layer
- [ ] `agency_instances` collection created in agency-specific databases
- [ ] `instance_agents` collection tracks agent→instance relationships
- [ ] Instance model with all required fields implemented
- [ ] Repository with CRUD operations functional
- [ ] Indexes created for efficient querying

### Business Logic
- [ ] InstanceService with all methods implemented
- [ ] StartInstance spawns agents from tag snapshot
- [ ] StopInstance gracefully stops all instance agents
- [ ] Multiple instances can run from same tag simultaneously
- [ ] Instance health monitoring working
- [ ] Instance state transitions correct (pending→starting→running→stopped)

### API Layer
- [ ] All 8 HTTP endpoints implemented
- [ ] Request validation working
- [ ] Error handling comprehensive
- [ ] API documentation complete

### UI Layer
- [ ] Tag list page displays all tags
- [ ] "Start Instance" button on each tag
- [ ] Start instance dialog functional
- [ ] Running instances listed under each tag
- [ ] Instance status badges show correct state
- [ ] Stop/restart buttons work
- [ ] Real-time instance updates (optional: WebSocket)

### Integration
- [ ] Lifecycle manager tracks instances
- [ ] Agents correctly linked to instances
- [ ] Instance cleanup on stop
- [ ] No memory leaks from stopped instances
- [ ] Multiple instances isolated from each other

### Testing
- [ ] Unit tests for InstanceService (80%+ coverage)
- [ ] Repository tests with mock database
- [ ] Handler tests for all endpoints
- [ ] Integration test: start instance from tag
- [ ] Integration test: stop instance gracefully
- [ ] Load test: multiple instances from same tag

## Migration Plan

1. **Database Migration** - Create collections and indexes
2. **Service Layer** - Implement InstanceService
3. **Repository Layer** - Implement InstanceRepository
4. **Handler Layer** - Implement HTTP endpoints
5. **UI Layer** - Build tag list page and dialogs
6. **Integration** - Wire up lifecycle manager
7. **Testing** - Comprehensive test suite
8. **Documentation** - API docs and user guide

## Future Enhancements (Post-MVP)

- [ ] Instance templates (save instance configuration for reuse)
- [ ] Scheduled instance deployment (cron-style)
- [ ] Instance cloning (duplicate running instance)
- [ ] Instance snapshots (checkpoint running instance)
- [ ] Auto-scaling (spawn instances based on load)
- [ ] Blue-green deployments (switch traffic between instances)
- [ ] Instance metrics dashboard (Prometheus/Grafana)
- [ ] Instance cost tracking (resource consumption)

## Dependencies

**Requires:**
- MVP-PUB-006 ✅ (Tag system in place)
- Lifecycle Manager (already exists)
- Agent Registry (already exists)

**Enables:**
- Production deployments
- Testing/staging environments
- Load testing
- Multi-tenancy at instance level

## Related Documentation

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md`
- Tag System: Completed in MVP-PUB-002
- Activation System: Completed in MVP-PUB-004
- Publishing UI: Completed in MVP-PUB-005

---

**Estimated Timeline**: 2-3 days (16-24 hours development)

**Complexity**: High (requires coordination across data, service, API, and UI layers)
