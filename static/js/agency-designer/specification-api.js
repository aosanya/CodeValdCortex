/**
 * Centralized API client for the unified AgencySpecification model
 * Replaces separate calls to /overview, /goals, /work-items endpoints
 */

/**
 * Specification API client - handles all specification-related API calls
 */
window.SpecificationAPI = class SpecificationAPI {
    constructor() {
        this.agencyId = window.getCurrentAgencyId ? window.getCurrentAgencyId() : null;
        this.baseUrl = `/api/v1/agencies/${this.agencyId}/specification`;
    }

    /**
     * Get the complete agency specification
     */
    async getSpecification() {
        try {
            const response = await fetch(this.baseUrl);
            if (!response.ok) {
                throw new Error(`Failed to fetch specification: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            // Return empty specification as fallback
            return {
                introduction: '',
                goals: [],
                work_items: [],
                roles: [],
                raci_matrix: null,
                version: 1,
                updated_by: 'system'
            };
        }
    }

    /**
     * Update the complete specification
     */
    async updateSpecification(updates, updatedBy = 'user') {
        try {
            const response = await fetch(this.baseUrl, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    ...updates,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update specification: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    /**
     * Update just the introduction
     */
    async updateIntroduction(introduction, updatedBy = 'user') {
        try {
            const response = await fetch(`${this.baseUrl}/introduction`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    introduction: introduction,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update introduction: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    /**
     * Update just the goals array
     */
    async updateGoals(goals, updatedBy = 'user') {
        try {
            const response = await fetch(`${this.baseUrl}/goals`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    goals: goals,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update goals: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    /**
     * Update just the work items array
     */
    async updateWorkItems(workItems, updatedBy = 'user') {
        try {
            const response = await fetch(`${this.baseUrl}/work-items`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    work_items: workItems,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update work items: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    /**
     * Update just the roles array
     */
    async updateRoles(roles, updatedBy = 'user') {
        try {
            const response = await fetch(`${this.baseUrl}/roles`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    roles: roles,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update roles: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    /**
     * Update just the workflows array
     */
    async updateWorkflows(workflows, updatedBy = 'user') {
        try {
            console.log('🚀 updateWorkflows called with', workflows.length, 'workflows');
            console.log('📦 Workflows to save:', workflows.map(w => ({
                key: w.key,
                _key: w._key,
                id: w.id,
                name: w.name,
                description: w.description?.substring(0, 50) + '...'
            })));

            // Workflows are stored in the specification, so we need to update them there
            const spec = await this.getSpecification();
            spec.workflows = workflows;
            spec.updated_by = updatedBy;

            console.log('📤 Sending PUT request to:', this.baseUrl);

            const response = await fetch(this.baseUrl, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(spec)
            });

            if (!response.ok) {
                console.error('❌ Failed to update workflows:', response.status);
                throw new Error(`Failed to update workflows: ${response.status}`);
            }

            const result = await response.json();
            console.log('✅ Workflows updated successfully');
            return result;
        } catch (error) {
            console.error('❌ Error updating workflows:', error);
            throw error;
        }
    }

    /**
     * Update just the RACI matrix
     */
    async updateRACIMatrix(raciMatrix, updatedBy = 'user') {
        try {
            const response = await fetch(`${this.baseUrl}/raci-matrix`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    raci_matrix: raciMatrix,
                    updated_by: updatedBy
                })
            });

            if (!response.ok) {
                throw new Error(`Failed to update RACI matrix: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            throw error;
        }
    }

    // Convenience methods for backward compatibility

    /**
     * Get just the introduction (from specification)
     */
    async getIntroduction() {
        const spec = await this.getSpecification();
        return { introduction: spec.introduction };
    }

    /**
     * Get just the goals (from specification)
     */
    async getGoals() {
        const spec = await this.getSpecification();
        return spec.goals || [];
    }

    /**
     * Get just the work items (from specification)
     */
    async getWorkItems() {
        const spec = await this.getSpecification();
        return spec.work_items || [];
    }

    /**
     * Get just the roles (from specification)
     */
    async getRoles() {
        const spec = await this.getSpecification();
        return spec.roles || [];
    }

    /**
     * Get just the workflows (from specification)
     */
    async getWorkflows() {
        const spec = await this.getSpecification();
        return spec.workflows || [];
    }

    /**
     * Get just the RACI matrix (from specification)
     */
    async getRACIMatrix() {
        const spec = await this.getSpecification();
        return spec.raci_matrix;
    }

    /**
     * Add a new goal to the specification
     */
    async addGoal(goal, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const updatedGoals = [...(spec.goals || []), goal];
        return await this.updateGoals(updatedGoals, updatedBy);
    }

    /**
     * Update a specific goal in the specification
     */
    async updateGoal(goalKey, updatedGoal, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const goals = spec.goals || [];
        const goalIndex = goals.findIndex(g => g._key === goalKey);

        if (goalIndex === -1) {
            throw new Error(`Goal with key ${goalKey} not found`);
        }

        goals[goalIndex] = { ...goals[goalIndex], ...updatedGoal };
        return await this.updateGoals(goals, updatedBy);
    }

    /**
     * Remove a goal from the specification
     */
    async deleteGoal(goalKey, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const goals = spec.goals || [];
        const filteredGoals = goals.filter(g => g._key !== goalKey);
        return await this.updateGoals(filteredGoals, updatedBy);
    }

    /**
     * Add a new work item to the specification
     */
    async addWorkItem(workItem, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const updatedWorkItems = [...(spec.work_items || []), workItem];
        return await this.updateWorkItems(updatedWorkItems, updatedBy);
    }

    /**
     * Update a specific work item in the specification
     */
    async updateWorkItem(workItemKey, updatedWorkItem, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const workItems = spec.work_items || [];
        const workItemIndex = workItems.findIndex(wi => wi._key === workItemKey);

        if (workItemIndex === -1) {
            throw new Error(`Work item with key ${workItemKey} not found`);
        }

        workItems[workItemIndex] = { ...workItems[workItemIndex], ...updatedWorkItem };
        return await this.updateWorkItems(workItems, updatedBy);
    }

    /**
     * Remove a work item from the specification
     */
    async deleteWorkItem(workItemKey, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const workItems = spec.work_items || [];
        const filteredWorkItems = workItems.filter(wi => wi._key !== workItemKey);
        return await this.updateWorkItems(filteredWorkItems, updatedBy);
    }

    /**
     * Add a new workflow to the specification
     */
    async addWorkflow(workflow, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const updatedWorkflows = [...(spec.workflows || []), workflow];
        return await this.updateWorkflows(updatedWorkflows, updatedBy);
    }

    /**
     * Update a specific workflow in the specification
     */
    async updateWorkflow(workflowKey, updatedWorkflow, updatedBy = 'user') {
        console.log('🔧 updateWorkflow called:', { workflowKey, updatedWorkflow });

        const spec = await this.getSpecification();
        const workflows = spec.workflows || [];

        console.log('📋 Current workflows:', workflows.map(w => ({
            key: w.key,
            _key: w._key,
            id: w.id,
            name: w.name
        })));

        const workflowIndex = workflows.findIndex(w => {
            // Match the pattern used for goals - check _key first
            const match = w._key === workflowKey || w.key === workflowKey || w.id === workflowKey;
            console.log(`  Checking workflow: _key=${w._key}, key=${w.key}, id=${w.id} against ${workflowKey}: ${match}`);
            return match;
        });

        console.log('🔍 Found workflow at index:', workflowIndex);

        if (workflowIndex === -1) {
            console.error('❌ Workflow not found:', workflowKey);
            throw new Error(`Workflow with key ${workflowKey} not found`);
        }

        const originalWorkflow = workflows[workflowIndex];
        console.log('📝 Original workflow:', originalWorkflow);

        // Match goals pattern: simple merge preserving original fields
        workflows[workflowIndex] = {
            ...originalWorkflow,
            ...updatedWorkflow
        };

        console.log('✅ Updated workflow:', workflows[workflowIndex]);

        return await this.updateWorkflows(workflows, updatedBy);
    }

    /**
     * Remove a workflow from the specification
     */
    async deleteWorkflow(workflowKey, updatedBy = 'user') {
        const spec = await this.getSpecification();
        const workflows = spec.workflows || [];
        const filteredWorkflows = workflows.filter(w => {
            const id = w.key || w._key || w.id;
            return id !== workflowKey;
        });
        return await this.updateWorkflows(filteredWorkflows, updatedBy);
    }
}

// Create a singleton instance for easy access
window.specificationAPI = new window.SpecificationAPI();

// Backward compatibility functions for global access
window.getOverview = async function () {
    return await window.specificationAPI.getIntroduction();
};

window.updateOverview = async function (introduction) {
    return await window.specificationAPI.updateIntroduction(introduction);
};

window.getGoals = async function () {
    return await window.specificationAPI.getGoals();
};

window.updateGoals = async function (goals) {
    return await window.specificationAPI.updateGoals(goals);
};

window.getWorkItems = async function () {
    return await window.specificationAPI.getWorkItems();
};

window.updateWorkItems = async function (workItems) {
    return await window.specificationAPI.updateWorkItems(workItems);
};