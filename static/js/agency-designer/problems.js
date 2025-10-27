// Problems functionality
// Handles problem definition management

import { getCurrentAgencyId, showNotification } from './utils.js';

// Problem editor state management
let problemEditorState = {
    mode: 'add', // 'add' or 'edit'
    problemKey: null,
    originalCode: '',
    originalDescription: ''
};

// Load problems list
export function loadProblems() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const problemsTableBody = document.getElementById('problems-table-body');
    if (!problemsTableBody) {
        console.error('Problems table body not found');
        return;
    }

    // Show loading state
    problemsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading problems...</p></td></tr>';

    // Fetch problems HTML from API
    fetch(`/api/v1/agencies/${agencyId}/problems/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load problems');
            }
            return response.text();
        })
        .then(html => {
            problemsTableBody.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading problems:', error);
            problemsTableBody.innerHTML = '<tr><td colspan="3" class="has-text-danger has-text-centered py-5"><p>Error loading problems</p></td></tr>';
        });
}

// Show problem editor
export function showProblemEditor(mode, problemKey = null, code = '', description = '') {
    problemEditorState.mode = mode;
    problemEditorState.problemKey = problemKey;
    problemEditorState.originalCode = code;
    problemEditorState.originalDescription = description;

    const editorCard = document.getElementById('problem-editor-card');
    const listCard = document.getElementById('problems-list-card');
    const editorTitle = document.getElementById('problem-editor-title');
    const codeEditor = document.getElementById('problem-code-editor');
    const descriptionEditor = document.getElementById('problem-description-editor');

    if (!editorCard || !listCard || !editorTitle || !codeEditor || !descriptionEditor) {
        console.error('Problem editor elements not found');
        return;
    }

    // Update editor title and content
    if (mode === 'add') {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Problem</span>';
        codeEditor.value = '';
        descriptionEditor.value = '';
    } else {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Problem</span>';
        codeEditor.value = code;
        descriptionEditor.value = description;
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on editor
    descriptionEditor.focus();
}

// Save problem from editor
export function saveProblemFromEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const codeEditor = document.getElementById('problem-code-editor');
    const descriptionEditor = document.getElementById('problem-description-editor');
    if (!codeEditor || !descriptionEditor) {
        console.error('Editor elements not found');
        return;
    }

    const code = codeEditor.value.trim();
    const description = descriptionEditor.value.trim();

    if (!code) {
        showNotification('Please enter a problem code', 'warning');
        codeEditor.focus();
        return;
    }

    if (!description) {
        showNotification('Please enter a problem description', 'warning');
        descriptionEditor.focus();
        return;
    }

    const saveBtn = document.getElementById('save-problem-btn');
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
    }

    const isAddMode = problemEditorState.mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/problems`
        : `/api/v1/agencies/${agencyId}/problems/${problemEditorState.problemKey}`;
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
                throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} problem`);
            }
            return response.json();
        })
        .then(() => {
            showNotification(`Problem ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');
            cancelProblemEdit(); // Hide editor
            loadProblems(); // Reload the list
        })
        .catch(error => {
            console.error(`Error ${isAddMode ? 'creating' : 'updating'} problem:`, error);
            showNotification(`Error ${isAddMode ? 'adding' : 'updating'} problem`, 'error');
        })
        .finally(() => {
            if (saveBtn) {
                saveBtn.classList.remove('is-loading');
            }
        });
}

// Cancel problem edit
export function cancelProblemEdit() {
    const editorCard = document.getElementById('problem-editor-card');
    const listCard = document.getElementById('problems-list-card');
    const codeEditor = document.getElementById('problem-code-editor');
    const descriptionEditor = document.getElementById('problem-description-editor');

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');
    if (codeEditor) codeEditor.value = '';
    if (descriptionEditor) descriptionEditor.value = '';

    // Reset state
    problemEditorState = {
        mode: 'add',
        problemKey: null,
        originalCode: '',
        originalDescription: ''
    };
}

// Delete problem
export function deleteProblem(problemKey, problemNumber) {
    if (!confirm(`Are you sure you want to delete problem #${problemNumber}? This will renumber all subsequent problems.`)) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    fetch(`/api/v1/agencies/${agencyId}/problems/${problemKey}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete problem');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Problem deleted successfully!', 'success');
            loadProblems(); // Reload the list
        })
        .catch(error => {
            console.error('Error deleting problem:', error);
            showNotification('Error deleting problem', 'error');
        });
}