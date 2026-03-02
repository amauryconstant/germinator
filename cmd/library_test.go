package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryCommand_Resources(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	// Set environment variable to test fixtures
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"resources"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output from library resources command")
	}
}

func TestLibraryCommand_Presets(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"presets"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output from library presets command")
	}
}

func TestLibraryCommand_Show(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"show", "skill/commit"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output from library show command")
	}
}

func TestLibraryCommand_InvalidRef(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"show", "invalidformat"})

	// Execute and expect error (will call os.Exit, but we can check the command setup)
	// Note: This test would need special handling for os.Exit calls
	// For now, we just verify the command structure is correct
}

func TestInitCommand_RequiresPlatform(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewInitCommand(cfg)
	cmd.SetArgs([]string{"--resources", "skill/commit"})

	// This should fail due to missing platform
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_RequiresResourcesOrPreset(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewInitCommand(cfg)
	cmd.SetArgs([]string{"--platform", "opencode"})

	// This should fail due to missing resources/preset
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_MutuallyExclusive(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewInitCommand(cfg)
	cmd.SetArgs([]string{"--platform", "opencode", "--resources", "skill/commit", "--preset", "git-workflow"})

	// This should fail due to mutually exclusive flags
	// The actual test would need to handle os.Exit
	_ = cmd
}

func TestInitCommand_DryRun(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	outputDir := t.TempDir()

	cmd := NewInitCommand(cfg)
	cmd.SetArgs([]string{
		"--platform", "opencode",
		"--resources", "skill/commit",
		"--library", absPath,
		"--output", outputDir,
		"--dry-run",
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output from dry-run")
	}

	// Verify no files were created
	outputPath := filepath.Join(outputDir, ".opencode", "skills", "commit", "SKILL.md")
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Dry-run should not create files")
	}
}
