package ratelimit

import (
	"context"
	"time"
)

// Limiter defines the interface for rate limiting
type Limiter interface {
	// Allow checks if a request is allowed for the given key
	Allow(ctx context.Context, key string) (bool, error)

	// AllowN checks if N requests are allowed for the given key
	AllowN(ctx context.Context, key string, n int) (bool, error)

	// Wait blocks until the request is allowed
	Wait(ctx context.Context, key string) error

	// WaitN blocks until N requests are allowed
	WaitN(ctx context.Context, key string, n int) error

	// Reset resets the rate limit for the given key
	Reset(ctx context.Context, key string) error

	// GetLimit returns the current limit for the given key
	GetLimit(ctx context.Context, key string) (int, error)
}

// Config represents rate limiter configuration
type Config struct {
	// Rate is the number of requests allowed
	Rate int

	// Per is the time window for the rate limit
	Per time.Duration

	// Burst is the maximum burst size (for token bucket)
	Burst int
}
