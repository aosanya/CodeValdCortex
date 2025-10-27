// AI Agency Designer - VS Code Style Interactions
// Handles chat, agent selection, and HTMX events

document.addEventListener('DOMContentLoaded', function () {
    console.log('Agency Designer: Initializing...');

    // Log initial active view
    const activeView = document.querySelector('.view-content.is-active');
    if (activeView) {
        console.log('Initial active view:', activeView.getAttribute('data-view-content'));
    }

    initializeChatScroll();
    initializeHTMXEvents();
    initializeViewSwitcher();
    initializeAgentSelection();
    initializeOverview();

    console.log('Agency Designer: Initialization complete');
});

// Initialize overview section
function initializeOverview() {
    // Check if we're on the overview view and introduction is active
    const overviewView = document.querySelector('.view-content[data-view-content="overview"]');
    const introEditor = document.getElementById('introduction-editor');

    if (overviewView && overviewView.classList.contains('is-active') && introEditor) {
        // Load introduction data
        loadIntroductionEditor();
    }
}

// Initialize auto-scroll for chat messages
function initializeChatScroll() {
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        // Scroll to bottom on page load
        scrollToBottom(chatContainer);
    }
}

// Scroll chat container to bottom
function scrollToBottom(container) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }
}

// Handle agent type selection in sidebar
function selectAgentType(element) {
    // Remove active class from all items
    const allItems = document.querySelectorAll('.agent-type-item');
    allItems.forEach(item => item.classList.remove('is-active'));

    // Add active class to selected item
    element.classList.add('is-active');

    // Update details title
    const agentName = element.querySelector('.agent-type-name')?.textContent || 'Agent Details';
    const detailsTitle = document.getElementById('details-title');
    if (detailsTitle) {
        detailsTitle.innerHTML = `
            <span class="icon"><i class="fas fa-robot"></i></span>
            <span>${agentName}</span>
        `;
    }
}

// Initialize agent selection handlers
function initializeAgentSelection() {
    // Only auto-click first agent if we're already on the agent-types view
    const agentTypesView = document.querySelector('.view-content[data-view-content="agent-types"]');
    if (agentTypesView && agentTypesView.classList.contains('is-active')) {
        const firstAgent = document.querySelector('.agent-type-item');
        if (firstAgent) {
            setTimeout(() => {
                firstAgent.click();
            }, 300);
        }
    }
}

// Initialize view switcher tabs
function initializeViewSwitcher() {
    console.log('Initializing view switcher...');
    const viewTabs = document.querySelectorAll('.view-tab');

    viewTabs.forEach(tab => {
        tab.addEventListener('click', function () {
            console.log('View tab clicked:', this.getAttribute('data-view'));

            // Remove active class from all tabs
            viewTabs.forEach(t => t.classList.remove('is-active'));

            // Add active class to clicked tab
            this.classList.add('is-active');

            // Get the selected view
            const view = this.getAttribute('data-view');

            // Handle view switching
            switchView(view);
        });
    });

    console.log('View switcher initialized. Active views:');
    document.querySelectorAll('.view-content').forEach(vc => {
        console.log(`  - ${vc.getAttribute('data-view-content')}: ${vc.classList.contains('is-active') ? 'ACTIVE' : 'inactive'}`);
    });
}

// Switch between different views
function switchView(view) {
    console.log('Switching to view:', view);

    // Remove is-active from all view content containers
    const allViewContents = document.querySelectorAll('.view-content');
    allViewContents.forEach(content => content.classList.remove('is-active'));

    // Add is-active to the selected view content
    const selectedViewContent = document.querySelector(`.view-content[data-view-content="${view}"]`);
    if (selectedViewContent) {
        selectedViewContent.classList.add('is-active');
    }

    // Handle specific view logic
    switch (view) {
        case 'overview':
            console.log('Showing overview');
            // Overview is always available
            break;
        case 'agent-types':
            console.log('Showing agent types');
            // Re-select first agent if needed
            const firstAgent = document.querySelector('.agent-type-item');
            if (firstAgent && !document.querySelector('.agent-type-item.is-active')) {
                firstAgent.click();
            }
            break;
        case 'layout':
            console.log('Showing layout diagram');
            // Layout diagram will be rendered here
            break;
    }
}

// Initialize HTMX event listeners
function initializeHTMXEvents() {
    // Show typing indicator when request starts
    document.body.addEventListener('htmx:beforeRequest', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            indicator.style.display = 'block';

            // Scroll to show typing indicator
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                setTimeout(() => scrollToBottom(chatContainer), 100);
            }
        }
    });

    // Hide typing indicator and scroll when new message arrives
    document.body.addEventListener('htmx:afterSwap', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.target.id === 'chat-messages') {
            indicator.style.display = 'none';
        }

        // Scroll to bottom to show new message
        const chatContainer = document.getElementById('chat-messages');
        if (chatContainer && evt.detail.target.id === 'chat-messages') {
            setTimeout(() => scrollToBottom(chatContainer), 100);
        }

        // Re-initialize agent selection if sidebar was updated
        if (evt.detail.target.closest('.sidebar-content')) {
            initializeAgentSelection();
        }
    });

    // Handle errors
    document.body.addEventListener('htmx:responseError', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.style.display = 'none';
        }

        // Show error message
        console.error('Request failed:', evt.detail);

        // Show error in UI
        const target = evt.detail.target;
        if (target) {
            const errorMsg = document.createElement('div');
            errorMsg.className = 'notification is-danger is-light';
            errorMsg.textContent = 'Request failed. Please try again.';
            target.appendChild(errorMsg);

            setTimeout(() => errorMsg.remove(), 3000);
        }
    });

    // Clear input after successful send
    document.body.addEventListener('htmx:afterRequest', function (evt) {
        if (evt.detail.successful && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            const input = evt.detail.elt.querySelector('input[name="message"]');
            if (input) {
                input.value = '';
                input.focus();
            }
        }
    });

    // Handle Enter key to submit
    document.body.addEventListener('keydown', function (evt) {
        const input = evt.target;
        if (input.matches('input[name="message"]') && evt.key === 'Enter' && !evt.shiftKey) {
            evt.preventDefault();
            const form = input.closest('form');
            if (form && typeof htmx !== 'undefined') {
                // Trigger HTMX submit
                htmx.trigger(form, 'submit');
            }
        }
    });
}

// Export for use in templates
window.selectAgentType = selectAgentType;
window.selectOverviewSection = selectOverviewSection;
window.saveOverviewIntroduction = saveOverviewIntroduction;
window.undoOverviewIntroduction = undoOverviewIntroduction;
window.showProblemEditor = showProblemEditor;
window.saveProblemFromEditor = saveProblemFromEditor;
window.cancelProblemEdit = cancelProblemEdit;
window.deleteProblem = deleteProblem;
window.showUnitEditor = showUnitEditor;
window.saveUnitFromEditor = saveUnitFromEditor;
window.cancelUnitEdit = cancelUnitEdit;
window.deleteUnit = deleteUnit;

// Handle overview section selection
function selectOverviewSection(element, section) {
    // Remove active class from all overview nav items
    const allItems = document.querySelectorAll('.overview-nav-item');
    allItems.forEach(item => item.classList.remove('is-active'));

    // Add active class to selected item
    element.classList.add('is-active');

    // Update the title
    const overviewTitle = document.getElementById('overview-title');

    // Update title based on section
    const titles = {
        'introduction': '<span class="icon"><i class="fas fa-info-circle"></i></span><span>Introduction</span>',
        'problem-definition': '<span class="icon"><i class="fas fa-exclamation-triangle"></i></span><span>Problem Definition</span>',
        'units-of-work': '<span class="icon"><i class="fas fa-clipboard-list"></i></span><span>Units of Work</span>'
    };

    if (titles[section] && overviewTitle) {
        overviewTitle.innerHTML = titles[section];
    }

    // Hide all content sections
    const allSections = document.querySelectorAll('.overview-content-section');
    allSections.forEach(sec => {
        sec.style.display = 'none';
        sec.classList.remove('is-active');
    });

    // Show the selected section
    const selectedSection = document.getElementById(`content-${section}`);
    if (selectedSection) {
        selectedSection.style.display = 'block';
        selectedSection.classList.add('is-active');

        // Load data if needed
        if (section === 'introduction') {
            loadIntroductionEditor();
        } else if (section === 'problem-definition') {
            loadProblems();
        } else if (section === 'units-of-work') {
            loadUnits();
        }
    }

    console.log('Selected overview section:', section);
}

// Load introduction editor and data
// Store original introduction for undo
let originalIntroduction = '';

function loadIntroductionEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    // Fetch the current overview/introduction
    fetch(`/api/v1/agencies/${agencyId}/overview`)
        .then(response => {
            if (!response.ok) {
                // If 404 or error, just show empty editor
                return { introduction: '' };
            }
            return response.json();
        })
        .then(data => {
            const editor = document.getElementById('introduction-editor');
            if (editor) {
                const introText = data.introduction || '';
                editor.value = introText;
                // Store original value for undo
                originalIntroduction = introText;
            }
        })
        .catch(error => {
            console.error('Error loading introduction:', error);
        });
}

// Save overview introduction
function saveOverviewIntroduction() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        console.error('Introduction editor not found');
        return;
    }

    const introduction = editor.value;
    const saveBtn = document.getElementById('save-introduction-btn');

    // Disable button while saving
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
        saveBtn.disabled = true;
    }

    fetch(`/api/v1/agencies/${agencyId}/overview`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ introduction: introduction })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to save introduction');
            }
            return response.json();
        })
        .then(data => {
            console.log('Introduction saved successfully:', data);
            showNotification('Introduction saved successfully!', 'success');
            // Update original value after successful save
            originalIntroduction = editor.value;
        })
        .catch(error => {
            console.error('Error saving introduction:', error);
            showNotification('Error saving introduction', 'error');
        })
        .finally(() => {
            // Re-enable button
            if (saveBtn) {
                saveBtn.classList.remove('is-loading');
                saveBtn.disabled = false;
            }
        });
}

// Undo changes to overview introduction
function undoOverviewIntroduction() {
    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        console.error('Introduction editor not found');
        return;
    }

    // Restore original value
    editor.value = originalIntroduction;
    showNotification('Changes reverted', 'info');
}

// Problem Management Functions
let currentEditingProblem = null;

// Load problems list
function loadProblems() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const problemsList = document.getElementById('problems-list');
    if (!problemsList) {
        console.error('Problems list container not found');
        return;
    }

    // Show loading state
    problemsList.innerHTML = '<div class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading problems...</p></div>';

    // Fetch problems HTML from API
    fetch(`/api/v1/agencies/${agencyId}/problems/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load problems');
            }
            return response.text();
        })
        .then(html => {
            problemsList.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading problems:', error);
            problemsList.innerHTML = '<div class="has-text-danger has-text-centered py-5"><p>Error loading problems</p></div>';
        });
}

// Problem editor state management
let problemEditorState = {
    mode: 'add', // 'add' or 'edit'
    problemKey: null,
    originalDescription: ''
};

// Show problem editor
function showProblemEditor(mode, problemKey = null, description = '') {
    problemEditorState.mode = mode;
    problemEditorState.problemKey = problemKey;
    problemEditorState.originalDescription = description;

    const editorCard = document.getElementById('problem-editor-card');
    const listCard = document.getElementById('problems-list-card');
    const editorTitle = document.getElementById('problem-editor-title');
    const descriptionEditor = document.getElementById('problem-description-editor');

    if (!editorCard || !listCard || !editorTitle || !descriptionEditor) {
        console.error('Problem editor elements not found');
        return;
    }

    // Update editor title and content
    if (mode === 'add') {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-plus"></i></span><span>Add New Problem</span>';
        descriptionEditor.value = '';
    } else {
        editorTitle.innerHTML = '<span class="icon"><i class="fas fa-edit"></i></span><span>Edit Problem</span>';
        descriptionEditor.value = description;
    }

    // Show editor, hide list
    editorCard.style.display = 'block';
    listCard.style.display = 'none';

    // Focus on editor
    descriptionEditor.focus();
}

// Save problem from editor
function saveProblemFromEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const descriptionEditor = document.getElementById('problem-description-editor');
    if (!descriptionEditor) {
        console.error('Description editor not found');
        return;
    }

    const description = descriptionEditor.value.trim();
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
        body: JSON.stringify({ description: description })
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
function cancelProblemEdit() {
    const editorCard = document.getElementById('problem-editor-card');
    const listCard = document.getElementById('problems-list-card');
    const descriptionEditor = document.getElementById('problem-description-editor');

    if (editorCard) editorCard.style.display = 'none';
    if (listCard) listCard.style.display = 'block';
    if (descriptionEditor) descriptionEditor.value = '';

    // Reset state
    problemEditorState = {
        mode: 'add',
        problemKey: null,
        originalDescription: ''
    };
}

// Delete problem
function deleteProblem(problemKey, problemNumber) {
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

// ===========================
// Units of Work Functions
// ===========================

// Load units of work from the server
function loadUnits() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    fetch(`/api/v1/agencies/${agencyId}/units/html`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load units');
            }
            return response.text();
        })
        .then(html => {
            const unitsListContainer = document.getElementById('units-list-container');
            if (unitsListContainer) {
                unitsListContainer.innerHTML = html;
            }
        })
        .catch(error => {
            console.error('Error loading units:', error);
            showNotification('Error loading units of work', 'error');
        });
}

// Show unit editor (add or edit mode)
function showUnitEditor(mode, unitKey = null, description = '') {
    const editorCard = document.getElementById('unit-editor-card');
    const listCard = document.getElementById('units-list-card');
    const editorTitle = document.getElementById('unit-editor-title');
    const descriptionInput = document.getElementById('unit-description-input');
    const saveButton = document.getElementById('save-unit-button');

    if (!editorCard || !listCard || !editorTitle || !descriptionInput || !saveButton) {
        console.error('Unit editor elements not found');
        return;
    }

    // Store editor state
    window.unitEditorState = {
        mode: mode,
        unitKey: unitKey
    };

    // Update editor UI
    if (mode === 'add') {
        editorTitle.textContent = 'Add Unit of Work';
        descriptionInput.value = '';
        saveButton.textContent = 'Add Unit';
    } else {
        editorTitle.textContent = 'Edit Unit of Work';
        descriptionInput.value = description;
        saveButton.textContent = 'Save Changes';
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');
    descriptionInput.focus();
}

// Cancel unit editing
function cancelUnitEdit() {
    const editorCard = document.getElementById('unit-editor-card');
    const listCard = document.getElementById('units-list-card');

    if (editorCard && listCard) {
        editorCard.classList.add('is-hidden');
        listCard.classList.remove('is-hidden');
    }

    // Clear editor state
    window.unitEditorState = null;
}

// Save unit from editor
function saveUnitFromEditor() {
    const descriptionInput = document.getElementById('unit-description-input');
    const description = descriptionInput.value.trim();

    if (!description) {
        showNotification('Please enter a unit description', 'warning');
        descriptionInput.focus();
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const editorState = window.unitEditorState;
    if (!editorState) {
        showNotification('Error: Editor state not found', 'error');
        return;
    }

    const url = editorState.mode === 'add'
        ? `/api/v1/agencies/${agencyId}/units`
        : `/api/v1/agencies/${agencyId}/units/${editorState.unitKey}`;

    const method = editorState.mode === 'add' ? 'POST' : 'PUT';

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ description: description })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to ${editorState.mode} unit`);
            }
            return response.json();
        })
        .then(() => {
            showNotification(`Unit ${editorState.mode === 'add' ? 'added' : 'updated'} successfully!`, 'success');
            cancelUnitEdit();
            loadUnits(); // Reload the list
        })
        .catch(error => {
            console.error(`Error ${editorState.mode}ing unit:`, error);
            showNotification(`Error ${editorState.mode}ing unit`, 'error');
        });
}

// Delete a unit of work
function deleteUnit(unitKey, unitNumber) {
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

// Helper function to escape HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Helper function to escape text for JavaScript strings
function escapeForJS(text) {
    return text.replace(/'/g, "\\'").replace(/"/g, '\\"').replace(/\n/g, '\\n').replace(/\r/g, '\\r');
}

// Get current agency ID from URL or context
function getCurrentAgencyId() {
    // Try to get from URL path /agencies/:id/designer
    const pathMatch = window.location.pathname.match(/\/agencies\/([^\/]+)/);
    if (pathMatch && pathMatch[1]) {
        return pathMatch[1];
    }
    return null;
}

// Show notification to user
function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification is-${type}`;
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '9999';
    notification.style.minWidth = '300px';
    notification.innerHTML = `
        <button class="delete"></button>
        ${message}
    `;

    document.body.appendChild(notification);

    // Add delete functionality
    const deleteBtn = notification.querySelector('.delete');
    if (deleteBtn) {
        deleteBtn.addEventListener('click', () => {
            notification.remove();
        });
    }

    // Auto-remove after 3 seconds
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transition = 'opacity 0.3s';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Handle design preview updates
document.body.addEventListener('htmx:afterSwap', function (evt) {
    if (evt.detail.target.id === 'design-preview') {
        // Animate preview update
        evt.detail.target.classList.add('updated');
        setTimeout(() => {
            evt.detail.target.classList.remove('updated');
        }, 500);
    }
});

// Add visual feedback for phase transitions
function highlightActivePhase() {
    const phases = document.querySelectorAll('.phase-step');
    phases.forEach(phase => {
        if (phase.classList.contains('is-active')) {
            phase.style.transform = 'scale(1.1)';
            setTimeout(() => {
                phase.style.transform = 'scale(1)';
            }, 300);
        }
    });
}

// Call on page load
document.addEventListener('DOMContentLoaded', highlightActivePhase);

// Smooth scroll behavior for chat
if (document.getElementById('chat-messages')) {
    document.getElementById('chat-messages').style.scrollBehavior = 'smooth';
}
