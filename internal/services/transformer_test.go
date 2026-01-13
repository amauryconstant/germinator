package services

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/models"
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

	errs, err := ValidateDocument(inputFile)
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

	errs, err := ValidateDocument(inputFile)
	if err != nil {
		t.Fatalf("ValidateDocument failed: %v", err)
	}

	if len(errs) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestValidateDocumentMissingFile(t *testing.T) {
	nonExistentFile := "/nonexistent/file.md"

	_, err := ValidateDocument(nonExistentFile)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestTransformAndRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input-agent.md")
	outputFile := filepath.Join(tmpDir, "output-agent.md")

	originalContent := `---
name: test-agent
description: A test agent
tools:
  - editor
  - bash
  - grep
model: sonnet
permissionMode: default
---
This is the agent content
It has multiple lines
And preserves markdown **formatting**
`

	if err := os.WriteFile(inputFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err := TransformDocument(inputFile, outputFile, "claude-code")
	if err != nil {
		t.Fatalf("TransformDocument failed: %v", err)
	}

	doc1, err := core.LoadDocument(inputFile)
	if err != nil {
		t.Fatalf("Failed to load original document: %v", err)
	}

	doc2, err := core.LoadDocument(outputFile)
	if err != nil {
		t.Fatalf("Failed to load transformed document: %v", err)
	}

	agent1, ok1 := doc1.(*models.Agent)
	agent2, ok2 := doc2.(*models.Agent)

	if !ok1 || !ok2 {
		t.Fatal("Documents are not of expected type")
	}

	if agent1.Name != agent2.Name {
		t.Errorf("Name mismatch: %q != %q", agent1.Name, agent2.Name)
	}
	if agent1.Description != agent2.Description {
		t.Errorf("Description mismatch: %q != %q", agent1.Description, agent2.Description)
	}
	if len(agent1.Tools) != len(agent2.Tools) {
		t.Errorf("Tools count mismatch: %d != %d", len(agent1.Tools), len(agent2.Tools))
	}
	if agent1.Content != agent2.Content {
		t.Errorf("Content mismatch:\nOriginal: %q\nGot:      %q", agent1.Content, agent2.Content)
	}
}
