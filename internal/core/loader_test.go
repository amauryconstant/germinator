package core

import (
	"errors"
	"os"
	"strings"
	"testing"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

func TestDetectTypeFromFilename(t *testing.T) {
	tests := []struct {
		name         string
		filepath     string
		expectedType string
	}{
		// Agent patterns - prefix
		{
			name:         "agent prefix with full name",
			filepath:     "agent-test-document.md",
			expectedType: "agent",
		},
		{
			name:         "agent prefix with simple name",
			filepath:     "agent-test.md",
			expectedType: "agent",
		},
		// Agent patterns - suffix
		{
			name:         "agent suffix with full name",
			filepath:     "test-document-agent.md",
			expectedType: "agent",
		},
		{
			name:         "agent suffix with simple name",
			filepath:     "test-agent.md",
			expectedType: "agent",
		},
		// Command patterns - prefix
		{
			name:         "command prefix with full name",
			filepath:     "command-test-document.md",
			expectedType: "command",
		},
		{
			name:         "command prefix with simple name",
			filepath:     "command-test.md",
			expectedType: "command",
		},
		// Command patterns - suffix
		{
			name:         "command suffix with full name",
			filepath:     "test-document-command.md",
			expectedType: "command",
		},
		{
			name:         "command suffix with simple name",
			filepath:     "test-command.md",
			expectedType: "command",
		},
		// Memory patterns - prefix
		{
			name:         "memory prefix with full name",
			filepath:     "memory-test-document.md",
			expectedType: "memory",
		},
		{
			name:         "memory prefix with simple name",
			filepath:     "memory-test.md",
			expectedType: "memory",
		},
		// Memory patterns - suffix
		{
			name:         "memory suffix with full name",
			filepath:     "test-document-memory.md",
			expectedType: "memory",
		},
		{
			name:         "memory suffix with simple name",
			filepath:     "test-memory.md",
			expectedType: "memory",
		},
		// Skill patterns - prefix
		{
			name:         "skill prefix with full name",
			filepath:     "skill-test-document.md",
			expectedType: "skill",
		},
		{
			name:         "skill prefix with simple name",
			filepath:     "skill-test.md",
			expectedType: "skill",
		},
		// Skill patterns - suffix
		{
			name:         "skill suffix with full name",
			filepath:     "test-document-skill.md",
			expectedType: "skill",
		},
		{
			name:         "skill suffix with simple name",
			filepath:     "test-skill.md",
			expectedType: "skill",
		},
		// Edge cases
		{
			name:         "unrecognizable filename",
			filepath:     "my-document.md",
			expectedType: "",
		},
		{
			name:         "no extension",
			filepath:     "agent-test",
			expectedType: "",
		},
		{
			name:         "wrong extension",
			filepath:     "agent-test.txt",
			expectedType: "",
		},
		{
			name:         "empty string",
			filepath:     "",
			expectedType: "",
		},
		{
			name:         "no hyphen or dash in name",
			filepath:     "agenttest.md",
			expectedType: "",
		},
		{
			name:         "path with directory",
			filepath:     "/path/to/agent-test.md",
			expectedType: "agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType := DetectType(tt.filepath)
			if docType != tt.expectedType {
				t.Errorf("DetectType(%q) = %v, want %v", tt.filepath, docType, tt.expectedType)
			}
		})
	}
}

func TestDetectTypeCaseSensitivity(t *testing.T) {
	tests := []struct {
		name         string
		filepath     string
		expectedType string
	}{
		{
			name:         "uppercase AGENT prefix not matched",
			filepath:     "AGENT-test.md",
			expectedType: "",
		},
		{
			name:         "uppercase AGENT suffix not matched",
			filepath:     "test-AGENT.md",
			expectedType: "",
		},
		{
			name:         "mixed case Agent prefix not matched",
			filepath:     "Agent-test.md",
			expectedType: "",
		},
		{
			name:         "mixed case Agent suffix not matched",
			filepath:     "test-Agent.md",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType := DetectType(tt.filepath)
			if docType != tt.expectedType {
				t.Errorf("DetectType(%q) = %v, want %v", tt.filepath, docType, tt.expectedType)
			}
		})
	}
}

func TestDetectTypeSpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		filepath     string
		expectedType string
	}{
		{
			name:         "agent with numbers",
			filepath:     "agent-123.md",
			expectedType: "agent",
		},
		{
			name:         "agent with underscores in name",
			filepath:     "agent-test_document.md",
			expectedType: "agent",
		},
		{
			name:         "agent with multiple hyphens",
			filepath:     "agent-test-document-name.md",
			expectedType: "agent",
		},
		{
			name:         "agent with dots",
			filepath:     "agent.test.md",
			expectedType: "",
		},
		{
			name:         "agent with spaces",
			filepath:     "agent test.md",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType := DetectType(tt.filepath)
			if docType != tt.expectedType {
				t.Errorf("DetectType(%q) = %v, want %v", tt.filepath, docType, tt.expectedType)
			}
		})
	}
}

func TestLoadDocumentEmptyPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-agent.md"

	content := `---
name: test-agent
description: A test agent
---
Content`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		platform string
		wantErr  bool
	}{
		{
			name:     "empty platform",
			platform: "",
			wantErr:  true,
		},
		{
			name:     "claude-code platform",
			platform: "claude-code",
			wantErr:  false,
		},
		{
			name:     "opencode platform",
			platform: "opencode",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := LoadDocument(testFile, tt.platform)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadDocument() expected error for platform %q, got nil", tt.platform)
				}
				return
			}

			if err != nil {
				t.Errorf("LoadDocument() unexpected error for platform %q: %v", tt.platform, err)
				return
			}

			if doc == nil {
				t.Errorf("LoadDocument() returned nil document")
			}
		})
	}
}

func TestLoadDocumentUnrecognizableFilename(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/unrecognizable.md"

	content := `---
name: test
description: Test
---
Content`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	doc, err := LoadDocument(testFile, "claude-code")

	if err == nil {
		t.Errorf("LoadDocument() expected error for unrecognizable filename, got nil")
	}

	if doc != nil {
		t.Errorf("LoadDocument() expected nil document for error, got %v", doc)
	}

	if err != nil && err.Error() == "" {
		t.Errorf("LoadDocument() error message should not be empty")
	}
}

func TestLoadDocumentInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-agent.md"

	content := `---
invalid yaml content
---
Content`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	doc, err := LoadDocument(testFile, "claude-code")

	if err == nil {
		t.Errorf("LoadDocument() expected error for invalid YAML, got nil")
	}

	if doc != nil {
		t.Errorf("LoadDocument() expected nil document for error, got %v", doc)
	}
}

func TestLoadDocumentValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-agent.md"

	content := `---
description: Missing name field
---
Content`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	doc, err := LoadDocument(testFile, "claude-code")

	if err == nil {
		t.Errorf("LoadDocument() expected error for validation failure, got nil")
	}

	if doc == nil {
		t.Errorf("LoadDocument() expected document even with validation errors, got nil")
	}

	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "validation error") {
			t.Errorf("LoadDocument() error should mention validation error, got: %s", errMsg)
		}
	}
}

func TestLoadDocumentAllDocumentTypes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "agent document",
			filename: "agent-test.md",
			content: `---
name: test-agent
description: Test agent
---
Agent content`,
		},
		{
			name:     "command document",
			filename: "command-test.md",
			content: `---
name: test-command
description: Test command
template: test template
---
Command content`,
		},
		{
			name:     "memory document",
			filename: "memory-test.md",
			content: `---
paths:
  - README.md
---
Memory content`,
		},
		{
			name:     "skill document",
			filename: "skill-test.md",
			content: `---
name: test-skill
description: Test skill
---
Skill content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := tmpDir + "/" + tt.filename

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			doc, err := LoadDocument(testFile, "claude-code")
			if err != nil {
				t.Errorf("LoadDocument() failed for %s: %v", tt.name, err)
			}

			if doc == nil {
				t.Errorf("LoadDocument() returned nil document for %s", tt.name)
			}
		})
	}
}

func TestLoadDocumentReturnsTypedParseError(t *testing.T) {
	t.Run("unrecognizable filename returns ParseError", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := tmpDir + "/unrecognizable.md"

		content := `---
name: test
description: Test
---
Content`

		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := LoadDocument(testFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for unrecognizable filename")
		}

		var parseErr *gerrors.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if parseErr.Path != testFile {
				t.Errorf("ParseError.Path = %q, want %q", parseErr.Path, testFile)
			}
			if !strings.Contains(parseErr.Message, "expected") {
				t.Errorf("ParseError.Message should mention expected patterns, got: %q", parseErr.Message)
			}
		}
	})

	t.Run("invalid YAML returns ParseError", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := tmpDir + "/test-agent.md"

		invalidYAML := `---
name: "unclosed string
---
Content`

		if err := os.WriteFile(testFile, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := LoadDocument(testFile, "claude-code")

		if err == nil {
			t.Fatal("Expected error for invalid YAML")
		}

		var parseErr *gerrors.ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got %T: %v", err, err)
		} else {
			if parseErr.Path != testFile {
				t.Errorf("ParseError.Path = %q, want %q", parseErr.Path, testFile)
			}
			if parseErr.Cause == nil {
				t.Error("ParseError.Cause should not be nil for YAML errors")
			}
		}
	})
}

func TestLoadDocumentReturnsTypedFileError(t *testing.T) {
	t.Run("file not found returns FileError", func(t *testing.T) {
		nonExistentFile := "/nonexistent/path/test-agent.md"

		_, err := LoadDocument(nonExistentFile, "claude-code")

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
}

func TestLoadDocumentReturnsTypedValidationError(t *testing.T) {
	t.Run("missing required field returns ValidationError", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := tmpDir + "/test-agent.md"

		content := `---
description: Missing name field
---
Content`

		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		doc, err := LoadDocument(testFile, "claude-code")

		if err == nil {
			t.Fatal("Expected validation error")
		}

		var validationErr *gerrors.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected ValidationError, got %T: %v", err, err)
		}

		if doc == nil {
			t.Error("Document should be returned even with validation errors")
		}
	})
}
