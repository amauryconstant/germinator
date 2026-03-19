package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLibrary(t *testing.T) {
	// Get absolute path to fixtures
	fixturePath := filepath.Join("..", "..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	lib, err := LoadLibrary(absPath)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	// Verify version
	if lib.Version != "1" {
		t.Errorf("Library version = %v, want 1", lib.Version)
	}

	// Verify resources exist
	if len(lib.Resources) == 0 {
		t.Error("Library has no resources")
	}

	// Verify skills
	skills, ok := lib.Resources["skill"]
	if !ok {
		t.Error("Library has no skills")
	}
	if len(skills) < 2 {
		t.Errorf("Expected at least 2 skills, got %d", len(skills))
	}

	// Verify specific skill
	commit, ok := skills["commit"]
	if !ok {
		t.Error("Library missing skill/commit")
	}
	if commit.Path != "skills/skill-commit.md" {
		t.Errorf("skill/commit path = %v, want skills/skill-commit.md", commit.Path)
	}
	if commit.Description != "Git commit best practices" {
		t.Errorf("skill/commit description = %v, want 'Git commit best practices'", commit.Description)
	}

	// Verify agents
	agents, ok := lib.Resources["agent"]
	if !ok {
		t.Error("Library has no agents")
	}
	if _, ok := agents["reviewer"]; !ok {
		t.Error("Library missing agent/reviewer")
	}

	// Verify presets
	if len(lib.Presets) == 0 {
		t.Error("Library has no presets")
	}

	gitWorkflow, ok := lib.Presets["git-workflow"]
	if !ok {
		t.Fatal("Library missing git-workflow preset")
	}
	if len(gitWorkflow.Resources) != 2 {
		t.Errorf("git-workflow preset has %d resources, want 2", len(gitWorkflow.Resources))
	}
}

func TestLoadLibrary_MissingDirectory(t *testing.T) {
	_, err := LoadLibrary("/nonexistent/path/to/library")
	if err == nil {
		t.Error("LoadLibrary() expected error for missing directory")
	}
}

func TestLoadLibrary_MissingYAML(t *testing.T) {
	// Create temp directory without library.yaml
	tmpDir := t.TempDir()

	_, err := LoadLibrary(tmpDir)
	if err == nil {
		t.Error("LoadLibrary() expected error for missing library.yaml")
	}
}

func TestLoadLibrary_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid YAML
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte("invalid: [yaml: content"), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	_, err := LoadLibrary(tmpDir)
	if err == nil {
		t.Error("LoadLibrary() expected error for invalid YAML")
	}
}

func TestLoadLibrary_MissingVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML without version
	yamlContent := `
resources:
  skill:
    test:
      path: skills/test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	_, err := LoadLibrary(tmpDir)
	if err == nil {
		t.Error("LoadLibrary() expected error for missing version")
	}
}

func TestLoadLibrary_UnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML with unsupported version
	yamlContent := `
version: "2"
resources:
  skill:
    test:
      path: skills/test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	_, err := LoadLibrary(tmpDir)
	if err == nil {
		t.Error("LoadLibrary() expected error for unsupported version")
	}
}

func TestLoadLibrary_InvalidResourceType(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML with invalid resource type
	yamlContent := `
version: "1"
resources:
  invalid-type:
    test:
      path: test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	_, err := LoadLibrary(tmpDir)
	if err == nil {
		t.Error("LoadLibrary() expected error for invalid resource type")
	}
}

func TestLoadLibrary_EmptyLibrary(t *testing.T) {
	tmpDir := t.TempDir()

	// Write minimal valid YAML
	yamlContent := `
version: "1"
resources: {}
presets: {}
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	lib, err := LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	if len(lib.Resources) != 0 {
		t.Errorf("Empty library should have 0 resources, got %d", len(lib.Resources))
	}
	if len(lib.Presets) != 0 {
		t.Errorf("Empty library should have 0 presets, got %d", len(lib.Presets))
	}
}

func TestLoadLibrary_NotADirectory(t *testing.T) {
	// Create temp file (not directory)
	tmpFile := filepath.Join(t.TempDir(), "notadir")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadLibrary(tmpFile)
	if err == nil {
		t.Error("LoadLibrary() expected error for file path")
	}
}
