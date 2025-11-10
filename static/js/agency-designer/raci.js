// RACI Matrix Editor for Agency Designer
// Loads work items and roles, allows assignment of R/A/C/I responsibilities
// Uses global specificationAPI

let raciState = {
    agencyId: null,
    roles: [], // array of role objects: {key, name, ...}
    workItems: [], // array of work item objects: {key, name, type, ...}
    assignments: {} // workItemKey -> roleKey -> {raci: 'R'|'A'|'C'|'I', objective: string}
};

// Get agency ID from URL or context
function getAgencyId() {
    // Try to get from URL path (e.g., /agencies/UC-CHAR-001/designer)
    const pathMatch = window.location.pathname.match(/\/agencies\/([^\/]+)/);
    if (pathMatch) {
        return pathMatch[1];
    }

    // Try to get from data attribute
    const designerElement = document.querySelector('[data-agency-id]');
    if (designerElement) {
        return designerElement.getAttribute('data-agency-id');
    }

    return null;
}

// Make loadRACIMatrix available globally for overview.js to call
window.loadRACIMatrix = function () {
    // Get agency ID if not already set
    if (!raciState.agencyId) {
        raciState.agencyId = getAgencyId();
        if (!raciState.agencyId) {
            console.error('[RACI] Unable to determine agency ID');
            return;
        }
    }

    const tableBody = document.getElementById('raci-matrix-body');
    if (!tableBody) {
        console.error('[RACI] Table body element not found');
        return;
    }

    // Show simple loading state in table
    tableBody.innerHTML = '<tr><td colspan="3" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading work items...</p></td></tr>';

    // Fetch work items and roles using specification API
    Promise.all([
        window.specificationAPI.getWorkItems(),
        window.specificationAPI.getRoles()
    ])
        .then(([workItems, roles]) => {
            raciState.workItems = workItems || [];
            raciState.roles = roles || [];

            // Render work items immediately
            renderRACITable();

            // Load RACI assignments in background
            loadExistingAssignments();
        })
        .catch(error => {
            console.error('[RACI] Error loading RACI matrix:', error);
            tableBody.innerHTML = '<tr><td colspan="3" class="has-text-danger has-text-centered py-5"><p>Error loading work items</p></td></tr>';
        });
}

function loadExistingAssignments() {
    // Load RACI assignments and re-render when complete
    fetch(`/api/v1/agencies/${raciState.agencyId}/raci-matrix`)
        .then(response => {
            if (response.ok) {
                return response.json();
            }
            return { assignments: {} };
        })
        .then(data => {
            raciState.assignments = data.assignments || {};
            // Re-render table with assignments
            renderRACITable();
        })
        .catch(error => {
            console.error('[RACI] Error loading RACI assignments:', error);
            raciState.assignments = {};
        });
}

function renderRACITable() {
    const tableBody = document.getElementById('raci-matrix-body');
    const emptyState = document.getElementById('raci-empty-state');
    const tableContainer = document.getElementById('raci-matrix-table');

    if (!tableBody) {
        console.error('[RACI] Table body not found in renderRACITable');
        return;
    }

    // Check if we have work items
    if (!raciState.workItems || raciState.workItems.length === 0) {
        if (tableContainer) tableContainer.style.display = 'none';
        if (emptyState) emptyState.style.display = 'block';
        return;
    }

    if (tableContainer) tableContainer.style.display = 'block';
    if (emptyState) emptyState.style.display = 'none';

    // Render work item rows with collapsible role assignments
    tableBody.innerHTML = raciState.workItems.map(workItem => {
        // Work items from ArangoDB have _key field
        const workItemKey = workItem._key || workItem.key || workItem.id;
        const assignments = raciState.assignments[workItemKey] || {};
        const assignmentCount = Object.keys(assignments).length;

        return `
            <tr data-work-item="${workItemKey}">
                <td colspan="3" class="p-0">
                    <div class="box mb-0">
                        <!-- Work Item Header (always visible) -->
                        <div class="level is-mobile mb-0" style="cursor: pointer;" onclick="toggleAssignments('${workItemKey}')">
                            <div class="level-left">
                                <div class="level-item">
                                    <span class="icon">
                                        <i id="toggle-icon-${workItemKey}" class="fas fa-chevron-right"></i>
                                    </span>
                                </div>
                                <div class="level-item">
                                    <div>
                                        <strong>${escapeHtml(workItem.name || workItem.title)}</strong>
                                        <span class="tag is-info is-light ml-2">${escapeHtml(workItem.type || 'Task')}</span>
                                        ${workItem.description ? `<p class="help is-size-7 mt-1">${escapeHtml(workItem.description)}</p>` : ''}
                                    </div>
                                </div>
                            </div>
                            <div class="level-right">
                                <div class="level-item">
                                    <span class="tag ${assignmentCount > 0 ? 'is-primary' : 'is-light'}">
                                        <span class="icon"><i class="fas fa-users"></i></span>
                                        <span>${assignmentCount} role${assignmentCount !== 1 ? 's' : ''}</span>
                                    </span>
                                </div>
                            </div>
                        </div>

                        <!-- Collapsible Assignments Panel -->
                        <div id="assignments-${workItemKey}" style="display: none;" class="mt-4 pt-4" style="border-top: 1px solid #dbdbdb;">
                            ${renderAssignmentsPanel(workItemKey, assignments)}
                        </div>
                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

function toggleAssignments(workItemKey) {
    const panel = document.getElementById(`assignments-${workItemKey}`);
    const icon = document.getElementById(`toggle-icon-${workItemKey}`);

    if (panel.style.display === 'none') {
        panel.style.display = 'block';
        icon.classList.remove('fa-chevron-right');
        icon.classList.add('fa-chevron-down');
    } else {
        panel.style.display = 'none';
        icon.classList.remove('fa-chevron-down');
        icon.classList.add('fa-chevron-right');
    }
}

function renderAssignmentsPanel(workItemKey, assignments) {
    const roles = Object.entries(assignments);

    let html = `
        <div class="level is-mobile mb-3">
            <div class="level-left">
                <div class="level-item">
                    <button class="button is-small is-primary" onclick="addRoleToWorkItem('${workItemKey}')">
                        <span class="icon"><i class="fas fa-plus"></i></span>
                        <span>Add Role</span>
                    </button>
                </div>
            </div>
        </div>
    `;

    if (roles.length === 0) {
        html += '<p class="help has-text-grey">No roles assigned yet. Click "Add Role" to assign responsibilities.</p>';
    } else {
        // Grid header
        html += `
            <div class="table-container">
                <table class="table is-fullwidth is-striped is-hoverable">
                    <thead>
                        <tr>
                            <th style="width: 30%">Role</th>
                            <th style="width: 10%">RACI</th>
                            <th style="width: 50%">Objective</th>
                            <th style="width: 10%">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        roles.forEach(([roleKey, data]) => {
            // Match using _key (ArangoDB key) or id (logical identifier)
            const role = raciState.roles.find(r => (r._key || r.key || r.id) === roleKey);
            const roleName = role ? (role.name || role._key || role.key) : roleKey;
            const raciType = data.raci || 'R';
            const objective = data.objective || '';
            const raciColors = {
                'R': 'is-info',
                'A': 'is-success',
                'C': 'is-warning',
                'I': 'is-light'
            };
            const raciLabels = {
                'R': 'Responsible',
                'A': 'Accountable',
                'C': 'Consulted',
                'I': 'Informed'
            };

            html += `
                <tr>
                    <td>
                        <strong>${escapeHtml(roleName)}</strong>
                    </td>
                    <td>
                        <span class="tag ${raciColors[raciType]}" title="${raciLabels[raciType]}">${raciType}</span>
                    </td>
                    <td>
                        ${objective ? `<p class="mb-0">${escapeHtml(objective)}</p>` : '<p class="help has-text-grey mb-0">No objective defined</p>'}
                    </td>
                    <td>
                        <div class="buttons are-small">
                            <button class="button is-light" onclick="editRoleObjective('${workItemKey}', '${roleKey}')" title="Edit assignment">
                                <span class="icon"><i class="fas fa-edit"></i></span>
                            </button>
                            <button class="button is-danger is-light" onclick="removeRoleFromWorkItem('${workItemKey}', '${roleKey}')" title="Remove role">
                                <span class="icon"><i class="fas fa-times"></i></span>
                            </button>
                        </div>
                    </td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;
    }

    return html;
}

window.addRoleToWorkItem = function (workItemKey) {
    if (!raciState.roles || raciState.roles.length === 0) {
        showError('No roles available. Please create roles first.');
        return;
    }

    // Create modal for role selection
    const modalHtml = `
        <div class="modal is-active" id="add-role-modal">
            <div class="modal-background" onclick="closeAddRoleModal()"></div>
            <div class="modal-card">
                <header class="modal-card-head">
                    <p class="modal-card-title">Add Role to Work Item</p>
                    <button class="delete" aria-label="close" onclick="closeAddRoleModal()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="field">
                        <label class="label">Select Role</label>
                        <div class="control">
                            <div class="select is-fullwidth">
                                <select id="role-select">
                                    ${raciState.roles.map(role => {
        const roleKey = role._key || role.key || role.id;
        const roleName = role.name || role._key || role.key;
        return `<option value="${roleKey}">${escapeHtml(roleName)}</option>`;
    }).join('')}
                                </select>
                            </div>
                        </div>
                    </div>
                    <div class="field">
                        <label class="label">RACI Type</label>
                        <div class="control">
                            <div class="select is-fullwidth">
                                <select id="raci-type-select">
                                    <option value="R">Responsible - Does the work</option>
                                    <option value="A">Accountable - Ultimately answerable</option>
                                    <option value="C">Consulted - Provides input</option>
                                    <option value="I">Informed - Kept updated</option>
                                </select>
                            </div>
                        </div>
                    </div>
                    <div class="field">
                        <label class="label">Objective</label>
                        <div class="control">
                            <textarea 
                                id="role-objective-input" 
                                class="textarea" 
                                rows="3"
                                placeholder="Define what this role needs to achieve for this work item..."></textarea>
                        </div>
                        <p class="help">Explain the specific goal or responsibility this role has for this work item.</p>
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="confirmAddRole('${workItemKey}')">Add Role</button>
                    <button class="button" onclick="closeAddRoleModal()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHtml);
}

window.closeAddRoleModal = function () {
    const modal = document.getElementById('add-role-modal');
    if (modal) {
        modal.remove();
    }
}

window.confirmAddRole = function (workItemKey) {
    const roleKey = document.getElementById('role-select').value;
    const raciType = document.getElementById('raci-type-select').value;
    const objective = document.getElementById('role-objective-input').value;

    if (!raciState.assignments[workItemKey]) {
        raciState.assignments[workItemKey] = {};
    }

    raciState.assignments[workItemKey][roleKey] = {
        raci: raciType,
        objective: objective
    };

    closeAddRoleModal();
    renderRACITable();
    showSuccess('Role added successfully');

    // Auto-save the new assignment
    saveRACIAssignment(workItemKey, roleKey);
}

window.removeRoleFromWorkItem = function (workItemKey, roleKey) {
    if (raciState.assignments[workItemKey]) {
        delete raciState.assignments[workItemKey][roleKey];
        if (Object.keys(raciState.assignments[workItemKey]).length === 0) {
            delete raciState.assignments[workItemKey];
        }
    }
    renderRACITable();

    // Auto-save after removal
    saveRACIAssignment(workItemKey, roleKey);
}

window.updateRoleRaci = function (workItemKey, roleKey, raciType) {
    if (raciState.assignments[workItemKey] && raciState.assignments[workItemKey][roleKey]) {
        raciState.assignments[workItemKey][roleKey].raci = raciType;
        renderRACITable();
    }
}

window.updateRoleObjective = function (workItemKey, roleKey, objective) {
    if (!raciState.assignments[workItemKey]) {
        raciState.assignments[workItemKey] = {};
    }
    if (!raciState.assignments[workItemKey][roleKey]) {
        raciState.assignments[workItemKey][roleKey] = { raci: 'R' };
    }
    raciState.assignments[workItemKey][roleKey].objective = objective;
}

// Save a single RACI assignment to the backend
window.saveRACIAssignment = async function (workItemKey, roleKey) {
    try {
        // Save entire matrix (this handles create/update/delete logic on backend)
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/raci-matrix`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                assignments: raciState.assignments
            })
        });

        if (!response.ok) {
            throw new Error('Failed to save assignment');
        }

        const data = await response.json();

        // Show subtle success indicator
        showNotification('Changes saved', 'success');

    } catch (error) {
        console.error('[RACI] Error saving assignment:', error);
        showNotification('Failed to save changes. Click "Save Matrix" to retry.', 'warning');
    }
}

window.editRoleObjective = function (workItemKey, roleKey) {
    // Get current assignment data
    const assignment = raciState.assignments[workItemKey]?.[roleKey] || {};
    const currentObjective = assignment.objective || '';
    const currentRaci = assignment.raci || 'R';

    // Get role name for display
    const role = raciState.roles.find(r => (r._key || r.key || r.id) === roleKey);
    const roleName = role ? (role.name || role._key || role.key) : roleKey;

    // Get work item name for display
    const workItem = raciState.workItems.find(w => (w._key || w.key || w.id) === workItemKey);
    const workItemName = workItem ? (workItem.name || workItem.title) : workItemKey;

    // Create modal for editing objective and RACI type
    const modalHtml = `
        <div class="modal is-active" id="edit-objective-modal">
            <div class="modal-background" onclick="closeEditObjectiveModal()"></div>
            <div class="modal-card">
                <header class="modal-card-head">
                    <p class="modal-card-title">Edit Role Assignment</p>
                    <button class="delete" aria-label="close" onclick="closeEditObjectiveModal()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="field">
                        <label class="label">Work Item</label>
                        <div class="content">
                            <p class="has-text-weight-semibold">${escapeHtml(workItemName)}</p>
                        </div>
                    </div>
                    <div class="field">
                        <label class="label">Role</label>
                        <div class="content">
                            <p class="has-text-weight-semibold">${escapeHtml(roleName)}</p>
                        </div>
                    </div>
                    <div class="field">
                        <label class="label">RACI Type</label>
                        <div class="control">
                            <div class="select is-fullwidth">
                                <select id="edit-raci-type-select">
                                    <option value="R" ${currentRaci === 'R' ? 'selected' : ''}>Responsible - Does the work</option>
                                    <option value="A" ${currentRaci === 'A' ? 'selected' : ''}>Accountable - Ultimately answerable</option>
                                    <option value="C" ${currentRaci === 'C' ? 'selected' : ''}>Consulted - Provides input</option>
                                    <option value="I" ${currentRaci === 'I' ? 'selected' : ''}>Informed - Kept updated</option>
                                </select>
                            </div>
                        </div>
                    </div>
                    <div class="field">
                        <label class="label">Objective</label>
                        <div class="control">
                            <textarea 
                                id="edit-objective-input" 
                                class="textarea" 
                                rows="4"
                                placeholder="Define what this role needs to achieve for this work item..."
                            >${escapeHtml(currentObjective)}</textarea>
                        </div>
                        <p class="help">Explain the specific goal or responsibility this role has for this work item.</p>
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="confirmEditObjective('${workItemKey}', '${roleKey}')">
                        <span class="icon"><i class="fas fa-save"></i></span>
                        <span>Save Changes</span>
                    </button>
                    <button class="button" onclick="closeEditObjectiveModal()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHtml);

    // Focus on textarea
    setTimeout(() => {
        const textarea = document.getElementById('edit-objective-input');
        if (textarea) {
            textarea.focus();
            textarea.setSelectionRange(textarea.value.length, textarea.value.length);
        }
    }, 100);
}

window.closeEditObjectiveModal = function () {
    const modal = document.getElementById('edit-objective-modal');
    if (modal) {
        modal.remove();
    }
}

window.confirmEditObjective = function (workItemKey, roleKey) {
    const textarea = document.getElementById('edit-objective-input');
    const raciSelect = document.getElementById('edit-raci-type-select');
    if (!textarea || !raciSelect) return;

    const objective = textarea.value.trim();
    const raciType = raciSelect.value;

    // Update state with both objective and RACI type
    if (!raciState.assignments[workItemKey]) {
        raciState.assignments[workItemKey] = {};
    }
    if (!raciState.assignments[workItemKey][roleKey]) {
        raciState.assignments[workItemKey][roleKey] = {};
    }
    raciState.assignments[workItemKey][roleKey].raci = raciType;
    raciState.assignments[workItemKey][roleKey].objective = objective;

    // Save to backend
    saveRACIAssignment(workItemKey, roleKey);

    // Re-render table to show updated values
    renderRACITable();

    // Close modal
    closeEditObjectiveModal();
}

// Keep the updateRoleObjective helper function
window.saveRoleObjective = function (workItemKey, roleKey) {
    // This function is no longer used with modal approach but keeping for compatibility
    const objective = raciState.assignments[workItemKey]?.[roleKey]?.objective || '';
    updateRoleObjective(workItemKey, roleKey, objective);
    saveRACIAssignment(workItemKey, roleKey);
    renderRACITable();
}

window.cancelEditObjective = function (workItemKey, roleKey) {
    // This function is no longer used with modal approach but keeping for compatibility
    closeEditObjectiveModal();
}

window.setRACIAssignment = function (workItemKey, roleKey, raciRole) {
    if (!raciState.assignments[workItemKey]) {
        raciState.assignments[workItemKey] = {};
    }
    raciState.assignments[workItemKey][roleKey] = raciRole;
    renderRACITable();
}

window.clearRACIAssignment = function (workItemKey, roleKey) {
    if (raciState.assignments[workItemKey]) {
        delete raciState.assignments[workItemKey][roleKey];
    }
    renderRACITable();
}

window.clearWorkItemRow = function (workItemKey) {
    delete raciState.assignments[workItemKey];
    renderRACITable();
}

window.saveRACIMatrix = async function () {
    try {
        // Validate before saving
        const validation = validateRACIMatrix();
        if (!validation.valid) {
            showValidationErrors(validation.errors);
            return;
        }

        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/raci-matrix`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                assignments: raciState.assignments
            })
        });

        if (!response.ok) throw new Error('Failed to save RACI matrix');

        showSuccess('RACI matrix saved successfully');
    } catch (error) {
        console.error('Error saving RACI matrix:', error);
        showError('Failed to save RACI matrix');
    }
}

window.validateRACIMatrix = function () {
    const errors = [];
    const warnings = [];

    raciState.workItems.forEach(workItem => {
        const workItemKey = workItem._key || workItem.key || workItem.id;
        const assignments = raciState.assignments[workItemKey] || {};
        const roleEntries = Object.entries(assignments);

        if (roleEntries.length === 0) {
            warnings.push(`"${workItem.name}" has no roles assigned`);
            return;
        }

        // Check for exactly one Accountable
        const accountableCount = roleEntries.filter(([_, data]) => data.raci === 'A').length;
        if (accountableCount === 0) {
            errors.push(`"${workItem.name}" has no Accountable (A) role assigned`);
        } else if (accountableCount > 1) {
            errors.push(`"${workItem.name}" has multiple Accountable (A) roles (should be exactly one)`);
        }

        // Check for at least one Responsible
        const responsibleCount = roleEntries.filter(([_, data]) => data.raci === 'R').length;
        if (responsibleCount === 0) {
            errors.push(`"${workItem.name}" has no Responsible (R) role assigned`);
        }

        // Check for objectives
        roleEntries.forEach(([roleKey, data]) => {
            if (!data.objective || data.objective.trim() === '') {
                const role = raciState.roles.find(r => (r._key || r.key || r.id) === roleKey);
                const roleName = role ? (role.name || role._key || role.key) : roleKey;
                warnings.push(`"${workItem.name}" - Role "${roleName}" has no objective defined`);
            }
        });
    });

    if (errors.length > 0 || warnings.length > 0) {
        showValidationErrors(errors, warnings);
        return { valid: errors.length === 0, errors, warnings };
    }

    showSuccess('RACI matrix validation passed');
    return { valid: true, errors: [], warnings: [] };
}

function showValidationErrors(errors, warnings = []) {
    const container = document.getElementById('raci-validation-messages');

    const hasErrors = errors.length > 0;
    const hasWarnings = warnings.length > 0;

    if (!hasErrors && !hasWarnings) {
        container.innerHTML = '';
        return;
    }

    let html = '';

    if (hasErrors) {
        html += `
            <div class="notification is-danger mb-3">
                <button class="delete" onclick="this.parentElement.remove()"></button>
                <p class="has-text-weight-bold mb-2">❌ Validation Errors:</p>
                <ul class="ml-4">
                    ${errors.map(err => `<li>${escapeHtml(err)}</li>`).join('')}
                </ul>
                <p class="help mt-2">
                    Fix these errors before saving. Each work item must have exactly one Accountable 
                    and at least one Responsible role.
                </p>
            </div>
        `;
    }

    if (hasWarnings) {
        html += `
            <div class="notification is-warning">
                <button class="delete" onclick="this.parentElement.remove()"></button>
                <p class="has-text-weight-bold mb-2">⚠️ Validation Warnings:</p>
                <ul class="ml-4">
                    ${warnings.map(warn => `<li>${escapeHtml(warn)}</li>`).join('')}
                </ul>
                <p class="help mt-2">
                    These are recommendations. You can save with warnings, but it's better to address them.
                </p>
            </div>
        `;
    }

    container.innerHTML = html;
}

// AI-powered RACI mapping creation
window.createRACIMappings = async function () {
    if (!raciState.agencyId) {
        raciState.agencyId = getAgencyId();
        if (!raciState.agencyId) {
            showError('Error: No agency selected');
            return;
        }
    }

    if (!raciState.workItems || raciState.workItems.length === 0) {
        showError('No work items available. Please create work items first.');
        return;
    }

    if (!raciState.roles || raciState.roles.length === 0) {
        showError('No roles available. Please create roles first.');
        return;
    }

    const statusMessage = 'AI is generating RACI mappings from your work items and roles...';

    // Show AI processing status in the chat area
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    try {
        // Disable button and show loading state
        const btn = document.getElementById('ai-create-mappings-btn');
        const originalHTML = btn.innerHTML;
        btn.disabled = true;
        btn.innerHTML = '<span class="icon"><i class="fas fa-spinner fa-spin"></i></span><span>Creating...</span>';

        // Call AI endpoint to create RACI mappings
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/raci-matrix/ai-generate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                operations: ['create']
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to generate RACI mappings: ${response.statusText}`);
        }

        const data = await response.json();

        // Update status to show we're processing results
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('Processing results and updating RACI mappings...');
        }

        // Update assignments
        raciState.assignments = data.assignments || {};

        // Re-render table with new mappings
        renderRACITable();

        // Refresh chat messages so AI explanation appears in the chat
        try {
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                const chatResp = await fetch(`/agencies/${raciState.agencyId}/chat-messages`);
                if (chatResp.ok) {
                    const chatHtml = await chatResp.text();
                    chatContainer.innerHTML = chatHtml;
                    // Scroll to bottom to show latest assistant message
                    if (window.scrollToBottom) {
                        try { window.scrollToBottom(chatContainer); } catch (e) { /* ignore */ }
                    }
                }
            }
        } catch (err) {
            console.error('[RACI] Error refreshing chat:', err);
        }

        // Hide AI processing status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        // Show success message
        const mappingCount = Object.keys(raciState.assignments).length;
        if (mappingCount > 0) {
            showSuccess(`Successfully created ${mappingCount} RACI mapping(s)! Review and adjust as needed.`);
        } else {
            showSuccess('RACI mappings created successfully!');
        }

        // Restore button
        btn.disabled = false;
        btn.innerHTML = originalHTML;

    } catch (error) {
        console.error('[RACI] Error creating mappings:', error);

        // Hide AI processing status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        showError(`AI processing failed: ${error.message}`);

        // Restore button
        const btn = document.getElementById('ai-create-mappings-btn');
        if (btn) {
            btn.disabled = false;
            btn.innerHTML = '<span class="icon"><i class="fas fa-robot"></i></span><span>Create Mappings</span>';
        }
    }
}

window.exportRACIMatrix = async function (format) {
    try {
        const url = `/api/v1/agencies/${raciState.agencyId}/raci-matrix/export/${format}`;
        window.open(url, '_blank');
    } catch (error) {
        console.error(`Error exporting RACI matrix as ${format}:`, error);
        showError(`Failed to export as ${format}`);
    }
}

window.toggleExportDropdown = function () {
    const dropdown = document.getElementById('export-dropdown');
    dropdown.classList.toggle('is-active');
}

window.applyTemplate = function () {
    const modal = document.getElementById('template-modal');
    modal.classList.add('is-active');
    loadTemplates();
}

window.closeTemplateModal = function () {
    const modal = document.getElementById('template-modal');
    modal.classList.remove('is-active');
}

async function loadTemplates() {
    try {
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/raci-templates`);
        if (!response.ok) throw new Error('Failed to load templates');

        const templates = await response.json();
        const container = document.getElementById('template-list');

        container.innerHTML = templates.map((template, index) => `
            <label class="radio">
                <input type="radio" name="template" value="${index}">
                ${escapeHtml(template.name)} - ${escapeHtml(template.description)}
            </label>
        `).join('<br>');

    } catch (error) {
        console.error('Error loading templates:', error);
        showError('Failed to load templates');
    }
}

function showLoading(show) {
    const loading = document.getElementById('raci-loading');
    const table = document.getElementById('raci-matrix-table');

    if (show) {
        loading.style.display = 'block';
        // Don't hide the table - let work items show while relationships load
        table.style.display = 'table';
    } else {
        loading.style.display = 'none';
        table.style.display = 'table';
    }
}

function showSuccess(message) {
    showNotification(message, 'success');
}

function showError(message) {
    showNotification(message, 'danger');
}

function showNotification(message, type = 'info') {
    // Create a notification toast
    const notification = document.createElement('div');
    notification.className = `notification is-${type} is-light`;
    notification.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 1000; min-width: 300px;';
    notification.innerHTML = `
        <button class="delete" onclick="this.parentElement.remove()"></button>
        ${escapeHtml(message)}
    `;
    document.body.appendChild(notification);

    // Auto-remove after 5 seconds
    setTimeout(() => notification.remove(), 5000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Close dropdowns when clicking outside
document.addEventListener('click', function (event) {
    if (!event.target.closest('.dropdown')) {
        document.querySelectorAll('.dropdown.is-active').forEach(dropdown => {
            dropdown.classList.remove('is-active');
        });
    }
});
