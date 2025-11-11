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
            await handleStreamingChatResponse(endpoint, formData, chatMessages, agencyID);
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
 * For chat, we stream JSON and extract the message at the end
 */
async function handleStreamingChatResponse(endpoint, formData, chatMessages, agencyID) {
    console.log('  📡 Using STREAMING mode for chat');

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
            body: formData
        });

        if (!response.ok) {
            // If we get a 500 error with an existing conversation, it might be lost (server restart)
            // Try again with a new conversation
            if (response.status === 500 && streamEndpoint.includes('/conversations/')) {
                console.warn('  ⚠️  Existing conversation failed, retrying with new conversation...');
                chatMessages.dataset.conversationId = ''; // Clear the old conversation ID
                const newEndpoint = `/api/v1/agencies/${agencyID}/designer/conversations/web?stream=true`;
                const retryResponse = await fetch(newEndpoint, {
                    method: 'POST',
                    body: formData
                });
                if (!retryResponse.ok) {
                    throw new Error(`HTTP error! status: ${retryResponse.status}`);
                }
                // Use the retry response
                return await processStreamingResponse(retryResponse, messageBubble, streamingText, chatMessages);
            }
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await processStreamingResponse(response, messageBubble, streamingText, chatMessages);

    } catch (error) {
        console.error('Streaming failed:', error);
        messageBubble.innerHTML = `<p class="has-text-danger">❌ ${error.message}</p>`;
        throw error;
    }
}

/**
 * Process the streaming response from the server
 */
async function processStreamingResponse(response, messageBubble, streamingText, chatMessages) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let currentEvent = '';
    let finalResult = null;

    console.log('  🔍 Starting SSE stream parsing...');

    while (true) {
        const { done, value } = await reader.read();
        if (done) {
            console.log('  ✅ Stream reading complete');
            break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || ''; // Keep incomplete line in buffer

        for (const line of lines) {
            if (!line.trim()) {
                // Empty line - don't reset event immediately, just log
                if (currentEvent) {
                    console.log(`  📦 Empty line after event: ${currentEvent}`);
                }
                continue; // Keep currentEvent for next data line
            }

            console.log(`  📥 Received line: "${line.substring(0, 100)}${line.length > 100 ? '...' : ''}"`);

            if (line.startsWith('event:')) {
                currentEvent = line.substring(6).trim();
                console.log(`  🏷️  Event type: ${currentEvent}`);
            } else if (line.startsWith('data:')) {
                const data = line.substring(5).trim();

                // If no event type yet, treat as chunk continuation
                if (!currentEvent) {
                    currentEvent = 'chunk';
                }

                console.log(`  📊 Data for event '${currentEvent}': ${data.substring(0, 100)}${data.length > 100 ? '...(truncated for log)' : ''}`);

                if (currentEvent === 'chunk') {
                    // Display streaming text
                    streamingText.textContent += data;
                } else if (currentEvent === 'complete') {
                    // Parse final result
                    console.log('  🎯 Parsing completion data...');
                    console.log('  📄 Full completion JSON (length:', data.length, ')');
                    try {
                        finalResult = JSON.parse(data);
                        console.log('  ✅ Successfully parsed completion:', finalResult);
                    } catch (e) {
                        console.error('  ❌ Failed to parse completion data:', e);
                        console.error('  📄 Problematic data (first 500 chars):', data.substring(0, 500));
                    }
                } else if (currentEvent === 'error') {
                    console.error('  ❌ Server sent error event:', data);
                } else if (currentEvent === 'start') {
                    console.log('  🎬 Stream started:', data);
                }
            }
        }
    }

    console.log('  📋 Final result:', finalResult);

    // Display the final message
    if (finalResult) {
        console.log('  📝 Processing final result...');
        const message = finalResult.explanation || finalResult.message || 'Changes applied successfully';

        // Store conversation ID if this was the first message
        if (finalResult.conversation_id) {
            console.log('  💾 Storing conversation ID:', finalResult.conversation_id);
            chatMessages.dataset.conversationId = finalResult.conversation_id;
        } else {
            console.log('  ⚠️  No conversation_id in final result');
        }

        // Update the introduction textarea if it was changed
        if (finalResult.was_changed && finalResult.introduction) {
            console.log('  📝 Updating introduction textarea with new content');
            console.log('  📏 New introduction length:', finalResult.introduction.length);
            const introTextarea = document.getElementById('introduction-editor');
            if (introTextarea) {
                introTextarea.value = finalResult.introduction;
                console.log('  ✅ Textarea updated successfully');
            } else {
                console.error('  ❌ Could not find introduction-editor textarea');
            }
        } else {
            console.log('  ℹ️  No introduction update needed:', {
                was_changed: finalResult.was_changed,
                has_introduction: !!finalResult.introduction
            });
        }

        // Show if changes were made
        if (finalResult.was_changed && finalResult.changed_sections) {
            console.log('  ✅ Changes detected in sections:', finalResult.changed_sections);
            const sections = finalResult.changed_sections.join(', ');
            messageBubble.innerHTML = `
                <p><strong>${message}</strong></p>
                <p class="has-text-grey-light mt-2"><small>✓ Updated: ${sections}</small></p>
            `;
        } else {
            console.log('  ℹ️  No changes made');
            messageBubble.innerHTML = `<p>${message}</p>`;
        }
    } else {
        console.error('  ❌ No final result received from stream!');
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
        console.log('🧹 Clearing stale conversation state from DOM');
        delete chatMessages.dataset.conversationId;
        chatForm.dataset.hasConversation = 'false'; // Reset the form flag
    }

    console.log('💬 Chat initialized:', {
        backendHasConversation,
        frontendConversationId: frontendConversationId || 'none',
        finalState: chatForm.dataset.hasConversation
    });
}

// Initialize chat when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeChat);
} else {
    initializeChat();
}

console.log('✅ Chat streaming utilities loaded');
