package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"gitlab.com/amoconst/germinator/internal/core"
)

// Manager defines the interface for configuration management.
type Manager interface {
	// Load reads and parses the configuration file.
	// Returns an error if the file exists but cannot be read or parsed.
	// Missing config file is not an error - defaults are used.
	Load() error

	// GetConfig returns the loaded configuration.
	// Returns nil if Load() has not been called.
	GetConfig() *Config
}

// koanfConfigManager implements Manager using Koanf.
type koanfConfigManager struct {
	config *Config
}

// Compile-time confirmation that *koanfConfigManager satisfies the
// Manager contract. If either side changes (interface or struct),
// the build fails immediately. *koanfConfigManager is the live
// receiver used by NewConfigManager and Load, so no suppression
// directive is required.
var _ Manager = (*koanfConfigManager)(nil)

// NewConfigManager creates a new Manager.
func NewConfigManager() Manager {
	return &koanfConfigManager{
		config: DefaultConfig(),
	}
}

// Load reads and parses the configuration file from the XDG-compliant
// location. The merge order is: defaults → file → env vars (prefix
// `GERMINATOR_`, delimiter `.`, lowercased keys via the koanf env
// provider). Missing config file is not an error - defaults are used.
//
// Env key mapping: `Config.Library` maps to env `GERMINATOR_LIBRARY`;
// `Config.PlatformDefault` (koanf tag `platform`) maps to
// `GERMINATOR_PLATFORM` (NOT `GERMINATOR_PLATFORM_DEFAULT`); the
// prefix is stripped and the remaining key is lowercased.
//
// Bool truthiness rule for env-derived bool fields (e.g.,
// `Config.Debug`): koanf parses via `strconv.ParseBool` semantics —
// `1` / `t` / `T` / `true` / `TRUE` / `True` resolve to `true`; all
// other non-empty strings resolve to `false`; unset defaults to the
// struct default. See `TestConfig_EnvVarBoolTruthinessRule`.
func (m *koanfConfigManager) Load() error {
	configPath := resolveConfigPath()

	k := koanf.New(".")

	// Tier 1: defaults are already seeded into m.config by
	// NewConfigManager (the empty `Config` struct field zero-values
	// are not used here — `DefaultConfig()` populates every field).

	// Tier 2: load from config file when present.
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			if !os.IsNotExist(err) {
				return core.NewFileError(configPath, "reading", "failed to read config file", err)
			}
			// File does not exist; fall through to env tier so
			// env-only configs still apply (a missing file is not
			// an error per the documented "missing file = defaults"
			// contract, but env vars must still be honored).
		} else {
			if err := k.Unmarshal("", m.config); err != nil {
				return core.NewParseError(configPath, "failed to parse config", err)
			}
		}
	}

	// Tier 3: env vars override file values. The koanf env provider
	// strips the GERMINATOR_ prefix and lowercases the key (via the
	// callback), so GERMINATOR_LIBRARY maps to Config.Library,
	// GERMINATOR_PLATFORM maps to Config.PlatformDefault (koanf tag
	// "platform"), and GERMINATOR_DEBUG maps to Config.Debug. The
	// delimiter is "." so nested keys (e.g.,
	// GERMINATOR_COMPLETION.TIMEOUT) become nested map entries.
	if err := k.Load(env.Provider("GERMINATOR_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "GERMINATOR_"))
	}), nil); err != nil {
		return core.NewParseError("env", "failed to load env vars", err)
	}
	if err := k.Unmarshal("", m.config); err != nil {
		return core.NewParseError("env", "failed to apply env vars", err)
	}

	// Validate the config
	if err := m.config.Validate(); err != nil {
		return err
	}

	// Expand paths (e.g., ~ to home directory)
	if err := m.config.ExpandPaths(); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the loaded configuration.
func (m *koanfConfigManager) GetConfig() *Config {
	return m.config
}

// resolveConfigPath returns the path to the config file.
//
// It first asks `adrg/xdg` for the canonical config location (which
// resolves `$XDG_CONFIG_HOME` per the XDG Base Directory
// Specification). On error or empty result, it falls back to the
// current working directory `./config.toml` for projects that ship
// their own config alongside `germinator`.
//
// The returned path may not exist — a missing config file is not an
// error at the caller level (`Load()` falls through to defaults).
func resolveConfigPath() string {
	xdg.Reload()
	if path, err := xdg.ConfigFile("germinator/config.toml"); err == nil && path != "" {
		return path
	}
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "config.toml")
	}
	return ""
}

// GetConfigPath returns the path where the config file would be located.
// This is useful for displaying messages about config location.
// It returns the XDG-resolved path even if the file does not exist
// (does NOT attempt to create the directory).
func GetConfigPath() (string, error) {
	xdg.Reload()
	if xdg.ConfigHome != "" {
		return filepath.Join(xdg.ConfigHome, "germinator", "config.toml"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "germinator", "config.toml"), nil
}
