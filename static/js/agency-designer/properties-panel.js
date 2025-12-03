/**
 * Properties Panel Manager
 * Generic, reusable properties panel with configurable save handlers
 */

window.PropertiesPanel = {
    // Store current configuration
    _currentConfig: null,

    /**
     * Show properties in the panel with custom configuration
     * @param {Object} config - Configuration object
     * @param {string} config.title - Panel title
     * @param {string} config.icon - FontAwesome icon class
     * @param {string} config.iconColor - Bulma color class for icon
     * @param {Array} config.fields - Array of field definitions
     * @param {Array} config.buttons - Array of button definitions
     * @param {Function} config.onUpdate - Called when field changes (field, value)
     * @param {Function} config.onSave - Called when save button clicked
     * @param {Function} config.onDelete - Called when delete button clicked
     * @param {Object} config.data - The data object being edited
     */
    showProperties: function (config) {
        const panel = document.getElementById('properties-panel-content');
        if (!panel) {
            console.error('Properties panel not found');
            return;
        }

        // Store configuration for callbacks
        this._currentConfig = config;

        // Auto-switch to properties tab if enabled
        if (config.autoSwitchTab !== false) {
            const chatPanel = document.querySelector('.chat-panel');
            if (chatPanel && chatPanel._x_dataStack && chatPanel._x_dataStack[0]) {
                chatPanel._x_dataStack[0].activeTab = 'properties';
            }
        }

        // Build fields HTML
        const fieldsHtml = config.fields.map(field => this._renderField(field, config.data)).join('');

        // Build buttons HTML
        const buttonsHtml = config.buttons ? config.buttons.map(btn => this._renderButton(btn)).join('') : '';

        const html = `
            <div class="box">
                <h4 class="title is-6 mb-4">
                    <span class="icon-text">
                        <span class="icon has-text-${config.iconColor || 'primary'}">
                            <i class="fas fa-${config.icon}"></i>
                        </span>
                        <span>${config.title}</span>
                    </span>
                </h4>

                ${fieldsHtml}

                ${buttonsHtml ? `<hr><div class="buttons">${buttonsHtml}</div>` : ''}
            </div>
        `;

        panel.innerHTML = html;
    },

    /**
     * Render a single field based on its type
     */
    _renderField: function (field, data) {
        const value = data[field.key] || field.default || '';
        const fieldId = `prop-field-${field.key}`;

        switch (field.type) {
            case 'text':
            case 'static':
                return `
                    <div class="field">
                        <label class="label">${field.label}</label>
                        <div class="control">
                            <input type="text" 
                                   id="${fieldId}"
                                   class="input ${field.type === 'static' ? 'is-static' : ''}" 
                                   value="${this._escapeHtml(value)}" 
                                   ${field.type === 'static' ? 'readonly' : ''}
                                   ${field.disabled ? 'disabled' : ''}
                                   ${field.type !== 'static' && !field.disabled ? `onchange="window.PropertiesPanel._handleFieldChange('${field.key}', this.value)"` : ''}
                                   placeholder="${field.placeholder || ''}">
                        </div>
                        ${field.help ? `<p class="help">${field.help}</p>` : ''}
                    </div>
                `;

            case 'textarea':
                return `
                    <div class="field">
                        <label class="label">${field.label}</label>
                        <div class="control">
                            <textarea 
                                id="${fieldId}"
                                class="textarea" 
                                rows="${field.rows || 4}"
                                ${field.disabled ? 'disabled' : ''}
                                ${!field.disabled ? `onchange="window.PropertiesPanel._handleFieldChange('${field.key}', this.value)"` : ''}
                                placeholder="${field.placeholder || ''}">${this._escapeHtml(value)}</textarea>
                        </div>
                        ${field.help ? `<p class="help">${field.help}</p>` : ''}
                    </div>
                `;

            case 'select':
                const options = field.options.map(opt => {
                    const optValue = typeof opt === 'string' ? opt : opt.value;
                    const optLabel = typeof opt === 'string' ? opt : opt.label;
                    return `<option value="${optValue}" ${value === optValue ? 'selected' : ''}>${optLabel}</option>`;
                }).join('');

                return `
                    <div class="field">
                        <label class="label">${field.label}</label>
                        <div class="control">
                            <div class="select is-fullwidth">
                                <select id="${fieldId}" 
                                        ${field.disabled ? 'disabled' : ''}
                                        ${!field.disabled ? `onchange="window.PropertiesPanel._handleFieldChange('${field.key}', this.value)"` : ''}>
                                    ${options}
                                </select>
                            </div>
                        </div>
                        ${field.help ? `<p class="help">${field.help}</p>` : ''}
                    </div>
                `;

            case 'tags':
                const tags = Array.isArray(value) ? value : [];
                return `
                    <div class="field">
                        <label class="label">${field.label}</label>
                        <div class="control">
                            <div class="tags">
                                ${tags.length > 0
                        ? tags.map(tag => `
                                        <span class="tag is-info is-light">
                                            ${this._escapeHtml(tag)}
                                            ${!field.disabled ? `<button class="delete is-small" 
                                                    onclick="window.PropertiesPanel._handleRemoveTag('${field.key}', '${this._escapeHtml(tag)}')"></button>` : ''}
                                        </span>
                                      `).join('')
                        : '<span class="has-text-grey-light is-size-7">No tags</span>'
                    }
                            </div>
                        </div>
                        ${field.help ? `<p class="help">${field.help}</p>` : ''}
                    </div>
                `;

            case 'badge':
                return `
                    <div class="field">
                        <label class="label">${field.label}</label>
                        <div class="control">
                            <span class="tag is-${field.color || 'primary'} is-light">
                                ${field.icon ? `<span class="icon"><i class="fas fa-${field.icon}"></i></span>` : ''}
                                <span>${field.format ? field.format(value) : value}</span>
                            </span>
                        </div>
                        ${field.help ? `<p class="help">${field.help}</p>` : ''}
                    </div>
                `;

            case 'custom':
                return field.render ? field.render(value, data) : '';

            default:
                return '';
        }
    },

    /**
     * Render a button
     */
    _renderButton: function (btn) {
        return `
            <button class="button is-small ${btn.class || 'is-light'}" 
                    onclick="window.PropertiesPanel._handleButtonClick('${btn.action}')">
                ${btn.icon ? `<span class="icon"><i class="fas fa-${btn.icon}"></i></span>` : ''}
                <span>${btn.label}</span>
            </button>
        `;
    },

    /**
     * Handle field change events
     */
    _handleFieldChange: function (field, value) {
        if (this._currentConfig && this._currentConfig.onUpdate) {
            this._currentConfig.onUpdate(field, value);
        }
    },

    /**
     * Handle tag removal
     */
    _handleRemoveTag: function (field, tag) {
        if (this._currentConfig && this._currentConfig.onRemoveTag) {
            this._currentConfig.onRemoveTag(field, tag);
        }
    },

    /**
     * Handle button clicks
     */
    _handleButtonClick: function (action) {
        if (!this._currentConfig) return;

        switch (action) {
            case 'save':
                if (this._currentConfig.onSave) {
                    this._currentConfig.onSave();
                }
                break;
            case 'delete':
                if (this._currentConfig.onDelete) {
                    if (confirm('Are you sure you want to delete this item?')) {
                        this._currentConfig.onDelete();
                    }
                }
                break;
            case 'close':
                this.clear();
                break;
            case 'chat':
                this.switchToChat();
                break;
            default:
                // Custom action
                if (this._currentConfig.onAction) {
                    this._currentConfig.onAction(action);
                }
        }
    },    /**
     * Escape HTML to prevent XSS
     */
    _escapeHtml: function (text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    /**
     * Show work item properties in the properties panel (legacy wrapper)
     */
    showWorkItemProperties: function (workItem) {
        this.showProperties({
            title: 'Work Item Properties',
            icon: 'tasks',
            iconColor: 'primary',
            data: workItem,
            fields: [
                {
                    key: 'code',
                    label: 'Code',
                    type: 'static'
                },
                {
                    key: 'title',
                    label: 'Title',
                    type: 'text',
                    placeholder: 'Work item title'
                },
                {
                    key: 'description',
                    label: 'Description',
                    type: 'textarea',
                    rows: 4,
                    placeholder: 'Work item description'
                },
                {
                    key: 'status',
                    label: 'Status',
                    type: 'select',
                    options: [
                        { value: 'not-started', label: 'Not Started' },
                        { value: 'in-progress', label: 'In Progress' },
                        { value: 'completed', label: 'Completed' },
                        { value: 'blocked', label: 'Blocked' }
                    ]
                },
                {
                    key: 'priority',
                    label: 'Priority',
                    type: 'select',
                    options: [
                        { value: 'low', label: 'Low' },
                        { value: 'medium', label: 'Medium' },
                        { value: 'high', label: 'High' },
                        { value: 'critical', label: 'Critical' }
                    ]
                },
                {
                    key: 'tags',
                    label: 'Tags',
                    type: 'tags',
                    help: 'Manage tags in the work item editor'
                },
                {
                    key: 'deliverables_structured',
                    label: 'Deliverables Count',
                    type: 'badge',
                    icon: 'folder-tree',
                    color: 'primary',
                    format: (value) => `${this.countDeliverables(value || [])} items`,
                    help: 'View and edit in the deliverables tree'
                }
            ],
            buttons: [
                {
                    action: 'chat',
                    label: 'Back to Chat',
                    icon: 'comments'
                }
            ],
            onUpdate: (field, value) => {
                // Get the current work item from the editor
                const titleInput = document.getElementById('work-item-title');
                const descInput = document.getElementById('work-item-description');

                // Update the form fields
                if (field === 'title' && titleInput) {
                    titleInput.value = value;
                } else if (field === 'description' && descInput) {
                    descInput.value = value;
                }

                // Trigger change event to update any listeners
                const event = new Event('change', { bubbles: true });
                if (field === 'title' && titleInput) {
                    titleInput.dispatchEvent(event);
                } else if (field === 'description' && descInput) {
                    descInput.dispatchEvent(event);
                }
            },
            onRemoveTag: (field, tag) => {
                if (window.WorkItems && window.WorkItems.removeTag) {
                    window.WorkItems.removeTag(tag);
                }
            }
        });
    },

    /**
     * Show deliverable node properties in the properties panel
     */
    showDeliverableNodeProperties: function (node) {
        const fields = [
            {
                key: 'name',
                label: 'Name',
                type: 'text',
                placeholder: 'Node name'
            }
        ];

        // Add file extension field for files
        if (node.type === 'file') {
            fields.push({
                key: 'file_extension',
                label: 'File Extension',
                type: 'text',
                disabled: true,
                default: '.md',
                help: 'Currently only .md (Markdown) files are supported'
            });
        }

        fields.push(
            {
                key: 'description',
                label: 'Description',
                type: 'text',
                placeholder: 'Brief description'
            },
            {
                key: 'prompt_instructions',
                label: 'AI Prompt Instructions',
                type: 'textarea',
                rows: 6,
                placeholder: 'Describe what content should be included in this file/folder...',
                help: 'These instructions guide AI agents on what to include when generating this deliverable.'
            },
            {
                key: 'path',
                label: 'Path',
                type: 'static',
                default: node.name,
                help: 'Full path in the deliverables tree'
            }
        );

        this.showProperties({
            title: `Deliverable: ${node.name}${node.type === 'file' ? (node.file_extension || '') : ''}`,
            icon: node.type === 'folder' ? 'folder' : 'file-lines',
            iconColor: node.type === 'folder' ? 'info' : 'primary',
            data: node,
            fields: fields,
            buttons: [
                {
                    action: 'save',
                    label: 'Save',
                    icon: 'save',
                    class: 'is-primary'
                },
                {
                    action: 'delete',
                    label: 'Delete Node',
                    icon: 'trash',
                    class: 'is-danger is-light'
                },
                {
                    action: 'close',
                    label: 'Close',
                    icon: 'times'
                }
            ],
            onUpdate: (field, value) => {
                // Get the Alpine.js tree component
                const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
                if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
                    const alpineData = treeContainer._x_dataStack[0];
                    if (alpineData.updateNodeField) {
                        alpineData.updateNodeField(node.id, field, value);
                    }
                }
            },
            onSave: async () => {

                // Show saving status
                window.showNotification('Saving...', 'info');

                // Get the tree component and call its save callback
                const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
                if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
                    const alpineData = treeContainer._x_dataStack[0];

                    // Call the save function passed to deliverable tree
                    if (alpineData.onSave && typeof alpineData.onSave === 'function') {
                        try {
                            await alpineData.onSave();
                            window.showNotification('Saved successfully', 'success');
                        } catch (error) {
                            console.error('Error saving:', error);
                            window.showNotification('Save failed', 'error');
                        }
                    } else {
                        console.error('No save handler configured. AlpineData:', alpineData);
                        window.showNotification('No save handler configured', 'warning');
                    }
                } else {
                    console.error('Tree component not found. Container:', treeContainer);
                    window.showNotification('Tree component not found', 'error');
                }
            },
            onDelete: () => {
                const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
                if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
                    const alpineData = treeContainer._x_dataStack[0];
                    if (alpineData.deleteNode) {
                        alpineData.deleteNode(node.id);
                        this.clear(); // Clear properties panel after delete
                    }
                }
            }
        });
    },

    // Legacy methods for backward compatibility

    /**
     * Update a deliverable node field (legacy)
     */
    updateDeliverableNodeField: function (nodeId, field, value) {
        // Get the Alpine.js tree component
        const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
        if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
            const alpineData = treeContainer._x_dataStack[0];
            if (alpineData.updateNodeField) {
                alpineData.updateNodeField(nodeId, field, value);
            }
        }
    },

    /**
     * Select a deliverable node (switch both tree and properties panel)
     */
    selectDeliverableNode: function (nodeId) {
        const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
        if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
            const alpineData = treeContainer._x_dataStack[0];
            alpineData.selectedNodeId = nodeId;

            // Show properties for this node
            const node = alpineData.findNodeById(nodeId);
            if (node) {
                this.showDeliverableNodeProperties(node);
            }
        }
    },

    /**
     * Delete a deliverable node (legacy)
     */
    deleteDeliverableNode: function (nodeId) {
        const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
        if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
            const alpineData = treeContainer._x_dataStack[0];
            if (alpineData.deleteNode) {
                alpineData.deleteNode(nodeId);
                this.clear(); // Clear properties panel after delete
            }
        }
    },

    /**
     * Save deliverable node (legacy)
     */
    saveDeliverableNode: async function (nodeId) {
        // Show saving status
        window.showNotification('Saving...', 'info');

        // Get the tree component and call its save callback
        const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
        if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack[0]) {
            const alpineData = treeContainer._x_dataStack[0];

            if (alpineData.onSave && typeof alpineData.onSave === 'function') {
                try {
                    await alpineData.onSave();
                    window.showNotification('Saved successfully', 'success');
                } catch (error) {
                    console.error('Error saving:', error);
                    window.showNotification('Save failed', 'error');
                }
            } else {
                window.showNotification('No save handler configured', 'warning');
            }
        }
    },

    /**
     * Count total deliverables (files and folders)
     */
    countDeliverables: function (deliverables) {
        let count = deliverables.length;
        deliverables.forEach(item => {
            if (item.children && item.children.length > 0) {
                count += this.countDeliverables(item.children);
            }
        });
        return count;
    },

    /**
     * Update a work item field
     */
    updateWorkItemField: function (field, value) {
        // Get the current work item from the editor
        const titleInput = document.getElementById('work-item-title');
        const descInput = document.getElementById('work-item-description');

        // Update the form fields
        if (field === 'title' && titleInput) {
            titleInput.value = value;
        } else if (field === 'description' && descInput) {
            descInput.value = value;
        }

        // Trigger change event to update any listeners
        const event = new Event('change', { bubbles: true });
        if (field === 'title' && titleInput) {
            titleInput.dispatchEvent(event);
        } else if (field === 'description' && descInput) {
            descInput.dispatchEvent(event);
        }
    },

    /**
     * Remove a tag from work item
     */
    removeTag: function (tag) {
        // This would integrate with the work items module
        if (window.WorkItems && window.WorkItems.removeTag) {
            window.WorkItems.removeTag(tag);
        }
    },

    /**
     * Switch back to chat tab
     */
    switchToChat: function () {
        // Trigger Alpine.js to switch tabs
        const chatPanel = document.querySelector('.chat-panel');
        if (chatPanel && chatPanel._x_dataStack && chatPanel._x_dataStack[0]) {
            chatPanel._x_dataStack[0].activeTab = 'chat';
        }
    },

    /**
     * Clear properties panel
     */
    clear: function () {
        const panel = document.getElementById('properties-panel-content');
        if (panel) {
            panel.innerHTML = `
                <div class="notification is-light has-text-centered">
                    <p class="has-text-grey">
                        <span class="icon is-large">
                            <i class="fas fa-mouse-pointer fa-2x"></i>
                        </span>
                    </p>
                    <p class="has-text-grey mt-3">
                        Select an item to view its properties
                    </p>
                </div>
            `;
        }
    }
};

// Auto-populate properties when work item editor is opened
document.addEventListener('work-item-editor-opened', function (event) {
    if (event.detail && event.detail.workItem) {
        window.PropertiesPanel.showWorkItemProperties(event.detail.workItem);

        // Auto-switch to properties tab
        const chatPanel = document.querySelector('.chat-panel');
        if (chatPanel && chatPanel._x_dataStack && chatPanel._x_dataStack[0]) {
            chatPanel._x_dataStack[0].activeTab = 'properties';
        }
    }
});

// Clear properties when editor is closed
document.addEventListener('work-item-editor-closed', function () {
    window.PropertiesPanel.clear();
});
