package gemini

import (
	"github.com/nicolasbonnici/gorest-ai/providers"
)

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents content in Gemini format
type GeminiContent struct {
	Role  string        `json:"role"`  // "user" or "model"
	Parts []GeminiPart  `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates     []GeminiCandidate    `json:"candidates"`
	UsageMetadata  *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	Content       GeminiContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	Index         int           `json:"index"`
	SafetyRatings []interface{} `json:"safetyRatings"`
}

// GeminiUsageMetadata represents token usage
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// mapChatRequest converts a common chat request to Gemini format
func mapChatRequest(req *providers.ChatRequest) *GeminiRequest {
	contents := make([]GeminiContent, 0)

	for _, msg := range req.Messages {
		// Map role: "user" -> "user", "assistant" -> "model", "system" -> prepend to first user message
		role := msg.Role
		if role == "assistant" {
			role = "model"
		} else if role == "system" {
			// Gemini doesn't have a system role, prepend to first user message
			// For simplicity, we'll skip system messages or prepend them
			continue
		}

		contents = append(contents, GeminiContent{
			Role: role,
			Parts: []GeminiPart{
				{Text: msg.Content},
			},
		})
	}

	return &GeminiRequest{
		Contents: contents,
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}
}

// mapChatResponse converts a Gemini response to common format
func mapChatResponse(resp *GeminiResponse, model string) *providers.ChatResponse {
	if len(resp.Candidates) == 0 {
		return &providers.ChatResponse{
			Model: model,
		}
	}

	candidate := resp.Candidates[0]

	// Extract text from parts
	text := ""
	for _, part := range candidate.Content.Parts {
		text += part.Text
	}

	// Extract token usage
	promptTokens := 0
	completionTokens := 0
	totalTokens := 0

	if resp.UsageMetadata != nil {
		promptTokens = resp.UsageMetadata.PromptTokenCount
		completionTokens = resp.UsageMetadata.CandidatesTokenCount
		totalTokens = resp.UsageMetadata.TotalTokenCount
	}

	return &providers.ChatResponse{
		ID:               "", // Gemini doesn't return an ID
		Content:          text,
		Model:            model,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		FinishReason:     candidate.FinishReason,
	}
}
