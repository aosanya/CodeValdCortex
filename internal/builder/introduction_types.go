package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// RefineIntroductionRequest contains the context for refining an introduction
type RefineIntroductionRequest struct {
	AgencyID            string    `json:"agency_id"`
	ConversationHistory []Message `json:"conversation_history,omitempty"` // Recent chat messages for context
}

// RefineIntroductionResponse contains the AI-refined introduction
type RefineIntroductionResponse struct {
	WasChanged      bool                `json:"was_changed"`
	Explanation     string              `json:"explanation"`
	ChangedSections []string            `json:"changed_sections"` // Array of section codes that were changed
	Data            *AgencyDataResponse `json:"data"`             // Complete updated agency data
}

// AgencyDataResponse contains the complete agency data structure
type AgencyDataResponse struct {
	Introduction string                   `json:"introduction"`
	Goals        []*agency.Goal           `json:"goals"`
	WorkItems    []*agency.WorkItem       `json:"work_items"`
	Roles        []*registry.Role         `json:"roles"`
	Assignments  []*agency.RACIAssignment `json:"assignments"`
}

// aiRefinementResponse represents the JSON structure returned by the AI
type aiRefinementResponse struct {
	Data            *AgencyDataResponse `json:"data"`
	Explanation     string              `json:"explanation"`
	Changed         bool                `json:"changed"`
	ChangedSections []string            `json:"changed_sections"`
}
