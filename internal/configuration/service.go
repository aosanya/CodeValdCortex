package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service manages agent configurations
type Service struct {
	repository Repository
	validator  Validator
	notifier   NotificationService
	cache      Cache
	mu         sync.RWMutex
}

// Repository defines the interface for configuration storage
type Repository interface {
	// Store saves an agent configuration
	Store(ctx context.Context, config *AgentConfiguration) error

	// Get retrieves an agent configuration by ID
	Get(ctx context.Context, id string) (*AgentConfiguration, error)

	// List retrieves agent configurations with optional filtering
	List(ctx context.Context, filter *ListFilter) ([]*AgentConfiguration, error)

	// Update updates an existing configuration
	Update(ctx context.Context, config *AgentConfiguration) error

	// Delete removes a configuration
	Delete(ctx context.Context, id string) error

	// GetVersions retrieves all versions of a configuration
	GetVersions(ctx context.Context, configID string) ([]*AgentConfiguration, error)

	// GetByLabels retrieves configurations matching labels
	GetByLabels(ctx context.Context, labels map[string]string) ([]*AgentConfiguration, error)
}

// Validator defines the interface for configuration validation
type Validator interface {
	// Validate validates a configuration
	Validate(config *AgentConfiguration) error

	// ValidateCompatibility checks if a configuration is compatible with the current system
	ValidateCompatibility(config *AgentConfiguration) error

	// ValidateResources checks if resources are available for the configuration
	ValidateResources(config *AgentConfiguration) error
}

// NotificationService defines the interface for configuration change notifications
type NotificationService interface {
	// NotifyConfigurationCreated notifies about new configuration
	NotifyConfigurationCreated(config *AgentConfiguration) error

	// NotifyConfigurationUpdated notifies about configuration update
	NotifyConfigurationUpdated(oldConfig, newConfig *AgentConfiguration) error

	// NotifyConfigurationDeleted notifies about configuration deletion
	NotifyConfigurationDeleted(configID string) error

	// NotifyConfigurationApplied notifies about configuration application
	NotifyConfigurationApplied(agentID, configID string) error
}

// Cache defines the interface for configuration caching
type Cache interface {
	// Get retrieves a cached configuration
	Get(id string) (*AgentConfiguration, bool)

	// Set stores a configuration in cache
	Set(id string, config *AgentConfiguration, ttl time.Duration)

	// Delete removes a configuration from cache
	Delete(id string)

	// Clear clears all cached configurations
	Clear()
}

// ListFilter defines filtering options for listing configurations
type ListFilter struct {
	// AgentType filters by agent type
	AgentType string

	// Labels filters by labels (all must match)
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

// ConfigurationEvent represents a configuration change event
type ConfigurationEvent struct {
	// Type of event (created, updated, deleted, applied)
	Type string `json:"type"`

	// ConfigurationID is the ID of the affected configuration
	ConfigurationID string `json:"configuration_id"`

	// AgentID is the ID of the agent (for applied events)
	AgentID string `json:"agent_id,omitempty"`

	// Timestamp when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Data contains event-specific data
	Data map[string]interface{} `json:"data,omitempty"`
}

// NewService creates a new configuration service
func NewService(repo Repository, validator Validator, notifier NotificationService, cache Cache) *Service {
	return &Service{
		repository: repo,
		validator:  validator,
		notifier:   notifier,
		cache:      cache,
	}
}

// CreateConfiguration creates a new agent configuration
func (s *Service) CreateConfiguration(ctx context.Context, config *AgentConfiguration) (*AgentConfiguration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set ID and timestamps if not provided
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = config.CreatedAt

	// Validate the configuration
	if err := s.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Check compatibility
	if err := s.validator.ValidateCompatibility(config); err != nil {
		return nil, fmt.Errorf("configuration compatibility check failed: %w", err)
	}

	// Store the configuration
	if err := s.repository.Store(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to store configuration: %w", err)
	}

	// Cache the configuration
	s.cache.Set(config.ID, config, time.Hour)

	// Send notification
	if err := s.notifier.NotifyConfigurationCreated(config); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return config, nil
}

// GetConfiguration retrieves a configuration by ID
func (s *Service) GetConfiguration(ctx context.Context, id string) (*AgentConfiguration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try cache first
	if config, found := s.cache.Get(id); found {
		return config, nil
	}

	// Get from repository
	config, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	// Cache the result
	s.cache.Set(id, config, time.Hour)

	return config, nil
}

// UpdateConfiguration updates an existing configuration
func (s *Service) UpdateConfiguration(ctx context.Context, config *AgentConfiguration) (*AgentConfiguration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the existing configuration
	oldConfig, err := s.repository.Get(ctx, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing configuration: %w", err)
	}

	// Update version and timestamp
	config.UpdatedAt = time.Now()
	// TODO: Implement proper semantic versioning

	// Validate the updated configuration
	if err := s.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Check compatibility
	if err := s.validator.ValidateCompatibility(config); err != nil {
		return nil, fmt.Errorf("configuration compatibility check failed: %w", err)
	}

	// Update in repository
	if err := s.repository.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update configuration: %w", err)
	}

	// Update cache
	s.cache.Set(config.ID, config, time.Hour)

	// Send notification
	if err := s.notifier.NotifyConfigurationUpdated(oldConfig, config); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return config, nil
}

// DeleteConfiguration deletes a configuration
func (s *Service) DeleteConfiguration(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete from repository
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	// Remove from cache
	s.cache.Delete(id)

	// Send notification
	if err := s.notifier.NotifyConfigurationDeleted(id); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return nil
}

// ListConfigurations lists configurations with optional filtering
func (s *Service) ListConfigurations(ctx context.Context, filter *ListFilter) ([]*AgentConfiguration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}

	// Cache the results
	for _, config := range configs {
		s.cache.Set(config.ID, config, time.Hour)
	}

	return configs, nil
}

// GetConfigurationVersions retrieves all versions of a configuration
func (s *Service) GetConfigurationVersions(ctx context.Context, configID string) ([]*AgentConfiguration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, err := s.repository.GetVersions(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration versions: %w", err)
	}

	return versions, nil
}

// GetConfigurationsByLabels retrieves configurations matching labels
func (s *Service) GetConfigurationsByLabels(ctx context.Context, labels map[string]string) ([]*AgentConfiguration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs, err := s.repository.GetByLabels(ctx, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations by labels: %w", err)
	}

	// Cache the results
	for _, config := range configs {
		s.cache.Set(config.ID, config, time.Hour)
	}

	return configs, nil
}

// ApplyConfiguration applies a configuration to an agent
func (s *Service) ApplyConfiguration(ctx context.Context, agentID, configID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the configuration
	config, err := s.GetConfiguration(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Validate resources are available
	if err := s.validator.ValidateResources(config); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}

	// TODO: Apply configuration to agent through agent service
	// This would involve updating the agent's runtime configuration

	// Send notification
	if err := s.notifier.NotifyConfigurationApplied(agentID, configID); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return nil
}

// ValidateConfiguration validates a configuration without storing it
func (s *Service) ValidateConfiguration(config *AgentConfiguration) error {
	return s.validator.Validate(config)
}

// ExportConfiguration exports a configuration to JSON
func (s *Service) ExportConfiguration(ctx context.Context, id string) ([]byte, error) {
	config, err := s.GetConfiguration(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration: %w", err)
	}

	return data, nil
}

// ImportConfiguration imports a configuration from JSON
func (s *Service) ImportConfiguration(ctx context.Context, data []byte) (*AgentConfiguration, error) {
	var config AgentConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Generate new ID to avoid conflicts
	config.ID = uuid.New().String()

	return s.CreateConfiguration(ctx, &config)
}

// CloneConfiguration creates a copy of an existing configuration
func (s *Service) CloneConfiguration(ctx context.Context, sourceID, name string) (*AgentConfiguration, error) {
	source, err := s.GetConfiguration(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source configuration: %w", err)
	}

	clone, err := source.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone configuration: %w", err)
	}

	// Update clone properties
	clone.ID = uuid.New().String()
	clone.Name = name
	clone.Version = "1.0.0"
	clone.CreatedAt = time.Time{}
	clone.UpdatedAt = time.Time{}

	return s.CreateConfiguration(ctx, clone)
}
