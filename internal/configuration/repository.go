package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Database defines a simple interface for database operations
type Database interface {
	Insert(ctx context.Context, collection string, document map[string]interface{}) error
	Query(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error)
	Update(ctx context.Context, collection string, key string, document map[string]interface{}) error
	Delete(ctx context.Context, collection string, key string) error
}

// ArangoRepository implements the Repository interface using ArangoDB
type ArangoRepository struct {
	db         Database
	collection string
	mu         sync.RWMutex
}

// NewArangoRepository creates a new ArangoDB-based configuration repository
func NewArangoRepository(db Database) *ArangoRepository {
	return &ArangoRepository{
		db:         db,
		collection: "agent_configurations",
	}
}

// Store saves an agent configuration
func (r *ArangoRepository) Store(ctx context.Context, config *AgentConfiguration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert to document
	doc := r.configToDocument(config)

	// Store in database
	if err := r.db.Insert(ctx, r.collection, doc); err != nil {
		return fmt.Errorf("failed to store configuration: %w", err)
	}

	return nil
}

// Get retrieves an agent configuration by ID
func (r *ArangoRepository) Get(ctx context.Context, id string) (*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Query by ID
	query := fmt.Sprintf(`
		FOR doc IN %s
		FILTER doc._key == @id
		RETURN doc
	`, r.collection)

	params := map[string]interface{}{
		"id": id,
	}

	results, err := r.db.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query configuration: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("configuration not found: %s", id)
	}

	// Convert document to configuration
	config, err := r.documentToConfig(results[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to configuration: %w", err)
	}

	return config, nil
}

// List retrieves agent configurations with optional filtering
func (r *ArangoRepository) List(ctx context.Context, filter *ListFilter) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build query
	query, params := r.buildListQuery(filter)

	results, err := r.db.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query configurations: %w", err)
	}

	// Convert documents to configurations
	configs := make([]*AgentConfiguration, 0, len(results))
	for _, result := range results {
		config, err := r.documentToConfig(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to configuration: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Update updates an existing configuration
func (r *ArangoRepository) Update(ctx context.Context, config *AgentConfiguration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert to document
	doc := r.configToDocument(config)

	// Update in database
	if err := r.db.Update(ctx, r.collection, config.ID, doc); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	return nil
}

// Delete removes a configuration
func (r *ArangoRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.db.Delete(ctx, r.collection, id); err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	return nil
}

// GetVersions retrieves all versions of a configuration
func (r *ArangoRepository) GetVersions(ctx context.Context, configID string) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Query by configuration name/base ID
	query := fmt.Sprintf(`
		FOR doc IN %s
		FILTER doc.name == @name OR doc.base_id == @config_id
		SORT doc.created_at DESC
		RETURN doc
	`, r.collection)

	params := map[string]interface{}{
		"name":      configID,
		"config_id": configID,
	}

	results, err := r.db.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query configuration versions: %w", err)
	}

	// Convert documents to configurations
	configs := make([]*AgentConfiguration, 0, len(results))
	for _, result := range results {
		config, err := r.documentToConfig(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to configuration: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// GetByLabels retrieves configurations matching labels
func (r *ArangoRepository) GetByLabels(ctx context.Context, labels map[string]string) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(labels) == 0 {
		return nil, fmt.Errorf("no labels provided")
	}

	// Build label filter conditions
	var conditions []string
	params := make(map[string]interface{})

	i := 0
	for key, value := range labels {
		condition := fmt.Sprintf("doc.labels.`%s` == @label_value_%d", key, i)
		conditions = append(conditions, condition)
		params[fmt.Sprintf("label_value_%d", i)] = value
		i++
	}

	query := fmt.Sprintf(`
		FOR doc IN %s
		FILTER %s
		SORT doc.created_at DESC
		RETURN doc
	`, r.collection, fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0])))

	// Add remaining conditions with AND
	for j := 1; j < len(conditions); j++ {
		query = query[:len(query)-len("RETURN doc")] + fmt.Sprintf(" AND %s\nRETURN doc", conditions[j])
	}

	results, err := r.db.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query configurations by labels: %w", err)
	}

	// Convert documents to configurations
	configs := make([]*AgentConfiguration, 0, len(results))
	for _, result := range results {
		config, err := r.documentToConfig(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to configuration: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Helper methods

// configToDocument converts a configuration to a database document
func (r *ArangoRepository) configToDocument(config *AgentConfiguration) map[string]interface{} {
	// Convert to JSON and back to map for easy storage
	data, _ := json.Marshal(config)
	var doc map[string]interface{}
	json.Unmarshal(data, &doc)

	// Set the document key to the configuration ID
	doc["_key"] = config.ID

	return doc
}

// documentToConfig converts a database document to a configuration
func (r *ArangoRepository) documentToConfig(doc map[string]interface{}) (*AgentConfiguration, error) {
	// Convert to JSON and unmarshal to configuration
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var config AgentConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &config, nil
}

// buildListQuery builds a query for listing configurations with filters
func (r *ArangoRepository) buildListQuery(filter *ListFilter) (string, map[string]interface{}) {
	params := make(map[string]interface{})
	conditions := []string{}

	// Base query
	query := fmt.Sprintf("FOR doc IN %s", r.collection)

	// Add filters
	if filter != nil {
		if filter.AgentType != "" {
			conditions = append(conditions, "doc.agent_type == @agent_type")
			params["agent_type"] = filter.AgentType
		}

		if filter.CreatedAfter != nil {
			conditions = append(conditions, "doc.created_at >= @created_after")
			params["created_after"] = filter.CreatedAfter.Format(time.RFC3339)
		}

		if filter.CreatedBefore != nil {
			conditions = append(conditions, "doc.created_at <= @created_before")
			params["created_before"] = filter.CreatedBefore.Format(time.RFC3339)
		}

		// Add label filters
		if len(filter.Labels) > 0 {
			i := 0
			for key, value := range filter.Labels {
				condition := fmt.Sprintf("doc.labels.`%s` == @label_value_%d", key, i)
				conditions = append(conditions, condition)
				params[fmt.Sprintf("label_value_%d", i)] = value
				i++
			}
		}
	}

	// Add WHERE clause if there are conditions
	if len(conditions) > 0 {
		query += "\nFILTER " + fmt.Sprintf("(%s)", conditions[0])
		for i := 1; i < len(conditions); i++ {
			query += fmt.Sprintf(" AND (%s)", conditions[i])
		}
	}

	// Add sorting
	if filter != nil && filter.SortBy != "" {
		direction := "ASC"
		if filter.SortOrder == "desc" {
			direction = "DESC"
		}
		query += fmt.Sprintf("\nSORT doc.%s %s", filter.SortBy, direction)
	} else {
		query += "\nSORT doc.created_at DESC"
	}

	// Add pagination
	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf("\nLIMIT %d, %d", filter.Offset, filter.Limit)
		}
	}

	query += "\nRETURN doc"

	return query, params
}

// InMemoryRepository implements the Repository interface using in-memory storage
// This is useful for testing and development
type InMemoryRepository struct {
	configs map[string]*AgentConfiguration
	mu      sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory configuration repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		configs: make(map[string]*AgentConfiguration),
	}
}

// Store saves an agent configuration
func (r *InMemoryRepository) Store(ctx context.Context, config *AgentConfiguration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clone the configuration to avoid external modifications
	clone, err := config.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone configuration: %w", err)
	}

	r.configs[config.ID] = clone
	return nil
}

// Get retrieves an agent configuration by ID
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.configs[id]
	if !exists {
		return nil, fmt.Errorf("configuration not found: %s", id)
	}

	// Clone to avoid external modifications
	clone, err := config.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone configuration: %w", err)
	}

	return clone, nil
}

// List retrieves agent configurations with optional filtering
func (r *InMemoryRepository) List(ctx context.Context, filter *ListFilter) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var configs []*AgentConfiguration

	// Collect all configs that match the filter
	for _, config := range r.configs {
		if r.matchesFilter(config, filter) {
			clone, err := config.Clone()
			if err != nil {
				return nil, fmt.Errorf("failed to clone configuration: %w", err)
			}
			configs = append(configs, clone)
		}
	}

	// Sort the results
	if filter != nil && filter.SortBy != "" {
		r.sortConfigs(configs, filter.SortBy, filter.SortOrder)
	} else {
		// Default sort by created_at desc
		sort.Slice(configs, func(i, j int) bool {
			return configs[i].CreatedAt.After(configs[j].CreatedAt)
		})
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(configs) {
			return []*AgentConfiguration{}, nil
		}
		end := start + filter.Limit
		if end > len(configs) {
			end = len(configs)
		}
		configs = configs[start:end]
	}

	return configs, nil
}

// Update updates an existing configuration
func (r *InMemoryRepository) Update(ctx context.Context, config *AgentConfiguration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[config.ID]; !exists {
		return fmt.Errorf("configuration not found: %s", config.ID)
	}

	// Clone the configuration to avoid external modifications
	clone, err := config.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone configuration: %w", err)
	}

	r.configs[config.ID] = clone
	return nil
}

// Delete removes a configuration
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[id]; !exists {
		return fmt.Errorf("configuration not found: %s", id)
	}

	delete(r.configs, id)
	return nil
}

// GetVersions retrieves all versions of a configuration
func (r *InMemoryRepository) GetVersions(ctx context.Context, configID string) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var versions []*AgentConfiguration

	// Find configurations with the same name (different versions)
	for _, config := range r.configs {
		if config.Name == configID || config.ID == configID {
			clone, err := config.Clone()
			if err != nil {
				return nil, fmt.Errorf("failed to clone configuration: %w", err)
			}
			versions = append(versions, clone)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreatedAt.After(versions[j].CreatedAt)
	})

	return versions, nil
}

// GetByLabels retrieves configurations matching labels
func (r *InMemoryRepository) GetByLabels(ctx context.Context, labels map[string]string) ([]*AgentConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var configs []*AgentConfiguration

	for _, config := range r.configs {
		if r.matchesLabels(config, labels) {
			clone, err := config.Clone()
			if err != nil {
				return nil, fmt.Errorf("failed to clone configuration: %w", err)
			}
			configs = append(configs, clone)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].CreatedAt.After(configs[j].CreatedAt)
	})

	return configs, nil
}

// Helper methods for InMemoryRepository

func (r *InMemoryRepository) matchesFilter(config *AgentConfiguration, filter *ListFilter) bool {
	if filter == nil {
		return true
	}

	// Check agent type
	if filter.AgentType != "" && config.AgentType != filter.AgentType {
		return false
	}

	// Check created after
	if filter.CreatedAfter != nil && config.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	// Check created before
	if filter.CreatedBefore != nil && config.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	// Check labels
	if !r.matchesLabels(config, filter.Labels) {
		return false
	}

	return true
}

func (r *InMemoryRepository) matchesLabels(config *AgentConfiguration, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}

	for key, value := range labels {
		if configValue, exists := config.Labels[key]; !exists || configValue != value {
			return false
		}
	}

	return true
}

func (r *InMemoryRepository) sortConfigs(configs []*AgentConfiguration, sortBy, sortOrder string) {
	ascending := sortOrder != "desc"

	switch sortBy {
	case "name":
		sort.Slice(configs, func(i, j int) bool {
			if ascending {
				return configs[i].Name < configs[j].Name
			}
			return configs[i].Name > configs[j].Name
		})
	case "agent_type":
		sort.Slice(configs, func(i, j int) bool {
			if ascending {
				return configs[i].AgentType < configs[j].AgentType
			}
			return configs[i].AgentType > configs[j].AgentType
		})
	case "created_at":
		sort.Slice(configs, func(i, j int) bool {
			if ascending {
				return configs[i].CreatedAt.Before(configs[j].CreatedAt)
			}
			return configs[i].CreatedAt.After(configs[j].CreatedAt)
		})
	case "updated_at":
		sort.Slice(configs, func(i, j int) bool {
			if ascending {
				return configs[i].UpdatedAt.Before(configs[j].UpdatedAt)
			}
			return configs[i].UpdatedAt.After(configs[j].UpdatedAt)
		})
	default:
		// Default to created_at desc
		sort.Slice(configs, func(i, j int) bool {
			return configs[i].CreatedAt.After(configs[j].CreatedAt)
		})
	}
}
