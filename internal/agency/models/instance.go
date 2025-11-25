package models

import "time"

// InstanceState represents the runtime state of an agency instance
type InstanceState string

const (
	// Optimistic start: instances immediately enter "running" state (no "starting" state)
	// See: instance-research-session.md Q1-Q2 for architecture rationale
	InstanceStateRunning  InstanceState = "running"  // Instance active (accepts jobs, agents spawn on-demand)
	InstanceStateStopping InstanceState = "stopping" // Graceful shutdown (rejects new jobs, 30s timeout)
	InstanceStateStopped  InstanceState = "stopped"  // All agents stopped
	InstanceStateFailed   InstanceState = "failed"   // Failed to start/crashed
)

// AgencyInstance represents a running instance of an agency from a tag
type AgencyInstance struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty"`
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`

	// Instance metadata
	AgencyID    string `json:"agency_id"`
	TagID       string `json:"tag_id"`      // Immutable reference to source tag
	TagName     string `json:"tag_name"`    // Cached for display
	InstanceID  string `json:"instance_id"` // Unique identifier (UUID)
	Name        string `json:"name"`        // User-friendly name (MUST be unique per agency)
	Description string `json:"description"`

	// Runtime state
	State        InstanceState `json:"state"`         // running, stopping, stopped, failed
	HealthStatus string        `json:"health_status"` // healthy, degraded, unhealthy (calculated on-demand)

	// Deployment info
	DeployedAt time.Time  `json:"deployed_at"`
	DeployedBy string     `json:"deployed_by"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	StoppedAt  *time.Time `json:"stopped_at,omitempty"`

	// Runtime tracking
	AgentCount     int       `json:"agent_count"`    // Active agent references
	WorkflowCount  int       `json:"workflow_count"` // Active workflows
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	UptimeSeconds  int64     `json:"uptime_seconds"`   // Computed: current_time - started_at
	AcceptsNewJobs bool      `json:"accepts_new_jobs"` // False when stopping

	// Resource allocation (from tag snapshot)
	ResourceLimits ResourceAllocation `json:"resource_limits"`

	// Metadata and tagging
	Tags     []string               `json:"tags,omitempty"`     // Uses existing tag system
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional tracking data

	// Soft delete
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	DeletedBy string     `json:"deleted_by,omitempty"`
}

// InstanceResourceStatus tracks current resource usage (kept for backward compatibility)
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
	TagName     string                 `json:"tag_name"`       // Source tag to instantiate
	Name        string                 `json:"name"`           // User-friendly name (must be unique per agency)
	Description string                 `json:"description"`    // Optional purpose/notes
	Tags        []string               `json:"tags,omitempty"` // Uses existing tag system
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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
