package agency

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RACIService handles RACI matrix operations
type RACIService struct {
	repo      Repository
	validator *RACIValidator
}

// NewRACIService creates a new RACI service
func NewRACIService(repo Repository) *RACIService {
	return &RACIService{
		repo:      repo,
		validator: NewRACIValidator(),
	}
}

// CreateMatrix creates a new RACI matrix
func (s *RACIService) CreateMatrix(ctx context.Context, agencyID string, req *CreateRACIMatrixRequest) (*RACIMatrix, error) {
	// Generate key
	key := fmt.Sprintf("raci_%s", uuid.New().String()[:8])

	// Assign IDs to activities if missing
	for i := range req.Activities {
		if req.Activities[i].ID == "" {
			req.Activities[i].ID = fmt.Sprintf("activity_%d", i+1)
		}
		if req.Activities[i].Order == 0 {
			req.Activities[i].Order = i + 1
		}
	}

	matrix := &RACIMatrix{
		Key:         key,
		AgencyID:    agencyID,
		WorkItemKey: req.WorkItemKey,
		Name:        req.Name,
		Description: req.Description,
		Activities:  req.Activities,
		Roles:       req.Roles,
		TemplateID:  req.TemplateID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate the matrix
	validationResult := s.validator.ValidateMatrix(matrix)
	matrix.IsValid = validationResult.IsValid

	// Store validation errors
	if !validationResult.IsValid {
		matrix.Errors = make([]string, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			matrix.Errors[i] = err.Message
		}
	}

	// Save to database (implementation depends on repository)
	if err := s.repo.SaveRACIMatrix(ctx, agencyID, matrix); err != nil {
		return nil, fmt.Errorf("failed to save RACI matrix: %w", err)
	}

	return matrix, nil
}

// GetMatrix retrieves a RACI matrix by key
func (s *RACIService) GetMatrix(ctx context.Context, agencyID, key string) (*RACIMatrix, error) {
	matrix, err := s.repo.GetRACIMatrix(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RACI matrix: %w", err)
	}
	return matrix, nil
}

// ListMatrices lists all RACI matrices for an agency
func (s *RACIService) ListMatrices(ctx context.Context, agencyID string) ([]*RACIMatrix, error) {
	matrices, err := s.repo.ListRACIMatrices(ctx, agencyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list RACI matrices: %w", err)
	}
	return matrices, nil
}

// UpdateMatrix updates an existing RACI matrix
func (s *RACIService) UpdateMatrix(ctx context.Context, agencyID, key string, req *UpdateRACIMatrixRequest) (*RACIMatrix, error) {
	// Get existing matrix
	matrix, err := s.repo.GetRACIMatrix(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RACI matrix: %w", err)
	}

	// Update fields
	matrix.Name = req.Name
	matrix.Description = req.Description
	matrix.WorkItemKey = req.WorkItemKey
	matrix.Activities = req.Activities
	matrix.Roles = req.Roles
	matrix.UpdatedAt = time.Now()

	// Assign IDs to activities if missing
	for i := range matrix.Activities {
		if matrix.Activities[i].ID == "" {
			matrix.Activities[i].ID = fmt.Sprintf("activity_%d", i+1)
		}
		if matrix.Activities[i].Order == 0 {
			matrix.Activities[i].Order = i + 1
		}
	}

	// Validate the updated matrix
	validationResult := s.validator.ValidateMatrix(matrix)
	matrix.IsValid = validationResult.IsValid

	// Store validation errors
	matrix.Errors = []string{}
	if !validationResult.IsValid {
		matrix.Errors = make([]string, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			matrix.Errors[i] = err.Message
		}
	}

	// Save updated matrix
	if err := s.repo.UpdateRACIMatrix(ctx, agencyID, matrix); err != nil {
		return nil, fmt.Errorf("failed to update RACI matrix: %w", err)
	}

	return matrix, nil
}

// DeleteMatrix deletes a RACI matrix
func (s *RACIService) DeleteMatrix(ctx context.Context, agencyID, key string) error {
	if err := s.repo.DeleteRACIMatrix(ctx, agencyID, key); err != nil {
		return fmt.Errorf("failed to delete RACI matrix: %w", err)
	}
	return nil
}

// ValidateMatrix validates a RACI matrix and returns detailed results
func (s *RACIService) ValidateMatrix(ctx context.Context, agencyID, key string) (*RACIValidationResult, error) {
	matrix, err := s.repo.GetRACIMatrix(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RACI matrix: %w", err)
	}

	return s.validator.ValidateMatrix(matrix), nil
}

// GetTemplates returns available RACI templates
func (s *RACIService) GetTemplates(ctx context.Context) ([]*RACITemplate, error) {
	// Return built-in templates
	templates := s.getBuiltInTemplates()

	// TODO: Also fetch custom templates from database
	// customTemplates, err := s.repo.ListRACITemplates(ctx)
	// if err == nil {
	//     templates = append(templates, customTemplates...)
	// }

	return templates, nil
}

// ApplyTemplate applies a template to create a new RACI matrix
func (s *RACIService) ApplyTemplate(ctx context.Context, agencyID, templateID string, name string) (*RACIMatrix, error) {
	// Get template
	templates := s.getBuiltInTemplates()
	var template *RACITemplate
	for _, t := range templates {
		if t.Key == templateID {
			template = t
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Create matrix from template
	req := &CreateRACIMatrixRequest{
		Name:       name,
		Activities: template.Activities,
		Roles:      template.Roles,
		TemplateID: templateID,
	}

	return s.CreateMatrix(ctx, agencyID, req)
}

// ExportToJSON exports a RACI matrix to JSON
func (s *RACIService) ExportToJSON(ctx context.Context, agencyID, key string) ([]byte, error) {
	matrix, err := s.repo.GetRACIMatrix(ctx, agencyID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RACI matrix: %w", err)
	}

	data, err := json.MarshalIndent(matrix, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// ExportToMarkdown exports a RACI matrix to Markdown format
func (s *RACIService) ExportToMarkdown(ctx context.Context, agencyID, key string) (string, error) {
	matrix, err := s.repo.GetRACIMatrix(ctx, agencyID, key)
	if err != nil {
		return "", fmt.Errorf("failed to get RACI matrix: %w", err)
	}

	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# RACI Matrix: %s\n\n", matrix.Name))

	if matrix.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", matrix.Description))
	}

	// Legend
	sb.WriteString("## RACI Legend\n\n")
	sb.WriteString("- **R** = Responsible (Does the work)\n")
	sb.WriteString("- **A** = Accountable (Ultimately answerable)\n")
	sb.WriteString("- **C** = Consulted (Provides input)\n")
	sb.WriteString("- **I** = Informed (Kept in the loop)\n\n")

	// Table header
	sb.WriteString("## RACI Matrix\n\n")
	sb.WriteString("| Activity | ")
	for _, role := range matrix.Roles {
		sb.WriteString(fmt.Sprintf("%s | ", role))
	}
	sb.WriteString("\n")

	// Table separator
	sb.WriteString("|----------|")
	for range matrix.Roles {
		sb.WriteString("---------|")
	}
	sb.WriteString("\n")

	// Table rows
	for _, activity := range matrix.Activities {
		sb.WriteString(fmt.Sprintf("| **%s**", activity.Name))
		if activity.Description != "" {
			sb.WriteString(fmt.Sprintf("<br/>_%s_", activity.Description))
		}
		sb.WriteString(" | ")

		for _, role := range matrix.Roles {
			if raciRole, exists := activity.Assignments[role]; exists {
				sb.WriteString(fmt.Sprintf("%s | ", raciRole))
			} else {
				sb.WriteString("- | ")
			}
		}
		sb.WriteString("\n")
	}

	// Validation status
	sb.WriteString(fmt.Sprintf("\n---\n\n**Validation Status**: %s\n", map[bool]string{true: "✅ Valid", false: "❌ Invalid"}[matrix.IsValid]))
	if !matrix.IsValid && len(matrix.Errors) > 0 {
		sb.WriteString("\n**Validation Errors**:\n")
		for _, err := range matrix.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	return sb.String(), nil
}

// getBuiltInTemplates returns the built-in RACI templates
func (s *RACIService) getBuiltInTemplates() []*RACITemplate {
	return []*RACITemplate{
		{
			Key:         "software-dev",
			Name:        "Software Development Project",
			Description: "Standard RACI matrix for software development projects",
			Category:    "Software Development",
			IsPublic:    true,
			Roles:       []string{"Project Manager", "Tech Lead", "Developer", "QA Engineer", "DevOps"},
			Activities: []RACIActivity{
				{ID: "1", Name: "Requirements Gathering", Description: "Collect and document project requirements", Order: 1, Assignments: map[string]RACIRole{"Project Manager": RACIAccountable, "Tech Lead": RACIConsulted, "Developer": RACIInformed}},
				{ID: "2", Name: "Design Architecture", Description: "Design system architecture", Order: 2, Assignments: map[string]RACIRole{"Tech Lead": RACIAccountable, "Developer": RACIConsulted, "Project Manager": RACIInformed}},
				{ID: "3", Name: "Implementation", Description: "Write code and implement features", Order: 3, Assignments: map[string]RACIRole{"Developer": RACIResponsible, "Tech Lead": RACIAccountable, "QA Engineer": RACIInformed}},
				{ID: "4", Name: "Code Review", Description: "Review and approve code changes", Order: 4, Assignments: map[string]RACIRole{"Tech Lead": RACIAccountable, "Developer": RACIConsulted}},
				{ID: "5", Name: "Testing", Description: "Execute test plans and report bugs", Order: 5, Assignments: map[string]RACIRole{"QA Engineer": RACIResponsible, "Tech Lead": RACIAccountable, "Developer": RACIConsulted}},
				{ID: "6", Name: "Deployment", Description: "Deploy to production environment", Order: 6, Assignments: map[string]RACIRole{"DevOps": RACIResponsible, "Tech Lead": RACIAccountable, "Project Manager": RACIInformed}},
			},
		},
		{
			Key:         "research-analysis",
			Name:        "Research & Analysis",
			Description: "RACI matrix for research and analysis projects",
			Category:    "Research",
			IsPublic:    true,
			Roles:       []string{"Research Lead", "Analyst", "Stakeholder", "Reviewer"},
			Activities: []RACIActivity{
				{ID: "1", Name: "Define Research Scope", Description: "Define research objectives and scope", Order: 1, Assignments: map[string]RACIRole{"Research Lead": RACIAccountable, "Stakeholder": RACIConsulted}},
				{ID: "2", Name: "Data Collection", Description: "Gather and organize data", Order: 2, Assignments: map[string]RACIRole{"Analyst": RACIResponsible, "Research Lead": RACIAccountable}},
				{ID: "3", Name: "Analysis", Description: "Analyze data and draw insights", Order: 3, Assignments: map[string]RACIRole{"Analyst": RACIResponsible, "Research Lead": RACIAccountable, "Reviewer": RACIConsulted}},
				{ID: "4", Name: "Report Writing", Description: "Document findings and recommendations", Order: 4, Assignments: map[string]RACIRole{"Research Lead": RACIAccountable, "Analyst": RACIResponsible, "Reviewer": RACIConsulted}},
				{ID: "5", Name: "Presentation", Description: "Present findings to stakeholders", Order: 5, Assignments: map[string]RACIRole{"Research Lead": RACIAccountable, "Stakeholder": RACIInformed}},
			},
		},
		{
			Key:         "infrastructure",
			Name:        "Infrastructure Deployment",
			Description: "RACI matrix for infrastructure and deployment projects",
			Category:    "Infrastructure",
			IsPublic:    true,
			Roles:       []string{"DevOps Lead", "System Admin", "Security Engineer", "Developer", "Manager"},
			Activities: []RACIActivity{
				{ID: "1", Name: "Infrastructure Planning", Description: "Plan infrastructure requirements", Order: 1, Assignments: map[string]RACIRole{"DevOps Lead": RACIAccountable, "System Admin": RACIConsulted, "Manager": RACIInformed}},
				{ID: "2", Name: "Security Review", Description: "Review security requirements", Order: 2, Assignments: map[string]RACIRole{"Security Engineer": RACIAccountable, "DevOps Lead": RACIConsulted}},
				{ID: "3", Name: "Infrastructure Setup", Description: "Configure servers and services", Order: 3, Assignments: map[string]RACIRole{"System Admin": RACIResponsible, "DevOps Lead": RACIAccountable}},
				{ID: "4", Name: "Deployment Configuration", Description: "Configure deployment pipelines", Order: 4, Assignments: map[string]RACIRole{"DevOps Lead": RACIResponsible, "Developer": RACIConsulted}},
				{ID: "5", Name: "Monitoring Setup", Description: "Set up monitoring and alerting", Order: 5, Assignments: map[string]RACIRole{"System Admin": RACIResponsible, "DevOps Lead": RACIAccountable}},
			},
		},
	}
}
