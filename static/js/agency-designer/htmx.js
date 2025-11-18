// HTMX events and interactions
// Handles HTMX-related functionality

// Uses global functions: scrollToBottom, initializeAgentSelection, loadEntityList

// Initialize HTMX event listeners
window.initializeHTMXEvents = function () {
    // Log context being sent with chat messages
    document.body.addEventListener('htmx:configRequest', function (evt) {
        if (evt.detail.path && evt.detail.path.includes('/messages/web')) {
        }
    });

    // Log what's actually being sent
    document.body.addEventListener('htmx:beforeRequest', function (evt) {

        if (evt.detail.path && (evt.detail.path.includes('/messages/web') || evt.detail.path.includes('/conversations/web'))) {

            // Try to log the actual form data
            const formData = new FormData(evt.detail.elt);
            for (let [key, value] of formData.entries()) {
            }
        }

        const indicator = document.getElementById('typing-indicator');
        const isChatForm = evt.detail.elt.matches('form[hx-post*="conversations"]') ||
            evt.detail.elt.matches('form[hx-post*="messages"]');

        if (isChatForm) {

            // Get the input field and message
            const input = evt.detail.elt.querySelector('input[name="message"]');
            const message = input ? input.value.trim() : '';

            // Add user message to chat immediately
            if (message && message.length > 0) {
                const chatContainer = document.getElementById('chat-messages');
                if (chatContainer) {
                    // Add to global state
                    if (window.addChatMessage) {
                        window.addChatMessage('user', message);
                    }

                    // Create user message element
                    const userMessageDiv = document.createElement('div');
                    userMessageDiv.className = 'message user-message';
                    userMessageDiv.innerHTML = `
                        <div class="message-content">
                            <div class="message-bubble">
                                <p>${message.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p>
                            </div>
                        </div>
                    `;
                    chatContainer.appendChild(userMessageDiv);

                    // Scroll to show the new message
                    setTimeout(() => scrollToBottom(chatContainer), 50);
                }

                // Clear the input immediately
                if (input) {
                    input.value = '';
                }
            }

            // Show typing indicator if it exists
            if (indicator) {
                indicator.style.display = 'block';
            }

            // Show AI process status for chat requests with context-aware message
            if (window.showAIProcessStatus) {
                // Get the current context to show appropriate message
                const context = window.currentAgencyContext || '';
                let statusMessage = 'AI is processing your message...';

                switch (context) {
                    case 'introduction':
                        statusMessage = 'AI is refining your introduction...';
                        break;
                    case 'goal-definition':
                        statusMessage = 'AI is generating goals...';
                        break;
                    case 'work-items':
                        statusMessage = 'AI is processing work items...';
                        break;
                    case 'roles':
                        statusMessage = 'AI is working on roles...';
                        break;
                    case 'raci-matrix':
                        statusMessage = 'AI is updating RACI matrix...';
                        break;
                    default:
                        statusMessage = 'AI is processing your message...';
                }

                window.showAIProcessStatus(statusMessage);
            } else {
            }

            // Scroll to show typing indicator
            const chatContainer = document.getElementById('chat-messages');
            if (chatContainer) {
                setTimeout(() => scrollToBottom(chatContainer), 100);
            }
        } else {
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

            const textarea = document.getElementById('introduction-editor');
            if (textarea) {
            } else {
            }

            // NOTE: We don't refresh chat messages here anymore
            // The AI message is already appended by HTMX via OOB swap or separate response
            // Refreshing would remove the user message that was added by JavaScript
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
            // Track AI message in global state if it was just added
            if (window.addChatMessage) {
                const newMessages = chatContainer.querySelectorAll('.message.ai-message');
                if (newMessages.length > 0) {
                    const lastMessage = newMessages[newMessages.length - 1];
                    const bubble = lastMessage.querySelector('.message-bubble');
                    if (bubble) {
                        const content = bubble.textContent.trim();
                        // Check if this message is already in our global state
                        const existingMessages = window.getChatMessages ? window.getChatMessages() : [];
                        const isDuplicate = existingMessages.some(msg =>
                            msg.role === 'assistant' && msg.content === content
                        );
                        if (!isDuplicate) {
                            window.addChatMessage('assistant', content);
                        }
                    }
                }
            }

            // Check if messages were lost and restore from global state
            if (window.restoreChatMessagesFromState) {
                setTimeout(() => {
                    window.restoreChatMessagesFromState();
                    scrollToBottom(chatContainer);
                }, 50);
            } else {
                setTimeout(() => scrollToBottom(chatContainer), 100);
            }

            // Refresh goals list if we're in goal-definition context
            const context = window.currentAgencyContext || '';
            if (context === 'goal-definition') {
                const agencyId = window.location.pathname.match(/agencies\/([^\/]+)/)?.[1];
                const goalsTableBody = document.getElementById('goals-table-body');

                if (agencyId && goalsTableBody) {
                    window.loadEntityList('goals', 'goals-table-body', 3)
                        .then(() => {
                        })
                        .catch(error => {
                        });
                }
            }

            // Refresh work items list if we're in work-items context
            if (context === 'work-items') {
                const agencyId = window.location.pathname.match(/agencies\/([^\/]+)/)?.[1];
                const workItemsTableBody = document.getElementById('work-items-table-body');

                if (agencyId && workItemsTableBody) {
                    window.loadEntityList('work-items', 'work-items-table-body', 3)
                        .then(() => {
                        })
                        .catch(error => {
                        });
                }
            }
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

    // Focus input after successful send (input already cleared in beforeRequest)
    document.body.addEventListener('htmx:afterRequest', function (evt) {
        if (evt.detail.successful && evt.detail.elt.matches('form[hx-post*="conversations"]')) {
            const input = evt.detail.elt.querySelector('input[name="message"]');
            if (input) {
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