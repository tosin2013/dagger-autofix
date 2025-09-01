package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/sirupsen/logrus"
)

// LLMProvider represents different LLM providers
type LLMProvider string

const (
	OpenAI    LLMProvider = "openai"
	Anthropic LLMProvider = "anthropic"
	Gemini    LLMProvider = "gemini"
	DeepSeek  LLMProvider = "deepseek"
	LiteLLM   LLMProvider = "litellm"
)

// LLMClient provides unified interface for multiple LLM providers
type LLMClient struct {
	provider   LLMProvider
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	config     *LLMConfig
}

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Model       string                 `json:"model"`
	Temperature float64               `json:"temperature"`
	MaxTokens   int                   `json:"max_tokens"`
	Timeout     time.Duration         `json:"timeout"`
	RetryCount  int                   `json:"retry_count"`
	ProviderConfig map[string]interface{} `json:"provider_config"`
}

// LLMRequest represents a request to an LLM
type LLMRequest struct {
	Prompt     string            `json:"prompt"`
	SystemMsg  string            `json:"system_message,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Tools      []LLMTool         `json:"tools,omitempty"`
	Model      string            `json:"model,omitempty"`
}

// LLMResponse represents a response from an LLM
type LLMResponse struct {
	Content      string                 `json:"content"`
	ToolCalls    []LLMToolCall          `json:"tool_calls,omitempty"`
	Usage        *LLMUsage              `json:"usage,omitempty"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LLMTool represents a tool that can be called by the LLM
type LLMTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"` // JSON schema
}

// LLMToolCall represents a tool call made by the LLM
type LLMToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	CallID    string                 `json:"call_id,omitempty"`
}

// LLMUsage represents token usage information
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewLLMClient creates a new LLM client for the specified provider
func NewLLMClient(ctx context.Context, provider LLMProvider, apiKey *dagger.Secret) (*LLMClient, error) {
	keyStr, err := apiKey.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	config := getDefaultConfig(provider)
	baseURL := getProviderBaseURL(provider)

	client := &LLMClient{
		provider: provider,
		apiKey:   keyStr,
		baseURL:  baseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logrus.New(),
		config: config,
	}

	// Test connection
	if err := client.testConnection(ctx); err != nil {
		return nil, fmt.Errorf("LLM client connection test failed: %w", err)
	}

	return client, nil
}

// Chat sends a chat request to the LLM and returns the response
func (c *LLMClient) Chat(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	start := time.Now()
	defer func() {
		c.logger.WithFields(logrus.Fields{
			"provider": c.provider,
			"duration": time.Since(start),
			"model":    request.Model,
		}).Debug("LLM request completed")
	}()

	switch c.provider {
	case OpenAI:
		return c.chatOpenAI(ctx, request)
	case Anthropic:
		return c.chatAnthropic(ctx, request)
	case Gemini:
		return c.chatGemini(ctx, request)
	case DeepSeek:
		return c.chatDeepSeek(ctx, request)
	case LiteLLM:
		return c.chatLiteLLM(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", c.provider)
	}
}

// WithModel sets the model to use for requests
func (c *LLMClient) WithModel(model string) *LLMClient {
	c.config.Model = model
	return c
}

// WithTemperature sets the temperature for requests
func (c *LLMClient) WithTemperature(temp float64) *LLMClient {
	c.config.Temperature = temp
	return c
}

// WithMaxTokens sets the maximum tokens for responses
func (c *LLMClient) WithMaxTokens(tokens int) *LLMClient {
	c.config.MaxTokens = tokens
	return c
}

// Provider-specific implementations

func (c *LLMClient) chatOpenAI(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	model := request.Model
	if model == "" {
		model = c.config.Model
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": request.SystemMsg,
			},
			{
				"role":    "user",
				"content": request.Prompt,
			},
		},
		"temperature": c.config.Temperature,
		"max_tokens":  c.config.MaxTokens,
	}

	// Add tools if provided
	if len(request.Tools) > 0 {
		tools := make([]map[string]interface{}, len(request.Tools))
		for i, tool := range request.Tools {
			tools[i] = map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  json.RawMessage(tool.Parameters),
				},
			}
		}
		payload["tools"] = tools
		payload["tool_choice"] = "auto"
	}

	resp, err := c.makeRequest(ctx, "POST", "/v1/chat/completions", payload)
	if err != nil {
		return nil, err
	}

	return c.parseOpenAIResponse(resp)
}

func (c *LLMClient) chatAnthropic(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	model := request.Model
	if model == "" {
		model = c.config.Model
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": request.Prompt,
			},
		},
		"max_tokens": c.config.MaxTokens,
		"temperature": c.config.Temperature,
	}

	if request.SystemMsg != "" {
		payload["system"] = request.SystemMsg
	}

	resp, err := c.makeRequest(ctx, "POST", "/v1/messages", payload)
	if err != nil {
		return nil, err
	}

	return c.parseAnthropicResponse(resp)
}

func (c *LLMClient) chatGemini(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	model := request.Model
	if model == "" {
		model = c.config.Model
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": request.Prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": c.config.Temperature,
			"maxOutputTokens": c.config.MaxTokens,
		},
	}

	if request.SystemMsg != "" {
		payload["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{
					"text": request.SystemMsg,
				},
			},
		}
	}

	url := fmt.Sprintf("/v1beta/models/%s:generateContent", model)
	resp, err := c.makeRequest(ctx, "POST", url, payload)
	if err != nil {
		return nil, err
	}

	return c.parseGeminiResponse(resp)
}

func (c *LLMClient) chatDeepSeek(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	// DeepSeek uses OpenAI-compatible API
	return c.chatOpenAI(ctx, request)
}

func (c *LLMClient) chatLiteLLM(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	// LiteLLM proxy uses OpenAI-compatible API
	return c.chatOpenAI(ctx, request)
}

// Helper methods

func (c *LLMClient) makeRequest(ctx context.Context, method, path string, payload interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers based on provider
	switch c.provider {
	case OpenAI, DeepSeek, LiteLLM:
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
	case Anthropic:
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("anthropic-version", "2023-06-01")
	case Gemini:
		q := req.URL.Query()
		q.Add("key", c.apiKey)
		req.URL.RawQuery = q.Encode()
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

func (c *LLMClient) parseOpenAIResponse(resp map[string]interface{}) (*LLMResponse, error) {
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	response := &LLMResponse{
		Content:      content,
		Provider:     string(c.provider),
		Model:        c.config.Model,
		FinishReason: choice["finish_reason"].(string),
	}

	// Parse usage if available
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		response.Usage = &LLMUsage{
			PromptTokens:     int(usage["prompt_tokens"].(float64)),
			CompletionTokens: int(usage["completion_tokens"].(float64)),
			TotalTokens:      int(usage["total_tokens"].(float64)),
		}
	}

	// Parse tool calls if available
	if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
		for _, tc := range toolCalls {
			toolCall := tc.(map[string]interface{})
			function := toolCall["function"].(map[string]interface{})
			
			var args map[string]interface{}
			if argsStr, ok := function["arguments"].(string); ok {
				json.Unmarshal([]byte(argsStr), &args)
			}

			response.ToolCalls = append(response.ToolCalls, LLMToolCall{
				Name:      function["name"].(string),
				Arguments: args,
				CallID:    toolCall["id"].(string),
			})
		}
	}

	return response, nil
}

func (c *LLMClient) parseAnthropicResponse(resp map[string]interface{}) (*LLMResponse, error) {
	content, ok := resp["content"].([]interface{})
	if !ok || len(content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	firstContent := content[0].(map[string]interface{})
	text := firstContent["text"].(string)

	response := &LLMResponse{
		Content:      text,
		Provider:     string(c.provider),
		Model:        c.config.Model,
		FinishReason: resp["stop_reason"].(string),
	}

	// Parse usage if available
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		response.Usage = &LLMUsage{
			PromptTokens:     int(usage["input_tokens"].(float64)),
			CompletionTokens: int(usage["output_tokens"].(float64)),
		}
		response.Usage.TotalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens
	}

	return response, nil
}

func (c *LLMClient) parseGeminiResponse(resp map[string]interface{}) (*LLMResponse, error) {
	candidates, ok := resp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := candidates[0].(map[string]interface{})
	content := candidate["content"].(map[string]interface{})
	parts := content["parts"].([]interface{})
	text := parts[0].(map[string]interface{})["text"].(string)

	response := &LLMResponse{
		Content:      text,
		Provider:     string(c.provider),
		Model:        c.config.Model,
		FinishReason: candidate["finishReason"].(string),
	}

	return response, nil
}

func (c *LLMClient) testConnection(ctx context.Context) error {
	// Simple test request to verify connectivity
	testReq := &LLMRequest{
		Prompt: "Hello, world!",
		SystemMsg: "You are a helpful assistant. Respond with just 'OK'.",
	}

	_, err := c.Chat(ctx, testReq)
	return err
}

func getDefaultConfig(provider LLMProvider) *LLMConfig {
	switch provider {
	case OpenAI:
		return &LLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	case Anthropic:
		return &LLMConfig{
			Model:       "claude-3-5-sonnet-20241022",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	case Gemini:
		return &LLMConfig{
			Model:       "gemini-2.0-flash-exp",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	case DeepSeek:
		return &LLMConfig{
			Model:       "deepseek-chat",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	case LiteLLM:
		return &LLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	default:
		return &LLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.1,
			MaxTokens:   4000,
			Timeout:     60 * time.Second,
			RetryCount:  3,
		}
	}
}

func getProviderBaseURL(provider LLMProvider) string {
	switch provider {
	case OpenAI:
		return "https://api.openai.com"
	case Anthropic:
		return "https://api.anthropic.com"
	case Gemini:
		return "https://generativelanguage.googleapis.com"
	case DeepSeek:
		return "https://api.deepseek.com"
	case LiteLLM:
		return "http://localhost:4000" // Default LiteLLM proxy URL
	default:
		return "https://api.openai.com"
	}
}
