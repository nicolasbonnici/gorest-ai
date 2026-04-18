package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// Client implements the Google Gemini provider
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Gemini client
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "gemini"
}

// Chat sends a chat completion request to Gemini
func (c *Client) Chat(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, error) {
	// Map to Gemini API format
	apiReq := mapChatRequest(req)

	// Marshal request
	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with API key
	url := fmt.Sprintf("%s/v1/models/%s:generateContent?key=%s", c.baseURL, req.Model, c.apiKey)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var apiResp GeminiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Map to common response format
	return mapChatResponse(&apiResp, req.Model), nil
}

// ChatStream sends a streaming chat request
func (c *Client) ChatStream(ctx context.Context, req *providers.ChatRequest) (<-chan providers.StreamChunk, <-chan error) {
	chunkChan := make(chan providers.StreamChunk)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		// Map to Gemini API format
		apiReq := mapChatRequest(req)

		// Marshal request
		body, err := json.Marshal(apiReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Build URL with streaming endpoint
		url := fmt.Sprintf("%s/v1/models/%s:streamGenerateContent?key=%s", c.baseURL, req.Model, c.apiKey)

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := c.client.Do(httpReq)
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

		// Process stream
		if err := processStream(resp.Body, chunkChan); err != nil {
			errChan <- err
		}
	}()

	return chunkChan, errChan
}

// CountTokens estimates token count for messages
func (c *Client) CountTokens(ctx context.Context, messages []providers.Message) (int, error) {
	// Gemini has a token counting API, but for simplicity, use approximation
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
		Model:       "gemini-pro",
		Temperature: 0.0,
		MaxTokens:   10,
	}

	_, err := c.Chat(ctx, req)
	return err
}
