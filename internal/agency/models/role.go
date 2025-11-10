package models

import "time"

// Role represents a team role in the agency (e.g., Technical Lead, Domain Expert)
// These are the roles referenced in RACI assignments and work item responsibilities
// Different from agent roles (internal/registry/roles.go) which define agent types/schemas
type Role struct {
	Key         string    `json:"_key,omitempty"` // ArangoDB document key
	ID          string    `json:"_id,omitempty"`  // Full document ID: "roles/{_key}"
	AgencyID    string    `json:"agency_id"`      // Reference to parent agency
	Code        string    `json:"code"`           // Unique role code (e.g., "TECH-LEAD", "QA-ENGINEER")
	Name        string    `json:"name"`           // Display name (e.g., "Technical Lead")
	Description string    `json:"description"`    // Role responsibilities and purpose
	Skills      []string  `json:"skills"`         // Required skills/capabilities
	Level       string    `json:"level"`          // Seniority level (Junior, Mid, Senior, Lead)
	Department  string    `json:"department"`     // Organizational unit (optional)
	IsActive    bool      `json:"is_active"`      // Whether role is currently active
	CreatedAt   time.Time `json:"created_at"`     // Creation timestamp
	UpdatedAt   time.Time `json:"updated_at"`     // Last update timestamp
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Level       string   `json:"level"`      // Junior, Mid, Senior, Lead
	Department  string   `json:"department"` // Optional organizational unit
	IsActive    bool     `json:"is_active"`  // Default: true
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Level       string   `json:"level"`
	Department  string   `json:"department"`
	IsActive    bool     `json:"is_active"`
}

// StandardAgencyRoles defines the common roles used in agency operations
// Based on RACI matrix documentation
var StandardAgencyRoles = []Role{
	{
		Code:        "AGENCY-LEAD",
		Name:        "Agency Lead",
		Description: "Overall responsible for agency strategy and decisions",
		Level:       "Lead",
		IsActive:    true,
	},
	{
		Code:        "TECH-LEAD",
		Name:        "Technical Lead",
		Description: "Responsible for technical architecture and implementation",
		Level:       "Lead",
		IsActive:    true,
	},
	{
		Code:        "DOMAIN-EXPERT",
		Name:        "Domain Expert",
		Description: "Subject matter expert for the specific problem domain",
		Level:       "Senior",
		IsActive:    true,
	},
	{
		Code:        "QA",
		Name:        "Quality Assurance",
		Description: "Ensures deliverable quality and compliance",
		Level:       "Mid",
		IsActive:    true,
	},
	{
		Code:        "STAKEHOLDER-REP",
		Name:        "Stakeholder Representative",
		Description: "Represents end-user or client interests",
		Level:       "Mid",
		IsActive:    true,
	},
	{
		Code:        "AGENT-COORD",
		Name:        "Agent Coordinator",
		Description: "Manages AI agent assignments and orchestration",
		Level:       "Mid",
		IsActive:    true,
	},
	{
		Code:        "DATA-ANALYST",
		Name:        "Data Analyst",
		Description: "Handles data requirements and analysis",
		Level:       "Mid",
		IsActive:    true,
	},
	{
		Code:        "SECURITY-OFFICER",
		Name:        "Security Officer",
		Description: "Ensures security compliance and risk management",
		Level:       "Senior",
		IsActive:    true,
	},
}

// RoleLevel represents the seniority levels for roles
type RoleLevel string

const (
	RoleLevelJunior RoleLevel = "Junior"
	RoleLevelMid    RoleLevel = "Mid"
	RoleLevelSenior RoleLevel = "Senior"
	RoleLevelLead   RoleLevel = "Lead"
)

// ValidateRoleLevel checks if a role level is valid
func ValidateRoleLevel(level string) bool {
	switch RoleLevel(level) {
	case RoleLevelJunior, RoleLevelMid, RoleLevelSenior, RoleLevelLead:
		return true
	default:
		return level == "" // Empty is also valid (optional field)
	}
}
