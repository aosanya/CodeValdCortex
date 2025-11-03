package registry

import (
	"context"
	"encoding/json"
	"time"
)

// Role represents a role that can be assigned to agents (both AI and human workers)
// with its schema and validation rules
type Role struct {
	// Key is the ArangoDB document key (same as ID)
	Key string `json:"_key,omitempty"`

	// ID is the unique identifier for this role (e.g., "pipe", "sensor")
	ID string `json:"id"`

	// Name is a human-readable name for this role
	Name string `json:"name"`

	// Description provides context about this role
	Description string `json:"description,omitempty"`

	// Category groups related roles (e.g., "infrastructure", "monitoring", "coordination")
	Category string `json:"category"`

	// Version is the schema version for this role
	Version string `json:"version"`

	// Schema defines the expected structure for agent configuration (JSON Schema format)
	Schema json.RawMessage `json:"schema,omitempty"`

	// RequiredCapabilities lists capabilities this role must have
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`

	// OptionalCapabilities lists optional capabilities for this role
	OptionalCapabilities []string `json:"optional_capabilities,omitempty"`

	// DefaultConfig provides default configuration values
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`

	// ValidationRules defines custom validation rules beyond schema
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`

	// Metadata contains additional type information (supports strings, arrays, and nested objects)
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// IsSystemType indicates if this is a core system role (cannot be deleted)
	IsSystemType bool `json:"is_system_type"`

	// IsEnabled indicates if agents in this role can be created
	IsEnabled bool `json:"is_enabled"`

	// CreatedAt is when this role was registered
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this role was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// CreatedBy tracks who created this role
	CreatedBy string `json:"created_by,omitempty"`
}

// ValidationRule defines a custom validation rule for a role
type ValidationRule struct {
	// Field is the configuration field this rule applies to
	Field string `json:"field"`

	// RuleType specifies the type of validation (e.g., "range", "pattern", "custom")
	RuleType string `json:"rule_type"`

	// Parameters contains rule-specific parameters
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// ErrorMessage is shown when validation fails
	ErrorMessage string `json:"error_message,omitempty"`
}

// RoleRepository defines the interface for role storage
type RoleRepository interface {
	// Create registers a new role
	Create(ctx context.Context, role *Role) error

	// Get retrieves a role by ID
	Get(ctx context.Context, id string) (*Role, error)

	// Update modifies an existing role
	Update(ctx context.Context, role *Role) error

	// Delete removes a role (only if not a system type)
	Delete(ctx context.Context, id string) error

	// List returns all registered roles
	List(ctx context.Context) ([]*Role, error)

	// ListByCategory returns roles in a specific category
	ListByCategory(ctx context.Context, category string) ([]*Role, error)

	// ListEnabled returns all enabled roles
	ListEnabled(ctx context.Context) ([]*Role, error)

	// Exists checks if a role is registered
	Exists(ctx context.Context, id string) (bool, error)

	// Count returns the total number of roles
	Count(ctx context.Context) (int64, error)
}

// RoleService provides business logic for role management
type RoleService interface {
	// RegisterType registers a new role with validation
	RegisterType(ctx context.Context, role *Role) error

	// GetType retrieves a role by ID
	GetType(ctx context.Context, id string) (*Role, error)

	// UpdateType updates an existing role
	UpdateType(ctx context.Context, role *Role) error

	// UnregisterType removes a role (with safety checks)
	UnregisterType(ctx context.Context, id string) error

	// ListTypes returns all roles
	ListTypes(ctx context.Context) ([]*Role, error)

	// ListTypesByCategory returns roles by category
	ListTypesByCategory(ctx context.Context, category string) ([]*Role, error)

	// IsValidType checks if a role ID is valid and enabled
	IsValidType(ctx context.Context, typeID string) (bool, error)

	// ValidateAgentConfig validates agent configuration against role schema
	ValidateAgentConfig(ctx context.Context, typeID string, config map[string]interface{}) error

	// EnableType enables a role
	EnableType(ctx context.Context, id string) error

	// DisableType disables a role
	DisableType(ctx context.Context, id string) error
}
