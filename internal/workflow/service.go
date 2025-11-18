package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/sirupsen/logrus"
)

// Service provides business logic for workflow operations
type Service struct {
	repo   Repository
	logger *logrus.Logger
}

// NewService creates a new workflow service
func NewService(repo Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateWorkflow creates a new workflow with validation
func (s *Service) CreateWorkflow(ctx context.Context, workflow *models.Workflow) error {
	// Validate workflow
	if err := s.ValidateWorkflow(workflow); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create in repository
	if err := s.repo.Create(ctx, workflow); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	return nil
}

// GetWorkflow retrieves a workflow by ID
func (s *Service) GetWorkflow(ctx context.Context, id string) (*models.Workflow, error) {
	return s.repo.GetByID(ctx, id)
}

// GetWorkflowsByAgency retrieves all workflows for an agency
func (s *Service) GetWorkflowsByAgency(ctx context.Context, agencyID string) ([]*models.Workflow, error) {
	return s.repo.GetByAgencyID(ctx, agencyID)
}

// UpdateWorkflow updates an existing workflow with validation
func (s *Service) UpdateWorkflow(ctx context.Context, workflow *models.Workflow) error {
	// Validate workflow
	if err := s.ValidateWorkflow(workflow); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, workflow); err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	return nil
}

// DeleteWorkflow deletes a workflow (soft delete)
func (s *Service) DeleteWorkflow(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// DuplicateWorkflow creates a copy of an existing workflow
func (s *Service) DuplicateWorkflow(ctx context.Context, id string) (*models.Workflow, error) {
	// Get original workflow
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original workflow: %w", err)
	}

	// Create duplicate
	duplicate := &models.Workflow{
		Name:        original.Name + " (Copy)",
		Version:     "1.0.0",
		Description: original.Description,
		Steps:       make(models.Steps, len(original.Steps)),
		AgencyID:    original.AgencyID,
		CreatedBy:   original.CreatedBy,
	}

	// Deep copy steps
	copy(duplicate.Steps, original.Steps)

	// Create the duplicate
	if err := s.CreateWorkflow(ctx, duplicate); err != nil {
		return nil, fmt.Errorf("failed to create duplicate: %w", err)
	}

	return duplicate, nil
}

// ListWorkflows retrieves workflows with pagination
func (s *Service) ListWorkflows(ctx context.Context, limit, offset int) ([]*models.Workflow, error) {
	return s.repo.List(ctx, limit, offset)
}

// ValidateWorkflow validates a workflow definition
func (s *Service) ValidateWorkflow(workflow *models.Workflow) error {
	result := s.ValidateWorkflowStructure(workflow)
	if !result.Valid {
		return fmt.Errorf("workflow validation failed: %d errors", len(result.Errors))
	}
	return nil
}

// ValidateWorkflowStructure performs comprehensive validation and returns detailed results
func (s *Service) ValidateWorkflowStructure(workflow *models.Workflow) *models.WorkflowValidationResult {
	result := &models.WorkflowValidationResult{
		Valid:  true,
		Errors: []models.ValidationError{},
	}

	// Validate name
	if strings.TrimSpace(workflow.Name) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, models.ValidationError{
			Field:   "name",
			Message: "models.Workflow name is required",
		})
	} else if len(workflow.Name) < 3 {
		result.Valid = false
		result.Errors = append(result.Errors, models.ValidationError{
			Field:   "name",
			Message: "models.Workflow name must be at least 3 characters",
		})
	}

	// Validate version format (semantic versioning)
	if workflow.Version != "" {
		parts := strings.Split(workflow.Version, ".")
		if len(parts) != 3 {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "version",
				Message: "Version must be in semantic versioning format (x.y.z)",
			})
		}
	}

	// Validate agency_id
	if strings.TrimSpace(workflow.AgencyID) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, models.ValidationError{
			Field:   "agency_id",
			Message: "Agency ID is required",
		})
	}

	// Validate steps
	if len(workflow.Steps) > 0 {
		s.validateSteps(workflow, result)
	}

	return result
}

// validateSteps validates all steps in the workflow
func (s *Service) validateSteps(workflow *models.Workflow, result *models.WorkflowValidationResult) {
	stepIDs := make(map[string]bool)
	orders := make(map[int]bool)

	for _, step := range workflow.Steps {
		// Check for duplicate step IDs
		if stepIDs[step.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "steps",
				Message: fmt.Sprintf("Duplicate step ID: %s", step.ID),
				StepID:  step.ID,
			})
			continue
		}
		stepIDs[step.ID] = true

		// Check for duplicate order values
		if orders[step.Order] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "steps",
				Message: fmt.Sprintf("Duplicate order value: %d", step.Order),
				StepID:  step.ID,
			})
		}
		orders[step.Order] = true

		// Validate step has at least one item
		if len(step.Items) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "steps",
				Message: "Step must have at least one work item",
				StepID:  step.ID,
			})
			continue
		}

		// Validate items within step
		itemIDs := make(map[string]bool)
		for _, item := range step.Items {
			// Check for duplicate item IDs within step
			if itemIDs[item.ID] {
				result.Valid = false
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "steps",
					Message: fmt.Sprintf("Duplicate item ID in step: %s", item.ID),
					StepID:  step.ID,
					ItemID:  item.ID,
				})
				continue
			}
			itemIDs[item.ID] = true

			// Validate work item reference
			if strings.TrimSpace(item.WorkItemID) == "" {
				result.Valid = false
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "steps",
					Message: "Work item ID is required",
					StepID:  step.ID,
					ItemID:  item.ID,
				})
			}
		}
	}
}
