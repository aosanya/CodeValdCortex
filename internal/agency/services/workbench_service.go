package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	driver "github.com/arangodb/go-driver"
)

// DBClient interface for getting agency-specific databases
type DBClient interface {
	Client() driver.Client
	GetDatabase(ctx context.Context, dbName string) (driver.Database, error)
}

// WorkbenchService handles workbench board generation
type WorkbenchService struct {
	tagService       TagService
	instanceService  InstanceService
	dbClient         DBClient
	specRepo         agency.Repository
	issueRepoFactory agency.IssueRepositoryFactory
}

// NewWorkbenchService creates a new workbench service
func NewWorkbenchService(tagService TagService, instanceService InstanceService, dbClient DBClient, specRepo agency.Repository) *WorkbenchService {
	return &WorkbenchService{
		tagService:       tagService,
		instanceService:  instanceService,
		dbClient:         dbClient,
		specRepo:         specRepo,
		issueRepoFactory: nil, // Will be set by app initialization
	}
}

// SetIssueRepositoryFactory sets the factory function for creating issue repositories
func (s *WorkbenchService) SetIssueRepositoryFactory(factory agency.IssueRepositoryFactory) {
	s.issueRepoFactory = factory
}

// getIssueRepo creates an agency-specific issue repository
func (s *WorkbenchService) getIssueRepo(ctx context.Context, agencyID string) (agency.IssueRepository, error) {
	if s.issueRepoFactory == nil {
		return nil, fmt.Errorf("issue repository factory not configured")
	}

	// Agency database name is the agency ID
	dbName := agencyID

	// Get agency-specific database
	db, err := s.dbClient.GetDatabase(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency database %s: %w", dbName, err)
	}

	// Use factory to create the repository
	return s.issueRepoFactory(s.dbClient.Client(), db)
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

	// Get issue repository for this agency
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Track seen work items to avoid duplicate columns
	seenWorkItems := make(map[string]bool)

	for _, step := range workflow.Steps {
		for _, item := range step.Items {
			// Get the step identifier to use for deduplication and querying
			// Use WorkItemID if WorkItemKey is empty
			stepIdentifier := item.WorkItemKey
			if stepIdentifier == "" {
				stepIdentifier = item.WorkItemID
			}

			// Skip if we've already created a column for this work item
			if seenWorkItems[stepIdentifier] {
				continue
			}
			seenWorkItems[stepIdentifier] = true

			// Find work item details in specification
			var workItem *models.WorkItem

			// Try to match by WorkItemKey first, then by WorkItemID if Key is empty
			matchKey := item.WorkItemKey
			if matchKey == "" && item.WorkItemID != "" {
				// WorkItemKey is empty, try to extract key from WorkItemID
				// WorkItemID format is typically: "work_items/{key}"
				// For now, just use WorkItemID as-is for matching
				matchKey = item.WorkItemID
			}

			for i := range spec.WorkItems {
				// Match by Key field (ArangoDB _key) or by ID
				if spec.WorkItems[i].Key == matchKey || spec.WorkItems[i].ID == matchKey {
					workItem = &spec.WorkItems[i]
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemName, // Use the name as fallback for code
					Title: item.WorkItemName,
				}
			}

			// Query issues for this step
			issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, stepIdentifier)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", stepIdentifier, err)
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
				WorkItemCode: workItem.Code, // Use the actual Code field from WorkItem
				WorkItemKey:  item.WorkItemKey,
				Order:        columnOrder,
				Issues:       issueValues,
			}

			columns = append(columns, column)
			columnOrder++
		}
	}

	// Get instance to retrieve name
	instance, err := s.instanceService.GetInstance(ctx, agencyID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Create board
	board := &models.WorkbenchBoard{
		AgencyID:     agencyID,
		InstanceID:   instanceID,
		InstanceName: instance.InstanceName,
		WorkflowID:   workflow.Key,
		WorkflowName: workflow.Name,
		TagKey:       tag.Key,
		Columns:      columns,
		GeneratedAt:  time.Now(),
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

	// Get issue repository for this agency
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Track seen work items to avoid duplicate columns
	seenWorkItems := make(map[string]bool)

	for _, step := range workflow.Steps {
		for _, item := range step.Items {

			// Get the step identifier to use for deduplication and querying
			// Use WorkItemID if WorkItemKey is empty
			stepIdentifier := item.WorkItemKey
			if stepIdentifier == "" {
				stepIdentifier = item.WorkItemID
			}

			// Skip if we've already created a column for this work item
			if seenWorkItems[stepIdentifier] {
				continue
			}
			seenWorkItems[stepIdentifier] = true

			// Find work item details in specification
			var workItem *models.WorkItem

			// Try to match by WorkItemKey first, then by WorkItemID if Key is empty
			matchKey := item.WorkItemKey
			if matchKey == "" && item.WorkItemID != "" {
				// WorkItemKey is empty, try to extract key from WorkItemID
				// WorkItemID format is typically: "work_items/{key}"
				// For now, just use WorkItemID as-is for matching
				matchKey = item.WorkItemID
			}

			for i := range spec.WorkItems {
				// Match by Key field (ArangoDB _key) or by ID
				if spec.WorkItems[i].Key == matchKey || spec.WorkItems[i].ID == matchKey {
					workItem = &spec.WorkItems[i]
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemName, // Use the name as fallback for code
					Title: item.WorkItemName,
				}
			}
			// Query issues for this step
			issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, stepIdentifier)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", stepIdentifier, err)
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
				WorkItemCode: workItem.Code, // Use the actual Code field from WorkItem
				WorkItemKey:  item.WorkItemKey,
				Order:        columnOrder,
				Issues:       issueValues,
			}

			columns = append(columns, column)
			columnOrder++
		}
	}

	// Get instance to retrieve name
	instance, err := s.instanceService.GetInstance(ctx, agencyID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Create board
	board := &models.WorkbenchBoard{
		AgencyID:     agencyID,
		InstanceID:   instanceID,
		InstanceName: instance.InstanceName,
		WorkflowID:   workflow.Key,
		WorkflowName: workflow.Name,
		TagKey:       "", // No tag used
		Columns:      columns,
		GeneratedAt:  time.Now(),
	}

	return board, nil
}

// GenerateBoardForWorkflow generates a board for a specific workflow from a tag
func (s *WorkbenchService) GenerateBoardForWorkflow(ctx context.Context, agencyID, instanceID, tagName, workflowID string) (*models.WorkbenchBoard, error) {
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

	// Find the specified workflow
	var workflow *models.Workflow
	for i := range spec.Workflows {
		if spec.Workflows[i].Key == workflowID {
			workflow = &spec.Workflows[i]
			break
		}
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow %s not found in tag", workflowID)
	}

	// Generate board using the common logic
	return s.generateBoardForWorkflow(ctx, agencyID, instanceID, spec, workflow, tagName)
}

// GenerateBoardForWorkflowFromSpecification generates a board for a specific workflow from current specification
func (s *WorkbenchService) GenerateBoardForWorkflowFromSpecification(ctx context.Context, agencyID, instanceID, workflowID string) (*models.WorkbenchBoard, error) {
	// Get agency specification
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	// Find the specified workflow
	var workflow *models.Workflow
	for i := range spec.Workflows {
		if spec.Workflows[i].Key == workflowID {
			workflow = &spec.Workflows[i]
			break
		}
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow %s not found in specification", workflowID)
	}

	// Generate board using the common logic
	return s.generateBoardForWorkflow(ctx, agencyID, instanceID, spec, workflow, "")
}

// generateBoardForWorkflow is the common logic for generating a board from a workflow
func (s *WorkbenchService) generateBoardForWorkflow(ctx context.Context, agencyID, instanceID string, spec *models.AgencySpecification, workflow *models.Workflow, tagKey string) (*models.WorkbenchBoard, error) {
	// Create columns from workflow steps
	var columns []models.BoardColumn
	columnOrder := 0

	// Get issue repository for this agency
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Track seen work items to avoid duplicate columns
	seenWorkItems := make(map[string]bool)

	for _, step := range workflow.Steps {
		for _, item := range step.Items {
			stepIdentifier := item.WorkItemKey
			if stepIdentifier == "" {
				stepIdentifier = item.WorkItemID
			}

			if seenWorkItems[stepIdentifier] {
				continue
			}
			seenWorkItems[stepIdentifier] = true

			// Find work item details in specification
			var workItem *models.WorkItem
			matchKey := item.WorkItemKey
			if matchKey == "" && item.WorkItemID != "" {
				matchKey = item.WorkItemID
			}

			for i := range spec.WorkItems {
				if spec.WorkItems[i].Key == matchKey || spec.WorkItems[i].ID == matchKey {
					workItem = &spec.WorkItems[i]
					break
				}
			}

			if workItem == nil {
				workItem = &models.WorkItem{
					Code:  item.WorkItemName,
					Title: item.WorkItemName,
				}
			}

			// Query issues for this step
			issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, stepIdentifier)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", stepIdentifier, err)
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
				WorkItemCode: workItem.Code,
				WorkItemKey:  item.WorkItemKey,
				Order:        columnOrder,
				Issues:       issueValues,
			}

			columns = append(columns, column)
			columnOrder++
		}
	}

	// Get instance name
	instance, err := s.instanceService.GetInstance(ctx, agencyID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Create board
	board := &models.WorkbenchBoard{
		AgencyID:     agencyID,
		InstanceID:   instanceID,
		InstanceName: instance.InstanceName,
		WorkflowID:   workflow.Key,
		WorkflowName: workflow.Name,
		TagKey:       tagKey,
		Columns:      columns,
		GeneratedAt:  time.Now(),
	}

	return board, nil
}

// GetIssuesForColumn retrieves all issues in a specific board column
func (s *WorkbenchService) GetIssuesForColumn(ctx context.Context, agencyID, instanceID, workflowID, workItemKey string) ([]*models.WorkIssue, error) {
	// Get issue repository for this agency
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflowID, workItemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get column issues: %w", err)
	}
	return issues, nil
}

// GetWorkflowsFromTag retrieves all workflows from a tag snapshot
func (s *WorkbenchService) GetWorkflowsFromTag(ctx context.Context, agencyID, tagName string) ([]models.Workflow, error) {
	// Get tag snapshot
	tag, err := s.tagService.GetTag(ctx, agencyID, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Verify tag has a specification
	if tag.Snapshot.Specification.Key == "" {
		return nil, fmt.Errorf("tag has no specification snapshot")
	}

	return tag.Snapshot.Specification.Workflows, nil
}

// GetWorkflowsFromSpecification retrieves all workflows from current specification
func (s *WorkbenchService) GetWorkflowsFromSpecification(ctx context.Context, agencyID string) ([]models.Workflow, error) {
	// Get agency specification
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	return spec.Workflows, nil
}
