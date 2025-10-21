package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EchoHandler is a simple handler for testing that returns the input payload
type EchoHandler struct{}

// NewEchoHandler creates a new echo handler
func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

// Type returns the task type this handler supports
func (h *EchoHandler) Type() string {
	return "echo"
}

// Validate checks if the task payload is valid
func (h *EchoHandler) Validate(task *Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	// Echo handler accepts any payload
	return nil
}

// Execute processes the task and returns the result
func (h *EchoHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Simple echo - return the payload
	result := &TaskResult{
		TaskID:  task.ID,
		AgentID: task.AgentID,
		Status:  TaskStatusCompleted,
		Result: map[string]interface{}{
			"echo":     task.Payload,
			"metadata": task.Metadata,
			"type":     task.Type,
		},
	}

	return result, nil
}

// HTTPRequestHandler makes HTTP requests to external services
type HTTPRequestHandler struct {
	client *http.Client
}

// NewHTTPRequestHandler creates a new HTTP request handler
func NewHTTPRequestHandler(client *http.Client) *HTTPRequestHandler {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &HTTPRequestHandler{
		client: client,
	}
}

// Type returns the task type this handler supports
func (h *HTTPRequestHandler) Type() string {
	return "http_request"
}

// Validate checks if the task payload is valid
func (h *HTTPRequestHandler) Validate(task *Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	payload := task.Payload
	if payload == nil {
		return errors.New("payload cannot be nil")
	}

	// Check required fields
	if _, ok := payload["url"]; !ok {
		return errors.New("missing required field: url")
	}

	if _, ok := payload["method"]; !ok {
		return errors.New("missing required field: method")
	}

	return nil
}

// Execute processes the task and returns the result
func (h *HTTPRequestHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	payload := task.Payload

	// Extract request parameters
	url, _ := payload["url"].(string)
	method, _ := payload["method"].(string)
	headers, _ := payload["headers"].(map[string]interface{})
	body, _ := payload["body"].(string)

	// Create request
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		if strValue, ok := value.(string); ok {
			req.Header.Set(key, strValue)
		}
	}

	// Make request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response if possible
	var jsonResponse interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		// Not JSON, return as string
		jsonResponse = string(respBody)
	}

	result := &TaskResult{
		TaskID:  task.ID,
		AgentID: task.AgentID,
		Status:  TaskStatusCompleted,
		Result: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        jsonResponse,
			"url":         url,
			"method":      method,
		},
	}

	return result, nil
}

// DelayHandler simulates work by sleeping for a specified duration
type DelayHandler struct{}

// NewDelayHandler creates a new delay handler
func NewDelayHandler() *DelayHandler {
	return &DelayHandler{}
}

// Type returns the task type this handler supports
func (h *DelayHandler) Type() string {
	return "delay"
}

// Validate checks if the task payload is valid
func (h *DelayHandler) Validate(task *Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	payload := task.Payload
	if payload == nil {
		return errors.New("payload cannot be nil")
	}

	// Check duration field
	durationValue, ok := payload["duration"]
	if !ok {
		return errors.New("missing required field: duration")
	}

	// Try to parse as duration string or number
	switch v := durationValue.(type) {
	case string:
		if _, err := time.ParseDuration(v); err != nil {
			return fmt.Errorf("invalid duration format: %v", err)
		}
	case float64:
		if v < 0 {
			return errors.New("duration cannot be negative")
		}
	case int:
		if v < 0 {
			return errors.New("duration cannot be negative")
		}
	default:
		return errors.New("duration must be a string or number")
	}

	return nil
}

// Execute processes the task and returns the result
func (h *DelayHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	payload := task.Payload
	durationValue := payload["duration"]

	var duration time.Duration
	var err error

	// Parse duration
	switch v := durationValue.(type) {
	case string:
		duration, err = time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %w", err)
		}
	case float64:
		duration = time.Duration(v * float64(time.Second))
	case int:
		duration = time.Duration(v) * time.Second
	}

	// Sleep with context cancellation support
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		// Delay completed
	}

	result := &TaskResult{
		TaskID:  task.ID,
		AgentID: task.AgentID,
		Status:  TaskStatusCompleted,
		Result: map[string]interface{}{
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
			"completed":   true,
		},
	}

	return result, nil
}

// ErrorHandler simulates task failures for testing
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Type returns the task type this handler supports
func (h *ErrorHandler) Type() string {
	return "error"
}

// Validate checks if the task payload is valid
func (h *ErrorHandler) Validate(task *Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}
	// Error handler accepts any payload
	return nil
}

// Execute processes the task and returns the result
func (h *ErrorHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	payload := task.Payload

	// Get error message from payload or use default
	message := "simulated error"
	if msg, ok := payload["message"].(string); ok && msg != "" {
		message = msg
	}

	// Check if we should delay before failing
	if delayValue, ok := payload["delay"]; ok {
		var delay time.Duration
		switch v := delayValue.(type) {
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				delay = d
			}
		case float64:
			delay = time.Duration(v * float64(time.Second))
		case int:
			delay = time.Duration(v) * time.Second
		}

		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timer.C:
				// Continue to error
			}
		}
	}

	return nil, errors.New(message)
}
