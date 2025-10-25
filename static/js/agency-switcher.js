// Agency Switcher and Filtering Logic

document.addEventListener('DOMContentLoaded', () => {
    // Initialize agency filtering
    initializeAgencyFiltering();

    // Initialize agency switcher
    initializeAgencySwitcher();

    // Initialize agency selection
    initializeAgencySelection();
});

/**
 * Initialize agency filtering functionality
 */
function initializeAgencyFiltering() {
    const searchInput = document.getElementById('agency-search');
    const categoryFilter = document.getElementById('category-filter');
    const statusFilter = document.getElementById('status-filter');
    const sortBy = document.getElementById('sort-by');
    const agencyGrid = document.getElementById('agency-grid');

    if (!agencyGrid) return;

    let agencies = Array.from(agencyGrid.querySelectorAll('.agency-card-column'));

    /**
     * Filter agencies based on search, category, and status
     */
    function filterAgencies() {
        const searchTerm = searchInput ? searchInput.value.toLowerCase() : '';
        const category = categoryFilter ? categoryFilter.value : '';
        const status = statusFilter ? statusFilter.value : '';

        agencies.forEach(card => {
            const name = card.dataset.name ? card.dataset.name.toLowerCase() : '';
            const description = card.dataset.description ? card.dataset.description.toLowerCase() : '';
            const cardCategory = card.dataset.category || '';
            const cardStatus = card.dataset.status || '';

            const matchesSearch = searchTerm === '' ||
                name.includes(searchTerm) ||
                description.includes(searchTerm);
            const matchesCategory = category === '' || cardCategory === category;
            const matchesStatus = status === '' || cardStatus === status;

            if (matchesSearch && matchesCategory && matchesStatus) {
                card.style.display = '';
                card.classList.remove('is-hidden');
            } else {
                card.style.display = 'none';
                card.classList.add('is-hidden');
            }
        });

        // Show empty state if no results
        showEmptyStateIfNeeded();
    }

    /**
     * Sort agencies based on selected criteria
     */
    function sortAgencies(sortValue) {
        const visibleAgencies = agencies.filter(card =>
            card.style.display !== 'none' && !card.classList.contains('is-hidden')
        );

        visibleAgencies.sort((a, b) => {
            switch (sortValue) {
                case 'name':
                    return (a.dataset.name || '').localeCompare(b.dataset.name || '');
                case 'category':
                    return (a.dataset.category || '').localeCompare(b.dataset.category || '');
                case 'agents':
                    // Extract agent count from tags (would need data attribute in production)
                    return 0; // Placeholder
                case 'recent':
                    // Would need timestamp data attribute
                    return 0; // Placeholder
                default:
                    return 0;
            }
        });

        // Reorder DOM elements
        visibleAgencies.forEach(card => {
            agencyGrid.appendChild(card);
        });
    }

    /**
     * Show empty state message if no agencies match filters
     */
    function showEmptyStateIfNeeded() {
        const visibleCount = agencies.filter(card =>
            card.style.display !== 'none' && !card.classList.contains('is-hidden')
        ).length;

        let emptyState = agencyGrid.querySelector('.empty-state');

        if (visibleCount === 0 && !emptyState) {
            emptyState = document.createElement('div');
            emptyState.className = 'column is-full empty-state';
            emptyState.innerHTML = `
				<div class="notification is-warning">
					<p class="has-text-centered is-size-5">No agencies match your filters</p>
					<p class="has-text-centered">Try adjusting your search or filter criteria.</p>
				</div>
			`;
            agencyGrid.appendChild(emptyState);
        } else if (visibleCount > 0 && emptyState) {
            emptyState.remove();
        }
    }

    // Attach event listeners
    if (searchInput) {
        searchInput.addEventListener('input', filterAgencies);
    }

    if (categoryFilter) {
        categoryFilter.addEventListener('change', filterAgencies);
    }

    if (statusFilter) {
        statusFilter.addEventListener('change', filterAgencies);
    }

    if (sortBy) {
        sortBy.addEventListener('change', (e) => {
            sortAgencies(e.target.value);
        });
    }

    // Expose to Alpine.js if needed
    window.filterAgencies = filterAgencies;
    window.sortAgencies = sortAgencies;
}

/**
 * Initialize agency switcher modal
 */
function initializeAgencySwitcher() {
    const switchButton = document.getElementById('switch-agency-button');
    const switcherModal = document.getElementById('agency-switcher-modal');

    if (!switchButton || !switcherModal) return;

    // Open modal
    switchButton.addEventListener('click', (e) => {
        e.preventDefault();
        switcherModal.classList.add('is-active');
    });

    // Close modal
    const closeButtons = switcherModal.querySelectorAll('.modal-close, .modal-background, [data-dismiss="modal"]');
    closeButtons.forEach(button => {
        button.addEventListener('click', () => {
            switcherModal.classList.remove('is-active');
        });
    });

    // Handle agency selection in modal
    const switcherCards = switcherModal.querySelectorAll('.agency-switcher-card');
    switcherCards.forEach(card => {
        card.addEventListener('click', () => {
            const agencyId = card.dataset.agencyId;
            if (agencyId) {
                selectAgencyFromModal(agencyId);
            }
        });
    });

    // Modal search functionality
    const modalSearch = document.getElementById('modal-agency-search');
    if (modalSearch) {
        modalSearch.addEventListener('input', (e) => {
            const searchTerm = e.target.value.toLowerCase();
            switcherCards.forEach(card => {
                const name = card.dataset.agencyName ? card.dataset.agencyName.toLowerCase() : '';
                if (name.includes(searchTerm)) {
                    card.style.display = '';
                } else {
                    card.style.display = 'none';
                }
            });
        });
    }
}/**
 * Initialize agency selection from homepage cards
 */
function initializeAgencySelection() {
    // Add loading state to cards when clicked
    const agencyCards = document.querySelectorAll('.agency-card');

    agencyCards.forEach(card => {
        const selectButton = card.closest('.card').querySelector('[hx-post]');
        if (selectButton) {
            selectButton.addEventListener('click', () => {
                card.classList.add('is-loading');
            });
        }
    });
}

/**
 * Select an agency and redirect to its dashboard
 */
async function selectAgency(agencyId) {
    try {
        // Show loading indicator
        showLoadingIndicator();

        // Make API call to select agency
        const response = await fetch(`/agencies/${agencyId}/select`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            throw new Error('Failed to select agency');
        }

        const data = await response.json();

        // Store selection in localStorage for "Remember my agency"
        const rememberAgency = localStorage.getItem('remember_agency') === 'true';
        if (rememberAgency) {
            localStorage.setItem('last_agency_id', agencyId);
        }

        // Redirect to agency dashboard
        window.location.href = `/agencies/${agencyId}/dashboard`;

    } catch (error) {
        console.error('Error selecting agency:', error);
        hideLoadingIndicator();
        showErrorNotification('Failed to select agency. Please try again.');
    }
}

/**
 * Select agency from modal (used by agency switcher)
 */
function selectAgencyFromModal(agencyId) {
    // Check for unsaved changes
    if (hasUnsavedChanges()) {
        if (!confirm('You have unsaved changes. Are you sure you want to switch agencies?')) {
            return;
        }
    }

    // Close modal
    const modal = document.getElementById('agency-switcher-modal');
    if (modal) {
        modal.classList.remove('is-active');
    }

    // Select the agency
    selectAgency(agencyId);
}

/**
 * Check if there are unsaved changes (placeholder - to be implemented per page)
 */
function hasUnsavedChanges() {
    // This should be overridden by individual pages that track changes
    // For now, return false
    return window.hasUnsavedChanges ? window.hasUnsavedChanges() : false;
}/**
 * Show loading indicator overlay
 */
function showLoadingIndicator() {
    let indicator = document.getElementById('loading-indicator');
    if (!indicator) {
        indicator = document.createElement('div');
        indicator.id = 'loading-indicator';
        indicator.className = 'modal is-active';
        indicator.innerHTML = `
			<div class="modal-background"></div>
			<div class="has-text-centered" style="z-index: 1000; position: relative;">
				<div class="loader" style="margin: 0 auto; width: 60px; height: 60px; border: 4px solid #dbdbdb; border-top-color: #3273dc; border-radius: 50%; animation: spin 0.8s linear infinite;"></div>
				<p class="has-text-white mt-4">Loading agency...</p>
			</div>
		`;
        document.body.appendChild(indicator);
    }
    indicator.classList.add('is-active');
}

/**
 * Hide loading indicator overlay
 */
function hideLoadingIndicator() {
    const indicator = document.getElementById('loading-indicator');
    if (indicator) {
        indicator.classList.remove('is-active');
    }
}

/**
 * Show error notification
 */
function showErrorNotification(message) {
    const notification = document.createElement('div');
    notification.className = 'notification is-danger';
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '10000';
    notification.style.minWidth = '300px';
    notification.innerHTML = `
		<button class="delete"></button>
		${message}
	`;

    document.body.appendChild(notification);

    const deleteButton = notification.querySelector('.delete');
    deleteButton.addEventListener('click', () => {
        notification.remove();
    });

    // Auto-remove after 5 seconds
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

/**
 * Check if user has a remembered agency and redirect
 */
function checkRememberedAgency() {
    const rememberAgency = localStorage.getItem('remember_agency') === 'true';
    const lastAgencyId = localStorage.getItem('last_agency_id');

    if (rememberAgency && lastAgencyId) {
        // Only auto-redirect if on homepage
        if (window.location.pathname === '/') {
            selectAgency(lastAgencyId);
        }
    }
}

// Check for remembered agency on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', checkRememberedAgency);
} else {
    checkRememberedAgency();
}

// Export functions for use in other scripts or inline handlers
window.AgencySwitcher = {
    selectAgency,
    selectAgencyFromModal,
    showLoadingIndicator,
    hideLoadingIndicator,
    showErrorNotification
};

// Also export selectAgencyFromModal globally for onclick handlers
window.selectAgencyFromModal = selectAgencyFromModal;