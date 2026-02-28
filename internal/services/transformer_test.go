package services

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
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

func TestTransformDocumentReturnsTypedParseError(t *testing.T) {
	t.Run("invalid YAML returns ParseError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test-agent.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		invalidYAML := `---
name: "unclosed string
---
Content`

		if err := os.WriteFile(inputFile, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		err := TransformDocument(inputFile, outputFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for invalid YAML")
		}

		var parseErr *gerrors.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if parseErr.Path != inputFile {
				t.Errorf("ParseError.Path = %q, want %q", parseErr.Path, inputFile)
			}
		}

		if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
			t.Error("Output file should not be created on parse error")
		}
	})

	t.Run("unrecognizable filename returns ParseError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "unrecognizable.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		content := `---
name: test
description: Test
---
Content`

		if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		err := TransformDocument(inputFile, outputFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for unrecognizable filename")
		}

		var parseErr *gerrors.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if !strings.Contains(parseErr.Message, "expected") {
				t.Errorf("ParseError.Message should mention expected patterns, got: %q", parseErr.Message)
			}
		}
	})
}

func TestTransformDocumentReturnsTypedConfigError(t *testing.T) {
	t.Run("invalid platform returns ConfigError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test-agent.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		content := `---
name: test-agent
description: Test agent
---
Content`

		if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		err := TransformDocument(inputFile, outputFile, "invalid-platform")

		if err == nil {
			t.Fatal("Expected error for invalid platform")
		}

		var configErr *gerrors.ConfigError
		if !errors.As(err, &configErr) {
			t.Errorf("Expected ConfigError, got %T: %v", err, err)
		} else {
			if configErr.Field != "platform" {
				t.Errorf("ConfigError.Field = %q, want 'platform'", configErr.Field)
			}
			if len(configErr.Available) == 0 {
				t.Error("ConfigError.Available should list valid platforms")
			}
		}
	})
}

func TestTransformDocumentReturnsTypedFileError(t *testing.T) {
	t.Run("file not found returns FileError", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentFile := filepath.Join(tmpDir, "nonexistent-agent.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		err := TransformDocument(nonExistentFile, outputFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for non-existent file")
		}

		var fileErr *gerrors.FileError
		if !errors.As(err, &fileErr) {
			t.Errorf("Expected FileError, got %T: %v", err, err)
		} else {
			if fileErr.Path != nonExistentFile {
				t.Errorf("FileError.Path = %q, want %q", fileErr.Path, nonExistentFile)
			}
			if fileErr.Operation != "read" {
				t.Errorf("FileError.Operation = %q, want 'read'", fileErr.Operation)
			}
			if !fileErr.IsNotFound() {
				t.Error("FileError.IsNotFound() should return true")
			}
		}
	})

	t.Run("write error returns FileError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test-agent.md")
		outputFile := "/nonexistent/directory/output.md"

		content := `---
name: test-agent
description: Test agent
---
Content`

		if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		err := TransformDocument(inputFile, outputFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for non-existent output directory")
		}

		var fileErr *gerrors.FileError
		if !errors.As(err, &fileErr) {
			t.Errorf("Expected FileError, got %T: %v", err, err)
		} else {
			if fileErr.Operation != "write" {
				t.Errorf("FileError.Operation = %q, want 'write'", fileErr.Operation)
			}
		}
	})
}

func TestValidateDocumentReturnsTypedConfigError(t *testing.T) {
	t.Run("invalid platform returns ConfigError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "test-agent.md")

		content := `---
name: test-agent
description: Test agent
---
Content`

		if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		errs, err := ValidateDocument(inputFile, "invalid-platform")

		if err != nil {
			t.Fatalf("ValidateDocument should not return fatal error: %v", err)
		}

		if len(errs) == 0 {
			t.Fatal("Expected validation errors for invalid platform")
		}

		foundConfigError := false
		for _, e := range errs {
			var configErr *gerrors.ConfigError
			if errors.As(e, &configErr) {
				foundConfigError = true
				if configErr.Field != "platform" {
					t.Errorf("ConfigError.Field = %q, want 'platform'", configErr.Field)
				}
				break
			}
		}

		if !foundConfigError {
			t.Error("Expected ConfigError in validation errors")
		}
	})
}

func TestValidateDocumentReturnsTypedParseError(t *testing.T) {
	t.Run("unrecognizable filename returns ParseError", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "unrecognizable.md")

		content := `---
name: test
description: Test
---
Content`

		if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		_, err := ValidateDocument(inputFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for unrecognizable filename")
		}

		var parseErr *gerrors.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		}
	})
}
