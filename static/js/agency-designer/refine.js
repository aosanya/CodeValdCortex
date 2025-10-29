// Chat refinement functionality
// Handles design refinement and alternative suggestions

import { getCurrentAgencyId, showNotification } from './utils.js';

// Refine the current design
export function refineCurrentDesign() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Send a refinement message
    const refinementMessage = "Please analyze the current design and suggest improvements. Focus on optimization, efficiency, and best practices for agent coordination.";
    sendChatMessage(refinementMessage);
}

// Request alternative design approaches
export function requestAlternativeDesign() {
    const agencyId = getCurrentAgencyId();
    if (!agencyId) {
        showNotification('Error: No agency selected', 'error');
        return;
    }

    // Send an alternative request message
    const alternativeMessage = "Please suggest alternative architectural approaches for this agency design. Consider different agent patterns, communication strategies, and organizational structures.";
    sendChatMessage(alternativeMessage);
}

// Helper function to send chat messages programmatically
function sendChatMessage(message) {

    const userInput = document.getElementById('user-input');
    const chatForm = userInput.closest('form');

    if (!userInput || !chatForm) {
        console.error('Chat form not found');
        return;
    }

    // Set the message in the input
    userInput.value = message;

    // Trigger the form submission
    const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
    chatForm.dispatchEvent(submitEvent);
}