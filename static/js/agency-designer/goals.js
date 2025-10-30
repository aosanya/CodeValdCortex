// Goals functionality
// Handles goal definition management

import { getCurrentAgencyId, showNotification } from './utils.js';

// Goal editor state management
let goalEditorState = {
    mode: 'add', // 'add' or 'edit'
    goalKey: null,
    originalCode: '',
    originalDescription: ''
};

// Load goals list
export function loadGoals() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const goalsTableBody = document.getElementById('goals-table-body');
    if (!goalsTableBody) {
        console.error('Goals table body not found');
        return;
    }

    // Show loading state
    goalsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading goals...</p></td></tr>';

    // Fetch goals HTML from API
    fetch(`/api/v1/agencies/${agencyId}/goals/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load goals');
            }
            return response.text();
        })
        .then(html => {
            goalsTableBody.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading goals:', error);
            goalsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-danger has-text-centered py-5"><p>Error loading goals</p></td></tr>';
        });
}

// Show goal editor
export function showGoalEditor(mode, goalKey = null, code = '', description = '') {
    goalEditorState.mode = mode;
    goalEditorState.goalKey = goalKey;
    goalEditorState.originalCode = code;
    goalEditorState.originalDescription = description;

    const editorCard = document.getElementById('goal-editor-card');
    const listCard = document.getElementById('goals-list-card');
    const editorTitle = document.getElementById('goal-editor-title');
    const codeEditor = document.getElementById('goal-code-editor');
    const descriptionEditor = document.getElementById('goal-description-editor');

    if (!editorCard || !listCard || !editorTitle || !codeEditor || !descriptionEditor) {
        console.error('Goal editor elements not found');
        return;
    }

    // Update editor title and content
    if (mode === 'add') {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Goal</span>';
        codeEditor.value = '';
        descriptionEditor.value = '';
    } else {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Goal</span>';
        codeEditor.value = code;
        descriptionEditor.value = description;
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on editor
    descriptionEditor.focus();
}

// Save goal from editor
export function saveGoalFromEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const codeEditor = document.getElementById('goal-code-editor');
    const descriptionEditor = document.getElementById('goal-description-editor');
    if (!codeEditor || !descriptionEditor) {
        console.error('Editor elements not found');
        return;
    }

    const code = codeEditor.value.trim();
    const description = descriptionEditor.value.trim();

    if (!code) {
        showNotification('Please enter a goal code', 'warning');
        codeEditor.focus();
        return;
    }

    if (!description) {
        showNotification('Please enter a goal description', 'warning');
        descriptionEditor.focus();
        return;
    }

    const saveBtn = document.getElementById('save-goal-btn');
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
    }

    const isAddMode = goalEditorState.mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/goals`
        : `/api/v1/agencies/${agencyId}/goals/${goalEditorState.goalKey}`;
    const method = isAddMode ? 'POST' : 'PUT';

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            code: code,
            description: description
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} goal`);
            }
            return response.json();
        })
        .then(() => {
            showNotification(`Goal ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelGoalEdit(); // Hide editor
            loadGoals(); // Reload the list
        })
        .catch(error => {
            console.error(`Error ${isAddMode ? 'creating' : 'updating'} goal:`, error);
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} goal`, 'error');
        })
        .finally(() => {
            if (saveBtn) {
                saveBtn.classList.remove('is-loading');
            }
        });
}

// Cancel goal edit
export function cancelGoalEdit() {
    const editorCard = document.getElementById('goal-editor-card');
    const listCard = document.getElementById('goals-list-card');
    const codeEditor = document.getElementById('goal-code-editor');
    const descriptionEditor = document.getElementById('goal-description-editor');

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');
    if (codeEditor) codeEditor.value = '';
    if (descriptionEditor) descriptionEditor.value = '';

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
    if (!confirm(`Are you sure you want to delete goal #${goalNumber}? This will renumber all subsequent goals.`)) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }


    fetch(`/api/v1/agencies/${agencyId}/goals/${goalKey}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete goal');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Goal deleted successfully!', 'success');
            loadGoals(); // Reload the list
        })
        .catch(error => {
            console.error('Error deleting goal:', error);
            showNotification('Error deleting goal', 'error');
        });
}

// Generate goal using AI
export function showGenerateGoalModal() {
    // Create a modal for AI goal generation with checkbox options
    const modal = document.createElement('div');
    modal.className = 'modal is-active';
    modal.id = 'ai-generate-modal';
    modal.innerHTML = `
        <div class="modal-background" onclick="closeGenerateGoalModal()"></div>
        <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">
                    <span class="icon has-text-info">
                        <i class="fas fa-magic"></i>
                    </span>
                    <span>AI Goal Assistant</span>
                </p>
                <button class="delete" aria-label="close" onclick="closeGenerateGoalModal()"></button>
            </header>
            <section class="modal-card-body">
                <div class="content">
                    <p class="mb-4">Select what you'd like the AI to help with:</p>
                    
                    <div class="field">
                        <label class="checkbox box has-background-light mb-3 p-4" style="display: block; cursor: pointer;">
                            <div class="is-flex is-align-items-center">
                                <input type="checkbox" id="option-create" class="mr-3" checked>
                                <span class="icon is-large has-text-success mr-3">
                                    <i class="fas fa-plus-circle fa-2x"></i>
                                </span>
                                <div>
                                    <p class="has-text-weight-bold mb-1">Create New Goals</p>
                                    <p class="is-size-7 has-text-grey mb-0">AI will analyze the introduction and create structured goals</p>
                                </div>
                            </div>
                        </label>
                    </div>
                    
                    <div class="field">
                        <label class="checkbox box has-background-light mb-3 p-4" style="display: block; cursor: pointer;">
                            <div class="is-flex is-align-items-center">
                                <input type="checkbox" id="option-enhance" class="mr-3">
                                <span class="icon is-large has-text-info mr-3">
                                    <i class="fas fa-lightbulb fa-2x"></i>
                                </span>
                                <div>
                                    <p class="has-text-weight-bold mb-1">Enhance Existing Goals</p>
                                    <p class="is-size-7 has-text-grey mb-0">AI will suggest improvements to existing goals</p>
                                </div>
                            </div>
                        </label>
                    </div>
                    
                    <div class="field">
                        <label class="checkbox box has-background-light p-4" style="display: block; cursor: pointer;">
                            <div class="is-flex is-align-items-center">
                                <input type="checkbox" id="option-consolidate" class="mr-3">
                                <span class="icon is-large has-text-warning mr-3">
                                    <i class="fas fa-layer-group fa-2x"></i>
                                </span>
                                <div>
                                    <p class="has-text-weight-bold mb-1">Consolidate Goals</p>
                                    <p class="is-size-7 has-text-grey mb-0">AI will suggest consolidations or reorganization</p>
                                </div>
                            </div>
                        </label>
                    </div>
                    
                    <div id="ai-generate-status" class="notification is-info is-light mt-4" style="display: none;">
                        <div class="is-flex is-align-items-center">
                            <span class="icon has-text-info mr-2">
                                <i class="fas fa-spinner fa-spin"></i>
                            </span>
                            <span>AI is processing your request...</span>
                        </div>
                    </div>
                </div>
            </section>
            <footer class="modal-card-foot">
                <button class="button is-success" onclick="processAIGoalRequest()" id="process-ai-btn">
                    <span class="icon"><i class="fas fa-magic"></i></span>
                    <span>Generate with AI</span>
                </button>
                <button class="button" onclick="closeGenerateGoalModal()">Cancel</button>
            </footer>
        </div>
    `;
    document.body.appendChild(modal);
}

// Close the generate goal modal
export function closeGenerateGoalModal() {
    const modal = document.getElementById('ai-generate-modal');
    if (modal) {
        modal.remove();
    }
}

// Process AI goal request with selected options
export function processAIGoalRequest() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Get selected options
    const createChecked = document.getElementById('option-create')?.checked || false;
    const enhanceChecked = document.getElementById('option-enhance')?.checked || false;
    const consolidateChecked = document.getElementById('option-consolidate')?.checked || false;

    // Validate at least one option is selected
    if (!createChecked && !enhanceChecked && !consolidateChecked) {
        showNotification('Please select at least one option', 'warning');
        return;
    }

    // Close the modal immediately
    closeGenerateGoalModal();

    // Show AI processing status in the chat area
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus('AI is generating goals from your introduction...');
    }

    // Build request with selected operations
    const operations = [];
    if (createChecked) operations.push('create');
    if (enhanceChecked) operations.push('enhance');
    if (consolidateChecked) operations.push('consolidate');

    // Call AI endpoint with consolidated request
    fetch(`/api/v1/agencies/${agencyId}/goals/ai-process`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            operations: operations
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to process AI goal request');
            }
            return response.json();
        })
        .then(data => {
            // Hide AI processing status
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

            // Reload goals list to show changes
            loadGoals();
        })
        .catch(error => {
            console.error('Error processing AI goal request:', error);
            
            // Hide AI processing status
            if (window.hideAIProcessStatus) {
                window.hideAIProcessStatus();
            }
            
            showNotification('Error processing AI goal request', 'error');
        });
}

// Handle AI goal option selection (deprecated - kept for compatibility)
export function selectAIGoalOption(option) {
    closeGenerateGoalModal();

    switch (option) {
        case 'create':
            showCreateGoalDialog();
            break;
        case 'enhance':
            enhanceExistingGoals();
            break;
        case 'consolidate':
            consolidateGoals();
            break;
    }
}

// Show dialog for creating a new goal
function showCreateGoalDialog() {
    const modal = document.createElement('div');
    modal.className = 'modal is-active';
    modal.id = 'ai-create-goal-modal';
    modal.innerHTML = `
        <div class="modal-background" onclick="closeCreateGoalDialog()"></div>
        <div class="modal-card">
            <header class="modal-card-head">
                <p class="modal-card-title">
                    <span class="icon has-text-success">
                        <i class="fas fa-plus-circle"></i>
                    </span>
                    <span>Create New Goal with AI</span>
                </p>
                <button class="delete" aria-label="close" onclick="closeCreateGoalDialog()"></button>
            </header>
            <section class="modal-card-body">
                <div class="field">
                    <label class="label">Describe the goal you want to create</label>
                    <div class="control">
                        <textarea
                            class="textarea"
                            id="ai-goal-input"
                            placeholder="Example: We need to improve customer response times and communication quality..."
                            rows="5"></textarea>
                    </div>
                    <p class="help">Describe the goal in natural language. The AI will structure it with proper scope and metrics.</p>
                </div>
                <div id="ai-generate-status" class="notification is-info is-light" style="display: none;">
                    <div class="is-flex is-align-items-center">
                        <span class="icon has-text-info mr-2">
                            <i class="fas fa-spinner fa-spin"></i>
                        </span>
                        <span>AI is generating your goal...</span>
                    </div>
                </div>
            </section>
            <footer class="modal-card-foot">
                <button class="button is-success" onclick="generateGoalWithAI()" id="generate-ai-btn">
                    <span class="icon"><i class="fas fa-magic"></i></span>
                    <span>Generate Goal</span>
                </button>
                <button class="button" onclick="closeCreateGoalDialog()">Cancel</button>
            </footer>
        </div>
    `;
    document.body.appendChild(modal);

    setTimeout(() => {
        document.getElementById('ai-goal-input')?.focus();
    }, 100);
}

function closeCreateGoalDialog() {
    const modal = document.getElementById('ai-create-goal-modal');
    if (modal) {
        modal.remove();
    }
}

// Enhance existing goals based on introduction
function enhanceExistingGoals() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    showNotification('AI is analyzing your introduction and goals...', 'info');

    // TODO: Implement enhance existing goals endpoint
    // This would call an API that:
    // 1. Gets the introduction
    // 2. Gets existing goals
    // 3. AI suggests enhancements to each goal
    // 4. Returns suggestions for review

    setTimeout(() => {
        showNotification('Feature coming soon! AI will analyze your introduction and suggest goal improvements.', 'info');
    }, 1000);
}

// Consolidate goals
function consolidateGoals() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    showNotification('AI is analyzing your goals for consolidation opportunities...', 'info');

    // TODO: Implement consolidate goals endpoint
    // This would call an API that:
    // 1. Gets all existing goals
    // 2. AI analyzes for overlaps and consolidation opportunities
    // 3. Suggests which goals to merge or reorganize
    // 4. Returns consolidation plan for review

    setTimeout(() => {
        showNotification('Feature coming soon! AI will analyze your goals and suggest consolidations.', 'info');
    }, 1000);
}

// Generate goal with AI
export function generateGoalWithAI() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const input = document.getElementById('ai-goal-input');
    if (!input) return;

    const userInput = input.value.trim();
    if (!userInput) {
        showNotification('Please describe the goal you want to create', 'warning');
        input.focus();
        return;
    }

    const generateBtn = document.getElementById('generate-ai-btn');
    const status = document.getElementById('ai-generate-status');

    if (generateBtn) generateBtn.classList.add('is-loading');
    if (status) status.style.display = 'block';

    fetch(`/api/v1/agencies/${agencyId}/goals/generate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            userInput: userInput
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to generate goal');
            }
            return response.json();
        })
        .then(data => {
            if (generateBtn) generateBtn.classList.remove('is-loading');
            if (status) status.style.display = 'none';

            closeGenerateGoalModal();
            showNotification('Goal generated successfully! Opening editor to review...', 'success');


            // Open the editor with the generated goal data
            if (data.goal) {
                showGoalEditor('edit', data.goal._key, data.goal.code, data.goal.description);
            } else {
                // Fallback: just reload goals list
                loadGoals();
            }
        })
        .catch(error => {
            console.error('Error generating goal:', error);
            if (generateBtn) generateBtn.classList.remove('is-loading');
            if (status) status.style.display = 'none';
            showNotification('Error generating goal with AI', 'error');
        });
}

// Make functions globally available
window.showGenerateGoalModal = showGenerateGoalModal;
window.closeGenerateGoalModal = closeGenerateGoalModal;
window.processAIGoalRequest = processAIGoalRequest;
window.selectAIGoalOption = selectAIGoalOption;
window.closeCreateGoalDialog = closeCreateGoalDialog;
window.generateGoalWithAI = generateGoalWithAI;
