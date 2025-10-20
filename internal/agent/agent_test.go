package agent_test

import (
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      100,
		HeartbeatInterval:  30 * time.Second,
		TaskTimeout:        5 * time.Minute,
	}

	a := agent.New("test-agent", "worker", config)

	assert.NotEmpty(t, a.ID)
	assert.Equal(t, "test-agent", a.Name)
	assert.Equal(t, "worker", a.Type)
	assert.Equal(t, agent.StateCreated, a.GetState())
	assert.NotNil(t, a.Metadata)
	assert.Equal(t, config.MaxConcurrentTasks, a.Config.MaxConcurrentTasks)
}

func TestNewWithDefaultConfig(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	// Check defaults are applied
	assert.Equal(t, 5, a.Config.MaxConcurrentTasks)
	assert.Equal(t, 100, a.Config.TaskQueueSize)
	assert.Equal(t, 30*time.Second, a.Config.HeartbeatInterval)
	assert.Equal(t, 5*time.Minute, a.Config.TaskTimeout)
}

func TestGetSetState(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	assert.Equal(t, agent.StateCreated, a.GetState())

	a.SetState(agent.StateRunning)
	assert.Equal(t, agent.StateRunning, a.GetState())

	a.SetState(agent.StatePaused)
	assert.Equal(t, agent.StatePaused, a.GetState())

	a.SetState(agent.StateStopped)
	assert.Equal(t, agent.StateStopped, a.GetState())

	a.SetState(agent.StateFailed)
	assert.Equal(t, agent.StateFailed, a.GetState())
}

func TestUpdateHeartbeat(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	initialHeartbeat := a.LastHeartbeat
	time.Sleep(10 * time.Millisecond)

	a.UpdateHeartbeat()
	assert.True(t, a.LastHeartbeat.After(initialHeartbeat))
}

func TestIsHealthy(t *testing.T) {
	config := agent.Config{
		HeartbeatInterval: 100 * time.Millisecond,
	}
	a := agent.New("test-agent", "worker", config)

	// Initially healthy (just created)
	assert.True(t, a.IsHealthy())

	// Wait longer than 2x heartbeat interval
	time.Sleep(250 * time.Millisecond)
	assert.False(t, a.IsHealthy())

	// Update heartbeat should make it healthy again
	a.UpdateHeartbeat()
	assert.True(t, a.IsHealthy())
}

func TestSubmitTask(t *testing.T) {
	config := agent.Config{
		TaskQueueSize: 5,
	}
	a := agent.New("test-agent", "worker", config)

	task := agent.Task{
		ID:        "task-1",
		Type:      "test",
		Payload:   "test payload",
		Priority:  1,
		Timeout:   1 * time.Minute,
		CreatedAt: time.Now().UTC(),
	}

	err := a.SubmitTask(task)
	assert.NoError(t, err)
}

func TestSubmitTaskQueueFull(t *testing.T) {
	config := agent.Config{
		TaskQueueSize: 2,
	}
	a := agent.New("test-agent", "worker", config)

	// Fill the queue
	for i := 0; i < 2; i++ {
		task := agent.Task{
			ID:        "task-" + string(rune(i)),
			Type:      "test",
			CreatedAt: time.Now().UTC(),
		}
		err := a.SubmitTask(task)
		require.NoError(t, err)
	}

	// Try to add one more (should fail)
	task := agent.Task{
		ID:        "task-overflow",
		Type:      "test",
		CreatedAt: time.Now().UTC(),
	}
	err := a.SubmitTask(task)
	assert.ErrorIs(t, err, agent.ErrTaskQueueFull)
}

func TestSubmitTaskAfterCancel(t *testing.T) {
	config := agent.Config{
		TaskQueueSize: 1,
	}
	a := agent.New("test-agent", "worker", config)

	// Fill the queue first
	task1 := agent.Task{
		ID:        "task-1",
		Type:      "test",
		CreatedAt: time.Now().UTC(),
	}
	err := a.SubmitTask(task1)
	require.NoError(t, err)

	// Cancel the agent context
	a.Cancel()

	// Now try to submit another task - should fail with ErrAgentStopped
	task2 := agent.Task{
		ID:        "task-2",
		Type:      "test",
		CreatedAt: time.Now().UTC(),
	}

	err = a.SubmitTask(task2)
	// Since the queue is full and context is cancelled,
	// we might get either error depending on select order
	assert.Error(t, err)
	assert.True(t, err == agent.ErrAgentStopped || err == agent.ErrTaskQueueFull)
}

func TestAgentContext(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	ctx := a.Context()
	assert.NotNil(t, ctx)

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done")
	default:
		// Context is not done, as expected
	}

	// Cancel the agent
	a.Cancel()

	// Context should now be done
	select {
	case <-ctx.Done():
		// Context is done, as expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be done after cancel")
	}
}

func TestConcurrentStateAccess(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	// Test concurrent state reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				a.SetState(agent.StateRunning)
				_ = a.GetState()
				a.UpdateHeartbeat()
				_ = a.IsHealthy()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic (race detector will catch issues)
}

func TestMetadata(t *testing.T) {
	a := agent.New("test-agent", "worker", agent.Config{})

	assert.NotNil(t, a.Metadata)
	assert.Empty(t, a.Metadata)

	a.Metadata["key1"] = "value1"
	a.Metadata["key2"] = "value2"

	assert.Equal(t, "value1", a.Metadata["key1"])
	assert.Equal(t, "value2", a.Metadata["key2"])
}
