package lifecycle

import (
	"context"
	"fmt"
	"sync"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// InMemoryRepository is a simple in-memory implementation of Repository
// Used for MVP until proper ArangoDB implementation is created
type InMemoryRepository struct {
	mu     sync.RWMutex
	agents map[string]*agent.Agent
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		agents: make(map[string]*agent.Agent),
	}
}

func (r *InMemoryRepository) Create(_ context.Context, a *agent.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[a.ID]; exists {
		return fmt.Errorf("agent already exists: %s", a.ID)
	}

	r.agents[a.ID] = a
	return nil
}

func (r *InMemoryRepository) Get(_ context.Context, id string) (*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, exists := r.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return a, nil
}

func (r *InMemoryRepository) Update(_ context.Context, a *agent.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[a.ID]; !exists {
		return fmt.Errorf("agent not found: %s", a.ID)
	}

	r.agents[a.ID] = a
	return nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[id]; !exists {
		return fmt.Errorf("agent not found: %s", id)
	}

	delete(r.agents, id)
	return nil
}

func (r *InMemoryRepository) List(_ context.Context) ([]*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*agent.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}

	return result, nil
}

func (r *InMemoryRepository) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents), nil
}

func (r *InMemoryRepository) FindByType(_ context.Context, agentType string) ([]*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.Agent
	for _, a := range r.agents {
		if a.Type == agentType {
			result = append(result, a)
		}
	}

	return result, nil
}

func (r *InMemoryRepository) FindByState(_ context.Context, state string) ([]*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.Agent
	for _, a := range r.agents {
		if string(a.State) == state {
			result = append(result, a)
		}
	}

	return result, nil
}

func (r *InMemoryRepository) FindHealthy(_ context.Context) ([]*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.Agent
	for _, a := range r.agents {
		// Agent is healthy if it's in running state
		if a.State == agent.StateRunning {
			result = append(result, a)
		}
	}

	return result, nil
}

func (r *InMemoryRepository) FindByTypeAndState(_ context.Context, agentType, state string) ([]*agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*agent.Agent
	for _, a := range r.agents {
		if a.Type == agentType && string(a.State) == state {
			result = append(result, a)
		}
	}

	return result, nil
}
