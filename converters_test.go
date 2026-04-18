package ai

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToProviderResponseDTO(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	baseURL := "https://api.example.com"

	provider := &AIProvider{
		ID:          id,
		Name:        "anthropic",
		DisplayName: "Claude",
		APIKey:      "secret-key",
		BaseURL:     &baseURL,
		Enabled:     true,
		Priority:    1,
		MaxTokens:   4096,
		Temperature: 0.7,
		RateLimit:   60,
		CreatedAt:   now,
		UpdatedAt:   &now,
	}

	dto := ToProviderResponseDTO(provider)

	assert.Equal(t, id, dto.ID)
	assert.Equal(t, "anthropic", dto.Name)
	assert.Equal(t, "Claude", dto.DisplayName)
	assert.Equal(t, &baseURL, dto.BaseURL)
	assert.True(t, dto.Enabled)
	assert.Equal(t, 1, dto.Priority)
	assert.Equal(t, 4096, dto.MaxTokens)
	assert.Equal(t, 0.7, dto.Temperature)
	assert.Equal(t, 60, dto.RateLimit)
	assert.True(t, dto.HasAPIKey)
	assert.Equal(t, now, dto.CreatedAt)
	assert.Equal(t, &now, dto.UpdatedAt)
}

func TestToProviderModel(t *testing.T) {
	baseURL := "https://api.example.com"
	dto := &ProviderCreateDTO{
		Name:        "openai",
		DisplayName: "OpenAI",
		APIKey:      "secret-key",
		BaseURL:     &baseURL,
		Enabled:     true,
		Priority:    2,
		MaxTokens:   8192,
		Temperature: 0.8,
		RateLimit:   100,
	}

	model := ToProviderModel(dto)

	assert.NotEqual(t, uuid.Nil, model.ID)
	assert.Equal(t, "openai", model.Name)
	assert.Equal(t, "OpenAI", model.DisplayName)
	assert.Equal(t, "secret-key", model.APIKey)
	assert.Equal(t, &baseURL, model.BaseURL)
	assert.True(t, model.Enabled)
	assert.Equal(t, 2, model.Priority)
	assert.Equal(t, 8192, model.MaxTokens)
	assert.Equal(t, 0.8, model.Temperature)
	assert.Equal(t, 100, model.RateLimit)
	assert.Equal(t, 0.0, model.CostPerToken)
	assert.False(t, model.CreatedAt.IsZero())
	assert.Nil(t, model.UpdatedAt)
}

func TestUpdateProviderModel(t *testing.T) {
	provider := &AIProvider{
		ID:          uuid.New(),
		Name:        "anthropic",
		DisplayName: "Claude",
		APIKey:      "old-key",
		Enabled:     false,
		Priority:    1,
		MaxTokens:   2048,
		Temperature: 0.5,
		RateLimit:   30,
		CreatedAt:   time.Now(),
	}

	newDisplayName := "Claude AI"
	newAPIKey := "new-key"
	newEnabled := true
	newPriority := 2
	newMaxTokens := 4096
	newTemperature := 0.7
	newRateLimit := 60

	dto := &ProviderUpdateDTO{
		DisplayName: &newDisplayName,
		APIKey:      &newAPIKey,
		Enabled:     &newEnabled,
		Priority:    &newPriority,
		MaxTokens:   &newMaxTokens,
		Temperature: &newTemperature,
		RateLimit:   &newRateLimit,
	}

	UpdateProviderModel(provider, dto)

	assert.Equal(t, newDisplayName, provider.DisplayName)
	assert.Equal(t, newAPIKey, provider.APIKey)
	assert.True(t, provider.Enabled)
	assert.Equal(t, 2, provider.Priority)
	assert.Equal(t, 4096, provider.MaxTokens)
	assert.Equal(t, 0.7, provider.Temperature)
	assert.Equal(t, 60, provider.RateLimit)
	assert.NotNil(t, provider.UpdatedAt)
}

func TestToChatResponseDTO(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	responseText := "Hello, world!"

	request := &AIRequest{
		ID:               id,
		ProviderName:     "anthropic",
		Model:            "claude-3-5-sonnet",
		ResponseText:     &responseText,
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
		Cost:             0.001,
		DurationMs:       1234,
		Cached:           false,
		CreatedAt:        now,
	}

	dto := ToChatResponseDTO(request)

	assert.Equal(t, id, dto.ID)
	assert.Equal(t, "anthropic", dto.Provider)
	assert.Equal(t, "claude-3-5-sonnet", dto.Model)
	assert.Equal(t, "Hello, world!", dto.Content)
	assert.Equal(t, 10, dto.PromptTokens)
	assert.Equal(t, 20, dto.CompletionTokens)
	assert.Equal(t, 30, dto.TotalTokens)
	assert.Equal(t, 0.001, dto.Cost)
	assert.Equal(t, 1234, dto.DurationMs)
	assert.False(t, dto.Cached)
	assert.Equal(t, now, dto.CreatedAt)
}

func TestToQuotaStatusDTO(t *testing.T) {
	now := time.Now()
	resetDaily := now.Add(24 * time.Hour)
	resetMonthly := now.Add(720 * time.Hour)

	quota := &AIQuota{
		DailyLimit:        100,
		MonthlyLimit:      1000,
		DailyTokenLimit:   10000,
		MonthlyTokenLimit: 100000,
		DailyUsed:         30,
		MonthlyUsed:       300,
		DailyTokensUsed:   3000,
		MonthlyTokensUsed: 30000,
		ResetDaily:        resetDaily,
		ResetMonthly:      resetMonthly,
	}

	dto := ToQuotaStatusDTO(quota)

	assert.Equal(t, 100, dto.DailyLimit)
	assert.Equal(t, 1000, dto.MonthlyLimit)
	assert.Equal(t, 10000, dto.DailyTokenLimit)
	assert.Equal(t, 100000, dto.MonthlyTokenLimit)
	assert.Equal(t, 30, dto.DailyUsed)
	assert.Equal(t, 300, dto.MonthlyUsed)
	assert.Equal(t, 3000, dto.DailyTokensUsed)
	assert.Equal(t, 30000, dto.MonthlyTokensUsed)
	assert.Equal(t, 70, dto.DailyRemaining)
	assert.Equal(t, 700, dto.MonthlyRemaining)
	assert.Equal(t, 7000, dto.DailyTokensRemaining)
	assert.Equal(t, 70000, dto.MonthlyTokensRemaining)
	assert.Equal(t, resetDaily, dto.ResetDaily)
	assert.Equal(t, resetMonthly, dto.ResetMonthly)
}
