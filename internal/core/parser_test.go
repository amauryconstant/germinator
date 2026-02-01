package core

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
)

func TestParseMemoryWithPaths(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name          string
		filepath      string
		expectError   bool
		expectedPaths []string
	}{
		{
			name:          "parse memory with paths",
			filepath:      filepath.Join(fixturesDir, "memory-with-paths.md"),
			expectError:   false,
			expectedPaths: []string{"src/**/*.go", "internal/**/*.go"},
		},
		{
			name:          "parse memory without frontmatter",
			filepath:      filepath.Join(fixturesDir, "memory-test.md"),
			expectError:   false,
			expectedPaths: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.filepath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			doc, err := parseMemory(tt.filepath, string(content))

			if tt.expectError {
				if err == nil {
					t.Errorf("parseMemory() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseMemory() unexpected error: %v", err)
				return
			}

			memory, ok := doc.(*models.Memory)
			if !ok {
				t.Errorf("parseMemory() expected *models.Memory, got %T", doc)
				return
			}

			if len(memory.Paths) != len(tt.expectedPaths) {
				t.Errorf("parseMemory() len(memory.Paths) = %d, want %d", len(memory.Paths), len(tt.expectedPaths))
			} else {
				for i, expectedPath := range tt.expectedPaths {
					if memory.Paths[i] != expectedPath {
						t.Errorf("parseMemory() memory.Paths[%d] = %s, want %s", i, memory.Paths[i], expectedPath)
					}
				}
			}

			if len(memory.Content) == 0 {
				t.Errorf("parseMemory() memory.Content is empty, want non-empty")
			}
		})
	}
}
