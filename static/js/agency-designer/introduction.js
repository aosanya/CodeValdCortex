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
        console.error('No agency ID found');
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
        console.error('Error loading introduction:', error);
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
        console.log('[Introduction] Text input detected');

        // Clear any existing timer
        clearTimeout(typingTimer);

        // Set a new timer to clear contexts after user stops typing
        typingTimer = setTimeout(function () {
            console.log('[Introduction] Typing stopped - checking for contexts and selections');

            if (!window.ContextManager) {
                console.warn('[Introduction] ContextManager not available');
                return;
            }

            const contexts = window.ContextManager.getAllContexts();
            const selections = window.ContextManager.getSelections();
            console.log('[Introduction] Current contexts:', contexts, 'selections:', selections);

            const hasContextsOrSelections = (contexts && contexts.length > 0) || (selections && selections.length > 0);

            if (hasContextsOrSelections) {
                console.log('[Introduction] Text changed - clearing contexts and selections');

                // Clear both contexts and selections
                window.ContextManager.clearAllContexts();
                window.ContextManager.clearSelections();

                console.log('[Introduction] ✅ Contexts and selections cleared');

                // Verify it was cleared
                setTimeout(() => {
                    const afterContexts = window.ContextManager.getAllContexts();
                    const afterSelections = window.ContextManager.getSelections();
                    console.log('[Introduction] After clear - contexts:', afterContexts, 'selections:', afterSelections);
                }, 100);
            } else {
                console.log('[Introduction] No contexts or selections to clear');
            }
        }, typingDelay);
    });    // Mark as attached
    editor._contextClearAttached = true;
    console.log('[Introduction] Context clear listener attached');
}

// Save overview introduction
window.saveOverviewIntroduction = async function () {
    const agencyId = window.getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        window.showNotification('Error: No agency selected', 'error');
        return;
    }

    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        console.error('Introduction editor not found');
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
        console.error('Error saving introduction:', error);
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
    console.log('🎯 [INTRODUCTION] AI Refine button clicked');

    const agencyId = window.getCurrentAgencyId();
    console.log('  🏢 Agency ID:', agencyId);
    if (!agencyId) {
        console.error('  ❌ No agency ID found');
        return;
    }

    const editor = document.getElementById('introduction-editor');
    console.log('  ✏️ Editor element:', editor);
    if (!editor) {
        console.error('  ❌ Introduction editor not found');
        return;
    }

    const contentElement = document.getElementById('introduction-content');
    console.log('  📄 Content element:', contentElement);
    if (!contentElement) {
        console.error('  ❌ Introduction content element not found');
        return;
    }

    try {
        // Show AI processing status
        console.log('  🔄 Showing AI processing status');
        if (window.showAIProcessStatus) {
            window.showAIProcessStatus('AI is refining your introduction...');
        }

        // Check if there's a pending user request from chat
        const pendingRequest = window.sessionStorage.getItem('pendingIntroductionRequest');
        console.log('  💬 Pending request from chat:', pendingRequest);

        const formData = new URLSearchParams({
            'introduction-editor': editor.value
        });
        console.log('  📝 Introduction text length:', editor.value.length);

        if (pendingRequest) {
            formData.append('user-request', pendingRequest);
            window.sessionStorage.removeItem('pendingIntroductionRequest');
            console.log('  ✅ Added pending request to form data');
        }

        console.log('  🚀 Calling executeAIRefine with streaming utility');
        // Use shared streaming utility with single endpoint
        await window.executeAIRefine({
            url: `/api/v1/agencies/${agencyId}/overview/refine`,
            formData: formData,
            displayElement: contentElement,
            onComplete: (result) => {
                console.log('  ✅ AI Refine completed');
                console.log('  📊 Result:', result);

                // Update editor with new content if available
                if (result.introduction) {
                    console.log('  ✏️ Updating editor with new introduction (length:', result.introduction.length, ')');
                    editor.value = result.introduction;
                } else {
                    console.log('  ℹ️ No introduction in result');
                }

                // Handle post-refinement tasks
                console.log('  🔄 Calling handlePostRefinement');
                handlePostRefinement();
            },
            onError: (error) => {
                console.error('  ❌ Error refining introduction:', error);
                window.showNotification('Failed to refine introduction. Please try again.', 'error');
            }
        });

    } catch (error) {
        console.error('  ❌ Caught error in handleAIRefineClick:', error);
        console.error('  📚 Error stack:', error.stack);
        window.showNotification('Failed to refine introduction. Please try again.', 'error');
    } finally {
        // Hide AI processing status
        console.log('  🔚 Hiding AI processing status');
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
            console.error('Error reloading chat messages:', error);
        });
}

// Undo changes to overview introduction
window.undoOverviewIntroduction = function () {
    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        console.error('Introduction editor not found');
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
        console.log('[Introduction] DOMContentLoaded - attaching context clear listener');
        attachContextClearListener(editor);
    }
});