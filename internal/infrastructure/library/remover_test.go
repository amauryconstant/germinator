package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveResource_Success(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Add a resource first
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-test.md")
	srcContent := `---
name: to-remove
description: A skill to remove
tools:
  - bash
---
# Test Skill
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	if err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	// Verify resource exists
	lib, err := LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if _, exists := lib.Resources["skill"]["to-remove"]; !exists {
		t.Fatal("Resource should exist before removal")
	}

	// Remove the resource
	result, err := RemoveResource(RemoveResourceOptions{
		Ref:         "skill/to-remove",
		LibraryPath: tmpLibDir,
	})
	if err != nil {
		t.Fatalf("RemoveResource() error = %v", err)
	}

	if result.Type != "resource" {
		t.Errorf("Type = %v, want resource", result.Type)
	}
	if result.ResourceType != "skill" {
		t.Errorf("ResourceType = %v, want skill", result.ResourceType)
	}
	if result.Name != "to-remove" {
		t.Errorf("Name = %v, want to-remove", result.Name)
	}

	// Verify resource no longer exists in library
	lib, err = LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if _, exists := lib.Resources["skill"]["to-remove"]; exists {
		t.Error("Resource should have been removed from library")
	}

	// Verify physical file was deleted
	physicalPath := filepath.Join(tmpLibDir, "skills", "to-remove.md")
	if _, err := os.Stat(physicalPath); !os.IsNotExist(err) {
		t.Error("Physical file should have been deleted")
	}
}

func TestRemoveResource_NotFound(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemoveResource(RemoveResourceOptions{
		Ref:         "skill/nonexistent",
		LibraryPath: tmpLibDir,
	})
	if err == nil {
		t.Fatal("Expected error when resource not found")
	}
}

func TestRemoveResource_PresetReferenceConflict(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Add a resource first
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-conflict.md")
	srcContent := `---
name: conflict-skill
description: A skill
tools:
  - bash
---
# Test Skill
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	if err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	// Create a preset that references the resource
	lib, err := LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	err = AddPreset(lib, Preset{
		Name:        "test-preset",
		Description: "Test preset",
		Resources:   []string{"skill/conflict-skill"},
	})
	if err != nil {
		t.Fatalf("AddPreset() error = %v", err)
	}

	err = SaveLibrary(lib)
	if err != nil {
		t.Fatalf("SaveLibrary() error = %v", err)
	}

	// Try to remove the resource - should fail
	_, err = RemoveResource(RemoveResourceOptions{
		Ref:         "skill/conflict-skill",
		LibraryPath: tmpLibDir,
	})
	if err == nil {
		t.Fatal("Expected error when resource is referenced by preset")
	}
}

func TestRemoveResource_InvalidRefFormat(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	tests := []struct {
		name string
		ref  string
	}{
		{name: "no slash", ref: "skill"},
		{name: "empty name", ref: "skill/"},
		{name: "empty type", ref: "/commit"},
		{name: "too many parts", ref: "skill/commit/extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RemoveResource(RemoveResourceOptions{
				Ref:         tt.ref,
				LibraryPath: tmpLibDir,
			})
			if err == nil {
				t.Errorf("Expected error for ref %q", tt.ref)
			}
		})
	}
}

func TestRemovePreset_Success(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	// Add a resource first
	tmpSrcDir := t.TempDir()
	srcPath := filepath.Join(tmpSrcDir, "skill-test.md")
	srcContent := `---
name: preset-skill
description: A skill
tools:
  - bash
---
# Test Skill
`
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := AddResource(AddOptions{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	if err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	// Create a preset
	lib, err := LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	err = AddPreset(lib, Preset{
		Name:        "workflow-preset",
		Description: "Test preset",
		Resources:   []string{"skill/preset-skill"},
	})
	if err != nil {
		t.Fatalf("AddPreset() error = %v", err)
	}

	err = SaveLibrary(lib)
	if err != nil {
		t.Fatalf("SaveLibrary() error = %v", err)
	}

	// Verify preset exists
	lib, err = LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if !PresetExists(lib, "workflow-preset") {
		t.Fatal("Preset should exist before removal")
	}

	// Remove the preset
	result, err := RemovePreset(RemovePresetOptions{
		Name:        "workflow-preset",
		LibraryPath: tmpLibDir,
	})
	if err != nil {
		t.Fatalf("RemovePreset() error = %v", err)
	}

	if result.Type != "preset" {
		t.Errorf("Type = %v, want preset", result.Type)
	}
	if result.Name != "workflow-preset" {
		t.Errorf("Name = %v, want workflow-preset", result.Name)
	}
	if len(result.ResourcesRemoved) != 1 || result.ResourcesRemoved[0] != "skill/preset-skill" {
		t.Errorf("ResourcesRemoved = %v, want [skill/preset-skill]", result.ResourcesRemoved)
	}

	// Verify preset no longer exists
	lib, err = LoadLibrary(tmpLibDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if PresetExists(lib, "workflow-preset") {
		t.Error("Preset should have been removed")
	}

	// Verify resource still exists (presets are just references)
	if _, exists := lib.Resources["skill"]["preset-skill"]; !exists {
		t.Error("Resource should still exist after preset removal")
	}
}

func TestRemovePreset_NotFound(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemovePreset(RemovePresetOptions{
		Name:        "nonexistent-preset",
		LibraryPath: tmpLibDir,
	})
	if err == nil {
		t.Fatal("Expected error when preset not found")
	}
}

func TestRemovePreset_EmptyName(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemovePreset(RemovePresetOptions{
		Name:        "",
		LibraryPath: tmpLibDir,
	})
	if err == nil {
		t.Fatal("Expected error when preset name is empty")
	}
}
