package agency

import (
	"context"
)

// Repository defines the interface for agency data persistence
type Repository interface {
	Create(ctx context.Context, agency *Agency) error
	GetByID(ctx context.Context, id string) (*Agency, error)
	List(ctx context.Context, filters AgencyFilters) ([]*Agency, error)
	Update(ctx context.Context, agency *Agency) error
	Delete(ctx context.Context, id string) error
	GetStatistics(ctx context.Context, id string) (*AgencyStatistics, error)
	Exists(ctx context.Context, id string) (bool, error)
}
