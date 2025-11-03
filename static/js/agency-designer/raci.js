// RACI Matrix Editor for Agency Designer
// Loads work items and roles, allows assignment of R/A/C/I responsibilities

let raciState = {
    agencyId: null,
    roles: [], // array of role objects: {key, name, ...}
    workItems: [], // array of work item objects: {key, name, type, ...}
    assignments: {} // workItemKey -> roleKey -> 'R'|'A'|'C'|'I'
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', function () {
    initializeRACIMatrix();
});

function initializeRACIMatrix() {
    // Extract agency ID from URL path: /agencies/{id}/raci
    const pathParts = window.location.pathname.split('/');
    const agenciesIndex = pathParts.indexOf('agencies');
    if (agenciesIndex !== -1 && pathParts[agenciesIndex + 1]) {
        raciState.agencyId = pathParts[agenciesIndex + 1];
        loadRACIMatrix();
    } else {
        showError('Unable to determine agency ID from URL');
    }
}

async function loadRACIMatrix() {
    showLoading(true);

    try {
        // Load roles and work items in parallel
        const [roles, workItems] = await Promise.all([
            fetchRoles(),
            fetchWorkItems()
        ]);

        raciState.roles = roles || [];
        raciState.workItems = workItems || [];

        // Show work items immediately even if no assignments yet
        renderRACITable();
        showLoading(false);

        // Load existing RACI assignments in the background
        await loadExistingAssignments();

        // Re-render with assignments
        renderRACITable();

    } catch (error) {
        console.error('Error loading RACI matrix:', error);
        showError('Failed to load RACI matrix data');
        showLoading(false);
    }
}

async function fetchRoles() {
    try {
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/roles`);
        if (!response.ok) throw new Error('Failed to fetch roles');
        return await response.json();
    } catch (error) {
        console.error('Error fetching roles:', error);
        return [];
    }
}

async function fetchWorkItems() {
    try {
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/work-items`);
        if (!response.ok) throw new Error('Failed to fetch work items');
        return await response.json();
    } catch (error) {
        console.error('Error fetching work items:', error);
        return [];
    }
}

async function loadExistingAssignments() {
    try {
        const response = await fetch(`/api/v1/agencies/${raciState.agencyId}/raci-matrix`);
        if (response.ok) {
            const data = await response.json();
            raciState.assignments = data.assignments || {};
        }
    } catch (error) {
        console.error('Error loading RACI assignments:', error);
        // Start with empty assignments
        raciState.assignments = {};
    }
}

function renderRACITable() {
    const tableBody = document.getElementById('raci-matrix-body');
    const emptyState = document.getElementById('raci-empty-state');
    const tableContainer = document.getElementById('raci-matrix-table');

    // Check if we have work items
    if (!raciState.workItems || raciState.workItems.length === 0) {
        tableContainer.style.display = 'none';
        emptyState.style.display = 'block';
        return;
    }

    tableContainer.style.display = 'block';
    emptyState.style.display = 'none';

    // Render work item rows with role assignments
    tableBody.innerHTML = raciState.workItems.map(workItem => {
        const workItemKey = workItem.key || workItem.id;
        const assignments = raciState.assignments[workItemKey] || {};

        return `
            <tr data-work-item="${workItemKey}">
                <td class="is-vcentered">
                    <div class="mb-2">
                        <strong>${escapeHtml(workItem.name || workItem.title)}</strong>
                        <span class="tag is-info is-light ml-2">${escapeHtml(workItem.type || 'Task')}</span>
                    </div>
                    ${workItem.description ? `<p class="help is-size-7">${escapeHtml(workItem.description)}</p>` : ''}
                </td>
                <td class="is-vcentered">
                    <div id="roles-list-${workItemKey}" class="mb-2">
                        ${renderAssignedRoles(workItemKey, assignments)}
                    </div>
                    <button class="button is-small is-primary is-light" onclick="addRoleToWorkItem('${workItemKey}')">
                        <span class="icon"><i class="fas fa-plus"></i></span>
                        <span>Add Role</span>
                    </button>
                </td>
                <td>
                    <div id="objectives-${workItemKey}">
                        ${renderRoleObjectives(workItemKey, assignments)}
                    </div>
                </td>
            </tr>
        `;
    }).join('');
}

function renderAssignedRoles(workItemKey, assignments) {
    const roles = Object.entries(assignments);
    if (roles.length === 0) {
        return '<p class="help is-size-7 has-text-grey">No roles assigned yet</p>';
    }

    return roles.map(([roleKey, data]) => {
        const role = raciState.roles.find(r => (r.key || r.id) === roleKey);
        const roleName = role ? (role.name || role.key) : roleKey;
        const raciType = data.raci || 'R';
        const raciColors = {
            'R': 'is-info',
            'A': 'is-success',
            'C': 'is-warning',
            'I': 'is-light'
        };

        return `
            <div class="tags has-addons mb-2">
                <span class="tag ${raciColors[raciType]}">${raciType}</span>
                <span class="tag">${escapeHtml(roleName)}</span>
                <a class="tag is-delete" onclick="removeRoleFromWorkItem('${workItemKey}', '${roleKey}')"></a>
            </div>
        `;
    }).join('');
}

function renderRoleObjectives(workItemKey, assignments) {
    const roles = Object.entries(assignments);
    if (roles.length === 0) {
        return '<p class="help is-size-7 has-text-grey">Add roles to define their objectives</p>';
    }

    return roles.map(([roleKey, data]) => {
        const role = raciState.roles.find(r => (r.key || r.id) === roleKey);
        const roleName = role ? (role.name || role.key) : roleKey;
        const objective = data.objective || '';
        const raciType = data.raci || 'R';

        return `
            <div class="box p-3 mb-2">
                <div class="level is-mobile mb-2">
                    <div class="level-left">
                        <div class="level-item">
                            <span class="tag is-${raciType === 'R' ? 'info' : raciType === 'A' ? 'success' : raciType === 'C' ? 'warning' : 'light'} mr-2">
                                ${raciType}
                            </span>
                            <strong class="is-size-7">${escapeHtml(roleName)}</strong>
                        </div>
                    </div>
                    <div class="level-right">
                        <div class="level-item">
                            <div class="select is-small">
                                <select onchange="updateRoleRaci('${workItemKey}', '${roleKey}', this.value)">
                                    <option value="R" ${raciType === 'R' ? 'selected' : ''}>Responsible</option>
                                    <option value="A" ${raciType === 'A' ? 'selected' : ''}>Accountable</option>
                                    <option value="C" ${raciType === 'C' ? 'selected' : ''}>Consulted</option>
                                    <option value="I" ${raciType === 'I' ? 'selected' : ''}>Informed</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="field">
                    <label class="label is-size-7">Objective</label>
                    <div class="control">
                        <textarea 
                            class="textarea is-small" 
                            rows="2"
                            placeholder="Define what this role needs to achieve for this work item..."
                            onblur="updateRoleObjective('${workItemKey}', '${roleKey}', this.value)"
                        >${escapeHtml(objective)}</textarea>
                    </div>
                </div>
            </div>
        `;
    }).join('');
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
        const roleKey = role.key || role.id;
        const roleName = role.name || role.key;
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
}

window.removeRoleFromWorkItem = function (workItemKey, roleKey) {
    if (raciState.assignments[workItemKey]) {
        delete raciState.assignments[workItemKey][roleKey];
        if (Object.keys(raciState.assignments[workItemKey]).length === 0) {
            delete raciState.assignments[workItemKey];
        }
    }
    renderRACITable();
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
        const workItemKey = workItem.key || workItem.id;
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
                const role = raciState.roles.find(r => (r.key || r.id) === roleKey);
                const roleName = role ? (role.name || role.key) : roleKey;
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
