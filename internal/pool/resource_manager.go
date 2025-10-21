package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceAllocation represents resource allocation for an agent
type ResourceAllocation struct {
	// AgentID identifies the agent
	AgentID string

	// AllocatedCPU in millicores
	AllocatedCPU int

	// AllocatedMemory in megabytes
	AllocatedMemory int

	// AllocatedTasks maximum concurrent tasks
	AllocatedTasks int

	// Utilized resources tracking
	UtilizedCPU    float64 // percentage 0-100
	UtilizedMemory float64 // percentage 0-100
	ActiveTasks    int     // current active tasks

	// Allocation metadata
	AllocatedAt time.Time
	LastUpdated time.Time
}

// ResourceMonitor tracks resource usage across agents
type ResourceMonitor struct {
	// allocations maps agent ID to resource allocation
	allocations map[string]*ResourceAllocation

	// mutex protects concurrent access
	mutex sync.RWMutex

	// updateInterval defines monitoring frequency
	updateInterval time.Duration

	// ctx for monitor lifecycle
	ctx context.Context

	// cancel function for shutdown
	cancel context.CancelFunc

	// logger for resource monitoring
	logger Logger
}

// ResourceManager handles resource allocation and optimization for agent pools
type ResourceManager struct {
	// pools maps pool ID to pool instance
	pools map[string]*AgentPool

	// monitor tracks resource utilization
	monitor *ResourceMonitor

	// optimizer handles resource optimization
	optimizer *ResourceOptimizer

	// mutex protects concurrent access
	mutex sync.RWMutex

	// logger for resource management
	logger Logger
}

// ResourceOptimizer optimizes resource allocation across pools
type ResourceOptimizer struct {
	// enableAutoScaling controls automatic optimization
	enableAutoScaling bool

	// scaleUpThreshold CPU utilization to trigger scale up
	scaleUpThreshold float64

	// scaleDownThreshold CPU utilization to trigger scale down
	scaleDownThreshold float64

	// optimizationInterval how often to run optimization
	optimizationInterval time.Duration

	// logger for optimization operations
	logger Logger
}

// AllocationRequest represents a resource allocation request
type AllocationRequest struct {
	// AgentID requesting resources
	AgentID string

	// PoolID the agent belongs to
	PoolID string

	// RequestedCPU in millicores
	RequestedCPU int

	// RequestedMemory in megabytes
	RequestedMemory int

	// RequestedTasks maximum concurrent tasks
	RequestedTasks int

	// Priority of the allocation request (1-10, higher = more important)
	Priority int
}

// AllocationResult contains the result of a resource allocation
type AllocationResult struct {
	// Success indicates if allocation succeeded
	Success bool

	// Allocation contains the allocated resources (if successful)
	Allocation *ResourceAllocation

	// Reason explains allocation failure (if unsuccessful)
	Reason string

	// Recommendations suggest optimizations
	Recommendations []string
}

// NewResourceManager creates a new resource manager
func NewResourceManager(logger Logger) *ResourceManager {
	return &ResourceManager{
		pools:     make(map[string]*AgentPool),
		monitor:   NewResourceMonitor(30*time.Second, logger),
		optimizer: NewResourceOptimizer(logger),
		logger:    logger,
	}
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(updateInterval time.Duration, logger Logger) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &ResourceMonitor{
		allocations:    make(map[string]*ResourceAllocation),
		updateInterval: updateInterval,
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
	}

	// Start monitoring goroutine
	go monitor.startMonitoring()

	return monitor
}

// NewResourceOptimizer creates a new resource optimizer
func NewResourceOptimizer(logger Logger) *ResourceOptimizer {
	return &ResourceOptimizer{
		enableAutoScaling:    true,
		scaleUpThreshold:     80.0, // 80% CPU utilization
		scaleDownThreshold:   20.0, // 20% CPU utilization
		optimizationInterval: 60 * time.Second,
		logger:               logger,
	}
}

// RegisterPool registers a pool with the resource manager
func (rm *ResourceManager) RegisterPool(ctx context.Context, pool *AgentPool) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.pools[pool.ID]; exists {
		return fmt.Errorf("pool %s already registered", pool.ID)
	}

	rm.pools[pool.ID] = pool

	rm.logger.Info("Registered pool with resource manager",
		"pool_id", pool.ID,
		"pool_name", pool.Config.Name)

	return nil
}

// UnregisterPool removes a pool from the resource manager
func (rm *ResourceManager) UnregisterPool(ctx context.Context, poolID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.pools[poolID]; !exists {
		return fmt.Errorf("pool %s not registered", poolID)
	}

	delete(rm.pools, poolID)

	rm.logger.Info("Unregistered pool from resource manager", "pool_id", poolID)

	return nil
}

// AllocateResources allocates resources for an agent
func (rm *ResourceManager) AllocateResources(ctx context.Context, request *AllocationRequest) (*AllocationResult, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	pool, exists := rm.pools[request.PoolID]
	if !exists {
		return &AllocationResult{
			Success: false,
			Reason:  fmt.Sprintf("pool %s not found", request.PoolID),
		}, nil
	}

	// Check if pool has capacity for the request
	canAllocate, reason := rm.canAllocateResources(pool, request)
	if !canAllocate {
		return &AllocationResult{
			Success:         false,
			Reason:          reason,
			Recommendations: rm.generateRecommendations(pool, request),
		}, nil
	}

	// Create allocation
	allocation := &ResourceAllocation{
		AgentID:         request.AgentID,
		AllocatedCPU:    request.RequestedCPU,
		AllocatedMemory: request.RequestedMemory,
		AllocatedTasks:  request.RequestedTasks,
		AllocatedAt:     time.Now(),
		LastUpdated:     time.Now(),
	}

	// Register allocation with monitor
	rm.monitor.RegisterAllocation(allocation)

	rm.logger.Info("Allocated resources for agent",
		"agent_id", request.AgentID,
		"pool_id", request.PoolID,
		"cpu", request.RequestedCPU,
		"memory", request.RequestedMemory,
		"tasks", request.RequestedTasks)

	return &AllocationResult{
		Success:    true,
		Allocation: allocation,
	}, nil
}

// DeallocateResources releases resources for an agent
func (rm *ResourceManager) DeallocateResources(ctx context.Context, agentID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.monitor.UnregisterAllocation(agentID)

	rm.logger.Info("Deallocated resources for agent", "agent_id", agentID)

	return nil
}

// GetResourceUtilization returns current resource utilization for a pool
func (rm *ResourceManager) GetResourceUtilization(ctx context.Context, poolID string) (*ResourceUtilization, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	pool, exists := rm.pools[poolID]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolID)
	}

	metrics := pool.GetMetrics(ctx)
	return &metrics.ResourceUtilization, nil
}

// OptimizeAllocations optimizes resource allocation across pools
func (rm *ResourceManager) OptimizeAllocations(ctx context.Context) error {
	if !rm.optimizer.enableAutoScaling {
		return nil
	}

	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	for poolID, pool := range rm.pools {
		err := rm.optimizePoolResources(ctx, poolID, pool)
		if err != nil {
			rm.logger.Error("Failed to optimize pool resources",
				"pool_id", poolID,
				"error", err)
		}
	}

	return nil
}

// StartOptimizer starts the resource optimization process
func (rm *ResourceManager) StartOptimizer(ctx context.Context) {
	ticker := time.NewTicker(rm.optimizer.optimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.OptimizeAllocations(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// RegisterAllocation registers an allocation with the monitor
func (rm *ResourceMonitor) RegisterAllocation(allocation *ResourceAllocation) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.allocations[allocation.AgentID] = allocation

	rm.logger.Debug("Registered resource allocation",
		"agent_id", allocation.AgentID,
		"cpu", allocation.AllocatedCPU,
		"memory", allocation.AllocatedMemory)
}

// UnregisterAllocation removes an allocation from monitoring
func (rm *ResourceMonitor) UnregisterAllocation(agentID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.allocations, agentID)

	rm.logger.Debug("Unregistered resource allocation", "agent_id", agentID)
}

// UpdateUtilization updates resource utilization for an agent
func (rm *ResourceMonitor) UpdateUtilization(agentID string, cpuUsage, memoryUsage float64, activeTasks int) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	allocation, exists := rm.allocations[agentID]
	if !exists {
		return
	}

	allocation.UtilizedCPU = cpuUsage
	allocation.UtilizedMemory = memoryUsage
	allocation.ActiveTasks = activeTasks
	allocation.LastUpdated = time.Now()
}

// GetAllocation returns resource allocation for an agent
func (rm *ResourceMonitor) GetAllocation(agentID string) (*ResourceAllocation, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	allocation, exists := rm.allocations[agentID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	allocationCopy := *allocation
	return &allocationCopy, true
}

// GetAllAllocations returns all current resource allocations
func (rm *ResourceMonitor) GetAllAllocations() map[string]*ResourceAllocation {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	allocations := make(map[string]*ResourceAllocation)
	for agentID, allocation := range rm.allocations {
		// Create copies to avoid race conditions
		allocationCopy := *allocation
		allocations[agentID] = &allocationCopy
	}

	return allocations
}

// Stop stops the resource monitor
func (rm *ResourceMonitor) Stop() {
	rm.cancel()
}

// startMonitoring starts the resource monitoring loop
func (rm *ResourceMonitor) startMonitoring() {
	ticker := time.NewTicker(rm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.collectMetrics()
		case <-rm.ctx.Done():
			return
		}
	}
}

// collectMetrics collects resource utilization metrics
func (rm *ResourceMonitor) collectMetrics() {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	for agentID, allocation := range rm.allocations {
		// In a real implementation, this would collect actual metrics
		// from the agent or system monitoring
		rm.logger.Debug("Collecting metrics for agent",
			"agent_id", agentID,
			"cpu_utilization", allocation.UtilizedCPU,
			"memory_utilization", allocation.UtilizedMemory,
			"active_tasks", allocation.ActiveTasks)
	}
}

// Helper methods

// canAllocateResources checks if a pool can accommodate a resource request
func (rm *ResourceManager) canAllocateResources(pool *AgentPool, request *AllocationRequest) (bool, string) {
	metrics := pool.GetMetrics(context.Background())

	// Calculate current resource usage
	currentCPU := int(float64(pool.Config.ResourceLimits.TotalCPU) * metrics.ResourceUtilization.CPUUsage / 100)
	currentMemory := int(float64(pool.Config.ResourceLimits.TotalMemory) * metrics.ResourceUtilization.MemoryUsage / 100)
	currentTasks := int(float64(pool.Config.ResourceLimits.MaxConcurrentTasks) * metrics.ResourceUtilization.TaskLoad / 100)

	// Check if request would exceed limits
	if currentCPU+request.RequestedCPU > pool.Config.ResourceLimits.TotalCPU {
		return false, fmt.Sprintf("insufficient CPU: requested %d, available %d",
			request.RequestedCPU, pool.Config.ResourceLimits.TotalCPU-currentCPU)
	}

	if currentMemory+request.RequestedMemory > pool.Config.ResourceLimits.TotalMemory {
		return false, fmt.Sprintf("insufficient memory: requested %d, available %d",
			request.RequestedMemory, pool.Config.ResourceLimits.TotalMemory-currentMemory)
	}

	if currentTasks+request.RequestedTasks > pool.Config.ResourceLimits.MaxConcurrentTasks {
		return false, fmt.Sprintf("insufficient task capacity: requested %d, available %d",
			request.RequestedTasks, pool.Config.ResourceLimits.MaxConcurrentTasks-currentTasks)
	}

	return true, ""
}

// generateRecommendations generates optimization recommendations
func (rm *ResourceManager) generateRecommendations(pool *AgentPool, request *AllocationRequest) []string {
	var recommendations []string

	metrics := pool.GetMetrics(context.Background())

	if metrics.ResourceUtilization.CPUUsage > 80 {
		recommendations = append(recommendations, "Consider adding more agents to the pool to reduce CPU utilization")
	}

	if metrics.ResourceUtilization.MemoryUsage > 80 {
		recommendations = append(recommendations, "Consider increasing memory limits or optimizing memory usage")
	}

	if metrics.ResourceUtilization.TaskLoad > 80 {
		recommendations = append(recommendations, "Consider increasing max concurrent tasks or adding more agents")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Pool resources are optimally allocated")
	}

	return recommendations
}

// optimizePoolResources optimizes resources for a specific pool
func (rm *ResourceManager) optimizePoolResources(ctx context.Context, poolID string, pool *AgentPool) error {
	metrics := pool.GetMetrics(ctx)

	// Check if scaling is needed
	if metrics.ResourceUtilization.CPUUsage > rm.optimizer.scaleUpThreshold {
		return rm.triggerScaleUp(ctx, poolID, pool)
	}

	if metrics.ResourceUtilization.CPUUsage < rm.optimizer.scaleDownThreshold {
		return rm.triggerScaleDown(ctx, poolID, pool)
	}

	return nil
}

// triggerScaleUp initiates scale up for a pool
func (rm *ResourceManager) triggerScaleUp(ctx context.Context, poolID string, pool *AgentPool) error {
	rm.logger.Info("Triggering scale up for pool",
		"pool_id", poolID,
		"current_agents", pool.Metrics.TotalAgents,
		"cpu_utilization", pool.Metrics.ResourceUtilization.CPUUsage)

	// In a real implementation, this would:
	// 1. Calculate optimal number of new agents
	// 2. Request new agent instances
	// 3. Add them to the pool
	// For now, just log the action

	return nil
}

// triggerScaleDown initiates scale down for a pool
func (rm *ResourceManager) triggerScaleDown(ctx context.Context, poolID string, pool *AgentPool) error {
	rm.logger.Info("Triggering scale down for pool",
		"pool_id", poolID,
		"current_agents", pool.Metrics.TotalAgents,
		"cpu_utilization", pool.Metrics.ResourceUtilization.CPUUsage)

	// In a real implementation, this would:
	// 1. Identify agents to remove
	// 2. Gracefully drain connections
	// 3. Remove agents from pool
	// For now, just log the action

	return nil
}
