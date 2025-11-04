// Roles functionality
// Handles roles management in Agency Designer

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';
import { loadEntityList, showEntityEditor, cancelEntityEdit, deleteEntity, saveEntity, populateForm, clearForm } from './crud-helpers.js';

// Role editor state management
let roleEditorState = {
    mode: 'add', // 'add' or 'edit'
    roleKey: null,
    originalData: {}
};

// Load roles list
export function loadRoles() {
    return loadEntityList('roles', 'roles-table-body', 5);
}

// Show role editor
export function showRoleEditor(mode, roleKey = null) {
    roleEditorState.mode = mode;
    roleEditorState.roleKey = roleKey;

    showEntityEditor(
        mode,
        'role-editor-card',
        'roles-list-card',
        'role-editor-title',
        'Add New Role',
        'Edit Role',
        'role-name-editor'
    );

    if (mode === 'add') {
        clearRoleForm();
    } else if (mode === 'edit' && roleKey) {
        loadRoleData(roleKey);
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
    populateForm({
        'role-name-editor': role.name || '',
        'role-tags-editor': (role.tags || []).join(', '),
        'role-description-editor': role.description || '',
        'role-autonomy-level-editor': role.autonomy_level || '',
        'role-capabilities-editor': (role.capabilities || []).join('\n'),
        'role-required-skills-editor': (role.required_skills || []).join(', '),
        'role-token-budget-editor': role.token_budget || 0,
        'role-icon-editor': role.icon || '',
        'role-color-editor': role.color || '#3298dc'
    });
}

// Clear role form
function clearRoleForm() {
    clearForm([
        'role-name-editor',
        'role-tags-editor',
        'role-description-editor',
        'role-autonomy-level-editor',
        'role-capabilities-editor',
        'role-required-skills-editor',
        'role-token-budget-editor',
        'role-icon-editor',
        'role-color-editor'
    ], {
        'role-color-editor': '#3298dc',
        'role-token-budget-editor': 0
    });
}

// Save role from editor
export function saveRoleFromEditor() {
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

    // Construct payload
    const payload = {
        name,
        tags,
        description,
        autonomy_level: autonomyLevel,
        capabilities,
        required_skills: requiredSkills,
        token_budget: tokenBudget,
        icon,
        color,
        // Include version: preserve existing for edit, default for add
        version: roleEditorState.mode === 'edit' && roleEditorState.originalData?.version
            ? roleEditorState.originalData.version
            : '1.0.0'
    };

    saveEntity('roles', roleEditorState.mode, roleEditorState.roleKey, payload, 'save-role-btn', () => {
        cancelRoleEdit();
        loadRoles();
        scrollToBottom();
    });
}

// Cancel role edit
export function cancelRoleEdit() {
    cancelEntityEdit('role-editor-card', 'roles-list-card', [
        'role-name-editor',
        'role-tags-editor',
        'role-description-editor',
        'role-autonomy-level-editor',
        'role-capabilities-editor',
        'role-required-skills-editor',
        'role-token-budget-editor',
        'role-icon-editor',
        'role-color-editor'
    ]);

    roleEditorState = {
        mode: 'add',
        roleKey: null,
        originalData: {}
    };
}

// Delete role
export function deleteRole(roleKey) {
    deleteEntity('roles', roleKey, 'this role', loadRoles);
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
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Get selected role keys if needed
    const selectedRoleKeys = operations.includes('enhance') || operations.includes('consolidate')
        ? getSelectedRoleKeys()
        : [];

    // Validate selection for operations that require it
    if ((operations.includes('enhance') || operations.includes('consolidate')) && selectedRoleKeys.length === 0) {
        showNotification('Please select roles first', 'warning');
        return;
    }

    // Determine status message
    let statusMessage = 'AI is processing your request...';
    if (operations.length === 1) {
        switch (operations[0]) {
            case 'create':
                statusMessage = 'AI is generating roles from your work items...';
                break;
            case 'enhance':
                statusMessage = `AI is enhancing ${selectedRoleKeys.length} role(s)...`;
                break;
            case 'consolidate':
                statusMessage = `AI is consolidating ${selectedRoleKeys.length} role(s)...`;
                break;
        }
    } else if (operations.length > 1) {
        statusMessage = `AI is performing ${operations.length} operations on your roles...`;
    }

    // Show AI processing status in the chat area
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    const payload = {
        operations,
        role_keys: selectedRoleKeys
    };

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/roles/ai-process`, {
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

        // Update status to show we're processing results
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('Processing results and updating roles...');
        }

        // Reload roles list
        await loadRoles();

        // After roles are reloaded, refresh chat messages so AI explanation appears in the chat
        try {
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                const chatResp = await fetch(`/agencies/${agencyId}/chat-messages`);
                if (chatResp.ok) {
                    const chatHtml = await chatResp.text();
                    chatContainer.innerHTML = chatHtml;
                    // Scroll to bottom to show latest assistant message
                    try { scrollToBottom(); } catch (e) { /* ignore */ }
                }
            }
        } catch (err) {
            console.error('[Roles] Error refreshing chat messages:', err);
        }

        // Hide AI processing status after roles and chat are updated
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        showNotification('AI operation completed', 'success');

    } catch (error) {
        console.error('[Roles] Error processing AI operation:', error);

        // Hide processing status on error
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

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
