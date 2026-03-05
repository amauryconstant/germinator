package config

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
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
	if cfg.Library != "~/.config/germinator/library" {
		t.Errorf("default Library = %q, want %q", cfg.Library, "~/.config/germinator/library")
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
	if cfg.Platform != "" {
		t.Errorf("expected empty Platform with no config, got %q", cfg.Platform)
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

	// Set HOME to temp dir
	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", origHome) }()

	mgr := NewConfigManager()
	err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	cfg := mgr.GetConfig()
	if cfg.Platform != models.PlatformOpenCode {
		t.Errorf("Platform = %q, want %q", cfg.Platform, models.PlatformOpenCode)
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
	if cfg.Platform != models.PlatformClaudeCode {
		t.Errorf("Platform = %q, want %q", cfg.Platform, models.PlatformClaudeCode)
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

	// Set HOME to temp dir
	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", origHome) }()

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

	// Set HOME to temp dir
	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", origHome) }()

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

	// Set HOME to temp dir
	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", origHome) }()

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
	if cfg.Platform != models.PlatformClaudeCode {
		t.Errorf("Platform = %q, want %q", cfg.Platform, models.PlatformClaudeCode)
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

	path, err := resolveConfigPath()
	if err != nil {
		t.Fatalf("resolveConfigPath() error: %v", err)
	}

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
