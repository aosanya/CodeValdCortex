/**
 * Deliverable Tree Builder
 * Alpine.js component for managing hierarchical deliverable structures
 */

function deliverableTree() {
    return {
        nodes: [],
        validationErrors: [],
        draggedNode: null,

        /**
         * Initialize the tree with existing deliverables
         */
        initTree(data) {
            if (data && data.deliverables && Array.isArray(data.deliverables)) {
                this.nodes = JSON.parse(JSON.stringify(data.deliverables));
                this.computeAllPaths();
            }
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
            const paddingLeft = depth * 20;
            const hasChildren = node.children && node.children.length > 0;
            const nodeId = node.id;
            // Sanitize node ID for use in Alpine.js variable names (replace hyphens with underscores)
            const safeNodeId = nodeId.replace(/-/g, '_');

            return `
                <div class="tree-node ${node.type === 'folder' ? 'is-folder' : 'is-file'}" 
                     style="padding-left: ${paddingLeft}px"
                     data-node-id="${nodeId}"
                     x-data="{ editing_${safeNodeId}: false, expanded_${safeNodeId}: true }">
                    
                    <div class="node-row level is-mobile mb-1">
                        <div class="level-left">
                            ${node.type === 'folder' && hasChildren ? `
                                <div class="level-item">
                                    <button type="button" class="button is-small is-ghost" 
                                            @click="expanded_${safeNodeId} = !expanded_${safeNodeId}">
                                        <span class="icon is-small">
                                            <i class="fas" :class="expanded_${safeNodeId} ? 'fa-chevron-down' : 'fa-chevron-right'"></i>
                                        </span>
                                    </button>
                                </div>
                            ` : `<div class="level-item" style="width: 32px;"></div>`}
                            
                            <div class="level-item">
                                <span class="icon ${node.type === 'folder' ? 'has-text-info' : 'has-text-grey'}">
                                    <i class="fas ${node.type === 'folder' ? 'fa-folder' : 'fa-file-lines'}"></i>
                                </span>
                            </div>
                            
                            <div class="level-item">
                                <div x-show="!editing_${safeNodeId}" class="node-name">
                                    <span>${node.name}${node.type === 'file' ? node.file_extension : ''}</span>
                                    ${node.description ? `<span class="has-text-grey is-size-7 ml-2">- ${node.description}</span>` : ''}
                                </div>
                                <div x-show="editing_${safeNodeId}" class="field is-grouped">
                                    <p class="control">
                                        <input type="text" class="input is-small" 
                                               value="${node.name}"
                                               @keydown.enter="editing_${safeNodeId} = false; updateNodeName('${nodeId}', $event.target.value)"
                                               @keydown.escape="editing_${safeNodeId} = false"
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
                                            @click="editing_${safeNodeId} = !editing_${safeNodeId}"
                                            title="Edit name">
                                        <span class="icon is-small">
                                            <i class="fas fa-pen"></i>
                                        </span>
                                    </button>
                                    
                                    <button type="button" class="button is-small is-ghost"
                                            @click="editPrompt('${nodeId}')"
                                            title="${node.prompt_instructions ? 'Edit instructions' : 'Add instructions'}">
                                        <span class="icon is-small ${node.prompt_instructions ? 'has-text-success' : 'has-text-grey-light'}">
                                            <i class="fas fa-message-lines"></i>
                                        </span>
                                    </button>
                                    
                                    ${node.type === 'folder' ? `
                                        <div class="dropdown is-hoverable is-right">
                                            <div class="dropdown-trigger">
                                                <button type="button" class="button is-small is-ghost" title="Add child">
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
                                    
                                    <button type="button" class="button is-small is-ghost has-text-danger"
                                            @click="deleteNode('${nodeId}')"
                                            title="Delete">
                                        <span class="icon is-small">
                                            <i class="fas fa-trash"></i>
                                        </span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    ${node.prompt_instructions && !editing ? `
                        <div x-show="!editing_${safeNodeId}" class="node-prompt-preview ml-4 mb-2">
                            <div class="message is-small is-light">
                                <div class="message-body py-2 px-3 is-size-7">
                                    <span class="icon is-small has-text-grey-light">
                                        <i class="fas fa-quote-left"></i>
                                    </span>
                                    <span>${node.prompt_instructions.substring(0, 150)}${node.prompt_instructions.length > 150 ? '...' : ''}</span>
                                </div>
                            </div>
                        </div>
                    ` : ''}
                    
                    ${node.type === 'folder' && hasChildren ? `
                        <div x-show="expanded_${safeNodeId}" class="node-children">
                            ${node.children.map((child, childIndex) => this.renderNode(child, depth + 1, childIndex)).join('')}
                        </div>
                    ` : ''}
                </div>
            `;
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
            const node = {
                id: this.generateId(),
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
            this.computeAllPaths();
            this.validate();
        },

        /**
         * Add a new root file
         */
        addRootFile() {
            const node = {
                id: this.generateId(),
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

            const node = {
                id: this.generateId(),
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

            const node = {
                id: this.generateId(),
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
