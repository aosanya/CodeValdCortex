package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/sirupsen/logrus"
)

// WorkflowsBuilder handles AI-powered workflow generation and refinement
type WorkflowsBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewAIWorkflowsBuilder creates a new workflow builder with AI capabilities
func NewAIWorkflowsBuilder(llmClient LLMClient, logger *logrus.Logger) *WorkflowsBuilder {
	return &WorkflowsBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// GenerateWorkflowsFromContext generates workflow suggestions based on agency context
func (b *WorkflowsBuilder) GenerateWorkflowsFromContext(ctx context.Context, ag *agency.Agency, overview *agency.Overview, workItems []agency.WorkItem) ([]workflow.Workflow, error) {
	prompt := b.buildContextPrompt(ag, overview, workItems)

	systemPrompt := `You are an expert workflow architect specializing in designing efficient work item orchestration flows.
Your task is to analyze the agency's work items and create logical workflows that connect them.

Return ONLY a valid JSON array of workflows. Each workflow should follow this structure:
{
	"name": "workflow name",
	"description": "detailed description",
	"version": "1.0.0",
	"status": "draft",
	"nodes": [
		{
			"id": "node_1",
			"type": "start|work_item|decision|parallel|end",
			"position": {"x": 100, "y": 100},
			"data": {
				"name": "node name",
				"work_item_id": "work_item_key" (for work_item type),
				"condition": "condition expression" (for decision type),
				"trigger": "manual|scheduled|event" (for start type),
				"status": "success|failure" (for end type)
			}
		}
	],
	"edges": [
		{
			"id": "edge_1",
			"source": "node_1",
			"target": "node_2",
			"type": "sequential|conditional|dataflow",
			"data": {
				"condition": "optional condition",
				"label": "optional label"
			}
		}
	],
	"variables": {}
}

Create 2-3 workflows that make sense for this agency. Ensure:
- Each workflow has a clear start and end
- Work items are connected in logical order
- Use decision nodes for branching logic
- Use parallel nodes for concurrent tasks
- Position nodes in a left-to-right flow (increment x by 200 for each step)`

	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflows: %w", err)
	}

	// Parse response
	workflows, err := b.parseWorkflowsResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflows: %w", err)
	}

	b.logger.WithField("count", len(workflows)).Info("Generated workflows from context")
	return workflows, nil
}

// GenerateWorkflowWithPrompt generates a workflow based on user's natural language prompt
func (b *WorkflowsBuilder) GenerateWorkflowWithPrompt(ctx context.Context, ag *agency.Agency, userPrompt string, workItems []agency.WorkItem) (*workflow.Workflow, error) {
	prompt := b.buildPromptWithContext(ag, userPrompt, workItems)

	systemPrompt := `You are an expert workflow designer. Based on the user's request and available work items, create a single workflow.

Return ONLY a valid JSON object (not an array) with this structure:
{
	"name": "workflow name",
	"description": "detailed description",
	"version": "1.0.0",
	"status": "draft",
	"nodes": [...],
	"edges": [...],
	"variables": {}
}

Ensure the workflow:
- Has a clear start and end node
- Uses available work items logically
- Includes appropriate decision points
- Has proper sequential or conditional connections
- Positions nodes for left-to-right flow`

	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow: %w", err)
	}

	// Parse single workflow
	wf, err := b.parseSingleWorkflowResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	b.logger.WithField("workflow", wf.Name).Info("Generated workflow from prompt")
	return wf, nil
}

// RefineWorkflow refines an existing workflow based on user feedback
func (b *WorkflowsBuilder) RefineWorkflow(ctx context.Context, wf *workflow.Workflow, refinementPrompt string) (*workflow.Workflow, error) {
	currentJSON, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal current workflow: %w", err)
	}

	prompt := fmt.Sprintf(`Current workflow:
%s

User request: %s

Modify the workflow according to the user's request while maintaining valid structure.`, string(currentJSON), refinementPrompt)

	systemPrompt := `You are an expert workflow designer. Refine the given workflow based on the user's feedback.

Return ONLY the modified workflow as valid JSON (not an array).
Preserve the workflow structure and ensure all connections remain valid.`

	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.5,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to refine workflow: %w", err)
	}

	refined, err := b.parseSingleWorkflowResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refined workflow: %w", err)
	}

	// Preserve original IDs
	refined.ID = wf.ID
	refined.AgencyID = wf.AgencyID
	refined.CreatedBy = wf.CreatedBy
	refined.CreatedAt = wf.CreatedAt

	b.logger.WithField("workflow", refined.Name).Info("Refined workflow")
	return refined, nil
}

// buildContextPrompt creates a prompt with agency context
func (b *WorkflowsBuilder) buildContextPrompt(ag *agency.Agency, overview *agency.Overview, workItems []agency.WorkItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Agency: %s\n\n", ag.Name))

	if overview != nil && overview.Introduction != "" {
		sb.WriteString(fmt.Sprintf("Introduction: %s\n\n", overview.Introduction))
	}

	sb.WriteString("Available Work Items:\n")
	for _, wi := range workItems {
		sb.WriteString(fmt.Sprintf("- %s (key: %s): %s\n", wi.Title, wi.Key, wi.Description))
		if len(wi.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("  Dependencies: %v\n", wi.Dependencies))
		}
	}

	sb.WriteString("\nCreate workflows that orchestrate these work items in logical sequences.")

	return sb.String()
}

// buildPromptWithContext creates a prompt for user-requested workflow
func (b *WorkflowsBuilder) buildPromptWithContext(ag *agency.Agency, userPrompt string, workItems []agency.WorkItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Agency: %s\n\n", ag.Name))
	sb.WriteString(fmt.Sprintf("User Request: %s\n\n", userPrompt))

	sb.WriteString("Available Work Items:\n")
	for _, wi := range workItems {
		sb.WriteString(fmt.Sprintf("- %s (key: %s): %s\n", wi.Title, wi.Key, wi.Description))
	}

	return sb.String()
}

// parseWorkflowsResponse parses AI response into workflow array
func (b *WorkflowsBuilder) parseWorkflowsResponse(response string) ([]workflow.Workflow, error) {
	// Clean response
	cleaned := b.cleanJSONResponse(response)

	var workflows []workflow.Workflow
	if err := json.Unmarshal([]byte(cleaned), &workflows); err != nil {
		b.logger.WithError(err).WithField("response", cleaned).Error("Failed to parse workflows JSON")
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return workflows, nil
}

// parseSingleWorkflowResponse parses AI response into single workflow
func (b *WorkflowsBuilder) parseSingleWorkflowResponse(response string) (*workflow.Workflow, error) {
	// Clean response
	cleaned := b.cleanJSONResponse(response)

	var wf workflow.Workflow
	if err := json.Unmarshal([]byte(cleaned), &wf); err != nil {
		b.logger.WithError(err).WithField("response", cleaned).Error("Failed to parse workflow JSON")
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return &wf, nil
}

// cleanJSONResponse removes markdown code blocks and extra whitespace
func (b *WorkflowsBuilder) cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// SuggestWorkflowImprovements suggests improvements for an existing workflow
func (b *WorkflowsBuilder) SuggestWorkflowImprovements(ctx context.Context, wf *workflow.Workflow) ([]string, error) {
	currentJSON, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze this workflow and suggest improvements:
%s

Provide 3-5 specific, actionable suggestions as a JSON array of strings.
Focus on: efficiency, error handling, parallel execution opportunities, better branching logic.`, string(currentJSON))

	systemPrompt := `You are a workflow optimization expert. Analyze workflows and suggest concrete improvements.
Return ONLY a JSON array of suggestion strings. Example: ["Add error handling after step X", "Parallelize tasks Y and Z"]`

	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.6,
		MaxTokens:   2000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate suggestions: %w", err)
	}

	cleaned := b.cleanJSONResponse(response.Content)

	var suggestions []string
	if err := json.Unmarshal([]byte(cleaned), &suggestions); err != nil {
		b.logger.WithError(err).Error("Failed to parse suggestions")
		return nil, fmt.Errorf("invalid suggestions response: %w", err)
	}

	return suggestions, nil
}
