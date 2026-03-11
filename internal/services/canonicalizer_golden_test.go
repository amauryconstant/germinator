package services

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/models"
)

func TestCanonicalizeGoldenFiles(t *testing.T) {
	if _, err := os.Stat("../../test/fixtures/canonical"); os.IsNotExist(err) {
		t.Skip("Golden file tests require running from project root")
	}

	fixturesDir := filepath.Join("..", "..", "test", "fixtures")
	goldenDir := filepath.Join("..", "..", "test", "golden", "canonical")

	tests := []struct {
		name     string
		fixture  string
		golden   string
		platform string
		docType  string
	}{
		{
			name:     "agent-claude-code",
			fixture:  filepath.Join(fixturesDir, "claude-code", "agent.md"),
			golden:   filepath.Join(goldenDir, "agent-claude-code.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "agent",
		},
		{
			name:     "command-claude-code",
			fixture:  filepath.Join(fixturesDir, "claude-code", "command.md"),
			golden:   filepath.Join(goldenDir, "command-claude-code.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "command",
		},
		{
			name:     "skill-claude-code",
			fixture:  filepath.Join(fixturesDir, "claude-code", "skill.md"),
			golden:   filepath.Join(goldenDir, "skill-claude-code.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "skill",
		},
		{
			name:     "memory-claude-code",
			fixture:  filepath.Join(fixturesDir, "claude-code", "memory.md"),
			golden:   filepath.Join(goldenDir, "memory-claude-code.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "memory",
		},
		{
			name:     "agent-opencode",
			fixture:  filepath.Join(fixturesDir, "opencode", "agent.md"),
			golden:   filepath.Join(goldenDir, "agent-opencode.yaml.golden"),
			platform: models.PlatformOpenCode,
			docType:  "agent",
		},
		{
			name:     "command-opencode",
			fixture:  filepath.Join(fixturesDir, "opencode", "command.md"),
			golden:   filepath.Join(goldenDir, "command-opencode.yaml.golden"),
			platform: models.PlatformOpenCode,
			docType:  "command",
		},
		{
			name:     "skill-opencode",
			fixture:  filepath.Join(fixturesDir, "opencode", "skill.md"),
			golden:   filepath.Join(goldenDir, "skill-opencode.yaml.golden"),
			platform: models.PlatformOpenCode,
			docType:  "skill",
		},
		{
			name:     "memory-opencode",
			fixture:  filepath.Join(fixturesDir, "opencode", "memory.md"),
			golden:   filepath.Join(goldenDir, "memory-opencode.yaml.golden"),
			platform: models.PlatformOpenCode,
			docType:  "memory",
		},
		{
			name:     "agent-generic",
			fixture:  filepath.Join(fixturesDir, "agent-valid.md"),
			golden:   filepath.Join(goldenDir, "agent.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "agent",
		},
		{
			name:     "command-generic",
			fixture:  filepath.Join(fixturesDir, "command-valid.md"),
			golden:   filepath.Join(goldenDir, "command.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "command",
		},
		{
			name:     "skill-generic",
			fixture:  filepath.Join(fixturesDir, "skill-valid.md"),
			golden:   filepath.Join(goldenDir, "skill.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "skill",
		},
		{
			name:     "memory-generic",
			fixture:  filepath.Join(fixturesDir, "memory-valid.md"),
			golden:   filepath.Join(goldenDir, "memory.yaml.golden"),
			platform: models.PlatformClaudeCode,
			docType:  "memory",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.yaml")

			_, err := c.Canonicalize(context.Background(), &application.CanonicalizeRequest{
				InputPath:  tt.fixture,
				OutputPath: outputPath,
				Platform:   tt.platform,
				DocType:    tt.docType,
			})
			if err != nil {
				t.Fatalf("Canonicalize() failed: %v", err)
			}

			actual, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			expected, err := os.ReadFile(tt.golden)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			if os.Getenv("UPDATE_GOLDEN") == "true" {
				if err := os.MkdirAll(filepath.Dir(tt.golden), 0755); err != nil {
					t.Fatalf("Failed to create golden directory: %v", err)
				}
				if err := os.WriteFile(tt.golden, actual, 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", tt.golden)
				return
			}

			if !bytes.Equal(actual, expected) {
				t.Errorf("Output does not match golden file")
				t.Logf("Expected:\n%s", string(expected))
				t.Logf("Actual:\n%s", string(actual))

				linesActual := bytes.Split(actual, []byte("\n"))
				linesExpected := bytes.Split(expected, []byte("\n"))
				for i := 0; i < len(linesActual) && i < len(linesExpected); i++ {
					if !bytes.Equal(linesActual[i], linesExpected[i]) {
						t.Logf("First difference at line %d", i+1)
						t.Logf("Expected: %q", string(linesExpected[i]))
						t.Logf("Actual:   %q", string(linesActual[i]))
						break
					}
				}
			}
		})
	}
}
