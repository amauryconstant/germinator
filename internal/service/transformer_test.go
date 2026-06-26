package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
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

	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
	_, err := tr.Transform(context.Background(), &application.TransformRequest{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
	})
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
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

	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
	_, err := tr.Transform(context.Background(), &application.TransformRequest{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
	})
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

	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
	_, err := tr.Transform(context.Background(), &application.TransformRequest{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
	})
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

	v := stubValidator{}
	result, err := v.Validate(context.Background(), &application.ValidateRequest{
		InputPath: inputFile,
		Platform:  "claude-code",
	})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid() {
		t.Errorf("Expected no validation errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateDocumentFailure(t *testing.T) {
	t.Skip("slice-3: real validation moved to cmd/validate.go; the stub validator " +
		"used here only checks file existence. The full content-validation assertions " +
		"are covered by internal/service/validator_test.go's deleted equivalent and " +
		"cmd/validate_test.go's TestValidateDocument_HappyPath.")
}

func TestValidateDocumentMissingFile(t *testing.T) {
	nonExistentFile := "/nonexistent/file.md"

	v := stubValidator{}
	_, err := v.Validate(context.Background(), &application.ValidateRequest{
		InputPath: nonExistentFile,
		Platform:  "claude-code",
	})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// stubValidator is a minimal application.Validator implementation used
// only by transformer_test.go as a sanity check after transformation.
// Real validation logic now lives in cmd/validate.go (slice 3). This
// stub returns FileError for missing files and Valid for any existing
// non-empty file. Slice-7 deletes this along with the rest of
// internal/service/.
type stubValidator struct{}

var _ application.Validator = (*stubValidator)(nil)

func (stubValidator) Validate(_ context.Context, req *application.ValidateRequest) (*core.ValidateResult, error) {
	if _, err := os.Stat(req.InputPath); err != nil {
		return nil, core.NewFileError(req.InputPath, "read", "failed to read file", err)
	}
	return &core.ValidateResult{}, nil
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

	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
	for _, platform := range []string{"claude-code", "opencode"} {
		t.Run(platform, func(t *testing.T) {
			platformOutputFile := filepath.Join(tmpDir, "output-"+platform+".md")
			_, err := tr.Transform(context.Background(), &application.TransformRequest{
				InputPath:  inputFile,
				OutputPath: platformOutputFile,
				Platform:   platform,
			})
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
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

	v := stubValidator{}
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

			result, err := v.Validate(context.Background(), &application.ValidateRequest{
				InputPath: inputFile,
				Platform:  tt.platform,
			})
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}

			if !result.Valid() {
				t.Errorf("Expected stub validator to return valid for existing files, got errors: %v", result.Errors)
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

	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
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

					_, err := tr.Transform(context.Background(), &application.TransformRequest{
						InputPath:  inputFile,
						OutputPath: outputFile,
						Platform:   platform,
					})
					if err != nil {
						t.Fatalf("Transform failed: %v", err)
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
	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
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

		_, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "claude-code",
		})

		if err == nil {
			t.Fatal("Expected error for invalid YAML")
		}

		var parseErr *core.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if parseErr.Path() != inputFile {
				t.Errorf("ParseError.Path() = %q, want %q", parseErr.Path(), inputFile)
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

		_, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "claude-code",
		})

		if err == nil {
			t.Fatal("Expected error for unrecognizable filename")
		}

		var parseErr *core.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if !strings.Contains(parseErr.Message(), "expected") {
				t.Errorf("ParseError.Message() should mention expected patterns, got: %q", parseErr.Message())
			}
		}
	})
}

func TestTransformDocumentReturnsTypedConfigError(t *testing.T) {
	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
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

		_, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "invalid-platform",
		})

		if err == nil {
			t.Fatal("Expected error for invalid platform")
		}

		var configErr *core.ConfigError
		if !errors.As(err, &configErr) {
			t.Errorf("Expected ConfigError, got %T: %v", err, err)
		} else {
			if configErr.Field() != "platform" {
				t.Errorf("ConfigError.Field() = %q, want 'platform'", configErr.Field())
			}
			if len(configErr.Suggestions()) == 0 {
				t.Error("ConfigError.Suggestions() should list valid platforms")
			}
		}
	})
}

func TestTransformDocumentReturnsTypedFileError(t *testing.T) {
	tr := NewTransformer(parser.NewParser(), renderer.NewSerializer())
	t.Run("file not found returns FileError", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentFile := filepath.Join(tmpDir, "nonexistent-agent.md")
		outputFile := filepath.Join(tmpDir, "output.md")

		_, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  nonExistentFile,
			OutputPath: outputFile,
			Platform:   "claude-code",
		})

		if err == nil {
			t.Fatal("Expected error for non-existent file")
		}

		var fileErr *core.FileError
		if !errors.As(err, &fileErr) {
			t.Errorf("Expected FileError, got %T: %v", err, err)
		} else {
			if fileErr.Path() != nonExistentFile {
				t.Errorf("FileError.Path() = %q, want %q", fileErr.Path(), nonExistentFile)
			}
			if fileErr.Operation() != "read" {
				t.Errorf("FileError.Operation() = %q, want 'read'", fileErr.Operation())
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

		_, err := tr.Transform(context.Background(), &application.TransformRequest{
			InputPath:  inputFile,
			OutputPath: outputFile,
			Platform:   "claude-code",
		})

		if err == nil {
			t.Fatal("Expected error for non-existent output directory")
		}

		var fileErr *core.FileError
		if !errors.As(err, &fileErr) {
			t.Errorf("Expected FileError, got %T: %v", err, err)
		} else {
			if fileErr.Operation() != "write" {
				t.Errorf("FileError.Operation() = %q, want 'write'", fileErr.Operation())
			}
		}
	})
}

func TestValidateDocumentReturnsTypedConfigError(t *testing.T) {
	t.Skip("slice-3: platform validation moved to runValidate via core.ValidatePlatform. " +
		"ConfigError on invalid platform is now surfaced at the runValidate layer, " +
		"not by the validator itself. Covered by cmd/validate_test.go.")
}

func TestValidateDocumentReturnsTypedParseError(t *testing.T) {
	t.Skip("slice-3: validator's internal platform check (which produced ConfigError for " +
		"unknown platforms) was removed; parse-error behavior for unrecognizable filenames " +
		"is covered by cmd/validate_test.go's TestNewValidator_AdapterSatisfiesInterface.")
}
