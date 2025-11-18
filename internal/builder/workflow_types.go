package builder

import (
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// RefineWorkflowRequest contains the context for refining a workflow
type RefineWorkflowRequest struct {
	AgencyID          string             `json:"agency_id"`
	CurrentWorkflow   *models.Workflow   `json:"current_workflow"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	ExistingWorkflows []*models.Workflow `json:"existing_workflows"`
	Goals             []*models.Goal     `json:"goals"`
	WorkItems         []*models.WorkItem `json:"work_items"`
	AgencyContext     *models.Agency     `json:"agency_context"`
}

// RefineWorkflowResponse contains the AI-refined workflow
type RefineWorkflowResponse struct {
	RefinedName        string                `json:"refined_name"`
	RefinedDescription string                `json:"refined_description"`
	RefinedNodes       []models.WorkflowNode `json:"refined_nodes"`
	RefinedEdges       []models.WorkflowEdge `json:"refined_edges"`
	WasChanged         bool                  `json:"was_changed"`
	Explanation        string                `json:"explanation"`
}

// GenerateWorkflowRequest contains the context for generating a new workflow
type GenerateWorkflowRequest struct {
	AgencyID          string             `json:"agency_id"`
	AgencyContext     *models.Agency     `json:"agency_context"`
	ExistingWorkflows []*models.Workflow `json:"existing_workflows"`
	Goals             []*models.Goal     `json:"goals"`
	WorkItems         []*models.WorkItem `json:"work_items"`
	UserInput         string             `json:"user_input"`
}

// GenerateWorkflowResponse contains the AI-generated workflow
type GenerateWorkflowResponse struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Version     string                `json:"version"`
	Nodes       []models.WorkflowNode `json:"nodes"`
	Edges       []models.WorkflowEdge `json:"edges"`
	Explanation string                `json:"explanation"`
}

// GenerateWorkflowsResponse contains multiple AI-generated workflows
type GenerateWorkflowsResponse struct {
	Workflows   []GenerateWorkflowResponse `json:"workflows"`
	Explanation string                     `json:"explanation"`
}

// ConsolidateWorkflowsRequest contains the context for consolidating workflows
type ConsolidateWorkflowsRequest struct {
	AgencyID         string             `json:"agency_id"`
	AgencyContext    *models.Agency     `json:"agency_context"`
	CurrentWorkflows []*models.Workflow `json:"current_workflows"`
	Goals            []*models.Goal     `json:"goals"`
	WorkItems        []*models.WorkItem `json:"work_items"`
}

// ConsolidateWorkflowsResponse contains the consolidated workflows
type ConsolidateWorkflowsResponse struct {
	ConsolidatedWorkflows []ConsolidatedWorkflow `json:"consolidated_workflows"`
	RemovedWorkflows      []string               `json:"removed_workflows"` // Keys of workflows that were consolidated/removed
	Summary               string                 `json:"summary"`
	Explanation           string                 `json:"explanation"`
}

// ConsolidatedWorkflow represents a workflow after consolidation
type ConsolidatedWorkflow struct {
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Version          string                `json:"version"`
	Nodes            []models.WorkflowNode `json:"nodes"`
	Edges            []models.WorkflowEdge `json:"edges"`
	ConsolidatedFrom []string              `json:"consolidated_from"` // Keys of original workflows
	Rationale        string                `json:"rationale"`
}

// RefineWorkflowsRequest contains the context for dynamic workflow processing
type RefineWorkflowsRequest struct {
	AgencyID          string             `json:"agency_id"`
	UserMessage       string             `json:"user_message"`
	TargetWorkflows   []*models.Workflow `json:"target_workflows"`   // Specific workflows to operate on (nil means all)
	ExistingWorkflows []*models.Workflow `json:"existing_workflows"` // All current workflows for context
	Goals             []*models.Goal     `json:"goals"`              // Agency goals for context
	WorkItems         []*models.WorkItem `json:"work_items"`         // Work items for workflow generation context
	AgencyContext     *models.Agency     `json:"agency_context"`
}

// RefineWorkflowsResponse contains the dynamic workflow processing results
type RefineWorkflowsResponse struct {
	Action             string                        `json:"action"`              // What action was determined (refine, generate, consolidate, enhance_all, etc.)
	RefinedWorkflows   []RefinedWorkflowResult       `json:"refined_workflows"`   // Workflows that were refined
	GeneratedWorkflows []GenerateWorkflowResponse    `json:"generated_workflows"` // Newly generated workflows
	ConsolidatedData   *ConsolidateWorkflowsResponse `json:"consolidated_data"`   // Consolidation results if applicable
	Explanation        string                        `json:"explanation"`         // What was done and why
	NoActionNeeded     bool                          `json:"no_action_needed"`    // True if workflows are already optimal
}

// RefinedWorkflowResult represents a single refined workflow
type RefinedWorkflowResult struct {
	OriginalKey        string                `json:"original_key"`
	RefinedName        string                `json:"refined_name"`
	RefinedDescription string                `json:"refined_description"`
	RefinedNodes       []models.WorkflowNode `json:"refined_nodes"`
	RefinedEdges       []models.WorkflowEdge `json:"refined_edges"`
	WasChanged         bool                  `json:"was_changed"`
	Explanation        string                `json:"explanation"`
}
