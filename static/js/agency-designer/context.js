// Context Manager
// Handles text selection and context building for AI conversations

import { showNotification } from './utils.js';

// Global context state - accessible via window.contextSelections
let contextState = {
    contexts: [], // Array of context objects
    nextId: 1
};

// Make contexts globally available
if (typeof window !== 'undefined') {
    window.contextSelections = contextState.contexts;
}

/**
 * Log context state changes
 */
function logContextStateChange(action, context = null) {
    console.log('🔄 Context State Changed:', {
        action: action,
        totalContexts: contextState.contexts.length,
        contexts: contextState.contexts.map(ctx => ({
            id: ctx.id,
            type: ctx.type,
            code: ctx.code,
            contentPreview: ctx.content.substring(0, 50) + '...'
        })),
        newContext: context ? {
            id: context.id,
            type: context.type,
            code: context.code
        } : null
    });

    // Update global reference
    if (typeof window !== 'undefined') {
        window.contextSelections = [...contextState.contexts];
    }
}

// Context types
export const ContextType = {
    INTRODUCTION: 'Introduction',
    GOAL: 'Goal Definition',
    UNIT_OF_WORK: 'Unit of Work',
    AGENT_TYPE: 'Agent Type',
    GENERIC: 'Generic'
};

/**
 * Create a new context object from selection
 * @param {string} type - Context type (from ContextType enum)
 * @param {string} code - Item code (e.g., G001, U001)
 * @param {string} content - Selected text content
 * @param {object} metadata - Additional metadata (optional)
 * @returns {object} Context object
 */
export function createContext(type, code, content, metadata = {}) {
    console.log('🔨 Creating context:', { type, code, contentLength: content.length });

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
 * Clear all contexts
 */
export function clearAllContexts() {
    const previousCount = contextState.contexts.length;
    contextState.contexts = [];
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
    if (contextState.contexts.length === 0) {
        return '';
    }

    let formatted = '\n\n**Context:**\n';
    contextState.contexts.forEach((ctx, index) => {
        formatted += `\n${index + 1}. **${ctx.type}** [${ctx.code}]:\n`;
        formatted += `   ${ctx.content}\n`;
    });

    return formatted;
}

/**
 * Update context display in UI
 */
function updateContextDisplay() {
    const contextContainer = document.getElementById('context-container');
    if (!contextContainer) return;

    if (contextState.contexts.length === 0) {
        contextContainer.innerHTML = `
            <div class="has-text-grey has-text-centered py-3">
                <p><i class="fas fa-info-circle"></i> No contexts selected</p>
                <p class="is-size-7 mt-2">Select text from goals, units, or other items to add context</p>
            </div>
        `;
        return;
    }

    // Render contexts
    let html = '';
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
                        <div class="level-item">
                            <strong class="has-text-weight-semibold">${ctx.code}</strong>
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
    console.log('🚀 Context manager initialized');
    console.log('📋 Initial context state:', {
        contexts: contextState.contexts,
        globalAccess: 'window.contextSelections'
    });

    // Listen for text selection (mouseup event)
    document.addEventListener('mouseup', handleTextSelection);

    // Hide context menu when clicking elsewhere
    document.addEventListener('click', function (event) {
        if (!event.target.closest('.context-menu') && !event.target.closest('.context-add-button')) {
            hideContextMenu();
        }
    });

    console.log('✅ Context selection listeners attached');
}/**
 * Handle text selection
 * @param {Event} event - Mouse up event
 */
function handleTextSelection(event) {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    console.log('🔍 Text selection detected:', {
        selectedText: selectedText,
        length: selectedText.length,
        target: event.target.tagName
    });

    // Only proceed if there's actual text selected
    if (selectedText.length === 0) {
        console.log('⚠️ No text selected, hiding context menu');
        hideContextMenu();
        return;
    }

    // Check if selection is within a context-selectable area
    const target = event.target;
    const goalItem = target.closest('[data-goal-code]');
    const unitItem = target.closest('[data-unit-code]');
    const agentItem = target.closest('[data-agent-code]');
    const introCard = target.closest('[data-intro-text]');

    console.log('📍 Context area detection:', {
        isGoal: !!goalItem,
        isUnit: !!unitItem,
        isAgent: !!agentItem,
        isIntro: !!introCard
    });

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

        console.log('✅ Showing context menu:', {
            type: contextType,
            code: code,
            selectedText: selectedText.substring(0, 50) + '...',
            position: { top: rect.bottom, left: rect.left }
        });

        // Show context menu near the selection
        showContextMenu(rect, selectedText, contextType, code);
    } else {
        console.log('❌ Selection not in a context-selectable area');
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
    console.log('🎯 Creating context menu:', {
        type: contextType,
        code: code,
        textLength: selectedText.length,
        position: { top: rect.bottom, left: rect.left }
    });

    // Remove existing menu if any
    hideContextMenu();

    // Create context menu
    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.innerHTML = `
        <button class="button is-small is-info context-add-button">
            <span class="icon is-small">
                <i class="fas fa-layer-group"></i>
            </span>
            <span>Add to Context</span>
        </button>
    `;

    // Position the menu near the selection
    menu.style.position = 'fixed';
    menu.style.top = `${rect.bottom + window.scrollY + 5}px`;
    menu.style.left = `${rect.left + window.scrollX}px`;
    menu.style.zIndex = '9999';

    // Add click handler
    menu.querySelector('.context-add-button').addEventListener('click', function (e) {
        e.preventDefault();
        e.stopPropagation();

        console.log('➕ Adding context:', {
            type: contextType,
            code: code,
            text: selectedText.substring(0, 50) + '...'
        });

        // Create context based on type
        createContext(contextType, code, selectedText);

        // Clear selection and hide menu
        window.getSelection().removeAllRanges();
        hideContextMenu();

        console.log('✅ Context added and menu hidden');
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
    window.ContextManager = {
        createContext,
        addGoalContext,
        addIntroductionContext,
        addUnitContext,
        addAgentContext,
        removeContext,
        clearAllContexts,
        getAllContexts,
        getFormattedContexts,
        initializeContextSelection,
        ContextType
    };
}
