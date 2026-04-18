package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "anthropic", config.DefaultProvider)
	assert.Equal(t, 4, len(config.EnabledProviders))
	assert.True(t, config.EnableCache)
	assert.Equal(t, 3600, config.CacheTTL)
	assert.True(t, config.EnableFallback)
	assert.True(t, config.EnableQuota)
	assert.Equal(t, 4096, config.MaxTokens)
	assert.Equal(t, 0.7, config.DefaultTemperature)
	assert.Equal(t, 60, config.RateLimitPerMin)
	assert.Equal(t, 30, config.RequestTimeout)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "Empty default provider",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "Invalid provider in enabled list",
			config: Config{
				DefaultProvider:  "anthropic",
				EnabledProviders: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "Default provider not in enabled list",
			config: Config{
				DefaultProvider:  "anthropic",
				EnabledProviders: []string{"openai"},
			},
			wantErr: true,
		},
		{
			name: "Missing API key for enabled provider",
			config: Config{
				DefaultProvider:  "anthropic",
				EnabledProviders: []string{"anthropic"},
				AnthropicAPIKey:  "",
			},
			wantErr: true,
		},
		{
			name: "Invalid max tokens",
			config: Config{
				DefaultProvider:  "anthropic",
				EnabledProviders: []string{"anthropic"},
				AnthropicAPIKey:  "test-key",
				MaxTokens:        -1,
			},
			wantErr: true,
		},
		{
			name: "Invalid temperature",
			config: Config{
				DefaultProvider:  "anthropic",
				EnabledProviders: []string{"anthropic"},
				AnthropicAPIKey:  "test-key",
				MaxTokens:        100,
				DefaultTemperature: 3.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigGetProviderAPIKey(t *testing.T) {
	config := Config{
		AnthropicAPIKey: "anthropic-key",
		OpenAIAPIKey:    "openai-key",
		GeminiAPIKey:    "gemini-key",
		MistralAPIKey:   "mistral-key",
	}

	assert.Equal(t, "anthropic-key", config.GetProviderAPIKey("anthropic"))
	assert.Equal(t, "openai-key", config.GetProviderAPIKey("openai"))
	assert.Equal(t, "gemini-key", config.GetProviderAPIKey("gemini"))
	assert.Equal(t, "mistral-key", config.GetProviderAPIKey("mistral"))
	assert.Equal(t, "", config.GetProviderAPIKey("unknown"))
}

func TestConfigGetProviderBaseURL(t *testing.T) {
	config := Config{
		AnthropicBaseURL: "https://custom-anthropic.com",
	}

	assert.Equal(t, "https://custom-anthropic.com", config.GetProviderBaseURL("anthropic"))
	assert.Equal(t, "https://api.openai.com/v1", config.GetProviderBaseURL("openai"))
	assert.Equal(t, "https://generativelanguage.googleapis.com", config.GetProviderBaseURL("gemini"))
	assert.Equal(t, "https://api.mistral.ai", config.GetProviderBaseURL("mistral"))
}

func TestConfigIsProviderEnabled(t *testing.T) {
	config := Config{
		EnabledProviders: []string{"anthropic", "openai"},
	}

	assert.True(t, config.IsProviderEnabled("anthropic"))
	assert.True(t, config.IsProviderEnabled("openai"))
	assert.False(t, config.IsProviderEnabled("gemini"))
	assert.False(t, config.IsProviderEnabled("mistral"))
}
