package services

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// WorkItemService handles work item operations
type WorkItemService struct {
	repo agency.Repository
}

// NewWorkItemService creates a new work item service
func NewWorkItemService(repo agency.Repository) *WorkItemService {
	return &WorkItemService{
		repo: repo,
	}
}

// CreateWorkItem creates a new work item for an agency
func (s *WorkItemService) CreateWorkItem(ctx context.Context, agencyID string, req models.CreateWorkItemRequest) (*models.WorkItem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	// Validate dependencies if provided
	if len(req.Dependencies) > 0 {
		if err := s.repo.ValidateDependencies(ctx, agencyID, "", req.Dependencies); err != nil {
			return nil, fmt.Errorf("invalid dependencies: %w", err)
		}
	}

	workItem := &models.WorkItem{
		AgencyID:     agencyID,
		Title:        req.Title,
		Description:  req.Description,
		Deliverables: req.Deliverables,
		Dependencies: req.Dependencies,
		Tags:         req.Tags,
	}

	if err := s.repo.CreateWorkItem(ctx, workItem); err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}

	return workItem, nil
}

// GetWorkItems retrieves all work items for an agency
func (s *WorkItemService) GetWorkItems(ctx context.Context, agencyID string) ([]*models.WorkItem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	workItems, err := s.repo.GetWorkItems(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work items: %w", err)
	}

	return workItems, nil
}

// GetWorkItem retrieves a single work item by key
func (s *WorkItemService) GetWorkItem(ctx context.Context, agencyID string, key string) (*models.WorkItem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	workItem, err := s.repo.GetWorkItem(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get work item: %w", err)
	}

	return workItem, nil
}

// GetWorkItemByCode retrieves a single work item by code
func (s *WorkItemService) GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*models.WorkItem, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	workItem, err := s.repo.GetWorkItemByCode(ctx, agencyID, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get work item: %w", err)
	}

	return workItem, nil
}

// UpdateWorkItem updates a work item
func (s *WorkItemService) UpdateWorkItem(ctx context.Context, agencyID string, key string, req models.UpdateWorkItemRequest) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	// Get the work item
	workItem, err := s.repo.GetWorkItem(ctx, agencyID, key)
	if err != nil {
		return fmt.Errorf("failed to get work item: %w", err)
	}

	// Validate dependencies if changed
	if len(req.Dependencies) > 0 {
		if err := s.repo.ValidateDependencies(ctx, agencyID, workItem.Code, req.Dependencies); err != nil {
			return fmt.Errorf("invalid dependencies: %w", err)
		}
	}

	// Update fields
	workItem.Title = req.Title
	workItem.Description = req.Description
	workItem.Deliverables = req.Deliverables
	workItem.Dependencies = req.Dependencies
	workItem.Tags = req.Tags

	// Save
	if err := s.repo.UpdateWorkItem(ctx, workItem); err != nil {
		return fmt.Errorf("failed to update work item: %w", err)
	}

	return nil
}

// DeleteWorkItem deletes a work item
func (s *WorkItemService) DeleteWorkItem(ctx context.Context, agencyID string, key string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	if err := s.repo.DeleteWorkItem(ctx, agencyID, key); err != nil {
		return fmt.Errorf("failed to delete work item: %w", err)
	}

	return nil
}

// ValidateDependencies validates work item dependencies
func (s *WorkItemService) ValidateDependencies(ctx context.Context, agencyID string, workItemCode string, dependencies []string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.ValidateDependencies(ctx, agencyID, workItemCode, dependencies)
}

// CreateWorkItemGoalLink creates a new link between a work item and a goal
func (s *WorkItemService) CreateWorkItemGoalLink(ctx context.Context, agencyID string, link *models.WorkItemGoalLink) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.CreateWorkItemGoalLink(ctx, agencyID, link)
}

// GetWorkItemGoalLinks retrieves all goal links for a work item
func (s *WorkItemService) GetWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) ([]*models.WorkItemGoalLink, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.GetWorkItemGoalLinks(ctx, agencyID, workItemKey)
}

// GetGoalWorkItems retrieves all work items linked to a goal
func (s *WorkItemService) GetGoalWorkItems(ctx context.Context, agencyID, goalKey string) ([]*models.WorkItemGoalLink, error) {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.GetGoalWorkItems(ctx, agencyID, goalKey)
}

// DeleteWorkItemGoalLink deletes a specific work item-goal link
func (s *WorkItemService) DeleteWorkItemGoalLink(ctx context.Context, agencyID, linkKey string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.DeleteWorkItemGoalLink(ctx, agencyID, linkKey)
}

// DeleteWorkItemGoalLinks deletes all goal links for a work item
func (s *WorkItemService) DeleteWorkItemGoalLinks(ctx context.Context, agencyID, workItemKey string) error {
	// Verify agency exists
	_, err := s.repo.GetByID(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to verify agency: %w", err)
	}

	return s.repo.DeleteWorkItemGoalLinks(ctx, agencyID, workItemKey)
}
