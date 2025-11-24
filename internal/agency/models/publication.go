package models

import "time"

// AgencyPublication represents a published version of an agency
type AgencyPublication struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty"`
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`

	// Publication metadata
	AgencyID    string `json:"agency_id"`
	Version     string `json:"version"`          // Semantic version (v1.0.0)
	TagID       string `json:"tag_id,omitempty"` // Optional tag reference
	Description string `json:"description"`

	// Snapshot of agency configuration at publication time
	Snapshot AgencySnapshot `json:"snapshot"`

	// Deployment manifest (computed at publish time)
	Manifest DeploymentManifest `json:"manifest"`

	// Lifecycle tracking
	PublishedAt   time.Time  `json:"published_at"`
	PublishedBy   string     `json:"published_by"`
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgencySnapshot captures complete agency state at publication
type AgencySnapshot struct {
	Specification AgencySpecification `json:"specification"`
	AIPolicy      AIPolicy            `json:"ai_policy,omitempty"`
	Settings      AgencySettings      `json:"settings"`
	Metadata      AgencyMetadata      `json:"metadata"`
}

// AIPolicy represents AI governance policy
type AIPolicy struct {
	Enabled          bool                   `json:"enabled"`
	RiskLevel        string                 `json:"risk_level"`
	ApprovalRequired bool                   `json:"approval_required"`
	Rules            []PolicyRule           `json:"rules"`
	Config           map[string]interface{} `json:"config,omitempty"`
}

// PolicyRule represents a single policy rule
type PolicyRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// DeploymentManifest contains computed deployment plan
type DeploymentManifest struct {
	AgentSpawnPlan     AgentSpawnPlan     `json:"agent_spawn_plan"`
	WorkflowExecution  WorkflowExecution  `json:"workflow_execution"`
	ResourceAllocation ResourceAllocation `json:"resource_allocation"`
	MonitoringConfig   MonitoringConfig   `json:"monitoring_config"`
}

// AgentSpawnPlan defines which agents to spawn
type AgentSpawnPlan struct {
	TotalAgents int               `json:"total_agents"`
	Agents      []AgentDefinition `json:"agents"`
}

// AgentDefinition defines an agent to be spawned
type AgentDefinition struct {
	RoleCode       string                 `json:"role_code"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	AutonomyLevel  string                 `json:"autonomy_level"`
	ResourceLimits ResourceLimits         `json:"resource_limits"`
	Configuration  map[string]interface{} `json:"configuration"`
}

// ResourceLimits defines resource constraints for an agent
type ResourceLimits struct {
	CPULimit    string `json:"cpu_limit"`    // e.g., "1000m"
	MemoryLimit string `json:"memory_limit"` // e.g., "512Mi"
	TokenBudget int    `json:"token_budget"`
}

// WorkflowExecution defines workflow initialization
type WorkflowExecution struct {
	Workflows []WorkflowConfig `json:"workflows"`
}

// WorkflowConfig defines workflow runtime configuration
type WorkflowConfig struct {
	WorkflowID string `json:"workflow_id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	AutoStart  bool   `json:"auto_start"`
}

// ResourceAllocation defines quotas and limits
type ResourceAllocation struct {
	TotalCPU    string `json:"total_cpu"`
	TotalMemory string `json:"total_memory"`
	MaxAgents   int    `json:"max_agents"`
}

// MonitoringConfig defines monitoring setup
type MonitoringConfig struct {
	Enabled         bool     `json:"enabled"`
	MetricsEndpoint string   `json:"metrics_endpoint"`
	Alerts          []string `json:"alerts"`
}
