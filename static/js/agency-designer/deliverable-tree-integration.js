// Deliverable Tree Integration for Work Items
// Handles integration between deliverable tree builder and work item editor
// Updated: 2025-12-02 - Removed simple text mode, tree builder only

// Initialize deliverable tree builder for work item editor
window.initDeliverableTreeBuilder = function (agencyId, workItemCode, existingDeliverables = []) {
    console.log('[MVP-054] initDeliverableTreeBuilder called with:', {
        agencyId,
        workItemCode,
        deliverables: existingDeliverables,
        count: existingDeliverables.length
    });

    const container = document.getElementById('deliverable-tree-container');
    if (!container) {
        console.error('[MVP-054] Deliverable tree container not found');
        return;
    }

    // Create the tree builder HTML structure
    const treeHTML = createTreeBuilderHTML(agencyId, workItemCode, existingDeliverables);
    container.innerHTML = treeHTML;

    console.log('[MVP-054] Tree HTML inserted, Alpine.js should initialize now');

    // Initialize Alpine.js component if not already initialized
    // Alpine.js will auto-initialize when the HTML is rendered
};

// Create tree builder HTML structure
function createTreeBuilderHTML(agencyId, workItemCode, deliverables) {
    // Store deliverables in a global temporary variable that Alpine can access
    // This avoids HTML escaping issues with complex JSON in attributes
    window.__tempDeliverables = deliverables || [];

    console.log('[MVP-054] createTreeBuilderHTML: deliverables count=', deliverables?.length || 0);

    return `
        <div class="deliverable-tree-builder" 
             x-data="deliverableTree()"
             x-init="nodes = window.__tempDeliverables || []; computeAllPaths(); window.__tempDeliverables = null; console.log('[MVP-054] Alpine initialized with', nodes.length, 'nodes')">
            <!-- Toolbar -->
            <div class="level mb-4">
                <div class="level-left">
                    <div class="level-item">
                        <h3 class="title is-5">
                            <span class="icon-text">
                                <span class="icon">
                                    <i class="fas fa-folder-tree"></i>
                                </span>
                                <span>Deliverables Structure</span>
                            </span>
                        </h3>
                    </div>
                </div>
                <div class="level-right">
                    <div class="level-item">
                        <div class="buttons">
                            <button 
                                type="button"
                                class="button is-small is-info"
                                @click="addRootFolder()"
                                title="Add root folder">
                                <span class="icon is-small">
                                    <i class="fas fa-folder-plus"></i>
                                </span>
                                <span>Add Folder</span>
                            </button>
                            <button 
                                type="button"
                                class="button is-small is-primary"
                                @click="addRootFile()"
                                title="Add root file">
                                <span class="icon is-small">
                                    <i class="fas fa-file-circle-plus"></i>
                                </span>
                                <span>Add File</span>
                            </button>
                            <button 
                                type="button"
                                class="button is-small is-light"
                                @click="expandAll()"
                                title="Expand all folders">
                                <span class="icon is-small">
                                    <i class="fas fa-angles-down"></i>
                                </span>
                            </button>
                            <button 
                                type="button"
                                class="button is-small is-light"
                                @click="collapseAll()"
                                title="Collapse all folders">
                                <span class="icon is-small">
                                    <i class="fas fa-angles-up"></i>
                                </span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Tree Container -->
            <div class="deliverable-tree-container" 
                 id="deliverable-tree-root"
                 x-init="if (window.__tempDeliverables && window.__tempDeliverables.length > 0) { nodes = window.__tempDeliverables; computeAllPaths(); }">
                
                <!-- Empty state -->
                <template x-if="nodes.length === 0">
                    <div class="notification is-light has-text-centered">
                        <p class="mb-3">
                            <span class="icon is-large has-text-grey-light">
                                <i class="fas fa-folder-tree fa-3x"></i>
                            </span>
                        </p>
                        <p class="has-text-grey">
                            No deliverables defined yet. Add folders and files to define the structure of artifacts this work item will produce.
                        </p>
                        <p class="mt-4">
                            <button 
                                type="button"
                                class="button is-info"
                                @click="addRootFolder()">
                                <span class="icon">
                                    <i class="fas fa-folder-plus"></i>
                                </span>
                                <span>Create First Folder</span>
                            </button>
                            <button 
                                type="button"
                                class="button is-primary"
                                @click="addRootFile()">
                                <span class="icon">
                                    <i class="fas fa-file-circle-plus"></i>
                                </span>
                                <span>Create First File</span>
                            </button>
                        </p>
                    </div>
                </template>

                <!-- Tree nodes will be rendered by JavaScript -->
                <div id="tree-nodes-container" x-html="renderTree()"></div>
            </div>

            <!-- Hidden input to store tree data -->
            <input 
                type="hidden" 
                name="deliverables_structured"
                :value="JSON.stringify(nodes)"
                id="deliverables-structured-data"/>

            <!-- Validation summary -->
            <div class="mt-4" x-show="validationErrors.length > 0">
                <div class="notification is-warning is-light">
                    <p class="has-text-weight-semibold mb-2">
                        <span class="icon">
                            <i class="fas fa-triangle-exclamation"></i>
                        </span>
                        Validation Issues:
                    </p>
                    <ul>
                        <template x-for="error in validationErrors" :key="error">
                            <li x-text="error" class="ml-4"></li>
                        </template>
                    </ul>
                </div>
            </div>

            <!-- Stats -->
            <div class="level mt-4 is-size-7 has-text-grey">
                <div class="level-left">
                    <div class="level-item">
                        <span x-text="\`\${countNodes(nodes)} total items\`"></span>
                    </div>
                    <div class="level-item">
                        <span x-text="\`\${countFiles(nodes)} files\`"></span>
                    </div>
                    <div class="level-item">
                        <span x-text="\`\${countFolders(nodes)} folders\`"></span>
                    </div>
                    <div class="level-item">
                        <span x-text="\`Max depth: \${getMaxDepth(nodes)}\`"></span>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// Get deliverables structured data from tree builder
window.getDeliverablesStructuredData = function () {
    // Try to get data from Alpine.js component directly
    const treeContainer = document.querySelector('[x-data*="deliverableTree"]');
    if (treeContainer && treeContainer._x_dataStack && treeContainer._x_dataStack.length > 0) {
        const alpineData = treeContainer._x_dataStack[0];
        if (alpineData && alpineData.nodes) {
            console.log('Got deliverables from Alpine.js:', alpineData.nodes);
            return alpineData.nodes;
        }
    }

    // Fallback: try to get from hidden input
    const hiddenInput = document.getElementById('deliverables-structured-data');
    if (hiddenInput && hiddenInput.value) {
        try {
            const data = JSON.parse(hiddenInput.value);
            console.log('Got deliverables from hidden input:', data);
            return data;
        } catch (error) {
            console.error('Failed to parse deliverables structured data:', error);
        }
    }

    console.warn('No deliverables data found');
    return null;
};

// Handle move-to-folder modal
window.addEventListener('show-move-to-folder-modal', function (event) {
    const { nodeId } = event.detail;
    const treeContainer = document.querySelector('[x-data*="deliverableTree"]');

    if (!treeContainer || !treeContainer._x_dataStack || treeContainer._x_dataStack.length === 0) {
        alert('Tree not initialized');
        return;
    }

    const alpineData = treeContainer._x_dataStack[0];
    const node = alpineData.findNodeById(nodeId);
    const folders = alpineData.getAllFolders().filter(f => f.id !== nodeId);

    if (folders.length === 0) {
        alert('No folders available to move to.');
        return;
    }

    // Create modal HTML
    showMoveToFolderModal(nodeId, node, folders, alpineData);
});

function showMoveToFolderModal(nodeId, node, folders, alpineData) {
    const modalHTML = `
        <div class="modal is-active" id="move-to-folder-modal">
            <div class="modal-background" onclick="document.getElementById('move-to-folder-modal').remove()"></div>
            <div class="modal-card">
                <header class="modal-card-head">
                    <p class="modal-card-title">Move "${node.name}" to Folder</p>
                    <button class="delete" aria-label="close" onclick="document.getElementById('move-to-folder-modal').remove()"></button>
                </header>
                <section class="modal-card-body">
                    <div class="field">
                        <label class="label">Select Target Folder</label>
                        <div class="control">
                            <div class="select is-fullwidth">
                                <select id="target-folder-select">
                                    ${folders.map(f => `<option value="${f.id}">${f.path || f.name}</option>`).join('')}
                                </select>
                            </div>
                        </div>
                    </div>
                </section>
                <footer class="modal-card-foot">
                    <button class="button is-success" onclick="executeMove('${nodeId}')">Move</button>
                    <button class="button" onclick="document.getElementById('move-to-folder-modal').remove()">Cancel</button>
                </footer>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // Store alpineData reference for executeMove
    window.__moveAlpineData = alpineData;
}

window.executeMove = function (nodeId) {
    const select = document.getElementById('target-folder-select');
    const targetFolderId = select.value;
    const alpineData = window.__moveAlpineData;

    if (alpineData && targetFolderId) {
        alpineData.moveNodeToFolder(nodeId, targetFolderId);
        document.getElementById('move-to-folder-modal').remove();
        delete window.__moveAlpineData;
    }
};
