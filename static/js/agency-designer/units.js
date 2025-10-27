// Units of Work functionality
// Handles units of work management

import { getCurrentAgencyId, showNotification } from './utils.js';

// Unit editor state management
let unitEditorState = {
    mode: 'add', // 'add' or 'edit'
    unitKey: null,
    originalCode: '',
    originalDescription: ''
};

// Load units of work list
export function loadUnits() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const unitsTableBody = document.getElementById('units-table-body');
    if (!unitsTableBody) {
        console.error('Units table body not found');
        return;
    }

    // Show loading state
    unitsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading units of work...</p></td></tr>';

    // Fetch units HTML from API
    fetch(`/api/v1/agencies/${agencyId}/units/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load units');
            }
            return response.text();
        })
        .then(html => {
            unitsTableBody.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading units:', error);
            unitsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-danger has-text-centered py-5"><p>Error loading units of work</p></td></tr>';
        });
}

// Show unit editor
export function showUnitEditor(mode, unitKey = null, code = '', description = '') {
    unitEditorState.mode = mode;
    unitEditorState.unitKey = unitKey;
    unitEditorState.originalCode = code;
    unitEditorState.originalDescription = description;

    const editorCard = document.getElementById('unit-editor-card');
    const listCard = document.getElementById('units-list-card');
    const editorTitle = document.getElementById('unit-editor-title');
    const codeEditor = document.getElementById('unit-code-input');
    const descriptionEditor = document.getElementById('unit-description-input');

    if (!editorCard || !listCard || !editorTitle || !codeEditor || !descriptionEditor) {
        console.error('Unit editor elements not found');
        return;
    }

    // Update editor title and content
    if (mode === 'add') {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Unit of Work</span>';
        codeEditor.value = '';
        descriptionEditor.value = '';
    } else {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Unit of Work</span>';
        codeEditor.value = code;
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

    const codeEditor = document.getElementById('unit-code-input');
    const descriptionEditor = document.getElementById('unit-description-input');
    if (!codeEditor || !descriptionEditor) {
        console.error('Editor elements not found');
        return;
    }

    const code = codeEditor.value.trim();
    const description = descriptionEditor.value.trim();

    if (!code) {
        showNotification('Please enter a unit code', 'warning');
        codeEditor.focus();
        return;
    }

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
        body: JSON.stringify({
            code: code,
            description: description
        })
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
    const codeEditor = document.getElementById('unit-code-input');
    const descriptionEditor = document.getElementById('unit-description-input');

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');
    if (codeEditor) codeEditor.value = '';
    if (descriptionEditor) descriptionEditor.value = '';

    // Reset state
    unitEditorState = {
        mode: 'add',
        unitKey: null,
        originalCode: '',
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