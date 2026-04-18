package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nicolasbonnici/gorest-ai/providers"
)

// processStream processes the SSE stream from Anthropic API
func processStream(reader io.Reader, chunkChan chan<- providers.StreamChunk) error {
	scanner := bufio.NewScanner(reader)
	var totalInputTokens, totalOutputTokens int

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

		// Skip ping events
		if data == "ping" || data == "" {
			continue
		}

		// Parse event
		var event AnthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("failed to parse stream event: %w", err)
		}

		// Handle different event types
		switch event.Type {
		case "message_start":
			// Initial message metadata
			if event.Message != nil && event.Message.Usage.InputTokens > 0 {
				totalInputTokens = event.Message.Usage.InputTokens
			}

		case "content_block_start":
			// Content block started, no action needed

		case "content_block_delta":
			// Send content delta
			if event.Delta != nil && event.Delta.Text != "" {
				chunkChan <- providers.StreamChunk{
					Delta: event.Delta.Text,
					Done:  false,
				}
			}

		case "content_block_stop":
			// Content block finished

		case "message_delta":
			// Message delta with usage info
			if event.Usage != nil {
				totalOutputTokens = event.Usage.OutputTokens
			}

		case "message_stop":
			// Stream finished, send final metrics
			chunkChan <- providers.StreamChunk{
				Delta: "",
				Done:  true,
				Metrics: &providers.TokenMetrics{
					PromptTokens:     totalInputTokens,
					CompletionTokens: totalOutputTokens,
					TotalTokens:      totalInputTokens + totalOutputTokens,
				},
			}

		case "error":
			return fmt.Errorf("stream error: %s", data)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream scanner error: %w", err)
	}

	return nil
}

// parseSSELine parses a single SSE line
func parseSSELine(line string) (string, string, error) {
	parts := bytes.SplitN([]byte(line), []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid SSE format")
	}

	field := string(bytes.TrimSpace(parts[0]))
	value := string(bytes.TrimSpace(parts[1]))

	return field, value, nil
}
