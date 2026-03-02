package services

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/library"
)

func TestInitializeResources_DryRun(t *testing.T) {
	// Load test library
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	lib, err := library.LoadLibrary(absPath)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	// Create temp output directory
	outputDir := t.TempDir()

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		DryRun:    true,
		Force:     false,
	}

	refs := []string{"skill/commit"}
	results, err := InitializeResources(opts, refs)
	if err != nil {
		t.Fatalf("InitializeResources() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Verify no files were written in dry-run mode
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Dry-run should not write files")
	}
}

func TestInitializeResources_FileExists(t *testing.T) {
	// Load test library
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	lib, err := library.LoadLibrary(absPath)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	outputDir := t.TempDir()

	// Create the output file to simulate existing file
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		DryRun:    false,
		Force:     false,
	}

	refs := []string{"skill/commit"}
	_, err = InitializeResources(opts, refs)
	if err == nil {
		t.Error("InitializeResources() should return error when file exists without force")
	}
}

func TestInitializeResources_ForceOverwrite(t *testing.T) {
	// Load test library
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	lib, err := library.LoadLibrary(absPath)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	outputDir := t.TempDir()

	// Create the output file to simulate existing file
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		DryRun:    false,
		Force:     true,
	}

	refs := []string{"skill/commit"}
	results, err := InitializeResources(opts, refs)
	if err != nil {
		t.Fatalf("InitializeResources() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Verify file was overwritten
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) == "existing" {
		t.Error("File should have been overwritten")
	}
}

func TestInitializeResources_ResourceNotFound(t *testing.T) {
	lib := &library.Library{
		RootPath:  t.TempDir(),
		Resources: map[string]map[string]library.Resource{},
	}

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: t.TempDir(),
		DryRun:    false,
		Force:     false,
	}

	refs := []string{"skill/nonexistent"}
	_, err := InitializeResources(opts, refs)
	if err == nil {
		t.Error("InitializeResources() should return error for missing resource")
	}
}

func TestInitializeFromPreset(t *testing.T) {
	// Load test library
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	lib, err := library.LoadLibrary(absPath)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	outputDir := t.TempDir()

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		DryRun:    true,
		Force:     false,
	}

	results, err := InitializeFromPreset(opts, "git-workflow")
	if err != nil {
		t.Fatalf("InitializeFromPreset() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results from git-workflow preset, got %d", len(results))
	}
}

func TestInitializeFromPreset_PresetNotFound(t *testing.T) {
	lib := &library.Library{
		RootPath: t.TempDir(),
		Presets:  map[string]library.Preset{},
	}

	opts := InitOptions{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: t.TempDir(),
	}

	_, err := InitializeFromPreset(opts, "nonexistent")
	if err == nil {
		t.Error("InitializeFromPreset() should return error for missing preset")
	}
}

func TestFormatDryRunOutput(t *testing.T) {
	results := []InitResult{
		{
			Ref:        "skill/commit",
			InputPath:  "/lib/skills/commit.yaml",
			OutputPath: ".opencode/skills/commit/SKILL.md",
		},
	}

	output := FormatDryRunOutput(results)

	if output == "" {
		t.Error("FormatDryRunOutput() should return non-empty string")
	}
}

func TestFormatSuccessOutput(t *testing.T) {
	results := []InitResult{
		{
			Ref:        "skill/commit",
			OutputPath: ".opencode/skills/commit/SKILL.md",
		},
	}

	output := FormatSuccessOutput(results)

	if output == "" {
		t.Error("FormatSuccessOutput() should return non-empty string")
	}
}
