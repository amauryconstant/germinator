package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	testPath := filepath.Join(wd, "..", "test")
	if _, err := os.Stat(testPath); err == nil {
		return filepath.Abs(filepath.Join(wd, ".."))
	}

	altTestPath := filepath.Join(wd, "..", "..", "test")
	if _, err := os.Stat(altTestPath); err == nil {
		return filepath.Abs(filepath.Join(wd, "..", ".."))
	}

	return "", os.ErrNotExist
}

func newTestConfig() *CommandConfig {
	return &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      0,
	}
}

func TestValidateCommandWithActualServices(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
This is valid content`

	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
		InputPath: validFile,
		Platform:  "claude-code",
	})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid() {
		t.Errorf("Expected no validation errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestAdaptCommand(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/input-agent.md"
	outputFile := tmpDir + "/output-agent.md"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
---
This is test content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		platform    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing platform flag",
			platform:    "",
			expectError: true,
			errorMsg:    "platform is required",
		},
		{
			name:        "invalid platform",
			platform:    "invalid-platform",
			expectError: true,
			errorMsg:    "unknown platform",
		},
		{
			name:        "valid claude-code platform",
			platform:    "claude-code",
			expectError: true,
			errorMsg:    "template file not found",
		},
		{
			name:        "valid opencode platform",
			platform:    "opencode",
			expectError: true,
			errorMsg:    "template file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()
			if tt.expectError {
				_, err := cfg.Services.Transformer.Transform(context.Background(), &application.TransformRequest{
					InputPath:  inputFile,
					OutputPath: outputFile,
					Platform:   tt.platform,
				})
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				_, err := cfg.Services.Transformer.Transform(context.Background(), &application.TransformRequest{
					InputPath:  inputFile,
					OutputPath: outputFile,
					Platform:   tt.platform,
				})
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateCommandWithPlatformVariants(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := tmpDir + "/test-command.md"

	content := `---
name: test-command
description: A test command
template: command template
---
Command content`

	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		platform string
	}{
		{"claude-code platform", "claude-code"},
		{"opencode platform", "opencode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()
			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: validFile,
				Platform:  tt.platform,
			})
			if err != nil {
				t.Errorf("Validate failed for %s: %v", tt.platform, err)
			}

			if !result.Valid() {
				t.Errorf("Expected no validation errors for %s, got %d: %v", tt.platform, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := newTestConfig()
	versionCmd := NewVersionCommand(cfg)
	versionCmd.Run(versionCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "germinator") {
		t.Errorf("Version command output should contain 'germinator', got: %s", output)
	}
}

func TestRootCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := newTestConfig()
	rootCmd := NewRootCommand(cfg)
	rootCmd.Run(rootCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	if !strings.Contains(output, "Germinator is a configuration adapter") {
		t.Errorf("Root command should show help, got: %s", output)
	}
}

func TestValidateCommandValidDocument(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := tmpDir + "/test-skill.md"

	content := `---
name: test-skill
description: A test skill
---
This is valid content`

	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
		InputPath: validFile,
		Platform:  "claude-code",
	})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if !result.Valid() {
		t.Errorf("Expected no validation errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestCLIPlatformFlagValidation(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	validFile := tmpDir + "/agent-test.md"

	content := `---
name: test-agent
description: A test agent
---
This is valid content`

	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		platform    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing platform flag",
			platform:    "",
			expectError: true,
			errorMsg:    "platform is required",
		},
		{
			name:        "invalid platform",
			platform:    "invalid-platform",
			expectError: true,
			errorMsg:    "unknown platform",
		},
		{
			name:        "valid claude-code platform",
			platform:    "claude-code",
			expectError: false,
		},
		{
			name:        "valid opencode platform",
			platform:    "opencode",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()
			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: validFile,
				Platform:  tt.platform,
			})
			if err != nil {
				t.Errorf("Validate failed: %v", err)
			}
			if tt.expectError && result.Valid() {
				t.Errorf("Expected validation errors for platform %s but got none", tt.platform)
			}
			if !tt.expectError && !result.Valid() {
				t.Errorf("Unexpected validation errors for platform %s: %v", tt.platform, result.Errors)
			}
			if tt.errorMsg != "" && !result.Valid() {
				var errMsgs strings.Builder
				for _, e := range result.Errors {
					errMsgs.WriteString(e.Error())
					errMsgs.WriteString("; ")
				}
				if !strings.Contains(errMsgs.String(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, errMsgs.String())
				}
			}
		})
	}
}

func TestCLIDescriptiveErrorMessages(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		filename    string
		platform    string
		expectError bool
		errorMsg    string
	}{
		{
			name:     "invalid agent name format",
			filename: "agent-invalid-name.md",
			content: `---
name: Invalid_Name
description: Test agent
---
Test content`,
			platform:    "opencode",
			expectError: true,
			errorMsg:    "name must match pattern",
		},
		{
			name:     "invalid agent temperature",
			filename: "agent-invalid-temp.md",
			content: `---
name: test-agent
description: Test agent
behavior:
  temperature: 1.5
---
Test content`,
			platform:    "opencode",
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 1.0",
		},
		{
			name:     "invalid agent mode",
			filename: "agent-invalid-mode.md",
			content: `---
name: test-agent
description: Test agent
behavior:
  mode: invalid-mode
---
Test content`,
			platform:    "opencode",
			expectError: true,
			errorMsg:    "mode must be one of",
		},
		{
			name:     "unknown platform error lists available",
			filename: "agent-unknown-platform.md",
			content: `---
name: test-agent
description: Test agent
---
Test content`,
			platform:    "unknown-platform",
			expectError: true,
			errorMsg:    "💡",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.filename
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			cfg := newTestConfig()
			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: testFile,
				Platform:  tt.platform,
			})
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}
			if tt.expectError && result.Valid() {
				t.Errorf("Expected validation errors but got none")
			}
			if tt.expectError && !result.Valid() {
				var errMsgs strings.Builder
				for _, e := range result.Errors {
					errMsgs.WriteString(e.Error())
					errMsgs.WriteString("; ")
				}
				if !strings.Contains(errMsgs.String(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, errMsgs.String())
				}
			}
		})
	}
}

func TestCLIAdaptEndToEnd(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		content     string
		platform    string
		expectError bool
		validate    func(t *testing.T, output string)
	}{
		{
			name:     "adapt agent to claude-code",
			filename: "test-agent.md",
			platform: "claude-code",
			content: `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
---
Agent content`,
			expectError: false,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "name: test-agent") {
					t.Errorf("Expected output to contain name field")
				}
				if !strings.Contains(output, "bash") && !strings.Contains(output, "Bash") {
					t.Errorf("Expected output to contain bash or Bash, got:\n%s", output)
				}
				if !strings.Contains(output, "read") && !strings.Contains(output, "Read") {
					t.Errorf("Expected output to contain read or Read, got:\n%s", output)
				}
			},
		},
		{
			name:     "adapt agent to opencode",
			filename: "test-agent.md",
			platform: "opencode",
			content: `---
name: test-agent
description: A test agent
tools:
  - Bash
  - Read
permissionPolicy: restrictive
---
Agent content`,
			expectError: false,
			validate: func(t *testing.T, output string) {
				if strings.Contains(output, "name: test-agent") {
					t.Errorf("Expected OpenCode output to omit name field")
				}
				if !strings.Contains(output, "bash:") || !strings.Contains(output, "read:") {
					t.Errorf("Expected output to contain lowercase tools")
				}
				if !strings.Contains(output, "permission:") {
					t.Errorf("Expected output to contain permission field")
				}
			},
		},
		{
			name:     "adapt command to opencode omits tool fields",
			filename: "test-command.md",
			platform: "opencode",
			content: `---
name: test-command
description: A test command
allowed-tools:
  - bash
  - read
disallowed-tools:
  - write
---
Command content`,
			expectError: false,
			validate: func(t *testing.T, output string) {
				if strings.Contains(output, "name: test-command") {
					t.Errorf("Expected OpenCode output to omit name field")
				}
				if strings.Contains(output, "bash:") || strings.Contains(output, "read:") || strings.Contains(output, "write:") {
					t.Errorf("Expected OpenCode output to omit tool fields (OpenCode commands don't support tools)")
				}
			},
		},
		{
			name:     "adapt skill to opencode",
			filename: "test-skill.md",
			platform: "opencode",
			content: `---
name: test-skill
description: A test skill
extensions:
  license: MIT
  compatibility:
    - claude-code
    - opencode
  metadata:
    author: test
---
Skill content`,
			expectError: false,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "name: test-skill") {
					t.Errorf("Expected output to contain name field")
				}
				if !strings.Contains(output, "license: MIT") {
					t.Errorf("Expected output to contain license")
				}
				if !strings.Contains(output, "compatibility:") {
					t.Errorf("Expected output to contain compatibility")
				}
				if !strings.Contains(output, "metadata:") {
					t.Errorf("Expected output to contain metadata")
				}
			},
		},
		{
			name:     "adapt memory to opencode",
			filename: "test-memory.md",
			platform: "opencode",
			content: `---
paths:
  - README.md
  - src/
content: |
  Project context`,
			expectError: false,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "@README.md") {
					t.Errorf("Expected output to contain @ file references")
				}
				if !strings.Contains(output, "Project context") {
					t.Errorf("Expected output to contain content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile := tmpDir + "/" + tt.filename
			outputFile := tmpDir + "/output.md"

			if err := os.WriteFile(inputFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			cfg := newTestConfig()
			_, err := cfg.Services.Transformer.Transform(context.Background(), &application.TransformRequest{
				InputPath:  inputFile,
				OutputPath: outputFile,
				Platform:   tt.platform,
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.validate(t, string(output))
		})
	}
}

func TestCLIValidateEndToEnd(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		content     string
		platform    string
		expectError bool
		errorCount  int
	}{
		{
			name:     "valid agent opencode",
			filename: "test-agent.md",
			content: `---
name: test-agent
description: A test agent
---
Agent content`,
			platform:    "opencode",
			expectError: false,
			errorCount:  0,
		},
		{
			name:     "invalid agent missing description",
			filename: "test-agent.md",
			content: `---
name: test-agent
---
Agent content`,
			platform:    "opencode",
			expectError: true,
			errorCount:  1,
		},
		{
			name:     "invalid agent multiple errors",
			filename: "test-agent.md",
			content: `---
name: Invalid_Name
behavior:
  temperature: 2.0
---
Agent content`,
			platform:    "opencode",
			expectError: true,
			errorCount:  3,
		},
		{
			name:     "valid command claude-code",
			filename: "test-command.md",
			content: `---
name: test-command
description: A test command
---
Command content`,
			platform:    "claude-code",
			expectError: false,
			errorCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.filename

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			cfg := newTestConfig()
			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: testFile,
				Platform:  tt.platform,
			})
			if err != nil {
				t.Fatalf("Validate failed: %v", err)
			}

			if tt.expectError && len(result.Errors) != tt.errorCount {
				t.Errorf("Expected %d errors, got %d: %v", tt.errorCount, len(result.Errors), result.Errors)
			}
			if !tt.expectError && !result.Valid() {
				t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
			}
		})
	}
}

func TestCanonicalizeCommandWithAllFlags(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	outputFile := tmpDir + "/canonical-agent.yaml"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := newTestConfig()
	_, err = cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
		DocType:    "agent",
	})

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	_ = output

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file to be created: %s", outputFile)
	}

	result, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "name: test-agent") {
		t.Errorf("Expected output to contain name, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "description: A test agent") {
		t.Errorf("Expected output to contain description, got: %s", resultStr)
	}
}

func TestCanonicalizeCommandMissingPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/output.yaml",
		Platform:   "",
		DocType:    "agent",
	})
	if err == nil {
		t.Errorf("Expected error when platform is empty")
	}

	if !strings.Contains(err.Error(), "unsupported platform") {
		t.Errorf("Expected unsupported platform error, got: %v", err)
	}
}

func TestCanonicalizeCommandMissingType(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/output.yaml",
		Platform:   "claude-code",
		DocType:    "",
	})
	if err == nil {
		t.Errorf("Expected error when type is empty")
	}

	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}

func TestCanonicalizeCommandInvalidPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/output.yaml",
		Platform:   "invalid-platform",
		DocType:    "agent",
	})
	if err == nil {
		t.Errorf("Expected error when platform is invalid")
	}

	if !strings.Contains(err.Error(), "unsupported platform") {
		t.Errorf("Expected unsupported platform error, got: %v", err)
	}
}

func TestCanonicalizeCommandInvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/output.yaml",
		Platform:   "claude-code",
		DocType:    "invalid-type",
	})
	if err == nil {
		t.Errorf("Expected error when type is invalid")
	}

	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}

func TestCanonicalizeCommandFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/non-existent-file.md"

	cfg := newTestConfig()
	_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: tmpDir + "/output.yaml",
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err == nil {
		t.Errorf("Expected error when input file not found")
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected file read error, got: %v", err)
	}
}

func TestCanonicalizeCommandSuccessfulConversion(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-agent.md"
	outputFile := tmpDir + "/canonical-agent.yaml"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
  - read
permissionMode: default
---
Agent content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := newTestConfig()
	_, err = cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
		InputPath:  inputFile,
		OutputPath: outputFile,
		Platform:   "claude-code",
		DocType:    "agent",
	})
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}

	result, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "name: test-agent") {
		t.Errorf("Expected output to contain name, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "description: A test agent") {
		t.Errorf("Expected output to contain description, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "bash") && !strings.Contains(resultStr, "read") {
		t.Errorf("Expected output to contain tools, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "permissionPolicy: restrictive") {
		t.Errorf("Expected output to contain permissionPolicy, got: %s", resultStr)
	}
	if !strings.Contains(resultStr, "Agent content") {
		t.Errorf("Expected output to contain content, got: %s", resultStr)
	}
}

func TestValidateCommandVerboseFlag(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	validFile := tmpDir + "/agent-test.md"

	content := `---
name: test-agent
description: A test agent
---
Test content`

	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name              string
		verboseFlag       string
		expectStderrEmpty bool
		expectContains    []string
	}{
		{
			name:              "no verbose flag",
			verboseFlag:       "",
			expectStderrEmpty: true,
		},
		{
			name:           "level 1 verbose (-v)",
			verboseFlag:    "-v",
			expectContains: []string{"Validating file:", "Platform:"},
		},
		{
			name:           "level 2 verbose (-vv)",
			verboseFlag:    "-vv",
			expectContains: []string{"Validating file:", "Loading document...", "Parsing document structure...", "Running validation..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			stdoutR, stdoutW, _ := os.Pipe()
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			args := []string{"validate", validFile, "--platform", "claude-code"}
			if tt.verboseFlag != "" {
				args = append([]string{tt.verboseFlag}, args...)
			}

			cfg := newTestConfig()
			rootCmd := NewRootCommand(cfg)
			rootCmd.SetArgs(args)
			_ = rootCmd.Execute()

			_ = stdoutW.Close()
			_ = stderrW.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var stdoutBuf, stderrBuf bytes.Buffer
			_, _ = io.Copy(&stdoutBuf, stdoutR)
			_, _ = io.Copy(&stderrBuf, stderrR)

			stderr := stderrBuf.String()

			if tt.expectStderrEmpty && stderr != "" {
				t.Errorf("Expected empty stderr, got: %q", stderr)
			}

			for _, want := range tt.expectContains {
				if !strings.Contains(stderr, want) {
					t.Errorf("Expected stderr to contain %q, got: %q", want, stderr)
				}
			}
		})
	}
}

func TestAdaptCommandVerboseFlag(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	inputFile := tmpDir + "/agent-test.md"
	outputFile := tmpDir + "/output.md"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
---
Test content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name              string
		verboseFlag       string
		expectStderrEmpty bool
		expectContains    []string
	}{
		{
			name:              "no verbose flag",
			verboseFlag:       "",
			expectStderrEmpty: true,
		},
		{
			name:           "level 1 verbose (-v)",
			verboseFlag:    "-v",
			expectContains: []string{"Transforming document...", "Output path:"},
		},
		{
			name:           "level 2 verbose (-vv)",
			verboseFlag:    "-vv",
			expectContains: []string{"Transforming document...", "Loading source document...", "Rendering template", "Writing output file..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove(outputFile)

			oldStdout := os.Stdout
			oldStderr := os.Stderr
			stdoutR, stdoutW, _ := os.Pipe()
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			args := []string{"adapt", inputFile, outputFile, "--platform", "opencode"}
			if tt.verboseFlag != "" {
				args = append([]string{tt.verboseFlag}, args...)
			}

			cfg := newTestConfig()
			rootCmd := NewRootCommand(cfg)
			rootCmd.SetArgs(args)
			_ = rootCmd.Execute()

			_ = stdoutW.Close()
			_ = stderrW.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var stdoutBuf, stderrBuf bytes.Buffer
			_, _ = io.Copy(&stdoutBuf, stdoutR)
			_, _ = io.Copy(&stderrBuf, stderrR)

			stderr := stderrBuf.String()

			if tt.expectStderrEmpty && stderr != "" {
				t.Errorf("Expected empty stderr, got: %q", stderr)
			}

			for _, want := range tt.expectContains {
				if !strings.Contains(stderr, want) {
					t.Errorf("Expected stderr to contain %q, got: %q", want, stderr)
				}
			}
		})
	}
}

func TestCanonicalizeCommandVerboseFlag(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()
	inputFile := tmpDir + "/agent-test.md"
	outputFile := tmpDir + "/output.yaml"

	content := `---
name: test-agent
description: A test agent
tools:
  - bash
---
Test content`

	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name              string
		verboseFlag       string
		expectStderrEmpty bool
		expectContains    []string
	}{
		{
			name:              "no verbose flag",
			verboseFlag:       "",
			expectStderrEmpty: true,
		},
		{
			name:           "level 1 verbose (-v)",
			verboseFlag:    "-v",
			expectContains: []string{"Canonicalizing document...", "Output path:"},
		},
		{
			name:           "level 2 verbose (-vv)",
			verboseFlag:    "-vv",
			expectContains: []string{"Canonicalizing document...", "Parsing platform document...", "Validating document...", "Marshalling to canonical YAML..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove(outputFile)

			oldStdout := os.Stdout
			oldStderr := os.Stderr
			stdoutR, stdoutW, _ := os.Pipe()
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			args := []string{"canonicalize", inputFile, outputFile, "--platform", "claude-code", "--type", "agent"}
			if tt.verboseFlag != "" {
				args = append([]string{tt.verboseFlag}, args...)
			}

			cfg := newTestConfig()
			rootCmd := NewRootCommand(cfg)
			rootCmd.SetArgs(args)
			_ = rootCmd.Execute()

			_ = stdoutW.Close()
			_ = stderrW.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var stdoutBuf, stderrBuf bytes.Buffer
			_, _ = io.Copy(&stdoutBuf, stdoutR)
			_, _ = io.Copy(&stderrBuf, stderrR)

			stderr := stderrBuf.String()

			if tt.expectStderrEmpty && stderr != "" {
				t.Errorf("Expected empty stderr, got: %q", stderr)
			}

			for _, want := range tt.expectContains {
				if !strings.Contains(stderr, want) {
					t.Errorf("Expected stderr to contain %q, got: %q", want, stderr)
				}
			}
		})
	}
}

func TestValidateCommandExitCodes(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		filename       string
		content        string
		platform       string
		expectedCode   int
		expectErrorMsg string
	}{
		{
			name:         "unrecognizable filename - exit code 3",
			filename:     "invalid-name.md",
			content:      "",
			platform:     "claude-code",
			expectedCode: 3,
		},
		{
			name:     "missing description - exit code 2",
			filename: "agent-invalid.md",
			content: `---
name: test-agent
---
content`,
			platform:     "opencode",
			expectedCode: 2,
		},
		{
			name:     "invalid platform - exit code 2",
			filename: "agent-test.md",
			content: `---
name: test-agent
description: test
---
content`,
			platform:     "invalid-platform",
			expectedCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.filename
			if tt.content != "" {
				if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			cfg := newTestConfig()
			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: testFile,
				Platform:  tt.platform,
			})
			if err != nil {
				code := GetExitCodeForError(err)
				if int(code) != tt.expectedCode {
					t.Errorf("Expected exit code %d for error, got %d (error: %v)", tt.expectedCode, code, err)
				}
			} else if !result.Valid() {
				code := GetExitCodeForError(gerrors.NewValidationError("", "", "", result.Errors[0].Error()))
				if int(code) != tt.expectedCode {
					t.Errorf("Expected exit code %d for validation errors, got %d", tt.expectedCode, code)
				}
			} else {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestAdaptCommandExitCodes(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		inputFile    string
		outputFile   string
		content      string
		platform     string
		expectedCode int
	}{
		{
			name:         "unrecognizable filename - exit code 3",
			inputFile:    "invalid-name.md",
			outputFile:   tmpDir + "/output.md",
			content:      "",
			platform:     "claude-code",
			expectedCode: 3,
		},
		{
			name:       "invalid platform - exit code 2",
			inputFile:  "agent-test.md",
			outputFile: tmpDir + "/output.md",
			content: `---
name: test-agent
description: test
---
content`,
			platform:     "invalid-platform",
			expectedCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.inputFile
			if tt.content != "" {
				if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			cfg := newTestConfig()
			_, err := cfg.Services.Transformer.Transform(context.Background(), &application.TransformRequest{
				InputPath:  testFile,
				OutputPath: tt.outputFile,
				Platform:   tt.platform,
			})
			if err != nil {
				code := GetExitCodeForError(err)
				if int(code) != tt.expectedCode {
					t.Errorf("Expected exit code %d, got %d (error: %v)", tt.expectedCode, code, err)
				}
			} else {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestExitCodeForErrorTypes(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode ExitCode
	}{
		{
			name:         "ParseError returns exit code 3",
			err:          gerrors.NewParseError("test.yaml", "parse failed", nil),
			expectedCode: ExitCodeParse,
		},
		{
			name:         "ValidationError returns exit code 2",
			err:          gerrors.NewValidationError("", "name", "", "invalid field"),
			expectedCode: ExitCodeUsage,
		},
		{
			name:         "ConfigError returns exit code 2",
			err:          gerrors.NewConfigError("platform", "invalid", "unknown platform").WithSuggestions([]string{"claude-code"}),
			expectedCode: ExitCodeUsage,
		},
		{
			name:         "TransformError returns exit code 1",
			err:          gerrors.NewTransformError("render", "opencode", "failed", nil),
			expectedCode: ExitCodeError,
		},
		{
			name:         "FileError returns exit code 1",
			err:          gerrors.NewFileError("test.yaml", "read", "not found", nil),
			expectedCode: ExitCodeError,
		},
		{
			name:         "generic error returns exit code 1",
			err:          fmt.Errorf("something went wrong"),
			expectedCode: ExitCodeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetExitCodeForError(tt.err)
			if code != tt.expectedCode {
				t.Errorf("GetExitCodeForError() = %d, want %d", code, tt.expectedCode)
			}
		})
	}
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedCategory ErrorCategory
	}{
		{
			name:             "ParseError categorizes as CategoryParse",
			err:              gerrors.NewParseError("test.yaml", "failed", nil),
			expectedCategory: CategoryParse,
		},
		{
			name:             "ValidationError categorizes as CategoryValidation",
			err:              gerrors.NewValidationError("", "field", "", "invalid"),
			expectedCategory: CategoryValidation,
		},
		{
			name:             "TransformError categorizes as CategoryTransform",
			err:              gerrors.NewTransformError("op", "platform", "failed", nil),
			expectedCategory: CategoryTransform,
		},
		{
			name:             "FileError categorizes as CategoryFile",
			err:              gerrors.NewFileError("path", "read", "failed", nil),
			expectedCategory: CategoryFile,
		},
		{
			name:             "ConfigError categorizes as CategoryConfig",
			err:              gerrors.NewConfigError("field", "value", "invalid"),
			expectedCategory: CategoryConfig,
		},
		{
			name:             "generic error categorizes as CategoryGeneric",
			err:              fmt.Errorf("generic error"),
			expectedCategory: CategoryGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := CategorizeError(tt.err)
			if category != tt.expectedCategory {
				t.Errorf("CategorizeError() = %v, want %v", category, tt.expectedCategory)
			}
		})
	}
}

func TestCommandConfigInitialization(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name           string
		verbosity      Verbosity
		expectVerbose  bool
		expectVeryVerb bool
	}{
		{
			name:           "level 0",
			verbosity:      0,
			expectVerbose:  false,
			expectVeryVerb: false,
		},
		{
			name:           "level 1",
			verbosity:      1,
			expectVerbose:  true,
			expectVeryVerb: false,
		},
		{
			name:           "level 2",
			verbosity:      2,
			expectVerbose:  true,
			expectVeryVerb: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CommandConfig{
				Services:       NewServiceContainer(),
				ErrorFormatter: formatter,
				Verbosity:      tt.verbosity,
			}

			if cfg.ErrorFormatter == nil {
				t.Error("CommandConfig.ErrorFormatter should not be nil")
			}

			if cfg.Verbosity.IsVerbose() != tt.expectVerbose {
				t.Errorf("CommandConfig.Verbosity.IsVerbose() = %v, want %v", cfg.Verbosity.IsVerbose(), tt.expectVerbose)
			}

			if cfg.Verbosity.IsVeryVerbose() != tt.expectVeryVerb {
				t.Errorf("CommandConfig.Verbosity.IsVeryVerbose() = %v, want %v", cfg.Verbosity.IsVeryVerbose(), tt.expectVeryVerb)
			}
		})
	}
}

func TestCommandConfigContainsErrorFormatter(t *testing.T) {
	cfg := newTestConfig()

	if cfg.ErrorFormatter == nil {
		t.Fatal("ErrorFormatter should not be nil")
	}

	testErr := gerrors.NewParseError("test.yaml", "test error", nil)
	formatted := cfg.ErrorFormatter.Format(testErr)

	if !strings.Contains(formatted, "Parse error:") {
		t.Errorf("ErrorFormatter.Format() should format ParseError, got: %q", formatted)
	}
}

func TestHandleErrorWritesToStderr(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg := newTestConfig()

	testErr := gerrors.NewConfigError("platform", "invalid", "unknown platform").WithSuggestions([]string{"claude-code", "opencode"})

	formatted := cfg.ErrorFormatter.Format(testErr)
	exitCode := GetExitCodeForError(testErr)

	_, _ = fmt.Fprintln(os.Stderr, formatted)
	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Config error:") {
		t.Errorf("Expected stderr to contain 'Config error:', got: %q", output)
	}

	if !strings.Contains(output, "Hint:") {
		t.Errorf("Expected stderr to contain 'Hint:', got: %q", output)
	}

	if exitCode != ExitCodeUsage {
		t.Errorf("GetExitCodeForError(ConfigError) = %d, want %d", exitCode, ExitCodeUsage)
	}
}

func TestHandleErrorWithNilError(t *testing.T) {
	exitCode := GetExitCodeForError(nil)
	if exitCode != ExitCodeError {
		t.Errorf("GetExitCodeForError(nil) = %d, want %d", exitCode, ExitCodeError)
	}
}

func TestHandleErrorExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expectCode ExitCode
	}{
		{
			name:       "ParseError exits with code 3",
			err:        gerrors.NewParseError("file.yaml", "parse failed", nil),
			expectCode: ExitCodeParse,
		},
		{
			name:       "ValidationError exits with code 2",
			err:        gerrors.NewValidationError("", "name", "", "invalid field"),
			expectCode: ExitCodeUsage,
		},
		{
			name:       "ConfigError exits with code 2",
			err:        gerrors.NewConfigError("platform", "bad", "invalid").WithSuggestions([]string{"claude-code"}),
			expectCode: ExitCodeUsage,
		},
		{
			name:       "TransformError exits with code 1",
			err:        gerrors.NewTransformError("render", "opencode", "failed", nil),
			expectCode: ExitCodeError,
		},
		{
			name:       "FileError exits with code 1",
			err:        gerrors.NewFileError("file.yaml", "read", "not found", nil),
			expectCode: ExitCodeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig()

			formatted := cfg.ErrorFormatter.Format(tt.err)
			if formatted == "" {
				t.Error("ErrorFormatter.Format() should return non-empty string")
			}

			code := GetExitCodeForError(tt.err)
			if code != tt.expectCode {
				t.Errorf("GetExitCodeForError() = %d, want %d", code, tt.expectCode)
			}
		})
	}
}
