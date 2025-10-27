// Agent selection functionality
// Handles agent type selection and details display

// Handle agent type selection in sidebar
export function selectAgentType(element) {
    // Remove active class from all items
    const allItems = document.querySelectorAll('.agent-type-item');
    allItems.forEach(item => item.classList.remove('is-active'));

    // Add active class to selected item
    element.classList.add('is-active');

    // Update details title
    const agentName = element.querySelector('.agent-type-name')?.textContent || 'Agent Details';
    const detailsTitle = document.getElementById('details-title');
    if (detailsTitle) {
        detailsTitle.innerHTML = `
            <span class="icon"><i class="fas fa-robot"></i></span>
            <span>${agentName}</span>
        `;
    }
}

// Initialize agent selection handlers
export function initializeAgentSelection() {
    // Only auto-click first agent if we're already on the agent-types view
    const agentTypesView = document.querySelector('.view-content[data-view-content="agent-types"]');
    if (agentTypesView && agentTypesView.classList.contains('is-active')) {
        const firstAgent = document.querySelector('.agent-type-item');
        if (firstAgent) {
            setTimeout(() => {
                firstAgent.click();
            }, 300);
        }
    }
}