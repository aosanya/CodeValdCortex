package ai

import (
	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// AIContext is a shared context structure used when building prompts for AI calls.
// It centralizes commonly used fields so all prompt builders can pass a typed
// structure instead of ad-hoc maps.
type AIContext struct {
	// Agency metadata
	AgencyName        string `json:"agency_name,omitempty"`
	AgencyCategory    string `json:"agency_category,omitempty"`
	AgencyDescription string `json:"agency_description,omitempty"`

	// Agency working data
	Introduction string                   `json:"introduction,omitempty"`
	Goals        []*agency.Goal           `json:"goals,omitempty"`
	WorkItems    []*agency.WorkItem       `json:"work_items,omitempty"`
	Roles        []*registry.Role         `json:"roles,omitempty"`
	Assignments  []*agency.RACIAssignment `json:"assignments,omitempty"`
	UserInput    string                   `json:"user_input,omitempty"`
}
