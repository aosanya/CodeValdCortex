// HTMX events and interactions
// Handles HTMX-related functionality

import { scrollToBottom } from './chat.js';
import { initializeAgentSelection } from './agents.js';

// Initialize HTMX event listeners
export function initializeHTMXEvents() {
    // Log context being sent with chat messages
    document.body.addEventListener('htmx:configRequest', function (evt) {
        if (evt.detail.path && evt.detail.path.includes('/messages/web')) {
            console.log('[HTMX] Chat message request config:', {
                path: evt.detail.path,
                verb: evt.detail.verb,
                parameters: evt.detail.parameters,
                headers: evt.detail.headers
            });
        }
    });

    // Log what's actually being sent
    document.body.addEventListener('htmx:beforeRequest', function (evt) {
        if (evt.detail.path && evt.detail.path.includes('/messages/web')) {
            console.log('[HTMX] About to send chat request:', {
                path: evt.detail.path,
                parameters: evt.detail.parameters,
                target: evt.detail.target
            });

            // Try to log the actual form data
            const formData = new FormData(evt.detail.elt);
            console.log('[HTMX] Form data entries:');
            for (let [key, value] of formData.entries()) {
                console.log(`  ${key}: ${value}`);
            }
        }

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
        if (evt.detail.elt.matches('[hx-post*="overview/refine"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is refining your introduction...');
            }
        } else if (evt.detail.elt.matches('[hx-post*="refine"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is refining the design...');
            }
        }

        if (evt.detail.elt.matches('[hx-post*="generate"]')) {
            if (window.showAIProcessStatus) {
                window.showAIProcessStatus('AI is generating the final design...');
            }
        }
    });

    // Hide typing indicator and scroll when new message arrives
    document.body.addEventListener('htmx:afterSwap', function (evt) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator && evt.detail.target.id === 'chat-messages') {
            indicator.style.display = 'none';
        }

        // Check if this is an introduction refine operation
        const isIntroductionRefine = (
            evt.detail.target.id === 'introduction-content' ||
            evt.detail.target.classList.contains('introduction-content')
        );

        // For introduction refine, refresh chat messages to show AI explanation
        if (isIntroductionRefine) {
            const agencyId = window.location.pathname.match(/agencies\/([^\/]+)/)?.[1];
            const chatContainer = document.getElementById('chat-messages');

            if (agencyId && chatContainer) {
                fetch(`/agencies/${agencyId}/chat-messages`)
                    .then(response => response.text())
                    .then(html => {
                        chatContainer.innerHTML = html;
                        scrollToBottom(chatContainer);
                    })
                    .catch(error => {
                        console.error('Error refreshing chat after introduction refine:', error);
                    });
            }
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