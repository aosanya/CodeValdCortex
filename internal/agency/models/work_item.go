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
	GoalKeys     []string  `json:"goal_keys,omitempty"` // Keys of goals this work item addresses
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateWorkItemRequest is the request body for creating a work item
type CreateWorkItemRequest struct {
	Code         string   `json:"code" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
	GoalKeys     []string `json:"goal_keys,omitempty"` // Keys of goals this work item addresses
	Tags         []string `json:"tags,omitempty"`
}

// UpdateWorkItemRequest is the request body for updating a work item
type UpdateWorkItemRequest struct {
	Code         string   `json:"code" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
	GoalKeys     []string `json:"goal_keys,omitempty"` // Keys of goals this work item addresses
	Tags         []string `json:"tags,omitempty"`
}

// WorkItemRefineRequest is the request body for AI work item refinement
type WorkItemRefineRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	Deliverables []string `json:"deliverables"`
}
