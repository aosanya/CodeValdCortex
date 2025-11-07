// Context Manager
// Handles text selection and context building for AI conversations

import { showNotification } from './utils.js';

/**
 * @typedef {Object} ContextMetadata
 * @property {number} [selectionCount] - Number of selections combined
 * @property {Selection[]} [selections] - Array of individual selections
 * @property {boolean} [isNavigational] - If true, context is automatically removed when navigating away
 * @property {string} [section] - The section this context belongs to (introduction, work-items, goals, etc.)
 */

/**
 * @typedef {Object} Context
 * @property {number} id - Unique context ID
 * @property {string} type - Context type (Introduction, Goal Definition, etc.)
 * @property {string} code - Item code (G001, INTRO, etc.)
 * @property {string} content - The selected text content
 * @property {string} timestamp - ISO timestamp of creation
 * @property {ContextMetadata} metadata - Additional metadata
 */

/**
 * @typedef {Object} Selection
 * @property {string} text - Selected text
 * @property {string} type - Context type
 * @property {string} code - Item code
 * @property {string} timestamp - ISO timestamp
 */

/**
 * @typedef {Object} ContextState
 * @property {Context[]} contexts - Array of context objects
 * @property {number} nextId - Next available context ID
 * @property {Selection[]} selections - Array of selections being accumulated
 */

/** @type {ContextState} */
let contextState = {
    contexts: [], // Array of context objects
    nextId: 1,
    selections: [] // Array of selections being accumulated
};

// Make arrays globally available
if (typeof window !== 'undefined') {
    window.contextSelections = contextState.contexts;
    window.selections = contextState.selections;
}

/**
 * Log context state changes
 */
function logContextStateChange(action, context = null) {
    // Update global reference
    if (typeof window !== 'undefined') {
        window.contextSelections = [...contextState.contexts];
    }
}

// Context types
export const ContextType = {
    INTRODUCTION: 'Introduction',
    GOAL: 'Goal Definition',
    WORK_ITEM: 'Work Item',
    UNIT_OF_WORK: 'Unit of Work',
    AGENT_TYPE: 'Agent Type',
    GENERIC: 'Generic'
};

/**
 * Create a new context object from selection
 * @param {string} type - Context type (from ContextType enum)
 * @param {string} code - Item code (e.g., G001, U001)
 * @param {string} content - Selected text content
 * @param {ContextMetadata} [metadata={}] - Additional metadata (optional)
 * @returns {Context} Context object
 */
export function createContext(type, code, content, metadata = {}) {
    /** @type {Context} */
    const context = {
        id: contextState.nextId++,
        type: type,
        code: code,
        content: content.trim(),
        timestamp: new Date().toISOString(),
        metadata: metadata
    };

    contextState.contexts.push(context);
    logContextStateChange('ADDED', context);

    updateContextDisplay();
    showNotification(`Context added: ${type} ${code}`, 'success');

    return context;
}

/**
 * Add context from selected text in goal item
 * @param {string} goalCode - Goal code
 * @param {string} goalDescription - Full goal description
 */
export function addGoalContext(goalCode, goalDescription) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    // If text is selected, use it; otherwise use full description
    const content = selectedText || goalDescription;

    // Check if context already exists
    const exists = contextState.contexts.some(ctx =>
        ctx.type === ContextType.GOAL && ctx.code === goalCode && ctx.content === content
    );

    if (exists) {
        showNotification('This context is already added', 'warning');
        return null;
    }

    return createContext(ContextType.GOAL, goalCode, content);
}

/**
 * Add context from work item
 * @param {string} workItemCode - Work item code (e.g., WI-001)
 * @param {string} workItemDescription - Full work item description
 * @param {boolean} [isNavigational=true] - If true, context is removed when navigating away from work items
 */
export function addWorkItemContext(workItemCode, workItemDescription, isNavigational = true) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    // If text is selected, use it; otherwise use full description
    const content = selectedText || workItemDescription;

    // Check if context already exists
    const exists = contextState.contexts.some(ctx =>
        ctx.type === ContextType.WORK_ITEM && ctx.code === workItemCode && ctx.content === content
    );

    if (exists) {
        showNotification('This context is already added', 'warning');
        return null;
    }

    const result = createContext(ContextType.WORK_ITEM, workItemCode, content, {
        isNavigational: isNavigational,
        section: 'work-items'
    });
    return result;
}

/**
 * Add context from introduction
 * @param {string} introText - Introduction text (full or selected)
 */
export function addIntroductionContext(introText) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    const content = selectedText || introText;

    // Check if context already exists
    const exists = contextState.contexts.some(ctx =>
        ctx.type === ContextType.INTRODUCTION && ctx.content === content
    );

    if (exists) {
        showNotification('This context is already added', 'warning');
        return null;
    }

    return createContext(ContextType.INTRODUCTION, 'INTRO', content);
}

/**
 * Add context from unit of work
 * @param {string} unitCode - Unit code
 * @param {string} unitDescription - Full unit description
 */
export function addUnitContext(unitCode, unitDescription) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    const content = selectedText || unitDescription;

    const exists = contextState.contexts.some(ctx =>
        ctx.type === ContextType.UNIT_OF_WORK && ctx.code === unitCode && ctx.content === content
    );

    if (exists) {
        showNotification('This context is already added', 'warning');
        return null;
    }

    return createContext(ContextType.UNIT_OF_WORK, unitCode, content);
}

/**
 * Add context from agent type
 * @param {string} agentCode - Agent type code
 * @param {string} agentDescription - Full agent description
 */
export function addAgentContext(agentCode, agentDescription) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    const content = selectedText || agentDescription;

    const exists = contextState.contexts.some(ctx =>
        ctx.type === ContextType.AGENT_TYPE && ctx.code === agentCode && ctx.content === content
    );

    if (exists) {
        showNotification('This context is already added', 'warning');
        return null;
    }

    return createContext(ContextType.AGENT_TYPE, agentCode, content);
}

/**
 * Remove context by ID
 * @param {number} contextId - Context ID to remove
 */
export function removeContext(contextId) {
    const index = contextState.contexts.findIndex(ctx => ctx.id === contextId);
    if (index !== -1) {
        const removed = contextState.contexts.splice(index, 1)[0];
        logContextStateChange('REMOVED', removed);
        updateContextDisplay();
        showNotification(`Context removed: ${removed.type} ${removed.code}`, 'info');
    }
}

/**
 * Clear navigational contexts for a specific section
 * @param {string} [section] - Section to clear (e.g., 'work-items'). If not provided, clears all navigational contexts.
 */
export function clearNavigationalContexts(section = null) {
    const beforeCount = contextState.contexts.length;

    if (section) {
        // Clear only contexts for specific section
        contextState.contexts = contextState.contexts.filter(ctx =>
            !ctx.metadata?.isNavigational || ctx.metadata?.section !== section
        );
    } else {
        // Clear all navigational contexts
        contextState.contexts = contextState.contexts.filter(ctx => !ctx.metadata?.isNavigational);
    }

    const removedCount = beforeCount - contextState.contexts.length;

    if (removedCount > 0) {
        logContextStateChange('CLEARED_NAVIGATIONAL');
        updateContextDisplay();
        console.log(`[ContextManager] Cleared ${removedCount} navigational context(s)${section ? ` from section: ${section}` : ''}`);
    }
}

/**
 * Remove selection by index
 * @param {number} index - Selection index to remove
 */
export function removeSelection(index) {
    if (index >= 0 && index < contextState.selections.length) {
        const removed = contextState.selections.splice(index, 1)[0];
        updateContextDisplay();
        showNotification(`Selection removed: ${removed.type} ${removed.code}`, 'info');
    }
}

/**
 * Clear all contexts
 */
export function clearAllContexts() {
    const previousCount = contextState.contexts.length;
    contextState.contexts = [];
    contextState.selections = []; // Also clear selections
    contextState.nextId = 1;
    logContextStateChange('CLEARED_ALL');
    updateContextDisplay();
    showNotification('All contexts cleared', 'info');
}

/**
 * Get all contexts
 * @returns {Array} Array of context objects
 */
export function getAllContexts() {
    return contextState.contexts;
}

/**
 * Get contexts formatted for API/Chat
 * @returns {string} Formatted context string for AI
 */
export function getFormattedContexts() {
    // Include both finalized contexts and pending selections
    const hasContexts = contextState.contexts.length > 0;
    const hasSelections = contextState.selections.length > 0;

    console.log('[ContextManager] getFormattedContexts called:', {
        contexts: contextState.contexts.length,
        selections: contextState.selections.length
    });

    // If no explicit contexts/selections, return empty string
    // (Context is already being appended by the backend/form handler)
    if (!hasContexts && !hasSelections) {
        console.log('[ContextManager] No contexts or selections, returning empty');
        return '';
    }

    let formatted = '\n\n**Context:**\n';

    // Add finalized contexts
    contextState.contexts.forEach((ctx, index) => {
        formatted += `\n${index + 1}. **${ctx.type}** [${ctx.code}]:\n`;
        formatted += `   ${ctx.content}\n`;
    });

    // Add pending selections (these are the yellow boxes)
    contextState.selections.forEach((sel, index) => {
        const num = contextState.contexts.length + index + 1;
        formatted += `\n${num}. **${sel.type}** [${sel.code}]:\n`;
        formatted += `   ${sel.text}\n`;
    });

    console.log('[ContextManager] Formatted context:', formatted.substring(0, 200));
    return formatted;
}

/**
 * Update context display in UI
 */
function updateContextDisplay() {
    const contextContainer = document.getElementById('context-container');
    if (!contextContainer) return;

    // Get the clear button in the header
    const clearButton = document.querySelector('.context-panel .button[onclick*="clearAllContexts"]');

    if (contextState.contexts.length === 0 && contextState.selections.length === 0) {
        // Empty state - clear container and hide clear button
        contextContainer.innerHTML = '';
        if (clearButton) {
            clearButton.style.display = 'none';
        }
        return;
    }

    // Show clear button when there's content
    if (clearButton) {
        clearButton.style.display = '';
    }

    let html = '';

    // Render pending selections first (if any)
    if (contextState.selections.length > 0) {
        html += `<div class="">`;

        contextState.selections.forEach((sel, index) => {
            const typeColor = getContextTypeColor(sel.type);
            const truncatedText = sel.text.length > 80
                ? sel.text.substring(0, 80) + '...'
                : sel.text;

            html += `
                <div class="box p-2 mb-2 has-background-warning-light" data-selection-index="${index}">
                    <div class="is-flex is-justify-content-space-between is-align-items-center">
                        <div class="is-flex is-align-items-center" style="gap: 0.5rem; flex: 1; min-width: 0;">
                            <span class="tag is-small ${typeColor}" style="flex-shrink: 0;">${sel.type}</span>
                            <span class="is-size-7" style="word-break: break-word; overflow: hidden;">${escapeHtml(truncatedText)}</span>
                        </div>
                        <button 
                            class="delete is-small" 
                            onclick="window.ContextManager.removeSelection(${index})"
                            title="Remove selection"
                            style="flex-shrink: 0;"
                        ></button>
                    </div>
                </div>
            `;
        });

        html += `</div>`;
    }

    // Render finalized contexts (if any)
    if (contextState.contexts.length > 0) {
        if (contextState.selections.length > 0) {
            html += `
                <hr class="my-3">
                <p class="has-text-weight-semibold is-size-7 has-text-grey-dark mb-2">
                    <span class="icon is-small">
                        <i class="fas fa-layer-group"></i>
                    </span>
                    <span>Finalized Contexts (${contextState.contexts.length})</span>
                </p>
            `;
        }

        contextState.contexts.forEach(ctx => {
            const typeColor = getContextTypeColor(ctx.type);
            const truncatedContent = ctx.content.length > 100
                ? ctx.content.substring(0, 100) + '...'
                : ctx.content;

            html += `
                <div class="context-item box p-3 mb-2" data-context-id="${ctx.id}">
                    <div class="level is-mobile">
                        <div class="level-left">
                            <div class="level-item">
                                <span class="tag ${typeColor}">${ctx.type}</span>
                            </div>
                        </div>
                        <div class="level-right">
                            <div class="level-item">
                                <button 
                                    class="delete is-small" 
                                    onclick="window.AgencyDesigner.removeContext(${ctx.id})"
                                    title="Remove context"
                                ></button>
                            </div>
                        </div>
                    </div>
                    <div class="content is-small mt-2">
                        <p class="context-content">${escapeHtml(truncatedContent)}</p>
                    </div>
                </div>
            `;
        });
    }

    contextContainer.innerHTML = html;
}

/**
 * Get Bulma color class for context type
 * @param {string} type - Context type
 * @returns {string} Bulma color class
 */
function getContextTypeColor(type) {
    switch (type) {
        case ContextType.INTRODUCTION:
            return 'is-info';
        case ContextType.GOAL:
            return 'is-primary';
        case ContextType.WORK_ITEM:
            return 'is-warning';
        case ContextType.UNIT_OF_WORK:
            return 'is-link';
        case ContextType.AGENT_TYPE:
            return 'is-success';
        default:
            return 'is-light';
    }
}

/**
 * Escape HTML to prevent XSS
 * @param {string} text - Text to escape
 * @returns {string} Escaped text
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Initialize context selection listeners
 */
export function initializeContextSelection() {
    // Listen for text selection (mouseup event)
    document.addEventListener('mouseup', handleTextSelection);

    // Hide context menu when clicking elsewhere
    document.addEventListener('click', function (event) {
        if (!event.target.closest('.context-menu') && !event.target.closest('.context-add-button')) {
            hideContextMenu();
        }
    });
}/**
 * Handle text selection
 * @param {Event} event - Mouse up event
 */
function handleTextSelection(event) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    // Only proceed if there's actual text selected
    if (selectedText.length === 0) {
        hideContextMenu();
        return;
    }

    // Check if selection is within a context-selectable area
    const target = event.target;
    const goalItem = target.closest('[data-goal-code]');
    const unitItem = target.closest('[data-unit-code]');
    const agentItem = target.closest('[data-agent-code]');
    const introCard = target.closest('[data-intro-text]');

    if (goalItem || unitItem || agentItem || introCard) {
        // Get the selection range and position
        const range = selection.getRangeAt(0);
        const rect = range.getBoundingClientRect();

        // Determine context type and code
        let contextType, code, fullText;
        if (goalItem) {
            contextType = ContextType.GOAL;
            code = goalItem.getAttribute('data-goal-code');
            fullText = goalItem.getAttribute('data-goal-description');
        } else if (unitItem) {
            contextType = ContextType.UNIT_OF_WORK;
            code = unitItem.getAttribute('data-unit-code');
            fullText = unitItem.getAttribute('data-unit-description');
        } else if (agentItem) {
            contextType = ContextType.AGENT_TYPE;
            code = agentItem.getAttribute('data-agent-code');
            fullText = agentItem.getAttribute('data-agent-description');
        } else if (introCard) {
            contextType = ContextType.INTRODUCTION;
            code = 'INTRO';
            fullText = introCard.getAttribute('data-intro-text');
        }

        // Automatically add selection to pending array
        contextState.selections.push({
            text: selectedText,
            type: contextType,
            code: code,
            timestamp: new Date().toISOString()
        });

        // Update display to show the new selection
        updateContextDisplay();

        // Show context menu near the selection
        showContextMenu(rect, selectedText, contextType, code);
    } else {
        hideContextMenu();
    }
}
/**
 * Show context menu to add selection as context
 * @param {DOMRect} rect - Selection bounding rect
 * @param {string} selectedText - Selected text
 * @param {string} contextType - Type of context
 * @param {string} code - Context code
 */
function showContextMenu(rect, selectedText, contextType, code) {
    // Remove existing menu if any
    hideContextMenu();

    // Create context menu with single button
    const menu = document.createElement('div');
    menu.className = 'context-menu';

    const pendingCount = contextState.selections.length;

    menu.innerHTML = `
        <button class="button is-small is-info create-context-button" title="Create context with ${pendingCount} accumulated selection${pendingCount !== 1 ? 's' : ''}">
            <span class="icon is-small">
                <i class="fas fa-layer-group"></i>
            </span>
            <span>Create Context (${pendingCount})</span>
        </button>
    `;

    // Position the menu near the selection
    menu.style.position = 'fixed';
    menu.style.top = `${rect.bottom + window.scrollY + 5}px`;
    menu.style.left = `${rect.left + window.scrollX}px`;
    menu.style.zIndex = '9999';

    // Create context button - creates context with all selections
    menu.querySelector('.create-context-button').addEventListener('click', function (e) {
        e.preventDefault();
        e.stopPropagation();

        // Combine all selections
        const combinedText = contextState.selections.map(s => s.text).join('\n\n---\n\n');

        // Create context with combined text (use first selection's type and code)
        const firstSelection = contextState.selections[0];
        createContext(firstSelection.type, firstSelection.code, combinedText, {
            selectionCount: contextState.selections.length,
            selections: contextState.selections
        });

        // Clear pending selections
        contextState.selections = [];
        if (typeof window !== 'undefined') {
            window.selections = contextState.selections;
        }

        // Update display
        updateContextDisplay();

        // Hide menu
        hideContextMenu();
    });

    // Add to document
    document.body.appendChild(menu);
}

/**
 * Hide context menu
 */
function hideContextMenu() {
    const existingMenu = document.querySelector('.context-menu');
    if (existingMenu) {
        existingMenu.remove();
    }
}

// Export for global access
if (typeof window !== 'undefined') {
    // Primary export as ContextManager
    window.ContextManager = {
        createContext,
        addGoalContext,
        addWorkItemContext,
        addIntroductionContext,
        addUnitContext,
        addAgentContext,
        removeContext,
        removeSelection,
        clearAllContexts,
        clearNavigationalContexts,
        getAllContexts,
        getFormattedContexts,
        initializeContextSelection,
        updateDisplay: updateContextDisplay,
        ContextType,
        getSelections: () => contextState.selections,
        clearSelections: () => {
            contextState.selections = [];
            updateContextDisplay();
        }
    };

    // Also export as AgencyDesigner for backward compatibility
    if (!window.AgencyDesigner) {
        window.AgencyDesigner = {};
    }
    window.AgencyDesigner.removeContext = removeContext;
    window.AgencyDesigner.clearAllContexts = clearAllContexts;

    // Export direct array references
    window.contextSelections = contextState.contexts;
    window.selections = contextState.selections;
}
