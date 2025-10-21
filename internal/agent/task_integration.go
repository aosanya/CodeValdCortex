package agent

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/task"
)

// TaskManager provides advanced task management capabilities
// Note: This is defined here to avoid circular imports
type TaskManager struct {
	agent   *Agent
	manager *task.Manager
	repo    *task.Repository
}

// SetupTaskManager initializes advanced task management for the agent
func (a *Agent) SetupTaskManager(db *database.ArangoClient) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create task repository
	repo, err := task.NewRepository(db.Database())
	if err != nil {
		return fmt.Errorf("failed to create task repository: %w", err)
	}

	// Create task manager with default configuration
	config := task.ManagerConfig{
		WorkerPool: task.WorkerPoolConfig{
			MaxWorkers:  a.Config.MaxConcurrentTasks,
			MinWorkers:  1,
			IdleTimeout: a.Config.TaskTimeout,
			QueueSize:   a.Config.TaskQueueSize,
		},
		Executor: task.ExecutorConfig{
			DefaultTimeout:     a.Config.TaskTimeout,
			MaxTimeout:         a.Config.TaskTimeout * 2,
			DefaultRetryPolicy: task.DefaultRetryPolicy(),
			MetricsEnabled:     true,
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	taskManager := task.NewManager(config, repo)

	// Store in agent
	a.taskManager = &TaskManager{
		agent:   a,
		manager: taskManager,
		repo:    repo,
	}

	return nil
}

// StartTaskManager starts the advanced task management system
func (a *Agent) StartTaskManager() error {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return fmt.Errorf("task manager not initialized - call SetupTaskManager first")
	}

	return taskManager.manager.Start()
}

// StopTaskManager stops the advanced task management system
func (a *Agent) StopTaskManager() error {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return nil
	}

	return taskManager.manager.Stop()
}

// SubmitAdvancedTask submits a task using the advanced task system
func (a *Agent) SubmitAdvancedTask(ctx context.Context, taskReq *task.Task) error {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	// Set agent ID
	taskReq.AgentID = a.ID

	return taskManager.manager.Submit(ctx, taskReq)
}

// RegisterTaskHandler registers a custom task handler
func (a *Agent) RegisterTaskHandler(handler task.TaskHandler) error {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.RegisterHandler(handler)
}

// GetAdvancedTask retrieves a task by ID
func (a *Agent) GetAdvancedTask(ctx context.Context, taskID string) (*task.Task, error) {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return nil, fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.GetTask(ctx, taskID)
}

// GetTaskResult retrieves task execution result
func (a *Agent) GetTaskResult(ctx context.Context, taskID string) (*task.TaskResult, error) {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return nil, fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.GetTaskResult(ctx, taskID)
}

// CancelAdvancedTask cancels a task
func (a *Agent) CancelAdvancedTask(ctx context.Context, taskID string) error {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.CancelTask(ctx, taskID)
}

// ListAdvancedTasks lists tasks with filters
func (a *Agent) ListAdvancedTasks(ctx context.Context, filters task.TaskFilters) ([]*task.Task, error) {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return nil, fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.ListTasks(ctx, filters)
}

// GetTaskMetrics retrieves task execution metrics
func (a *Agent) GetTaskMetrics(ctx context.Context) (*task.AgentTaskMetrics, error) {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return nil, fmt.Errorf("task manager not initialized")
	}

	return taskManager.manager.GetMetrics(ctx, a.ID)
}

// GetTaskManagerStatus returns the status of the task manager
func (a *Agent) GetTaskManagerStatus() map[string]interface{} {
	a.mu.RLock()
	taskManager := a.taskManager
	a.mu.RUnlock()

	if taskManager == nil {
		return map[string]interface{}{
			"initialized": false,
		}
	}

	status := taskManager.manager.GetStatus()
	status["initialized"] = true
	return status
}
