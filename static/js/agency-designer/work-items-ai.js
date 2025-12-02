// Work Items AI Operations
// Handles AI-powered work item operations and deliverable refinement

/**
 * Process AI work item operations (create, enhance, consolidate)
 */
window.processAIWorkItemOperation = async function (operations) {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    // Validate operations array
    if (!operations || operations.length === 0) {
        window.showNotification('Error: No operation specified', 'error');
        return;
    }

    // For enhance/consolidate operations, get selected work items
    let selectedWorkItemKeys = [];
    if (operations.includes('enhance') || operations.includes('consolidate')) {
        selectedWorkItemKeys = window.getSelectedWorkItemKeys();
        if (selectedWorkItemKeys.length === 0) {
            window.showNotification('Please select work items first', 'warning');
            return;
        }
    }

    let statusMessage = 'AI is processing your request...';
    if (operations.length === 1) {
        switch (operations[0]) {
            case 'create':
                statusMessage = 'AI is generating work items from your goals...';
                break;
            case 'enhance':
                statusMessage = `AI is enhancing ${selectedWorkItemKeys.length} work item(s)...`;
                break;
            case 'consolidate':
                statusMessage = `AI is consolidating ${selectedWorkItemKeys.length} work item(s)...`;
                break;
        }
    } else if (operations.length > 1) {
        statusMessage = `AI is performing ${operations.length} operations on your work items...`;
    }

    // Show AI processing status in the chat area
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus(statusMessage);
    }

    try {
        const requestBody = { operations };

        // Include selected work item keys if applicable
        if (selectedWorkItemKeys.length > 0) {
            requestBody.work_item_keys = selectedWorkItemKeys;
        }

        const response = await fetch(`/api/v1/agencies/${agencyId}/work-items/ai-process`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            throw new Error(`Failed to process AI work item operations: ${response.statusText}`);
        }

        const data = await response.json();

        // Update status to show we're processing results
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('Processing results and updating work items...');
        }

        // Reload work items to show updates
        await window.loadWorkItems();

        // After work items are reloaded, refresh chat messages so AI explanation appears in the chat
        try {
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                const chatResp = await fetch(`/agencies/${agencyId}/chat-messages`);
                if (chatResp.ok) {
                    const chatHtml = await chatResp.text();
                    chatContainer.innerHTML = chatHtml;
                    // Scroll to bottom to show latest assistant message
                    try { scrollToBottom(chatContainer); } catch (e) { /* ignore */ }
                }
            }
        } catch (err) {
        }

        // Hide AI processing status after work items and chat are updated
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        // Show success message with results
        if (data.created_count > 0) {
            window.showNotification(`Successfully created ${data.created_count} work items!`, 'success');
        } else if (data.enhanced_count > 0) {
            window.showNotification(`Successfully enhanced ${data.enhanced_count} work items!`, 'success');
        } else if (data.results && data.results.consolidate_success) {
            window.showNotification(data.results.consolidate_success, 'success');
        } else {
            window.showNotification('AI operations completed!', 'success');
        }

    } catch (error) {
        // Hide AI processing status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        window.showNotification(`AI processing failed: ${error.message}`, 'danger');
    }
}

/**
 * AI Refinement for Deliverables
 * Generates or refines deliverable structure based on work item context
 */
window.refineDeliverablesWithAI = async function () {
    // Get work item context
    const title = document.getElementById('work-item-title-editor')?.value.trim();
    const description = document.getElementById('work-item-description-editor')?.value.trim();
    const code = document.getElementById('work-item-code-editor')?.value.trim();

    if (!title || !description) {
        window.showNotification('Please enter work item title and description first', 'warning');
        return;
    }

    // Build context for AI
    let contextMessage = `Generate a structured deliverable hierarchy for this work item:\n\n`;
    contextMessage += `**Work Item Code**: ${code}\n`;
    contextMessage += `**Title**: ${title}\n`;
    contextMessage += `**Description**: ${description}\n\n`;
    contextMessage += `Please analyze the work item and suggest a complete deliverable structure (folders and files) that would be appropriate for this work. `;
    contextMessage += `Consider documentation requirements, code artifacts, configuration files, and any other relevant deliverables. `;
    contextMessage += `Format the response as a hierarchical structure that can be used to build a directory tree.`;

    // Add current deliverables as context if they exist
    const currentDeliverables = window.getDeliverablesStructuredData?.();
    if (currentDeliverables && currentDeliverables.length > 0) {
        contextMessage += `\n\n**Current Deliverables Structure**:\n`;
        contextMessage += JSON.stringify(currentDeliverables, null, 2);
        contextMessage += `\n\nYou can refine or expand this structure as needed.`;
    }

    // Add context to context manager
    if (window.ContextManager) {
        window.ContextManager.createContext(
            window.ContextManager.ContextType.WORK_ITEM,
            code || 'DELIVERABLES',
            contextMessage,
            {
                isNavigational: true,
                section: 'work-items',
                tab: 'deliverables'
            }
        );
    }

    // Show notification to user
    window.showNotification('Context added to AI chat. Please use the chat panel to generate or refine deliverables.', 'info');

    // Optionally, trigger chat focus
    const chatInput = document.getElementById('chat-input-field') || document.querySelector('[data-chat-input]');
    if (chatInput) {
        chatInput.focus();
        chatInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
};
