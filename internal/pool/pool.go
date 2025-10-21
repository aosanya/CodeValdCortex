package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/google/uuid"
)

// PoolStatus represents the current status of an agent pool
type PoolStatus string

const (
	PoolStatusActive   PoolStatus = "active"
	PoolStatusPaused   PoolStatus = "paused"
	PoolStatusDraining PoolStatus = "draining"
	PoolStatusStopped  PoolStatus = "stopped"
)

// LoadBalancingStrategy defines how tasks are distributed across agents
type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin      LoadBalancingStrategy = "round_robin"
	LoadBalancingLeastConnection LoadBalancingStrategy = "least_connection"
	LoadBalancingWeighted        LoadBalancingStrategy = "weighted"
	LoadBalancingRandom          LoadBalancingStrategy = "random"
)

// PoolConfig holds configuration for an agent pool
type PoolConfig struct {
	// Name is the pool identifier
	Name string

	// Description provides pool details
	Description string

	// LoadBalancingStrategy defines how to distribute work
	LoadBalancingStrategy LoadBalancingStrategy

	// MinAgents is the minimum number of agents in the pool
	MinAgents int

	// MaxAgents is the maximum number of agents in the pool
	MaxAgents int

	// HealthCheckInterval defines how often to check agent health
	HealthCheckInterval time.Duration

	// ResourceLimits define aggregate limits for the pool
	ResourceLimits ResourceLimits

	// AutoScaling enables automatic pool scaling
	AutoScaling AutoScalingConfig
}

// ResourceLimits define resource constraints for a pool
type ResourceLimits struct {
	// TotalCPU in millicores across all agents
	TotalCPU int

	// TotalMemory in megabytes across all agents
	TotalMemory int

	// MaxConcurrentTasks across all agents
	MaxConcurrentTasks int
}

// AutoScalingConfig defines auto-scaling behavior
type AutoScalingConfig struct {
	// Enabled indicates if auto-scaling is active
	Enabled bool

	// ScaleUpThreshold CPU percentage to trigger scale up
	ScaleUpThreshold float64

	// ScaleDownThreshold CPU percentage to trigger scale down
	ScaleDownThreshold float64

	// CooldownPeriod minimum time between scaling operations
	CooldownPeriod time.Duration
}

// AgentPoolMember represents an agent within a pool
type AgentPoolMember struct {
	// Agent is the actual agent instance
	Agent *agent.Agent

	// Weight for weighted load balancing (1-100)
	Weight int

	// JoinedAt is when the agent joined the pool
	JoinedAt time.Time

	// ActiveConnections tracks current workload
	ActiveConnections int

	// LastHealthCheck is the last health check time
	LastHealthCheck time.Time

	// Healthy indicates if the agent is healthy
	Healthy bool
}

// AgentPool manages a group of agents with load balancing and resource allocation
type AgentPool struct {
	// ID is the unique pool identifier
	ID string

	// Config holds pool configuration
	Config PoolConfig

	// Status is the current pool status
	Status PoolStatus

	// Members maps agent ID to pool member info
	Members map[string]*AgentPoolMember

	// LoadBalancer handles task distribution
	LoadBalancer LoadBalancer

	// CreatedAt is when the pool was created
	CreatedAt time.Time

	// UpdatedAt is when the pool was last modified
	UpdatedAt time.Time

	// Metrics tracks pool performance
	Metrics *PoolMetrics

	// mutex protects concurrent access
	mutex sync.RWMutex

	// ctx for pool operations
	ctx context.Context

	// cancel function for pool shutdown
	cancel context.CancelFunc

	// logger for pool operations
	logger Logger
}

// PoolMetrics tracks pool performance statistics
type PoolMetrics struct {
	// TotalRequests processed by the pool
	TotalRequests int64

	// ActiveRequests currently being processed
	ActiveRequests int64

	// FailedRequests that encountered errors
	FailedRequests int64

	// AverageResponseTime in milliseconds
	AverageResponseTime float64

	// TotalAgents currently in the pool
	TotalAgents int

	// HealthyAgents currently healthy
	HealthyAgents int

	// ResourceUtilization current resource usage
	ResourceUtilization ResourceUtilization

	// LastUpdated when metrics were last calculated
	LastUpdated time.Time
}

// ResourceUtilization tracks current resource usage
type ResourceUtilization struct {
	// CPUUsage percentage (0-100)
	CPUUsage float64

	// MemoryUsage percentage (0-100)
	MemoryUsage float64

	// TaskLoad percentage (0-100)
	TaskLoad float64
}

// Logger interface for pool logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewAgentPool creates a new agent pool with the given configuration
func NewAgentPool(config PoolConfig, logger Logger) (*AgentPool, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("pool name cannot be empty")
	}

	if config.MinAgents < 0 || config.MaxAgents <= 0 || config.MinAgents > config.MaxAgents {
		return nil, fmt.Errorf("invalid agent limits: min=%d, max=%d", config.MinAgents, config.MaxAgents)
	}

	if config.HealthCheckInterval <= 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	if config.LoadBalancingStrategy == "" {
		config.LoadBalancingStrategy = LoadBalancingRoundRobin
	}

	ctx, cancel := context.WithCancel(context.Background())

	poolID := uuid.New().String()

	pool := &AgentPool{
		ID:      poolID,
		Config:  config,
		Status:  PoolStatusActive,
		Members: make(map[string]*AgentPoolMember),
		Metrics: &PoolMetrics{
			LastUpdated: time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}

	// Initialize load balancer based on strategy
	var err error
	pool.LoadBalancer, err = NewLoadBalancer(config.LoadBalancingStrategy, pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer: %w", err)
	}

	// Start health monitoring
	go pool.startHealthMonitoring()

	pool.logger.Info("Created new agent pool",
		"pool_id", poolID,
		"name", config.Name,
		"strategy", config.LoadBalancingStrategy,
		"min_agents", config.MinAgents,
		"max_agents", config.MaxAgents)

	return pool, nil
}

// AddAgent adds an agent to the pool
func (p *AgentPool) AddAgent(ctx context.Context, agent *agent.Agent, weight int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status == PoolStatusStopped {
		return fmt.Errorf("cannot add agent to stopped pool")
	}

	if len(p.Members) >= p.Config.MaxAgents {
		return fmt.Errorf("pool at maximum capacity (%d agents)", p.Config.MaxAgents)
	}

	if _, exists := p.Members[agent.ID]; exists {
		return fmt.Errorf("agent %s already in pool", agent.ID)
	}

	if weight <= 0 || weight > 100 {
		weight = 1 // Default weight
	}

	member := &AgentPoolMember{
		Agent:           agent,
		Weight:          weight,
		JoinedAt:        time.Now(),
		LastHealthCheck: time.Now(),
		Healthy:         true,
	}

	p.Members[agent.ID] = member
	p.UpdatedAt = time.Now()

	// Update metrics
	p.updateMetrics()

	p.logger.Info("Added agent to pool",
		"pool_id", p.ID,
		"agent_id", agent.ID,
		"weight", weight,
		"total_agents", len(p.Members))

	return nil
}

// RemoveAgent removes an agent from the pool
func (p *AgentPool) RemoveAgent(ctx context.Context, agentID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	member, exists := p.Members[agentID]
	if !exists {
		return fmt.Errorf("agent %s not in pool", agentID)
	}

	// Wait for active connections to drain if any
	if member.ActiveConnections > 0 {
		p.logger.Warn("Removing agent with active connections",
			"pool_id", p.ID,
			"agent_id", agentID,
			"active_connections", member.ActiveConnections)
	}

	delete(p.Members, agentID)
	p.UpdatedAt = time.Now()

	// Update metrics
	p.updateMetrics()

	p.logger.Info("Removed agent from pool",
		"pool_id", p.ID,
		"agent_id", agentID,
		"total_agents", len(p.Members))

	return nil
}

// GetAgent returns an agent from the pool using the load balancing strategy
func (p *AgentPool) GetAgent(ctx context.Context) (*agent.Agent, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.Status != PoolStatusActive {
		return nil, fmt.Errorf("pool is not active (status: %s)", p.Status)
	}

	if len(p.Members) == 0 {
		return nil, fmt.Errorf("no agents available in pool")
	}

	// Get agent using load balancing strategy
	selectedAgent, err := p.LoadBalancer.SelectAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("load balancer failed to select agent: %w", err)
	}

	// Update connection count
	if member, exists := p.Members[selectedAgent.ID]; exists {
		member.ActiveConnections++
	}

	return selectedAgent, nil
}

// ReleaseAgent marks an agent as available after completing work
func (p *AgentPool) ReleaseAgent(ctx context.Context, agentID string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	member, exists := p.Members[agentID]
	if !exists {
		return fmt.Errorf("agent %s not in pool", agentID)
	}

	if member.ActiveConnections > 0 {
		member.ActiveConnections--
	}

	return nil
}

// ListAgents returns all agents in the pool
func (p *AgentPool) ListAgents(ctx context.Context) ([]*AgentPoolMember, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	members := make([]*AgentPoolMember, 0, len(p.Members))
	for _, member := range p.Members {
		members = append(members, member)
	}

	return members, nil
}

// GetHealthyAgents returns only healthy agents
func (p *AgentPool) GetHealthyAgents(ctx context.Context) ([]*AgentPoolMember, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var healthyMembers []*AgentPoolMember
	for _, member := range p.Members {
		if member.Healthy {
			healthyMembers = append(healthyMembers, member)
		}
	}

	return healthyMembers, nil
}

// UpdateConfig updates the pool configuration
func (p *AgentPool) UpdateConfig(ctx context.Context, config PoolConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Validate new config
	if config.MinAgents < 0 || config.MaxAgents <= 0 || config.MinAgents > config.MaxAgents {
		return fmt.Errorf("invalid agent limits: min=%d, max=%d", config.MinAgents, config.MaxAgents)
	}

	// Check if we need to remove agents due to new max limit
	if len(p.Members) > config.MaxAgents {
		return fmt.Errorf("cannot reduce max agents below current pool size (%d)", len(p.Members))
	}

	oldStrategy := p.Config.LoadBalancingStrategy
	p.Config = config
	p.UpdatedAt = time.Now()

	// Recreate load balancer if strategy changed
	if oldStrategy != config.LoadBalancingStrategy {
		var err error
		p.LoadBalancer, err = NewLoadBalancer(config.LoadBalancingStrategy, p)
		if err != nil {
			return fmt.Errorf("failed to update load balancer: %w", err)
		}
	}

	p.logger.Info("Updated pool configuration",
		"pool_id", p.ID,
		"strategy", config.LoadBalancingStrategy,
		"min_agents", config.MinAgents,
		"max_agents", config.MaxAgents)

	return nil
}

// GetMetrics returns current pool metrics
func (p *AgentPool) GetMetrics(ctx context.Context) *PoolMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *p.Metrics
	return &metrics
}

// Stop gracefully stops the pool
func (p *AgentPool) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status == PoolStatusStopped {
		return nil
	}

	p.Status = PoolStatusStopped
	p.cancel()

	p.logger.Info("Stopped agent pool", "pool_id", p.ID)
	return nil
}

// startHealthMonitoring runs health checks on pool agents
func (p *AgentPool) startHealthMonitoring() {
	ticker := time.NewTicker(p.Config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performHealthChecks()
		case <-p.ctx.Done():
			return
		}
	}
}

// performHealthChecks checks health of all agents in the pool
func (p *AgentPool) performHealthChecks() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	healthyCount := 0

	for agentID, member := range p.Members {
		// Check agent health (simplified - could integrate with agent health system)
		isHealthy := p.checkAgentHealth(member.Agent)

		member.Healthy = isHealthy
		member.LastHealthCheck = now

		if isHealthy {
			healthyCount++
		}

		p.logger.Debug("Health check completed",
			"pool_id", p.ID,
			"agent_id", agentID,
			"healthy", isHealthy)
	}

	// Update metrics
	p.Metrics.HealthyAgents = healthyCount
	p.Metrics.TotalAgents = len(p.Members)
	p.Metrics.LastUpdated = now
}

// checkAgentHealth performs health check on an individual agent
func (p *AgentPool) checkAgentHealth(agentInstance *agent.Agent) bool {
	// Simplified health check - in real implementation would check:
	// - Agent status
	// - Resource utilization
	// - Response time
	// - Error rates
	return agentInstance.State == agent.StateRunning
}

// updateMetrics updates pool performance metrics
func (p *AgentPool) updateMetrics() {
	now := time.Now()
	healthyCount := 0

	for _, member := range p.Members {
		if member.Healthy {
			healthyCount++
		}
	}

	p.Metrics.TotalAgents = len(p.Members)
	p.Metrics.HealthyAgents = healthyCount
	p.Metrics.LastUpdated = now

	// Calculate resource utilization (simplified)
	p.calculateResourceUtilization()
}

// calculateResourceUtilization calculates current resource usage
func (p *AgentPool) calculateResourceUtilization() {
	if len(p.Members) == 0 {
		p.Metrics.ResourceUtilization = ResourceUtilization{}
		return
	}

	var totalCPU, totalMemory, totalTasks int
	var activeTasks int

	for _, member := range p.Members {
		if member.Agent != nil {
			totalCPU += member.Agent.Config.Resources.CPU
			totalMemory += member.Agent.Config.Resources.Memory
			totalTasks += member.Agent.Config.Resources.MaxTasks
			activeTasks += member.ActiveConnections
		}
	}

	// Calculate percentages
	var cpuUsage, memoryUsage, taskLoad float64

	if p.Config.ResourceLimits.TotalCPU > 0 {
		cpuUsage = float64(totalCPU) / float64(p.Config.ResourceLimits.TotalCPU) * 100
	}

	if p.Config.ResourceLimits.TotalMemory > 0 {
		memoryUsage = float64(totalMemory) / float64(p.Config.ResourceLimits.TotalMemory) * 100
	}

	if totalTasks > 0 {
		taskLoad = float64(activeTasks) / float64(totalTasks) * 100
	}

	p.Metrics.ResourceUtilization = ResourceUtilization{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		TaskLoad:    taskLoad,
	}
}
