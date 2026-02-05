package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTransformDocumentSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	content := `---
name: test-agent
description: A test agent
tools:
  - editor
  - bash
---
This is test content
`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err != nil {
		t.Fatalf("TransformDocument failed: %v", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(outputContent) == 0 {
		t.Error("Output file is empty")
	}
}

func TestTransformDocumentParseFailure(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	invalidContent := `---
name: "test-agent" "invalid yaml
---
content`

	if err := os.WriteFile(inputFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}

	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Error("Output file should not exist on parse failure")
	}
}

func TestTransformDocumentWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")
	outputFile := "/nonexistent/directory/output.md"

	content := `---
name: test-agent
description: A test agent
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err == nil {
		t.Error("Expected error for non-existent output directory")
	}
}

func TestValidateDocumentSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")

	content := `---
name: test-agent
description: A test agent
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	errs, err := ValidateDocument(inputFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateDocumentFailure(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-agent.md")

	content := `---
name: TEST-AGENT
description: ""
---
content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	errs, err := ValidateDocument(inputFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestValidateDocumentMissingFile(t *testing.T) {
	nonExistentFile := "/nonexistent/file.md"

	_, err := ValidateDocument(nonExistentFile, "claude-code")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestTransformDocumentCanonicalAgent(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "code-reviewer-agent.md")

	content := `---
name: code-reviewer
description: Reviews code changes and provides feedback
permissionPolicy: balanced
tools:
  - bash
  - editor
behavior:
  mode: primary
  temperature: 0.3
  steps: 100
---
Reviews code changes and provides constructive feedback.
`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	for _, platform := range []string{"claude-code", "opencode"} {
		t.Run(platform, func(t *testing.T) {
			platformOutputFile := filepath.Join(tmpDir, "output-"+platform+".md")
			err := TransformDocument(inputFile, platformOutputFile, platform)
			if err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
			}

			output, err := os.ReadFile(platformOutputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			if len(output) == 0 {
				t.Error("Output file is empty")
			}
		})
	}
}

func TestValidateDocumentCanonical(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		platform    string
		expectError bool
	}{
		{
			name: "valid canonical agent",
			content: `---
name: test-agent
description: Test agent
permissionPolicy: permissive
---
Content`,
			platform:    "claude-code",
			expectError: false,
		},
		{
			name: "invalid agent name",
			content: `---
name: Invalid_Name
description: Test agent
---
Content`,
			platform:    "claude-code",
			expectError: true,
		},
		{
			name: "valid canonical skill",
			content: `---
name: git-workflow
description: Git workflow skill
---
Content`,
			platform:    "opencode",
			expectError: false,
		},
		{
			name: "invalid permission policy",
			content: `---
name: test-agent
description: Test agent
permissionPolicy: invalid
---
Content`,
			platform:    "claude-code",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var inputFile string
			switch tt.name {
			case "valid canonical agent", "invalid agent name", "invalid permission policy":
				inputFile = filepath.Join(tmpDir, "test-agent.md")
			case "valid canonical skill":
				inputFile = filepath.Join(tmpDir, "git-workflow-skill.md")
			}

			if err := os.WriteFile(inputFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			errs, err := ValidateDocument(inputFile, tt.platform)
			if err != nil {
				t.Fatalf("ValidateDocument failed: %v", err)
			}

			if tt.expectError && len(errs) == 0 {
				t.Error("Expected validation errors, got none")
			}
			if !tt.expectError && len(errs) > 0 {
				t.Errorf("Expected no validation errors, got %d: %v", len(errs), errs)
			}
		})
	}
}

func TestTransformDocumentPlatforms(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		docType string
		content string
	}{
		{
			name:    "agent",
			docType: "agent",
			content: `---
name: test-agent
description: Test agent
permissionPolicy: balanced
behavior:
  steps: 50
---
Agent content`,
		},
		{
			name:    "command",
			docType: "command",
			content: `---
name: test-command
description: Test command
execution:
  subtask: true
---
Command template: echo "hello"`,
		},
		{
			name:    "skill",
			docType: "skill",
			content: `---
name: test-skill
description: Test skill
extensions:
  license: MIT
---
Skill content`,
		},
		{
			name:    "memory",
			docType: "memory",
			content: `---
paths:
  - README.md
  - AGENTS.md
---
Memory content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, platform := range []string{"claude-code", "opencode"} {
				t.Run(platform, func(t *testing.T) {
					var inputFile string
					switch tt.docType {
					case "agent":
						inputFile = filepath.Join(tmpDir, "test-agent.md")
					case "command":
						inputFile = filepath.Join(tmpDir, "test-command.md")
					case "skill":
						inputFile = filepath.Join(tmpDir, "test-skill.md")
					case "memory":
						inputFile = filepath.Join(tmpDir, "project-memory.md")
					}
					outputFile := filepath.Join(tmpDir, tt.docType+"-"+platform+".md")

					if err := os.WriteFile(inputFile, []byte(tt.content), 0644); err != nil {
						t.Fatalf("Failed to create input file: %v", err)
					}

					err := TransformDocument(inputFile, outputFile, platform)
					if err != nil {
						t.Fatalf("TransformDocument failed: %v", err)
					}

					if _, err := os.Stat(outputFile); os.IsNotExist(err) {
						t.Error("Output file was not created")
					}
				})
			}
		})
	}
}
