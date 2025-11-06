package workflow

import "time"

// NodeType represents the type of a workflow node
type NodeType string

const (
	NodeTypeStart    NodeType = "start"
	NodeTypeWorkItem NodeType = "work_item"
	NodeTypeDecision NodeType = "decision"
	NodeTypeParallel NodeType = "parallel"
	NodeTypeEnd      NodeType = "end"
)

// EdgeType represents the type of connection between nodes
type EdgeType string

const (
	EdgeTypeSequential  EdgeType = "sequential"
	EdgeTypeConditional EdgeType = "conditional"
	EdgeTypeDataFlow    EdgeType = "dataflow"
)

// WorkflowStatus represents the state of a workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusPaused    WorkflowStatus = "paused"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
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

// Position represents the x,y coordinates of a node on the canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData contains node-specific configuration
type NodeData struct {
	// Common fields
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	
	// Start node fields
	Trigger string `json:"trigger,omitempty"` // manual, scheduled, event, api
	
	// Work item node fields
	WorkItemID   string                 `json:"work_item_id,omitempty"`
	WorkItemType string                 `json:"work_item_type,omitempty"`
	Role         string                 `json:"role,omitempty"`
	SLAHours     int                    `json:"sla_hours,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	
	// Decision node fields
	Condition string `json:"condition,omitempty"`
	
	// Parallel node fields
	GatewayType string `json:"gateway_type,omitempty"` // fork, join
	
	// End node fields
	Status string `json:"status,omitempty"` // success, failure
}

// EdgeData contains edge-specific configuration
type EdgeData struct {
	Condition string `json:"condition,omitempty"`
	Label     string `json:"label,omitempty"`
}

// Node represents a workflow node
type Node struct {
	ID       string   `json:"id"`
	Type     NodeType `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

// Edge represents a connection between nodes
type Edge struct {
	ID     string   `json:"id"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   EdgeType `json:"type"`
	Data   EdgeData `json:"data"`
}

// Workflow represents a complete workflow definition
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Status      WorkflowStatus         `json:"status"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
	Variables   map[string]interface{} `json:"variables"`
	AgencyID    string                 `json:"agency_id"` // Link to agency
}

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

// ValidationError represents a workflow validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	EdgeID  string `json:"edge_id,omitempty"`
}

// ValidationResult contains the results of workflow validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}
