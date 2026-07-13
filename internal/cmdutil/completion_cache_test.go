package cmdutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/amoconst/germinator/internal/library"
)

func TestCompletionCache_GetSet(t *testing.T) {
	t.Parallel()

	cache := NewCompletionCache()
	lib := &library.Library{Version: "1"}

	t.Run("cache miss returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, cache.Get("/nonexistent"))
	})

	t.Run("cache hit returns library", func(t *testing.T) {
		t.Parallel()
		libPath := "/test/hit"
		cache.Set(libPath, lib, 5*time.Second)
		got := cache.Get(libPath)
		assert.Equal(t, lib, got)
	})

	t.Run("expired cache returns nil", func(t *testing.T) {
		t.Parallel()
		libPath := "/test/expired"
		// Set with a TTL that has already elapsed (negative offset via
		// 0 duration forces immediate expiration on the next Get
		// because time.Now().After(expiresAt) is true when expiresAt
		// equals Now).
		cache.Set(libPath, lib, 0)
		time.Sleep(1 * time.Millisecond)
		assert.Nil(t, cache.Get(libPath))
	})
}

func TestCompletionCache_Reset(t *testing.T) {
	t.Parallel()

	cache := NewCompletionCache()
	lib := &library.Library{Version: "1"}

	cache.Set("/a", lib, 5*time.Second)
	cache.Set("/b", lib, 5*time.Second)
	assert.NotNil(t, cache.Get("/a"))
	assert.NotNil(t, cache.Get("/b"))

	cache.Reset()

	assert.Nil(t, cache.Get("/a"))
	assert.Nil(t, cache.Get("/b"))
}

func TestCompletionCache_ResetIsReusable(t *testing.T) {
	t.Parallel()

	cache := NewCompletionCache()
	lib := &library.Library{Version: "1"}

	cache.Set("/a", lib, 5*time.Second)
	cache.Reset()
	// After Reset the cache MUST still accept new entries.
	cache.Set("/a", lib, 5*time.Second)
	assert.Equal(t, lib, cache.Get("/a"))
}

func TestCompletionCache_Invalidate(t *testing.T) {
	t.Parallel()

	cache := NewCompletionCache()
	lib := &library.Library{Version: "1"}

	cache.Set("/a", lib, 5*time.Second)
	assert.NotNil(t, cache.Get("/a"))

	// Invalidate is a semantic alias for Reset: it clears all entries.
	cache.Invalidate()
	assert.Nil(t, cache.Get("/a"))
}

func TestCompletionCache_PerFactoryIsolation(t *testing.T) {
	t.Parallel()

	c1 := NewCompletionCache()
	c2 := NewCompletionCache()
	lib := &library.Library{Version: "1"}

	c1.Set("/key", lib, 5*time.Second)
	// c2 is an independent cache: the entry set on c1 MUST NOT be
	// visible to c2.
	assert.Nil(t, c2.Get("/key"))
	assert.Equal(t, lib, c1.Get("/key"))
}

// TestCompletionCache_EntryExpiresAfterTTL drives a fake clock
// forward to prove the cache evicts entries exactly at the configured
// TTL boundary (not earlier, not later). Pins the spec scenarios
// "Default cache TTL" (5s) and "Configurable cache TTL" (configured
// duration).
func TestCompletionCache_EntryExpiresAfterTTL(t *testing.T) {
	t.Parallel()

	t.Run("5s default TTL expires at boundary", func(t *testing.T) {
		t.Parallel()
		fake := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		cache := NewCompletionCache().WithClock(func() time.Time { return fake })
		lib := &library.Library{Version: "1"}

		cache.Set("/lib", lib, 5*time.Second)
		// Just before the boundary: hit.
		fake = fake.Add(4*time.Second + 999*time.Millisecond)
		assert.NotNil(t, cache.Get("/lib"),
			"entry must remain valid 1ms before TTL boundary")

		// Past the boundary: miss.
		fake = fake.Add(2 * time.Millisecond)
		assert.Nil(t, cache.Get("/lib"),
			"entry must be expired 1ms past TTL boundary")
	})

	t.Run("configured 10s TTL expires at boundary", func(t *testing.T) {
		t.Parallel()
		fake := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		cache := NewCompletionCache().WithClock(func() time.Time { return fake })
		lib := &library.Library{Version: "2"}

		cache.Set("/lib", lib, 10*time.Second)
		fake = fake.Add(9*time.Second + 999*time.Millisecond)
		assert.NotNil(t, cache.Get("/lib"))

		fake = fake.Add(2 * time.Millisecond)
		assert.Nil(t, cache.Get("/lib"))
	})
}

// TestCompletionCache_DifferentTTLsExpireIndependently verifies that
// entries with different TTLs expire independently under a shared clock.
func TestCompletionCache_DifferentTTLsExpireIndependently(t *testing.T) {
	t.Parallel()

	fake := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cache := NewCompletionCache().WithClock(func() time.Time { return fake })

	cache.Set("/short", &library.Library{Version: "short"}, 2*time.Second)
	cache.Set("/long", &library.Library{Version: "long"}, 10*time.Second)

	// At t+1s both are alive.
	fake = fake.Add(1 * time.Second)
	assert.NotNil(t, cache.Get("/short"))
	assert.NotNil(t, cache.Get("/long"))

	// At t+3s, /short has expired but /long remains.
	fake = fake.Add(2 * time.Second)
	assert.Nil(t, cache.Get("/short"), "short TTL entry should have expired")
	assert.NotNil(t, cache.Get("/long"), "long TTL entry should still be valid")

	// At t+11s, both have expired.
	fake = fake.Add(8 * time.Second)
	assert.Nil(t, cache.Get("/short"))
	assert.Nil(t, cache.Get("/long"))
}

// TestCompletionCache_WithClockReturnsIndependentCopy verifies that
// WithClock produces a clone that does not mutate the original cache
// when entries are Set through the clone.
func TestCompletionCache_WithClockReturnsIndependentCopy(t *testing.T) {
	t.Parallel()

	original := NewCompletionCache()
	original.Set("/orig", &library.Library{Version: "orig"}, 5*time.Second)

	fake := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clone := original.WithClock(func() time.Time { return fake })

	clone.Set("/clone", &library.Library{Version: "clone"}, 5*time.Second)

	assert.Nil(t, original.Get("/clone"),
		"original cache MUST NOT see entries Set via the clone")
	assert.NotNil(t, clone.Get("/orig"),
		"clone MUST inherit entries Set on the original")
}
