// Chat streaming functionality
// Handles chat form submission with SSE streaming support

/**
 * Handle chat form submission with streaming support
 * @param {Event} event - Form submit event
 * @returns {boolean} false to prevent default form submission
 */
window.handleChatSubmit = async function (event) {
    event.preventDefault();
    
    const form = event.target;
    const agencyID = form.dataset.agencyId;
    const hasExistingConversation = form.dataset.hasConversation === 'true';

    console.log('💬 [CHAT] Form submitted', { agencyID, hasExistingConversation });

    const messageInput = document.getElementById('user-input');
    const chatMessages = document.getElementById('chat-messages');
    const submitBtn = document.getElementById('chat-submit-btn');

    if (!messageInput || !chatMessages) {
        console.error('Required elements not found');
        return false;
    }

    const originalMessage = messageInput.value.trim();
    if (!originalMessage) {
        return false;
    }

    // Get current context
    const context = window.currentAgencyContext || 'introduction';
    console.log('  📍 Context:', context);

    // Build form data
    const formData = new URLSearchParams();
    formData.append('message', originalMessage);
    formData.append('context', context);

    // Include editor content based on context
    if (context === 'introduction') {
        const editor = document.getElementById('introduction-editor');
        if (editor && editor.value) {
            formData.append('introduction-editor', editor.value);
            console.log('  📝 Included introduction editor content');
        }
    }

    // Append formatted contexts if available
    let fullMessage = originalMessage;
    if (window.ContextManager) {
        const formattedContexts = window.ContextManager.getFormattedContexts();
        if (formattedContexts) {
            fullMessage = originalMessage + formattedContexts;
            formData.set('message', fullMessage);
            console.log('  🔗 Added context selections');
        }
    }

    // Add user message to chat immediately
    addUserMessageToChat(originalMessage, chatMessages);

    // Clear input and disable button
    messageInput.value = '';
    if (submitBtn) {
        submitBtn.classList.add('is-loading');
        submitBtn.disabled = true;
    }

    // Show processing indicator
    if (window.showAIProcessStatus) {
        window.showAIProcessStatus('AI is processing your message...');
    }

    try {
        // Determine endpoint based on conversation state
        let endpoint;
        if (hasExistingConversation) {
            // Get conversation ID from data attribute or URL
            const conversationID = chatMessages.dataset.conversationId || '';
            endpoint = `/api/v1/conversations/${conversationID}/messages/web`;
        } else {
            endpoint = `/api/v1/agencies/${agencyID}/designer/conversations/web`;
        }

        // Check if streaming is enabled
        const useStreaming = window.isStreamingEnabled && window.isStreamingEnabled();
        console.log('  🌊 Streaming enabled:', useStreaming);

        if (useStreaming && context === 'introduction') {
            // Use streaming for introduction refinement
            await handleStreamingChatResponse(endpoint, formData, chatMessages);
        } else {
            // Use non-streaming for other contexts or when disabled
            await handleNonStreamingChatResponse(endpoint, formData, chatMessages, hasExistingConversation);
        }

        // Clear context selections
        if (window.ContextManager) {
            window.ContextManager.clearSelections();
        }

        // Scroll to bottom
        chatMessages.scrollTop = chatMessages.scrollHeight;

    } catch (error) {
        console.error('❌ Chat submission error:', error);
        addErrorMessageToChat('Failed to send message. Please try again.', chatMessages);
    } finally {
        // Re-enable button
        if (submitBtn) {
            submitBtn.classList.remove('is-loading');
            submitBtn.disabled = false;
        }
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }
    }

    return false;
};

/**
 * Handle streaming chat response using SSE
 */
async function handleStreamingChatResponse(endpoint, formData, chatMessages) {
    console.log('  📡 Using STREAMING mode');

    // Add streaming query parameter
    const streamEndpoint = `${endpoint}?stream=true`;

    // Create AI message container for streaming
    const aiMessageDiv = createAIMessageContainer(chatMessages);
    const messageBubble = aiMessageDiv.querySelector('.message-bubble');

    try {
        // Use shared streaming utility
        await window.executeAIStream({
            url: streamEndpoint,
            formData: formData,
            displayElement: messageBubble,
            onComplete: (result) => {
                console.log('  ✅ Streaming complete', result);
                // Update timestamp
                const timeDiv = aiMessageDiv.querySelector('.message-time');
                if (timeDiv) {
                    timeDiv.textContent = new Date().toLocaleTimeString('en-US', {
                        hour: 'numeric',
                        minute: '2-digit'
                    });
                }
            },
            onError: (error) => {
                console.error('  ❌ Streaming error:', error);
                messageBubble.innerHTML = '<p class="has-text-danger">⚠️ Error processing message</p>';
            }
        });
    } catch (error) {
        console.error('Streaming failed:', error);
        throw error;
    }
}

/**
 * Handle non-streaming chat response
 */
async function handleNonStreamingChatResponse(endpoint, formData, chatMessages, hasExistingConversation) {
    console.log('  📄 Using NON-STREAMING mode');

    const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: formData
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    const html = await response.text();
    console.log('  📥 Received HTML response');

    // For new conversations, replace entire chat
    // For existing conversations, append new messages
    if (!hasExistingConversation) {
        chatMessages.innerHTML = html;
    } else {
        // Extract and append only new messages
        const temp = document.createElement('div');
        temp.innerHTML = html;
        const newMessages = temp.querySelectorAll('.message');
        newMessages.forEach(msg => chatMessages.appendChild(msg));
    }
}

/**
 * Add user message to chat UI
 */
function addUserMessageToChat(message, container) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message user-message';

    const now = new Date();
    const timeStr = now.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit'
    });

    messageDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                <p>${escapeHtml(message)}</p>
            </div>
            <div class="message-time">${timeStr}</div>
        </div>
    `;

    container.appendChild(messageDiv);
    console.log('  💬 Added user message to chat');
}

/**
 * Create AI message container for streaming
 */
function createAIMessageContainer(container) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ai-message';

    const now = new Date();
    const timeStr = now.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit'
    });

    messageDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                <p class="has-text-grey-light">
                    <span class="icon"><i class="fas fa-spinner fa-pulse"></i></span>
                    Thinking...
                </p>
            </div>
            <div class="message-time">${timeStr}</div>
        </div>
    `;

    container.appendChild(messageDiv);
    return messageDiv;
}

/**
 * Add error message to chat
 */
function addErrorMessageToChat(errorMessage, container) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ai-message';

    const now = new Date();
    const timeStr = now.toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit'
    });

    messageDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                <p class="has-text-danger">
                    <span class="icon"><i class="fas fa-exclamation-triangle"></i></span>
                    ${escapeHtml(errorMessage)}
                </p>
            </div>
            <div class="message-time">${timeStr}</div>
        </div>
    `;

    container.appendChild(messageDiv);
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

console.log('✅ Chat streaming utilities loaded');
