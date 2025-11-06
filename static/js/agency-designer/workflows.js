// Workflows functionality
// Handles workflow management

import { getCurrentAgencyId, showNotification } from './utils.js';
import { scrollToBottom } from './chat.js';
import { loadEntityList, showEntityEditor, cancelEntityEdit, populateForm, clearForm } from './crud-helpers.js';

// Workflow editor state management
let workflowEditorState = {
    mode: 'add', // 'add' or 'edit'
    workflowId: null,
    originalData: {}
};

// Load workflows list
export function loadWorkflows() {
    return loadEntityList('workflows', 'workflows-table-body', 5);
}

// Show workflow editor
export function showWorkflowEditor(mode, workflowId = null) {
    workflowEditorState.mode = mode;
    workflowEditorState.workflowId = workflowId;

    showEntityEditor(
        mode,
        'workflow-editor-card',
        'workflows-list-card',
        'workflow-editor-title',
        'Add New Workflow',
        'Edit Workflow',
        'workflow-name-editor'
    );

    if (mode === 'add') {
        clearWorkflowForm();
    } else if (mode === 'edit') {
        loadWorkflowData(workflowId);
    }
}

// Load workflow data for editing
function loadWorkflowData(workflowId) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId || !workflowId) {
        return;
    }

    // Fetch workflow data
    fetch(`/api/v1/workflows/${workflowId}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load workflow');
            }
            return response.json();
        })
        .then(workflow => {
            populateWorkflowForm(workflow);
            workflowEditorState.originalData = workflow;

            // Enable refine button when editing
            const refineBtn = document.getElementById('refine-workflow-btn');
            if (refineBtn) {
                refineBtn.disabled = false;
            }
        })
        .catch(error => {
            console.error('Error loading workflow:', error);
            showNotification('Error loading workflow data', 'error');
        });
}

// Populate form with workflow data
function populateWorkflowForm(workflow) {
    populateForm({
        'workflow-name-editor': workflow.name || '',
        'workflow-description-editor': workflow.description || '',
        'workflow-version-editor': workflow.version || '1.0.0',
        'workflow-status-editor': workflow.status || 'draft'
    });

    // Handle workflow structure (nodes and edges)
    if (workflow.nodes || workflow.edges) {
        const structure = {
            nodes: workflow.nodes || [],
            edges: workflow.edges || []
        };
        document.getElementById('workflow-structure-editor').value = JSON.stringify(structure, null, 2);
    }
}

// Clear workflow form
function clearWorkflowForm() {
    clearForm([
        'workflow-name-editor',
        'workflow-description-editor',
        'workflow-version-editor',
        'workflow-structure-editor'
    ]);

    // Reset to defaults
    document.getElementById('workflow-version-editor').value = '1.0.0';
    document.getElementById('workflow-status-editor').value = 'draft';

    // Disable refine button
    const refineBtn = document.getElementById('refine-workflow-btn');
    if (refineBtn) {
        refineBtn.disabled = true;
    }

    workflowEditorState.originalData = {};
}

// Save workflow from editor
export function saveWorkflowFromEditor() {
    // Get form values
    const name = document.getElementById('workflow-name-editor').value.trim();
    const description = document.getElementById('workflow-description-editor').value.trim();
    const version = document.getElementById('workflow-version-editor').value.trim();
    const status = document.getElementById('workflow-status-editor').value;
    const structureJson = document.getElementById('workflow-structure-editor').value.trim();

    // Validate required fields
    if (!name) {
        showNotification('Please provide a workflow name', 'error');
        return;
    }

    if (!description) {
        showNotification('Please provide a workflow description', 'error');
        return;
    }

    // Parse workflow structure
    let nodes = [];
    let edges = [];
    if (structureJson) {
        try {
            const structure = JSON.parse(structureJson);
            nodes = structure.nodes || [];
            edges = structure.edges || [];
        } catch (e) {
            showNotification('Invalid workflow structure JSON', 'error');
            return;
        }
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const workflow = {
        name,
        description,
        version,
        status,
        nodes,
        edges,
        variables: {},
        agency_id: agencyId // Always include agency_id
    };

    // Determine endpoint and method
    let url, method;
    if (workflowEditorState.mode === 'add') {
        url = `/api/v1/agencies/${agencyId}/workflows`;
        method = 'POST';
    } else {
        url = `/api/v1/workflows/${workflowEditorState.workflowId}`;
        method = 'PUT';
        workflow.id = workflowEditorState.workflowId;
    }

    // Save workflow
    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(workflow)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to ${method === 'POST' ? 'create' : 'update'} workflow`);
            }
            return response.json();
        })
        .then(() => {
            showNotification(`Workflow ${method === 'POST' ? 'created' : 'updated'} successfully`, 'success');
            cancelWorkflowEdit();
            loadWorkflows();
        })
        .catch(error => {
            console.error('Error saving workflow:', error);
            showNotification(`Error ${method === 'POST' ? 'creating' : 'updating'} workflow`, 'error');
        });
}

// Cancel workflow editing
export function cancelWorkflowEdit() {
    cancelEntityEdit('workflow-editor-card', 'workflows-list-card');
    clearWorkflowForm();
}

// Delete workflow
export function deleteWorkflow(workflowId) {
    const confirmDelete = confirm('Are you sure you want to delete this workflow?');
    if (!confirmDelete) {
        return;
    }

    fetch(`/api/v1/workflows/${workflowId}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete workflow');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Workflow deleted successfully', 'success');
            loadWorkflows();
        })
        .catch(error => {
            console.error('Error deleting workflow:', error);
            showNotification('Error deleting workflow', 'error');
        });
}

// Duplicate workflow
export function duplicateWorkflow(workflowId) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Unable to determine current agency', 'error');
        return;
    }

    fetch(`/api/v1/workflows/${workflowId}/duplicate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to duplicate workflow');
            }
            return response.json();
        })
        .then(() => {
            showNotification('Workflow duplicated successfully', 'success');
            loadWorkflows();
        })
        .catch(error => {
            console.error('Error duplicating workflow:', error);
            showNotification('Error duplicating workflow', 'error');
        });
}

// Filter workflows based on search input
export function filterWorkflows() {
    const searchInput = document.getElementById('workflows-search').value.toLowerCase();
    const rows = document.querySelectorAll('#workflows-table-body tr');

    rows.forEach(row => {
        const name = row.querySelector('td:nth-child(2)')?.textContent.toLowerCase() || '';
        const version = row.querySelector('td:nth-child(3)')?.textContent.toLowerCase() || '';
        const status = row.querySelector('td:nth-child(4)')?.textContent.toLowerCase() || '';

        if (name.includes(searchInput) || version.includes(searchInput) || status.includes(searchInput)) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// Toggle all workflows selection
export function toggleAllWorkflows(checked) {
    const checkboxes = document.querySelectorAll('#workflows-table-body input[type="checkbox"]');
    checkboxes.forEach(cb => {
        cb.checked = checked;
    });
    updateWorkflowSelectionUI();
}

// Update workflow selection UI
function updateWorkflowSelectionUI() {
    const checkboxes = document.querySelectorAll('#workflows-table-body input[type="checkbox"]:checked');
    const count = checkboxes.length;
    const countSpan = document.getElementById('workflow-selection-count');
    const refineBtn = document.getElementById('ai-refine-workflows-btn');
    const suggestBtn = document.getElementById('ai-suggest-workflows-btn');

    if (count > 0) {
        countSpan.textContent = `${count} selected`;
        countSpan.style.display = 'inline-flex';

        // Enable refine/suggest buttons when exactly 1 workflow selected
        if (count === 1) {
            refineBtn?.classList.remove('is-static');
            refineBtn?.removeAttribute('disabled');
            suggestBtn?.classList.remove('is-static');
            suggestBtn?.removeAttribute('disabled');
        } else {
            refineBtn?.classList.add('is-static');
            refineBtn?.setAttribute('disabled', 'disabled');
            suggestBtn?.classList.add('is-static');
            suggestBtn?.setAttribute('disabled', 'disabled');
        }
    } else {
        countSpan.style.display = 'none';
        refineBtn?.classList.add('is-static');
        refineBtn?.setAttribute('disabled', 'disabled');
        suggestBtn?.classList.add('is-static');
        suggestBtn?.setAttribute('disabled', 'disabled');
    }
}

// AI Workflow Operations
export function processAIWorkflowOperation(operation) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Unable to determine current agency', 'error');
        return;
    }

    const responseContainer = document.getElementById('workflow-ai-response');

    // Get selected workflows
    const selectedCheckboxes = document.querySelectorAll('#workflows-table-body input[type="checkbox"]:checked');
    const selectedKeys = Array.from(selectedCheckboxes).map(cb => cb.value);

    let userMessage = '';

    switch (operation) {
        case 'create':
            userMessage = 'Generate workflows from my work items';
            break;
        case 'refine':
            if (selectedKeys.length !== 1) {
                showNotification('Please select exactly one workflow to refine', 'warning');
                return;
            }
            userMessage = 'Refine this workflow to improve its structure and efficiency';
            break;
        case 'suggest':
            if (selectedKeys.length !== 1) {
                showNotification('Please select exactly one workflow to analyze', 'warning');
                return;
            }
            userMessage = 'Analyze this workflow and suggest improvements';
            break;
        default:
            showNotification('Unknown operation', 'error');
            return;
    }

    // Show loading state
    responseContainer.innerHTML = '<div class="notification is-info"><i class="fas fa-spinner fa-spin mr-2"></i>Processing...</div>';
    responseContainer.style.display = 'block';

    // Make AI request
    const requestData = {
        user_message: userMessage,
        workflow_keys: selectedKeys
    };

    fetch(`/api/v1/agencies/${agencyId}/workflows/refine-dynamic`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Server error: ${response.status}`);
            }
            return response.text();
        })
        .then(html => {
            responseContainer.innerHTML = html;
            loadWorkflows(); // Reload list to show new/updated workflows

            // Clear selection after successful operation
            const selectAll = document.getElementById('select-all-workflows');
            if (selectAll) {
                selectAll.checked = false;
                toggleAllWorkflows(false);
            }
        })
        .catch(error => {
            console.error('AI workflow operation error:', error);
            responseContainer.innerHTML = '<div class="notification is-danger"><i class="fas fa-exclamation-triangle mr-2"></i>Failed to process AI request. The response may have been too large or the AI service encountered an error.</div>';
        });
}

// Generate workflow with AI from description
export function generateWorkflowWithAI() {
    const description = document.getElementById('workflow-description-editor').value.trim();

    if (!description) {
        showNotification('Please provide a workflow description first', 'warning');
        return;
    }

    const agencyId = getCurrentAgencyId();
    const userMessage = `Generate a workflow structure for: ${description}`;

    // Make AI request
    fetch(`/api/v1/agencies/${agencyId}/workflows/refine-dynamic`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            user_message: userMessage,
            workflow_keys: []
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('AI generation failed');
            }
            return response.text();
        })
        .then(html => {
            showNotification('Workflow generated! Check the workflows list.', 'success');
            cancelWorkflowEdit();
            loadWorkflows();
        })
        .catch(error => {
            console.error('AI workflow generation error:', error);
            showNotification('Failed to generate workflow with AI', 'error');
        });
}

// Refine workflow with AI
export function refineWorkflowWithAI() {
    if (workflowEditorState.mode !== 'edit' || !workflowEditorState.workflowId) {
        showNotification('Please load a workflow first', 'warning');
        return;
    }

    const agencyId = getCurrentAgencyId();
    const userMessage = 'Refine and improve this workflow';

    fetch(`/api/v1/agencies/${agencyId}/workflows/refine-dynamic`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            user_message: userMessage,
            workflow_keys: [workflowEditorState.workflowId]
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('AI refinement failed');
            }
            return response.text();
        })
        .then(html => {
            showNotification('Workflow refined successfully!', 'success');
            // Reload the workflow data
            loadWorkflowData(workflowEditorState.workflowId);
        })
        .catch(error => {
            console.error('AI workflow refinement error:', error);
            showNotification('Failed to refine workflow with AI', 'error');
        });
}

// Make functions available globally
window.showWorkflowEditor = showWorkflowEditor;
window.saveWorkflowFromEditor = saveWorkflowFromEditor;
window.cancelWorkflowEdit = cancelWorkflowEdit;
window.deleteWorkflow = deleteWorkflow;
window.duplicateWorkflow = duplicateWorkflow;
window.filterWorkflows = filterWorkflows;
window.toggleAllWorkflows = toggleAllWorkflows;
window.processAIWorkflowOperation = processAIWorkflowOperation;
window.generateWorkflowWithAI = generateWorkflowWithAI;
window.refineWorkflowWithAI = refineWorkflowWithAI;
