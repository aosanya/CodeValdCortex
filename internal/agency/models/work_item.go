package models

import "time"

// WorkItem represents a work item in the agency
type WorkItem struct {
	Key          string    `json:"_key,omitempty"`
	ID           string    `json:"_id,omitempty"`
	AgencyID     string    `json:"agency_id"`
	Number       int       `json:"number"`
	Code         string    `json:"code"` // e.g., "WI-001"
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Deliverables []string  `json:"deliverables"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateWorkItemRequest is the request body for creating a work item
type CreateWorkItemRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
	Tags         []string `json:"tags,omitempty"`
}

// UpdateWorkItemRequest is the request body for updating a work item
type UpdateWorkItemRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
	Tags         []string `json:"tags,omitempty"`
}

// WorkItemRefineRequest is the request body for AI work item refinement
type WorkItemRefineRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
}

// WorkItemGoalLink represents an edge between a work item and a goal
// Stored in ArangoDB as an edge in the work_item_goals collection within the agency's database
type WorkItemGoalLink struct {
	Key          string    `json:"_key,omitempty"`
	ID           string    `json:"_id,omitempty"`
	From         string    `json:"_from"`         // Full ID: work_items/{work_item_key}
	To           string    `json:"_to"`           // Full ID: goals/{goal_key}
	WorkItemKey  string    `json:"work_item_key"` // Work item _key (denormalized for queries)
	GoalKey      string    `json:"goal_key"`      // Goal _key (denormalized for queries)
	Relationship string    `json:"relationship"`  // Type of relationship: "addresses", "supports", "implements", "depends_on"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateWorkItemGoalLinkRequest is the request to create a work item-goal link edge
type CreateWorkItemGoalLinkRequest struct {
	WorkItemKey  string `json:"work_item_key" binding:"required"`
	GoalKey      string `json:"goal_key" binding:"required"`
	Relationship string `json:"relationship"` // Default: "addresses"
}
