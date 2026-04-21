package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAIProviderTableName(t *testing.T) {
	provider := AIProvider{}
	assert.Equal(t, "ai_providers", provider.TableName())
}

func TestAIRequestTableName(t *testing.T) {
	request := AIRequest{}
	assert.Equal(t, "ai_requests", request.TableName())
}

func TestAICacheTableName(t *testing.T) {
	cache := AICache{}
	assert.Equal(t, "ai_cache", cache.TableName())
}

func TestAIQuotaTableName(t *testing.T) {
	quota := AIQuota{}
	assert.Equal(t, "ai_quotas", quota.TableName())
}

func TestAIQuotaIsExceeded(t *testing.T) {
	tests := []struct {
		name     string
		quota    AIQuota
		expected bool
	}{
		{
			name: "Daily requests exceeded",
			quota: AIQuota{
				DailyLimit:        100,
				DailyUsed:         100,
				MonthlyLimit:      1000,
				MonthlyUsed:       50,
				DailyTokenLimit:   10000,
				DailyTokensUsed:   5000,
				MonthlyTokenLimit: 100000,
				MonthlyTokensUsed: 50000,
			},
			expected: true,
		},
		{
			name: "Monthly requests exceeded",
			quota: AIQuota{
				DailyLimit:        100,
				DailyUsed:         50,
				MonthlyLimit:      1000,
				MonthlyUsed:       1000,
				DailyTokenLimit:   10000,
				DailyTokensUsed:   5000,
				MonthlyTokenLimit: 100000,
				MonthlyTokensUsed: 50000,
			},
			expected: true,
		},
		{
			name: "Daily tokens exceeded",
			quota: AIQuota{
				DailyLimit:        100,
				DailyUsed:         50,
				MonthlyLimit:      1000,
				MonthlyUsed:       500,
				DailyTokenLimit:   10000,
				DailyTokensUsed:   10000,
				MonthlyTokenLimit: 100000,
				MonthlyTokensUsed: 50000,
			},
			expected: true,
		},
		{
			name: "Monthly tokens exceeded",
			quota: AIQuota{
				DailyLimit:        100,
				DailyUsed:         50,
				MonthlyLimit:      1000,
				MonthlyUsed:       500,
				DailyTokenLimit:   10000,
				DailyTokensUsed:   5000,
				MonthlyTokenLimit: 100000,
				MonthlyTokensUsed: 100000,
			},
			expected: true,
		},
		{
			name: "Not exceeded",
			quota: AIQuota{
				DailyLimit:        100,
				DailyUsed:         50,
				MonthlyLimit:      1000,
				MonthlyUsed:       500,
				DailyTokenLimit:   10000,
				DailyTokensUsed:   5000,
				MonthlyTokenLimit: 100000,
				MonthlyTokensUsed: 50000,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.quota.IsExceeded())
		})
	}
}

func TestAIQuotaNeedsReset(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		quota         AIQuota
		expectedDaily bool
		expectedMonthly bool
	}{
		{
			name: "Both need reset",
			quota: AIQuota{
				ResetDaily:   now.Add(-1 * time.Hour),
				ResetMonthly: now.Add(-1 * time.Hour),
			},
			expectedDaily:   true,
			expectedMonthly: true,
		},
		{
			name: "Only daily needs reset",
			quota: AIQuota{
				ResetDaily:   now.Add(-1 * time.Hour),
				ResetMonthly: now.Add(1 * time.Hour),
			},
			expectedDaily:   true,
			expectedMonthly: false,
		},
		{
			name: "Only monthly needs reset",
			quota: AIQuota{
				ResetDaily:   now.Add(1 * time.Hour),
				ResetMonthly: now.Add(-1 * time.Hour),
			},
			expectedDaily:   false,
			expectedMonthly: true,
		},
		{
			name: "Neither needs reset",
			quota: AIQuota{
				ResetDaily:   now.Add(1 * time.Hour),
				ResetMonthly: now.Add(1 * time.Hour),
			},
			expectedDaily:   false,
			expectedMonthly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daily, monthly := tt.quota.NeedsReset()
			assert.Equal(t, tt.expectedDaily, daily)
			assert.Equal(t, tt.expectedMonthly, monthly)
		})
	}
}

func TestAIQuotaResetDailyCounters(t *testing.T) {
	quota := AIQuota{
		DailyUsed:       50,
		DailyTokensUsed: 5000,
	}

	quota.ResetDailyCounters()

	assert.Equal(t, 0, quota.DailyUsed)
	assert.Equal(t, 0, quota.DailyTokensUsed)
	assert.True(t, quota.ResetDaily.After(time.Now()))
}

func TestAIQuotaResetMonthlyCounters(t *testing.T) {
	quota := AIQuota{
		MonthlyUsed:       500,
		MonthlyTokensUsed: 50000,
	}

	quota.ResetMonthlyCounters()

	assert.Equal(t, 0, quota.MonthlyUsed)
	assert.Equal(t, 0, quota.MonthlyTokensUsed)
	assert.True(t, quota.ResetMonthly.After(time.Now()))
}
