/**
 * Deliverable Tree Builder
 * Alpine.js component for managing hierarchical deliverable structures
 */

function deliverableTree() {
    return {
        nodes: [],
        validationErrors: [],
        draggedNode: null,
        newlyCreatedNodeId: null,
        expandedNodes: {}, // Track expanded state of all nodes
        editingNodeId: null, // Track which node is currently being edited
        selectedNodeId: null, // Track which node is selected for detail view
        onSave: null, // Callback function for saving (set via x-init)

        /**
         * Initialize the tree with existing deliverables
         */
        initTree(data) {
            if (data && data.deliverables && Array.isArray(data.deliverables)) {
                this.nodes = JSON.parse(JSON.stringify(data.deliverables));
                this.computeAllPaths();

                // Expand root nodes by default
                this.nodes.forEach(node => {
                    if (node.type === 'folder') {
                        this.expandedNodes[node.id] = true;
                    }
                });
            }
        },

        /**
         * Initialize watchers (called by Alpine.js)
         */
        init() {
            // Watch for selectedNodeId changes and update properties panel
            this.$watch('selectedNodeId', (newNodeId) => {
                if (newNodeId && window.PropertiesPanel && window.PropertiesPanel.showDeliverableNodeProperties) {
                    const node = this.findNodeById(newNodeId);
                    if (node) {
                        window.PropertiesPanel.showDeliverableNodeProperties(node);
                    }
                }
            });
        },

        /**
         * Get the currently selected node
         */
        getSelectedNode() {
            if (!this.selectedNodeId) return null;
            return this.findNodeById(this.selectedNodeId);
        },

        /**
         * Update a specific field of a node
         */
        updateNodeField(nodeId, field, value) {
            const node = this.findNodeById(nodeId);
            if (node) {
                node[field] = value;
                if (field === 'name') {
                    this.computeAllPaths();
                }

                // Don't regenerate the properties panel - it's already showing this node
                // and regenerating would lose event handlers and user state
                console.log('Field updated:', field, '=', value);
            }
        },

        /**
         * Watch for selected node changes and update properties panel
         */
        $watch(prop, callback) {
            // Alpine.js will handle this automatically through Alpine's reactivity
        },

        /**
         * Toggle node expanded state
         */
        toggleExpanded(nodeId) {
            this.expandedNodes[nodeId] = !this.expandedNodes[nodeId];
        },

        /**
         * Check if node is expanded
         */
        isExpanded(nodeId) {
            return this.expandedNodes[nodeId] === true;
        },

        /**
         * Start editing a node (closes all other editing states)
         */
        startEditing(nodeId) {
            this.editingNodeId = nodeId;
        },

        /**
         * Check if a node is being edited
         */
        isEditing(nodeId) {
            return this.editingNodeId === nodeId;
        },

        /**
         * Stop editing the current node
         */
        stopEditing() {
            this.editingNodeId = null;
        },

        /**
         * Render the tree as HTML (for Alpine x-html)
         */
        renderTree() {
            if (this.nodes.length === 0) {
                return '';
            }

            return `<div class="tree-nodes">${this.nodes.map((node, index) => this.renderNode(node, 0, index)).join('')}</div>`;
        },

        /**
         * Render a single node with its children
         */
        renderNode(node, depth, nodeIndex) {
            const paddingLeft = depth * 12;
            const hasChildren = node.children && node.children.length > 0;
            const nodeId = node.id;
            // Check if this is a newly created node that should start in edit mode
            const isNewlyCreated = this.newlyCreatedNodeId === nodeId;

            return `
                <div class="tree-node ${node.type === 'folder' ? 'is-folder' : 'is-file'}" 
                     data-node-id="${nodeId}"
                     data-depth="${depth}"
                     :class="selectedNodeId === '${nodeId}' ? 'is-selected' : ''"
                     x-init="${isNewlyCreated ? `$nextTick(() => { startEditing('${nodeId}'); $el.querySelector('input[type=text]')?.focus(); }); newlyCreatedNodeId = null;` : ''}">
                    
                    <div class="node-row level is-mobile" 
                         style="margin-bottom: 0.25rem; cursor: pointer;"
                         @click="selectedNodeId = '${nodeId}'">
                        <div class="level-left">
                            ${node.type === 'folder' ? `
                                <div class="level-item">
                                    <button type="button" class="button is-small is-ghost" 
                                            @click.stop="toggleExpanded('${nodeId}')">
                                        <span class="icon is-small">
                                            <i class="fas" :class="isExpanded('${nodeId}') ? 'fa-chevron-down' : 'fa-chevron-right'"></i>
                                        </span>
                                    </button>
                                </div>
                            ` : `<div class="level-item" style="width: 32px;"></div>`}
                            
                            <div class="level-item">
                                <span class="icon ${node.type === 'folder' ? (hasChildren ? 'has-text-info' : 'has-text-grey-light') : 'has-text-grey'}">
                                    <i class="fas ${node.type === 'folder' ? (hasChildren ? 'fa-folder' : 'fa-folder-open') : 'fa-file-lines'}"></i>
                                </span>
                            </div>
                            
                            <div class="level-item">
                                <div x-show="!isEditing('${nodeId}')" class="node-name">
                                    <span>${node.name}${node.type === 'file' ? node.file_extension : ''}</span>
                                </div>
                                <div x-show="isEditing('${nodeId}')" class="field is-grouped">
                                    <p class="control">
                                        <input type="text" class="input is-small" 
                                               value="${node.name}"
                                               @keydown.enter="stopEditing(); updateNodeName('${nodeId}', $event.target.value)"
                                               @keydown.escape="stopEditing()"
                                               @blur="stopEditing(); updateNodeName('${nodeId}', $event.target.value)"
                                               placeholder="Name">
                                    </p>
                                    ${node.type === 'file' ? '<p class="control"><span class="button is-small is-static">.md</span></p>' : ''}
                                </div>
                            </div>
                        </div>
                        
                        <div class="level-right">
                            <div class="level-item">
                                <div class="buttons are-small">
                                    <button type="button" class="button is-small is-ghost"
                                            @click.stop="startEditing('${nodeId}')"
                                            title="Edit name">
                                        <span class="icon is-small">
                                            <i class="fas fa-pen"></i>
                                        </span>
                                    </button>
                                    
                                    <button type="button" class="button is-small is-ghost"
                                            @click.stop="editPrompt('${nodeId}')"
                                            title="${node.prompt_instructions ? 'Edit instructions' : 'Add instructions'}">
                                        <span class="icon is-small ${node.prompt_instructions ? 'has-text-success' : 'has-text-grey-light'}">
                                            <i class="fas fa-message-lines"></i>
                                        </span>
                                    </button>
                                    
                                    ${node.type === 'folder' ? `
                                        <div class="dropdown is-hoverable is-right">
                                            <div class="dropdown-trigger">
                                                <button type="button" class="button is-small is-ghost" title="Add child" @click.stop="">
                                                    <span class="icon is-small">
                                                        <i class="fas fa-plus"></i>
                                                    </span>
                                                </button>
                                            </div>
                                            <div class="dropdown-menu">
                                                <div class="dropdown-content">
                                                    <a class="dropdown-item" @click.prevent="addChildFolder('${nodeId}')">
                                                        <span class="icon-text">
                                                            <span class="icon"><i class="fas fa-folder"></i></span>
                                                            <span>Folder</span>
                                                        </span>
                                                    </a>
                                                    <a class="dropdown-item" @click.prevent="addChildFile('${nodeId}')">
                                                        <span class="icon-text">
                                                            <span class="icon"><i class="fas fa-file"></i></span>
                                                            <span>File</span>
                                                        </span>
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                    ` : ''}
                                    
                                    <div class="dropdown is-hoverable is-right">
                                        <div class="dropdown-trigger">
                                            <button type="button" class="button is-small is-ghost" title="Move" @click.stop="">
                                                <span class="icon is-small">
                                                    <i class="fas fa-arrows-up-down-left-right"></i>
                                                </span>
                                            </button>
                                        </div>
                                        <div class="dropdown-menu">
                                            <div class="dropdown-content">
                                                ${depth > 0 ? `
                                                    <a class="dropdown-item" @click.prevent="moveNodeUp('${nodeId}')">
                                                        <span class="icon-text">
                                                            <span class="icon"><i class="fas fa-arrow-up"></i></span>
                                                            <span>Move Up One Level</span>
                                                        </span>
                                                    </a>
                                                    <hr class="dropdown-divider">
                                                ` : ''}
                                                <a class="dropdown-item" @click.prevent="window.dispatchEvent(new CustomEvent('show-move-to-folder-modal', { detail: { nodeId: '${nodeId}' } }))">
                                                    <span class="icon-text">
                                                        <span class="icon"><i class="fas fa-folder-arrow-up"></i></span>
                                                        <span>Move to Folder...</span>
                                                    </span>
                                                </a>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <button type="button" class="button is-small is-ghost has-text-danger"
                                            @click.stop="deleteNode('${nodeId}')"
                                            title="Delete">
                                        <span class="icon is-small">
                                            <i class="fas fa-trash"></i>
                                        </span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    ${node.type === 'folder' && hasChildren ? `
                        <div x-show="isExpanded('${nodeId}')" class="node-children">
                            ${this.getSortedChildren(node.children).map((child, childIndex) => this.renderNode(child, depth + 1, childIndex)).join('')}
                        </div>
                    ` : ''}
                </div>
            `;
        },

        /**
         * Sort children alphabetically (folders first, then files)
         */
        getSortedChildren(children) {
            if (!children || children.length === 0) return [];

            return [...children].sort((a, b) => {
                // Folders first, then files
                if (a.type === 'folder' && b.type === 'file') return -1;
                if (a.type === 'file' && b.type === 'folder') return 1;

                // Within same type, sort alphabetically by name
                return a.name.localeCompare(b.name, undefined, { sensitivity: 'base' });
            });
        },

        /**
         * Update node name
         */
        updateNodeName(nodeId, newName) {
            const node = this.findNodeById(nodeId);
            if (node) {
                node.name = newName;
                this.computeAllPaths();
                this.validate();
            }
        },

        /**
         * Edit prompt for a node (dispatches event for modal)
         */
        editPrompt(nodeId) {
            window.dispatchEvent(new CustomEvent('edit-prompt', { detail: { nodeId } }));
        },

        /**
         * Generate a unique ID for new nodes
         */
        generateId() {
            return 'node-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        },

        /**
         * Add a new root folder
         */
        addRootFolder() {
            const nodeId = this.generateId();
            const node = {
                id: nodeId,
                name: 'New Folder',
                description: '',
                path: 'New Folder',
                type: 'folder',
                prompt_instructions: '',
                children: [],
                file_extension: '',
                order: this.nodes.length
            };
            this.nodes.push(node);
            this.newlyCreatedNodeId = nodeId;
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Add a new root file
         */
        addRootFile() {
            const nodeId = this.generateId();
            const node = {
                id: nodeId,
                name: 'new-file',
                description: '',
                path: 'new-file.md',
                type: 'file',
                prompt_instructions: '',
                children: [],
                file_extension: '.md',
                order: this.nodes.length
            };
            this.nodes.push(node);
            this.newlyCreatedNodeId = nodeId;
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Add a child folder to a parent node
         */
        addChildFolder(parentId) {
            const parent = this.findNodeById(parentId);
            if (!parent || parent.type !== 'folder') {
                return;
            }

            if (!parent.children) {
                parent.children = [];
            }

            const nodeId = this.generateId();
            const node = {
                id: nodeId,
                name: 'New Folder',
                description: '',
                path: '',
                type: 'folder',
                prompt_instructions: '',
                children: [],
                file_extension: '',
                parent_id: parentId,
                order: parent.children.length
            };

            parent.children.push(node);
            this.newlyCreatedNodeId = nodeId;
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Add a child file to a parent folder
         */
        addChildFile(parentId) {
            const parent = this.findNodeById(parentId);
            if (!parent || parent.type !== 'folder') {
                return;
            }

            if (!parent.children) {
                parent.children = [];
            }

            const nodeId = this.generateId();
            const node = {
                id: nodeId,
                name: 'new-file',
                description: '',
                path: '',
                type: 'file',
                prompt_instructions: '',
                children: [],
                file_extension: '.md',
                parent_id: parentId,
                order: parent.children.length
            };

            parent.children.push(node);
            this.newlyCreatedNodeId = nodeId;
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Delete a node by ID
         */
        deleteNode(nodeId) {
            if (!confirm('Are you sure you want to delete this item and all its children?')) {
                return;
            }

            // Try to find and remove from root
            const rootIndex = this.nodes.findIndex(n => n.id === nodeId);
            if (rootIndex !== -1) {
                this.nodes.splice(rootIndex, 1);
                this.computeAllPaths();
                this.validate();
                return;
            }

            // Recursively search and remove from children
            this.deleteNodeRecursive(this.nodes, nodeId);
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Recursively delete a node from the tree
         */
        deleteNodeRecursive(nodes, nodeId) {
            for (let i = 0; i < nodes.length; i++) {
                if (nodes[i].children && Array.isArray(nodes[i].children)) {
                    const childIndex = nodes[i].children.findIndex(c => c.id === nodeId);
                    if (childIndex !== -1) {
                        nodes[i].children.splice(childIndex, 1);
                        return true;
                    }
                    if (this.deleteNodeRecursive(nodes[i].children, nodeId)) {
                        return true;
                    }
                }
            }
            return false;
        },

        /**
         * Find a node by ID
         */
        findNodeById(nodeId) {
            return this.findNodeByIdRecursive(this.nodes, nodeId);
        },

        /**
         * Move node up one level (to parent's parent)
         */
        moveNodeUp(nodeId) {
            const result = this.findNodeWithParent(nodeId);
            if (!result || !result.parent || !result.grandparent) {
                alert('Cannot move this node up - it is already at the top level or has no valid parent.');
                return;
            }

            const { node, parent, grandparent } = result;

            // Remove from current parent
            const childIndex = parent.children.findIndex(c => c.id === nodeId);
            if (childIndex !== -1) {
                parent.children.splice(childIndex, 1);
            }

            // Add to grandparent's children (after the parent)
            const parentIndex = grandparent.children.findIndex(c => c.id === parent.id);
            grandparent.children.splice(parentIndex + 1, 0, node);

            this.computeAllPaths();
            this.validate();
        },

        /**
         * Move node to another folder
         */
        moveNodeToFolder(nodeId, targetFolderId) {
            if (nodeId === targetFolderId) {
                alert('Cannot move a node to itself.');
                return;
            }

            // Check if target is a descendant of the node being moved
            if (this.isDescendant(nodeId, targetFolderId)) {
                alert('Cannot move a node into its own descendant.');
                return;
            }

            const node = this.findNodeById(nodeId);
            const targetFolder = this.findNodeById(targetFolderId);

            if (!node) {
                alert('Source node not found.');
                return;
            }

            if (!targetFolder || targetFolder.type !== 'folder') {
                alert('Target must be a folder.');
                return;
            }

            // Remove from current location
            this.removeNodeFromParent(nodeId);

            // Add to target folder
            if (!targetFolder.children) {
                targetFolder.children = [];
            }
            targetFolder.children.push(node);

            this.computeAllPaths();
            this.validate();
        },

        /**
         * Check if targetId is a descendant of nodeId
         */
        isDescendant(nodeId, targetId) {
            const node = this.findNodeById(nodeId);
            if (!node || !node.children) return false;

            return this.isDescendantRecursive(node.children, targetId);
        },

        /**
         * Recursively check if targetId exists in children
         */
        isDescendantRecursive(children, targetId) {
            for (const child of children) {
                if (child.id === targetId) return true;
                if (child.children && this.isDescendantRecursive(child.children, targetId)) {
                    return true;
                }
            }
            return false;
        },

        /**
         * Remove node from its parent (keeps the node object intact)
         */
        removeNodeFromParent(nodeId) {
            // Try to remove from root
            const rootIndex = this.nodes.findIndex(n => n.id === nodeId);
            if (rootIndex !== -1) {
                this.nodes.splice(rootIndex, 1);
                return true;
            }

            // Recursively search and remove from children
            return this.removeNodeFromParentRecursive(this.nodes, nodeId);
        },

        /**
         * Recursively remove node from parent's children
         */
        removeNodeFromParentRecursive(nodes, nodeId) {
            for (let i = 0; i < nodes.length; i++) {
                if (nodes[i].children && Array.isArray(nodes[i].children)) {
                    const childIndex = nodes[i].children.findIndex(c => c.id === nodeId);
                    if (childIndex !== -1) {
                        nodes[i].children.splice(childIndex, 1);
                        return true;
                    }
                    if (this.removeNodeFromParentRecursive(nodes[i].children, nodeId)) {
                        return true;
                    }
                }
            }
            return false;
        },

        /**
         * Find node along with its parent and grandparent
         */
        findNodeWithParent(nodeId) {
            return this.findNodeWithParentRecursive(this.nodes, nodeId, null, null);
        },

        /**
         * Recursively find node with parent context
         */
        findNodeWithParentRecursive(nodes, targetId, parent, grandparent) {
            for (const node of nodes) {
                if (node.id === targetId) {
                    return { node, parent, grandparent };
                }
                if (node.children && Array.isArray(node.children)) {
                    const result = this.findNodeWithParentRecursive(
                        node.children,
                        targetId,
                        node,
                        parent
                    );
                    if (result) {
                        return result;
                    }
                }
            }
            return null;
        },

        /**
         * Get all folders (for move-to dropdown)
         */
        getAllFolders() {
            const folders = [];
            this.collectFoldersRecursive(this.nodes, folders);
            return folders;
        },

        /**
         * Recursively collect all folders
         */
        collectFoldersRecursive(nodes, result) {
            for (const node of nodes) {
                if (node.type === 'folder') {
                    result.push(node);
                    if (node.children && Array.isArray(node.children)) {
                        this.collectFoldersRecursive(node.children, result);
                    }
                }
            }
        },

        /**
         * Recursively find a node by ID
         */
        findNodeByIdRecursive(nodes, nodeId) {
            for (const node of nodes) {
                if (node.id === nodeId) {
                    return node;
                }
                if (node.children && Array.isArray(node.children)) {
                    const found = this.findNodeByIdRecursive(node.children, nodeId);
                    if (found) {
                        return found;
                    }
                }
            }
            return null;
        },

        /**
         * Compute paths for all nodes
         */
        computeAllPaths() {
            this.computePathsRecursive(this.nodes, '');
        },

        /**
         * Recursively compute paths
         */
        computePathsRecursive(nodes, parentPath) {
            for (const node of nodes) {
                if (parentPath === '') {
                    node.path = node.type === 'file'
                        ? node.name + node.file_extension
                        : node.name;
                } else {
                    node.path = node.type === 'file'
                        ? parentPath + '/' + node.name + node.file_extension
                        : parentPath + '/' + node.name;
                }

                if (node.children && Array.isArray(node.children)) {
                    this.computePathsRecursive(node.children, node.path);
                }
            }
        },

        /**
         * Validate the entire tree
         */
        validate() {
            this.validationErrors = [];

            // Check for duplicate IDs
            const ids = new Set();
            this.checkDuplicateIds(this.nodes, ids);

            // Validate each node
            this.validateNodesRecursive(this.nodes, 1);

            return this.validationErrors.length === 0;
        },

        /**
         * Check for duplicate IDs
         */
        checkDuplicateIds(nodes, ids) {
            for (const node of nodes) {
                if (ids.has(node.id)) {
                    this.validationErrors.push(`Duplicate ID found: ${node.id}`);
                }
                ids.add(node.id);

                if (node.children && Array.isArray(node.children)) {
                    this.checkDuplicateIds(node.children, ids);
                }
            }
        },

        /**
         * Validate nodes recursively
         */
        validateNodesRecursive(nodes, depth) {
            for (const node of nodes) {
                // Check depth
                if (depth > 10) {
                    this.validationErrors.push(`Node "${node.name}" exceeds maximum nesting depth of 10`);
                }

                // Check name
                if (!node.name || node.name.trim() === '') {
                    this.validationErrors.push(`Node has empty name`);
                }

                if (node.name && node.name.length > 255) {
                    this.validationErrors.push(`Node "${node.name}" name exceeds 255 characters`);
                }

                // Check prompt length
                if (node.prompt_instructions && node.prompt_instructions.length > 5000) {
                    this.validationErrors.push(`Node "${node.name}" prompt exceeds 5000 characters`);
                }

                // Check file extension for files
                if (node.type === 'file' && !node.file_extension) {
                    this.validationErrors.push(`File "${node.name}" is missing extension`);
                }

                // Check folder doesn't have extension
                if (node.type === 'folder' && node.file_extension) {
                    this.validationErrors.push(`Folder "${node.name}" cannot have file extension`);
                }

                // Check file doesn't have children
                if (node.type === 'file' && node.children && node.children.length > 0) {
                    this.validationErrors.push(`File "${node.name}" cannot have children`);
                }

                // Check children count
                if (node.children && node.children.length > 100) {
                    this.validationErrors.push(`Node "${node.name}" has too many children (max 100)`);
                }

                // Recurse into children
                if (node.children && Array.isArray(node.children)) {
                    this.validateNodesRecursive(node.children, depth + 1);
                }
            }
        },

        /**
         * Expand all folders
         */
        expandAll() {
            // This is handled by Alpine's x-data in the template
            // We dispatch an event to notify all nodes
            window.dispatchEvent(new CustomEvent('expand-all-nodes'));
        },

        /**
         * Collapse all folders
         */
        collapseAll() {
            window.dispatchEvent(new CustomEvent('collapse-all-nodes'));
        },

        /**
         * Count total nodes
         */
        countNodes(nodes) {
            let count = nodes.length;
            for (const node of nodes) {
                if (node.children && Array.isArray(node.children)) {
                    count += this.countNodes(node.children);
                }
            }
            return count;
        },

        /**
         * Count files
         */
        countFiles(nodes) {
            let count = 0;
            for (const node of nodes) {
                if (node.type === 'file') {
                    count++;
                }
                if (node.children && Array.isArray(node.children)) {
                    count += this.countFiles(node.children);
                }
            }
            return count;
        },

        /**
         * Count folders
         */
        countFolders(nodes) {
            let count = 0;
            for (const node of nodes) {
                if (node.type === 'folder') {
                    count++;
                }
                if (node.children && Array.isArray(node.children)) {
                    count += this.countFolders(node.children);
                }
            }
            return count;
        },

        /**
         * Get maximum depth
         */
        getMaxDepth(nodes) {
            if (nodes.length === 0) {
                return 0;
            }

            let maxDepth = 1;
            for (const node of nodes) {
                if (node.children && Array.isArray(node.children) && node.children.length > 0) {
                    const childDepth = 1 + this.getMaxDepth(node.children);
                    maxDepth = Math.max(maxDepth, childDepth);
                }
            }
            return maxDepth;
        },

        /**
         * Check if a node has parent instructions
         */
        hasParentInstructions(nodeId) {
            const parents = this.buildParentChain(nodeId);
            return parents.some(p => p.prompt_instructions && p.prompt_instructions.trim() !== '');
        },

        /**
         * Build parent chain for a node
         */
        buildParentChain(nodeId) {
            const parents = [];
            this.buildParentChainRecursive(this.nodes, nodeId, parents, []);
            return parents;
        },

        /**
         * Recursively build parent chain
         */
        buildParentChainRecursive(nodes, targetId, result, currentChain) {
            for (const node of nodes) {
                if (node.id === targetId) {
                    result.push(...currentChain);
                    return true;
                }

                if (node.children && Array.isArray(node.children)) {
                    const newChain = [...currentChain, node];
                    if (this.buildParentChainRecursive(node.children, targetId, result, newChain)) {
                        return true;
                    }
                }
            }
            return false;
        }
    };
}

// Make it available globally
window.deliverableTree = deliverableTree;
