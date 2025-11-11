package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type claudeClient struct {
	config     *LLMConfig
	httpClient *http.Client
}

// NewClaudeClient creates a new Claude (Anthropic) client
func NewClaudeClient(config *LLMConfig) (LLMClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("claude api key is required")
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

	// Build Claude request with streaming enabled
	claudeReq := map[string]interface{}{
		"model":      c.getModel(req),
		"messages":   messages,
		"max_tokens": c.getMaxTokens(req),
		"stream":     true, // Enable streaming
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
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Read SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			break
		}

		// Parse the event
		var event struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			Delta struct {
				Type       string `json:"type"`
				Text       string `json:"text"`
				StopReason string `json:"stop_reason"`
			} `json:"delta"`
			ContentBlock struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content_block"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// Skip malformed events
			continue
		}

		// Handle different event types
		switch event.Type {
		case "content_block_start":
			// Initial content block, may have text
			if event.ContentBlock.Text != "" {
				if err := callback(event.ContentBlock.Text); err != nil {
					return err
				}
			}
		case "content_block_delta":
			// Text chunk received
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				if err := callback(event.Delta.Text); err != nil {
					return err
				}
			}
		case "message_delta":
			// Message metadata update (e.g., stop_reason)
			// We can ignore this for now
			continue
		case "message_stop":
			// End of stream - exit the loop
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream reading error: %w", err)
	}

	return nil
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
