package builder

import (
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// Message represents a chat message (shared type for AI interactions)
type Message struct {
	Role      string    `json:"role"`           // "system", "user", "assistant"
	Content   string    `json:"content"`        // Message content
	Name      string    `json:"name,omitempty"` // Optional speaker name
	Timestamp time.Time `json:"timestamp"`      // When message was created
}

// BuilderContext is a shared context structure used when building prompts for AI calls.
// It centralizes commonly used fields so all prompt builders can pass a typed
// structure instead of ad-hoc maps.
type BuilderContext struct {
	// Agency metadata
	AgencyName        string `json:"agency_name,omitempty"`
	AgencyCategory    string `json:"agency_category,omitempty"`
	AgencyDescription string `json:"agency_description,omitempty"`

	// Agency working data
	Introduction string                   `json:"introduction,omitempty"`
	Goals        []*models.Goal           `json:"goals,omitempty"`
	WorkItems    []*models.WorkItem       `json:"work_items,omitempty"`
	Roles        []*models.Role           `json:"roles,omitempty"` // Changed from registry.Role to models.Role
	Assignments  []*models.RACIAssignment `json:"assignments,omitempty"`
	UserInput    string                   `json:"user_input,omitempty"`
}
