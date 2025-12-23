package overpass

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CacheConfig holds cache behavior configuration.
type CacheConfig struct {
	Enabled    bool          // Enable/disable caching (default: false)
	TTL        time.Duration // Time-to-live for cache entries (default: 5 minutes)
	MaxEntries int           // Maximum cache entries (0 = unlimited, default: 1000)
}

// DefaultCacheConfig returns sensible defaults (DISABLED by default).
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:    false,
		TTL:        5 * time.Minute,
		MaxEntries: 1000,
	}
}

// cacheEntry holds cached result with expiration.
type cacheEntry struct {
	result    Result
	expiresAt time.Time
}

// cache implements thread-safe in-memory cache.
type cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	config  CacheConfig
}

// newCache creates new cache instance.
func newCache(config CacheConfig) *cache {
	return &cache{
		entries: make(map[string]*cacheEntry),
		config:  config,
	}
}

// generateKey creates cache key from query and endpoint.
func (c *cache) generateKey(endpoint, query string) string {
	h := sha256.New()
	h.Write([]byte(endpoint))
	h.Write([]byte(query))

	return hex.EncodeToString(h.Sum(nil))
}

// get retrieves cached result if exists and not expired.
func (c *cache) get(endpoint, query string) (Result, bool) {
	if !c.config.Enabled {
		return Result{}, false
	}

	key := c.generateKey(endpoint, query)

	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return Result{}, false
	}

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		// Expired - remove and return miss
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()

		return Result{}, false
	}

	return entry.result, true
}

// set stores result in cache with TTL.
func (c *cache) set(endpoint, query string, result Result) {
	if !c.config.Enabled {
		return
	}

	key := c.generateKey(endpoint, query)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Enforce max entries using simple FIFO eviction
	if c.config.MaxEntries > 0 && len(c.entries) >= c.config.MaxEntries {
		// Find and remove oldest entry
		var oldestKey string
		var oldestTime time.Time

		for k, e := range c.entries {
			if oldestKey == "" || e.expiresAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = e.expiresAt
			}
		}

		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}

	c.entries[key] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.config.TTL),
	}
}

// clear removes all cache entries.
func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// size returns current number of cached entries.
func (c *cache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// cleanup removes expired entries (called periodically).
func (c *cache) cleanup() {
	if !c.config.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// startCleanupRoutine starts background goroutine for periodic cleanup.
func (c *cache) startCleanupRoutine(ctx context.Context) {
	if !c.config.Enabled {
		return
	}

	ticker := time.NewTicker(c.config.TTL / 2) // cleanup at half-TTL intervals

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}
