package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgencyDesignerService manages AI-powered agency design conversations
type AgencyDesignerService struct {
	llmClient     LLMClient
	logger        *logrus.Logger
	conversations map[string]*ConversationContext // In-memory for MVP, should be persistent
}

// NewAgencyDesignerService creates a new agency designer service
func NewAgencyDesignerService(llmClient LLMClient, logger *logrus.Logger) *AgencyDesignerService {
	return &AgencyDesignerService{
		llmClient:     llmClient,
		logger:        logger,
		conversations: make(map[string]*ConversationContext),
	}
}

// StartConversation begins a new agency design conversation
func (s *AgencyDesignerService) StartConversation(ctx context.Context, agencyID string) (*ConversationContext, error) {
	conversationID := uuid.New().String()

	conversation := &ConversationContext{
		ID:        conversationID,
		AgencyID:  agencyID,
		Phase:     PhaseInitial,
		Messages:  []Message{},
		State:     make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add initial system message
	systemPrompt := s.getSystemPrompt(PhaseInitial)
	conversation.Messages = append(conversation.Messages, Message{
		Role:      "system",
		Content:   systemPrompt,
		Timestamp: time.Now(),
	})

	// Get initial AI greeting
	initialGreeting := s.getInitialGreeting()
	conversation.Messages = append(conversation.Messages, Message{
		Role:      "assistant",
		Content:   initialGreeting,
		Timestamp: time.Now(),
	})

	s.conversations[conversationID] = conversation

	s.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"agency_id":       agencyID,
	}).Info("Started new agency design conversation")

	return conversation, nil
}

// SendMessage sends a user message and gets AI response
func (s *AgencyDesignerService) SendMessage(ctx context.Context, conversationID, userMessage string) (*Message, error) {
	conversation, exists := s.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	// Add user message
	conversation.Messages = append(conversation.Messages, Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	})

	// Get AI response
	response, err := s.llmClient.Chat(ctx, &ChatRequest{
		Messages:    conversation.Messages,
		Temperature: 0.7,
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	// Add assistant response
	assistantMsg := Message{
		Role:      "assistant",
		Content:   response.Content,
		Timestamp: time.Now(),
	}
	conversation.Messages = append(conversation.Messages, assistantMsg)
	conversation.UpdatedAt = time.Now()

	// Extract information and update phase
	s.extractInformation(conversation, userMessage, response.Content)
	s.updatePhase(conversation)

	s.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"phase":           conversation.Phase,
		"tokens":          response.Usage.TotalTokens,
	}).Debug("Processed message")

	return &assistantMsg, nil
}

// GetConversation retrieves a conversation by ID
func (s *AgencyDesignerService) GetConversation(conversationID string) (*ConversationContext, error) {
	conversation, exists := s.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}
	return conversation, nil
}

// GenerateAgencyDesign creates the final agency design from conversation
func (s *AgencyDesignerService) GenerateAgencyDesign(ctx context.Context, conversationID string) (*AgencyDesign, error) {
	conversation, exists := s.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	if conversation.Phase != PhaseComplete && conversation.Phase != PhaseValidation {
		return nil, fmt.Errorf("conversation not ready for design generation, current phase: %s", conversation.Phase)
	}

	// Request structured output from LLM
	designPrompt := s.getDesignGenerationPrompt(conversation)

	messages := append(conversation.Messages, Message{
		Role:    "user",
		Content: designPrompt,
	})

	response, err := s.llmClient.Chat(ctx, &ChatRequest{
		Messages:    messages,
		Temperature: 0.3, // Lower temperature for structured output
		MaxTokens:   4096,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate design: %w", err)
	}

	// Parse the JSON response
	design, err := s.parseAgencyDesign(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse design: %w", err)
	}

	design.AgencyID = conversation.AgencyID

	s.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"agency_id":       conversation.AgencyID,
		"agent_types":     len(design.AgentTypes),
		"relationships":   len(design.Relationships),
	}).Info("Generated agency design")

	return design, nil
}

// extractInformation extracts key information from the conversation
func (s *AgencyDesignerService) extractInformation(conversation *ConversationContext, userMsg, aiMsg string) {
	// Simple keyword-based extraction for MVP
	// TODO: Use LLM to extract structured information

	state := conversation.State

	// Extract business domain
	if state["domain"] == nil {
		keywords := []string{"warehouse", "logistics", "healthcare", "water", "infrastructure", "agriculture"}
		for _, keyword := range keywords {
			if contains(userMsg, keyword) {
				state["domain"] = keyword
				break
			}
		}
	}

	// Extract agent type mentions
	if state["agent_types"] == nil {
		state["agent_types"] = []string{}
	}
	agentTypes := state["agent_types"].([]string)

	// Look for agent type patterns in AI response
	if contains(aiMsg, "Agent") || contains(aiMsg, "agent") {
		// Extract agent types (simplified)
		// In production, use better NLP or structured output
	}

	state["agent_types"] = agentTypes
}

// updatePhase transitions the conversation to the next phase
func (s *AgencyDesignerService) updatePhase(conversation *ConversationContext) {
	state := conversation.State
	messageCount := len(conversation.Messages)

	// Simple heuristic-based phase transitions
	switch conversation.Phase {
	case PhaseInitial:
		if messageCount > 2 {
			conversation.Phase = PhaseRequirements
		}
	case PhaseRequirements:
		if messageCount > 6 && state["domain"] != nil {
			conversation.Phase = PhaseAgentBrainstorm
		}
	case PhaseAgentBrainstorm:
		if messageCount > 12 {
			conversation.Phase = PhaseRelationshipMapping
		}
	case PhaseRelationshipMapping:
		if messageCount > 18 {
			conversation.Phase = PhaseValidation
		}
	case PhaseValidation:
		// Manual transition to complete
		if state["approved"] == true {
			conversation.Phase = PhaseComplete
		}
	}
}

// parseAgencyDesign parses JSON design from LLM response
func (s *AgencyDesignerService) parseAgencyDesign(content string) (*AgencyDesign, error) {
	// Extract JSON from markdown code blocks if present
	jsonStr := extractJSON(content)

	var design AgencyDesign
	if err := json.Unmarshal([]byte(jsonStr), &design); err != nil {
		return nil, fmt.Errorf("failed to unmarshal design JSON: %w", err)
	}

	return &design, nil
}

// Helper functions

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		(text == substr || len(text) >= len(substr) &&
			(text[:len(substr)] == substr || text[len(text)-len(substr):] == substr ||
				len(text) > len(substr) && findSubstring(text, substr)))
}

func findSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractJSON(content string) string {
	// Look for JSON in markdown code blocks
	start := -1
	end := -1

	// Find ```json or ```
	markers := []string{"```json", "```"}
	for _, marker := range markers {
		idx := findIndex(content, marker)
		if idx != -1 {
			start = idx + len(marker)
			// Find closing ```
			closeIdx := findIndex(content[start:], "```")
			if closeIdx != -1 {
				end = start + closeIdx
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return content[start:end]
	}

	// If no code blocks, return as is (assume it's JSON)
	return content
}

func findIndex(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
