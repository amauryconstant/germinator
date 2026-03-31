package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectType(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		flag    string
		want    string
		wantErr bool
	}{
		{
			name:    "flag takes precedence",
			source:  "/path/to/some-file.md",
			flag:    "agent",
			want:    "agent",
			wantErr: false,
		},
		{
			name:    "invalid flag type",
			source:  "/path/to/some-file.md",
			flag:    "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "frontmatter type",
			source:  "/path/to/skill-test.md",
			flag:    "",
			want:    "skill",
			wantErr: false,
		},
		{
			name:    "filename agent pattern",
			source:  "/path/to/agent-reviewer.md",
			flag:    "",
			want:    "agent",
			wantErr: false,
		},
		{
			name:    "filename skill pattern",
			source:  "/path/to/skill-test.md",
			flag:    "",
			want:    "skill",
			wantErr: false,
		},
		{
			name:    "filename command pattern",
			source:  "/path/to/command-test.md",
			flag:    "",
			want:    "command",
			wantErr: false,
		},
		{
			name:    "filename memory pattern",
			source:  "/path/to/memory-test.md",
			flag:    "",
			want:    "memory",
			wantErr: false,
		},
		{
			name:    "cannot detect type",
			source:  "/path/to/test.md",
			flag:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with appropriate content
			tmpDir := t.TempDir()
			sourcePath := filepath.Join(tmpDir, filepath.Base(tt.source))

			var content string
			if tt.flag == "agent" || tt.want == "agent" {
				content = "---\ntype: agent\nname: test\ndescription: Test\n---\n"
			} else if tt.flag == "skill" || tt.want == "skill" {
				content = "---\ntype: skill\nname: test\ndescription: Test\n---\n"
			} else if tt.flag == "command" || tt.want == "command" {
				content = "---\ntype: command\nname: test\ndescription: Test\n---\n"
			} else if tt.flag == "memory" || tt.want == "memory" {
				content = "---\ntype: memory\nname: test\ndescription: Test\n---\n"
			} else if tt.flag != "" && tt.flag != "invalid" {
				content = "---\nname: test\ndescription: Test\n---\n"
			} else {
				content = "---\nname: test\ndescription: Test\n---\n"
			}
			if err := os.WriteFile(sourcePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got, err := detectType(sourcePath, tt.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectName(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		flag    string
		want    string
		wantErr bool
	}{
		{
			name:    "flag takes precedence",
			source:  "/path/to/skill-test.md",
			flag:    "my-name",
			want:    "my-name",
			wantErr: false,
		},
		{
			name:    "frontmatter name",
			source:  "/path/to/skill-test.md",
			flag:    "",
			want:    "commit-skill",
			wantErr: false,
		},
		{
			name:    "filename extraction",
			source:  "/path/to/skill-test.md",
			flag:    "",
			want:    "test",
			wantErr: false,
		},
		{
			name:    "filename extraction without type prefix",
			source:  "/path/to/skill-unknown.md",
			flag:    "",
			want:    "unknown",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			// Use basename from source for the filename
			baseName := filepath.Base(tt.source)
			sourcePath := filepath.Join(tmpDir, baseName)

			var content string
			if tt.flag != "" {
				content = "---\nname: frontmatter-name\ndescription: Test\n---\n"
			} else if tt.want == "commit-skill" {
				content = "---\nname: commit-skill\ndescription: Test\n---\n"
			} else {
				content = "---\ndescription: Test\n---\n"
			}
			if err := os.WriteFile(sourcePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got, err := detectName(sourcePath, tt.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectDescription(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		flag    string
		want    string
		wantErr bool
	}{
		{
			name:    "flag takes precedence",
			source:  "/path/to/test.md",
			flag:    "My description",
			want:    "My description",
			wantErr: false,
		},
		{
			name:    "frontmatter description",
			source:  "/path/to/test.md",
			flag:    "",
			want:    "Frontmatter description",
			wantErr: false,
		},
		{
			name:    "no description",
			source:  "/path/to/test.md",
			flag:    "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			sourcePath := filepath.Join(tmpDir, "test.md")

			var content string
			if tt.flag != "" {
				content = "---\nname: test\ndescription: Should be ignored\n---\n"
			} else if tt.want != "" {
				content = "---\nname: test\ndescription: Frontmatter description\n---\n"
			} else {
				content = "---\nname: test\n---\n"
			}
			if err := os.WriteFile(sourcePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got := detectDescription(sourcePath, tt.flag)
			if got != tt.want {
				t.Errorf("detectDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		content string
		want    string
	}{
		{
			name:    "opencode platform in frontmatter",
			source:  "/path/to/agent-test.md",
			content: "---\nplatform: opencode\nname: test\ndescription: Test\n---\n",
			want:    "opencode",
		},
		{
			name:    "claude-code platform in frontmatter",
			source:  "/path/to/agent-test.md",
			content: "---\nplatform: claude-code\nname: test\ndescription: Test\n---\n",
			want:    "claude-code",
		},
		{
			name:    "opencode in filename",
			source:  "/path/to/agent-opencode-test.md",
			content: "---\nname: test\ndescription: Test\n---\n",
			want:    "opencode",
		},
		{
			name:    "claude-code in filename",
			source:  "/path/to/agent-claude-code-test.md",
			content: "---\nname: test\ndescription: Test\n---\n",
			want:    "claude-code",
		},
		{
			name:    "no platform detected",
			source:  "/path/to/agent-test.md",
			content: "---\nname: test\ndescription: Test\n---\n",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			sourcePath := filepath.Join(tmpDir, filepath.Base(tt.source))
			if err := os.WriteFile(sourcePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got := DetectPlatform(sourcePath)
			if got != tt.want {
				t.Errorf("DetectPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCanonicalFormat(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		content string
		docType string
		want    bool
	}{
		{
			name:    "canonical skill format",
			source:  "/path/to/skill-test.md",
			content: "---\nname: test\ndescription: Test\ntools:\n  - bash\n---\n",
			docType: "skill",
			want:    true,
		},
		{
			name:    "opencode skill format with allowed-tools",
			source:  "/path/to/skill-test.md",
			content: "---\nname: test\ndescription: Test\nallowed-tools:\n  - bash\n---\n",
			docType: "skill",
			want:    false,
		},
		{
			name:    "canonical agent format",
			source:  "/path/to/agent-test.md",
			content: "---\nname: test\ndescription: Test\ntools:\n  - bash\n---\n",
			docType: "agent",
			want:    true,
		},
		{
			name:    "opencode agent format with mode",
			source:  "/path/to/agent-test.md",
			content: "---\nname: test\ndescription: Test\nmode: primary\n---\n",
			docType: "agent",
			want:    false,
		},
		{
			name:    "claude-code agent format with permissionMode",
			source:  "/path/to/agent-test.md",
			content: "---\nname: test\ndescription: Test\npermissionMode: default\n---\n",
			docType: "agent",
			want:    false,
		},
		{
			name:    "missing name",
			source:  "/path/to/test.md",
			content: "---\ndescription: Test\n---\n",
			docType: "skill",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			sourcePath := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(sourcePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got := IsCanonicalFormat(sourcePath, tt.docType)
			if got != tt.want {
				t.Errorf("IsCanonicalFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddResource_DryRun(t *testing.T) {
	// Create a temp library
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Create a source file
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-test.md")
	srcContent := `---
name: test-skill
description: A test skill
tools:
  - bash
---
# Test Skill
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run AddResource with DryRun
	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("AddResource() error = %v", err)
	}

	// Verify library.yaml was NOT modified
	lib, err := LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if _, exists := lib.Resources["skill"]["test-skill"]; exists {
		t.Error("Dry-run should not have added resource to library")
	}
}

func TestAddResource_Success(t *testing.T) {
	// Create a temp library
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Create a source file
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-test.md")
	srcContent := `---
name: new-skill
description: A new skill
tools:
  - bash
---
# New Skill
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run AddResource
	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
		DryRun:      false,
	})
	if err != nil {
		t.Fatalf("AddResource() error = %v", err)
	}

	// Verify library was updated
	lib, err := LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if _, exists := lib.Resources["skill"]["new-skill"]; !exists {
		t.Error("Resource should have been added to library")
	}

	// Verify file was copied
	expectedPath := filepath.Join(tmpLibDir, "skills", "new-skill.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Resource file should exist at %s", expectedPath)
	}
}

func TestAddResource_ForceOverwrite(t *testing.T) {
	// Create a temp library
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Create an initial source file
	tmpSrcDir := t.TempDir()
	initialSrcPath := filepath.Join(tmpSrcDir, "skill-existing.md")
	initialContent := `---
name: existing
description: Existing skill
tools:
  - bash
---
# Existing
`
	if err := os.WriteFile(initialSrcPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial source file: %v", err)
	}

	// Add the initial resource
	err := AddResource(AddOptions{
		Source:      initialSrcPath,
		LibraryPath: tmpLibDir,
		Force:       false,
	})
	if err != nil {
		t.Fatalf("Initial AddResource() error = %v", err)
	}

	// Create a new source file with same name but different content
	newSrcPath := filepath.Join(tmpSrcDir, "skill-existing-2.md")
	newContent := `---
name: existing
description: Updated description
tools:
  - read
---
# Updated
`
	if err := os.WriteFile(newSrcPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to write new source file: %v", err)
	}

	// Try to add without force - should fail
	err = AddResource(AddOptions{
		Source:      newSrcPath,
		LibraryPath: tmpLibDir,
		Force:       false,
	})
	if err == nil {
		t.Error("Expected error when resource exists without force flag")
	}

	// Add with force - should succeed
	err = AddResource(AddOptions{
		Source:      newSrcPath,
		LibraryPath: tmpLibDir,
		Force:       true,
	})
	if err != nil {
		t.Fatalf("AddResource() with force error = %v", err)
	}

	// Verify content was updated
	existingPath := filepath.Join(tmpLibDir, "skills", "existing.md")
	content, _ := os.ReadFile(existingPath)
	if !contains(string(content), "Updated description") {
		t.Error("Resource content should have been updated")
	}
}

func TestAddResource_ConflictDetection(t *testing.T) {
	// Create a temp library
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Create a source file
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-conflict.md")
	srcContent := `---
name: conflict
description: A skill
tools:
  - bash
---
# Conflict
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// First add should succeed
	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
		Force:       false,
	})
	if err != nil {
		t.Fatalf("First AddResource() error = %v", err)
	}

	// Second add without force should fail
	err = AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
		Force:       false,
	})
	if err == nil {
		t.Error("Expected error when adding duplicate resource without force")
	}
}

// Helper function to create a test library
func createTestLibrary(t *testing.T, tmpDir string) {
	t.Helper()

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "skills"), 0750); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "agents"), 0750); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "commands"), 0750); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "memory"), 0750); err != nil {
		t.Fatalf("Failed to create memory directory: %v", err)
	}

	// Create library.yaml
	yamlContent := `version: "1"
resources:
  skill: {}
  agent: {}
  command: {}
  memory: {}
presets: {}
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write library.yaml: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
