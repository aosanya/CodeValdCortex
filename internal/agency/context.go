package agency

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// AgencyContextKey is the key for storing agency in context
	AgencyContextKey ContextKey = "agency"
	// AgencyIDContextKey is the key for storing agency ID in context
	AgencyIDContextKey ContextKey = "agency_id"
)

// ContextManager manages agency context for requests
type ContextManager struct {
	service Service
}

// NewContextManager creates a new context manager
func NewContextManager(service Service) *ContextManager {
	return &ContextManager{
		service: service,
	}
}

// WithAgency adds an agency to the context
func (cm *ContextManager) WithAgency(ctx context.Context, agencyID string) (context.Context, error) {
	agency, err := cm.service.GetAgency(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}

	ctx = context.WithValue(ctx, AgencyContextKey, agency)
	ctx = context.WithValue(ctx, AgencyIDContextKey, agencyID)

	return ctx, nil
}

// GetAgencyFromContext retrieves the agency from context
func GetAgencyFromContext(ctx context.Context) (*models.Agency, error) {
	agency, ok := ctx.Value(AgencyContextKey).(*models.Agency)
	if !ok || agency == nil {
		return nil, fmt.Errorf("no agency in context")
	}
	return agency, nil
}

// GetAgencyIDFromContext retrieves the agency ID from context
func GetAgencyIDFromContext(ctx context.Context) (string, error) {
	agencyID, ok := ctx.Value(AgencyIDContextKey).(string)
	if !ok || agencyID == "" {
		return "", fmt.Errorf("no agency ID in context")
	}
	return agencyID, nil
}

// HasAgencyContext checks if there's an agency in context
func HasAgencyContext(ctx context.Context) bool {
	_, err := GetAgencyFromContext(ctx)
	return err == nil
}
