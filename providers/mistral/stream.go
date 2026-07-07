package mistral

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// processStream processes the SSE stream from Mistral API
func processStream(reader io.Reader, chunkChan chan<- providers.StreamChunk) error {
	scanner := bufio.NewScanner(reader)
	// Only the streamed content length is needed (for token estimation), so
	// track bytes rather than concatenating the whole body — the latter is
	// quadratic over the number of deltas.
	var completionLen int

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
					// Mistral doesn't provide token metrics in streaming mode; estimate from content.
					CompletionTokens: completionLen / 4,
				},
			}
			break
		}

		// Parse chunk
		var chunk MistralStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Extract delta content
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				completionLen += len(delta)
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
						CompletionTokens: completionLen / 4,
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
