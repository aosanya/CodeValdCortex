// Chat streaming functionality
// Handles chat form submission with SSE streaming support

// Global abort controller for stopping requests
let currentAbortController = null;

/**
 * Convert send button to stop button
 */
function convertToStopButton() {
    const submitBtn = document.getElementById('chat-submit-btn');
    if (submitBtn) {
        const submitIcon = document.getElementById('chat-submit-icon');
        if (submitIcon) {
            // Change to stop icon
            submitIcon.innerHTML = '<i class="fas fa-stop"></i>';
        }
        submitBtn.classList.remove('is-primary');
        submitBtn.classList.add('is-danger');
        // Don't disable - allow stopping
        submitBtn.onclick = function (e) {
            e.preventDefault();
            console.log('Stop button clicked');
            window.stopChatProcessing();
            return false;
        };
        console.log('Converted to stop button');
    } else {
        console.error('Submit button not found');
    }
}

/**
 * Restore the send button to its default state
 */
function restoreSendButton() {
    const submitBtn = document.getElementById('chat-submit-btn');
    if (submitBtn) {
        const submitIcon = document.getElementById('chat-submit-icon');
        if (submitIcon) {
            submitIcon.innerHTML = '<i class="fas fa-paper-plane"></i>';
        }
        submitBtn.classList.remove('is-danger');
        submitBtn.classList.add('is-primary');
        submitBtn.onclick = null;
    }
}

/**
 * Stop the current chat processing
 */
window.stopChatProcessing = function () {
    console.log('stopChatProcessing called', { hasController: !!currentAbortController });

    if (currentAbortController) {
        console.log('Aborting request...');
        currentAbortController.abort();
        currentAbortController = null;

        // Show cancellation message
        const chatMessages = document.getElementById('chat-messages');
        if (chatMessages) {
            addErrorMessageToChat('Request cancelled by user.', chatMessages);
        }

        // Hide AI status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        // Restore send button
        restoreSendButton();
    } else {
        console.warn('No active abort controller to cancel');
    }
};

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

    const messageInput = document.getElementById('user-input');
    const chatMessages = document.getElementById('chat-messages');
    const submitBtn = document.getElementById('chat-submit-btn');

    if (!messageInput || !chatMessages) {
        return false;
    }

    const originalMessage = messageInput.value.trim();
    if (!originalMessage) {
        return false;
    }

    // Get current context
    const context = window.currentAgencyContext || 'introduction';

    // Build form data
    const formData = new URLSearchParams();
    formData.append('message', originalMessage);
    formData.append('context', context);

    // Include editor content based on context
    if (context === 'introduction') {
        const editor = document.getElementById('introduction-editor');
        if (editor && editor.value) {
            formData.append('introduction-editor', editor.value);
        }
    }

    // Append formatted contexts if available
    let fullMessage = originalMessage;
    if (window.ContextManager) {
        const formattedContexts = window.ContextManager.getFormattedContexts();
        if (formattedContexts) {
            fullMessage = originalMessage + formattedContexts;
            formData.set('message', fullMessage);
        }
    }

    // Add user message to chat immediately
    addUserMessageToChat(originalMessage, chatMessages);

    // Create new abort controller for this request
    currentAbortController = new AbortController();
    console.log('Created new AbortController', currentAbortController);

    // Clear input and convert send button to stop button
    messageInput.value = '';
    convertToStopButton();

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

        // Check if streaming is enabled (enabled by default for all contexts)
        const useStreaming = window.isStreamingEnabled ? window.isStreamingEnabled() : true;

        if (useStreaming) {
            // Use streaming for all contexts
            await handleStreamingChatResponse(endpoint, formData, chatMessages, agencyID, currentAbortController);
        } else {
            // Use non-streaming when explicitly disabled
            await handleNonStreamingChatResponse(endpoint, formData, chatMessages, hasExistingConversation, currentAbortController);
        }

        // Clear context selections
        if (window.ContextManager) {
            window.ContextManager.clearSelections();
        }

        // Scroll to bottom
        chatMessages.scrollTop = chatMessages.scrollHeight;

    } catch (error) {
        // Check if it was an abort
        if (error.name === 'AbortError') {
            console.log('Request was cancelled');
            // Message already shown in stopChatProcessing
        } else {
            addErrorMessageToChat('Failed to send message. Please try again.', chatMessages);
        }
    } finally {
        // Clear abort controller
        currentAbortController = null;

        // Restore send button
        restoreSendButton();

        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }
    }

    return false;
};

/**
 * Handle streaming chat response using SSE
 * For chat, we stream JSON and extract the message at the end
 */
async function handleStreamingChatResponse(endpoint, formData, chatMessages, agencyID, abortController) {

    // Add streaming query parameter
    const streamEndpoint = `${endpoint}?stream=true`;

    // Create AI message container for streaming
    const aiMessageDiv = createAIMessageContainer(chatMessages);
    const messageBubble = aiMessageDiv.querySelector('.message-bubble');

    // Create streaming content area
    messageBubble.innerHTML = `
        <div class="streaming-content">
            <div class="is-flex is-align-items-center mb-2">
                <span class="icon has-text-info mr-2">
                    <i class="fas fa-brain fa-pulse"></i>
                </span>
                <strong>AI is processing...</strong>
            </div>
            <div class="streaming-text" style="white-space: pre-wrap; font-family: inherit;"></div>
        </div>
    `;

    const streamingText = messageBubble.querySelector('.streaming-text');

    try {
        const response = await fetch(streamEndpoint, {
            method: 'POST',
            body: formData,
            signal: abortController.signal
        });

        if (!response.ok) {
            // If we get a 500 error with an existing conversation, it might be lost (server restart)
            // Try again with a new conversation
            if (response.status === 500 && streamEndpoint.includes('/conversations/')) {
                chatMessages.dataset.conversationId = ''; // Clear the old conversation ID
                const newEndpoint = `/api/v1/agencies/${agencyID}/designer/conversations/web?stream=true`;
                const retryResponse = await fetch(newEndpoint, {
                    method: 'POST',
                    body: formData,
                    signal: abortController.signal
                });
                if (!retryResponse.ok) {
                    throw new Error(`HTTP error! status: ${retryResponse.status}`);
                }
                // Use the retry response
                return await processStreamingResponse(retryResponse, messageBubble, streamingText, chatMessages, abortController);
            }
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await processStreamingResponse(response, messageBubble, streamingText, chatMessages, abortController);

    } catch (error) {
        messageBubble.innerHTML = `<p class="has-text-danger">❌ ${error.message}</p>`;
        throw error;
    }
}

/**
 * Process the streaming response from the server
 */
async function processStreamingResponse(response, messageBubble, streamingText, chatMessages, abortController) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let currentEvent = '';
    let finalResult = null;

    try {
        while (true) {
            // Check if aborted
            if (abortController.signal.aborted) {
                reader.cancel();
                throw new DOMException('Request aborted', 'AbortError');
            }

            const { done, value } = await reader.read();
            if (done) {
                break;
            }

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop() || ''; // Keep incomplete line in buffer

            for (const line of lines) {
                if (!line.trim()) {
                    // Empty line - don't reset event immediately, just log
                    if (currentEvent) {
                    }
                    continue; // Keep currentEvent for next data line
                }

                if (line.startsWith('event:')) {
                    currentEvent = line.substring(6).trim();
                } else if (line.startsWith('data:')) {
                    const data = line.substring(5).trim();

                    // If no event type yet, treat as chunk continuation
                    if (!currentEvent) {
                        currentEvent = 'chunk';
                    }

                    if (currentEvent === 'chunk') {
                        // Display streaming text
                        streamingText.textContent += data;
                    } else if (currentEvent === 'complete') {
                        // Parse final result
                        try {
                            finalResult = JSON.parse(data);
                        } catch (e) {
                        }
                    } else if (currentEvent === 'error') {
                    } else if (currentEvent === 'start') {
                    }
                }
            }
        }

        // Display the final message
        if (finalResult) {
            const message = finalResult.explanation || finalResult.message || 'Changes applied successfully';

            // Store conversation ID if this was the first message
            if (finalResult.conversation_id) {
                chatMessages.dataset.conversationId = finalResult.conversation_id;
            }

            // Update the introduction textarea if it was changed
            if (finalResult.was_changed && finalResult.introduction) {
                const introTextarea = document.getElementById('introduction-editor');
                if (introTextarea) {
                    introTextarea.value = finalResult.introduction;
                }
            }

            // Refresh goals list if goals were changed
            const context = window.currentAgencyContext || '';
            if (finalResult.was_changed && context === 'goal-definition') {
                const agencyId = window.location.pathname.match(/agencies\/([^\/]+)/)?.[1];
                const goalsTableBody = document.getElementById('goals-table-body');

                if (agencyId && goalsTableBody && window.loadEntityList) {
                    console.log('Refreshing goals list after update');
                    window.loadEntityList('goals', 'goals-table-body', 3)
                        .catch(error => {
                            console.error('Failed to refresh goals list:', error);
                        });
                }
            }

            // Show if changes were made
            if (finalResult.was_changed && finalResult.changed_sections) {
                const sections = finalResult.changed_sections.join(', ');
                messageBubble.innerHTML = `
                    <p><strong>${message}</strong></p>
                    <p class="has-text-grey-light mt-2"><small>✓ Updated: ${sections}</small></p>
                `;
            } else {
                messageBubble.innerHTML = `<p>${message}</p>`;
            }
        } else {
            messageBubble.innerHTML = '<p class="has-text-grey">Response received</p>';
        }

        // Update timestamp
        const timeDiv = messageBubble.closest('.ai-message').querySelector('.message-time');
        if (timeDiv) {
            timeDiv.textContent = new Date().toLocaleTimeString('en-US', {
                hour: 'numeric',
                minute: '2-digit'
            });
        }
    } catch (error) {
        // If aborted, clean up and re-throw
        if (error.name === 'AbortError') {
            reader.cancel();
            throw error;
        }
        // For other errors, display them
        messageBubble.innerHTML = `<p class="has-text-danger">❌ ${error.message}</p>`;
        throw error;
    }
}

/**
 * Handle non-streaming chat response
 */
async function handleNonStreamingChatResponse(endpoint, formData, chatMessages, hasExistingConversation, abortController) {

    const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: formData,
        signal: abortController.signal
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    const html = await response.text();

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

/**
 * Initialize chat on page load
 * Clears stale conversation state that might persist in DOM after page refresh
 */
function initializeChat() {
    const chatMessages = document.getElementById('chat-messages');
    const chatForm = document.getElementById('chat-form');

    if (!chatMessages || !chatForm) {
        return; // Chat not present on this page
    }

    // Get conversation state from backend and frontend
    const backendHasConversation = chatForm.dataset.hasConversation === 'true';
    const frontendConversationId = chatMessages.dataset.conversationId;

    // Clear stale state if:
    // 1. Backend says no conversation exists, OR
    // 2. Frontend has no conversation ID but backend thinks there is one
    if (!backendHasConversation || (!frontendConversationId && backendHasConversation)) {
        delete chatMessages.dataset.conversationId;
        chatForm.dataset.hasConversation = 'false'; // Reset the form flag
    }

}

// Initialize chat when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeChat);
} else {
    initializeChat();
}

