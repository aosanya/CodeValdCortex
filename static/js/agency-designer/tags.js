/**
 * Tag Dialog and Management
 * 
 * Handles tag creation, listing, comparison, and restoration
 * 
 * Depends on: shared.js (for currentAgencyId)
 */

// State (currentAgencyId is in shared.js)
let metadataFieldCount = 0;

/**
 * Initialize tag features
 */
function initializeTagFeatures() {
    // currentAgencyId is already initialized in shared.js
    setupTagFormListeners();
}

/**
 * Open tag creation dialog
 */
function openTagDialog() {
    const dialog = document.getElementById('tag-dialog');
    dialog.classList.add('is-active');
    resetTagForm();
}

/**
 * Close tag dialog
 */
function closeTagDialog() {
    const dialog = document.getElementById('tag-dialog');
    dialog.classList.remove('is-active');
}

/**
 * Reset tag form
 */
function resetTagForm() {
    document.getElementById('tag-name').value = '';
    document.getElementById('tag-type').value = 'release';
    document.getElementById('tag-version').value = '';
    document.getElementById('tag-description').value = '';
    document.getElementById('tag-show-metadata').checked = false;
    document.getElementById('tag-metadata-fields').style.display = 'none';
    document.getElementById('metadata-pairs').innerHTML = '';
    metadataFieldCount = 0;
    updateTagTypeHelp();
    updateTagPreview();
}

/**
 * Set up form listeners
 */
function setupTagFormListeners() {
    const nameInput = document.getElementById('tag-name');
    if (nameInput) {
        nameInput.addEventListener('input', updateTagPreview);
    }

    const typeSelect = document.getElementById('tag-type');
    if (typeSelect) {
        typeSelect.addEventListener('change', updateTagPreview);
    }

    const versionInput = document.getElementById('tag-version');
    if (versionInput) {
        versionInput.addEventListener('input', updateTagPreview);
    }

    const descInput = document.getElementById('tag-description');
    if (descInput) {
        descInput.addEventListener('input', updateTagPreview);
    }
}

/**
 * Update tag type help text
 */
function updateTagTypeHelp() {
    const tagType = document.getElementById('tag-type').value;
    const helpText = document.getElementById('tag-type-help');

    const helpTexts = {
        'release': 'Production-ready versions intended for deployment',
        'snapshot': 'Point-in-time backups for safety or comparison',
        'experimental': 'Testing experimental changes before committing',
        'checkpoint': 'Stable state markers before major refactoring'
    };

    helpText.textContent = helpTexts[tagType] || '';
}

/**
 * Update tag preview
 */
function updateTagPreview() {
    const name = document.getElementById('tag-name').value || '-';
    const type = document.getElementById('tag-type').value;
    const version = document.getElementById('tag-version').value || '-';
    const description = document.getElementById('tag-description').value || '-';

    document.getElementById('preview-tag-name-text').textContent = name;
    document.getElementById('preview-tag-type').textContent =
        type.charAt(0).toUpperCase() + type.slice(1);
    document.getElementById('preview-tag-version').textContent = version;
    document.getElementById('preview-tag-description').textContent = description;
}

/**
 * Toggle metadata fields visibility
 */
function toggleMetadataFields() {
    const show = document.getElementById('tag-show-metadata').checked;
    const fields = document.getElementById('tag-metadata-fields');

    if (show) {
        fields.style.display = 'block';
        if (metadataFieldCount === 0) {
            addMetadataField(); // Add first field automatically
        }
    } else {
        fields.style.display = 'none';
    }
}

/**
 * Add metadata key-value field
 */
function addMetadataField() {
    metadataFieldCount++;
    const container = document.getElementById('metadata-pairs');

    const fieldHtml = `
        <div class="field is-horizontal mb-3" id="metadata-field-${metadataFieldCount}">
            <div class="field-body">
                <div class="field">
                    <div class="control">
                        <input 
                            class="input is-small" 
                            type="text" 
                            placeholder="Key"
                            id="metadata-key-${metadataFieldCount}">
                    </div>
                </div>
                <div class="field">
                    <div class="control">
                        <input 
                            class="input is-small" 
                            type="text" 
                            placeholder="Value"
                            id="metadata-value-${metadataFieldCount}">
                    </div>
                </div>
                <div class="field is-narrow">
                    <div class="control">
                        <button 
                            class="button is-small is-danger is-outlined"
                            type="button"
                            onclick="removeMetadataField(${metadataFieldCount})">
                            <span class="icon"><i class="fas fa-times"></i></span>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;

    container.insertAdjacentHTML('beforeend', fieldHtml);
}

/**
 * Remove metadata field
 */
function removeMetadataField(id) {
    const field = document.getElementById(`metadata-field-${id}`);
    if (field) {
        field.remove();
    }
}

/**
 * Collect metadata from form
 */
function collectMetadata() {
    const metadata = {};
    const container = document.getElementById('metadata-pairs');
    const fields = container.querySelectorAll('[id^="metadata-field-"]');

    fields.forEach(field => {
        const id = field.id.split('-')[2];
        const key = document.getElementById(`metadata-key-${id}`).value.trim();
        const value = document.getElementById(`metadata-value-${id}`).value.trim();

        if (key && value) {
            metadata[key] = value;
        }
    });

    return metadata;
}

/**
 * Handle tag form submission
 */
async function handleTagSubmit() {
    const name = document.getElementById('tag-name').value.trim();
    const type = document.getElementById('tag-type').value;
    const version = document.getElementById('tag-version').value.trim();
    const description = document.getElementById('tag-description').value.trim();

    // Validate required fields
    if (!name || !description) {
        alert('Tag name and description are required');
        return;
    }

    // Validate name format
    const namePattern = /^[a-zA-Z0-9._-]+$/;
    if (!namePattern.test(name)) {
        alert('Tag name can only contain letters, numbers, dots, dashes, and underscores');
        return;
    }

    // Validate version format if provided
    if (version) {
        const versionPattern = /^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$/;
        if (!versionPattern.test(version)) {
            alert('Version must be in semantic versioning format (e.g., 1.0.0)');
            return;
        }
    }

    const tagBtn = document.getElementById('confirm-tag-btn');
    const progress = document.getElementById('tag-progress');
    tagBtn.disabled = true;
    progress.style.display = 'block';

    try {
        // Collect metadata
        const metadata = document.getElementById('tag-show-metadata').checked
            ? collectMetadata()
            : {};

        // Build tag data
        const tagData = {
            name: name,
            type: type,
            version: version || null,
            description: description,
            metadata: metadata
        };

        // Create tag
        const response = await fetch(`/api/v1/agencies/${currentAgencyId}/tags`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(tagData)
        });

        const result = await response.json();

        if (response.ok) {
            closeTagDialog();
            showNotification('Success', `Tag "${name}" created successfully`, 'success');

            // Refresh tag list if on tag management page
            if (typeof refreshTagList === 'function') {
                refreshTagList();
            }
        } else {
            throw new Error(result.error || 'Failed to create tag');
        }
    } catch (error) {
        alert(`Failed to create tag: ${error.message}`);
    } finally {
        tagBtn.disabled = false;
        progress.style.display = 'none';
    }
}/**
 * Pause agency
 */
async function handlePauseAgency() {
    if (!confirm('Pause this agency? Agents will be temporarily suspended.')) {
        return;
    }

    await performLifecycleAction('pause', 'Agency paused successfully');
}

/**
 * Resume agency
 */
async function handleResumeAgency() {
    if (!confirm('Resume this agency? Agents will be reactivated.')) {
        return;
    }

    await performLifecycleAction('resume', 'Agency resumed successfully');
}

/**
 * Drain agency
 */
async function handleDrainAgency() {
    if (!confirm('Drain this agency? No new work will be accepted, but existing work will complete.')) {
        return;
    }

    await performLifecycleAction('drain', 'Agency draining... Monitor status bar for completion.');
}

/**
 * Stop agency
 */
async function handleStopAgency() {
    if (!confirm('Stop this agency? All agents will be shut down gracefully.')) {
        return;
    }

    await performLifecycleAction('stop', 'Agency stopped successfully');
}

/**
 * Force stop agency (during draining)
 */
async function handleForceStopAgency() {
    if (!confirm('FORCE STOP this agency? This will immediately terminate all agents, potentially interrupting work!')) {
        return;
    }

    await performLifecycleAction('stop', 'Agency force-stopped', true);
}

/**
 * Perform lifecycle action
 */
async function performLifecycleAction(action, successMessage, force = false) {
    const btnId = force ? 'force-stop-btn' : `${action}-btn`;
    const btn = document.getElementById(btnId);

    if (btn) {
        btn.classList.add('is-loading');
    }

    try {
        const url = `/api/v1/agencies/${currentAgencyId}/lifecycle/${action}`;
        const body = force ? JSON.stringify({ force: true }) : null;

        const response = await fetch(url, {
            method: 'POST',
            headers: body ? { 'Content-Type': 'application/json' } : {},
            body: body
        });

        if (response.ok) {
            showNotification('Success', successMessage, 'success');
            setTimeout(() => window.location.reload(), 1500);
        } else {
            const error = await response.json();
            throw new Error(error.error || `${action} failed`);
        }
    } catch (error) {
        alert(`Failed to ${action}: ${error.message}`);
    } finally {
        if (btn) {
            btn.classList.remove('is-loading');
        }
    }
}/**
 * Show notification
 */
function showNotification(title, message, type = 'info') {
    if (window.showToast) {
        window.showToast(message, type);
    }
    // Silently ignore if no notification system available
}

// Initialize
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeTagFeatures);
} else {
    initializeTagFeatures();
}
