package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryCache implements an in-memory cache with TTL support
type MemoryCache struct {
	data    map[string]*CacheEntry
	mu      sync.RWMutex
	janitor *time.Ticker
	done    chan struct{}
}

// NewMemoryCache creates a new in-memory cache with periodic cleanup
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:    make(map[string]*CacheEntry),
		janitor: time.NewTicker(cleanupInterval),
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a cached value
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false, nil
	}

	// Check if expired
	if entry.IsExpired() {
		// Remove expired entry (needs write lock, so do it later)
		go func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			delete(c.data, key)
		}()
		return nil, false, nil
	}

	return entry.Value, true, nil
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		HitCount:  0,
		CreatedAt: time.Now(),
	}

	return nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
	return nil
}

// Exists checks if a key exists in the cache
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if entry.IsExpired() {
		return false, nil
	}

	return true, nil
}

// IncrementHit increments the hit count for a cache key
func (c *MemoryCache) IncrementHit(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.data[key]
	if !exists {
		return fmt.Errorf("cache entry not found: %s", key)
	}

	entry.HitCount++
	return nil
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
	for {
		select {
		case <-c.janitor.C:
			c.removeExpired()
		case <-c.done:
			c.janitor.Stop()
			return
		}
	}
}

// removeExpired removes all expired entries
func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
		}
	}
}

// Close stops the cleanup goroutine
func (c *MemoryCache) Close() {
	close(c.done)
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalHits := 0
	for _, entry := range c.data {
		totalHits += entry.HitCount
	}

	return CacheStats{
		Size:      len(c.data),
		TotalHits: totalHits,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size      int
	TotalHits int
}
