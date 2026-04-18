package ai

import (
	"time"

	"github.com/google/uuid"
)

func ToProviderResponseDTO(provider *AIProvider) *ProviderResponseDTO {
	return &ProviderResponseDTO{
		ID:          provider.ID,
		Name:        provider.Name,
		DisplayName: provider.DisplayName,
		BaseURL:     provider.BaseURL,
		Enabled:     provider.Enabled,
		Priority:    provider.Priority,
		MaxTokens:   provider.MaxTokens,
		Temperature: provider.Temperature,
		RateLimit:   provider.RateLimit,
		HasAPIKey:   provider.APIKey != "",
		CreatedAt:   provider.CreatedAt,
		UpdatedAt:   provider.UpdatedAt,
	}
}

func ToProviderModel(dto *ProviderCreateDTO) *AIProvider {
	now := time.Now()
	return &AIProvider{
		ID:          uuid.New(),
		Name:        dto.Name,
		DisplayName: dto.DisplayName,
		APIKey:      dto.APIKey,
		BaseURL:     dto.BaseURL,
		Enabled:      dto.Enabled,
		Priority:     dto.Priority,
		MaxTokens:    dto.MaxTokens,
		Temperature:  dto.Temperature,
		RateLimit:    dto.RateLimit,
		CostPerToken: 0.0,
		CreatedAt:    now,
		UpdatedAt:    nil,
	}
}

func UpdateProviderModel(provider *AIProvider, dto *ProviderUpdateDTO) {
	now := time.Now()

	if dto.DisplayName != nil {
		provider.DisplayName = *dto.DisplayName
	}
	if dto.APIKey != nil {
		provider.APIKey = *dto.APIKey
	}
	if dto.BaseURL != nil {
		provider.BaseURL = dto.BaseURL
	}
	if dto.Enabled != nil {
		provider.Enabled = *dto.Enabled
	}
	if dto.Priority != nil {
		provider.Priority = *dto.Priority
	}
	if dto.MaxTokens != nil {
		provider.MaxTokens = *dto.MaxTokens
	}
	if dto.Temperature != nil {
		provider.Temperature = *dto.Temperature
	}
	if dto.RateLimit != nil {
		provider.RateLimit = *dto.RateLimit
	}

	provider.UpdatedAt = &now
}

func ToChatResponseDTO(request *AIRequest) *ChatResponseDTO {
	return &ChatResponseDTO{
		ID:               request.ID,
		Provider:         request.ProviderName,
		Model:            request.Model,
		Content:          *request.ResponseText,
		PromptTokens:     request.PromptTokens,
		CompletionTokens: request.CompletionTokens,
		TotalTokens:      request.TotalTokens,
		Cost:             request.Cost,
		DurationMs:       request.DurationMs,
		Cached:           request.Cached,
		CreatedAt:        request.CreatedAt,
	}
}

func ToRequestResponseDTO(request *AIRequest) *RequestResponseDTO {
	return &RequestResponseDTO{
		ID:               request.ID,
		UserID:           request.UserID,
		Provider:         request.ProviderName,
		Model:            request.Model,
		Prompt:           request.Prompt,
		PromptTokens:     request.PromptTokens,
		CompletionTokens: request.CompletionTokens,
		TotalTokens:      request.TotalTokens,
		Cost:             request.Cost,
		DurationMs:       request.DurationMs,
		Status:           request.Status,
		Cached:           request.Cached,
		CreatedAt:        request.CreatedAt,
	}
}

func ToQuotaStatusDTO(quota *AIQuota) *QuotaStatusDTO {
	return &QuotaStatusDTO{
		DailyLimit:             quota.DailyLimit,
		MonthlyLimit:           quota.MonthlyLimit,
		DailyTokenLimit:        quota.DailyTokenLimit,
		MonthlyTokenLimit:      quota.MonthlyTokenLimit,
		DailyUsed:              quota.DailyUsed,
		MonthlyUsed:            quota.MonthlyUsed,
		DailyTokensUsed:        quota.DailyTokensUsed,
		MonthlyTokensUsed:      quota.MonthlyTokensUsed,
		DailyRemaining:         max(0, quota.DailyLimit-quota.DailyUsed),
		MonthlyRemaining:       max(0, quota.MonthlyLimit-quota.MonthlyUsed),
		DailyTokensRemaining:   max(0, quota.DailyTokenLimit-quota.DailyTokensUsed),
		MonthlyTokensRemaining: max(0, quota.MonthlyTokenLimit-quota.MonthlyTokensUsed),
		ResetDaily:             quota.ResetDaily,
		ResetMonthly:           quota.ResetMonthly,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
