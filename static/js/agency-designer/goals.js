// Goals functionality
// Handles goal definition management

// Functions available from global window namespace
// getCurrentAgencyId, showNotification from utils.js
// specificationAPI from specification-api.js
// loadEntityList, showEntityEditor, etc. from crud-helpers.js

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
    const agencyId = window.getCurrentAgencyId();
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
window.loadGoals = function () {
    return window.loadEntityList('goals', 'goals-table-body', 4);
}

// Show goal editor
window.showGoalEditor = function (mode, goalKey = null, code = '', description = '') {
    goalEditorState.mode = mode;
    goalEditorState.goalKey = goalKey;
    goalEditorState.originalCode = code;
    goalEditorState.originalDescription = description;

    const addTitle = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Goal</span>';
    const editTitle = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Goal</span>';

    window.showEntityEditor(mode, 'goal-editor-card', 'goals-list-card', 'goal-editor-title', addTitle, editTitle, 'goal-description-editor');

    // Set field values
    const codeEditor = document.getElementById('goal-code-editor');
    const descriptionEditor = document.getElementById('goal-description-editor');

    if (codeEditor) codeEditor.value = code;
    if (descriptionEditor) descriptionEditor.value = description;
}

// Save goal from editor
window.saveGoalFromEditor = function () {
    const code = document.getElementById('goal-code-editor')?.value.trim();
    const description = document.getElementById('goal-description-editor')?.value.trim();

    if (!code) {
        window.showNotification('Please enter a goal code', 'warning');
        document.getElementById('goal-code-editor')?.focus();
        return;
    }

    if (!description) {
        window.showNotification('Please enter a goal description', 'warning');
        document.getElementById('goal-description-editor')?.focus();
        return;
    }

    const data = { code, description };

    window.saveEntity('goals', goalEditorState.mode, goalEditorState.goalKey, data, 'save-goal-btn', () => {
        cancelGoalEdit();
        loadGoals();
    });
}

// Cancel goal edit
window.cancelGoalEdit = function () {
    window.cancelEntityEdit('goal-editor-card', 'goals-list-card', ['goal-code-editor', 'goal-description-editor']);

    // Reset state
    goalEditorState = {
        mode: 'add',
        goalKey: null,
        originalCode: '',
        originalDescription: ''
    };
}

// Delete goal
window.deleteGoal = function (goalKey) {
    // Use generic "goal" as display name since we're using _key (not user-friendly)
    const displayName = `goal`;
    window.deleteEntity('goals', goalKey, displayName, loadGoals);
}

// Process AI Goal Operation - Direct operation without modal
window.processAIGoalOperation = function (operations, userRequest = '') {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    // Validate operations array
    if (!operations || operations.length === 0) {
        window.showNotification('Error: No operation specified', 'error');
        return;
    }

    // Get selected goal keys for enhance/consolidate operations
    const selectedGoalKeys = (operations.includes('enhance') || operations.includes('consolidate'))
        ? getSelectedGoalKeys()
        : [];

    // Validate selection for enhance/consolidate
    if ((operations.includes('enhance') || operations.includes('consolidate')) && selectedGoalKeys.length === 0) {
        window.showNotification('Please select goals first', 'warning');
        return;
    }

    // Show AI processing status
    const statusMessage = getStatusMessage(operations, selectedGoalKeys.length);
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    // Determine endpoint based on operation
    let endpoint;
    let requestBody;

    if (operations.includes('create')) {
        // Use generate endpoint for creating new goals
        endpoint = `/api/v1/agencies/${agencyId}/goals/generate`;
        requestBody = {
            userInput: userRequest || "Generate 3-5 strategic goals based on the agency's introduction and purpose"
        };
    } else if (operations.includes('consolidate')) {
        // Use consolidate endpoint
        endpoint = `/api/v1/agencies/${agencyId}/goals/consolidate`;
        requestBody = {}; // Consolidate endpoint uses preset prompt
    } else if (operations.includes('enhance')) {
        // Use refine-dynamic endpoint for enhance
        endpoint = `/api/v1/agencies/${agencyId}/goals/refine-dynamic`;
        requestBody = {
            user_message: userRequest || "Enhance and improve the selected goals to be clearer, more specific, and better aligned with the agency's purpose",
            goal_keys: selectedGoalKeys
        };
    } else {
        // Default to refine-dynamic for other operations
        endpoint = `/api/v1/agencies/${agencyId}/goals/refine-dynamic`;
        requestBody = {
            user_message: userRequest || "Process the goals based on the context",
            goal_keys: selectedGoalKeys
        };
    }

    // Call AI endpoint
    fetch(endpoint, {
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
            document.querySelectorAll('#goals-table-body input[type="checkbox"]:checked').forEach(cb => cb.checked = false);
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

            window.showNotification(`AI successfully ${operationText}!`, 'success');
        })
        .catch(error => {

            // Hide AI process status
            if (window.hideAIProcessStatus) {
                window.hideAIProcessStatus();
            }

            window.showNotification('Failed to process AI goal operation. Please try again.', 'error');
        });
}

// Make functions globally available
window.processAIGoalOperation = processAIGoalOperation;

// Goal selection management
window.getSelectedGoalKeys = function () {
    const checkboxes = document.querySelectorAll('#goals-table-body input[type="checkbox"]:checked');
    return Array.from(checkboxes).map(cb => cb.value);
}

window.updateGoalSelectionButtons = function () {
    const selectedKeys = window.getSelectedGoalKeys();
    const hasSelection = selectedKeys.length > 0;

    // Update "Select All" checkbox state
    const selectAllCheckbox = document.getElementById('select-all-goals');
    const allCheckboxes = document.querySelectorAll('#goals-table-body input[type="checkbox"]');
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

window.toggleAllGoals = function (checked) {
    const checkboxes = document.querySelectorAll('#goals-table-body input[type="checkbox"]');
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
window.filterGoals = function () {
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
window.refineGoalDescription = function () {
    const description = document.getElementById('goal-description-editor')?.value.trim();

    if (!description) {
        window.showNotification('Please enter a description first', 'warning');
        return;
    }

    window.showNotification('AI refinement for goal descriptions coming soon!', 'info');
}

// Validate goal code format
window.validateGoalCode = function (code) {
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

// Attach context clear listener to goals table
// Similar to introduction.js, but for table-based editing
function attachGoalsContextClearListener() {
    const goalsTable = document.getElementById('goals-table-body');
    if (!goalsTable) {
        return;
    }

    // Check if already attached
    if (goalsTable._contextClearAttached) {
        return;
    }

    // Use MutationObserver to detect changes to the goals table
    const observer = new MutationObserver(function (mutations) {
        // Check if there are actual goal row changes
        const hasGoalChanges = mutations.some(mutation => {
            return mutation.type === 'childList' ||
                (mutation.type === 'characterData' && mutation.target.parentElement);
        });

        if (!hasGoalChanges) {
            return;
        }

        if (!window.ContextManager) {
            return;
        }

        const contexts = window.ContextManager.getAllContexts();
        const selections = window.ContextManager.getSelections();

        const hasContextsOrSelections = (contexts && contexts.length > 0) || (selections && selections.length > 0);

        if (hasContextsOrSelections) {

            // Clear both contexts and selections
            window.ContextManager.clearAllContexts();
            window.ContextManager.clearSelections();

        }
    });

    // Observe the table for changes
    observer.observe(goalsTable, {
        childList: true,
        subtree: true,
        characterData: true
    });

    // Mark as attached
    goalsTable._contextClearAttached = true;
}

// Initialize button states on page load
document.addEventListener('DOMContentLoaded', function () {
    updateGoalSelectionButtons();
    attachGoalsContextClearListener();
});

// Functions are already assigned to window namespace above
