package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoHandler(t *testing.T) {
	handler := NewEchoHandler()

	// Test Type()
	assert.Equal(t, "echo", handler.Type())

	// Test Validate()
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "echo",
		Payload: map[string]interface{}{"test": "data"},
	}
	err := handler.Validate(task)
	assert.NoError(t, err)

	// Test Validate() with nil task
	err = handler.Validate(nil)
	assert.Error(t, err)

	// Test Execute()
	result, err := handler.Execute(context.Background(), task)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.Equal(t, "test-1", result.TaskID)
	assert.Equal(t, "agent-1", result.AgentID)

	// Check echo result
	resultData, ok := result.Result["echo"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "data", resultData["test"])
}

func TestHTTPRequestHandler(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "test response"}`))
	}))
	defer server.Close()

	handler := NewHTTPRequestHandler(nil)

	// Test Type()
	assert.Equal(t, "http_request", handler.Type())

	// Test valid task
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "http_request",
		Payload: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
			"headers": map[string]interface{}{
				"User-Agent": "test-agent",
			},
		},
	}

	// Test Validate()
	err := handler.Validate(task)
	assert.NoError(t, err)

	// Test Execute()
	result, err := handler.Execute(context.Background(), task)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)

	// Check response
	assert.Equal(t, 200, result.Result["status_code"])
	assert.Equal(t, server.URL, result.Result["url"])
	assert.Equal(t, "GET", result.Result["method"])

	// Check JSON response
	body, ok := result.Result["body"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, body["success"].(bool))
	assert.Equal(t, "test response", body["message"])
}

func TestHTTPRequestHandler_Validation(t *testing.T) {
	handler := NewHTTPRequestHandler(nil)

	// Test missing URL
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "http_request",
		Payload: map[string]interface{}{
			"method": "GET",
		},
	}
	err := handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: url")

	// Test missing method
	task.Payload = map[string]interface{}{
		"url": "http://example.com",
	}
	err = handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: method")

	// Test nil payload
	task.Payload = nil
	err = handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload cannot be nil")
}

func TestDelayHandler(t *testing.T) {
	handler := NewDelayHandler()

	// Test Type()
	assert.Equal(t, "delay", handler.Type())

	// Test with duration string
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "delay",
		Payload: map[string]interface{}{
			"duration": "100ms",
		},
	}

	// Test Validate()
	err := handler.Validate(task)
	assert.NoError(t, err)

	// Test Execute()
	start := time.Now()
	result, err := handler.Execute(context.Background(), task)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.True(t, duration >= 100*time.Millisecond)
	assert.True(t, duration < 200*time.Millisecond) // Some buffer for test execution

	// Check result
	assert.Equal(t, "100ms", result.Result["duration"])
	assert.Equal(t, int64(100), result.Result["duration_ms"])
	assert.True(t, result.Result["completed"].(bool))
}

func TestDelayHandler_NumericDuration(t *testing.T) {
	handler := NewDelayHandler()

	// Test with numeric duration (seconds)
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "delay",
		Payload: map[string]interface{}{
			"duration": 0.05, // 50ms
		},
	}

	// Test Validate()
	err := handler.Validate(task)
	assert.NoError(t, err)

	// Test Execute()
	start := time.Now()
	result, err := handler.Execute(context.Background(), task)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.True(t, duration >= 50*time.Millisecond)
}

func TestDelayHandler_Validation(t *testing.T) {
	handler := NewDelayHandler()

	// Test missing duration
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "delay",
		Payload: map[string]interface{}{},
	}
	err := handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: duration")

	// Test invalid duration format
	task.Payload = map[string]interface{}{
		"duration": "invalid",
	}
	err = handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration format")

	// Test negative duration
	task.Payload = map[string]interface{}{
		"duration": -1,
	}
	err = handler.Validate(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duration cannot be negative")
}

func TestDelayHandler_Cancellation(t *testing.T) {
	handler := NewDelayHandler()

	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "delay",
		Payload: map[string]interface{}{
			"duration": "1s",
		},
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test Execute() with cancellation
	start := time.Now()
	result, err := handler.Execute(ctx, task)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, duration < 200*time.Millisecond) // Should cancel quickly
}

func TestErrorHandler(t *testing.T) {
	handler := NewErrorHandler()

	// Test Type()
	assert.Equal(t, "error", handler.Type())

	// Test with default error message
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "error",
		Payload: map[string]interface{}{},
	}

	// Test Validate()
	err := handler.Validate(task)
	assert.NoError(t, err)

	// Test Execute()
	result, err := handler.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "simulated error")
}

func TestErrorHandler_CustomMessage(t *testing.T) {
	handler := NewErrorHandler()

	// Test with custom error message
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "error",
		Payload: map[string]interface{}{
			"message": "custom error message",
		},
	}

	// Test Execute()
	result, err := handler.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "custom error message")
}

func TestErrorHandler_WithDelay(t *testing.T) {
	handler := NewErrorHandler()

	// Test with delay before error
	task := &Task{
		ID:      "test-1",
		AgentID: "agent-1",
		Type:    "error",
		Payload: map[string]interface{}{
			"message": "delayed error",
			"delay":   "50ms",
		},
	}

	// Test Execute()
	start := time.Now()
	result, err := handler.Execute(context.Background(), task)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "delayed error")
	assert.True(t, duration >= 50*time.Millisecond)
}
