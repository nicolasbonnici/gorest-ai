package anthropic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// Client implements the Anthropic (Claude) provider
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Anthropic client
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  providers.NewHTTPClient(60 * time.Second),
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "anthropic"
}

// Chat sends a chat completion request to Claude
func (c *Client) Chat(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, error) {
	apiReq := mapChatRequest(req)

	body, release, err := providers.EncodeJSON(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	defer release()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp AnthropicResponse
	if err := providers.DecodeJSONResponse(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return mapChatResponse(&apiResp), nil
}

// ChatStream sends a streaming chat request
func (c *Client) ChatStream(ctx context.Context, req *providers.ChatRequest) (<-chan providers.StreamChunk, <-chan error) {
	chunkChan := make(chan providers.StreamChunk)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		apiReq := mapChatRequest(req)
		apiReq.Stream = true

		body, release, err := providers.EncodeJSON(apiReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", body)
		if err != nil {
			release()
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := c.client.Do(httpReq)
		release()
		if err != nil {
			errChan <- fmt.Errorf("failed to execute request: %w", err)
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
			return
		}

		// Process SSE stream
		if err := processStream(resp.Body, chunkChan); err != nil {
			errChan <- err
		}
	}()

	return chunkChan, errChan
}

// CountTokens estimates token count for messages
func (c *Client) CountTokens(ctx context.Context, messages []providers.Message) (int, error) {
	// Anthropic doesn't provide a token counting API
	// Use approximate estimation: ~4 characters per token
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	return totalChars / 4, nil
}

// ValidateConfig validates the client configuration
func (c *Client) ValidateConfig() error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

// HealthCheck performs a health check on the API
func (c *Client) HealthCheck(ctx context.Context) error {
	// Send a minimal request to check API availability
	req := &providers.ChatRequest{
		Messages: []providers.Message{
			{Role: "user", Content: "Hi"},
		},
		Model:       "claude-3-haiku-20240307", // Use cheapest model for health check
		Temperature: 0.0,
		MaxTokens:   10,
	}

	_, err := c.Chat(ctx, req)
	return err
}
