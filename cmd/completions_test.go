package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	values := actionValuesAsStrings(t, actionPlatforms(f).Invoke(carapace.Context{}))
	assert.Contains(t, values, "claude-code",
		"actionPlatforms MUST expose the claude-code value")
	assert.Contains(t, values, "opencode",
		"actionPlatforms MUST expose the opencode value")
}

// actionValuesAsStrings invokes the given action and extracts the
// "value" field of each entry from the JSON-marshalled InvokedAction.
// carapace.Action keeps RawValues private; the public JSON surface
// is the documented inspection seam.
func actionValuesAsStrings(t *testing.T, ia carapace.InvokedAction) []string {
	t.Helper()
	data, err := json.Marshal(ia)
	require.NoError(t, err, "json.Marshal of InvokedAction must succeed")
	var decoded struct {
		Values []struct {
			Value string `json:"value"`
		} `json:"values"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded),
		"json.Unmarshal of InvokedAction must succeed")
	out := make([]string, 0, len(decoded.Values))
	for _, v := range decoded.Values {
		out = append(out, v.Value)
	}
	return out
}

func TestResourceActionFromLibrary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		lib            *library.Library
		wantContains   []string
		wantNotPresent []string
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
			wantContains: []string{"skill/commit"},
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
			wantContains: []string{"skill/commit", "skill/review", "agent/planner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			values := actionValuesAsStrings(t, resourceActionFromLibrary(tt.lib).Invoke(carapace.Context{}))
			for _, want := range tt.wantContains {
				assert.Contains(t, values, want,
					"resource action MUST include %q", want)
			}
			for _, dontWant := range tt.wantNotPresent {
				assert.NotContains(t, values, dontWant)
			}
		})
	}
}

func TestPresetActionFromLibrary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		lib          *library.Library
		wantContains []string
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
			wantContains: []string{"git-workflow"},
		},
		{
			name: "multiple presets",
			lib: &library.Library{
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
					"code-review":  {Name: "code-review", Description: "Code review workflow"},
				},
			},
			wantContains: []string{"git-workflow", "code-review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			values := actionValuesAsStrings(t, presetActionFromLibrary(tt.lib).Invoke(carapace.Context{}))
			for _, want := range tt.wantContains {
				assert.Contains(t, values, want,
					"preset action MUST include %q", want)
			}
		})
	}
}

func TestLibraryRefActionFromLibrary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		lib          *library.Library
		wantContains []string
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
			wantContains: []string{"skill/commit"},
		},
		{
			name: "presets only",
			lib: &library.Library{
				Resources: map[string]map[string]library.Resource{},
				Presets: map[string]library.Preset{
					"git-workflow": {Name: "git-workflow", Description: "Git workflow tools"},
				},
			},
			wantContains: []string{"preset/git-workflow"},
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
			wantContains: []string{"skill/commit", "preset/git-workflow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			values := actionValuesAsStrings(t, libraryRefActionFromLibrary(tt.lib).Invoke(carapace.Context{}))
			for _, want := range tt.wantContains {
				assert.Contains(t, values, want,
					"library-ref action MUST include %q", want)
			}
		})
	}
}

// =============================================================================
// loadLibraryForCompletion — slice 9.1 implementation tests
// (covers shell-completion spec "actionResources loads library with timeout
// and bypasses f.Library" + "Timeout returns empty completion")
// =============================================================================

// newFactoryWithCache returns a Factory wired with a fresh CompletionCache.
// RootContext is the caller-supplied ctx (used by timeout tests).
func newFactoryWithCache(ctx context.Context) *cmdutil.Factory {
	return &cmdutil.Factory{
		IOStreams:       nil,
		RootContext:     ctx,
		CompletionCache: cmdutil.NewCompletionCache(),
	}
}

// TestLoadLibraryForCompletion_CacheHit exercises the cache-first branch:
// when f.CompletionCache already holds the libPath entry, loadLibraryForCompletion
// must return it WITHOUT calling library.LoadLibrary.
func TestLoadLibraryForCompletion_CacheHit(t *testing.T) {
	t.Parallel()

	libDir := makeCompletionTestLibrary(t)
	cached := &library.Library{Version: "1", RootPath: libDir}
	cache := cmdutil.NewCompletionCache()
	cache.Set(libDir, cached, 5*time.Second)

	f := newFactoryWithCache(context.Background())
	f.CompletionCache = cache

	got := loadLibraryForCompletion(f, libDir)
	require.NotNil(t, got, "cached entry must be returned")
	assert.Same(t, cached, got,
		"loadLibraryForCompletion must return the exact cached library pointer")
}

// TestLoadLibraryForCompletion_CacheMissLoads covers the cache-miss + load path:
// the function must load the library from disk and store it in the cache.
func TestLoadLibraryForCompletion_CacheMissLoads(t *testing.T) {
	t.Parallel()

	libDir := makeCompletionTestLibrary(t)
	f := newFactoryWithCache(context.Background())

	require.Nil(t, f.CompletionCache.Get(libDir),
		"precondition: cache must NOT hold the entry before the call")

	got := loadLibraryForCompletion(f, libDir)
	require.NotNil(t, got, "library must be loaded on cache miss")

	cached := f.CompletionCache.Get(libDir)
	require.NotNil(t, cached,
		"successful load must populate the cache for subsequent calls")
	assert.Equal(t, got.RootPath, cached.RootPath,
		"cached library must be the same one returned to the caller")
}

// TestLoadLibraryForCompletion_NonexistentPath covers silent failure: an
// invalid library path returns nil so the action surfaces an empty
// completion rather than an error.
func TestLoadLibraryForCompletion_NonexistentPath(t *testing.T) {
	t.Parallel()

	f := newFactoryWithCache(context.Background())
	bogus := "/no/such/path/" + t.Name()

	assert.Nil(t, loadLibraryForCompletion(f, bogus),
		"nonexistent path must return nil (silent failure)")
}

// TestLoadLibraryForCompletion_Timeout covers the timeout branch: when the
// wrapped context is already cancelled, library.LoadLibrary returns
// context.Canceled and loadLibraryForCompletion surfaces nil.
func TestLoadLibraryForCompletion_Timeout(t *testing.T) {
	t.Parallel()

	libDir := makeCompletionTestLibrary(t)
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	f := newFactoryWithCache(cancelledCtx)

	assert.Nil(t, loadLibraryForCompletion(f, libDir),
		"cancelled context must produce nil (timeout branch)")
}

// TestLoadLibraryForCompletion_BypassesFLibrary pins the design decision
// documented in cmd/completions.go:88-95 — completion lookups MUST bypass
// f.Library() because it is sync.OnceValues-cached and would permanently
// pin the first error. We force f.Library to return an error and verify
// loadLibraryForCompletion still succeeds by loading the library directly.
func TestLoadLibraryForCompletion_BypassesFLibrary(t *testing.T) {
	t.Parallel()

	libDir := makeCompletionTestLibrary(t)

	f := newFactoryWithCache(context.Background())
	f.Library = func() (*library.Library, error) {
		return nil, errors.New("f.Library deliberately pinned to error")
	}

	got := loadLibraryForCompletion(f, libDir)
	require.NotNil(t, got,
		"loadLibraryForCompletion MUST bypass f.Library and use library.LoadLibrary directly")
	assert.Equal(t, libDir, got.RootPath)
}

// makeCompletionTestLibrary scaffolds a minimal library dir on disk and
// returns its RootPath. Used by the loadLibraryForCompletion tests.
func makeCompletionTestLibrary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"skills", "agents", "commands", "memory"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, sub), 0o750))
	}
	lib := &library.Library{
		Version:   "1",
		RootPath:  dir,
		Resources: map[string]map[string]library.Resource{},
		Presets:   map[string]library.Preset{},
	}
	require.NoError(t, library.SaveLibrary(lib))
	return dir
}

// =============================================================================
// resolveLibraryPath — shell-completion spec "Library Path Resolution for Completions"
// (4 priority-chain scenarios)
// =============================================================================

// newCmdWithLibraryFlag constructs a minimal cobra command with a registered
// --library string flag. Used by resolveLibraryPath tests.
func newCmdWithLibraryFlag(t *testing.T, value string, changed bool) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("library", "", "library path")
	if value != "" {
		require.NoError(t, cmd.Flags().Set("library", value))
	}
	if changed {
		cmd.Flag("library").Changed = true
	}
	return cmd
}

// TestResolveLibraryPath_FlagWins — the --library flag is the highest
// priority in the resolution chain.
func TestResolveLibraryPath_FlagWins(t *testing.T) {
	t.Parallel()

	cmd := newCmdWithLibraryFlag(t, "/flag/path", true)
	cfg := &config.Config{Library: "/config/path"}

	assert.Equal(t, "/flag/path", resolveLibraryPath(cmd, cfg),
		"--library flag MUST win over config")
}

// TestResolveLibraryPath_EnvWinsOverConfig — the GERMINATOR_LIBRARY env
// var is the second priority, ahead of config but behind the flag.
func TestResolveLibraryPath_EnvWinsOverConfig(t *testing.T) {
	t.Setenv("GERMINATOR_LIBRARY", "/env/path")

	cmd := newCmdWithLibraryFlag(t, "", false)
	cfg := &config.Config{Library: "/config/path"}

	assert.Equal(t, "/env/path", resolveLibraryPath(cmd, cfg),
		"env var MUST win over config when flag is unset")
}

// TestResolveLibraryPath_ConfigWinsOverDefault — the config file is the
// third priority, ahead of the built-in default.
func TestResolveLibraryPath_ConfigWinsOverDefault(t *testing.T) {
	t.Setenv("GERMINATOR_LIBRARY", "")

	cmd := newCmdWithLibraryFlag(t, "", false)
	cfg := &config.Config{Library: "/config/path"}

	assert.Equal(t, "/config/path", resolveLibraryPath(cmd, cfg),
		"config MUST win over default when flag and env are unset")
}

// TestResolveLibraryPath_DefaultFallback — with no flag, env, or config,
// the built-in default library path is returned.
func TestResolveLibraryPath_DefaultFallback(t *testing.T) {
	t.Setenv("GERMINATOR_LIBRARY", "")

	cmd := newCmdWithLibraryFlag(t, "", false)

	got := resolveLibraryPath(cmd, nil)
	assert.Equal(t, library.DefaultLibraryPath(), got,
		"default path MUST be returned when no other source is set")
}

// TestResolveLibraryPath_TildeExpansion covers the ~/ expansion helper
// used by all three priority levels.
func TestResolveLibraryPath_TildeExpansion(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("UserHomeDir unavailable; skipping tilde expansion test")
	}

	got := expandTildeInPath("~/my-lib")
	assert.Equal(t, home+"/my-lib", got,
		"~/my-lib MUST expand to <home>/my-lib")
}

// =============================================================================
// Silent Failure on Completion Errors — shell-completion spec scenarios
// (3 scenarios; library not found, timeout, malformed)
// =============================================================================

// TestActionResources_SilentOnNonexistentLibrary covers the "Library not
// found returns empty completions" scenario. The carapace Action should
// produce zero values when the library cannot be loaded.
func TestActionResources_SilentOnNonexistentLibrary(t *testing.T) {
	t.Parallel()

	f := newFactoryWithCache(context.Background())
	cmd := newCmdWithLibraryFlag(t, "/no/such/path", true)

	values := actionValuesAsStrings(t,
		actionResources(f, cmd).Invoke(carapace.Context{}))

	assert.Empty(t, values,
		"actionResources MUST return empty values when library is missing")
}

// TestActionResources_SilentOnCancelledContext covers the "Timeout
// returns empty completions" scenario via a cancelled RootContext.
func TestActionResources_SilentOnCancelledContext(t *testing.T) {
	t.Parallel()

	libDir := makeCompletionTestLibrary(t)

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	f := newFactoryWithCache(cancelledCtx)

	cmd := newCmdWithLibraryFlag(t, libDir, true)

	values := actionValuesAsStrings(t,
		actionResources(f, cmd).Invoke(carapace.Context{}))

	assert.Empty(t, values,
		"actionResources MUST return empty values when the context is cancelled")
}

// TestActionResources_SilentOnMalformedLibrary covers the "Invalid
// library returns empty completions" scenario.
func TestActionResources_SilentOnMalformedLibrary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "library.yaml"),
		[]byte("this is :: not valid [ yaml at all"), 0o600))

	f := newFactoryWithCache(context.Background())
	cmd := newCmdWithLibraryFlag(t, dir, true)

	values := actionValuesAsStrings(t,
		actionResources(f, cmd).Invoke(carapace.Context{}))

	assert.Empty(t, values,
		"actionResources MUST return empty values when library.yaml is malformed")
}
