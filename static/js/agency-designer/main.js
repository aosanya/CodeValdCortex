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
import {
    refineCurrentDesign,
    requestAlternativeDesign
} from './refine.js';
import { getCurrentAgencyId, showNotification } from './utils.js';

document.addEventListener('DOMContentLoaded', function () {
    console.log('Agency Designer: Initializing...');

    // Log initial active view
    const activeView = document.querySelector('.view-content.is-active');
    if (activeView) {
        console.log('Initial active view:', activeView.getAttribute('data-view-content'));
    }

    // Initialize all modules
    initializeChatScroll();
    initializeHTMXEvents();
    initializeViewSwitcher();
    initializeAgentSelection();
    initializeOverview();
    initializeAIProcessControls();

    console.log('Agency Designer: Initialization complete');
});

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
    console.log('Stopping AI process...');

    // Hide the AI process status bar
    const processStatus = document.getElementById('ai-process-status');
    if (processStatus) {
        processStatus.style.display = 'none';
    }

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
    showNotification('AI process stopped', 'warning');
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
window.refineCurrentDesign = refineCurrentDesign;
window.requestAlternativeDesign = requestAlternativeDesign;
window.stopAIProcess = stopAIProcess;