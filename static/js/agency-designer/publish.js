/**
 * Publish Dialog and Operations
 * 
 * Handles agency validation, publishing, and tag creation workflows
 * Interacts with /api/v1/agencies/:id/validate and /api/v1/agencies/:id/publish endpoints
 */

// State management
let currentAgencyId = null;
let validationResult = null;

/**
 * Initialize publish functionality
 */
function initializePublishFeatures() {
    // Get agency ID from URL or data attribute
    const urlParts = window.location.pathname.split('/');
    currentAgencyId = urlParts[urlParts.indexOf('agencies') + 1];

    // Set up form listeners for live preview
    setupPublishFormListeners();
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
    document.getElementById('publish-create-tag').checked = false;
    document.getElementById('publish-tag-name').value = '';
    document.getElementById('tag-name-field').style.display = 'none';
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

        if (response.ok && result.is_valid) {
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
            errorList.innerHTML = '<ul class="is-size-7">' +
                result.errors.map(err => `<li><strong>${err.field}:</strong> ${err.message}</li>`).join('') +
                '</ul>';

            publishForm.style.display = 'none';
        }
    } catch (error) {
        console.error('Validation check failed:', error);
        statusDiv.innerHTML = `
            <p class="has-text-danger">
                <span class="icon"><i class="fas fa-exclamation-circle"></i></span>
                Failed to check validation: ${error.message}
            </p>
        `;
    }
}

/**
 * Set up form field listeners for live preview
 */
function setupPublishFormListeners() {
    // Version input
    const versionInput = document.getElementById('publish-version');
    if (versionInput) {
        versionInput.addEventListener('input', updatePublishPreview);
    }

    // Description input
    const descInput = document.getElementById('publish-description');
    if (descInput) {
        descInput.addEventListener('input', updatePublishPreview);
    }

    // Auto-activate checkbox
    const activateCheck = document.getElementById('publish-auto-activate');
    if (activateCheck) {
        activateCheck.addEventListener('change', updatePublishPreview);
    }

    // Create tag checkbox
    const tagCheck = document.getElementById('publish-create-tag');
    if (tagCheck) {
        tagCheck.addEventListener('change', updatePublishPreview);
    }

    // Tag name input
    const tagNameInput = document.getElementById('publish-tag-name');
    if (tagNameInput) {
        tagNameInput.addEventListener('input', updatePublishPreview);
    }
}

/**
 * Update live preview of publication
 */
function updatePublishPreview() {
    const version = document.getElementById('publish-version').value || '-';
    const description = document.getElementById('publish-description').value || '-';
    const autoActivate = document.getElementById('publish-auto-activate').checked;
    const createTag = document.getElementById('publish-create-tag').checked;
    const tagName = document.getElementById('publish-tag-name').value || version;

    // Update preview fields
    document.getElementById('preview-version').textContent = version;
    document.getElementById('preview-description').textContent = description;

    // Update action list
    const tagAction = document.getElementById('preview-tag-action');
    const activateAction = document.getElementById('preview-activate-action');

    if (createTag) {
        tagAction.style.display = 'list-item';
        document.getElementById('preview-tag-name').textContent = tagName;
    } else {
        tagAction.style.display = 'none';
    }

    if (autoActivate) {
        activateAction.style.display = 'list-item';
    } else {
        activateAction.style.display = 'none';
    }
}

/**
 * Toggle tag name input visibility
 */
function toggleTagNameInput() {
    const createTag = document.getElementById('publish-create-tag').checked;
    const tagField = document.getElementById('tag-name-field');

    if (createTag) {
        tagField.style.display = 'block';
    } else {
        tagField.style.display = 'none';
    }

    updatePublishPreview();
}

/**
 * Handle publish form submission
 */
async function handlePublishSubmit() {
    const version = document.getElementById('publish-version').value.trim();
    const description = document.getElementById('publish-description').value.trim();
    const autoActivate = document.getElementById('publish-auto-activate').checked;
    const createTag = document.getElementById('publish-create-tag').checked;
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

        // Optionally create tag first
        if (createTag) {
            await createTagBeforePublish(tagName, version, description);
        }

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
        console.error('Publish failed:', error);
        alert(`Failed to publish agency: ${error.message}`);
    } finally {
        publishBtn.disabled = false;
        progress.style.display = 'none';
    }
}

/**
 * Create tag before publishing
 */
async function createTagBeforePublish(tagName, version, description) {
    const tagData = {
        name: tagName,
        type: 'snapshot',
        version: version,
        description: `Pre-publish snapshot: ${description}`,
        metadata: {
            created_before_publish: true
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
        throw new Error(`Failed to create tag: ${error.error}`);
    }
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

        if (response.ok && result.is_valid) {
            showNotification('Success', 'Agency validated successfully', 'success');

            // Transition state to validated
            await updateAgencyState('validated');

            // Reload to show new buttons
            setTimeout(() => window.location.reload(), 1000);
        } else {
            // Show validation errors
            const errors = result.errors.map(e => `${e.field}: ${e.message}`).join('\n');
            alert(`Validation failed:\n\n${errors}`);
        }
    } catch (error) {
        console.error('Validation failed:', error);
        alert(`Validation error: ${error.message}`);
    } finally {
        btn.classList.remove('is-loading');
    }
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
        console.error('Activation failed:', error);
        alert(`Failed to activate: ${error.message}`);
    } finally {
        btn.classList.remove('is-loading');
    }
}

/**
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
