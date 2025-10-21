package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHandler is a mock implementation of TaskHandler for testing
type MockHandler struct {
	taskType      string
	shouldFail    bool
	shouldTimeout bool
	delay         time.Duration
	validateError error
}

func NewMockHandler(taskType string) *MockHandler {
	return &MockHandler{
		taskType: taskType,
	}
}

func (h *MockHandler) Type() string {
	return h.taskType
}

func (h *MockHandler) Validate(task *Task) error {
	if h.validateError != nil {
		return h.validateError
	}
	return nil
}

func (h *MockHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	if h.delay > 0 {
		select {
		case <-time.After(h.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if h.shouldTimeout {
		// Simulate a long-running task
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if h.shouldFail {
		return nil, errors.New("mock handler error")
	}

	result := &TaskResult{
		TaskID:  task.ID,
		AgentID: task.AgentID,
		Status:  TaskStatusCompleted,
		Result: map[string]interface{}{
			"mock": true,
			"type": task.Type,
		},
	}

	return result, nil
}

func (h *MockHandler) SetShouldFail(fail bool) {
	h.shouldFail = fail
}

func (h *MockHandler) SetShouldTimeout(timeout bool) {
	h.shouldTimeout = timeout
}

func (h *MockHandler) SetDelay(delay time.Duration) {
	h.delay = delay
}

func (h *MockHandler) SetValidateError(err error) {
	h.validateError = err
}

func TestExecutor_RegisterHandler(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
		MaxTimeout:     10 * time.Second,
	}
	executor := NewExecutor(config)

	handler := NewMockHandler("test")

	// Register handler
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	// Try to register duplicate
	err = executor.RegisterHandler(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Register nil handler
	err = executor.RegisterHandler(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Register handler with empty type
	emptyHandler := &MockHandler{taskType: ""}
	err = executor.RegisterHandler(emptyHandler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestExecutor_GetHandler(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)

	handler := NewMockHandler("test")
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	// Get existing handler
	retrieved, err := executor.GetHandler("test")
	require.NoError(t, err)
	assert.Equal(t, handler, retrieved)

	// Get non-existing handler
	_, err = executor.GetHandler("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler not found")
}

func TestExecutor_Execute_Success(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
		MetricsEnabled: true,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	handler := NewMockHandler("test")
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
		Payload: map[string]interface{}{"key": "value"},
	}

	result, err := executor.Execute(context.Background(), task)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.Equal(t, "task-1", result.TaskID)
	assert.Equal(t, "agent-1", result.AgentID)
	assert.True(t, result.Duration > 0)
}

func TestExecutor_Execute_Failure(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	handler := NewMockHandler("test")
	handler.SetShouldFail(true)
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
	}

	result, err := executor.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusFailed, result.Status)
	assert.Contains(t, result.Error, "mock handler error")
}

func TestExecutor_Execute_Timeout(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 100 * time.Millisecond,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	handler := NewMockHandler("test")
	handler.SetShouldTimeout(true)
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
	}

	start := time.Now()
	result, err := executor.Execute(context.Background(), task)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusTimeout, result.Status)
	assert.Contains(t, result.Error, "timeout")
	// Should timeout within reasonable time
	assert.True(t, duration < 1*time.Second)
}

func TestExecutor_Execute_CustomTimeout(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	handler := NewMockHandler("test")
	handler.SetDelay(200 * time.Millisecond)
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
		Timeout: 100 * time.Millisecond, // Shorter than delay
	}

	result, err := executor.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusTimeout, result.Status)
}

func TestExecutor_Execute_ValidationError(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	handler := NewMockHandler("test")
	handler.SetValidateError(errors.New("validation failed"))
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
	}

	result, err := executor.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusFailed, result.Status)
	assert.Contains(t, result.Error, "validation failed")
}

func TestExecutor_Execute_HandlerNotFound(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)
	executor.Start()
	defer executor.Stop()

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "nonexistent",
	}

	result, err := executor.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusFailed, result.Status)
	assert.Contains(t, result.Error, "handler not found")
}

func TestExecutor_Execute_NotStarted(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)
	// Don't start executor

	handler := NewMockHandler("test")
	err := executor.RegisterHandler(handler)
	require.NoError(t, err)

	task := &Task{
		ID:      "task-1",
		AgentID: "agent-1",
		Type:    "test",
	}

	result, err := executor.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrExecutorStopped, err)
}

func TestExecutor_ListHandlers(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)

	// No handlers initially
	handlers := executor.ListHandlers()
	assert.Empty(t, handlers)

	// Register some handlers
	handler1 := NewMockHandler("type1")
	handler2 := NewMockHandler("type2")

	err := executor.RegisterHandler(handler1)
	require.NoError(t, err)
	err = executor.RegisterHandler(handler2)
	require.NoError(t, err)

	handlers = executor.ListHandlers()
	assert.Len(t, handlers, 2)
	assert.Contains(t, handlers, "type1")
	assert.Contains(t, handlers, "type2")
}

func TestExecutor_StartStop(t *testing.T) {
	config := ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	}
	executor := NewExecutor(config)

	// Start executor
	err := executor.Start()
	assert.NoError(t, err)

	// Start again (should be no-op)
	err = executor.Start()
	assert.NoError(t, err)

	// Stop executor
	err = executor.Stop()
	assert.NoError(t, err)

	// Stop again (should be no-op)
	err = executor.Stop()
	assert.NoError(t, err)
}
