// Introduction functionality
// Handles introduction editor and saving

// Uses global functions: getCurrentAgencyId, showNotification, specificationAPI

// Store original introduction for undo
let originalIntroduction = '';

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
        }
    } catch (error) {
        console.error('Error loading introduction:', error);
    }
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