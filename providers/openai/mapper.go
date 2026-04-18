package openai

import (
	"github.com/nicolasbonnici/gorest-ai/providers"
)

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice represents a choice in the response
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents token usage in the response
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamChunk represents a streaming chunk
type OpenAIStreamChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
}

// OpenAIStreamChoice represents a choice in a streaming chunk
type OpenAIStreamChoice struct {
	Index        int                `json:"index"`
	Delta        OpenAIMessageDelta `json:"delta"`
	FinishReason *string            `json:"finish_reason"`
}

// OpenAIMessageDelta represents a message delta in streaming
type OpenAIMessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// mapChatRequest converts a common chat request to OpenAI format
func mapChatRequest(req *providers.ChatRequest) *OpenAIRequest {
	messages := make([]OpenAIMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return &OpenAIRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}
}

// mapChatResponse converts an OpenAI response to common format
func mapChatResponse(resp *OpenAIResponse) *providers.ChatResponse {
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
