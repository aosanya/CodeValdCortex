package runtime

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/task"
)

// TaskManagerOptions holds configuration for task management setup
type TaskManagerOptions struct {
	// EnablePersistence enables task persistence to database
	EnablePersistence bool

	// Database connection for task persistence
	Database *database.ArangoClient
}

// SetupAgentTaskManager initializes advanced task management for an agent
func (m *Manager) SetupAgentTaskManager(agentID string, opts TaskManagerOptions) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Only setup if database is provided
	if opts.Database == nil {
		return fmt.Errorf("database connection required for task management")
	}

	// Setup task manager on the agent
	if err := a.SetupTaskManager(opts.Database); err != nil {
		return fmt.Errorf("failed to setup task manager: %w", err)
	}

	// Start task manager if agent is running
	if a.GetState() == agent.StateRunning {
		if err := a.StartTaskManager(); err != nil {
			return fmt.Errorf("failed to start task manager: %w", err)
		}
	}

	m.logger.WithField("agent_id", agentID).Info("Task manager setup completed")
	return nil
}

// StartAgentTaskManager starts the task manager for an agent
func (m *Manager) StartAgentTaskManager(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return a.StartTaskManager()
}

// StopAgentTaskManager stops the task manager for an agent
func (m *Manager) StopAgentTaskManager(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return a.StopTaskManager()
}

// SubmitTaskToAgent submits a task to an agent using the advanced task system
func (m *Manager) SubmitTaskToAgent(ctx context.Context, agentID string, taskReq *task.Task) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return a.SubmitAdvancedTask(ctx, taskReq)
}

// GetAgentTask retrieves a task from an agent
func (m *Manager) GetAgentTask(ctx context.Context, agentID, taskID string) (*task.Task, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return a.GetAdvancedTask(ctx, taskID)
}

// GetAgentTaskResult retrieves a task result from an agent
func (m *Manager) GetAgentTaskResult(ctx context.Context, agentID, taskID string) (*task.TaskResult, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return a.GetTaskResult(ctx, taskID)
}

// CancelAgentTask cancels a task on an agent
func (m *Manager) CancelAgentTask(ctx context.Context, agentID, taskID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return a.CancelAdvancedTask(ctx, taskID)
}

// ListAgentTasks lists tasks for an agent
func (m *Manager) ListAgentTasks(ctx context.Context, agentID string, filters task.TaskFilters) ([]*task.Task, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return a.ListAdvancedTasks(ctx, filters)
}

// GetAgentTaskMetrics retrieves task metrics for an agent
func (m *Manager) GetAgentTaskMetrics(ctx context.Context, agentID string) (*task.AgentTaskMetrics, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return a.GetTaskMetrics(ctx)
}

// RegisterTaskHandlerForAgent registers a task handler for an agent
func (m *Manager) RegisterTaskHandlerForAgent(agentID string, handler task.TaskHandler) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return a.RegisterTaskHandler(handler)
}

// GetAgentTaskManagerStatus returns the task manager status for an agent
func (m *Manager) GetAgentTaskManagerStatus(agentID string) map[string]interface{} {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return map[string]interface{}{
			"error": "agent not found",
		}
	}

	return a.GetTaskManagerStatus()
}

// SetupTaskManagementForAllAgents sets up task management for all existing agents
func (m *Manager) SetupTaskManagementForAllAgents(opts TaskManagerOptions) error {
	m.mu.RLock()
	agents := make([]*agent.Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	m.mu.RUnlock()

	errors := make([]error, 0)
	for _, a := range agents {
		if err := m.SetupAgentTaskManager(a.ID, opts); err != nil {
			errors = append(errors, fmt.Errorf("agent %s: %w", a.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to setup task management for some agents: %v", errors)
	}

	return nil
}
