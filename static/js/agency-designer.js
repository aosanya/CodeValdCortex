// Agency Designer - Modular Version Entry Point
// This file loads the modular agency designer components

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

// Add handleRefineClick function for the button
// This function is called when the "Refine" button is clicked
// It shows the AI process status indicator
// The actual introduction text is taken from the textarea by HTMX via hx-include
window.handleRefineClick = function () {
    // Show AI processing status
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus('AI is refining your introduction...');
    }

    // Note: The current textarea value is automatically included in the POST request
    // by HTMX via the hx-include="#introduction-editor" attribute
    // No need to manually read or send the textarea value here
};

// Since browsers don't fully support ES6 modules without bundling,
// we'll create a simple loader that imports all functionality

// Import main module which coordinates everything
import('./agency-designer/main.js').then((module) => {
    // Module loaded successfully
}).catch(error => {
    // Error loading modules
});