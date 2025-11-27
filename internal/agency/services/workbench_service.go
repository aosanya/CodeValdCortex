package services

import (
	"context"
	"fmt"
	"log"
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

	log.Printf("[MVP-WI-008] GenerateBoard - workflow.Key: %s, workflow.Name: %s, spec has %d work items",
		workflow.Key, workflow.Name, len(spec.WorkItems))
	log.Printf("[MVP-WI-008] Workflow has %d steps", len(workflow.Steps))

	for stepIdx, step := range workflow.Steps {
		log.Printf("[MVP-WI-008] Step %d (Order: %d) has %d items", stepIdx, step.Order, len(step.Items))
		for itemIdx, item := range step.Items {
			log.Printf("[MVP-WI-008] Step item [%d][%d] - ID: '%s', WorkItemID: '%s', WorkItemKey: '%s', WorkItemName: '%s'",
				stepIdx, itemIdx, item.ID, item.WorkItemID, item.WorkItemKey, item.WorkItemName)

			// Find work item details in specification
			var workItem *models.WorkItem

			// Try to match by WorkItemKey first, then by WorkItemID if Key is empty
			matchKey := item.WorkItemKey
			if matchKey == "" && item.WorkItemID != "" {
				// WorkItemKey is empty, try to extract key from WorkItemID
				// WorkItemID format is typically: "work_items/{key}"
				log.Printf("[MVP-WI-008] WorkItemKey is empty, trying to extract from WorkItemID: '%s'", item.WorkItemID)
				// For now, just use WorkItemID as-is for matching
				matchKey = item.WorkItemID
			}

			log.Printf("[MVP-WI-008] Attempting to match with key: '%s'", matchKey)

			for i := range spec.WorkItems {
				// Match by Key field (ArangoDB _key) or by ID
				if spec.WorkItems[i].Key == matchKey || spec.WorkItems[i].ID == matchKey {
					workItem = &spec.WorkItems[i]
					log.Printf("[MVP-WI-008] Found matching work item in spec - Key: %s, ID: %s, Code: %s, Title: %s",
						workItem.Key, workItem.ID, workItem.Code, workItem.Title)
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemName, // Use the name as fallback for code
					Title: item.WorkItemName,
				}
				log.Printf("[MVP-WI-008] Work item NOT found in spec, using fallback - Code: %s, Title: %s",
					workItem.Code, workItem.Title)
			}

			// Get the step identifier to use for querying issues
			// Use WorkItemID if WorkItemKey is empty
			stepIdentifier := item.WorkItemKey
			if stepIdentifier == "" {
				stepIdentifier = item.WorkItemID
			}
			log.Printf("[MVP-WI-008] GenerateBoard - Querying issues with stepIdentifier: '%s' (WorkItemKey='%s', WorkItemID='%s')",
				stepIdentifier, item.WorkItemKey, item.WorkItemID)

			// Query issues for this step
			issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, stepIdentifier)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", stepIdentifier, err)
			}

			log.Printf("[MVP-WI-008] GenerateBoard - Found %d issues for step '%s'", len(issues), stepIdentifier)

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

			log.Printf("[MVP-WI-008] Created column - ID: %s, Name: '%s', Code: '%s', Order: %d, Issues: %d",
				column.ID, column.Name, column.WorkItemCode, column.Order, len(column.Issues))

			columns = append(columns, column)
			columnOrder++
		}
	}

	log.Printf("[MVP-WI-008] Total columns created: %d", len(columns))

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

	log.Printf("[MVP-WI-008] GenerateBoardFromSpecification - workflow.Key: %s, workflow.Name: %s, spec has %d work items",
		workflow.Key, workflow.Name, len(spec.WorkItems))
	log.Printf("[MVP-WI-008] Workflow has %d steps", len(workflow.Steps))

	for stepIdx, step := range workflow.Steps {
		log.Printf("[MVP-WI-008] Step %d (Order: %d) has %d items", stepIdx, step.Order, len(step.Items))
		for itemIdx, item := range step.Items {
			log.Printf("[MVP-WI-008] Step item [%d][%d] - ID: '%s', WorkItemID: '%s', WorkItemKey: '%s', WorkItemName: '%s'",
				stepIdx, itemIdx, item.ID, item.WorkItemID, item.WorkItemKey, item.WorkItemName)

			// Find work item details in specification
			var workItem *models.WorkItem

			// Try to match by WorkItemKey first, then by WorkItemID if Key is empty
			matchKey := item.WorkItemKey
			if matchKey == "" && item.WorkItemID != "" {
				// WorkItemKey is empty, try to extract key from WorkItemID
				// WorkItemID format is typically: "work_items/{key}"
				log.Printf("[MVP-WI-008] WorkItemKey is empty, trying to extract from WorkItemID: '%s'", item.WorkItemID)
				// For now, just use WorkItemID as-is for matching
				matchKey = item.WorkItemID
			}

			log.Printf("[MVP-WI-008] Attempting to match with key: '%s'", matchKey)

			for i := range spec.WorkItems {
				// Match by Key field (ArangoDB _key) or by ID
				if spec.WorkItems[i].Key == matchKey || spec.WorkItems[i].ID == matchKey {
					workItem = &spec.WorkItems[i]
					log.Printf("[MVP-WI-008] Found matching work item in spec - Key: %s, ID: %s, Code: %s, Title: %s",
						workItem.Key, workItem.ID, workItem.Code, workItem.Title)
					break
				}
			}

			if workItem == nil {
				// Work item not found in spec, use basic info from workflow
				workItem = &models.WorkItem{
					Code:  item.WorkItemName, // Use the name as fallback for code
					Title: item.WorkItemName,
				}
				log.Printf("[MVP-WI-008] Work item NOT found in spec, using fallback - Code: %s, Title: %s",
					workItem.Code, workItem.Title)
			}

			// Get the step identifier to use for querying issues
			// Use WorkItemID if WorkItemKey is empty
			stepIdentifier := item.WorkItemKey
			if stepIdentifier == "" {
				stepIdentifier = item.WorkItemID
			}
			log.Printf("[MVP-WI-008] GenerateBoardFromSpecification - Querying issues with stepIdentifier: '%s' (WorkItemKey='%s', WorkItemID='%s')",
				stepIdentifier, item.WorkItemKey, item.WorkItemID)

			// Query issues for this step
			issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflow.Key, stepIdentifier)
			if err != nil {
				return nil, fmt.Errorf("failed to get issues for step %s: %w", stepIdentifier, err)
			}

			log.Printf("[MVP-WI-008] GenerateBoardFromSpecification - Found %d issues for step '%s'", len(issues), stepIdentifier)

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

			log.Printf("[MVP-WI-008] Created column - ID: %s, Name: '%s', Code: '%s', Order: %d, Issues: %d",
				column.ID, column.Name, column.WorkItemCode, column.Order, len(column.Issues))

			columns = append(columns, column)
			columnOrder++
		}
	}

	log.Printf("[MVP-WI-008] Total columns created: %d", len(columns))

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
