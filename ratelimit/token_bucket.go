package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm
type TokenBucket struct {
	buckets map[string]*bucket
	config  Config
	mu      sync.RWMutex
}

// bucket represents a token bucket for a single key
type bucket struct {
	tokens    int
	lastRefill time.Time
	mu        sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(config Config) *TokenBucket {
	return &TokenBucket{
		buckets: make(map[string]*bucket),
		config:  config,
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	return tb.AllowN(ctx, key, 1)
}

// AllowN checks if N requests are allowed
func (tb *TokenBucket) AllowN(ctx context.Context, key string, n int) (bool, error) {
	b := tb.getBucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	tb.refillBucket(b)

	// Check if we have enough tokens
	if b.tokens >= n {
		b.tokens -= n
		return true, nil
	}

	return false, nil
}

// Wait blocks until the request is allowed
func (tb *TokenBucket) Wait(ctx context.Context, key string) error {
	return tb.WaitN(ctx, key, 1)
}

// WaitN blocks until N requests are allowed
func (tb *TokenBucket) WaitN(ctx context.Context, key string, n int) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			allowed, err := tb.AllowN(ctx, key, n)
			if err != nil {
				return err
			}
			if allowed {
				return nil
			}
		}
	}
}

// Reset resets the rate limit for the given key
func (tb *TokenBucket) Reset(ctx context.Context, key string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, exists := tb.buckets[key]
	if !exists {
		return fmt.Errorf("bucket not found for key: %s", key)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.tokens = tb.config.Burst
	b.lastRefill = time.Now()

	return nil
}

// GetLimit returns the current limit for the given key
func (tb *TokenBucket) GetLimit(ctx context.Context, key string) (int, error) {
	b := tb.getBucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	tb.refillBucket(b)
	return b.tokens, nil
}

// getBucket gets or creates a bucket for the given key
func (tb *TokenBucket) getBucket(key string) *bucket {
	tb.mu.RLock()
	b, exists := tb.buckets[key]
	tb.mu.RUnlock()

	if exists {
		return b
	}

	// Create new bucket
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Double-check after acquiring write lock
	b, exists = tb.buckets[key]
	if exists {
		return b
	}

	b = &bucket{
		tokens:    tb.config.Burst,
		lastRefill: time.Now(),
	}
	tb.buckets[key] = b

	return b
}

// refillBucket refills tokens based on time elapsed
func (tb *TokenBucket) refillBucket(b *bucket) {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	if elapsed < tb.config.Per {
		// Not enough time has passed for refill
		return
	}

	// Calculate tokens to add based on elapsed time
	tokensToAdd := int(elapsed / tb.config.Per) * tb.config.Rate

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > tb.config.Burst {
			b.tokens = tb.config.Burst
		}
		b.lastRefill = now
	}
}

// Cleanup removes old buckets (for periodic cleanup)
func (tb *TokenBucket) Cleanup(maxAge time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	for key, b := range tb.buckets {
		b.mu.Lock()
		age := now.Sub(b.lastRefill)
		b.mu.Unlock()

		if age > maxAge {
			delete(tb.buckets, key)
		}
	}
}
