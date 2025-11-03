// Work Items functionality
// Handles work items management

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';

// Work item editor state management
let workItemEditorState = {
    mode: 'add', // 'add' or 'edit'
    workItemKey: null,
    originalData: {}
};

// Load work items list
export function loadWorkItems() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const workItemsTableBody = document.getElementById('work-items-table-body');
    if (!workItemsTableBody) {
        return;
    }

    // Show loading state
    workItemsTableBody.innerHTML = '<tr><td colspan="4" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading work items...</p></td></tr>';

    // Fetch work items HTML from API
    const url = `/api/v1/agencies/${agencyId}/work-items/html`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load work items');
            }
            return response.text();
        })
        .then(html => {
            workItemsTableBody.innerHTML = html;
        })
        .catch(error => {
            workItemsTableBody.innerHTML = '<tr><td colspan="4" class="has-text-danger has-text-centered py-5"><p>Error loading work items</p></td></tr>';
        });
}

// Show work item editor
export function showWorkItemEditor(mode, workItemKey = null) {

    workItemEditorState.mode = mode;
    workItemEditorState.workItemKey = workItemKey;

    const editorCard = document.getElementById('work-item-editor-card');
    const listCard = document.getElementById('work-items-list-card');
    const editorTitle = document.getElementById('work-item-editor-title');

    if (!editorCard || !listCard || !editorTitle) {
        return;
    }

    if (mode === 'add') {
        // Clear form for new work item
        editorTitle.textContent = 'Add New Work Item';
        clearWorkItemForm();
    } else if (mode === 'edit') {
        // Load existing work item data
        editorTitle.textContent = 'Edit Work Item';
        loadWorkItemData(workItemKey);
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on title field
    const titleEditor = document.getElementById('work-item-title-editor');
    if (titleEditor) {
        titleEditor.focus();
    } else {
    }

}

// Load work item data for editing
function loadWorkItemData(workItemKey) {

    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) {
        return;
    }


    // Fetch work item data
    fetch(`/api/v1/agencies/${agencyId}/work-items`)
        .then(response => {
            return response.json();
        })
        .then(workItems => {

            // The key field comes as "_key" from JSON
            const workItem = workItems.find(wi => wi._key === workItemKey || wi.key === workItemKey);


            if (workItem) {
                populateWorkItemForm(workItem);
                workItemEditorState.originalData = workItem;
            } else {
                showNotification('Work item not found', 'error');
            }
        })
        .catch(error => {
            showNotification('Error loading work item data', 'error');
        });
}

// Populate form with work item data
function populateWorkItemForm(workItem) {

    const fields = {
        'work-item-title-editor': workItem.title || '',
        'work-item-description-editor': workItem.description || '',
        'work-item-deliverables-editor': workItem.deliverables ? workItem.deliverables.join('\n') : '',
        'work-item-dependencies-editor': workItem.dependencies ? workItem.dependencies.join(', ') : '',
        'work-item-tags-editor': workItem.tags ? workItem.tags.join(', ') : ''
    };


    for (const [id, value] of Object.entries(fields)) {
        const element = document.getElementById(id);
        if (element) {
            element.value = value;
        } else {
        }
    }

}

// Clear work item form
function clearWorkItemForm() {
    const fields = [
        'work-item-title-editor',
        'work-item-description-editor',
        'work-item-deliverables-editor',
        'work-item-dependencies-editor',
        'work-item-tags-editor'
    ];

    fields.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            element.value = '';
        }
    });

    workItemEditorState.originalData = {};
}

// Save work item from editor
export function saveWorkItemFromEditor() {

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Get form values
    const title = document.getElementById('work-item-title-editor')?.value.trim();
    const description = document.getElementById('work-item-description-editor')?.value.trim();
    const deliverables = document.getElementById('work-item-deliverables-editor')?.value
        .split('\n')
        .map(d => d.trim())
        .filter(d => d.length > 0);
    const dependencies = document.getElementById('work-item-dependencies-editor')?.value
        .split(',')
        .map(d => d.trim())
        .filter(d => d.length > 0);
    const tags = document.getElementById('work-item-tags-editor')?.value
        .split(',')
        .map(t => t.trim())
        .filter(t => t.length > 0);

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

    const isAddMode = workItemEditorState.mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/work-items`
        : `/api/v1/agencies/${agencyId}/work-items/${workItemEditorState.workItemKey}`;
    const method = isAddMode ? 'POST' : 'PUT';

    const requestBody = {
        title,
        description,
        deliverables,
        dependencies,
        tags
    };


    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} work item`);
            }
            return response.json();
        })
        .then(data => {
            showNotification(`Work item ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelWorkItemEdit();
            loadWorkItems();
        })
        .catch(error => {
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} work item`, 'error');
        });
}

// Cancel work item edit
export function cancelWorkItemEdit() {

    const editorCard = document.getElementById('work-item-editor-card');
    const listCard = document.getElementById('work-items-list-card');

    if (editorCard) {
        editorCard.classList.add('is-hidden');
    }
    if (listCard) {
        listCard.classList.remove('is-hidden');
    }

    clearWorkItemForm();

    // Reset state
    workItemEditorState = {
        mode: 'add',
        workItemKey: null,
        originalData: {}
    };

}

// Delete work item
export function deleteWorkItem(workItemKey) {
    if (!confirm(`Are you sure you want to delete this work item?`)) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    fetch(`/api/v1/agencies/${agencyId}/work-items/${workItemKey}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete work item');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Work item deleted successfully!', 'success');
            loadWorkItems();
        })
        .catch(error => {
            showNotification('Error deleting work item', 'error');
        });
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

// Validate work item dependencies
export function validateWorkItemDependencies(dependencies) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !dependencies || dependencies.length === 0) {
        return Promise.resolve(true);
    }

    return fetch(`/api/v1/agencies/${agencyId}/work-items/validate-deps`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ dependencies })
    })
        .then(response => response.json())
        .then(data => {
            if (!data.valid) {
                showNotification(`Invalid dependencies: ${data.error}`, 'warning');
                return false;
            }
            return true;
        })
        .catch(error => {
            return true; // Don't block on validation errors
        });
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
window.validateWorkItemDependencies = validateWorkItemDependencies;
window.getSelectedWorkItemKeys = getSelectedWorkItemKeys;
window.updateWorkItemSelectionButtons = updateWorkItemSelectionButtons;
window.toggleAllWorkItems = toggleAllWorkItems;
