package templates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/configuration"
	"github.com/google/uuid"
)

// Template represents an agent configuration template
type Template struct {
	// ID is the unique identifier for the template
	ID string `json:"id" yaml:"id"`

	// Name is a human-readable name for the template
	Name string `json:"name" yaml:"name"`

	// Description provides context about the template
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Version is the template version
	Version string `json:"version" yaml:"version"`

	// AgentType specifies the type of agent this template is for
	AgentType string `json:"agent_type" yaml:"agent_type"`

	// Content is the template content (Go template syntax)
	Content string `json:"content" yaml:"content"`

	// Variables defines the variables that can be substituted
	Variables []TemplateVariable `json:"variables" yaml:"variables"`

	// BaseTemplateID references a parent template for inheritance
	BaseTemplateID string `json:"base_template_id,omitempty" yaml:"base_template_id,omitempty"`

	// Labels for categorization and selection
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`

	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`

	// CreatedBy tracks who created this template
	CreatedBy string `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

// TemplateVariable defines a variable that can be substituted in the template
type TemplateVariable struct {
	// Name is the variable name
	Name string `json:"name" yaml:"name"`

	// Type is the variable type (string, int, bool, etc.)
	Type string `json:"type" yaml:"type"`

	// Description explains what the variable is for
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// DefaultValue is the default value if not provided
	DefaultValue interface{} `json:"default_value,omitempty" yaml:"default_value,omitempty"`

	// Required indicates if the variable must be provided
	Required bool `json:"required" yaml:"required"`

	// ValidValues lists allowed values (for enum-type variables)
	ValidValues []interface{} `json:"valid_values,omitempty" yaml:"valid_values,omitempty"`

	// MinValue for numeric variables
	MinValue *float64 `json:"min_value,omitempty" yaml:"min_value,omitempty"`

	// MaxValue for numeric variables
	MaxValue *float64 `json:"max_value,omitempty" yaml:"max_value,omitempty"`

	// Pattern for string validation (regex)
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
}

// Engine manages template operations
type Engine struct {
	repository Repository
	validator  Validator
}

// Repository defines the interface for template storage
type Repository interface {
	// Store saves a template
	Store(ctx context.Context, template *Template) error

	// Get retrieves a template by ID
	Get(ctx context.Context, id string) (*Template, error)

	// List retrieves templates with optional filtering
	List(ctx context.Context, filter *ListFilter) ([]*Template, error)

	// Update updates an existing template
	Update(ctx context.Context, template *Template) error

	// Delete removes a template
	Delete(ctx context.Context, id string) error

	// GetByLabels retrieves templates matching labels
	GetByLabels(ctx context.Context, labels map[string]string) ([]*Template, error)
}

// Validator defines the interface for template validation
type Validator interface {
	// ValidateTemplate validates a template
	ValidateTemplate(template *Template) error

	// ValidateVariables validates template variables
	ValidateVariables(variables map[string]interface{}, templateVars []TemplateVariable) error
}

// ListFilter defines filtering options for listing templates
type ListFilter struct {
	// AgentType filters by agent type
	AgentType string

	// Labels filters by labels
	Labels map[string]string

	// CreatedAfter filters by creation time
	CreatedAfter *time.Time

	// CreatedBefore filters by creation time
	CreatedBefore *time.Time

	// Limit limits the number of results
	Limit int

	// Offset for pagination
	Offset int

	// SortBy field to sort by
	SortBy string

	// SortOrder (asc or desc)
	SortOrder string
}

// NewEngine creates a new template engine
func NewEngine(repo Repository, validator Validator) *Engine {
	return &Engine{
		repository: repo,
		validator:  validator,
	}
}

// CreateTemplate creates a new template
func (e *Engine) CreateTemplate(ctx context.Context, template *Template) (*Template, error) {
	// Set ID and timestamps if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	if template.Version == "" {
		template.Version = "1.0.0"
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt

	// Validate the template
	if err := e.validator.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Store the template
	if err := e.repository.Store(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to store template: %w", err)
	}

	return template, nil
}

// GetTemplate retrieves a template by ID
func (e *Engine) GetTemplate(ctx context.Context, id string) (*Template, error) {
	template, err := e.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return template, nil
}

// RenderTemplate renders a template with provided variables
func (e *Engine) RenderTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*configuration.AgentConfiguration, error) {
	// Get the template
	tmpl, err := e.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Validate variables
	if err := e.validator.ValidateVariables(variables, tmpl.Variables); err != nil {
		return nil, fmt.Errorf("variable validation failed: %w", err)
	}

	// Apply default values for missing variables
	mergedVars := e.mergeWithDefaults(variables, tmpl.Variables)

	// Handle template inheritance
	content, err := e.resolveInheritance(ctx, tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template inheritance: %w", err)
	}

	// Parse and execute the template
	goTemplate, err := template.New(tmpl.ID).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := goTemplate.Execute(&buf, mergedVars); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Parse the rendered JSON into a configuration
	var config configuration.AgentConfiguration
	if err := json.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rendered template: %w", err)
	}

	// Set generated ID if not present
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = config.CreatedAt

	return &config, nil
}

// ListTemplates lists templates with optional filtering
func (e *Engine) ListTemplates(ctx context.Context, filter *ListFilter) ([]*Template, error) {
	templates, err := e.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// UpdateTemplate updates an existing template
func (e *Engine) UpdateTemplate(ctx context.Context, template *Template) (*Template, error) {
	// Update timestamp
	template.UpdatedAt = time.Now()

	// Validate the template
	if err := e.validator.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Update in repository
	if err := e.repository.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return template, nil
}

// DeleteTemplate deletes a template
func (e *Engine) DeleteTemplate(ctx context.Context, id string) error {
	if err := e.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

// GetTemplatesByLabels retrieves templates matching labels
func (e *Engine) GetTemplatesByLabels(ctx context.Context, labels map[string]string) ([]*Template, error) {
	templates, err := e.repository.GetByLabels(ctx, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates by labels: %w", err)
	}

	return templates, nil
}

// ValidateTemplate validates a template without storing it
func (e *Engine) ValidateTemplate(template *Template) error {
	return e.validator.ValidateTemplate(template)
}

// ExportTemplate exports a template to JSON
func (e *Engine) ExportTemplate(ctx context.Context, id string) ([]byte, error) {
	template, err := e.GetTemplate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	return data, nil
}

// ImportTemplate imports a template from JSON
func (e *Engine) ImportTemplate(ctx context.Context, data []byte) (*Template, error) {
	var template Template
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	// Generate new ID to avoid conflicts
	template.ID = uuid.New().String()

	return e.CreateTemplate(ctx, &template)
}

// Helper methods

// mergeWithDefaults merges provided variables with default values
func (e *Engine) mergeWithDefaults(variables map[string]interface{}, templateVars []TemplateVariable) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy provided variables
	for k, v := range variables {
		merged[k] = v
	}

	// Add default values for missing variables
	for _, tv := range templateVars {
		if _, exists := merged[tv.Name]; !exists && tv.DefaultValue != nil {
			merged[tv.Name] = tv.DefaultValue
		}
	}

	return merged
}

// resolveInheritance resolves template inheritance by merging with base templates
func (e *Engine) resolveInheritance(ctx context.Context, tmpl *Template) (string, error) {
	content := tmpl.Content

	// If no base template, return content as-is
	if tmpl.BaseTemplateID == "" {
		return content, nil
	}

	// Get base template
	baseTemplate, err := e.GetTemplate(ctx, tmpl.BaseTemplateID)
	if err != nil {
		return "", fmt.Errorf("failed to get base template: %w", err)
	}

	// Recursively resolve base template inheritance
	baseContent, err := e.resolveInheritance(ctx, baseTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base template inheritance: %w", err)
	}

	// Simple inheritance: if current template content is empty, use base
	// For more complex inheritance, we could implement block replacement
	if strings.TrimSpace(content) == "" {
		return baseContent, nil
	}

	// For now, just return the current template content
	// TODO: Implement proper template block inheritance
	return content, nil
}

// GetTemplateVariables extracts variables from a template
func (e *Engine) GetTemplateVariables(ctx context.Context, templateID string) ([]TemplateVariable, error) {
	template, err := e.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return template.Variables, nil
}

// CloneTemplate creates a copy of an existing template
func (e *Engine) CloneTemplate(ctx context.Context, sourceID, name string) (*Template, error) {
	source, err := e.GetTemplate(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source template: %w", err)
	}

	// Create a clone
	clone := &Template{
		ID:             uuid.New().String(),
		Name:           name,
		Description:    source.Description,
		Version:        "1.0.0",
		AgentType:      source.AgentType,
		Content:        source.Content,
		Variables:      make([]TemplateVariable, len(source.Variables)),
		BaseTemplateID: source.BaseTemplateID,
		Labels:         make(map[string]string),
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      source.CreatedBy,
	}

	// Deep copy variables
	copy(clone.Variables, source.Variables)

	// Deep copy labels
	for k, v := range source.Labels {
		clone.Labels[k] = v
	}

	return e.CreateTemplate(ctx, clone)
}
