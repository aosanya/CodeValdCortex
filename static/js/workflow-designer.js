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
        saveTimeout: null,

        /**
         * Initialize the designer
         */
        async init() {
            console.log('[WF-DESIGNER] 🚀 Initializing workflow designer component');

            // Get workflow data from data attributes
            const container = this.$el;
            console.log('[WF-DESIGNER] 📦 Container element:', container);

            this.agencyID = container.dataset.agencyId;
            this.workflowID = container.dataset.workflowId;
            this.workflowKey = container.dataset.workflowKey;
            this.workflowName = container.dataset.workflowName;
            this.workflowDescription = container.dataset.workflowDescription;
            this.workflowVersion = container.dataset.workflowVersion;

            console.log('[WF-DESIGNER] 📋 Workflow metadata:', {
                agencyID: this.agencyID,
                workflowID: this.workflowID,
                workflowKey: this.workflowKey,
                workflowName: this.workflowName
            });

            // Parse workflow steps if available
            try {
                const stepsData = container.dataset.workflowSteps;
                if (stepsData && stepsData !== 'null') {
                    this.workflowSteps = JSON.parse(stepsData) || [];
                    console.log('[WF-DESIGNER] ✅ Parsed workflow steps:', this.workflowSteps.length, 'steps');
                }
            } catch (e) {
                console.error('[WF-DESIGNER] ❌ Failed to parse workflow steps:', e);
                this.workflowSteps = [];
            }

            // Check PropertiesPanel availability
            console.log('[WF-DESIGNER] 🪟 PropertiesPanel available:', !!window.PropertiesPanel);
            console.log('[WF-DESIGNER] 🛠️ PropertiesPanel methods:', window.PropertiesPanel ? Object.keys(window.PropertiesPanel) : 'N/A');

            // Load available work items from specification API
            await this.loadWorkItems();

            console.log('[WF-DESIGNER] ✅ Initialization complete');
        },

        /**
         * Load available work items from specification
         */
        async loadWorkItems() {
            console.log('[WF-DESIGNER] 📚 Loading work items...');
            try {
                // Use existing specification API
                if (typeof window.specificationAPI !== 'undefined') {
                    // Set the agency ID on the API instance
                    window.specificationAPI.agencyId = this.agencyID;

                    console.log('[WF-DESIGNER] 🔍 Fetching specification for agency:', this.agencyID);
                    const spec = await window.specificationAPI.getSpecification();
                    console.log('[WF-DESIGNER] 📦 Specification received:', spec);

                    this.availableWorkItems = spec.work_items || [];
                    this.filteredWorkItems = [...this.availableWorkItems];

                    console.log('[WF-DESIGNER] ✅ Loaded', this.availableWorkItems.length, 'work items');
                    console.log('[WF-DESIGNER] 📋 Work items:', this.availableWorkItems);

                    // Load all workflows for specification updates
                    // Normalize workflows to ensure they have a 'key' property
                    this.allWorkflows = (spec.workflows || []).map(wf => ({
                        ...wf,
                        key: wf._key
                    }));

                    // Enrich existing workflow steps with work item details
                    this.enrichWorkflowSteps();
                } else {
                    console.error('[WF-DESIGNER] ❌ specificationAPI not available');
                    this.availableWorkItems = [];
                    this.filteredWorkItems = [];
                    this.allWorkflows = [];
                }
            } catch (error) {
                console.error('[WF-DESIGNER] ❌ Error loading work items:', error);
                this.availableWorkItems = [];
                this.filteredWorkItems = [];
                this.allWorkflows = [];
            }
        },

        /**
         * Enrich workflow steps with work item details
         */
        enrichWorkflowSteps() {
            if (!this.availableWorkItems || this.availableWorkItems.length === 0) {
                return;
            }

            this.workflowSteps.forEach(step => {
                if (step.items) {
                    step.items.forEach(item => {
                        // Find the work item in availableWorkItems
                        const workItem = this.availableWorkItems.find(wi =>
                            wi._key === item.work_item_id ||
                            wi.key === item.work_item_id ||
                            wi._key === item.work_item_key ||
                            wi.key === item.work_item_key
                        );

                        if (workItem) {
                            // Enrich with work item details
                            item.work_item_title = workItem.title || workItem.name || item.work_item_name || 'Unnamed Work Item';
                            item.description = workItem.description || '';
                            item.showDescription = item.showDescription || false;
                        } else {
                            // Fallback to existing name or default
                            item.work_item_title = item.work_item_name || 'Unknown Work Item';
                            item.description = item.description || '';
                            item.showDescription = item.showDescription || false;
                        }
                    });
                }
            });
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
            console.log('[WF-DESIGNER] 🏁 handleDragStart called:', { item, event });
            this.isDragging = true;
            this.draggedItem = item;
            this.draggedFromStep = null; // Not from workflow

            // Store in window for cross-instance access (Alpine.js may create multiple instances)
            window.__workflowDraggedItem = item;
            window.__workflowDraggedFromStep = null;

            event.dataTransfer.effectAllowed = 'move';
            event.dataTransfer.setData('text/plain', JSON.stringify(item));
            event.target.classList.add('is-dragging');
            console.log('[WF-DESIGNER] ✅ Dragging item:', this.draggedItem);
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

            // Store in window for cross-instance access
            window.__workflowDraggedItem = item;
            window.__workflowDraggedFromStep = { stepIndex, itemIndex };

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

            // Clean up window storage
            window.__workflowDraggedItem = null;
            window.__workflowDraggedFromStep = null;
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

            // Get dragged item from window storage (handles cross-instance drops)
            const draggedItem = window.__workflowDraggedItem || this.draggedItem;
            const draggedFromStep = window.__workflowDraggedFromStep || this.draggedFromStep;

            console.log('[WF-DESIGNER] 🎯 handleDrop called:', { position, dropType, draggedItem, draggedFromStep });

            if (!draggedItem) {
                console.warn('[WF-DESIGNER] ⚠️ No dragged item');
                return;
            }

            // Remove from original position if dragging from workflow
            if (draggedFromStep !== null) {
                console.log('[WF-DESIGNER] 🔄 Moving from existing step:', draggedFromStep);
                this.removeItemFromStep(
                    draggedFromStep.stepIndex,
                    draggedFromStep.itemIndex,
                    false // Don't save yet
                );
            }

            // Add to new position based on drop type
            console.log('[WF-DESIGNER] 📦 Drop type:', dropType);
            switch (dropType) {
                case 'empty':
                    // First item in empty workflow
                    console.log('[WF-DESIGNER] ➕ Adding to empty workflow');
                    this.addStepAt(0, [draggedItem]);
                    break;

                case 'before':
                    // Insert new step before position
                    console.log('[WF-DESIGNER] ⬆️ Inserting before position:', position);
                    this.addStepAt(position, [draggedItem]);
                    break;

                case 'after':
                    // Insert new step after last position
                    console.log('[WF-DESIGNER] ⬇️ Inserting after position:', position);
                    this.addStepAt(position, [draggedItem]);
                    break;

                case 'left':
                case 'right':
                    // Add as parallel item
                    console.log('[WF-DESIGNER] ↔️ Adding as parallel item:', dropType);
                    this.addParallelItem(position, draggedItem, dropType);
                    break;

                case 'parallel':
                    // Add to existing parallel execution
                    console.log('[WF-DESIGNER] 🔀 Adding to parallel execution');
                    this.addParallelItem(position, draggedItem, 'add');
                    break;
            }

            // Clear drag state (both instance and window)
            this.isDragging = false;
            this.draggedItem = null;
            this.draggedFromStep = null;
            this.dragOverTarget = null;
            this.sideDropTarget = null;
            window.__workflowDraggedItem = null;
            window.__workflowDraggedFromStep = null;

            // Save workflow
            console.log('[WF-DESIGNER] 💾 Saving workflow');
            this.saveWorkflow();
        },

        /**
         * Add a new step at specified position
         */
        addStepAt(position, items) {
            console.log('[WF-DESIGNER] 📍 addStepAt called:', { position, items });

            const newStep = {
                id: this.generateID(),
                order: position,
                name: '', // Empty by default, user can set via properties panel
                description: '',
                items: items.map(item => {
                    const newItem = {
                        id: this.generateID(),
                        work_item_id: item.key || item._key || item.work_item_id,
                        work_item_key: item.key || item._key || '',
                        work_item_name: item.name || item.title || '',
                        work_item_title: item.title || item.name || '',
                        description: item.description || '',
                        showDescription: false,
                        autonomy_level: item.autonomy_level || 'L0' // Default autonomy level for work item
                    };
                    console.log('[WF-DESIGNER] ✨ Created item:', newItem);
                    return newItem;
                }),
                routes: {}, // Empty routes map
                aggregation: '', // No aggregation by default
                requires_human_decision: false,
                available_routes: []
            };

            console.log('[WF-DESIGNER] ✅ Created new step:', newStep);

            // Insert at position
            this.workflowSteps.splice(position, 0, newStep);

            // Reorder all steps
            this.reorderSteps();
        },

        /**
         * Add parallel item to existing step
         */
        addParallelItem(stepIndex, item, side) {
            console.log('[WF-DESIGNER] 🔀 addParallelItem called:', { stepIndex, item, side });

            const step = this.workflowSteps[stepIndex];
            if (!step) {
                console.error('[WF-DESIGNER] ❌ Step not found at index:', stepIndex);
                return;
            }

            const newItem = {
                id: this.generateID(),
                work_item_id: item.key || item._key || item.work_item_id,
                work_item_key: item.key || item._key || '',
                work_item_name: item.name || item.title || '',
                work_item_title: item.title || item.name || '',
                description: item.description || '',
                showDescription: false,
                autonomy_level: item.autonomy_level || 'L0' // Default autonomy level for work item
            };

            console.log('[WF-DESIGNER] ✨ Created parallel item:', newItem);

            // Add to items array (left = prepend, right/add = append)
            if (side === 'left') {
                step.items.unshift(newItem);
            } else {
                // 'right' or 'add' both append to the end
                step.items.push(newItem);
            }

            console.log('[WF-DESIGNER] ✅ Step now has', step.items.length, 'items');
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
            // Debounce: Cancel previous save timer and set a new one
            if (this.saveTimeout) {
                clearTimeout(this.saveTimeout);
            }

            this.saveTimeout = setTimeout(async () => {
                await this._performSave();
            }, 300); // 300ms debounce
        },

        /**
         * Perform the actual save operation
         */
        async _performSave() {
            // Prevent multiple simultaneous saves
            if (this.saving) {
                return;
            }

            this.saving = true;

            try {
                // Validate required data
                if (!this.workflowKey || !this.agencyID) {
                    throw new Error('Missing required workflow key or agency ID');
                }

                // Find and update the current workflow in the allWorkflows array
                const workflowIndex = this.allWorkflows.findIndex(wf => wf.key === this.workflowKey);

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
                } else {
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

                // Filter out workflows with empty keys and remove duplicates
                const workflowsToSave = this.allWorkflows.filter((workflow, index, arr) => {
                    // Remove workflows with no key
                    if (!workflow.key || workflow.key.trim() === '') {
                        return false;
                    }

                    // Remove duplicates by key (keep first occurrence)
                    const firstIndex = arr.findIndex(w => w.key === workflow.key);
                    if (firstIndex !== index) {
                        return false;
                    }

                    return true;
                });

                // Save all workflows via specification endpoint
                const response = await fetch(`/api/v1/agencies/${this.agencyID}/specification/workflows`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        workflows: workflowsToSave,
                        updated_by: 'system' // TODO: Get from auth context
                    })
                });

                if (!response.ok) {
                    const errorData = await response.json();
                    throw new Error(errorData.error || `Failed to save workflow: ${response.statusText}`);
                }

                const updatedSpec = await response.json();

                // Update allWorkflows with the response to stay in sync
                // Normalize workflows to ensure they have a 'key' property
                this.allWorkflows = (updatedSpec.workflows || []).map(wf => ({
                    ...wf,
                    key: wf._key
                }));

            } catch (error) {
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
        },

        // ===== Autonomy Level & Route Helpers =====

        /**
         * Get Bulma CSS class for autonomy badge
         */
        getAutonomyBadgeClass(level) {
            const classes = {
                'L0': 'is-light',      // Gray - Manual
                'L1': 'is-info',       // Blue - Assisted
                'L2': 'is-warning',    // Yellow - Conditional
                'L3': 'is-warning',    // Orange - High Auto (same as warning)
                'L4': 'is-success'     // Green - Full Auto
            };
            return `tag ${classes[level] || 'is-light'}`;
        },

        /**
         * Get autonomy level label
         */
        getAutonomyLabel(level) {
            const labels = {
                'L0': 'Manual',
                'L1': 'Assisted',
                'L2': 'Conditional',
                'L3': 'High Auto',
                'L4': 'Full Auto'
            };
            return labels[level] || 'Unknown';
        },

        /**
         * Get route count for a step
         */
        getRouteCount(step) {
            if (!step.routes) return 0;
            return Object.keys(step.routes).length;
        },

        /**
         * Check if step has routes defined
         */
        hasRoutes(step) {
            return this.getRouteCount(step) > 0;
        },

        /**
         * Get autonomy levels for all work items in a step
         * Returns array of unique autonomy levels
         */
        getStepAutonomyLevels(step) {
            if (!step.items || step.items.length === 0) return [];

            const levels = step.items
                .map(item => item.autonomy_level || 'L0')
                .filter((level, index, self) => self.indexOf(level) === index); // unique

            return levels.sort(); // Sort L0, L1, L2, L3, L4
        },

        /**
         * Get display text for step autonomy (shows all unique levels)
         */
        getStepAutonomyDisplay(step) {
            const levels = this.getStepAutonomyLevels(step);
            if (levels.length === 0) return 'L0';
            if (levels.length === 1) return levels[0];
            return levels.join(', '); // e.g., "L0, L2"
        },

        /**
         * Get CSS class for step autonomy badge (uses highest level)
         */
        getStepAutonomyBadgeClass(step) {
            const levels = this.getStepAutonomyLevels(step);
            if (levels.length === 0) return 'tag is-light';

            // Use highest autonomy level for color
            const highest = levels[levels.length - 1];
            return this.getAutonomyBadgeClass(highest);
        },

        /**
         * Select a step and open properties panel
         */
        selectStep(stepId) {
            console.log('🔍 selectStep called:', stepId);
            const step = this.workflowSteps.find(s => s.id === stepId);
            console.log('📌 Found step:', step);
            console.log('🪟 PropertiesPanel available:', !!window.PropertiesPanel);

            if (step && window.PropertiesPanel) {
                console.log('✅ Opening step properties');
                this.openStepProperties(step);
            } else {
                if (!step) console.error('❌ Step not found:', stepId);
                if (!window.PropertiesPanel) console.error('❌ PropertiesPanel not available');
            }
        },

        /**
         * Select a work item and open properties panel
         */
        selectWorkItem(stepId, itemId) {
            const step = this.workflowSteps.find(s => s.id === stepId);
            if (!step) return;

            const item = step.items.find(i => i.id === itemId);
            if (item && window.PropertiesPanel) {
                this.openWorkItemProperties(step, item);
            }
        },

        /**
         * Open properties panel for a step
         */
        openStepProperties(step) {
            if (!window.PropertiesPanel) {
                console.error('PropertiesPanel not available');
                return;
            }

            // Create a copy to track changes
            const stepCopy = { ...step };

            // Build step properties configuration
            const config = {
                title: `Step ${step.order} Properties`,
                icon: 'cog',
                iconColor: 'info',
                data: stepCopy,
                autoSwitchTab: false, // Don't auto-switch tabs (no chat panel in workflow designer)

                // Track field changes
                onUpdate: (field, value) => {
                    stepCopy[field] = value;
                },

                fields: [
                    // Basic Information
                    {
                        key: 'name',
                        label: 'Step Name',
                        type: 'text',
                        placeholder: 'e.g., Review Documentation',
                        help: 'Optional: Give this step a descriptive name'
                    },
                    {
                        key: 'description',
                        label: 'Description',
                        type: 'textarea',
                        placeholder: 'Describe what happens in this step...',
                        help: 'Optional: Detailed description of the step'
                    },

                    // Info: Work items in this step
                    {
                        key: 'work_items_info',
                        label: 'Work Items in This Step',
                        type: 'info',
                        value: step.items.map(item => item.work_item_title || item.work_item_name || 'Untitled').join(', '),
                        help: 'Click individual work items to edit their autonomy levels'
                    },

                    // Parallel Execution (only if multiple items)
                    ...(step.items && step.items.length > 1 ? [{
                        key: 'aggregation',
                        label: 'Aggregation Rule',
                        type: 'select',
                        options: [
                            { value: '', label: 'None (sequential)' },
                            { value: 'any', label: 'Any (first completion)' },
                            { value: 'all', label: 'All (wait for all)' },
                            { value: 'majority', label: 'Majority (>50%)' },
                            { value: 'first', label: 'First (immediate)' }
                        ],
                        help: 'How to proceed when multiple work items complete'
                    }] : []),

                    // Conditional Routing
                    {
                        key: 'routes_editor',
                        label: 'Conditional Routes',
                        type: 'custom',
                        html: this._renderRouteEditor(step, stepCopy),
                        help: 'Define where workflow goes based on step completion status'
                    }
                ],

                // Save handler
                onSave: () => {
                    this.saveStepProperties(step, stepCopy);
                },

                buttons: [
                    {
                        label: 'Save',
                        class: 'is-primary',
                        icon: 'save',
                        action: 'save'
                    },
                    {
                        label: 'Delete Step',
                        class: 'is-danger',
                        icon: 'trash',
                        action: () => this.deleteStepFromProperties(step)
                    },
                    {
                        label: 'Cancel',
                        class: 'is-light',
                        icon: 'times',
                        action: () => this.closePropertiesPanel()
                    }
                ]
            };

            window.PropertiesPanel.showProperties(config);
        },

        /**
         * Delete step from properties panel
         */
        deleteStepFromProperties(step) {
            if (!confirm(`Are you sure you want to delete Step ${step.order}? This will remove all work items in this step.`)) {
                return;
            }

            const stepIndex = this.workflowSteps.findIndex(s => s.id === step.id);
            if (stepIndex !== -1) {
                this.workflowSteps.splice(stepIndex, 1);

                // Reorder remaining steps
                this.workflowSteps.forEach((s, idx) => {
                    s.order = idx + 1;
                });

                this.saveWorkflow();
                this.closePropertiesPanel();

                if (window.showNotification) {
                    window.showNotification('Step deleted successfully', 'success');
                }
            }
        },

        /**
         * Save step properties from panel
         */
        saveStepProperties(originalStep, updatedData) {
            // Update step in workflow
            const stepIndex = this.workflowSteps.findIndex(s => s.id === originalStep.id);
            if (stepIndex !== -1) {
                // Merge updated properties (preserve items array and other data)
                this.workflowSteps[stepIndex] = {
                    ...this.workflowSteps[stepIndex],
                    name: updatedData.name || '',
                    description: updatedData.description || '',
                    autonomy_level: updatedData.autonomy_level || 'L0',
                    aggregation: updatedData.aggregation || ''
                };

                // Save workflow
                this.saveWorkflow();

                // Show success notification
                if (window.showNotification) {
                    window.showNotification('Step properties saved successfully', 'success');
                } else {
                    console.log('Step properties saved');
                }

                // Close panel
                this.closePropertiesPanel();
            }
        },

        /**
         * Close properties panel
         */
        closePropertiesPanel() {
            const panel = document.getElementById('properties-panel-content');
            if (panel) {
                panel.innerHTML = '<div class="box has-text-centered has-text-grey-light"><p>Click a step or work item to view properties</p></div>';
            }
        },

        /**
         * Open properties panel for a work item
         */
        openWorkItemProperties(step, item) {
            if (!window.PropertiesPanel) {
                console.error('PropertiesPanel not available');
                return;
            }

            // Create a copy to track changes
            const itemCopy = { ...item };

            // Build work item properties configuration
            const config = {
                title: `Work Item: ${item.work_item_title || item.work_item_name}`,
                icon: 'tasks',
                iconColor: 'primary',
                data: itemCopy,
                autoSwitchTab: false,

                // Track field changes
                onUpdate: (field, value) => {
                    itemCopy[field] = value;
                },

                fields: [
                    // Work Item Info (read-only)
                    {
                        key: 'work_item_name',
                        label: 'Work Item',
                        type: 'static',
                        help: 'Work item from specification (read-only)'
                    },
                    {
                        key: 'description',
                        label: 'Description',
                        type: 'static',
                        help: 'Work item description from specification'
                    },

                    // Autonomy Level (editable)
                    {
                        key: 'autonomy_level',
                        label: 'Autonomy Level',
                        type: 'select',
                        options: [
                            { value: 'L0', label: 'L0 - Manual (Human executes)' },
                            { value: 'L1', label: 'L1 - Assisted (AI suggests, human approves)' },
                            { value: 'L2', label: 'L2 - Conditional (AI autonomous within constraints)' },
                            { value: 'L3', label: 'L3 - High Automation (AI handles most scenarios)' },
                            { value: 'L4', label: 'L4 - Full Autonomy (AI fully independent)' }
                        ],
                        help: 'Controls how much autonomy AI agents have for this specific work item'
                    },

                    // Step Context Info
                    {
                        key: 'step_context',
                        label: 'Step Context',
                        type: 'info',
                        value: `Part of: Step ${step.order} - ${step.name || 'Unnamed Step'}`,
                        help: 'This work item belongs to the step shown above'
                    }
                ],

                // Save handler
                onSave: () => {
                    this.saveWorkItemProperties(step, item, itemCopy);
                },

                buttons: [
                    {
                        label: 'Save',
                        class: 'is-primary',
                        icon: 'save',
                        action: 'save'
                    },
                    {
                        label: 'Remove from Step',
                        class: 'is-danger',
                        icon: 'trash',
                        action: () => this.deleteWorkItemFromProperties(step, originalItem)
                    },
                    {
                        label: 'Cancel',
                        class: 'is-light',
                        icon: 'times',
                        action: () => this.closePropertiesPanel()
                    }
                ]
            };

            window.PropertiesPanel.showProperties(config);
        },

        /**
         * Delete work item from properties panel
         */
        deleteWorkItemFromProperties(step, item) {
            if (!confirm(`Remove "${item.work_item_title || item.work_item_name}" from this step?`)) {
                return;
            }

            const stepIndex = this.workflowSteps.findIndex(s => s.id === step.id);
            if (stepIndex !== -1) {
                const itemIndex = this.workflowSteps[stepIndex].items.findIndex(i => i.id === item.id);
                if (itemIndex !== -1) {
                    this.workflowSteps[stepIndex].items.splice(itemIndex, 1);

                    // If step is now empty, remove the step
                    if (this.workflowSteps[stepIndex].items.length === 0) {
                        this.workflowSteps.splice(stepIndex, 1);

                        // Reorder remaining steps
                        this.workflowSteps.forEach((s, idx) => {
                            s.order = idx + 1;
                        });

                        if (window.showNotification) {
                            window.showNotification('Work item removed and empty step deleted', 'success');
                        }
                    } else {
                        if (window.showNotification) {
                            window.showNotification('Work item removed from step', 'success');
                        }
                    }

                    this.saveWorkflow();
                    this.closePropertiesPanel();
                }
            }
        },

        /**
         * Save work item properties from panel
         */
        saveWorkItemProperties(step, originalItem, updatedData) {
            // Find the step in workflow
            const stepIndex = this.workflowSteps.findIndex(s => s.id === step.id);
            if (stepIndex === -1) return;

            // Find the item in the step
            const itemIndex = this.workflowSteps[stepIndex].items.findIndex(i => i.id === originalItem.id);
            if (itemIndex === -1) return;

            // Update the work item's autonomy level
            this.workflowSteps[stepIndex].items[itemIndex].autonomy_level = updatedData.autonomy_level || 'L0';

            // Save workflow
            this.saveWorkflow();

            // Show success notification
            if (window.showNotification) {
                window.showNotification('Work item autonomy level updated successfully', 'success');
            } else {
                console.log('Work item autonomy level saved');
            }

            // Close panel
            this.closePropertiesPanel();
        },

        /**
         * Render route editor HTML for properties panel
         */
        _renderRouteEditor(step, stepCopy) {
            const routes = step.routes || {};
            const routeEntries = Object.entries(routes);
            const availableSteps = this.workflowSteps.filter(s => s.id !== step.id);

            let html = `
                <div class="route-editor">
                    <div class="table-container">
                        <table class="table is-fullwidth is-striped is-size-7">
                            <thead>
                                <tr>
                                    <th>Status</th>
                                    <th>Target Step</th>
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody id="route-table-body">`;

            // Render existing routes
            routeEntries.forEach(([status, targetStepId], index) => {
                const targetStep = this.workflowSteps.find(s => s.id === targetStepId);
                html += this._renderRouteRow(index, status, targetStepId, targetStep, availableSteps);
            });

            // Empty state
            if (routeEntries.length === 0) {
                html += `
                    <tr id="empty-routes-message">
                        <td colspan="3" class="has-text-centered has-text-grey-light">
                            <em>No routes defined. Will proceed to next step by default.</em>
                        </td>
                    </tr>`;
            }

            html += `
                            </tbody>
                        </table>
                    </div>
                    <button 
                        type="button" 
                        class="button is-small is-info is-light"
                        onclick="window.workflowDesignerInstance.addRouteRow()"
                    >
                        <span class="icon is-small"><i class="fas fa-plus"></i></span>
                        <span>Add Route</span>
                    </button>
                </div>
            `;

            // Store reference to stepCopy for route editing
            this._currentStepCopy = stepCopy;

            return html;
        },

        /**
         * Render a single route row
         */
        _renderRouteRow(index, status, targetStepId, targetStep, availableSteps) {
            return `
                <tr data-route-index="${index}">
                    <td>
                        <div class="select is-small is-fullwidth">
                            <select onchange="window.workflowDesignerInstance.updateRoute(${index}, 'status', this.value)">
                                <option value="">Select status...</option>
                                <option value="success" ${status === 'success' ? 'selected' : ''}>Success</option>
                                <option value="failure" ${status === 'failure' ? 'selected' : ''}>Failure</option>
                                <option value="error" ${status === 'error' ? 'selected' : ''}>Error</option>
                                <option value="retry" ${status === 'retry' ? 'selected' : ''}>Retry</option>
                                <option value="skip" ${status === 'skip' ? 'selected' : ''}>Skip</option>
                            </select>
                        </div>
                    </td>
                    <td>
                        <div class="select is-small is-fullwidth">
                            <select onchange="window.workflowDesignerInstance.updateRoute(${index}, 'target', this.value)">
                                <option value="">Select target step...</option>
                                ${availableSteps.map(s => `
                                    <option value="${s.id}" ${s.id === targetStepId ? 'selected' : ''}>
                                        Step ${s.order}: ${s.name || 'Unnamed'}
                                    </option>
                                `).join('')}
                            </select>
                        </div>
                    </td>
                    <td style="width: 40px;">
                        <button 
                            type="button"
                            class="delete is-small" 
                            onclick="window.workflowDesignerInstance.removeRoute(${index})"
                            title="Remove route"
                        ></button>
                    </td>
                </tr>
            `;
        },

        /**
         * Add a new route row
         */
        addRouteRow() {
            const tbody = document.getElementById('route-table-body');
            if (!tbody) return;

            // Remove empty message if present
            const emptyMsg = document.getElementById('empty-routes-message');
            if (emptyMsg) {
                emptyMsg.remove();
            }

            // Get available steps
            const currentStep = this._currentStepCopy;
            const availableSteps = this.workflowSteps.filter(s => s.id !== currentStep.id);

            // Add new row
            const index = tbody.children.length;
            const tr = document.createElement('tr');
            tr.setAttribute('data-route-index', index);
            tr.innerHTML = this._renderRouteRow(index, '', '', null, availableSteps).replace(/<\/?tr[^>]*>/g, '');
            tbody.appendChild(tr);
        },

        /**
         * Update a route
         */
        updateRoute(index, field, value) {
            if (!this._currentStepCopy) return;

            const tbody = document.getElementById('route-table-body');
            if (!tbody) return;

            const row = tbody.children[index];
            if (!row) return;

            // Get current route data from row
            const statusSelect = row.querySelector('select');
            const targetSelect = row.querySelectorAll('select')[1];

            let status = statusSelect.value;
            let targetStepId = targetSelect.value;

            // Update based on field
            if (field === 'status') {
                status = value;
            } else if (field === 'target') {
                targetStepId = value;
            }

            // Initialize routes if needed
            if (!this._currentStepCopy.routes) {
                this._currentStepCopy.routes = {};
            }

            // Validate: both status and target must be selected
            if (status && targetStepId) {
                // Remove old status key if changing status
                if (field === 'status') {
                    const oldStatus = Object.keys(this._currentStepCopy.routes).find((key, idx) => {
                        return Array.from(tbody.children).indexOf(row) === Object.keys(this._currentStepCopy.routes).indexOf(key);
                    });
                    if (oldStatus && oldStatus !== status) {
                        delete this._currentStepCopy.routes[oldStatus];
                    }
                }

                // Set new route
                this._currentStepCopy.routes[status] = targetStepId;
            }
        },

        /**
         * Remove a route
         */
        removeRoute(index) {
            if (!this._currentStepCopy) return;

            const tbody = document.getElementById('route-table-body');
            if (!tbody) return;

            const row = tbody.children[index];
            if (!row) return;

            // Get status from row
            const statusSelect = row.querySelector('select');
            const status = statusSelect.value;

            // Remove from stepCopy routes
            if (this._currentStepCopy.routes && status) {
                delete this._currentStepCopy.routes[status];
            }

            // Remove row
            row.remove();

            // Re-index remaining rows
            Array.from(tbody.children).forEach((r, i) => {
                r.setAttribute('data-route-index', i);
            });

            // Show empty message if no routes
            if (tbody.children.length === 0) {
                const tr = document.createElement('tr');
                tr.id = 'empty-routes-message';
                tr.innerHTML = `
                    <td colspan="3" class="has-text-centered has-text-grey-light">
                        <em>No routes defined. Will proceed to next step by default.</em>
                    </td>
                `;
                tbody.appendChild(tr);
            }
        }
    };
};

// Store global instance reference for inline event handlers
let workflowDesignerInstance = null;
document.addEventListener('alpine:init', () => {
    // Capture instance when Alpine initializes
    setTimeout(() => {
        const container = document.querySelector('[x-data="workflowDesigner()"]');
        if (container && container._x_dataStack && container._x_dataStack[0]) {
            window.workflowDesignerInstance = container._x_dataStack[0];
        }
    }, 100);
});
