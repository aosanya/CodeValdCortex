package services

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// UnitOfWorkService handles unit of work operations
type UnitOfWorkService struct {
	repo agency.Repository
}

// NewUnitOfWorkService creates a new unit of work service
func NewUnitOfWorkService(repo agency.Repository) *UnitOfWorkService {
	return &UnitOfWorkService{
		repo: repo,
	}
}

// CreateUnitOfWork creates a new unit of work for an agency
func (s *UnitOfWorkService) CreateUnitOfWork(ctx context.Context, agencyID string, code string, description string) (*agency.UnitOfWork, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	unit := &agency.UnitOfWork{
		AgencyID:    agencyID,
		Code:        code,
		Description: description,
	}

	if err := s.repo.CreateUnitOfWork(ctx, unit); err != nil {
		return nil, fmt.Errorf("failed to create unit of work: %w", err)
	}

	return unit, nil
}

// GetUnitsOfWork retrieves all units of work for an agency
func (s *UnitOfWorkService) GetUnitsOfWork(ctx context.Context, agencyID string) ([]*agency.UnitOfWork, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	units, err := s.repo.GetUnitsOfWork(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units of work: %w", err)
	}

	return units, nil
}

// UpdateUnitOfWork updates a unit of work's code and description
func (s *UnitOfWorkService) UpdateUnitOfWork(ctx context.Context, agencyID string, key string, code string, description string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get the unit of work
	unit, err := s.repo.GetUnitOfWork(ctx, agencyID, key)
	if err != nil {
		return fmt.Errorf("failed to get unit of work: %w", err)
	}

	// Update code and description
	unit.Code = code
	unit.Description = description

	// Save
	if err := s.repo.UpdateUnitOfWork(ctx, unit); err != nil {
		return fmt.Errorf("failed to update unit of work: %w", err)
	}

	return nil
}

// DeleteUnitOfWork deletes a unit of work
func (s *UnitOfWorkService) DeleteUnitOfWork(ctx context.Context, agencyID string, key string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	if err := s.repo.DeleteUnitOfWork(ctx, agencyID, key); err != nil {
		return fmt.Errorf("failed to delete unit of work: %w", err)
	}

	return nil
}
