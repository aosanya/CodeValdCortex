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

// Compile-time check to ensure AIWorkItemsBuilder implements WorkItemBuilderInterface
var _ builder.WorkItemBuilderInterface = (*WorkItemsBuilder)(nil)

// WorkItemsBuilder handles AI-powered work item definition and refinement
type WorkItemsBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewAIWorkItemsBuilder creates a new AI-powered work item builder
func NewAIWorkItemsBuilder(llmClient LLMClient, logger *logrus.Logger) *WorkItemsBuilder {
	return &WorkItemsBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineWorkItems dynamically determines and executes the appropriate work item operation based on user message
func (w *WorkItemsBuilder) RefineWorkItems(ctx context.Context, req *builder.RefineWorkItemsRequest, builderContext builder.BuilderContext) (*builder.RefineWorkItemsResponse, error) {
	w.logger.WithFields(logrus.Fields{
		"agency_id":           req.AgencyID,
		"user_message":        req.UserMessage,
		"target_work_items":   len(req.TargetWorkItems),
		"existing_work_items": len(req.ExistingWorkItems),
	}).Info("Starting dynamic work item refinement")

	// Build the prompt to determine what action to take
	prompt := w.buildDynamicWorkItemsPrompt(req, builderContext)

	// Make the LLM request to determine action
	response, err := w.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicWorkItemsSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		w.logger.WithError(err).Error("Failed to get AI response for dynamic work item refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var result builder.RefineWorkItemsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		w.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse dynamic work items response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	w.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedWorkItems),
		"generated_count":  len(result.GeneratedWorkItems),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Dynamic work item refinement completed")

	return &result, nil
}

// RefineWorkItemsStream performs dynamic work item refinement with streaming support
// Similar to RefineWorkItems but streams chunks to the callback as they arrive from the LLM
func (w *WorkItemsBuilder) RefineWorkItemsStream(ctx context.Context, req *builder.RefineWorkItemsRequest, builderContext builder.BuilderContext, streamCallback builder.StreamCallback) (*builder.RefineWorkItemsResponse, error) {
	w.logger.WithFields(logrus.Fields{
		"agency_id":           req.AgencyID,
		"user_message":        req.UserMessage,
		"target_work_items":   len(req.TargetWorkItems),
		"existing_work_items": len(req.ExistingWorkItems),
	}).Info("Starting streaming dynamic work item refinement")

	// Build the prompt
	prompt := w.buildDynamicWorkItemsPrompt(req, builderContext)

	// Stream the LLM response
	var contentBuilder strings.Builder
	err := w.llmClient.ChatStream(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: dynamicWorkItemsSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}, func(chunk string) error {
		// Accumulate content for final parsing
		contentBuilder.WriteString(chunk)

		// Forward chunk to the callback (for SSE streaming)
		if streamCallback != nil {
			return streamCallback(chunk)
		}
		return nil
	})

	if err != nil {
		w.logger.WithError(err).Error("Failed to stream AI response for dynamic work item refinement")
		return nil, fmt.Errorf("AI streaming refinement failed: %w", err)
	}

	// Parse the accumulated response
	fullContent := contentBuilder.String()
	cleanedContent := stripMarkdownFences(fullContent)

	var result builder.RefineWorkItemsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		w.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse streamed work items response")
		return nil, fmt.Errorf("failed to parse streamed response: %w", err)
	}

	w.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedWorkItems),
		"generated_count":  len(result.GeneratedWorkItems),
		"no_action_needed": result.NoActionNeeded,
	}).Info("Streaming dynamic work item refinement completed")

	return &result, nil
}
func (w *WorkItemsBuilder) buildDynamicWorkItemsPrompt(req *builder.RefineWorkItemsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\n\n### USER REQUEST\n")
	builder.WriteString(req.UserMessage)
	builder.WriteString("\n\n")

	if len(req.TargetWorkItems) > 0 {
		builder.WriteString("### TARGET WORK ITEMS FOR OPERATION\n")
		for _, item := range req.TargetWorkItems {
			builder.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", item.Key, item.Code, item.Title))
			if item.Description != "" {
				builder.WriteString(fmt.Sprintf("  Description: %s\n", item.Description))
			}
			if len(item.DeliverablesStructured) > 0 {
				deliverablesJSON, err := json.MarshalIndent(item.DeliverablesStructured, "  ", "  ")
				if err == nil {
					builder.WriteString("  Deliverables Structure (JSON):\n  ")
					builder.WriteString(string(deliverablesJSON))
					builder.WriteString("\n")
				}
			}
			if len(item.Tags) > 0 {
				builder.WriteString(fmt.Sprintf("  Tags: %v\n", item.Tags))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("Based on the user's request and the agency context, determine what needs to be done with the work items and execute the appropriate action.")

	return builder.String()
}

// aiWorkItemRefinementResponse represents the JSON structure returned by the AI
type aiWorkItemRefinementResponse struct {
	Title                  string                   `json:"title"`
	Description            string                   `json:"description"`
	DeliverablesStructured []models.DeliverableNode `json:"deliverables_structured,omitempty"` // Hierarchical tree
	GoalKeys               []string                 `json:"goal_keys"`
	SuggestedTags          []string                 `json:"suggested_tags"`
	Explanation            string                   `json:"explanation"`
	Changed                bool                     `json:"changed"`
}

// RefineWorkItem uses AI to refine a work item definition based on all available context
func (r *WorkItemsBuilder) RefineWorkItem(ctx context.Context, req *builder.RefineWorkItemRequest, builderContext builder.BuilderContext) (*builder.RefineWorkItemResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work item refinement")

	// Build the prompt for work item refinement
	prompt := r.buildWorkItemRefinementPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: workItemRefinementSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for work item refinement")
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse aiWorkItemRefinementResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to our response format
	result := &builder.RefineWorkItemResponse{
		Title:                  aiResponse.Title,
		Description:            aiResponse.Description,
		DeliverablesStructured: aiResponse.DeliverablesStructured,
		GoalKeys:               aiResponse.GoalKeys,
		SuggestedTags:          aiResponse.SuggestedTags,
		WasChanged:             aiResponse.Changed,
		Explanation:            aiResponse.Explanation,
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":    req.AgencyID,
		"was_changed":  result.WasChanged,
		"title":        len(result.Title),
		"description":  len(result.Description),
		"deliverables": len(result.DeliverablesStructured),
	}).Info("AI work item refinement completed")

	return result, nil
}

// GenerateWorkItem uses AI to generate a new work item from user input
func (r *WorkItemsBuilder) GenerateWorkItem(ctx context.Context, req *builder.GenerateWorkItemRequest, builderContext builder.BuilderContext) (*builder.GenerateWorkItemResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work item generation")

	// Build the prompt for work item generation
	prompt := r.buildWorkItemGenerationPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: workItemGenerationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for work item generation")
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse builder.GenerateWorkItemResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":      req.AgencyID,
		"suggested_code": aiResponse.SuggestedCode,
		"title":          len(aiResponse.Title),
		"description":    len(aiResponse.Description),
	}).Info("AI work item generation completed")

	return &aiResponse, nil
}

// GenerateWorkItems uses AI to generate multiple work items from goals
func (r *WorkItemsBuilder) GenerateWorkItems(ctx context.Context, req *builder.GenerateWorkItemRequest, builderContext builder.BuilderContext) (*builder.GenerateWorkItemsResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work items generation")

	// Build the prompt for multiple work items generation
	prompt := r.buildWorkItemsGenerationPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: workItemsGenerationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for work items generation")
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse builder.GenerateWorkItemsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":        req.AgencyID,
		"work_items_count": len(aiResponse.WorkItems),
	}).Info("AI work items generation completed")

	return &aiResponse, nil
}

// buildWorkItemRefinementPrompt creates a context-rich prompt for work item refinement
func (r *WorkItemsBuilder) buildWorkItemRefinementPrompt(_ *builder.RefineWorkItemRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Please refine this work item to be clear, actionable, and aligned with agency goals.")

	return builder.String()
} // buildWorkItemGenerationPrompt creates a prompt for generating a single work item
func (r *WorkItemsBuilder) buildWorkItemGenerationPrompt(_ *builder.GenerateWorkItemRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Please generate a work item based on this request.")

	return builder.String()
}

// buildWorkItemsGenerationPrompt creates a prompt for generating multiple work items
func (r *WorkItemsBuilder) buildWorkItemsGenerationPrompt(_ *builder.GenerateWorkItemRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Please generate 3-7 work items that would help achieve these goals. ")
	builder.WriteString("Create a balanced mix of tasks, features, and possibly epic-level work items. ")
	builder.WriteString("Ensure each work item is specific, actionable, and clearly contributes to one or more goals.")

	return builder.String()
}

// ConsolidateWorkItems analyzes and consolidates work items into a lean, concise list
func (r *WorkItemsBuilder) ConsolidateWorkItems(ctx context.Context, req *builder.ConsolidateWorkItemsRequest, builderContext builder.BuilderContext) (*builder.ConsolidateWorkItemsResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":        req.AgencyID,
		"total_work_items": len(req.CurrentWorkItems),
	}).Info("Starting work item consolidation")

	// Build the prompt for work item consolidation
	prompt := r.buildWorkItemConsolidationPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: workItemConsolidationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for work item consolidation")
		return nil, fmt.Errorf("AI consolidation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)

	var consolidationResp builder.ConsolidateWorkItemsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &consolidationResp); err != nil {
		r.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse work item consolidation response")
		return nil, fmt.Errorf("failed to parse consolidation response: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"original_count":     len(req.CurrentWorkItems),
		"consolidated_count": len(consolidationResp.ConsolidatedWorkItems),
		"removed_count":      len(consolidationResp.RemovedWorkItems),
	}).Info("Work item consolidation completed")

	return &consolidationResp, nil
}

// buildWorkItemConsolidationPrompt creates the prompt for work item consolidation
func (r *WorkItemsBuilder) buildWorkItemConsolidationPrompt(_ *builder.ConsolidateWorkItemsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("Analyze these work items and provide a consolidated, optimized list. ")
	builder.WriteString("Remove duplicates, merge related items, and ensure clear separation of concerns. ")
	builder.WriteString("Keep only essential work items that directly contribute to the goals. ")
	builder.WriteString("Return a lean, actionable set of work items prioritized by importance.")

	return builder.String()
}

// System prompts for work item operations
const dynamicWorkItemsSystemPrompt = SharedAgencyContext + `

Act as a strategic work item management AI. Modify work items based on user requests.

**CRITICAL**: Work items are AGENT ACTIONS for Kanban boards (To Do → In Progress → Done), NOT system implementation tasks.

**TIME LIMIT**: You have approximately 90 seconds to complete your analysis. If processing many items:
- Prioritize quality over quantity
- Process items in order of importance
- If running out of time, return what you've completed so far with a partial explanation
- It's better to return 5 well-refined items than timeout trying to process 20

Examples:
✅ Agent Actions: "Review API spec v2.1", "Execute auth tests", "Deploy v1.2 to staging"
❌ Implementation: "Build payment API", "Create dashboard", "Implement CI/CD"

## Actions:
- **remove**: Delete work items
- **refine**: Improve existing items
- **generate**: Create new items aligned with goals
- **consolidate**: Merge duplicates
- **enhance_all**: Refine all items
- **no_action**: Already optimal

## Response JSON:
{
  "action": "remove|refine|generate|consolidate|enhance_all|no_action",
  "refined_work_items": [
    {
      "original_key": "key",
      "title": "Clear action-oriented title",
      "description": "Specific description",
      "deliverables_structured": [
        {
          "id": "uuid",
          "name": "folder_name",
          "description": "Folder description",
          "type": "folder",
          "prompt_instructions": "What to organize in this folder",
          "children": [
            {
              "id": "uuid",
              "name": "file_name",
              "description": "File description",
              "type": "file",
              "file_extension": ".md",
              "prompt_instructions": "Specific content to include in this file",
              "order": 1
            }
          ],
          "order": 1
        }
      ],
      "goal_keys": ["goal_key1"],
      "suggested_code": "CODE",
      "suggested_tags": ["tag1"],
      "was_changed": true,
      "explanation": "Brief explanation"
    }
  ],
  "generated_work_items": [...],
  "consolidated_data": {"consolidated_work_items": [...], "removed_work_items": ["key1"], "summary": "Brief", "explanation": "Brief"},
  "explanation": "Brief summary. If time-limited, note: 'Processed X of Y items before time limit'",
  "no_action_needed": false
}

**PARTIAL RESPONSES**: If approaching time limit, return completed items with explanation noting how many were processed.

## Deliverables Guidelines:
deliverables_structured is a hierarchical tree of folders/files with AI prompt instructions.

**Structure**: Use folders to group related deliverables (planning/, execution/, reports/)
**Naming**: Descriptive lowercase with underscores (technical_spec.md, not doc.md)
**Prompt Instructions**: Explain WHAT content goes in each file/folder with specific sections
**Order**: Use order field to sequence items
**File Type**: Only .md (Markdown) supported

Requirements:
- Use action verbs: Review, Execute, Deploy, Analyze, Process, Validate, Monitor, Generate
- Be specific with scope
- Link to goals via goal_keys array
- Keep explanations concise (1-2 sentences)
- Codes: 2-4 uppercase letters`

const workItemRefinementSystemPrompt = `You are an expert project manager and technical architect helping to refine work items for software development.

Your task is to refine work items to be:
1. Clear and specific
2. Actionable with well-defined deliverables
3. Properly scoped (not too large or too small)
4. Aligned with agency goals

Return your response as a JSON object with this structure:
{
  "title": "Clear, concise title",
  "description": "Detailed description of what needs to be done",
  "deliverables_structured": [
    {
      "id": "unique-uuid",
      "name": "requirements",
      "description": "Requirements documentation folder",
      "type": "folder",
      "prompt_instructions": "Organize all requirements documents",
      "children": [
        {
          "id": "unique-uuid-2",
          "name": "stakeholders",
          "description": "Stakeholder requirements",
          "type": "file",
          "file_extension": ".md",
          "prompt_instructions": "Document stakeholder needs",
          "order": 1
        }
      ],
      "order": 1
    }
  ],
  "goal_keys": ["goal_key1", "goal_key2"],
  "suggested_tags": ["tag1", "tag2"],
  "explanation": "Brief explanation of changes made",
  "changed": true|false
}

Set "changed" to false if the work item is already well-defined and needs no improvements.
**IMPORTANT**: 
- Always include goal_keys array with the keys of goals this work item addresses.
- Use deliverables_structured for hierarchical folder/file structure with prompt instructions.
- Each deliverable node must have: id, name, description, type ("folder" or "file"), prompt_instructions, order.
- Files must have file_extension (currently only ".md" supported).
- Folders can have children array with nested deliverables.`

const workItemGenerationSystemPrompt = `You are an expert project manager helping to create agent action work items.

CRITICAL: Work items are AGENT ACTIONS for Kanban boards, NOT system features to build.

Your task is to create a clear, actionable work item that:
1. Describes a specific AGENT ACTION (Review, Execute, Deploy, Analyze, Process, Validate, Monitor, Generate)
2. Addresses the user's request with concrete operations
3. Is Kanban-ready (clear completion criteria)
4. Aligns with agency goals
5. Is agent-executable (autonomous or human-in-loop can perform)

Examples of GOOD work items:
- "Review API specification v2.1 for completeness"
- "Execute security scan on production codebase"
- "Deploy hotfix v1.2.4 to production environment"
- "Analyze error logs from last 24 hours"

Examples of BAD work items (system features, not actions):
- "Build user authentication system"
- "Create reporting dashboard"

Return your response as a JSON object with this structure:
{
  "title": "Action-oriented title starting with verb",
  "description": "Detailed description of the agent action",
  "deliverables_structured": [
    {
      "id": "unique-uuid",
      "name": "reports",
      "description": "Analysis reports folder",
      "type": "folder",
      "prompt_instructions": "Store all analysis outputs",
      "children": [
        {
          "id": "unique-uuid-2",
          "name": "security_scan_results",
          "description": "Security scan findings",
          "type": "file",
          "file_extension": ".md",
          "prompt_instructions": "Document all security vulnerabilities found",
          "order": 1
        }
      ],
      "order": 1
    }
  ],
  "goal_keys": ["goal_key1", "goal_key2"],
  "suggested_code": "SHORT-CODE",
  "suggested_tags": ["tag1", "tag2"],
  "explanation": "Brief explanation of the action"
}

Use short, memorable codes (2-4 uppercase letters).
**IMPORTANT**: 
- Always include goal_keys array with the keys of goals this work item addresses.
- Use deliverables_structured for hierarchical folder/file structure with prompt instructions.
- Each deliverable node must have: id, name, description, type ("folder" or "file"), prompt_instructions, order.
- Files must have file_extension (currently only ".md" supported).
- Folders can have children array with nested deliverables.`

const workItemsGenerationSystemPrompt = `You are an expert project manager helping to break down goals into actionable work items.

CRITICAL: Work items are AGENT ACTIONS for Kanban boards, NOT system implementation tasks.

Your task is to generate 3-7 work items that:
1. Are specific AGENT ACTIONS (Review, Execute, Deploy, Analyze, Process, Validate, Monitor, Generate)
2. Help agents achieve the stated goals through concrete operations
3. Are Kanban-ready (clear start/done states)
4. Create a logical workflow sequence
5. Are agent-executable (autonomous or human-in-loop can complete)

Examples of GOOD work items:
- "Review requirements document for technical feasibility"
- "Execute integration test suite for gRPC services"
- "Deploy application to staging environment"
- "Analyze performance metrics and identify bottlenecks"
- "Generate API documentation from code annotations"

Examples of BAD work items (these are system features, not agent actions):
- "Build authentication system"
- "Create monitoring dashboard"
- "Implement payment processing"

Return your response as a JSON object with this structure:
{
  "work_items": [
    {
      "title": "Action-oriented title starting with verb (Review, Execute, Deploy, etc.)",
      "description": "Detailed description of what the agent does",
      "deliverables_structured": [
        {
          "id": "unique-uuid",
          "name": "documentation",
          "description": "Project documentation",
          "type": "folder",
          "prompt_instructions": "Maintain all project docs",
          "children": [
            {
              "id": "unique-uuid-2",
              "name": "api_spec",
              "description": "API specification document",
              "type": "file",
              "file_extension": ".md",
              "prompt_instructions": "Document all API endpoints and schemas",
              "order": 1
            }
          ],
          "order": 1
        }
      ],
      "goal_keys": ["goal_key1", "goal_key2"],
      "suggested_code": "SHORT-CODE",
      "suggested_tags": ["tag1", "tag2"],
      "explanation": "How this action helps achieve goals"
    }
  ],
  "explanation": "Overall workflow strategy and how these actions relate to goals"
}

Use short, memorable codes (2-4 uppercase letters) that are unique.
**IMPORTANT**: 
- Always include goal_keys array with the keys of goals each work item addresses.
- Link work items to the relevant goals from the agency context provided.
- Use deliverables_structured for hierarchical folder/file structure with prompt instructions.
- Each deliverable node must have: id, name, description, type ("folder" or "file"), prompt_instructions, order.
- Files must have file_extension (currently only ".md" supported).
- Folders can have children array with nested deliverables.`

const workItemConsolidationSystemPrompt = `Act as an experienced project manager. Your task is to analyze work items and determine if consolidation is beneficial.

IMPORTANT: Only consolidate work items when it truly adds value. If work items are already well-defined and distinct, keep them separate.

Evaluate if consolidation is needed:

1. **Assess consolidation value**:
   - Check for duplicate or near-duplicate work items
   - Look for items with significant scope overlap
   - Identify items that are really subtasks of a larger item
   - Determine if items can be combined without losing clarity
   - **If items are distinct and well-scoped, DO NOT force consolidation**

2. **When consolidation IS beneficial**:
   - Merge duplicate or overlapping work items
   - Combine related tasks when appropriate
   - Ensure each consolidated item remains actionable
   - Preserve all deliverables and requirements
   - Maintain clear acceptance criteria

3. **When consolidation is NOT beneficial**:
   - Return empty arrays for consolidated_work_items and removed_work_items
   - Provide explanation that work items are already well-defined
   - Don't force consolidation just to reduce count

4. **Maintain work quality**:
   - Each work item should be SMART and actionable
   - Avoid creating overly broad or vague items
   - Balance scope and achievability
   - Support clear sprint planning

5. **Track merges accurately** (only when consolidating):
   - Record ALL original work item keys that were merged in "merged_from_keys"
   - List ALL work item keys to DELETE in "removed_work_items"
   - Provide clear explanations of consolidation decisions

Focus on practical project management. Do not force consolidation.

Respond ONLY with valid JSON (no markdown, no explanations outside JSON) in this exact format:

If consolidation is NOT beneficial:
{
  "consolidated_work_items": [],
  "removed_work_items": [],
  "summary": "No consolidation needed - work items are already distinct and well-scoped",
  "explanation": "Each work item addresses a specific deliverable and should remain independent"
}

If consolidation IS beneficial:
{
  "consolidated_work_items": [
    {
      "title": "Clear, actionable title",
      "description": "Detailed description of what needs to be done",
      "deliverables_structured": [
        {
          "id": "unique-uuid",
          "name": "deliverables",
          "description": "Project deliverables",
          "type": "folder",
          "prompt_instructions": "All consolidated outputs",
          "children": [
            {
              "id": "unique-uuid-2",
              "name": "merged_output",
              "description": "Consolidated deliverable",
              "type": "file",
              "file_extension": ".md",
              "prompt_instructions": "Combined requirements from merged work items",
              "order": 1
            }
          ],
          "order": 1
        }
      ],
      "goal_keys": ["goal_key1", "goal_key2"],
      "suggested_code": "SHORT-CODE",
      "suggested_tags": ["tag1", "tag2"],
      "merged_from_keys": ["original_key1", "original_key2"],
      "explanation": "Brief explanation of what was consolidated and why"
    }
  ],
  "removed_work_items": ["original_key1", "original_key2"],
  "summary": "Consolidated X work items into Y more focused items",
  "explanation": "Overall consolidation strategy and benefits"
}
`
