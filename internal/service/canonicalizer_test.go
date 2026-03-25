package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/internal/infrastructure/serialization"
	"gitlab.com/amoconst/germinator/internal/models"
)

func TestCanonicalizeDocument(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	c := NewCanonicalizer()

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

			_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
				InputPath:  tt.inputFile,
				OutputPath: outputPath,
				Platform:   tt.platform,
				DocType:    tt.docType,
			})

			if tt.expectError {
				if err == nil {
					t.Error("Canonicalize() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Canonicalize() unexpected error: %v", err)
			}

			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("Canonicalize() did not create output file")
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			if len(content) == 0 {
				t.Error("Canonicalize() output file is empty")
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
	c := NewCanonicalizer()

	_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  invalidFile,
		OutputPath: outputPath,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err == nil {
		t.Error("Canonicalize() expected error for invalid YAML, got nil")
	}

	if _, statErr := os.Stat(outputPath); statErr == nil {
		t.Error("Canonicalize() should not create output file on parse error")
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
	c := NewCanonicalizer()

	_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  invalidFile,
		OutputPath: outputPath,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err == nil {
		t.Error("Canonicalize() expected error for missing name, got nil")
	}

	if _, statErr := os.Stat(outputPath); statErr == nil {
		t.Error("Canonicalize() should not create output file on validation error")
	}
}

func TestCanonicalizeDocumentFileWriteError(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	inputFile := filepath.Join(fixturesDir, "claude-code", "agent.md")

	nonExistentDir := "/non/existent/directory/output.yaml"
	c := NewCanonicalizer()

	_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: nonExistentDir,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err == nil {
		t.Error("Canonicalize() expected error for unwritable path, got nil")
	}
}

func TestCanonicalizeDocumentRoundTrip(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	inputFile := filepath.Join(fixturesDir, "claude-code", "agent.md")

	tmpDir := t.TempDir()
	canonicalOutput := filepath.Join(tmpDir, "agent-test.yaml")
	c := NewCanonicalizer()

	_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: canonicalOutput,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err != nil {
		t.Fatalf("Canonicalize() failed: %v", err)
	}

	platformOutput := filepath.Join(tmpDir, "platform.md")
	t2 := NewTransformer(parsing.NewParser(), serialization.NewSerializer())

	_, err = t2.Transform(context.Background(), &application.TransformRequest{
		InputPath:  canonicalOutput,
		OutputPath: platformOutput,
		Platform:   "claude-code",
	})
	if err != nil {
		t.Fatalf("Transform() failed: %v", err)
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
