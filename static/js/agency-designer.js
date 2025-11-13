// Agency Designer - Modular Version Entry Point
// This file loads the modular agency designer components

// Define AI status functions immediately to ensure they're always available
window.showAIProcessStatus = function (message = 'AI is working on your request...') {
    const processStatus = document.getElementById('ai-process-status');
    const statusMessage = document.getElementById('ai-status-message');

    // Update message text
    if (statusMessage) {
        statusMessage.textContent = message;
    }

    // Show the process status bar
    if (processStatus) {
        processStatus.style.display = 'flex';
        processStatus.style.visibility = 'visible';
        // Remove htmx-indicator class temporarily to prevent HTMX from hiding it
        processStatus.classList.remove('htmx-indicator');

        // Clear any existing timeout
        if (processStatus._hideTimeout) {
            clearTimeout(processStatus._hideTimeout);
            processStatus._hideTimeout = null;
        }

    } else {
    }
};

window.hideAIProcessStatus = function () {
    const processStatus = document.getElementById('ai-process-status');

    if (processStatus) {
        // Clear any existing timeout
        if (processStatus._hideTimeout) {
            clearTimeout(processStatus._hideTimeout);
            processStatus._hideTimeout = null;
        }

        processStatus.style.display = 'none';
        // Add htmx-indicator class back
        processStatus.classList.add('htmx-indicator');

    }
};

// Add HTMX event listeners directly here as a fallback

// HTMX afterSwap event listener - direct implementation
document.body.addEventListener('htmx:afterSwap', function (evt) {
    // Hide AI process status for introduction content updates
    const shouldHideStatus = (
        evt.detail.target.id === 'chat-messages' ||
        evt.detail.target.id === 'design-preview' ||
        evt.detail.target.id === 'introduction-content' ||
        evt.detail.target.classList.contains('introduction-content') ||
        evt.detail.target.closest('.details-content')
    );

    if (shouldHideStatus) {
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        } else {
            const status = document.getElementById('ai-process-status');
            if (status) {
                status.style.display = 'none';
            }
        }
    }
});

// Add global manual hide function for debugging
window.manualHideStatus = function () {

    // Try multiple possible status elements
    const possibleIds = [
        'ai-process-status',
        'ai-status',
        'process-status',
        'chat-loading-indicator',
        'ai-refine-loading'
    ];

    let found = false;

    possibleIds.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            element.style.display = 'none';
            element.style.visibility = 'hidden';
            found = true;
        }
    });

    // Also try class-based selectors
    const possibleClasses = [
        '.ai-process-status',
        '.htmx-indicator',
        '.process-status',
        '.ai-status'
    ];

    possibleClasses.forEach(className => {
        const elements = document.querySelectorAll(className);
        if (elements.length > 0) {
            elements.forEach((element, index) => {
                element.style.display = 'none';
                element.style.visibility = 'hidden';
                found = true;
            });
        }
    });

    if (!found) {
        // List all visible elements that might be the status
        const allVisible = document.querySelectorAll('*:not([style*="display: none"]):not([style*="display:none"])');
    }

    return found;
};

// Since browsers don't fully support ES6 modules without bundling,
// we'll create a simple loader that imports all functionality

// Import main module which coordinates everything
import('./agency-designer/main.js').then((module) => {
    // Module loaded successfully
}).catch(error => {
});

// Overview functions are loaded directly as global functions from overview.js script
