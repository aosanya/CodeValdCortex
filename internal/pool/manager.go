package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// Manager manages agent pools and their lifecycle
type Manager struct {
	// pools maps pool ID to pool instance
	pools map[string]*AgentPool

	// repository handles pool persistence
	repository *Repository

	// resourceManager handles resource allocation
	resourceManager *ResourceManager

	// mutex protects concurrent access
	mutex sync.RWMutex

	// ctx for manager lifecycle
	ctx context.Context

	// cancel function for shutdown
	cancel context.CancelFunc

	// logger for pool management
	logger Logger
}

// ManagerConfig holds configuration for the pool manager
type ManagerConfig struct {
	// Repository configuration
	RepositoryConfig RepositoryConfig

	// EnableAutoScaling controls automatic pool scaling
	EnableAutoScaling bool

	// MetricsInterval defines how often to collect metrics
	MetricsInterval time.Duration

	// CleanupInterval defines how often to cleanup old data
	CleanupInterval time.Duration

	// MetricsRetention defines how long to keep metrics
	MetricsRetention time.Duration
}

// NewManager creates a new pool manager
func NewManager(config ManagerConfig, logger Logger) (*Manager, error) {
	// Create repository
	repository, err := NewRepository(config.RepositoryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create resource manager
	resourceManager := NewResourceManager(logger)

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		pools:           make(map[string]*AgentPool),
		repository:      repository,
		resourceManager: resourceManager,
		ctx:             ctx,
		cancel:          cancel,
		logger:          logger,
	}

	// Start background processes
	go manager.startMetricsCollection(config.MetricsInterval)
	go manager.startCleanupProcess(config.CleanupInterval, config.MetricsRetention)

	if config.EnableAutoScaling {
		go manager.resourceManager.StartOptimizer(ctx)
	}

	// Load existing pools from database
	if err := manager.loadPools(ctx); err != nil {
		logger.Warn("Failed to load existing pools", "error", err)
	}

	logger.Info("Pool manager started",
		"auto_scaling", config.EnableAutoScaling,
		"metrics_interval", config.MetricsInterval)

	return manager, nil
}

// CreatePool creates a new agent pool
func (m *Manager) CreatePool(ctx context.Context, config PoolConfig) (*AgentPool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if pool with same name already exists
	for _, pool := range m.pools {
		if pool.Config.Name == config.Name {
			return nil, fmt.Errorf("pool with name '%s' already exists", config.Name)
		}
	}

	// Create pool
	pool, err := NewAgentPool(config, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Store in memory
	m.pools[pool.ID] = pool

	// Persist to database
	if err := m.repository.StorePool(ctx, pool); err != nil {
		// Remove from memory if persistence fails
		delete(m.pools, pool.ID)
		return nil, fmt.Errorf("failed to persist pool: %w", err)
	}

	// Register with resource manager
	if err := m.resourceManager.RegisterPool(ctx, pool); err != nil {
		m.logger.Warn("Failed to register pool with resource manager",
			"pool_id", pool.ID,
			"error", err)
	}

	m.logger.Info("Created new pool",
		"pool_id", pool.ID,
		"name", config.Name,
		"strategy", config.LoadBalancingStrategy)

	return pool, nil
}

// GetPool retrieves a pool by ID
func (m *Manager) GetPool(ctx context.Context, poolID string) (*AgentPool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	pool, exists := m.pools[poolID]
	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolID)
	}

	return pool, nil
}

// ListPools returns all pools with optional status filter
func (m *Manager) ListPools(ctx context.Context, status PoolStatus) ([]*AgentPool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var pools []*AgentPool
	for _, pool := range m.pools {
		if status == "" || pool.Status == status {
			pools = append(pools, pool)
		}
	}

	return pools, nil
}

// DeletePool removes a pool
func (m *Manager) DeletePool(ctx context.Context, poolID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pool, exists := m.pools[poolID]
	if !exists {
		return fmt.Errorf("pool not found: %s", poolID)
	}

	// Stop the pool
	if err := pool.Stop(ctx); err != nil {
		m.logger.Warn("Failed to stop pool during deletion",
			"pool_id", poolID,
			"error", err)
	}

	// Unregister from resource manager
	if err := m.resourceManager.UnregisterPool(ctx, poolID); err != nil {
		m.logger.Warn("Failed to unregister pool from resource manager",
			"pool_id", poolID,
			"error", err)
	}

	// Remove from memory
	delete(m.pools, poolID)

	// Remove from database
	if err := m.repository.DeletePool(ctx, poolID); err != nil {
		m.logger.Warn("Failed to delete pool from database",
			"pool_id", poolID,
			"error", err)
	}

	m.logger.Info("Deleted pool", "pool_id", poolID)

	return nil
}

// AddAgentToPool adds an agent to a specific pool
func (m *Manager) AddAgentToPool(ctx context.Context, poolID string, agent *agent.Agent, weight int) error {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolID)
	}

	// Add agent to pool
	if err := pool.AddAgent(ctx, agent, weight); err != nil {
		return fmt.Errorf("failed to add agent to pool: %w", err)
	}

	// Persist membership
	member := &AgentPoolMember{
		Agent:           agent,
		Weight:          weight,
		JoinedAt:        time.Now(),
		LastHealthCheck: time.Now(),
		Healthy:         true,
	}

	if err := m.repository.StoreMembership(ctx, poolID, member); err != nil {
		m.logger.Warn("Failed to persist pool membership",
			"pool_id", poolID,
			"agent_id", agent.ID,
			"error", err)
	}

	// Allocate resources
	request := &AllocationRequest{
		AgentID:         agent.ID,
		PoolID:          poolID,
		RequestedCPU:    agent.Config.Resources.CPU,
		RequestedMemory: agent.Config.Resources.Memory,
		RequestedTasks:  agent.Config.Resources.MaxTasks,
		Priority:        5, // Default priority
	}

	result, err := m.resourceManager.AllocateResources(ctx, request)
	if err != nil {
		m.logger.Warn("Failed to allocate resources for agent",
			"agent_id", agent.ID,
			"pool_id", poolID,
			"error", err)
	} else if !result.Success {
		m.logger.Warn("Resource allocation unsuccessful",
			"agent_id", agent.ID,
			"pool_id", poolID,
			"reason", result.Reason)
	}

	m.logger.Info("Added agent to pool",
		"pool_id", poolID,
		"agent_id", agent.ID,
		"weight", weight)

	return nil
}

// RemoveAgentFromPool removes an agent from a specific pool
func (m *Manager) RemoveAgentFromPool(ctx context.Context, poolID, agentID string) error {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolID)
	}

	// Remove agent from pool
	if err := pool.RemoveAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to remove agent from pool: %w", err)
	}

	// Remove membership from database
	if err := m.repository.RemoveMembership(ctx, poolID, agentID); err != nil {
		m.logger.Warn("Failed to remove pool membership from database",
			"pool_id", poolID,
			"agent_id", agentID,
			"error", err)
	}

	// Deallocate resources
	if err := m.resourceManager.DeallocateResources(ctx, agentID); err != nil {
		m.logger.Warn("Failed to deallocate resources for agent",
			"agent_id", agentID,
			"error", err)
	}

	m.logger.Info("Removed agent from pool",
		"pool_id", poolID,
		"agent_id", agentID)

	return nil
}

// GetAgentFromPool gets an available agent from a pool using load balancing
func (m *Manager) GetAgentFromPool(ctx context.Context, poolID string) (*agent.Agent, error) {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolID)
	}

	return pool.GetAgent(ctx)
}

// ReleaseAgent marks an agent as available after completing work
func (m *Manager) ReleaseAgent(ctx context.Context, poolID, agentID string) error {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolID)
	}

	return pool.ReleaseAgent(ctx, agentID)
}

// GetPoolMetrics returns metrics for a specific pool
func (m *Manager) GetPoolMetrics(ctx context.Context, poolID string) (*PoolMetrics, error) {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolID)
	}

	return pool.GetMetrics(ctx), nil
}

// UpdatePoolConfig updates the configuration for a pool
func (m *Manager) UpdatePoolConfig(ctx context.Context, poolID string, config PoolConfig) error {
	m.mutex.RLock()
	pool, exists := m.pools[poolID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolID)
	}

	// Update pool configuration
	if err := pool.UpdateConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to update pool config: %w", err)
	}

	// Persist updated configuration
	if err := m.repository.StorePool(ctx, pool); err != nil {
		m.logger.Warn("Failed to persist updated pool configuration",
			"pool_id", poolID,
			"error", err)
	}

	m.logger.Info("Updated pool configuration",
		"pool_id", poolID,
		"strategy", config.LoadBalancingStrategy)

	return nil
}

// GetResourceUtilization returns resource utilization for a pool
func (m *Manager) GetResourceUtilization(ctx context.Context, poolID string) (*ResourceUtilization, error) {
	return m.resourceManager.GetResourceUtilization(ctx, poolID)
}

// Stop gracefully stops the pool manager
func (m *Manager) Stop(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Stop all pools
	for poolID, pool := range m.pools {
		if err := pool.Stop(ctx); err != nil {
			m.logger.Warn("Failed to stop pool during manager shutdown",
				"pool_id", poolID,
				"error", err)
		}
	}

	// Cancel background processes
	m.cancel()

	m.logger.Info("Pool manager stopped")

	return nil
}

// loadPools loads existing pools from the database
func (m *Manager) loadPools(ctx context.Context) error {
	pools, err := m.repository.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to load pools from database: %w", err)
	}

	for _, poolDoc := range pools {
		// Convert document to pool config
		config := PoolConfig{
			Name:                  poolDoc.Name,
			Description:           poolDoc.Description,
			LoadBalancingStrategy: poolDoc.LoadBalancingStrategy,
			MinAgents:             poolDoc.MinAgents,
			MaxAgents:             poolDoc.MaxAgents,
			HealthCheckInterval:   time.Duration(poolDoc.HealthCheckInterval) * time.Millisecond,
			ResourceLimits:        poolDoc.ResourceLimits,
			AutoScaling:           poolDoc.AutoScaling,
		}

		// Create pool instance
		pool, err := NewAgentPool(config, m.logger)
		if err != nil {
			m.logger.Warn("Failed to recreate pool from database",
				"pool_id", poolDoc.ID,
				"error", err)
			continue
		}

		// Restore pool state
		pool.ID = poolDoc.ID
		pool.Status = poolDoc.Status
		pool.CreatedAt = poolDoc.CreatedAt
		pool.UpdatedAt = poolDoc.UpdatedAt

		// Load memberships
		memberships, err := m.repository.GetMemberships(ctx, poolDoc.ID)
		if err != nil {
			m.logger.Warn("Failed to load pool memberships",
				"pool_id", poolDoc.ID,
				"error", err)
		} else {
			// Note: In a real implementation, you would need to
			// reconcile these memberships with actual agent instances
			// from the agent registry
			for _, membership := range memberships {
				m.logger.Debug("Found pool membership",
					"pool_id", poolDoc.ID,
					"agent_id", membership.AgentID,
					"weight", membership.Weight)
			}
		}

		// Store in memory
		m.pools[pool.ID] = pool

		// Register with resource manager
		if err := m.resourceManager.RegisterPool(ctx, pool); err != nil {
			m.logger.Warn("Failed to register loaded pool with resource manager",
				"pool_id", pool.ID,
				"error", err)
		}

		m.logger.Info("Loaded pool from database",
			"pool_id", pool.ID,
			"name", pool.Config.Name)
	}

	m.logger.Info("Loaded pools from database", "count", len(pools))

	return nil
}

// startMetricsCollection starts the metrics collection process
func (m *Manager) startMetricsCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectAndStoreMetrics()
		case <-m.ctx.Done():
			return
		}
	}
}

// collectAndStoreMetrics collects and stores metrics for all pools
func (m *Manager) collectAndStoreMetrics() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for poolID, pool := range m.pools {
		metrics := pool.GetMetrics(context.Background())

		if err := m.repository.StoreMetrics(context.Background(), poolID, metrics); err != nil {
			m.logger.Warn("Failed to store pool metrics",
				"pool_id", poolID,
				"error", err)
		}
	}
}

// startCleanupProcess starts the cleanup process for old data
func (m *Manager) startCleanupProcess(interval, retention time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.repository.CleanupOldMetrics(context.Background(), retention); err != nil {
				m.logger.Warn("Failed to cleanup old metrics", "error", err)
			}
		case <-m.ctx.Done():
			return
		}
	}
}
