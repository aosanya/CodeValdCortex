package services

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/sirupsen/logrus"
)

// SpecificationService handles agency specification operations
type SpecificationService struct {
	repo   agency.Repository
	logger *logrus.Logger
}

// NewSpecificationService creates a new specification service
func NewSpecificationService(repo agency.Repository, logger *logrus.Logger) *SpecificationService {
	return &SpecificationService{
		repo:   repo,
		logger: logger,
	}
}

// GetSpecification retrieves the complete specification for an agency
func (s *SpecificationService) GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error) {
	s.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"method":    "SpecificationService.GetSpecification",
	}).Info("Calling repository GetSpecification")

	spec, err := s.repo.GetSpecification(ctx, agencyID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
			"method":    "SpecificationService.GetSpecification",
		}).Error("Repository GetSpecification failed")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"method":    "SpecificationService.GetSpecification",
	}).Info("Repository GetSpecification completed successfully")

	return spec, nil
}

// CreateSpecification creates a new specification for an agency
func (s *SpecificationService) CreateSpecification(ctx context.Context, agencyID string, req *models.CreateSpecificationRequest) (*models.AgencySpecification, error) {
	return s.repo.CreateSpecification(ctx, agencyID, req)
}

// UpdateSpecification updates the entire specification
func (s *SpecificationService) UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error) {
	return s.repo.UpdateSpecification(ctx, agencyID, req)
}

// UpdateIntroduction updates only the introduction section
func (s *SpecificationService) UpdateIntroduction(ctx context.Context, agencyID, introduction, updatedBy string) (*models.AgencySpecification, error) {
	return s.repo.PatchSpecificationSection(ctx, agencyID, "introduction", introduction, updatedBy)
}

// UpdateGoals updates only the goals section
func (s *SpecificationService) UpdateGoals(ctx context.Context, agencyID string, goals []models.Goal, updatedBy string) (*models.AgencySpecification, error) {
	return s.repo.PatchSpecificationSection(ctx, agencyID, "goals", goals, updatedBy)
}

// UpdateWorkItems updates only the work items section
func (s *SpecificationService) UpdateWorkItems(ctx context.Context, agencyID string, workItems []models.WorkItem, updatedBy string) (*models.AgencySpecification, error) {
	return s.repo.PatchSpecificationSection(ctx, agencyID, "work_items", workItems, updatedBy)
}

// UpdateRoles updates only the roles section
func (s *SpecificationService) UpdateRoles(ctx context.Context, agencyID string, roles []models.Role, updatedBy string) (*models.AgencySpecification, error) {
	return s.repo.PatchSpecificationSection(ctx, agencyID, "roles", roles, updatedBy)
}

// UpdateRACIMatrix updates only the RACI matrix section
func (s *SpecificationService) UpdateRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix, updatedBy string) (*models.AgencySpecification, error) {
	return s.repo.PatchSpecificationSection(ctx, agencyID, "raci_matrix", matrix, updatedBy)
}

// DeleteSpecification deletes the specification for an agency
func (s *SpecificationService) DeleteSpecification(ctx context.Context, agencyID string) error {
	return s.repo.DeleteSpecification(ctx, agencyID)
}

// InitializeSpecificationWithDefaults creates a specification with standard roles
func (s *SpecificationService) InitializeSpecificationWithDefaults(ctx context.Context, agencyID string) (*models.AgencySpecification, error) {
	req := &models.CreateSpecificationRequest{
		Introduction: "",
		Goals:        []models.Goal{},
		WorkItems:    []models.WorkItem{},
		Roles:        models.StandardAgencyRoles, // Use predefined standard roles
		RACIMatrix:   nil,
	}
	return s.repo.CreateSpecification(ctx, agencyID, req)
}
