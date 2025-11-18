package models

import "time"

// ===== Core Workflow Definitions =====

// Workflow represents a simplified step-based workflow definition stored in ArangoDB
type Workflow struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty" db:"_key"`
	ID  string `json:"_id,omitempty" db:"_id"`
	Rev string `json:"_rev,omitempty" db:"_rev"`

	// Core fields
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	AgencyID    string `json:"agency_id" binding:"required"`
	Version     string `json:"version"`
	Steps       Steps  `json:"steps"` // Simplified step-based model
	// Status intentionally omitted: workflows currently do not have a runtime status.
	// Future: add published/draft states when workflow publishing is implemented.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
}

// Steps is an array of workflow steps
type Steps []Step

// Step represents a single step in the workflow
// Execution type is implicit: 1 item = sequential, 2+ items = parallel
type Step struct {
	ID    string     `json:"id" binding:"required"`
	Order int        `json:"order" binding:"required"`
	Items []StepItem `json:"items" binding:"required,min=1"` // 1 item = sequential, 2+ = parallel
}

// StepItem represents a work item within a step
type StepItem struct {
	ID           string `json:"id" binding:"required"`
	WorkItemID   string `json:"work_item_id" binding:"required"`   // Work item document ID
	WorkItemKey  string `json:"work_item_key" binding:"required"`  // ArangoDB _key reference
	WorkItemName string `json:"work_item_name" binding:"required"` // Denormalized for display
}

// ===== Validation =====

// WorkflowValidationResult contains validation errors and warnings
type WorkflowValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ValidationError represents a workflow validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	StepID  string `json:"step_id,omitempty"` // Step ID if error relates to a specific step
	ItemID  string `json:"item_id,omitempty"` // Item ID if error relates to a specific step item
}
