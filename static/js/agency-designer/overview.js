// Overview section functionality
// Handles overview navigation and section switching

import { loadIntroductionEditor } from './introduction.js';
import { loadGoals } from './goals.js';
import { loadWorkItems } from './work-items.js';
import { loadRoles } from './roles.js';

// Initialize overview section
export function initializeOverview() {
    // Initialize global context list and set default to introduction
    if (typeof window !== 'undefined') {
        if (!window.AGENCY_CONTEXTS) {
            window.AGENCY_CONTEXTS = [
                'introduction',
                'goal-definition',
                'work-items',
                'roles',
                'raci-matrix'
            ];
        }

        // Set default context to introduction
        window.currentAgencyContext = 'introduction';

        // Update the context display to show Introduction
        const contextCurrentEl = document.getElementById('context-current');
        if (contextCurrentEl) {
            contextCurrentEl.textContent = 'Introduction';
        }
    }

    // Check if we're on the overview view and introduction is active
    const overviewView = document.querySelector('.view-content[data-view-content="overview"]');
    const introEditor = document.getElementById('introduction-editor');

    if (overviewView && overviewView.classList.contains('is-active') && introEditor) {
        // Load introduction data
        loadIntroductionEditor();
    }
}

// Handle overview section selection
export function selectOverviewSection(element, section) {
    // Ensure a global default context list exists
    if (typeof window !== 'undefined') {
        if (!window.AGENCY_CONTEXTS) {
            window.AGENCY_CONTEXTS = [
                'introduction',
                'goal-definition',
                'work-items',
                'roles',
                'raci-matrix'
            ];
        }

        // Track current selected context (for backend calls to include as `context`)
        window.currentAgencyContext = section;
    }
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
        'goal-definition': '<span class="icon"><i class="fas fa-bullseye"></i></span><span>Goal Definition</span>',
        'work-items': '<span class="icon"><i class="fas fa-clipboard-list"></i></span><span>Work Items</span>',
        'roles': '<span class="icon"><i class="fas fa-user-tag"></i></span><span>Roles</span>',
        'raci-matrix': '<span class="icon"><i class="fas fa-table-cells"></i></span><span>RACI Matrix</span>'
    };

    if (titles[section] && overviewTitle) {
        overviewTitle.innerHTML = titles[section];
    }

    // Update the small Context header display (if present)
    const contextCurrentEl = document.getElementById('context-current');
    if (contextCurrentEl) {
        // Strip HTML tags from the title mapping and set a readable label
        const labelMap = {
            'introduction': 'Introduction',
            'goal-definition': 'Goal Definition',
            'work-items': 'Work Items',
            'roles': 'Roles',
            'raci-matrix': 'RACI Matrix'
        };
        contextCurrentEl.textContent = labelMap[section] || section;
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
        } else if (section === 'goal-definition') {
            loadGoals();
        } else if (section === 'work-items') {
            loadWorkItems();
        } else if (section === 'roles') {
            loadRoles();
        } else if (section === 'raci-matrix') {
            // Load RACI matrix data
            if (window.loadRACIMatrix) {
                window.loadRACIMatrix();
            }
        }
    }
}
