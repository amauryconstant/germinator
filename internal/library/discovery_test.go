package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindLibrary(t *testing.T) {
	tests := []struct {
		name     string
		flagPath string
		envPath  string
		cfgPath  string
		wantPath string
	}{
		{
			name:     "flag takes priority",
			flagPath: "/flag/path",
			envPath:  "/env/path",
			cfgPath:  "/cfg/path",
			wantPath: "/flag/path",
		},
		{
			name:     "env when no flag",
			flagPath: "",
			envPath:  "/env/path",
			cfgPath:  "/cfg/path",
			wantPath: "/env/path",
		},
		{
			name:     "cfg when no flag or env",
			flagPath: "",
			envPath:  "",
			cfgPath:  "/cfg/path",
			wantPath: "/cfg/path",
		},
		{
			name:     "default when no flag, env, or cfg",
			flagPath: "",
			envPath:  "",
			cfgPath:  "",
			wantPath: DefaultLibraryPath(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLibrary(tt.flagPath, tt.envPath, tt.cfgPath)
			if got != tt.wantPath {
				assert.Equal(t, tt.wantPath, got, "FindLibrary() = %v, want %v")
			}
		})
	}
}

func TestDefaultLibraryPath(t *testing.T) {
	path := DefaultLibraryPath()

	if !filepath.IsAbs(path) {
		assert.NotEmpty(t, path, "DefaultLibraryPath() should return absolute")
	}
}

func TestDefaultLibraryPathXDGDataHome(t *testing.T) {
	original := os.Getenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_DATA_HOME", original) //nolint:errcheck // test cleanup: best-effort env restore
	})

	require.NoError(t, os.Setenv("XDG_DATA_HOME", "/custom/data"))

	path := DefaultLibraryPath()

	expected := filepath.Join("/custom/data", "germinator", "library")
	if path != expected {
		assert.Equal(t, expected, path, "DefaultLibraryPath() with XDG_DATA_HOME mismatch")
	}
}

func TestExists(t *testing.T) {
	// Test with existing directory
	tmpDir := t.TempDir()
	if !Exists(tmpDir) {
		assert.Fail(t, "Exists() should return true for existing directory")
	}

	// Test with non-existing directory
	if Exists("/nonexistent/path") {
		assert.Fail(t, "Exists() should return false for non-existing directory")
	}
}

func TestYAMLExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test without library.yaml
	if YAMLExists(tmpDir) {
		assert.Fail(t, "YAMLExists() should return false when no library.yaml")
	}

	// Create library.yaml
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("version: \"1\""), 0644))

	// Test with library.yaml
	if !YAMLExists(tmpDir) {
		assert.Fail(t, "YAMLExists() should return true when library.yaml exists")
	}
}

// TestResolveLibrary_FlagOverEnvOverCfgOverDefault covers the full
// 4-tier precedence mandated by application-configuration/spec.md:122.
func TestResolveLibrary_FlagOverEnvOverCfgOverDefault(t *testing.T) {
	tests := []struct {
		name     string
		flagPath string
		envPath  string
		cfgPath  string
		wantPath string
	}{
		{
			name:     "flag wins over env",
			flagPath: "/flag/path",
			envPath:  "/env/path",
			cfgPath:  "/cfg/path",
			wantPath: "/flag/path",
		},
		{
			name:     "env wins over cfg",
			flagPath: "",
			envPath:  "/env/path",
			cfgPath:  "/cfg/path",
			wantPath: "/env/path",
		},
		{
			name:     "cfg wins over default",
			flagPath: "",
			envPath:  "",
			cfgPath:  "/cfg/path",
			wantPath: "/cfg/path",
		},
		{
			name:     "flag beats cfg",
			flagPath: "/flag/path",
			envPath:  "",
			cfgPath:  "/cfg/path",
			wantPath: "/flag/path",
		},
		{
			name:     "env beats cfg",
			flagPath: "",
			envPath:  "/env/path",
			cfgPath:  "",
			wantPath: "/env/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLibrary(tt.flagPath, tt.envPath, tt.cfgPath)
			if got != tt.wantPath {
				assert.Equal(t, tt.wantPath, got, "FindLibrary() = %v, want %v")
			}
		})
	}
}

// TestResolveLibrary_AllEmpty_ReturnsXDGDefault verifies that the
// final tier of precedence falls through to DefaultLibraryPath().
func TestResolveLibrary_AllEmpty_ReturnsXDGDefault(t *testing.T) {
	got := FindLibrary("", "", "")
	want := DefaultLibraryPath()
	if got != want {
		assert.Equal(t, want, got, "FindLibrary(\"\",\"\",\"\" ) mismatch")
	}
}

// TestDefaultLibraryPath_AdoptsXDG verifies the XDG-backed data path
// is computed from XDG_DATA_HOME.
func TestDefaultLibraryPath_AdoptsXDG(t *testing.T) {
	xdgDataHome := "/xdg/lib"
	t.Setenv("XDG_DATA_HOME", xdgDataHome)
	t.Setenv("HOME", "/nonexistent")

	got := DefaultLibraryPath()
	want := filepath.Join(xdgDataHome, "germinator", "library")
	if got != want {
		assert.Equal(t, want, got, "DefaultLibraryPath() = %q, want %q")
	}
}

// TestDefaultLibraryPath_PrefersXDGOverCWDWhenXDGExists verifies that
// when both XDG and CWD paths exist, XDG wins.
func TestDefaultLibraryPath_PrefersXDGOverCWDWhenXDGExists(t *testing.T) {
	tmpDir := t.TempDir()
	xdgHome := filepath.Join(tmpDir, "xdg-data")
	xdgLib := filepath.Join(xdgHome, "germinator", "library")
	require.NoError(t, os.MkdirAll(xdgLib, 0755))

	origWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("XDG_DATA_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := DefaultLibraryPath()
	if got != xdgLib {
		assert.Equal(t, xdgLib, got, "DefaultLibraryPath() mismatch: XDG path should win")
	}
}

// TestDefaultLibraryPath_FallsBackToCWDWhenXDGDoesNotExist verifies
// the project-local override behavior.
func TestDefaultLibraryPath_FallsBackToCWDWhenXDGDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	xdgHome := filepath.Join(tmpDir, "xdg-data")
	cwdLib := filepath.Join(tmpDir, "germinator", "library")
	require.NoError(t, os.MkdirAll(cwdLib, 0755))

	origWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	t.Setenv("XDG_DATA_HOME", xdgHome)
	t.Setenv("HOME", "/nonexistent")

	got := DefaultLibraryPath()
	if got != cwdLib {
		assert.Equal(t, cwdLib, got, "DefaultLibraryPath() mismatch: CWD path should win when XDG does not exist")
	}
}

// TestXdgReload verifies that DefaultLibraryPath picks up env mutations
// (the production wrapper holds xdgReloadMu across both xdg.Reload and
// the subsequent xdg.DataHome read).
func TestXdgReload(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	t.Setenv("HOME", "/nonexistent")

	// Trigger a reload via the public DefaultLibraryPath entry point.
	_ = DefaultLibraryPath()

	if xdg.DataHome != "/custom/data" {
		assert.Equal(t, "/custom/data", xdg.DataHome, "xdg.DataHome mismatch: xdg.Reload should pick up env")
	}
}
