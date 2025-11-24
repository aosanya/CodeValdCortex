package models

import "time"

// InstanceState represents the runtime state of an agency instance
type InstanceState string

const (
	InstanceStateStarting InstanceState = "starting" // Instance is being initialized
	InstanceStateRunning  InstanceState = "running"  // Instance is active and operational
	InstanceStateStopping InstanceState = "stopping" // Instance is shutting down
	InstanceStateStopped  InstanceState = "stopped"  // Instance is stopped
	InstanceStateFailed   InstanceState = "failed"   // Instance encountered an error
)

// AgencyInstance represents a running instance of an agency from a tag
type AgencyInstance struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty"`
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`

	// Instance identification
	InstanceID   string `json:"instance_id"`   // Unique instance identifier
	AgencyID     string `json:"agency_id"`     // Source agency
	TagName      string `json:"tag_name"`      // Source tag name
	InstanceName string `json:"instance_name"` // User-friendly instance name

	// State management
	State        InstanceState `json:"state"`
	StateMessage string        `json:"state_message,omitempty"` // Human-readable state info

	// Runtime configuration
	Snapshot  AgencySnapshot         `json:"snapshot"`  // Config snapshot from tag
	Config    InstanceConfiguration  `json:"config"`    // Runtime-specific config
	Resources InstanceResourceStatus `json:"resources"` // Resource allocation and usage

	// Lifecycle tracking
	StartedAt  time.Time  `json:"started_at"`
	StoppedAt  *time.Time `json:"stopped_at,omitempty"`
	LastSeenAt time.Time  `json:"last_seen_at"` // Last health check

	// Metadata
	CreatedBy   string                 `json:"created_by"`
	Environment string                 `json:"environment"` // e.g., "production", "staging", "test"
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// InstanceConfiguration holds runtime-specific settings
type InstanceConfiguration struct {
	// Resource allocation
	CPULimit    string `json:"cpu_limit"`    // e.g., "2000m"
	MemoryLimit string `json:"memory_limit"` // e.g., "1Gi"
	MaxAgents   int    `json:"max_agents"`   // Maximum concurrent agents

	// Auto-scaling settings
	AutoScaleEnabled bool `json:"auto_scale_enabled"`
	MinAgents        int  `json:"min_agents"`
	MaxScaleAgents   int  `json:"max_scale_agents"`

	// Networking
	ExposedPorts []int  `json:"exposed_ports,omitempty"`
	HostBinding  string `json:"host_binding,omitempty"`

	// Monitoring
	MetricsEnabled bool   `json:"metrics_enabled"`
	LogLevel       string `json:"log_level"` // debug, info, warn, error
}

// InstanceResourceStatus tracks current resource usage
type InstanceResourceStatus struct {
	// Current usage
	CPUUsage    string `json:"cpu_usage"`    // e.g., "850m"
	MemoryUsage string `json:"memory_usage"` // e.g., "512Mi"

	// Agent counts
	ActiveAgents     int `json:"active_agents"`
	SpawnedAgents    int `json:"spawned_agents"`
	TerminatedAgents int `json:"terminated_agents"`

	// Health indicators
	HealthScore         float64   `json:"health_score"` // 0.0 to 1.0
	LastHealthCheck     time.Time `json:"last_health_check"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

// InstanceHealth represents the health status of an instance
type InstanceHealth struct {
	InstanceID string        `json:"instance_id"`
	State      InstanceState `json:"state"`
	Healthy    bool          `json:"healthy"`
	Message    string        `json:"message,omitempty"`

	// Component health
	AgentsHealthy    bool `json:"agents_healthy"`
	WorkflowsHealthy bool `json:"workflows_healthy"`
	ResourcesHealthy bool `json:"resources_healthy"`

	// Metrics
	Uptime        time.Duration `json:"uptime"`
	RequestCount  int64         `json:"request_count"`
	ErrorCount    int64         `json:"error_count"`
	LastErrorTime *time.Time    `json:"last_error_time,omitempty"`

	CheckedAt time.Time `json:"checked_at"`
}

// InstanceListItem is a lightweight view for instance lists
type InstanceListItem struct {
	InstanceID   string        `json:"instance_id"`
	InstanceName string        `json:"instance_name"`
	AgencyID     string        `json:"agency_id"`
	TagName      string        `json:"tag_name"`
	State        InstanceState `json:"state"`
	Environment  string        `json:"environment"`
	ActiveAgents int           `json:"active_agents"`
	HealthScore  float64       `json:"health_score"`
	StartedAt    time.Time     `json:"started_at"`
	Uptime       time.Duration `json:"uptime"`
}

// StartInstanceRequest represents the request to start a new instance
type StartInstanceRequest struct {
	InstanceName string                 `json:"instance_name"` // User-friendly name
	Environment  string                 `json:"environment"`   // e.g., "production", "staging"
	Config       InstanceConfiguration  `json:"config"`        // Runtime configuration
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// InstanceAgent represents an agent belonging to an instance
type InstanceAgent struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty"`
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`

	// Agent identification
	AgentID    string `json:"agent_id"`
	InstanceID string `json:"instance_id"`
	AgencyID   string `json:"agency_id"`

	// Agent details
	RoleCode      string `json:"role_code"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	AutonomyLevel string `json:"autonomy_level"`

	// State
	State     string     `json:"state"`
	SpawnedAt time.Time  `json:"spawned_at"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`

	// Resource tracking
	CPUUsage    string `json:"cpu_usage,omitempty"`
	MemoryUsage string `json:"memory_usage,omitempty"`
	TaskCount   int    `json:"task_count"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
