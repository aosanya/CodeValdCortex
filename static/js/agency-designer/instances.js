// Instance Management JavaScript for Agency Designer
// Handles instance lifecycle operations (start, stop, restart, delete)

// Global state
let currentTagForInstance = null;
let instanceRefreshInterval = null;

// ==================== Dialog Management ====================

/**
 * Opens the start instance dialog for a specific tag
 * @param {string} tagName - The tag name to start an instance from
 */
function openStartInstanceDialog(tagName) {
    currentTagForInstance = tagName;

    // Update dialog to show the tag name
    const sourceTagEl = document.getElementById('instance-source-tag');
    if (sourceTagEl) {
        sourceTagEl.textContent = tagName;
    }

    // Reset form
    document.getElementById('instance-name').value = `${tagName}-instance-${Date.now()}`;
    document.getElementById('instance-environment').value = 'development';
    document.getElementById('instance-cpu-limit').value = '2000m';
    document.getElementById('instance-memory-limit').value = '1Gi';
    document.getElementById('instance-max-agents').value = '10';
    document.getElementById('instance-autoscale').checked = false;
    document.getElementById('instance-labels').value = '';

    // Hide autoscale settings
    const autoscaleSettings = document.getElementById('autoscale-settings');
    if (autoscaleSettings) {
        autoscaleSettings.style.display = 'none';
    }

    // Clear validation messages
    const validationEl = document.getElementById('instance-validation-messages');
    if (validationEl) {
        validationEl.innerHTML = '';
    }

    // Show dialog
    const dialog = document.getElementById('start-instance-dialog');
    if (dialog) {
        dialog.classList.add('is-active');
    }
}

/**
 * Closes the start instance dialog
 */
function closeStartInstanceDialog() {
    const dialog = document.getElementById('start-instance-dialog');
    if (dialog) {
        dialog.classList.remove('is-active');
    }
    currentTagForInstance = null;
}

/**
 * Toggles auto-scale settings visibility
 */
function toggleAutoScaleSettings() {
    const autoscaleCheckbox = document.getElementById('instance-autoscale');
    const autoscaleSettings = document.getElementById('autoscale-settings');

    if (autoscaleSettings && autoscaleCheckbox) {
        autoscaleSettings.style.display = autoscaleCheckbox.checked ? 'block' : 'none';
    }
}

// ==================== Instance Operations ====================

/**
 * Starts a new instance from the current tag
 */
async function startInstanceFromTag() {
    if (!currentTagForInstance) {
        showNotification('No tag selected', 'warning');
        return;
    }

    const agencyID = getAgencyID();
    if (!agencyID) {
        showNotification('Agency ID not found', 'danger');
        return;
    }

    // Gather form data
    const instanceName = document.getElementById('instance-name').value.trim();
    const environment = document.getElementById('instance-environment').value;
    const cpuLimit = document.getElementById('instance-cpu-limit').value.trim();
    const memoryLimit = document.getElementById('instance-memory-limit').value.trim();
    const maxAgents = parseInt(document.getElementById('instance-max-agents').value);
    const autoScaleEnabled = document.getElementById('instance-autoscale').checked;

    // Validation
    if (!instanceName) {
        showValidationError('Instance name is required');
        return;
    }

    // Build configuration
    const config = {
        cpu_limit: cpuLimit,
        memory_limit: memoryLimit,
        max_agents: maxAgents,
        auto_scale_enabled: autoScaleEnabled,
        metrics_enabled: true,
        log_level: 'info'
    };

    if (autoScaleEnabled) {
        config.min_agents = parseInt(document.getElementById('instance-min-agents').value);
        config.max_scale_agents = parseInt(document.getElementById('instance-max-scale-agents').value);
    }

    // Parse labels (optional)
    const labelsText = document.getElementById('instance-labels').value.trim();
    let labels = {};
    if (labelsText) {
        try {
            labels = JSON.parse(labelsText);
        } catch (e) {
            showValidationError('Invalid JSON format for labels');
            return;
        }
    }

    // Build request
    const request = {
        instance_name: instanceName,
        environment: environment,
        config: config,
        labels: labels,
        metadata: {}
    };

    // Disable button
    const btn = document.getElementById('start-instance-btn');
    const originalHTML = btn.innerHTML;
    btn.disabled = true;
    btn.innerHTML = '<span class="icon"><i class="fas fa-spinner fa-spin"></i></span><span>Starting...</span>';

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/tags/${currentTagForInstance}/instances`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(request)
        });

        const data = await response.json();

        if (response.ok) {
            showNotification(`Instance "${instanceName}" started successfully`, 'success');
            closeStartInstanceDialog();

            // Refresh instances list
            await loadInstancesForTag(currentTagForInstance);
        } else {
            showValidationError(data.details || data.error || 'Failed to start instance');
        }
    } catch (error) {
        console.error('Error starting instance:', error);
        showValidationError('Network error: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalHTML;
    }
}

/**
 * Stops a running instance
 * @param {string} instanceID - The instance ID to stop
 */
async function stopInstance(instanceID) {
    if (!confirm('Are you sure you want to stop this instance? All running agents will be terminated.')) {
        return;
    }

    const agencyID = getAgencyID();
    if (!agencyID) {
        showNotification('Agency ID not found', 'danger');
        return;
    }

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/instances/${instanceID}`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (response.ok) {
            showNotification('Instance stopped successfully', 'success');
            await loadAllInstances();
        } else {
            showNotification(data.details || data.error || 'Failed to stop instance', 'danger');
        }
    } catch (error) {
        console.error('Error stopping instance:', error);
        showNotification('Network error: ' + error.message, 'danger');
    }
}

/**
 * Restarts a stopped instance
 * @param {string} instanceID - The instance ID to restart
 */
async function restartInstance(instanceID) {
    const agencyID = getAgencyID();
    if (!agencyID) {
        showNotification('Agency ID not found', 'danger');
        return;
    }

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/instances/${instanceID}/restart`, {
            method: 'POST'
        });

        const data = await response.json();

        if (response.ok) {
            showNotification('Instance restarted successfully', 'success');
            await loadAllInstances();
        } else {
            showNotification(data.details || data.error || 'Failed to restart instance', 'danger');
        }
    } catch (error) {
        console.error('Error restarting instance:', error);
        showNotification('Network error: ' + error.message, 'danger');
    }
}

/**
 * Views instance details
 * @param {string} instanceID - The instance ID to view
 */
async function viewInstanceDetails(instanceID) {
    const agencyID = getAgencyID();
    if (!agencyID) {
        showNotification('Agency ID not found', 'danger');
        return;
    }

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/instances/${instanceID}`);
        const data = await response.json();

        if (response.ok) {
            // TODO: Display instance details in a modal or panel
            console.log('Instance details:', data.instance);
            alert('Instance details:\n' + JSON.stringify(data.instance, null, 2));
        } else {
            showNotification(data.details || data.error || 'Failed to load instance', 'danger');
        }
    } catch (error) {
        console.error('Error loading instance:', error);
        showNotification('Network error: ' + error.message, 'danger');
    }
}

// ==================== Data Loading ====================

/**
 * Loads all instances for the current agency
 */
async function loadAllInstances() {
    const agencyID = getAgencyID();
    if (!agencyID) {
        console.warn('No agency ID found, skipping instance load');
        return;
    }

    try {
        const url = `/api/v1/agencies/${agencyID}/instances`;
        console.log('Loading instances from:', url);

        const response = await fetch(url);

        if (!response.ok) {
            console.error(`Failed to load instances: ${response.status} ${response.statusText}`);
            const errorText = await response.text();
            console.error('Error response:', errorText);
            return;
        }

        const data = await response.json();
        console.log('Instances loaded:', data);

        if (data.instances) {
            // Group instances by tag
            const instancesByTag = {};
            data.instances.forEach(instance => {
                if (!instancesByTag[instance.tag_name]) {
                    instancesByTag[instance.tag_name] = [];
                }
                instancesByTag[instance.tag_name].push(instance);
            });

            // Update UI with instance counts
            updateInstanceCounts(instancesByTag);
        }
    } catch (error) {
        console.error('Error loading instances:', error);
        console.error('Error details:', {
            message: error.message,
            stack: error.stack
        });
    }
}

/**
 * Loads instances for a specific tag
 * @param {string} tagName - The tag name
 */
async function loadInstancesForTag(tagName) {
    const agencyID = getAgencyID();
    if (!agencyID) return;

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/tags/${tagName}/instances`);
        const data = await response.json();

        if (response.ok) {
            console.log(`Instances for tag ${tagName}:`, data.instances);
            // Update the instance count badge for this tag
            updateTagInstanceCount(tagName, data.count);
        }
    } catch (error) {
        console.error(`Error loading instances for tag ${tagName}:`, error);
    }
}

/**
 * Updates instance count badges in the tags table
 * @param {Object} instancesByTag - Map of tag names to instance arrays
 */
function updateInstanceCounts(instancesByTag) {
    Object.keys(instancesByTag).forEach(tagName => {
        const count = instancesByTag[tagName].length;
        updateTagInstanceCount(tagName, count);
    });
}

/**
 * Updates the instance count for a specific tag
 * @param {string} tagName - The tag name
 * @param {number} count - Number of instances
 */
function updateTagInstanceCount(tagName, count) {
    const badge = document.querySelector(`[data-tag-name="${tagName}"] .instance-count-badge`);
    if (badge) {
        badge.textContent = count;
        badge.className = count > 0 ? 'tag is-success instance-count-badge' : 'tag is-light instance-count-badge';
    }
}

// ==================== Helpers ====================

/**
 * Shows a validation error in the instance dialog
 * @param {string} message - Error message
 */
function showValidationError(message) {
    const validationEl = document.getElementById('instance-validation-messages');
    if (validationEl) {
        validationEl.innerHTML = `
            <div class="notification is-danger is-light">
                <button class="delete" onclick="this.parentElement.remove()"></button>
                ${message}
            </div>
        `;
    }
}

/**
 * Gets the current agency ID from the page
 * @returns {string|null} Agency ID or null
 */
function getAgencyID() {
    const pathParts = window.location.pathname.split('/');
    const agencyIndex = pathParts.indexOf('agencies');
    if (agencyIndex !== -1 && pathParts.length > agencyIndex + 1) {
        return pathParts[agencyIndex + 1];
    }
    return null;
}

// ==================== Initialization ====================

/**
 * Initializes instance management when the page loads
 */
function initializeInstanceManagement() {
    console.log('Instance management initialized');

    // Load instances on page load
    loadAllInstances();

    // Auto-refresh instances every 30 seconds
    if (instanceRefreshInterval) {
        clearInterval(instanceRefreshInterval);
    }
    instanceRefreshInterval = setInterval(loadAllInstances, 30000);
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeInstanceManagement);
} else {
    initializeInstanceManagement();
}
