// Goals functionality
// Handles goal definition management

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';
import { loadEntityList, showEntityEditor, cancelEntityEdit, deleteEntity, saveEntity } from './crud-helpers.js';

// Helper function to determine status message based on operations
function getStatusMessage(operations, selectedCount = 0) {
    if (operations.length === 1) {
        switch (operations[0]) {
            case 'create':
                return 'AI is generating goals...';
            case 'enhance':
                return selectedCount > 0
                    ? `AI is enhancing ${selectedCount} selected goal(s)...`
                    : 'AI is enhancing your existing goals...';
            case 'consolidate':
                return selectedCount > 0
                    ? `AI is consolidating ${selectedCount} selected goal(s) into a lean, strategic list...`
                    : 'AI is consolidating your goals into a lean, strategic list...';
            default:
                return 'AI is processing your request...';
        }
    } else if (operations.length > 1) {
        return `AI is performing ${operations.length} operations on your goals...`;
    }
    return 'AI is processing your request...';
}

// Helper function to reload chat messages
async function reloadChatMessages() {
    const agencyId = getCurrentAgencyId();
    const chatContainer = document.getElementById('chat-messages');

    if (!chatContainer || !agencyId) {
        return;
    }

    try {
        const response = await fetch(`/agencies/${agencyId}/chat-messages`);
        if (response.ok) {
            const html = await response.text();
            chatContainer.innerHTML = html;
            try {
                scrollToBottom(chatContainer);
            } catch (e) {
                // Fallback scroll
                chatContainer.scrollTop = chatContainer.scrollHeight;
            }
        }
    } catch (err) {
        console.error('[Goals] Error refreshing chat messages:', err);
    }
}

// Goal editor state management
let goalEditorState = {
    mode: 'add', // 'add' or 'edit'
    goalKey: null,
    originalCode: '',
    originalDescription: ''
};

// Load goals list
export function loadGoals() {
    return loadEntityList('goals', 'goals-table-body', 3);
}

// Show goal editor
export function showGoalEditor(mode, goalKey = null, code = '', description = '') {
    goalEditorState.mode = mode;
    goalEditorState.goalKey = goalKey;
    goalEditorState.originalCode = code;
    goalEditorState.originalDescription = description;

    const addTitle = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Goal</span>';
    const editTitle = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Goal</span>';

    showEntityEditor(mode, 'goal-editor-card', 'goals-list-card', 'goal-editor-title', addTitle, editTitle, 'goal-description-editor');

    // Set field values
    const codeEditor = document.getElementById('goal-code-editor');
    const descriptionEditor = document.getElementById('goal-description-editor');

    if (codeEditor) codeEditor.value = code;
    if (descriptionEditor) descriptionEditor.value = description;
}

// Save goal from editor
export function saveGoalFromEditor() {
    const code = document.getElementById('goal-code-editor')?.value.trim();
    const description = document.getElementById('goal-description-editor')?.value.trim();

    if (!code) {
        showNotification('Please enter a goal code', 'warning');
        document.getElementById('goal-code-editor')?.focus();
        return;
    }

    if (!description) {
        showNotification('Please enter a goal description', 'warning');
        document.getElementById('goal-description-editor')?.focus();
        return;
    }

    const data = { code, description };

    saveEntity('goals', goalEditorState.mode, goalEditorState.goalKey, data, 'save-goal-btn', () => {
        cancelGoalEdit();
        loadGoals();
    });
}

// Cancel goal edit
export function cancelGoalEdit() {
    cancelEntityEdit('goal-editor-card', 'goals-list-card', ['goal-code-editor', 'goal-description-editor']);

    // Reset state
    goalEditorState = {
        mode: 'add',
        goalKey: null,
        originalCode: '',
        originalDescription: ''
    };
}

// Delete goal
export function deleteGoal(goalKey, goalNumber) {
    deleteEntity('goals', goalKey, `goal #${goalNumber}`, loadGoals);
}

// Process AI Goal Operation - Direct operation without modal
export function processAIGoalOperation(operations, userRequest = '') {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Validate operations array
    if (!operations || operations.length === 0) {
        showNotification('Error: No operation specified', 'error');
        return;
    }

    // Get selected goal keys for enhance/consolidate operations
    const selectedGoalKeys = (operations.includes('enhance') || operations.includes('consolidate'))
        ? getSelectedGoalKeys()
        : [];

    // Validate selection for enhance/consolidate
    if ((operations.includes('enhance') || operations.includes('consolidate')) && selectedGoalKeys.length === 0) {
        showNotification('Please select goals first', 'warning');
        return;
    }

    // Show AI processing status
    const statusMessage = getStatusMessage(operations, selectedGoalKeys.length);
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    // Build request body
    const requestBody = {
        operations: operations,
        goal_keys: selectedGoalKeys
    };

    // Add user request if provided
    if (userRequest && userRequest.trim() !== '') {
        requestBody.user_request = userRequest.trim();
    }

    // Call AI endpoint with operations and selected goal keys
    fetch(`/api/v1/agencies/${agencyId}/goals/ai-process`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to process AI goal request');
            }
            return response.json();
        })
        .then(async data => {
            // Update status to show we're processing results
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('Processing results and updating goals...');
            }

            // Reload goals list to show changes
            await loadGoals();

            // Clear selections and update buttons after reload
            document.querySelectorAll('.goal-checkbox:checked').forEach(cb => cb.checked = false);
            updateGoalSelectionButtons();

            // Reload chat messages to show the AI explanation
            await reloadChatMessages();

            // Hide AI process status after goals and chat are updated
            if (window.hideAIProcessStatus) {
                window.hideAIProcessStatus();
            }

            // Show success message with what was done
            const operationText = operations.map(op => {
                switch (op) {
                    case 'create': return 'created new goals';
                    case 'enhance': return 'enhanced existing goals';
                    case 'consolidate': return 'consolidated goals';
                    default: return op;
                }
            }).join(', ');

            showNotification(`AI successfully ${operationText}!`, 'success');
        })
        .catch(error => {
            console.error('Error processing AI goal operation:', error);

            // Hide AI process status
            if (window.hideAIProcessStatus) {
                window.hideAIProcessStatus();
            }

            showNotification('Failed to process AI goal operation. Please try again.', 'error');
        });
}

// Make functions globally available
window.processAIGoalOperation = processAIGoalOperation;

// Goal selection management
function getSelectedGoalKeys() {
    const checkboxes = document.querySelectorAll('.goal-checkbox:checked');
    return Array.from(checkboxes).map(cb => cb.dataset.goalKey);
}

function updateGoalSelectionButtons() {
    const selectedKeys = getSelectedGoalKeys();
    const hasSelection = selectedKeys.length > 0;

    // Update "Select All" checkbox state
    const selectAllCheckbox = document.getElementById('select-all-goals');
    const allCheckboxes = document.querySelectorAll('.goal-checkbox');
    if (selectAllCheckbox && allCheckboxes.length > 0) {
        const allChecked = Array.from(allCheckboxes).every(cb => cb.checked);
        const someChecked = Array.from(allCheckboxes).some(cb => cb.checked);
        selectAllCheckbox.checked = allChecked;
        selectAllCheckbox.indeterminate = someChecked && !allChecked;
    }

    // Enable/disable Enhance and Consolidate buttons
    const enhanceBtn = document.getElementById('ai-enhance-goals-btn');
    const consolidateBtn = document.getElementById('ai-consolidate-goals-btn');

    if (enhanceBtn) {
        if (hasSelection) {
            enhanceBtn.disabled = false;
            enhanceBtn.classList.remove('is-static');
            enhanceBtn.title = `Enhance ${selectedKeys.length} selected goal(s)`;
        } else {
            enhanceBtn.disabled = true;
            enhanceBtn.classList.add('is-static');
            enhanceBtn.title = 'Select goals to enhance';
        }
    }

    if (consolidateBtn) {
        if (hasSelection) {
            consolidateBtn.disabled = false;
            consolidateBtn.classList.remove('is-static');
            consolidateBtn.title = `Consolidate ${selectedKeys.length} selected goal(s)`;
        } else {
            consolidateBtn.disabled = true;
            consolidateBtn.classList.add('is-static');
            consolidateBtn.title = 'Select goals to consolidate';
        }
    }

    // Update selection count display
    updateSelectionCount(selectedKeys.length);
}

function toggleAllGoals(checked) {
    const checkboxes = document.querySelectorAll('.goal-checkbox');
    checkboxes.forEach(cb => {
        cb.checked = checked;
    });
    updateGoalSelectionButtons();
}

function updateSelectionCount(count) {
    const countDisplay = document.getElementById('goal-selection-count');

    if (countDisplay) {
        if (count > 0) {
            countDisplay.textContent = `${count} selected`;
            countDisplay.style.display = 'inline-block';
        } else {
            countDisplay.style.display = 'none';
        }
    }
}

// Filter goals by search text
export function filterGoals() {
    const searchInput = document.getElementById('goal-search')?.value.toLowerCase() || '';
    const tbody = document.getElementById('goals-table-body');
    if (!tbody) return;

    const rows = tbody.querySelectorAll('.table-item');

    rows.forEach(row => {
        const code = row.querySelector('.goal-code')?.textContent.toLowerCase() || '';
        const description = row.querySelector('.goal-description')?.textContent.toLowerCase() || '';

        const matchesSearch = !searchInput || code.includes(searchInput) || description.includes(searchInput);

        if (matchesSearch) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// Refine goal description with AI
export function refineGoalDescription() {
    const description = document.getElementById('goal-description-editor')?.value.trim();

    if (!description) {
        showNotification('Please enter a description first', 'warning');
        return;
    }

    showNotification('AI refinement for goal descriptions coming soon!', 'info');
}

// Validate goal code format
export function validateGoalCode(code) {
    if (!code || code.trim().length === 0) {
        return { valid: false, error: 'Goal code cannot be empty' };
    }

    // Goal codes should follow a pattern like G001, G002, etc.
    const pattern = /^G\d{3}$/;
    if (!pattern.test(code)) {
        return {
            valid: false,
            error: 'Goal code should follow format G### (e.g., G001, G002)'
        };
    }

    return { valid: true };
}

// Initialize button states on page load
document.addEventListener('DOMContentLoaded', function () {
    updateGoalSelectionButtons();
});

// Make functions available globally
window.loadGoals = loadGoals;
window.showGoalEditor = showGoalEditor;
window.saveGoalFromEditor = saveGoalFromEditor;
window.cancelGoalEdit = cancelGoalEdit;
window.deleteGoal = deleteGoal;
window.filterGoals = filterGoals;
window.processAIGoalOperation = processAIGoalOperation;
window.refineGoalDescription = refineGoalDescription;
window.validateGoalCode = validateGoalCode;
window.getSelectedGoalKeys = getSelectedGoalKeys;
window.updateGoalSelectionButtons = updateGoalSelectionButtons;
window.toggleAllGoals = toggleAllGoals;
