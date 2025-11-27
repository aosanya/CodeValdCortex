package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// IssueService handles business logic for work issues
type IssueService struct {
	specRepo         agency.Repository
	dbClient         DBClient
	issueRepoFactory agency.IssueRepositoryFactory
}

// NewIssueService creates a new issue service
func NewIssueService(specRepo agency.Repository, dbClient DBClient) *IssueService {
	return &IssueService{
		specRepo:         specRepo,
		dbClient:         dbClient,
		issueRepoFactory: nil, // Will be set by app initialization
	}
}

// SetIssueRepositoryFactory sets the factory function for creating issue repositories
func (s *IssueService) SetIssueRepositoryFactory(factory agency.IssueRepositoryFactory) {
	s.issueRepoFactory = factory
}

// getIssueRepo creates an agency-specific issue repository
func (s *IssueService) getIssueRepo(ctx context.Context, agencyID string) (agency.IssueRepository, error) {
	if s.issueRepoFactory == nil {
		return nil, fmt.Errorf("issue repository factory not configured")
	}

	// Use agencyID as database name (standard pattern)
	db, err := s.dbClient.GetDatabase(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency database: %w", err)
	}

	// Create repository with agency-specific database
	repo, err := s.issueRepoFactory(s.dbClient.Client(), db)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue repository: %w", err)
	}

	return repo, nil
}

// CreateIssue creates a new issue at the workflow entry point
func (s *IssueService) CreateIssue(ctx context.Context, agencyID, instanceID string, req models.CreateIssueRequest) (*models.WorkIssue, error) {
	log.Printf("[MVP-WI-008] CreateIssue - agencyID: %s, instanceID: %s, title: '%s', workflowID: '%s'",
		agencyID, instanceID, req.Title, req.WorkflowID)

	// Get agency specification to access workflows
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		log.Printf("[MVP-WI-008] Failed to get agency specification: %v", err)
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	log.Printf("[MVP-WI-008] Retrieved specification with %d workflows", len(spec.Workflows))

	// Find workflow in specification
	var workflow *models.Workflow
	for i := range spec.Workflows {
		log.Printf("[MVP-WI-008] Checking workflow[%d] - Key: '%s', Name: '%s'", i, spec.Workflows[i].Key, spec.Workflows[i].Name)
		if spec.Workflows[i].Key == req.WorkflowID {
			workflow = &spec.Workflows[i]
			log.Printf("[MVP-WI-008] Found matching workflow: %s", workflow.Name)
			break
		}
	}

	if workflow == nil {
		log.Printf("[MVP-WI-008] Workflow not found for ID: %s", req.WorkflowID)
		return nil, fmt.Errorf("workflow not found")
	}

	// Enforce entry point: issues MUST start at first workflow step
	if len(workflow.Steps) == 0 {
		log.Printf("[MVP-WI-008] Workflow %s has no steps defined", workflow.Key)
		return nil, fmt.Errorf("workflow has no steps defined")
	}

	log.Printf("[MVP-WI-008] Workflow has %d steps", len(workflow.Steps))

	firstStep := workflow.Steps[0]
	if len(firstStep.Items) == 0 {
		log.Printf("[MVP-WI-008] First workflow step has no work items")
		return nil, fmt.Errorf("first workflow step has no work items")
	}

	log.Printf("[MVP-WI-008] First step has %d items", len(firstStep.Items))

	// Get entry point work item code - use WorkItemID if WorkItemKey is empty
	entryPointCode := firstStep.Items[0].WorkItemKey
	if entryPointCode == "" {
		entryPointCode = firstStep.Items[0].WorkItemID
	}
	log.Printf("[MVP-WI-008] Entry point WorkItemKey: '%s', WorkItemID: '%s', WorkItemName: '%s', Using: '%s'",
		firstStep.Items[0].WorkItemKey, firstStep.Items[0].WorkItemID, firstStep.Items[0].WorkItemName, entryPointCode)

	// Get issue repository for this agency
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		log.Printf("[MVP-WI-008] Failed to get issue repository: %v", err)
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Get next issue number
	number, err := issueRepo.GetNextIssueNumber(ctx, agencyID, instanceID)
	if err != nil {
		log.Printf("[MVP-WI-008] Failed to get next issue number: %v", err)
		return nil, fmt.Errorf("failed to get next issue number: %w", err)
	}

	log.Printf("[MVP-WI-008] Next issue number: %d", number)

	// Create issue at entry point
	issue := &models.WorkIssue{
		Title:          req.Title,
		Description:    req.Description,
		Number:         number,
		AgencyID:       agencyID,
		InstanceID:     instanceID,
		WorkflowID:     req.WorkflowID,
		CurrentStep:    entryPointCode,
		Status:         models.IssueStatusOpen,
		AssignedTo:     "",
		CompletedSteps: []string{},
	}

	log.Printf("[MVP-WI-008] Creating issue: Number=%d, Title='%s', CurrentStep='%s', Status='%s'",
		issue.Number, issue.Title, issue.CurrentStep, issue.Status)

	// Create in repository
	if err := issueRepo.Create(ctx, issue); err != nil {
		log.Printf("[MVP-WI-008] Failed to create issue in repository: %v", err)
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	log.Printf("[MVP-WI-008] Issue created successfully with Key: %s", issue.Key)
	return issue, nil
}

// GetIssue retrieves an issue by ID
func (s *IssueService) GetIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// UpdateIssue updates an existing issue
func (s *IssueService) UpdateIssue(ctx context.Context, agencyID, instanceID, issueID string, req models.UpdateIssueRequest) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Get current issue
	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Update fields if provided
	if req.Status != nil {
		issue.Status = *req.Status
	}

	if req.AssignedTo != nil {
		issue.AssignedTo = *req.AssignedTo
	}

	if req.BranchName != nil {
		issue.BranchName = *req.BranchName
	}

	if req.PullRequestID != nil {
		issue.PullRequestID = *req.PullRequestID
	}

	issue.UpdatedAt = time.Now()

	// Save updates
	if err := issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	return issue, nil
}

// DeleteIssue deletes an issue
func (s *IssueService) DeleteIssue(ctx context.Context, agencyID, instanceID, issueID string) error {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get issue repository: %w", err)
	}

	if err := issueRepo.Delete(ctx, agencyID, instanceID, issueID); err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}
	return nil
}

// AssignIssue assigns an issue to a worker
func (s *IssueService) AssignIssue(ctx context.Context, agencyID, instanceID, issueID, workerID string) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Get current issue
	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Check if already assigned
	if issue.AssignedTo != "" && issue.AssignedTo != workerID {
		return nil, fmt.Errorf("issue already assigned to another worker")
	}

	// Assign worker
	issue.AssignedTo = workerID
	issue.Status = models.IssueStatusAssigned
	issue.UpdatedAt = time.Now()

	// Save updates
	if err := issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to assign issue: %w", err)
	}

	return issue, nil
}

// ClaimIssue allows a worker to self-assign an available issue
func (s *IssueService) ClaimIssue(ctx context.Context, agencyID, instanceID, issueID, workerID string) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Get current issue
	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Verify issue is available (open and unassigned)
	if issue.Status != models.IssueStatusOpen {
		return nil, fmt.Errorf("issue is not available for claiming (status: %s)", issue.Status)
	}

	if issue.AssignedTo != "" {
		return nil, fmt.Errorf("issue already assigned to worker: %s", issue.AssignedTo)
	}

	// Claim issue
	issue.AssignedTo = workerID
	issue.Status = models.IssueStatusAssigned
	issue.UpdatedAt = time.Now()

	// Save updates
	if err := issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to claim issue: %w", err)
	}

	return issue, nil
}

// UpdateIssueStatus updates the status of an issue
func (s *IssueService) UpdateIssueStatus(ctx context.Context, agencyID, instanceID, issueID, status string) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Validate status
	validStatuses := []string{
		models.IssueStatusOpen,
		models.IssueStatusAssigned,
		models.IssueStatusInProgress,
		models.IssueStatusReadyForReview,
		models.IssueStatusCompleted,
		models.IssueStatusBlocked,
	}

	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	// Get current issue
	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Update status
	issue.Status = status
	issue.UpdatedAt = time.Now()

	// Save updates
	if err := issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return issue, nil
}

// ProgressIssue moves an issue to the next workflow step
func (s *IssueService) ProgressIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	// Get current issue
	issue, err := issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Get agency specification to access workflows
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	// Find workflow in specification
	var workflow *models.Workflow
	for i := range spec.Workflows {
		if spec.Workflows[i].Key == issue.WorkflowID {
			workflow = &spec.Workflows[i]
			break
		}
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found")
	}

	// Find current step index
	currentStepIndex := -1
	for i, step := range workflow.Steps {
		for _, item := range step.Items {
			// Use WorkItemID if WorkItemKey is empty
			stepID := item.WorkItemKey
			if stepID == "" {
				stepID = item.WorkItemID
			}
			if stepID == issue.CurrentStep {
				currentStepIndex = i
				break
			}
		}
		if currentStepIndex != -1 {
			break
		}
	}

	if currentStepIndex == -1 {
		return nil, fmt.Errorf("current step not found in workflow")
	}

	// Check if there's a next step
	if currentStepIndex >= len(workflow.Steps)-1 {
		// Already at final step, mark as completed
		issue.Status = models.IssueStatusCompleted
		issue.UpdatedAt = time.Now()
	} else {
		// Move to next step
		nextStep := workflow.Steps[currentStepIndex+1]
		if len(nextStep.Items) == 0 {
			return nil, fmt.Errorf("next step has no work items")
		}

		// Add current step to completed steps
		issue.CompletedSteps = append(issue.CompletedSteps, issue.CurrentStep)

		// Move to next step's first work item
		// Use WorkItemID if WorkItemKey is empty
		nextStepID := nextStep.Items[0].WorkItemKey
		if nextStepID == "" {
			nextStepID = nextStep.Items[0].WorkItemID
		}
		issue.CurrentStep = nextStepID
		issue.Status = models.IssueStatusOpen
		issue.AssignedTo = "" // Clear assignment for next step
		issue.UpdatedAt = time.Now()
	}

	// Save updates
	if err := issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to progress issue: %w", err)
	}

	return issue, nil
}

// GetIssuesByStep retrieves all issues at a specific workflow step
func (s *IssueService) GetIssuesByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) ([]*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	issues, err := issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflowID, step)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues by step: %w", err)
	}
	return issues, nil
}

// GetAvailableWork retrieves all available (unassigned) issues
func (s *IssueService) GetAvailableWork(ctx context.Context, agencyID, instanceID string, step string) ([]*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	issues, err := issueRepo.ListAvailable(ctx, agencyID, instanceID, step)
	if err != nil {
		return nil, fmt.Errorf("failed to get available work: %w", err)
	}
	return issues, nil
}

// ListIssues retrieves issues with filters
func (s *IssueService) ListIssues(ctx context.Context, agencyID, instanceID string, filters agency.IssueFilters) ([]*models.WorkIssue, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue repository: %w", err)
	}

	issues, err := issueRepo.List(ctx, agencyID, instanceID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	return issues, nil
}

// CountIssuesByStep counts issues at a specific workflow step
func (s *IssueService) CountIssuesByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) (int, error) {
	issueRepo, err := s.getIssueRepo(ctx, agencyID)
	if err != nil {
		return 0, fmt.Errorf("failed to get issue repository: %w", err)
	}

	count, err := issueRepo.CountByStep(ctx, agencyID, instanceID, workflowID, step)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	return count, nil
}
