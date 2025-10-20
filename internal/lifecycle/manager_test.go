package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of registry.Repository for testing
type MockRepository struct {
	agents map[string]*agent.Agent
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		agents: make(map[string]*agent.Agent),
	}
}

func (m *MockRepository) Create(ctx context.Context, a *agent.Agent) error {
	m.agents[a.ID] = a
	return nil
}

func (m *MockRepository) Get(ctx context.Context, id string) (*agent.Agent, error) {
	a, exists := m.agents[id]
	if !exists {
		return nil, assert.AnError
	}
	return a, nil
}

func (m *MockRepository) Update(ctx context.Context, a *agent.Agent) error {
	m.agents[a.ID] = a
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	delete(m.agents, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context) ([]*agent.Agent, error) {
	agents := make([]*agent.Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	return agents, nil
}

func (m *MockRepository) Count(ctx context.Context) (int, error) {
	return len(m.agents), nil
}

func (m *MockRepository) FindByType(ctx context.Context, agentType string) ([]*agent.Agent, error) {
	agents := make([]*agent.Agent, 0)
	for _, a := range m.agents {
		if a.Type == agentType {
			agents = append(agents, a)
		}
	}
	return agents, nil
}

func (m *MockRepository) FindByState(ctx context.Context, state string) ([]*agent.Agent, error) {
	agents := make([]*agent.Agent, 0)
	for _, a := range m.agents {
		if string(a.GetState()) == state {
			agents = append(agents, a)
		}
	}
	return agents, nil
}

func (m *MockRepository) FindHealthy(ctx context.Context) ([]*agent.Agent, error) {
	agents := make([]*agent.Agent, 0)
	for _, a := range m.agents {
		if a.IsHealthy() {
			agents = append(agents, a)
		}
	}
	return agents, nil
}

func (m *MockRepository) FindByTypeAndState(ctx context.Context, agentType, state string) ([]*agent.Agent, error) {
	agents := make([]*agent.Agent, 0)
	for _, a := range m.agents {
		if a.Type == agentType && string(a.GetState()) == state {
			agents = append(agents, a)
		}
	}
	return agents, nil
}

// TestCreateAgent tests agent creation
func TestCreateAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	config := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      100,
		HeartbeatInterval:  30 * time.Second,
	}

	a, err := manager.CreateAgent(ctx, "test-agent", "worker", config)
	require.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, "test-agent", a.Name)
	assert.Equal(t, "worker", a.Type)
	assert.Equal(t, agent.StateCreated, a.GetState())
}

// TestStartAgent tests starting an agent
func TestStartAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	// Start the agent
	err = manager.StartAgent(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, a.GetState())
}

// TestStopAgent tests stopping an agent
func TestStopAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create and start an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(ctx, a.ID)
	require.NoError(t, err)

	// Stop the agent
	err = manager.StopAgent(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateStopped, a.GetState())
}

// TestPauseAgent tests pausing an agent
func TestPauseAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create and start an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(ctx, a.ID)
	require.NoError(t, err)

	// Pause the agent
	err = manager.PauseAgent(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StatePaused, a.GetState())
}

// TestResumeAgent tests resuming a paused agent
func TestResumeAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create, start, and pause an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(ctx, a.ID)
	require.NoError(t, err)
	err = manager.PauseAgent(ctx, a.ID)
	require.NoError(t, err)

	// Resume the agent
	err = manager.ResumeAgent(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, a.GetState())
}

// TestRestartAgent tests restarting an agent
func TestRestartAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create and start an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)
	err = manager.StartAgent(ctx, a.ID)
	require.NoError(t, err)

	// Restart the agent
	err = manager.RestartAgent(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, a.GetState())
}

// TestDeleteAgent tests agent deletion
func TestDeleteAgent(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	// Delete the agent
	err = manager.DeleteAgent(ctx, a.ID)
	require.NoError(t, err)

	// Verify agent is deleted
	_, err = manager.GetAgent(ctx, a.ID)
	assert.Error(t, err)
}

// TestInvalidStateTransitions tests that invalid transitions are rejected
func TestInvalidStateTransitions(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	// Try to pause a created agent (should fail)
	err = manager.PauseAgent(ctx, a.ID)
	assert.Error(t, err)
	assert.IsType(t, &StateTransitionError{}, err)

	// Try to resume a created agent (should fail)
	err = manager.ResumeAgent(ctx, a.ID)
	assert.Error(t, err)
	assert.IsType(t, &StateTransitionError{}, err)
}

// TestListAgents tests listing all agents
func TestListAgents(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create multiple agents
	_, err := manager.CreateAgent(ctx, "agent1", "worker", agent.Config{})
	require.NoError(t, err)
	_, err = manager.CreateAgent(ctx, "agent2", "coordinator", agent.Config{})
	require.NoError(t, err)

	// List all agents
	agents, err := manager.ListAgents(ctx)
	require.NoError(t, err)
	assert.Len(t, agents, 2)
}

// TestGetAgentStatus tests retrieving agent status
func TestGetAgentStatus(t *testing.T) {
	repo := NewMockRepository()
	manager := NewManager(repo)
	ctx := context.Background()

	// Create an agent
	a, err := manager.CreateAgent(ctx, "test-agent", "worker", agent.Config{})
	require.NoError(t, err)

	// Get status
	status, err := manager.GetAgentStatus(ctx, a.ID)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, a.ID, status.ID)
	assert.Equal(t, "test-agent", status.Name)
	assert.Equal(t, "worker", status.Type)
	assert.Equal(t, agent.StateCreated, status.State)
}
