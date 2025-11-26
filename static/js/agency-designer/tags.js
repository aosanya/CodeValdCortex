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
}

/**
 * Load and display all tags for the current agency
 */
async function loadTags() {
    const agencyID = getAgencyID();
    if (!agencyID) {
        return;
    }

    const tbody = document.getElementById('tags-table-body');
    if (!tbody) {
        return;
    }

    try {
        const url = `/api/v1/agencies/${agencyID}/tags`;

        const response = await fetch(url);

        if (!response.ok) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="has-text-danger has-text-centered py-5">
                        <p><i class="fas fa-exclamation-triangle"></i> Failed to load tags</p>
                    </td>
                </tr>
            `;
            return;
        }

        const data = await response.json();

        if (!data.tags || data.tags.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="has-text-grey has-text-centered py-5">
                        <p><i class="fas fa-info-circle"></i> No tags found. Create your first tag to get started.</p>
                    </td>
                </tr>
            `;
            return;
        }

        // Render tags
        tbody.innerHTML = data.tags.map(tag => renderTagRow(tag)).join('');

        // Load instance counts for each tag
        for (const tag of data.tags) {
            if (typeof loadInstancesForTag === 'function') {
                await loadInstancesForTag(tag.name);
            }
        }

    } catch (error) {
        console.error('Error loading tags:', error);
        tbody.innerHTML = `
            <tr>
                <td colspan="7" class="has-text-danger has-text-centered py-5">
                    <p><i class="fas fa-exclamation-triangle"></i> Error: ${error.message}</p>
                </td>
            </tr>
        `;
    }
}

/**
 * Render a single tag row
 */
function renderTagRow(tag) {
    const createdDate = new Date(tag.created_at).toLocaleDateString();
    const typeColors = {
        'release': 'is-success',
        'snapshot': 'is-info',
        'experimental': 'is-warning',
        'checkpoint': 'is-primary'
    };
    const typeColor = typeColors[tag.type] || 'is-light';

    return `
        <tr data-tag-name="${tag.name}">
            <td><i class="fas fa-tag has-text-${typeColor.replace('is-', '')}"></i></td>
            <td><strong>${escapeHtml(tag.name)}</strong></td>
            <td><span class="tag ${typeColor}">${escapeHtml(tag.type)}</span></td>
            <td>${escapeHtml(tag.version || '-')}</td>
            <td>${createdDate}</td>
            <td>
                <span class="tag instance-count-badge" data-tag-name="${tag.name}">0</span>
            </td>
            <td>
                <div class="dropdown is-hoverable is-right">
                    <div class="dropdown-trigger">
                        <button class="button is-small" aria-haspopup="true" aria-controls="dropdown-menu-${escapeHtml(tag.name)}">
                            <span class="icon is-small">
                                <i class="fas fa-ellipsis-v"></i>
                            </span>
                        </button>
                    </div>
                    <div class="dropdown-menu" id="dropdown-menu-${escapeHtml(tag.name)}" role="menu">
                        <div class="dropdown-content">
                            <a class="dropdown-item" onclick="openStartInstanceDialog('${escapeHtml(tag.name)}')">
                                <span class="icon"><i class="fas fa-play"></i></span>
                                <span>Start Instance</span>
                            </a>
                            <a class="dropdown-item" onclick="viewTagInstances('${escapeHtml(tag._id)}')">
                                <span class="icon"><i class="fas fa-server"></i></span>
                                <span>View Instances</span>
                            </a>
                            <a class="dropdown-item" onclick="viewTagDetails('${escapeHtml(tag.name)}')">
                                <span class="icon"><i class="fas fa-eye"></i></span>
                                <span>View Details</span>
                            </a>
                            <hr class="dropdown-divider">
                            <a class="dropdown-item has-text-danger" onclick="deleteTag('${escapeHtml(tag.name)}')">
                                <span class="icon"><i class="fas fa-trash"></i></span>
                                <span>Delete Tag</span>
                            </a>
                        </div>
                    </div>
                </div>
            </td>
        </tr>
    `;
}

/**
 * Filter tags by search input
 */
function filterTags() {
    const searchInput = document.getElementById('tags-search');
    if (!searchInput) return;

    const searchTerm = searchInput.value.toLowerCase();
    const rows = document.querySelectorAll('#tags-table-body tr');

    rows.forEach(row => {
        const tagName = row.getAttribute('data-tag-name');
        if (!tagName) return; // Skip empty state rows

        const visible = tagName.toLowerCase().includes(searchTerm);
        row.style.display = visible ? '' : 'none';
    });
}

/**
 * View tag details (placeholder)
 */
function viewTagDetails(tagName) {
    alert(`View details for tag: ${tagName}\n\nThis feature will show full tag information and snapshot details.`);
}

/**
 * View instances for a specific tag
 */
function viewTagInstances(tagName) {
    const agencyID = getAgencyID();
    if (!agencyID) {
        return;
    }

    // Navigate to instances page with tag_key filter in URL
    // Note: tagName is actually the tag ID from the renderTagRow function
    window.location.href = `/agencies/${agencyID}/instances?tag_key=${encodeURIComponent(tagName)}`;
}
/**
 * Delete a tag
 */
async function deleteTag(tagName) {
    if (!confirm(`Are you sure you want to delete tag "${tagName}"?\n\nThis action cannot be undone.`)) {
        return;
    }

    const agencyID = getAgencyID();
    if (!agencyID) return;

    try {
        const response = await fetch(`/api/v1/agencies/${agencyID}/tags/${tagName}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showNotification('Success', `Tag "${tagName}" deleted successfully`, 'success');
            loadTags(); // Reload tags list
        } else {
            const data = await response.json();
            showNotification('Error', data.details || data.error || 'Failed to delete tag', 'danger');
        }
    } catch (error) {
        console.error('Error deleting tag:', error);
        showNotification('Error', 'Network error: ' + error.message, 'danger');
    }
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Get agency ID from URL
 */
function getAgencyID() {
    const pathParts = window.location.pathname.split('/');
    const agencyIndex = pathParts.indexOf('agencies');
    if (agencyIndex !== -1 && pathParts.length > agencyIndex + 1) {
        return pathParts[agencyIndex + 1];
    }
    return null;
}

/**
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
