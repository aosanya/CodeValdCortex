// Utility functions
// Common helper functions used across modules

// Get current agency ID from URL or data attributes
window.getCurrentAgencyId = function () {
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

    return null;
}

// Show notification message in status bar
window.showNotification = function (message, type = 'info') {
    // Wait for DOM to be ready if needed
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.showNotification(message, type);
        });
        return;
    }

    // Find the status bar right section
    const statusBarRight = document.querySelector('.vscode-status-bar .status-bar-right');

    if (!statusBarRight) {

        // Fallback to old notification method if status bar not found
        const notification = document.createElement('div');
        notification.className = `notification is-${type} is-light`;
        notification.style.cssText = 'position: fixed; top: 20px; left: 50%; transform: translateX(-50%); z-index: 9999; min-width: 300px;';
        notification.innerHTML = `
            <button class="delete" onclick="this.parentElement.remove()"></button>
            ${message}
        `;
        document.body.appendChild(notification);
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 5000);
        return;
    }

    // Clear any existing notification in status bar
    const existingNotification = statusBarRight.querySelector('.status-notification');
    if (existingNotification) {
        existingNotification.remove();
    }

    // Map notification types to icons and colors
    const typeConfig = {
        'success': { icon: 'fa-check-circle', class: 'has-text-success' },
        'error': { icon: 'fa-exclamation-circle', class: 'has-text-danger' },
        'warning': { icon: 'fa-exclamation-triangle', class: 'has-text-warning' },
        'info': { icon: 'fa-info-circle', class: 'has-text-info' }
    };

    const config = typeConfig[type] || typeConfig['info'];

    // Create status bar notification
    const statusNotification = document.createElement('span');
    statusNotification.className = `status-item status-notification ${config.class}`;
    statusNotification.innerHTML = `
        <i class="fas ${config.icon}"></i>
        <span class="status-text">${message}</span>
    `;

    // Add to status bar
    statusBarRight.appendChild(statusNotification);

    // Auto-remove after 5 seconds with fade out
    setTimeout(() => {
        if (statusNotification.parentElement) {
            statusNotification.style.opacity = '0';
            statusNotification.style.transition = 'opacity 0.3s ease-out';
            setTimeout(() => {
                statusNotification.remove();
            }, 300);
        }
    }, 5000);
}

// HTML escaping utility
window.escapeHtml = function (text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}