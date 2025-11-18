package ai_refine

// Change represents a modification to any agency component
// This unified type is used for introduction, goals, work items, roles, RACI, workflows
type Change struct {
	Key         string         `json:"key,omitempty"`         // For updates (empty for creates) - the database _key
	Code        string         `json:"code,omitempty"`        // Component code (G-001, WI-001, etc.)
	Explanation string         `json:"explanation,omitempty"` // Brief explanation of what changed
	Content     map[string]any `json:"content,omitempty"`     // Actual content changes (field name -> new value)
}

// AgencyChanges represents changes made to agency components
// This is a shared type used across all AI refinement handlers (introduction, goals, work items, roles, RACI, workflows)
type AgencyChanges struct {
	Introduction *Change  `json:"introduction,omitempty"` // Agency introduction change
	Goals        []Change `json:"goals,omitempty"`        // Goals that were updated/created
	WorkItems    []Change `json:"work_items,omitempty"`   // Work items that were updated/created
	Roles        []Change `json:"roles,omitempty"`        // Roles that were updated/created
	RACI         []Change `json:"raci,omitempty"`         // RACI assignments that were updated/created
	Workflows    []Change `json:"workflows,omitempty"`    // Workflows that were updated/created
}

// AgencyDeletions represents items to be deleted from agency components
type AgencyDeletions struct {
	Goals     []string `json:"goals,omitempty"`      // Goal keys/codes to delete
	WorkItems []string `json:"work_items,omitempty"` // Work item keys/codes to delete
	Roles     []string `json:"roles,omitempty"`      // Role keys/codes to delete
	RACI      []string `json:"raci,omitempty"`       // RACI assignment keys to delete
	Workflows []string `json:"workflows,omitempty"`  // Workflow keys to delete
}

// AIProcessResult represents the AI-processed operation result
// Used by all dynamic refine handlers to determine what changed
type AIProcessResult struct {
	Operation string          `json:"operation"` // "update", "create", "delete", or "mixed"
	Changes   AgencyChanges   `json:"changes"`   // Dictionary of what changed
	Deletions AgencyDeletions `json:"deletions"` // Dictionary of what to delete
}
