// Chat streaming functionality
// Handles chat form submission with SSE streaming support

// Global abort controller for stopping requests
let currentAbortController = null;

// Track if user is scrolled near bottom (for auto-scroll)
let userIsNearBottom = true;
let lastScrollTime = 0;

/**
 * Check if user is near the bottom of chat
 */
function isNearBottom(container) {
    if (!container) return true;
    const threshold = 100; // pixels from bottom
    const scrollBottom = container.scrollHeight - container.scrollTop - container.clientHeight;
    return scrollBottom < threshold;
}

/**
 * Auto-scroll to bottom if user hasn't manually scrolled away
 */
function autoScrollIfNearBottom(container) {
    if (!container) return;

    // Only auto-scroll if user is near bottom and hasn't scrolled recently
    const now = Date.now();
    const timeSinceLastScroll = now - lastScrollTime;

    if (userIsNearBottom || timeSinceLastScroll < 500) {
        container.scrollTop = container.scrollHeight;
        userIsNearBottom = true;
    }
}

/**
 * Setup scroll tracking for chat container
 */
function setupScrollTracking(container) {
    if (!container || container.dataset.scrollTracking === 'true') {
        return; // Already setup
    }

    container.dataset.scrollTracking = 'true';

    let scrollTimeout;
    container.addEventListener('scroll', () => {
        lastScrollTime = Date.now();

        // Clear existing timeout
        if (scrollTimeout) {
            clearTimeout(scrollTimeout);
        }

        // Check if user is near bottom after scroll settles
        scrollTimeout = setTimeout(() => {
            userIsNearBottom = isNearBottom(container);
        }, 100);
    });
}

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
            window.stopChatProcessing();
            return false;
        };
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
    if (currentAbortController) {
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

    // Setup scroll tracking and mark user as near bottom (they just submitted)
    setupScrollTracking(chatMessages);
    userIsNearBottom = true;

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

    // Setup scroll tracking for auto-scroll
    setupScrollTracking(chatMessages);

    // User just submitted, so they're at bottom
    userIsNearBottom = true;

    // Add streaming query parameter
    const streamEndpoint = `${endpoint}?stream=true`;

    // Create AI message container for streaming
    const aiMessageDiv = createAIMessageContainer(chatMessages);
    const messageBubble = aiMessageDiv.querySelector('.message-bubble');

    // Scroll to new message
    autoScrollIfNearBottom(chatMessages);

    // Create streaming content area
    messageBubble.innerHTML = `
        <div class="streaming-content">
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
    // Buffer for accumulating JSON-like output separately from visible streaming text
    let jsonBuffer = '';
    // Progress tag state machine
    let isAccumulatingProgress = false;
    let progressMessageBuffer = '';
    let progressTagCount = 0;
    // Buffer for accumulating partial chunks (to handle tags split across chunks)
    let chunkBuffer = '';

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
                        // Get the progress token from the global context
                        const progressToken = window.currentProgressToken;

                        if (!progressToken) {
                            // No progress token set, just accumulate normally
                            const looksLikeJson = /"name"|"description"|"prompt_instructions"|"suggested_children"|^\s*\{/.test(data);
                            if (looksLikeJson || jsonBuffer) {
                                streamingText.style.display = 'none';
                                jsonBuffer += data;
                            } else {
                                streamingText.style.display = '';
                                streamingText.textContent += data;
                            }
                            autoScrollIfNearBottom(chatMessages);
                            continue;
                        }

                        // Add chunk to buffer (tags may be split across chunks)
                        chunkBuffer += data;

                        // State machine for progress tag detection
                        const openTag = `<${progressToken}>`;
                        const closeTag = `</${progressToken}>`;
                        let remainingData = chunkBuffer;

                        while (remainingData.length > 0) {
                            if (isAccumulatingProgress) {
                                // We're inside a progress tag, look for closing tag
                                const closeIndex = remainingData.indexOf(closeTag);

                                if (closeIndex !== -1) {
                                    // Found closing tag - complete the progress message
                                    progressMessageBuffer += remainingData.substring(0, closeIndex);

                                    // Clean and format the progress message
                                    let progressMessage = progressMessageBuffer.trim()
                                        .replace(/([a-z])([A-Z])/g, '$1 $2')  // Add spaces between camelCase
                                        .replace(/([a-z])([a-z])([A-Z])/g, '$1$2 $3')
                                        .replace(/([a-z]{2,})([A-Z][a-z])/g, '$1 $2');

                                    if (progressMessage) {
                                        progressTagCount++;
                                        console.log('[PROGRESS] Tag #' + progressTagCount + ':', progressMessage);

                                        // Remove spinning icon from previous progress bubble
                                        const previousBubbles = chatMessages.querySelectorAll('.progress-bubble .fa-circle-notch');
                                        previousBubbles.forEach(icon => {
                                            icon.classList.remove('fa-circle-notch', 'fa-spin');
                                            icon.classList.add('fa-check-circle');
                                        });

                                        // Create progress bubble with spinning icon (will be the latest)
                                        const progressBubble = document.createElement('div');
                                        progressBubble.className = 'message ai-message mb-2 progress-bubble';
                                        progressBubble.dataset.progressId = `progress-${Date.now()}-${progressTagCount}`;
                                        progressBubble.innerHTML = `
                                            <div class="message-content">
                                                <div class="message-bubble">
                                                    <span class="tag is-info is-light">
                                                        <i class="fas fa-circle-notch fa-spin mr-1"></i> ${escapeHtml(progressMessage)}
                                                    </span>
                                                </div>
                                            </div>
                                        `;
                                        chatMessages.insertBefore(progressBubble, messageBubble.closest('.ai-message'));
                                        autoScrollIfNearBottom(chatMessages);
                                    }

                                    // Reset state and continue with remaining data after closing tag
                                    isAccumulatingProgress = false;
                                    progressMessageBuffer = '';
                                    remainingData = remainingData.substring(closeIndex + closeTag.length);
                                    // Clear chunkBuffer - we've processed up to the closing tag
                                    chunkBuffer = remainingData;
                                } else {
                                    // No closing tag yet
                                    // Check if we might have a partial closing tag at the end
                                    const maxTagLength = closeTag.length;
                                    if (remainingData.length > maxTagLength) {
                                        // Accumulate data except for last maxTagLength characters (might be partial tag)
                                        progressMessageBuffer += remainingData.substring(0, remainingData.length - maxTagLength);
                                        // Keep last maxTagLength chars in buffer (might be partial closing tag)
                                        chunkBuffer = remainingData.substring(remainingData.length - maxTagLength);
                                    } else {
                                        // Buffer is shorter than max tag length, keep everything for next chunk
                                        chunkBuffer = remainingData;
                                    }
                                    remainingData = '';
                                }
                            } else {
                                // We're not in a progress tag, look for opening tag
                                const openIndex = remainingData.indexOf(openTag);

                                if (openIndex !== -1) {
                                    console.log('[PROGRESS] Found opening tag at index:', openIndex, 'in chunk:', remainingData.substring(0, 100));
                                    // Found opening tag - process data before it, then start accumulating
                                    const beforeTag = remainingData.substring(0, openIndex);

                                    if (beforeTag) {
                                        // Process the data before the tag
                                        const cleaned = beforeTag.replace(/```(?:json)?/gi, '').replace(/\s+/g, ' ');
                                        if (cleaned.trim()) {
                                            const looksLikeJson = /"name"|"description"|"prompt_instructions"|"suggested_children"|^\s*\{/.test(cleaned);
                                            if (looksLikeJson || jsonBuffer) {
                                                streamingText.style.display = 'none';
                                                jsonBuffer += cleaned;
                                            } else {
                                                streamingText.style.display = '';
                                                streamingText.textContent += cleaned;
                                            }
                                        }
                                    }

                                    // Start accumulating progress message
                                    isAccumulatingProgress = true;
                                    progressMessageBuffer = '';
                                    remainingData = remainingData.substring(openIndex + openTag.length);
                                    // Clear chunkBuffer - we've processed up to the opening tag
                                    chunkBuffer = remainingData;
                                } else {
                                    // No opening tag found yet
                                    // Check if we might have a partial tag at the end
                                    const maxTagLength = openTag.length;
                                    if (remainingData.length > maxTagLength) {
                                        // Process data except for last maxTagLength characters (might be partial tag)
                                        const safeData = remainingData.substring(0, remainingData.length - maxTagLength);
                                        const cleaned = safeData.replace(/```(?:json)?/gi, '').replace(/\s+/g, ' ');
                                        if (cleaned.trim()) {
                                            const looksLikeJson = /"name"|"description"|"prompt_instructions"|"suggested_children"|^\s*\{/.test(cleaned);
                                            if (looksLikeJson || jsonBuffer) {
                                                streamingText.style.display = 'none';
                                                jsonBuffer += cleaned;
                                            } else {
                                                streamingText.style.display = '';
                                                streamingText.textContent += cleaned;
                                            }
                                        }
                                        // Keep last maxTagLength chars in buffer (might be partial tag)
                                        chunkBuffer = remainingData.substring(remainingData.length - maxTagLength);
                                    } else {
                                        // Buffer is shorter than max tag length, keep everything
                                        chunkBuffer = remainingData;
                                    }
                                    remainingData = '';
                                }
                            }
                        }

                        // Auto-scroll if user is near bottom
                        autoScrollIfNearBottom(chatMessages);
                    } else if (currentEvent === 'complete') {
                        // Parse final result
                        try {
                            finalResult = JSON.parse(data);
                        } catch (e) {
                            console.debug('[MVP-054] failed to parse complete event JSON:', e.message);
                        }
                    } else if (currentEvent === 'error') {
                    } else if (currentEvent === 'start') {
                    }
                }
            }
        }

        // If we accumulated a jsonBuffer, try to parse it as the final result
        if (!finalResult && jsonBuffer) {
            try {
                finalResult = JSON.parse(jsonBuffer);
            } catch (e) {
                console.debug('[MVP-054] failed to parse jsonBuffer into JSON:', e.message);
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
                    window.loadEntityList('goals', 'goals-table-body', 3)
                        .catch(error => {
                        });
                }
            }

            // Refresh work items list if work items were changed
            if (finalResult.was_changed && context === 'work-items') {
                // Check if this was a deliverable enhancement (has suggested_children or prompt_instructions)
                const isDeliverableEnhancement = finalResult.suggested_children !== undefined ||
                    finalResult.prompt_instructions !== undefined;

                if (isDeliverableEnhancement) {
                    console.log('[MVP-054] Deliverable enhancement detected, reloading work item editor...');

                    // Get the work item key from editor state
                    const workItemKey = window.workItemEditorState?.workItemKey;

                    if (workItemKey && window.loadWorkItemData) {
                        console.log('[MVP-054] Reloading work item:', workItemKey);

                        window.loadWorkItemData(workItemKey)
                            .then(() => {
                                console.log('[MVP-054] Work item reloaded successfully');
                                if (window.showNotification) {
                                    window.showNotification('Deliverable tree updated', 'success');
                                }
                            })
                            .catch(error => {
                                console.error('[MVP-054] Failed to reload work item:', error);
                            });
                    } else {
                        console.warn('[MVP-054] Cannot reload - missing dependencies:', {
                            workItemKey: !!workItemKey,
                            loadWorkItemData: !!window.loadWorkItemData
                        });
                    }
                } else if (window.loadWorkItems) {
                    // Regular work items list update (not deliverable enhancement)
                    window.loadWorkItems();
                }
            }

            // Refresh roles list if roles were changed
            if (finalResult.was_changed && context === 'roles') {
                if (window.loadRoles) {
                    window.loadRoles();
                }
            }

            // Refresh workflows list if workflows were changed
            if (finalResult.was_changed && context === 'workflows') {
                if (window.loadWorkflows) {
                    window.loadWorkflows();
                }
            }

            // Show the message and keep JSON hidden for AI enhancement detector
            // The JSON buffer or streaming text may contain the full JSON that needs to be detected
            messageBubble.innerHTML = `
                <p>${message}</p>
                <div class="json-data" style="display: none;">
                    \`\`\`json
                    ${jsonBuffer || ''}
                    \`\`\`
                </div>
            `;

            console.log('[MVP-054] Final result displayed, JSON kept in hidden div for enhancement detector');

            // Auto-scroll to show final message
            autoScrollIfNearBottom(chatMessages);
        } else {
            messageBubble.innerHTML = '<p class="has-text-grey">Response received</p>';
            autoScrollIfNearBottom(chatMessages);
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

    // Auto-scroll to show new user message
    autoScrollIfNearBottom(container);
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

    // Auto-scroll to show new AI message container
    autoScrollIfNearBottom(container);

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

