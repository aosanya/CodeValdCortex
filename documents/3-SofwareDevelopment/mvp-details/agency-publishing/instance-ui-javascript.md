# Instance Management: JavaScript & Routing

**Related Task**: MVP-PUB-007  
**Component**: Frontend JavaScript  
**Research Reference**: See [instance-research-session.md](instance-research-session.md) for architectural Q&A

---

## Overview

This document contains JavaScript implementation details, UI routes, and styling specifications for the instance management UI.

**Related Files**:
- Template specifications: [instance-ui-templates.md](instance-ui-templates.md)
- Data models: [instance-data-models.md](instance-data-models.md)
- API endpoints: [instance-api.md](instance-api.md)

---

## JavaScript Implementation

**File**: `static/js/instances.js`

### 1. Tab Switching

```javascript
/**
 * Switch between "By Tag" and "All Instances" tabs
 */
function switchTab(tabName) {
    // Update tab active state
    document.querySelectorAll('.tabs li').forEach(tab => {
        if (tab.dataset.tab === tabName) {
            tab.classList.add('is-active');
        } else {
            tab.classList.remove('is-active');
        }
    });
    
    // Show/hide tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        if (content.id === `tab-${tabName}`) {
            content.style.display = 'block';
            content.classList.add('is-active');
        } else {
            content.style.display = 'none';
            content.classList.remove('is-active');
        }
    });
    
    // Show/hide filter controls (only for "All Instances" tab)
    const filterControls = document.getElementById('filter-controls');
    if (tabName === 'all-instances') {
        filterControls.style.display = 'block';
    } else {
        filterControls.style.display = 'none';
    }
}
```

### 2. Client-Side Filtering

```javascript
/**
 * Apply filters to instances table (all instances tab)
 */
function applyFilters() {
    const stateFilter = document.getElementById('filter-state').value.toLowerCase();
    const tagFilter = document.getElementById('filter-tag').value;
    const searchQuery = document.getElementById('filter-search').value.toLowerCase();
    const sortOrder = document.getElementById('filter-sort').value;
    
    const table = document.getElementById('instances-table');
    const rows = Array.from(table.querySelectorAll('tbody tr'));
    
    // Filter rows
    rows.forEach(row => {
        const state = row.dataset.state || '';
        const tagId = row.dataset.tagId || '';
        const tagName = row.dataset.tagName || '';
        const name = row.dataset.name || '';
        
        let show = true;
        
        // State filter
        if (stateFilter && state !== stateFilter) {
            show = false;
        }
        
        // Tag filter
        if (tagFilter && tagId !== tagFilter) {
            show = false;
        }
        
        // Search filter
        if (searchQuery && !name.toLowerCase().includes(searchQuery) && !tagName.toLowerCase().includes(searchQuery)) {
            show = false;
        }
        
        row.style.display = show ? '' : 'none';
    });
    
    // Sort visible rows
    const visibleRows = rows.filter(r => r.style.display !== 'none');
    sortRows(visibleRows, sortOrder);
}

/**
 * Sort table rows
 */
function sortRows(rows, sortOrder) {
    const [field, direction] = sortOrder.split('-');
    
    rows.sort((a, b) => {
        let aVal, bVal;
        
        if (field === 'name') {
            aVal = a.dataset.name || '';
            bVal = b.dataset.name || '';
        } else if (field === 'date') {
            aVal = a.dataset.deployedAt || '';
            bVal = b.dataset.deployedAt || '';
        }
        
        if (direction === 'asc') {
            return aVal.localeCompare(bVal);
        } else {
            return bVal.localeCompare(aVal);
        }
    });
    
    // Reorder in DOM
    const tbody = rows[0].parentNode;
    rows.forEach(row => tbody.appendChild(row));
}
```

### 3. Auto-Refresh Control

```javascript
// Global state for auto-refresh
let autoRefreshEnabled = false;
let refreshTimers = {};

/**
 * Toggle auto-refresh for dashboard panels
 */
function toggleAutoRefresh() {
    autoRefreshEnabled = !autoRefreshEnabled;
    
    const btn = document.getElementById('toggle-auto-refresh');
    const label = document.getElementById('refresh-label');
    
    if (autoRefreshEnabled) {
        btn.classList.add('is-success');
        label.textContent = 'Auto-Refresh ON';
        startPanelPolling();
    } else {
        btn.classList.remove('is-success');
        label.textContent = 'Enable Auto-Refresh';
        stopPanelPolling();
    }
}

/**
 * Start polling for all panels
 */
function startPanelPolling() {
    // HTMX will handle polling via hx-trigger="every Xs [autoRefreshEnabled]"
    // This function just sets the global flag
    htmx.process(document.body);
}

/**
 * Stop polling for all panels
 */
function stopPanelPolling() {
    // HTMX automatically stops when autoRefreshEnabled = false
    // Clear any manual timers if we add them later
    Object.values(refreshTimers).forEach(timer => clearInterval(timer));
    refreshTimers = {};
}
```

### 4. Dialog Management

```javascript
/**
 * Open start instance dialog (from tag card)
 */
function startInstanceFromTag(tagName) {
    const select = document.getElementById('instance-tag-select');
    select.value = tagName;
    updateInstanceNamePlaceholder();
    document.getElementById('start-instance-dialog').classList.add('is-active');
}

/**
 * Open start instance dialog (generic)
 */
function createInstanceDialog() {
    document.getElementById('instance-name').value = '';
    document.getElementById('instance-description').value = '';
    document.getElementById('instance-tags').value = '';
    document.getElementById('start-instance-dialog').classList.add('is-active');
}

/**
 * Close start instance dialog
 */
function closeStartInstanceDialog() {
    document.getElementById('start-instance-dialog').classList.remove('is-active');
}

/**
 * Update instance name placeholder based on selected tag
 */
function updateInstanceNamePlaceholder() {
    const select = document.getElementById('instance-tag-select');
    const tagName = select.value;
    const today = new Date().toISOString().split('T')[0];
    
    if (tagName) {
        document.getElementById('instance-name').value = `${tagName} - ${today}`;
    }
}
```

### 5. Instance Creation

```javascript
/**
 * Start new instance from tag
 */
async function handleStartInstance() {
    const tagName = document.getElementById('instance-tag-select').value;
    const instanceName = document.getElementById('instance-name').value.trim();
    const description = document.getElementById('instance-description').value.trim();
    const tagsInput = document.getElementById('instance-tags').value.trim();
    const tags = tagsInput ? tagsInput.split(',').map(t => t.trim()) : [];
    
    if (!tagName) {
        alert('Please select a tag');
        return;
    }
    
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

### 6. Instance Control

```javascript
/**
 * Navigate to instance dashboard
 */
function viewInstance(instanceID) {
    window.location.href = `/agencies/${currentAgencyId}/instances/${instanceID}`;
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
 * Restart instance
 */
async function restartInstance(instanceID) {
    if (!confirm('Restart this instance?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/instances/${instanceID}/restart`, {
            method: 'POST'
        });
        
        if (response.ok) {
            showNotification('Success', 'Instance restarting...', 'success');
            setTimeout(() => window.location.reload(), 2000);
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to restart instance');
        }
    } catch (error) {
        alert(`Failed to restart instance: ${error.message}`);
    }
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

### 7. Helper Functions

```javascript
/**
 * Get CSS class for instance state badge
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
 * Get CSS class for health status badge
 */
function getHealthClass(health) {
    const classes = {
        'healthy': 'is-success',
        'degraded': 'is-warning',
        'unhealthy': 'is-danger'
    };
    return classes[health] || 'is-light';
}

/**
 * Format uptime seconds to human-readable string
 */
function formatUptime(uptimeSeconds) {
    const hours = Math.floor(uptimeSeconds / 3600);
    const minutes = Math.floor((uptimeSeconds % 3600) / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
}

/**
 * Get CSS class for agent status badge
 */
function getAgentStatusClass(status) {
    const classes = {
        'referenced': 'is-light',
        'instantiated': 'is-info',
        'healthy': 'is-success',
        'degraded': 'is-warning'
    };
    return classes[status] || 'is-light';
}

/**
 * Get CSS class for workflow status badge
 */
function getWorkflowStatusClass(status) {
    const classes = {
        'pending': 'is-light',
        'running': 'is-info',
        'completed': 'is-success',
        'failed': 'is-danger'
    };
    return classes[status] || 'is-light';
}

/**
 * Get CSS class for event timeline marker
 */
function getEventColorClass(eventType) {
    const classes = {
        'instance_created': 'is-success',
        'state_changed': 'is-info',
        'agent_spawned': 'is-info',
        'workflow_started': 'is-primary',
        'workflow_completed': 'is-success',
        'error': 'is-danger'
    };
    return classes[eventType] || 'is-light';
}

/**
 * Get FontAwesome icon for event type
 */
function getEventIcon(eventType) {
    const icons = {
        'instance_created': 'fa-plus-circle',
        'state_changed': 'fa-exchange-alt',
        'agent_spawned': 'fa-robot',
        'workflow_started': 'fa-play',
        'workflow_completed': 'fa-check-circle',
        'error': 'fa-exclamation-triangle'
    };
    return icons[eventType] || 'fa-circle';
}
```

### 8. Notification Helper

```javascript
/**
 * Show notification message (reuse existing notification component)
 */
function showNotification(title, message, type) {
    // Assumes notification component exists in layout
    // Type: success, warning, danger, info
    const notification = document.createElement('div');
    notification.className = `notification is-${type}`;
    notification.innerHTML = `
        <button class="delete" onclick="this.parentElement.remove()"></button>
        <strong>${title}</strong><br>${message}
    `;
    
    const container = document.getElementById('notification-container') || document.body;
    container.insertBefore(notification, container.firstChild);
    
    // Auto-dismiss after 5 seconds
    setTimeout(() => notification.remove(), 5000);
}
```

---

## UI Routes

**Handler**: `internal/web/handlers/instance_handler.go`

### Page Routes

```go
// Instances list page (top-level, all agencies)
GET  /instances                      
// Handler: GetInstancesList()
// Returns: InstancesList template with all instances

// Instance dashboard page (agency-scoped)
GET  /agencies/:id/instances/:instance_id    
// Handler: GetInstanceDashboard()
// Returns: InstanceDashboard template with 5 panels
```

### HTMX Component Routes (Partial Updates)

```go
// Panel partial updates (for auto-refresh)
GET  /agencies/:id/instances/:instance_id/overview     
// Handler: GetInstanceOverviewPartial()
// Returns: InstanceOverviewPanel component only

GET  /agencies/:id/instances/:instance_id/agents       
// Handler: GetInstanceAgentsPartial()
// Returns: InstanceAgentsPanel component only

GET  /agencies/:id/instances/:instance_id/workflows    
// Handler: GetInstanceWorkflowsPartial()
// Returns: InstanceWorkflowsPanel component only

GET  /agencies/:id/instances/:instance_id/events       
// Handler: GetInstanceActivityPartial()
// Returns: InstanceActivityPanel component only
```

**Handler Pattern** (example):
```go
func (h *InstanceHandler) GetInstanceOverviewPartial(c *gin.Context) {
    instanceID := c.Param("instance_id")
    agencyID := c.Param("id")
    
    // Fetch instance data
    instance, err := h.instanceService.GetInstance(c.Request.Context(), agencyID, instanceID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
        return
    }
    
    // Calculate real-time uptime
    instance.UptimeSeconds = calculateUptime(instance.StartedAt)
    
    // Render only the panel component
    config := DefaultDashboardConfig.OverviewPanel
    component := components.InstanceOverviewPanel(instance, config)
    component.Render(c.Request.Context(), c.Writer)
}
```

---

## Styling Reference

### CSS Classes Summary

**Bulma Framework Classes**:
```css
/* Layout */
.section, .container, .columns, .column, .level, .level-left, .level-right

/* Cards */
.card, .card-header, .card-header-title, .card-content, .card-footer, .card-footer-item

/* Tables */
.table, .is-fullwidth, .is-striped, .is-hoverable, .table-container

/* Forms */
.field, .control, .label, .input, .select, .textarea, .help

/* Buttons */
.button, .is-primary, .is-success, .is-warning, .is-danger, .is-info, .is-light, .buttons, .are-small

/* Tags/Badges */
.tag, .is-success, .is-warning, .is-danger, .is-info, .is-light, .is-medium, .tags

/* Tabs */
.tabs, .is-boxed, .is-active

/* Modal */
.modal, .modal-background, .modal-card, .modal-card-head, .modal-card-body, .modal-card-foot

/* Progress */
.progress, .is-small, .is-info

/* Notifications */
.notification, .is-info, .is-success, .is-warning, .is-danger, .is-light

/* Utility */
.box, .has-text-centered, .has-text-grey-light, .mt-4, .ml-2, .mb-2, .p-3
```

**Custom CSS Required** (minimal):
```css
/* Tab content visibility */
.tab-content {
    display: none;
}

.tab-content.is-active {
    display: block;
}

/* Timeline styling (if not provided by Bulma) */
.timeline {
    position: relative;
}

.timeline-item {
    display: flex;
    padding-bottom: 1rem;
}

.timeline-marker {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 1rem;
}

.timeline-content {
    flex: 1;
}

/* State-specific timeline markers */
.timeline-marker.is-success { background-color: #48c774; color: white; }
.timeline-marker.is-info { background-color: #3298dc; color: white; }
.timeline-marker.is-warning { background-color: #ffdd57; color: #333; }
.timeline-marker.is-danger { background-color: #f14668; color: white; }
```

---

## Real-time Updates (Future Enhancement)

For real-time dashboard updates without page refresh or polling:

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
            updateStateBadge(update.state);
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

**Benefits of WebSocket approach**:
- Real-time updates (no polling delay)
- Reduced server load (push instead of poll)
- Better UX for active monitoring

**Implementation Timeline**: Post-MVP (after basic polling is working)

---

## Related Files

- **Templates**: [instance-ui-templates.md](instance-ui-templates.md)
- **Data Models**: [instance-data-models.md](instance-data-models.md)
- **Service Layer**: [instance-services.md](instance-services.md)
- **API Endpoints**: [instance-api.md](instance-api.md)
- **Research Session**: [instance-research-session.md](instance-research-session.md)
