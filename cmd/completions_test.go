package cmd

import (
	"testing"
	"time"

	"gitlab.com/amoconst/germinator/internal/infrastructure/config"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

func TestGetCompletionTimeout(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected time.Duration
	}{
		{
			name:     "nil config returns default",
			cfg:      nil,
			expected: 500 * time.Millisecond,
		},
		{
			name: "empty timeout returns default",
			cfg: &config.Config{
				Completion: config.CompletionConfig{Timeout: ""},
			},
			expected: 500 * time.Millisecond,
		},
		{
			name: "valid timeout",
			cfg: &config.Config{
				Completion: config.CompletionConfig{Timeout: "1s"},
			},
			expected: 1 * time.Second,
		},
		{
			name: "invalid timeout returns default",
			cfg: &config.Config{
				Completion: config.CompletionConfig{Timeout: "invalid"},
			},
			expected: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCompletionTimeout(tt.cfg)
			if got != tt.expected {
				t.Errorf("getCompletionTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetCacheTTL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected time.Duration
	}{
		{
			name:     "nil config returns default",
			cfg:      nil,
			expected: 5 * time.Second,
		},
		{
			name: "empty cache_ttl returns default",
			cfg: &config.Config{
				Completion: config.CompletionConfig{CacheTTL: ""},
			},
			expected: 5 * time.Second,
		},
		{
			name: "valid cache_ttl",
			cfg: &config.Config{
				Completion: config.CompletionConfig{CacheTTL: "10s"},
			},
			expected: 10 * time.Second,
		},
		{
			name: "invalid cache_ttl returns default",
			cfg: &config.Config{
				Completion: config.CompletionConfig{CacheTTL: "invalid"},
			},
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCacheTTL(tt.cfg)
			if got != tt.expected {
				t.Errorf("getCacheTTL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActionPlatforms(_ *testing.T) {
	// Test that actionPlatforms returns a valid action
	_ = actionPlatforms()
}

func TestResourceActionFromLibrary(t *testing.T) {
	tests := []struct {
		name string
		lib  *library.Library
	}{
		{
			name: "empty library",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{},
			},
		},
		{
			name: "single resource",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{
					"skill": {
						"commit": {Path: "skills/commit.md", Description: "Git commit helper"},
					},
				},
			},
		},
		{
			name: "multiple resources",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{
					"skill": {
						"commit": {Path: "skills/commit.md", Description: "Git commit helper"},
						"review": {Path: "skills/review.md", Description: "Code review helper"},
					},
					"agent": {
						"planner": {Path: "agents/planner.md", Description: "Planning agent"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			_ = resourceActionFromLibrary(tt.lib)
		})
	}
}

func TestPresetActionFromLibrary(t *testing.T) {
	tests := []struct {
		name string
		lib  *library.Library
	}{
		{
			name: "empty presets",
			lib: &library.Library{
				Presets: map[string]library.Preset{},
			},
		},
		{
			name: "single preset",
			lib: &library.Library{
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
				},
			},
		},
		{
			name: "multiple presets",
			lib: &library.Library{
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
					"code-review":  {Name: "code-review", Description: "Code review workflow"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			_ = presetActionFromLibrary(tt.lib)
		})
	}
}

func TestLibraryRefActionFromLibrary(t *testing.T) {
	tests := []struct {
		name string
		lib  *library.Library
	}{
		{
			name: "empty library",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{},
				Presets:   map[string]library.Preset{},
			},
		},
		{
			name: "resources only",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{
					"skill": {
						"commit": {Path: "skills/commit.md", Description: "Git commit helper"},
					},
				},
				Presets: map[string]library.Preset{},
			},
		},
		{
			name: "presets only",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{},
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
				},
			},
		},
		{
			name: "both resources and presets",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{
					"skill": {
						"commit": {Path: "skills/commit.md", Description: "Git commit helper"},
					},
				},
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			_ = libraryRefActionFromLibrary(tt.lib)
		})
	}
}

func TestGetCachedLibrary(t *testing.T) {
	// Initialize the cache
	completionCache.entries = make(map[string]*cachedLibraryData)

	lib := &library.Library{
		Version: "1",
	}

	t.Run("cache miss returns nil", func(t *testing.T) {
		got := getCachedLibrary("/nonexistent", 5*time.Second)
		if got != nil {
			t.Errorf("getCachedLibrary() = %v, want nil", got)
		}
	})

	t.Run("cache hit returns library", func(t *testing.T) {
		libPath := "/test/library"
		setCachedLibrary(libPath, lib, 5*time.Second)

		got := getCachedLibrary(libPath, 5*time.Second)
		if got == nil {
			t.Error("getCachedLibrary() returned nil, expected library")
		}
	})

	t.Run("expired cache returns nil", func(t *testing.T) {
		libPath := "/test/expired"
		// Set with 0 TTL so it expires immediately
		setCachedLibrary(libPath, lib, 0)

		// Small sleep to ensure expiration
		time.Sleep(1 * time.Millisecond)

		got := getCachedLibrary(libPath, 5*time.Second)
		if got != nil {
			t.Errorf("getCachedLibrary() = %v, want nil (expired)", got)
		}
	})
}

func TestSetCachedLibrary(t *testing.T) {
	// Initialize the cache
	completionCache.entries = make(map[string]*cachedLibraryData)

	lib := &library.Library{
		Version: "1",
	}

	libPath := "/test/cache"
	ttl := 5 * time.Second

	setCachedLibrary(libPath, lib, ttl)

	// Verify the library was cached
	completionCache.mu.RLock()
	entry, exists := completionCache.entries[libPath]
	completionCache.mu.RUnlock()

	if !exists {
		t.Error("setCachedLibrary() did not cache the library")
	}

	if entry.data != lib {
		t.Error("setCachedLibrary() cached wrong library")
	}

	if entry.expiresAt.Before(time.Now()) {
		t.Error("setCachedLibrary() expiration time is in the past")
	}
}
