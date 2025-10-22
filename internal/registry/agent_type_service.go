package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// DefaultAgentTypeService implements AgentTypeService
type DefaultAgentTypeService struct {
	repo   AgentTypeRepository
	logger *logrus.Logger
}

// NewAgentTypeService creates a new agent type service
func NewAgentTypeService(repo AgentTypeRepository, logger *logrus.Logger) *DefaultAgentTypeService {
	return &DefaultAgentTypeService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterType registers a new agent type with validation
func (s *DefaultAgentTypeService) RegisterType(ctx context.Context, agentType *AgentType) error {
	// Validate agent type definition
	if err := s.validateAgentType(agentType); err != nil {
		return fmt.Errorf("invalid agent type: %w", err)
	}

	// Check if agent type already exists
	exists, err := s.repo.Exists(ctx, agentType.ID)
	if err != nil {
		return fmt.Errorf("failed to check if agent type exists: %w", err)
	}

	if exists {
		// Update existing agent type
		if err := s.repo.Update(ctx, agentType); err != nil {
			return fmt.Errorf("failed to update agent type: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"type_id":  agentType.ID,
			"name":     agentType.Name,
			"category": agentType.Category,
		}).Info("Agent type updated")
	} else {
		// Create new agent type
		if err := s.repo.Create(ctx, agentType); err != nil {
			return fmt.Errorf("failed to register agent type: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"type_id":  agentType.ID,
			"name":     agentType.Name,
			"category": agentType.Category,
		}).Info("Agent type registered")
	}

	return nil
}

// GetType retrieves an agent type by ID
func (s *DefaultAgentTypeService) GetType(ctx context.Context, id string) (*AgentType, error) {
	return s.repo.Get(ctx, id)
}

// UpdateType updates an existing agent type
func (s *DefaultAgentTypeService) UpdateType(ctx context.Context, agentType *AgentType) error {
	// Validate agent type definition
	if err := s.validateAgentType(agentType); err != nil {
		return fmt.Errorf("invalid agent type: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to update agent type: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"type_id":  agentType.ID,
		"name":     agentType.Name,
		"category": agentType.Category,
	}).Info("Agent type updated")

	return nil
}

// UnregisterType removes an agent type (with safety checks)
func (s *DefaultAgentTypeService) UnregisterType(ctx context.Context, id string) error {
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
		return fmt.Errorf("failed to unregister agent type: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Agent type unregistered")

	return nil
}

// ListTypes returns all agent types
func (s *DefaultAgentTypeService) ListTypes(ctx context.Context) ([]*AgentType, error) {
	return s.repo.List(ctx)
}

// ListTypesByCategory returns agent types by category
func (s *DefaultAgentTypeService) ListTypesByCategory(ctx context.Context, category string) ([]*AgentType, error) {
	return s.repo.ListByCategory(ctx, category)
}

// IsValidType checks if an agent type ID is valid and enabled
func (s *DefaultAgentTypeService) IsValidType(ctx context.Context, typeID string) (bool, error) {
	agentType, err := s.repo.Get(ctx, typeID)
	if err != nil {
		return false, nil // Type doesn't exist, but not an error
	}

	return agentType.IsEnabled, nil
}

// ValidateAgentConfig validates agent configuration against type schema
func (s *DefaultAgentTypeService) ValidateAgentConfig(ctx context.Context, typeID string, config map[string]interface{}) error {
	// Get agent type
	agentType, err := s.repo.Get(ctx, typeID)
	if err != nil {
		return fmt.Errorf("agent type %s not found: %w", typeID, err)
	}

	if !agentType.IsEnabled {
		return fmt.Errorf("agent type %s is disabled", typeID)
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

// EnableType enables an agent type
func (s *DefaultAgentTypeService) EnableType(ctx context.Context, id string) error {
	agentType, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	agentType.IsEnabled = true

	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to enable agent type: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Agent type enabled")

	return nil
}

// DisableType disables an agent type
func (s *DefaultAgentTypeService) DisableType(ctx context.Context, id string) error {
	agentType, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if agentType.IsSystemType {
		return fmt.Errorf("cannot disable system type %s", id)
	}

	agentType.IsEnabled = false

	if err := s.repo.Update(ctx, agentType); err != nil {
		return fmt.Errorf("failed to disable agent type: %w", err)
	}

	s.logger.WithField("type_id", id).Info("Agent type disabled")

	return nil
}

// validateAgentType validates an agent type definition
func (s *DefaultAgentTypeService) validateAgentType(agentType *AgentType) error {
	if agentType.ID == "" {
		return fmt.Errorf("agent type ID cannot be empty")
	}

	if agentType.Name == "" {
		return fmt.Errorf("agent type name cannot be empty")
	}

	if agentType.Category == "" {
		return fmt.Errorf("agent type category cannot be empty")
	}

	if agentType.Version == "" {
		return fmt.Errorf("agent type version cannot be empty")
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
func (s *DefaultAgentTypeService) applyValidationRules(rules []ValidationRule, config map[string]interface{}) error {
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
func (s *DefaultAgentTypeService) validateRange(rule ValidationRule, value interface{}) error {
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
func (s *DefaultAgentTypeService) validatePattern(rule ValidationRule, value interface{}) error {
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
