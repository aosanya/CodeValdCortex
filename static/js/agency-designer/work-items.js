// Work Items functionality
// Handles work items management

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';
import { loadEntityList, showEntityEditor, cancelEntityEdit, deleteEntity, saveEntity, populateForm, clearForm } from './crud-helpers.js';

// Work item editor state management
let workItemEditorState = {
    mode: 'add', // 'add' or 'edit'
    workItemKey: null,
    originalData: {}
};

// Load goals for the checkbox list
async function loadGoalsForSelection() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) return;

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/goals`);
        if (!response.ok) throw new Error('Failed to fetch goals');

        const goals = await response.json();
        const container = document.getElementById('work-item-goals-editor');
        if (!container) return;

        // Clear existing content
        container.innerHTML = '';

        if (goals.length === 0) {
            container.innerHTML = '<p class="has-text-grey-light">No goals available. Create goals first.</p>';
            return;
        }

        // Add goals as checkboxes
        goals.forEach(goal => {
            const label = document.createElement('label');
            label.className = 'checkbox is-block mb-2';

            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = goal._key || goal.key;
            checkbox.className = 'mr-2';
            checkbox.dataset.goalKey = goal._key || goal.key;

            const text = document.createTextNode(` ${goal.code} - ${goal.description.substring(0, 60)}${goal.description.length > 60 ? '...' : ''}`);

            label.appendChild(checkbox);
            label.appendChild(text);
            container.appendChild(label);
        });
    } catch (error) {
        console.error('Error loading goals:', error);
    }
}

// Load work items list
export function loadWorkItems() {
    return loadEntityList('work-items', 'work-items-table-body', 4);
}

// Show work item editor
export async function showWorkItemEditor(mode, workItemKey = null) {
    workItemEditorState.mode = mode;
    workItemEditorState.workItemKey = workItemKey;

    showEntityEditor(
        mode,
        'work-item-editor-card',
        'work-items-list-card',
        'work-item-editor-title',
        'Add New Work Item',
        'Edit Work Item',
        'work-item-title-editor'
    );

    // Load available goals for the dropdown first
    await loadGoalsForSelection();

    if (mode === 'add') {
        clearWorkItemForm();
    } else if (mode === 'edit') {
        await loadWorkItemData(workItemKey);
    }
}

// Load work item data for editing
async function loadWorkItemData(workItemKey) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) {
        return;
    }

    try {
        // Fetch work item data
        const response = await fetch(`/api/v1/agencies/${agencyId}/work-items`);
        const workItems = await response.json();

        // The key field comes as "_key" from JSON
        const workItem = workItems.find(wi => wi._key === workItemKey || wi.key === workItemKey);

        if (workItem) {
            populateWorkItemForm(workItem);
            workItemEditorState.originalData = workItem;

            // Load linked goals for this work item
            await loadLinkedGoals(workItemKey);
        } else {
            showNotification('Work item not found', 'error');
        }
    } catch (error) {
        console.error('Error loading work item:', error);
        showNotification('Error loading work item data', 'error');
    }
}

// Load and select linked goals for a work item
async function loadLinkedGoals(workItemKey) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) return;

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/work-items/${workItemKey}/goals`);
        if (!response.ok) return; // No links yet, that's OK

        const links = await response.json();
        const container = document.getElementById('work-item-goals-editor');
        if (!container) return;

        // Check the linked goals
        const linkedGoalKeys = links.map(link => link.goal_key);
        const checkboxes = container.querySelectorAll('input[type="checkbox"]');
        checkboxes.forEach(checkbox => {
            checkbox.checked = linkedGoalKeys.includes(checkbox.value);
        });
    } catch (error) {
        console.error('Error loading goal links:', error);
    }
}

// Populate form with work item data
function populateWorkItemForm(workItem) {
    populateForm({
        'work-item-title-editor': workItem.title || '',
        'work-item-description-editor': workItem.description || '',
        'work-item-deliverables-editor': workItem.deliverables ? workItem.deliverables.join('\n') : '',
        'work-item-tags-editor': workItem.tags ? workItem.tags.join(', ') : ''
    });
}

// Clear work item form
function clearWorkItemForm() {
    clearForm([
        'work-item-title-editor',
        'work-item-description-editor',
        'work-item-deliverables-editor',
        'work-item-tags-editor'
    ]);

    // Clear goals checkboxes
    const container = document.getElementById('work-item-goals-editor');
    if (container) {
        const checkboxes = container.querySelectorAll('input[type="checkbox"]');
        checkboxes.forEach(checkbox => {
            checkbox.checked = false;
        });
    }

    workItemEditorState.originalData = {};
}

// Save work item from editor
export async function saveWorkItemFromEditor() {
    // Get form values
    const title = document.getElementById('work-item-title-editor')?.value.trim();
    const description = document.getElementById('work-item-description-editor')?.value.trim();
    const deliverables = document.getElementById('work-item-deliverables-editor')?.value
        .split('\n')
        .map(d => d.trim())
        .filter(d => d.length > 0);
    const tags = document.getElementById('work-item-tags-editor')?.value
        .split(',')
        .map(t => t.trim())
        .filter(t => t.length > 0);

    // Get selected goals from checkboxes
    const container = document.getElementById('work-item-goals-editor');
    const checkboxes = container ? container.querySelectorAll('input[type="checkbox"]:checked') : [];
    const selectedGoals = Array.from(checkboxes).map(cb => cb.value);

    // Validation
    if (!title) {
        showNotification('Please enter a work item title', 'warning');
        document.getElementById('work-item-title-editor')?.focus();
        return;
    }

    if (!description) {
        showNotification('Please enter a work item description', 'warning');
        document.getElementById('work-item-description-editor')?.focus();
        return;
    }

    const data = {
        title,
        description,
        deliverables,
        tags
    };

    // Save the work item first
    try {
        const agencyId = getCurrentAgencyId();
        const url = workItemEditorState.mode === 'add'
            ? `/api/v1/agencies/${agencyId}/work-items`
            : `/api/v1/agencies/${agencyId}/work-items/${workItemEditorState.workItemKey}`;

        const method = workItemEditorState.mode === 'add' ? 'POST' : 'PUT';

        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            throw new Error('Failed to save work item');
        }

        const savedWorkItem = await response.json();
        const workItemKey = savedWorkItem._key || savedWorkItem.key || workItemEditorState.workItemKey;

        // Save goal links
        await saveGoalLinks(workItemKey, selectedGoals);

        showNotification('Work item saved successfully', 'success');
        cancelWorkItemEdit();
        loadWorkItems();
    } catch (error) {
        console.error('Error saving work item:', error);
        showNotification('Error saving work item', 'error');
    }
}

// Save goal links for a work item
async function saveGoalLinks(workItemKey, selectedGoalKeys) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) return;

    try {
        // Delete existing links
        await fetch(`/api/v1/agencies/${agencyId}/work-items/${workItemKey}/goals`, {
            method: 'DELETE'
        });

        // Create new links
        for (const goalKey of selectedGoalKeys) {
            await fetch(`/api/v1/agencies/${agencyId}/work-items/${workItemKey}/goals`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    work_item_key: workItemKey,
                    goal_key: goalKey,
                    relationship: 'addresses'
                })
            });
        }
    } catch (error) {
        console.error('Error saving goal links:', error);
        // Don't throw - work item was saved, links are optional
    }
}

// Cancel work item edit
export function cancelWorkItemEdit() {
    cancelEntityEdit('work-item-editor-card', 'work-items-list-card', [
        'work-item-title-editor',
        'work-item-description-editor',
        'work-item-deliverables-editor',
        'work-item-tags-editor'
    ]);

    // Clear all navigational contexts when returning to work items list
    if (window.ContextManager) {
        window.ContextManager.clearNavigationalContexts();
        console.log('[WorkItems] Cleared navigational contexts when returning to list');
    }

    // Reset state
    workItemEditorState = {
        mode: 'add',
        workItemKey: null,
        originalData: {}
    };
}

// Delete work item
export function deleteWorkItem(workItemKey) {
    deleteEntity('work-items', workItemKey, 'this work item', loadWorkItems);
}

// Filter work items
export function filterWorkItems() {
    const searchInput = document.getElementById('work-item-search')?.value.toLowerCase() || '';
    const typeFilter = document.getElementById('filter-type')?.value || '';
    const tbody = document.getElementById('work-items-tbody');
    if (!tbody) return;

    const rows = tbody.querySelectorAll('.table-item');

    rows.forEach(row => {
        const key = row.dataset.itemKey || '';
        const title = row.querySelector('.has-text-weight-semibold')?.textContent.toLowerCase() || '';

        const matchesSearch = !searchInput || key.toLowerCase().includes(searchInput) || title.includes(searchInput);

        if (matchesSearch) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// AI Work Item Operations
export async function processAIWorkItemOperation(operations) {
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

    // For enhance/consolidate operations, get selected work items
    let selectedWorkItemKeys = [];
    if (operations.includes('enhance') || operations.includes('consolidate')) {
        selectedWorkItemKeys = getSelectedWorkItemKeys();
        if (selectedWorkItemKeys.length === 0) {
            showNotification('Please select work items first', 'warning');
            return;
        }
    }

    let statusMessage = 'AI is processing your request...';
    if (operations.length === 1) {
        switch (operations[0]) {
            case 'create':
                statusMessage = 'AI is generating work items from your goals...';
                break;
            case 'enhance':
                statusMessage = `AI is enhancing ${selectedWorkItemKeys.length} work item(s)...`;
                break;
            case 'consolidate':
                statusMessage = `AI is consolidating ${selectedWorkItemKeys.length} work item(s)...`;
                break;
        }
    } else if (operations.length > 1) {
        statusMessage = `AI is performing ${operations.length} operations on your work items...`;
    }

    // Show AI processing status in the chat area
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    try {
        const requestBody = { operations };

        // Include selected work item keys if applicable
        if (selectedWorkItemKeys.length > 0) {
            requestBody.work_item_keys = selectedWorkItemKeys;
        }

        const response = await fetch(`/api/v1/agencies/${agencyId}/work-items/ai-process`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            throw new Error(`Failed to process AI work item operations: ${response.statusText}`);
        }

        const data = await response.json();

        // Update status to show we're processing results
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('Processing results and updating work items...');
        }

        // Reload work items to show updates
        await loadWorkItems();

        // After work items are reloaded, refresh chat messages so AI explanation appears in the chat
        try {
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                const chatResp = await fetch(`/agencies/${agencyId}/chat-messages`);
                if (chatResp.ok) {
                    const chatHtml = await chatResp.text();
                    chatContainer.innerHTML = chatHtml;
                    // Scroll to bottom to show latest assistant message
                    try { scrollToBottom(chatContainer); } catch (e) { /* ignore */ }
                }
            }
        } catch (err) {
        }

        // Hide AI processing status after work items and chat are updated
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        // Show success message with results
        if (data.created_count > 0) {
            showNotification(`Successfully created ${data.created_count} work items!`, 'success');
        } else if (data.enhanced_count > 0) {
            showNotification(`Successfully enhanced ${data.enhanced_count} work items!`, 'success');
        } else if (data.results && data.results.consolidate_success) {
            showNotification(data.results.consolidate_success, 'success');
        } else {
            showNotification('AI operations completed!', 'success');
        }

    } catch (error) {
        // Hide AI processing status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        showNotification(`AI processing failed: ${error.message}`, 'danger');
    }
}

// Refine work item description with AI
export function refineWorkItemDescription() {
    const description = document.getElementById('work-item-description-editor')?.value.trim();

    if (!description) {
        showNotification('Please enter a description first', 'warning');
        return;
    }

    showNotification('AI refinement for work items coming soon!', 'info');
}

// Get selected work item keys from checkboxes
function getSelectedWorkItemKeys() {
    const checkboxes = document.querySelectorAll('.work-item-checkbox:checked');
    return Array.from(checkboxes).map(cb => cb.dataset.workItemKey);
}

// Update selection buttons based on checkbox state
function updateWorkItemSelectionButtons() {
    const selectedKeys = getSelectedWorkItemKeys();
    const hasSelection = selectedKeys.length > 0;

    // Update "Select All" checkbox state
    const selectAllCheckbox = document.getElementById('select-all-work-items');
    const allCheckboxes = document.querySelectorAll('.work-item-checkbox');
    if (selectAllCheckbox && allCheckboxes.length > 0) {
        const allChecked = Array.from(allCheckboxes).every(cb => cb.checked);
        const someChecked = Array.from(allCheckboxes).some(cb => cb.checked);
        selectAllCheckbox.checked = allChecked;
        selectAllCheckbox.indeterminate = someChecked && !allChecked;
    }

    // Enable/disable Enhance and Consolidate buttons
    const enhanceBtn = document.getElementById('ai-enhance-work-items-btn');
    const consolidateBtn = document.getElementById('ai-consolidate-work-items-btn');

    if (enhanceBtn) {
        if (hasSelection) {
            enhanceBtn.disabled = false;
            enhanceBtn.classList.remove('is-static');
            enhanceBtn.title = `Enhance ${selectedKeys.length} selected work item(s)`;
        } else {
            enhanceBtn.disabled = true;
            enhanceBtn.classList.add('is-static');
            enhanceBtn.title = 'Select work items to enhance';
        }
    }

    if (consolidateBtn) {
        if (hasSelection) {
            consolidateBtn.disabled = false;
            consolidateBtn.classList.remove('is-static');
            consolidateBtn.title = `Consolidate ${selectedKeys.length} selected work item(s)`;
        } else {
            consolidateBtn.disabled = true;
            consolidateBtn.classList.add('is-static');
            consolidateBtn.title = 'Select work items to consolidate';
        }
    }

    // Update selection count display
    updateWorkItemSelectionCount(selectedKeys.length);
}

// Toggle all work item checkboxes
function toggleAllWorkItems(checked) {
    const checkboxes = document.querySelectorAll('.work-item-checkbox');
    checkboxes.forEach(cb => {
        cb.checked = checked;
    });
    updateWorkItemSelectionButtons();
}

// Update selection count display
function updateWorkItemSelectionCount(count) {
    const countDisplay = document.getElementById('work-item-selection-count');

    if (countDisplay) {
        if (count > 0) {
            countDisplay.textContent = `${count} selected`;
            countDisplay.style.display = 'inline-block';
        } else {
            countDisplay.style.display = 'none';
        }
    }
}

// Initialize button states on page load
document.addEventListener('DOMContentLoaded', function () {
    updateWorkItemSelectionButtons();
});

// Make functions available globally
window.loadWorkItems = loadWorkItems;
window.showWorkItemEditor = showWorkItemEditor;
window.saveWorkItemFromEditor = saveWorkItemFromEditor;
window.cancelWorkItemEdit = cancelWorkItemEdit;
window.deleteWorkItem = deleteWorkItem;
window.filterWorkItems = filterWorkItems;
window.processAIWorkItemOperation = processAIWorkItemOperation;
window.refineWorkItemDescription = refineWorkItemDescription;
window.getSelectedWorkItemKeys = getSelectedWorkItemKeys;
window.updateWorkItemSelectionButtons = updateWorkItemSelectionButtons;
window.toggleAllWorkItems = toggleAllWorkItems;
