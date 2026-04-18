package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// processStream processes the SSE stream from OpenAI API
func processStream(reader io.Reader, chunkChan chan<- providers.StreamChunk) error {
	scanner := bufio.NewScanner(reader)
	fullContent := ""

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse SSE data
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			// Send final chunk with Done flag
			chunkChan <- providers.StreamChunk{
				Delta: "",
				Done:  true,
				Metrics: &providers.TokenMetrics{
					// OpenAI doesn't provide token metrics in streaming mode
					// You would need to estimate or make a separate API call
					PromptTokens:     0,
					CompletionTokens: 0,
					TotalTokens:      0,
				},
			}
			break
		}

		// Parse chunk
		var chunk OpenAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Extract delta content
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta
				chunkChan <- providers.StreamChunk{
					Delta: delta,
					Done:  false,
				}
			}

			// Check if finished
			if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
				chunkChan <- providers.StreamChunk{
					Delta: "",
					Done:  true,
					Metrics: &providers.TokenMetrics{
						// Estimate token count from full content
						CompletionTokens: len(fullContent) / 4,
					},
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream scanner error: %w", err)
	}

	return nil
}
