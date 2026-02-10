package services

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
)

func TestCanonicalizeDocument(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name        string
		platform    string
		docType     string
		inputFile   string
		expectError bool
	}{
		{
			name:      "claude-code agent",
			platform:  models.PlatformClaudeCode,
			docType:   "agent",
			inputFile: filepath.Join(fixturesDir, "claude-code", "agent.md"),
		},
		{
			name:      "claude-code command",
			platform:  models.PlatformClaudeCode,
			docType:   "command",
			inputFile: filepath.Join(fixturesDir, "claude-code", "command.md"),
		},
		{
			name:      "claude-code skill",
			platform:  models.PlatformClaudeCode,
			docType:   "skill",
			inputFile: filepath.Join(fixturesDir, "claude-code", "skill.md"),
		},
		{
			name:      "claude-code memory",
			platform:  models.PlatformClaudeCode,
			docType:   "memory",
			inputFile: filepath.Join(fixturesDir, "claude-code", "memory.md"),
		},
		{
			name:      "opencode agent",
			platform:  models.PlatformOpenCode,
			docType:   "agent",
			inputFile: filepath.Join(fixturesDir, "opencode", "agent.md"),
		},
		{
			name:      "opencode command",
			platform:  models.PlatformOpenCode,
			docType:   "command",
			inputFile: filepath.Join(fixturesDir, "opencode", "command.md"),
		},
		{
			name:      "opencode skill",
			platform:  models.PlatformOpenCode,
			docType:   "skill",
			inputFile: filepath.Join(fixturesDir, "opencode", "skill.md"),
		},
		{
			name:      "opencode memory",
			platform:  models.PlatformOpenCode,
			docType:   "memory",
			inputFile: filepath.Join(fixturesDir, "opencode", "memory.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.yaml")

			err := CanonicalizeDocument(tt.inputFile, outputPath, tt.platform, tt.docType)

			if tt.expectError {
				if err == nil {
					t.Error("CanonicalizeDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CanonicalizeDocument() unexpected error: %v", err)
			}

			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("CanonicalizeDocument() did not create output file")
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			if len(content) == 0 {
				t.Error("CanonicalizeDocument() output file is empty")
			}
		})
	}
}

func TestCanonicalizeDocumentParseError(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.md")
	invalidContent := `---
name: test
invalid yaml [unclosed
---
content`

	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output.yaml")

	err := CanonicalizeDocument(invalidFile, outputPath, "claude-code", "agent")
	if err == nil {
		t.Error("CanonicalizeDocument() expected error for invalid YAML, got nil")
	}

	if _, statErr := os.Stat(outputPath); statErr == nil {
		t.Error("CanonicalizeDocument() should not create output file on parse error")
	}
}

func TestCanonicalizeDocumentValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.md")
	invalidContent := `---
description: Agent without name
---
content`

	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output.yaml")

	err := CanonicalizeDocument(invalidFile, outputPath, "claude-code", "agent")
	if err == nil {
		t.Error("CanonicalizeDocument() expected error for missing name, got nil")
	}

	if _, statErr := os.Stat(outputPath); statErr == nil {
		t.Error("CanonicalizeDocument() should not create output file on validation error")
	}
}

func TestCanonicalizeDocumentFileWriteError(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	inputFile := filepath.Join(fixturesDir, "claude-code", "agent.md")

	nonExistentDir := "/non/existent/directory/output.yaml"

	err := CanonicalizeDocument(inputFile, nonExistentDir, "claude-code", "agent")
	if err == nil {
		t.Error("CanonicalizeDocument() expected error for unwritable path, got nil")
	}
}

func TestCanonicalizeDocumentRoundTrip(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	inputFile := filepath.Join(fixturesDir, "claude-code", "agent.md")

	tmpDir := t.TempDir()
	canonicalOutput := filepath.Join(tmpDir, "agent-test.yaml")

	err := CanonicalizeDocument(inputFile, canonicalOutput, "claude-code", "agent")
	if err != nil {
		t.Fatalf("CanonicalizeDocument() failed: %v", err)
	}

	platformOutput := filepath.Join(tmpDir, "platform.md")

	err = TransformDocument(canonicalOutput, platformOutput, "claude-code")
	if err != nil {
		t.Fatalf("TransformDocument() failed: %v", err)
	}

	platformContent, err := os.ReadFile(platformOutput)
	if err != nil {
		t.Fatalf("failed to read platform output: %v", err)
	}

	if len(platformContent) == 0 {
		t.Error("Round-trip produced empty platform output")
	}
}

func getFixturesDir(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return filepath.Join(filepath.Join(cwd, "..", "..", "test", "fixtures"))
}
