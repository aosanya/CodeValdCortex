// Overview section functionality
// Handles overview navigation and section switching
// Uses global functions: loadIntroductionEditor, loadGoals, loadWorkItems, loadRoles, loadWorkflows

// Initialize overview section
window.initializeOverview = function () {
    // Initialize global context list and set default to introduction
    if (typeof window !== 'undefined') {
        if (!window.AGENCY_CONTEXTS) {
            window.AGENCY_CONTEXTS = [
                'introduction',
                'goal-definition',
                'work-items',
                'roles',
                'raci-matrix',
                'workflows'
            ];
        }

        // Check if there's a hash in the URL
        const hash = window.location.hash.slice(1); // Remove the '#'
        const validSections = window.AGENCY_CONTEXTS;

        // Set default context based on hash or default to introduction
        if (hash && validSections.includes(hash)) {
            window.currentAgencyContext = hash;

            // Update hidden context field in chat form
            const contextField = document.getElementById('chat-context-field');
            if (contextField) {
                contextField.value = hash;
            }

            // Update the context display
            const contextCurrentEl = document.getElementById('context-current');
            if (contextCurrentEl) {
                const labelMap = {
                    'introduction': 'Introduction',
                    'goal-definition': 'Goal Definition',
                    'work-items': 'Work Items',
                    'roles': 'Roles',
                    'raci-matrix': 'RACI Matrix',
                    'workflows': 'Workflows'
                };
                contextCurrentEl.textContent = labelMap[hash] || hash;
            }

            // Activate the correct nav item
            const navItem = document.querySelector(`.overview-nav-item[data-section="${hash}"]`);
            if (navItem) {
                selectOverviewSection(navItem, hash);
            }
        } else {
            // Set default context to introduction
            window.currentAgencyContext = 'introduction';
            window.location.hash = 'introduction';

            // Update hidden context field in chat form
            const contextField = document.getElementById('chat-context-field');
            if (contextField) {
                contextField.value = 'introduction';
            }

            // Update the context display to show Introduction
            const contextCurrentEl = document.getElementById('context-current');
            if (contextCurrentEl) {
                contextCurrentEl.textContent = 'Introduction';
            }
        }
    }

    // Check if we're on the overview view and introduction is active
    const overviewView = document.querySelector('.view-content[data-view-content="overview"]');
    const introEditor = document.getElementById('introduction-editor');

    if (overviewView && overviewView.classList.contains('is-active') && introEditor) {
        // Load introduction data if on introduction section
        if (!window.location.hash || window.location.hash === '#introduction') {
            if (window.loadIntroductionEditor) window.loadIntroductionEditor();
        }
    }
}

// Handle overview section selection
window.selectOverviewSection = function (element, section) {
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

        // Clear navigational contexts from the previous section
        const previousSection = window.currentAgencyContext;
        if (previousSection && previousSection !== section && window.ContextManager) {
            // Clear navigational contexts from the section we're leaving
            window.ContextManager.clearNavigationalContexts(previousSection);
        }

        // When entering work-items, clear ALL navigational contexts
        if (section === 'work-items' && window.ContextManager) {
            window.ContextManager.clearNavigationalContexts();
        }

        // Track current selected context (for backend calls to include as `context`)
        window.currentAgencyContext = section;

        // Update hidden context field in chat form immediately
        const contextField = document.getElementById('chat-context-field');
        if (contextField) {
            contextField.value = section;
        }
    }

    // Update URL hash
    window.location.hash = section;

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
            if (window.loadIntroductionEditor) window.loadIntroductionEditor();
        } else if (section === 'goal-definition') {
            if (window.loadGoals) window.loadGoals();
        } else if (section === 'work-items') {
            if (window.loadWorkItems) window.loadWorkItems();
        } else if (section === 'roles') {
            if (window.loadRoles) window.loadRoles();
        } else if (section === 'raci-matrix') {
            // Load RACI matrix data
            if (window.loadRACIMatrix) {
                window.loadRACIMatrix();
            }
        } else if (section === 'workflows') {
            if (window.loadWorkflows) window.loadWorkflows();
        }
    }
}

// Immediately export to window for inline onclick handlers
if (typeof window !== 'undefined') {
    window.selectOverviewSection = selectOverviewSection;
    window.initializeOverview = initializeOverview;
}
