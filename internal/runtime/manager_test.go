package runtime_test

import (
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := runtime.ManagerConfig{
		MaxAgents:           10,
		HealthCheckInterval: 100 * time.Millisecond,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	manager := runtime.NewManager(logger, config)
	assert.NotNil(t, manager)

	// Clean shutdown
	err := manager.Shutdown()
	assert.NoError(t, err)
}

func TestCreateAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{
		MaxAgents: 10,
	})
	defer manager.Shutdown()

	agentConfig := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      100,
	}

	a, err := manager.CreateAgent("test-agent", "worker", agentConfig)
	require.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, "test-agent", a.Name)
	assert.Equal(t, "worker", a.Type)
	assert.Equal(t, agent.StateCreated, a.GetState())
}

func TestCreateAgentLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{
		MaxAgents: 2,
	})
	defer manager.Shutdown()

	agentConfig := agent.Config{}

	// Create 2 agents (max)
	_, err := manager.CreateAgent("agent-1", "worker", agentConfig)
	require.NoError(t, err)

	_, err = manager.CreateAgent("agent-2", "worker", agentConfig)
	require.NoError(t, err)

	// Try to create a 3rd agent (should fail)
	_, err = manager.CreateAgent("agent-3", "worker", agentConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent limit reached")
}

func TestStartAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	a, err := manager.CreateAgent("test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	err = manager.StartAgent(a.ID)
	assert.NoError(t, err)
	assert.Equal(t, agent.StateRunning, a.GetState())
}

func TestStartNonExistentAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	err := manager.StartAgent("non-existent-id")
	assert.ErrorIs(t, err, agent.ErrAgentNotFound)
}

func TestStopAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	a, err := manager.CreateAgent("test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	err = manager.StartAgent(a.ID)
	require.NoError(t, err)

	err = manager.StopAgent(a.ID)
	assert.NoError(t, err)
	assert.Equal(t, agent.StateStopped, a.GetState())
}

func TestGetAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	a1, err := manager.CreateAgent("test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	a2, err := manager.GetAgent(a1.ID)
	require.NoError(t, err)
	assert.Equal(t, a1.ID, a2.ID)
	assert.Equal(t, a1.Name, a2.Name)
}

func TestGetNonExistentAgent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	_, err := manager.GetAgent("non-existent-id")
	assert.ErrorIs(t, err, agent.ErrAgentNotFound)
}

func TestListAgents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	// Initially empty
	agents := manager.ListAgents()
	assert.Empty(t, agents)

	// Create some agents
	_, err := manager.CreateAgent("agent-1", "worker", agent.Config{})
	require.NoError(t, err)

	_, err = manager.CreateAgent("agent-2", "coordinator", agent.Config{})
	require.NoError(t, err)

	agents = manager.ListAgents()
	assert.Len(t, agents, 2)
}

func TestGetMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	metrics := manager.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalAgentsCreated)
	assert.Equal(t, int64(0), metrics.CurrentActiveAgents)

	// Create an agent
	a, err := manager.CreateAgent("test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	metrics = manager.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalAgentsCreated)
	assert.Equal(t, int64(1), metrics.CurrentActiveAgents)

	// Stop the agent
	err = manager.StopAgent(a.ID)
	require.NoError(t, err)

	metrics = manager.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalAgentsStopped)
	assert.Equal(t, int64(0), metrics.CurrentActiveAgents)
}

func TestAgentTaskExecution(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{})
	defer manager.Shutdown()

	a, err := manager.CreateAgent("test-agent", "worker", agent.Config{
		MaxConcurrentTasks: 2,
		TaskQueueSize:      10,
	})
	require.NoError(t, err)

	err = manager.StartAgent(a.ID)
	require.NoError(t, err)

	// Submit tasks
	for i := 0; i < 5; i++ {
		task := agent.Task{
			ID:        "task-" + string(rune(i)),
			Type:      "test",
			Payload:   "test data",
			CreatedAt: time.Now().UTC(),
		}
		err = a.SubmitTask(task)
		require.NoError(t, err)
	}

	// Wait a bit for tasks to be processed
	time.Sleep(600 * time.Millisecond)

	// Check metrics
	metrics := manager.GetMetrics()
	assert.Greater(t, metrics.TotalTasksExecuted, int64(0))
}

func TestShutdown(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{
		ShutdownTimeout: 5 * time.Second,
	})

	// Create and start some agents
	a1, err := manager.CreateAgent("agent-1", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(a1.ID)
	require.NoError(t, err)

	a2, err := manager.CreateAgent("agent-2", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(a2.ID)
	require.NoError(t, err)

	// Shutdown should gracefully stop all agents
	err = manager.Shutdown()
	assert.NoError(t, err)

	// All agents should be stopped
	assert.Equal(t, agent.StateStopped, a1.GetState())
	assert.Equal(t, agent.StateStopped, a2.GetState())
}

func TestConcurrentOperations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := runtime.NewManager(logger, runtime.ManagerConfig{
		MaxAgents: 50,
	})
	defer manager.Shutdown()

	// Create agents concurrently
	done := make(chan string, 20)
	for i := 0; i < 20; i++ {
		go func(id int) {
			a, err := manager.CreateAgent("agent-"+string(rune(id)), "worker", agent.Config{})
			if err != nil {
				done <- ""
				return
			}
			done <- a.ID
		}(i)
	}

	// Collect agent IDs
	agentIDs := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		id := <-done
		if id != "" {
			agentIDs = append(agentIDs, id)
		}
	}

	assert.Greater(t, len(agentIDs), 0)

	// Start and stop agents concurrently
	for _, id := range agentIDs {
		go func(agentID string) {
			_ = manager.StartAgent(agentID)
		}(id)
	}

	time.Sleep(100 * time.Millisecond)

	agents := manager.ListAgents()
	assert.Equal(t, len(agentIDs), len(agents))
}
