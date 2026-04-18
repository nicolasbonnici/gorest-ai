package ai

import (
	"time"

	"github.com/google/uuid"
)

type AIProvider struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Name         string     `db:"name" json:"name"`
	DisplayName  string     `db:"display_name" json:"display_name"`
	APIKey       string     `db:"api_key" json:"-"`
	BaseURL      *string    `db:"base_url" json:"base_url,omitempty"`
	Enabled      bool       `db:"enabled" json:"enabled"`
	Priority     int        `db:"priority" json:"priority"`
	MaxTokens    int        `db:"max_tokens" json:"max_tokens"`
	Temperature  float64    `db:"temperature" json:"temperature"`
	RateLimit    int        `db:"rate_limit" json:"rate_limit"`
	CostPerToken float64    `db:"cost_per_token" json:"cost_per_token"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (AIProvider) TableName() string {
	return "ai_providers"
}

type AIRequest struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	UserID           *uuid.UUID `db:"user_id" json:"user_id,omitempty"`
	ProviderID       uuid.UUID  `db:"provider_id" json:"provider_id"`
	ProviderName     string     `db:"provider_name" json:"provider_name"`
	Model            string     `db:"model" json:"model"`
	RequestType      string     `db:"request_type" json:"request_type"`
	Prompt           string     `db:"prompt" json:"prompt"`
	ResponseText     *string    `db:"response_text" json:"response_text,omitempty"`
	PromptTokens     int        `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int        `db:"completion_tokens" json:"completion_tokens"`
	TotalTokens      int        `db:"total_tokens" json:"total_tokens"`
	Cost             float64    `db:"cost" json:"cost"`
	DurationMs       int        `db:"duration_ms" json:"duration_ms"`
	Status           string     `db:"status" json:"status"`
	ErrorMessage     *string    `db:"error_message" json:"error_message,omitempty"`
	Cached           bool       `db:"cached" json:"cached"`
	IPAddress        *string    `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}

func (AIRequest) TableName() string {
	return "ai_requests"
}

type AICache struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	CacheKey         string     `db:"cache_key" json:"cache_key"`
	ProviderName     string     `db:"provider_name" json:"provider_name"`
	Model            string     `db:"model" json:"model"`
	RequestType      string     `db:"request_type" json:"request_type"`
	ResponseText     string     `db:"response_text" json:"response_text"`
	PromptTokens     int        `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int        `db:"completion_tokens" json:"completion_tokens"`
	TotalTokens      int        `db:"total_tokens" json:"total_tokens"`
	HitCount         int        `db:"hit_count" json:"hit_count"`
	ExpiresAt        time.Time  `db:"expires_at" json:"expires_at"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (AICache) TableName() string {
	return "ai_cache"
}

type AIQuota struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	UserID            uuid.UUID  `db:"user_id" json:"user_id"`
	DailyLimit        int        `db:"daily_limit" json:"daily_limit"`
	MonthlyLimit      int        `db:"monthly_limit" json:"monthly_limit"`
	DailyTokenLimit   int        `db:"daily_token_limit" json:"daily_token_limit"`
	MonthlyTokenLimit int        `db:"monthly_token_limit" json:"monthly_token_limit"`
	DailyUsed         int        `db:"daily_used" json:"daily_used"`
	MonthlyUsed       int        `db:"monthly_used" json:"monthly_used"`
	DailyTokensUsed   int        `db:"daily_tokens_used" json:"daily_tokens_used"`
	MonthlyTokensUsed int        `db:"monthly_tokens_used" json:"monthly_tokens_used"`
	ResetDaily        time.Time  `db:"reset_daily" json:"reset_daily"`
	ResetMonthly      time.Time  `db:"reset_monthly" json:"reset_monthly"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (AIQuota) TableName() string {
	return "ai_quotas"
}

func (q *AIQuota) IsExceeded() bool {
	return q.DailyUsed >= q.DailyLimit ||
		q.MonthlyUsed >= q.MonthlyLimit ||
		q.DailyTokensUsed >= q.DailyTokenLimit ||
		q.MonthlyTokensUsed >= q.MonthlyTokenLimit
}

func (q *AIQuota) NeedsReset() (daily, monthly bool) {
	now := time.Now()
	daily = now.After(q.ResetDaily)
	monthly = now.After(q.ResetMonthly)
	return
}

func (q *AIQuota) ResetDailyCounters() {
	q.DailyUsed = 0
	q.DailyTokensUsed = 0
	q.ResetDaily = time.Now().Add(24 * time.Hour)
}

func (q *AIQuota) ResetMonthlyCounters() {
	q.MonthlyUsed = 0
	q.MonthlyTokensUsed = 0
	now := time.Now()
	q.ResetMonthly = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
}
