package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlatformDocumentClaudeCodeAgent(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name        string
		filepath    string
		platform    string
		docType     string
		expectError bool
		validateDoc func(t *testing.T, doc interface{})
	}{
		{
			name:        "valid claude-code agent",
			filepath:    filepath.Join(fixturesDir, "claude-code", "agent.md"),
			platform:    "claude-code",
			docType:     "agent",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				agent, ok := doc.(*CanonicalAgent)
				if !ok {
					t.Fatalf("expected *CanonicalAgent, got %T", doc)
				}
				if agent.Name != "code-reviewer" {
					t.Errorf("agent.Name = %s, want code-reviewer", agent.Name)
				}
				if agent.Description != "Expert code review specialist" {
					t.Errorf("agent.Description = %s, want 'Expert code review specialist'", agent.Description)
				}
				if len(agent.Tools) != 3 {
					t.Errorf("len(agent.Tools) = %d, want 3", len(agent.Tools))
				}
				if agent.PermissionPolicy != "restrictive" {
					t.Errorf("agent.PermissionPolicy = %s, want restrictive", agent.PermissionPolicy)
				}
				if len(agent.Content) == 0 {
					t.Error("agent.Content should not be empty")
				}
			},
		},
		{
			name:        "valid claude-code command",
			filepath:    filepath.Join(fixturesDir, "claude-code", "command.md"),
			platform:    "claude-code",
			docType:     "command",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				command, ok := doc.(*CanonicalCommand)
				if !ok {
					t.Fatalf("expected *CanonicalCommand, got %T", doc)
				}
				if command.Name != "" {
					t.Logf("command.Name = %s", command.Name)
				}
				if len(command.Content) == 0 {
					t.Error("command.Content should not be empty")
				}
			},
		},
		{
			name:        "valid claude-code skill",
			filepath:    filepath.Join(fixturesDir, "claude-code", "skill.md"),
			platform:    "claude-code",
			docType:     "skill",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				skill, ok := doc.(*CanonicalSkill)
				if !ok {
					t.Fatalf("expected *CanonicalSkill, got %T", doc)
				}
				if len(skill.Content) == 0 {
					t.Error("skill.Content should not be empty")
				}
			},
		},
		{
			name:        "valid claude-code memory",
			filepath:    filepath.Join(fixturesDir, "claude-code", "memory.md"),
			platform:    "claude-code",
			docType:     "memory",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				_, ok := doc.(*CanonicalMemory)
				if !ok {
					t.Fatalf("expected *CanonicalMemory, got %T", doc)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParsePlatformDocument(tt.filepath, tt.platform, tt.docType)

			if tt.expectError {
				if err == nil {
					t.Error("ParsePlatformDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePlatformDocument() unexpected error: %v", err)
			}

			if tt.validateDoc != nil {
				tt.validateDoc(t, doc)
			}
		})
	}
}

func TestParsePlatformDocumentOpenCodeAgent(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name        string
		filepath    string
		platform    string
		docType     string
		expectError bool
		validateDoc func(t *testing.T, doc interface{})
	}{
		{
			name:        "valid opencode agent",
			filepath:    filepath.Join(fixturesDir, "opencode", "agent.md"),
			platform:    "opencode",
			docType:     "agent",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				agent, ok := doc.(*CanonicalAgent)
				if !ok {
					t.Fatalf("expected *CanonicalAgent, got %T", doc)
				}
				if agent.Name != "code-reviewer" {
					t.Errorf("agent.Name = %s, want code-reviewer", agent.Name)
				}
				if agent.Description != "Expert code review specialist" {
					t.Errorf("agent.Description = %s, want 'Expert code review specialist'", agent.Description)
				}
				if len(agent.Tools) != 3 {
					t.Errorf("len(agent.Tools) = %d, want 3", len(agent.Tools))
				}
				if len(agent.Content) == 0 {
					t.Error("agent.Content should not be empty")
				}
			},
		},
		{
			name:        "valid opencode command",
			filepath:    filepath.Join(fixturesDir, "opencode", "command.md"),
			platform:    "opencode",
			docType:     "command",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				command, ok := doc.(*CanonicalCommand)
				if !ok {
					t.Fatalf("expected *CanonicalCommand, got %T", doc)
				}
				if len(command.Content) == 0 {
					t.Error("command.Content should not be empty")
				}
			},
		},
		{
			name:        "valid opencode skill",
			filepath:    filepath.Join(fixturesDir, "opencode", "skill.md"),
			platform:    "opencode",
			docType:     "skill",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				skill, ok := doc.(*CanonicalSkill)
				if !ok {
					t.Fatalf("expected *CanonicalSkill, got %T", doc)
				}
				if len(skill.Content) == 0 {
					t.Error("skill.Content should not be empty")
				}
			},
		},
		{
			name:        "valid opencode memory",
			filepath:    filepath.Join(fixturesDir, "opencode", "memory.md"),
			platform:    "opencode",
			docType:     "memory",
			expectError: false,
			validateDoc: func(t *testing.T, doc interface{}) {
				_, ok := doc.(*CanonicalMemory)
				if !ok {
					t.Fatalf("expected *CanonicalMemory, got %T", doc)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParsePlatformDocument(tt.filepath, tt.platform, tt.docType)

			if tt.expectError {
				if err == nil {
					t.Error("ParsePlatformDocument() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePlatformDocument() unexpected error: %v", err)
			}

			if tt.validateDoc != nil {
				tt.validateDoc(t, doc)
			}
		})
	}
}

func TestParsePlatformDocumentInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.md")
	invalidContent := `---
name: test
description: test
invalid yaml syntax: [unclosed bracket
---
content`

	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParsePlatformDocument(invalidFile, "claude-code", "agent")
	if err == nil {
		t.Error("ParsePlatformDocument() expected error for invalid YAML, got nil")
	}
	if err == nil || !containsString(err.Error(), "failed to parse YAML") && !containsString(err.Error(), "failed to extract frontmatter") {
		t.Errorf("ParsePlatformDocument() error = %v, want YAML parsing error", err)
	}
}

func TestParsePlatformDocumentFileNotFound(t *testing.T) {
	nonExistentPath := "/non/existent/path/file.md"

	_, err := ParsePlatformDocument(nonExistentPath, "claude-code", "agent")
	if err == nil {
		t.Error("ParsePlatformDocument() expected error for non-existent file, got nil")
	}
	if !containsString(err.Error(), "failed to read file") && !containsString(err.Error(), "no such file") {
		t.Errorf("ParsePlatformDocument() error = %v, want file not found error", err)
	}
}

func TestParsePlatformDocumentContentPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-agent.md")
	testContent := `---
name: test-agent
description: Test agent
---
This is the content
with multiple lines
and formatting
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	doc, err := ParsePlatformDocument(testFile, "claude-code", "agent")
	if err != nil {
		t.Fatalf("ParsePlatformDocument() unexpected error: %v", err)
	}

	agent, ok := doc.(*CanonicalAgent)
	if !ok {
		t.Fatalf("expected *CanonicalAgent, got %T", doc)
	}

	expectedContent := "This is the content\nwith multiple lines\nand formatting\n"
	if agent.Content != expectedContent {
		t.Errorf("agent.Content = %q, want %q", agent.Content, expectedContent)
	}
}

func TestParsePlatformDocumentUnsupportedPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	testContent := `---
name: test
description: test
---
content`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParsePlatformDocument(testFile, "unsupported-platform", "agent")
	if err == nil {
		t.Error("ParsePlatformDocument() expected error for unsupported platform, got nil")
	}
	if !containsString(err.Error(), "unsupported platform") {
		t.Errorf("ParsePlatformDocument() error = %v, want unsupported platform error", err)
	}
}

func TestParsePlatformDocumentUnsupportedDocType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	testContent := `---
name: test
description: test
---
content`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParsePlatformDocument(testFile, "claude-code", "unsupported-type")
	if err == nil {
		t.Error("ParsePlatformDocument() expected error for unsupported document type, got nil")
	}
	if !containsString(err.Error(), "unknown document type") {
		t.Errorf("ParsePlatformDocument() error = %v, want unknown document type error", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInString(s, substr))
}

func containsInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
