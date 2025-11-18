package services

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/sirupsen/logrus"
)

// CompositeService combines all sub-services and implements the agency.Service interface
type CompositeService struct {
	*AgencyService
	*SpecificationService
}

// New creates a new composite service with all sub-services
func New(repo agency.Repository, validator agency.Validator, logger *logrus.Logger) agency.Service {
	return &CompositeService{
		AgencyService:        NewAgencyService(repo, validator, nil),
		SpecificationService: NewSpecificationService(repo, logger),
	}
}

// NewWithDBInit creates a new composite service with database initialization support
func NewWithDBInit(repo agency.Repository, validator agency.Validator, dbInit agency.DatabaseInitializer, logger *logrus.Logger) agency.Service {
	return &CompositeService{
		AgencyService:        NewAgencyService(repo, validator, dbInit),
		SpecificationService: NewSpecificationService(repo, logger),
	}
}

// Ensure CompositeService implements agency.Service
var _ agency.Service = (*CompositeService)(nil)

// Forwarding methods to maintain the interface

func (c *CompositeService) CreateAgency(ctx context.Context, agencyDoc *models.Agency) error {
	return c.AgencyService.CreateAgency(ctx, agencyDoc)
}

func (c *CompositeService) GetAgency(ctx context.Context, id string) (*models.Agency, error) {
	return c.AgencyService.GetAgency(ctx, id)
}

func (c *CompositeService) ListAgencies(ctx context.Context, filters models.AgencyFilters) ([]*models.Agency, error) {
	return c.AgencyService.ListAgencies(ctx, filters)
}

func (c *CompositeService) UpdateAgency(ctx context.Context, id string, updates models.AgencyUpdates) error {
	return c.AgencyService.UpdateAgency(ctx, id, updates)
}

func (c *CompositeService) DeleteAgency(ctx context.Context, id string) error {
	return c.AgencyService.DeleteAgency(ctx, id)
}

func (c *CompositeService) SetActiveAgency(ctx context.Context, id string) error {
	return c.AgencyService.SetActiveAgency(ctx, id)
}

func (c *CompositeService) GetActiveAgency(ctx context.Context) (*models.Agency, error) {
	return c.AgencyService.GetActiveAgency(ctx)
}

func (c *CompositeService) GetAgencyStatistics(ctx context.Context, id string) (*models.AgencyStatistics, error) {
	return c.AgencyService.GetAgencyStatistics(ctx, id)
}

// Specification forwarding methods

func (c *CompositeService) GetSpecification(ctx context.Context, agencyID string) (*models.AgencySpecification, error) {
	return c.SpecificationService.GetSpecification(ctx, agencyID)
}

func (c *CompositeService) UpdateSpecification(ctx context.Context, agencyID string, req *models.SpecificationUpdateRequest) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateSpecification(ctx, agencyID, req)
}

func (c *CompositeService) UpdateIntroduction(ctx context.Context, agencyID, introduction, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateIntroduction(ctx, agencyID, introduction, updatedBy)
}

func (c *CompositeService) UpdateSpecificationGoals(ctx context.Context, agencyID string, goals []models.Goal, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateGoals(ctx, agencyID, goals, updatedBy)
}

func (c *CompositeService) UpdateSpecificationWorkItems(ctx context.Context, agencyID string, workItems []models.WorkItem, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateWorkItems(ctx, agencyID, workItems, updatedBy)
}

func (c *CompositeService) UpdateSpecificationWorkflows(ctx context.Context, agencyID string, workflows []models.Workflow, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateWorkflows(ctx, agencyID, workflows, updatedBy)
}

func (c *CompositeService) UpdateSpecificationRoles(ctx context.Context, agencyID string, roles []models.Role, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateRoles(ctx, agencyID, roles, updatedBy)
}

func (c *CompositeService) UpdateSpecificationRACIMatrix(ctx context.Context, agencyID string, matrix *models.RACIMatrix, updatedBy string) (*models.AgencySpecification, error) {
	return c.SpecificationService.UpdateRACIMatrix(ctx, agencyID, matrix, updatedBy)
}
