// Generic AI Operations Handler
// Routes AI operations to entity-specific handlers and renders operation buttons

/**
 * Entity configuration for AI operations
 */
const entityConfigs = {
    'goal': {
        plural: 'goals',
        displayName: 'goal',
        sourceContext: 'introduction'
    },
    'work-item': {
        plural: 'work-items',
        displayName: 'work item',
        sourceContext: 'goals'
    },
    'role': {
        plural: 'roles',
        displayName: 'role',
        sourceContext: 'work items'
    },
    'workflow': {
        plural: 'workflows',
        displayName: 'workflow',
        sourceContext: 'work items'
    }
};

/**
 * Render AI operation buttons for an entity type
 * @param {string} entityType - The entity type (goal, work-item, role, workflow)
 */
function renderAIOperationButtons(entityType) {
    const container = document.getElementById(`${entityType}-ai-buttons`);
    if (!container) return;

    const config = entityConfigs[entityType];
    if (!config) return;

    // Generate button HTML
    const buttons = [
        {
            id: `ai-create-${config.plural}-btn`,
            label: 'Create',
            icon: 'fas fa-sparkles',
            buttonClass: 'is-info',
            onclick: `processAIOperation('${entityType}', ['create'])`,
            title: `Generate new ${config.plural} from ${config.sourceContext}`,
            disabled: false
        },
        {
            id: `ai-enhance-${config.plural}-btn`,
            label: 'Enhance',
            icon: 'fas fa-wand-magic-sparkles',
            buttonClass: 'is-link is-static',
            onclick: `processAIOperation('${entityType}', ['enhance'])`,
            title: `Select ${config.plural} to enhance`,
            disabled: true
        },
        {
            id: `ai-consolidate-${config.plural}-btn`,
            label: 'Consolidate',
            icon: 'fas fa-compress',
            buttonClass: 'is-warning is-static',
            onclick: `processAIOperation('${entityType}', ['consolidate'])`,
            title: `Select ${config.plural} to consolidate`,
            disabled: true
        }
    ];

    // Render buttons
    container.innerHTML = buttons.map(btn => `
        <button 
            class="button is-small ${btn.buttonClass}"
            id="${btn.id}"
            onclick="${btn.onclick}"
            title="${btn.title}"
            ${btn.disabled ? 'disabled' : ''}>
            <span class="icon"><i class="${btn.icon}"></i></span>
            <span>${btn.label}</span>
        </button>
    `).join('');
}

/**
 * Initialize AI operation buttons for all entity types on page load
 */
function initializeAIOperationButtons() {
    Object.keys(entityConfigs).forEach(entityType => {
        renderAIOperationButtons(entityType);
    });
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeAIOperationButtons);
} else {
    initializeAIOperationButtons();
}

/**
 * Generic AI operation processor that routes to entity-specific handlers
 * @param {string} entityType - The entity type (goal, work-item, role, workflow)
 * @param {string[]} operations - Array of operations to perform (create, enhance, consolidate)
 */
window.processAIOperation = function (entityType, operations) {
    // Convert entity type to function name format
    const functionMap = {
        'goal': 'processAIGoalOperation',
        'work-item': 'processAIWorkItemOperation',
        'role': 'processAIRoleOperation',
        'workflow': 'processAIWorkflowOperation'
    };

    const handlerFunction = functionMap[entityType];

    if (!handlerFunction) {
        console.error(`Unknown entity type: ${entityType}`);
        window.showNotification(`Unknown entity type: ${entityType}`, 'error');
        return;
    }

    // Check if the handler function exists
    if (typeof window[handlerFunction] !== 'function') {
        console.error(`Handler function not found: ${handlerFunction}`);
        window.showNotification(`Handler function not found for ${entityType}`, 'error');
        return;
    }

    // Call the entity-specific handler
    window[handlerFunction](operations);
};
