// Roles functionality
// Handles roles management in Agency Designer

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';

// Role editor state management
let roleEditorState = {
    mode: 'add', // 'add' or 'edit'
    roleKey: null,
    originalData: {}
};

// Load roles list
export function loadRoles() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const rolesTableBody = document.getElementById('roles-table-body');
    if (!rolesTableBody) {
        return;
    }

    // Show loading state
    rolesTableBody.innerHTML = '<tr><td colspan="5" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading roles...</p></td></tr>';

    // Fetch roles HTML from API
    const url = `/api/v1/agencies/${agencyId}/roles/html`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load roles');
            }
            return response.text();
        })
        .then(html => {
            rolesTableBody.innerHTML = html;
        })
        .catch(error => {
            console.error('[Roles] Error loading roles:', error);
            rolesTableBody.innerHTML = '<tr><td colspan="5" class="has-text-danger has-text-centered py-5"><p>Error loading roles</p></td></tr>';
        });
}

// Show role editor
export function showRoleEditor(mode, roleKey = null) {
    console.log(`[Roles] Show editor in ${mode} mode`, roleKey);

    roleEditorState.mode = mode;
    roleEditorState.roleKey = roleKey;

    // Update title
    const title = document.getElementById('role-editor-title');
    if (title) {
        title.textContent = mode === 'add' ? 'Add New Role' : 'Edit Role';
    }

    // Clear form or load data
    if (mode === 'add') {
        clearRoleForm();
    } else if (mode === 'edit' && roleKey) {
        loadRoleData(roleKey);
    }

    // Show editor card
    const editorCard = document.getElementById('role-editor-card');
    if (editorCard) {
        editorCard.classList.remove('is-hidden');
        // Scroll to editor
        editorCard.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
}

// Load role data for editing
function loadRoleData(roleKey) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const url = `/api/v1/agencies/${agencyId}/roles/${roleKey}`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load role data');
            }
            return response.json();
        })
        .then(role => {
            roleEditorState.originalData = role;
            populateRoleForm(role);
        })
        .catch(error => {
            console.error('[Roles] Error loading role data:', error);
            showNotification('Error loading role data', 'danger');
        });
}

// Populate form with role data
function populateRoleForm(role) {
    console.log('[Roles] Populating form with role data:', role);

    const fields = {
        'role-name-editor': role.name || '',
        'role-tags-editor': (role.tags || []).join(', '),
        'role-description-editor': role.description || '',
        'role-autonomy-level-editor': role.autonomy_level || '',
        'role-capabilities-editor': (role.capabilities || []).join('\n'),
        'role-required-skills-editor': (role.required_skills || []).join(', '),
        'role-token-budget-editor': role.token_budget || 0,
        'role-icon-editor': role.icon || '',
        'role-color-editor': role.color || '#3298dc'
    };

    for (const [id, value] of Object.entries(fields)) {
        const element = document.getElementById(id);
        if (element) {
            element.value = value;
        } else {
            console.warn(`[Roles] Element not found: ${id}`);
        }
    }

    console.log('[Roles] Form population complete');
}

// Clear role form
function clearRoleForm() {
    const fields = [
        'role-name-editor',
        'role-tags-editor',
        'role-description-editor',
        'role-autonomy-level-editor',
        'role-capabilities-editor',
        'role-required-skills-editor',
        'role-token-budget-editor',
        'role-icon-editor',
        'role-color-editor'
    ];

    fields.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            if (element.type === 'color') {
                element.value = '#3298dc';
            } else if (element.type === 'number') {
                element.value = 0;
            } else {
                element.value = '';
            }
        }
    });
}

// Save role from editor
export function saveRoleFromEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    // Gather form data
    const name = document.getElementById('role-name-editor')?.value.trim();
    const tagsText = document.getElementById('role-tags-editor')?.value.trim();
    const description = document.getElementById('role-description-editor')?.value.trim();
    const autonomyLevel = document.getElementById('role-autonomy-level-editor')?.value;
    const capabilitiesText = document.getElementById('role-capabilities-editor')?.value.trim();
    const requiredSkillsText = document.getElementById('role-required-skills-editor')?.value.trim();
    const tokenBudget = parseInt(document.getElementById('role-token-budget-editor')?.value || '0');
    const icon = document.getElementById('role-icon-editor')?.value.trim();
    const color = document.getElementById('role-color-editor')?.value;

    // Validation
    if (!name) {
        showNotification('Please enter a role name', 'warning');
        return;
    }

    if (!autonomyLevel) {
        showNotification('Please select an autonomy level', 'warning');
        return;
    }

    // Parse tags, capabilities and skills
    const tags = tagsText
        ? tagsText.split(',').map(t => t.trim()).filter(t => t)
        : [];

    const capabilities = capabilitiesText
        ? capabilitiesText.split('\n').map(c => c.trim().replace(/^-\s*/, '')).filter(c => c)
        : [];

    const requiredSkills = requiredSkillsText
        ? requiredSkillsText.split(',').map(s => s.trim()).filter(s => s)
        : [];

    // Prepare payload
    const payload = {
        name,
        tags,
        description,
        autonomy_level: autonomyLevel,
        capabilities,
        required_skills: requiredSkills,
        token_budget: tokenBudget,
        icon,
        color
    };

    console.log('[Roles] Saving role:', payload);

    // Determine URL and method
    const { mode, roleKey } = roleEditorState;
    const url = mode === 'add'
        ? `/api/v1/agencies/${agencyId}/roles`
        : `/api/v1/agencies/${agencyId}/roles/${roleKey}`;
    const method = mode === 'add' ? 'POST' : 'PUT';

    // Send request
    fetch(url, {
        method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || 'Failed to save role');
                });
            }
            return response.json();
        })
        .then(result => {
            console.log('[Roles] Role saved successfully:', result);
            showNotification(
                mode === 'add' ? 'Role created successfully' : 'Role updated successfully',
                'success'
            );

            // Hide editor and reload list
            cancelRoleEdit();
            loadRoles();

            // Scroll chat to bottom to show any AI messages
            scrollToBottom();
        })
        .catch(error => {
            console.error('[Roles] Error saving role:', error);
            showNotification(`Error: ${error.message}`, 'danger');
        });
}

// Cancel role edit
export function cancelRoleEdit() {
    const editorCard = document.getElementById('role-editor-card');
    if (editorCard) {
        editorCard.classList.add('is-hidden');
    }

    clearRoleForm();
    roleEditorState = {
        mode: 'add',
        roleKey: null,
        originalData: {}
    };
}

// Delete role
export function deleteRole(roleKey) {
    if (!confirm('Are you sure you want to delete this role?')) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const url = `/api/v1/agencies/${agencyId}/roles/${roleKey}`;

    fetch(url, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete role');
            }
            showNotification('Role deleted successfully', 'success');
            loadRoles();
        })
        .catch(error => {
            console.error('[Roles] Error deleting role:', error);
            showNotification('Error deleting role', 'danger');
        });
}

// Filter roles
export function filterRoles() {
    const searchInput = document.getElementById('roles-search');
    if (!searchInput) return;

    const searchTerm = searchInput.value.toLowerCase();
    const tableRows = document.querySelectorAll('#roles-table-body tr.table-item');

    tableRows.forEach(row => {
        const text = row.textContent.toLowerCase();
        if (text.includes(searchTerm)) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// AI Role Operations
export async function processAIRoleOperation(operations) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    console.log('[Roles] Processing AI role operation:', operations);

    // Get selected role keys if needed
    const selectedRoleKeys = operations.includes('enhance') || operations.includes('consolidate')
        ? getSelectedRoleKeys()
        : [];

    // Validate selection for operations that require it
    if ((operations.includes('enhance') || operations.includes('consolidate')) && selectedRoleKeys.length === 0) {
        showNotification('Please select roles first', 'warning');
        return;
    }

    const payload = {
        operations,
        role_keys: selectedRoleKeys
    };

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/ai/process-roles`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error('AI operation failed');
        }

        const result = await response.json();
        console.log('[Roles] AI operation result:', result);

        showNotification('AI operation completed', 'success');

        // Reload roles list
        loadRoles();

        // Scroll to chat to show AI response
        scrollToBottom();

    } catch (error) {
        console.error('[Roles] Error processing AI operation:', error);
        showNotification('Error processing AI operation', 'danger');
    }
}

// Get selected role keys from checkboxes
function getSelectedRoleKeys() {
    const checkboxes = document.querySelectorAll('.role-checkbox:checked');
    return Array.from(checkboxes).map(cb => cb.dataset.roleKey);
}

// Update selection buttons based on checkbox state
function updateRoleSelectionButtons() {
    const selectedCount = document.querySelectorAll('.role-checkbox:checked').length;

    // Update count display
    updateRoleSelectionCount(selectedCount);

    // Enable/disable buttons based on selection
    const enhanceBtn = document.getElementById('ai-enhance-roles-btn');
    const consolidateBtn = document.getElementById('ai-consolidate-roles-btn');

    if (enhanceBtn) {
        if (selectedCount > 0) {
            enhanceBtn.classList.remove('is-static');
            enhanceBtn.disabled = false;
        } else {
            enhanceBtn.classList.add('is-static');
            enhanceBtn.disabled = true;
        }
    }

    if (consolidateBtn) {
        if (selectedCount > 1) {
            consolidateBtn.classList.remove('is-static');
            consolidateBtn.disabled = false;
        } else {
            consolidateBtn.classList.add('is-static');
            consolidateBtn.disabled = true;
        }
    }
}

// Toggle all role checkboxes
function toggleAllRoles(checked) {
    const checkboxes = document.querySelectorAll('.role-checkbox');
    checkboxes.forEach(cb => {
        cb.checked = checked;
    });
    updateRoleSelectionButtons();
}

// Update selection count display
function updateRoleSelectionCount(count) {
    const countDisplay = document.getElementById('role-selection-count');
    if (countDisplay) {
        if (count > 0) {
            countDisplay.textContent = `${count} selected`;
            countDisplay.style.display = '';
        } else {
            countDisplay.style.display = 'none';
        }
    }
}

// Initialize button states on page load
document.addEventListener('DOMContentLoaded', function () {
    updateRoleSelectionButtons();
});

// Make functions available globally
window.loadRoles = loadRoles;
window.showRoleEditor = showRoleEditor;
window.saveRoleFromEditor = saveRoleFromEditor;
window.cancelRoleEdit = cancelRoleEdit;
window.deleteRole = deleteRole;
window.filterRoles = filterRoles;
window.processAIRoleOperation = processAIRoleOperation;
window.updateRoleSelectionButtons = updateRoleSelectionButtons;
window.toggleAllRoles = toggleAllRoles;
