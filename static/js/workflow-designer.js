/**
 * Simplified Workflow Designer with Alpine.js
 * Uses vertical column layout with HTML5 drag-and-drop
 */

window.workflowDesigner = function () {
    return {
        // Workflow data
        agencyID: '',
        workflowID: '',
        workflowKey: '',
        workflowName: '',
        workflowDescription: '',
        workflowVersion: '',
        workflowSteps: [],

        // All agency workflows (for specification update)
        allWorkflows: [],

        // Available work items
        availableWorkItems: [],
        filteredWorkItems: [],
        searchQuery: '',

        // Drag state
        isDragging: false,
        draggedItem: null,
        draggedFromStep: null, // { stepIndex, itemIndex } if dragging from workflow
        dragOverTarget: null,
        sideDropTarget: null,

        // UI state
        saving: false,

        /**
         * Initialize the designer
         */
        async init() {
            // Get workflow data from data attributes
            const container = this.$el;

            console.log('Container element:', container);
            console.log('All data attributes:', container.dataset);

            this.agencyID = container.dataset.agencyId;
            this.workflowID = container.dataset.workflowId;
            this.workflowKey = container.dataset.workflowKey;
            this.workflowName = container.dataset.workflowName;
            this.workflowDescription = container.dataset.workflowDescription;
            this.workflowVersion = container.dataset.workflowVersion;

            console.log('Parsed workflow data:', {
                agencyID: this.agencyID,
                workflowID: this.workflowID,
                workflowKey: this.workflowKey,
                workflowName: this.workflowName
            });

            // Parse existing workflow steps
            try {
                const stepsData = container.dataset.workflowSteps;
                if (stepsData && stepsData !== 'null') {
                    this.workflowSteps = JSON.parse(stepsData) || [];
                }
            } catch (e) {
                console.error('Failed to parse workflow steps:', e);
                this.workflowSteps = [];
            }

            // Load available work items from specification API
            await this.loadWorkItems();

            console.log('Workflow Designer initialized:', {
                workflowID: this.workflowID,
                steps: this.workflowSteps.length,
                workItems: this.availableWorkItems.length
            });
        },

        /**
         * Load available work items from specification
         */
        async loadWorkItems() {
            console.log('loadWorkItems called with agencyID:', this.agencyID);

            try {
                // Use existing specification API
                if (typeof window.specificationAPI !== 'undefined') {
                    // Set the agency ID on the API instance
                    window.specificationAPI.agencyId = this.agencyID;

                    console.log('Fetching specification from:', window.specificationAPI.baseUrl);

                    const spec = await window.specificationAPI.getSpecification();

                    console.log('Raw specification response:', spec);

                    this.availableWorkItems = spec.work_items || [];
                    this.filteredWorkItems = [...this.availableWorkItems];

                    // Load all workflows for specification updates
                    this.allWorkflows = spec.workflows || [];

                    console.log('Loaded specification:', {
                        workItems: this.availableWorkItems.length,
                        workflows: this.allWorkflows.length,
                        firstWorkItem: this.availableWorkItems[0]
                    });
                } else {
                    console.error('Specification API not available!');
                    this.availableWorkItems = [];
                    this.filteredWorkItems = [];
                    this.allWorkflows = [];
                }
            } catch (error) {
                console.error('Failed to load work items:', error);
                this.availableWorkItems = [];
                this.filteredWorkItems = [];
                this.allWorkflows = [];
            }
        },

        /**
         * Toggle work item description visibility
         */
        toggleItemDescription(itemKey) {
            const item = this.filteredWorkItems.find(i => i._key === itemKey);
            if (item) {
                item.showDescription = !item.showDescription;
            }
        },

        /**
         * Toggle step item description visibility
         */
        toggleStepItemDescription(stepId, workItemId) {
            const step = this.workflowSteps.find(s => s.id === stepId);
            if (step) {
                const item = step.items.find(i => i.work_item_id === workItemId);
                if (item) {
                    item.showDescription = !item.showDescription;
                }
            }
        },

        /**
         * Filter work items based on search query
         */
        filterWorkItems() {
            const query = this.searchQuery.toLowerCase().trim();
            if (!query) {
                this.filteredWorkItems = [...this.availableWorkItems];
                return;
            }

            this.filteredWorkItems = this.availableWorkItems.filter(item => {
                const title = (item.title || '').toLowerCase();
                const description = (item.description || '').toLowerCase();
                return title.includes(query) || description.includes(query);
            });
        },

        /**
         * Handle drag start from work items panel
         */
        handleDragStart(event, item) {
            this.isDragging = true;
            this.draggedItem = item;
            this.draggedFromStep = null; // Not from workflow
            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData('text/plain', JSON.stringify(item));
            event.target.classList.add('is-dragging');
        },

        /**
         * Handle drag start from existing step item
         */
        handleStepDragStart(event, stepIndex, itemIndex) {
            const step = this.workflowSteps[stepIndex];
            if (!step) return;

            const item = step.items[itemIndex];
            if (!item) return;

            this.isDragging = true;
            this.draggedItem = item;
            this.draggedFromStep = { stepIndex, itemIndex };
            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData('text/plain', JSON.stringify(item));
            event.target.classList.add('is-dragging');
        },

        /**
         * Handle drag end
         */
        handleDragEnd(event) {
            this.isDragging = false;
            this.draggedItem = null;
            this.draggedFromStep = null;
            this.dragOverTarget = null;
            this.sideDropTarget = null;
            event.target.classList.remove('is-dragging');
        },

        /**
         * Handle drag over drop zone
         */
        handleDragOver(event, target) {
            event.preventDefault();
            this.dragOverTarget = target;
        },

        /**
         * Handle drag over item for side drop zones
         */
        handleItemDragOver(event, stepIndex) {
            if (!this.isDragging) return;
            event.preventDefault();

            // Determine which side based on mouse position
            const rect = event.currentTarget.getBoundingClientRect();
            const midX = rect.left + rect.width / 2;
            const side = event.clientX < midX ? 'left' : 'right';

            this.sideDropTarget = { stepIndex, side };
        },

        /**
         * Handle drag leave
         */
        handleDragLeave(event) {
            // Only clear if we're actually leaving the drop zone
            if (event.currentTarget === event.target) {
                this.dragOverTarget = null;
            }
        },

        /**
         * Handle drag leave from item
         */
        handleItemDragLeave(event) {
            if (event.currentTarget === event.target) {
                this.sideDropTarget = null;
            }
        },

        /**
         * Handle drop on workflow
         */
        handleDrop(event, position, dropType) {
            event.preventDefault();
            event.stopPropagation();

            if (!this.draggedItem) return;

            // Remove from original position if dragging from workflow
            if (this.draggedFromStep !== null) {
                this.removeItemFromStep(
                    this.draggedFromStep.stepIndex,
                    this.draggedFromStep.itemIndex,
                    false // Don't save yet
                );
            }

            // Add to new position based on drop type
            switch (dropType) {
                case 'empty':
                    // First item in empty workflow
                    this.addStepAt(0, [this.draggedItem]);
                    break;

                case 'before':
                    // Insert new step before position
                    this.addStepAt(position, [this.draggedItem]);
                    break;

                case 'after':
                    // Insert new step after last position
                    this.addStepAt(position, [this.draggedItem]);
                    break;

                case 'left':
                case 'right':
                    // Add as parallel item
                    this.addParallelItem(position, this.draggedItem, dropType);
                    break;

                case 'parallel':
                    // Add to existing parallel execution
                    this.addParallelItem(position, this.draggedItem, 'add');
                    break;
            }

            // Clear drag state
            this.isDragging = false;
            this.draggedItem = null;
            this.draggedFromStep = null;
            this.dragOverTarget = null;
            this.sideDropTarget = null;

            // Save workflow
            this.saveWorkflow();
        },

        /**
         * Add a new step at specified position
         */
        addStepAt(position, items) {
            const newStep = {
                id: this.generateID(),
                order: position,
                items: items.map(item => ({
                    id: this.generateID(),
                    work_item_id: item.key || item._key,
                    work_item_title: item.title,
                    description: item.description || '',
                    showDescription: false
                }))
            };

            // Insert at position
            this.workflowSteps.splice(position, 0, newStep);

            // Reorder all steps
            this.reorderSteps();
        },

        /**
         * Add parallel item to existing step
         */
        addParallelItem(stepIndex, item, side) {
            const step = this.workflowSteps[stepIndex];
            if (!step) return;

            const newItem = {
                id: this.generateID(),
                work_item_id: item.key || item._key,
                work_item_title: item.title,
                description: item.description || '',
                showDescription: false
            };

            // Add to items array (left = prepend, right/add = append)
            if (side === 'left') {
                step.items.unshift(newItem);
            } else {
                // 'right' or 'add' both append to the end
                step.items.push(newItem);
            }
        },

        /**
         * Remove item from step
         */
        removeItemFromStep(stepIndex, itemIndex, shouldSave = true) {
            const step = this.workflowSteps[stepIndex];
            if (!step) return;

            // Remove the item
            step.items.splice(itemIndex, 1);

            // If step is now empty, remove the step
            if (step.items.length === 0) {
                this.workflowSteps.splice(stepIndex, 1);
                this.reorderSteps();
            }

            if (shouldSave) {
                this.saveWorkflow();
            }
        },

        /**
         * Reorder steps to have sequential order values
         */
        reorderSteps() {
            this.workflowSteps.forEach((step, index) => {
                step.order = index;
            });
        },

        /**
         * Save workflow to backend (via specification workflows endpoint)
         */
        async saveWorkflow() {
            this.saving = true;

            try {
                console.log('🔍 saveWorkflow DEBUG:', {
                    workflowKey: this.workflowKey,
                    allWorkflowsCount: this.allWorkflows.length,
                    allWorkflowKeys: this.allWorkflows.map(wf => ({ key: wf.key, name: wf.name }))
                });

                // Find and update the current workflow in the allWorkflows array
                const workflowIndex = this.allWorkflows.findIndex(wf => wf.key === this.workflowKey);

                console.log('🔍 Workflow search result:', { 
                    workflowIndex, 
                    found: workflowIndex >= 0,
                    searchKey: this.workflowKey 
                });

                if (workflowIndex >= 0) {
                    // Update existing workflow - preserve the key explicitly!
                    const existingKey = this.allWorkflows[workflowIndex].key;
                    this.allWorkflows[workflowIndex] = {
                        ...this.allWorkflows[workflowIndex],
                        key: existingKey,  // Explicitly preserve the key
                        name: this.workflowName,
                        description: this.workflowDescription,
                        version: this.workflowVersion,
                        steps: this.workflowSteps
                    };
                    console.log('✅ Updated existing workflow at index', workflowIndex, 'with key:', existingKey);
                } else {
                    console.warn('⚠️ Workflow NOT found in allWorkflows, adding new entry');
                    // Workflow not found in allWorkflows, add it
                    this.allWorkflows.push({
                        key: this.workflowKey,
                        agency_id: this.agencyID,
                        name: this.workflowName,
                        description: this.workflowDescription,
                        version: this.workflowVersion,
                        steps: this.workflowSteps
                    });
                }

                console.log('📤 Sending workflows to server:', this.allWorkflows.map(wf => ({ 
                    key: wf.key, 
                    name: wf.name,
                    hasKey: !!wf.key 
                })));

                // Save all workflows via specification endpoint
                const response = await fetch(`/api/v1/agencies/${this.agencyID}/specification/workflows`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        workflows: this.allWorkflows,
                        updated_by: 'system' // TODO: Get from auth context
                    })
                });

                if (!response.ok) {
                    const errorData = await response.json();
                    throw new Error(errorData.error || `Failed to save workflow: ${response.statusText}`);
                }

                const updatedSpec = await response.json();
                console.log('✅ Workflow saved successfully to specification');
                console.log('📥 Server returned workflows:', updatedSpec.workflows?.map(wf => ({ 
                    key: wf.key, 
                    name: wf.name 
                })));

                // Update allWorkflows with the response to stay in sync
                this.allWorkflows = updatedSpec.workflows || [];

            } catch (error) {
                console.error('❌ Failed to save workflow:', error);
                alert(`Failed to save workflow: ${error.message}`);
            } finally {
                this.saving = false;
            }
        },

        /**
         * Export workflow as JSON
         */
        exportWorkflow() {
            const data = {
                name: this.workflowName,
                description: this.workflowDescription,
                version: this.workflowVersion,
                steps: this.workflowSteps
            };

            const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${this.workflowKey}-${new Date().toISOString().split('T')[0]}.json`;
            a.click();
            URL.revokeObjectURL(url);
        },

        /**
         * Generate unique ID
         */
        generateID() {
            return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        }
    };
};
