package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// RefineRACIMappingRequest contains the context for refining a RACI mapping (placeholder for future implementation)
type RefineRACIMappingRequest struct {
	AgencyID         string                   `json:"agency_id"`
	WorkItemKey      string                   `json:"work_item_key"`
	RoleKey          string                   `json:"role_key"`
	CurrentRaci      string                   `json:"current_raci"` // R, A, C, or I
	CurrentObjective string                   `json:"current_objective"`
	WorkItems        []*agency.WorkItem       `json:"work_items"`
	Roles            []*registry.Role         `json:"roles"`
	AllAssignments   []*agency.RACIAssignment `json:"all_assignments"`
	AgencyContext    *agency.Agency           `json:"agency_context"`
}

// RefineRACIMappingResponse contains the AI-refined RACI mapping (placeholder for future implementation)
type RefineRACIMappingResponse struct {
	RefinedRaci      string `json:"refined_raci"`
	RefinedObjective string `json:"refined_objective"`
	WasChanged       bool   `json:"was_changed"`
	Explanation      string `json:"explanation"`
}

// GenerateRACIMappingRequest contains the context for generating a single RACI mapping (placeholder for future implementation)
type GenerateRACIMappingRequest struct {
	AgencyID       string                   `json:"agency_id"`
	WorkItemKey    string                   `json:"work_item_key"`
	RoleKey        string                   `json:"role_key"`
	WorkItems      []*agency.WorkItem       `json:"work_items"`
	Roles          []*registry.Role         `json:"roles"`
	AllAssignments []*agency.RACIAssignment `json:"all_assignments"`
	AgencyContext  *agency.Agency           `json:"agency_context"`
	UserInput      string                   `json:"user_input"`
}

// GenerateRACIMappingResponse contains the AI-generated RACI mapping (placeholder for future implementation)
type GenerateRACIMappingResponse struct {
	Raci        string `json:"raci"`      // R, A, C, or I
	Objective   string `json:"objective"` // What this role needs to achieve
	Explanation string `json:"explanation"`
}

// CreateRACIMappingsRequest contains the context for creating RACI mappings
type CreateRACIMappingsRequest struct {
	AgencyID      string             `json:"agency_id"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	Roles         []*registry.Role   `json:"roles"`
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// RACIAssignment represents a role assignment with objective
type RACIAssignment struct {
	RACI      string `json:"raci"`      // R, A, C, or I
	Objective string `json:"objective"` // What this role needs to achieve
}

// CreateRACIMappingsResponse contains the AI-generated RACI mappings
type CreateRACIMappingsResponse struct {
	Assignments map[string]map[string]RACIAssignment `json:"assignments"` // workItemKey -> roleKey -> assignment
	Explanation string                               `json:"explanation"`
}

// ConsolidateRACIMappingsRequest contains the context for consolidating RACI mappings (placeholder for future implementation)
type ConsolidateRACIMappingsRequest struct {
	AgencyID           string                   `json:"agency_id"`
	AgencyContext      *agency.Agency           `json:"agency_context"`
	CurrentAssignments []*agency.RACIAssignment `json:"current_assignments"`
	WorkItems          []*agency.WorkItem       `json:"work_items"`
	Roles              []*registry.Role         `json:"roles"`
}

// ConsolidateRACIMappingsResponse contains the consolidated RACI mappings (placeholder for future implementation)
type ConsolidateRACIMappingsResponse struct {
	ConsolidatedAssignments map[string]map[string]RACIAssignment `json:"consolidated_assignments"` // workItemKey -> roleKey -> assignment
	RemovedAssignments      []string                             `json:"removed_assignments"`      // Keys of assignments that were removed
	Summary                 string                               `json:"summary"`
	Explanation             string                               `json:"explanation"`
}

// aiRACIMappingResponse represents the AI's response structure
type aiRACIMappingResponse struct {
	Mappings    []aiRACIMapping `json:"mappings"`
	Explanation string          `json:"explanation"`
}

// aiRACIMapping represents a single work item's RACI assignments
type aiRACIMapping struct {
	WorkItemKey string                      `json:"work_item_key"`
	Assignments map[string]aiRACIAssignment `json:"assignments"` // roleKey -> assignment
}

// aiRACIAssignment represents a role's assignment to a work item
type aiRACIAssignment struct {
	RACI      string `json:"raci"`
	Objective string `json:"objective"`
}
