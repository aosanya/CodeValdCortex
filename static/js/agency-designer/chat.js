// Chat functionality
// Handles chat interface and messaging

// Global chat messages array
window.chatMessages = [];

// Get the editor element to include based on current context
window.getContextEditor = function () {
    const context = window.currentAgencyContext || 'introduction';

    switch (context) {
        case 'introduction':
            return '#introduction-editor';
        case 'goal-definition':
            return '#goals-table-body';  // or whatever editor exists for goals
        case 'work-items':
        case 'workflows':
            return '#work-items-container';  // or whatever editor exists
        case 'roles':
            return '#roles-container';
        case 'raci-matrix':
            return '#raci-matrix-container';
        default:
            return '';  // No editor to include
    }
}

// Add message to global chat state
window.addChatMessage = function (role, content, timestamp = new Date()) {
    const message = {
        role: role, // 'user' or 'assistant'
        content: content,
        timestamp: timestamp
    };
    window.chatMessages.push(message);
    return message;
}

// Get all chat messages
window.getChatMessages = function () {
    return window.chatMessages;
}

// Clear all chat messages
window.clearChatMessages = function () {
    window.chatMessages = [];
}

// Load messages from DOM on page load
window.loadChatMessagesFromDOM = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (!chatContainer) return;

    window.chatMessages = [];

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

            window.addChatMessage(
                isUser ? 'user' : 'assistant',
                content,
                timeStr ? new Date() : new Date() // Could parse time if needed
            );
        }
    });

    console.log('Loaded', window.chatMessages.length, 'messages from DOM');
}

// Initialize auto-scroll for chat messages
window.initializeChatScroll = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        // Load existing messages into global state
        window.loadChatMessagesFromDOM();

        // Scroll to bottom on page load
        scrollToBottom(chatContainer);
    }
}

// Scroll chat container to bottom
window.scrollToBottom = function (container) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }
}

// Restore messages from global state to DOM
window.restoreChatMessagesFromState = function () {
    const chatContainer = document.getElementById('chat-messages');
    if (!chatContainer || !window.chatMessages || window.chatMessages.length === 0) {
        return;
    }

    // Count current DOM messages
    const currentDOMMessages = chatContainer.querySelectorAll('.message').length;
    const stateMessages = window.chatMessages.length;

    console.log('Chat state check:', { stateMessages, currentDOMMessages });

    // If we have more messages in state than in DOM, restore them
    if (stateMessages > currentDOMMessages) {
        console.warn('⚠️ Messages missing from DOM! Restoring from global state...');

        // Clear and rebuild
        chatContainer.innerHTML = '';

        window.chatMessages.forEach(msg => {
            const messageDiv = document.createElement('div');
            messageDiv.className = msg.role === 'user' ? 'message user-message' : 'message ai-message';

            const timestamp = msg.timestamp instanceof Date ?
                msg.timestamp.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' }) :
                new Date().toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });

            messageDiv.innerHTML = `
                <div class="message-content">
                    <div class="message-bubble">
                        <p>${msg.content.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p>
                    </div>
                    <div class="message-time">${timestamp}</div>
                </div>
            `;

            chatContainer.appendChild(messageDiv);
        });

        console.log('✅ Restored', window.chatMessages.length, 'messages from global state');
        window.scrollToBottom(chatContainer);
    }
}