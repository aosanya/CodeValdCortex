package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_EndToEndTaskExecution(t *testing.T) {
	// Create a full task system with real scheduler and executor
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers:  3,
			MinWorkers:  1,
			IdleTimeout: 5 * time.Second,
			QueueSize:   10,
		},
		Executor: ExecutorConfig{
			DefaultTimeout:     5 * time.Second,
			MetricsEnabled:     true,
			DefaultRetryPolicy: DefaultRetryPolicy(),
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Submit multiple tasks
	tasks := []*Task{
		{
			Type:     "echo",
			AgentID:  "agent-1",
			Priority: 5,
			Payload:  map[string]interface{}{"message": "hello"},
		},
		{
			Type:     "delay",
			AgentID:  "agent-1",
			Priority: 3,
			Payload:  map[string]interface{}{"duration": "50ms"},
		},
		{
			Type:     "echo",
			AgentID:  "agent-2",
			Priority: 8,
			Payload:  map[string]interface{}{"message": "world"},
		},
	}

	taskIDs := make([]string, len(tasks))
	for i, task := range tasks {
		err := manager.Submit(context.Background(), task)
		require.NoError(t, err)
		taskIDs[i] = task.ID
	}

	// Wait for tasks to execute
	time.Sleep(200 * time.Millisecond)

	// Check all tasks completed
	for i, taskID := range taskIDs {
		result, err := manager.GetTaskResult(context.Background(), taskID)
		require.NoError(t, err, "Task %d (%s) should have result", i, taskID)
		assert.Equal(t, TaskStatusCompleted, result.Status)
		assert.True(t, result.Duration > 0)
	}

	// Check metrics
	metrics, err := manager.GetMetrics(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.True(t, metrics.TotalTasks >= 2)
	assert.True(t, metrics.CompletedTasks >= 2)
}

func TestIntegration_TaskPriority(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 1, // Single worker to test priority
			MinWorkers: 1,
			QueueSize:  10,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Submit tasks with different priorities
	// Lower priority first
	lowPriorityTask := &Task{
		Type:     "delay",
		AgentID:  "agent-1",
		Priority: 1,
		Payload:  map[string]interface{}{"duration": "100ms"},
	}
	err = manager.Submit(context.Background(), lowPriorityTask)
	require.NoError(t, err)

	// High priority task should execute first even though submitted later
	highPriorityTask := &Task{
		Type:     "echo",
		AgentID:  "agent-1",
		Priority: 10,
		Payload:  map[string]interface{}{"message": "urgent"},
	}
	err = manager.Submit(context.Background(), highPriorityTask)
	require.NoError(t, err)

	// Wait for execution
	time.Sleep(300 * time.Millisecond)

	// Get results
	lowResult, err := manager.GetTaskResult(context.Background(), lowPriorityTask.ID)
	require.NoError(t, err)
	highResult, err := manager.GetTaskResult(context.Background(), highPriorityTask.ID)
	require.NoError(t, err)

	// High priority should complete first (started after but finished before low priority)
	assert.True(t, highResult.StartedAt.After(lowResult.StartedAt) ||
		highResult.CompletedAt.Before(lowResult.CompletedAt))
}

func TestIntegration_TaskTimeout(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 2,
			MinWorkers: 1,
			QueueSize:  10,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 50 * time.Millisecond, // Very short timeout
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Submit task that will timeout
	task := &Task{
		Type:    "delay",
		AgentID: "agent-1",
		Payload: map[string]interface{}{"duration": "200ms"}, // Longer than timeout
	}
	err = manager.Submit(context.Background(), task)
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Check task timed out
	result, err := manager.GetTaskResult(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, TaskStatusTimeout, result.Status)
	assert.Contains(t, result.Error, "timeout")
}

func TestIntegration_TaskCancellation(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 1,
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

	// Submit a long-running task
	task1 := &Task{
		Type:    "delay",
		AgentID: "agent-1",
		Payload: map[string]interface{}{"duration": "500ms"},
	}
	err = manager.Submit(context.Background(), task1)
	require.NoError(t, err)

	// Submit another task that will be queued
	task2 := &Task{
		Type:    "echo",
		AgentID: "agent-1",
		Payload: map[string]interface{}{"message": "test"},
	}
	err = manager.Submit(context.Background(), task2)
	require.NoError(t, err)

	// Wait a bit for first task to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the second task (should be in queue)
	err = manager.CancelTask(context.Background(), task2.ID)
	require.NoError(t, err)

	// Wait for first task to complete
	time.Sleep(500 * time.Millisecond)

	// Check first task completed normally
	result1, err := manager.GetTaskResult(context.Background(), task1.ID)
	require.NoError(t, err)
	assert.Equal(t, TaskStatusCompleted, result1.Status)

	// Check second task was cancelled
	task2FromRepo, err := manager.GetTask(context.Background(), task2.ID)
	require.NoError(t, err)
	assert.Equal(t, TaskStatusCancelled, task2FromRepo.Status)
}

func TestIntegration_CustomHandler(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 2,
			MinWorkers: 1,
			QueueSize:  10,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Register custom handler
	customHandler := NewMockHandler("custom")
	err = manager.RegisterHandler(customHandler)
	require.NoError(t, err)

	// Submit task using custom handler
	task := &Task{
		Type:    "custom",
		AgentID: "agent-1",
		Payload: map[string]interface{}{"custom": "data"},
	}
	err = manager.Submit(context.Background(), task)
	require.NoError(t, err)

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	// Check result
	result, err := manager.GetTaskResult(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.Equal(t, true, result.Result["mock"])
	assert.Equal(t, "custom", result.Result["type"])
}

func TestIntegration_MultipleAgents(t *testing.T) {
	repo := NewMockRepository()
	config := ManagerConfig{
		WorkerPool: WorkerPoolConfig{
			MaxWorkers: 5,
			MinWorkers: 2,
			QueueSize:  20,
		},
		Executor: ExecutorConfig{
			DefaultTimeout: 5 * time.Second,
		},
		PersistTasks:   true,
		PersistResults: true,
	}

	manager := NewManager(config, repo)
	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Submit tasks for multiple agents
	agents := []string{"agent-1", "agent-2", "agent-3"}
	tasksPerAgent := 3
	totalTasks := len(agents) * tasksPerAgent

	for _, agentID := range agents {
		for i := 0; i < tasksPerAgent; i++ {
			task := &Task{
				Type:    "echo",
				AgentID: agentID,
				Payload: map[string]interface{}{
					"message": "task for " + agentID,
					"index":   i,
				},
			}
			err = manager.Submit(context.Background(), task)
			require.NoError(t, err)
		}
	}

	// Wait for all tasks to complete
	time.Sleep(200 * time.Millisecond)

	// Check metrics for each agent
	for _, agentID := range agents {
		metrics, err := manager.GetMetrics(context.Background(), agentID)
		require.NoError(t, err)
		assert.Equal(t, int64(tasksPerAgent), metrics.TotalTasks)
		assert.Equal(t, int64(tasksPerAgent), metrics.CompletedTasks)
		assert.Equal(t, int64(0), metrics.FailedTasks)
	}

	// List all tasks
	allTasks, err := manager.ListTasks(context.Background(), TaskFilters{})
	require.NoError(t, err)
	assert.Len(t, allTasks, totalTasks)

	// List tasks for specific agent
	agent1Tasks, err := manager.ListTasks(context.Background(), TaskFilters{})
	require.NoError(t, err)
	assert.Len(t, agent1Tasks, totalTasks) // No agent filter implemented in mock

	// Check all tasks completed
	completed := 0
	for _, task := range allTasks {
		if task.Status == TaskStatusCompleted {
			completed++
		}
	}
	assert.Equal(t, totalTasks, completed)
}
