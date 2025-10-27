package agency

import (
	"context"
	"fmt"
	"time"
)

// Service defines the interface for agency management operations
type Service interface {
	CreateAgency(ctx context.Context, agency *Agency) error
	GetAgency(ctx context.Context, id string) (*Agency, error)
	ListAgencies(ctx context.Context, filters AgencyFilters) ([]*Agency, error)
	UpdateAgency(ctx context.Context, id string, updates AgencyUpdates) error
	DeleteAgency(ctx context.Context, id string) error
	SetActiveAgency(ctx context.Context, id string) error
	GetActiveAgency(ctx context.Context) (*Agency, error)
	GetAgencyStatistics(ctx context.Context, id string) (*AgencyStatistics, error)
}

// service implements the Service interface
type service struct {
	repo      Repository
	validator Validator
	dbInit    DatabaseInitializer
	active    string // Currently active agency ID
}

// NewService creates a new agency service
func NewService(repo Repository, validator Validator) Service {
	return &service{
		repo:      repo,
		validator: validator,
	}
}

// NewServiceWithDBInit creates a new agency service with database initialization support
func NewServiceWithDBInit(repo Repository, validator Validator, dbInit DatabaseInitializer) Service {
	return &service{
		repo:      repo,
		validator: validator,
		dbInit:    dbInit,
	}
}

// CreateAgency creates a new agency
func (s *service) CreateAgency(ctx context.Context, agency *Agency) error {
	// Validate agency configuration
	if err := s.validator.ValidateAgency(agency); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set timestamps
	now := time.Now()
	agency.CreatedAt = now
	agency.UpdatedAt = now

	// Set default status if not provided
	if agency.Status == "" {
		agency.Status = AgencyStatusActive
	}

	// Set database field if not provided
	// Database name uses the agency ID directly (which already has "agency_" prefix)
	if agency.Database == "" {
		agency.Database = agency.ID
	}

	// Initialize agency database if database initializer is available
	if s.dbInit != nil {
		if err := s.dbInit.InitializeAgencyDatabase(ctx, agency.ID); err != nil {
			return fmt.Errorf("failed to initialize agency database: %w", err)
		}
	}

	// Create in repository
	if err := s.repo.Create(ctx, agency); err != nil {
		return fmt.Errorf("failed to create agency: %w", err)
	}

	return nil
}

// GetAgency retrieves an agency by ID
func (s *service) GetAgency(ctx context.Context, id string) (*Agency, error) {
	agency, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}
	return agency, nil
}

// ListAgencies retrieves agencies with optional filtering
func (s *service) ListAgencies(ctx context.Context, filters AgencyFilters) ([]*Agency, error) {
	agencies, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list agencies: %w", err)
	}
	return agencies, nil
}

// UpdateAgency updates an existing agency
func (s *service) UpdateAgency(ctx context.Context, id string, updates AgencyUpdates) error {
	// Get existing agency
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("agency not found: %w", err)
	}

	// Apply updates
	s.applyUpdates(existing, updates)
	existing.UpdatedAt = time.Now()

	// Validate updated agency
	if err := s.validator.ValidateAgency(existing); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update agency: %w", err)
	}

	return nil
}

// DeleteAgency deletes an agency
func (s *service) DeleteAgency(ctx context.Context, id string) error {
	// Check if it's the active agency
	if s.active == id {
		return fmt.Errorf("cannot delete active agency")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete agency: %w", err)
	}

	return nil
}

// SetActiveAgency sets the currently active agency
func (s *service) SetActiveAgency(ctx context.Context, id string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("agency not found: %w", err)
	}

	s.active = id
	return nil
}

// GetActiveAgency returns the currently active agency
func (s *service) GetActiveAgency(ctx context.Context) (*Agency, error) {
	if s.active == "" {
		return nil, fmt.Errorf("no active agency set")
	}

	return s.repo.GetByID(ctx, s.active)
}

// GetAgencyStatistics retrieves operational statistics for an agency
func (s *service) GetAgencyStatistics(ctx context.Context, id string) (*AgencyStatistics, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agency not found: %w", err)
	}

	stats, err := s.repo.GetStatistics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// applyUpdates applies the updates to the agency
func (s *service) applyUpdates(agency *Agency, updates AgencyUpdates) {
	if updates.DisplayName != nil {
		agency.DisplayName = *updates.DisplayName
	}
	if updates.Description != nil {
		agency.Description = *updates.Description
	}
	if updates.Category != nil {
		agency.Category = *updates.Category
	}
	if updates.Icon != nil {
		agency.Icon = *updates.Icon
	}
	if updates.Status != nil {
		agency.Status = *updates.Status
	}
	if updates.Metadata != nil {
		agency.Metadata = *updates.Metadata
	}
	if updates.Settings != nil {
		agency.Settings = *updates.Settings
	}
}
