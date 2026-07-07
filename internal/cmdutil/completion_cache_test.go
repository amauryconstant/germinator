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
