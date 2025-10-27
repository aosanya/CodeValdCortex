package ai

import (
	"context"
	"time"
)

// Provider represents different LLM providers
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderClaude Provider = "claude"
	ProviderLocal  Provider = "local"
	ProviderCustom Provider = "custom"
)

// Message represents a chat message
type Message struct {
	Role      string    `json:"role"`           // "system", "user", "assistant"
	Content   string    `json:"content"`        // Message content
	Name      string    `json:"name,omitempty"` // Optional speaker name
	Timestamp time.Time `json:"timestamp"`      // When message was created
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Messages    []Message              `json:"messages"`
	Model       string                 `json:"model,omitempty"`       // Optional model override
	Temperature float32                `json:"temperature,omitempty"` // 0.0 to 2.0
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Additional context
}

// ChatResponse represents the LLM response
type ChatResponse struct {
	Content      string                 `json:"content"`
	FinishReason string                 `json:"finish_reason,omitempty"` // "stop", "length", "content_filter"
	Usage        *TokenUsage            `json:"usage,omitempty"`
	Model        string                 `json:"model,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamCallback is called for each chunk in streaming mode
type StreamCallback func(chunk string) error

// LLMClient defines the interface for LLM interactions
type LLMClient interface {
	// Chat sends messages and gets a response
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream sends messages and streams the response
	ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error

	// GetProvider returns the provider type
	GetProvider() Provider

	// GetModel returns the model being used
	GetModel() string
}

// LLMConfig holds configuration for LLM clients
type LLMConfig struct {
	Provider    Provider `json:"provider"`
	APIKey      string   `json:"api_key"`
	Model       string   `json:"model"`
	BaseURL     string   `json:"base_url,omitempty"`    // For custom endpoints
	Temperature float32  `json:"temperature,omitempty"` // Default temperature
	MaxTokens   int      `json:"max_tokens,omitempty"`  // Default max tokens
	Timeout     int      `json:"timeout,omitempty"`     // Request timeout in seconds
}

// ConversationContext holds the state of an ongoing conversation
type ConversationContext struct {
	ID            string                 `json:"id"`
	AgencyID      string                 `json:"agency_id"`
	Phase         DesignPhase            `json:"phase"`
	Messages      []Message              `json:"messages"`
	State         map[string]interface{} `json:"state"` // Extracted information
	CurrentDesign *AgencyDesign          `json:"current_design,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// DesignPhase represents the current phase of agency design
type DesignPhase string

const (
	PhaseInitial             DesignPhase = "initial"
	PhaseRequirements        DesignPhase = "requirements"
	PhaseAgentBrainstorm     DesignPhase = "agent_brainstorm"
	PhaseRelationshipMapping DesignPhase = "relationship_mapping"
	PhaseValidation          DesignPhase = "validation"
	PhaseComplete            DesignPhase = "complete"
)

// AgencyDesign represents the designed agency structure
type AgencyDesign struct {
	AgencyID      string                 `json:"agency_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	AgentTypes    []AgentTypeSpec        `json:"agent_types"`
	Relationships []AgentRelationship    `json:"relationships"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AgentTypeSpec represents a designed agent type
type AgentTypeSpec struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Capabilities  []string               `json:"capabilities"`
	Schema        map[string]interface{} `json:"schema"`
	DefaultConfig map[string]interface{} `json:"default_config,omitempty"`
	Count         int                    `json:"count,omitempty"` // Suggested instance count
}

// AgentRelationship defines communication between agent types
type AgentRelationship struct {
	From        string   `json:"from"`        // Source agent type ID
	To          string   `json:"to"`          // Target agent type ID
	Type        string   `json:"type"`        // "pub_sub", "direct", "broadcast"
	Topics      []string `json:"topics"`      // Pub/sub topics
	Description string   `json:"description"` // What is being communicated
}
