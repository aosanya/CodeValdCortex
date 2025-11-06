package workflow

import (
	"context"
	"fmt"
	"strings"

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
func (s *Service) CreateWorkflow(ctx context.Context, workflow *Workflow) error {
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
func (s *Service) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
	return s.repo.GetByID(ctx, id)
}

// GetWorkflowsByAgency retrieves all workflows for an agency
func (s *Service) GetWorkflowsByAgency(ctx context.Context, agencyID string) ([]*Workflow, error) {
	return s.repo.GetByAgencyID(ctx, agencyID)
}

// UpdateWorkflow updates an existing workflow with validation
func (s *Service) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
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

// DeleteWorkflow deletes a workflow
func (s *Service) DeleteWorkflow(ctx context.Context, id string) error {
	// Check for active executions
	executions, err := s.repo.GetExecutionsByWorkflowID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check executions: %w", err)
	}

	for _, exec := range executions {
		if exec.Status == WorkflowStatusActive {
			return fmt.Errorf("cannot delete workflow with active executions")
		}
	}

	// Delete workflow
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// DuplicateWorkflow creates a copy of an existing workflow
func (s *Service) DuplicateWorkflow(ctx context.Context, id string) (*Workflow, error) {
	// Get original workflow
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original workflow: %w", err)
	}

	// Create duplicate
	duplicate := &Workflow{
		Name:        original.Name + " (Copy)",
		Version:     "1.0.0",
		Description: original.Description,
		Status:      WorkflowStatusDraft,
		Nodes:       make([]Node, len(original.Nodes)),
		Edges:       make([]Edge, len(original.Edges)),
		Variables:   make(map[string]interface{}),
		AgencyID:    original.AgencyID,
		CreatedBy:   original.CreatedBy,
	}

	// Deep copy nodes
	copy(duplicate.Nodes, original.Nodes)

	// Deep copy edges
	copy(duplicate.Edges, original.Edges)

	// Deep copy variables
	for k, v := range original.Variables {
		duplicate.Variables[k] = v
	}

	// Create the duplicate
	if err := s.CreateWorkflow(ctx, duplicate); err != nil {
		return nil, fmt.Errorf("failed to create duplicate: %w", err)
	}

	return duplicate, nil
}

// ListWorkflows retrieves workflows with pagination
func (s *Service) ListWorkflows(ctx context.Context, limit, offset int) ([]*Workflow, error) {
	return s.repo.List(ctx, limit, offset)
}

// ValidateWorkflow validates a workflow definition
func (s *Service) ValidateWorkflow(workflow *Workflow) error {
	result := s.ValidateWorkflowStructure(workflow)
	if !result.Valid {
		return fmt.Errorf("workflow validation failed: %d errors", len(result.Errors))
	}
	return nil
}

// ValidateWorkflowStructure performs comprehensive validation and returns detailed results
func (s *Service) ValidateWorkflowStructure(workflow *Workflow) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Validate name
	if strings.TrimSpace(workflow.Name) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "Workflow name is required",
		})
	} else if len(workflow.Name) < 3 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "Workflow name must be at least 3 characters",
		})
	}

	// Validate version format (semantic versioning)
	if workflow.Version != "" {
		parts := strings.Split(workflow.Version, ".")
		if len(parts) != 3 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "version",
				Message: "Version must be in semantic versioning format (x.y.z)",
			})
		}
	}

	// Validate status
	validStatuses := map[WorkflowStatus]bool{
		WorkflowStatusDraft:     true,
		WorkflowStatusActive:    true,
		WorkflowStatusPaused:    true,
		WorkflowStatusCompleted: true,
		WorkflowStatusFailed:    true,
	}
	if !validStatuses[workflow.Status] {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("Invalid status: %s", workflow.Status),
		})
	}

	// Validate agency_id
	if strings.TrimSpace(workflow.AgencyID) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
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
func (s *Service) validateNodes(workflow *Workflow, result *ValidationResult) {
	nodeIDs := make(map[string]bool)
	hasStart := false
	hasEnd := false

	for _, node := range workflow.Nodes {
		// Check for duplicate node IDs
		if nodeIDs[node.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Duplicate node ID: %s", node.ID),
				NodeID:  node.ID,
			})
			continue
		}
		nodeIDs[node.ID] = true

		// Track start and end nodes
		if node.Type == NodeTypeStart {
			hasStart = true
		}
		if node.Type == NodeTypeEnd {
			hasEnd = true
		}

		// Validate node type
		validNodeTypes := map[NodeType]bool{
			NodeTypeStart:    true,
			NodeTypeWorkItem: true,
			NodeTypeDecision: true,
			NodeTypeParallel: true,
			NodeTypeEnd:      true,
		}
		if !validNodeTypes[node.Type] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Invalid node type: %s", node.Type),
				NodeID:  node.ID,
			})
		}

		// Validate node-specific data
		if node.Type == NodeTypeWorkItem {
			if node.Data.WorkItemID == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "nodes",
					Message: "Work item node must have work_item_id",
					NodeID:  node.ID,
				})
			}
		}
	}

	// Check for required start and end nodes if there are multiple nodes
	if len(workflow.Nodes) > 1 {
		if !hasStart {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "nodes",
				Message: "Workflow must have a start node",
			})
		}
		if !hasEnd {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "nodes",
				Message: "Workflow must have an end node",
			})
		}
	}
}

// validateEdges validates all edges in the workflow
func (s *Service) validateEdges(workflow *Workflow, result *ValidationResult) {
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
			result.Errors = append(result.Errors, ValidationError{
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
			result.Errors = append(result.Errors, ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("Edge source node not found: %s", edge.Source),
				EdgeID:  edge.ID,
			})
		}

		// Validate target node exists
		if !nodeIDs[edge.Target] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("Edge target node not found: %s", edge.Target),
				EdgeID:  edge.ID,
			})
		}

		// Validate edge type
		validEdgeTypes := map[EdgeType]bool{
			EdgeTypeSequential:  true,
			EdgeTypeConditional: true,
			EdgeTypeDataFlow:    true,
		}
		if !validEdgeTypes[edge.Type] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "edges",
				Message: fmt.Sprintf("Invalid edge type: %s", edge.Type),
				EdgeID:  edge.ID,
			})
		}

		// Validate conditional edges have conditions
		if edge.Type == EdgeTypeConditional && edge.Data.Condition == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "edges",
				Message: "Conditional edge must have a condition",
				EdgeID:  edge.ID,
			})
		}
	}

	// Check for orphaned nodes (nodes with no incoming or outgoing edges)
	if len(workflow.Edges) > 0 {
		s.checkOrphanedNodes(workflow, result)
	}
}

// checkOrphanedNodes checks for nodes that are not connected
func (s *Service) checkOrphanedNodes(workflow *Workflow, result *ValidationResult) {
	hasIncoming := make(map[string]bool)
	hasOutgoing := make(map[string]bool)

	for _, edge := range workflow.Edges {
		hasOutgoing[edge.Source] = true
		hasIncoming[edge.Target] = true
	}

	for _, node := range workflow.Nodes {
		// Start nodes don't need incoming edges
		if node.Type == NodeTypeStart {
			continue
		}
		// End nodes don't need outgoing edges
		if node.Type == NodeTypeEnd {
			continue
		}

		// Other nodes should have both
		if !hasIncoming[node.ID] && !hasOutgoing[node.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "nodes",
				Message: fmt.Sprintf("Orphaned node (no connections): %s", node.ID),
				NodeID:  node.ID,
			})
		}
	}
}

// StartExecution starts a new workflow execution
func (s *Service) StartExecution(ctx context.Context, workflowID, startedBy string, context map[string]interface{}) (*WorkflowExecution, error) {
	// Get workflow
	workflow, err := s.repo.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Validate workflow
	if err := s.ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Create execution
	execution := &WorkflowExecution{
		WorkflowID:      workflowID,
		WorkflowVersion: workflow.Version,
		Status:          WorkflowStatusActive,
		StartedBy:       startedBy,
		Context:         context,
		NodeExecutions:  []NodeExecution{},
		Errors:          []string{},
	}

	if err := s.repo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"workflow_id":  workflowID,
		"execution_id": execution.ID,
	}).Info("Started workflow execution")

	return execution, nil
}
