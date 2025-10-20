package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	log "github.com/sirupsen/logrus"
)

// Manager handles agent lifecycle operations
type Manager struct {
	registry Repository
	mu       sync.RWMutex
	agents   map[string]*agent.Agent // active agents in memory
}

// NewManager creates a new lifecycle manager
func NewManager(repo Repository) *Manager {
	return &Manager{
		registry: repo,
		agents:   make(map[string]*agent.Agent),
	}
}

// CreateAgent creates a new agent and registers it
func (m *Manager) CreateAgent(ctx context.Context, name, agentType string, config agent.Config) (*agent.Agent, error) {
	// Create agent instance
	a := agent.New(name, agentType, config)

	// Validate state transition
	if err := m.validateTransition(agent.State(""), agent.StateCreated); err != nil {
		return nil, fmt.Errorf("invalid state transition: %w", err)
	}

	// Store in registry
	if err := m.registry.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	// Track in memory
	m.mu.Lock()
	m.agents[a.ID] = a
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"agent_id":   a.ID,
		"agent_name": a.Name,
		"agent_type": a.Type,
		"state":      a.State,
	}).Info("Agent created")

	return a, nil
}

// StartAgent transitions an agent from created/stopped/paused to running
func (m *Manager) StartAgent(ctx context.Context, agentID string) error {
	// Get agent
	a, err := m.getAgent(agentID)
	if err != nil {
		return err
	}

	// Check current state
	currentState := a.GetState()

	// Validate state transition
	if err := m.validateTransition(currentState, agent.StateRunning); err != nil {
		return fmt.Errorf("cannot start agent: %w", err)
	}

	// Start the agent's goroutine
	if err := m.startAgentRuntime(a); err != nil {
		return fmt.Errorf("failed to start agent runtime: %w", err)
	}

	// Update state
	a.SetState(agent.StateRunning)

	// Persist state change
	if err := m.registry.Update(ctx, a); err != nil {
		// Rollback - stop the runtime
		m.stopAgentRuntime(a)
		return fmt.Errorf("failed to persist state change: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"state":    agent.StateRunning,
	}).Info("Agent started")

	return nil
}

// StopAgent gracefully stops a running or paused agent
func (m *Manager) StopAgent(ctx context.Context, agentID string) error {
	// Get agent
	a, err := m.getAgent(agentID)
	if err != nil {
		return err
	}

	// Check current state
	currentState := a.GetState()

	// Validate state transition
	if err := m.validateTransition(currentState, agent.StateStopped); err != nil {
		return fmt.Errorf("cannot stop agent: %w", err)
	}

	// Stop the agent's runtime
	if err := m.stopAgentRuntime(a); err != nil {
		return fmt.Errorf("failed to stop agent runtime: %w", err)
	}

	// Update state
	a.SetState(agent.StateStopped)

	// Persist state change
	if err := m.registry.Update(ctx, a); err != nil {
		log.WithFields(log.Fields{
			"agent_id": agentID,
			"error":    err,
		}).Error("Failed to persist stopped state")
		// Don't return error - agent is already stopped
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"state":    agent.StateStopped,
	}).Info("Agent stopped")

	return nil
}

// PauseAgent pauses a running agent
func (m *Manager) PauseAgent(ctx context.Context, agentID string) error {
	// Get agent
	a, err := m.getAgent(agentID)
	if err != nil {
		return err
	}

	// Check current state
	currentState := a.GetState()

	// Validate state transition
	if err := m.validateTransition(currentState, agent.StatePaused); err != nil {
		return fmt.Errorf("cannot pause agent: %w", err)
	}

	// Pause the agent's runtime (stop accepting new tasks)
	if err := m.pauseAgentRuntime(a); err != nil {
		return fmt.Errorf("failed to pause agent runtime: %w", err)
	}

	// Update state
	a.SetState(agent.StatePaused)

	// Persist state change
	if err := m.registry.Update(ctx, a); err != nil {
		// Rollback - resume the runtime
		m.resumeAgentRuntime(a)
		return fmt.Errorf("failed to persist state change: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"state":    agent.StatePaused,
	}).Info("Agent paused")

	return nil
}

// ResumeAgent resumes a paused agent
func (m *Manager) ResumeAgent(ctx context.Context, agentID string) error {
	// Get agent
	a, err := m.getAgent(agentID)
	if err != nil {
		return err
	}

	// Check current state
	currentState := a.GetState()

	// Validate state transition
	if err := m.validateTransition(currentState, agent.StateRunning); err != nil {
		return fmt.Errorf("cannot resume agent: %w", err)
	}

	// Resume the agent's runtime
	if err := m.resumeAgentRuntime(a); err != nil {
		return fmt.Errorf("failed to resume agent runtime: %w", err)
	}

	// Update state
	a.SetState(agent.StateRunning)

	// Persist state change
	if err := m.registry.Update(ctx, a); err != nil {
		// Rollback - pause again
		m.pauseAgentRuntime(a)
		return fmt.Errorf("failed to persist state change: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"state":    agent.StateRunning,
	}).Info("Agent resumed")

	return nil
}

// RestartAgent stops and then starts an agent
func (m *Manager) RestartAgent(ctx context.Context, agentID string) error {
	// Stop the agent
	if err := m.StopAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to stop agent during restart: %w", err)
	}

	// Small delay to ensure clean shutdown
	time.Sleep(100 * time.Millisecond)

	// Start the agent
	if err := m.StartAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to start agent during restart: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
	}).Info("Agent restarted")

	return nil
}

// GetAgent retrieves an agent by ID
func (m *Manager) GetAgent(ctx context.Context, agentID string) (*agent.Agent, error) {
	// Try in-memory first
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if exists {
		return a, nil
	}

	// Load from database
	dbAgent, err := m.registry.Get(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Store in memory
	m.mu.Lock()
	m.agents[dbAgent.ID] = dbAgent
	m.mu.Unlock()

	return dbAgent, nil
}

// GetAgentStatus returns the current status of an agent
func (m *Manager) GetAgentStatus(ctx context.Context, agentID string) (*AgentStatus, error) {
	a, err := m.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return &AgentStatus{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		State:         a.GetState(),
		IsHealthy:     a.IsHealthy(),
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		LastHeartbeat: a.LastHeartbeat,
	}, nil
}

// ListAgents returns all agents
func (m *Manager) ListAgents(ctx context.Context) ([]*agent.Agent, error) {
	return m.registry.List(ctx)
}

// DeleteAgent removes an agent from the system
func (m *Manager) DeleteAgent(ctx context.Context, agentID string) error {
	// Get agent
	a, err := m.getAgent(agentID)
	if err != nil {
		return err
	}

	// Ensure agent is stopped
	if a.GetState() != agent.StateStopped {
		if err := m.StopAgent(ctx, agentID); err != nil {
			return fmt.Errorf("failed to stop agent before deletion: %w", err)
		}
	}

	// Remove from registry
	if err := m.registry.Delete(ctx, agentID); err != nil {
		return fmt.Errorf("failed to delete agent from registry: %w", err)
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.agents, agentID)
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"agent_id": agentID,
	}).Info("Agent deleted")

	return nil
}

// getAgent is a helper to retrieve agent (internal use)
func (m *Manager) getAgent(agentID string) (*agent.Agent, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %s not found in memory", agentID)
	}

	return a, nil
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	State         agent.State `json:"state"`
	IsHealthy     bool        `json:"is_healthy"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	LastHeartbeat time.Time   `json:"last_heartbeat"`
}
