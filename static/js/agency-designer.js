// AI Agency Designer - Chat Interactions
// Handles real-time chat features and HTMX events

document.addEventListener('DOMContentLoaded', function () {
    initializeChatScroll();
    initializeHTMXEvents();
});

// Initialize auto-scroll for chat messages
function initializeChatScroll() {
    const chatContainer = document.getElementById('chat-messages');
    if (chatContainer) {
        // Scroll to bottom on page load
        scrollToBottom(chatContainer);
    }
}

// Scroll chat container to bottom
function scrollToBottom(container) {
    if (container) {
        container.scrollTop = container.scrollHeight;
    }
}

// Initialize HTMX event listeners
function initializeHTMXEvents() {
    // Show typing indicator when request starts
    document.body.addEventListener('htmx:beforeRequest', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            indicator.style.display = 'block';

            // Scroll to show typing indicator
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                setTimeout(() => scrollToBottom(chatContainer), 100);
            }
        }
    });

    // Hide typing indicator and scroll when new message arrives
    document.body.addEventListener('htmx:afterSwap', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.style.display = 'none';
        }

        // Scroll to bottom to show new message
        const chatContainer = document.getElementById('chat-messages');
        if (chatContainer) {
            setTimeout(() => scrollToBottom(chatContainer), 100);
        }
    });

    // Handle errors
    document.body.addEventListener('htmx:responseError', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.style.display = 'none';
        }

        // Show error message
        console.error('Chat request failed:', evt.detail);
        alert('Failed to send message. Please try again.');
    });

    // Clear input after successful send
    document.body.addEventListener('htmx:afterRequest', function (evt) {
        if (evt.detail.successful && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            const input = evt.detail.elt.querySelector('input[name="message"]');
            if (input) {
                input.value = '';
                input.focus();
            }
        }
    });

    // Auto-resize textarea (if we switch to textarea later)
    const messageInputs = document.querySelectorAll('textarea[name="message"]');
    messageInputs.forEach(input => {
        input.addEventListener('input', function () {
            this.style.height = 'auto';
            this.style.height = (this.scrollHeight) + 'px';
        });
    });

    // Handle Enter key to submit (without Shift)
    document.body.addEventListener('keydown', function (evt) {
        const input = evt.target;
        if (input.matches('input[name="message"]') && evt.key === 'Enter' && !evt.shiftKey) {
            evt.preventDefault();
            const form = input.closest('form');
            if (form) {
                // Trigger HTMX submit
                htmx.trigger(form, 'submit');
            }
        }
    });
}

// Handle design preview updates
document.body.addEventListener('htmx:afterSwap', function (evt) {
    if (evt.detail.target.id === 'design-preview') {
        // Animate preview update
        evt.detail.target.classList.add('updated');
        setTimeout(() => {
            evt.detail.target.classList.remove('updated');
        }, 500);
    }
});

// Add visual feedback for phase transitions
function highlightActivePhase() {
    const phases = document.querySelectorAll('.phase-step');
    phases.forEach(phase => {
        if (phase.classList.contains('is-active')) {
            phase.style.transform = 'scale(1.1)';
            setTimeout(() => {
                phase.style.transform = 'scale(1)';
            }, 300);
        }
    });
}

// Call on page load
document.addEventListener('DOMContentLoaded', highlightActivePhase);

// Smooth scroll behavior for chat
if (document.getElementById('chat-messages')) {
    document.getElementById('chat-messages').style.scrollBehavior = 'smooth';
}
