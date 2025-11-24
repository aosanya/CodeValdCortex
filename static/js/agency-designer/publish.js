/**
 * Publish Dialog and Operations
 * 
 * Handles agency validation, publishing, and tag creation workflows
 * Interacts with /api/v1/agencies/:id/validate and /api/v1/agencies/:id/publish endpoints
 * 
 * Depends on: shared.js (for currentAgencyId)
 */

// State management (currentAgencyId is in shared.js)
let validationResult = null;

/**
 * Initialize publish functionality
 */
function initializePublishFeatures() {
    // currentAgencyId is already initialized in shared.js
}

/**
 * Open publish dialog and run validation check
 */
async function openPublishDialog() {
    const dialog = document.getElementById('publish-dialog');
    dialog.classList.add('is-active');

    // Reset form
    resetPublishForm();

    // Run validation check
    await checkValidation();
}

/**
 * Close publish dialog
 */
function closePublishDialog() {
    const dialog = document.getElementById('publish-dialog');
    dialog.classList.remove('is-active');
    validationResult = null;
}

/**
 * Reset publish form to initial state
 */
function resetPublishForm() {
    document.getElementById('publish-version').value = '';
    document.getElementById('publish-description').value = '';
    document.getElementById('publish-auto-activate').checked = false;
    document.getElementById('publish-tag-name').value = '';
    document.getElementById('publish-form').style.display = 'none';
    document.getElementById('validation-status-info').style.display = 'block';
    document.getElementById('validation-errors').style.display = 'none';
    document.getElementById('confirm-publish-btn').disabled = true;
}

/**
 * Check agency validation status
 */
async function checkValidation() {
    const statusDiv = document.getElementById('validation-status-info');
    const errorsDiv = document.getElementById('validation-errors');
    const errorList = document.getElementById('validation-error-list');
    const publishForm = document.getElementById('publish-form');

    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/validate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();
        validationResult = result;

        if (response.ok && result.valid) {
            // Valid - show publish form
            statusDiv.style.display = 'none';
            errorsDiv.style.display = 'none';
            publishForm.style.display = 'block';
            document.getElementById('confirm-publish-btn').disabled = false;
        } else {
            // Invalid - show errors
            statusDiv.style.display = 'none';
            errorsDiv.style.display = 'block';

            // Render error list
            if (result.errors && result.errors.length > 0) {
                errorList.innerHTML = '<ul class="is-size-7">' +
                    result.errors.map(err => `<li><strong>${err.field}:</strong> ${err.message}</li>`).join('') +
                    '</ul>';
            } else {
                errorList.innerHTML = '<p class="is-size-7">No specific errors provided.</p>';
            }

            publishForm.style.display = 'none';
        }
    } catch (error) {
        statusDiv.innerHTML = `
        <p class="has-text-danger">
            <span class="icon"><i class="fas fa-exclamation-circle"></i></span>
            Failed to check validation: ${error.message}
        </p>
    `;
    }
}/**
 * Handle publish form submission
 */
async function handlePublishSubmit() {
    const version = document.getElementById('publish-version').value.trim();
    const description = document.getElementById('publish-description').value.trim();
    const autoActivate = document.getElementById('publish-auto-activate').checked;
    const tagName = document.getElementById('publish-tag-name').value.trim() || version;

    // Validate required fields
    if (!version || !description) {
        alert('Version and description are required');
        return;
    }

    // Validate version format
    const versionPattern = /^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$/;
    if (!versionPattern.test(version)) {
        alert('Version must be in semantic versioning format (e.g., v1.0.0)');
        return;
    }

    // Show progress
    const publishBtn = document.getElementById('confirm-publish-btn');
    const progress = document.getElementById('publish-progress');
    publishBtn.disabled = true;
    progress.style.display = 'block';

    try {
        // Build publish request
        const publishData = {
            version: version,
            description: description,
            auto_activate: autoActivate
        };

        // Always create tag before publishing
        await createTagBeforePublish(tagName, version, description);

        // Publish agency
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/publish`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(publishData)
        });

        const result = await response.json();

        if (response.ok) {
            // Success - close dialog and show notification
            closePublishDialog();

            // Show success notification
            showNotification('Success', `Agency published as ${version}`, 'success');

            // Reload page to reflect new state
            setTimeout(() => {
                window.location.reload();
            }, 1500);
        } else {
            // Error
            throw new Error(result.error || 'Failed to publish agency');
        }
    } catch (error) {
        alert(`Failed to publish agency: ${error.message}`);
    } finally {
        publishBtn.disabled = false;
        progress.style.display = 'none';
    }
}/**
 * Create tag before publishing (or use existing if it already exists)
 */
async function createTagBeforePublish(tagName, version, description) {
    // Check if tag already exists
    const checkResponse = await fetch(`/api/v1/agencies/${currentAgencyId}/tags/${tagName}`);

    if (checkResponse.ok) {
        const existingTag = await checkResponse.json();
        return existingTag;
    }

    // Tag doesn't exist, create it
    const tagData = {
        name: tagName,
        type: 'snapshot',
        description: `Pre-publish snapshot: ${description}`,
        metadata: {
            created_before_publish: 'true',
            version: version
        }
    };

    const response = await fetch(`/api/v1/agencies/${currentAgencyId}/tags`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(tagData)
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(`Failed to create tag: ${error.error || error.details || 'Unknown error'}`);
    }

    const result = await response.json();
    return result;
}

/**
 * Handle validate button click (Draft → Validated)
 */
async function handleValidateAgency() {
    const btn = document.getElementById('validate-btn');
    btn.classList.add('is-loading');

    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/validate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        // Navigate to validation section to show results
        navigateToValidationSection();

        // Display validation results
        displayValidationResults(result, response.ok);

        if (response.ok && result.is_valid) {
            showNotification('Success', 'Agency validated successfully', 'success');

            // Transition state to validated
            await updateAgencyState('validated');

            // Reload to show new buttons
            setTimeout(() => window.location.reload(), 1000);
        } else {
            showNotification('Validation Failed', `Found ${result.errors?.length || 0} errors and ${result.warnings?.length || 0} warnings`, 'warning');
        }
    } catch (error) {
        showNotification('Error', `Validation error: ${error.message}`, 'danger');
    } finally {
        btn.classList.remove('is-loading');
    }
}/**
 * Navigate to the validation section in overview
 */
function navigateToValidationSection() {
    // Switch to overview view if not already there
    const overviewView = document.querySelector('[data-view-content="overview"]');
    if (overviewView && !overviewView.classList.contains('is-active')) {
        const overviewTab = document.querySelector('[data-view="overview"]');
        if (overviewTab) overviewTab.click();
    }

    // Select validation section
    const validationNavItem = document.querySelector('[data-section="validation"]');
    if (validationNavItem) {
        validationNavItem.click();
    }
}

/**
 * Display validation results in the validation section
 */
function displayValidationResults(result, isSuccess) {
    const container = document.getElementById('validation-results');
    if (!container) return;

    const isValid = result.is_valid || false;
    const errors = result.errors || [];
    const warnings = result.warnings || [];

    let html = '';

    if (isValid) {
        html = `
            <div class="notification is-success is-light">
                <p class="mb-2"><strong><i class="fas fa-check-circle mr-2"></i>Validation Passed!</strong></p>
                <p class="is-size-7">Your agency specification is complete and ready for publishing.</p>
            </div>
        `;
    } else {
        // Show errors
        if (errors.length > 0) {
            html += `
                <div class="notification is-danger is-light mb-4">
                    <p class="mb-3"><strong><i class="fas fa-exclamation-circle mr-2"></i>Validation Errors (${errors.length})</strong></p>
                    <div class="content is-small">
                        <ul class="mb-0">
                            ${errors.map(e => `<li><strong>${e.field}:</strong> ${e.message}</li>`).join('')}
                        </ul>
                    </div>
                </div>
            `;
        }

        // Show warnings
        if (warnings.length > 0) {
            html += `
                <div class="notification is-warning is-light">
                    <p class="mb-3"><strong><i class="fas fa-exclamation-triangle mr-2"></i>Warnings (${warnings.length})</strong></p>
                    <div class="content is-small">
                        <ul class="mb-0">
                            ${warnings.map(w => `<li><strong>${w.field}:</strong> ${w.message}</li>`).join('')}
                        </ul>
                    </div>
                </div>
            `;
        }
    }

    container.innerHTML = html;
}

/**
 * Handle activate button click (Published → Active)
 */
async function handleActivateAgency() {
    if (!confirm('Activate this agency? This will spawn agents and start workflows.')) {
        return;
    }

    const btn = document.getElementById('activate-btn');
    btn.classList.add('is-loading');

    try {
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/activate`, {
            method: 'POST'
        });

        if (response.ok) {
            showNotification('Success', 'Agency activated successfully', 'success');
            setTimeout(() => window.location.reload(), 1500);
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Activation failed');
        }
    } catch (error) {
        alert(`Failed to activate: ${error.message}`);
    } finally {
        btn.classList.remove('is-loading');
    }
}/**
 * Update agency state via API
 */
async function updateAgencyState(newState) {
    const response = await fetch(`/api/v1/agencies/${currentAgencyId}`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ state: newState })
    });

    if (!response.ok) {
        throw new Error('Failed to update agency state');
    }
}

/**
 * Show notification (uses existing notification system if available)
 */
function showNotification(title, message, type = 'info') {
    // Use existing notification system or fallback
    if (window.showToast) {
        window.showToast(message, type);
    }
    // Silently ignore if no notification system available
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializePublishFeatures);
} else {
    initializePublishFeatures();
}
