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
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	AgencyID    string         `json:"agency_id" binding:"required"`
	Version     string         `json:"version"`
	Nodes       []WorkflowNode `json:"nodes"`
	Edges       []WorkflowEdge `json:"edges"`
	Status      WorkflowStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   string         `json:"created_by"`
	UpdatedBy   string         `json:"updated_by"`
}

// WorkflowNode represents a node in the workflow (combining both models)
type WorkflowNode struct {
	ID       string           `json:"id" binding:"required"`
	Type     NodeType         `json:"type" binding:"required"`
	Position NodePosition     `json:"position" binding:"required"`
	Data     WorkflowNodeData `json:"data" binding:"required"`
	Width    int              `json:"width,omitempty"`
	Height   int              `json:"height,omitempty"`
}

// WorkflowEdge represents a connection between nodes
type WorkflowEdge struct {
	ID       string   `json:"id" binding:"required"`
	Source   string   `json:"source" binding:"required"`
	Target   string   `json:"target" binding:"required"`
	Type     EdgeType `json:"type,omitempty"`
	Animated bool     `json:"animated,omitempty"`
	Label    string   `json:"label,omitempty"`
	Data     EdgeData `json:"data,omitempty"`
}

// ===== Node and Edge Types =====

// NodeType represents the type of a workflow node
type NodeType string

const (
	NodeTypeWorkItem NodeType = "work_item"
)

// EdgeType represents the type of connection between nodes
type EdgeType string

// NodePosition represents the x,y coordinates of a node
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ===== Node Data Structures =====

// WorkflowNodeData contains node-specific configuration (merged from both models)
type WorkflowNodeData struct {
	// Work item node fields
	WorkItemKey string `json:"work_item_key,omitempty"`
}

// EdgeData contains edge-specific configuration
type EdgeData struct {
	Label string `json:"label,omitempty"`
}

// ===== Workflow Status =====

// WorkflowStatus represents the execution state of the workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft  WorkflowStatus = "draft"
	WorkflowStatusActive WorkflowStatus = "active"
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
