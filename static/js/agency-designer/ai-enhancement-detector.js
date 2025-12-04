// AI Enhancement Detector
// Detects JSON enhancement responses in chat and provides apply button

/**
 * Parse AI response for enhancement JSON
 * @param {string} messageText - The AI message text
 * @returns {Object|null} - Parsed enhancement object or null
 */
function parseEnhancementJSON(messageText) {
    if (!messageText) return null;

    // Look for JSON blocks in markdown code fences or plain JSON
    const jsonPatterns = [
        /```json\s*\n([\s\S]*?)\n```/g,
        /```\s*\n([\s\S]*?)\n```/g,
        /(\{[\s\S]*"prompt_instructions"[\s\S]*\})/g
    ];

    for (const pattern of jsonPatterns) {
        const matches = messageText.matchAll(pattern);
        for (const match of matches) {
            try {
                const jsonText = match[1];
                const parsed = JSON.parse(jsonText);

                // Verify it looks like an enhancement response
                if (parsed.prompt_instructions !== undefined || parsed.description !== undefined) {
                    return parsed;
                }
            } catch (e) {
                // Not valid JSON, continue
                continue;
            }
        }
    }

    return null;
}

/**
 * Add "Apply Enhancement" button to AI message
 * @param {HTMLElement} messageElement - The message element
 * @param {Object} enhancement - The parsed enhancement object
 */
function addApplyEnhancementButton(messageElement, enhancement) {
    // Check if button already exists
    if (messageElement.querySelector('.apply-enhancement-btn')) {
        return;
    }

    // Create button container
    const buttonContainer = document.createElement('div');
    buttonContainer.className = 'mt-3';
    buttonContainer.innerHTML = `
        <button class="button is-primary is-small apply-enhancement-btn">
            <span class="icon">
                <i class="fas fa-check"></i>
            </span>
            <span>Apply Enhancement to Deliverable</span>
        </button>
    `;

    const button = buttonContainer.querySelector('button');
    button.onclick = function () {
        if (window.PropertiesPanel && window.PropertiesPanel.applyAIEnhancement) {
            // Disable button to prevent double-click
            button.disabled = true;
            button.innerHTML = `
                <span class="icon">
                    <i class="fas fa-spinner fa-spin"></i>
                </span>
                <span>Applying...</span>
            `;

            // Apply enhancement
            window.PropertiesPanel.applyAIEnhancement(enhancement);

            // Re-enable after a delay
            setTimeout(() => {
                button.disabled = false;
                button.innerHTML = `
                    <span class="icon">
                        <i class="fas fa-check-circle"></i>
                    </span>
                    <span>Applied!</span>
                `;
                button.classList.remove('is-primary');
                button.classList.add('is-success');
            }, 1000);
        }
    };

    // Append to message
    const messageContent = messageElement.querySelector('.message-content');
    if (messageContent) {
        messageContent.appendChild(buttonContainer);
    }
}

/**
 * Monitor chat messages for AI enhancement responses
 */
function monitorChatForEnhancements() {
    // Use MutationObserver to watch for new messages
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) {
        return;
    }


    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            mutation.addedNodes.forEach((node) => {
                if (node.nodeType === Node.ELEMENT_NODE) {
                        nodeName: node.nodeName,
                        className: node.className,
                        hasNodeBeingEnhanced: !!window.PropertiesPanel._nodeBeingEnhanced
                    });

                    // Check if this is an AI message - collect into array
                    let aiMessages = [];

                    if (node.querySelectorAll) {
                        // Convert NodeList to array - look for various message selectors
                        const possibleSelectors = [
                            '.message.ai-message',
                            '.message.assistant',
                            '.ai-message',
                            '.assistant-message',
                            '[data-role="assistant"]',
                            '.message' // Fallback to any message
                        ];

                        for (const selector of possibleSelectors) {
                            const found = Array.from(node.querySelectorAll(selector));
                            if (found.length > 0) {
                                aiMessages.push(...found);
                                break;
                            }
                        }
                    }

                    // Check if the node itself is an AI message
                    if (node.classList && (
                        node.classList.contains('ai-message') ||
                        node.classList.contains('assistant') ||
                        node.classList.contains('assistant-message') ||
                        node.getAttribute('data-role') === 'assistant'
                    )) {
                        aiMessages.push(node);
                    }

                    aiMessages.forEach((aiMessage) => {
                        // First try to find hidden JSON data div (added by chat-streaming.js)
                        const jsonDataDiv = aiMessage.querySelector('.json-data');
                        let messageText = '';

                        if (jsonDataDiv) {
                            messageText = jsonDataDiv.textContent || jsonDataDiv.innerText;
                        } else {
                            // Fallback to full message content
                            const messageContent = aiMessage.querySelector('.message-text') ||
                                aiMessage.querySelector('.message-content') ||
                                aiMessage.querySelector('.content') ||
                                aiMessage;
                            messageText = messageContent.textContent || messageContent.innerText;
                        }

                        if (messageText) {

                            const enhancement = parseEnhancementJSON(messageText);

                            if (enhancement) {
                                    hasNodeBeingEnhanced: !!window.PropertiesPanel._nodeBeingEnhanced,
                                    nodeId: window.PropertiesPanel._nodeBeingEnhanced?.id,
                                    enhancementKeys: Object.keys(enhancement)
                                });

                                if (window.PropertiesPanel._nodeBeingEnhanced) {
                                    addApplyEnhancementButton(aiMessage, enhancement);
                                } else {
                                }
                            } else {
                            }
                        }
                    });
                }
            });
        });
    });

    observer.observe(chatMessages, {
        childList: true,
        subtree: true
    });

}

// Start monitoring when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', monitorChatForEnhancements);
} else {
    monitorChatForEnhancements();
}

// Also try to start after a delay in case chat loads dynamically
setTimeout(monitorChatForEnhancements, 1000);
