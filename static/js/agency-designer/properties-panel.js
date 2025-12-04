/**
 * Properties Panel Manager
 * Generic, reusable properties panel with configurable save handlers
 */

window.PropertiesPanel = {
    // Store current configuration
    _currentConfig: null,

    // Store node being enhanced for AI workflow
    _nodeBeingEnhanced: null,

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
            case 'ai-enhance':
                // Use chat workflow for AI enhance
                this._sendAIEnhanceToChat();
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
                    action: 'ai-enhance',
                    label: 'AI Enhance',
                    icon: 'sparkles',
                    class: 'is-info'
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
     * Send AI enhance request through chat workflow
     */
    _sendAIEnhanceToChat: function () {
        if (!this._currentConfig || !this._currentConfig.data) {
            return;
        }

        const node = this._currentConfig.data;

        // Store the node being enhanced so we can update it when AI responds
        this._nodeBeingEnhanced = node;

        // Generate unique progress token for this request (timestamp + random)
        const progressToken = `PROG_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

        // Store token so chat-streaming can use it to parse progress tags
        this._currentProgressToken = progressToken;

        // Also store globally for chat-streaming.js to access
        window.currentProgressToken = progressToken;

        // Build enhancement message based on node data
        let enhanceMessage = `DELIVERABLE ENHANCEMENT REQUEST\n\n`;
        enhanceMessage += `Please enhance this deliverable node and return the result as JSON.\n\n`;
        enhanceMessage += `**Current Node:**\n`;
        enhanceMessage += `- Name: ${node.name}\n`;
        enhanceMessage += `- Type: ${node.type}\n`;

        if (node.description) {
            enhanceMessage += `- Description: ${node.description}\n`;
        }

        if (node.prompt_instructions) {
            enhanceMessage += `- Prompt Instructions: ${node.prompt_instructions}\n`;
        }

        enhanceMessage += `\n**Enhancement Tasks:**\n`;

        if (node.type === 'folder') {
            enhanceMessage += `1. Improve the description to be clear and actionable\n`;
            enhanceMessage += `2. Enhance the prompt instructions for AI generation\n`;
            enhanceMessage += `3. Suggest logical child nodes (files/subfolders) if needed\n\n`;
            enhanceMessage += `**IMPORTANT:** Return your response as valid JSON with these fields:\n`;
            enhanceMessage += `- name (string): Enhanced or original name\n`;
            enhanceMessage += `- description (string): Clear description\n`;
            enhanceMessage += `- prompt_instructions (string): Comprehensive AI instructions\n`;
            enhanceMessage += `- suggested_children (array): List of child nodes with name, description, type, file_extension, prompt_instructions, order\n`;
            enhanceMessage += `- enhancement_explanation (string): What was improved\n`;
            enhanceMessage += `- was_changed (boolean): Whether changes were made\n\n`;
            enhanceMessage += `Include progress tags using this EXACT format: <${progressToken}>message</${progressToken}> as you work. Return ONLY the JSON object, wrapped in \`\`\`json code fence.`;
        } else {
            enhanceMessage += `1. Improve the description to be clear and actionable\n`;
            enhanceMessage += `2. Enhance the prompt instructions for AI generation\n`;
            enhanceMessage += `3. Suggest any improvements to make this deliverable more effective\n\n`;
            enhanceMessage += `**IMPORTANT:** Return your response as valid JSON with these fields:\n`;
            enhanceMessage += `- name (string): Enhanced or original name\n`;
            enhanceMessage += `- description (string): Clear description\n`;
            enhanceMessage += `- prompt_instructions (string): Comprehensive AI instructions\n`;
            enhanceMessage += `- suggested_children (array): Empty for files\n`;
            enhanceMessage += `- enhancement_explanation (string): What was improved\n`;
            enhanceMessage += `- was_changed (boolean): Whether changes were made\n\n`;
            enhanceMessage += `Include progress tags using this EXACT format: <${progressToken}>message</${progressToken}> as you work. Return ONLY the JSON object, wrapped in \`\`\`json code fence.`;
        }

        // Show status in status bar
        if (window.showNotification) {
            window.showNotification('AI request: Enhancing deliverable node...', 'info');
        }

        // Add deliverable enhancement context
        if (window.ContextManager && window.ContextManager.ContextType) {
            window.ContextManager.createContext(
                window.ContextManager.ContextType.DELIVERABLE_NODE,
                `DELIV-${node.id}`,
                `Deliverable Enhancement: ${node.name} (${node.type})`,
                {
                    nodeId: node.id,
                    nodeName: node.name,
                    nodeType: node.type,
                    isEnhancementRequest: true
                }
            );
        }

        // Switch to chat tab
        this.switchToChat();

        // Send message through chat
        const userInput = document.getElementById('user-input');
        const chatForm = userInput ? userInput.closest('form') : null;

        if (!userInput || !chatForm) {
            if (window.showNotification) {
                window.showNotification('Chat interface not available', 'error');
            }
            return;
        }

        // Set the message in the input
        userInput.value = enhanceMessage;

        // Trigger the form submission
        const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
        chatForm.dispatchEvent(submitEvent);


        // Show status indicator
        const statusBar = document.querySelector('.status-notification');
        if (statusBar) {
            statusBar.textContent = 'AI Processing...';
            statusBar.classList.add('is-info');
            statusBar.classList.remove('is-hidden');

            // Hide after streaming completes (listen for streaming done event)
            setTimeout(() => {
                statusBar.classList.add('is-hidden');
            }, 30000); // Hide after 30 seconds max
        }
    },

    /**
     * Apply AI enhancement suggestions to the current node
     * @param {Object} enhancement - The enhancement object from AI response
     */
    applyAIEnhancement: function (enhancement) {
        if (!this._nodeBeingEnhanced) {
            window.showNotification('No node to enhance', 'warning');
            return;
        }

        const node = this._nodeBeingEnhanced;

        // Get the Alpine.js tree component
        const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
        if (!treeContainer || !treeContainer._x_dataStack || !treeContainer._x_dataStack[0]) {
            window.showNotification('Could not find deliverable tree', 'error');
            return;
        }

        const alpineData = treeContainer._x_dataStack[0];

        // Update node fields
        if (enhancement.name && enhancement.name !== node.name) {
            if (alpineData.updateNodeField) {
                alpineData.updateNodeField(node.id, 'name', enhancement.name);
            }
        }

        if (enhancement.description) {
            if (alpineData.updateNodeField) {
                alpineData.updateNodeField(node.id, 'description', enhancement.description);
            }
        }

        if (enhancement.prompt_instructions) {
            if (alpineData.updateNodeField) {
                alpineData.updateNodeField(node.id, 'prompt_instructions', enhancement.prompt_instructions);
            }
        }

        // Add suggested children if this is a folder and children were suggested
        if (node.type === 'folder' && enhancement.suggested_children && enhancement.suggested_children.length > 0) {

            // Find the parent node in the tree
            const parentNode = alpineData.findNodeById(node.id);
            if (parentNode) {
                if (!parentNode.children) {
                    parentNode.children = [];
                }

                enhancement.suggested_children.forEach((child, index) => {
                    // Generate a unique ID for the child
                    const childId = 'node-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9) + '-' + index;

                    const newChild = {
                        id: childId,
                        name: child.name || 'New Item',
                        description: child.description || '',
                        type: child.type || 'file',
                        prompt_instructions: child.prompt_instructions || '',
                        file_extension: child.type === 'file' ? (child.file_extension || '.md') : '',
                        children: child.type === 'folder' ? [] : undefined,
                        order: child.order !== undefined ? child.order : index
                    };

                    parentNode.children.push(newChild);
                });

                // Recompute paths and validate
                if (alpineData.computeAllPaths) {
                    alpineData.computeAllPaths();
                }
                if (alpineData.validate) {
                    alpineData.validate();
                }
            }
        }

        // Save the tree
        if (alpineData.onSave && typeof alpineData.onSave === 'function') {
            alpineData.onSave().then(async () => {
                window.showNotification('AI enhancements applied successfully!', 'success');

                // Reload the work item from server to get the fresh data with all enhancements
                // This ensures the tree shows the newly added children
                const agencyId = window.getCurrentAgencyId();
                const workItemCode = this._currentWorkItemCode || '';


                if (agencyId && workItemCode) {
                    try {
                        // Fetch fresh work item data from server
                        const workItems = await window.specificationAPI.getWorkItems();

                        // Work items are stored with 'code' field, so we search by code
                        const workItem = workItems.find(wi => wi.code === workItemCode);

                        if (workItem && workItem.deliverables_structured) {

                            // Reinitialize the tree with fresh data from server
                            if (typeof window.initDeliverableTreeBuilder === 'function') {
                                window.initDeliverableTreeBuilder(
                                    agencyId,
                                    workItemCode,
                                    workItem.deliverables_structured,
                                    alpineData.onSave
                                );

                                // Expand the parent node to show new children
                                setTimeout(() => {
                                    const treeData = window.Alpine ? window.Alpine.$data(document.querySelector('[x-data*="deliverableTree"]')) : null;
                                    if (treeData && node.type === 'folder' && enhancement.suggested_children && enhancement.suggested_children.length > 0) {
                                        treeData.expandedNodes[node.id] = true;
                                        // Force reactivity
                                        treeData.nodes = [...treeData.nodes];
                                    }
                                }, 100);
                            } else {
                            }
                        } else {
                        }
                    } catch (error) {
                        // Fallback to local update if server reload fails
                        if (alpineData.computeAllPaths) {
                            alpineData.computeAllPaths();
                        }
                        if (alpineData.nodes) {
                            alpineData.nodes = [...alpineData.nodes];
                        }
                    }
                } else {
                }

                // Switch back to properties tab to show updated node
                const chatPanel = document.querySelector('.chat-panel');
                if (chatPanel && chatPanel._x_dataStack && chatPanel._x_dataStack[0]) {
                    chatPanel._x_dataStack[0].activeTab = 'properties';
                }

                // Refresh properties panel with updated node
                setTimeout(() => {
                    const treeData = window.Alpine ? window.Alpine.$data(document.querySelector('[x-data*="deliverableTree"]')) : null;
                    if (treeData) {
                        const updatedNode = treeData.findNodeById(node.id);
                        if (updatedNode) {
                            this.showDeliverableNodeProperties(updatedNode);
                        }
                    }
                }, 200);
            }).catch(error => {
                window.showNotification('Failed to save enhancements', 'error');
            });
        } else {
            window.showNotification('Enhancement applied (save manually)', 'warning');
        }

        // Clear the node being enhanced
        this._nodeBeingEnhanced = null;
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
