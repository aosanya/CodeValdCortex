package services

import (
	"context"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency"
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
func (s *WorkItemService) CreateWorkItem(ctx context.Context, agencyID string, req agency.CreateWorkItemRequest) (*agency.WorkItem, error) {
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

	workItem := &agency.WorkItem{
		AgencyID:        agencyID,
		Title:           req.Title,
		Description:     req.Description,
		Type:            req.Type,
		Priority:        req.Priority,
		Status:          req.Status,
		Deliverables:    req.Deliverables,
		Dependencies:    req.Dependencies,
		EstimatedEffort: req.EstimatedEffort,
		AssignedTo:      req.AssignedTo,
		Tags:            req.Tags,
	}

	// Set default status if not provided
	if workItem.Status == "" {
		workItem.Status = agency.WorkItemStatusNotStarted
	}

	if err := s.repo.CreateWorkItem(ctx, workItem); err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}

	return workItem, nil
}

// GetWorkItems retrieves all work items for an agency
func (s *WorkItemService) GetWorkItems(ctx context.Context, agencyID string) ([]*agency.WorkItem, error) {
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
func (s *WorkItemService) GetWorkItem(ctx context.Context, agencyID string, key string) (*agency.WorkItem, error) {
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
func (s *WorkItemService) GetWorkItemByCode(ctx context.Context, agencyID string, code string) (*agency.WorkItem, error) {
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
func (s *WorkItemService) UpdateWorkItem(ctx context.Context, agencyID string, key string, req agency.UpdateWorkItemRequest) error {
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
	workItem.Type = req.Type
	workItem.Priority = req.Priority
	workItem.Status = req.Status
	workItem.Deliverables = req.Deliverables
	workItem.Dependencies = req.Dependencies
	workItem.EstimatedEffort = req.EstimatedEffort
	workItem.AssignedTo = req.AssignedTo
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

// GetWorkItemTemplates returns predefined work item templates
func (s *WorkItemService) GetWorkItemTemplates() map[string]WorkItemTemplate {
	return map[string]WorkItemTemplate{
		string(agency.WorkItemTypeTask): {
			Type:        agency.WorkItemTypeTask,
			Description: "A single unit of work to be completed",
			Deliverables: []string{
				"Implementation complete",
				"Unit tests passing",
				"Code reviewed",
			},
		},
		string(agency.WorkItemTypeFeature): {
			Type:        agency.WorkItemTypeFeature,
			Description: "A new feature or capability to be developed",
			Deliverables: []string{
				"Feature specification document",
				"Implementation complete",
				"Integration tests passing",
				"User documentation updated",
				"Feature deployed",
			},
		},
		string(agency.WorkItemTypeEpic): {
			Type:        agency.WorkItemTypeEpic,
			Description: "A large body of work that can be broken down into smaller tasks",
			Deliverables: []string{
				"Epic broken down into tasks/features",
				"Architecture design document",
				"All child work items completed",
				"Integration verified",
				"Release notes prepared",
			},
		},
		string(agency.WorkItemTypeBug): {
			Type:        agency.WorkItemTypeBug,
			Description: "A defect or issue that needs to be fixed",
			Deliverables: []string{
				"Root cause identified",
				"Fix implemented",
				"Regression tests added",
				"Fix verified in production",
			},
		},
		string(agency.WorkItemTypeResearch): {
			Type:        agency.WorkItemTypeResearch,
			Description: "Investigation or proof of concept work",
			Deliverables: []string{
				"Research findings documented",
				"Recommendations provided",
				"Proof of concept (if applicable)",
				"Decision recorded",
			},
		},
	}
}

// WorkItemTemplate represents a template for creating work items
type WorkItemTemplate struct {
	Type         agency.WorkItemType
	Description  string
	Deliverables []string
}
