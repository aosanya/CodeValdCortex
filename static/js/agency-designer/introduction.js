// Introduction functionality
// Handles introduction editor and saving

// Uses global functions: getCurrentAgencyId, showNotification, specificationAPI

// Store original introduction for undo
let originalIntroduction = '';

// Update original introduction (used after AI refinement)
window.updateOriginalIntroduction = function (text) {
    originalIntroduction = text || '';
}

// Load introduction editor and data
window.loadIntroductionEditor = async function () {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    try {
        const data = await window.specificationAPI.getIntroduction();
        const editor = document.getElementById('introduction-editor');
        if (editor) {
            const introText = data.introduction || '';
            editor.value = introText;
            // Store original value for undo
            originalIntroduction = introText;

            // Attach input event listener to clear contexts on edit
            attachContextClearListener(editor);
        }
    } catch (error) {
    }
}

// Attach event listener to clear contexts when editor changes
function attachContextClearListener(editor) {
    // Remove any existing listener to avoid duplicates
    if (editor._contextClearAttached) {
        return;
    }

    let typingTimer;
    const typingDelay = 500; // Wait 500ms after user stops typing

    editor.addEventListener('input', function () {

        // Clear any existing timer
        clearTimeout(typingTimer);

        // Set a new timer to clear contexts after user stops typing
        typingTimer = setTimeout(function () {

            if (!window.ContextManager) {
                return;
            }

            const contexts = window.ContextManager.getAllContexts();
            const selections = window.ContextManager.getSelections();

            const hasContextsOrSelections = (contexts && contexts.length > 0) || (selections && selections.length > 0);

            if (hasContextsOrSelections) {

                // Clear both contexts and selections
                window.ContextManager.clearAllContexts();
                window.ContextManager.clearSelections();

                // Verify it was cleared
                setTimeout(() => {
                    const afterContexts = window.ContextManager.getAllContexts();
                    const afterSelections = window.ContextManager.getSelections();
                }, 100);
            } else {
            }
        }, typingDelay);
    });    // Mark as attached
    editor._contextClearAttached = true;
}

// Save overview introduction
window.saveOverviewIntroduction = async function () {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        return;
    }

    const introduction = editor.value;
    const saveBtn = document.getElementById('save-introduction-btn');

    // Disable button while saving
    if (saveBtn) {
        saveBtn.classList.add('is-loading');
        saveBtn.disabled = true;
    }

    try {
        await window.specificationAPI.updateIntroduction(introduction, 'user');

        window.showNotification('Introduction saved successfully!', 'success');
        // Update original value after successful save
        originalIntroduction = editor.value;
    } catch (error) {
        window.showNotification('Error saving introduction', 'error');
    } finally {
        // Re-enable button
        if (saveBtn) {
            saveBtn.classList.remove('is-loading');
            saveBtn.disabled = false;
        }
    }
}

// Update agency name display in various places
// (name updates are not handled here; agency name is managed centrally)

// Handle AI refine button click
window.handleAIRefineClick = async function () {

    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        return;
    }

    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        return;
    }

    const contentElement = document.getElementById('introduction-content');
    if (!contentElement) {
        return;
    }

    try {
        // Show AI processing status
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('AI is refining your introduction...');
        }

        // Check if there's a pending user request from chat
        const pendingRequest = window.sessionStorage.getItem('pendingIntroductionRequest');

        const formData = new URLSearchParams({
            'introduction-editor': editor.value
        });

        if (pendingRequest) {
            formData.append('user-request', pendingRequest);
            window.sessionStorage.removeItem('pendingIntroductionRequest');
        }

        // Use shared streaming utility with single endpoint
        await window.executeAIRefine({
            url: `/api/v1/agencies/${agencyId}/overview/refine`,
            formData: formData,
            displayElement: contentElement,
            onComplete: (result) => {

                // Update editor with new content if available
                if (result.introduction) {
                    editor.value = result.introduction;
                } else {
                }

                // Handle post-refinement tasks
                handlePostRefinement();
            },
            onError: (error) => {
                window.showNotification('Failed to refine introduction. Please try again.', 'error');
            }
        });

    } catch (error) {
        window.showNotification('Failed to refine introduction. Please try again.', 'error');
    } finally {
        // Hide AI processing status
        if (window.hideAIProcessStatus) {
            window.hideAIProcessStatus();
        }
    }
}

// Handle post-refinement tasks (reload editor and chat)
function handlePostRefinement() {
    // Reload the introduction editor to sync with updated content
    if (window.loadIntroductionEditor) {
        window.loadIntroductionEditor();
    }

    // Reload chat messages to show AI response
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) return;

    const triggerElement = document.getElementById('ai-refine-complete');
    const agencyName = triggerElement?.dataset?.agencyName || '';

    fetch(`/agencies/${agencyId}/chat-messages?agencyName=${encodeURIComponent(agencyName)}`)
        .then(response => response.text())
        .then(html => {
            const chatMessages = document.getElementById('chat-messages');
            if (chatMessages) {
                chatMessages.innerHTML = html;
                // Scroll to bottom
                if (window.scrollToBottom) {
                    window.scrollToBottom(chatMessages);
                }
            }
        })
        .catch(error => {
        });
}

// Undo changes to overview introduction
window.undoOverviewIntroduction = function () {
    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        return;
    }

    // Restore original value
    editor.value = originalIntroduction;
    window.showNotification('Changes reverted', 'info');
}

// Initialize introduction editor event listeners
document.addEventListener('DOMContentLoaded', function () {
    const editor = document.getElementById('introduction-editor');
    if (editor) {
        attachContextClearListener(editor);
    }
});