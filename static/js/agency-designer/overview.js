// Overview section functionality
// Handles overview navigation and section switching

import { loadIntroductionEditor } from './introduction.js';
import { loadProblems } from './problems.js';
import { loadUnits } from './units.js';

// Initialize overview section
export function initializeOverview() {
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
}