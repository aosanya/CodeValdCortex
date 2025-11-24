# Instance Management: UI Components

**Related Task**: MVP-PUB-007  
**Component**: Frontend Layer

---

## Overview

The instance management UI provides interfaces for:
1. Viewing all tags with their running instances
2. Starting new instances from tags
3. Monitoring instance status and health
4. Controlling instance lifecycle (stop/restart/delete)
5. Viewing detailed instance dashboard

---

## Page Templates

### 1. Tags List Page with Instances

**File**: `internal/web/pages/agency_designer/tags_list.templ`

**Purpose**: Display all agency tags with running instances listed underneath each tag

**Key Components**:
- Tag cards showing metadata (name, version, SHA, created date)
- Instance list under each tag (collapsible)
- "Start Instance" button per tag
- Instance status badges (running, stopping, stopped, failed)
- Quick actions (view, stop) per instance

**Template Structure**:
```templ
templ TagsList(tags []*models.AgencyTag, instances map[string][]*models.AgencyInstance) {
    // Header with "Create Tag" button
    // Grid of tag cards
    // Each card contains:
    //   - Tag metadata
    //   - List of instances from this tag
    //   - Actions: Start Instance, View Details
}

templ instancesList(tag *models.AgencyTag, instances []*models.AgencyInstance) {
    // Shows instances for a specific tag
    // Displays: name, state, agent count, uptime
    // Actions: Stop (if running), View Dashboard
}
```

**State Visualization**:
- `running`: Green tag
- `stopping`: Yellow/orange tag
- `stopped`: Gray tag
- `failed`: Red tag

---

### 2. Start Instance Dialog

**File**: `internal/web/components/start_instance_dialog.templ`

**Purpose**: Modal dialog for creating a new instance from a tag

**Fields**:
1. **Instance Name** (required, unique per agency)
   - Placeholder: "Production Instance, Test Run, etc."
   - Validation: Must be unique
   
2. **Description** (optional)
   - Multi-line textarea
   - Purpose/notes about this instance

3. **Tags** (optional)
   - Comma-separated tags using existing tag system
   - Examples: production, testing, demo

**Behavior**:
- Pre-fills instance name with: `{tag_name} - {current_date}`
- Shows informational note about immediate "running" state
- On submit: POST to `/api/v1/agencies/:id/instances`
- On success: Reloads page to show new instance

---

### 3. Instance Dashboard Page

**File**: `internal/web/pages/instance_dashboard.templ`

**Purpose**: Comprehensive monitoring view for a single instance

**Sections**:

#### A. Instance Overview Card
- State badge (running/stopping/stopped/failed)
- Health badge (healthy/degraded/unhealthy)
- Tag reference (name, link to tag details)
- Uptime (real-time calculated)
- Deployment info (deployed at, deployed by)
- Job acceptance status
- Description
- Tags

#### B. Agent References Panel
- Table showing all agent role codes from tag
- Status column (referenced, instantiated, healthy, degraded)
- Added timestamp
- Count summary

#### C. Workflow Execution Panel
- Table of active workflows
- Columns: Workflow name, Status, Started time, Progress bar
- Real-time status updates

#### D. Recent Activity Feed
- Timeline of events:
  - Instance created
  - Agent references loaded
  - Workflows started
  - State changes
  - Health status changes
- Event icons and timestamps
- Reverse chronological order

#### F. Control Actions
- Buttons in header:
  - **Stop** (if running): Triggers graceful shutdown
  - **Restart** (if stopped): Restarts instance
  - **Delete** (if stopped): Soft deletes instance
- Confirmation dialogs for destructive actions

**Template Structure**:
```templ
templ InstanceDashboard(instance *models.AgencyInstance, agents []AgentRef, workflows []*Workflow, events []*Event) {
    @components.LayoutWithAgency("Instance Dashboard", instance.Agency) {
        // Header with title and control buttons
        // Overview card (grid layout, 2 columns)
        // Agent references table
        // Workflows table with progress bars
        // Activity timeline
    }
}
```

---

## JavaScript Implementation

**File**: `static/js/agency-designer/instances.js`

### Core Functions

#### 1. Dialog Management

```javascript
/**
 * Open start instance dialog
 */
function startInstanceDialog(tagName) {
    document.getElementById('start-instance-tag-name').value = tagName;
    document.getElementById('instance-name').value = `${tagName} - ${new Date().toISOString().split('T')[0]}`;
    document.getElementById('start-instance-dialog').classList.add('is-active');
}

function closeStartInstanceDialog() {
    document.getElementById('start-instance-dialog').classList.remove('is-active');
}
```

#### 2. Instance Creation

```javascript
/**
 * Start new instance from tag
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
```

#### 3. Instance Control

```javascript
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
 * Navigate to instance dashboard
 */
function viewInstance(instanceID) {
    window.location.href = `/agencies/${currentAgencyId}/instances/${instanceID}`;
}

/**
 * Delete instance (soft delete)
 */
async function deleteInstance(instanceID) {
    if (!confirm('Permanently delete this instance? This action marks it as deleted but preserves audit trail.')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/instances/${instanceID}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showNotification('Success', 'Instance deleted', 'success');
            setTimeout(() => window.location.href = `/agencies/${currentAgencyId}/tags`, 1500);
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to delete instance');
        }
    } catch (error) {
        alert(`Failed to delete instance: ${error.message}`);
    }
}
```

#### 4. Helper Functions

```javascript
/**
 * Get CSS class for instance state
 */
function getStateClass(state) {
    const classes = {
        'running': 'is-success',
        'stopping': 'is-warning',
        'stopped': 'is-light',
        'failed': 'is-danger'
    };
    return classes[state] || 'is-light';
}

/**
 * Format uptime seconds to human-readable
 */
function formatUptime(uptimeSeconds) {
    const hours = Math.floor(uptimeSeconds / 3600);
    const minutes = Math.floor((uptimeSeconds % 3600) / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
}
```

---

## UI Routes

**Handler**: `internal/web/handlers/instance_handler.go`

```go
// Page routes
GET  /agencies/:id/tags                      // Tags list page with instances
GET  /agencies/:id/instances/:instance_id    // Instance dashboard page

// Component routes (HTMX partials)
GET  /agencies/:id/instances/:instance_id/overview     // Overview card partial
GET  /agencies/:id/instances/:instance_id/agents       // Agents table partial
GET  /agencies/:id/instances/:instance_id/workflows    // Workflows partial
GET  /agencies/:id/instances/:instance_id/events       // Activity feed partial
```

---

## Styling Notes

- Uses Bulma CSS framework classes throughout
- Key classes:
  - `card`, `card-header`, `card-content`, `card-footer` for tag/instance cards
  - `modal`, `modal-card` for dialogs
  - `tag` for status badges (with state-specific colors)
  - `level` for horizontal layouts
  - `table is-fullwidth is-striped` for data tables
  - `progress` for workflow progress bars
  - `timeline` for activity feed

---

## Real-time Updates (Future Enhancement)

For real-time dashboard updates without page refresh:

```javascript
// WebSocket connection for live updates
const ws = new WebSocket(`wss://${window.location.host}/ws/instances/${instanceID}`);

ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    updateDashboard(update);
};

function updateDashboard(update) {
    // Update specific sections based on update type
    switch (update.type) {
        case 'state_change':
            updateStateB adge(update.state);
            break;
        case 'agent_update':
            refreshAgentsPanel();
            break;
        case 'workflow_progress':
            updateWorkflowProgress(update.workflow_id, update.progress);
            break;
    }
}
```

---

## Related Files

- **Data Models**: [instance-data-models.md](instance-data-models.md)
- **Service Layer**: [instance-services.md](instance-services.md)
- **API Endpoints**: [instance-api.md](instance-api.md)
- **Task Overview**: [instance-management.md](instance-management.md)
