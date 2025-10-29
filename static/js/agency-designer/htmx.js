// HTMX events and interactions
// Handles HTMX-related functionality

import { scrollToBottom } from './chat.js';
import { initializeAgentSelection } from './agents.js';

// Initialize HTMX event listeners
export function initializeHTMXEvents() {
    // Show typing indicator when request starts
    document.body.addEventListener('htmx:beforeRequest', function (evt) {

        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            indicator.style.display = 'block';

            // Show AI process status for chat requests
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is processing your message...');
            }

            // Scroll to show typing indicator
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                setTimeout(() => scrollToBottom(chatContainer), 100);
            }
        }

        // Handle other AI operations
        if (evt.detail.elt.matches('[hx-post*="refine"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is refining the design...');
            }
        }

        if (evt.detail.elt.matches('[hx-post*="generate"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is generating the final design...');
            }
        }

        if (evt.detail.elt.matches('[hx-post*="introduction"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is updating the introduction...');
            }
        }
    });

    // Hide typing indicator and scroll when new message arrives
    document.body.addEventListener('htmx:afterSwap', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.target.id === 'chat-messages') {
            indicator.style.display = 'none';
        }

        // Hide AI process status only for specific targets that indicate completion
        const shouldHideStatus = (
            evt.detail.target.id === 'chat-messages' ||
            evt.detail.target.id === 'design-preview' ||
            evt.detail.target.id === 'introduction-content' ||
            evt.detail.target.classList.contains('introduction-content') ||
            evt.detail.target.closest('.details-content')
        );

        if (shouldHideStatus) {
            if (window.hideAIProcessStatus) {
                window.hideAIProcessStatus();
            }
        }

        // Scroll to bottom to show new message
        const chatContainer = document.getElementById('chat-messages');
        if (chatContainer && evt.detail.target.id === 'chat-messages') {
            setTimeout(() => scrollToBottom(chatContainer), 100);
        }

        // Re-initialize agent selection if sidebar was updated
        if (evt.detail.target.closest('.sidebar-content')) {
            initializeAgentSelection();
        }

        // Animate preview update
        if (evt.detail.target.id === 'design-preview') {
            evt.detail.target.classList.add('updated');
            setTimeout(() => {
                evt.detail.target.classList.remove('updated');
            }, 500);
        }
    });

    // Handle errors
    document.body.addEventListener('htmx:responseError', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.style.display = 'none';
        }

        // Hide AI process status on error
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }

        // Show error message
        // Show error in UI
        const target = evt.detail.target;
        if (target) {
            const errorMsg = document.createElement('div');
            errorMsg.className = 'notification is-danger is-light';
            errorMsg.textContent = 'Request failed. Please try again.';
            target.appendChild(errorMsg);

            setTimeout(() => errorMsg.remove(), 3000);
        }
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

    // Handle Enter key to submit
    document.body.addEventListener('keydown', function (evt) {
        const input = evt.target;
        if (input.matches('input[name="message"]') && evt.key === 'Enter' && !evt.shiftKey) {
            evt.preventDefault();
            const form = input.closest('form');
            if (form && typeof htmx !== 'undefined') {
                // Trigger HTMX submit
                htmx.trigger(form, 'submit');
            }
        }
    });
}