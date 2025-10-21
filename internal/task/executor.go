package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrHandlerNotFound is returned when no handler is registered for a task type
	ErrHandlerNotFound = errors.New("handler not found for task type")
	// ErrInvalidTask is returned when task validation fails
	ErrInvalidTask = errors.New("invalid task")
	// ErrExecutorStopped is returned when executor is not running
	ErrExecutorStopped = errors.New("executor is stopped")
)

// Executor implements task execution with handler registry
type Executor struct {
	handlers map[string]TaskHandler
	config   ExecutorConfig
	mu       sync.RWMutex
	started  bool
}

// NewExecutor creates a new task executor
func NewExecutor(config ExecutorConfig) *Executor {
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 5 * time.Minute
	}
	if config.MaxTimeout <= 0 {
		config.MaxTimeout = 30 * time.Minute
	}
	if config.DefaultRetryPolicy == nil {
		config.DefaultRetryPolicy = DefaultRetryPolicy()
	}

	return &Executor{
		handlers: make(map[string]TaskHandler),
		config:   config,
	}
}

// Start starts the executor
func (e *Executor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return nil
	}

	e.started = true
	return nil
}

// Stop stops the executor
func (e *Executor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.started {
		return nil
	}

	e.started = false
	return nil
}

// Execute runs a task using the registered handler
func (e *Executor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	e.mu.RLock()
	if !e.started {
		e.mu.RUnlock()
		return nil, ErrExecutorStopped
	}
	e.mu.RUnlock()

	// Get handler for task type
	handler, err := e.GetHandler(task.Type)
	if err != nil {
		return &TaskResult{
			TaskID:      task.ID,
			AgentID:     task.AgentID,
			Status:      TaskStatusFailed,
			Error:       fmt.Sprintf("handler not found: %v", err),
			StartedAt:   time.Now(),
			CompletedAt: time.Now(),
		}, err
	}

	// Validate task
	if err := handler.Validate(task); err != nil {
		return &TaskResult{
			TaskID:      task.ID,
			AgentID:     task.AgentID,
			Status:      TaskStatusFailed,
			Error:       fmt.Sprintf("validation failed: %v", err),
			StartedAt:   time.Now(),
			CompletedAt: time.Now(),
		}, err
	}

	// Determine timeout
	timeout := task.Timeout
	if timeout <= 0 {
		timeout = e.config.DefaultTimeout
	}
	if timeout > e.config.MaxTimeout {
		timeout = e.config.MaxTimeout
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Update task context
	task.ctx = execCtx
	task.cancel = cancel

	startTime := time.Now()

	// Execute task
	result, err := handler.Execute(execCtx, task)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Create result if handler didn't return one
	if result == nil {
		result = &TaskResult{
			TaskID:    task.ID,
			AgentID:   task.AgentID,
			StartedAt: startTime,
		}
	}

	// Update result fields
	result.TaskID = task.ID
	result.AgentID = task.AgentID
	result.StartedAt = startTime
	result.CompletedAt = endTime
	result.Duration = duration

	// Determine final status
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.Status = TaskStatusTimeout
			result.Error = "task execution timeout"
		} else if errors.Is(err, context.Canceled) {
			result.Status = TaskStatusCancelled
			result.Error = "task execution cancelled"
		} else {
			result.Status = TaskStatusFailed
			if result.Error == "" {
				result.Error = err.Error()
			}
		}
	} else {
		result.Status = TaskStatusCompleted
	}

	// Collect metrics if enabled
	if e.config.MetricsEnabled {
		e.collectMetrics(task, result, duration)
	}

	return result, err
}

// RegisterHandler registers a handler for a task type
func (e *Executor) RegisterHandler(handler TaskHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}

	taskType := handler.Type()
	if taskType == "" {
		return errors.New("handler type cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Check for duplicate registration
	if _, exists := e.handlers[taskType]; exists {
		return fmt.Errorf("handler already registered for type: %s", taskType)
	}

	e.handlers[taskType] = handler
	return nil
}

// GetHandler retrieves the handler for a task type
func (e *Executor) GetHandler(taskType string) (TaskHandler, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	handler, exists := e.handlers[taskType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrHandlerNotFound, taskType)
	}

	return handler, nil
}

// ListHandlers returns all registered task types
func (e *Executor) ListHandlers() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	types := make([]string, 0, len(e.handlers))
	for taskType := range e.handlers {
		types = append(types, taskType)
	}
	return types
}

// collectMetrics collects execution metrics (placeholder implementation)
func (e *Executor) collectMetrics(task *Task, result *TaskResult, duration time.Duration) {
	// TODO: Implement metrics collection
	// This could integrate with monitoring systems like Prometheus
	_ = task
	_ = result
	_ = duration
}
