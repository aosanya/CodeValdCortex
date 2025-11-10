package models

import "time"

// ===== Core Workflow Definitions =====

// Workflow represents a visual workflow definition stored in ArangoDB
type Workflow struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty" db:"_key"`
	ID  string `json:"_id,omitempty" db:"_id"`
	Rev string `json:"_rev,omitempty" db:"_rev"`

	// Core fields
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	AgencyID    string                 `json:"agency_id" binding:"required"`
	Version     string                 `json:"version"`
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	Metadata    map[string]interface{} `json:"metadata"`
	Variables   map[string]interface{} `json:"variables"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
}

// WorkflowNode represents a node in the workflow (combining both models)
type WorkflowNode struct {
	ID       string                 `json:"id" binding:"required"`
	Type     NodeType               `json:"type" binding:"required"`
	Position NodePosition           `json:"position" binding:"required"`
	Data     WorkflowNodeData       `json:"data" binding:"required"`
	Width    int                    `json:"width,omitempty"`
	Height   int                    `json:"height,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowEdge represents a connection between nodes
type WorkflowEdge struct {
	ID       string                 `json:"id" binding:"required"`
	Source   string                 `json:"source" binding:"required"`
	Target   string                 `json:"target" binding:"required"`
	Type     EdgeType               `json:"type,omitempty"`
	Animated bool                   `json:"animated,omitempty"`
	Label    string                 `json:"label,omitempty"`
	Data     EdgeData               `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ===== Node and Edge Types =====

// NodeType represents the type of a workflow node
type NodeType string

const (
	NodeTypeStart    NodeType = "start"
	NodeTypeWorkItem NodeType = "work_item"
	NodeTypeDecision NodeType = "decision"
	NodeTypeParallel NodeType = "parallel"
	NodeTypeEnd      NodeType = "end"
	// Legacy types for compatibility
	NodeTypeDocument NodeType = "document"
	NodeTypeSoftware NodeType = "software"
	NodeTypeProposal NodeType = "proposal"
	NodeTypeAnalysis NodeType = "analysis"
)

// EdgeType represents the type of connection between nodes
type EdgeType string

const (
	EdgeTypeDefault     EdgeType = "default"
	EdgeTypeSequential  EdgeType = "sequential"
	EdgeTypeConditional EdgeType = "conditional"
	EdgeTypeDataFlow    EdgeType = "dataflow"
)

// NodePosition represents the x,y coordinates of a node
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ===== Node Data Structures =====

// WorkflowNodeData contains node-specific configuration (merged from both models)
type WorkflowNodeData struct {
	// Common fields
	Name        string `json:"name,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Start node fields
	Trigger string `json:"trigger,omitempty"` // manual, scheduled, event, api

	// Work item node fields
	Type         string                 `json:"type,omitempty"`
	WorkItemID   string                 `json:"work_item_id,omitempty"`
	WorkItemType string                 `json:"work_item_type,omitempty"`
	Role         string                 `json:"role,omitempty"`
	SLAHours     int                    `json:"sla_hours,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Labels       []string               `json:"labels,omitempty"`
	LabelsText   string                 `json:"labels_text,omitempty"` // For UI binding

	// GitOps configuration (supporting multiple backends)
	GitConfig   *GitConfig   `json:"git_config,omitempty"`
	GiteaConfig *GiteaConfig `json:"gitea_config,omitempty"` // Legacy support

	// Execution state
	Status   string `json:"status,omitempty"` // pending, executing, completed, failed
	IssueID  int64  `json:"issue_id,omitempty"`
	IssueURL string `json:"issue_url,omitempty"`
	Error    string `json:"error,omitempty"`

	// Decision node fields
	Condition string `json:"condition,omitempty"`

	// Parallel node fields
	GatewayType string `json:"gateway_type,omitempty"` // fork, join

	// End node fields
	EndStatus string `json:"end_status,omitempty"` // success, failure
}

// GitConfig contains GitOps configuration for work item execution
// Supports multiple Git backends: Gitea, GitLab, GitHub, etc.
type GitConfig struct {
	Backend        string `json:"backend,omitempty"` // gitea, gitlab, github, etc.
	Repo           string `json:"repo" binding:"required"`
	BranchPattern  string `json:"branch_pattern,omitempty"`  // e.g., "issue-{issue_id}-{slug}"
	MergeStrategy  string `json:"merge_strategy,omitempty"`  // squash, merge, rebase
	AutoMerge      bool   `json:"auto_merge,omitempty"`      // Auto-merge on CI success
	RequireReviews int    `json:"require_reviews,omitempty"` // Number of required approvals
	RequireCI      bool   `json:"require_ci,omitempty"`      // Require CI to pass
}

// GiteaConfig contains GitOps settings for Gitea (legacy support)
type GiteaConfig struct {
	Repo           string `json:"repo" binding:"required"`
	BranchPattern  string `json:"branch_pattern"`
	MergeStrategy  string `json:"merge_strategy"` // squash, merge, rebase
	AutoMerge      bool   `json:"auto_merge"`
	RequireReviews int    `json:"require_reviews"`
}

// EdgeData contains edge-specific configuration
type EdgeData struct {
	Condition string `json:"condition,omitempty"`
	Label     string `json:"label,omitempty"`
}

// ===== Workflow Status =====

// WorkflowStatus represents the execution state of the workflow
type WorkflowStatus string

const (
	// Design phase
	WorkflowStatusDraft      WorkflowStatus = "draft"
	WorkflowStatusValidating WorkflowStatus = "validating"
	WorkflowStatusValid      WorkflowStatus = "valid"
	WorkflowStatusInvalid    WorkflowStatus = "invalid"
	// Execution phase
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusExecuting WorkflowStatus = "executing"
	WorkflowStatusPaused    WorkflowStatus = "paused"
	// Completion phase
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// NodeStatus represents the execution state of a node
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusWaiting   NodeStatus = "waiting"
	NodeStatusCompleted NodeStatus = "completed"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
)

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
	NodeID  string `json:"node_id,omitempty"`
	EdgeID  string `json:"edge_id,omitempty"`
}

// ===== Execution =====

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

// ===== Runtime Execution =====

// NodeExecution represents the execution state of a node
type NodeExecution struct {
	NodeID      string                 `json:"node_id"`
	Status      NodeStatus             `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	AgentID     string                 `json:"assigned_agent,omitempty"`
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID              string                 `json:"id"`
	WorkflowID      string                 `json:"workflow_id"`
	WorkflowVersion string                 `json:"workflow_version"`
	Status          WorkflowStatus         `json:"status"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	StartedBy       string                 `json:"started_by"`
	Context         map[string]interface{} `json:"context"`
	NodeExecutions  []NodeExecution        `json:"node_executions"`
	Errors          []string               `json:"errors"`
}
