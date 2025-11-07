package models

import "time"

// RACIRole represents a RACI responsibility assignment
type RACIRole string

const (
	RACIResponsible RACIRole = "R" // Does the work
	RACIAccountable RACIRole = "A" // Ultimately answerable (one per activity)
	RACIConsulted   RACIRole = "C" // Provides input
	RACIInformed    RACIRole = "I" // Kept in the loop
)

// RACIActivity represents a single activity/task in the RACI matrix
type RACIActivity struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Assignments map[string]RACIRole `json:"assignments"` // role_name -> RACI role
	Order       int                 `json:"order"`       // For display ordering
}

// RACIMatrix represents a complete RACI matrix for an agency
type RACIMatrix struct {
	Key         string         `json:"_key,omitempty"`
	ID          string         `json:"_id,omitempty"`
	AgencyID    string         `json:"agency_id"`
	WorkItemKey string         `json:"work_item_key,omitempty"` // Optional: link to specific work item
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Activities  []RACIActivity `json:"activities"`
	Roles       []string       `json:"roles"` // List of role names
	TemplateID  string         `json:"template_id,omitempty"`
	IsValid     bool           `json:"is_valid"`
	Errors      []string       `json:"errors,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// RACITemplate represents a reusable RACI matrix template
type RACITemplate struct {
	Key         string         `json:"_key,omitempty"`
	ID          string         `json:"_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"` // e.g., "Software Development", "Research", "Infrastructure"
	Activities  []RACIActivity `json:"activities"`
	Roles       []string       `json:"roles"`
	IsPublic    bool           `json:"is_public"` // Available to all agencies
	AgencyID    string         `json:"agency_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CreateRACIMatrixRequest is the request body for creating a RACI matrix
type CreateRACIMatrixRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	WorkItemKey string         `json:"work_item_key,omitempty"`
	Activities  []RACIActivity `json:"activities"`
	Roles       []string       `json:"roles"`
	TemplateID  string         `json:"template_id,omitempty"`
}

// UpdateRACIMatrixRequest is the request body for updating a RACI matrix
type UpdateRACIMatrixRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	WorkItemKey string         `json:"work_item_key,omitempty"`
	Activities  []RACIActivity `json:"activities"`
	Roles       []string       `json:"roles"`
}

// RACIValidationResult contains validation results for a RACI matrix
type RACIValidationResult struct {
	IsValid  bool                    `json:"is_valid"`
	Errors   []RACIValidationError   `json:"errors,omitempty"`
	Warnings []RACIValidationWarning `json:"warnings,omitempty"`
	Summary  RACIValidationSummary   `json:"summary"`
}

// RACIValidationError represents a validation error
type RACIValidationError struct {
	ActivityID string `json:"activity_id"`
	Activity   string `json:"activity"`
	ErrorType  string `json:"error_type"` // "missing_accountable", "multiple_accountable", "missing_responsible"
	Message    string `json:"message"`
}

// RACIValidationWarning represents a validation warning
type RACIValidationWarning struct {
	ActivityID  string `json:"activity_id"`
	Activity    string `json:"activity"`
	WarningType string `json:"warning_type"` // "no_consulted", "no_informed"
	Message     string `json:"message"`
}

// RACIValidationSummary provides an overview of validation results
type RACIValidationSummary struct {
	TotalActivities        int `json:"total_activities"`
	ValidActivities        int `json:"valid_activities"`
	ActivitiesWithErrors   int `json:"activities_with_errors"`
	ActivitiesWithWarnings int `json:"activities_with_warnings"`
}

// RACIExportFormat represents the export format for RACI matrices
type RACIExportFormat string

const (
	RACIExportPDF      RACIExportFormat = "pdf"
	RACIExportMarkdown RACIExportFormat = "markdown"
	RACIExportJSON     RACIExportFormat = "json"
)

// RACIAssignment represents an edge between a work item and a role
// Stored in ArangoDB as an edge in the raci_assignments collection within the agency's database
type RACIAssignment struct {
	Key         string    `json:"_key,omitempty"`
	ID          string    `json:"_id,omitempty"`
	From        string    `json:"_from"`         // Full ID: work_items/{work_item_key}
	To          string    `json:"_to"`           // Full ID: roles/{role_key}
	WorkItemKey string    `json:"work_item_key"` // Work item _key (denormalized for queries)
	RoleKey     string    `json:"role_key"`      // Role _key (denormalized for queries)
	RACI        string    `json:"raci"`          // "R", "A", "C", or "I"
	Objective   string    `json:"objective"`     // Description of what the role does
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRACIAssignmentRequest is the request to create a RACI assignment edge
type CreateRACIAssignmentRequest struct {
	WorkItemKey string `json:"work_item_key" binding:"required"`
	RoleKey     string `json:"role_key" binding:"required"`
	RACI        string `json:"raci" binding:"required"`
	Objective   string `json:"objective"`
}

// AgencyRACIAssignments represents all RACI assignments for an agency
// This is a simpler model for the MVP that maps work items to role assignments
// DEPRECATED: Use RACIAssignment edges instead
type AgencyRACIAssignments struct {
	Key         string                               `json:"_key,omitempty"`
	ID          string                               `json:"_id,omitempty"`
	AgencyID    string                               `json:"agency_id"`
	Assignments map[string]map[string]RoleAssignment `json:"assignments"` // workItemKey -> roleKey -> assignment
	CreatedAt   time.Time                            `json:"created_at"`
	UpdatedAt   time.Time                            `json:"updated_at"`
}

// RoleAssignment represents a single RACI assignment for a role on a work item
// DEPRECATED: Use RACIAssignment edge instead
type RoleAssignment struct {
	RACI      string `json:"raci"`      // "R", "A", "C", or "I"
	Objective string `json:"objective"` // Description of what the role does for this work item
}
