package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
)

func TestLoadDocumentIntegration(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	fixturesDir := filepath.Join(wd, "..", "..", "test", "fixtures")

	tests := []struct {
		name         string
		filepath     string
		expectError  bool
		expectedType interface{}
		checkContent bool
	}{
		{
			name:         "load valid agent",
			filepath:     filepath.Join(fixturesDir, "agent-test.md"),
			expectError:  false,
			expectedType: &models.Agent{},
			checkContent: true,
		},
		{
			name:         "load valid command",
			filepath:     filepath.Join(fixturesDir, "command-test.md"),
			expectError:  false,
			expectedType: &models.Command{},
			checkContent: true,
		},
		{
			name:         "load valid memory",
			filepath:     filepath.Join(fixturesDir, "memory-test.md"),
			expectError:  false,
			expectedType: &models.Memory{},
			checkContent: true,
		},
		{
			name:         "load valid skill",
			filepath:     filepath.Join(fixturesDir, "skill-test.md"),
			expectError:  false,
			expectedType: &models.Skill{},
			checkContent: true,
		},
		{
			name:        "load invalid agent",
			filepath:    filepath.Join(fixturesDir, "agent-invalid.md"),
			expectError: true,
		},
		{
			name:        "load invalid command",
			filepath:    filepath.Join(fixturesDir, "command-invalid.md"),
			expectError: true,
		},
		{
			name:        "load unrecognizable filename",
			filepath:    filepath.Join(fixturesDir, "my-document.md"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := LoadDocument(tt.filepath)

			if tt.expectError {
				if err == nil {
					t.Errorf("LoadDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadDocument() unexpected error: %v", err)
				return
			}

			switch tt.expectedType.(type) {
			case *models.Agent:
				agent, ok := doc.(*models.Agent)
				if !ok {
					t.Errorf("LoadDocument() expected *models.Agent, got %T", doc)
					return
				}
				if agent.Name != "code-reviewer" {
					t.Errorf("LoadDocument() agent.Name = %v, want code-reviewer", agent.Name)
				}
				if agent.FilePath != tt.filepath {
					t.Errorf("LoadDocument() agent.FilePath = %v, want %v", agent.FilePath, tt.filepath)
				}
				if tt.checkContent && !strings.Contains(agent.Content, "test agent document") {
					t.Errorf("LoadDocument() agent.Content missing expected content")
				}

			case *models.Command:
				command, ok := doc.(*models.Command)
				if !ok {
					t.Errorf("LoadDocument() expected *models.Command, got %T", doc)
					return
				}
				if command.Name != "command-test" {
					t.Errorf("LoadDocument() command.Name = %v, want command-test", command.Name)
				}
				if command.FilePath != tt.filepath {
					t.Errorf("LoadDocument() command.FilePath = %v, want %v", command.FilePath, tt.filepath)
				}
				if tt.checkContent && !strings.Contains(command.Content, "test command document") {
					t.Errorf("LoadDocument() command.Content missing expected content")
				}

			case *models.Memory:
				memory, ok := doc.(*models.Memory)
				if !ok {
					t.Errorf("LoadDocument() expected *models.Memory, got %T", doc)
					return
				}
				if memory.FilePath != tt.filepath {
					t.Errorf("LoadDocument() memory.FilePath = %v, want %v", memory.FilePath, tt.filepath)
				}
				if tt.checkContent && !strings.Contains(memory.Content, "test memory document") {
					t.Errorf("LoadDocument() memory.Content missing expected content")
				}

			case *models.Skill:
				skill, ok := doc.(*models.Skill)
				if !ok {
					t.Errorf("LoadDocument() expected *models.Skill, got %T", doc)
					return
				}
				if skill.Name != "explaining-code" {
					t.Errorf("LoadDocument() skill.Name = %v, want explaining-code", skill.Name)
				}
				if skill.FilePath != tt.filepath {
					t.Errorf("LoadDocument() skill.FilePath = %v, want %v", skill.FilePath, tt.filepath)
				}
				if tt.checkContent && !strings.Contains(skill.Content, "test skill document") {
					t.Errorf("LoadDocument() skill.Content missing expected content")
				}
			}
		})
	}
}

func TestParseDocument(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	fixturesDir := filepath.Join(wd, "..", "..", "test", "fixtures")

	tests := []struct {
		name        string
		filepath    string
		docType     string
		expectError bool
	}{
		{
			name:        "parse agent",
			filepath:    filepath.Join(fixturesDir, "agent-test.md"),
			docType:     "agent",
			expectError: false,
		},
		{
			name:        "parse command",
			filepath:    filepath.Join(fixturesDir, "command-test.md"),
			docType:     "command",
			expectError: false,
		},
		{
			name:        "parse memory",
			filepath:    filepath.Join(fixturesDir, "memory-test.md"),
			docType:     "memory",
			expectError: false,
		},
		{
			name:        "parse skill",
			filepath:    filepath.Join(fixturesDir, "skill-test.md"),
			docType:     "skill",
			expectError: false,
		},
		{
			name:        "parse invalid yaml",
			filepath:    filepath.Join(fixturesDir, "command-invalid.md"),
			docType:     "command",
			expectError: true,
		},
		{
			name:        "parse unsupported type",
			filepath:    filepath.Join(fixturesDir, "agent-test.md"),
			docType:     "unsupported",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseDocument(tt.filepath, tt.docType)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDocument() unexpected error: %v", err)
				return
			}

			switch tt.docType {
			case "agent":
				if _, ok := doc.(*models.Agent); !ok {
					t.Errorf("ParseDocument() expected *models.Agent, got %T", doc)
				}
			case "command":
				if _, ok := doc.(*models.Command); !ok {
					t.Errorf("ParseDocument() expected *models.Command, got %T", doc)
				}
			case "memory":
				if _, ok := doc.(*models.Memory); !ok {
					t.Errorf("ParseDocument() expected *models.Memory, got %T", doc)
				}
			case "skill":
				if _, ok := doc.(*models.Skill); !ok {
					t.Errorf("ParseDocument() expected *models.Skill, got %T", doc)
				}
			}
		})
	}
}

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedYAML string
		expectedBody string
		expectError  bool
	}{
		{
			name: "valid frontmatter",
			content: `---
key: value
---
markdown body`,
			expectedYAML: "key: value",
			expectedBody: "markdown body",
			expectError:  false,
		},
		{
			name: "no frontmatter delimiters",
			content: `no delimiters here
just markdown`,
			expectedYAML: "",
			expectedBody: "no delimiters here\njust markdown",
			expectError:  false,
		},
		{
			name: "only first delimiter",
			content: `---
key: value
no closing delimiter`,
			expectedYAML: "",
			expectedBody: "---\nkey: value\nno closing delimiter",
			expectError:  false,
		},
		{
			name: "empty frontmatter",
			content: `---
---
empty body`,
			expectedYAML: "",
			expectedBody: "empty body",
			expectError:  false,
		},
		{
			name: "multiline yaml and body",
			content: `---
key1: value1
key2: value2
---
body line 1
body line 2`,
			expectedYAML: "key1: value1\nkey2: value2",
			expectedBody: "body line 1\nbody line 2",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml, body, err := extractFrontmatter(tt.content)

			if tt.expectError && err == nil {
				t.Errorf("extractFrontmatter() expected error, got nil")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("extractFrontmatter() unexpected error: %v", err)
				return
			}

			if yaml != tt.expectedYAML {
				t.Errorf("extractFrontmatter() yaml = %v, want %v", yaml, tt.expectedYAML)
			}

			if body != tt.expectedBody {
				t.Errorf("extractFrontmatter() body = %v, want %v", body, tt.expectedBody)
			}
		})
	}
}

func TestDetectType(t *testing.T) {
	tests := []struct {
		name         string
		filepath     string
		expectedType string
	}{
		{
			name:         "agent prefix",
			filepath:     "agent-test.md",
			expectedType: "agent",
		},
		{
			name:         "agent suffix",
			filepath:     "test-agent.md",
			expectedType: "agent",
		},
		{
			name:         "command prefix",
			filepath:     "command-test.md",
			expectedType: "command",
		},
		{
			name:         "command suffix",
			filepath:     "test-command.md",
			expectedType: "command",
		},
		{
			name:         "memory prefix",
			filepath:     "memory-test.md",
			expectedType: "memory",
		},
		{
			name:         "memory suffix",
			filepath:     "test-memory.md",
			expectedType: "memory",
		},
		{
			name:         "skill prefix",
			filepath:     "skill-test.md",
			expectedType: "skill",
		},
		{
			name:         "skill suffix",
			filepath:     "test-skill.md",
			expectedType: "skill",
		},
		{
			name:         "unrecognizable",
			filepath:     "my-document.md",
			expectedType: "",
		},
		{
			name:         "no extension",
			filepath:     "agent-test",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType := DetectType(tt.filepath)
			if docType != tt.expectedType {
				t.Errorf("DetectType() = %v, want %v", docType, tt.expectedType)
			}
		})
	}
}
