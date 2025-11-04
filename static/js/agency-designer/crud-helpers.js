// CRUD Helper Functions
// Reusable functions for managing entities (goals, work items, roles)

import { getCurrentAgencyId, showNotification } from './utils.js';

/**
 * Generic function to load entity list HTML
 * @param {string} entityType - Type of entity (goals, work-items, roles)
 * @param {string} tableBodyId - ID of the table body element
 * @param {number} colspan - Number of columns for loading/error messages
 */
export async function loadEntityList(entityType, tableBodyId, colspan = 3) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    const tableBody = document.getElementById(tableBodyId);
    if (!tableBody) {
        console.error(`Table body not found: ${tableBodyId}`);
        return;
    }

    // Show loading state
    tableBody.innerHTML = `<tr><td colspan="${colspan}" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading ${entityType}...</p></td></tr>`;

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/${entityType}/html`);
        if (!response.ok) {
            throw new Error(`Failed to load ${entityType}`);
        }
        const html = await response.text();
        tableBody.innerHTML = html;
    } catch (error) {
        console.error(`Error loading ${entityType}:`, error);
        tableBody.innerHTML = `<tr><td colspan="${colspan}" class="has-text-danger has-text-centered py-5"><p>Error loading ${entityType}</p></td></tr>`;
    }
}

/**
 * Generic function to show entity editor
 * @param {string} mode - 'add' or 'edit'
 * @param {string} editorCardId - ID of editor card element
 * @param {string} listCardId - ID of list card element
 * @param {string} titleElementId - ID of editor title element
 * @param {string} addTitle - Title text for add mode
 * @param {string} editTitle - Title text for edit mode
 * @param {string} focusElementId - ID of element to focus after showing
 */
export function showEntityEditor(mode, editorCardId, listCardId, titleElementId, addTitle, editTitle, focusElementId) {
    const editorCard = document.getElementById(editorCardId);
    const listCard = document.getElementById(listCardId);
    const editorTitle = document.getElementById(titleElementId);

    if (!editorCard || !listCard) {
        console.error('Editor or list card not found');
        return;
    }

    // Update editor title
    if (editorTitle) {
        editorTitle.innerHTML = mode === 'add' ? addTitle : editTitle;
    }

    // Show editor, hide list
    editorCard.classList.remove('is-hidden');
    listCard.classList.add('is-hidden');

    // Focus on specified element
    if (focusElementId) {
        setTimeout(() => {
            document.getElementById(focusElementId)?.focus();
        }, 100);
    }
}

/**
 * Generic function to cancel entity edit
 * @param {string} editorCardId - ID of editor card element
 * @param {string} listCardId - ID of list card element
 * @param {string[]} fieldIds - Array of field IDs to clear
 */
export function cancelEntityEdit(editorCardId, listCardId, fieldIds = []) {
    const editorCard = document.getElementById(editorCardId);
    const listCard = document.getElementById(listCardId);

    if (editorCard) editorCard.classList.add('is-hidden');
    if (listCard) listCard.classList.remove('is-hidden');

    // Clear form fields
    fieldIds.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            element.value = '';
        }
    });
}

/**
 * Generic function to delete entity
 * @param {string} entityType - Type of entity (goals, work-items, roles)
 * @param {string} entityKey - Key/ID of entity to delete
 * @param {string} entityName - Display name for confirmation
 * @param {Function} reloadCallback - Function to call after successful deletion
 */
export async function deleteEntity(entityType, entityKey, entityName, reloadCallback) {
    if (!confirm(`Are you sure you want to delete ${entityName}?`)) {
        return;
    }

    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    try {
        const response = await fetch(`/api/v1/agencies/${agencyId}/${entityType}/${entityKey}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            throw new Error(`Failed to delete ${entityType.slice(0, -1)}`);
        }

        await response.json();
        showNotification(`${entityName} deleted successfully!`, 'success');
        if (reloadCallback) {
            reloadCallback();
        }
    } catch (error) {
        console.error(`Error deleting ${entityType}:`, error);
        showNotification(`Error deleting ${entityType.slice(0, -1)}`, 'error');
    }
}

/**
 * Generic function to save entity
 * @param {string} entityType - Type of entity (goals, work-items, roles)
 * @param {string} mode - 'add' or 'edit'
 * @param {string} entityKey - Key/ID of entity (for edit mode)
 * @param {Object} data - Entity data to save
 * @param {string} saveBtnId - ID of save button (to show loading state)
 * @param {Function} successCallback - Function to call after successful save
 */
export async function saveEntity(entityType, mode, entityKey, data, saveBtnId, successCallback) {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    const saveBtn = saveBtnId ? document.getElementById(saveBtnId) : null;
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
    }

    const isAddMode = mode === 'add';
    const url = isAddMode
        ? `/api/v1/agencies/${agencyId}/${entityType}`
        : `/api/v1/agencies/${agencyId}/${entityType}/${entityKey}`;
    const method = isAddMode ? 'POST' : 'PUT';

    try {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            throw new Error(`Failed to ${isAddMode ? 'create' : 'update'} ${entityType.slice(0, -1)}`);
        }

        await response.json();
        showNotification(`${entityType.slice(0, -1)} ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');

        if (successCallback) {
            successCallback();
        }
    } catch (error) {
        console.error(`Error ${isAddMode ? 'creating' : 'updating'} ${entityType}:`, error);
        showNotification(`Error ${isAddMode ? 'adding' : 'updating'} ${entityType.slice(0, -1)}`, 'error');
    } finally {
        if (saveBtn) {
            saveBtn.classList.remove('is-loading');
        }
    }
}

/**
 * Helper to populate form fields from data object
 * @param {Object} fieldMap - Map of field IDs to data values
 */
export function populateForm(fieldMap) {
    for (const [id, value] of Object.entries(fieldMap)) {
        const element = document.getElementById(id);
        if (element) {
            element.value = value;
        }
    }
}

/**
 * Helper to clear form fields
 * @param {string[]} fieldIds - Array of field IDs to clear
 * @param {Object} defaults - Optional default values for specific fields
 */
export function clearForm(fieldIds, defaults = {}) {
    fieldIds.forEach(id => {
        const element = document.getElementById(id);
        if (element) {
            if (defaults[id] !== undefined) {
                element.value = defaults[id];
            } else if (element.type === 'color') {
                element.value = '#3298dc';
            } else if (element.type === 'number') {
                element.value = 0;
            } else {
                element.value = '';
            }
        }
    });
}
