package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/services"
)

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
			name:        "empty platform flag",
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
				t.Fatalf("ValidateDocument failed: %v", err)
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
