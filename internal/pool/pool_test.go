package pool

import (
	"context"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}
func (m *MockLogger) Debug(msg string, fields ...interface{}) {}

func createTestPoolConfig(name string) PoolConfig {
	return PoolConfig{
		Name:                  name,
		Description:           "Test pool",
		LoadBalancingStrategy: LoadBalancingRoundRobin,
		MinAgents:             0,
		MaxAgents:             10,
		HealthCheckInterval:   30 * time.Second,
		ResourceLimits: ResourceLimits{
			TotalCPU:           1000,
			TotalMemory:        1024,
			MaxConcurrentTasks: 50,
		},
		AutoScaling: AutoScalingConfig{
			Enabled:            false,
			ScaleUpThreshold:   80.0,
			ScaleDownThreshold: 20.0,
		},
	}
}

func createTestAgent(name string) *agent.Agent {
	config := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      10,
		HeartbeatInterval:  30 * time.Second,
		TaskTimeout:        60 * time.Second,
		Resources: agent.Resources{
			CPU:      100,
			Memory:   256,
			MaxTasks: 10,
		},
	}

	return agent.New(name, "test", config)
}

func TestAgentPool_NewPool(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}

	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	assert.NotEmpty(t, pool.ID)
	assert.Equal(t, config.Name, pool.Config.Name)
	assert.Equal(t, LoadBalancingRoundRobin, pool.Config.LoadBalancingStrategy)
	assert.Equal(t, PoolStatusActive, pool.Status)
	assert.Empty(t, pool.Members)
}

func TestAgentPool_NewPool_InvalidConfig(t *testing.T) {
	logger := &MockLogger{}

	// Test empty name
	config := createTestPoolConfig("")
	_, err := NewAgentPool(config, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pool name cannot be empty")

	// Test invalid agent limits
	config = createTestPoolConfig("test-pool")
	config.MinAgents = 10
	config.MaxAgents = 5
	_, err = NewAgentPool(config, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent limits")
}

func TestAgentPool_AddAgent(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent := createTestAgent("test-agent")
	ctx := context.Background()

	err = pool.AddAgent(ctx, testAgent, 1)
	require.NoError(t, err)

	assert.Len(t, pool.Members, 1)
	assert.Contains(t, pool.Members, testAgent.ID)
}

func TestAgentPool_AddAgent_Duplicate(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent := createTestAgent("test-agent")
	ctx := context.Background()

	// Add agent first time
	err = pool.AddAgent(ctx, testAgent, 1)
	require.NoError(t, err)

	// Try to add same agent again
	err = pool.AddAgent(ctx, testAgent, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in pool")
	assert.Len(t, pool.Members, 1)
}

func TestAgentPool_RemoveAgent(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent := createTestAgent("test-agent")
	ctx := context.Background()

	// Add agent first
	err = pool.AddAgent(ctx, testAgent, 1)
	require.NoError(t, err)

	// Remove agent
	err = pool.RemoveAgent(ctx, testAgent.ID)
	require.NoError(t, err)

	assert.Len(t, pool.Members, 0)
	assert.NotContains(t, pool.Members, testAgent.ID)
}

func TestAgentPool_RemoveAgent_NotFound(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	nonExistentID := uuid.New().String()

	err = pool.RemoveAgent(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pool")
}

func TestAgentPool_GetAgent(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent1 := createTestAgent("test-agent-1")
	testAgent2 := createTestAgent("test-agent-2")
	ctx := context.Background()

	// Add agents
	require.NoError(t, pool.AddAgent(ctx, testAgent1, 1))
	require.NoError(t, pool.AddAgent(ctx, testAgent2, 1))

	// Get an agent
	selectedAgent, err := pool.GetAgent(ctx)
	require.NoError(t, err)
	assert.NotNil(t, selectedAgent)
	assert.True(t, selectedAgent.ID == testAgent1.ID || selectedAgent.ID == testAgent2.ID)
}

func TestAgentPool_GetAgent_EmptyPool(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = pool.GetAgent(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no agents available")
}

func TestAgentPool_ListAgents(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent1 := createTestAgent("test-agent-1")
	testAgent2 := createTestAgent("test-agent-2")
	ctx := context.Background()

	require.NoError(t, pool.AddAgent(ctx, testAgent1, 1))
	require.NoError(t, pool.AddAgent(ctx, testAgent2, 1))

	agents, err := pool.ListAgents(ctx)
	require.NoError(t, err)
	assert.Len(t, agents, 2)
}

func TestAgentPool_GetHealthyAgents(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent1 := createTestAgent("test-agent-1")
	testAgent2 := createTestAgent("test-agent-2")
	ctx := context.Background()

	require.NoError(t, pool.AddAgent(ctx, testAgent1, 1))
	require.NoError(t, pool.AddAgent(ctx, testAgent2, 1))

	healthy, err := pool.GetHealthyAgents(ctx)
	require.NoError(t, err)
	assert.Len(t, healthy, 2)
}

func TestAgentPool_GetMetrics(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	testAgent1 := createTestAgent("test-agent-1")
	testAgent2 := createTestAgent("test-agent-2")
	ctx := context.Background()

	require.NoError(t, pool.AddAgent(ctx, testAgent1, 1))
	require.NoError(t, pool.AddAgent(ctx, testAgent2, 1))

	metrics := pool.GetMetrics(ctx)
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, metrics.TotalAgents)
	assert.Equal(t, 2, metrics.HealthyAgents)
	assert.NotZero(t, metrics.LastUpdated)
}

func TestAgentPool_Stop(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = pool.Stop(ctx)
	require.NoError(t, err)
	assert.Equal(t, PoolStatusStopped, pool.Status)
}

func TestAgentPool_UpdateConfig(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Update config
	newConfig := config
	newConfig.LoadBalancingStrategy = LoadBalancingLeastConnection
	newConfig.MaxAgents = 20

	err = pool.UpdateConfig(ctx, newConfig)
	require.NoError(t, err)

	assert.Equal(t, LoadBalancingLeastConnection, pool.Config.LoadBalancingStrategy)
	assert.Equal(t, 20, pool.Config.MaxAgents)
}

func TestAgentPool_ConcurrentAccess(t *testing.T) {
	config := createTestPoolConfig("test-pool")
	logger := &MockLogger{}
	pool, err := NewAgentPool(config, logger)
	require.NoError(t, err)

	agentCount := 5
	ctx := context.Background()

	// Add agents concurrently
	errChan := make(chan error, agentCount)
	for i := 0; i < agentCount; i++ {
		go func(index int) {
			testAgent := createTestAgent("test-agent")
			err := pool.AddAgent(ctx, testAgent, 1)
			errChan <- err
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < agentCount; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	assert.Len(t, pool.Members, agentCount)
}
