package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Role represents a team role in the agency (e.g., Technical Lead, Domain Expert)
// These are the roles referenced in RACI assignments and work item responsibilities
// Different from agent roles (internal/registry/roles.go) which define agent types/schemas
type Role struct {
	Key           string    `json:"_key,omitempty"`           // ArangoDB document key
	ID            string    `json:"_id,omitempty"`            // Full document ID: "roles/{_key}"
	AgencyID      string    `json:"agency_id"`                // Reference to parent agency
	Code          string    `json:"code"`                     // Unique role code (e.g., "TECH-LEAD", "QA-ENGINEER")
	Name          string    `json:"name"`                     // Display name (e.g., "Technical Lead")
	Description   string    `json:"description"`              // Role responsibilities and purpose
	Tags          []string  `json:"tags,omitempty"`           // Tags for categorizing and filtering roles
	AutonomyLevel string    `json:"autonomy_level,omitempty"` // Level of autonomy (L0-L4)
	TokenBudget   int64     `json:"token_budget,omitempty"`   // Token budget for agents in this role
	Icon          string    `json:"icon,omitempty"`           // Visual icon (emoji or FontAwesome class)
	Color         string    `json:"color,omitempty"`          // Visual identification color
	IsActive      bool      `json:"is_active"`                // Whether role is currently active
	CreatedAt     time.Time `json:"created_at"`               // Creation timestamp
	UpdatedAt     time.Time `json:"updated_at"`               // Last update timestamp
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Code          string   `json:"code" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	AutonomyLevel string   `json:"autonomy_level"`
	TokenBudget   int64    `json:"token_budget"`
	Icon          string   `json:"icon"`
	Color         string   `json:"color"`
	IsActive      bool     `json:"is_active"` // Default: true
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Code          string   `json:"code" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	AutonomyLevel string   `json:"autonomy_level"`
	TokenBudget   int64    `json:"token_budget"`
	Icon          string   `json:"icon"`
	Color         string   `json:"color"`
	IsActive      bool     `json:"is_active"`
}

// StandardAgencyRoles defines the common roles used in agency operations
// Based on RACI matrix documentation
var StandardAgencyRoles = []Role{
	{
		Code:        "AGENCY-LEAD",
		Name:        "Agency Lead",
		Description: "Overall responsible for agency strategy and decisions",
		IsActive:    true,
	},
	{
		Code:        "TECH-LEAD",
		Name:        "Technical Lead",
		Description: "Responsible for technical architecture and implementation",
		IsActive:    true,
	},
	{
		Code:        "DOMAIN-EXPERT",
		Name:        "Domain Expert",
		Description: "Subject matter expert for the specific problem domain",
		IsActive:    true,
	},
	{
		Code:        "QA",
		Name:        "Quality Assurance",
		Description: "Ensures deliverable quality and compliance",
		IsActive:    true,
	},
	{
		Code:        "STAKEHOLDER-REP",
		Name:        "Stakeholder Representative",
		Description: "Represents end-user or client interests",
		IsActive:    true,
	},
	{
		Code:        "AGENT-COORD",
		Name:        "Agent Coordinator",
		Description: "Manages AI agent assignments and orchestration",
		IsActive:    true,
	},
	{
		Code:        "DATA-ANALYST",
		Name:        "Data Analyst",
		Description: "Handles data requirements and analysis",
		IsActive:    true,
	},
	{
		Code:        "SECURITY-OFFICER",
		Name:        "Security Officer",
		Description: "Ensures security compliance and risk management",
		IsActive:    true,
	},
}

// UnmarshalJSON implements custom JSON unmarshaling for Role to handle type conversions
func (r *Role) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with the same fields but Code as interface{}
	type TempRole struct {
		Key           string      `json:"_key,omitempty"`
		ID            string      `json:"_id,omitempty"`
		AgencyID      string      `json:"agency_id"`
		Code          interface{} `json:"code"` // Can be string or number
		Name          string      `json:"name"`
		Description   string      `json:"description"`
		Tags          []string    `json:"tags,omitempty"`
		AutonomyLevel string      `json:"autonomy_level,omitempty"`
		TokenBudget   int64       `json:"token_budget,omitempty"`
		Icon          string      `json:"icon,omitempty"`
		Color         string      `json:"color,omitempty"`
		IsActive      bool        `json:"is_active"`
		CreatedAt     time.Time   `json:"created_at"`
		UpdatedAt     time.Time   `json:"updated_at"`
	}

	var temp TempRole
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy all fields except Code
	r.Key = temp.Key
	r.ID = temp.ID
	r.AgencyID = temp.AgencyID
	r.Name = temp.Name
	r.Description = temp.Description
	r.Tags = temp.Tags
	r.AutonomyLevel = temp.AutonomyLevel
	r.TokenBudget = temp.TokenBudget
	r.Icon = temp.Icon
	r.Color = temp.Color
	r.IsActive = temp.IsActive
	r.CreatedAt = temp.CreatedAt
	r.UpdatedAt = temp.UpdatedAt

	// Handle Code field conversion
	switch v := temp.Code.(type) {
	case string:
		r.Code = v
	case float64:
		r.Code = strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		r.Code = strconv.Itoa(v)
	case int64:
		r.Code = strconv.FormatInt(v, 10)
	case nil:
		r.Code = ""
	default:
		r.Code = fmt.Sprintf("%v", v)
	}

	return nil
}
