package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/library"
)

func TestGetCompletionTimeout(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got := getCompletionTimeout(tt.cfg)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetCacheTTL(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			got := getCacheTTL(tt.cfg)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestActionPlatforms(t *testing.T) {
	t.Parallel()

	f := &cmdutil.Factory{CompletionCache: cmdutil.NewCompletionCache()}
	// Test that actionPlatforms returns a valid action given a Factory.
	_ = actionPlatforms(f)
}

func TestResourceActionFromLibrary(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = resourceActionFromLibrary(tt.lib)
		})
	}
}

func TestPresetActionFromLibrary(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = presetActionFromLibrary(tt.lib)
		})
	}
}

func TestLibraryRefActionFromLibrary(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = libraryRefActionFromLibrary(tt.lib)
		})
	}
}
