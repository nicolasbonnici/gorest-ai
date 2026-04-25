package ai

import (
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
)

type Config struct {
	Database           database.Database `json:"-" yaml:"-"`
	DefaultProvider    string            `json:"default_provider" yaml:"default_provider"`
	EnabledProviders   []string          `json:"enabled_providers" yaml:"enabled_providers"`
	AnthropicAPIKey    string            `json:"anthropic_api_key" yaml:"anthropic_api_key"`
	AnthropicBaseURL   string            `json:"anthropic_base_url" yaml:"anthropic_base_url"`
	OpenAIAPIKey       string            `json:"openai_api_key" yaml:"openai_api_key"`
	OpenAIBaseURL      string            `json:"openai_base_url" yaml:"openai_base_url"`
	GeminiAPIKey       string            `json:"gemini_api_key" yaml:"gemini_api_key"`
	GeminiBaseURL      string            `json:"gemini_base_url" yaml:"gemini_base_url"`
	MistralAPIKey      string            `json:"mistral_api_key" yaml:"mistral_api_key"`
	MistralBaseURL     string            `json:"mistral_base_url" yaml:"mistral_base_url"`
	EnableCache        bool              `json:"enable_cache" yaml:"enable_cache"`
	CacheTTL           int               `json:"cache_ttl" yaml:"cache_ttl"`
	EnableFallback     bool              `json:"enable_fallback" yaml:"enable_fallback"`
	EnableQuota        bool              `json:"enable_quota" yaml:"enable_quota"`
	MaxTokens          int               `json:"max_tokens" yaml:"max_tokens"`
	DefaultTemperature float64           `json:"default_temperature" yaml:"default_temperature"`
	RateLimitPerMin    int               `json:"rate_limit_per_min" yaml:"rate_limit_per_min"`
	RequestTimeout     int               `json:"request_timeout" yaml:"request_timeout"`
	PaginationLimit    int               `json:"pagination_limit" yaml:"pagination_limit"`
	MaxPaginationLimit int               `json:"max_pagination_limit" yaml:"max_pagination_limit"`
	RequireAuth        bool              `json:"require_auth" yaml:"require_auth"`
	AllowAnonymous     bool              `json:"allow_anonymous" yaml:"allow_anonymous"`
	EnableAudit        bool              `json:"enable_audit" yaml:"enable_audit"`
	RetainAuditDays    int               `json:"retain_audit_days" yaml:"retain_audit_days"`
}

func DefaultConfig() Config {
	return Config{
		DefaultProvider:    "anthropic",
		EnabledProviders:   []string{"anthropic", "openai", "gemini", "mistral"},
		EnableCache:        true,
		CacheTTL:           3600,
		EnableFallback:     true,
		EnableQuota:        true,
		MaxTokens:          4096,
		DefaultTemperature: 0.7,
		RateLimitPerMin:    60,
		RequestTimeout:     30,
		PaginationLimit:    20,
		MaxPaginationLimit: 100,
		RequireAuth:        true,
		AllowAnonymous:     false,
		EnableAudit:        true,
		RetainAuditDays:    90,
	}
}

func (c *Config) Validate() error {
	if c.DefaultProvider == "" {
		return fmt.Errorf("default_provider is required")
	}

	if len(c.EnabledProviders) > 0 {
		found := false
		for _, provider := range c.EnabledProviders {
			if provider == c.DefaultProvider {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default_provider '%s' must be in enabled_providers list", c.DefaultProvider)
		}
	}

	validProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
		"gemini":    true,
		"mistral":   true,
	}

	for _, provider := range c.EnabledProviders {
		if !validProviders[provider] {
			return fmt.Errorf("invalid provider '%s', must be one of: anthropic, openai, gemini, mistral", provider)
		}

		switch provider {
		case "anthropic":
			if c.AnthropicAPIKey == "" {
				return fmt.Errorf("anthropic_api_key is required when anthropic provider is enabled")
			}
		case "openai":
			if c.OpenAIAPIKey == "" {
				return fmt.Errorf("openai_api_key is required when openai provider is enabled")
			}
		case "gemini":
			if c.GeminiAPIKey == "" {
				return fmt.Errorf("gemini_api_key is required when gemini provider is enabled")
			}
		case "mistral":
			if c.MistralAPIKey == "" {
				return fmt.Errorf("mistral_api_key is required when mistral provider is enabled")
			}
		}
	}

	if c.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}

	if c.DefaultTemperature < 0 || c.DefaultTemperature > 2 {
		return fmt.Errorf("default_temperature must be between 0 and 2")
	}

	if c.RateLimitPerMin < 0 {
		return fmt.Errorf("rate_limit_per_min must be non-negative")
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request_timeout must be greater than 0")
	}

	if c.EnableCache && c.CacheTTL <= 0 {
		return fmt.Errorf("cache_ttl must be greater than 0 when caching is enabled")
	}

	if c.PaginationLimit <= 0 {
		return fmt.Errorf("pagination_limit must be greater than 0")
	}

	if c.MaxPaginationLimit < c.PaginationLimit {
		return fmt.Errorf("max_pagination_limit must be greater than or equal to pagination_limit")
	}

	if c.EnableAudit && c.RetainAuditDays <= 0 {
		return fmt.Errorf("retain_audit_days must be greater than 0 when audit is enabled")
	}

	if c.Database == nil {
		return fmt.Errorf("database connection is required")
	}

	return nil
}

func (c *Config) GetProviderAPIKey(provider string) string {
	switch provider {
	case "anthropic":
		return c.AnthropicAPIKey
	case "openai":
		return c.OpenAIAPIKey
	case "gemini":
		return c.GeminiAPIKey
	case "mistral":
		return c.MistralAPIKey
	default:
		return ""
	}
}

func (c *Config) GetProviderBaseURL(provider string) string {
	switch provider {
	case "anthropic":
		if c.AnthropicBaseURL != "" {
			return c.AnthropicBaseURL
		}
		return "https://api.anthropic.com"
	case "openai":
		if c.OpenAIBaseURL != "" {
			return c.OpenAIBaseURL
		}
		return "https://api.openai.com/v1"
	case "gemini":
		if c.GeminiBaseURL != "" {
			return c.GeminiBaseURL
		}
		return "https://generativelanguage.googleapis.com"
	case "mistral":
		if c.MistralBaseURL != "" {
			return c.MistralBaseURL
		}
		return "https://api.mistral.ai"
	default:
		return ""
	}
}

func (c *Config) IsProviderEnabled(provider string) bool {
	for _, p := range c.EnabledProviders {
		if p == provider {
			return true
		}
	}
	return false
}
