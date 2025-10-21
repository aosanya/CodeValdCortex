package templates

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryRepository implements the Repository interface using in-memory storage
type InMemoryRepository struct {
	templates map[string]*Template
	mu        sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory template repository
func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		templates: make(map[string]*Template),
	}

	// Initialize with some default templates
	repo.initializeDefaultTemplates()

	return repo
}

// Store saves a template
func (r *InMemoryRepository) Store(ctx context.Context, template *Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates[template.ID] = template
	return nil
}

// Get retrieves a template by ID
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	template, exists := r.templates[id]
	if !exists {
		return nil, &TemplateNotFoundError{ID: id}
	}

	return template, nil
}

// List retrieves templates with optional filtering
func (r *InMemoryRepository) List(ctx context.Context, filter *ListFilter) ([]*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var templates []*Template
	for _, template := range r.templates {
		if r.matchesFilter(template, filter) {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

// Update updates an existing template
func (r *InMemoryRepository) Update(ctx context.Context, template *Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[template.ID]; !exists {
		return &TemplateNotFoundError{ID: template.ID}
	}

	r.templates[template.ID] = template
	return nil
}

// Delete removes a template
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[id]; !exists {
		return &TemplateNotFoundError{ID: id}
	}

	delete(r.templates, id)
	return nil
}

// GetByLabels retrieves templates matching labels
func (r *InMemoryRepository) GetByLabels(ctx context.Context, labels map[string]string) ([]*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var templates []*Template
	for _, template := range r.templates {
		if r.matchesLabels(template, labels) {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

// Helper methods

func (r *InMemoryRepository) matchesFilter(template *Template, filter *ListFilter) bool {
	if filter == nil {
		return true
	}

	if filter.AgentType != "" && template.AgentType != filter.AgentType {
		return false
	}

	if filter.CreatedAfter != nil && template.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	if filter.CreatedBefore != nil && template.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	if !r.matchesLabels(template, filter.Labels) {
		return false
	}

	return true
}

func (r *InMemoryRepository) matchesLabels(template *Template, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}

	for key, value := range labels {
		if templateValue, exists := template.Labels[key]; !exists || templateValue != value {
			return false
		}
	}

	return true
}

// initializeDefaultTemplates creates some default templates for common use cases
func (r *InMemoryRepository) initializeDefaultTemplates() {
	now := time.Now()

	// Basic Worker Agent Template
	workerTemplate := &Template{
		ID:          "template-basic-worker",
		Name:        "Basic Worker Agent",
		Description: "A basic worker agent configuration template",
		Version:     "1.0.0",
		AgentType:   "worker",
		Content: `{
  "id": "{{.agent_id}}",
  "name": "{{.agent_name}}",
  "agent_type": "worker",
  "base_config": {
    "max_concurrent_tasks": {{.max_concurrent_tasks}},
    "task_queue_size": {{.task_queue_size}},
    "heartbeat_interval": "{{.heartbeat_interval}}",
    "task_timeout": "{{.task_timeout}}",
    "resources": {
      "cpu": {{.cpu_millicores}},
      "memory": {{.memory_mb}},
      "max_tasks": {{.max_tasks}}
    }
  },
  "runtime_config": {
    "auto_restart": {{.auto_restart}},
    "restart_policy": {
      "policy": "{{.restart_policy}}",
      "max_retries": {{.max_retries}},
      "backoff_multiplier": {{.backoff_multiplier}},
      "initial_delay": "{{.initial_delay}}",
      "max_delay": "{{.max_delay}}"
    },
    "health_check": {
      "enabled": {{.health_check_enabled}},
      "path": "{{.health_check_path}}",
      "port": {{.health_check_port}},
      "interval": "{{.health_check_interval}}",
      "timeout": "{{.health_check_timeout}}",
      "failure_threshold": {{.failure_threshold}},
      "success_threshold": {{.success_threshold}},
      "initial_delay": "{{.health_check_initial_delay}}"
    },
    "logging": {
      "level": "{{.log_level}}",
      "format": "{{.log_format}}",
      "output": "{{.log_output}}"
    },
    "metrics": {
      "enabled": {{.metrics_enabled}},
      "port": {{.metrics_port}},
      "path": "{{.metrics_path}}",
      "interval": "{{.metrics_interval}}"
    }
  },
  "deployment_config": {
    "strategy": "{{.deployment_strategy}}",
    "replicas": {{.replicas}},
    "resources": {
      "requests": {
        "cpu": "{{.cpu_request}}",
        "memory": "{{.memory_request}}"
      },
      "limits": {
        "cpu": "{{.cpu_limit}}",
        "memory": "{{.memory_limit}}"
      }
    }
  },
  "labels": {
    "type": "worker",
    "environment": "{{.environment}}"
  }
}`,
		Variables: []TemplateVariable{
			{Name: "agent_id", Type: "string", Description: "Unique agent identifier", Required: false},
			{Name: "agent_name", Type: "string", Description: "Human-readable agent name", Required: true},
			{Name: "max_concurrent_tasks", Type: "int", Description: "Maximum concurrent tasks", DefaultValue: 5, MinValue: &[]float64{1}[0], MaxValue: &[]float64{100}[0]},
			{Name: "task_queue_size", Type: "int", Description: "Task queue buffer size", DefaultValue: 100, MinValue: &[]float64{10}[0]},
			{Name: "heartbeat_interval", Type: "string", Description: "Heartbeat interval duration", DefaultValue: "30s"},
			{Name: "task_timeout", Type: "string", Description: "Default task timeout", DefaultValue: "5m"},
			{Name: "cpu_millicores", Type: "int", Description: "CPU allocation in millicores", DefaultValue: 100, MinValue: &[]float64{50}[0]},
			{Name: "memory_mb", Type: "int", Description: "Memory allocation in MB", DefaultValue: 128, MinValue: &[]float64{64}[0]},
			{Name: "max_tasks", Type: "int", Description: "Maximum tasks in queue", DefaultValue: 1000, MinValue: &[]float64{100}[0]},
			{Name: "auto_restart", Type: "bool", Description: "Enable automatic restart", DefaultValue: true},
			{Name: "restart_policy", Type: "string", Description: "Restart policy", DefaultValue: "OnFailure", ValidValues: []interface{}{"Always", "OnFailure", "Never"}},
			{Name: "max_retries", Type: "int", Description: "Maximum restart retries", DefaultValue: 3, MinValue: &[]float64{0}[0]},
			{Name: "backoff_multiplier", Type: "float", Description: "Exponential backoff multiplier", DefaultValue: 2.0, MinValue: &[]float64{1.0}[0]},
			{Name: "initial_delay", Type: "string", Description: "Initial restart delay", DefaultValue: "1s"},
			{Name: "max_delay", Type: "string", Description: "Maximum restart delay", DefaultValue: "1m"},
			{Name: "health_check_enabled", Type: "bool", Description: "Enable health checks", DefaultValue: true},
			{Name: "health_check_path", Type: "string", Description: "Health check endpoint path", DefaultValue: "/health"},
			{Name: "health_check_port", Type: "int", Description: "Health check port", DefaultValue: 8080, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "health_check_interval", Type: "string", Description: "Health check interval", DefaultValue: "30s"},
			{Name: "health_check_timeout", Type: "string", Description: "Health check timeout", DefaultValue: "5s"},
			{Name: "failure_threshold", Type: "int", Description: "Health check failure threshold", DefaultValue: 3, MinValue: &[]float64{1}[0]},
			{Name: "success_threshold", Type: "int", Description: "Health check success threshold", DefaultValue: 1, MinValue: &[]float64{1}[0]},
			{Name: "health_check_initial_delay", Type: "string", Description: "Health check initial delay", DefaultValue: "10s"},
			{Name: "log_level", Type: "string", Description: "Log level", DefaultValue: "info", ValidValues: []interface{}{"debug", "info", "warn", "error"}},
			{Name: "log_format", Type: "string", Description: "Log format", DefaultValue: "json", ValidValues: []interface{}{"json", "text"}},
			{Name: "log_output", Type: "string", Description: "Log output destination", DefaultValue: "stdout", ValidValues: []interface{}{"stdout", "file", "syslog"}},
			{Name: "metrics_enabled", Type: "bool", Description: "Enable metrics collection", DefaultValue: true},
			{Name: "metrics_port", Type: "int", Description: "Metrics endpoint port", DefaultValue: 9090, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "metrics_path", Type: "string", Description: "Metrics endpoint path", DefaultValue: "/metrics"},
			{Name: "metrics_interval", Type: "string", Description: "Metrics collection interval", DefaultValue: "30s"},
			{Name: "deployment_strategy", Type: "string", Description: "Deployment strategy", DefaultValue: "RollingUpdate", ValidValues: []interface{}{"Recreate", "RollingUpdate", "BlueGreen", "Canary"}},
			{Name: "replicas", Type: "int", Description: "Number of replicas", DefaultValue: 1, MinValue: &[]float64{0}[0]},
			{Name: "cpu_request", Type: "string", Description: "CPU resource request", DefaultValue: "100m"},
			{Name: "memory_request", Type: "string", Description: "Memory resource request", DefaultValue: "128Mi"},
			{Name: "cpu_limit", Type: "string", Description: "CPU resource limit", DefaultValue: "200m"},
			{Name: "memory_limit", Type: "string", Description: "Memory resource limit", DefaultValue: "256Mi"},
			{Name: "environment", Type: "string", Description: "Deployment environment", DefaultValue: "development", ValidValues: []interface{}{"development", "staging", "production"}},
		},
		Labels: map[string]string{
			"category": "worker",
			"tier":     "basic",
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "system",
	}

	// High-Performance Worker Template
	highPerfTemplate := &Template{
		ID:             "template-high-perf-worker",
		Name:           "High-Performance Worker Agent",
		Description:    "A high-performance worker agent configuration template for demanding workloads",
		Version:        "1.0.0",
		AgentType:      "worker",
		BaseTemplateID: "template-basic-worker", // Inherits from basic worker
		Content: `{
  "id": "{{.agent_id}}",
  "name": "{{.agent_name}}",
  "agent_type": "worker",
  "base_config": {
    "max_concurrent_tasks": {{.max_concurrent_tasks}},
    "task_queue_size": {{.task_queue_size}},
    "heartbeat_interval": "{{.heartbeat_interval}}",
    "task_timeout": "{{.task_timeout}}",
    "resources": {
      "cpu": {{.cpu_millicores}},
      "memory": {{.memory_mb}},
      "max_tasks": {{.max_tasks}}
    }
  },
  "runtime_config": {
    "auto_restart": true,
    "restart_policy": {
      "policy": "Always",
      "max_retries": 5,
      "backoff_multiplier": 1.5,
      "initial_delay": "500ms",
      "max_delay": "30s"
    },
    "health_check": {
      "enabled": true,
      "path": "/health",
      "port": {{.health_check_port}},
      "interval": "10s",
      "timeout": "2s",
      "failure_threshold": 2,
      "success_threshold": 1,
      "initial_delay": "5s"
    },
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout"
    },
    "metrics": {
      "enabled": true,
      "port": {{.metrics_port}},
      "path": "/metrics",
      "interval": "10s"
    },
    "memory": {
      "max_memory_usage": {{.max_memory_usage}},
      "gc_interval": "{{.gc_interval}}",
      "persistence_enabled": {{.persistence_enabled}},
      "sync_interval": "{{.sync_interval}}"
    }
  },
  "deployment_config": {
    "strategy": "RollingUpdate",
    "replicas": {{.replicas}},
    "rolling_update": {
      "max_unavailable": "25%",
      "max_surge": "25%"
    },
    "resources": {
      "requests": {
        "cpu": "{{.cpu_request}}",
        "memory": "{{.memory_request}}"
      },
      "limits": {
        "cpu": "{{.cpu_limit}}",
        "memory": "{{.memory_limit}}"
      }
    },
    "affinity": {
      "node_affinity": {
        "preferred_during_scheduling_ignored_during_execution": [
          {
            "weight": 100,
            "preference": {
              "match_expressions": [
                {
                  "key": "node-type",
                  "operator": "In",
                  "values": ["high-performance"]
                }
              ]
            }
          }
        ]
      }
    }
  },
  "labels": {
    "type": "worker",
    "tier": "high-performance",
    "environment": "{{.environment}}"
  }
}`,
		Variables: []TemplateVariable{
			{Name: "agent_id", Type: "string", Description: "Unique agent identifier", Required: false},
			{Name: "agent_name", Type: "string", Description: "Human-readable agent name", Required: true},
			{Name: "max_concurrent_tasks", Type: "int", Description: "Maximum concurrent tasks", DefaultValue: 20, MinValue: &[]float64{10}[0], MaxValue: &[]float64{100}[0]},
			{Name: "task_queue_size", Type: "int", Description: "Task queue buffer size", DefaultValue: 1000, MinValue: &[]float64{100}[0]},
			{Name: "heartbeat_interval", Type: "string", Description: "Heartbeat interval duration", DefaultValue: "10s"},
			{Name: "task_timeout", Type: "string", Description: "Default task timeout", DefaultValue: "10m"},
			{Name: "cpu_millicores", Type: "int", Description: "CPU allocation in millicores", DefaultValue: 2000, MinValue: &[]float64{1000}[0]},
			{Name: "memory_mb", Type: "int", Description: "Memory allocation in MB", DefaultValue: 2048, MinValue: &[]float64{1024}[0]},
			{Name: "max_tasks", Type: "int", Description: "Maximum tasks in queue", DefaultValue: 10000, MinValue: &[]float64{1000}[0]},
			{Name: "health_check_port", Type: "int", Description: "Health check port", DefaultValue: 8080, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "metrics_port", Type: "int", Description: "Metrics endpoint port", DefaultValue: 9090, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "max_memory_usage", Type: "int", Description: "Maximum memory usage in MB", DefaultValue: 1536, MinValue: &[]float64{512}[0]},
			{Name: "gc_interval", Type: "string", Description: "Garbage collection interval", DefaultValue: "30s"},
			{Name: "persistence_enabled", Type: "bool", Description: "Enable memory persistence", DefaultValue: true},
			{Name: "sync_interval", Type: "string", Description: "Memory sync interval", DefaultValue: "60s"},
			{Name: "replicas", Type: "int", Description: "Number of replicas", DefaultValue: 3, MinValue: &[]float64{1}[0]},
			{Name: "cpu_request", Type: "string", Description: "CPU resource request", DefaultValue: "1000m"},
			{Name: "memory_request", Type: "string", Description: "Memory resource request", DefaultValue: "1Gi"},
			{Name: "cpu_limit", Type: "string", Description: "CPU resource limit", DefaultValue: "2000m"},
			{Name: "memory_limit", Type: "string", Description: "Memory resource limit", DefaultValue: "2Gi"},
			{Name: "environment", Type: "string", Description: "Deployment environment", DefaultValue: "production", ValidValues: []interface{}{"staging", "production"}},
		},
		Labels: map[string]string{
			"category":    "worker",
			"tier":        "high-performance",
			"performance": "high",
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "system",
	}

	// Coordinator Agent Template
	coordinatorTemplate := &Template{
		ID:          "template-coordinator",
		Name:        "Coordinator Agent",
		Description: "A coordinator agent configuration template for orchestrating other agents",
		Version:     "1.0.0",
		AgentType:   "coordinator",
		Content: `{
  "id": "{{.agent_id}}",
  "name": "{{.agent_name}}",
  "agent_type": "coordinator",
  "base_config": {
    "max_concurrent_tasks": {{.max_concurrent_tasks}},
    "task_queue_size": {{.task_queue_size}},
    "heartbeat_interval": "{{.heartbeat_interval}}",
    "task_timeout": "{{.task_timeout}}",
    "resources": {
      "cpu": {{.cpu_millicores}},
      "memory": {{.memory_mb}},
      "max_tasks": {{.max_tasks}}
    }
  },
  "runtime_config": {
    "auto_restart": true,
    "restart_policy": {
      "policy": "Always",
      "max_retries": 10,
      "backoff_multiplier": 2.0,
      "initial_delay": "2s",
      "max_delay": "5m"
    },
    "health_check": {
      "enabled": true,
      "path": "/health",
      "port": {{.health_check_port}},
      "interval": "15s",
      "timeout": "3s",
      "failure_threshold": 3,
      "success_threshold": 1,
      "initial_delay": "30s"
    },
    "logging": {
      "level": "{{.log_level}}",
      "format": "json",
      "output": "stdout"
    },
    "metrics": {
      "enabled": true,
      "port": {{.metrics_port}},
      "path": "/metrics",
      "interval": "15s"
    },
    "communication": {
      "protocols": ["http", "grpc"],
      "message_queue_size": {{.message_queue_size}},
      "connection_timeout": "{{.connection_timeout}}",
      "read_timeout": "{{.read_timeout}}",
      "write_timeout": "{{.write_timeout}}",
      "max_retries": {{.max_retries}}
    }
  },
  "deployment_config": {
    "strategy": "RollingUpdate",
    "replicas": {{.replicas}},
    "rolling_update": {
      "max_unavailable": "1",
      "max_surge": "1"
    },
    "resources": {
      "requests": {
        "cpu": "{{.cpu_request}}",
        "memory": "{{.memory_request}}"
      },
      "limits": {
        "cpu": "{{.cpu_limit}}",
        "memory": "{{.memory_limit}}"
      }
    },
    "affinity": {
      "pod_anti_affinity": {
        "preferred_during_scheduling_ignored_during_execution": [
          {
            "weight": 100,
            "pod_affinity_term": {
              "label_selector": {
                "match_expressions": [
                  {
                    "key": "type",
                    "operator": "In",
                    "values": ["coordinator"]
                  }
                ]
              },
              "topology_key": "kubernetes.io/hostname"
            }
          }
        ]
      }
    }
  },
  "labels": {
    "type": "coordinator",
    "role": "orchestrator",
    "environment": "{{.environment}}"
  }
}`,
		Variables: []TemplateVariable{
			{Name: "agent_id", Type: "string", Description: "Unique agent identifier", Required: false},
			{Name: "agent_name", Type: "string", Description: "Human-readable agent name", Required: true},
			{Name: "max_concurrent_tasks", Type: "int", Description: "Maximum concurrent orchestration tasks", DefaultValue: 10, MinValue: &[]float64{5}[0], MaxValue: &[]float64{50}[0]},
			{Name: "task_queue_size", Type: "int", Description: "Task queue buffer size", DefaultValue: 500, MinValue: &[]float64{100}[0]},
			{Name: "heartbeat_interval", Type: "string", Description: "Heartbeat interval duration", DefaultValue: "15s"},
			{Name: "task_timeout", Type: "string", Description: "Default task timeout", DefaultValue: "30m"},
			{Name: "cpu_millicores", Type: "int", Description: "CPU allocation in millicores", DefaultValue: 500, MinValue: &[]float64{200}[0]},
			{Name: "memory_mb", Type: "int", Description: "Memory allocation in MB", DefaultValue: 512, MinValue: &[]float64{256}[0]},
			{Name: "max_tasks", Type: "int", Description: "Maximum tasks in queue", DefaultValue: 5000, MinValue: &[]float64{500}[0]},
			{Name: "health_check_port", Type: "int", Description: "Health check port", DefaultValue: 8080, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "metrics_port", Type: "int", Description: "Metrics endpoint port", DefaultValue: 9090, MinValue: &[]float64{1}[0], MaxValue: &[]float64{65535}[0]},
			{Name: "log_level", Type: "string", Description: "Log level", DefaultValue: "info", ValidValues: []interface{}{"debug", "info", "warn", "error"}},
			{Name: "message_queue_size", Type: "int", Description: "Communication message queue size", DefaultValue: 1000, MinValue: &[]float64{100}[0]},
			{Name: "connection_timeout", Type: "string", Description: "Connection timeout", DefaultValue: "30s"},
			{Name: "read_timeout", Type: "string", Description: "Read timeout", DefaultValue: "30s"},
			{Name: "write_timeout", Type: "string", Description: "Write timeout", DefaultValue: "30s"},
			{Name: "max_retries", Type: "int", Description: "Maximum communication retries", DefaultValue: 3, MinValue: &[]float64{1}[0]},
			{Name: "replicas", Type: "int", Description: "Number of replicas", DefaultValue: 2, MinValue: &[]float64{1}[0], MaxValue: &[]float64{5}[0]},
			{Name: "cpu_request", Type: "string", Description: "CPU resource request", DefaultValue: "200m"},
			{Name: "memory_request", Type: "string", Description: "Memory resource request", DefaultValue: "256Mi"},
			{Name: "cpu_limit", Type: "string", Description: "CPU resource limit", DefaultValue: "500m"},
			{Name: "memory_limit", Type: "string", Description: "Memory resource limit", DefaultValue: "512Mi"},
			{Name: "environment", Type: "string", Description: "Deployment environment", DefaultValue: "production", ValidValues: []interface{}{"development", "staging", "production"}},
		},
		Labels: map[string]string{
			"category": "coordinator",
			"tier":     "control-plane",
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "system",
	}

	// Store the templates
	r.templates[workerTemplate.ID] = workerTemplate
	r.templates[highPerfTemplate.ID] = highPerfTemplate
	r.templates[coordinatorTemplate.ID] = coordinatorTemplate
}

// TemplateNotFoundError represents a template not found error
type TemplateNotFoundError struct {
	ID string
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s", e.ID)
}

func (e *TemplateNotFoundError) Is(target error) bool {
	_, ok := target.(*TemplateNotFoundError)
	return ok
}
