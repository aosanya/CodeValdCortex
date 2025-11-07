package models

import "time"

// Workflow represents a visual workflow definition
type Workflow struct {
	Key         string                 `json:"_key,omitempty" db:"_key"`
	ID          string                 `json:"_id,omitempty" db:"_id"`
	Rev         string                 `json:"_rev,omitempty" db:"_rev"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	AgencyID    string                 `json:"agency_id" binding:"required"`
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
}

// WorkflowNode represents a work item node in the workflow
type WorkflowNode struct {
	ID       string                 `json:"id" binding:"required"`
	Type     string                 `json:"type" binding:"required"` // document, software, proposal, analysis
	Position NodePosition           `json:"position" binding:"required"`
	Data     WorkflowNodeData       `json:"data" binding:"required"`
	Width    int                    `json:"width,omitempty"`
	Height   int                    `json:"height,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NodePosition represents the x,y coordinates of a node
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowNodeData contains the work item configuration
type WorkflowNodeData struct {
	Type        string      `json:"type" binding:"required"`
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description" binding:"required"`
	Labels      []string    `json:"labels,omitempty"`
	LabelsText  string      `json:"labels_text,omitempty"` // For UI binding
	GiteaConfig GiteaConfig `json:"gitea_config" binding:"required"`
	Status      string      `json:"status,omitempty"` // pending, executing, completed, failed
	IssueID     int64       `json:"issue_id,omitempty"`
	IssueURL    string      `json:"issue_url,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// GiteaConfig contains GitOps settings for the work item
type GiteaConfig struct {
	Repo           string `json:"repo" binding:"required"`
	BranchPattern  string `json:"branch_pattern"`
	MergeStrategy  string `json:"merge_strategy"` // squash, merge, rebase
	AutoMerge      bool   `json:"auto_merge"`
	RequireReviews int    `json:"require_reviews"`
}

// WorkflowEdge represents a dependency between work items
type WorkflowEdge struct {
	ID       string                 `json:"id" binding:"required"`
	Source   string                 `json:"source" binding:"required"`
	Target   string                 `json:"target" binding:"required"`
	Type     string                 `json:"type,omitempty"` // default, conditional, etc.
	Animated bool                   `json:"animated,omitempty"`
	Label    string                 `json:"label,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStatus represents the execution state of the workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft      WorkflowStatus = "draft"
	WorkflowStatusValidating WorkflowStatus = "validating"
	WorkflowStatusValid      WorkflowStatus = "valid"
	WorkflowStatusInvalid    WorkflowStatus = "invalid"
	WorkflowStatusExecuting  WorkflowStatus = "executing"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"
)

// WorkflowValidationResult contains validation errors and warnings
type WorkflowValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// WorkflowExecutionRequest contains parameters for workflow execution
type WorkflowExecutionRequest struct {
	WorkflowID string                 `json:"workflow_id" binding:"required"`
	DryRun     bool                   `json:"dry_run"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// WorkflowExecutionResult contains the result of workflow execution
type WorkflowExecutionResult struct {
	WorkflowID   string               `json:"workflow_id"`
	ExecutionID  string               `json:"execution_id"`
	Status       string               `json:"status"`
	CreatedNodes []WorkflowNodeResult `json:"created_nodes"`
	Errors       []string             `json:"errors,omitempty"`
	StartedAt    time.Time            `json:"started_at"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
}

// WorkflowNodeResult contains the execution result for a single node
type WorkflowNodeResult struct {
	NodeID   string `json:"node_id"`
	IssueID  int64  `json:"issue_id"`
	IssueURL string `json:"issue_url"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}
