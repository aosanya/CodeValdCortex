package ai

import (
	"context"
	"fmt"
)

type customClient struct {
	config *LLMConfig
}

// NewCustomClient creates a client for custom LLM endpoints
// This is a placeholder for users to implement their own integrations
func NewCustomClient(config *LLMConfig) (LLMClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required for custom client")
	}

	return &customClient{
		config: config,
	}, nil
}

func (c *customClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return nil, fmt.Errorf("custom client not implemented - please implement based on your API")
}

func (c *customClient) ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error {
	return fmt.Errorf("custom client streaming not implemented")
}

func (c *customClient) GetProvider() Provider {
	return ProviderCustom
}

func (c *customClient) GetModel() string {
	return c.config.Model
}
