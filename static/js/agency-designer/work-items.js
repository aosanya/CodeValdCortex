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
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const workItemsTableBody = document.getElementById('work-items-table-body');
    if (!workItemsTableBody) {
        console.error('Work items table body not found');
        return;
    }

    // Show loading state
    workItemsTableBody.innerHTML = '<tr><td colspan="4" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading work items...</p></td></tr>';

    // Fetch work items HTML from API
    fetch(`/api/v1/agencies/${agencyId}/work-items/html`)
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
            console.error('Error loading work items:', error);
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
        console.error('Work item editor elements not found');
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
    }
}

// Load work item data for editing
function loadWorkItemData(workItemKey) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workItemKey) return;

    // Fetch work item data
    fetch(`/api/v1/agencies/${agencyId}/work-items`)
        .then(response => response.json())
        .then(workItems => {
            const workItem = workItems.find(wi => wi.key === workItemKey);
            if (workItem) {
                populateWorkItemForm(workItem);
                workItemEditorState.originalData = workItem;
            }
        })
        .catch(error => {
            console.error('Error loading work item:', error);
            showNotification('Error loading work item data', 'error');
        });
}

// Populate form with work item data
function populateWorkItemForm(workItem) {
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

    for (const [id, value] of Object.entries(fields)) {
        const element = document.getElementById(id);
        if (element) {
            element.value = value;
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
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

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
        type,
        priority,
        status,
        description,
        deliverables,
        dependencies,
        estimated_effort: effort || '',
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
        .then(() => {
            showNotification(`Work item ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelWorkItemEdit();
            loadWorkItems();
        })
        .catch(error => {
            console.error(`Error ${isAddMode ? 'creating' : 'updating'} work item:`, error);
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} work item`, 'error');
        });
}

// Cancel work item edit
export function cancelWorkItemEdit() {
    const editorCard = document.getElementById('work-item-editor-card');
    const listCard = document.getElementById('work-items-list-card');

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');

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
export function processAIWorkItemOperation(operations) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // For now, show a placeholder notification
    // This will be implemented when the AI refinement handler is created
    showNotification('AI Work Item operations coming soon!', 'info');

    console.log('AI Work Item operations requested:', operations);
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
