package agents

import (
	"sync"
	"time"
)

// HealthCache provides TTL-based caching for agent health checks
// to reduce redundant API calls during frequent availability checks.
type HealthCache struct {
	mu        sync.RWMutex
	available bool
	checkedAt time.Time
	ttl       time.Duration
}

// NewHealthCache creates a new HealthCache with the specified TTL.
// A TTL of 0 effectively disables caching.
func NewHealthCache(ttl time.Duration) *HealthCache {
	return &HealthCache{
		ttl: ttl,
	}
}

// IsValid returns true if the cached result is still within TTL.
func (c *HealthCache) IsValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.checkedAt.IsZero() && time.Since(c.checkedAt) < c.ttl
}

// Get returns the cached availability status and whether the cache is valid.
func (c *HealthCache) Get() (available bool, valid bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	valid = !c.checkedAt.IsZero() && time.Since(c.checkedAt) < c.ttl
	return c.available, valid
}

// Set updates the cached availability status.
func (c *HealthCache) Set(available bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.available = available
	c.checkedAt = time.Now()
}

// Invalidate clears the cache, forcing the next check to make a live call.
func (c *HealthCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checkedAt = time.Time{}
}

// TTL returns the cache's time-to-live duration.
func (c *HealthCache) TTL() time.Duration {
	return c.ttl
}

// DefaultHealthCacheTTL is the default TTL for health check caching (30 seconds).
const DefaultHealthCacheTTL = 30 * time.Second
