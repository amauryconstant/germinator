package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// To update golden files: go test ./internal/services -update-golden
// To run specific test: go test ./internal/services -run TestGoldenFiles/TestAgentFull

func TestGoldenFiles(t *testing.T) {
	// Ensure tests run from project root for correct fixture paths
	if _, err := os.Stat("../../test/fixtures/opencode"); os.IsNotExist(err) {
		t.Skip("Golden file tests require running from project root")
	}

	tests := []struct {
		name     string
		fixture  string // Germinator format fixture
		golden   string // Golden file path
		platform string // Platform to test
	}{
		// Agent tests
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
			golden:   "../../test/golden/opencode/git-workflow-skill/git-workflow-skill.md.golden",
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

		// Permission mode tests
		{
			name:     "permission-default",
			fixture:  "../../test/fixtures/opencode/agent-permission-default.md",
			golden:   "../../test/golden/opencode/permission-default.md.golden",
			platform: "opencode",
		},
		{
			name:     "permission-acceptedits",
			fixture:  "../../test/fixtures/opencode/agent-permission-acceptedits.md",
			golden:   "../../test/golden/opencode/permission-acceptedits.md.golden",
			platform: "opencode",
		},
		{
			name:     "permission-dontask",
			fixture:  "../../test/fixtures/opencode/agent-permission-dontask.md",
			golden:   "../../test/golden/opencode/permission-dontask.md.golden",
			platform: "opencode",
		},
		{
			name:     "permission-bypasspermissions",
			fixture:  "../../test/fixtures/opencode/agent-permission-bypasspermissions.md",
			golden:   "../../test/golden/opencode/permission-bypasspermissions.md.golden",
			platform: "opencode",
		},
		{
			name:     "permission-plan",
			fixture:  "../../test/fixtures/opencode/agent-permission-plan.md",
			golden:   "../../test/golden/opencode/permission-plan.md.golden",
			platform: "opencode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output file
			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "output.md")

			// Transform document
			err := TransformDocument(tt.fixture, outputFile, tt.platform)
			if err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
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

// TestGoldenFileUpdate tests the golden file update mechanism
func TestGoldenFileUpdate(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "true" {
		t.Skip("Set UPDATE_GOLDEN=true to test update mechanism")
	}

	// This test just verifies the mechanism works by running a subset of golden tests
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "output.md")

			err := TransformDocument(tt.fixture, outputFile, tt.platform)
			if err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
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
