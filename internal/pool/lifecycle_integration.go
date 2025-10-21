package pool

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// LifecycleIntegration integrates pool management with agent lifecycle events
type LifecycleIntegration struct {
	// poolManager manages agent pools
	poolManager *Manager

	// logger for lifecycle integration
	logger Logger
}

// NewLifecycleIntegration creates a new lifecycle integration
func NewLifecycleIntegration(poolManager *Manager, logger Logger) *LifecycleIntegration {
	return &LifecycleIntegration{
		poolManager: poolManager,
		logger:      logger,
	}
}

// OnAgentCreated handles agent creation events
func (li *LifecycleIntegration) OnAgentCreated(ctx context.Context, agent *agent.Agent) error {
	// For now, we don't automatically add agents to pools on creation
	// This should be done explicitly through pool management APIs
	li.logger.Debug("Agent created, available for pool assignment",
		"agent_id", agent.ID,
		"agent_type", agent.Type)

	return nil
}

// OnAgentStarted handles agent start events
func (li *LifecycleIntegration) OnAgentStarted(ctx context.Context, agent *agent.Agent) error {
	// Find pools that contain this agent and update their status
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agent.ID {
				// Update health status to reflect that agent is running
				member.Healthy = true

				li.logger.Info("Updated pool member status after agent start",
					"pool_id", pool.ID,
					"agent_id", agent.ID)
				break
			}
		}
	}

	return nil
}

// OnAgentStopped handles agent stop events
func (li *LifecycleIntegration) OnAgentStopped(ctx context.Context, agentID string) error {
	// Find pools that contain this agent and mark as unhealthy
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agentID {
				// Mark as unhealthy since agent is stopped
				member.Healthy = false

				li.logger.Info("Updated pool member status after agent stop",
					"pool_id", pool.ID,
					"agent_id", agentID)
				break
			}
		}
	}

	return nil
}

// OnAgentPaused handles agent pause events
func (li *LifecycleIntegration) OnAgentPaused(ctx context.Context, agentID string) error {
	// Similar to stopped, mark as unhealthy while paused
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agentID {
				// Mark as unhealthy while paused
				member.Healthy = false

				li.logger.Info("Updated pool member status after agent pause",
					"pool_id", pool.ID,
					"agent_id", agentID)
				break
			}
		}
	}

	return nil
}

// OnAgentResumed handles agent resume events
func (li *LifecycleIntegration) OnAgentResumed(ctx context.Context, agent *agent.Agent) error {
	// Mark as healthy when resumed
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agent.ID {
				// Mark as healthy when resumed
				member.Healthy = true

				li.logger.Info("Updated pool member status after agent resume",
					"pool_id", pool.ID,
					"agent_id", agent.ID)
				break
			}
		}
	}

	return nil
}

// OnAgentDeleted handles agent deletion events
func (li *LifecycleIntegration) OnAgentDeleted(ctx context.Context, agentID string) error {
	// Remove agent from all pools when deleted
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agentID {
				// Remove from pool
				err := li.poolManager.RemoveAgentFromPool(ctx, pool.ID, agentID)
				if err != nil {
					li.logger.Warn("Failed to remove deleted agent from pool",
						"pool_id", pool.ID,
						"agent_id", agentID,
						"error", err)
				} else {
					li.logger.Info("Removed deleted agent from pool",
						"pool_id", pool.ID,
						"agent_id", agentID)
				}
				break
			}
		}
	}

	return nil
}

// OnAgentHealthChanged handles agent health status changes
func (li *LifecycleIntegration) OnAgentHealthChanged(ctx context.Context, agentID string, healthy bool) error {
	// Update health status in all pools containing this agent
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pools: %w", err)
	}

	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agentID {
				// Update health status
				member.Healthy = healthy

				li.logger.Debug("Updated pool member health status",
					"pool_id", pool.ID,
					"agent_id", agentID,
					"healthy", healthy)
				break
			}
		}
	}

	return nil
}

// GetAgentPools returns all pools that contain a specific agent
func (li *LifecycleIntegration) GetAgentPools(ctx context.Context, agentID string) ([]*AgentPool, error) {
	pools, err := li.poolManager.ListPools(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	var agentPools []*AgentPool
	for _, pool := range pools {
		members, err := pool.ListAgents(ctx)
		if err != nil {
			continue
		}

		for _, member := range members {
			if member.Agent.ID == agentID {
				agentPools = append(agentPools, pool)
				break
			}
		}
	}

	return agentPools, nil
}

// EnsurePoolCompliance ensures all pools meet their minimum agent requirements
func (li *LifecycleIntegration) EnsurePoolCompliance(ctx context.Context) error {
	pools, err := li.poolManager.ListPools(ctx, PoolStatusActive)
	if err != nil {
		return fmt.Errorf("failed to list active pools: %w", err)
	}

	for _, pool := range pools {
		healthyAgents, err := pool.GetHealthyAgents(ctx)
		if err != nil {
			li.logger.Warn("Failed to get healthy agents for pool",
				"pool_id", pool.ID,
				"error", err)
			continue
		}

		healthyCount := len(healthyAgents)
		if healthyCount < pool.Config.MinAgents {
			li.logger.Warn("Pool below minimum agent requirement",
				"pool_id", pool.ID,
				"healthy_agents", healthyCount,
				"min_required", pool.Config.MinAgents)

			// In a real implementation, this could trigger:
			// 1. Alert to operations team
			// 2. Automatic agent provisioning
			// 3. Pool status change to degraded
		}

		if healthyCount > pool.Config.MaxAgents {
			li.logger.Warn("Pool exceeds maximum agent limit",
				"pool_id", pool.ID,
				"healthy_agents", healthyCount,
				"max_allowed", pool.Config.MaxAgents)

			// In a real implementation, this could trigger:
			// 1. Automatic scaling down
			// 2. Load redistribution
		}
	}

	return nil
}

// AutoAssignAgent automatically assigns an agent to the most suitable pool
func (li *LifecycleIntegration) AutoAssignAgent(ctx context.Context, agent *agent.Agent) error {
	pools, err := li.poolManager.ListPools(ctx, PoolStatusActive)
	if err != nil {
		return fmt.Errorf("failed to list active pools: %w", err)
	}

	// Find the best pool for this agent based on:
	// 1. Agent type compatibility
	// 2. Pool capacity
	// 3. Resource requirements
	// 4. Load distribution

	var bestPool *AgentPool
	var bestScore float64

	for _, pool := range pools {
		score := li.calculatePoolScore(pool, agent)
		if score > bestScore {
			bestScore = score
			bestPool = pool
		}
	}

	if bestPool == nil {
		return fmt.Errorf("no suitable pool found for agent %s", agent.ID)
	}

	// Add agent to the best pool
	err = li.poolManager.AddAgentToPool(ctx, bestPool.ID, agent, 1) // Default weight
	if err != nil {
		return fmt.Errorf("failed to add agent to pool: %w", err)
	}

	li.logger.Info("Auto-assigned agent to pool",
		"agent_id", agent.ID,
		"pool_id", bestPool.ID,
		"pool_name", bestPool.Config.Name,
		"score", bestScore)

	return nil
}

// calculatePoolScore calculates a suitability score for an agent-pool combination
func (li *LifecycleIntegration) calculatePoolScore(pool *AgentPool, agent *agent.Agent) float64 {
	score := 0.0

	// Check capacity - higher score for pools with more available space
	members, err := pool.ListAgents(context.Background())
	if err != nil {
		return 0
	}

	currentCount := len(members)
	if currentCount >= pool.Config.MaxAgents {
		return 0 // Pool is full
	}

	capacityScore := float64(pool.Config.MaxAgents-currentCount) / float64(pool.Config.MaxAgents)
	score += capacityScore * 40 // 40% weight for capacity

	// Check resource availability
	metrics := pool.GetMetrics(context.Background())
	resourceScore := (100 - metrics.ResourceUtilization.CPUUsage) / 100
	score += resourceScore * 30 // 30% weight for resources

	// Check load distribution
	healthyCount := metrics.HealthyAgents
	loadScore := 1.0
	if healthyCount > 0 {
		averageLoad := float64(metrics.ActiveRequests) / float64(healthyCount)
		loadScore = 1.0 / (1.0 + averageLoad/10) // Diminishing returns for high load
	}
	score += loadScore * 30 // 30% weight for load distribution

	return score
}
