package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/core"
)

func TestNewConfigManager(t *testing.T) {
	mgr := NewConfigManager()
	if mgr == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	cfg := mgr.GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil before Load()")
	}

	// Should have defaults
	if cfg.Library != "" {
		t.Errorf("default Library = %q, want %q (empty — XDG falls through at resolution time)", cfg.Library, "")
	}
}

func TestConfigManagerLoad_NoConfigFile(t *testing.T) {
	// Create a temp directory and set it as home to ensure no config exists
	tmpDir := t.TempDir()

	// Save original env vars
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")

	// Set temp as home
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	// Change to temp dir to avoid picking up project config
	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWd)
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() with no config file returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	if cfg.PlatformDefault != "" {
		t.Errorf("expected empty PlatformDefault with no config, got %q", cfg.PlatformDefault)
	}
}

func TestConfigManagerLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write valid config
	configContent := `
platform = "opencode"
library = "/custom/library/path"
`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and clear env vars to ensure test isolation
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	if cfg.PlatformDefault != core.PlatformOpenCode {
		t.Errorf("PlatformDefault = %q, want %q", cfg.PlatformDefault, core.PlatformOpenCode)
	}
	if cfg.Library != "/custom/library/path" {
		t.Errorf("Library = %q, want %q", cfg.Library, "/custom/library/path")
	}
}

func TestConfigManagerLoad_XDGConfigHome(t *testing.T) {
	tmpDir := t.TempDir()

	// Create XDG config directory
	configDir := filepath.Join(tmpDir, "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write valid config
	configContent := `platform = "claude-code"`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set XDG_CONFIG_HOME
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", origXDG) }()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	if cfg.PlatformDefault != core.PlatformClaudeCode {
		t.Errorf("PlatformDefault = %q, want %q", cfg.PlatformDefault, core.PlatformClaudeCode)
	}
}

func TestConfigManagerLoad_InvalidPlatform(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config with invalid platform
	configContent := `platform = "invalid"`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and clear env vars to ensure test isolation
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid platform, got nil")
	}

	// Should mention available platforms
	if !containsStr(err.Error(), "💡") {
		t.Errorf("error should mention available platforms (with 💡), got: %v", err)
	}
}

func TestConfigManagerLoad_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write invalid TOML
	configContent := `platform = "unclosed`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and clear env vars to ensure test isolation
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid TOML, got nil")
	}
}

func TestConfigManagerLoad_TildeExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config with tilde path
	configContent := `library = "~/my-library"`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and clear env vars to ensure test isolation
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	wantLibrary := filepath.Join(tmpDir, "my-library")
	if cfg.Library != wantLibrary {
		t.Errorf("Library = %q, want %q", cfg.Library, wantLibrary)
	}
}

func TestConfigManagerLoad_CurrentDirConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Write config in current dir
	configContent := `platform = "claude-code"`
	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Change to temp dir
	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Clear HOME to ensure we pick up local config
	origHome := os.Getenv("HOME")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("HOME", "/nonexistent"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
	}()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	if cfg.PlatformDefault != core.PlatformClaudeCode {
		t.Errorf("PlatformDefault = %q, want %q", cfg.PlatformDefault, core.PlatformClaudeCode)
	}
}

func TestResolveConfigPath_Precedence(t *testing.T) {
	// This tests that XDG_CONFIG_HOME takes precedence over HOME/.config
	// when both exist

	tmpDir := t.TempDir()
	xdgDir := filepath.Join(tmpDir, "xdg")
	homeDir := filepath.Join(tmpDir, "home")

	// Create both config directories
	xdgConfigDir := filepath.Join(xdgDir, "germinator")
	homeConfigDir := filepath.Join(homeDir, ".config", "germinator")
	if err := os.MkdirAll(xdgConfigDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(homeConfigDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Write different configs to each
	if err := os.WriteFile(filepath.Join(xdgConfigDir, "config.toml"), []byte(`platform = "opencode"`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeConfigDir, "config.toml"), []byte(`platform = "claude-code"`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Set env vars
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", xdgDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
		_ = os.Setenv("HOME", origHome)
	}()

	path := resolveConfigPath()

	// Should return XDG path, not HOME path
	wantPath := filepath.Join(xdgDir, "germinator", "config.toml")
	if path != wantPath {
		t.Errorf("resolveConfigPath() = %q, want %q", path, wantPath)
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		xdgHome  string
		home     string
		wantPath string
	}{
		{
			name:     "XDG_CONFIG_HOME set",
			xdgHome:  "/custom/xdg",
			home:     "/home/user",
			wantPath: "/custom/xdg/germinator/config.toml",
		},
		{
			name:     "XDG_CONFIG_HOME not set, uses HOME",
			xdgHome:  "",
			home:     "/home/user",
			wantPath: "/home/user/.config/germinator/config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origXDG := os.Getenv("XDG_CONFIG_HOME")
			origHome := os.Getenv("HOME")

			if err := os.Setenv("XDG_CONFIG_HOME", tt.xdgHome); err != nil {
				t.Fatalf("Setenv failed: %v", err)
			}
			if err := os.Setenv("HOME", tt.home); err != nil {
				t.Fatalf("Setenv failed: %v", err)
			}

			defer func() {
				_ = os.Setenv("XDG_CONFIG_HOME", origXDG)
				_ = os.Setenv("HOME", origHome)
			}()

			got, err := GetConfigPath()
			if err != nil {
				t.Fatalf("GetConfigPath() error: %v", err)
			}

			if got != tt.wantPath {
				t.Errorf("GetConfigPath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// clearEnv unsets a set of env vars for the duration of the test, so
// callers can isolate env-provider behavior. Uses t.Setenv with empty
// value because t.Setenv requires a non-empty name; the helper is
// only useful for vars that should be considered "unset" — koanf's
// env provider treats empty values as present-but-empty (which parses
// as zero value), but tests below use this only for confirming the
// provider does NOT see the var.
func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		orig, hadOrig := os.LookupEnv(k)
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("Unsetenv(%s) failed: %v", k, err)
		}
		t.Cleanup(func() {
			if hadOrig {
				_ = os.Setenv(k, orig)
			}
		})
	}
}

// TestDefaultConfig_LibraryIsEmpty pins the Library: "" shape change.
// This test is read-only (no env mutation) so it MAY use t.Parallel().
func TestDefaultConfig_LibraryIsEmpty(t *testing.T) {
	t.Parallel()
	if got := DefaultConfig().Library; got != "" {
		t.Errorf("DefaultConfig().Library = %q, want %q (empty — XDG falls through at resolution time)", got, "")
	}
}

// TestLoad_EnvOverridesFile verifies the merge order defaults → file → env.
func TestLoad_EnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	configContent := `library = "/file/lib"`
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmpDir)
	t.Setenv("GERMINATOR_LIBRARY", "/env/lib")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Library != "/env/lib" {
		t.Errorf("Load().Library = %q, want %q (env should override file)", cfg.Library, "/env/lib")
	}
}

// TestLoad_EnvOverridesDefault verifies env vars override defaults.
func TestLoad_EnvOverridesDefault(t *testing.T) {
	clearEnv(t, "XDG_CONFIG_HOME", "HOME")
	origWd, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("GERMINATOR_DEBUG", "1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.Debug {
		t.Errorf("Load().Debug = false, want true (env should override default)")
	}
}

// TestLoad_NoEnvNoFile verifies defaults are returned when neither
// env nor file is set.
func TestLoad_NoEnvNoFile(t *testing.T) {
	clearEnv(t, "XDG_CONFIG_HOME", "HOME", "GERMINATOR_LIBRARY", "GERMINATOR_DEBUG", "GERMINATOR_PLATFORM")
	origWd, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Library != "" {
		t.Errorf("Load().Library = %q, want %q (default is empty)", cfg.Library, "")
	}
	if cfg.Debug {
		t.Errorf("Load().Debug = true, want false (default)")
	}
	if cfg.PlatformDefault != "" {
		t.Errorf("Load().PlatformDefault = %q, want empty", cfg.PlatformDefault)
	}
}

// TestLoad_MissingFile verifies a missing config file is not an error.
func TestLoad_MissingFile(t *testing.T) {
	clearEnv(t, "XDG_CONFIG_HOME", "HOME", "GERMINATOR_LIBRARY", "GERMINATOR_DEBUG", "GERMINATOR_PLATFORM")
	origWd, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("GERMINATOR_LIBRARY", "/env/lib")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with missing file error: %v", err)
	}
	if cfg.Library != "/env/lib" {
		t.Errorf("Load().Library = %q, want %q (env should still apply)", cfg.Library, "/env/lib")
	}
}

// TestLoad_EnvKeyMapping_PlatformDefault pins the lowercase-after-
// prefix-stripping key mapping: GERMINATOR_PLATFORM maps to the
// `platform` koanf tag (Config.PlatformDefault), NOT
// GERMINATOR_PLATFORM_DEFAULT.
func TestLoad_EnvKeyMapping_PlatformDefault(t *testing.T) {
	clearEnv(t, "XDG_CONFIG_HOME", "HOME", "GERMINATOR_LIBRARY", "GERMINATOR_DEBUG")
	origWd, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("GERMINATOR_PLATFORM", "opencode")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.PlatformDefault != "opencode" {
		t.Errorf("Load().PlatformDefault = %q, want %q", cfg.PlatformDefault, "opencode")
	}
}

// TestConfig_EnvVarBoolTruthinessRule pins the koanf bool parsing
// rule via strconv.ParseBool semantics. The recognized values are:
//
//   - "true" (any case): 1, t, T, TRUE, True, true → true
//   - "false" (any case): 0, f, F, FALSE, False, false → false
//   - unparseable values (no, garbage, empty, etc.) → Load() returns
//     a *core.ParseError. The bool field is left at the default
//     (false), but the error chain is the authoritative signal.
//   - unset env var → struct default (false)
//
// This documents the koanf bool parsing rule at test level so future
// koanf upgrades can detect regressions.
func TestConfig_EnvVarBoolTruthinessRule(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		setEnv  bool
		want    bool
		wantErr bool
	}{
		{name: "1 is true", envVal: "1", setEnv: true, want: true},
		{name: "t is true", envVal: "t", setEnv: true, want: true},
		{name: "T is true", envVal: "T", setEnv: true, want: true},
		{name: "true is true", envVal: "true", setEnv: true, want: true},
		{name: "TRUE is true", envVal: "TRUE", setEnv: true, want: true},
		{name: "True is true", envVal: "True", setEnv: true, want: true},
		{name: "0 is false", envVal: "0", setEnv: true, want: false},
		{name: "f is false", envVal: "f", setEnv: true, want: false},
		{name: "F is false", envVal: "F", setEnv: true, want: false},
		{name: "false is false", envVal: "false", setEnv: true, want: false},
		{name: "no errors (unparseable)", envVal: "no", setEnv: true, want: false, wantErr: true},
		{name: "garbage errors (unparseable)", envVal: "garbage", setEnv: true, want: false, wantErr: true},
		{name: "empty string is false (koanf graceful default)", envVal: "", setEnv: true, want: false},
		{name: "unset is default (false)", envVal: "", setEnv: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t, "XDG_CONFIG_HOME", "HOME", "GERMINATOR_LIBRARY", "GERMINATOR_PLATFORM")
			origWd, _ := os.Getwd()
			tmpDir := t.TempDir()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Chdir failed: %v", err)
			}
			t.Cleanup(func() { _ = os.Chdir(origWd) })

			if tt.setEnv {
				t.Setenv("GERMINATOR_DEBUG", tt.envVal)
			}

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() expected parse error, got nil")
				}
				var parseErr *core.ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("Load() error = %T, want *core.ParseError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.Debug != tt.want {
				t.Errorf("Load().Debug = %v, want %v", cfg.Debug, tt.want)
			}
		})
	}
}

// TestLoad_TopLevelWrapper pins the contract that the top-level
// Load() returns a never-nil *Config (the same one as Manager.GetConfig()).
func TestLoad_TopLevelWrapper(t *testing.T) {
	clearEnv(t, "XDG_CONFIG_HOME", "HOME", "GERMINATOR_LIBRARY", "GERMINATOR_DEBUG", "GERMINATOR_PLATFORM")
	origWd, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil *Config")
	}
	if cfg.Library != "" {
		t.Errorf("Load().Library = %q, want empty default", cfg.Library)
	}
}

// TestLoad_TopLevelWrapperReturnsErrorOnInvalidConfig verifies the
// typed error chain documented in the cli-cli-factory spec.
func TestLoad_TopLevelWrapperReturnsErrorOnInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "germinator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	configContent := `platform = "invalid"`
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmpDir)
	clearEnv(t, "GERMINATOR_LIBRARY", "GERMINATOR_DEBUG", "GERMINATOR_PLATFORM")

	cfg, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid platform, got nil")
	}
	if cfg == nil {
		t.Fatal("Load() must return non-nil *Config even on error")
	}

	var configErr *core.ConfigError
	if !errors.As(err, &configErr) {
		t.Errorf("Load() error = %T, want *core.ConfigError", err)
	}
}
