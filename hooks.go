package ai

import (
	"fmt"

	"github.com/nicolasbonnici/gorest/hooks"
)

func RegisterHooks() {
	hooks.Register(hooks.BeforeCreate, "ai_providers", validateProviderBeforeCreate)
	hooks.Register(hooks.BeforeUpdate, "ai_providers", validateProviderBeforeUpdate)
	hooks.Register(hooks.BeforeDelete, "ai_providers", validateProviderBeforeDelete)
	hooks.Register(hooks.BeforeCreate, "ai_requests", validateRequestBeforeCreate)
	hooks.Register(hooks.BeforeCreate, "ai_cache", validateCacheBeforeCreate)
	hooks.Register(hooks.BeforeCreate, "ai_quotas", validateQuotaBeforeCreate)
	hooks.Register(hooks.BeforeUpdate, "ai_quotas", validateQuotaBeforeUpdate)
}

func validateProviderBeforeCreate(data interface{}) error {
	provider, ok := data.(*AIProvider)
	if !ok {
		return fmt.Errorf("invalid data type for provider")
	}

	validProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
		"gemini":    true,
		"mistral":   true,
	}

	if !validProviders[provider.Name] {
		return fmt.Errorf("invalid provider name: %s", provider.Name)
	}

	if provider.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if provider.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}

	if provider.Temperature < 0 || provider.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if provider.RateLimit < 0 {
		return fmt.Errorf("rate_limit must be non-negative")
	}

	return nil
}

func validateProviderBeforeUpdate(data interface{}) error {
	provider, ok := data.(*AIProvider)
	if !ok {
		return fmt.Errorf("invalid data type for provider")
	}

	if provider.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}

	if provider.Temperature < 0 || provider.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if provider.RateLimit < 0 {
		return fmt.Errorf("rate_limit must be non-negative")
	}

	return nil
}

func validateProviderBeforeDelete(data interface{}) error {
	return nil
}

func validateRequestBeforeCreate(data interface{}) error {
	request, ok := data.(*AIRequest)
	if !ok {
		return fmt.Errorf("invalid data type for request")
	}

	if request.ProviderName == "" {
		return fmt.Errorf("provider_name is required")
	}

	if request.Model == "" {
		return fmt.Errorf("model is required")
	}

	if request.RequestType == "" {
		return fmt.Errorf("request_type is required")
	}

	validStatuses := map[string]bool{
		"success": true,
		"error":   true,
		"cached":  true,
	}

	if !validStatuses[request.Status] {
		return fmt.Errorf("invalid status: %s", request.Status)
	}

	return nil
}

func validateCacheBeforeCreate(data interface{}) error {
	cache, ok := data.(*AICache)
	if !ok {
		return fmt.Errorf("invalid data type for cache")
	}

	if cache.CacheKey == "" {
		return fmt.Errorf("cache_key is required")
	}

	if cache.ProviderName == "" {
		return fmt.Errorf("provider_name is required")
	}

	if cache.Model == "" {
		return fmt.Errorf("model is required")
	}

	if cache.ResponseText == "" {
		return fmt.Errorf("response_text is required")
	}

	return nil
}

func validateQuotaBeforeCreate(data interface{}) error {
	quota, ok := data.(*AIQuota)
	if !ok {
		return fmt.Errorf("invalid data type for quota")
	}

	if quota.DailyLimit <= 0 {
		return fmt.Errorf("daily_limit must be greater than 0")
	}

	if quota.MonthlyLimit <= 0 {
		return fmt.Errorf("monthly_limit must be greater than 0")
	}

	if quota.DailyTokenLimit <= 0 {
		return fmt.Errorf("daily_token_limit must be greater than 0")
	}

	if quota.MonthlyTokenLimit <= 0 {
		return fmt.Errorf("monthly_token_limit must be greater than 0")
	}

	if quota.DailyUsed > quota.DailyLimit {
		return fmt.Errorf("daily_used cannot exceed daily_limit")
	}

	if quota.MonthlyUsed > quota.MonthlyLimit {
		return fmt.Errorf("monthly_used cannot exceed monthly_limit")
	}

	return nil
}

func validateQuotaBeforeUpdate(data interface{}) error {
	return validateQuotaBeforeCreate(data)
}
