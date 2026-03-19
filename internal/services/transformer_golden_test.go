//go:build golden

package services

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
)

// To update golden files: go test ./internal/services -update-golden
// To run specific test: go test ./internal/services -run TestGoldenFiles/TestAgentFull

func TestGoldenFiles(t *testing.T) {
	// Ensure tests run from project root for correct fixture paths
	if _, err := os.Stat("../../test/fixtures/canonical"); os.IsNotExist(err) {
		t.Skip("Golden file tests require running from project root")
	}

	tests := []struct {
		name     string
		fixture  string // Canonical format fixture
		golden   string // Golden file path
		platform string // Platform to test
	}{
		// Agent tests - Canonical permission policies
		{
			name:     "agent-permission-restrictive",
			fixture:  "../../test/fixtures/canonical/agent-permission-restrictive.md",
			golden:   "../../test/golden/opencode/agent-permission-restrictive.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-permission-balanced",
			fixture:  "../../test/fixtures/canonical/agent-permission-balanced.md",
			golden:   "../../test/golden/opencode/agent-permission-balanced.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-permission-permissive",
			fixture:  "../../test/fixtures/canonical/agent-permission-permissive.md",
			golden:   "../../test/golden/opencode/agent-permission-permissive.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-permission-analysis",
			fixture:  "../../test/fixtures/canonical/agent-permission-analysis.md",
			golden:   "../../test/golden/opencode/agent-permission-analysis.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-permission-unrestricted",
			fixture:  "../../test/fixtures/canonical/agent-permission-unrestricted.md",
			golden:   "../../test/golden/opencode/agent-permission-unrestricted.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-with-targets-claude-code",
			fixture:  "../../test/fixtures/canonical/agent-with-targets-claude-code.md",
			golden:   "../../test/golden/opencode/agent-with-targets-claude-code.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-with-targets-opencode",
			fixture:  "../../test/fixtures/canonical/agent-with-targets-opencode.md",
			golden:   "../../test/golden/opencode/agent-with-targets-opencode.md.golden",
			platform: "opencode",
		},

		// Command tests
		{
			name:     "command-full",
			fixture:  "../../test/fixtures/opencode/command-full.md",
			golden:   "../../test/golden/opencode/command-full.md.golden",
			platform: "opencode",
		},
		{
			name:     "command-with-arguments",
			fixture:  "../../test/fixtures/opencode/command-with-arguments.md",
			golden:   "../../test/golden/opencode/command-with-arguments.md.golden",
			platform: "opencode",
		},
		{
			name:     "run-tests-command",
			fixture:  "../../test/fixtures/opencode/run-tests-command.md",
			golden:   "../../test/golden/opencode/run-tests-command.md.golden",
			platform: "opencode",
		},

		// Skill tests
		{
			name:     "skill-full",
			fixture:  "../../test/fixtures/opencode/skill-full.md",
			golden:   "../../test/golden/opencode/skill-full.md.golden",
			platform: "opencode",
		},
		{
			name:     "git-workflow-skill",
			fixture:  "../../test/fixtures/opencode/git-workflow-skill/git-workflow-skill.md",
			golden:   "../../test/golden/opencode/git-workflow-skill.md.golden",
			platform: "opencode",
		},

		// Memory tests
		{
			name:     "memory-paths-only",
			fixture:  "../../test/fixtures/opencode/memory-paths-only.md",
			golden:   "../../test/golden/opencode/memory-paths-only.md.golden",
			platform: "opencode",
		},
		{
			name:     "memory-content-only",
			fixture:  "../../test/fixtures/opencode/memory-content-only.md",
			golden:   "../../test/golden/opencode/memory-content-only.md.golden",
			platform: "opencode",
		},
		{
			name:     "memory-both",
			fixture:  "../../test/fixtures/opencode/memory-both.md",
			golden:   "../../test/golden/opencode/memory-both.md.golden",
			platform: "opencode",
		},

		// Agent tests - basic fixtures
		{
			name:     "agent-full",
			fixture:  "../../test/fixtures/opencode/agent-full.md",
			golden:   "../../test/golden/opencode/agent-full.md.golden",
			platform: "opencode",
		},
		{
			name:     "agent-mixed-tools",
			fixture:  "../../test/fixtures/opencode/agent-mixed-tools.md",
			golden:   "../../test/golden/opencode/agent-mixed-tools.md.golden",
			platform: "opencode",
		},
		{
			name:     "code-reviewer-agent",
			fixture:  "../../test/fixtures/opencode/code-reviewer-agent.md",
			golden:   "../../test/golden/opencode/code-reviewer-agent.md.golden",
			platform: "opencode",
		},

		// Claude-Code platform tests
		{
			name:     "claude-code-agent-full",
			fixture:  "../../test/fixtures/claude-code/agent-full.md",
			golden:   "../../test/golden/claude-code/agent-full.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-agent-permission-restrictive",
			fixture:  "../../test/fixtures/claude-code/agent-permission-restrictive.md",
			golden:   "../../test/golden/claude-code/agent-permission-restrictive.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-agent-permission-balanced",
			fixture:  "../../test/fixtures/claude-code/agent-permission-balanced.md",
			golden:   "../../test/golden/claude-code/agent-permission-balanced.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-agent-permission-permissive",
			fixture:  "../../test/fixtures/claude-code/agent-permission-permissive.md",
			golden:   "../../test/golden/claude-code/agent-permission-permissive.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-agent-permission-analysis",
			fixture:  "../../test/fixtures/claude-code/agent-permission-analysis.md",
			golden:   "../../test/golden/claude-code/agent-permission-analysis.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-agent-permission-unrestricted",
			fixture:  "../../test/fixtures/claude-code/agent-permission-unrestricted.md",
			golden:   "../../test/golden/claude-code/agent-permission-unrestricted.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-command-full",
			fixture:  "../../test/fixtures/claude-code/command-full.md",
			golden:   "../../test/golden/claude-code/command-full.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-command-with-arguments",
			fixture:  "../../test/fixtures/claude-code/command-with-arguments.md",
			golden:   "../../test/golden/claude-code/command-with-arguments.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-skill-full",
			fixture:  "../../test/fixtures/claude-code/skill-full.md",
			golden:   "../../test/golden/claude-code/skill-full.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-memory-paths-only",
			fixture:  "../../test/fixtures/claude-code/memory-paths-only.md",
			golden:   "../../test/golden/claude-code/memory-paths-only.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-memory-content-only",
			fixture:  "../../test/fixtures/claude-code/memory-content-only.md",
			golden:   "../../test/golden/claude-code/memory-content-only.md.golden",
			platform: "claude-code",
		},
		{
			name:     "claude-code-memory-both",
			fixture:  "../../test/fixtures/claude-code/memory-both.md",
			golden:   "../../test/golden/claude-code/memory-both.md.golden",
			platform: "claude-code",
		},
	}

	tr := NewTransformer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output file
			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "output.md")

			// Transform document
			_, err := tr.Transform(context.Background(), &application.TransformRequest{
				InputPath:  tt.fixture,
				OutputPath: outputFile,
				Platform:   tt.platform,
			})
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Read actual output
			actual, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			// Read golden file
			expected, err := os.ReadFile(tt.golden)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			// Check if UPDATE_GOLDEN flag is set
			if os.Getenv("UPDATE_GOLDEN") == "true" {
				// Create directory if it doesn't exist
				if err := os.MkdirAll(filepath.Dir(tt.golden), 0755); err != nil {
					t.Fatalf("Failed to create golden directory: %v", err)
				}
				// Write actual output to golden file
				if err := os.WriteFile(tt.golden, actual, 0644); err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", tt.golden)
				return
			}

			// Compare actual with expected (byte-by-byte comparison)
			if !bytes.Equal(actual, expected) {
				t.Errorf("Output does not match golden file")
				t.Logf("Expected:\n%s", string(expected))
				t.Logf("Actual:\n%s", string(actual))
				t.Logf("Diff length: expected=%d, actual=%d", len(expected), len(actual))

				// Show first differing line
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

// TestGoldenFileUpdate tests golden file update mechanism
func TestGoldenFileUpdate(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "true" {
		t.Skip("Set UPDATE_GOLDEN=true to test update mechanism")
	}

	// This test just verifies mechanism works by running a subset of golden tests
	tests := []struct {
		name     string
		fixture  string
		golden   string
		platform string
	}{
		{
			name:     "agent-full-update",
			fixture:  "../../test/fixtures/opencode/agent-full.md",
			golden:   "../../test/golden/opencode/agent-full.md.golden",
			platform: "opencode",
		},
	}

	tr := NewTransformer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "output.md")

			_, err := tr.Transform(context.Background(), &application.TransformRequest{
				InputPath:  tt.fixture,
				OutputPath: outputFile,
				Platform:   tt.platform,
			})
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			actual, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			// Ensure golden directory exists
			if err := os.MkdirAll(filepath.Dir(tt.golden), 0755); err != nil {
				t.Fatalf("Failed to create golden directory: %v", err)
			}

			// Update golden file
			if err := os.WriteFile(tt.golden, actual, 0644); err != nil {
				t.Fatalf("Failed to update golden file: %v", err)
			}

			t.Logf("Successfully updated golden file: %s", tt.golden)
		})
	}
}
