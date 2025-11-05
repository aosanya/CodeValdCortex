package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// aiWorkItemRefinementResponse represents the JSON structure returned by the AI
type aiWorkItemRefinementResponse struct {
	RefinedTitle        string   `json:"refined_title"`
	RefinedDescription  string   `json:"refined_description"`
	RefinedDeliverables []string `json:"refined_deliverables"`
	SuggestedType       string   `json:"suggested_type"`
	SuggestedPriority   string   `json:"suggested_priority"`
	SuggestedEffort     int      `json:"suggested_effort"`
	SuggestedTags       []string `json:"suggested_tags"`
	Explanation         string   `json:"explanation"`
	Changed             bool     `json:"changed"`
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
		RefinedTitle:        aiResponse.RefinedTitle,
		RefinedDescription:  aiResponse.RefinedDescription,
		RefinedDeliverables: aiResponse.RefinedDeliverables,
		SuggestedType:       aiResponse.SuggestedType,
		SuggestedPriority:   aiResponse.SuggestedPriority,
		SuggestedEffort:     aiResponse.SuggestedEffort,
		SuggestedTags:       aiResponse.SuggestedTags,
		WasChanged:          aiResponse.Changed,
		Explanation:         aiResponse.Explanation,
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":    req.AgencyID,
		"was_changed":  result.WasChanged,
		"title":        len(result.RefinedTitle),
		"description":  len(result.RefinedDescription),
		"deliverables": len(result.RefinedDeliverables),
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
		"suggested_type": aiResponse.SuggestedType,
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
const workItemRefinementSystemPrompt = `You are an expert project manager and technical architect helping to refine work items for software development.

Your task is to refine work items to be:
1. Clear and specific
2. Actionable with well-defined deliverables
3. Properly scoped (not too large or too small)
4. Aligned with agency goals
5. Categorized correctly (Task, Feature, Epic, Bug, or Research)

Return your response as a JSON object with this structure:
{
  "refined_title": "Clear, concise title",
  "refined_description": "Detailed description of what needs to be done",
  "refined_deliverables": ["Deliverable 1", "Deliverable 2"],
  "suggested_type": "Task|Feature|Epic|Bug|Research",
  "suggested_priority": "P0|P1|P2|P3",
  "suggested_effort": 1-13,
  "suggested_tags": ["tag1", "tag2"],
  "explanation": "Brief explanation of changes made",
  "changed": true|false
}

Set "changed" to false if the work item is already well-defined and needs no improvements.
Effort should follow Fibonacci numbers (1, 2, 3, 5, 8, 13) representing story points.`

const workItemGenerationSystemPrompt = `You are an expert project manager and technical architect helping to create work items for software development.

Your task is to create a clear, actionable work item that:
1. Addresses the user's request
2. Is properly scoped
3. Aligns with agency goals
4. Has clear deliverables
5. Is correctly categorized

Return your response as a JSON object with this structure:
{
  "title": "Clear, concise title",
  "description": "Detailed description of what needs to be done",
  "deliverables": ["Deliverable 1", "Deliverable 2"],
  "suggested_code": "SHORT-CODE",
  "suggested_type": "Task|Feature|Epic|Bug|Research",
  "suggested_priority": "P0|P1|P2|P3",
  "suggested_effort": 1-13,
  "suggested_tags": ["tag1", "tag2"],
  "explanation": "Brief explanation of the work item"
}

Use short, memorable codes (2-4 uppercase letters).
Effort should follow Fibonacci numbers (1, 2, 3, 5, 8, 13) representing story points.`

const workItemsGenerationSystemPrompt = `You are an expert project manager and technical architect helping to break down goals into actionable work items.

Your task is to generate 3-7 work items that:
1. Help achieve the stated goals
2. Are properly scoped and prioritized
3. Create a logical development sequence
4. Mix different types (Tasks, Features, possibly Epics)
5. Have clear deliverables

Return your response as a JSON object with this structure:
{
  "work_items": [
    {
      "title": "Clear, concise title",
      "description": "Detailed description",
      "deliverables": ["Deliverable 1", "Deliverable 2"],
      "suggested_code": "SHORT-CODE",
      "suggested_type": "Task|Feature|Epic|Bug|Research",
      "suggested_priority": "P0|P1|P2|P3",
      "suggested_effort": 1-13,
      "suggested_tags": ["tag1", "tag2"],
      "explanation": "How this contributes to goals"
    }
  ],
  "explanation": "Overall strategy for these work items and how they relate to goals"
}

Use short, memorable codes (2-4 uppercase letters) that are unique.
Prioritize foundational work as P0/P1, enhancements as P2/P3.
Effort should follow Fibonacci numbers (1, 2, 3, 5, 8, 13) representing story points.`

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
   - Combine related tasks into Features or Epics when appropriate
   - Ensure each consolidated item remains actionable
   - Preserve all deliverables and requirements
   - Maintain clear acceptance criteria
   - Keep effort estimates reasonable (1, 2, 3, 5, 8, 13 story points)

3. **When consolidation is NOT beneficial**:
   - Return empty arrays for consolidated_work_items and removed_work_items
   - Provide explanation that work items are already well-defined
   - Don't force consolidation just to reduce count

4. **Maintain work quality**:
   - Each work item should be SMART and actionable
   - Avoid creating overly broad or vague items
   - Ensure proper categorization (Task, Feature, Epic, Bug, Research)
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
      "deliverables": ["Deliverable 1", "Deliverable 2", "Deliverable 3"],
      "suggested_code": "SHORT-CODE",
      "suggested_type": "Task|Feature|Epic|Bug|Research",
      "suggested_priority": "P0|P1|P2|P3",
      "suggested_effort": 1-13,
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
