package configuration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/configuration"
	"github.com/aosanya/CodeValdCortex/internal/templates"
)

// MockResourceChecker implements ResourceChecker for testing
type MockResourceChecker struct{}

func (m *MockResourceChecker) CheckCPUAvailability(cpu string) error {
	return nil // Always available for testing
}

func (m *MockResourceChecker) CheckMemoryAvailability(memory string) error {
	return nil // Always available for testing
}

func (m *MockResourceChecker) CheckStorageAvailability(storage string) error {
	return nil // Always available for testing
}

// MockNotificationService implements NotificationService for testing
type MockNotificationService struct{}

func (m *MockNotificationService) NotifyConfigurationCreated(config *configuration.AgentConfiguration) error {
	fmt.Printf("Configuration created: %s\n", config.Name)
	return nil
}

func (m *MockNotificationService) NotifyConfigurationUpdated(oldConfig, newConfig *configuration.AgentConfiguration) error {
	fmt.Printf("Configuration updated: %s (version %s -> %s)\n", newConfig.Name, oldConfig.Version, newConfig.Version)
	return nil
}

func (m *MockNotificationService) NotifyConfigurationDeleted(configID string) error {
	fmt.Printf("Configuration deleted: %s\n", configID)
	return nil
}

func (m *MockNotificationService) NotifyConfigurationApplied(agentID, configID string) error {
	fmt.Printf("Configuration %s applied to agent %s\n", configID, agentID)
	return nil
}

// MockCache implements Cache for testing
type MockCache struct {
	data map[string]*configuration.AgentConfiguration
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]*configuration.AgentConfiguration),
	}
}

func (c *MockCache) Get(id string) (*configuration.AgentConfiguration, bool) {
	config, exists := c.data[id]
	return config, exists
}

func (c *MockCache) Set(id string, config *configuration.AgentConfiguration, ttl time.Duration) {
	c.data[id] = config
}

func (c *MockCache) Delete(id string) {
	delete(c.data, id)
}

func (c *MockCache) Clear() {
	c.data = make(map[string]*configuration.AgentConfiguration)
}

// TestConfigurationUsage demonstrates how to use the configuration management system
func TestConfigurationUsage(t *testing.T) {
	ctx := context.Background()

	// Set up dependencies
	repo := configuration.NewInMemoryRepository()
	resourceChecker := &MockResourceChecker{}
	validator := configuration.NewDefaultValidator(resourceChecker)
	notifier := &MockNotificationService{}
	cache := NewMockCache()

	// Create configuration service
	configService := configuration.NewService(repo, validator, notifier, cache)

	// Set up template system
	templateRepo := templates.NewInMemoryRepository()
	templateValidator := templates.NewDefaultValidator()
	templateEngine := templates.NewEngine(templateRepo, templateValidator)

	// Example 1: Create configuration from template
	fmt.Println("=== Creating Configuration from Template ===")

	templateVariables := map[string]interface{}{
		"agent_name":           "worker-001",
		"max_concurrent_tasks": 10,
		"task_queue_size":      200,
		"heartbeat_interval":   "20s",
		"cpu_millicores":       250,
		"memory_mb":            256,
		"environment":          "production",
		"health_check_port":    8080,
		"metrics_port":         9090,
		"replicas":             2,
		"cpu_request":          "150m",
		"memory_request":       "200Mi",
		"cpu_limit":            "300m",
		"memory_limit":         "400Mi",
	}

	config, err := templateEngine.RenderTemplate(ctx, "template-basic-worker", templateVariables)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return
	}

	// Store the configuration
	storedConfig, err := configService.CreateConfiguration(ctx, config)
	if err != nil {
		fmt.Printf("Error creating configuration: %v\n", err)
		return
	}

	fmt.Printf("Created configuration: %s (ID: %s)\n", storedConfig.Name, storedConfig.ID)

	// Example 2: Create high-performance configuration
	fmt.Println("\n=== Creating High-Performance Configuration ===")

	highPerfVariables := map[string]interface{}{
		"agent_name":           "worker-high-perf-001",
		"max_concurrent_tasks": 25,
		"task_queue_size":      2000,
		"heartbeat_interval":   "10s",
		"cpu_millicores":       2000,
		"memory_mb":            2048,
		"environment":          "production",
		"health_check_port":    8080,
		"metrics_port":         9090,
		"max_memory_usage":     1800,
		"gc_interval":          "20s",
		"persistence_enabled":  true,
		"sync_interval":        "45s",
		"replicas":             3,
		"cpu_request":          "1500m",
		"memory_request":       "1.5Gi",
		"cpu_limit":            "2500m",
		"memory_limit":         "2.5Gi",
	}

	highPerfConfig, err := templateEngine.RenderTemplate(ctx, "template-high-perf-worker", highPerfVariables)
	if err != nil {
		fmt.Printf("Error rendering high-performance template: %v\n", err)
		return
	}

	storedHighPerfConfig, err := configService.CreateConfiguration(ctx, highPerfConfig)
	if err != nil {
		fmt.Printf("Error creating high-performance configuration: %v\n", err)
		return
	}

	fmt.Printf("Created high-performance configuration: %s (ID: %s)\n", storedHighPerfConfig.Name, storedHighPerfConfig.ID)

	// Example 3: Create coordinator configuration
	fmt.Println("\n=== Creating Coordinator Configuration ===")

	coordinatorVariables := map[string]interface{}{
		"agent_name":           "coordinator-001",
		"max_concurrent_tasks": 8,
		"task_queue_size":      300,
		"heartbeat_interval":   "15s",
		"cpu_millicores":       400,
		"memory_mb":            384,
		"environment":          "production",
		"health_check_port":    8080,
		"metrics_port":         9090,
		"message_queue_size":   800,
		"connection_timeout":   "20s",
		"read_timeout":         "25s",
		"write_timeout":        "25s",
		"max_retries":          2,
		"replicas":             2,
		"cpu_request":          "200m",
		"memory_request":       "256Mi",
		"cpu_limit":            "500m",
		"memory_limit":         "512Mi",
	}

	coordinatorConfig, err := templateEngine.RenderTemplate(ctx, "template-coordinator", coordinatorVariables)
	if err != nil {
		fmt.Printf("Error rendering coordinator template: %v\n", err)
		return
	}

	storedCoordinatorConfig, err := configService.CreateConfiguration(ctx, coordinatorConfig)
	if err != nil {
		fmt.Printf("Error creating coordinator configuration: %v\n", err)
		return
	}

	fmt.Printf("Created coordinator configuration: %s (ID: %s)\n", storedCoordinatorConfig.Name, storedCoordinatorConfig.ID)

	// Example 4: List configurations
	fmt.Println("\n=== Listing Configurations ===")

	configs, err := configService.ListConfigurations(ctx, &configuration.ListFilter{
		SortBy:    "created_at",
		SortOrder: "desc",
	})
	if err != nil {
		fmt.Printf("Error listing configurations: %v\n", err)
		return
	}

	for i, cfg := range configs {
		fmt.Printf("%d. %s (Type: %s, ID: %s)\n", i+1, cfg.Name, cfg.AgentType, cfg.ID)
	}

	// Example 5: Export and clone configuration
	fmt.Println("\n=== Exporting and Cloning Configuration ===")

	exportData, err := configService.ExportConfiguration(ctx, storedConfig.ID)
	if err != nil {
		fmt.Printf("Error exporting configuration: %v\n", err)
		return
	}

	fmt.Printf("Exported configuration size: %d bytes\n", len(exportData))

	// Clone the configuration
	clonedConfig, err := configService.CloneConfiguration(ctx, storedConfig.ID, "worker-001-clone")
	if err != nil {
		fmt.Printf("Error cloning configuration: %v\n", err)
		return
	}

	fmt.Printf("Cloned configuration: %s (ID: %s)\n", clonedConfig.Name, clonedConfig.ID)

	// Example 6: Update configuration
	fmt.Println("\n=== Updating Configuration ===")

	clonedConfig.Description = "Updated clone of worker configuration"
	clonedConfig.RuntimeConfig.Logging.Level = "debug"
	clonedConfig.DeploymentConfig.Replicas = 3

	updatedConfig, err := configService.UpdateConfiguration(ctx, clonedConfig)
	if err != nil {
		fmt.Printf("Error updating configuration: %v\n", err)
		return
	}

	fmt.Printf("Updated configuration: %s\n", updatedConfig.Name)

	// Example 7: Search by labels
	fmt.Println("\n=== Searching by Labels ===")

	workerConfigs, err := configService.GetConfigurationsByLabels(ctx, map[string]string{
		"type": "worker",
	})
	if err != nil {
		fmt.Printf("Error searching configurations by labels: %v\n", err)
		return
	}

	fmt.Printf("Found %d worker configurations:\n", len(workerConfigs))
	for i, cfg := range workerConfigs {
		fmt.Printf("%d. %s\n", i+1, cfg.Name)
	}

	// Example 8: Apply configuration to agent (mock)
	fmt.Println("\n=== Applying Configuration to Agent ===")

	err = configService.ApplyConfiguration(ctx, "agent-001", storedConfig.ID)
	if err != nil {
		fmt.Printf("Error applying configuration: %v\n", err)
		return
	}

	fmt.Println("Configuration applied successfully!")
}

// TestBasicConfigurationFlow tests the basic configuration management flow
func TestBasicConfigurationFlow(t *testing.T) {
	ctx := context.Background()

	// Set up dependencies
	repo := configuration.NewInMemoryRepository()
	resourceChecker := &MockResourceChecker{}
	validator := configuration.NewDefaultValidator(resourceChecker)
	notifier := &MockNotificationService{}
	cache := NewMockCache()

	// Create configuration service
	service := configuration.NewService(repo, validator, notifier, cache)

	// Create a test configuration
	config := &configuration.AgentConfiguration{
		Name:      "test-config",
		AgentType: "worker",
		BaseConfig: agent.Config{
			MaxConcurrentTasks: 5,
			TaskQueueSize:      100,
			HeartbeatInterval:  30 * time.Second,
			TaskTimeout:        5 * time.Minute,
			Resources: agent.Resources{
				CPU:      100,
				Memory:   128,
				MaxTasks: 1000,
			},
		},
		RuntimeConfig: configuration.RuntimeConfiguration{
			AutoRestart: true,
			RestartPolicy: configuration.RestartPolicy{
				Policy:            "OnFailure",
				MaxRetries:        3,
				BackoffMultiplier: 2.0,
				InitialDelay:      1 * time.Second,
				MaxDelay:          1 * time.Minute,
			},
			HealthCheck: configuration.HealthCheckConfig{
				Enabled:          true,
				Path:             "/health",
				Port:             8080,
				Interval:         30 * time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
				SuccessThreshold: 1,
				InitialDelay:     10 * time.Second,
			},
			Logging: configuration.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			Metrics: configuration.MetricsConfig{
				Enabled:  true,
				Port:     9090,
				Path:     "/metrics",
				Interval: 30 * time.Second,
			},
		},
		DeploymentConfig: configuration.DeploymentConfiguration{
			Strategy: configuration.DeploymentStrategyRollingUpdate,
			Replicas: 1,
			Resources: configuration.ResourceRequirements{
				Requests: configuration.ResourceList{
					CPU:    "100m",
					Memory: "128Mi",
				},
				Limits: configuration.ResourceList{
					CPU:    "200m",
					Memory: "256Mi",
				},
			},
		},
		Labels: map[string]string{
			"type":        "worker",
			"environment": "test",
		},
	}

	// Test create
	storedConfig, err := service.CreateConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create configuration: %v", err)
	}

	if storedConfig.ID == "" {
		t.Error("Configuration ID should not be empty")
	}

	if storedConfig.Version == "" {
		t.Error("Configuration version should not be empty")
	}

	// Test get
	retrievedConfig, err := service.GetConfiguration(ctx, storedConfig.ID)
	if err != nil {
		t.Fatalf("Failed to get configuration: %v", err)
	}

	if retrievedConfig.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, retrievedConfig.Name)
	}

	// Test update
	retrievedConfig.Description = "Updated test configuration"
	updatedConfig, err := service.UpdateConfiguration(ctx, retrievedConfig)
	if err != nil {
		t.Fatalf("Failed to update configuration: %v", err)
	}

	if updatedConfig.Description != "Updated test configuration" {
		t.Error("Configuration description was not updated")
	}

	// Test list
	configs, err := service.ListConfigurations(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list configurations: %v", err)
	}

	if len(configs) == 0 {
		t.Error("Expected at least one configuration")
	}

	// Test delete
	err = service.DeleteConfiguration(ctx, storedConfig.ID)
	if err != nil {
		t.Fatalf("Failed to delete configuration: %v", err)
	}

	// Verify deletion
	_, err = service.GetConfiguration(ctx, storedConfig.ID)
	if err == nil {
		t.Error("Expected error when getting deleted configuration")
	}
}

// PrintConfigurationJSON prints a configuration as formatted JSON
func PrintConfigurationJSON(config *configuration.AgentConfiguration) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling configuration: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
