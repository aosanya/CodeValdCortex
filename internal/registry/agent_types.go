package registry

import (
	"context"
	"encoding/json"
	"time"
)

// AgentType represents a type of agent with its schema and validation rules
type AgentType struct {
	// Key is the ArangoDB document key (same as ID)
	Key string `json:"_key,omitempty"`

	// ID is the unique identifier for this agent type (e.g., "pipe", "sensor")
	ID string `json:"id"`

	// Name is a human-readable name for this agent type
	Name string `json:"name"`

	// Description provides context about this agent type
	Description string `json:"description,omitempty"`

	// Category groups related agent types (e.g., "infrastructure", "monitoring", "coordination")
	Category string `json:"category"`

	// Version is the schema version for this agent type
	Version string `json:"version"`

	// Schema defines the expected structure for agent configuration (JSON Schema format)
	Schema json.RawMessage `json:"schema,omitempty"`

	// RequiredCapabilities lists capabilities this agent type must have
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`

	// OptionalCapabilities lists optional capabilities for this agent type
	OptionalCapabilities []string `json:"optional_capabilities,omitempty"`

	// DefaultConfig provides default configuration values
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`

	// ValidationRules defines custom validation rules beyond schema
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`

	// Metadata contains additional type information
	Metadata map[string]string `json:"metadata,omitempty"`

	// IsSystemType indicates if this is a core system type (cannot be deleted)
	IsSystemType bool `json:"is_system_type"`

	// IsEnabled indicates if agents of this type can be created
	IsEnabled bool `json:"is_enabled"`

	// CreatedAt is when this type was registered
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this type was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// CreatedBy tracks who created this type
	CreatedBy string `json:"created_by,omitempty"`
}

// ValidationRule defines a custom validation rule for an agent type
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

// AgentTypeRepository defines the interface for agent type storage
type AgentTypeRepository interface {
	// Create registers a new agent type
	Create(ctx context.Context, agentType *AgentType) error

	// Get retrieves an agent type by ID
	Get(ctx context.Context, id string) (*AgentType, error)

	// Update modifies an existing agent type
	Update(ctx context.Context, agentType *AgentType) error

	// Delete removes an agent type (only if not a system type)
	Delete(ctx context.Context, id string) error

	// List returns all registered agent types
	List(ctx context.Context) ([]*AgentType, error)

	// ListByCategory returns agent types in a specific category
	ListByCategory(ctx context.Context, category string) ([]*AgentType, error)

	// ListEnabled returns all enabled agent types
	ListEnabled(ctx context.Context) ([]*AgentType, error)

	// Exists checks if an agent type is registered
	Exists(ctx context.Context, id string) (bool, error)

	// Count returns the total number of agent types
	Count(ctx context.Context) (int64, error)
}

// AgentTypeService provides business logic for agent type management
type AgentTypeService interface {
	// RegisterType registers a new agent type with validation
	RegisterType(ctx context.Context, agentType *AgentType) error

	// GetType retrieves an agent type by ID
	GetType(ctx context.Context, id string) (*AgentType, error)

	// UpdateType updates an existing agent type
	UpdateType(ctx context.Context, agentType *AgentType) error

	// UnregisterType removes an agent type (with safety checks)
	UnregisterType(ctx context.Context, id string) error

	// ListTypes returns all agent types
	ListTypes(ctx context.Context) ([]*AgentType, error)

	// ListTypesByCategory returns agent types by category
	ListTypesByCategory(ctx context.Context, category string) ([]*AgentType, error)

	// IsValidType checks if an agent type ID is valid and enabled
	IsValidType(ctx context.Context, typeID string) (bool, error)

	// ValidateAgentConfig validates agent configuration against type schema
	ValidateAgentConfig(ctx context.Context, typeID string, config map[string]interface{}) error

	// EnableType enables an agent type
	EnableType(ctx context.Context, id string) error

	// DisableType disables an agent type
	DisableType(ctx context.Context, id string) error
}
