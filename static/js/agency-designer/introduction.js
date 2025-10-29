// Introduction functionality
// Handles introduction editor and saving

import { getCurrentAgencyId, showNotification } from './utils.js';

// Store original introduction for undo
let originalIntroduction = '';

// Load introduction editor and data
export function loadIntroductionEditor() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        return;
    }

    // Fetch the current overview/introduction
    fetch(`/api/v1/agencies/${agencyId}/overview`)
        .then(response => {
            if (!response.ok) {
                // If 404 or error, just show empty editor
                return { introduction: '' };
            }
            return response.json();
        })
        .then(data => {
            const editor = document.getElementById('introduction-editor');
            if (editor) {
                const introText = data.introduction || '';
                editor.value = introText;
                // Store original value for undo
                originalIntroduction = introText;
            }
        })
        .catch(error => {
            console.error('Error loading introduction:', error);
        });
}

// Save overview introduction
export function saveOverviewIntroduction() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        console.error('No agency ID found');
        showNotification('Error: No agency selected', 'error');
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

    fetch(`/api/v1/agencies/${agencyId}/overview`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ introduction: introduction })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to save introduction');
            }
            return response.json();
        })
        .then(data => {
            showNotification('Introduction saved successfully!', 'success');
            // Update original value after successful save
            originalIntroduction = editor.value;
        })
        .catch(error => {
            console.error('Error saving introduction:', error);
            showNotification('Error saving introduction', 'error');
        })
        .finally(() => {
            // Re-enable button
            if (saveBtn) {
                saveBtn.classList.remove('is-loading');
                saveBtn.disabled = false;
            }
        });
}

// Undo changes to overview introduction
export function undoOverviewIntroduction() {
    const editor = document.getElementById('introduction-editor');
    if (!editor) {
        console.error('Introduction editor not found');
        return;
    }

    // Restore original value
    editor.value = originalIntroduction;
    showNotification('Changes reverted', 'info');
}