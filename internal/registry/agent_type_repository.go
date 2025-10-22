package registry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryAgentTypeRepository implements AgentTypeRepository using in-memory storage
type InMemoryAgentTypeRepository struct {
	mu         sync.RWMutex
	types      map[string]*AgentType
	categories map[string][]*AgentType // Index by category
}

// NewInMemoryAgentTypeRepository creates a new in-memory repository
func NewInMemoryAgentTypeRepository() *InMemoryAgentTypeRepository {
	return &InMemoryAgentTypeRepository{
		types:      make(map[string]*AgentType),
		categories: make(map[string][]*AgentType),
	}
}

// Create registers a new agent type
func (r *InMemoryAgentTypeRepository) Create(ctx context.Context, agentType *AgentType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists
	if _, exists := r.types[agentType.ID]; exists {
		return fmt.Errorf("agent type %s already exists", agentType.ID)
	}

	// Set timestamps
	now := time.Now()
	agentType.CreatedAt = now
	agentType.UpdatedAt = now

	// Store type
	r.types[agentType.ID] = agentType

	// Update category index
	r.categories[agentType.Category] = append(r.categories[agentType.Category], agentType)

	return nil
}

// Get retrieves an agent type by ID
func (r *InMemoryAgentTypeRepository) Get(ctx context.Context, id string) (*AgentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentType, exists := r.types[id]
	if !exists {
		return nil, fmt.Errorf("agent type %s not found", id)
	}

	return agentType, nil
}

// Update modifies an existing agent type
func (r *InMemoryAgentTypeRepository) Update(ctx context.Context, agentType *AgentType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.types[agentType.ID]
	if !exists {
		return fmt.Errorf("agent type %s not found", agentType.ID)
	}

	// Prevent modification of system types' core fields
	if existing.IsSystemType && agentType.IsSystemType != existing.IsSystemType {
		return fmt.Errorf("cannot change system type flag for %s", agentType.ID)
	}

	// Update category index if category changed
	if existing.Category != agentType.Category {
		r.removeCategoryIndex(existing)
		r.categories[agentType.Category] = append(r.categories[agentType.Category], agentType)
	}

	// Preserve creation metadata
	agentType.CreatedAt = existing.CreatedAt
	agentType.CreatedBy = existing.CreatedBy
	agentType.UpdatedAt = time.Now()

	// Store updated type
	r.types[agentType.ID] = agentType

	return nil
}

// Delete removes an agent type
func (r *InMemoryAgentTypeRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agentType, exists := r.types[id]
	if !exists {
		return fmt.Errorf("agent type %s not found", id)
	}

	// Prevent deletion of system types
	if agentType.IsSystemType {
		return fmt.Errorf("cannot delete system type %s", id)
	}

	// Remove from category index
	r.removeCategoryIndex(agentType)

	// Delete type
	delete(r.types, id)

	return nil
}

// List returns all registered agent types
func (r *InMemoryAgentTypeRepository) List(ctx context.Context) ([]*AgentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]*AgentType, 0, len(r.types))
	for _, agentType := range r.types {
		types = append(types, agentType)
	}

	return types, nil
}

// ListByCategory returns agent types in a specific category
func (r *InMemoryAgentTypeRepository) ListByCategory(ctx context.Context, category string) ([]*AgentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := r.categories[category]
	if types == nil {
		return []*AgentType{}, nil
	}

	// Return a copy to prevent external modification
	result := make([]*AgentType, len(types))
	copy(result, types)

	return result, nil
}

// ListEnabled returns all enabled agent types
func (r *InMemoryAgentTypeRepository) ListEnabled(ctx context.Context) ([]*AgentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]*AgentType, 0)
	for _, agentType := range r.types {
		if agentType.IsEnabled {
			types = append(types, agentType)
		}
	}

	return types, nil
}

// Exists checks if an agent type is registered
func (r *InMemoryAgentTypeRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.types[id]
	return exists, nil
}

// Count returns the total number of agent types
func (r *InMemoryAgentTypeRepository) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.types)), nil
}

// removeCategoryIndex removes an agent type from the category index
// Must be called with lock held
func (r *InMemoryAgentTypeRepository) removeCategoryIndex(agentType *AgentType) {
	types := r.categories[agentType.Category]
	for i, t := range types {
		if t.ID == agentType.ID {
			r.categories[agentType.Category] = append(types[:i], types[i+1:]...)
			break
		}
	}
}
