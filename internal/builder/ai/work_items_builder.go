package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai/prompts"
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
				Content: SharedAgencyContext + "\n\n" + prompts.WorkItemsDynamicSystem,
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
				Content: SharedAgencyContext + "\n\n" + prompts.WorkItemsDynamicSystem,
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
				Content: prompts.WorkItemsRefinementSystem,
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
				Content: prompts.WorkItemsGenerationSingleSystem,
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
				Content: prompts.WorkItemsGenerationMultipleSystem,
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
				Content: prompts.WorkItemsConsolidationSystem,
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
