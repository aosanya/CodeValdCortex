// CRUD Helper Functions
// Reusable functions for managing entities (goals, work items, roles)

// Uses global functions: getCurrentAgencyId, showNotification, specificationAPI

/**
 * Generic function to load entity list HTML
 * @param {string} entityType - Type of entity (goals, work-items, roles)
 * @param {string} tableBodyId - ID of the table body element
 * @param {number} colspan - Number of columns for loading/error messages
 */
window.loadEntityList = async function (entityType, tableBodyId, colspan = 3) {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const tableBody = document.getElementById(tableBodyId);
    if (!tableBody) {
        return;
    }

    // Show loading state
    tableBody.innerHTML = `<tr><td colspan="${colspan}" class="has-text-grey has-text-centered py-5"><p><i class="fas fa-spinner fa-spin"></i> Loading ${entityType}...</p></td></tr>`;

    try {
        // Use the specification API to get data and render HTML locally
        const specification = await window.specificationAPI.getSpecification();

        // Cache specification for use in rendering (especially for goal tags in work items)
        window._cachedSpecification = specification;

        let entities = [];

        // Handle workflows separately since they have their own API
        if (entityType === 'workflows') {
            const response = await fetch(`/api/v1/agencies/${agencyId}/workflows/html`);
            if (!response.ok) {
                throw new Error(`Failed to load workflows: ${response.status}`);
            }
            const html = await response.text();
            tableBody.innerHTML = html;
            return;
        }

        switch (entityType) {
            case 'goals':
                entities = specification.goals || [];
                break;
            case 'work-items':
                entities = specification.work_items || [];
                break;
            case 'roles':
                entities = specification.roles || [];
                break;
            default:
                throw new Error(`Unknown entity type: ${entityType}`);
        }

        // Generate HTML based on entity type
        const html = generateEntityListHTML(entityType, entities);
        tableBody.innerHTML = html;
    } catch (error) {
        tableBody.innerHTML = `<tr><td colspan="${colspan}" class="has-text-danger has-text-centered py-5"><p>Error loading ${entityType}</p></td></tr>`;
    }
}

/**
 * Generate HTML for entity list based on type and data
 * @param {string} entityType - Type of entity
 * @param {Array} entities - Array of entity objects
 */
function generateEntityListHTML(entityType, entities) {
    if (!entities || entities.length === 0) {
        const entityDisplay = entityType.replace('-', ' ');
        return `<tr><td colspan="4" class="has-text-grey has-text-centered py-5"><p>No ${entityDisplay} defined yet.</p></td></tr>`;
    }

    switch (entityType) {
        case 'goals':
            return entities.map(goal => {
                // Use _key as the primary unique identifier for goals (guaranteed unique by ArangoDB)
                const id = goal._key;

                return `
                <tr>
                    <td>
                        <label class="checkbox">
                            <input type="checkbox" value="${id}" onchange="window.updateGoalSelectionButtons && window.updateGoalSelectionButtons()">
                        </label>
                    </td>
                    <td>
                        <strong>${escapeHtml(goal.code || '')}</strong>
                    </td>
                    <td>${escapeHtml(goal.description || '')}</td>
                    <td>
                        <div class="buttons">
                            <button class="button is-small" onclick="window.showGoalEditor('edit', '${id}', '${escapeHtml(goal.code)}', '${escapeHtml(goal.description)}')">
                                <span class="icon"><i class="fas fa-edit"></i></span>
                            </button>
                            <button class="button is-small is-danger" onclick="window.deleteGoal('${id}')">
                                <span class="icon"><i class="fas fa-trash"></i></span>
                            </button>
                        </div>
                    </td>
                </tr>
                `;
            }).join('');

        case 'work-items':
            return entities.map(item => {
                // Use _key as the primary unique identifier for work items (guaranteed unique by UUID)
                const id = item._key;

                // Get goals data for displaying goal tags
                const goalKeys = item.goal_keys || [];
                let goalTagsHTML = '';

                if (goalKeys.length > 0) {
                    // We need to get the specification to map goal keys to goal codes
                    // Since we're in a map function, we'll use the cached specification
                    const goals = window._cachedSpecification?.goals || [];
                    const goalMap = {};
                    goals.forEach(g => {
                        goalMap[g._key] = g.code || g._key;
                    });

                    goalTagsHTML = '<hr class="m-0"/><div class="mt-2">' +
                        '<span class="has-text-grey is-size-6">Goals: </span>' +
                        '<div class="tags is-inline">' +
                        goalKeys.map(gk => {
                            const goalCode = goalMap[gk] || gk.substring(0, 8);
                            return `<span class="tag is-link is-light" title="Linked Goal: ${goalCode}">${goalCode}</span>`;
                        }).join('') +
                        '</div>' +
                        '</div>';
                }

                return `
                <tr>
                    <td>
                        <label class="checkbox">
                            <input type="checkbox" value="${id}" onchange="window.updateWorkItemSelectionButtons && window.updateWorkItemSelectionButtons()">
                        </label>
                    </td>
                    <td><strong>${escapeHtml(item.code || '')}</strong></td>
                    <td>
                        <div>
                            <strong>${escapeHtml(item.title || '')}</strong>
                            ${item.description ? `<br><small class="has-text-grey">${escapeHtml(item.description)}</small>` : ''}
                            ${goalTagsHTML}
                        </div>
                    </td>
                    <td>
                        <div class="buttons">
                            <button class="button is-small" onclick="window.showWorkItemEditor('edit', '${id}')">
                                <span class="icon"><i class="fas fa-edit"></i></span>
                            </button>
                            <button class="button is-small is-danger" onclick="window.deleteWorkItem('${id}')">
                                <span class="icon"><i class="fas fa-trash"></i></span>
                            </button>
                        </div>
                    </td>
                </tr>
                `;
            }).join('');

        case 'roles':
            return entities.map(role => {
                // Use _key as primary identifier for roles
                const id = role._key || role.key || role._id || role.code || role.id || role.name;

                return `
                <tr class="table-item">
                    <td>
                        <label class="checkbox">
                            <input type="checkbox" class="role-checkbox" value="${id}" data-role-key="${id}" onchange="window.updateRoleSelectionButtons && window.updateRoleSelectionButtons()">
                        </label>
                    </td>
                    <td><code>${escapeHtml(role.code || '')}</code></td>
                    <td>
                        <strong>${escapeHtml(role.name || '')}</strong><br>
                        <small class="has-text-grey">${escapeHtml(role.description || '')}</small>
                    </td>
                    <td><span class="tag is-small">${escapeHtml(role.autonomy_level || '')}</span></td>
                    <td>
                        <div class="buttons">
                            <button class="button is-small" onclick="window.showRoleEditor('edit', '${id}')">
                                <span class="icon"><i class="fas fa-edit"></i></span>
                            </button>
                            <button class="button is-small is-danger" onclick="window.deleteRole('${id}')">
                                <span class="icon"><i class="fas fa-trash"></i></span>
                            </button>
                        </div>
                    </td>
                </tr>
                `;
            }).join('');

        default:
            return '<tr><td colspan="3" class="has-text-danger">Unknown entity type</td></tr>';
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
 * Generic function to show entity editor
 * @param {string} mode - 'add' or 'edit'
 * @param {string} editorCardId - ID of editor card element
 * @param {string} listCardId - ID of list card element
 * @param {string} titleElementId - ID of editor title element
 * @param {string} addTitle - Title text for add mode
 * @param {string} editTitle - Title text for edit mode
 * @param {string} focusElementId - ID of element to focus after showing
 */
window.showEntityEditor = function (mode, editorCardId, listCardId, titleElementId, addTitle, editTitle, focusElementId) {
    const editorCard = document.getElementById(editorCardId);
    const listCard = document.getElementById(listCardId);
    const editorTitle = document.getElementById(titleElementId);

    if (!editorCard || !listCard) {
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
window.cancelEntityEdit = function (editorCardId, listCardId, fieldIds = []) {
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
window.deleteEntity = async function (entityType, entityKey, entityName, reloadCallback) {
    if (!confirm(`Are you sure you want to delete ${entityName}?`)) {
        return;
    }

    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    try {
        // Use specification API to delete entity
        switch (entityType) {
            case 'goals':
                await window.specificationAPI.deleteGoal(entityKey);
                break;
            case 'work-items':
                await window.specificationAPI.deleteWorkItem(entityKey);
                break;
            case 'roles':
                // For roles, we need to get current roles and filter out the deleted one
                const spec = await window.specificationAPI.getSpecification();
                const updatedRoles = (spec.roles || []).filter(r => {
                    const id = r._key || r.key || r._id || r.name;
                    return id !== entityKey;
                });
                await window.specificationAPI.updateRoles(updatedRoles);
                break;
            default:
                throw new Error(`Unknown entity type: ${entityType}`);
        }

        window.showNotification(`${entityName} deleted successfully!`, 'success');
        if (reloadCallback) {
            reloadCallback();
        }
    } catch (error) {
        window.showNotification(`Error deleting ${entityType.slice(0, -1)}`, 'error');
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
window.saveEntity = async function (entityType, mode, entityKey, data, saveBtnId, successCallback) {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    const saveBtn = saveBtnId ? document.getElementById(saveBtnId) : null;
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
    }

    const isAddMode = mode === 'add';

    try {
        // Use specification API to save entity
        if (isAddMode) {
            switch (entityType) {
                case 'goals':
                    await window.specificationAPI.addGoal(data);
                    break;
                case 'work-items':
                    await window.specificationAPI.addWorkItem(data);
                    break;
                case 'roles':
                    const spec = await window.specificationAPI.getSpecification();
                    const updatedRoles = [...(spec.roles || []), data];
                    await window.specificationAPI.updateRoles(updatedRoles);
                    break;
                default:
                    throw new Error(`Unknown entity type: ${entityType}`);
            }
        } else {
            switch (entityType) {
                case 'goals':
                    await window.specificationAPI.updateGoal(entityKey, data);
                    break;
                case 'work-items':
                    await window.specificationAPI.updateWorkItem(entityKey, data);
                    break;
                case 'roles':
                    const spec = await window.specificationAPI.getSpecification();
                    const roles = spec.roles || [];
                    const roleIndex = roles.findIndex(r => {
                        const id = r._key || r.key || r._id || r.name;
                        return id === entityKey;
                    });
                    if (roleIndex === -1) {
                        throw new Error(`Role with key ${entityKey} not found`);
                    }
                    roles[roleIndex] = { ...roles[roleIndex], ...data };
                    await window.specificationAPI.updateRoles(roles);
                    break;
                default:
                    throw new Error(`Unknown entity type: ${entityType}`);
            }
        }

        window.showNotification(`${entityType.slice(0, -1)} ${isAddMode ? 'added' : 'updated'} successfully!`, 'success');

        if (successCallback) {
            successCallback();
        }
    } catch (error) {
        window.showNotification(`Error ${isAddMode ? 'adding' : 'updating'} ${entityType.slice(0, -1)}`, 'error');
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
window.populateForm = function (fieldMap) {
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
window.clearForm = function (fieldIds, defaults = {}) {
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
