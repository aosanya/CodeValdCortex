package services

import (
	"context"

	"github.com/aosanya/CodeValdCortex/internal/agency"
)

// CompositeService combines all sub-services and implements the agency.Service interface
type CompositeService struct {
	*AgencyService
	*OverviewService
	*GoalService
	*WorkItemService
	*RACIService
}

// New creates a new composite service with all sub-services
func New(repo agency.Repository, validator agency.Validator) agency.Service {
	return &CompositeService{
		AgencyService:   NewAgencyService(repo, validator, nil),
		OverviewService: NewOverviewService(repo),
		GoalService:     NewGoalService(repo),
		WorkItemService: NewWorkItemService(repo),
		RACIService:     NewRACIService(repo),
	}
}

// NewWithDBInit creates a new composite service with database initialization support
func NewWithDBInit(repo agency.Repository, validator agency.Validator, dbInit agency.DatabaseInitializer) agency.Service {
	return &CompositeService{
		AgencyService:   NewAgencyService(repo, validator, dbInit),
		OverviewService: NewOverviewService(repo),
		GoalService:     NewGoalService(repo),
		WorkItemService: NewWorkItemService(repo),
		RACIService:     NewRACIService(repo),
	}
}

// Ensure CompositeService implements agency.Service
var _ agency.Service = (*CompositeService)(nil)

// Forwarding methods to maintain the interface

func (c *CompositeService) CreateAgency(ctx context.Context, agencyDoc *agency.Agency) error {
	return c.AgencyService.CreateAgency(ctx, agencyDoc)
}

func (c *CompositeService) GetAgency(ctx context.Context, id string) (*agency.Agency, error) {
	return c.AgencyService.GetAgency(ctx, id)
}

func (c *CompositeService) ListAgencies(ctx context.Context, filters agency.AgencyFilters) ([]*agency.Agency, error) {
	return c.AgencyService.ListAgencies(ctx, filters)
}

func (c *CompositeService) UpdateAgency(ctx context.Context, id string, updates agency.AgencyUpdates) error {
	return c.AgencyService.UpdateAgency(ctx, id, updates)
}

func (c *CompositeService) DeleteAgency(ctx context.Context, id string) error {
	return c.AgencyService.DeleteAgency(ctx, id)
}

func (c *CompositeService) SetActiveAgency(ctx context.Context, id string) error {
	return c.AgencyService.SetActiveAgency(ctx, id)
}

func (c *CompositeService) GetActiveAgency(ctx context.Context) (*agency.Agency, error) {
	return c.AgencyService.GetActiveAgency(ctx)
}

func (c *CompositeService) GetAgencyStatistics(ctx context.Context, id string) (*agency.AgencyStatistics, error) {
	return c.AgencyService.GetAgencyStatistics(ctx, id)
}

func (c *CompositeService) GetAgencyOverview(ctx context.Context, agencyID string) (*agency.Overview, error) {
	return c.OverviewService.GetAgencyOverview(ctx, agencyID)
}

func (c *CompositeService) UpdateAgencyOverview(ctx context.Context, agencyID string, introduction string) error {
	return c.OverviewService.UpdateAgencyOverview(ctx, agencyID, introduction)
}

func (c *CompositeService) CreateGoal(ctx context.Context, agencyID string, code string, description string) (*agency.Goal, error) {
	return c.GoalService.CreateGoal(ctx, agencyID, code, description)
}

func (c *CompositeService) GetGoals(ctx context.Context, agencyID string) ([]*agency.Goal, error) {
	return c.GoalService.GetGoals(ctx, agencyID)
}

func (c *CompositeService) GetGoal(ctx context.Context, agencyID string, key string) (*agency.Goal, error) {
	return c.GoalService.GetGoal(ctx, agencyID, key)
}

func (c *CompositeService) UpdateGoal(ctx context.Context, agencyID string, key string, code string, description string) error {
	return c.GoalService.UpdateGoal(ctx, agencyID, key, code, description)
}

func (c *CompositeService) DeleteGoal(ctx context.Context, agencyID string, key string) error {
	return c.GoalService.DeleteGoal(ctx, agencyID, key)
}

// WorkItem forwarding methods

func (c *CompositeService) CreateWorkItem(ctx context.Context, agencyID string, req agency.CreateWorkItemRequest) (*agency.WorkItem, error) {
	return c.WorkItemService.CreateWorkItem(ctx, agencyID, req)
}

func (c *CompositeService) GetWorkItems(ctx context.Context, agencyID string) ([]*agency.WorkItem, error) {
	return c.WorkItemService.GetWorkItems(ctx, agencyID)
}

func (c *CompositeService) GetWorkItem(ctx context.Context, agencyID string, key string) (*agency.WorkItem, error) {
	return c.WorkItemService.GetWorkItem(ctx, agencyID, key)
}

func (c *CompositeService) GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*agency.WorkItem, error) {
	return c.WorkItemService.GetWorkItemByCode(ctx, agencyID, code)
}

func (c *CompositeService) UpdateWorkItem(ctx context.Context, agencyID string, key string, req agency.UpdateWorkItemRequest) error {
	return c.WorkItemService.UpdateWorkItem(ctx, agencyID, key, req)
}

func (c *CompositeService) DeleteWorkItem(ctx context.Context, agencyID string, key string) error {
	return c.WorkItemService.DeleteWorkItem(ctx, agencyID, key)
}

func (c *CompositeService) ValidateWorkItemDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error {
	return c.WorkItemService.ValidateDependencies(ctx, agencyID, workItemCode, dependencies)
}
