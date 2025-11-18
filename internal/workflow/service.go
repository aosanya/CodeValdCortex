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
		Nodes:       make([]models.WorkflowNode, len(original.Nodes)),
		Edges:       make([]models.WorkflowEdge, len(original.Edges)),
		AgencyID:    original.AgencyID,
		CreatedBy:   original.CreatedBy,
	}

	// Deep copy nodes
	copy(duplicate.Nodes, original.Nodes)

	// Deep copy edges
	copy(duplicate.Edges, original.Edges)

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

	// Validate nodes
	if len(workflow.Nodes) > 0 {
		s.validateNodes(workflow, result)
	}

	// Validate edges
	if len(workflow.Edges) > 0 {
		s.validateEdges(workflow, result)
	}

	return result
}

// validateNodes validates all nodes in the workflow
func (s *Service) validateNodes(workflow *models.Workflow, result *models.WorkflowValidationResult) {
	nodeIDs := make(map[string]bool)
	hasStart := false
	hasEnd := false

	for _, node := range workflow.Nodes {
		// Check for duplicate node IDs
		if nodeIDs[node.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Duplicate node ID: %s", node.ID),
				NodeID:  node.ID,
			})
			continue
		}
		nodeIDs[node.ID] = true

		// Validate node type - only work_item nodes are supported
		if node.Type != models.NodeTypeWorkItem {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Invalid node type: %s. Only 'work_item' nodes are supported", node.Type),
				NodeID:  node.ID,
			})
		}

		// Validate node-specific data
		if node.Type == models.NodeTypeWorkItem {
			if node.Data.WorkItemKey == "" {
				result.Valid = false
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "nodes",
					Message: "Work item node must have work_item_key",
					NodeID:  node.ID,
				})
			}
		}
	}

	// Check for required start and end nodes if there are multiple nodes
	if len(workflow.Nodes) > 1 {
		if !hasStart {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "nodes",
				Message: "models.Workflow must have a start node",
			})
		}
		if !hasEnd {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "nodes",
				Message: "models.Workflow must have an end node",
			})
		}
	}
}

// validateEdges validates all edges in the workflow
func (s *Service) validateEdges(workflow *models.Workflow, result *models.WorkflowValidationResult) {
	// Build node ID map for validation
	nodeIDs := make(map[string]bool)
	for _, node := range workflow.Nodes {
		nodeIDs[node.ID] = true
	}

	edgeIDs := make(map[string]bool)

	for _, edge := range workflow.Edges {
		// Check for duplicate edge IDs
		if edgeIDs[edge.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("Duplicate edge ID: %s", edge.ID),
				EdgeID:  edge.ID,
			})
			continue
		}
		edgeIDs[edge.ID] = true

		// Validate source node exists
		if !nodeIDs[edge.Source] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("models.WorkflowEdge source node not found: %s", edge.Source),
				EdgeID:  edge.ID,
			})
		}

		// Validate target node exists
		if !nodeIDs[edge.Target] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("models.WorkflowEdge target node not found: %s", edge.Target),
				EdgeID:  edge.ID,
			})
		}
	}

	// Validate edges
	if len(workflow.Edges) > 0 {
		s.checkOrphanedNodes(workflow, result)
	}
}

// checkOrphanedNodes checks for nodes that are not connected
func (s *Service) checkOrphanedNodes(workflow *models.Workflow, result *models.WorkflowValidationResult) {
	hasIncoming := make(map[string]bool)
	hasOutgoing := make(map[string]bool)

	for _, edge := range workflow.Edges {
		hasOutgoing[edge.Source] = true
		hasIncoming[edge.Target] = true
	}

	for _, node := range workflow.Nodes {
		// All work item nodes should have connections
		if !hasIncoming[node.ID] && !hasOutgoing[node.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Orphaned node (no connections): %s", node.ID),
				NodeID:  node.ID,
			})
		}
	}
}
