package ai

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content" validate:"required"`
}

type ChatRequestDTO struct {
	Provider    string        `json:"provider" validate:"required,oneof=anthropic openai gemini mistral auto"`
	Model       *string       `json:"model,omitempty"`
	Messages    []ChatMessage `json:"messages" validate:"required,min=1,dive"`
	Temperature *float64      `json:"temperature,omitempty" validate:"omitempty,gte=0,lte=2"`
	MaxTokens   *int          `json:"max_tokens,omitempty" validate:"omitempty,gt=0"`
	Stream      bool          `json:"stream"`
	UseCache    bool          `json:"use_cache"`
}

type ChatResponseDTO struct {
	ID               uuid.UUID `json:"id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	Content          string    `json:"content"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	Cost             float64   `json:"cost"`
	DurationMs       int       `json:"duration_ms"`
	Cached           bool      `json:"cached"`
	CreatedAt        time.Time `json:"created_at"`
}

type ProviderCreateDTO struct {
	Name        string  `json:"name" validate:"required,oneof=anthropic openai gemini mistral"`
	DisplayName string  `json:"display_name" validate:"required"`
	APIKey      string  `json:"api_key" validate:"required"`
	BaseURL     *string `json:"base_url,omitempty" validate:"omitempty,url"`
	Enabled     bool    `json:"enabled"`
	Priority    int     `json:"priority" validate:"gte=0"`
	MaxTokens   int     `json:"max_tokens" validate:"required,gt=0"`
	Temperature float64 `json:"temperature" validate:"required,gte=0,lte=2"`
	RateLimit   int     `json:"rate_limit" validate:"gte=0"`
}

type ProviderUpdateDTO struct {
	DisplayName *string  `json:"display_name,omitempty"`
	APIKey      *string  `json:"api_key,omitempty"`
	BaseURL     *string  `json:"base_url,omitempty" validate:"omitempty,url"`
	Enabled     *bool    `json:"enabled,omitempty"`
	Priority    *int     `json:"priority,omitempty" validate:"omitempty,gte=0"`
	MaxTokens   *int     `json:"max_tokens,omitempty" validate:"omitempty,gt=0"`
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,gte=0,lte=2"`
	RateLimit   *int     `json:"rate_limit,omitempty" validate:"omitempty,gte=0"`
}

type ProviderResponseDTO struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	BaseURL     *string    `json:"base_url,omitempty"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
	MaxTokens   int        `json:"max_tokens"`
	Temperature float64    `json:"temperature"`
	RateLimit   int        `json:"rate_limit"`
	HasAPIKey   bool       `json:"has_api_key"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type UsageStatsDTO struct {
	TotalRequests      int                         `json:"total_requests"`
	SuccessfulRequests int                         `json:"successful_requests"`
	FailedRequests     int                         `json:"failed_requests"`
	CachedRequests     int                         `json:"cached_requests"`
	TotalTokens        int                         `json:"total_tokens"`
	PromptTokens       int                         `json:"prompt_tokens"`
	CompletionTokens   int                         `json:"completion_tokens"`
	TotalCost          float64                     `json:"total_cost"`
	AverageDuration    int                         `json:"average_duration_ms"`
	ProviderBreakdown  map[string]ProviderUsageDTO `json:"provider_breakdown"`
}

type ProviderUsageDTO struct {
	Provider    string  `json:"provider"`
	Requests    int     `json:"requests"`
	Tokens      int     `json:"tokens"`
	Cost        float64 `json:"cost"`
	AvgDuration int     `json:"avg_duration_ms"`
}

type QuotaStatusDTO struct {
	DailyLimit             int       `json:"daily_limit"`
	MonthlyLimit           int       `json:"monthly_limit"`
	DailyTokenLimit        int       `json:"daily_token_limit"`
	MonthlyTokenLimit      int       `json:"monthly_token_limit"`
	DailyUsed              int       `json:"daily_used"`
	MonthlyUsed            int       `json:"monthly_used"`
	DailyTokensUsed        int       `json:"daily_tokens_used"`
	MonthlyTokensUsed      int       `json:"monthly_tokens_used"`
	DailyRemaining         int       `json:"daily_remaining"`
	MonthlyRemaining       int       `json:"monthly_remaining"`
	DailyTokensRemaining   int       `json:"daily_tokens_remaining"`
	MonthlyTokensRemaining int       `json:"monthly_tokens_remaining"`
	ResetDaily             time.Time `json:"reset_daily"`
	ResetMonthly           time.Time `json:"reset_monthly"`
}

type RequestFilterDTO struct {
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	Provider *string    `json:"provider,omitempty"`
	Status   *string    `json:"status,omitempty" validate:"omitempty,oneof=success error cached"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Limit    int        `json:"limit" validate:"omitempty,gt=0,lte=100"`
	Offset   int        `json:"offset" validate:"omitempty,gte=0"`
}

type RequestResponseDTO struct {
	ID               uuid.UUID  `json:"id"`
	UserID           *uuid.UUID `json:"user_id,omitempty"`
	Provider         string     `json:"provider"`
	Model            string     `json:"model"`
	Prompt           string     `json:"prompt"`
	PromptTokens     int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	TotalTokens      int        `json:"total_tokens"`
	Cost             float64    `json:"cost"`
	DurationMs       int        `json:"duration_ms"`
	Status           string     `json:"status"`
	Cached           bool       `json:"cached"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ErrorResponseDTO struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginatedResponseDTO struct {
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}
