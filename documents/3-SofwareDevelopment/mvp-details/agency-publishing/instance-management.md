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
5. View instance dashboard with comprehensive monitoring
6. Manage instance lifecycle (graceful shutdown with job rejection)

## Architecture Concept

```
Tag (Immutable Snapshot)
    ↓
    ├─> Instance 1 (Running)
    ├─> Instance 2 (Running)
    └─> Instance 3 (Stopped)

Each instance:
- References immutable tag configuration
- Stored in agency_instances collection (agency database)
- Independent lifecycle (start/stop/restart)
- Unique instance ID with unique name per agency
- Deployment timestamp and metadata
- Agents are references to tag configurations (not physical entities)
- Instance marked as "running" immediately upon creation
- Health calculated on-demand based on agent states
```

### Key Design Decisions

1. **Instance Storage**: All instances stored in `agency_instances` collection within the agency's database
2. **Agent Model**: Agents are references to tag configurations, not physically created until workflows trigger them
3. **State Management**: Instance transitions to "running" state immediately upon creation (agent spawning is asynchronous)
4. **Health Monitoring**: Health status calculated on-demand when requested, based on current agent states
5. **Graceful Shutdown**: When stopping, instance rejects new jobs while completing current tasks (with timeout)
6. **Naming**: Instance names must be unique per agency (enforced at creation)
7. **Deletion**: Soft delete only - instances marked as deleted but preserved for audit trail
8. **Uptime**: Real-time calculation (`current_time - started_at`), not stored
9. **Tagging**: Uses existing tag system (like goals, workitems, roles), not separate Labels field

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

type InstanceState string

const (
    InstanceStateRunning  InstanceState = "running"   // Instance active (set immediately on creation)
    InstanceStateStopping InstanceState = "stopping"  // Graceful shutdown (rejects new jobs, completes current work)
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
- **Purpose**: Track which agent references belong to which instance
- **Schema**:
  ```json
  {
    "instance_id": "inst-uuid-001",
    "agent_role_code": "developer-agent-001",
    "tag_id": "tag-uuid-123",
    "state": "referenced",
    "added_at": "2025-11-24T10:00:00Z"
  }
  ```
- **Note**: Agents are references to tag configurations, not physical agent entities. Agents are created when workflows trigger them.

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

#### Implementation Flow
```go
func (s *instanceService) StartInstance(ctx, agencyID, tagName, req) (*models.AgencyInstance, error) {
    // 1. Validate instance name is unique per agency
    exists, err := s.instanceRepo.ExistsByName(ctx, agencyID, req.Name, agencyDB)
    if exists {
        return nil, errors.New("instance name must be unique per agency")
    }
    
    // 2. Retrieve tag from agency-specific database
    tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDB)
    
    // 3. Create instance record with immediate "running" state
    instance := &models.AgencyInstance{
        AgencyID:       agencyID,
        TagID:          tag.ID,
        TagName:        tag.Name,
        InstanceID:     generateInstanceID(),
        Name:           req.Name,
        State:          models.InstanceStateRunning, // Running immediately
        DeployedAt:     time.Now(),
        StartedAt:      &now,
        DeployedBy:     req.DeployedBy,
        AcceptsNewJobs: true,
    }
    
    // 4. Save instance to agency_instances collection
    err = s.instanceRepo.Create(ctx, instance, agencyID, agencyDB)
    
    // 5. Store agent references from tag snapshot (async)
    go s.storeAgentReferences(ctx, instance, tag)
    
    return instance, nil
}

func (s *instanceService) StopInstance(ctx, agencyID, instanceID) error {
    // 1. Mark instance as stopping and reject new jobs
    instance.State = models.InstanceStateStopping
    instance.AcceptsNewJobs = false
    s.instanceRepo.Update(ctx, instance, agencyDB)
    
    // 2. Signal all active agents to complete current work
    go s.gracefulShutdown(ctx, instance, 30*time.Second) // 30s timeout
    
    return nil
}

func (s *instanceService) gracefulShutdown(ctx, instance, timeout) {
    // Wait for current tasks with timeout
    select {
    case <-s.waitForCompletion(instance):
        // All tasks completed gracefully
    case <-time.After(timeout):
        // Force stop after timeout
        s.forceStop(instance)
    }
    
    // Mark as stopped
    instance.State = models.InstanceStateStopped
    now := time.Now()
    instance.StoppedAt = &now
    s.instanceRepo.Update(ctx, instance, agencyDB)
}
```
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
    ExistsByName(ctx context.Context, agencyID, name, agencyDB string) (bool, error) // Validate unique names
    List(ctx context.Context, agencyID, agencyDB string, filters *services.InstanceFilters) ([]*models.AgencyInstance, error)
    Update(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error
    SoftDelete(ctx context.Context, agencyID, instanceID, deletedBy, agencyDB string) error // Soft delete only
    
    // Agent reference tracking
    LinkAgentReference(ctx context.Context, instanceID, agentRoleCode, tagID, agencyDB string) error
    GetInstanceAgentReferences(ctx context.Context, instanceID, agencyDB string) ([]string, error)
}
```

### 5. HTTP API Endpoints

```go
// internal/handlers/instance_handler.go
POST   /api/v1/agencies/:id/instances                        // Start new instance from tag
GET    /api/v1/agencies/:id/instances                        // List instances (supports filtering)
GET    /api/v1/agencies/:id/instances/:instance_id           // Get instance dashboard
DELETE /api/v1/agencies/:id/instances/:instance_id           // Soft delete stopped instance
POST   /api/v1/agencies/:id/instances/:instance_id/stop      // Stop instance (graceful shutdown)
POST   /api/v1/agencies/:id/instances/:instance_id/restart   // Restart instance
GET    /api/v1/agencies/:id/instances/:instance_id/health    // Get health status (calculated on-demand)
GET    /api/v1/agencies/:id/instances/:instance_id/agents    // List instance agent references
POST   /api/v1/agencies/:id/instances/:instance_id/accept-job // Check if instance accepts new jobs
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
                    <label class="label">Instance Name <span class="has-text-danger">*</span></label>
                    <div class="control">
                        <input class="input" type="text" id="instance-name" 
                               placeholder="Production Instance, Test Run, etc." required />
                    </div>
                    <p class="help">Must be unique per agency</p>
                </div>
                
                <div class="field">
                    <label class="label">Description</label>
                    <div class="control">
                        <textarea class="textarea" id="instance-description" 
                                  placeholder="Purpose of this instance..."></textarea>
                    </div>
                </div>
                
                <div class="field">
                    <label class="label">Tags</label>
                    <div class="control">
                        <input class="input" type="text" id="instance-tags" 
                               placeholder="production, testing, etc. (comma-separated)" />
                    </div>
                </div>
                
                <div class="notification is-info is-light is-size-7">
                    <p><strong>Note:</strong> Instance will be marked as "running" immediately. Agent references will be loaded asynchronously from the tag snapshot.</p>
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
    const tagsInput = document.getElementById('instance-tags').value.trim();
    const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()) : [];
    
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
                tags: tags
            })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            closeStartInstanceDialog();
            showNotification('Success', `Instance "${instanceName}" created and running`, 'success');
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
 * Stop instance with graceful shutdown
 */
async function stopInstance(instanceID) {
    if (!confirm('Stop this instance? New jobs will be rejected and current work will complete gracefully.')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/instances/${instanceID}/stop`, {
            method: 'POST'
        });
        
        if (response.ok) {
            showNotification('Success', 'Instance entering graceful shutdown (30s timeout)...', 'success');
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
 * View instance dashboard
 */
function viewInstance(instanceID) {
    window.location.href = `/agencies/${currentAgencyId}/instances/${instanceID}`;
}
```

### 8. Instance Dashboard UI

#### Template: `internal/web/pages/instance_dashboard.templ`

```templ
templ InstanceDashboard(instance *models.AgencyInstance, agents []string, workflows []*models.Workflow, events []*models.Event) {
    @components.LayoutWithAgency("Instance Dashboard", instance.Agency) {
        <section class="section">
            <div class="container">
                <!-- Header with controls -->
                <div class="level mb-5">
                    <div class="level-left">
                        <div class="level-item">
                            <div>
                                <h1 class="title is-3">{ instance.Name }</h1>
                                <p class="subtitle is-6">Instance Dashboard</p>
                            </div>
                        </div>
                    </div>
                    <div class="level-right">
                        <div class="level-item">
                            <div class="buttons">
                                if instance.State == models.InstanceStateRunning {
                                    <button class="button is-warning" onclick={ stopInstance(instance.InstanceID) }>
                                        <span class="icon"><i class="fas fa-stop"></i></span>
                                        <span>Stop Instance</span>
                                    </button>
                                } else if instance.State == models.InstanceStateStopped {
                                    <button class="button is-success" onclick={ restartInstance(instance.InstanceID) }>
                                        <span class="icon"><i class="fas fa-play"></i></span>
                                        <span>Restart</span>
                                    </button>
                                    <button class="button is-danger" onclick={ deleteInstance(instance.InstanceID) }>
                                        <span class="icon"><i class="fas fa-trash"></i></span>
                                        <span>Delete</span>
                                    </button>
                                }
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- A. Instance Overview Card -->
                <div class="box mb-5">
                    <h2 class="title is-5 mb-4">Instance Overview</h2>
                    <div class="columns">
                        <div class="column">
                            <div class="content">
                                <p><strong>Status:</strong> <span class={ "tag", getStateClass(instance.State) }>{ string(instance.State) }</span></p>
                                <p><strong>Health:</strong> <span class={ "tag", getHealthClass(instance.HealthStatus) }>{ instance.HealthStatus }</span></p>
                                <p><strong>Tag:</strong> { instance.TagName }</p>
                                <p><strong>Uptime:</strong> { formatUptime(instance) }</p>
                            </div>
                        </div>
                        <div class="column">
                            <div class="content">
                                <p><strong>Deployed:</strong> { instance.DeployedAt.Format("Jan 2, 2006 15:04 MST") }</p>
                                <p><strong>Deployed By:</strong> { instance.DeployedBy }</p>
                                <p><strong>Accepts Jobs:</strong> { fmt.Sprintf("%t", instance.AcceptsNewJobs) }</p>
                                <p><strong>Description:</strong> { instance.Description }</p>
                            </div>
                        </div>
                    </div>
                    if len(instance.Tags) > 0 {
                        <div class="tags mt-3">
                            for _, tag := range instance.Tags {
                                <span class="tag is-info">{ tag }</span>
                            }
                        </div>
                    }
                </div>
                
                <!-- B. Agent References Panel -->
                <div class="box mb-5">
                    <h2 class="title is-5 mb-4">Agent References ({ strconv.Itoa(len(agents)) })</h2>
                    if len(agents) == 0 {
                        <p class="has-text-grey">No agent references loaded yet.</p>
                    } else {
                        <div class="table-container">
                            <table class="table is-fullwidth is-striped">
                                <thead>
                                    <tr>
                                        <th>Role Code</th>
                                        <th>Status</th>
                                        <th>Added At</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    for _, agentRef := range agents {
                                        <tr>
                                            <td>{ agentRef.RoleCode }</td>
                                            <td><span class="tag">{ agentRef.State }</span></td>
                                            <td>{ agentRef.AddedAt.Format("15:04:05") }</td>
                                        </tr>
                                    }
                                </tbody>
                            </table>
                        </div>
                    }
                </div>
                
                <!-- C. Workflow Execution Panel -->
                <div class="box mb-5">
                    <h2 class="title is-5 mb-4">Active Workflows ({ strconv.Itoa(len(workflows)) })</h2>
                    if len(workflows) == 0 {
                        <p class="has-text-grey">No active workflows.</p>
                    } else {
                        <div class="table-container">
                            <table class="table is-fullwidth is-striped">
                                <thead>
                                    <tr>
                                        <th>Workflow</th>
                                        <th>Status</th>
                                        <th>Started</th>
                                        <th>Progress</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    for _, wf := range workflows {
                                        <tr>
                                            <td>{ wf.Name }</td>
                                            <td><span class={ "tag", getWorkflowStateClass(wf.State) }>{ string(wf.State) }</span></td>
                                            <td>{ wf.StartedAt.Format("15:04:05") }</td>
                                            <td>
                                                <progress class="progress is-small is-primary" value={ strconv.Itoa(wf.Progress) } max="100">{ strconv.Itoa(wf.Progress) }%</progress>
                                            </td>
                                        </tr>
                                    }
                                </tbody>
                            </table>
                        </div>
                    }
                </div>
                
                <!-- D. Real-time Activity Feed -->
                <div class="box">
                    <h2 class="title is-5 mb-4">Recent Activity</h2>
                    if len(events) == 0 {
                        <p class="has-text-grey">No recent events.</p>
                    } else {
                        <div class="timeline">
                            for _, event := range events {
                                <div class="timeline-item">
                                    <div class="timeline-marker is-icon">
                                        <i class={ "fas", getEventIcon(event.Type) }></i>
                                    </div>
                                    <div class="timeline-content">
                                        <p class="heading">{ event.Timestamp.Format("15:04:05") }</p>
                                        <p>{ event.Message }</p>
                                    </div>
                                </div>
                            }
                        </div>
                    }
                </div>
            </div>
        </section>
    }
}
```

### 9. Integration with Lifecycle Manager

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
- [ ] `instance_agents` collection tracks agent references to tag configurations
- [ ] Instance model with all required fields implemented (including soft delete)
- [ ] Repository with CRUD operations functional (including ExistsByName for uniqueness)
- [ ] Indexes created for efficient querying

### Business Logic
- [ ] InstanceService with all methods implemented
- [ ] StartInstance validates unique names per agency
- [ ] Instance marked as "running" immediately upon creation
- [ ] Agent references stored asynchronously from tag snapshot
- [ ] StopInstance implements graceful shutdown:
  - Sets AcceptsNewJobs = false
  - Waits for current tasks to complete (30s timeout)
  - Force stops if timeout exceeded
  - Marks as "stopped"
- [ ] Multiple instances can run from same tag simultaneously
- [ ] Instance health calculated on-demand based on agent states
- [ ] Instance uptime computed in real-time (current_time - started_at)
- [ ] Soft delete implementation (marks is_deleted, preserves audit trail)

### API Layer
- [ ] All 9 HTTP endpoints implemented
- [ ] Instance name uniqueness validation
- [ ] Request validation working
- [ ] Error handling comprehensive
- [ ] API documentation complete

### UI Layer
- [ ] Tag list page displays all tags
- [ ] "Start Instance" button on each tag
- [ ] Start instance dialog functional with:
  - Unique name validation
  - Tag support (not Labels)
  - Clear messaging about immediate "running" state
- [ ] Running instances listed under each tag
- [ ] Instance status badges show correct state
- [ ] Stop button triggers graceful shutdown confirmation
- [ ] Instance dashboard page with:
  - A. Instance overview (metadata, state, uptime, health)
  - B. Agent references panel
  - C. Workflow execution panel
  - D. Real-time activity/events feed
  - F. Control actions (stop, restart, soft delete)

### Integration
- [ ] Lifecycle manager tracks instances
- [ ] Agent references correctly linked to instances
- [ ] Instance cleanup on stop (graceful shutdown flow)
- [ ] No memory leaks from stopped instances
- [ ] Multiple instances isolated from each other
- [ ] Job acceptance check honors AcceptsNewJobs flag

### Testing
- [ ] Unit tests for InstanceService (80%+ coverage)
- [ ] Repository tests with mock database
- [ ] Handler tests for all endpoints
- [ ] Integration test: start instance from tag
- [ ] Integration test: graceful shutdown with timeout
- [ ] Integration test: unique name validation
- [ ] Integration test: soft delete preserves audit trail
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
