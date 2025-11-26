/**
 * Instance Management JavaScript
 * Handles tab switching, filtering, dialogs, and instance operations
 */

// Global state
let autoRefreshEnabled = false;
let refreshTimers = {};

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
    if (filterControls) {
        if (tabName === 'all-instances') {
            filterControls.style.display = 'block';
        } else {
            filterControls.style.display = 'none';
        }
    }
}

/**
 * Apply filters to instances table (all instances tab)
 */
function applyFilters() {
    const stateFilter = document.getElementById('filter-state').value.toLowerCase();
    const tagFilter = document.getElementById('filter-tag').value;
    const searchQuery = document.getElementById('filter-search').value.toLowerCase();
    const sortOrder = document.getElementById('filter-sort').value;

    const table = document.getElementById('instances-table');
    if (!table) {
        return;
    }

    const rows = Array.from(table.querySelectorAll('tbody tr'));

    let visibleCount = 0;

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
        if (show) visibleCount++;
    });

    // Update subtitle to show filtered count
    updateInstanceCountSubtitle(visibleCount, rows.length);

    // Sort visible rows
    const visibleRows = rows.filter(r => r.style.display !== 'none');
    sortRows(visibleRows, sortOrder);
}

/**
 * Update the instance count subtitle
 */
function updateInstanceCountSubtitle(visibleCount, totalCount) {
    const subtitle = document.querySelector('.subtitle');
    if (subtitle) {
        if (visibleCount < totalCount) {
            subtitle.textContent = `${visibleCount} of ${totalCount} instances`;
        } else {
            subtitle.textContent = `${totalCount} instances`;
        }
    }
}

/**
 * Sort table rows
 */
function sortRows(rows, sortOrder) {
    if (!rows || rows.length === 0) return;

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

/**
 * Open start instance dialog (from tag card)
 */
function startInstanceFromTag(tagName) {
    const select = document.getElementById('instance-tag-select');
    if (select) {
        select.value = tagName;
        updateInstanceNamePlaceholder();
    }
    openStartInstanceDialog();
}

/**
 * Open start instance dialog (generic)
 */
function openStartInstanceDialog() {
    const dialog = document.getElementById('start-instance-dialog');
    if (dialog) {
        dialog.classList.add('is-active');
    }
}

/**
 * Close start instance dialog
 */
function closeStartInstanceDialog() {
    const dialog = document.getElementById('start-instance-dialog');
    if (dialog) {
        dialog.classList.remove('is-active');
    }
}

/**
 * Update instance name placeholder based on selected tag
 */
function updateInstanceNamePlaceholder() {
    const select = document.getElementById('instance-tag-select');
    const nameField = document.getElementById('instance-name');

    if (select && nameField) {
        const tagName = select.value;
        const today = new Date().toISOString().split('T')[0];

        if (tagName) {
            nameField.value = `${tagName} - ${today}`;
        }
    }
}

/**
 * Start new instance from tag
 */
async function submitStartInstance() {
    const tagName = document.getElementById('instance-tag-select').value;
    const instanceName = document.getElementById('instance-name').value.trim();
    const description = document.getElementById('instance-description').value.trim();

    if (!tagName) {
        alert('Please select a tag');
        return;
    }

    if (!instanceName) {
        alert('Instance name is required');
        return;
    }

    const btn = document.getElementById('confirm-start-instance-btn');
    if (btn) {
        btn.classList.add('is-loading');
    }

    try {
        // Get agency ID from page context
        const agencyId = window.currentAgencyId || document.body.dataset.agencyId;

        const response = await fetch(`/api/v1/agencies/${agencyId}/tags/${encodeURIComponent(tagName)}/instances`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name: instanceName,
                description: description
            })
        });

        const result = await response.json();

        if (response.ok) {
            closeStartInstanceDialog();

            // Update instance count in subtitle immediately
            const subtitle = document.querySelector('.subtitle');
            if (subtitle) {
                const currentCount = parseInt(subtitle.textContent.match(/\d+/)?.[0] || '0');
                subtitle.textContent = `${currentCount + 1} instances`;
            }

            showNotification('Success', `Instance "${instanceName}" created and running`, 'success');

            // Reload page immediately to show new instance in lists
            setTimeout(() => window.location.reload(true), 500);
        } else {
            throw new Error(result.error || 'Failed to start instance');
        }
    } catch (error) {
        alert(`Failed to start instance: ${error.message}`);
    } finally {
        if (btn) {
            btn.classList.remove('is-loading');
        }
    }
}

/**
 * Navigate to instance dashboard
 */
function viewInstance(instanceID) {
    const agencyId = window.currentAgencyId || document.body.dataset.agencyId;
    window.location.href = `/agencies/${agencyId}/instances/${instanceID}`;
}

/**
 * Open file explorer for instance
 */
function openExplorer(instanceID) {
    const agencyId = window.currentAgencyId || document.body.dataset.agencyId;
    window.location.href = `/agencies/${agencyId}/instances/${instanceID}/explorer`;
}

/**
 * Stop instance with graceful shutdown
 */
async function stopInstance(instanceID) {
    if (!confirm('Stop this instance? New jobs will be rejected and current work will complete gracefully.')) {
        return;
    }

    try {
        const agencyId = window.currentAgencyId || document.body.dataset.agencyId;

        if (!agencyId) {
            throw new Error('Agency ID not found');
        }

        const url = `/api/v1/agencies/${agencyId}/instances/${instanceID}/stop`;

        const response = await fetch(url, {
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
        const agencyId = window.currentAgencyId || document.body.dataset.agencyId;

        const response = await fetch(`/api/v1/agencies/${agencyId}/instances/${instanceID}/restart`, {
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
        const agencyId = window.currentAgencyId || document.body.dataset.agencyId;

        const response = await fetch(`/api/v1/agencies/${agencyId}/instances/${instanceID}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showNotification('Success', 'Instance deleted', 'success');
            setTimeout(() => window.location.href = `/agencies/${agencyId}/instances`, 1500);
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to delete instance');
        }
    } catch (error) {
        alert(`Failed to delete instance: ${error.message}`);
    }
}

/**
 * Show notification message
 */
function showNotification(title, message, type) {
    // Type: success, warning, danger, info
    const notification = document.createElement('div');
    notification.className = `notification is-${type}`;
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '9999';
    notification.style.minWidth = '300px';
    notification.innerHTML = `
        <button class="delete" onclick="this.parentElement.remove()"></button>
        <strong>${title}</strong><br>${message}
    `;

    document.body.appendChild(notification);

    // Auto-dismiss after 5 seconds
    setTimeout(() => notification.remove(), 5000);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function () {
    // Set default tab to "by-tag"
    switchTab('by-tag');

    // Store agency ID in window for easy access
    const agencyIdMeta = document.querySelector('meta[name="agency-id"]');
    if (agencyIdMeta) {
        window.currentAgencyId = agencyIdMeta.content;
    }

    // Check for tag filter in URL query parameter
    const urlParams = new URLSearchParams(window.location.search);
    const tagKey = urlParams.get('tag_key');

    if (tagKey) {
        // Switch to all-instances tab (better for filtered results)
        switchTab('all-instances');

        // Apply the tag filter after a short delay to ensure DOM is ready
        setTimeout(() => {
            const tagFilterSelect = document.getElementById('filter-tag');

            if (tagFilterSelect) {
                tagFilterSelect.value = tagKey;

                // Trigger filter
                applyFilters();
            }
        }, 100);
    }
});
