package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// DefaultRoleService implements RoleService
type DefaultRoleService struct {
	repo   RoleRepository
	logger *logrus.Logger
}

// NewRoleService creates a new role service
func NewRoleService(repo RoleRepository, logger *logrus.Logger) *DefaultRoleService {
	return &DefaultRoleService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterType registers a new role with validation
func (s *DefaultRoleService) RegisterType(ctx context.Context, agentType *Role) error {
	// Validate role definition
	if err := s.validateRole(agentType); err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	// Check if role already exists
	exists, err := s.repo.Exists(ctx, agentType.ID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if exists {
		// Update existing role
		if err := s.repo.Update(ctx, agentType); err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"type_id": agentType.ID,
			"name":    agentType.Name,
		}).Info("Role updated")
	} else {
		// Create new role
		if err := s.repo.Create(ctx, agentType); err != nil {
			return fmt.Errorf("failed to register role: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"type_id": agentType.ID,
			"name":    agentType.Name,
		}).Info("Role registered")
	}

	return nil
}

// GetType retrieves an role by ID
func (s *DefaultRoleService) GetType(ctx context.Context, id string) (*Role, error) {
	return s.repo.Get(ctx, id)
}

// UpdateType updates an existing role
func (s *DefaultRoleService) UpdateType(ctx context.Context, agentType *Role) error {
	// Validate role definition
	if err := s.validateRole(agentType); err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"type_id": agentType.ID,
		"name":    agentType.Name,
	}).Info("Role updated")

	return nil
}

// UnregisterType removes an role (with safety checks)
func (s *DefaultRoleService) UnregisterType(ctx context.Context, id string) error {
	// Get type to check if it's a system type
	agentType, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if agentType.IsSystemType {
		return fmt.Errorf("cannot unregister system type %s", id)
	}

	// TODO: Check if any agents are using this type
	// This would require integration with the agent registry

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to unregister role: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Role unregistered")

	return nil
}

// ListTypes returns all roles
func (s *DefaultRoleService) ListTypes(ctx context.Context) ([]*Role, error) {
	return s.repo.List(ctx)
}

// ListTypesByTags returns roles that have any of the specified tags
func (s *DefaultRoleService) ListTypesByTags(ctx context.Context, tags []string) ([]*Role, error) {
	return s.repo.ListByTags(ctx, tags)
}

// IsValidType checks if an role ID is valid and enabled
func (s *DefaultRoleService) IsValidType(ctx context.Context, typeID string) (bool, error) {
	agentType, err := s.repo.Get(ctx, typeID)
	if err != nil {
		return false, nil // Type doesn't exist, but not an error
	}

	return agentType.IsEnabled, nil
}

// ValidateAgentConfig validates agent configuration against type schema
func (s *DefaultRoleService) ValidateAgentConfig(ctx context.Context, typeID string, config map[string]interface{}) error {
	// Get role
	agentType, err := s.repo.Get(ctx, typeID)
	if err != nil {
		return fmt.Errorf("role %s not found: %w", typeID, err)
	}

	if !agentType.IsEnabled {
		return fmt.Errorf("role %s is disabled", typeID)
	}

	// If no schema defined, skip schema validation
	if len(agentType.Schema) == 0 {
		return nil
	}

	// Validate against JSON schema
	schemaLoader := gojsonschema.NewBytesLoader(agentType.Schema)
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configLoader := gojsonschema.NewBytesLoader(configBytes)

	result, err := gojsonschema.Validate(schemaLoader, configLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		errMsg := "configuration validation failed:"
		for _, desc := range result.Errors() {
			errMsg += fmt.Sprintf("\n  - %s", desc)
		}
		return fmt.Errorf(errMsg)
	}

	// Apply custom validation rules
	if err := s.applyValidationRules(agentType.ValidationRules, config); err != nil {
		return fmt.Errorf("custom validation failed: %w", err)
	}

	return nil
}

// EnableType enables an role
func (s *DefaultRoleService) EnableType(ctx context.Context, id string) error {
	agentType, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	agentType.IsEnabled = true

	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to enable role: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Agent type enabled")

	return nil
}

// DisableType disables an role
func (s *DefaultRoleService) DisableType(ctx context.Context, id string) error {
	agentType, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if agentType.IsSystemType {
		return fmt.Errorf("cannot disable system type %s", id)
	}

	agentType.IsEnabled = false

	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to disable role: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Agent type disabled")

	return nil
}

// validateRole validates an role definition
func (s *DefaultRoleService) validateRole(agentType *Role) error {
	if agentType.ID == "" {
		return fmt.Errorf("role ID cannot be empty")
	}

	if agentType.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if agentType.Version == "" {
		return fmt.Errorf("role version cannot be empty")
	}

	// Validate JSON schema if provided
	if len(agentType.Schema) > 0 {
		schemaLoader := gojsonschema.NewBytesLoader(agentType.Schema)
		_, err := gojsonschema.NewSchema(schemaLoader)
		if err != nil {
			return fmt.Errorf("invalid JSON schema: %w", err)
		}
	}

	return nil
}

// applyValidationRules applies custom validation rules to configuration
func (s *DefaultRoleService) applyValidationRules(rules []ValidationRule, config map[string]interface{}) error {
	for _, rule := range rules {
		value, exists := config[rule.Field]
		if !exists {
			continue // Field not present, skip
		}

		switch rule.RuleType {
		case "range":
			if err := s.validateRange(rule, value); err != nil {
				if rule.ErrorMessage != "" {
					return fmt.Errorf("%s: %w", rule.ErrorMessage, err)
				}
				return err
			}
		case "pattern":
			if err := s.validatePattern(rule, value); err != nil {
				if rule.ErrorMessage != "" {
					return fmt.Errorf("%s: %w", rule.ErrorMessage, err)
				}
				return err
			}
			// Add more rule types as needed
		}
	}

	return nil
}

// validateRange validates that a numeric value is within a range
func (s *DefaultRoleService) validateRange(rule ValidationRule, value interface{}) error {
	numValue, ok := value.(float64)
	if !ok {
		return fmt.Errorf("field %s must be numeric", rule.Field)
	}

	if min, ok := rule.Parameters["min"].(float64); ok {
		if numValue < min {
			return fmt.Errorf("field %s must be >= %v", rule.Field, min)
		}
	}

	if max, ok := rule.Parameters["max"].(float64); ok {
		if numValue > max {
			return fmt.Errorf("field %s must be <= %v", rule.Field, max)
		}
	}

	return nil
}

// validatePattern validates that a string value matches a pattern
func (s *DefaultRoleService) validatePattern(rule ValidationRule, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("field %s must be a string", rule.Field)
	}

	// Pattern validation would use regex here
	// For now, just a placeholder
	pattern, ok := rule.Parameters["pattern"].(string)
	if !ok {
		return fmt.Errorf("pattern rule requires pattern parameter")
	}

	// TODO: Implement actual regex matching
	_ = pattern
	_ = strValue

	return nil
}
