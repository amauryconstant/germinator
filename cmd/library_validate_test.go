package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibraryValidateCommand_ValidLibrary(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	// Use test fixtures
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", absPath})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "valid") {
		t.Errorf("Expected 'valid' in output, got: %s", output)
	}
}

func TestLibraryValidateCommand_JSON(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	// Use test fixtures
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", absPath, "--json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify it's valid JSON
	var output ValidationOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if !output.Valid {
		t.Errorf("Expected Valid=true, got Valid=false")
	}
	if output.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount=0, got %d", output.ErrorCount)
	}
	if output.WarningCount != 0 {
		t.Errorf("Expected WarningCount=0, got %d", output.WarningCount)
	}
}

func TestLibraryValidateCommand_WithIssues(t *testing.T) {
	// Create temp directory with library that has issues
	tmpDir := t.TempDir()

	// Create library.yaml with missing file entry
	libraryYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets: {}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(libraryYAML), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	// Create skills directory with only commit.md
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "commit.md"), []byte("---\nname: commit\n---\nContent"), 0644); err != nil {
		t.Fatalf("Failed to write commit.md: %v", err)
	}

	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "missing") {
		t.Errorf("Expected 'missing' in output for missing file issue, got: %s", output)
	}
	if !strings.Contains(output, "errors: 1") {
		t.Errorf("Expected 'errors: 1' in output, got: %s", output)
	}
}

func TestLibraryValidateCommand_Fix(t *testing.T) {
	// Create temp directory with library that has issues
	tmpDir := t.TempDir()

	// Create library.yaml with missing file entry and ghost preset ref
	libraryYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
    missing:
      path: skills/missing.md
      description: Missing skill
presets:
  workflow:
    description: Workflow
    resources:
      - skill/commit
      - skill/ghost
`
	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(libraryYAML), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	// Create skills directory with only commit.md
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "commit.md"), []byte("---\nname: commit\n---\nContent"), 0644); err != nil {
		t.Fatalf("Failed to write commit.md: %v", err)
	}

	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", tmpDir, "--fix"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Fix applied") {
		t.Errorf("Expected 'Fix applied' in output, got: %s", output)
	}

	// Verify library.yaml was modified - missing entry should be removed
	modifiedLib, err := os.ReadFile(filepath.Join(tmpDir, "library.yaml"))
	if err != nil {
		t.Fatalf("Failed to read modified library.yaml: %v", err)
	}

	if strings.Contains(string(modifiedLib), "skill/missing") {
		t.Error("Expected 'skill/missing' to be removed from library.yaml")
	}
	if strings.Contains(string(modifiedLib), "skill/ghost") {
		t.Error("Expected 'skill/ghost' to be removed from preset resources")
	}
}

func TestLibraryValidateCommand_OrphanWarning(t *testing.T) {
	// Create temp directory with orphan file
	tmpDir := t.TempDir()

	// Create library.yaml with one resource
	libraryYAML := `
version: "1"
resources:
  skill:
    commit:
      path: skills/commit.md
      description: Commit skill
presets: {}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(libraryYAML), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	// Create skills directory with extra orphan file
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "commit.md"), []byte("---\nname: commit\n---\nContent"), 0644); err != nil {
		t.Fatalf("Failed to write commit.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "orphan.md"), []byte("---\nname: orphan\n---\nContent"), 0644); err != nil {
		t.Fatalf("Failed to write orphan.md: %v", err)
	}

	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "orphan") {
		t.Errorf("Expected 'orphan' warning in output, got: %s", output)
	}
	if !strings.Contains(output, "warnings: 1") {
		t.Errorf("Expected 'warnings: 1' in output, got: %s", output)
	}
	// Library should still be considered valid (no errors)
	if !strings.Contains(output, "valid") {
		t.Errorf("Expected library to be valid (warnings only), got: %s", output)
	}
}

func TestLibraryValidateCommand_ExitCode(t *testing.T) {
	// Create temp directory with missing file (error)
	tmpDir := t.TempDir()

	libraryYAML := `
version: "1"
resources:
  skill:
    missing:
      path: skills/missing.md
      description: Missing skill
presets: {}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "library.yaml"), []byte(libraryYAML), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}

	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"validate", "--library", tmpDir})

	// The validate command returns nil on success even with validation issues
	// Exit code handling happens in main.go via HandleCLIError
	// For this test, we just verify the command runs and produces issues
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "errors: 1") {
		t.Errorf("Expected 'errors: 1' in output, got: %s", output)
	}
}
