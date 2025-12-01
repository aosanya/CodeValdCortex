// Deliverable Tree Integration for Work Items
// Handles integration between deliverable tree builder and work item editor

// Toggle between tree builder and simple text mode
window.toggleDeliverableMode = function (useSimple) {
    const treeContainer = document.getElementById('deliverable-tree-container');
    const simpleInput = document.getElementById('deliverable-simple-input');

    if (useSimple) {
        treeContainer.classList.add('is-hidden');
        simpleInput.classList.remove('is-hidden');
    } else {
        treeContainer.classList.remove('is-hidden');
        simpleInput.classList.add('is-hidden');
    }
};

// Initialize deliverable tree builder for work item editor
window.initDeliverableTreeBuilder = function (agencyId, workItemCode, existingDeliverables = []) {
    const container = document.getElementById('deliverable-tree-container');
    if (!container) {
        console.error('Deliverable tree container not found');
        return;
    }

    // Create the tree builder HTML structure
    const treeHTML = createTreeBuilderHTML(agencyId, workItemCode, existingDeliverables);
    container.innerHTML = treeHTML;

    // Initialize Alpine.js component if not already initialized
    // Alpine.js will auto-initialize when the HTML is rendered
};

// Create tree builder HTML structure
function createTreeBuilderHTML(agencyId, workItemCode, deliverables) {
    return `
        <div class="deliverable-tree-builder box" x-data="deliverableTree()">
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
                 x-init="initTree(${JSON.stringify(deliverables)})">
                
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
    const hiddenInput = document.getElementById('deliverables-structured-data');
    if (!hiddenInput || !hiddenInput.value) {
        return null;
    }

    try {
        return JSON.parse(hiddenInput.value);
    } catch (error) {
        console.error('Failed to parse deliverables structured data:', error);
        return null;
    }
};

// Check if using simple mode
window.isUsingSimpleDeliverables = function () {
    const checkbox = document.getElementById('use-simple-deliverables');
    return checkbox ? checkbox.checked : false;
};
