package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/sirupsen/logrus"
)

// WorkItemConsolidator handles AI-powered work item consolidation
type WorkItemConsolidator struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewWorkItemConsolidator creates a new work item consolidator service
func NewWorkItemConsolidator(llmClient LLMClient, logger *logrus.Logger) *WorkItemConsolidator {
	return &WorkItemConsolidator{
		llmClient: llmClient,
		logger:    logger,
	}
}

// ConsolidateWorkItemsRequest contains the context for consolidating work items
type ConsolidateWorkItemsRequest struct {
	AgencyID         string             `json:"agency_id"`
	AgencyContext    *agency.Agency     `json:"agency_context"`
	CurrentWorkItems []*agency.WorkItem `json:"current_work_items"`
	Goals            []*agency.Goal     `json:"goals"`
}

// ConsolidateWorkItemsResponse contains the consolidated work items
type ConsolidateWorkItemsResponse struct {
	ConsolidatedWorkItems []ConsolidatedWorkItem `json:"consolidated_work_items"`
	RemovedWorkItems      []string               `json:"removed_work_items"` // Keys of work items that were consolidated/removed
	Summary               string                 `json:"summary"`
	Explanation           string                 `json:"explanation"`
}

// ConsolidatedWorkItem represents a work item after consolidation
type ConsolidatedWorkItem struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Deliverables      []string `json:"deliverables"`
	SuggestedCode     string   `json:"suggested_code"`
	SuggestedType     string   `json:"suggested_type"`
	SuggestedPriority string   `json:"suggested_priority"`
	SuggestedEffort   int      `json:"suggested_effort"`
	SuggestedTags     []string `json:"suggested_tags"`
	MergedFromKeys    []string `json:"merged_from_keys"` // Keys of original work items that were merged
	Explanation       string   `json:"explanation"`
}

// ConsolidateWorkItems analyzes and consolidates work items into a lean, concise list
func (c *WorkItemConsolidator) ConsolidateWorkItems(ctx context.Context, req *ConsolidateWorkItemsRequest) (*ConsolidateWorkItemsResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"agency_id":        req.AgencyID,
		"total_work_items": len(req.CurrentWorkItems),
	}).Info("Starting work item consolidation")

	// Build the prompt for work item consolidation
	prompt := c.buildConsolidationPrompt(req)

	// Make the LLM request
	response, err := c.llmClient.Chat(ctx, &ChatRequest{
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
		c.logger.WithError(err).Error("Failed to get AI response for work item consolidation")
		return nil, fmt.Errorf("AI consolidation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)

	var consolidationResp ConsolidateWorkItemsResponse
	if err := json.Unmarshal([]byte(cleanedContent), &consolidationResp); err != nil {
		c.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse work item consolidation response")
		return nil, fmt.Errorf("failed to parse consolidation response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"original_count":     len(req.CurrentWorkItems),
		"consolidated_count": len(consolidationResp.ConsolidatedWorkItems),
		"removed_count":      len(consolidationResp.RemovedWorkItems),
	}).Info("Work item consolidation completed")

	return &consolidationResp, nil
}

// buildConsolidationPrompt creates the prompt for work item consolidation
func (c *WorkItemConsolidator) buildConsolidationPrompt(req *ConsolidateWorkItemsRequest) string {
	// Create context map with relevant data
	contextData := map[string]interface{}{
		"current_work_items": req.CurrentWorkItems,
		"goals":              req.Goals,
	}

	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	builder.WriteString("\n\nPlease analyze these work items and consolidate them into a lean, manageable list. Look for duplicates, overlaps, and opportunities to combine related items.")

	return builder.String()
} // workItemConsolidationSystemPrompt defines the AI's role for work item consolidation
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
}`
