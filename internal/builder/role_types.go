package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// RefineRoleRequest contains the context for refining a role (placeholder for future implementation)
type RefineRoleRequest struct {
	AgencyID      string             `json:"agency_id"`
	CurrentRole   *registry.Role     `json:"current_role"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	ExistingRoles []*registry.Role   `json:"existing_roles"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// RefineRoleResponse contains the AI-refined role (placeholder for future implementation)
type RefineRoleResponse struct {
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

// GenerateRoleRequest contains the context for generating a single role (placeholder for future implementation)
type GenerateRoleRequest struct {
	AgencyID      string             `json:"agency_id"`
	AgencyContext *agency.Agency     `json:"agency_context"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	ExistingRoles []*registry.Role   `json:"existing_roles"`
	UserInput     string             `json:"user_input"`
}

// GenerateRoleResponse contains the AI-generated role (placeholder for future implementation)
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

// GenerateRolesRequest contains the context for generating roles
type GenerateRolesRequest struct {
	AgencyID      string             `json:"agency_id"`
	AgencyContext *agency.Agency     `json:"agency_context"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	ExistingRoles []*registry.Role   `json:"existing_roles"`
}

// GenerateRolesResponse contains the AI-generated roles
type GenerateRolesResponse struct {
	Roles       []GeneratedRole `json:"roles"`
	Explanation string          `json:"explanation"`
}

// GeneratedRole represents a single AI-generated role
type GeneratedRole struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	AutonomyLevel  string   `json:"autonomy_level"`
	Capabilities   []string `json:"capabilities"`
	RequiredSkills []string `json:"required_skills"`
	TokenBudget    int64    `json:"token_budget"`
}

// ConsolidateRolesRequest contains the context for consolidating roles (placeholder for future implementation)
type ConsolidateRolesRequest struct {
	AgencyID      string             `json:"agency_id"`
	AgencyContext *agency.Agency     `json:"agency_context"`
	CurrentRoles  []*registry.Role   `json:"current_roles"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
}

// ConsolidateRolesResponse contains the consolidated roles (placeholder for future implementation)
type ConsolidateRolesResponse struct {
	ConsolidatedRoles []ConsolidatedRole `json:"consolidated_roles"`
	RemovedRoles      []string           `json:"removed_roles"` // Keys of roles that were consolidated/removed
	Summary           string             `json:"summary"`
	Explanation       string             `json:"explanation"`
}

// ConsolidatedRole represents a role after consolidation (placeholder for future implementation)
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
