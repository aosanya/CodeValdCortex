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

    console.log('Agency Designer: Initialization complete');
});

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