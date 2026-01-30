package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai/prompts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FlexibleIntroductionBuilder handles AI-powered flexible introduction generation and refinement
type FlexibleIntroductionBuilder struct {
	llmClient LLMClient
	logger    *logrus.Logger
}

// NewFlexibleIntroductionBuilder creates a new flexible introduction builder
func NewFlexibleIntroductionBuilder(llmClient LLMClient, logger *logrus.Logger) *FlexibleIntroductionBuilder {
	return &FlexibleIntroductionBuilder{
		llmClient: llmClient,
		logger:    logger,
	}
}

// GenerateIntroductionRequest represents a request to generate a complete introduction
type GenerateIntroductionRequest struct {
	AgencyID      string                 `json:"agency_id"`
	Template      string                 `json:"template"`
	Keywords      []string               `json:"keywords"`
	AgencyContext map[string]interface{} `json:"agency_context"`
}

// GenerateIntroductionResponse represents the AI response for generation
type GenerateIntroductionResponse struct {
	Introduction *models.AgencyIntroduction `json:"introduction"`
	Confidence   float64                    `json:"confidence"`
	Explanation  string                     `json:"explanation"`
}

// RefineSectionRequest represents a request to refine a specific section
type RefineSectionRequest struct {
	AgencyID         string                       `json:"agency_id"`
	Section          *models.IntroductionSection  `json:"section"`
	RefinementText   string                       `json:"refinement_text"`
	AgencyContext    map[string]interface{}       `json:"agency_context"`
}

// RefineSectionResponse represents the AI response for section refinement
type RefineSectionResponse struct {
	Section     *models.IntroductionSection `json:"section"`
	Changed     bool                        `json:"changed"`
	Explanation string                      `json:"explanation"`
	Confidence  float64                     `json:"confidence"`
}

// aiGenerateResponse matches the expected AI output for generation
type aiGenerateResponse struct {
	Sections    []aiSectionOutput `json:"sections"`
	Confidence  float64           `json:"confidence"`
	Explanation string            `json:"explanation"`
}

// aiSectionOutput represents a section in AI generation output
type aiSectionOutput struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Content  map[string]interface{} `json:"content"`
	Order    int                    `json:"order"`
	Required bool                   `json:"required"`
}

// aiRefineResponse matches the expected AI output for refinement
type aiRefineResponse struct {
	Content     map[string]interface{} `json:"content"`
	Changed     bool                   `json:"changed"`
	Explanation string                 `json:"explanation"`
	Confidence  float64                `json:"confidence"`
}

// GenerateFromKeywords generates a complete introduction based on keywords and template
func (b *FlexibleIntroductionBuilder) GenerateFromKeywords(ctx context.Context, req *GenerateIntroductionRequest) (*GenerateIntroductionResponse, error) {
	b.logger.WithFields(logrus.Fields{
		"agency_id": req.AgencyID,
		"template":  req.Template,
		"keywords":  req.Keywords,
	}).Info("Starting AI introduction generation")

	// Build the prompt
	keywordsStr := strings.Join(req.Keywords, ", ")
	prompt := prompts.GenerateIntroductionPromptTemplate(req.Template, keywordsStr, req.AgencyContext)

	// Add template descriptions and examples for context
	prompt += "\n\n" + prompts.TemplateDescriptions
	prompt += "\n\n" + prompts.ExampleOutputs

	b.logger.WithFields(logrus.Fields{
		"prompt_length": len(prompt),
	}).Debug("Built generation prompt")

	// Request AI generation
	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: prompts.FlexibleIntroductionSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7, // Some creativity for generation
		MaxTokens:   4096,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation request failed: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"response_length": len(response.Content),
		"tokens_used":     response.Usage.TotalTokens,
	}).Debug("Received AI response")

	// Parse the response
	intro, confidence, explanation, err := b.parseGenerateResponse(response.Content, req.AgencyID, req.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"agency_id":       req.AgencyID,
		"sections_count":  len(intro.Sections),
		"confidence":      confidence,
	}).Info("Successfully generated introduction")

	return &GenerateIntroductionResponse{
		Introduction: intro,
		Confidence:   confidence,
		Explanation:  explanation,
	}, nil
}

// RefineSection refines a specific section based on user input
func (b *FlexibleIntroductionBuilder) RefineSection(ctx context.Context, req *RefineSectionRequest) (*RefineSectionResponse, error) {
	b.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"section_id":  req.Section.ID,
		"section_type": req.Section.Type,
	}).Info("Starting AI section refinement")

	// Serialize current content
	currentContent, err := json.MarshalIndent(req.Section.Content, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize current content: %w", err)
	}

	// Build the prompt
	prompt := prompts.RefineSectionPromptTemplate(
		string(req.Section.Type),
		string(currentContent),
		req.RefinementText,
		req.AgencyContext,
	)

	// Add example outputs for reference
	prompt += "\n\n" + prompts.ExampleOutputs

	b.logger.WithFields(logrus.Fields{
		"prompt_length": len(prompt),
	}).Debug("Built refinement prompt")

	// Request AI refinement
	response, err := b.llmClient.Chat(ctx, &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: prompts.FlexibleIntroductionSystemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower creativity for refinement
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, fmt.Errorf("AI refinement request failed: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"response_length": len(response.Content),
		"tokens_used":     response.Usage.TotalTokens,
	}).Debug("Received AI response")

	// Parse the response
	refinedSection, changed, explanation, confidence, err := b.parseRefineResponse(response.Content, req.Section)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	b.logger.WithFields(logrus.Fields{
		"agency_id":   req.AgencyID,
		"section_id":  req.Section.ID,
		"changed":     changed,
		"confidence":  confidence,
	}).Info("Successfully refined section")

	return &RefineSectionResponse{
		Section:     refinedSection,
		Changed:     changed,
		Explanation: explanation,
		Confidence:  confidence,
	}, nil
}

// parseGenerateResponse parses the AI response for generation
func (b *FlexibleIntroductionBuilder) parseGenerateResponse(response, agencyID, template string) (*models.AgencyIntroduction, float64, string, error) {
	// Extract JSON from response if wrapped in other text
	jsonStr := b.extractJSON(response)

	var aiResp aiGenerateResponse
	if err := json.Unmarshal([]byte(jsonStr), &aiResp); err != nil {
		b.logger.WithError(err).Error("Failed to parse AI generation response")
		return nil, 0, "", fmt.Errorf("invalid AI response format: %w", err)
	}

	// Convert AI sections to models
	sections := make([]models.IntroductionSection, 0, len(aiResp.Sections))
	for _, aiSection := range aiResp.Sections {
		// Generate ID if not provided
		id := aiSection.ID
		if id == "" {
			id = uuid.New().String()
		}

		// Convert type string to SectionType
		var sectionType models.SectionType
		switch strings.ToLower(aiSection.Type) {
		case "text":
			sectionType = models.SectionTypeText
		case "list":
			sectionType = models.SectionTypeList
		case "nested":
			sectionType = models.SectionTypeNested
		case "table":
			sectionType = models.SectionTypeTable
		default:
			b.logger.WithField("type", aiSection.Type).Warn("Unknown section type, defaulting to text")
			sectionType = models.SectionTypeText
		}

		section := models.IntroductionSection{
			ID:       id,
			Type:     sectionType,
			Title:    aiSection.Title,
			Content:  aiSection.Content,
			Order:    aiSection.Order,
			Required: aiSection.Required,
		}

		sections = append(sections, section)
	}

	// Create the introduction
	intro := models.NewAgencyIntroduction(agencyID, template)
	intro.Sections = sections

	return intro, aiResp.Confidence, aiResp.Explanation, nil
}

// parseRefineResponse parses the AI response for refinement
func (b *FlexibleIntroductionBuilder) parseRefineResponse(response string, originalSection *models.IntroductionSection) (*models.IntroductionSection, bool, string, float64, error) {
	// Extract JSON from response if wrapped in other text
	jsonStr := b.extractJSON(response)

	var aiResp aiRefineResponse
	if err := json.Unmarshal([]byte(jsonStr), &aiResp); err != nil {
		b.logger.WithError(err).Error("Failed to parse AI refinement response")
		return nil, false, "", 0, fmt.Errorf("invalid AI response format: %w", err)
	}

	// Create refined section
	refinedSection := *originalSection
	refinedSection.Content = aiResp.Content

	return &refinedSection, aiResp.Changed, aiResp.Explanation, aiResp.Confidence, nil
}

// extractJSON attempts to extract JSON from response that may contain additional text
func (b *FlexibleIntroductionBuilder) extractJSON(response string) string {
	// Trim whitespace
	response = strings.TrimSpace(response)

	// If it starts with {, assume it's pure JSON
	if strings.HasPrefix(response, "{") {
		return response
	}

	// Try to find JSON within the response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		return response[startIdx : endIdx+1]
	}

	// Return as-is if no JSON found
	return response
}
