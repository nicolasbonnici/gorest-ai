package anthropic

import (
	"github.com/nicolasbonnici/gorest-ai/providers"
)

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []AnthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
	System      string              `json:"system,omitempty"`
}

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`    // "user" or "assistant" (no system)
	Content string `json:"content"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContent      `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence *string                 `json:"stop_sequence"`
	Usage        AnthropicUsage          `json:"usage"`
}

// AnthropicContent represents content in the response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage in the response
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicStreamEvent represents a streaming event
type AnthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Message      *AnthropicResponse      `json:"message,omitempty"`
	Index        int                     `json:"index,omitempty"`
	Delta        *AnthropicDelta         `json:"delta,omitempty"`
	Usage        *AnthropicUsage         `json:"usage,omitempty"`
}

// AnthropicDelta represents a content delta in streaming
type AnthropicDelta struct {
	Type       string  `json:"type"`
	Text       string  `json:"text,omitempty"`
	StopReason *string `json:"stop_reason,omitempty"`
}

// mapChatRequest converts a common chat request to Anthropic format
func mapChatRequest(req *providers.ChatRequest) *AnthropicRequest {
	apiReq := &AnthropicRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// Extract system message and convert user/assistant messages
	var systemMessage string
	messages := make([]AnthropicMessage, 0)

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			// Anthropic uses a separate system field
			if systemMessage != "" {
				systemMessage += "\n\n" + msg.Content
			} else {
				systemMessage = msg.Content
			}
		} else {
			messages = append(messages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	apiReq.System = systemMessage
	apiReq.Messages = messages

	return apiReq
}

// mapChatResponse converts an Anthropic response to common format
func mapChatResponse(resp *AnthropicResponse) *providers.ChatResponse {
	// Extract text from content blocks
	text := ""
	for _, content := range resp.Content {
		if content.Type == "text" {
			text += content.Text
		}
	}

	return &providers.ChatResponse{
		ID:               resp.ID,
		Content:          text,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.InputTokens,
		CompletionTokens: resp.Usage.OutputTokens,
		TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		FinishReason:     resp.StopReason,
	}
}
