package ai

import (
	"fmt"
)

// NewLLMClient creates an LLM client based on the configuration
func NewLLMClient(config *LLMConfig) (LLMClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	switch config.Provider {
	case ProviderOpenAI:
		return NewOpenAIClient(config)
	case ProviderClaude:
		return NewClaudeClient(config)
	case ProviderLocal:
		return NewLocalClient(config)
	case ProviderCustom:
		return NewCustomClient(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// GetDefaultConfig returns default configuration for a provider
func GetDefaultConfig(provider Provider, apiKey string) *LLMConfig {
	config := &LLMConfig{
		Provider:    provider,
		APIKey:      apiKey,
		Temperature: 0.7,
		MaxTokens:   4096,
		Timeout:     60,
	}

	switch provider {
	case ProviderOpenAI:
		config.Model = "gpt-4-turbo-preview"
		config.BaseURL = "https://api.openai.com/v1"
	case ProviderClaude:
		config.Model = "claude-3-5-sonnet-20241022"
		config.BaseURL = "https://api.anthropic.com/v1"
	case ProviderLocal:
		config.Model = "llama-3.1-70b"
		config.BaseURL = "http://localhost:11434" // Ollama default
	}

	return config
}
