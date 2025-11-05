package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/sirupsen/logrus"
)

// WorkItemRefiner handles AI-powered work item definition and refinement
type WorkItemRefiner struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewWorkItemRefiner creates a new work item refiner service
func NewWorkItemRefiner(llmClient LLMClient, logger *logrus.Logger) *WorkItemRefiner {
	return &WorkItemRefiner{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineWorkItemRequest contains the context for refining a work item
type RefineWorkItemRequest struct {
	AgencyID          string             `json:"agency_id"`
	CurrentWorkItem   *agency.WorkItem   `json:"current_work_item"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Deliverables      []string           `json:"deliverables"`
	ExistingWorkItems []*agency.WorkItem `json:"existing_work_items"`
	Goals             []*agency.Goal     `json:"goals"`
	AgencyContext     *agency.Agency     `json:"agency_context"`
}

// RefineWorkItemResponse contains the AI-refined work item
type RefineWorkItemResponse struct {
	RefinedTitle        string   `json:"refined_title"`
	RefinedDescription  string   `json:"refined_description"`
	RefinedDeliverables []string `json:"refined_deliverables"`
	SuggestedType       string   `json:"suggested_type"`
	SuggestedPriority   string   `json:"suggested_priority"`
	SuggestedEffort     int      `json:"suggested_effort"`
	SuggestedTags       []string `json:"suggested_tags"`
	WasChanged          bool     `json:"was_changed"`
	Explanation         string   `json:"explanation"`
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

// GenerateWorkItemRequest contains the context for generating a new work item
type GenerateWorkItemRequest struct {
	AgencyID          string             `json:"agency_id"`
	AgencyContext     *agency.Agency     `json:"agency_context"`
	ExistingWorkItems []*agency.WorkItem `json:"existing_work_items"`
	Goals             []*agency.Goal     `json:"goals"`
	UserInput         string             `json:"user_input"`
}

// GenerateWorkItemResponse contains the AI-generated work item
type GenerateWorkItemResponse struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Deliverables      []string `json:"deliverables"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedType     string   `json:"suggested_type"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedEffort   int      `json:"suggested_effort"`
	SuggestedTags     []string `json:"suggested_tags"`
	Explanation       string   `json:"explanation"`
}

// GenerateWorkItemsResponse contains multiple AI-generated work items
type GenerateWorkItemsResponse struct {
	WorkItems   []GenerateWorkItemResponse `json:"work_items"`
	Explanation string                     `json:"explanation"`
}

// RefineWorkItem uses AI to refine a work item definition based on all available context
func (r *WorkItemRefiner) RefineWorkItem(ctx context.Context, req *RefineWorkItemRequest) (*RefineWorkItemResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work item refinement")

	// Build the prompt for work item refinement
	prompt := r.buildWorkItemRefinementPrompt(req)

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
	result := &RefineWorkItemResponse{
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
func (r *WorkItemRefiner) GenerateWorkItem(ctx context.Context, req *GenerateWorkItemRequest) (*GenerateWorkItemResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work item generation")

	// Build the prompt for work item generation
	prompt := r.buildWorkItemGenerationPrompt(req)

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
	var aiResponse GenerateWorkItemResponse
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
func (r *WorkItemRefiner) GenerateWorkItems(ctx context.Context, req *GenerateWorkItemRequest) (*GenerateWorkItemsResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting AI work items generation")

	// Build the prompt for multiple work items generation
	prompt := r.buildWorkItemsGenerationPrompt(req)

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
	var aiResponse GenerateWorkItemsResponse
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
func (r *WorkItemRefiner) buildWorkItemRefinementPrompt(req *RefineWorkItemRequest) string {
	// Create context map with relevant data
	contextData := map[string]interface{}{
		"current_work_item":   req.CurrentWorkItem,
		"title":               req.Title,
		"description":         req.Description,
		"deliverables":        req.Deliverables,
		"existing_work_items": req.ExistingWorkItems,
		"goals":               req.Goals,
	}

	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	builder.WriteString("Please refine this work item to be clear, actionable, and aligned with agency goals.")

	return builder.String()
} // buildWorkItemGenerationPrompt creates a prompt for generating a single work item
func (r *WorkItemRefiner) buildWorkItemGenerationPrompt(req *GenerateWorkItemRequest) string {
	// Create context map with relevant data
	contextData := map[string]interface{}{
		"goals":               req.Goals,
		"existing_work_items": req.ExistingWorkItems,
		"user_input":          req.UserInput,
	}

	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	builder.WriteString("Please generate a work item based on this request.")

	return builder.String()
}

// buildWorkItemsGenerationPrompt creates a prompt for generating multiple work items
func (r *WorkItemRefiner) buildWorkItemsGenerationPrompt(req *GenerateWorkItemRequest) string {
	// Create context map with relevant data
	contextData := map[string]interface{}{
		"goals":               req.Goals,
		"existing_work_items": req.ExistingWorkItems,
		"user_input":          req.UserInput,
	}

	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	builder.WriteString("Please generate 3-7 work items that would help achieve these goals. ")
	builder.WriteString("Create a balanced mix of tasks, features, and possibly epic-level work items. ")
	builder.WriteString("Ensure each work item is specific, actionable, and clearly contributes to one or more goals.")

	return builder.String()
} // System prompts for work item operations
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
