package models

import "time"

// WorkIssue represents a runtime instance of work (issue/ticket) that progresses through workflow stages
type WorkIssue struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty" db:"_key"`
	ID  string `json:"_id,omitempty" db:"_id"`
	Rev string `json:"_rev,omitempty" db:"_rev"`

	// Core metadata
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Number      int    `json:"number"` // Auto-incrementing issue number (e.g., #123)

	// Workflow context
	AgencyID   string `json:"agency_id" binding:"required"`
	InstanceID string `json:"instance_id" binding:"required"` // Agency instance ID
	WorkflowID string `json:"workflow_id" binding:"required"` // Links to workflow definition

	// Current state
	CurrentStep string `json:"current_step" binding:"required"` // Work item code (REQ1, IMPL1, etc.)
	Status      string `json:"status" binding:"required"`       // Issue status (see constants below)

	// Assignment
	AssignedTo string `json:"assigned_to,omitempty"` // Agent ID or user ID

	// Git integration
	BranchName    string `json:"branch_name,omitempty"`     // Git branch for this issue
	PullRequestID string `json:"pull_request_id,omitempty"` // Associated PR

	// History tracking
	CompletedSteps []string `json:"completed_steps"` // Work item codes completed (REQ1, REV1, ...)

	// Audit fields
	CreatedBy string    `json:"created_by" binding:"required"` // User/agent who created issue
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

// Issue status constants
const (
	IssueStatusOpen           = "open"             // Created, waiting for assignment
	IssueStatusAssigned       = "assigned"         // Worker assigned but work not started
	IssueStatusInProgress     = "in_progress"      // Active work in Git branch
	IssueStatusReadyForReview = "ready_for_review" // PR created, awaiting review
	IssueStatusCompleted      = "completed"        // PR merged, all workflow steps done
	IssueStatusBlocked        = "blocked"          // Cannot proceed (dependency or issue)
)

// WorkbenchBoard represents a generated Kanban board from a workflow
type WorkbenchBoard struct {
	AgencyID    string        `json:"agency_id"`
	InstanceID  string        `json:"instance_id"`
	WorkflowID  string        `json:"workflow_id"`
	TagKey      string        `json:"tag_key"` // Agency tag (specification snapshot)
	Columns     []BoardColumn `json:"columns"` // Kanban columns
	GeneratedAt time.Time     `json:"generated_at"`
}

// BoardColumn represents a single column in the Workbench board
type BoardColumn struct {
	ID           string      `json:"id"`             // Column identifier
	Name         string      `json:"name"`           // Display name (Requirements Gathering)
	WorkItemCode string      `json:"work_item_code"` // Work item code (REQ1)
	WorkItemKey  string      `json:"work_item_key"`  // Work item ArangoDB _key
	Order        int         `json:"order"`          // Display order (0, 1, 2, ...)
	Issues       []WorkIssue `json:"issues"`         // Issues in this column
}

// CreateIssueRequest represents the request to create a new issue
type CreateIssueRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	WorkflowID  string `json:"workflow_id" binding:"required"`
}

// UpdateIssueRequest represents the request to update an issue
type UpdateIssueRequest struct {
	Status        *string `json:"status,omitempty"`
	AssignedTo    *string `json:"assigned_to,omitempty"`
	BranchName    *string `json:"branch_name,omitempty"`
	PullRequestID *string `json:"pull_request_id,omitempty"`
}

// AssignIssueRequest represents the request to assign an issue to a worker
type AssignIssueRequest struct {
	WorkerID string `json:"worker_id" binding:"required"` // Agent ID or user ID
}

// IssueListResponse represents a list of issues with pagination
type IssueListResponse struct {
	Issues []WorkIssue `json:"issues"`
	Total  int         `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}
