package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of TaskRepository for testing
type MockRepository struct {
	tasks       map[string]*Task
	results     map[string]*TaskResult
	metrics     map[string]*AgentTaskMetrics
	shouldError bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		tasks:   make(map[string]*Task),
		results: make(map[string]*TaskResult),
		metrics: make(map[string]*AgentTaskMetrics),
	}
}

func (r *MockRepository) StoreTask(ctx context.Context, task *Task) error {
	if r.shouldError {
		return assert.AnError
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *MockRepository) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if r.shouldError {
		return nil, assert.AnError
	}
	task, exists := r.tasks[taskID]
	if !exists {
		return nil, assert.AnError
	}
	return task, nil
}

func (r *MockRepository) UpdateTask(ctx context.Context, task *Task) error {
	if r.shouldError {
		return assert.AnError
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *MockRepository) ListTasks(ctx context.Context, agentID string, filters TaskFilters) ([]*Task, error) {
	if r.shouldError {
		return nil, assert.AnError
	}
	var tasks []*Task
	for _, task := range r.tasks {
		// Agent filter
		if agentID != "" && task.AgentID != agentID {
			continue
		}

		// Status filter
		if len(filters.Status) > 0 {
			matchStatus := false
			for _, status := range filters.Status {
				if task.Status == status {
					matchStatus = true
					break
				}
			}
			if !matchStatus {
				continue
			}
		}

		// Type filter
		if filters.Type != "" && task.Type != filters.Type {
			continue
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *MockRepository) StoreResult(ctx context.Context, result *TaskResult) error {
	if r.shouldError {
		return assert.AnError
	}
	r.results[result.TaskID] = result
	return nil
}

func (r *MockRepository) GetResult(ctx context.Context, taskID string) (*TaskResult, error) {
	if r.shouldError {
		return nil, assert.AnError
	}
	result, exists := r.results[taskID]
	if !exists {
		return nil, assert.AnError
	}
	return result, nil
}

func (r *MockRepository) GetMetrics(ctx context.Context, agentID string) (*AgentTaskMetrics, error) {
	if r.shouldError {
		return nil, assert.AnError
	}
	metrics, exists := r.metrics[agentID]
	if !exists {
		return &AgentTaskMetrics{
			AgentID:     agentID,
			TasksByType: make(map[string]int64),
		}, nil
	}
	return metrics, nil
}

func (r *MockRepository) UpdateMetrics(ctx context.Context, metrics *AgentTaskMetrics) error {
	if r.shouldError {
		return assert.AnError
	}
	r.metrics[metrics.AgentID] = metrics
	return nil
}

func (r *MockRepository) CleanupOldResults(ctx context.Context, before time.Time) (int, error) {
	if r.shouldError {
		return 0, assert.AnError
	}
	return 0, nil
}

func (r *MockRepository) SetShouldError(shouldError bool) {
	r.shouldError = shouldError
}

func TestManager_Submit(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
			MinWorkers: 1,
			QueueSize:  10,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
		PersistTasks: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	task := &Task{
		Type:    "echo",
		AgentID: "agent-1",
		Payload: map[string]interface{}{"test": "data"},
	}

	// Submit task
	err = manager.Submit(context.Background(), task)
	require.NoError(t, err)

	// Check task was assigned ID
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, TaskStatusQueued, task.Status) // Task gets queued immediately
	assert.False(t, task.CreatedAt.IsZero())

	// Check task was stored
	storedTask, exists := repo.tasks[task.ID]
	require.True(t, exists)
	assert.Equal(t, task.ID, storedTask.ID)
}

func TestManager_RegisterHandler(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)

	handler := NewMockHandler("custom")
	err := manager.RegisterHandler(handler)
	require.NoError(t, err)

	// Try to register duplicate
	err = manager.RegisterHandler(handler)
	assert.Error(t, err)
}

func TestManager_GetTask(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	task := &Task{
		ID:      "test-task",
		Type:    "echo",
		AgentID: "agent-1",
	}

	// Store task directly in repo
	repo.tasks[task.ID] = task

	// Get task
	retrieved, err := manager.GetTask(context.Background(), "test-task")
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Type, retrieved.Type)
}

func TestManager_CancelTask(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	task := &Task{
		ID:      "test-task",
		Type:    "echo",
		AgentID: "agent-1",
		Status:  TaskStatusPending,
	}

	// Store task in repo
	repo.tasks[task.ID] = task

	// Cancel task
	err = manager.CancelTask(context.Background(), "test-task")
	require.NoError(t, err)

	// Check task was marked as cancelled
	cancelled := repo.tasks[task.ID]
	assert.Equal(t, TaskStatusCancelled, cancelled.Status)
	assert.False(t, cancelled.CompletedAt.IsZero())
}

func TestManager_ListTasks(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)

	// Add some tasks to repo
	tasks := []*Task{
		{ID: "task-1", Type: "echo", AgentID: "agent-1", Status: TaskStatusCompleted},
		{ID: "task-2", Type: "delay", AgentID: "agent-1", Status: TaskStatusFailed},
		{ID: "task-3", Type: "echo", AgentID: "agent-2", Status: TaskStatusCompleted},
	}

	for _, task := range tasks {
		repo.tasks[task.ID] = task
	}

	// List all tasks
	filters := TaskFilters{}
	retrieved, err := manager.ListTasks(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// List by status
	filters.Status = []TaskStatus{TaskStatusCompleted}
	retrieved, err = manager.ListTasks(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// List by type
	filters = TaskFilters{Type: "echo"}
	retrieved, err = manager.ListTasks(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestManager_GetMetrics(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)

	// Add metrics to repo
	metrics := &AgentTaskMetrics{
		AgentID:        "agent-1",
		TotalTasks:     10,
		CompletedTasks: 8,
		FailedTasks:    2,
		TasksByType:    map[string]int64{"echo": 5, "delay": 5},
	}
	repo.metrics["agent-1"] = metrics

	// Get metrics
	retrieved, err := manager.GetMetrics(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(10), retrieved.TotalTasks)
	assert.Equal(t, int64(8), retrieved.CompletedTasks)
	assert.Equal(t, int64(2), retrieved.FailedTasks)
}

func TestManager_GetStatus(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	status := manager.GetStatus()
	assert.True(t, status["started"].(bool))
	assert.NotNil(t, status["queue_size"])
	assert.NotNil(t, status["handler_types"])

	handlerTypes := status["handler_types"].([]string)
	assert.Contains(t, handlerTypes, "echo")
	assert.Contains(t, handlerTypes, "http_request")
	assert.Contains(t, handlerTypes, "delay")
	assert.Contains(t, handlerTypes, "error")
}

func TestManager_NotStarted(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
	}

	manager := NewManager(config, repo)
	// Don't start manager

	task := &Task{
		Type:    "echo",
		AgentID: "agent-1",
	}

	// Try to submit task
	err := manager.Submit(context.Background(), task)
	assert.Error(t, err)
	assert.Equal(t, ErrManagerStopped, err)
}
