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
