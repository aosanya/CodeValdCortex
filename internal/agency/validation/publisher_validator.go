package validation

import (
	"fmt"
	"log/slog"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// ValidationError represents a validation error preventing publication
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a non-critical issue
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult contains the results of pre-publish validation
type ValidationResult struct {
	Valid           bool                `json:"valid"`
	Errors          []ValidationError   `json:"errors"`
	Warnings        []ValidationWarning `json:"warnings"`
	Recommendations []string            `json:"recommendations"`
}

// PublisherValidator validates agencies for publication
type PublisherValidator struct {
	logger *slog.Logger
}

// NewPublisherValidator creates a new PublisherValidator instance
func NewPublisherValidator(logger *slog.Logger) *PublisherValidator {
	return &PublisherValidator{
		logger: logger,
	}
}

// ValidateForPublish performs comprehensive pre-publish validation
func (v *PublisherValidator) ValidateForPublish(
	agencyDoc *models.Agency,
	spec *models.AgencySpecification,
) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:           true,
		Errors:          []ValidationError{},
		Warnings:        []ValidationWarning{},
		Recommendations: []string{},
	}

	if agencyDoc == nil {
		return nil, fmt.Errorf("agency document is nil")
	}

	// Check 1: Specification must exist
	if spec == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "specification",
			Message: "Agency must have a specification",
			Code:    "SPEC_MISSING",
		})
		result.Valid = false
		return result, nil
	}

	// Check 2: Introduction must be non-empty
	if spec.IntroductionText() == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "specification.introduction",
			Message: "Agency must have an introduction",
			Code:    "INTRO_MISSING",
		})
		result.Valid = false
	} else if len(spec.IntroductionText()) < 50 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "specification.introduction",
			Message: "Introduction is very short (less than 50 characters)",
		})
		result.Recommendations = append(result.Recommendations,
			"Consider providing a more detailed introduction to help users understand the agency's purpose")
	}

	// Check 3: At least one goal
	if len(spec.Goals) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "specification.goals",
			Message: "Agency must have at least one goal",
			Code:    "GOALS_MISSING",
		})
		result.Valid = false
	} else {
		v.validateGoals(spec.Goals, result)
	}

	// Check 4: At least one role
	if len(spec.Roles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "specification.roles",
			Message: "Agency must have at least one role",
			Code:    "ROLES_MISSING",
		})
		result.Valid = false
	} else {
		v.validateRoles(spec.Roles, result)
	}

	// Check 5: At least one workflow
	if len(spec.Workflows) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "specification.workflows",
			Message: "Agency has no workflows defined",
		})
		result.Recommendations = append(result.Recommendations,
			"Consider adding workflows to orchestrate agent activities")
	} else {
		v.validateWorkflows(spec.Workflows, result)
	}

	// Check 6: Agency state validation (allow draft, published for republishing)
	// Empty state is treated as draft
	if agencyDoc.State != "" &&
		agencyDoc.State != models.AgencyStateDraft &&
		agencyDoc.State != models.AgencyStatePublished {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "state",
			Message: fmt.Sprintf("Cannot publish agency in state: %s", agencyDoc.State),
			Code:    "INVALID_STATE",
		})
		result.Valid = false
	}

	// Check 7: Agency name must be non-empty
	if agencyDoc.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "Agency must have a name",
			Code:    "NAME_MISSING",
		})
		result.Valid = false
	}

	// Log validation summary
	v.logger.Info("validation completed",
		"agency_id", agencyDoc.Key,
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings))

	return result, nil
}

// validateGoals checks individual goals for completeness
func (v *PublisherValidator) validateGoals(goals []models.Goal, result *ValidationResult) {
	for i, goal := range goals {
		if goal.Description == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("specification.goals[%d].description", i),
				Message: "Goal description cannot be empty",
				Code:    "GOAL_DESC_MISSING",
			})
			result.Valid = false
		}

		if len(goal.SuccessMetrics) == 0 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Goal '%s' would benefit from defined success metrics", goal.Description))
		}
	}
}

// validateRoles checks individual roles for completeness
func (v *PublisherValidator) validateRoles(roles []models.Role, result *ValidationResult) {
	roleCodesSeen := make(map[string]bool)

	for i, role := range roles {
		// Check for required fields
		if role.Code == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("specification.roles[%d].code", i),
				Message: "Role code cannot be empty",
				Code:    "ROLE_CODE_MISSING",
			})
			result.Valid = false
		} else {
			// Check for duplicate role codes
			if roleCodesSeen[role.Code] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("specification.roles[%d].code", i),
					Message: fmt.Sprintf("Duplicate role code: %s", role.Code),
					Code:    "ROLE_CODE_DUPLICATE",
				})
				result.Valid = false
			}
			roleCodesSeen[role.Code] = true
		}

		if role.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("specification.roles[%d].name", i),
				Message: "Role name cannot be empty",
				Code:    "ROLE_NAME_MISSING",
			})
			result.Valid = false
		}

		if role.Description == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.roles[%d].description", i),
				Message: "Role has no description",
			})
		}

		// Recommendations for roles (no critical validation needed)
		if role.Description != "" && len(role.Description) < 20 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Role '%s' would benefit from a more detailed description", role.Name))
		}

		// Check autonomy level
		if role.AutonomyLevel == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.roles[%d].autonomy_level", i),
				Message: "Role has no autonomy level set",
			})
		}

		// Check token budget
		if role.TokenBudget <= 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.roles[%d].token_budget", i),
				Message: "Role has no token budget set",
			})
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Role '%s' should have a token budget defined for resource management", role.Name))
		}
	}
}

// validateWorkflows checks individual workflows for completeness
func (v *PublisherValidator) validateWorkflows(workflows []models.Workflow, result *ValidationResult) {
	workflowNamesSeen := make(map[string]bool)

	for i, workflow := range workflows {
		// Check for required fields
		if workflow.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("specification.workflows[%d].name", i),
				Message: "Workflow name cannot be empty",
				Code:    "WORKFLOW_NAME_MISSING",
			})
			result.Valid = false
		} else {
			// Check for duplicate workflow names
			if workflowNamesSeen[workflow.Name] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("specification.workflows[%d].name", i),
					Message: fmt.Sprintf("Duplicate workflow name: %s", workflow.Name),
					Code:    "WORKFLOW_NAME_DUPLICATE",
				})
				result.Valid = false
			}
			workflowNamesSeen[workflow.Name] = true
		}

		if workflow.Description == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.workflows[%d].description", i),
				Message: "Workflow has no description",
			})
		}

		if len(workflow.Steps) == 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.workflows[%d].steps", i),
				Message: "Workflow has no steps defined",
			})
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Workflow '%s' should define steps for execution", workflow.Name))
		} else {
			v.validateWorkflowSteps(workflow, i, result)
		}
	}
}

// validateWorkflowSteps checks workflow steps for completeness
func (v *PublisherValidator) validateWorkflowSteps(workflow models.Workflow, workflowIndex int, result *ValidationResult) {
	stepIDsSeen := make(map[string]bool)

	for j, step := range workflow.Steps {
		if step.ID == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("specification.workflows[%d].steps[%d].id", workflowIndex, j),
				Message: "Workflow step ID cannot be empty",
				Code:    "STEP_ID_MISSING",
			})
			result.Valid = false
		} else {
			// Check for duplicate step IDs within workflow
			if stepIDsSeen[step.ID] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("specification.workflows[%d].steps[%d].id", workflowIndex, j),
					Message: fmt.Sprintf("Duplicate step ID in workflow '%s': %s", workflow.Name, step.ID),
					Code:    "STEP_ID_DUPLICATE",
				})
				result.Valid = false
			}
			stepIDsSeen[step.ID] = true
		}

		// Validate step items
		if len(step.Items) == 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fmt.Sprintf("specification.workflows[%d].steps[%d].items", workflowIndex, j),
				Message: "Workflow step has no items",
			})
		}

		// Validate each step item
		for k, item := range step.Items {
			if item.WorkItemID == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("specification.workflows[%d].steps[%d].items[%d].work_item_id", workflowIndex, j, k),
					Message: "Step item must have a work_item_id",
					Code:    "STEP_ITEM_MISSING_WORK_ITEM",
				})
				result.Valid = false
			}
		}
	}
}
