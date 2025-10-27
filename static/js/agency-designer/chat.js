// Chat functionality
// Handles chat interface and messaging

// Initialize auto-scroll for chat messages
export function initializeChatScroll() {
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        // Scroll to bottom on page load
        scrollToBottom(chatContainer);
    }
}

// Scroll chat container to bottom
export function scrollToBottom(container) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }
}