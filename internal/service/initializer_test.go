package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/internal/infrastructure/serialization"
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

	init := NewInitializer(parsing.NewParser(), serialization.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		DryRun:    true,
		Force:     false,
	})
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
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

	init := NewInitializer(parsing.NewParser(), serialization.NewSerializer())
	_, err = init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		DryRun:    false,
		Force:     false,
	})
	if err == nil {
		t.Error("Initialize() should return error when file exists without force")
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

	init := NewInitializer(parsing.NewParser(), serialization.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		DryRun:    false,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
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

	init := NewInitializer(parsing.NewParser(), serialization.NewSerializer())
	_, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/nonexistent"},
		DryRun:    false,
		Force:     false,
	})
	if err == nil {
		t.Error("Initialize() should return error for missing resource")
	}
}

func TestInitialize_WithPresetRefs(t *testing.T) {
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

	// Resolve preset refs (this would happen in command layer)
	refs, err := library.ResolvePreset(lib, "git-workflow")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}

	init := NewInitializer(parsing.NewParser(), serialization.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      refs,
		DryRun:    true,
		Force:     false,
	})
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results from git-workflow preset, got %d", len(results))
	}
}

func TestInitialize_PresetNotFound(t *testing.T) {
	lib := &library.Library{
		RootPath: t.TempDir(),
		Presets:  map[string]library.Preset{},
	}

	// Resolve preset refs (this would happen in command layer)
	_, err := library.ResolvePreset(lib, "nonexistent")
	if err == nil {
		t.Error("ResolvePreset() should return error for missing preset")
	}
}

func TestFormatDryRunOutput(t *testing.T) {
	results := []domain.InitializeResult{
		{
			Ref:        "skill/commit",
			InputPath:  "/lib/skills/commit.yaml",
			OutputPath: ".opencode/skills/commit/SKILL.md",
		},
	}

	output := formatDryRunOutput(results)

	if output == "" {
		t.Error("formatDryRunOutput() should return non-empty string")
	}
}

func TestFormatSuccessOutput(t *testing.T) {
	results := []domain.InitializeResult{
		{
			Ref:        "skill/commit",
			OutputPath: ".opencode/skills/commit/SKILL.md",
		},
	}

	output := formatSuccessOutput(results)

	if output == "" {
		t.Error("formatSuccessOutput() should return non-empty string")
	}
}

func formatDryRunOutput(results []domain.InitializeResult) string {
	output := ""
	for _, result := range results {
		output += "Would write: " + result.OutputPath + "\n  from: " + result.InputPath + "\n"
	}
	return output
}

func formatSuccessOutput(results []domain.InitializeResult) string {
	output := ""
	for _, result := range results {
		output += "Installed: " + result.Ref + " -> " + result.OutputPath + "\n"
	}
	return output
}
