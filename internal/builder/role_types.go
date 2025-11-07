package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// GenerateRoleResponse contains the AI-generated role (used in dynamic responses)
type GenerateRoleResponse struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	AutonomyLevel  string   `json:"autonomy_level"`
	Capabilities   []string `json:"capabilities"`
	RequiredSkills []string `json:"required_skills"`
	TokenBudget    int64    `json:"token_budget"`
	Explanation    string   `json:"explanation"`
}

// ConsolidateRolesResponse contains the consolidated roles (used in dynamic responses)
type ConsolidateRolesResponse struct {
	ConsolidatedRoles []ConsolidatedRole `json:"consolidated_roles"`
	RemovedRoles      []string           `json:"removed_roles"` // Keys of roles that were consolidated/removed
	Summary           string             `json:"summary"`
	Explanation       string             `json:"explanation"`
}

// ConsolidatedRole represents a role after consolidation (used in dynamic responses)
type ConsolidatedRole struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	AutonomyLevel    string   `json:"autonomy_level"`
	Capabilities     []string `json:"capabilities"`
	RequiredSkills   []string `json:"required_skills"`
	TokenBudget      int64    `json:"token_budget"`
	ConsolidatedFrom []string `json:"consolidated_from"` // Keys of original roles
	Rationale        string   `json:"rationale"`
}

// RefineRolesRequest contains the context for dynamic role processing
type RefineRolesRequest struct {
	AgencyID      string             `json:"agency_id"`
	UserMessage   string             `json:"user_message"`
	TargetRoles   []*registry.Role   `json:"target_roles"`   // Specific roles to operate on (nil means all)
	ExistingRoles []*registry.Role   `json:"existing_roles"` // All current roles for context
	WorkItems     []*models.WorkItem `json:"work_items"`     // Work items for context
	AgencyContext *models.Agency     `json:"agency_context"`
}

// RefineRolesResponse contains the dynamic role processing results
type RefineRolesResponse struct {
	Action           string                    `json:"action"`            // What action was determined (refine, generate, consolidate, enhance_all, etc.)
	RefinedRoles     []RefinedRoleResult       `json:"refined_roles"`     // Roles that were refined
	GeneratedRoles   []GenerateRoleResponse    `json:"generated_roles"`   // Newly generated roles
	ConsolidatedData *ConsolidateRolesResponse `json:"consolidated_data"` // Consolidation results if applicable
	Explanation      string                    `json:"explanation"`       // What was done and why
	NoActionNeeded   bool                      `json:"no_action_needed"`  // True if roles are already optimal
}

// RefinedRoleResult represents a single refined role
type RefinedRoleResult struct {
	OriginalKey            string   `json:"original_key"`
	RefinedName            string   `json:"refined_name"`
	RefinedDescription     string   `json:"refined_description"`
	SuggestedAutonomyLevel string   `json:"suggested_autonomy_level"`
	SuggestedCapabilities  []string `json:"suggested_capabilities"`
	SuggestedSkills        []string `json:"suggested_skills"`
	SuggestedTokenBudget   int64    `json:"suggested_token_budget"`
	SuggestedTags          []string `json:"suggested_tags"`
	WasChanged             bool     `json:"was_changed"`
	Explanation            string   `json:"explanation"`
}
