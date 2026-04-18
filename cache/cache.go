package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)
	IncrementHit(ctx context.Context, key string) error
}

type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	HitCount  int
	CreatedAt time.Time
}

func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}
