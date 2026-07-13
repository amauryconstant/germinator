//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// TestResolveConfigPath_HonorsXDGConfigHome verifies that when
// XDG_CONFIG_HOME is set, resolveConfigPath returns
// $XDG_CONFIG_HOME/germinator/config.toml.
func TestResolveConfigPath_HonorsXDGConfigHome(t *testing.T) {
	xdgHome := t.TempDir()

	// adrg/xdg requires the parent of the returned path to exist
	// when computing the config location; create the germinator
	// subdir so ConfigFile resolves successfully.
	germinatorDir := filepath.Join(xdgHome, "germinator")
	if err := os.MkdirAll(germinatorDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := resolveConfigPath()
	want := filepath.Join(xdgHome, "germinator", "config.toml")
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
	germinatorDir := filepath.Join(xdgHome, "germinator")
	if err := os.MkdirAll(germinatorDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := resolveConfigPath()
	want := filepath.Join(xdgHome, "germinator", "config.toml")
	if got != want {
		t.Errorf("resolveConfigPath() = %q, want %q", got, want)
	}

	// Confirm xdg.Reload was called — ConfigHome now reflects XDG_CONFIG_HOME.
	if xdg.ConfigHome != xdgHome {
		t.Errorf("xdg.ConfigHome = %q, want %q (xdg.Reload should pick up env)", xdg.ConfigHome, xdgHome)
	}
}
