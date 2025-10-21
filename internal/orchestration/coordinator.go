package orchestration

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/health"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	log "github.com/sirupsen/logrus"
)

// Coordinator implements the AgentCoordinator interface
type Coordinator struct {
	// Configuration
	config CoordinatorConfig

	// Dependencies
	runtimeManager *runtime.Manager
	healthMonitor  *health.Monitor

	// Runtime state
	agentLoads map[string]*AgentLoad
	loadMutex  sync.RWMutex

	// Load balancing
	roundRobinIndex map[string]int
	rrMutex         sync.Mutex

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Logger
	logger *log.Logger
}

// CoordinatorConfig configures the agent coordinator
type CoordinatorConfig struct {
	// LoadUpdateInterval for refreshing agent load information
	LoadUpdateInterval time.Duration

	// HealthThreshold minimum health score for agent selection
	HealthThreshold float64

	// MaxTasksPerAgent limits tasks assigned to a single agent
	MaxTasksPerAgent int

	// LoadBalancingStrategy default strategy for task distribution
	LoadBalancingStrategy AgentSelectionStrategy

	// EnableHealthAwareSelection considers agent health in selection
	EnableHealthAwareSelection bool

	// CapabilityMatching enables strict capability matching
	CapabilityMatching bool
}

// DefaultCoordinatorConfig returns default coordinator configuration
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		LoadUpdateInterval:         30 * time.Second,
		HealthThreshold:            0.7, // 70% health threshold
		MaxTasksPerAgent:           10,
		LoadBalancingStrategy:      AgentSelectionLeastLoaded,
		EnableHealthAwareSelection: true,
		CapabilityMatching:         true,
	}
}

// NewCoordinator creates a new agent coordinator
func NewCoordinator(config CoordinatorConfig, runtimeManager *runtime.Manager, healthMonitor *health.Monitor, logger *log.Logger) *Coordinator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Coordinator{
		config:          config,
		runtimeManager:  runtimeManager,
		healthMonitor:   healthMonitor,
		agentLoads:      make(map[string]*AgentLoad),
		roundRobinIndex: make(map[string]int),
		ctx:             ctx,
		cancel:          cancel,
		logger:          logger,
	}
}

// Start starts the coordinator
func (c *Coordinator) Start() error {
	c.logger.Info("Starting agent coordinator")

	// Start load monitoring
	c.wg.Add(1)
	go c.loadMonitorWorker()

	c.logger.Info("Agent coordinator started successfully")
	return nil
}

// Stop stops the coordinator
func (c *Coordinator) Stop() error {
	c.logger.Info("Stopping agent coordinator")

	// Cancel context to signal workers to stop
	c.cancel()

	// Wait for all workers to finish
	c.wg.Wait()

	c.logger.Info("Agent coordinator stopped successfully")
	return nil
}

// SelectAgents chooses agents for task execution based on criteria
func (c *Coordinator) SelectAgents(ctx context.Context, selector AgentSelector, count int) ([]*agent.Agent, error) {
	c.logger.WithFields(log.Fields{
		"strategy": selector.Strategy,
		"count":    count,
	}).Debug("Selecting agents for task execution")

	// Get available agents
	availableAgents, err := c.GetAvailableAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}

	// Filter agents based on selector criteria
	candidateAgents, err := c.filterAgents(availableAgents, selector)
	if err != nil {
		return nil, fmt.Errorf("failed to filter agents: %w", err)
	}

	if len(candidateAgents) == 0 {
		return nil, fmt.Errorf("no agents match selection criteria")
	}

	// Select agents based on strategy
	selectedAgents, err := c.selectByStrategy(candidateAgents, selector.Strategy, count)
	if err != nil {
		return nil, fmt.Errorf("failed to select agents by strategy: %w", err)
	}

	c.logger.WithFields(log.Fields{
		"strategy":     selector.Strategy,
		"candidates":   len(candidateAgents),
		"selected":     len(selectedAgents),
		"selected_ids": c.getAgentIDs(selectedAgents),
	}).Debug("Agent selection completed")

	return selectedAgents, nil
}

// AssignTask assigns a task to a specific agent
func (c *Coordinator) AssignTask(ctx context.Context, agentID string, task *WorkflowTask, execution *WorkflowExecution) error {
	c.logger.WithFields(log.Fields{
		"agent_id":     agentID,
		"task_id":      task.ID,
		"execution_id": execution.ID,
	}).Debug("Assigning task to agent")

	// Get agent load
	agentLoad, err := c.GetAgentLoad(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent load: %w", err)
	}

	// Check if agent can handle more tasks
	if agentLoad.ActiveTasks >= c.config.MaxTasksPerAgent {
		return fmt.Errorf("agent %s is at maximum capacity (%d tasks)", agentID, c.config.MaxTasksPerAgent)
	}

	// Check agent health if health-aware selection is enabled
	if c.config.EnableHealthAwareSelection && agentLoad.HealthScore < c.config.HealthThreshold {
		return fmt.Errorf("agent %s health score %.2f is below threshold %.2f", agentID, agentLoad.HealthScore, c.config.HealthThreshold)
	}

	// Update agent load (optimistic)
	c.updateAgentLoad(agentID, agentLoad.ActiveTasks+1, agentLoad.QueuedTasks+1)

	// In a real implementation, this would submit the task to the agent's task system
	// For now, we simulate the assignment
	c.logger.WithFields(log.Fields{
		"agent_id":  agentID,
		"task_id":   task.ID,
		"task_type": task.Type,
	}).Info("Task assigned to agent successfully")

	return nil
}

// GetAgentLoad returns current load information for an agent
func (c *Coordinator) GetAgentLoad(ctx context.Context, agentID string) (*AgentLoad, error) {
	c.loadMutex.RLock()
	load, exists := c.agentLoads[agentID]
	c.loadMutex.RUnlock()

	if !exists {
		// Load information not available, create default
		load = &AgentLoad{
			AgentID:      agentID,
			ActiveTasks:  0,
			QueuedTasks:  0,
			CPUUsage:     0.0,
			MemoryUsage:  0.0,
			HealthScore:  1.0, // Assume healthy by default
			Capabilities: []string{},
			LastUpdated:  time.Now(),
		}

		// Try to get real load information
		if err := c.refreshAgentLoad(ctx, agentID); err != nil {
			c.logger.WithError(err).WithField("agent_id", agentID).Warn("Failed to refresh agent load")
		}
	}

	return load, nil
}

// GetAvailableAgents returns agents available for task assignment
func (c *Coordinator) GetAvailableAgents(ctx context.Context) ([]*agent.Agent, error) {
	// Get all agents from runtime manager
	agents := c.runtimeManager.ListAgents()

	// Filter to only running agents
	availableAgents := make([]*agent.Agent, 0)
	for _, ag := range agents {
		if ag.GetState() == agent.StateRunning {
			availableAgents = append(availableAgents, ag)
		}
	}

	c.logger.WithFields(log.Fields{
		"total_agents":     len(agents),
		"available_agents": len(availableAgents),
	}).Debug("Retrieved available agents")

	return availableAgents, nil
}

// RebalanceLoad redistributes tasks across agents
func (c *Coordinator) RebalanceLoad(ctx context.Context) error {
	c.logger.Info("Starting load rebalancing")

	// Get all agents and their loads
	agents, err := c.GetAvailableAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(agents) < 2 {
		c.logger.Debug("Not enough agents for load balancing")
		return nil
	}

	// Calculate load distribution
	totalTasks := 0
	agentLoads := make(map[string]*AgentLoad)

	for _, ag := range agents {
		load, err := c.GetAgentLoad(ctx, ag.ID)
		if err != nil {
			c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to get agent load")
			continue
		}
		agentLoads[ag.ID] = load
		totalTasks += load.ActiveTasks + load.QueuedTasks
	}

	if totalTasks == 0 {
		c.logger.Debug("No tasks to rebalance")
		return nil
	}

	// Calculate target load per agent
	targetLoad := totalTasks / len(agents)
	c.logger.WithFields(log.Fields{
		"total_tasks":  totalTasks,
		"total_agents": len(agents),
		"target_load":  targetLoad,
	}).Debug("Load rebalancing analysis")

	// For now, just log the analysis - actual task migration would require
	// integration with the task management system
	for agentID, load := range agentLoads {
		currentLoad := load.ActiveTasks + load.QueuedTasks
		if currentLoad > targetLoad+1 {
			c.logger.WithFields(log.Fields{
				"agent_id":     agentID,
				"current_load": currentLoad,
				"target_load":  targetLoad,
				"excess":       currentLoad - targetLoad,
			}).Info("Agent is overloaded - would migrate tasks")
		} else if currentLoad < targetLoad-1 {
			c.logger.WithFields(log.Fields{
				"agent_id":     agentID,
				"current_load": currentLoad,
				"target_load":  targetLoad,
				"capacity":     targetLoad - currentLoad,
			}).Info("Agent has capacity - could receive tasks")
		}
	}

	c.logger.Info("Load rebalancing analysis completed")
	return nil
}

// Helper methods

func (c *Coordinator) filterAgents(agents []*agent.Agent, selector AgentSelector) ([]*agent.Agent, error) {
	candidates := make([]*agent.Agent, 0)

	for _, ag := range agents {
		// Check specific agents constraint
		if len(selector.SpecificAgents) > 0 {
			found := false
			for _, specificID := range selector.SpecificAgents {
				if ag.ID == specificID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check pool constraint
		if selector.PoolID != "" {
			// In a real implementation, would check agent's pool membership
			// For now, assume all agents are in the default pool
		}

		// Check required capabilities
		if len(selector.RequiredCapabilities) > 0 && c.config.CapabilityMatching {
			agentLoad, err := c.GetAgentLoad(context.Background(), ag.ID)
			if err != nil {
				c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to get agent capabilities")
				continue
			}

			if !c.hasRequiredCapabilities(agentLoad.Capabilities, selector.RequiredCapabilities) {
				continue
			}
		}

		// Check health threshold
		if c.config.EnableHealthAwareSelection && selector.HealthThreshold > 0 {
			agentLoad, err := c.GetAgentLoad(context.Background(), ag.ID)
			if err != nil {
				c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to get agent health")
				continue
			}

			if agentLoad.HealthScore < selector.HealthThreshold {
				continue
			}
		}

		// Check tags
		if len(selector.Tags) > 0 {
			// In a real implementation, would check agent tags
			// For now, assume all agents match
		}

		candidates = append(candidates, ag)
	}

	return candidates, nil
}

func (c *Coordinator) selectByStrategy(candidates []*agent.Agent, strategy AgentSelectionStrategy, count int) ([]*agent.Agent, error) {
	if count <= 0 {
		return []*agent.Agent{}, nil
	}

	if count > len(candidates) {
		count = len(candidates)
	}

	switch strategy {
	case AgentSelectionRoundRobin:
		return c.selectRoundRobin(candidates, count)
	case AgentSelectionLeastLoaded:
		return c.selectLeastLoaded(candidates, count)
	case AgentSelectionHealthAware:
		return c.selectHealthAware(candidates, count)
	case AgentSelectionSpecific:
		// For specific selection, just return first N candidates
		return candidates[:count], nil
	case AgentSelectionCapabilityBased:
		// For capability-based, candidates are already filtered
		return candidates[:count], nil
	default:
		return c.selectRoundRobin(candidates, count)
	}
}

func (c *Coordinator) selectRoundRobin(candidates []*agent.Agent, count int) ([]*agent.Agent, error) {
	c.rrMutex.Lock()
	defer c.rrMutex.Unlock()

	strategyKey := "default"
	startIndex := c.roundRobinIndex[strategyKey]

	selected := make([]*agent.Agent, 0, count)
	for i := 0; i < count; i++ {
		index := (startIndex + i) % len(candidates)
		selected = append(selected, candidates[index])
	}

	// Update round-robin index
	c.roundRobinIndex[strategyKey] = (startIndex + count) % len(candidates)

	return selected, nil
}

func (c *Coordinator) selectLeastLoaded(candidates []*agent.Agent, count int) ([]*agent.Agent, error) {
	// Create slice with load information
	type agentWithLoad struct {
		agent *agent.Agent
		load  int
	}

	agentsWithLoad := make([]agentWithLoad, 0, len(candidates))

	for _, ag := range candidates {
		agentLoad, err := c.GetAgentLoad(context.Background(), ag.ID)
		if err != nil {
			c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to get agent load for selection")
			continue
		}

		agentsWithLoad = append(agentsWithLoad, agentWithLoad{
			agent: ag,
			load:  agentLoad.ActiveTasks + agentLoad.QueuedTasks,
		})
	}

	// Sort by load (ascending)
	sort.Slice(agentsWithLoad, func(i, j int) bool {
		return agentsWithLoad[i].load < agentsWithLoad[j].load
	})

	// Select least loaded agents
	selected := make([]*agent.Agent, 0, count)
	for i := 0; i < count && i < len(agentsWithLoad); i++ {
		selected = append(selected, agentsWithLoad[i].agent)
	}

	return selected, nil
}

func (c *Coordinator) selectHealthAware(candidates []*agent.Agent, count int) ([]*agent.Agent, error) {
	// Create slice with health information
	type agentWithHealth struct {
		agent  *agent.Agent
		health float64
	}

	agentsWithHealth := make([]agentWithHealth, 0, len(candidates))

	for _, ag := range candidates {
		agentLoad, err := c.GetAgentLoad(context.Background(), ag.ID)
		if err != nil {
			c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to get agent health for selection")
			continue
		}

		agentsWithHealth = append(agentsWithHealth, agentWithHealth{
			agent:  ag,
			health: agentLoad.HealthScore,
		})
	}

	// Sort by health (descending)
	sort.Slice(agentsWithHealth, func(i, j int) bool {
		return agentsWithHealth[i].health > agentsWithHealth[j].health
	})

	// Select healthiest agents
	selected := make([]*agent.Agent, 0, count)
	for i := 0; i < count && i < len(agentsWithHealth); i++ {
		selected = append(selected, agentsWithHealth[i].agent)
	}

	return selected, nil
}

func (c *Coordinator) hasRequiredCapabilities(agentCapabilities, requiredCapabilities []string) bool {
	// Create map of agent capabilities for fast lookup
	agentCaps := make(map[string]bool)
	for _, cap := range agentCapabilities {
		agentCaps[cap] = true
	}

	// Check if all required capabilities are present
	for _, required := range requiredCapabilities {
		if !agentCaps[required] {
			return false
		}
	}

	return true
}

func (c *Coordinator) getAgentIDs(agents []*agent.Agent) []string {
	ids := make([]string, len(agents))
	for i, ag := range agents {
		ids[i] = ag.ID
	}
	return ids
}

func (c *Coordinator) updateAgentLoad(agentID string, activeTasks, queuedTasks int) {
	c.loadMutex.Lock()
	defer c.loadMutex.Unlock()

	load, exists := c.agentLoads[agentID]
	if !exists {
		load = &AgentLoad{
			AgentID:      agentID,
			ActiveTasks:  0,
			QueuedTasks:  0,
			CPUUsage:     0.0,
			MemoryUsage:  0.0,
			HealthScore:  1.0,
			Capabilities: []string{},
			LastUpdated:  time.Now(),
		}
		c.agentLoads[agentID] = load
	}

	load.ActiveTasks = activeTasks
	load.QueuedTasks = queuedTasks
	load.LastUpdated = time.Now()
}

func (c *Coordinator) refreshAgentLoad(ctx context.Context, agentID string) error {
	// In a real implementation, this would:
	// 1. Query the agent's task manager for current task counts
	// 2. Get resource usage from the runtime manager
	// 3. Get health score from the health monitor
	// 4. Get capabilities from agent metadata

	// For now, simulate load information
	c.loadMutex.Lock()
	defer c.loadMutex.Unlock()

	// Simulate some load
	activeTasks := rand.Intn(5)             // 0-4 active tasks
	queuedTasks := rand.Intn(3)             // 0-2 queued tasks
	cpuUsage := rand.Float64() * 100        // 0-100% CPU
	memoryUsage := rand.Float64() * 100     // 0-100% memory
	healthScore := 0.7 + rand.Float64()*0.3 // 0.7-1.0 health score

	c.agentLoads[agentID] = &AgentLoad{
		AgentID:      agentID,
		ActiveTasks:  activeTasks,
		QueuedTasks:  queuedTasks,
		CPUUsage:     cpuUsage,
		MemoryUsage:  memoryUsage,
		HealthScore:  healthScore,
		Capabilities: []string{"http_request", "data_processing", "file_io"}, // Default capabilities
		LastUpdated:  time.Now(),
	}

	return nil
}

// Worker methods

func (c *Coordinator) loadMonitorWorker() {
	defer c.wg.Done()

	c.logger.Debug("Load monitor worker started")

	ticker := time.NewTicker(c.config.LoadUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.updateAllAgentLoads()
		}
	}
}

func (c *Coordinator) updateAllAgentLoads() {
	agents, err := c.GetAvailableAgents(context.Background())
	if err != nil {
		c.logger.WithError(err).Error("Failed to get agents for load update")
		return
	}

	for _, ag := range agents {
		if err := c.refreshAgentLoad(context.Background(), ag.ID); err != nil {
			c.logger.WithError(err).WithField("agent_id", ag.ID).Warn("Failed to refresh agent load")
		}
	}

	c.logger.WithField("agent_count", len(agents)).Debug("Updated agent loads")
}
