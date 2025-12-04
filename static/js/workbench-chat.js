// Workbench Chat functionality
// Handles chat interface for workbench workflow instances

// Global chat messages array for workbench
window.workbenchChatMessages = [];

// Add message to global chat state
window.addWorkbenchChatMessage = function (role, content, timestamp = new Date()) {
    const message = {
        role: role, // 'user' or 'assistant'
        content: content,
        timestamp: timestamp
    };
    window.workbenchChatMessages.push(message);
    return message;
}

// Get all chat messages
window.getWorkbenchChatMessages = function () {
    return window.workbenchChatMessages;
}

// Clear all chat messages
window.clearWorkbenchChatMessages = function () {
    window.workbenchChatMessages = [];
}

// Load messages from DOM on page load
window.loadWorkbenchChatMessagesFromDOM = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (!chatContainer) return;

    window.workbenchChatMessages = [];

    // Parse existing messages from DOM
    const messageElements = chatContainer.querySelectorAll('.message');
    messageElements.forEach(el => {
        const isUser = el.classList.contains('user-message');
        const isAI = el.classList.contains('ai-message');

        if (isUser || isAI) {
            const bubble = el.querySelector('.message-bubble');
            const content = bubble ? bubble.textContent.trim() : '';
            const timeEl = el.querySelector('.message-time');
            const timeStr = timeEl ? timeEl.textContent.trim() : '';

            window.addWorkbenchChatMessage(
                isUser ? 'user' : 'assistant',
                content,
                timeStr ? new Date() : new Date()
            );
        }
    });
}

// Initialize auto-scroll for chat messages
window.initializeWorkbenchChatScroll = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        // Load existing messages into global state
        window.loadWorkbenchChatMessagesFromDOM();

        // Scroll to bottom on page load
        scrollToBottomWorkbench(chatContainer);
    }
}

// Scroll chat container to bottom
window.scrollToBottomWorkbench = function (container) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }
}

// Restore messages from global state to DOM
window.restoreWorkbenchChatMessagesFromState = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (!chatContainer || !window.workbenchChatMessages || window.workbenchChatMessages.length === 0) {
        return;
    }

    // Count current DOM messages
    const currentDOMMessages = chatContainer.querySelectorAll('.message').length;
    const stateMessages = window.workbenchChatMessages.length;

    // If we have more messages in state than in DOM, restore them
    if (stateMessages > currentDOMMessages) {
        // Clear and rebuild
        chatContainer.innerHTML = '';

        window.workbenchChatMessages.forEach(msg => {
            const messageDiv = document.createElement('div');
            messageDiv.className = msg.role === 'user' ? 'message user-message' : 'message ai-message';

            const timestamp = msg.timestamp instanceof Date ?
                msg.timestamp.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' }) :
                new Date().toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });

            messageDiv.innerHTML = `
                <div class="message-content">
                    <div class="message-bubble">
                        <p>${escapeHtml(msg.content)}</p>
                    </div>
                    <div class="message-time">${timestamp}</div>
                </div>
            `;

            chatContainer.appendChild(messageDiv);
        });

        window.scrollToBottomWorkbench(chatContainer);
    }
}

// Handle workbench chat form submission
window.handleWorkbenchChatSubmit = async function (event) {
    event.preventDefault();

    const form = event.target;
    const input = form.querySelector('#user-input');
    const submitBtn = form.querySelector('#chat-submit-btn');
    const submitIcon = form.querySelector('#chat-submit-icon i');
    const chatMessages = document.getElementById('chat-messages');

    const message = input.value.trim();
    if (!message) return false;

    const agencyID = form.dataset.agencyId;
    const instanceID = form.dataset.instanceId;
    const workflowID = form.dataset.workflowId;

    // Disable input and show loading state
    input.disabled = true;
    submitBtn.disabled = true;
    submitIcon.className = 'fas fa-spinner fa-spin';

    try {
        // Add user message to chat
        addUserMessage(message, chatMessages);
        window.addWorkbenchChatMessage('user', message);

        // Clear input
        input.value = '';

        // Add loading indicator
        const loadingDiv = document.createElement('div');
        loadingDiv.className = 'message ai-message';
        loadingDiv.id = 'ai-loading';
        loadingDiv.innerHTML = `
            <div class="message-content">
                <div class="message-bubble">
                    <p class="has-text-grey-light">
                        <i class="fas fa-circle-notch fa-spin mr-2"></i>
                        Thinking...
                    </p>
                </div>
            </div>
        `;
        chatMessages.appendChild(loadingDiv);
        window.scrollToBottomWorkbench(chatMessages);

        // Send message to server
        const response = await fetch(`/api/agencies/${agencyID}/workbench/${instanceID}/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                workflow_id: workflowID,
                instance_id: instanceID
            })
        });

        // Remove loading indicator
        const loading = document.getElementById('ai-loading');
        if (loading) {
            loading.remove();
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Add AI response to chat
        if (data.response) {
            addAIMessage(data.response, chatMessages);
            window.addWorkbenchChatMessage('assistant', data.response);
        }

        // Scroll to bottom
        window.scrollToBottomWorkbench(chatMessages);

    } catch (error) {
        console.error('Chat error:', error);

        // Remove loading indicator
        const loading = document.getElementById('ai-loading');
        if (loading) {
            loading.remove();
        }

        // Show error message
        addAIMessage('Sorry, I encountered an error. Please try again.', chatMessages);

    } finally {
        // Re-enable input
        input.disabled = false;
        submitBtn.disabled = false;
        submitIcon.className = 'fas fa-paper-plane';
        input.focus();
    }

    return false;
}

// Add user message to chat display
function addUserMessage(content, container) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message user-message';

    const timestamp = new Date().toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit'
    });

    messageDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                <p>${escapeHtml(content)}</p>
            </div>
            <div class="message-time">${timestamp}</div>
        </div>
    `;

    container.appendChild(messageDiv);
    window.scrollToBottomWorkbench(container);
}

// Add AI message to chat display
function addAIMessage(content, container) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ai-message';

    const timestamp = new Date().toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit'
    });

    messageDiv.innerHTML = `
        <div class="message-content">
            <div class="message-bubble">
                ${formatWorkbenchAIMessage(content)}
            </div>
            <div class="message-time">${timestamp}</div>
        </div>
    `;

    container.appendChild(messageDiv);
    window.scrollToBottomWorkbench(container);
}

// Format AI message content (support markdown-like formatting)
function formatWorkbenchAIMessage(content) {
    // Basic markdown-style formatting
    let formatted = escapeHtml(content);

    // Bold: **text**
    formatted = formatted.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

    // Italic: *text*
    formatted = formatted.replace(/\*([^*]+)\*/g, '<em>$1</em>');

    // Code: `code`
    formatted = formatted.replace(/`([^`]+)`/g, '<code>$1</code>');

    // Line breaks
    formatted = formatted.replace(/\n/g, '<br>');

    return `<p>${formatted}</p>`;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function () {
    window.initializeWorkbenchChatScroll();

    // Focus on input
    const input = document.getElementById('user-input');
    if (input) {
        input.focus();
    }

    // Handle Enter key for submission (without Shift)
    if (input) {
        input.addEventListener('keydown', function (e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                const form = document.getElementById('chat-form');
                if (form) {
                    form.dispatchEvent(new Event('submit', { cancelable: true }));
                }
            }
        });
    }
});
