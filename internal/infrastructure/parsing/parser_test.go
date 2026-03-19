package parsing

import (
	"os"
	"path/filepath"
	"testing"
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
			filepath:      filepath.Join(fixturesDir, "memory-valid.md"),
			expectError:   false,
			expectedPaths: []string{"src/**/*.go", "cmd/**/*.go", "pkg/**/*.go", "go.mod", "go.sum"},
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

			memory, ok := doc.(*CanonicalMemory)
			if !ok {
				t.Errorf("parseMemory() expected *CanonicalMemory, got %T", doc)
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

func TestParseAgentWithFrontmatter(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name           string
		filepath       string
		expectError    bool
		expectedName   string
		expectedPolicy string
	}{
		{
			name:           "parse valid agent",
			filepath:       filepath.Join(fixturesDir, "agent-valid.md"),
			expectError:    false,
			expectedName:   "code-reviewer",
			expectedPolicy: "restrictive",
		},
		{
			name:        "parse invalid agent",
			filepath:    filepath.Join(fixturesDir, "agent-invalid.md"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.filepath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			doc, err := parseDocumentWithFrontmatter(tt.filepath, string(content), "agent")

			if tt.expectError {
				if err == nil {
					t.Errorf("parseDocumentWithFrontmatter() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseDocumentWithFrontmatter() unexpected error: %v", err)
				return
			}

			agent, ok := doc.(*CanonicalAgent)
			if !ok {
				t.Errorf("parseDocumentWithFrontmatter() expected *CanonicalAgent, got %T", doc)
				return
			}

			if tt.expectedName != "" && agent.Name != tt.expectedName {
				t.Errorf("parseDocumentWithFrontmatter() agent.Name = %s, want %s", agent.Name, tt.expectedName)
			}

			if tt.expectedPolicy != "" && string(agent.PermissionPolicy) != tt.expectedPolicy {
				t.Errorf("parseDocumentWithFrontmatter() agent.PermissionPolicy = %s, want %s", agent.PermissionPolicy, tt.expectedPolicy)
			}

			if len(agent.Content) == 0 {
				t.Errorf("parseDocumentWithFrontmatter() agent.Content is empty, want non-empty")
			}
		})
	}
}

func TestParseSkillWithFrontmatter(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name        string
		filepath    string
		expectError bool
	}{
		{
			name:        "parse valid skill",
			filepath:    filepath.Join(fixturesDir, "skill-valid.md"),
			expectError: false,
		},
		{
			name:        "parse invalid skill",
			filepath:    filepath.Join(fixturesDir, "skill-invalid.md"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.filepath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			doc, err := parseDocumentWithFrontmatter(tt.filepath, string(content), "skill")

			if tt.expectError {
				if err == nil {
					t.Errorf("parseDocumentWithFrontmatter() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseDocumentWithFrontmatter() unexpected error: %v", err)
				return
			}

			skill, ok := doc.(*CanonicalSkill)
			if !ok {
				t.Errorf("parseDocumentWithFrontmatter() expected *CanonicalSkill, got %T", doc)
				return
			}

			if len(skill.Content) == 0 {
				t.Errorf("parseDocumentWithFrontmatter() skill.Content is empty, want non-empty")
			}
		})
	}
}

func TestParseCommandWithFrontmatter(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name        string
		filepath    string
		expectError bool
	}{
		{
			name:        "parse valid command",
			filepath:    filepath.Join(fixturesDir, "command-valid.md"),
			expectError: false,
		},
		{
			name:        "parse invalid command",
			filepath:    filepath.Join(fixturesDir, "command-invalid.md"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.filepath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			doc, err := parseDocumentWithFrontmatter(tt.filepath, string(content), "command")

			if tt.expectError {
				if err == nil {
					t.Errorf("parseDocumentWithFrontmatter() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseDocumentWithFrontmatter() unexpected error: %v", err)
				return
			}

			command, ok := doc.(*CanonicalCommand)
			if !ok {
				t.Errorf("parseDocumentWithFrontmatter() expected *CanonicalCommand, got %T", doc)
				return
			}

			if len(command.Content) == 0 {
				t.Errorf("parseDocumentWithFrontmatter() command.Content is empty, want non-empty")
			}
		})
	}
}
