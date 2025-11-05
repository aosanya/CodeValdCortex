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

type localClient struct {
	config     *LLMConfig
	httpClient *http.Client
}

// NewLocalClient creates a client for local LLM servers (e.g., Ollama, LM Studio)
func NewLocalClient(config *LLMConfig) (LLMClient, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434" // Ollama default
	}

	if config.Model == "" {
		config.Model = "llama3.1:70b"
	}

	timeout := 120 * time.Second // Local models may be slower
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &localClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *localClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Build request compatible with Ollama API
	localReq := map[string]interface{}{
		"model":    c.getModel(req),
		"messages": c.convertMessages(req.Messages),
		"stream":   false,
	}

	if req.Temperature > 0 {
		localReq["temperature"] = req.Temperature
	}

	// Make HTTP request
	body, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
	var localResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&localResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ChatResponse{
		Content:      localResp.Message.Content,
		FinishReason: "stop",
		Model:        localResp.Model,
	}, nil
}

func (c *localClient) ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error {
	// TODO: Implement streaming support
	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}
	return callback(resp.Content)
}

func (c *localClient) GetProvider() Provider {
	return ProviderLocal
}

func (c *localClient) GetModel() string {
	return c.config.Model
}

func (c *localClient) getModel(req *ChatRequest) string {
	if req.Model != "" {
		return req.Model
	}
	return c.config.Model
}

func (c *localClient) convertMessages(messages []Message) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		converted[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	return converted
}
