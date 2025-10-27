package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type claudeClient struct {
	config     *LLMConfig
	httpClient *http.Client
}

// NewClaudeClient creates a new Claude (Anthropic) client
func NewClaudeClient(config *LLMConfig) (LLMClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}

	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}

	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &claudeClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *claudeClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Separate system message from other messages
	systemMsg := ""
	messages := []map[string]interface{}{}

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
		} else {
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	// Build Claude request
	claudeReq := map[string]interface{}{
		"model":      c.getModel(req),
		"messages":   messages,
		"max_tokens": c.getMaxTokens(req),
	}

	if systemMsg != "" {
		claudeReq["system"] = systemMsg
	}

	if req.Temperature > 0 {
		claudeReq["temperature"] = req.Temperature
	} else if c.config.Temperature > 0 {
		claudeReq["temperature"] = c.config.Temperature
	}

	// Make HTTP request
	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Extract text from content blocks
	var content string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &ChatResponse{
		Content:      content,
		FinishReason: claudeResp.StopReason,
		Model:        claudeResp.Model,
		Usage: &TokenUsage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

func (c *claudeClient) ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error {
	// TODO: Implement streaming support for Claude
	// For MVP, we'll use non-streaming and call callback once
	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}
	return callback(resp.Content)
}

func (c *claudeClient) GetProvider() Provider {
	return ProviderClaude
}

func (c *claudeClient) GetModel() string {
	return c.config.Model
}

func (c *claudeClient) getModel(req *ChatRequest) string {
	if req.Model != "" {
		return req.Model
	}
	return c.config.Model
}

func (c *claudeClient) getMaxTokens(req *ChatRequest) int {
	if req.MaxTokens > 0 {
		return req.MaxTokens
	}
	if c.config.MaxTokens > 0 {
		return c.config.MaxTokens
	}
	return 4096 // Claude default
}
