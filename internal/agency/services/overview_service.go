package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// OverviewService handles agency overview operations
type OverviewService struct {
	repo agency.Repository
}

// NewOverviewService creates a new overview service
func NewOverviewService(repo agency.Repository) *OverviewService {
	return &OverviewService{
		repo: repo,
	}
}

// GetAgencyOverview retrieves the overview for an agency
func (s *OverviewService) GetAgencyOverview(ctx context.Context, agencyID string) (*models.Overview, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get overview from repository
	overview, err := s.repo.GetOverview(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overview: %w", err)
	}

	return overview, nil
}

// UpdateAgencyOverview updates the introduction in the agency overview
func (s *OverviewService) UpdateAgencyOverview(ctx context.Context, agencyID string, introduction string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get current overview or create new one
	overview, err := s.repo.GetOverview(ctx, agencyID)
	if err != nil {
		// If overview doesn't exist, create a new one
		overview = &models.Overview{
			AgencyID:     agencyID,
			Introduction: introduction,
			UpdatedAt:    time.Now(),
		}
	} else {
		// Update existing overview
		if introduction != "" {
			overview.Introduction = introduction
		}
		overview.UpdatedAt = time.Now()
	}

	// Update overview in repository
	err = s.repo.UpdateOverview(ctx, overview)
	if err != nil {
		return fmt.Errorf("failed to update overview: %w", err)
	}

	return nil
}
