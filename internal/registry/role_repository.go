package registry

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryRoleRepository implements RoleRepository using in-memory storage
type InMemoryRoleRepository struct {
	mu    sync.RWMutex
	types map[string]*Role
}

// NewInMemoryRoleRepository creates a new in-memory repository
func NewInMemoryRoleRepository() *InMemoryRoleRepository {
	return &InMemoryRoleRepository{
		types: make(map[string]*Role),
	}
}

// Create registers a new role
func (r *InMemoryRoleRepository) Create(ctx context.Context, agentType *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists
	if _, exists := r.types[agentType.ID]; exists {
		return fmt.Errorf("role %s already exists", agentType.ID)
	}

	// Set timestamps
	now := time.Now()
	agentType.CreatedAt = now
	agentType.UpdatedAt = now

	// Store type
	r.types[agentType.ID] = agentType

	return nil
}

// Get retrieves an role by ID
func (r *InMemoryRoleRepository) Get(ctx context.Context, id string) (*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentType, exists := r.types[id]
	if !exists {
		return nil, fmt.Errorf("role %s not found", id)
	}

	return agentType, nil
}

// Update modifies an existing role
func (r *InMemoryRoleRepository) Update(ctx context.Context, agentType *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.types[agentType.ID]
	if !exists {
		return fmt.Errorf("role %s not found", agentType.ID)
	}

	// Prevent modification of system types' core fields
	if existing.IsSystemType && agentType.IsSystemType != existing.IsSystemType {
		return fmt.Errorf("cannot change system type flag for %s", agentType.ID)
	}

	// Preserve creation metadata
	agentType.CreatedAt = existing.CreatedAt
	agentType.CreatedBy = existing.CreatedBy
	agentType.UpdatedAt = time.Now()

	// Store updated type
	r.types[agentType.ID] = agentType

	return nil
}

// Delete removes an role
func (r *InMemoryRoleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agentType, exists := r.types[id]
	if !exists {
		return fmt.Errorf("role %s not found", id)
	}

	// Prevent deletion of system types
	if agentType.IsSystemType {
		return fmt.Errorf("cannot delete system type %s", id)
	}

	// Delete type
	delete(r.types, id)

	return nil
}

// List returns all registered roles
func (r *InMemoryRoleRepository) List(ctx context.Context) ([]*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]*Role, 0, len(r.types))
	for _, agentType := range r.types {
		types = append(types, agentType)
	}

	return types, nil
}

// ListByTags returns roles that have any of the specified tags
func (r *InMemoryRoleRepository) ListByTags(ctx context.Context, tags []string) ([]*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(tags) == 0 {
		return []*Role{}, nil
	}

	result := make([]*Role, 0)
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	for _, agentType := range r.types {
		// Check if role has any of the requested tags
		for _, roleTag := range agentType.Tags {
			if tagSet[roleTag] {
				result = append(result, agentType)
				break
			}
		}
	}

	return result, nil
}

// ListEnabled returns all enabled roles
func (r *InMemoryRoleRepository) ListEnabled(ctx context.Context) ([]*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]*Role, 0)
	for _, agentType := range r.types {
		if agentType.IsEnabled {
			types = append(types, agentType)
		}
	}

	return types, nil
}

// Exists checks if an role is registered
func (r *InMemoryRoleRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.types[id]
	return exists, nil
}

// Count returns the total number of roles
func (r *InMemoryRoleRepository) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.types)), nil
}
