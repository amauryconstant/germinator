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
