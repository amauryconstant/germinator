package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/services"
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

	errs, err := services.ValidateDocument(validFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errs), errs)
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
			errorMsg:    "unknown platform: invalid-platform",
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
			if tt.expectError {
				err := services.TransformDocument(inputFile, outputFile, tt.platform)
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
					}
				}
			} else {
				err := services.TransformDocument(inputFile, outputFile, tt.platform)
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
			errs, err := services.ValidateDocument(validFile, tt.platform)
			if err != nil {
				t.Errorf("ValidateDocument failed for %s: %v", tt.platform, err)
			}

			if len(errs) != 0 {
				t.Errorf("Expected no validation errors for %s, got %d: %v", tt.platform, len(errs), errs)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

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

	errs, err := services.ValidateDocument(validFile, "claude-code")
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errs), errs)
	}
}

// CLI Integration Tests - Platform Flag Validation

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
			errorMsg:    "unknown platform: invalid-platform",
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
			errs, err := services.ValidateDocument(validFile, tt.platform)
			if err != nil {
				t.Errorf("ValidateDocument failed: %v", err)
			}
			if tt.expectError && len(errs) == 0 {
				t.Errorf("Expected validation errors for platform %s but got none", tt.platform)
			}
			if !tt.expectError && len(errs) > 0 {
				t.Errorf("Unexpected validation errors for platform %s: %v", tt.platform, errs)
			}
			if tt.errorMsg != "" && len(errs) > 0 {
				var errMsgs strings.Builder
				for _, e := range errs {
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
mode: invalid-mode
---
Test content`,
			platform:    "opencode",
			expectError: true,
			errorMsg:    "invalid mode",
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
			errorMsg:    "available: claude-code, opencode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.filename
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			errs, err := services.ValidateDocument(testFile, tt.platform)
			if err != nil {
				t.Fatalf("ValidateDocument failed: %v", err)
			}
			if tt.expectError && len(errs) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			if tt.expectError && len(errs) > 0 {
				var errMsgs strings.Builder
				for _, e := range errs {
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
				if !strings.Contains(output, "bash") || !strings.Contains(output, "read") {
					t.Errorf("Expected output to contain tools")
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
permissionMode: default
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

			err := services.TransformDocument(inputFile, outputFile, tt.platform)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("TransformDocument failed: %v", err)
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
		{
			name:     "invalid skill missing content",
			filename: "test-skill.md",
			content: `---
name: test-skill
description: A test skill
---
`,
			platform:    "opencode",
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := tmpDir + "/" + tt.filename

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			errs, err := services.ValidateDocument(testFile, tt.platform)
			if err != nil {
				t.Fatalf("ValidateDocument failed: %v", err)
			}

			if tt.expectError && len(errs) != tt.errorCount {
				t.Errorf("Expected %d errors, got %d: %v", tt.errorCount, len(errs), errs)
			}
			if !tt.expectError && len(errs) > 0 {
				t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
			}
		})
	}
}
