/**
 * AI Streaming Utilities
 * Shared streaming functionality for all AI refine operations
 * Handles SSE (Server-Sent Events) streaming from backend AI services
 */

/**
 * StreamingOptions - Configuration for AI streaming requests
 * @typedef {Object} StreamingOptions
 * @property {string} url - The streaming endpoint URL
 * @property {Object} formData - Form data to send in POST request
 * @property {Function} onStart - Called when streaming starts
 * @property {Function} onChunk - Called for each text chunk received
 * @property {Function} onComplete - Called when streaming completes with final result
 * @property {Function} onError - Called on error
 * @property {HTMLElement} displayElement - Element to display streaming content (optional)
 */

/**
 * Execute an AI streaming request with SSE
 * @param {StreamingOptions} options - Streaming configuration
 * @returns {Promise<Object>} Final result object
 */
window.executeAIStream = async function (options) {
    const {
        url,
        formData,
        onStart = () => { },
        onChunk = () => { },
        onComplete = () => { },
        onError = () => { },
        displayElement = null
    } = options;

    // Track accumulated text
    let accumulatedText = '';
    let streamingTextElement = null;

    try {
        // Call onStart callback
        onStart();

        // Create streaming display if displayElement provided
        if (displayElement) {
            streamingTextElement = createStreamingDisplay(displayElement);
        } else {
        }

        // Make streaming request
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Read streaming response
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let finalResult = null;
        let chunkCount = 0;
        let lineCount = 0;

        while (true) {
            const { done, value } = await reader.read();
            if (done) {
                break;
            }

            chunkCount++;
            const chunk = decoder.decode(value, { stream: true });

            const lines = chunk.split('\n');

            for (const line of lines) {
                lineCount++;

                // Skip empty lines
                if (!line.trim()) {
                    continue;
                }

                // Parse SSE format
                if (line.startsWith('event: ')) {
                    const eventType = line.substring(7).trim();
                    continue;
                }

                if (line.startsWith('data: ')) {
                    const data = line.substring(6);

                    try {
                        const parsed = JSON.parse(data);

                        // Handle error events
                        if (parsed.error) {
                            throw new Error(parsed.error);
                        }

                        // Handle start event
                        if (parsed.status === 'streaming') {
                            if (streamingTextElement) {
                                streamingTextElement.textContent = 'Connecting to AI...';
                            }
                        }
                        // Handle complete event (has was_changed or similar completion fields)
                        else if (parsed.was_changed !== undefined || parsed.complete) {
                            finalResult = parsed;
                            if (displayElement) {
                                showCompletionResult(parsed, displayElement);
                            }
                            onComplete(parsed);
                            break;
                        }
                    } catch (e) {
                        // Not JSON, it's a text chunk
                        accumulatedText += data;

                        // Update display
                        if (streamingTextElement) {
                            streamingTextElement.textContent = accumulatedText;
                            // Auto-scroll to bottom
                            streamingTextElement.scrollTop = streamingTextElement.scrollHeight;
                        } else {
                        }

                        // Call chunk callback
                        onChunk(data, accumulatedText);
                    }
                } else {
                }
            }
        }

        return finalResult;
    } catch (error) {
        onError(error);
        if (displayElement) {
            showErrorResult(error.message, displayElement);
        }
        throw error;
    } finally {
    }
}

/**
 * Create streaming display UI
 * @param {HTMLElement} container - Container element
 * @returns {HTMLElement} The text display element
 */
function createStreamingDisplay(container) {

    const streamingDisplay = document.createElement('div');
    streamingDisplay.innerHTML = `
        <div class="is-flex is-align-items-center mb-2">
            <span class="icon has-text-info mr-2">
                <i class="fas fa-brain fa-pulse"></i>
            </span>
            <strong>AI is thinking...</strong>
        </div>
        <div id="streaming-text" class="content" style="white-space: pre-wrap; font-size: 0.9em;"></div>
    `;

    container.innerHTML = '';
    container.appendChild(streamingDisplay);

    const textElement = streamingDisplay.querySelector('#streaming-text');

    return textElement;
}/**
 * Show completion result in UI
 * @param {Object} result - Final result object
 * @param {HTMLElement} container - Container element
 */
function showCompletionResult(result, container) {
    const wasChanged = result.was_changed || result.changed || false;
    const explanation = result.explanation || result.message || 'AI has processed your request.';

    const notification = document.createElement('div');
    notification.className = wasChanged ? 'notification is-success' : 'notification is-info';
    notification.innerHTML = `
        <div class="is-flex is-align-items-center">
            <span class="icon has-text-${wasChanged ? 'success' : 'info'} mr-2">
                <i class="fas fa-${wasChanged ? 'check-circle' : 'info-circle'}"></i>
            </span>
            <div>
                <strong>${wasChanged ? 'Updated Successfully' : 'No Changes Needed'}</strong>
                <p class="mb-0">${explanation}</p>
            </div>
        </div>
    `;

    container.innerHTML = '';
    container.appendChild(notification);
}

/**
 * Show error result in UI
 * @param {string} errorMessage - Error message
 * @param {HTMLElement} container - Container element
 */
function showErrorResult(errorMessage, container) {
    const notification = document.createElement('div');
    notification.className = 'notification is-danger';
    notification.innerHTML = `
        <div class="is-flex is-align-items-center">
            <span class="icon has-text-danger mr-2">
                <i class="fas fa-exclamation-triangle"></i>
            </span>
            <div>
                <strong>Streaming Error</strong>
                <p class="mb-0">${errorMessage}</p>
            </div>
        </div>
    `;

    container.innerHTML = '';
    container.appendChild(notification);
}

/**
 * Check if streaming is enabled in user preferences
 * @returns {boolean} True if streaming is enabled
 */
window.isStreamingEnabled = function () {
    return window.localStorage.getItem('ai-use-streaming') !== 'false';
}

/**
 * Enable/disable streaming preference
 * @param {boolean} enabled - Whether to enable streaming
 */
window.setStreamingEnabled = function (enabled) {
    window.localStorage.setItem('ai-use-streaming', enabled ? 'true' : 'false');
}

/**
 * Execute AI refine with automatic streaming/non-streaming selection
 * @param {Object} options - Configuration options
 * @param {string} options.url - Base URL for the endpoint
 * @param {Object} options.formData - Form data to send
 * @param {Function} options.onComplete - Called with final result
 * @param {Function} options.onError - Called on error
 * @param {HTMLElement} options.displayElement - Display element (for streaming)
 * @returns {Promise<Object>} Final result
 */
window.executeAIRefine = async function (options) {
    const {
        url,
        formData,
        onComplete = () => { },
        onError = () => { },
        displayElement = null
    } = options;

    const useStreaming = window.isStreamingEnabled();

    // Add stream query parameter if streaming is enabled
    const requestUrl = useStreaming ? `${url}?stream=true` : url;

    if (useStreaming) {
        // Use streaming version
        return await window.executeAIStream({
            url: requestUrl,
            formData: formData,
            onComplete: onComplete,
            onError: onError,
            displayElement: displayElement
        });
    } else {
        // Use non-streaming version
        const response = await fetch(requestUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // For non-streaming, response is HTML to replace content
        const html = await response.text();

        if (displayElement) {
            displayElement.innerHTML = html;
        }

        onComplete({ html });
        return { html };
    }
}

