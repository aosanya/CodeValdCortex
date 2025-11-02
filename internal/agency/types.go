package agency

import (
	"time"
)

// Agency represents a use case operating as an independent entity with its own configuration
type Agency struct {
	Key         string         `json:"_key,omitempty"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Icon        string         `json:"icon"`
	Status      AgencyStatus   `json:"status"`
	Database    string         `json:"database"` // Database name for this agency
	Metadata    AgencyMetadata `json:"metadata"`
	Settings    AgencySettings `json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   string         `json:"created_by"`
}

// AgencyStatus represents the current state of an agency
type AgencyStatus string

const (
	AgencyStatusActive   AgencyStatus = "active"
	AgencyStatusInactive AgencyStatus = "inactive"
	AgencyStatusPaused   AgencyStatus = "paused"
	AgencyStatusArchived AgencyStatus = "archived"
)

// AgencyMetadata contains additional information about the agency
type AgencyMetadata struct {
	Location    string   `json:"location,omitempty"`
	AgentTypes  []string `json:"agent_types,omitempty"`
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
	Status   AgencyStatus
	Search   string // Search in name/description
	Tags     []string
	Limit    int
	Offset   int
}

// AgencyUpdates defines fields that can be updated
type AgencyUpdates struct {
	DisplayName *string         `json:"display_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	Status      *AgencyStatus   `json:"status,omitempty"`
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
	DisplayName *string         `json:"display_name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	Status      *AgencyStatus   `json:"status,omitempty"`
	Metadata    *AgencyMetadata `json:"metadata,omitempty"`
	Settings    *AgencySettings `json:"settings,omitempty"`
}

// Overview represents the overview document in an agency's database
type Overview struct {
	Key          string    `json:"_key,omitempty"`
	AgencyID     string    `json:"agency_id"`
	Introduction string    `json:"introduction"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Goal represents a goal statement that the agency is solving
type Goal struct {
	Key            string    `json:"_key,omitempty"`
	ID             string    `json:"_id,omitempty"`
	AgencyID       string    `json:"agency_id"`
	Number         int       `json:"number"`
	Code           string    `json:"code"`
	Description    string    `json:"description"`
	Scope          string    `json:"scope"`
	SuccessMetrics []string  `json:"success_metrics"`
	Priority       string    `json:"priority"` // High, Medium, Low
	Status         string    `json:"status"`   // Draft, Active, Resolved, Archived
	Category       string    `json:"category"` // Operational, Strategic, Technical, etc.
	Tags           []string  `json:"tags"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateGoalRequest is the request body for creating a goal
type CreateGoalRequest struct {
	Code           string   `json:"code" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	Scope          string   `json:"scope"`
	SuccessMetrics []string `json:"success_metrics"`
	Priority       string   `json:"priority"` // High, Medium, Low
	Status         string   `json:"status"`   // Draft, Active, Resolved, Archived
	Category       string   `json:"category"` // Operational, Strategic, Technical, etc.
	Tags           []string `json:"tags"`
}

// UpdateGoalRequest is the request body for updating a goal
type UpdateGoalRequest struct {
	Code           string   `json:"code" binding:"required"`
	Description    string   `json:"description" binding:"required"`
	Scope          string   `json:"scope"`
	SuccessMetrics []string `json:"success_metrics"`
	Priority       string   `json:"priority"` // High, Medium, Low
	Status         string   `json:"status"`   // Draft, Active, Resolved, Archived
	Category       string   `json:"category"` // Operational, Strategic, Technical, etc.
	Tags           []string `json:"tags"`
}

// GoalRefineRequest is the request body for AI goal refinement
type GoalRefineRequest struct {
	Description    string   `json:"description" binding:"required"`
	Scope          string   `json:"scope"`
	SuccessMetrics []string `json:"success_metrics"`
}

// UpdateOverviewRequest is the request body for updating overview
type UpdateOverviewRequest struct {
	Introduction string `json:"introduction"`
}

// WorkItemType represents the type of work item
type WorkItemType string

const (
	WorkItemTypeTask     WorkItemType = "Task"
	WorkItemTypeFeature  WorkItemType = "Feature"
	WorkItemTypeEpic     WorkItemType = "Epic"
	WorkItemTypeBug      WorkItemType = "Bug"
	WorkItemTypeResearch WorkItemType = "Research"
)

// WorkItemPriority represents the priority level of a work item
type WorkItemPriority string

const (
	WorkItemPriorityP0 WorkItemPriority = "P0" // Critical/Blocking
	WorkItemPriorityP1 WorkItemPriority = "P1" // High
	WorkItemPriorityP2 WorkItemPriority = "P2" // Medium
	WorkItemPriorityP3 WorkItemPriority = "P3" // Low
)

// WorkItemStatus represents the current status of a work item
type WorkItemStatus string

const (
	WorkItemStatusNotStarted WorkItemStatus = "Not Started"
	WorkItemStatusInProgress WorkItemStatus = "In Progress"
	WorkItemStatusBlocked    WorkItemStatus = "Blocked"
	WorkItemStatusInReview   WorkItemStatus = "In Review"
	WorkItemStatusDone       WorkItemStatus = "Done"
	WorkItemStatusCancelled  WorkItemStatus = "Cancelled"
)

// WorkItem represents a work item in the agency
type WorkItem struct {
	Key             string           `json:"_key,omitempty"`
	ID              string           `json:"_id,omitempty"`
	AgencyID        string           `json:"agency_id"`
	Number          int              `json:"number"`
	Code            string           `json:"code"` // e.g., "WI-001"
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Type            WorkItemType     `json:"type"`
	Priority        WorkItemPriority `json:"priority"`
	Status          WorkItemStatus   `json:"status"`
	Deliverables    []string         `json:"deliverables"`
	Dependencies    []string         `json:"dependencies"`     // References to other work item codes
	EstimatedEffort string           `json:"estimated_effort"` // e.g., "2 weeks", "40 hours"
	AssignedTo      string           `json:"assigned_to,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// CreateWorkItemRequest is the request body for creating a work item
type CreateWorkItemRequest struct {
	Title           string           `json:"title" binding:"required"`
	Description     string           `json:"description" binding:"required"`
	Type            WorkItemType     `json:"type" binding:"required"`
	Priority        WorkItemPriority `json:"priority" binding:"required"`
	Status          WorkItemStatus   `json:"status"`
	Deliverables    []string         `json:"deliverables"`
	Dependencies    []string         `json:"dependencies"`
	EstimatedEffort string           `json:"estimated_effort"`
	AssignedTo      string           `json:"assigned_to,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
}

// UpdateWorkItemRequest is the request body for updating a work item
type UpdateWorkItemRequest struct {
	Title           string           `json:"title" binding:"required"`
	Description     string           `json:"description" binding:"required"`
	Type            WorkItemType     `json:"type" binding:"required"`
	Priority        WorkItemPriority `json:"priority" binding:"required"`
	Status          WorkItemStatus   `json:"status" binding:"required"`
	Deliverables    []string         `json:"deliverables"`
	Dependencies    []string         `json:"dependencies"`
	EstimatedEffort string           `json:"estimated_effort"`
	AssignedTo      string           `json:"assigned_to,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
}

// WorkItemRefineRequest is the request body for AI work item refinement
type WorkItemRefineRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Type         string   `json:"type" binding:"required"`
	Deliverables []string `json:"deliverables"`
}
