package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// processStream processes the stream from Gemini API
func processStream(reader io.Reader, chunkChan chan<- providers.StreamChunk) error {
	scanner := bufio.NewScanner(reader)
	var totalPromptTokens, totalCompletionTokens int

	for scanner.Scan() {
		line := scanner.Bytes()

		// Gemini returns JSON objects separated by newlines
		if len(line) == 0 {
			continue
		}

		// Parse response chunk
		var chunk GeminiResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Extract content from candidates
		if len(chunk.Candidates) > 0 {
			candidate := chunk.Candidates[0]

			// Extract text from parts
			text := ""
			for _, part := range candidate.Content.Parts {
				text += part.Text
			}

			if text != "" {
				chunkChan <- providers.StreamChunk{
					Delta: text,
					Done:  false,
				}
			}

			// Update token counts if available
			if chunk.UsageMetadata != nil {
				totalPromptTokens = chunk.UsageMetadata.PromptTokenCount
				totalCompletionTokens = chunk.UsageMetadata.CandidatesTokenCount
			}

			// Check if finished
			if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
				chunkChan <- providers.StreamChunk{
					Delta: "",
					Done:  true,
					Metrics: &providers.TokenMetrics{
						PromptTokens:     totalPromptTokens,
						CompletionTokens: totalCompletionTokens,
						TotalTokens:      totalPromptTokens + totalCompletionTokens,
					},
				}
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream scanner error: %w", err)
	}

	// Send final chunk if not already sent
	chunkChan <- providers.StreamChunk{
		Delta: "",
		Done:  true,
		Metrics: &providers.TokenMetrics{
			PromptTokens:     totalPromptTokens,
			CompletionTokens: totalCompletionTokens,
			TotalTokens:      totalPromptTokens + totalCompletionTokens,
		},
	}

	return nil
}
