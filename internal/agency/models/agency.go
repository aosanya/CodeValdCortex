package models

import "time"

// Agency represents a use case operating as an independent entity with its own configuration
type Agency struct {
	Key         string `json:"_key,omitempty"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`

	// Lifecycle state
	State AgencyState `json:"state"`

	// Publishing metadata
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	PublishedBy   string     `json:"published_by,omitempty"`
	PublicationID string     `json:"publication_id,omitempty"`
	CurrentTagID  string     `json:"current_tag_id,omitempty"`

	// Activation metadata
	ActivatedAt      *time.Time `json:"activated_at,omitempty"`
	ActivatedBy      string     `json:"activated_by,omitempty"`
	ActiveAgentCount int        `json:"active_agent_count"`

	// Infrastructure
	Database string         `json:"database"` // Database name for this agency
	Metadata AgencyMetadata `json:"metadata"`
	Settings AgencySettings `json:"settings"`

	// Audit trail
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

// AgencyState represents the lifecycle state of an agency
type AgencyState string

const (
	AgencyStateDraft     AgencyState = "draft"     // Design in progress
	AgencyStateValidated AgencyState = "validated" // Ready for publishing
	AgencyStatePublished AgencyState = "published" // Published but not active
	AgencyStateActive    AgencyState = "active"    // Agents running
	AgencyStatePaused    AgencyState = "paused"    // Temporarily suspended
	AgencyStateDraining  AgencyState = "draining"  // Completing existing work
	AgencyStateStopped   AgencyState = "stopped"   // Shut down gracefully
	AgencyStateArchived  AgencyState = "archived"  // Historical record
)

// ValidAgencyStates returns all valid agency states
func ValidAgencyStates() []AgencyState {
	return []AgencyState{
		AgencyStateDraft,
		AgencyStateValidated,
		AgencyStatePublished,
		AgencyStateActive,
		AgencyStatePaused,
		AgencyStateDraining,
		AgencyStateStopped,
		AgencyStateArchived,
	}
}

// IsValidAgencyState checks if the given state is valid
func IsValidAgencyState(state AgencyState) bool {
	for _, validState := range ValidAgencyStates() {
		if state == validState {
			return true
		}
	}
	return false
}

// AgencyMetadata contains additional information about the agency
type AgencyMetadata struct {
	Location    string   `json:"location,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	TotalAgents int      `json:"total_agents"`
	Zones       int      `json:"zones,omitempty"`
	APIEndpoint string   `json:"api_endpoint,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AgencySettings contains configuration options for the agency
type AgencySettings struct {
	AutoStart         bool `json:"auto_start"`
	MonitoringEnabled bool `json:"monitoring_enabled"`
	DashboardEnabled  bool `json:"dashboard_enabled"`
	VisualizerEnabled bool `json:"visualizer_enabled"`
}

// AgencyFilters defines criteria for filtering agencies in queries
type AgencyFilters struct {
	Category string
	State    AgencyState
	Search   string // Search in name/description
	Tags     []string
	Limit    int
	Offset   int
}

// AgencyUpdates defines fields that can be updated
type AgencyUpdates struct {
	Name        *string         `json:"name,omitempty"`
	DisplayName *string         `json:"display_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	State       *AgencyState    `json:"state,omitempty"`
	Metadata    *AgencyMetadata `json:"metadata,omitempty"`
	Settings    *AgencySettings `json:"settings,omitempty"`
}

// AgencyStatistics contains operational statistics for an agency
type AgencyStatistics struct {
	AgencyID       string    `json:"agency_id"`
	ActiveAgents   int       `json:"active_agents"`
	InactiveAgents int       `json:"inactive_agents"`
	TotalTasks     int       `json:"total_tasks"`
	CompletedTasks int       `json:"completed_tasks"`
	FailedTasks    int       `json:"failed_tasks"`
	LastActivity   time.Time `json:"last_activity"`
	Uptime         float64   `json:"uptime"` // Percentage
}

// CreateAgencyRequest is the request body for creating a new agency
type CreateAgencyRequest struct {
	ID          string         `json:"id" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	DisplayName string         `json:"display_name" binding:"required"`
	Description string         `json:"description"`
	Category    string         `json:"category" binding:"required"`
	Icon        string         `json:"icon"`
	Metadata    AgencyMetadata `json:"metadata"`
	Settings    AgencySettings `json:"settings"`
}

// UpdateAgencyRequest is the request body for updating an agency
type UpdateAgencyRequest struct {
	Name        *string         `json:"name,omitempty"`
	DisplayName *string         `json:"display_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	State       *AgencyState    `json:"state,omitempty"`
	Metadata    *AgencyMetadata `json:"metadata,omitempty"`
	Settings    *AgencySettings `json:"settings,omitempty"`
}
