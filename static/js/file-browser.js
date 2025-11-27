// File Browser JavaScript
// Handles file explorer interactions, file editing, and API calls

/**
 * Get agency ID from meta tag
 */
function getAgencyID() {
    const meta = document.querySelector('meta[name="agency-id"]');
    return meta ? meta.content : null;
}

/**
 * Get instance ID from meta tag
 */
function getInstanceID() {
    const meta = document.querySelector('meta[name="instance-id"]');
    return meta ? meta.content : null;
}

/**
 * Navigate to a directory path
 */
function navigateToPath(instanceID, path) {
    if (!instanceID || !path) {
        return;
    }

    // Normalize path
    path = path.replace(/\/+/g, '/');

    // For now, just reload the page
    // TODO: Implement HTMX partial updates for better UX
    const agencyId = getAgencyID();
    window.location.href = `/agencies/${agencyId}/instances/${instanceID}/explorer?path=${encodeURIComponent(path)}`;
}/**
 * Open file for viewing
 */
function openFile(instanceID, filePath) {
    if (!instanceID || !filePath) {
        return;
    }

    const agencyID = getAgencyID();

    // Fetch file content
    fetch(`/api/files/${encodeURIComponent(filePath)}?instance_id=${encodeURIComponent(instanceID)}&agency_id=${encodeURIComponent(agencyID)}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to get file content');
            }
            return response.json();
        })
        .then(data => {
            // Show file content in modal (read-only)
            showFileViewer(data.path, data.content, data.mime_type);
        })
        .catch(error => {
            showNotification('Failed to open file', 'danger');
        });
}

/**
 * Edit file content
 */
function editFile(instanceID, filePath) {
    if (!instanceID || !filePath) {
        return;
    }

    const agencyID = getAgencyID();

    // Fetch file content
    fetch(`/api/files/${encodeURIComponent(filePath)}?instance_id=${encodeURIComponent(instanceID)}&agency_id=${encodeURIComponent(agencyID)}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to get file content');
            }
            return response.json();
        })
        .then(data => {
            // Show file editor modal
            showFileEditor(instanceID, data.path, data.content);
        })
        .catch(error => {
            showNotification('Failed to load file for editing', 'danger');
        });
}

/**
 * Save file changes
 */
function saveFile(instanceID, filePath) {
    const content = document.getElementById('file-content-editor').value;
    const author = 'user'; // TODO: Get from session
    const agencyID = getAgencyID();

    fetch(`/api/files/${encodeURIComponent(filePath)}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            instance_id: instanceID,
            agency_id: agencyID,
            path: filePath,
            content: content,
            author: author,
            message: `Update ${filePath}`
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to save file');
            }
            return response.json();
        })
        .then(data => {
            showNotification('File saved successfully', 'success');
            closeFileEditor();

            // Refresh file list
            const currentPath = filePath.substring(0, filePath.lastIndexOf('/')) || '/';
            navigateToPath(instanceID, currentPath);
        })
        .catch(error => {
            showNotification('Failed to save file', 'danger');
        });
}

/**
 * Delete file or folder
 */
function deleteFileOrFolder(instanceID, path, type) {
    const itemType = type === 'directory' ? 'folder' : 'file';

    if (!confirm(`Are you sure you want to delete this ${itemType}?`)) {
        return;
    }

    const author = 'user'; // TODO: Get from session
    const agencyID = getAgencyID();

    fetch(`/api/files/${encodeURIComponent(path)}?instance_id=${encodeURIComponent(instanceID)}&agency_id=${encodeURIComponent(agencyID)}&author=${author}`, {
        method: 'DELETE'
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed to delete ${itemType}`);
            }
            return response.json();
        })
        .then(data => {
            showNotification(`${itemType.charAt(0).toUpperCase() + itemType.slice(1)} deleted successfully`, 'success');

            // Refresh current directory
            const currentPath = path.substring(0, path.lastIndexOf('/')) || '/';
            navigateToPath(instanceID, currentPath);
        })
        .catch(error => {
            showNotification(`Failed to delete ${itemType}`, 'danger');
        });
}

/**
 * Show create file dialog
 */
function showCreateFileDialog() {
    const instanceID = document.querySelector('meta[name="instance-id"]').content;
    const currentPath = getCurrentPath();

    const fileName = prompt('Enter file name:');
    if (!fileName) return;

    const fullPath = currentPath === '/' ? `/${fileName}` : `${currentPath}/${fileName}`;

    // Create empty file
    createFile(instanceID, fullPath, '');
}

/**
 * Show create folder dialog
 */
function showCreateFolderDialog() {
    const instanceID = document.querySelector('meta[name="instance-id"]').content;
    const currentPath = getCurrentPath();

    const folderName = prompt('Enter folder name:');
    if (!folderName) return;

    const fullPath = currentPath === '/' ? `/${folderName}` : `${currentPath}/${folderName}`;

    createDirectory(instanceID, fullPath);
}

/**
 * Create new file
 */
function createFile(instanceID, path, content) {
    const author = 'user'; // TODO: Get from session
    const agencyID = getAgencyID();

    fetch('/api/files', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            instance_id: instanceID,
            agency_id: agencyID,
            path: path,
            content: content || '',
            author: author,
            message: `Create ${path}`
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to create file');
            }
            return response.json();
        })
        .then(data => {
            showNotification('File created successfully', 'success');

            // Refresh directory
            const currentPath = path.substring(0, path.lastIndexOf('/')) || '/';
            navigateToPath(instanceID, currentPath);
        })
        .catch(error => {
            showNotification('Failed to create file', 'danger');
        });
}

/**
 * Create new directory
 */
function createDirectory(instanceID, path) {
    const author = 'user'; // TODO: Get from session
    const agencyID = getAgencyID();

    fetch('/api/files/directory', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            instance_id: instanceID,
            agency_id: agencyID,
            path: path,
            author: author,
            message: `Create directory ${path}`
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to create directory');
            }
            return response.json();
        })
        .then(data => {
            showNotification('Directory created successfully', 'success');

            // Refresh current directory
            const currentPath = path.substring(0, path.lastIndexOf('/')) || '/';
            navigateToPath(instanceID, currentPath);
        })
        .catch(error => {
            showNotification('Failed to create directory', 'danger');
        });
}

/**
 * Show file editor modal
 */
function showFileEditor(instanceID, filePath, content) {
    // Create modal HTML
    const modalHTML = `
        <div class="modal is-active" id="file-editor-modal">
            <div class="modal-background" onclick="closeFileEditor()"></div>
            <div class="modal-card" style="width: 80%; max-width: 1200px;">
                <header class="modal-card-head">
                    <p class="modal-card-title">
                        <span class="icon"><i class="fas fa-edit"></i></span>
                        <span>${filePath.split('/').pop()}</span>
                    </p>
                    <button class="delete" aria-label="close" onclick="closeFileEditor()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="field">
                        <label class="label">File Path</label>
                        <p class="is-family-monospace is-size-7 has-text-grey">${filePath}</p>
                    </div>
                    <div class="field">
                        <label class="label">Content</label>
                        <div class="control">
                            <textarea 
                                class="textarea is-family-monospace" 
                                id="file-content-editor" 
                                rows="20" 
                                style="font-size: 0.875rem;">${content}</textarea>
                        </div>
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="saveFile('${instanceID}', '${filePath}')">
                        <span class="icon"><i class="fas fa-save"></i></span>
                        <span>Save Changes</span>
                    </button>
                    <button class="button" onclick="closeFileEditor()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

    // Append to body
    document.body.insertAdjacentHTML('beforeend', modalHTML);
}

/**
 * Show file viewer (read-only)
 */
function showFileViewer(filePath, content, mimeType) {
    // Similar to editor but read-only
    const modalHTML = `
        <div class="modal is-active" id="file-viewer-modal">
            <div class="modal-background" onclick="closeFileViewer()"></div>
            <div class="modal-card" style="width: 80%; max-width: 1200px;">
                <header class="modal-card-head">
                    <p class="modal-card-title">
                        <span class="icon"><i class="fas fa-eye"></i></span>
                        <span>${filePath.split('/').pop()}</span>
                    </p>
                    <button class="delete" aria-label="close" onclick="closeFileViewer()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="field">
                        <label class="label">File Path</label>
                        <p class="is-family-monospace is-size-7 has-text-grey">${filePath}</p>
                    </div>
                    <div class="field">
                        <label class="label">Content</label>
                        <div class="control">
                            <pre class="has-background-light p-4" style="max-height: 500px; overflow-y: auto;"><code>${escapeHtml(content)}</code></pre>
                        </div>
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button" onclick="closeFileViewer()">Close</button>
                </footer>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);
}

/**
 * Close file editor
 */
function closeFileEditor() {
    const modal = document.getElementById('file-editor-modal');
    if (modal) {
        modal.remove();
    }
}

/**
 * Close file viewer
 */
function closeFileViewer() {
    const modal = document.getElementById('file-viewer-modal');
    if (modal) {
        modal.remove();
    }
}

/**
 * Get current path from breadcrumb or URL
 */
function getCurrentPath() {
    // Try to get from URL parameter first
    const urlParams = new URLSearchParams(window.location.search);
    const pathFromUrl = urlParams.get('path');
    if (pathFromUrl) {
        return pathFromUrl;
    }

    // Fallback: Build path from breadcrumb items (including active)
    const items = Array.from(document.querySelectorAll('.breadcrumb li'));
    if (items.length === 0) {
        return '/';
    }

    // Collect all path segments (excluding "Root" text)
    const pathSegments = items
        .map(item => {
            const link = item.querySelector('a');
            return link ? link.textContent.trim() : '';
        })
        .filter(text => text && text !== 'Root');

    if (pathSegments.length === 0) {
        return '/';
    }

    return '/' + pathSegments.join('/');
}

/**
 * Escape HTML for safe display
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Show notification (reuse from instances.js if available)
 */
function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification is-${type} is-light`;
    notification.style.position = 'fixed';
    notification.style.top = '20px';
    notification.style.right = '20px';
    notification.style.zIndex = '9999';
    notification.style.minWidth = '300px';

    notification.innerHTML = `
        <button class="delete" onclick="this.parentElement.remove()"></button>
        ${message}
    `;

    document.body.appendChild(notification);

    // Auto-remove after 5 seconds
    setTimeout(() => {
        notification.remove();
    }, 5000);
}
