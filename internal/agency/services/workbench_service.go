package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// WorkbenchService handles workbench board generation
type WorkbenchService struct {
	tagService TagService
	issueRepo  agency.IssueRepository
	specRepo   agency.Repository
}

// NewWorkbenchService creates a new workbench service
func NewWorkbenchService(tagService TagService, issueRepo agency.IssueRepository, specRepo agency.Repository) *WorkbenchService {
	return &WorkbenchService{
		tagService: tagService,
		issueRepo:  issueRepo,
		specRepo:   specRepo,
	}
}

// GenerateBoard creates a Kanban board from a tag snapshot and current issues
func (s *WorkbenchService) GenerateBoard(ctx context.Context, agencyID, instanceID, tagName string) (*models.WorkbenchBoard, error) {
	// Get tag snapshot
	tag, err := s.tagService.GetTag(ctx, agencyID, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Verify tag has a specification
	if tag.Snapshot.Specification.Key == "" {
		return nil, fmt.Errorf("tag has no specification snapshot")
	}

	spec := &tag.Snapshot.Specification

	// Verify tag has workflows
	if len(spec.Workflows) == 0 {
		return nil, fmt.Errorf("tag specification has no workflows")
	}

	// Use first workflow (in future, allow workflow selection)
	workflow := spec.Workflows[0]

	// Create columns from workflow steps
	var columns []models.BoardColumn
	columnOrder := 0

	for _, step := range workflow.Steps {
		for _, item := range step.Items {
			// Find work item details in specification
			var workItem *models.WorkItem
			for i := range spec.WorkItems {
				if spec.WorkItems[i].Code == item.WorkItemKey {
					workItem = &spec.WorkItems[i]
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemKey,
					Title: item.WorkItemName,
				}
			}

			// Query issues for this step
			issues, err := s.issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, item.WorkItemKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", item.WorkItemKey, err)
			}

			// Convert pointers to values for BoardColumn
			issueValues := make([]models.WorkIssue, len(issues))
			for i, issue := range issues {
				issueValues[i] = *issue
			}

			// Create column
			column := models.BoardColumn{
				ID:           fmt.Sprintf("col-%s", item.WorkItemKey),
				Name:         workItem.Title,
				WorkItemCode: item.WorkItemKey,
				WorkItemKey:  item.WorkItemKey,
				Order:        columnOrder,
				Issues:       issueValues,
			}

			columns = append(columns, column)
			columnOrder++
		}
	}

	// Create board
	board := &models.WorkbenchBoard{
		AgencyID:    agencyID,
		InstanceID:  instanceID,
		WorkflowID:  workflow.Key,
		TagKey:      tag.Key,
		Columns:     columns,
		GeneratedAt: time.Now(),
	}

	return board, nil
}

// GenerateBoardFromLatestTag generates board using the latest published tag
func (s *WorkbenchService) GenerateBoardFromLatestTag(ctx context.Context, agencyID, instanceID string) (*models.WorkbenchBoard, error) {
	// Get latest published tag
	filters := &TagFilters{
		Type:   models.TagTypeRelease,
		Limit:  1,
		Offset: 0,
	}

	tags, err := s.tagService.ListTags(ctx, agencyID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %w", err)
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("no published tags found")
	}

	latestTag := tags[0]

	// Generate board from this tag
	return s.GenerateBoard(ctx, agencyID, instanceID, latestTag.Name)
}

// GenerateBoardFromSpecification generates board directly from agency specification (no tag)
func (s *WorkbenchService) GenerateBoardFromSpecification(ctx context.Context, agencyID, instanceID string) (*models.WorkbenchBoard, error) {
	// Get agency specification
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	// Verify spec has workflows
	if len(spec.Workflows) == 0 {
		return nil, fmt.Errorf("agency specification has no workflows")
	}

	// Use first workflow
	workflow := spec.Workflows[0]

	// Create columns from workflow steps
	var columns []models.BoardColumn
	columnOrder := 0

	for _, step := range workflow.Steps {
		for _, item := range step.Items {
			// Find work item details in specification
			var workItem *models.WorkItem
			for i := range spec.WorkItems {
				if spec.WorkItems[i].Code == item.WorkItemKey {
					workItem = &spec.WorkItems[i]
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemKey,
					Title: item.WorkItemName,
				}
			}

			// Query issues for this step
			issues, err := s.issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, item.WorkItemKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", item.WorkItemKey, err)
			}

			// Convert pointers to values for BoardColumn
			issueValues := make([]models.WorkIssue, len(issues))
			for i, issue := range issues {
				issueValues[i] = *issue
			}

			// Create column
			column := models.BoardColumn{
				ID:           fmt.Sprintf("col-%s", item.WorkItemKey),
				Name:         workItem.Title,
				WorkItemCode: item.WorkItemKey,
				WorkItemKey:  item.WorkItemKey,
				Order:        columnOrder,
				Issues:       issueValues,
			}

			columns = append(columns, column)
			columnOrder++
		}
	}

	// Create board
	board := &models.WorkbenchBoard{
		AgencyID:    agencyID,
		InstanceID:  instanceID,
		WorkflowID:  workflow.Key,
		TagKey:      "", // No tag used
		Columns:     columns,
		GeneratedAt: time.Now(),
	}

	return board, nil
}

// GetIssuesForColumn retrieves all issues in a specific board column
func (s *WorkbenchService) GetIssuesForColumn(ctx context.Context, agencyID, instanceID, workflowID, workItemKey string) ([]*models.WorkIssue, error) {
	issues, err := s.issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflowID, workItemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get column issues: %w", err)
	}
	return issues, nil
}
