package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrManagerStopped is returned when manager is not running
	ErrManagerStopped = errors.New("task manager is stopped")
)

// Manager implements comprehensive task management
type Manager struct {
	scheduler TaskScheduler
	executor  TaskExecutor
	repo      TaskRepository
	config    ManagerConfig

	started bool
	mu      sync.RWMutex
}

// NewManager creates a new task manager
func NewManager(config ManagerConfig, repo TaskRepository) *Manager {
	// Set default configuration
	if config.WorkerPool.MaxWorkers <= 0 {
		config.WorkerPool.MaxWorkers = 10
	}
	if config.WorkerPool.MinWorkers <= 0 {
		config.WorkerPool.MinWorkers = 2
	}
	if config.Executor.DefaultTimeout <= 0 {
		config.Executor.DefaultTimeout = 5 * time.Minute
	}

	// Create executor
	executor := NewExecutor(config.Executor)

	// Register built-in handlers
	executor.RegisterHandler(NewEchoHandler())
	executor.RegisterHandler(NewHTTPRequestHandler(nil))
	executor.RegisterHandler(NewDelayHandler())
	executor.RegisterHandler(NewErrorHandler())

	// Create scheduler
	scheduler := NewScheduler(config.WorkerPool, executor, repo)

	return &Manager{
		scheduler: scheduler,
		executor:  executor,
		repo:      repo,
		config:    config,
	}
}

// Start starts the task manager
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	// Start executor
	if err := m.executor.Start(); err != nil {
		return fmt.Errorf("failed to start executor: %w", err)
	}

	// Start scheduler
	if err := m.scheduler.Start(); err != nil {
		m.executor.Stop() // Cleanup
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	m.started = true
	return nil
}

// Stop stops the task manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	// Stop scheduler first (to stop accepting new tasks)
	if err := m.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	// Stop executor
	if err := m.executor.Stop(); err != nil {
		return fmt.Errorf("failed to stop executor: %w", err)
	}

	m.started = false
	return nil
}

// Submit adds a task to the execution queue
func (m *Manager) Submit(ctx context.Context, task *Task) error {
	m.mu.RLock()
	if !m.started {
		m.mu.RUnlock()
		return ErrManagerStopped
	}
	m.mu.RUnlock()

	// Validate task
	if task == nil {
		return errors.New("task cannot be nil")
	}

	// Generate ID if not provided
	if task.ID == "" {
		task.ID = generateTaskID()
	}

	// Set default values
	if task.Status == "" {
		task.Status = TaskStatusPending
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Apply default retry policy if not set
	if task.RetryPolicy == nil && m.config.Executor.DefaultRetryPolicy != nil {
		task.RetryPolicy = m.config.Executor.DefaultRetryPolicy
	}

	// Store task if persistence is enabled
	if m.config.PersistTasks && m.repo != nil {
		if err := m.repo.StoreTask(ctx, task); err != nil {
			return fmt.Errorf("failed to store task: %w", err)
		}
	}

	// Schedule task
	if err := m.scheduler.Schedule(task); err != nil {
		return fmt.Errorf("failed to schedule task: %w", err)
	}

	return nil
}

// RegisterHandler registers a task handler
func (m *Manager) RegisterHandler(handler TaskHandler) error {
	return m.executor.RegisterHandler(handler)
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	// Try to get from scheduler queue first
	if task, err := m.scheduler.GetTask(taskID); err == nil {
		return task, nil
	}

	// Fall back to repository if available
	if m.repo != nil {
		return m.repo.GetTask(ctx, taskID)
	}

	return nil, errors.New("task not found")
}

// GetTaskResult retrieves task result
func (m *Manager) GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error) {
	if m.repo == nil {
		return nil, errors.New("repository not available")
	}

	return m.repo.GetResult(ctx, taskID)
}

// CancelTask cancels a pending/running task
func (m *Manager) CancelTask(ctx context.Context, taskID string) error {
	// Try to cancel from scheduler queue
	if err := m.scheduler.Cancel(taskID); err == nil {
		return nil
	}

	// If not in queue, try to get from repository and mark as cancelled
	if m.repo != nil {
		task, err := m.repo.GetTask(ctx, taskID)
		if err != nil {
			return err
		}

		// Only cancel if not in terminal state
		if !task.IsTerminal() {
			task.Status = TaskStatusCancelled
			task.CompletedAt = time.Now()

			// Cancel the task's context if available
			if task.cancel != nil {
				task.cancel()
			}

			return m.repo.UpdateTask(ctx, task)
		}
	}

	return errors.New("task not found or already completed")
}

// ListTasks lists tasks with filters
func (m *Manager) ListTasks(ctx context.Context, filters TaskFilters) ([]*Task, error) {
	if m.repo == nil {
		// Return queued tasks if no repository
		queuedTasks := m.scheduler.GetQueuedTasks()

		// Apply basic filtering
		var filtered []*Task
		for _, task := range queuedTasks {
			if m.matchesFilters(task, filters) {
				filtered = append(filtered, task)
			}
		}

		return filtered, nil
	}

	// Use repository for comprehensive listing
	return m.repo.ListTasks(ctx, "", filters) // Empty agentID means all agents
}

// GetMetrics retrieves task execution metrics
func (m *Manager) GetMetrics(ctx context.Context, agentID string) (*AgentTaskMetrics, error) {
	if m.repo == nil {
		return nil, errors.New("repository not available")
	}

	return m.repo.GetMetrics(ctx, agentID)
}

// GetStatus returns the current status of the task manager
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := map[string]interface{}{
		"started":       m.started,
		"queue_size":    m.scheduler.Size(),
		"handler_types": m.executor.ListHandlers(),
	}

	return status
}

// matchesFilters checks if a task matches the given filters
func (m *Manager) matchesFilters(task *Task, filters TaskFilters) bool {
	// Status filter
	if len(filters.Status) > 0 {
		matched := false
		for _, status := range filters.Status {
			if task.Status == status {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Type filter
	if filters.Type != "" && task.Type != filters.Type {
		return false
	}

	// Priority filter
	if filters.MinPriority > 0 && task.Priority < filters.MinPriority {
		return false
	}

	// Time filters
	if filters.CreatedAfter != nil && task.CreatedAt.Before(*filters.CreatedAfter) {
		return false
	}
	if filters.CreatedBefore != nil && task.CreatedAt.After(*filters.CreatedBefore) {
		return false
	}

	return true
}
