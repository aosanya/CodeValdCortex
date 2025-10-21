package pool

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// LoadBalancer interface defines methods for selecting agents from a pool
type LoadBalancer interface {
	// SelectAgent selects an agent based on the load balancing strategy
	SelectAgent(ctx context.Context) (*agent.Agent, error)

	// GetStrategy returns the current load balancing strategy
	GetStrategy() LoadBalancingStrategy

	// Reset resets the load balancer state
	Reset()
}

// NewLoadBalancer creates a load balancer based on the specified strategy
func NewLoadBalancer(strategy LoadBalancingStrategy, pool *AgentPool) (LoadBalancer, error) {
	switch strategy {
	case LoadBalancingRoundRobin:
		return NewRoundRobinBalancer(pool), nil
	case LoadBalancingLeastConnection:
		return NewLeastConnectionBalancer(pool), nil
	case LoadBalancingWeighted:
		return NewWeightedBalancer(pool), nil
	case LoadBalancingRandom:
		return NewRandomBalancer(pool), nil
	default:
		return nil, fmt.Errorf("unsupported load balancing strategy: %s", strategy)
	}
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
	pool     *AgentPool
	position int64
	mutex    sync.Mutex
}

// NewRoundRobinBalancer creates a new round-robin load balancer
func NewRoundRobinBalancer(pool *AgentPool) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		pool:     pool,
		position: 0,
	}
}

// SelectAgent selects the next agent in round-robin order
func (rb *RoundRobinBalancer) SelectAgent(ctx context.Context) (*agent.Agent, error) {
	healthyAgents, err := rb.pool.GetHealthyAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy agents: %w", err)
	}

	if len(healthyAgents) == 0 {
		return nil, fmt.Errorf("no healthy agents available")
	}

	// Atomic increment and modulo to get next position
	nextPos := atomic.AddInt64(&rb.position, 1) - 1
	selectedIndex := int(nextPos) % len(healthyAgents)

	return healthyAgents[selectedIndex].Agent, nil
}

// GetStrategy returns the round-robin strategy
func (rb *RoundRobinBalancer) GetStrategy() LoadBalancingStrategy {
	return LoadBalancingRoundRobin
}

// Reset resets the round-robin position
func (rb *RoundRobinBalancer) Reset() {
	atomic.StoreInt64(&rb.position, 0)
}

// LeastConnectionBalancer implements least-connection load balancing
type LeastConnectionBalancer struct {
	pool  *AgentPool
	mutex sync.RWMutex
}

// NewLeastConnectionBalancer creates a new least-connection load balancer
func NewLeastConnectionBalancer(pool *AgentPool) *LeastConnectionBalancer {
	return &LeastConnectionBalancer{
		pool: pool,
	}
}

// SelectAgent selects the agent with the fewest active connections
func (lcb *LeastConnectionBalancer) SelectAgent(ctx context.Context) (*agent.Agent, error) {
	healthyAgents, err := lcb.pool.GetHealthyAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy agents: %w", err)
	}

	if len(healthyAgents) == 0 {
		return nil, fmt.Errorf("no healthy agents available")
	}

	// Find agent with minimum connections
	var selectedAgent *agent.Agent
	minConnections := int(^uint(0) >> 1) // Max int

	for _, member := range healthyAgents {
		if member.ActiveConnections < minConnections {
			minConnections = member.ActiveConnections
			selectedAgent = member.Agent
		}
	}

	if selectedAgent == nil {
		return nil, fmt.Errorf("failed to select agent")
	}

	return selectedAgent, nil
}

// GetStrategy returns the least-connection strategy
func (lcb *LeastConnectionBalancer) GetStrategy() LoadBalancingStrategy {
	return LoadBalancingLeastConnection
}

// Reset resets the least-connection balancer state
func (lcb *LeastConnectionBalancer) Reset() {
	// No state to reset for least-connection
}

// WeightedBalancer implements weighted load balancing
type WeightedBalancer struct {
	pool          *AgentPool
	totalWeight   int
	currentIndex  int
	currentWeight int
	maxWeight     int
	gcd           int
	mutex         sync.Mutex
}

// NewWeightedBalancer creates a new weighted load balancer
func NewWeightedBalancer(pool *AgentPool) *WeightedBalancer {
	wb := &WeightedBalancer{
		pool: pool,
	}
	wb.calculateWeights()
	return wb
}

// SelectAgent selects an agent based on weighted distribution
func (wb *WeightedBalancer) SelectAgent(ctx context.Context) (*agent.Agent, error) {
	wb.mutex.Lock()
	defer wb.mutex.Unlock()

	healthyAgents, err := wb.pool.GetHealthyAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy agents: %w", err)
	}

	if len(healthyAgents) == 0 {
		return nil, fmt.Errorf("no healthy agents available")
	}

	// Recalculate weights if needed
	wb.calculateWeightsLocked(healthyAgents)

	if wb.totalWeight == 0 {
		// Fall back to round-robin if no weights
		return healthyAgents[wb.currentIndex%len(healthyAgents)].Agent, nil
	}

	// Weighted round-robin algorithm
	for {
		wb.currentIndex = (wb.currentIndex + 1) % len(healthyAgents)
		if wb.currentIndex == 0 {
			wb.currentWeight = wb.currentWeight - wb.gcd
			if wb.currentWeight <= 0 {
				wb.currentWeight = wb.maxWeight
			}
		}

		member := healthyAgents[wb.currentIndex]
		if member.Weight >= wb.currentWeight {
			return member.Agent, nil
		}
	}
}

// GetStrategy returns the weighted strategy
func (wb *WeightedBalancer) GetStrategy() LoadBalancingStrategy {
	return LoadBalancingWeighted
}

// Reset resets the weighted balancer state
func (wb *WeightedBalancer) Reset() {
	wb.mutex.Lock()
	defer wb.mutex.Unlock()

	wb.currentIndex = 0
	wb.currentWeight = 0
	wb.calculateWeights()
}

// calculateWeights calculates weight-related values
func (wb *WeightedBalancer) calculateWeights() {
	wb.mutex.Lock()
	defer wb.mutex.Unlock()

	healthyAgents, _ := wb.pool.GetHealthyAgents(context.Background())
	wb.calculateWeightsLocked(healthyAgents)
}

// calculateWeightsLocked calculates weights (must be called with mutex held)
func (wb *WeightedBalancer) calculateWeightsLocked(healthyAgents []*AgentPoolMember) {
	if len(healthyAgents) == 0 {
		wb.totalWeight = 0
		wb.maxWeight = 0
		wb.gcd = 1
		return
	}

	weights := make([]int, len(healthyAgents))
	wb.totalWeight = 0
	wb.maxWeight = 0

	for i, member := range healthyAgents {
		weight := member.Weight
		if weight <= 0 {
			weight = 1
		}
		weights[i] = weight
		wb.totalWeight += weight
		if weight > wb.maxWeight {
			wb.maxWeight = weight
		}
	}

	wb.gcd = gcdSlice(weights)
}

// RandomBalancer implements random load balancing
type RandomBalancer struct {
	pool *AgentPool
}

// NewRandomBalancer creates a new random load balancer
func NewRandomBalancer(pool *AgentPool) *RandomBalancer {
	return &RandomBalancer{
		pool: pool,
	}
}

// SelectAgent selects a random healthy agent
func (rb *RandomBalancer) SelectAgent(ctx context.Context) (*agent.Agent, error) {
	healthyAgents, err := rb.pool.GetHealthyAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy agents: %w", err)
	}

	if len(healthyAgents) == 0 {
		return nil, fmt.Errorf("no healthy agents available")
	}

	// Select random agent
	selectedIndex := rand.Intn(len(healthyAgents))
	return healthyAgents[selectedIndex].Agent, nil
}

// GetStrategy returns the random strategy
func (rb *RandomBalancer) GetStrategy() LoadBalancingStrategy {
	return LoadBalancingRandom
}

// Reset resets the random balancer state
func (rb *RandomBalancer) Reset() {
	// No state to reset for random
}

// Helper functions

// gcd calculates the greatest common divisor of two numbers
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// gcdSlice calculates the GCD of a slice of integers
func gcdSlice(numbers []int) int {
	if len(numbers) == 0 {
		return 1
	}

	result := numbers[0]
	for i := 1; i < len(numbers); i++ {
		result = gcd(result, numbers[i])
		if result == 1 {
			break
		}
	}

	if result == 0 {
		return 1
	}
	return result
}
