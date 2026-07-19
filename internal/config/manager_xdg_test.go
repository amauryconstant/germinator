//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// TestResolveConfigPath_HonorsXDGConfigHome verifies that when
// XDG_CONFIG_HOME is set and a config file exists there,
// resolveConfigPath returns $XDG_CONFIG_HOME/germinator/config.toml.
// The file must exist because resolveConfigPath falls back to CWD
// when the XDG path has no config.toml (see
// TestResolveConfigPath_FallsBackToCWD).
func TestResolveConfigPath_HonorsXDGConfigHome(t *testing.T) {
	xdgHome := t.TempDir()

	// Write a config file at the XDG location so the existence check
	// in resolveConfigPath accepts the XDG path.
	xdgConfig := filepath.Join(xdgHome, "germinator", "config.toml")
	if err := os.MkdirAll(filepath.Dir(xdgConfig), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(xdgConfig, []byte(`platform = "claude-code"`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := resolveConfigPath()
	want := xdgConfig
	if got != want {
		t.Errorf("resolveConfigPath() = %q, want %q", got, want)
	}
}

// TestResolveConfigPath_FallsBackToCWD verifies that when both
// XDG_CONFIG_HOME and HOME are unset (or HOME points to a nonexistent
// dir), resolveConfigPath falls back to ./config.toml in the CWD.
func TestResolveConfigPath_FallsBackToCWD(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a config in CWD
	cwdConfig := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(cwdConfig, []byte(`platform = "claude-code"`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/nonexistent")

	got := resolveConfigPath()
	if got != cwdConfig {
		t.Errorf("resolveConfigPath() = %q, want %q", got, cwdConfig)
	}
}

// TestResolveConfigPath_EnvReloadAfterChange confirms that xdg's
// Reload picks up env mutations between calls (used in tests where
// t.Setenv may run mid-suite).
func TestResolveConfigPath_EnvReloadAfterChange(t *testing.T) {
	tmpDir := t.TempDir()
	xdgHome := filepath.Join(tmpDir, "xdg")
	xdgConfig := filepath.Join(xdgHome, "germinator", "config.toml")
	if err := os.MkdirAll(filepath.Dir(xdgConfig), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	// resolveConfigPath only returns the XDG path when the file exists.
	if err := os.WriteFile(xdgConfig, []byte(`platform = "claude-code"`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := resolveConfigPath()
	want := xdgConfig
	if got != want {
		t.Errorf("resolveConfigPath() = %q, want %q", got, want)
	}

	// Confirm xdg.Reload was called — ConfigHome now reflects XDG_CONFIG_HOME.
	if xdg.ConfigHome != xdgHome {
		t.Errorf("xdg.ConfigHome = %q, want %q (xdg.Reload should pick up env)", xdg.ConfigHome, xdgHome)
	}
}
