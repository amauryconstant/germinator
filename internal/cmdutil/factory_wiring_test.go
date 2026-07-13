package cmdutil

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// TestBuildFactory_ConfigIsLazy verifies that f.Config is set after
// BuildFactory returns and is backed by sync.OnceValues-style caching.
// Pins the cli-cli-factory/spec.md "Config is a lazy function" scenario.
func TestBuildFactory_ConfigIsLazy(t *testing.T) {
	ios := iostreams.Test()

	var calls atomic.Int32
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		calls.Add(1)
		return config.DefaultConfig(), nil
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	require.NotNil(t, f)
	defer f.Close()

	require.NotNil(t, f.Config, "BuildFactory MUST wire f.Config")

	// Three successive calls must invoke the inner loader exactly once
	// because OnceValuesFunc caches the result.
	_, _ = f.Config()
	_, _ = f.Config()
	_, _ = f.Config()
	assert.Equal(t, int32(1), calls.Load(),
		"f.Config MUST cache the inner config.Load via OnceValuesFunc")
}

// TestBuildFactory_ConfigCachedAcrossCallsConcurrent stresses the
// cache under concurrency: 50 goroutines invoke f.Config() and the
// inner loader runs exactly once.
func TestBuildFactory_ConfigCachedAcrossCallsConcurrent(t *testing.T) {
	ios := iostreams.Test()

	var calls atomic.Int32
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		calls.Add(1)
		return config.DefaultConfig(), nil
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	defer f.Close()

	done := make(chan struct{}, 50)
	for i := 0; i < 50; i++ {
		go func() {
			_, _ = f.Config()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	assert.Equal(t, int32(1), calls.Load(),
		"concurrent f.Config() calls MUST invoke the loader exactly once")
}

// TestBuildFactory_DebugEnablesLogger verifies that cfg.Debug=true
// flows through to IOStreams.SetDebug, enabling the debug-level logger.
// Pins the spec scenario "Debug activation flows through Config".
func TestBuildFactory_DebugEnablesLogger(t *testing.T) {
	ios := iostreams.Test()
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Debug = true
		return cfg, nil
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	defer f.Close()

	require.NotNil(t, f.IOStreams.Logger)
	assert.True(t, f.IOStreams.Logger.Enabled(context.Background(), slog.LevelDebug),
		"BuildFactory MUST activate debug-level Logger when cfg.Debug=true")
}

// TestBuildFactory_DebugDisabledByDefault verifies that the default
// cfg.Debug=false leaves the Logger as a discard handler (no debug
// emission). Pins the "logger uses noop when debug unset" subtest.
func TestBuildFactory_DebugDisabledByDefault(t *testing.T) {
	ios := iostreams.Test()
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		return config.DefaultConfig(), nil
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	defer f.Close()

	require.NotNil(t, f.IOStreams.Logger)
	assert.False(t, f.IOStreams.Logger.Enabled(context.Background(), slog.LevelDebug),
		"BuildFactory MUST keep Logger as discard when cfg.Debug=false")
}

// TestBuildFactory_ConfigLoadErrorSurfaces verifies that BuildFactory
// returns the error from config.Load unchanged so main.go can map it
// to the appropriate exit code via cmdutil.ExitCodeFor. Pins the
// spec scenario "Config load errors propagate from BuildFactory".
func TestBuildFactory_ConfigLoadErrorSurfaces(t *testing.T) {
	ios := iostreams.Test()
	wantErr := core.NewConfigError("platform", "invalid", "unknown platform")
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		return config.DefaultConfig(), wantErr
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.Error(t, err)
	require.NotNil(t, f, "BuildFactory MUST return a non-nil Factory even when config load errors")
	defer f.Close()

	var cfgErr *core.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("BuildFactory error = %T, want *core.ConfigError", err)
	}
}

// TestBuildFactory_LibraryIsLazy verifies that f.Library is wired
// (non-nil) and respects the flag > env > cfg > XDG default priority
// chain via library.FindLibrary. Pins task 4.4's closure migration.
func TestBuildFactory_LibraryIsLazy(t *testing.T) {
	ios := iostreams.Test()
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Library = "/from/cfg/path"
		return cfg, nil
	}))

	t.Setenv("GERMINATOR_LIBRARY", "")
	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	defer f.Close()

	require.NotNil(t, f.Library, "BuildFactory MUST wire f.Library")

	// The closure must NOT be called eagerly; verify by checking that
	// calling it surfaces an error from library.LoadLibrary (the path
	// does not exist). The error message embeds the resolved path,
	// which must come from cfg.Library — the XDG-fallthrough tier is
	// bypassed because cfg.Library is non-empty.
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", "/nonexistent")
	_, err = f.Library()
	require.Error(t, err, "loadLibrary for nonexistent path should error")
	assert.Contains(t, err.Error(), "/from/cfg/path",
		"f.Library() resolved path MUST reflect cfg.Library when flag/env are unset")
}

// TestBuildFactory_CompletionCacheAssigned verifies that
// BuildFactory populates Factory.CompletionCache so completion actions
// have a working cache immediately.
func TestBuildFactory_CompletionCacheAssigned(t *testing.T) {
	ios := iostreams.Test()
	t.Cleanup(swapConfigLoadForTest(func() (*config.Config, error) {
		return config.DefaultConfig(), nil
	}))

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err)
	defer f.Close()

	assert.NotNil(t, f.CompletionCache,
		"BuildFactory MUST populate CompletionCache")
}

// TestBuildFactory_NoEnvStillBuilds verifies the resilience property
// that BuildFactory succeeds even with no config file, no env vars,
// and no XDG path (a fresh install with zero configuration).
func TestBuildFactory_NoEnvStillBuilds(t *testing.T) {
	ios := iostreams.Test()

	// Restore production loader — defaults apply when env and file
	// are absent.
	t.Cleanup(swapConfigLoadForTest(config.Load))

	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmpDir)
	// Working dir is t.TempDir so ./config.toml does not exist.
	origWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	f, err := BuildFactory(context.Background(), ios, "v0.0.0", "germinator")
	require.NoError(t, err, "BuildFactory MUST succeed with zero configuration")
	defer f.Close()

	require.NotNil(t, f.Config)
	// Subsequent call must return the cached *Config (default values).
	cfg, err := f.Config()
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Library)
	assert.False(t, cfg.Debug)
}
