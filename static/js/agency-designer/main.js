// AI Agency Designer - Main Entry Point
// Handles initialization and module coordination

// Import all modules
import { initializeChatScroll, scrollToBottom } from './chat.js';
import { initializeHTMXEvents } from './htmx.js';
import { initializeViewSwitcher, switchView } from './views.js';
import { initializeAgentSelection, selectAgentType } from './agents.js';
import { initializeOverview, selectOverviewSection } from './overview.js';
import {
    saveOverviewIntroduction,
    undoOverviewIntroduction
} from './introduction.js';
import {
    showProblemEditor,
    saveProblemFromEditor,
    cancelProblemEdit,
    deleteProblem
} from './problems.js';
import {
    showUnitEditor,
    saveUnitFromEditor,
    cancelUnitEdit,
    deleteUnit
} from './units.js';
import { getCurrentAgencyId, showNotification } from './utils.js';

// Check if DOM is already loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeAgencyDesigner);
} else {
    initializeAgencyDesigner();
}

function initializeAgencyDesigner() {
    try {
        // Initialize all modules
        initializeChatScroll();
        initializeHTMXEvents();
        initializeViewSwitcher();
        initializeAgentSelection();
        initializeOverview();
        initializeAIProcessControls();
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

        // Add fallback timeout to hide status after 15 seconds
        setTimeout(() => {
            hideAIProcessStatus();
        }, 15000);
    }
}

// Hide AI process status
function hideAIProcessStatus() {
    const processStatus = document.getElementById('ai-process-status');

    if (processStatus) {
        processStatus.style.display = 'none';
        // Add htmx-indicator class back
        processStatus.classList.add('htmx-indicator');
    }
}

// Export functions to global scope for onclick handlers
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

// Export AI process control functions
window.showAIProcessStatus = showAIProcessStatus;
window.hideAIProcessStatus = hideAIProcessStatus;
window.stopAIProcess = stopAIProcess;