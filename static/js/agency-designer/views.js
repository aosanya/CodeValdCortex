// View switching functionality
// Handles navigation between different views (overview, agent-types, layout)

// Initialize view switcher tabs
export function initializeViewSwitcher() {
    console.log('Initializing view switcher...');
    const viewTabs = document.querySelectorAll('.view-tab');

    viewTabs.forEach(tab => {
        tab.addEventListener('click', function () {
            console.log('View tab clicked:', this.getAttribute('data-view'));

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

    console.log('View switcher initialized. Active views:');
    document.querySelectorAll('.view-content').forEach(vc => {
        console.log(`  - ${vc.getAttribute('data-view-content')}: ${vc.classList.contains('is-active') ? 'ACTIVE' : 'inactive'}`);
    });
}

// Switch between different views
export function switchView(view) {
    console.log('Switching to view:', view);

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
            console.log('Showing overview');
            // Overview is always available
            break;
        case 'agent-types':
            console.log('Showing agent types');
            // Re-select first agent if needed
            const firstAgent = document.querySelector('.agent-type-item');
            if (firstAgent && !document.querySelector('.agent-type-item.is-active')) {
                firstAgent.click();
            }
            break;
        case 'layout':
            console.log('Showing layout diagram');
            // Layout diagram will be rendered here
            break;
    }
}