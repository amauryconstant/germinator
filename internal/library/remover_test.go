package library

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gerrors "gitlab.com/amoconst/germinator/internal/core"
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
	require.NoError(t, os.WriteFile(srcPath, []byte(srcContent), 0644))

	err := AddResource(context.Background(), AddRequest{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	require.NoError(t, err)

	// Verify resource exists
	lib, err := LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)
	if _, exists := lib.Resources["skill"]["to-remove"]; !exists {
		require.Fail(t, "Resource should exist before removal")
	}

	// Remove the resource
	result, err := RemoveResource(context.Background(), RemoveResourceOptions{
		Ref:         "skill/to-remove",
		LibraryPath: tmpLibDir,
	})
	require.NoError(t, err)

	if result.Type != "resource" {
		assert.Equal(t, "resource", result.Type, "Type mismatch")
	}
	if result.ResourceType != "skill" {
		assert.Equal(t, "skill", result.ResourceType, "ResourceType mismatch")
	}
	if result.Name != "to-remove" {
		assert.Equal(t, "to-remove", result.Name, "Name mismatch")
	}

	// Verify resource no longer exists in library
	lib, err = LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)
	if _, exists := lib.Resources["skill"]["to-remove"]; exists {
		assert.Fail(t, "Resource should have been removed from library")
	}

	// Verify physical file was deleted
	physicalPath := filepath.Join(tmpLibDir, "skills", "to-remove.md")
	if _, err := os.Stat(physicalPath); !os.IsNotExist(err) {
		assert.Fail(t, "Physical file should have been deleted")
	}
}

func TestRemoveResource_NotFound(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemoveResource(context.Background(), RemoveResourceOptions{
		Ref:         "skill/nonexistent",
		LibraryPath: tmpLibDir,
	})
	require.Error(t, err)

	var nf *gerrors.NotFoundError
	require.True(t, errors.As(err, &nf),
		"missing-resource removal MUST surface as *core.NotFoundError (Phase 3 migration)")
	assert.Equal(t, "library ref", nf.Entity)
	assert.Equal(t, "skill/nonexistent", nf.Key)
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
	require.NoError(t, os.WriteFile(srcPath, []byte(srcContent), 0644))

	err := AddResource(context.Background(), AddRequest{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	require.NoError(t, err)

	// Create a preset that references the resource
	lib, err := LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)

	err = AddPreset(lib, Preset{
		Name:        "test-preset",
		Description: "Test preset",
		Resources:   []string{"skill/conflict-skill"},
	})
	require.NoError(t, err)

	err = SaveLibrary(lib)
	require.NoError(t, err)

	// Try to remove the resource - should fail
	_, err = RemoveResource(context.Background(), RemoveResourceOptions{
		Ref:         "skill/conflict-skill",
		LibraryPath: tmpLibDir,
	})
	require.Error(t, err)
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
			_, err := RemoveResource(context.Background(), RemoveResourceOptions{
				Ref:         tt.ref,
				LibraryPath: tmpLibDir,
			})
			if err == nil {
				require.Error(t, err, "Expected error for ref %q", tt.ref)
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
	require.NoError(t, os.WriteFile(srcPath, []byte(srcContent), 0644))

	err := AddResource(context.Background(), AddRequest{
		Source:      srcPath,
		LibraryPath: tmpLibDir,
	})
	require.NoError(t, err)

	// Create a preset
	lib, err := LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)

	err = AddPreset(lib, Preset{
		Name:        "workflow-preset",
		Description: "Test preset",
		Resources:   []string{"skill/preset-skill"},
	})
	require.NoError(t, err)

	err = SaveLibrary(lib)
	require.NoError(t, err)

	// Verify preset exists
	lib, err = LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)
	require.True(t, PresetExists(lib, "workflow-preset"), "Preset should exist before removal")

	// Remove the preset
	result, err := RemovePreset(context.Background(), RemovePresetOptions{
		Name:        "workflow-preset",
		LibraryPath: tmpLibDir,
	})
	require.NoError(t, err)

	if result.Type != "preset" {
		assert.Equal(t, "preset", result.Type, "Type mismatch")
	}
	if result.Name != "workflow-preset" {
		assert.Equal(t, "workflow-preset", result.Name, "Name mismatch")
	}
	if len(result.ResourcesRemoved) != 1 || result.ResourcesRemoved[0] != "skill/preset-skill" {
		assert.Equal(t, []string{"skill/preset-skill"}, result.ResourcesRemoved, "ResourcesRemoved mismatch")
	}

	// Verify preset no longer exists
	lib, err = LoadLibrary(context.Background(), tmpLibDir)
	require.NoError(t, err)
	assert.False(t, PresetExists(lib, "workflow-preset"), "Preset should have been removed")

	// Verify resource still exists (presets are just references)
	if _, exists := lib.Resources["skill"]["preset-skill"]; !exists {
		assert.Fail(t, "Resource should still exist after preset removal")
	}
}

func TestRemovePreset_NotFound(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemovePreset(context.Background(), RemovePresetOptions{
		Name:        "nonexistent-preset",
		LibraryPath: tmpLibDir,
	})
	require.Error(t, err)

	var nf *gerrors.NotFoundError
	require.True(t, errors.As(err, &nf),
		"missing-preset removal MUST surface as *core.NotFoundError (Phase 3 migration)")
	assert.Equal(t, "preset", nf.Entity)
	assert.Equal(t, "nonexistent-preset", nf.Key)
}

func TestRemovePreset_EmptyName(t *testing.T) {
	tmpLibDir := t.TempDir()
	createTestLibrary(t, tmpLibDir)

	_, err := RemovePreset(context.Background(), RemovePresetOptions{
		Name:        "",
		LibraryPath: tmpLibDir,
	})
	require.Error(t, err)
}
