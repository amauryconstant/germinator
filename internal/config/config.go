// Package config provides configuration loading and management for germinator.
package config

import (
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// Config holds the application configuration.
//
// Env-var mapping (via the koanf env provider at `Manager.Load()`):
//   - `Config.Library` ← `GERMINATOR_LIBRARY`
//   - `Config.PlatformDefault` (koanf tag `platform`) ←
//     `GERMINATOR_PLATFORM` (NOT `GERMINATOR_PLATFORM_DEFAULT`; the
//     prefix is stripped and the remaining key is lowercased)
//   - `Config.Debug` ← `GERMINATOR_DEBUG`
//   - `Config.Completion.Timeout` ← `GERMINATOR_COMPLETION.TIMEOUT`
//   - `Config.Completion.CacheTTL` ← `GERMINATOR_COMPLETION.CACHE_TTL`
//
// Bool truthiness rule for env-derived bool fields (`Config.Debug`):
// koanf parses via `strconv.ParseBool` semantics — `1` / `t` / `T` /
// `true` / `TRUE` / `True` resolve to `true`; all other non-empty
// strings resolve to `false`; unset defaults to the struct default.
type Config struct {
	// Library is the path to the library directory. When empty (the
	// `DefaultConfig()` seed), library-path resolution falls through
	// to `DefaultLibraryPath()` (XDG-resolved via
	// `adrg/xdg.DataFile("germinator/library")`).
	Library string `koanf:"library"`

	// PlatformDefault is the default target platform
	// (`claude-code` or `opencode`) for commands that opt in via a
	// follow-up change. Empty means platform must be specified via
	// flag (the historical default). The koanf tag remains `platform`
	// so existing config files continue to bind the same key.
	PlatformDefault string `koanf:"platform"`

	// Debug enables debug-level structured logging when true. Driven
	// by `GERMINATOR_DEBUG` (env) or `debug = true` (config file) per
	// the bool truthiness rule documented on the package.
	Debug bool `koanf:"debug"`

	// Completion holds the shell completion configuration.
	Completion CompletionConfig `koanf:"completion"`
}

// CompletionConfig holds configuration for shell completion.
type CompletionConfig struct {
	// Timeout is the maximum time for library loading during completion.
	// Default: "500ms"
	Timeout string `koanf:"timeout"`

	// CacheTTL is the duration to cache library data during completion.
	// Default: "5s"
	CacheTTL string `koanf:"cache_ttl"`
}

// DefaultConfig returns a Config with sensible defaults.
//
// `Library: ""` is the canonical "no config-file override" signal —
// when the value is empty at resolution time, library-path
// resolution falls through to `library.DefaultLibraryPath()` (XDG
// via `adrg/xdg.DataFile`).
func DefaultConfig() *Config {
	return &Config{
		Library:         "",
		PlatformDefault: "",
		Debug:           false,
		Completion: CompletionConfig{
			Timeout:  "500ms",
			CacheTTL: "5s",
		},
	}
}

// Validate checks that the configuration is valid.
// Returns *core.ConfigError if PlatformDefault is non-empty and
// not one of the supported platforms. Debug is always valid (bool);
// Library is always valid (empty falls through to XDG default at
// resolution time).
func (c *Config) Validate() error {
	if c.PlatformDefault == "" {
		return nil
	}

	if c.PlatformDefault != gerrors.PlatformClaudeCode && c.PlatformDefault != gerrors.PlatformOpenCode {
		return gerrors.NewConfigError(
			"platform",
			c.PlatformDefault,
			"unknown platform",
		).WithSuggestions([]string{gerrors.PlatformClaudeCode, gerrors.PlatformOpenCode})
	}

	return nil
}

// ExpandPaths expands the tilde (~) in paths to the user's home directory.
// This should be called after loading the config.
func (c *Config) ExpandPaths() error {
	expanded, err := expandTilde(c.Library)
	if err != nil {
		return err
	}
	c.Library = expanded
	return nil
}

// expandTilde expands a leading ~ in a path to the user's home directory.
func expandTilde(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", gerrors.NewConfigError("path", path, "cannot determine home directory")
		}
		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
