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

// Handle overview section selection
function selectOverviewSection(element, section) {
    // Remove active class from all overview nav items
    const allItems = document.querySelectorAll('.overview-nav-item');
    allItems.forEach(item => item.classList.remove('is-active'));

    // Add active class to selected item
    element.classList.add('is-active');

    // Update the content area
    const overviewContent = document.getElementById('overview-content');
    const overviewTitle = document.getElementById('overview-title');

    if (!overviewContent || !overviewTitle) return;

    // Update title based on section
    const titles = {
        'introduction': '<span class="icon"><i class="fas fa-info-circle"></i></span><span>Introduction</span>',
        'problem-definition': '<span class="icon"><i class="fas fa-exclamation-triangle"></i></span><span>Problem Definition</span>',
        'requirements': '<span class="icon"><i class="fas fa-clipboard-list"></i></span><span>Requirements</span>'
    };

    if (titles[section]) {
        overviewTitle.innerHTML = titles[section];
    }

    // Load the section content based on section type
    if (section === 'introduction') {
        loadIntroductionEditor();
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
