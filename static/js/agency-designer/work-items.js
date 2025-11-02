// Work Items functionality
// Handles work items management

import { getCurrentAgencyId, showNotification } from './utils.js';

// Work item editor state management
let workItemEditorState = {
    mode: 'add', // 'add' or 'edit'
    workItemKey: null,
    originalData: {}
};

// Load work items list
export function loadWorkItems() {
    console.log('[Work Items] loadWorkItems() called');
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('[Work Items] No agency ID found');
        return;
    }
    console.log('[Work Items] Agency ID:', agencyId);

    const workItemsTableBody = document.getElementById('work-items-table-body');
    if (!workItemsTableBody) {
        console.error('[Work Items] Work items table body not found');
        return;
    }
    console.log('[Work Items] Table body element found');

    // Show loading state
    workItemsTableBody.innerHTML = '<tr><td colspan="4" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading work items...</p></td></tr>';

    // Fetch work items HTML from API
    const url = `/api/v1/agencies/${agencyId}/work-items/html`;
    console.log('[Work Items] Fetching from URL:', url);

    fetch(url)
        .then(response => {
            console.log('[Work Items] Response status:', response.status);
            if (!response.ok) {
                throw new Error('Failed to load work items');
            }
            return response.text();
        })
        .then(html => {
            console.log('[Work Items] Received HTML, length:', html.length);
            workItemsTableBody.innerHTML = html;
            console.log('[Work Items] Table updated successfully');
        })
        .catch(error => {
            console.error('[Work Items] Error loading work items:', error);
            workItemsTableBody.innerHTML = '<tr><td colspan="4" class="has-text-danger has-text-centered py-5"><p>Error loading work items</p></td></tr>';
        });
}

// Show work item editor
export function showWorkItemEditor(mode, workItemKey = null) {
    console.log('[Work Items] showWorkItemEditor() called with mode:', mode, 'key:', workItemKey);

    workItemEditorState.mode = mode;
    workItemEditorState.workItemKey = workItemKey;

    const editorCard = document.getElementById('work-item-editor-card');
    const listCard = document.getElementById('work-items-list-card');
    const editorTitle = document.getElementById('work-item-editor-title');

    console.log('[Work Items] Editor elements found:', {
        editorCard: !!editorCard,
        listCard: !!listCard,
        editorTitle: !!editorTitle
    });

    if (!editorCard || !listCard || !editorTitle) {
        console.error('[Work Items] Work item editor elements not found:', {
            editorCard: editorCard,
            listCard: listCard,
            editorTitle: editorTitle
        });
        return;
    }

    if (mode === 'add') {
        console.log('[Work Items] Setting up ADD mode');
        // Clear form for new work item
        editorTitle.textContent = 'Add New Work Item';
        clearWorkItemForm();
    } else if (mode === 'edit') {
        console.log('[Work Items] Setting up EDIT mode for key:', workItemKey);
        // Load existing work item data
        editorTitle.textContent = 'Edit Work Item';
        loadWorkItemData(workItemKey);
    }

    // Show editor, hide list
    console.log('[Work Items] Toggling visibility: showing editor, hiding list');
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on title field
    const titleEditor = document.getElementById('work-item-title-editor');
    if (titleEditor) {
        console.log('[Work Items] Focusing on title editor');
        titleEditor.focus();
    } else {
        console.warn('[Work Items] Title editor not found for focus');
    }

    console.log('[Work Items] Editor should now be visible');
}

// Load work item data for editing
function loadWorkItemData(workItemKey) {
    console.log('[Work Items] loadWorkItemData() called with key:', workItemKey);

    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) {
        console.error('[Work Items] Missing agencyId or workItemKey:', { agencyId, workItemKey });
        return;
    }

    console.log('[Work Items] Fetching work items for agency:', agencyId);

    // Fetch work item data
    fetch(`/api/v1/agencies/${agencyId}/work-items`)
        .then(response => {
            console.log('[Work Items] Fetch response status:', response.status);
            return response.json();
        })
        .then(workItems => {
            console.log('[Work Items] Received work items:', workItems.length);
            console.log('[Work Items] Looking for key:', workItemKey);
            console.log('[Work Items] Sample work item structure:', workItems[0]);

            // The key field comes as "_key" from JSON
            const workItem = workItems.find(wi => wi._key === workItemKey || wi.key === workItemKey);

            console.log('[Work Items] Found work item:', workItem);

            if (workItem) {
                populateWorkItemForm(workItem);
                workItemEditorState.originalData = workItem;
                console.log('[Work Items] Form populated successfully');
            } else {
                console.error('[Work Items] Work item not found with key:', workItemKey);
                showNotification('Work item not found', 'error');
            }
        })
        .catch(error => {
            console.error('[Work Items] Error loading work item:', error);
            showNotification('Error loading work item data', 'error');
        });
}

// Populate form with work item data
function populateWorkItemForm(workItem) {
    console.log('[Work Items] populateWorkItemForm() called with:', workItem);

    const fields = {
        'work-item-title-editor': workItem.title || '',
        'work-item-type-editor': workItem.type || 'task',
        'work-item-priority-editor': workItem.priority || 'P2',
        'work-item-status-editor': workItem.status || 'not-started',
        'work-item-description-editor': workItem.description || '',
        'work-item-deliverables-editor': workItem.deliverables ? workItem.deliverables.join('\n') : '',
        'work-item-dependencies-editor': workItem.dependencies ? workItem.dependencies.join(', ') : '',
        'work-item-effort-editor': workItem.estimated_effort || '',
        'work-item-tags-editor': workItem.tags ? workItem.tags.join(', ') : ''
    };

    console.log('[Work Items] Field values to populate:', fields);

    for (const [id, value] of Object.entries(fields)) {
        const element = document.getElementById(id);
        if (element) {
            element.value = value;
            console.log(`[Work Items] Set ${id} = ${value.substring ? value.substring(0, 50) : value}`);
        } else {
            console.warn(`[Work Items] Element not found: ${id}`);
        }
    }

    console.log('[Work Items] Form population complete');
}

// Clear work item form
function clearWorkItemForm() {
    const fields = [
        'work-item-title-editor',
        'work-item-description-editor',
        'work-item-deliverables-editor',
        'work-item-dependencies-editor',
        'work-item-effort-editor',
        'work-item-tags-editor'
    ];

    fields.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            element.value = '';
        }
    });

    // Reset selects to defaults
    const typeSelect = document.getElementById('work-item-type-editor');
    if (typeSelect) typeSelect.value = 'task';

    const prioritySelect = document.getElementById('work-item-priority-editor');
    if (prioritySelect) prioritySelect.value = 'P2';

    const statusSelect = document.getElementById('work-item-status-editor');
    if (statusSelect) statusSelect.value = 'not-started';

    workItemEditorState.originalData = {};
}

// Save work item from editor
export function saveWorkItemFromEditor() {
    console.log('[Work Items] saveWorkItemFromEditor() called');

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('[Work Items] No agency ID found');
        showNotification('Error: No agency selected', 'error');
        return;
    }
    console.log('[Work Items] Agency ID:', agencyId);

    // Get form values
    const title = document.getElementById('work-item-title-editor')?.value.trim();
    const type = document.getElementById('work-item-type-editor')?.value;
    const priority = document.getElementById('work-item-priority-editor')?.value;
    const status = document.getElementById('work-item-status-editor')?.value;
    const description = document.getElementById('work-item-description-editor')?.value.trim();
    const deliverables = document.getElementById('work-item-deliverables-editor')?.value
        .split('\n')
        .map(d => d.trim())
        .filter(d => d.length > 0);
    const dependencies = document.getElementById('work-item-dependencies-editor')?.value
        .split(',')
        .map(d => d.trim())
        .filter(d => d.length > 0);
    const effort = document.getElementById('work-item-effort-editor')?.value;
    const tags = document.getElementById('work-item-tags-editor')?.value
        .split(',')
        .map(t => t.trim())
        .filter(t => t.length > 0);

    console.log('[Work Items] Form values:', {
        title,
        type,
        priority,
        status,
        descriptionLength: description?.length || 0,
        deliverablesCount: deliverables?.length || 0,
        dependenciesCount: dependencies?.length || 0,
        effort,
        tagsCount: tags?.length || 0
    });

    // Validation
    if (!title) {
        console.warn('[Work Items] Validation failed: no title');
        showNotification('Please enter a work item title', 'warning');
        document.getElementById('work-item-title-editor')?.focus();
        return;
    }

    if (!description) {
        console.warn('[Work Items] Validation failed: no description');
        showNotification('Please enter a work item description', 'warning');
        document.getElementById('work-item-description-editor')?.focus();
        return;
    }

    const isAddMode = workItemEditorState.mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/work-items`
        : `/api/v1/agencies/${agencyId}/work-items/${workItemEditorState.workItemKey}`;
    const method = isAddMode ? 'POST' : 'PUT';

    console.log('[Work Items] Sending request:', {
        mode: workItemEditorState.mode,
        method,
        url
    });

    const requestBody = {
        title,
        type,
        priority,
        status,
        description,
        deliverables,
        dependencies,
        estimated_effort: effort || '',
        tags
    };

    console.log('[Work Items] Request body:', requestBody);

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    })
        .then(response => {
            console.log('[Work Items] Response status:', response.status);
            if (!response.ok) {
                throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} work item`);
            }
            return response.json();
        })
        .then(data => {
            console.log('[Work Items] Success! Response data:', data);
            showNotification(`Work item ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelWorkItemEdit();
            loadWorkItems();
        })
        .catch(error => {
            console.error(`[Work Items] Error ${isAddMode ? 'creating' : 'updating'} work item:`, error);
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} work item`, 'error');
        });
}

// Cancel work item edit
export function cancelWorkItemEdit() {
    console.log('[Work Items] cancelWorkItemEdit() called');

    const editorCard = document.getElementById('work-item-editor-card');
    const listCard = document.getElementById('work-items-list-card');

    if (editorCard) {
        console.log('[Work Items] Hiding editor card');
        editorCard.classList.add('is-hidden');
    }
    if (listCard) {
        console.log('[Work Items] Showing list card');
        listCard.classList.remove('is-hidden');
    }

    clearWorkItemForm();

    // Reset state
    workItemEditorState = {
        mode: 'add',
        workItemKey: null,
        originalData: {}
    };

    console.log('[Work Items] Editor cancelled, state reset');
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
            console.error('Error deleting work item:', error);
            showNotification('Error deleting work item', 'error');
        });
}

// Filter work items
export function filterWorkItems() {
    const searchInput = document.getElementById('work-items-search')?.value.toLowerCase() || '';
    const statusFilter = document.getElementById('filter-status')?.value || '';
    const priorityFilter = document.getElementById('filter-priority')?.value || '';
    const typeFilter = document.getElementById('filter-type')?.value || '';

    const rows = document.querySelectorAll('#work-items-table-body tr.table-item');

    rows.forEach(row => {
        const key = row.dataset.itemKey || '';
        const title = row.querySelector('.has-text-weight-semibold')?.textContent.toLowerCase() || '';
        const type = row.dataset.type || '';
        const priority = row.dataset.priority || '';
        const status = row.dataset.status || '';

        const matchesSearch = !searchInput || key.toLowerCase().includes(searchInput) || title.includes(searchInput);
        const matchesStatus = !statusFilter || status === statusFilter;
        const matchesPriority = !priorityFilter || priority === priorityFilter;
        const matchesType = !typeFilter || type === typeFilter;

        if (matchesSearch && matchesStatus && matchesPriority && matchesType) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// AI Work Item Operations
export async function processAIWorkItemOperation(operations) {
    console.log('[Work Items] AI operation called - this is for AI-GENERATED work items, not manual creation');
    console.log('[Work Items] To manually create a work item, click the green "Add" button instead');

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

    console.log('[Work Items] AI Work Item operations requested:', operations);
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
        console.log('[Work Items] AI processing result:', data);

        // Hide AI processing status
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

        // Reload work items to show updates
        await loadWorkItems();

    } catch (error) {
        console.error('[Work Items] AI processing error:', error);

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
            console.error('Error validating dependencies:', error);
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
