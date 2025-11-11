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
        onError = (error) => console.error('Streaming error:', error),
        displayElement = null
    } = options;

    console.log('🌊 [AI STREAMING] Starting stream request');
    console.log('  📍 URL:', url);
    console.log('  📝 FormData:', formData.toString());
    console.log('  🖼️ Display Element:', displayElement);

    // Track accumulated text
    let accumulatedText = '';
    let streamingTextElement = null;

    try {
        // Call onStart callback
        console.log('  ▶️ Calling onStart callback');
        onStart();

        // Create streaming display if displayElement provided
        if (displayElement) {
            console.log('  🎨 Creating streaming display in element:', displayElement.id);
            streamingTextElement = createStreamingDisplay(displayElement);
            console.log('  ✅ Streaming text element created:', streamingTextElement);
        } else {
            console.warn('  ⚠️ No display element provided - streaming will not be visible');
        }

        // Make streaming request
        console.log('  🌐 Sending POST request to:', url);
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData
        });

        console.log('  📥 Response received - Status:', response.status);
        console.log('  📋 Response headers:', {
            contentType: response.headers.get('Content-Type'),
            cacheControl: response.headers.get('Cache-Control'),
            connection: response.headers.get('Connection')
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Read streaming response
        console.log('  📖 Starting to read stream...');
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let finalResult = null;
        let chunkCount = 0;
        let lineCount = 0;

        while (true) {
            const { done, value } = await reader.read();
            if (done) {
                console.log('  ✅ Stream complete - Total chunks:', chunkCount, 'Total lines:', lineCount);
                break;
            }

            chunkCount++;
            const chunk = decoder.decode(value, { stream: true });
            console.log(`  📦 Chunk ${chunkCount} received (${chunk.length} bytes):`, chunk.substring(0, 100));

            const lines = chunk.split('\n');
            console.log(`  📄 Split into ${lines.length} lines`);

            for (const line of lines) {
                lineCount++;

                // Skip empty lines
                if (!line.trim()) {
                    console.log(`    ⏭️ Line ${lineCount}: Empty, skipping`);
                    continue;
                }

                console.log(`    📝 Line ${lineCount}:`, line.substring(0, 80));

                // Parse SSE format
                if (line.startsWith('event: ')) {
                    const eventType = line.substring(7).trim();
                    console.log(`    🏷️ Event type: ${eventType}`);
                    continue;
                }

                if (line.startsWith('data: ')) {
                    const data = line.substring(6);
                    console.log(`    💾 Data field (${data.length} chars):`, data.substring(0, 100));

                    try {
                        const parsed = JSON.parse(data);
                        console.log('    ✅ Parsed as JSON:', parsed);

                        // Handle error events
                        if (parsed.error) {
                            console.error('    ❌ Error event received:', parsed.error);
                            throw new Error(parsed.error);
                        }

                        // Handle start event
                        if (parsed.status === 'streaming') {
                            console.log('    🎬 Start event received');
                            if (streamingTextElement) {
                                streamingTextElement.textContent = 'Connecting to AI...';
                                console.log('    ✅ Updated streaming text element');
                            }
                        }
                        // Handle complete event (has was_changed or similar completion fields)
                        else if (parsed.was_changed !== undefined || parsed.complete) {
                            console.log('    🏁 Complete event received:', parsed);
                            finalResult = parsed;
                            if (displayElement) {
                                console.log('    🎨 Showing completion result in display element');
                                showCompletionResult(parsed, displayElement);
                            }
                            console.log('    📞 Calling onComplete callback');
                            onComplete(parsed);
                            break;
                        }
                    } catch (e) {
                        // Not JSON, it's a text chunk
                        console.log('    📝 Not JSON - treating as text chunk:', data.substring(0, 50));
                        accumulatedText += data;
                        console.log(`    📊 Accumulated text now ${accumulatedText.length} chars`);

                        // Update display
                        if (streamingTextElement) {
                            streamingTextElement.textContent = accumulatedText;
                            // Auto-scroll to bottom
                            streamingTextElement.scrollTop = streamingTextElement.scrollHeight;
                            console.log('    ✅ Updated streaming display');
                        } else {
                            console.warn('    ⚠️ No streaming text element to update!');
                        }

                        // Call chunk callback
                        console.log('    📞 Calling onChunk callback');
                        onChunk(data, accumulatedText);
                    }
                } else {
                    console.log(`    ⚠️ Unexpected line format:`, line.substring(0, 50));
                }
            }
        }

        console.log('  🎯 Final result:', finalResult);
        return finalResult;
    } catch (error) {
        console.error('  ❌ Streaming error:', error);
        console.error('  📚 Error stack:', error.stack);
        onError(error);
        if (displayElement) {
            console.log('  🎨 Showing error in display element');
            showErrorResult(error.message, displayElement);
        }
        throw error;
    } finally {
        console.log('🌊 [AI STREAMING] Request complete');
    }
}

/**
 * Create streaming display UI
 * @param {HTMLElement} container - Container element
 * @returns {HTMLElement} The text display element
 */
function createStreamingDisplay(container) {
    console.log('    🎨 Creating streaming display UI');
    console.log('    📦 Container:', container);

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

    console.log('    🗑️ Clearing container');
    container.innerHTML = '';
    console.log('    ➕ Appending streaming display');
    container.appendChild(streamingDisplay);

    const textElement = streamingDisplay.querySelector('#streaming-text');
    console.log('    🔍 Found streaming text element:', textElement);

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
        onError = (error) => console.error('AI refine error:', error),
        displayElement = null
    } = options;

    const useStreaming = window.isStreamingEnabled();

    console.log('🔀 [AI REFINE] Routing request');
    console.log('  ⚙️ Streaming enabled:', useStreaming);
    console.log('  � Base URL:', url);
    console.log('  �️ Display element:', displayElement);

    // Add stream query parameter if streaming is enabled
    const requestUrl = useStreaming ? `${url}?stream=true` : url;
    console.log('  � Request URL:', requestUrl);

    if (useStreaming) {
        // Use streaming version
        console.log('  ➡️ Using STREAMING version');
        return await window.executeAIStream({
            url: requestUrl,
            formData: formData,
            onComplete: onComplete,
            onError: onError,
            displayElement: displayElement
        });
    } else {
        // Use non-streaming version
        console.log('  ➡️ Using NON-STREAMING version');
        const response = await fetch(requestUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData
        });

        console.log('  📥 Non-streaming response status:', response.status);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // For non-streaming, response is HTML to replace content
        const html = await response.text();
        console.log('  📝 Received HTML response (length:', html.length, ')');

        if (displayElement) {
            console.log('  🎨 Updating display element with HTML');
            displayElement.innerHTML = html;
        }

        console.log('  📞 Calling onComplete callback');
        onComplete({ html });
        return { html };
    }
}

console.log('✅ AI Streaming utilities loaded');
