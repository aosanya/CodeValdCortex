// View switching functionality
// Handles navigation between different views (overview, agent-types, layout)

// Initialize view switcher tabs
export function initializeViewSwitcher() {
    const viewTabs = document.querySelectorAll('.view-tab');

    viewTabs.forEach(tab => {
        tab.addEventListener('click', function () {
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
}

// Switch between different views
export function switchView(view) {
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
            // Overview is always available
            break;
        case 'layout':
            // Layout diagram will be rendered here
            break;
    }
}