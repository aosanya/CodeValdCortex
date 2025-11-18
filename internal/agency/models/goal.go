package models

import "time"

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
