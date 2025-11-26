# Instance Management: UI Templates (Templ Components)

**Related Task**: MVP-PUB-007  
**Component**: Frontend Templates  
**Research Reference**: See [instance-research-session.md](instance-research-session.md) for architectural Q&A

---

## Overview

This document contains all Templ template specifications for the instance management UI, including:
1. **Instances List Page** with hybrid view (by-tag + flat table tabs)
2. **Instance Dashboard** with 5 configurable panels
3. **Start Instance Dialog** modal component
4. **Panel Components** with HTMX polling support

**Related Files**:
- JavaScript implementation: [instance-ui-javascript.md](instance-ui-javascript.md)
- Data models: [instance-data-models.md](instance-data-models.md)
- API endpoints: [instance-api.md](instance-api.md)

---

## Navigation Structure

### Routes

```
GET /instances                                    # Instances list page (all instances)
    └─ Tab: "By Tag" (grouped view)
    └─ Tab: "All Instances" (flat table view)
  
GET /agencies/:id/instances/:instance_id          # Instance dashboard
    └─ 5 panels with individual HTMX refresh endpoints
```

**Data Loading**:
- **Instances List**: Full server render, all data loaded initially, client-side tab switching
- **Instance Dashboard**: Server render + optional HTMX polling (user toggle)
- **Future**: Pagination for instances list when counts grow

---

## 1. Instances List Page (Hybrid View)

**File**: `internal/web/pages/instances_list.templ`

**Purpose**: Display all instances with two viewing modes and standard filtering

**Layout Structure**:
```templ
templ InstancesList(agency *models.Agency, tags []*models.AgencyTag, instances []*models.AgencyInstance) {
    @components.LayoutWithAgency("Instances", agency) {
        <section class="section">
            <!-- Page Header -->
            <div class="level">
                <div class="level-left">
                    <h1 class="title">Instance Management</h1>
                    <p class="subtitle ml-4">{ strconv.Itoa(len(instances)) } instances</p>
                </div>
                <div class="level-right">
                    <button class="button is-primary" onclick="createInstanceDialog()">
                        <span class="icon"><i class="fas fa-plus"></i></span>
                        <span>New Instance</span>
                    </button>
                </div>
            </div>
            
            <!-- Filtering Controls (for "All Instances" tab) -->
            <div class="box" id="filter-controls">
                <div class="columns">
                    <div class="column is-3">
                        <div class="field">
                            <label class="label">State</label>
                            <div class="control">
                                <div class="select is-fullwidth">
                                    <select id="filter-state" onchange="applyFilters()">
                                        <option value="">All States</option>
                                        <option value="running">Running</option>
                                        <option value="stopping">Stopping</option>
                                        <option value="stopped">Stopped</option>
                                        <option value="failed">Failed</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="column is-3">
                        <div class="field">
                            <label class="label">Tag</label>
                            <div class="control">
                                <div class="select is-fullwidth">
                                    <select id="filter-tag" onchange="applyFilters()">
                                        <option value="">All Tags</option>
                                        for _, tag := range tags {
                                            <option value={ tag.ID }>{ tag.Name }</option>
                                        }
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="column is-4">
                        <div class="field">
                            <label class="label">Search</label>
                            <div class="control has-icons-left">
                                <input class="input" type="text" id="filter-search" 
                                       placeholder="Search by name..." 
                                       oninput="applyFilters()"/>
                                <span class="icon is-left">
                                    <i class="fas fa-search"></i>
                                </span>
                            </div>
                        </div>
                    </div>
                    <div class="column is-2">
                        <div class="field">
                            <label class="label">Sort</label>
                            <div class="control">
                                <div class="select is-fullwidth">
                                    <select id="filter-sort" onchange="applyFilters()">
                                        <option value="name-asc">Name (A-Z)</option>
                                        <option value="name-desc">Name (Z-A)</option>
                                        <option value="date-desc">Newest First</option>
                                        <option value="date-asc">Oldest First</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Tabs -->
            <div class="tabs is-boxed">
                <ul>
                    <li class="is-active" data-tab="by-tag">
                        <a onclick="switchTab('by-tag')">
                            <span class="icon"><i class="fas fa-tags"></i></span>
                            <span>By Tag</span>
                        </a>
                    </li>
                    <li data-tab="all-instances">
                        <a onclick="switchTab('all-instances')">
                            <span class="icon"><i class="fas fa-list"></i></span>
                            <span>All Instances</span>
                        </a>
                    </li>
                </ul>
            </div>
            
            <!-- Tab Content -->
            <div id="tab-by-tag" class="tab-content is-active">
                @instancesByTagView(tags, instances)
            </div>
            
            <div id="tab-all-instances" class="tab-content" style="display: none;">
                @instancesTableView(instances, tags)
            </div>
        </section>
        
        <!-- Start Instance Dialog -->
        @components.StartInstanceDialog()
    }
}
```

### Tab 1: By Tag Grouped View

```templ
templ instancesByTagView(tags []*models.AgencyTag, instances []*models.AgencyInstance) {
    <div class="columns is-multiline">
        for _, tag := range tags {
            <div class="column is-4">
                @tagCardWithInstances(tag, getInstancesForTag(instances, tag.ID))
            </div>
        }
        if len(tags) == 0 {
            <div class="column">
                <div class="notification is-warning">
                    <p>No tags available. Create a tag first to start instances.</p>
                </div>
            </div>
        }
    </div>
}

templ tagCardWithInstances(tag *models.AgencyTag, instances []*models.AgencyInstance) {
    <div class="card">
        <header class="card-header">
            <p class="card-header-title">
                <span class="icon"><i class="fas fa-tag"></i></span>
                <span>{ tag.Name }</span>
            </p>
        </header>
        <div class="card-content">
            <!-- Tag metadata -->
            <div class="content">
                <p><strong>Version:</strong> { tag.Version }</p>
                <p><strong>Type:</strong> <span class="tag">{ string(tag.Type) }</span></p>
                <p><strong>Created:</strong> { tag.CreatedAt.Format("2006-01-02 15:04") }</p>
            </div>
            
            <!-- Instances from this tag -->
            if len(instances) > 0 {
                <div class="mt-4">
                    <p class="heading">Running Instances ({ strconv.Itoa(len(instances)) })</p>
                    for _, inst := range instances {
                        @instanceListItem(inst)
                    }
                </div>
            } else {
                <p class="has-text-grey-light">No instances running</p>
            }
        </div>
        <footer class="card-footer">
            <a class="card-footer-item" onclick={ startInstanceFromTag(tag.Name) }>
                <span class="icon"><i class="fas fa-play"></i></span>
                <span>Start Instance</span>
            </a>
        </footer>
    </div>
}

templ instanceListItem(inst *models.AgencyInstance) {
    <div class="box mb-2 p-3">
        <div class="level is-mobile">
            <div class="level-left">
                <div>
                    <p class="is-size-6"><strong>{ inst.Name }</strong></p>
                    <p class="is-size-7 has-text-grey">
                        <span class={ fmt.Sprintf("tag %s", getStateClass(inst.State)) }>
                            { string(inst.State) }
                        </span>
                        <span class="ml-2">{ formatUptime(inst.UptimeSeconds) } uptime</span>
                    </p>
                </div>
            </div>
            <div class="level-right">
                <div class="buttons are-small">
                    <button class="button is-info" onclick={ viewInstance(inst.InstanceID) } title="View Dashboard">
                        <span class="icon"><i class="fas fa-eye"></i></span>
                    </button>
                    if inst.State == models.InstanceStateRunning {
                        <button class="button is-warning" onclick={ stopInstance(inst.InstanceID) } title="Stop">
                            <span class="icon"><i class="fas fa-stop"></i></span>
                        </button>
                    }
                </div>
            </div>
        </div>
    </div>
}
```

### Tab 2: All Instances Flat View

```templ
templ instancesTableView(instances []*models.AgencyInstance, tags []*models.AgencyTag) {
    <div class="table-container">
        <table class="table is-fullwidth is-striped is-hoverable" id="instances-table">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Tag</th>
                    <th>State</th>
                    <th>Health</th>
                    <th>Agents</th>
                    <th>Uptime</th>
                    <th>Deployed</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                for _, inst := range instances {
                    @instanceTableRow(inst)
                }
                if len(instances) == 0 {
                    <tr>
                        <td colspan="8" class="has-text-centered has-text-grey-light">
                            No instances found
                        </td>
                    </tr>
                }
            </tbody>
        </table>
    </div>
}

templ instanceTableRow(inst *models.AgencyInstance) {
    <tr data-instance-id={ inst.InstanceID } 
        data-state={ string(inst.State) }
        data-tag-id={ inst.TagID }
        data-tag-name={ inst.TagName }
        data-name={ inst.Name }
        data-deployed-at={ inst.DeployedAt.Format("2006-01-02") }>
        <td>
            <a href={ templ.URL(fmt.Sprintf("/agencies/%s/instances/%s", inst.AgencyID, inst.InstanceID)) }>
                <strong>{ inst.Name }</strong>
            </a>
        </td>
        <td>
            <span class="tag is-info">{ inst.TagName }</span>
        </td>
        <td>
            <span class={ fmt.Sprintf("tag %s", getStateClass(inst.State)) }>
                { string(inst.State) }
            </span>
        </td>
        <td>
            <span class={ fmt.Sprintf("tag %s", getHealthClass(inst.HealthStatus)) }>
                { inst.HealthStatus }
            </span>
        </td>
        <td>{ strconv.Itoa(inst.AgentCount) }</td>
        <td>{ formatUptime(inst.UptimeSeconds) }</td>
        <td>{ inst.DeployedAt.Format("2006-01-02 15:04") }</td>
        <td>
            <div class="buttons are-small">
                if inst.State == models.InstanceStateRunning {
                    <button class="button is-warning" onclick={ stopInstance(inst.InstanceID) } title="Stop">
                        <span class="icon"><i class="fas fa-stop"></i></span>
                    </button>
                }
                if inst.State == models.InstanceStateStopped {
                    <button class="button is-success" onclick={ restartInstance(inst.InstanceID) } title="Restart">
                        <span class="icon"><i class="fas fa-redo"></i></span>
                    </button>
                    <button class="button is-danger" onclick={ deleteInstance(inst.InstanceID) } title="Delete">
                        <span class="icon"><i class="fas fa-trash"></i></span>
                    </button>
                }
                <button class="button is-info" onclick={ viewInstance(inst.InstanceID) } title="View Dashboard">
                    <span class="icon"><i class="fas fa-eye"></i></span>
                </button>
            </div>
        </td>
    </tr>
}
```

**State Badge CSS Classes**:
- `running`: `is-success` (green)
- `stopping`: `is-warning` (yellow/orange)
- `stopped`: `is-light` (gray)
- `failed`: `is-danger` (red)

**Health Badge CSS Classes**:
- `healthy`: `is-success` (green)
- `degraded`: `is-warning` (yellow)
- `unhealthy`: `is-danger` (red)

---

## 2. Instance Dashboard Page (Configurable Panels)

**File**: `internal/web/pages/instance_dashboard.templ`

**Purpose**: Comprehensive monitoring view for a single instance with 5 independently refreshable panels

**Auto-Refresh Strategy**:
- **Default**: Static page (manual refresh only)
- **Opt-in**: "Enable Auto-Refresh" toggle button
- **Polling**: Staggered intervals per panel when enabled
- **Future**: WebSocket push updates

**Dashboard Structure**:
```templ
templ InstanceDashboard(instance *models.AgencyInstance, config DashboardConfig) {
    @components.LayoutWithAgency("Instance Dashboard", instance.Agency) {
        <section class="section">
            <!-- Header with Controls -->
            <div class="level">
                <div class="level-left">
                    <div>
                        <h1 class="title">{ instance.Name }</h1>
                        <p class="subtitle">
                            Instance Dashboard
                            <span class={ fmt.Sprintf("tag %s ml-2", getStateClass(instance.State)) }>
                                { string(instance.State) }
                            </span>
                        </p>
                    </div>
                </div>
                <div class="level-right">
                    <div class="buttons">
                        <!-- Auto-refresh toggle -->
                        <button class="button" id="toggle-auto-refresh" onclick="toggleAutoRefresh()">
                            <span class="icon"><i class="fas fa-sync"></i></span>
                            <span id="refresh-label">Enable Auto-Refresh</span>
                        </button>
                        
                        <!-- Instance controls -->
                        if instance.State == models.InstanceStateRunning {
                            <button class="button is-warning" onclick={ stopInstance(instance.InstanceID) }>
                                <span class="icon"><i class="fas fa-stop"></i></span>
                                <span>Stop</span>
                            </button>
                        }
                        if instance.State == models.InstanceStateStopped {
                            <button class="button is-success" onclick={ restartInstance(instance.InstanceID) }>
                                <span class="icon"><i class="fas fa-redo"></i></span>
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
            
            <!-- 5 Dashboard Panels (Each independently refreshable) -->
            <div class="columns is-multiline">
                <!-- Panel 1: Overview Card -->
                <div class="column is-12">
                    @components.InstanceOverviewPanel(instance, config.OverviewPanel)
                </div>
                
                <!-- Panel 2: Agent References -->
                <div class="column is-6">
                    @components.InstanceAgentsPanel(instance, config.AgentsPanel)
                </div>
                
                <!-- Panel 3: Workflow Execution -->
                <div class="column is-6">
                    @components.InstanceWorkflowsPanel(instance, config.WorkflowsPanel)
                </div>
                
                <!-- Panel 4: Recent Activity Feed -->
                <div class="column is-12">
                    @components.InstanceActivityPanel(instance, config.ActivityPanel)
                </div>
            </div>
        </section>
    }
}
```

**Dashboard Configuration Model**:
```go
type DashboardConfig struct {
    OverviewPanel  PanelConfig
    AgentsPanel    PanelConfig
    WorkflowsPanel PanelConfig
    ActivityPanel  PanelConfig
}

type PanelConfig struct {
    RefreshInterval int    // Seconds (0 = no auto-refresh)
    HTMXEndpoint    string // Partial update endpoint
    Enabled         bool   // Can be toggled off
}

// Default configuration
var DefaultDashboardConfig = DashboardConfig{
    OverviewPanel: PanelConfig{
        RefreshInterval: 30,  // 30 seconds
        HTMXEndpoint:    "/api/v1/agencies/{agencyID}/instances/{instanceID}/overview",
        Enabled:         true,
    },
    AgentsPanel: PanelConfig{
        RefreshInterval: 30,  // 30 seconds
        HTMXEndpoint:    "/api/v1/agencies/{agencyID}/instances/{instanceID}/agents",
        Enabled:         true,
    },
    WorkflowsPanel: PanelConfig{
        RefreshInterval: 20,  // 20 seconds (more frequent for progress)
        HTMXEndpoint:    "/api/v1/agencies/{agencyID}/instances/{instanceID}/workflows",
        Enabled:         true,
    },
    ActivityPanel: PanelConfig{
        RefreshInterval: 60,  // 60 seconds (less frequent)
        HTMXEndpoint:    "/api/v1/agencies/{agencyID}/instances/{instanceID}/events",
        Enabled:         true,
    },
}
```

---

## 3. Dashboard Panel Components

Each panel is a separate templ component with HTMX attributes for optional polling.

### Panel 1: Instance Overview

**File**: `internal/web/components/instance_overview_panel.templ`

```templ
templ InstanceOverviewPanel(instance *models.AgencyInstance, config PanelConfig) {
    <div class="card" 
         id="overview-panel"
         if config.RefreshInterval > 0 {
             hx-get={ config.HTMXEndpoint }
             hx-trigger="every { strconv.Itoa(config.RefreshInterval) }s [autoRefreshEnabled]"
             hx-swap="outerHTML"
         }>
        <header class="card-header">
            <p class="card-header-title">
                <span class="icon"><i class="fas fa-info-circle"></i></span>
                <span>Overview</span>
            </p>
        </header>
        <div class="card-content">
            <div class="columns">
                <div class="column">
                    <div class="field">
                        <label class="label">State</label>
                        <span class={ fmt.Sprintf("tag is-medium %s", getStateClass(instance.State)) }>
                            { string(instance.State) }
                        </span>
                    </div>
                    <div class="field">
                        <label class="label">Health</label>
                        <span class={ fmt.Sprintf("tag is-medium %s", getHealthClass(instance.HealthStatus)) }>
                            { instance.HealthStatus }
                        </span>
                    </div>
                    <div class="field">
                        <label class="label">Tag Reference</label>
                        <a href={ templ.URL(fmt.Sprintf("/agencies/%s/tags/%s", instance.AgencyID, instance.TagID)) }>
                            <span class="tag is-info is-medium">{ instance.TagName }</span>
                        </a>
                    </div>
                </div>
                <div class="column">
                    <div class="field">
                        <label class="label">Uptime</label>
                        <p id="uptime-display">{ formatUptime(instance.UptimeSeconds) }</p>
                    </div>
                    <div class="field">
                        <label class="label">Deployed At</label>
                        <p>{ instance.DeployedAt.Format("2006-01-02 15:04:05") }</p>
                    </div>
                    <div class="field">
                        <label class="label">Deployed By</label>
                        <p>{ instance.DeployedBy }</p>
                    </div>
                </div>
                <div class="column">
                    <div class="field">
                        <label class="label">Job Acceptance</label>
                        <p>
                            if instance.AcceptsNewJobs {
                                <span class="tag is-success">Accepting New Jobs</span>
                            } else {
                                <span class="tag is-warning">Rejecting New Jobs</span>
                            }
                        </p>
                    </div>
                    if instance.Description != "" {
                        <div class="field">
                            <label class="label">Description</label>
                            <p>{ instance.Description }</p>
                        </div>
                    }
                    if len(instance.Tags) > 0 {
                        <div class="field">
                            <label class="label">Tags</label>
                            <div class="tags">
                                for _, tag := range instance.Tags {
                                    <span class="tag">{ tag }</span>
                                }
                            </div>
                        </div>
                    }
                </div>
            </div>
        </div>
    </div>
}
```

### Panel 2: Agent References

**File**: `internal/web/components/instance_agents_panel.templ`

```templ
templ InstanceAgentsPanel(instance *models.AgencyInstance, agents []AgentRef, config PanelConfig) {
    <div class="card"
         id="agents-panel"
         if config.RefreshInterval > 0 {
             hx-get={ config.HTMXEndpoint }
             hx-trigger="every { strconv.Itoa(config.RefreshInterval) }s [autoRefreshEnabled]"
             hx-swap="outerHTML"
         }>
        <header class="card-header">
            <p class="card-header-title">
                <span class="icon"><i class="fas fa-robot"></i></span>
                <span>Agent References ({ strconv.Itoa(len(agents)) })</span>
            </p>
        </header>
        <div class="card-content">
            <div class="table-container">
                <table class="table is-fullwidth is-striped">
                    <thead>
                        <tr>
                            <th>Role Code</th>
                            <th>Status</th>
                            <th>Added</th>
                        </tr>
                    </thead>
                    <tbody>
                        for _, agent := range agents {
                            <tr>
                                <td><code>{ agent.RoleCode }</code></td>
                                <td>
                                    <span class={ fmt.Sprintf("tag %s", getAgentStatusClass(agent.Status)) }>
                                        { agent.Status }
                                    </span>
                                </td>
                                <td>{ agent.AddedAt.Format("2006-01-02 15:04") }</td>
                            </tr>
                        }
                        if len(agents) == 0 {
                            <tr>
                                <td colspan="3" class="has-text-centered has-text-grey-light">
                                    No agents loaded yet
                                </td>
                            </tr>
                        }
                    </tbody>
                </table>
            </div>
        </div>
    </div>
}
```

### Panel 3: Workflow Execution

**File**: `internal/web/components/instance_workflows_panel.templ`

```templ
templ InstanceWorkflowsPanel(instance *models.AgencyInstance, workflows []*Workflow, config PanelConfig) {
    <div class="card"
         id="workflows-panel"
         if config.RefreshInterval > 0 {
             hx-get={ config.HTMXEndpoint }
             hx-trigger="every { strconv.Itoa(config.RefreshInterval) }s [autoRefreshEnabled]"
             hx-swap="outerHTML"
         }>
        <header class="card-header">
            <p class="card-header-title">
                <span class="icon"><i class="fas fa-project-diagram"></i></span>
                <span>Active Workflows ({ strconv.Itoa(len(workflows)) })</span>
            </p>
        </header>
        <div class="card-content">
            <div class="table-container">
                <table class="table is-fullwidth">
                    <thead>
                        <tr>
                            <th>Workflow</th>
                            <th>Status</th>
                            <th>Progress</th>
                            <th>Started</th>
                        </tr>
                    </thead>
                    <tbody>
                        for _, wf := range workflows {
                            <tr>
                                <td>{ wf.Name }</td>
                                <td>
                                    <span class={ fmt.Sprintf("tag %s", getWorkflowStatusClass(wf.Status)) }>
                                        { wf.Status }
                                    </span>
                                </td>
                                <td>
                                    <progress class="progress is-small is-info" 
                                              value={ strconv.Itoa(wf.Progress) } 
                                              max="100">
                                        { strconv.Itoa(wf.Progress) }%
                                    </progress>
                                </td>
                                <td>{ wf.StartedAt.Format("15:04:05") }</td>
                            </tr>
                        }
                        if len(workflows) == 0 {
                            <tr>
                                <td colspan="4" class="has-text-centered has-text-grey-light">
                                    No active workflows
                                </td>
                            </tr>
                        }
                    </tbody>
                </table>
            </div>
        </div>
    </div>
}
```

### Panel 4: Recent Activity Feed

**File**: `internal/web/components/instance_activity_panel.templ`

```templ
templ InstanceActivityPanel(instance *models.AgencyInstance, events []*Event, config PanelConfig) {
    <div class="card"
         id="activity-panel"
         if config.RefreshInterval > 0 {
             hx-get={ config.HTMXEndpoint }
             hx-trigger="every { strconv.Itoa(config.RefreshInterval) }s [autoRefreshEnabled]"
             hx-swap="outerHTML"
         }>
        <header class="card-header">
            <p class="card-header-title">
                <span class="icon"><i class="fas fa-history"></i></span>
                <span>Recent Activity</span>
            </p>
        </header>
        <div class="card-content">
            <div class="timeline">
                for _, event := range events {
                    <div class="timeline-item">
                        <div class="timeline-marker { getEventColorClass(event.Type) }">
                            <i class={ fmt.Sprintf("fas %s", getEventIcon(event.Type)) }></i>
                        </div>
                        <div class="timeline-content">
                            <p class="heading">{ event.Timestamp.Format("15:04:05") }</p>
                            <p>{ event.Message }</p>
                        </div>
                    </div>
                }
                if len(events) == 0 {
                    <p class="has-text-centered has-text-grey-light">No recent activity</p>
                }
            </div>
        </div>
    </div>
}
```

---

## 4. Start Instance Dialog

**File**: `internal/web/components/start_instance_dialog.templ`

**Purpose**: Modal dialog for creating a new instance from a tag

**Fields**:
1. **Tag Selection** (required, dropdown populated from available tags)
2. **Instance Name** (required, unique per agency)
3. **Description** (optional, multi-line textarea)
4. **Tags** (optional, comma-separated)

**Template Structure**:
```templ
templ StartInstanceDialog(agency *models.Agency, tags []*models.AgencyTag) {
    <div class="modal" id="start-instance-dialog">
        <div class="modal-background" onclick="closeStartInstanceDialog()"></div>
        <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">Start New Instance</p>
                <button class="delete" onclick="closeStartInstanceDialog()"></button>
            </header>
            <section class="modal-card-body">
                <!-- Tag Selection -->
                <div class="field">
                    <label class="label">Tag <span class="has-text-danger">*</span></label>
                    <div class="control">
                        <div class="select is-fullwidth">
                            <select id="instance-tag-select" onchange="updateInstanceNamePlaceholder()">
                                <option value="">Select a tag...</option>
                                for _, tag := range tags {
                                    <option value={ tag.Name } data-version={ tag.Version }>
                                        { tag.Name } ({ tag.Version }) - { string(tag.Type) }
                                    </option>
                                }
                            </select>
                        </div>
                    </div>
                    <p class="help">Select which tag snapshot to instantiate</p>
                </div>
                
                <!-- Instance Name -->
                <div class="field">
                    <label class="label">Instance Name <span class="has-text-danger">*</span></label>
                    <div class="control">
                        <input class="input" 
                               type="text" 
                               id="instance-name" 
                               placeholder="Production Instance, Test Run, etc."
                               required/>
                    </div>
                    <p class="help">Must be unique within this agency</p>
                </div>
                
                <!-- Description -->
                <div class="field">
                    <label class="label">Description</label>
                    <div class="control">
                        <textarea class="textarea" 
                                  id="instance-description" 
                                  rows="3"
                                  placeholder="Purpose or notes about this instance..."></textarea>
                    </div>
                </div>
                
                <!-- Tags -->
                <div class="field">
                    <label class="label">Tags</label>
                    <div class="control">
                        <input class="input" 
                               type="text" 
                               id="instance-tags" 
                               placeholder="production, testing, demo"/>
                    </div>
                    <p class="help">Comma-separated tags for organization</p>
                </div>
                
                <!-- Info Message -->
                <div class="notification is-info is-light">
                    <p><strong>Note:</strong> Instance will start in "running" state immediately. Agents will spawn on-demand when jobs arrive.</p>
                </div>
            </section>
            <footer class="modal-card-foot">
                <button class="button is-primary" id="confirm-start-instance-btn" onclick="handleStartInstance()">
                    <span class="icon"><i class="fas fa-play"></i></span>
                    <span>Start Instance</span>
                </button>
                <button class="button" onclick="closeStartInstanceDialog()">Cancel</button>
            </footer>
        </div>
    </div>
}
```

---

## Styling Notes

All templates use Bulma CSS framework classes:

**Key Classes Used**:
- **Layout**: `section`, `container`, `columns`, `column`
- **Cards**: `card`, `card-header`, `card-content`, `card-footer`
- **Tables**: `table is-fullwidth is-striped is-hoverable`
- **Forms**: `field`, `control`, `label`, `input`, `select`, `textarea`
- **Buttons**: `button is-primary`, `button is-warning`, `button is-danger`
- **Tags/Badges**: `tag is-success`, `tag is-warning`, `tag is-danger`, `tag is-light`
- **Tabs**: `tabs is-boxed`, `is-active`
- **Modal**: `modal`, `modal-card`, `modal-background`
- **Progress**: `progress is-small is-info`
- **Timeline**: `timeline`, `timeline-item`, `timeline-marker`, `timeline-content`
- **Utility**: `level`, `buttons`, `notification`, `box`

**Custom CSS** (minimal):
- Tab content visibility toggling (`.tab-content { display: none; }`)
- Timeline custom styling (if Bulma doesn't provide it)

---

## Related Files

- **JavaScript Implementation**: [instance-ui-javascript.md](instance-ui-javascript.md)
- **Data Models**: [instance-data-models.md](instance-data-models.md)
- **API Endpoints**: [instance-api.md](instance-api.md)
- **Service Layer**: [instance-services.md](instance-services.md)
- **Research Session**: [instance-research-session.md](instance-research-session.md)
