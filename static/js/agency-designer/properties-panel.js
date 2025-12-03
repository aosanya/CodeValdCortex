/**
 * Properties Panel Manager
 * Handles displaying and updating properties in the right sidebar
 */

window.PropertiesPanel = {
    /**
     * Show work item properties in the properties panel
     */
    showWorkItemProperties: function (workItem) {
        const panel = document.getElementById('properties-panel-content');
        if (!panel) {
            console.error('Properties panel not found');
            return;
        }

        const html = `
            <div class="box">
                <h4 class="title is-6 mb-4">
                    <span class="icon-text">
                        <span class="icon has-text-primary">
                            <i class="fas fa-tasks"></i>
                        </span>
                        <span>Work Item Properties</span>
                    </span>
                </h4>

                <div class="field">
                    <label class="label">Code</label>
                    <div class="control">
                        <input type="text" 
                               class="input is-static" 
                               value="${workItem.code || ''}" 
                               readonly>
                    </div>
                </div>

                <div class="field">
                    <label class="label">Title</label>
                    <div class="control">
                        <input type="text" 
                               class="input" 
                               value="${workItem.title || ''}" 
                               onchange="window.PropertiesPanel.updateWorkItemField('title', this.value)"
                               placeholder="Work item title">
                    </div>
                </div>

                <div class="field">
                    <label class="label">Description</label>
                    <div class="control">
                        <textarea 
                            class="textarea" 
                            rows="4"
                            onchange="window.PropertiesPanel.updateWorkItemField('description', this.value)"
                            placeholder="Work item description">${workItem.description || ''}</textarea>
                    </div>
                </div>

                <div class="field">
                    <label class="label">Status</label>
                    <div class="control">
                        <div class="select is-fullwidth">
                            <select onchange="window.PropertiesPanel.updateWorkItemField('status', this.value)">
                                <option value="not-started" ${workItem.status === 'not-started' ? 'selected' : ''}>Not Started</option>
                                <option value="in-progress" ${workItem.status === 'in-progress' ? 'selected' : ''}>In Progress</option>
                                <option value="completed" ${workItem.status === 'completed' ? 'selected' : ''}>Completed</option>
                                <option value="blocked" ${workItem.status === 'blocked' ? 'selected' : ''}>Blocked</option>
                            </select>
                        </div>
                    </div>
                </div>

                <div class="field">
                    <label class="label">Priority</label>
                    <div class="control">
                        <div class="select is-fullwidth">
                            <select onchange="window.PropertiesPanel.updateWorkItemField('priority', this.value)">
                                <option value="low" ${workItem.priority === 'low' ? 'selected' : ''}>Low</option>
                                <option value="medium" ${workItem.priority === 'medium' ? 'selected' : ''}>Medium</option>
                                <option value="high" ${workItem.priority === 'high' ? 'selected' : ''}>High</option>
                                <option value="critical" ${workItem.priority === 'critical' ? 'selected' : ''}>Critical</option>
                            </select>
                        </div>
                    </div>
                </div>

                <div class="field">
                    <label class="label">Tags</label>
                    <div class="control">
                        <div class="tags">
                            ${workItem.tags && workItem.tags.length > 0
                ? workItem.tags.map(tag => `
                                    <span class="tag is-info is-light">
                                        ${tag}
                                        <button class="delete is-small" 
                                                onclick="window.PropertiesPanel.removeTag('${tag}')"></button>
                                    </span>
                                  `).join('')
                : '<span class="has-text-grey-light is-size-7">No tags</span>'
            }
                        </div>
                    </div>
                    <p class="help">Manage tags in the work item editor</p>
                </div>

                <div class="field">
                    <label class="label">Deliverables Count</label>
                    <div class="control">
                        <span class="tag is-primary is-light">
                            <span class="icon">
                                <i class="fas fa-folder-tree"></i>
                            </span>
                            <span>${this.countDeliverables(workItem.deliverables_structured || [])} items</span>
                        </span>
                    </div>
                    <p class="help">View and edit in the deliverables tree</p>
                </div>

                <hr>

                <div class="buttons">
                    <button class="button is-small is-light" 
                            onclick="window.PropertiesPanel.switchToChat()">
                        <span class="icon">
                            <i class="fas fa-comments"></i>
                        </span>
                        <span>Back to Chat</span>
                    </button>
                </div>
            </div>
        `;

        panel.innerHTML = html;
    },

    /**
     * Show deliverable node properties in the properties panel
     */
    showDeliverableNodeProperties: function (node) {
        const panel = document.getElementById('properties-panel-content');
        if (!panel) {
            console.error('Properties panel not found');
            return;
        }

        // Auto-switch to properties tab
        const chatPanel = document.querySelector('.chat-panel');
        if (chatPanel && chatPanel._x_dataStack && chatPanel._x_dataStack[0]) {
            chatPanel._x_dataStack[0].activeTab = 'properties';
        }

        const html = `
            <div class="box">
                <h4 class="title is-6 mb-4">
                    <span class="icon-text">
                        <span class="icon has-text-${node.type === 'folder' ? 'info' : 'primary'}">
                            <i class="fas fa-${node.type === 'folder' ? 'folder' : 'file-lines'}"></i>
                        </span>
                        <span>Deliverable: ${node.name}${node.type === 'file' ? (node.file_extension || '') : ''}</span>
                    </span>
                </h4>

                <div class="field">
                    <label class="label">Name</label>
                    <div class="control">
                        <input type="text" 
                               class="input" 
                               value="${node.name || ''}" 
                               onchange="window.PropertiesPanel.updateDeliverableNodeField('${node.id}', 'name', this.value)"
                               placeholder="Node name">
                    </div>
                </div>

                ${node.type === 'file' ? `
                <div class="field">
                    <label class="label">File Extension</label>
                    <div class="control">
                        <input type="text" 
                               class="input" 
                               value="${node.file_extension || '.md'}" 
                               disabled
                               placeholder=".md">
                    </div>
                    <p class="help">Currently only .md (Markdown) files are supported</p>
                </div>
                ` : ''}

                <div class="field">
                    <label class="label">Description</label>
                    <div class="control">
                        <input type="text" 
                               class="input" 
                               value="${node.description || ''}" 
                               onchange="window.PropertiesPanel.updateDeliverableNodeField('${node.id}', 'description', this.value)"
                               placeholder="Brief description">
                    </div>
                </div>

                <div class="field">
                    <label class="label">AI Prompt Instructions</label>
                    <div class="control">
                        <textarea 
                            class="textarea" 
                            rows="6"
                            onchange="window.PropertiesPanel.updateDeliverableNodeField('${node.id}', 'prompt_instructions', this.value)"
                            placeholder="Describe what content should be included in this file/folder...">${node.prompt_instructions || ''}</textarea>
                    </div>
                    <p class="help">
                        These instructions guide AI agents on what to include when generating this deliverable.
                    </p>
                </div>

                <div class="field">
                    <label class="label">Path</label>
                    <div class="control">
                        <input type="text" 
                               class="input is-static" 
                               value="${node.path || node.name}" 
                               readonly>
                    </div>
                    <p class="help">Full path in the deliverables tree</p>
                </div>

                <hr>

                <div class="buttons">
                    <button class="button is-small is-danger is-light" 
                            onclick="if(confirm('Delete this node?')) { window.PropertiesPanel.deleteDeliverableNode('${node.id}'); }">
                        <span class="icon">
                            <i class="fas fa-trash"></i>
                        </span>
                        <span>Delete Node</span>
                    </button>
                    <button class="button is-small is-light" 
                            onclick="window.PropertiesPanel.clear()">
                        <span class="icon">
                            <i class="fas fa-times"></i>
                        </span>
                        <span>Close</span>
                    </button>
                </div>
            </div>
        `;

        panel.innerHTML = html;
    },

    /**
     * Update a deliverable node field
     */
    updateDeliverableNodeField: function (nodeId, field, value) {
        console.log(`Updating deliverable node field: ${field} =`, value);

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
     * Delete a deliverable node
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
        console.log(`Updating work item field: ${field} =`, value);

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
        console.log('Removing tag:', tag);
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
