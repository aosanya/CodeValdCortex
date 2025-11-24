/**
 * Shared State and Utilities
 * 
 * Global state and helper functions used across agency designer modules
 * This file should be loaded before other agency-designer scripts
 */

// Global agency context
let currentAgencyId = null;

/**
 * Initialize shared state from URL
 */
function initializeSharedState() {
    const urlParts = window.location.pathname.split('/');
    const agencyIndex = urlParts.indexOf('agencies');

    if (agencyIndex !== -1 && urlParts.length > agencyIndex + 1) {
        currentAgencyId = urlParts[agencyIndex + 1];
    }
}

/**
 * Get current agency ID
 * @returns {string|null} Current agency ID
 */
function getCurrentAgencyId() {
    if (!currentAgencyId) {
        initializeSharedState();
    }
    return currentAgencyId;
}

/**
 * Set current agency ID (useful for testing or manual override)
 * @param {string} agencyId - Agency ID to set
 */
function setCurrentAgencyId(agencyId) {
    currentAgencyId = agencyId;
}

// Initialize on load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeSharedState);
} else {
    initializeSharedState();
}
