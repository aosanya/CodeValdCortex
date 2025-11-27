package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// IssueService handles business logic for work issues
type IssueService struct {
	issueRepo agency.IssueRepository
	specRepo  agency.Repository
}

// NewIssueService creates a new issue service
func NewIssueService(issueRepo agency.IssueRepository, specRepo agency.Repository) *IssueService {
	return &IssueService{
		issueRepo: issueRepo,
		specRepo:  specRepo,
	}
}

// CreateIssue creates a new issue at the workflow entry point
func (s *IssueService) CreateIssue(ctx context.Context, agencyID, instanceID string, req models.CreateIssueRequest) (*models.WorkIssue, error) {
	// Get agency specification to access workflows
	spec, err := s.specRepo.GetSpecification(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agency specification: %w", err)
	}

	// Find workflow in specification
	var workflow *models.Workflow
	for i := range spec.Workflows {
		if spec.Workflows[i].Key == req.WorkflowID {
			workflow = &spec.Workflows[i]
			break
		}
	}

	if workflow == nil {
		return nil, fmt.Errorf("workflow not found")
	}

	// Enforce entry point: issues MUST start at first workflow step
	if len(workflow.Steps) == 0 {
		return nil, fmt.Errorf("workflow has no steps defined")
	}

	firstStep := workflow.Steps[0]
	if len(firstStep.Items) == 0 {
		return nil, fmt.Errorf("first workflow step has no work items")
	}

	// Get entry point work item code
	entryPointCode := firstStep.Items[0].WorkItemKey

	// Get next issue number
	number, err := s.issueRepo.GetNextIssueNumber(ctx, agencyID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next issue number: %w", err)
	}

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

	// Create in repository
	if err := s.issueRepo.Create(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

// GetIssue retrieves an issue by ID
func (s *IssueService) GetIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// UpdateIssue updates an existing issue
func (s *IssueService) UpdateIssue(ctx context.Context, agencyID, instanceID, issueID string, req models.UpdateIssueRequest) (*models.WorkIssue, error) {
	// Get current issue
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
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
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	return issue, nil
}

// DeleteIssue deletes an issue
func (s *IssueService) DeleteIssue(ctx context.Context, agencyID, instanceID, issueID string) error {
	if err := s.issueRepo.Delete(ctx, agencyID, instanceID, issueID); err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}
	return nil
}

// AssignIssue assigns an issue to a worker
func (s *IssueService) AssignIssue(ctx context.Context, agencyID, instanceID, issueID, workerID string) (*models.WorkIssue, error) {
	// Get current issue
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
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
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to assign issue: %w", err)
	}

	return issue, nil
}

// ClaimIssue allows a worker to self-assign an available issue
func (s *IssueService) ClaimIssue(ctx context.Context, agencyID, instanceID, issueID, workerID string) (*models.WorkIssue, error) {
	// Get current issue
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
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
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to claim issue: %w", err)
	}

	return issue, nil
}

// UpdateIssueStatus updates the status of an issue
func (s *IssueService) UpdateIssueStatus(ctx context.Context, agencyID, instanceID, issueID, status string) (*models.WorkIssue, error) {
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
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Update status
	issue.Status = status
	issue.UpdatedAt = time.Now()

	// Save updates
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return issue, nil
}

// ProgressIssue moves an issue to the next workflow step
func (s *IssueService) ProgressIssue(ctx context.Context, agencyID, instanceID, issueID string) (*models.WorkIssue, error) {
	// Get current issue
	issue, err := s.issueRepo.GetByID(ctx, agencyID, instanceID, issueID)
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
			if item.WorkItemKey == issue.CurrentStep {
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
		issue.CurrentStep = nextStep.Items[0].WorkItemKey
		issue.Status = models.IssueStatusOpen
		issue.AssignedTo = "" // Clear assignment for next step
		issue.UpdatedAt = time.Now()
	}

	// Save updates
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to progress issue: %w", err)
	}

	return issue, nil
}

// GetIssuesByStep retrieves all issues at a specific workflow step
func (s *IssueService) GetIssuesByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) ([]*models.WorkIssue, error) {
	issues, err := s.issueRepo.ListByWorkflowStep(ctx, agencyID, instanceID, workflowID, step)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues by step: %w", err)
	}
	return issues, nil
}

// GetAvailableWork retrieves all available (unassigned) issues
func (s *IssueService) GetAvailableWork(ctx context.Context, agencyID, instanceID string, step string) ([]*models.WorkIssue, error) {
	issues, err := s.issueRepo.ListAvailable(ctx, agencyID, instanceID, step)
	if err != nil {
		return nil, fmt.Errorf("failed to get available work: %w", err)
	}
	return issues, nil
}

// ListIssues retrieves issues with filters
func (s *IssueService) ListIssues(ctx context.Context, agencyID, instanceID string, filters agency.IssueFilters) ([]*models.WorkIssue, error) {
	issues, err := s.issueRepo.List(ctx, agencyID, instanceID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	return issues, nil
}

// CountIssuesByStep counts issues at a specific workflow step
func (s *IssueService) CountIssuesByStep(ctx context.Context, agencyID, instanceID, workflowID, step string) (int, error) {
	count, err := s.issueRepo.CountByStep(ctx, agencyID, instanceID, workflowID, step)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	return count, nil
}
