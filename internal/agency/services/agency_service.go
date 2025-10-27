package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// AgencyService handles core agency CRUD operations
type AgencyService struct {
	repo      agency.Repository
	validator agency.Validator
	dbInit    agency.DatabaseInitializer
	active    string // Currently active agency ID
}

// NewAgencyService creates a new agency service
func NewAgencyService(repo agency.Repository, validator agency.Validator, dbInit agency.DatabaseInitializer) *AgencyService {
	return &AgencyService{
		repo:      repo,
		validator: validator,
		dbInit:    dbInit,
	}
}

// CreateAgency creates a new agency
func (s *AgencyService) CreateAgency(ctx context.Context, agencyDoc *agency.Agency) error {
	// Validate agency configuration
	if err := s.validator.ValidateAgency(agencyDoc); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set timestamps
	now := time.Now()
	agencyDoc.CreatedAt = now
	agencyDoc.UpdatedAt = now

	// Set default status if not provided
	if agencyDoc.Status == "" {
		agencyDoc.Status = agency.AgencyStatusActive
	}

	// Set database field if not provided
	// Database name uses the agency ID directly (which already has "agency_" prefix)
	if agencyDoc.Database == "" {
		agencyDoc.Database = agencyDoc.ID
	}

	// Initialize agency database if database initializer is available
	if s.dbInit != nil {
		if err := s.dbInit.InitializeAgencyDatabase(ctx, agencyDoc.ID); err != nil {
			return fmt.Errorf("failed to initialize agency database: %w", err)
		}
	}

	// Create in repository
	if err := s.repo.Create(ctx, agencyDoc); err != nil {
		return fmt.Errorf("failed to create agency: %w", err)
	}

	return nil
}

// GetAgency retrieves an agency by ID
func (s *AgencyService) GetAgency(ctx context.Context, id string) (*agency.Agency, error) {
	agencyDoc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency: %w", err)
	}
	return agencyDoc, nil
}

// ListAgencies retrieves agencies with optional filtering
func (s *AgencyService) ListAgencies(ctx context.Context, filters agency.AgencyFilters) ([]*agency.Agency, error) {
	agencies, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list agencies: %w", err)
	}
	return agencies, nil
}

// UpdateAgency updates an existing agency
func (s *AgencyService) UpdateAgency(ctx context.Context, id string, updates agency.AgencyUpdates) error {
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
func (s *AgencyService) DeleteAgency(ctx context.Context, id string) error {
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
func (s *AgencyService) SetActiveAgency(ctx context.Context, id string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("agency not found: %w", err)
	}

	s.active = id
	return nil
}

// GetActiveAgency returns the currently active agency
func (s *AgencyService) GetActiveAgency(ctx context.Context) (*agency.Agency, error) {
	if s.active == "" {
		return nil, fmt.Errorf("no active agency set")
	}

	return s.repo.GetByID(ctx, s.active)
}

// GetAgencyStatistics retrieves operational statistics for an agency
func (s *AgencyService) GetAgencyStatistics(ctx context.Context, id string) (*agency.AgencyStatistics, error) {
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
func (s *AgencyService) applyUpdates(agencyDoc *agency.Agency, updates agency.AgencyUpdates) {
	if updates.DisplayName != nil {
		agencyDoc.DisplayName = *updates.DisplayName
	}
	if updates.Description != nil {
		agencyDoc.Description = *updates.Description
	}
	if updates.Category != nil {
		agencyDoc.Category = *updates.Category
	}
	if updates.Icon != nil {
		agencyDoc.Icon = *updates.Icon
	}
	if updates.Status != nil {
		agencyDoc.Status = *updates.Status
	}
	if updates.Metadata != nil {
		agencyDoc.Metadata = *updates.Metadata
	}
	if updates.Settings != nil {
		agencyDoc.Settings = *updates.Settings
	}
}
