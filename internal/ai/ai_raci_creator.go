package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/sirupsen/logrus"
)

// RACICreator handles AI-powered RACI matrix creation
type RACICreator struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewRACICreator creates a new RACI creator service
func NewRACICreator(llmClient LLMClient, logger *logrus.Logger) *RACICreator {
	return &RACICreator{
		llmClient: llmClient,
		logger:    logger,
	}
}

// CreateRACIMappingsRequest contains the context for creating RACI mappings
type CreateRACIMappingsRequest struct {
	AgencyID      string             `json:"agency_id"`
	WorkItems     []*agency.WorkItem `json:"work_items"`
	Roles         []*registry.Role   `json:"roles"`
	AgencyContext *agency.Agency     `json:"agency_context"`
}

// RACIAssignment represents a role assignment with objective
type RACIAssignment struct {
	RACI      string `json:"raci"`      // R, A, C, or I
	Objective string `json:"objective"` // What this role needs to achieve
}

// CreateRACIMappingsResponse contains the AI-generated RACI mappings
type CreateRACIMappingsResponse struct {
	Assignments map[string]map[string]RACIAssignment `json:"assignments"` // workItemKey -> roleKey -> assignment
	Explanation string                               `json:"explanation"`
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
func (c *RACICreator) CreateRACIMappings(ctx context.Context, req *CreateRACIMappingsRequest) (*CreateRACIMappingsResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"agency_id":  req.AgencyID,
		"work_items": len(req.WorkItems),
		"roles":      len(req.Roles),
	}).Info("Creating RACI mappings with AI")

	// Build the prompt with context
	prompt := c.buildRACICreationPrompt(req)

	// Make the LLM request
	response, err := c.llmClient.Chat(ctx, &ChatRequest{
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
		c.logger.WithError(err).Error("Failed to get AI response for RACI creation")
		return nil, fmt.Errorf("AI RACI creation failed: %w", err)
	}

	// Parse the AI response
	cleanedContent := stripMarkdownFences(response.Content)
	var aiResponse aiRACIMappingResponse
	if err := json.Unmarshal([]byte(cleanedContent), &aiResponse); err != nil {
		c.logger.WithError(err).WithField("response", response.Content).Error("Failed to parse AI RACI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to our response format
	assignments := make(map[string]map[string]RACIAssignment)
	for _, mapping := range aiResponse.Mappings {
		if assignments[mapping.WorkItemKey] == nil {
			assignments[mapping.WorkItemKey] = make(map[string]RACIAssignment)
		}
		assignments[mapping.WorkItemKey][mapping.RoleKey] = RACIAssignment{
			RACI:      mapping.RACI,
			Objective: mapping.Objective,
		}
	}

	result := &CreateRACIMappingsResponse{
		Assignments: assignments,
		Explanation: aiResponse.Explanation,
	}

	c.logger.WithFields(logrus.Fields{
		"agency_id":         req.AgencyID,
		"mappings_created":  len(aiResponse.Mappings),
		"work_items_mapped": len(assignments),
	}).Info("AI RACI creation completed")

	return result, nil
}

func (c *RACICreator) buildRACICreationPrompt(req *CreateRACIMappingsRequest) string {
	// Create context map with relevant data
	contextData := map[string]interface{}{
		"work_items": req.WorkItems,
		"roles":      req.Roles,
	}

	var builder strings.Builder

	// Use the reusable agency context formatter
	builder.WriteString(FormatAgencyContextBlock(req.AgencyContext, contextData))

	builder.WriteString("\nPlease analyze these work items and roles, then create appropriate RACI assignments.\n")
	builder.WriteString("Ensure each work item has exactly one Accountable role and at least one Responsible role.\n")
	builder.WriteString("Provide clear objectives for each assignment that explain what the role needs to achieve.")

	return builder.String()
}
