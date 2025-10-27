// Units of Work functionality
// Handles units of work management

import { getCurrentAgencyId, showNotification } from './utils.js';

// Unit editor state management
let unitEditorState = {
    mode: 'add', // 'add' or 'edit'
    unitKey: null,
    originalDescription: ''
};

// Load units of work list
export function loadUnits() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const unitsList = document.getElementById('units-list');
    if (!unitsList) {
        console.error('Units list container not found');
        return;
    }

    // Show loading state
    unitsList.innerHTML = '<div class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading units of work...</p></div>';

    // Fetch units HTML from API
    fetch(`/api/v1/agencies/${agencyId}/units/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load units');
            }
            return response.text();
        })
        .then(html => {
            unitsList.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading units:', error);
            unitsList.innerHTML = '<div class="has-text-danger has-text-centered py-5"><p>Error loading units of work</p></div>';
        });
}

// Show unit editor
export function showUnitEditor(mode, unitKey = null, description = '') {
    unitEditorState.mode = mode;
    unitEditorState.unitKey = unitKey;
    unitEditorState.originalDescription = description;

    const editorCard = document.getElementById('unit-editor-card');
    const listCard = document.getElementById('units-list-card');
    const editorTitle = document.getElementById('unit-editor-title');
    const descriptionEditor = document.getElementById('unit-description-input');

    if (!editorCard || !listCard || !editorTitle || !descriptionEditor) {
        console.error('Unit editor elements not found');
        return;
    }

    // Update editor title and content
    if (mode === 'add') {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Unit of Work</span>';
        descriptionEditor.value = '';
    } else {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Unit of Work</span>';
        descriptionEditor.value = description;
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on editor
    descriptionEditor.focus();
}

// Save unit from editor
export function saveUnitFromEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const descriptionEditor = document.getElementById('unit-description-input');
    if (!descriptionEditor) {
        console.error('Description editor not found');
        return;
    }

    const description = descriptionEditor.value.trim();
    if (!description) {
        showNotification('Please enter a unit description', 'warning');
        descriptionEditor.focus();
        return;
    }

    const saveBtn = document.getElementById('save-unit-btn');
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
    }

    const isAddMode = unitEditorState.mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/units`
        : `/api/v1/agencies/${agencyId}/units/${unitEditorState.unitKey}`;
    const method = isAddMode ? 'POST' : 'PUT';

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ description: description })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} unit`);
            }
            return response.json();
        })
        .then(() => {
            showNotification(`Unit ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelUnitEdit(); // Hide editor
            loadUnits(); // Reload the list
        })
        .catch(error => {
            console.error(`Error ${isAddMode ? 'creating' : 'updating'} unit:`, error);
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} unit`, 'error');
        })
        .finally(() => {
            if (saveBtn) {
                saveBtn.classList.remove('is-loading');
            }
        });
}

// Cancel unit edit
export function cancelUnitEdit() {
    const editorCard = document.getElementById('unit-editor-card');
    const listCard = document.getElementById('units-list-card');
    const descriptionEditor = document.getElementById('unit-description-input');

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');
    if (descriptionEditor) descriptionEditor.value = '';

    // Reset state
    unitEditorState = {
        mode: 'add',
        unitKey: null,
        originalDescription: ''
    };
}

// Delete unit
export function deleteUnit(unitKey, unitNumber) {
    if (!confirm(`Are you sure you want to delete unit #${unitNumber}? This will renumber all subsequent units.`)) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    fetch(`/api/v1/agencies/${agencyId}/units/${unitKey}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete unit');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Unit deleted successfully!', 'success');
            loadUnits(); // Reload the list
        })
        .catch(error => {
            console.error('Error deleting unit:', error);
            showNotification('Error deleting unit', 'error');
        });
}