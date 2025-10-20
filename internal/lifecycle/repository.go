package lifecycle

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// Repository defines the interface for agent persistence
type Repository interface {
	Create(ctx context.Context, a *agent.Agent) error
	Get(ctx context.Context, id string) (*agent.Agent, error)
	Update(ctx context.Context, a *agent.Agent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*agent.Agent, error)
	Count(ctx context.Context) (int, error)
	FindByType(ctx context.Context, agentType string) ([]*agent.Agent, error)
	FindByState(ctx context.Context, state string) ([]*agent.Agent, error)
	FindHealthy(ctx context.Context) ([]*agent.Agent, error)
	FindByTypeAndState(ctx context.Context, agentType, state string) ([]*agent.Agent, error)
}
