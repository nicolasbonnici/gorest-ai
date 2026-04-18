package mistral

import (
	"github.com/nicolasbonnici/gorest-ai/providers"
)

// MistralRequest represents a request to the Mistral API
type MistralRequest struct {
	Model       string           `json:"model"`
	Messages    []MistralMessage `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

// MistralMessage represents a message in Mistral format
type MistralMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"`
}

// MistralResponse represents a response from the Mistral API
type MistralResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []MistralChoice `json:"choices"`
	Usage   MistralUsage    `json:"usage"`
}

// MistralChoice represents a choice in the response
type MistralChoice struct {
	Index        int            `json:"index"`
	Message      MistralMessage `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

// MistralUsage represents token usage in the response
type MistralUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// MistralStreamChunk represents a streaming chunk
type MistralStreamChunk struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []MistralStreamChoice `json:"choices"`
}

// MistralStreamChoice represents a choice in a streaming chunk
type MistralStreamChoice struct {
	Index        int                 `json:"index"`
	Delta        MistralMessageDelta `json:"delta"`
	FinishReason *string             `json:"finish_reason"`
}

// MistralMessageDelta represents a message delta in streaming
type MistralMessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// mapChatRequest converts a common chat request to Mistral format
func mapChatRequest(req *providers.ChatRequest) *MistralRequest {
	messages := make([]MistralMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = MistralMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return &MistralRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}
}

// mapChatResponse converts a Mistral response to common format
func mapChatResponse(resp *MistralResponse) *providers.ChatResponse {
	if len(resp.Choices) == 0 {
		return &providers.ChatResponse{
			ID:    resp.ID,
			Model: resp.Model,
		}
	}

	choice := resp.Choices[0]

	return &providers.ChatResponse{
		ID:               resp.ID,
		Content:          choice.Message.Content,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
		FinishReason:     choice.FinishReason,
	}
}
