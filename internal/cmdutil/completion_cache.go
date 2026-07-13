package cmdutil

import (
	"sync"
	"time"

	"gitlab.com/amoconst/germinator/internal/library"
)

// cacheEntry holds a cached library snapshot with its expiration time.
type cacheEntry struct {
	lib       *library.Library
	expiresAt time.Time
}

// CompletionCache memoizes library data for shell completion lookups
// with an optional per-entry TTL. All methods are safe for concurrent
// use. Each Factory instance owns its own CompletionCache (populated
// in main.go) so tests construct a fresh cache via NewFactory plus
// main.go wiring.
//
// Time is read from the now function field (default time.Now) so tests
// can drive expiry deterministically via WithClock. WithClock returns a
// new pointer so callers must update their reference; the original
// cache is unchanged (immutable-config pattern, no concurrent-mutation
// hazard with the shared cache).
type CompletionCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	now     func() time.Time
}

// NewCompletionCache returns a fresh cache ready for use with time.Now
// as the default clock.
func NewCompletionCache() *CompletionCache {
	return &CompletionCache{
		entries: make(map[string]cacheEntry),
		now:     time.Now,
	}
}

// WithClock returns a copy of the cache that reads time from fn instead
// of time.Now. Useful for deterministic TTL tests. The original cache
// is unchanged.
func (c *CompletionCache) WithClock(fn func() time.Time) *CompletionCache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clone := &CompletionCache{
		entries: make(map[string]cacheEntry, len(c.entries)),
		now:     fn,
	}
	for k, v := range c.entries {
		clone.entries[k] = v
	}
	return clone
}

// Get returns the cached library for key, or nil if missing or expired.
func (c *CompletionCache) Get(key string) *library.Library {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || c.now().After(e.expiresAt) {
		return nil
	}
	return e.lib
}

// Set stores lib under key for ttl duration.
func (c *CompletionCache) Set(key string, lib *library.Library, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{lib: lib, expiresAt: c.now().Add(ttl)}
}

// Reset removes all cached entries so the cache is reusable as if
// newly constructed.
func (c *CompletionCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]cacheEntry)
}

// Invalidate clears all entries; called by mutating library commands
// so the next completion reflects the new state immediately. It is a
// semantic alias for Reset, kept distinct so future per-key
// invalidation strategies can diverge.
func (c *CompletionCache) Invalidate() {
	c.Reset()
}
