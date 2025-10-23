package runtime

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// Manager manages the lifecycle of all agents in the system
type Manager struct {
	// agents maps agent ID to agent instance (in-memory cache)
	agents map[string]*agent.Agent

	// registry provides persistent storage for agents
	registry *registry.Repository

	// mu protects concurrent access to agents map
	mu sync.RWMutex

	// ctx is the manager's context for shutdown
	ctx context.Context

	// cancel is the function to cancel manager context
	cancel context.CancelFunc

	// wg tracks running agent goroutines
	wg sync.WaitGroup

	// logger for manager operations
	logger *logrus.Logger

	// config holds manager configuration
	config ManagerConfig

	// metrics tracks runtime metrics
	metrics *metricsHolder
}

// ManagerConfig holds runtime manager configuration
type ManagerConfig struct {
	// MaxAgents limits the total number of agents
	MaxAgents int

	// HealthCheckInterval defines how often to check agent health
	HealthCheckInterval time.Duration

	// ShutdownTimeout defines grace period for shutdown
	ShutdownTimeout time.Duration

	// EnableMetrics toggles metrics collection
	EnableMetrics bool
}

// Metrics tracks runtime statistics
type Metrics struct {
	TotalAgentsCreated  int64 `json:"total_agents_created"`
	TotalAgentsStopped  int64 `json:"total_agents_stopped"`
	TotalTasksExecuted  int64 `json:"total_tasks_executed"`
	TotalTasksFailed    int64 `json:"total_tasks_failed"`
	CurrentActiveAgents int64 `json:"current_active_agents"`
	CurrentRunningTasks int64 `json:"current_running_tasks"`
}

// metricsHolder wraps Metrics with a mutex for thread-safe access
type metricsHolder struct {
	mu      sync.RWMutex
	metrics Metrics
}

// NewManager creates a new runtime manager
func NewManager(logger *logrus.Logger, config ManagerConfig, reg *registry.Repository) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default configuration
	if config.MaxAgents == 0 {
		config.MaxAgents = 100
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	m := &Manager{
		agents:   make(map[string]*agent.Agent),
		registry: reg,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		config:   config,
		metrics:  &metricsHolder{},
	}

	// Load existing agents from registry into memory cache
	if reg != nil {
		if err := m.loadAgentsFromRegistry(); err != nil {
			logger.WithError(err).Warn("Failed to load agents from registry")
		}
	}

	// Start health check loop
	m.wg.Add(1)
	go m.healthCheckLoop()

	return m
}

// loadAgentsFromRegistry loads all agents from persistent storage into memory cache
func (m *Manager) loadAgentsFromRegistry() error {
	if m.registry == nil {
		return nil
	}

	agents, err := m.registry.List(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to list agents from registry: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, a := range agents {
		m.agents[a.ID] = a
		m.logger.WithFields(logrus.Fields{
			"agent_id":   a.ID,
			"agent_name": a.Name,
			"state":      a.State,
		}).Debug("Loaded agent from registry")
	}

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.metrics.CurrentActiveAgents = int64(len(agents))
	m.metrics.mu.Unlock()

	m.logger.WithField("count", len(agents)).Info("Loaded agents from registry")
	return nil
}

// CreateAgent creates and registers a new agent
func (m *Manager) CreateAgent(name, agentType string, config agent.Config) (*agent.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check agent limit
	if len(m.agents) >= m.config.MaxAgents {
		return nil, fmt.Errorf("agent limit reached: %d", m.config.MaxAgents)
	}

	// Create new agent
	a := agent.New(name, agentType, config)

	// Persist to registry if available
	if m.registry != nil {
		if err := m.registry.Create(m.ctx, a); err != nil {
			return nil, fmt.Errorf("failed to persist agent to registry: %w", err)
		}
	}

	// Add to in-memory cache
	m.agents[a.ID] = a

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.metrics.TotalAgentsCreated++
	m.metrics.metrics.CurrentActiveAgents++
	m.metrics.mu.Unlock()

	m.logger.WithFields(logrus.Fields{
		"agent_id":   a.ID,
		"agent_name": name,
		"agent_type": agentType,
	}).Info("Agent created")

	return a, nil
}

// StartAgent starts an agent and begins processing tasks
func (m *Manager) StartAgent(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return agent.ErrAgentNotFound
	}

	// Check current state
	currentState := a.GetState()
	if currentState != agent.StateCreated && currentState != agent.StatePaused {
		return fmt.Errorf("cannot start agent in state: %s", currentState)
	}

	// Update state to running
	a.SetState(agent.StateRunning)

	// Persist state change to registry
	if m.registry != nil {
		if err := m.registry.Update(m.ctx, a); err != nil {
			m.logger.WithError(err).Warn("Failed to persist agent state to registry")
		}
	}

	// Start agent worker goroutines
	m.wg.Add(1)
	go m.runAgent(a)

	m.logger.WithField("agent_id", agentID).Info("Agent started")

	return nil
}

// runAgent is the main execution loop for an agent
func (m *Manager) runAgent(a *agent.Agent) {
	defer m.wg.Done()

	// Create worker pool based on agent config
	workerCount := a.Config.MaxConcurrentTasks
	taskResults := make(chan *agent.TaskResult, workerCount*2)

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		m.wg.Add(1)
		go m.agentWorker(a, i, taskResults)
	}

	// Start heartbeat goroutine
	m.wg.Add(1)
	go m.agentHeartbeat(a)

	// Process task results
	for {
		select {
		case <-m.ctx.Done():
			m.logger.WithField("agent_id", a.ID).Info("Manager shutting down, stopping agent")
			a.SetState(agent.StateStopped)
			return

		case <-a.Context().Done():
			m.logger.WithField("agent_id", a.ID).Info("Agent context cancelled, stopping")
			a.SetState(agent.StateStopped)
			return

		case err := <-a.Errors():
			m.logger.WithFields(logrus.Fields{
				"agent_id": a.ID,
				"error":    err,
			}).Error("Agent error")
			a.SetState(agent.StateFailed)
			return

		case result := <-taskResults:
			m.handleTaskResult(a, result)
		}
	}
}

// agentWorker processes tasks from the agent's task channel
func (m *Manager) agentWorker(a *agent.Agent, workerID int, results chan<- *agent.TaskResult) {
	defer m.wg.Done()

	m.logger.WithFields(logrus.Fields{
		"agent_id":  a.ID,
		"worker_id": workerID,
	}).Debug("Agent worker started")

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-a.Context().Done():
			return
		case task := <-a.TaskChan():
			result := m.executeTask(a, task)
			results <- result
		}
	}
}

// executeTask executes a single task
func (m *Manager) executeTask(a *agent.Agent, task agent.Task) *agent.TaskResult {
	result := &agent.TaskResult{
		TaskID:    task.ID,
		AgentID:   a.ID,
		StartedAt: time.Now().UTC(),
	}

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.metrics.CurrentRunningTasks++
	m.metrics.mu.Unlock()

	defer func() {
		result.CompletedAt = time.Now().UTC()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)

		m.metrics.mu.Lock()
		m.metrics.metrics.CurrentRunningTasks--
		if result.Success {
			m.metrics.metrics.TotalTasksExecuted++
		} else {
			m.metrics.metrics.TotalTasksFailed++
		}
		m.metrics.mu.Unlock()
	}()

	// Create task timeout context
	timeout := task.Timeout
	if timeout == 0 {
		timeout = a.Config.TaskTimeout
	}

	ctx, cancel := context.WithTimeout(a.Context(), timeout)
	defer cancel()

	// Execute task with timeout
	done := make(chan struct{})
	go func() {
		// TODO: Implement actual task execution logic
		// For now, simulate work
		time.Sleep(100 * time.Millisecond)
		result.Success = true
		result.Result = "Task completed successfully"
		close(done)
	}()

	select {
	case <-done:
		// Task completed successfully
	case <-ctx.Done():
		result.Success = false
		result.Error = agent.ErrTaskTimeout
	}

	return result
}

// agentHeartbeat maintains agent health status
func (m *Manager) agentHeartbeat(a *agent.Agent) {
	defer m.wg.Done()

	ticker := time.NewTicker(a.Config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-a.Context().Done():
			return
		case <-ticker.C:
			a.UpdateHeartbeat()
		}
	}
}

// handleTaskResult processes task execution results
func (m *Manager) handleTaskResult(a *agent.Agent, result *agent.TaskResult) {
	if result.Success {
		m.logger.WithFields(logrus.Fields{
			"agent_id": a.ID,
			"task_id":  result.TaskID,
			"duration": result.Duration,
		}).Debug("Task completed successfully")
	} else {
		m.logger.WithFields(logrus.Fields{
			"agent_id": a.ID,
			"task_id":  result.TaskID,
			"error":    result.Error,
			"duration": result.Duration,
		}).Error("Task failed")
	}

	// TODO: Send result to result storage or event bus
}

// StopAgent gracefully stops an agent
func (m *Manager) StopAgent(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return agent.ErrAgentNotFound
	}

	// Cancel agent context
	a.Cancel()

	// Update state
	a.SetState(agent.StateStopped)

	// Persist state change to registry
	if m.registry != nil {
		if err := m.registry.Update(m.ctx, a); err != nil {
			m.logger.WithError(err).Warn("Failed to persist agent state to registry")
		}
	}

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.metrics.TotalAgentsStopped++
	m.metrics.metrics.CurrentActiveAgents--
	m.metrics.mu.Unlock()

	m.logger.WithField("agent_id", agentID).Info("Agent stopped")

	return nil
}

// PauseAgent pauses a running agent
func (m *Manager) PauseAgent(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return agent.ErrAgentNotFound
	}

	// Check current state
	currentState := a.GetState()
	if currentState != agent.StateRunning {
		return fmt.Errorf("cannot pause agent in state: %s", currentState)
	}

	// Update state to paused
	a.SetState(agent.StatePaused)

	// Persist state change to registry
	if m.registry != nil {
		if err := m.registry.Update(m.ctx, a); err != nil {
			m.logger.WithError(err).Warn("Failed to persist agent state to registry")
		}
	}

	m.logger.WithField("agent_id", agentID).Info("Agent paused")

	return nil
}

// ResumeAgent resumes a paused agent
func (m *Manager) ResumeAgent(agentID string) error {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return agent.ErrAgentNotFound
	}

	// Check current state
	currentState := a.GetState()
	if currentState != agent.StatePaused {
		return fmt.Errorf("cannot resume agent in state: %s", currentState)
	}

	// Update state to running
	a.SetState(agent.StateRunning)

	// Persist state change to registry
	if m.registry != nil {
		if err := m.registry.Update(m.ctx, a); err != nil {
			m.logger.WithError(err).Warn("Failed to persist agent state to registry")
		}
	}

	m.logger.WithField("agent_id", agentID).Info("Agent resumed")

	return nil
}

// RestartAgent stops and restarts an agent
func (m *Manager) RestartAgent(agentID string) error {
	// Stop the agent
	if err := m.StopAgent(agentID); err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}

	// Brief pause to ensure cleanup
	time.Sleep(100 * time.Millisecond)

	// Start the agent
	if err := m.StartAgent(agentID); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	m.logger.WithField("agent_id", agentID).Info("Agent restarted")

	return nil
}

// GetAgent retrieves an agent by ID
func (m *Manager) GetAgent(agentID string) (*agent.Agent, error) {
	m.mu.RLock()
	a, exists := m.agents[agentID]
	m.mu.RUnlock()

	if exists {
		return a, nil
	}

	// Try to load from registry if not in memory
	if m.registry != nil {
		a, err := m.registry.Get(m.ctx, agentID)
		if err != nil {
			return nil, agent.ErrAgentNotFound
		}

		// Add to cache
		m.mu.Lock()
		m.agents[agentID] = a
		m.mu.Unlock()

		return a, nil
	}

	return nil, agent.ErrAgentNotFound
}

// ListAgents returns all registered agents sorted by ID for consistent ordering
func (m *Manager) ListAgents() []*agent.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*agent.Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}

	// Sort by ID to ensure consistent order across requests (prevent shuffling in UI)
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}

// ListAgentsFromRegistry returns all agents from persistent storage
func (m *Manager) ListAgentsFromRegistry() ([]*agent.Agent, error) {
	if m.registry == nil {
		// Fallback to in-memory list
		return m.ListAgents(), nil
	}

	agents, err := m.registry.List(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents from registry: %w", err)
	}

	return agents, nil
}

// healthCheckLoop periodically checks agent health
func (m *Manager) healthCheckLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAgentHealth()
		}
	}
}

// checkAgentHealth checks all agents for health issues
func (m *Manager) checkAgentHealth() {
	m.mu.RLock()
	agents := make([]*agent.Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	m.mu.RUnlock()

	for _, a := range agents {
		if !a.IsHealthy() {
			m.logger.WithFields(logrus.Fields{
				"agent_id":       a.ID,
				"last_heartbeat": a.LastHeartbeat,
				"current_state":  a.GetState(),
			}).Warn("Agent health check failed")

			// TODO: Implement recovery strategy
		}
	}
}

// GetMetrics returns current runtime metrics
func (m *Manager) GetMetrics() Metrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// Return a copy of the metrics
	return m.metrics.metrics
}

// Shutdown gracefully shuts down the manager and all agents
func (m *Manager) Shutdown() error {
	m.logger.Info("Shutting down runtime manager")

	// Stop all agents
	m.mu.RLock()
	agentIDs := make([]string, 0, len(m.agents))
	for id := range m.agents {
		agentIDs = append(agentIDs, id)
	}
	m.mu.RUnlock()

	for _, id := range agentIDs {
		if err := m.StopAgent(id); err != nil {
			m.logger.WithError(err).Errorf("Error stopping agent %s", id)
		}
	}

	// Cancel manager context
	m.cancel()

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Runtime manager shutdown complete")
		return nil
	case <-time.After(m.config.ShutdownTimeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
