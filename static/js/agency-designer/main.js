// AI Agency Designer - Main Entry Point
// Handles initialization and module coordination
// Uses global functions from other script files loaded before this one

// Check if DOM is already loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeAgencyDesigner);
} else {
    initializeAgencyDesigner();
}

function initializeAgencyDesigner() {
    try {
        // Initialize all modules using global functions
        if (window.initializeChatScroll) window.initializeChatScroll();
        if (window.initializeHTMXEvents) window.initializeHTMXEvents();
        if (window.initializeViewSwitcher) window.initializeViewSwitcher();
        if (window.initializeOverview) window.initializeOverview();

        // Listen for introduction updates from chat refinement
        document.body.addEventListener('introductionUpdated', function () {
            console.log('[Main] Introduction updated event received - reloading editor');
            if (window.loadIntroductionEditor) window.loadIntroductionEditor();
        });

        // Listen for goals updates from chat processing
        document.body.addEventListener('goalsUpdated', function () {
            if (typeof loadGoals === 'function') {
                loadGoals();
            }
        });

        loadIntroductionEditor(); // Initialize introduction editor
        initializeAIProcessControls();
        if (typeof window.initializeContextSelection === 'function') {
            window.initializeContextSelection(); // Initialize context selection system
        } else {
            console.warn('initializeContextSelection function not available');
        }
    } catch (error) {
        console.error('❌ Error during initialization:', error);
        console.error('Error stack:', error.stack);
    }
}

// Initialize AI process controls (stop button, etc.)
function initializeAIProcessControls() {
    const stopButton = document.getElementById('stop-ai-process');

    if (stopButton) {
        stopButton.addEventListener('click', function () {
            stopAIProcess();
        });
    }
}

// Stop AI processing
function stopAIProcess() {
    // Hide the AI process status bar
    hideAIProcessStatus();

    // Try to abort any ongoing HTMX requests
    if (window.htmx) {
        // Get all elements with active HTMX requests and abort them
        const elements = document.querySelectorAll('[hx-indicator="#ai-process-status"]');
        elements.forEach(element => {
            if (element.classList.contains('htmx-request')) {
                htmx.trigger(element, 'htmx:abort');
            }
        });
    }

    // Show notification
    if (window.showNotification) {
        showNotification('AI process stopped', 'warning');
    }
}

// Show AI process status with custom message
function showAIProcessStatus(message = 'AI is working on your request...') {
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

        // No timeout - status will remain visible until explicitly hidden
    }
}// Hide AI process status
function hideAIProcessStatus() {
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
}

// Functions are defined globally in their respective files
// No need to re-assign them here

// Export AI process control functions
window.showAIProcessStatus = showAIProcessStatus;
window.hideAIProcessStatus = hideAIProcessStatus;
window.stopAIProcess = stopAIProcess;

// Create AgencyDesigner namespace for utility functions
window.AgencyDesigner = {
    getCurrentTab: function () {
        // Check which overview section is active
        const activeSection = document.querySelector('.overview-section-button.is-active');
        if (activeSection) {
            const section = activeSection.getAttribute('data-section');
            return section || 'introduction';
        }

        // Fallback to checking view tabs
        const activeViewTab = document.querySelector('.view-tab.is-active');
        if (activeViewTab) {
            return activeViewTab.getAttribute('data-view') || 'overview';
        }

        // Default to introduction
        return 'introduction';
    }
};