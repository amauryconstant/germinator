package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLibrary(t *testing.T) {
	tests := []struct {
		name     string
		flagPath string
		envPath  string
		wantPath string
	}{
		{
			name:     "flag takes priority",
			flagPath: "/flag/path",
			envPath:  "/env/path",
			wantPath: "/flag/path",
		},
		{
			name:     "env when no flag",
			flagPath: "",
			envPath:  "/env/path",
			wantPath: "/env/path",
		},
		{
			name:     "default when no flag or env",
			flagPath: "",
			envPath:  "",
			wantPath: DefaultLibraryPath(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLibrary(tt.flagPath, tt.envPath)
			if got != tt.wantPath {
				t.Errorf("FindLibrary() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestDefaultLibraryPath(t *testing.T) {
	path := DefaultLibraryPath()

	if !filepath.IsAbs(path) {
		t.Errorf("DefaultLibraryPath() should return absolute path, got %s", path)
	}
}

func TestDefaultLibraryPathXDGDataHome(t *testing.T) {
	original := os.Getenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_DATA_HOME", original) //nolint:errcheck
	})

	if err := os.Setenv("XDG_DATA_HOME", "/custom/data"); err != nil {
		t.Fatalf("Failed to set XDG_DATA_HOME: %v", err)
	}

	path := DefaultLibraryPath()

	expected := filepath.Join("/custom/data", "germinator", "library")
	if path != expected {
		t.Errorf("DefaultLibraryPath() with XDG_DATA_HOME = %q, got %q, want %q", "/custom/data", path, expected)
	}
}

func TestExists(t *testing.T) {
	// Test with existing directory
	tmpDir := t.TempDir()
	if !Exists(tmpDir) {
		t.Error("Exists() should return true for existing directory")
	}

	// Test with non-existing directory
	if Exists("/nonexistent/path") {
		t.Error("Exists() should return false for non-existing directory")
	}
}

func TestYAMLExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test without library.yaml
	if YAMLExists(tmpDir) {
		t.Error("YAMLExists() should return false when no library.yaml")
	}

	// Create library.yaml
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte("version: \"1\""), 0644); err != nil {
		t.Fatalf("Failed to create library.yaml: %v", err)
	}

	// Test with library.yaml
	if !YAMLExists(tmpDir) {
		t.Error("YAMLExists() should return true when library.yaml exists")
	}
}
