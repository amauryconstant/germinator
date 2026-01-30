package main

import (
	"os"
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
					if !contains(err.Error(), tt.errorMsg) {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
