// Utility functions
// Common helper functions used across modules

// Get current agency ID from URL or data attributes
export function getCurrentAgencyId() {
    // Try to get from URL path (e.g., /agencies/UC-CHAR-001/designer)
    const pathMatch = window.location.pathname.match(/\/agencies\/([^\/]+)/);
    if (pathMatch) {
        return pathMatch[1];
    }

    // Try to get from data attribute
    const designerElement = document.querySelector('[data-agency-id]');
    if (designerElement) {
        return designerElement.getAttribute('data-agency-id');
    }

    // Try to get from meta tag
    const metaTag = document.querySelector('meta[name="agency-id"]');
    if (metaTag) {
        return metaTag.getAttribute('content');
    }

    console.warn('Could not determine agency ID');
    return null;
}

// Show notification message
export function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification is-${type} is-light`;
    notification.innerHTML = `
        <button class="delete" onclick="this.parentElement.remove()"></button>
        ${message}
    `;

    // Add to page
    const container = document.querySelector('.notifications-container') || document.body;
    container.appendChild(notification);

    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}