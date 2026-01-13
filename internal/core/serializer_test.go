package core

import (
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
)

func TestRenderDocumentAgent(t *testing.T) {
	agent := &models.Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Tools:       []string{"editor", "bash"},
		Model:       "sonnet",
		Content:     "This is test content\nWith multiple lines",
	}

	result, err := RenderDocument(agent, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "name: test-agent") {
		t.Error("Expected name field in output")
	}
	if !strings.Contains(result, "description: Test agent") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "tools:") {
		t.Error("Expected tools section in output")
	}
	if !strings.Contains(result, "  - editor") {
		t.Error("Expected editor tool in output")
	}
	if !strings.Contains(result, "This is test content") {
		t.Error("Expected content in output")
	}
	if !strings.Contains(result, "---") {
		t.Error("Expected frontmatter delimiters in output")
	}
}

func TestRenderDocumentCommand(t *testing.T) {
	command := &models.Command{
		Name:         "test-command",
		Description:  "Test command",
		AllowedTools: []string{"bash"},
		Content:      "Command content here",
	}

	result, err := RenderDocument(command, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "description: Test command") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "allowed-tools:") {
		t.Error("Expected allowed-tools section in output")
	}
	if !strings.Contains(result, "Command content here") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentSkill(t *testing.T) {
	skill := &models.Skill{
		Name:        "test-skill",
		Description: "Test skill description",
		Model:       "haiku",
		Content:     "Skill instructions",
	}

	result, err := RenderDocument(skill, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "name: test-skill") {
		t.Error("Expected name field in output")
	}
	if !strings.Contains(result, "description: Test skill description") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "Skill instructions") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentMemory(t *testing.T) {
	memory := &models.Memory{
		Paths:   []string{"src/**/*.go", "README.md"},
		Content: "Memory content\nWith multiple lines",
	}

	result, err := RenderDocument(memory, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "paths:") {
		t.Error("Expected paths section in output")
	}
	if !strings.Contains(result, "  - src/**/*.go") {
		t.Error("Expected path in output")
	}
	if !strings.Contains(result, "Memory content") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentUnknownType(t *testing.T) {
	unknownDoc := "not a document"

	_, err := RenderDocument(unknownDoc, "claude-code")
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected 'unknown document type' error, got: %v", err)
	}
}

func TestRenderDocumentInvalidTemplate(t *testing.T) {
	agent := &models.Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Content:     "Content",
	}

	_, err := RenderDocument(agent, "invalid-platform")
	if err == nil {
		t.Error("Expected error for invalid platform")
	}
}

func TestRenderDocumentMarkdownBodyPreservation(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"simple text", "Simple text content"},
		{"multiline", "Line 1\nLine 2\nLine 3"},
		{"markdown formatting", "# Header\n\nSome **bold** text"},
		{"code blocks", "```go\nfunc test() {}\n```"},
		{"empty content", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &models.Agent{
				Name:        "test-agent",
				Description: "Test",
				Content:     tt.content,
			}

			result, err := RenderDocument(agent, "claude-code")
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}

			if !strings.Contains(result, tt.content) {
				t.Errorf("Expected content %q not found in output", tt.content)
			}

			parts := strings.Split(result, "---")
			if len(parts) < 3 {
				t.Error("Expected frontmatter with --- delimiters")
			}

			body := strings.TrimSpace(parts[len(parts)-1])
			if body != tt.content && tt.content != "" {
				t.Errorf("Expected body %q, got %q", tt.content, body)
			}
		})
	}
}

func TestGetDocType(t *testing.T) {
	tests := []struct {
		name     string
		doc      interface{}
		expected string
		wantErr  bool
	}{
		{"agent", &models.Agent{}, "agent", false},
		{"command", &models.Command{}, "command", false},
		{"skill", &models.Skill{}, "skill", false},
		{"memory", &models.Memory{}, "memory", false},
		{"unknown", "string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDocType(tt.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDocType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getDocType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
