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

type openAIClient struct {
	config     *LLMConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config *LLMConfig) (LLMClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-4-turbo-preview"
	}

	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &openAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *openAIClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Build OpenAI request
	openAIReq := map[string]interface{}{
		"model":    c.getModel(req),
		"messages": c.convertMessages(req.Messages),
	}

	if req.Temperature > 0 {
		openAIReq["temperature"] = req.Temperature
	} else if c.config.Temperature > 0 {
		openAIReq["temperature"] = c.config.Temperature
	}

	if req.MaxTokens > 0 {
		openAIReq["max_tokens"] = req.MaxTokens
	} else if c.config.MaxTokens > 0 {
		openAIReq["max_tokens"] = c.config.MaxTokens
	}

	// Make HTTP request
	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatResponse{
		Content:      openAIResp.Choices[0].Message.Content,
		FinishReason: openAIResp.Choices[0].FinishReason,
		Model:        openAIResp.Model,
		Usage: &TokenUsage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}, nil
}

func (c *openAIClient) ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error {
	// Build OpenAI request with streaming enabled
	openAIReq := map[string]interface{}{
		"model":    c.getModel(req),
		"messages": c.convertMessages(req.Messages),
		"stream":   true, // Enable streaming
	}

	if req.Temperature > 0 {
		openAIReq["temperature"] = req.Temperature
	} else if c.config.Temperature > 0 {
		openAIReq["temperature"] = c.config.Temperature
	}

	if req.MaxTokens > 0 {
		openAIReq["max_tokens"] = req.MaxTokens
	} else if c.config.MaxTokens > 0 {
		openAIReq["max_tokens"] = c.config.MaxTokens
	}

	// Make HTTP request
	body, err := json.Marshal(openAIReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

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

		// Parse the chunk
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Skip malformed chunks
			continue
		}

		// Extract content from delta
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := callback(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}

		// Check for finish
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != nil {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream reading error: %w", err)
	}

	return nil
}

func (c *openAIClient) GetProvider() Provider {
	return ProviderOpenAI
}

func (c *openAIClient) GetModel() string {
	return c.config.Model
}

func (c *openAIClient) getModel(req *ChatRequest) string {
	if req.Model != "" {
		return req.Model
	}
	return c.config.Model
}

func (c *openAIClient) convertMessages(messages []Message) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		converted[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.Name != "" {
			converted[i]["name"] = msg.Name
		}
	}
	return converted
}
