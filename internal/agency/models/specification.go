package models

import (
	"time"

	"github.com/google/uuid"
)

// AgencySpecification represents the complete specification document for an agency
// This is a single unified document containing all agency definition data:
// - Introduction (problem statement/context)
// - Goals (what the agency aims to achieve)
// - Work Items (tasks/activities to accomplish goals)
// - Roles (team structure and responsibilities)
// - RACI Matrix (responsibility assignments)
//
// Stored as a single document in ArangoDB for:
// - Atomic updates
// - Complete AI context in one fetch
// - Version control of entire specification
// - Simpler data model
type AgencySpecification struct {
	// ArangoDB fields
	Key string `json:"_key,omitempty"` // ArangoDB document key
	ID  string `json:"_id,omitempty"`  // Full document ID: "specifications/{_key}"
	Rev string `json:"_rev,omitempty"` // ArangoDB revision for optimistic locking

	// Ownership
	AgencyID string `json:"agency_id"` // Reference to parent agency

	// Version control
	Version   int       `json:"version"`    // Incremented on each update
	UpdatedAt time.Time `json:"updated_at"` // Last modification timestamp
	UpdatedBy string    `json:"updated_by"` // User/system that made the update
	CreatedAt time.Time `json:"created_at"` // Initial creation timestamp

	// Specification sections (embedded)
	Introduction string      `json:"introduction"` // Problem statement and context
	Goals        []Goal      `json:"goals"`        // Strategic objectives (reuses existing Goal model)
	WorkItems    []WorkItem  `json:"work_items"`   // Tasks and activities (reuses existing WorkItem model)
	Roles        []Role      `json:"roles"`        // Team roles and structure (uses Role model from role.go)
	RACIMatrix   *RACIMatrix `json:"raci_matrix"`  // Responsibility assignments (reuses existing RACIMatrix model)
}

// SpecificationUpdateRequest represents a request to update the entire specification
type SpecificationUpdateRequest struct {
	Introduction *string     `json:"introduction,omitempty"`
	Goals        *[]Goal     `json:"goals,omitempty"`
	WorkItems    *[]WorkItem `json:"work_items,omitempty"`
	Roles        *[]Role     `json:"roles,omitempty"`
	RACIMatrix   *RACIMatrix `json:"raci_matrix,omitempty"`
	UpdatedBy    string      `json:"updated_by,omitempty"`
}

// SpecificationPatchRequest represents a partial update to specific sections
type SpecificationPatchRequest struct {
	Section   string      `json:"section"` // "introduction", "goals", "work_items", "roles", "raci_matrix"
	Data      interface{} `json:"data"`    // Section-specific data
	UpdatedBy string      `json:"updated_by,omitempty"`
}

// GetSpecificationResponse is the API response for specification retrieval
type GetSpecificationResponse struct {
	Specification *AgencySpecification `json:"specification"`
	Message       string               `json:"message,omitempty"`
}

// CreateSpecificationRequest is the request to create a new specification
type CreateSpecificationRequest struct {
	Introduction string      `json:"introduction"`
	Goals        []Goal      `json:"goals,omitempty"`
	WorkItems    []WorkItem  `json:"work_items,omitempty"`
	Roles        []Role      `json:"roles,omitempty"`
	RACIMatrix   *RACIMatrix `json:"raci_matrix,omitempty"`
}

// NewAgencySpecification creates a new specification with default values
func NewAgencySpecification(agencyID string) *AgencySpecification {
	now := time.Now()
	return &AgencySpecification{
		AgencyID:     agencyID,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Introduction: "",
		Goals:        []Goal{},
		WorkItems:    []WorkItem{},
		Roles:        []Role{},
		RACIMatrix:   nil,
	}
}

// IncrementVersion increments the version and updates the timestamp
func (s *AgencySpecification) IncrementVersion(updatedBy string) {
	s.Version++
	s.UpdatedAt = time.Now()
	s.UpdatedBy = updatedBy
}

// UpdateIntroduction updates the introduction section
func (s *AgencySpecification) UpdateIntroduction(intro string, updatedBy string) {
	s.Introduction = intro
	s.IncrementVersion(updatedBy)
}

// SetGoals replaces all goals
func (s *AgencySpecification) SetGoals(goals []Goal, updatedBy string) {
	// Generate keys for goals that don't have them
	for i := range goals {
		if goals[i].Key == "" {
			goals[i].Key = uuid.New().String()
		}
		// Set timestamps
		if goals[i].CreatedAt.IsZero() {
			goals[i].CreatedAt = time.Now()
		}
		goals[i].UpdatedAt = time.Now()
	}

	s.Goals = goals
	s.IncrementVersion(updatedBy)
}

// SetWorkItems replaces all work items
func (s *AgencySpecification) SetWorkItems(items []WorkItem, updatedBy string) {
	// Generate keys for work items that don't have them
	for i := range items {
		if items[i].Key == "" {
			items[i].Key = uuid.New().String()
		}
		// Set timestamps
		if items[i].CreatedAt.IsZero() {
			items[i].CreatedAt = time.Now()
		}
		items[i].UpdatedAt = time.Now()
	}

	s.WorkItems = items
	s.IncrementVersion(updatedBy)
}

// SetRoles replaces all roles
func (s *AgencySpecification) SetRoles(roles []Role, updatedBy string) {
	// Generate keys for roles that don't have them
	for i := range roles {
		if roles[i].Key == "" {
			roles[i].Key = uuid.New().String()
		}
		// Set timestamps
		if roles[i].CreatedAt.IsZero() {
			roles[i].CreatedAt = time.Now()
		}
		roles[i].UpdatedAt = time.Now()
	}

	s.Roles = roles
	s.IncrementVersion(updatedBy)
}

// SetRACIMatrix replaces the RACI matrix
func (s *AgencySpecification) SetRACIMatrix(matrix *RACIMatrix, updatedBy string) {
	s.RACIMatrix = matrix
	s.IncrementVersion(updatedBy)
}
