package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestManagerConfig() ManagerConfig {
	return ManagerConfig{
		RepositoryConfig: RepositoryConfig{
			DatabaseURL:  "http://localhost:8529",
			DatabaseName: "test_pool_db",
			Username:     "root",
			Password:     "test",
		},
		EnableAutoScaling: false,
		MetricsInterval:   30 * time.Second,
		CleanupInterval:   1 * time.Hour,
		MetricsRetention:  24 * time.Hour,
	}
}

func TestManager_CreatePool(t *testing.T) {
	logger := &MockLogger{}
	config := createTestManagerConfig()
	manager, err := NewManager(config, logger)
	require.NoError(t, err)

	poolConfig := createTestPoolConfig("test-pool")
	ctx := context.Background()

	pool, err := manager.CreatePool(ctx, poolConfig)
	require.NoError(t, err)
	assert.NotNil(t, pool)
	assert.NotEmpty(t, pool.ID)

	// Verify pool was created
	retrievedPool, err := manager.GetPool(ctx, pool.ID)
	require.NoError(t, err)
	assert.Equal(t, poolConfig.Name, retrievedPool.Config.Name)
	assert.Equal(t, PoolStatusActive, retrievedPool.Status)
}

func TestManager_CreatePool_DuplicateName(t *testing.T) {
	logger := &MockLogger{}
	config := createTestManagerConfig()
	manager, err := NewManager(config, logger)
	require.NoError(t, err)

	poolConfig := createTestPoolConfig("test-pool")
	ctx := context.Background()

	// Create first pool
	_, err = manager.CreatePool(ctx, poolConfig)
	require.NoError(t, err)

	// Try to create pool with same name
	_, err = manager.CreatePool(ctx, poolConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_DeletePool(t *testing.T) {
	logger := &MockLogger{}
	config := createTestManagerConfig()
	manager, err := NewManager(config, logger)
	require.NoError(t, err)

	poolConfig := createTestPoolConfig("test-pool")
	ctx := context.Background()

	// Create pool
	pool, err := manager.CreatePool(ctx, poolConfig)
	require.NoError(t, err)

	// Delete pool
	err = manager.DeletePool(ctx, pool.ID)
	require.NoError(t, err)

	// Verify pool is deleted
	_, err = manager.GetPool(ctx, pool.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetPoolNotFound(t *testing.T) {
	logger := &MockLogger{}
	config := createTestManagerConfig()
	manager, err := NewManager(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = manager.GetPool(ctx, "non-existent-pool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
