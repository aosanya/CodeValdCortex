package workflows

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/web/templates/pages/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles workflow-related HTTP requests
type Handler struct {
	db *database.ArangoClient
}

// NewHandler creates a new workflow handler
func NewHandler(db *database.ArangoClient) *Handler {
	return &Handler{db: db}
}

// ShowDesigner renders the workflow designer page
func (h *Handler) ShowDesigner(c *gin.Context) {
	agencyID := c.Param("agency_id")
	workflowID := c.Query("workflow_id")

	var workflow *models.Workflow

	if workflowID != "" {
		// Load existing workflow
		var err error
		workflow, err = h.getWorkflowByID(c.Request.Context(), workflowID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
			return
		}
	}

	// Render template
	component := workflows.WorkflowDesigner(workflow, agencyID)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
		return
	}
}

// ListWorkflows returns all workflows for an agency
func (h *Handler) ListWorkflows(c *gin.Context) {
	agencyID := c.Param("agency_id")

	query := `
		FOR w IN workflows
			FILTER w.agency_id == @agency_id
			SORT w.updated_at DESC
			RETURN w
	`

	cursor, err := h.db.Database().Query(context.Background(), query, map[string]interface{}{
		"agency_id": agencyID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query workflows"})
		return
	}
	defer cursor.Close()

	var workflowsList []models.Workflow
	for cursor.HasMore() {
		var workflow models.Workflow
		_, err := cursor.ReadDocument(context.Background(), &workflow)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read workflow"})
			return
		}
		workflowsList = append(workflowsList, workflow)
	}

	c.JSON(http.StatusOK, workflowsList)
}

// GetWorkflow returns a specific workflow by ID
func (h *Handler) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	workflow, err := h.getWorkflowByID(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// CreateWorkflow creates a new workflow
func (h *Handler) CreateWorkflow(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var workflow models.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set metadata
	workflow.Key = uuid.New().String()
	workflow.AgencyID = agencyID
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	// TODO: Get from auth context
	workflow.CreatedBy = "system"
	workflow.UpdatedBy = "system"

	// Save to database
	collection, err := h.db.Database().Collection(context.Background(), "workflows")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access workflows collection"})
		return
	}

	meta, err := collection.CreateDocument(context.Background(), workflow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	workflow.ID = meta.ID.String()
	workflow.Rev = meta.Rev

	c.JSON(http.StatusCreated, workflow)
}

// UpdateWorkflow updates an existing workflow
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	var workflow models.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update metadata
	workflow.UpdatedAt = time.Now()
	workflow.UpdatedBy = "system" // TODO: Get from auth context

	// Save to database
	collection, err := h.db.Database().Collection(context.Background(), "workflows")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access workflows collection"})
		return
	}

	meta, err := collection.UpdateDocument(context.Background(), workflowID, workflow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
		return
	}

	workflow.Rev = meta.Rev

	c.JSON(http.StatusOK, workflow)
}

// DeleteWorkflow deletes a workflow
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	collection, err := h.db.Database().Collection(context.Background(), "workflows")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access workflows collection"})
		return
	}

	_, err = collection.RemoveDocument(context.Background(), workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ValidateWorkflow validates a workflow structure
func (h *Handler) ValidateWorkflow(c *gin.Context) {
	var workflow models.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.validateWorkflowStructure(&workflow)
	c.JSON(http.StatusOK, result)
}

// Helper functions

func (h *Handler) getWorkflowByID(ctx context.Context, workflowID string) (*models.Workflow, error) {
	collection, err := h.db.Database().Collection(ctx, "workflows")
	if err != nil {
		return nil, err
	}

	var workflow models.Workflow
	_, err = collection.ReadDocument(ctx, workflowID, &workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (h *Handler) validateWorkflowStructure(workflow *models.Workflow) models.WorkflowValidationResult {
	result := models.WorkflowValidationResult{
		Valid:    true,
		Errors:   []models.ValidationError{},
		Warnings: []string{},
	}

	// Check if workflow has nodes
	if len(workflow.Nodes) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, models.ValidationError{
			Field:   "nodes",
			Message: "Workflow must have at least one node",
		})
		return result
	}

	// Check for cycles
	if h.hasCycle(workflow) {
		result.Valid = false
		result.Errors = append(result.Errors, models.ValidationError{
			Field:   "edges",
			Message: "Workflow contains circular dependencies",
		})
	}

	// Check for orphaned nodes (except root nodes)
	orphanedNodes := h.findOrphanedNodes(workflow)
	if len(orphanedNodes) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Found %d orphaned nodes that have no incoming or outgoing edges", len(orphanedNodes)))
	}

	return result
}

func (h *Handler) hasCycle(workflow *models.Workflow) bool {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, edge := range workflow.Edges {
		graph[edge.Source] = append(graph[edge.Source], edge.Target)
	}

	// Track visited nodes
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// DFS for cycle detection
	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Check all nodes
	for _, node := range workflow.Nodes {
		if !visited[node.ID] {
			if dfs(node.ID) {
				return true
			}
		}
	}

	return false
}

func (h *Handler) findOrphanedNodes(workflow *models.Workflow) []string {
	// Build set of nodes with edges
	hasEdge := make(map[string]bool)
	for _, edge := range workflow.Edges {
		hasEdge[edge.Source] = true
		hasEdge[edge.Target] = true
	}

	// Find orphaned nodes
	orphaned := []string{}
	for _, node := range workflow.Nodes {
		if !hasEdge[node.ID] && len(workflow.Edges) > 0 {
			orphaned = append(orphaned, node.ID)
		}
	}

	return orphaned
}
