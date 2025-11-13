package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
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
func (b *WorkflowsBuilder) GenerateWorkflowsFromContext(ctx context.Context, ag *models.Agency, overview *models.Overview, workItems []models.WorkItem) ([]models.Workflow, error) {
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

Create ONLY 1 simple workflow that makes sense for this agency. Keep it minimal and focused:
- Maximum 5-7 nodes total
- Each workflow must have exactly 1 start and 1 end node
- Work items are connected in logical sequential order
- Avoid complex decision trees - use decision nodes sparingly
- NO parallel nodes unless absolutely essential
- Position nodes in a simple left-to-right flow (increment x by 200-250 for each step)
- Focus on the single most important workflow for this agency
- Return a JSON array with ONLY 1 workflow object`

	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.5,
		MaxTokens:   2500,
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
func (b *WorkflowsBuilder) GenerateWorkflowWithPrompt(ctx context.Context, ag *models.Agency, userPrompt string, workItems []models.WorkItem) (*models.Workflow, error) {
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
func (b *WorkflowsBuilder) RefineWorkflow(ctx context.Context, wf *models.Workflow, refinementPrompt string) (*models.Workflow, error) {
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
func (b *WorkflowsBuilder) buildContextPrompt(ag *models.Agency, overview *models.Overview, workItems []models.WorkItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Agency: %s\n\n", ag.Name))

	if overview != nil && overview.Introduction != "" {
		sb.WriteString(fmt.Sprintf("Introduction: %s\n\n", overview.Introduction))
	}

	sb.WriteString("Available Work Items:\n")
	for _, wi := range workItems {
		sb.WriteString(fmt.Sprintf("- %s (key: %s): %s\n", wi.Title, wi.Key, wi.Description))
	}

	sb.WriteString("\nCreate workflows that orchestrate these work items in logical sequences.")

	return sb.String()
}

// buildPromptWithContext creates a prompt for user-requested workflow
func (b *WorkflowsBuilder) buildPromptWithContext(ag *models.Agency, userPrompt string, workItems []models.WorkItem) string {
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
func (b *WorkflowsBuilder) parseWorkflowsResponse(response string) ([]models.Workflow, error) {
	// Clean response
	cleaned := b.cleanJSONResponse(response)

	// Check for truncated JSON (common error)
	if !strings.HasSuffix(strings.TrimSpace(cleaned), "]") {
		b.logger.WithField("response_length", len(cleaned)).Warn("JSON response appears truncated - missing closing bracket")
		return nil, fmt.Errorf("invalid JSON response: response appears truncated (likely too large). Try creating fewer or simpler workflows")
	}

	var workflows []models.Workflow
	if err := json.Unmarshal([]byte(cleaned), &workflows); err != nil {
		b.logger.WithError(err).WithField("response", cleaned).Error("Failed to parse workflows JSON")

		// Provide more helpful error for truncated JSON
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return nil, fmt.Errorf("invalid JSON response: response was truncated (too large). Try generating fewer or simpler workflows")
		}

		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return workflows, nil
}

// parseSingleWorkflowResponse parses AI response into single workflow
func (b *WorkflowsBuilder) parseSingleWorkflowResponse(response string) (*models.Workflow, error) {
	// Clean response
	cleaned := b.cleanJSONResponse(response)

	var wf models.Workflow
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
func (b *WorkflowsBuilder) SuggestWorkflowImprovements(ctx context.Context, wf *models.Workflow) ([]string, error) {
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

// RefineWorkflows dynamically determines and executes the appropriate workflow operation based on user message
func (b *WorkflowsBuilder) RefineWorkflows(ctx context.Context, req *builder.RefineWorkflowsRequest, builderContext builder.BuilderContext) (*builder.RefineWorkflowsResponse, error) {
	b.logger.WithFields(logrus.Fields{
		"agency_id":          req.AgencyID,
		"user_message":       req.UserMessage,
		"target_workflows":   len(req.TargetWorkflows),
		"existing_workflows": len(req.ExistingWorkflows),
	}).Info("Starting dynamic workflow refinement")

	// Build the prompt to determine what action to take
	prompt := b.buildDynamicWorkflowsPrompt(req, builderContext)

	// Make the LLM request to determine action
	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicWorkflowsSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		b.logger.WithError(err).Error("Failed to get AI response for dynamic workflow refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var result builder.RefineWorkflowsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		b.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse dynamic workflows response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedWorkflows),
		"generated_count":  len(result.GeneratedWorkflows),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Dynamic workflow refinement completed")

	return &result, nil
}

// RefineWorkflowsStream performs dynamic workflow refinement with streaming support
// Similar to RefineWorkflows but streams chunks to the callback as they arrive from the LLM
func (b *WorkflowsBuilder) RefineWorkflowsStream(ctx context.Context, req *builder.RefineWorkflowsRequest, builderContext builder.BuilderContext, streamCallback StreamCallback) (*builder.RefineWorkflowsResponse, error) {
	b.logger.WithFields(logrus.Fields{
		"agency_id":          req.AgencyID,
		"user_message":       req.UserMessage,
		"target_workflows":   len(req.TargetWorkflows),
		"existing_workflows": len(req.ExistingWorkflows),
	}).Info("Starting streaming dynamic workflow refinement")

	// Build the prompt
	prompt := b.buildDynamicWorkflowsPrompt(req, builderContext)

	// Stream the LLM response
	var contentBuilder strings.Builder
	chunkCount := 0
	totalBytes := 0

	err := b.llmClient.ChatStream(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicWorkflowsSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}, func(chunk string) error {
		chunkCount++
		chunkBytes := len(chunk)
		totalBytes += chunkBytes

		// Log every 10 chunks
		if chunkCount%10 == 0 {
			b.logger.WithFields(logrus.Fields{
				"chunk_number": chunkCount,
				"chunk_bytes":  chunkBytes,
				"total_bytes":  totalBytes,
			}).Debug("Streaming workflow chunk received")
		}

		// Accumulate content for final parsing
		contentBuilder.WriteString(chunk)

		// Forward chunk to the callback (for SSE streaming)
		if streamCallback != nil {
			return streamCallback(chunk)
		}
		return nil
	})

	if err != nil {
		b.logger.WithError(err).WithFields(logrus.Fields{
			"total_chunks": chunkCount,
			"total_bytes":  totalBytes,
		}).Error("Failed to stream AI response for dynamic workflow refinement")
		return nil, fmt.Errorf("AI streaming refinement failed: %w", err)
	}

	// Parse the accumulated response
	fullContent := contentBuilder.String()
	contentLength := len(fullContent)

	b.logger.WithFields(logrus.Fields{
		"total_chunks":    chunkCount,
		"total_bytes":     totalBytes,
		"content_length":  contentLength,
		"content_preview": truncateString(fullContent, 100),
		"content_suffix":  getSuffix(fullContent, 100),
	}).Info("🔍 DEBUG: Completed streaming, parsing accumulated content")

	cleanedContent := stripMarkdownFences(fullContent)
	cleanedLength := len(cleanedContent)

	b.logger.WithFields(logrus.Fields{
		"original_length": contentLength,
		"cleaned_length":  cleanedLength,
		"bytes_removed":   contentLength - cleanedLength,
	}).Info("🔍 DEBUG: Content cleaned, attempting JSON parse")

	var result builder.RefineWorkflowsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		// Log detailed error info
		b.logger.WithError(err).WithFields(logrus.Fields{
			"content_length": cleanedLength,
			"content_start":  truncateString(cleanedContent, 200),
			"content_end":    getSuffix(cleanedContent, 200),
			"last_100_chars": getSuffix(cleanedContent, 100),
			"is_truncated":   !strings.HasSuffix(strings.TrimSpace(cleanedContent), "}"),
		}).Error("Failed to parse streamed workflows response")
		return nil, fmt.Errorf("failed to parse streamed response: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedWorkflows),
		"generated_count":  len(result.GeneratedWorkflows),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Streaming dynamic workflow refinement completed")

	return &result, nil
}

// buildDynamicWorkflowsPrompt creates the prompt for dynamic workflow processing
func (b *WorkflowsBuilder) buildDynamicWorkflowsPrompt(req *builder.RefineWorkflowsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\n\n### USER REQUEST\n")
	builder.WriteString(req.UserMessage)
	builder.WriteString("\n\n")

	if len(req.TargetWorkflows) > 0 {
		builder.WriteString("### TARGET WORKFLOWS FOR OPERATION\n")
		for _, workflow := range req.TargetWorkflows {
			builder.WriteString(fmt.Sprintf("- **%s** (v%s): %s\n", workflow.Name, workflow.Version, workflow.Description))
			builder.WriteString(fmt.Sprintf("  Nodes: %d, Edges: %d\n", len(workflow.Nodes), len(workflow.Edges)))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Based on the user's request and the agency context, determine what needs to be done with the workflows and execute the appropriate action.")

	return builder.String()
}

// System prompts for workflow operations
const dynamicWorkflowsSystemPrompt = SharedAgencyContext + `

Act as a strategic workflow management AI. Modify workflows based on user requests.

CRITICAL: Workflows ORCHESTRATE work item sequences. They define how work items connect and flow.

**IMPORTANT: When generating workflows, create a MAXIMUM of 3 workflows to keep response sizes manageable.**

## Actions:
**remove** - Delete workflows (return in consolidated_data.removed_workflows)
**refine** - Improve existing workflow structures and connections
**generate** - Create new workflows from work items and goals (MAX 3 workflows)
**consolidate** - Merge duplicate or overlapping workflows
**enhance_all** - Refine all workflows
**no_action** - Already optimal

## Response JSON:
{
  "action": "remove|refine|generate|consolidate|enhance_all|no_action",
  "refined_workflows": [{"original_key": "key", "refined_name": "...", "refined_description": "...", "refined_nodes": [...], "refined_edges": [...], "was_changed": true, "explanation": "Brief"}],
  "generated_workflows": [{"name": "...", "description": "...", "version": "1.0.0", "nodes": [...], "edges": [...], "explanation": "Brief"}],
  "consolidated_data": {"consolidated_workflows": [...], "removed_workflows": ["key1"], "summary": "Brief", "explanation": "Brief"},
  "explanation": "Brief overall summary",
  "no_action_needed": false
}

Guidelines:
- Workflows = ORCHESTRATION (how work items connect), not individual work items
- Each workflow should connect 3-7 work items in a logical sequence
- **Generate MAXIMUM 3 workflows** to avoid response truncation
- Start with start node, end with end node
- Use decision nodes for conditional branching
- Align with goals: Each workflow should support agency objectives
- Keep explanations concise (1-2 sentences)
- Node positions should flow left-to-right (increment x by 200-250)
- Prioritize the most important workflows that cover core agency operations
`

// truncateString returns the first n characters of a string, or the full string if shorter
func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// getSuffix returns the last n characters of a string, or the full string if shorter
func getSuffix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}
