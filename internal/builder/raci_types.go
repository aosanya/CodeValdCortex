package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// CreateRACIMappingsRequest contains the context for creating RACI mappings
type CreateRACIMappingsRequest struct {
	AgencyID      string             `json:"agency_id"`
	WorkItems     []*models.WorkItem `json:"work_items"`
	Roles         []*models.Role     `json:"roles"`
	AgencyContext *models.Agency     `json:"agency_context"`
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
	AgencyContext      *models.Agency           `json:"agency_context"`
	CurrentAssignments []*models.RACIAssignment `json:"current_assignments"`
	WorkItems          []*models.WorkItem       `json:"work_items"`
	Roles              []*models.Role           `json:"roles"`
}

// ConsolidateRACIMappingsResponse contains the consolidated RACI mappings (placeholder for future implementation)
type ConsolidateRACIMappingsResponse struct {
	ConsolidatedAssignments map[string]map[string]RACIAssignment `json:"consolidated_assignments"` // workItemKey -> roleKey -> assignment
	RemovedAssignments      []string                             `json:"removed_assignments"`      // Keys of assignments that were removed
	Summary                 string                               `json:"summary"`
	Explanation             string                               `json:"explanation"`
}

// RefineRACIMappingsRequest contains the context for dynamic RACI processing
type RefineRACIMappingsRequest struct {
	AgencyID            string                   `json:"agency_id"`
	UserMessage         string                   `json:"user_message"`
	TargetWorkItemKeys  []string                 `json:"target_work_item_keys"` // Specific work items to operate on (nil means all)
	TargetRoleKeys      []string                 `json:"target_role_keys"`      // Specific roles to operate on (nil means all)
	ExistingAssignments []*models.RACIAssignment `json:"existing_assignments"`  // All current RACI assignments for context
	WorkItems           []*models.WorkItem       `json:"work_items"`            // Work items for context
	Roles               []*models.Role           `json:"roles"`                 // Roles for context
	AgencyContext       *models.Agency           `json:"agency_context"`
}

// RefineRACIMappingsResponse contains the dynamic RACI processing results
type RefineRACIMappingsResponse struct {
	Action             string                           `json:"action"`              // What action was determined (refine, generate, consolidate, create_all, etc.)
	RefinedAssignments []RefinedRACIAssignmentResult    `json:"refined_assignments"` // RACI assignments that were refined
	GeneratedMappings  *CreateRACIMappingsResponse      `json:"generated_mappings"`  // Newly generated RACI mappings
	ConsolidatedData   *ConsolidateRACIMappingsResponse `json:"consolidated_data"`   // Consolidation results if applicable
	Explanation        string                           `json:"explanation"`         // What was done and why
	NoActionNeeded     bool                             `json:"no_action_needed"`    // True if RACI assignments are already optimal
}

// RefinedRACIAssignmentResult represents a single refined RACI assignment
type RefinedRACIAssignmentResult struct {
	WorkItemKey       string `json:"work_item_key"`
	RoleKey           string `json:"role_key"`
	OriginalRaci      string `json:"original_raci"`
	RefinedRaci       string `json:"refined_raci"`
	OriginalObjective string `json:"original_objective"`
	RefinedObjective  string `json:"refined_objective"`
	WasChanged        bool   `json:"was_changed"`
	Explanation       string `json:"explanation"`
}
