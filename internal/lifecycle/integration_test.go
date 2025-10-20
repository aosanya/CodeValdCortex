//go:build integration
// +build integration

package lifecycle

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest sets up a test database connection
func setupIntegrationTest(t *testing.T) (*database.ArangoClient, *registry.Repository, func()) {
	// Use test database
	cfg := &config.DatabaseConfig{
		Host:     getEnv("CVXC_DATABASE_HOST", "localhost"),
		Port:     8529,
		Database: getEnv("CVXC_DATABASE_NAME", "codevaldcortex_test"),
		Username: getEnv("CVXC_DATABASE_USERNAME", "root"),
		Password: getEnv("CVXC_DATABASE_PASSWORD", "openSesame"),
	}

	// Create database client
	dbClient, err := database.NewArangoClient(cfg)
	require.NoError(t, err, "Failed to connect to test database")

	// Create repository
	repo, err := registry.NewRepository(dbClient)
	require.NoError(t, err, "Failed to create repository")

	// Cleanup function
	cleanup := func() {
		// Clean up all test agents
		ctx := context.Background()
		agents, _ := repo.List(ctx)
		for _, a := range agents {
			_ = repo.Delete(ctx, a.ID)
		}
		dbClient.Close()
	}

	return dbClient, repo, cleanup
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestIntegration_FullLifecycle(t *testing.T) {
	_, repo, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewManager(repo)

	// Create agent
	cfg := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      10,
		HeartbeatInterval:  5 * time.Second,
		TaskTimeout:        30 * time.Second,
	}

	a, err := manager.Create(ctx, "integration-test-agent", "worker", cfg)
	require.NoError(t, err)
	assert.Equal(t, agent.StateCreated, a.State)

	// Verify persistence
	retrieved, err := repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, a.ID, retrieved.ID)
	assert.Equal(t, agent.StateCreated, retrieved.State)

	// Start agent
	err = manager.Start(ctx, a.ID)
	require.NoError(t, err)

	// Verify state in database
	retrieved, err = repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, retrieved.State)

	// Pause agent
	err = manager.Pause(ctx, a.ID)
	require.NoError(t, err)

	// Verify pause in database
	retrieved, err = repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StatePaused, retrieved.State)

	// Resume agent
	err = manager.Resume(ctx, a.ID)
	require.NoError(t, err)

	// Verify resume in database
	retrieved, err = repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, retrieved.State)

	// Stop agent
	err = manager.Stop(ctx, a.ID)
	require.NoError(t, err)

	// Verify stop in database
	retrieved, err = repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateStopped, retrieved.State)

	// Delete agent
	err = manager.Delete(ctx, a.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(ctx, a.ID)
	assert.Error(t, err, "Agent should be deleted")
}

func TestIntegration_ConcurrentAgents(t *testing.T) {
	_, repo, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewManager(repo)

	cfg := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      10,
		HeartbeatInterval:  5 * time.Second,
		TaskTimeout:        30 * time.Second,
	}

	// Create multiple agents
	numAgents := 10
	agents := make([]*agent.Agent, numAgents)

	for i := 0; i < numAgents; i++ {
		a, err := manager.Create(ctx, fmt.Sprintf("concurrent-agent-%d", i), "worker", cfg)
		require.NoError(t, err)
		agents[i] = a
	}

	// Start all agents concurrently
	errChan := make(chan error, numAgents)
	for _, a := range agents {
		go func(agentID string) {
			errChan <- manager.Start(ctx, agentID)
		}(a.ID)
	}

	// Wait for all starts to complete
	for i := 0; i < numAgents; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify all agents are running
	for _, a := range agents {
		retrieved, err := repo.Get(ctx, a.ID)
		require.NoError(t, err)
		assert.Equal(t, agent.StateRunning, retrieved.State)
	}

	// Stop all agents concurrently
	for _, a := range agents {
		go func(agentID string) {
			errChan <- manager.Stop(ctx, agentID)
		}(a.ID)
	}

	// Wait for all stops to complete
	for i := 0; i < numAgents; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify all agents are stopped
	for _, a := range agents {
		retrieved, err := repo.Get(ctx, a.ID)
		require.NoError(t, err)
		assert.Equal(t, agent.StateStopped, retrieved.State)
	}

	// Cleanup
	for _, a := range agents {
		err := manager.Delete(ctx, a.ID)
		assert.NoError(t, err)
	}
}

func TestIntegration_RestartWithPersistence(t *testing.T) {
	_, repo, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewManager(repo)

	cfg := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      10,
		HeartbeatInterval:  5 * time.Second,
		TaskTimeout:        30 * time.Second,
	}

	// Create and start agent
	a, err := manager.Create(ctx, "restart-test-agent", "worker", cfg)
	require.NoError(t, err)

	err = manager.Start(ctx, a.ID)
	require.NoError(t, err)

	// Simulate application restart by creating new manager instance
	manager2 := NewManager(repo)

	// Restart agent using new manager instance
	err = manager2.Restart(ctx, a.ID)
	require.NoError(t, err)

	// Verify agent is running
	retrieved, err := repo.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, agent.StateRunning, retrieved.State)

	// Cleanup
	err = manager2.Stop(ctx, a.ID)
	require.NoError(t, err)
	err = manager2.Delete(ctx, a.ID)
	require.NoError(t, err)
}

func TestIntegration_StateTransitionValidation(t *testing.T) {
	_, repo, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	manager := NewManager(repo)

	cfg := agent.Config{
		MaxConcurrentTasks: 5,
		TaskQueueSize:      10,
		HeartbeatInterval:  5 * time.Second,
		TaskTimeout:        30 * time.Second,
	}

	// Create agent
	a, err := manager.Create(ctx, "transition-test-agent", "worker", cfg)
	require.NoError(t, err)

	// Invalid transitions
	testCases := []struct {
		name      string
		action    func() error
		shouldErr bool
	}{
		{
			name:      "Cannot pause created agent",
			action:    func() error { return manager.Pause(ctx, a.ID) },
			shouldErr: true,
		},
		{
			name:      "Cannot resume created agent",
			action:    func() error { return manager.Resume(ctx, a.ID) },
			shouldErr: true,
		},
		{
			name:      "Can start created agent",
			action:    func() error { return manager.Start(ctx, a.ID) },
			shouldErr: false,
		},
		{
			name:      "Cannot start running agent",
			action:    func() error { return manager.Start(ctx, a.ID) },
			shouldErr: true,
		},
		{
			name:      "Can pause running agent",
			action:    func() error { return manager.Pause(ctx, a.ID) },
			shouldErr: false,
		},
		{
			name:      "Cannot pause paused agent",
			action:    func() error { return manager.Pause(ctx, a.ID) },
			shouldErr: true,
		},
		{
			name:      "Can resume paused agent",
			action:    func() error { return manager.Resume(ctx, a.ID) },
			shouldErr: false,
		},
		{
			name:      "Can stop running agent",
			action:    func() error { return manager.Stop(ctx, a.ID) },
			shouldErr: false,
		},
		{
			name:      "Cannot resume stopped agent",
			action:    func() error { return manager.Resume(ctx, a.ID) },
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action()
			if tc.shouldErr {
				assert.Error(t, err, "Expected error for: %s", tc.name)
			} else {
				assert.NoError(t, err, "Unexpected error for: %s", tc.name)
			}
		})
	}

	// Cleanup
	err = manager.Delete(ctx, a.ID)
	assert.NoError(t, err)
}
