package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/sirupsen/logrus"
)

// Ensure AIRACIBuilder implements the RACIBuilderInterface
var _ builder.RACIBuilderInterface = (*RACIBuilder)(nil)

// RACIBuilder handles AI-powered RACI matrix creation
type RACIBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewAIRACIBuilder creates a new RACI builder service
func NewAIRACIBuilder(llmClient LLMClient, logger *logrus.Logger) *RACIBuilder {
	return &RACIBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// RefineRACIMappings is the main dynamic method for all RACI operations
// It analyzes the user message to determine what action to take and handles
// RACI refinement, generation, consolidation, and creation
func (r *RACIBuilder) RefineRACIMappings(ctx context.Context, req *builder.RefineRACIMappingsRequest, builderContext builder.BuilderContext) (*builder.RefineRACIMappingsResponse, error) {
	r.logger.WithField("agency_id", req.AgencyID).Info("Starting dynamic RACI processing")

	// For now, return a placeholder response
	// TODO: Implement dynamic RACI processing following the pattern from goals_builder.go
	response := &builder.RefineRACIMappingsResponse{
		Action:         "under_construction",
		Explanation:    "RACI processing is under construction. This will analyze the user message to determine whether to refine existing RACI assignments, generate new assignments, consolidate duplicate assignments, or create all assignments.",
		NoActionNeeded: false,
	}

	r.logger.Info("Dynamic RACI processing completed (placeholder)")
	return response, nil
}

// aiRACIMappingResponse represents the AI's response structure
type aiRACIMappingResponse struct {
	Mappings    []aiRACIMapping `json:"mappings"`
	Explanation string          `json:"explanation"`
}

type aiRACIMapping struct {
	WorkItemKey string `json:"work_item_key"`
	RoleKey     string `json:"role_key"`
	RACI        string `json:"raci"` // R, A, C, or I
	Objective   string `json:"objective"`
}

const raciCreationSystemPrompt = `You are an AI assistant specialized in creating RACI (Responsible, Accountable, Consulted, Informed) matrices for agency operations.

Your task is to analyze work items and roles, then create appropriate RACI assignments that clearly define responsibilities.

RACI Definitions:
- Responsible (R): The person/role who does the work to complete the task
- Accountable (A): The person/role ultimately answerable for the task completion (only ONE per work item)
- Consulted (C): People/roles who provide input and expertise
- Informed (I): People/roles who are kept updated on progress

Rules:
1. Each work item MUST have exactly ONE Accountable (A) role
2. Each work item SHOULD have at least one Responsible (R) role
3. Multiple roles can be Consulted (C) or Informed (I)
4. Provide clear objectives for each role-work item assignment
5. Consider the role's capabilities and the work item's requirements

Return your response as a JSON object with this structure:
{
  "mappings": [
    {
      "work_item_key": "string",
      "role_key": "string",
      "raci": "R|A|C|I",
      "objective": "Clear description of what this role needs to achieve for this work item"
    }
  ],
  "explanation": "Brief explanation of your RACI assignment strategy"
}`

// CreateRACIMappings generates RACI assignments using AI
func (r *RACIBuilder) CreateRACIMappings(ctx context.Context, req *builder.CreateRACIMappingsRequest, builderContext builder.BuilderContext) (*builder.CreateRACIMappingsResponse, error) {
	r.logger.WithFields(logrus.Fields{
		"agency_id":  req.AgencyID,
		"work_items": len(builderContext.WorkItems),
		"roles":      len(builderContext.Roles),
	}).Info("Creating RACI mappings with AI")

	// Build the prompt with context
	prompt := r.buildRACICreationPrompt(req, builderContext)

	// Make the LLM request
	response, err := r.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: raciCreationSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		r.logger.WithError(err).Error("Failed to get AI response for RACI creation")
		return nil, fmt.Errorf("AI RACI creation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse aiRACIMappingResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		r.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI RACI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to our response format
	assignments := make(map[string]map[string]builder.RACIAssignment)
	for _, mapping := range aiResponse.Mappings {
		if assignments[mapping.WorkItemKey] == nil {
			assignments[mapping.WorkItemKey] = make(map[string]builder.RACIAssignment)
		}
		assignments[mapping.WorkItemKey][mapping.RoleKey] = builder.RACIAssignment{
			RACI:      mapping.RACI,
			Objective: mapping.Objective,
		}
	}

	result := &builder.CreateRACIMappingsResponse{
		Assignments: assignments,
		Explanation: aiResponse.Explanation,
	}

	r.logger.WithFields(logrus.Fields{
		"agency_id":         req.AgencyID,
		"mappings_created":  len(aiResponse.Mappings),
		"work_items_mapped": len(assignments),
	}).Info("AI RACI creation completed")

	return result, nil
}

func (r *RACIBuilder) buildRACICreationPrompt(_ *builder.CreateRACIMappingsRequest, contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(contextData))

	builder.WriteString("\nPlease analyze these work items and roles, then create appropriate RACI assignments.\n")
	builder.WriteString("Ensure each work item has exactly one Accountable role and at least one Responsible role.\n")
	builder.WriteString("Provide clear objectives for each assignment that explain what the role needs to achieve.")

	return builder.String()
}
