package registry

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTypeRepository(t *testing.T) {
	repo := NewInMemoryAgentTypeRepository()
	ctx := context.Background()

	t.Run("Create and Get", func(t *testing.T) {
		agentType := &AgentType{
			ID:          "test-type",
			Name:        "Test Type",
			Description: "A test agent type",
			Category:    "test",
			Version:     "1.0.0",
			IsEnabled:   true,
		}

		err := repo.Create(ctx, agentType)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, "test-type")
		require.NoError(t, err)
		assert.Equal(t, "test-type", retrieved.ID)
		assert.Equal(t, "Test Type", retrieved.Name)
	})

	t.Run("Duplicate Create", func(t *testing.T) {
		agentType := &AgentType{
			ID:       "duplicate",
			Name:     "Duplicate",
			Category: "test",
			Version:  "1.0.0",
		}

		err := repo.Create(ctx, agentType)
		require.NoError(t, err)

		err = repo.Create(ctx, agentType)
		assert.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		types, err := repo.List(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(types), 2)
	})

	t.Run("List By Category", func(t *testing.T) {
		types, err := repo.ListByCategory(ctx, "test")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(types), 2)
	})

	t.Run("Update", func(t *testing.T) {
		agentType, err := repo.Get(ctx, "test-type")
		require.NoError(t, err)

		agentType.Description = "Updated description"
		err = repo.Update(ctx, agentType)
		require.NoError(t, err)

		updated, err := repo.Get(ctx, "test-type")
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updated.Description)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, "duplicate")
		require.NoError(t, err)

		_, err = repo.Get(ctx, "duplicate")
		assert.Error(t, err)
	})
}

func TestAgentTypeService(t *testing.T) {
	repo := NewInMemoryAgentTypeRepository()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	service := NewAgentTypeService(repo, logger)
	ctx := context.Background()

	t.Run("Register Valid Type", func(t *testing.T) {
		agentType := &AgentType{
			ID:          "service-test",
			Name:        "Service Test",
			Description: "Test type for service",
			Category:    "test",
			Version:     "1.0.0",
			IsEnabled:   true,
		}

		err := service.RegisterType(ctx, agentType)
		require.NoError(t, err)
	})

	t.Run("Register Invalid Type - No ID", func(t *testing.T) {
		agentType := &AgentType{
			Name:     "No ID",
			Category: "test",
			Version:  "1.0.0",
		}

		err := service.RegisterType(ctx, agentType)
		assert.Error(t, err)
	})

	t.Run("Register Invalid Type - No Name", func(t *testing.T) {
		agentType := &AgentType{
			ID:       "no-name",
			Category: "test",
			Version:  "1.0.0",
		}

		err := service.RegisterType(ctx, agentType)
		assert.Error(t, err)
	})

	t.Run("IsValidType", func(t *testing.T) {
		isValid, err := service.IsValidType(ctx, "service-test")
		require.NoError(t, err)
		assert.True(t, isValid)

		isValid, err = service.IsValidType(ctx, "non-existent")
		require.NoError(t, err)
		assert.False(t, isValid)
	})

	t.Run("Enable and Disable Type", func(t *testing.T) {
		err := service.DisableType(ctx, "service-test")
		require.NoError(t, err)

		isValid, err := service.IsValidType(ctx, "service-test")
		require.NoError(t, err)
		assert.False(t, isValid)

		err = service.EnableType(ctx, "service-test")
		require.NoError(t, err)

		isValid, err = service.IsValidType(ctx, "service-test")
		require.NoError(t, err)
		assert.True(t, isValid)
	})
}

func TestDefaultAgentTypes(t *testing.T) {
	repo := NewInMemoryAgentTypeRepository()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	service := NewAgentTypeService(repo, logger)
	ctx := context.Background()

	t.Run("Initialize Default Types", func(t *testing.T) {
		err := InitializeDefaultAgentTypes(ctx, service, logger)
		require.NoError(t, err)

		// Check core types (only 5 core types in framework defaults now)
		coreTypes := []string{"worker", "coordinator", "monitor", "proxy", "gateway"}
		for _, typeID := range coreTypes {
			agentType, err := service.GetType(ctx, typeID)
			require.NoError(t, err)
			assert.Equal(t, typeID, agentType.ID)
			assert.True(t, agentType.IsSystemType)
			assert.True(t, agentType.IsEnabled)
			assert.Equal(t, "core", agentType.Category)
		}
	})

	t.Run("List Types By Category", func(t *testing.T) {
		coreTypes, err := service.ListTypesByCategory(ctx, "core")
		require.NoError(t, err)
		assert.Equal(t, 5, len(coreTypes))
	})
}
