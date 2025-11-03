package registry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// InitializeDefaultRoles registers default roles in the system
func InitializeDefaultRoles(ctx context.Context, service RoleService, logger *logrus.Logger) error {
	logger.Info("Initializing default roles")

	defaultTypes := getDefaultRoles()

	for _, agentType := range defaultTypes {
		if err := service.RegisterType(ctx, agentType); err != nil {
			logger.WithError(err).Warnf("Failed to register role %s (may already exist)", agentType.ID)
			// Continue with other types even if one fails
		} else {
			logger.WithField("type_id", agentType.ID).Info("Registered role")
		}
	}

	logger.Infof("Initialized %d default roles", len(defaultTypes))
	return nil
}

// getDefaultRoles returns the default system roles
func getDefaultRoles() []*Role {
	now := time.Now()

	return []*Role{
		// Core System Types
		{
			ID:          "worker",
			Name:        "Worker Agent",
			Description: "General-purpose worker agent for task execution",
			Category:    "core",
			Version:     "1.0.0",
			Schema:      getWorkerSchema(),
			RequiredCapabilities: []string{
				"task_execution",
				"heartbeat",
			},
			OptionalCapabilities: []string{
				"metrics_reporting",
				"log_streaming",
			},
			DefaultConfig: map[string]interface{}{
				"max_concurrent_tasks": 10,
				"task_queue_size":      100,
				"heartbeat_interval":   "30s",
			},
			IsSystemType: true,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
			CreatedBy:    "system",
		},
		{
			ID:          "coordinator",
			Name:        "Coordinator Agent",
			Description: "Coordinator agent for orchestrating other agents",
			Category:    "core",
			Version:     "1.0.0",
			Schema:      getCoordinatorSchema(),
			RequiredCapabilities: []string{
				"agent_management",
				"task_distribution",
				"monitoring",
			},
			OptionalCapabilities: []string{
				"auto_scaling",
				"load_balancing",
			},
			DefaultConfig: map[string]interface{}{
				"max_managed_agents":    50,
				"health_check_interval": "15s",
			},
			IsSystemType: true,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
			CreatedBy:    "system",
		},
		{
			ID:          "monitor",
			Name:        "Monitor Agent",
			Description: "Monitoring and observability agent",
			Category:    "core",
			Version:     "1.0.0",
			Schema:      getMonitorSchema(),
			RequiredCapabilities: []string{
				"metrics_collection",
				"health_monitoring",
			},
			OptionalCapabilities: []string{
				"alerting",
				"log_aggregation",
			},
			DefaultConfig: map[string]interface{}{
				"collection_interval": "10s",
				"retention_period":    "24h",
			},
			IsSystemType: true,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
			CreatedBy:    "system",
		},
		{
			ID:          "proxy",
			Name:        "Proxy Agent",
			Description: "Proxy agent for external system integration",
			Category:    "core",
			Version:     "1.0.0",
			Schema:      getProxySchema(),
			RequiredCapabilities: []string{
				"request_forwarding",
				"response_handling",
			},
			OptionalCapabilities: []string{
				"caching",
				"rate_limiting",
			},
			DefaultConfig: map[string]interface{}{
				"timeout":         "30s",
				"max_connections": 100,
			},
			IsSystemType: true,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
			CreatedBy:    "system",
		},
		{
			ID:          "gateway",
			Name:        "Gateway Agent",
			Description: "API gateway agent for external access",
			Category:    "core",
			Version:     "1.0.0",
			Schema:      getGatewaySchema(),
			RequiredCapabilities: []string{
				"api_routing",
				"authentication",
				"authorization",
			},
			OptionalCapabilities: []string{
				"rate_limiting",
				"request_validation",
				"response_transformation",
			},
			DefaultConfig: map[string]interface{}{
				"port":           8080,
				"rate_limit_rps": 1000,
			},
			IsSystemType: true,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
			CreatedBy:    "system",
		},
	}
}

// Schema definitions for each role
// These use JSON Schema format for validation

func getWorkerSchema() json.RawMessage {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"max_concurrent_tasks": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
				"maximum": 100,
			},
			"task_queue_size": map[string]interface{}{
				"type":    "integer",
				"minimum": 10,
			},
			"heartbeat_interval": map[string]interface{}{
				"type":    "string",
				"pattern": "^[0-9]+(s|m|h)$",
			},
		},
	}
	bytes, _ := json.Marshal(schema)
	return bytes
}

func getCoordinatorSchema() json.RawMessage {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"max_managed_agents": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
			},
			"health_check_interval": map[string]interface{}{
				"type":    "string",
				"pattern": "^[0-9]+(s|m|h)$",
			},
		},
	}
	bytes, _ := json.Marshal(schema)
	return bytes
}

func getMonitorSchema() json.RawMessage {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"collection_interval": map[string]interface{}{
				"type":    "string",
				"pattern": "^[0-9]+(s|m|h)$",
			},
			"retention_period": map[string]interface{}{
				"type":    "string",
				"pattern": "^[0-9]+(h|d)$",
			},
		},
	}
	bytes, _ := json.Marshal(schema)
	return bytes
}

func getProxySchema() json.RawMessage {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"timeout": map[string]interface{}{
				"type":    "string",
				"pattern": "^[0-9]+(s|m)$",
			},
			"max_connections": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
			},
		},
	}
	bytes, _ := json.Marshal(schema)
	return bytes
}

func getGatewaySchema() json.RawMessage {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"port": map[string]interface{}{
				"type":    "integer",
				"minimum": 1024,
				"maximum": 65535,
			},
			"rate_limit_rps": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
			},
		},
	}
	bytes, _ := json.Marshal(schema)
	return bytes
}
