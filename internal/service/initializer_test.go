package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
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

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
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

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit"},
		DryRun:    false,
		Force:     false,
	})
	// Per the slice-5 contract: error is nil even on per-resource
	// failure; per-resource outcomes live in result.Error.
	if err != nil {
		t.Fatalf("Initialize() should not return error on per-resource failure (it lives in result.Error): %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Fatal("expected result[0].Error to be non-nil when file exists without --force")
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

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
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

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/nonexistent"},
		DryRun:    false,
		Force:     false,
	})
	// Per the slice-5 contract: error is nil even on per-resource
	// failure; per-resource outcomes live in result.Error.
	if err != nil {
		t.Fatalf("Initialize() should not return error on missing resource: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Fatal("expected result[0].Error to be non-nil for missing resource")
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
	refs, err := lib.ResolvePreset(context.Background(), "git-workflow")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
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
	_, err := lib.ResolvePreset(context.Background(), "nonexistent")
	if err == nil {
		t.Error("ResolvePreset() should return error for missing preset")
	}
}

func TestFormatDryRunOutput(t *testing.T) {
	results := []core.InitializeResult{
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
	results := []core.InitializeResult{
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

// formatDryRunOutput and formatSuccessOutput are local copies for testing
// since the actual formatters are in cmd/formatters.go
func formatDryRunOutput(results []core.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		output.WriteString("Would write: ")
		output.WriteString(result.OutputPath)
		output.WriteString("\n  from: ")
		output.WriteString(result.InputPath)
		output.WriteString("\n")
	}
	return output.String()
}

func formatSuccessOutput(results []core.InitializeResult) string {
	var output strings.Builder
	for _, result := range results {
		output.WriteString("Installed: ")
		output.WriteString(result.Ref)
		output.WriteString(" -> ")
		output.WriteString(result.OutputPath)
		output.WriteString("\n")
	}
	return output.String()
}

// Partial processing tests

func TestInitialize_PartialSuccess(t *testing.T) {
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

	// Create a valid resource path for one of the refs
	validOutputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(validOutputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	// First ref exists, second ref doesn't - partial success
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit", "skill/nonexistent"},
		DryRun:    false,
		Force:     false,
	})

	// Should return nil error (at least one succeeded)
	if err != nil {
		t.Errorf("Initialize() expected nil error on partial success, got: %v", err)
	}

	// Should return 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should be successful
	if results[0].Error != nil {
		t.Errorf("First result expected no error, got: %v", results[0].Error)
	}

	// Second result should have error
	if results[1].Error == nil {
		t.Error("Second result expected error for nonexistent resource")
	}
}

func TestInitialize_AllResourcesFail(t *testing.T) {
	lib := &library.Library{
		RootPath:  t.TempDir(),
		Resources: map[string]map[string]library.Resource{},
	}

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: t.TempDir(),
		Refs:      []string{"skill/nonexistent1", "skill/nonexistent2"},
		DryRun:    false,
		Force:     false,
	})

	// Per the slice-5 contract: error is nil even when ALL resources
	// fail; the caller (runInit) inspects the slice and synthesizes
	// *core.PartialSuccessError{Succeeded: 0, Failed: N}.
	if err != nil {
		t.Errorf("Initialize() should not return error on per-resource failure: %v", err)
	}

	// Should still return 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// All results should have errors
	for i, r := range results {
		if r.Error == nil {
			t.Errorf("Result %d expected error, got nil", i)
		}
	}
}

func TestInitialize_AllResultsReturnedRegardlessOfErrors(t *testing.T) {
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

	// Create a file exists scenario for one resource
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit", "skill/nonexistent", "skill/merge-request"},
		DryRun:    false,
		Force:     false,
	})

	// Should return nil error (at least one succeeded - merge-request)
	if err != nil {
		t.Errorf("Initialize() expected nil error on partial success, got: %v", err)
	}

	// Should return 3 results - all resources processed
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// All 3 refs should be represented
	refs := make(map[string]bool)
	for _, r := range results {
		refs[r.Ref] = true
	}
	if !refs["skill/commit"] || !refs["skill/nonexistent"] || !refs["skill/merge-request"] {
		t.Error("Expected results for all 3 refs")
	}
}

func TestInitialize_ContinuesAfterFileExistsError(t *testing.T) {
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

	// Create the first output file to simulate existing file
	outputPath1 := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(outputPath1), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outputPath1, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	init := NewInitializer(parser.NewParser(), renderer.NewSerializer())
	results, err := init.Initialize(context.Background(), &application.InitializeRequest{
		Library:   lib,
		Platform:  "opencode",
		OutputDir: outputDir,
		Refs:      []string{"skill/commit", "skill/merge-request"},
		DryRun:    false,
		Force:     false,
	})

	// Should return nil error (merge-request succeeded)
	if err != nil {
		t.Errorf("Initialize() expected nil error on partial success, got: %v", err)
	}

	// Should return 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should have file exists error
	if results[0].Error == nil {
		t.Error("First result expected file exists error")
	}

	// Second result should be successful
	if results[1].Error != nil {
		t.Errorf("Second result expected no error, got: %v", results[1].Error)
	}
}
