package library

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// withRenameFunc installs a test-only rename function for
// atomicWriteFile and restores the original on test cleanup. The seam
// lets us inject syscall.EXDEV deterministically in environments where
// cross-filesystem setup is unreliable.
func withRenameFunc(t *testing.T, fn func(string, string) error) {
	t.Helper()
	prev := renameFunc
	renameFunc = fn
	t.Cleanup(func() { renameFunc = prev })
}

// TestAtomicWriteFile_EXDEV exercises the cross-filesystem fallback
// in atomicWriteFile. It overrides the rename seam to return
// syscall.EXDEV (mimicking a cross-device rename failure), then
// verifies that atomicWriteFile transparently falls back to copy+remove,
// leaving the target with the new content and no stale temp files.
func TestAtomicWriteFile_EXDEV(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "library.yaml")

	// First rename attempt returns EXDEV (cross-filesystem); second
	// attempt (after the fallback path triggered io.Copy + os.Remove)
	// shouldn't be reached, but the seam must still return nil for
	// atomicWriteFile to consider the operation successful.
	withRenameFunc(t, func(string, string) error {
		return syscall.EXDEV
	})

	if err := atomicWriteFile(targetPath, []byte("data\n"), 0o600); err != nil {
		require.NoError(t, err)
	}

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, "data\n", string(got), "target content mismatch")

	// Temp file should be cleaned up by the fallback path.
	tmpPath := targetPath + ".tmp"
	_, err = os.Stat(tmpPath)
	assert.True(t, errors.Is(err, os.ErrNotExist), "temp file should be removed after EXDEV fallback")
}

// TestAtomicWriteFile_HappyPath verifies that atomicWriteFile with
// the default rename behaves identically to the legacy temp+rename
// pattern: target is updated, temp is removed.
func TestAtomicWriteFile_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "library.yaml")

	if err := atomicWriteFile(targetPath, []byte("hello\n"), 0o600); err != nil {
		require.NoError(t, err)
	}

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got), "target content mismatch")

	tmpPath := targetPath + ".tmp"
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err), "temp file should be removed")
}

// TestAtomicWriteFile_RenameFailNoEXDEV verifies that non-EXDEV
// rename failures are surfaced (not swallowed) so callers see the
// underlying error.
func TestAtomicWriteFile_RenameFailNoEXDEV(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "library.yaml")

	wantErr := errors.New("simulated permission denied")
	withRenameFunc(t, func(string, string) error {
		return wantErr
	})

	if err := atomicWriteFile(targetPath, []byte("data\n"), 0o600); err == nil {
		require.Fail(t, "atomicWriteFile() expected error when rename fails non-EXDEV")
	}

	// Target should not exist (rename failed before any copy).
	_, statErr := os.Stat(targetPath)
	assert.True(t, os.IsNotExist(statErr), "target should not exist after non-EXDEV rename failure")
}

func TestSaveLibrary(t *testing.T) {
	// Create a temporary library directory
	tmpDir := t.TempDir()

	// Create minimal library structure
	lib := &Library{
		Version:  "1",
		RootPath: tmpDir,
		Resources: map[string]map[string]Resource{
			"skill": {
				"test": {Path: "skills/test.md", Description: "Test skill"},
			},
		},
		Presets: map[string]Preset{
			"test-preset": {
				Name:        "test-preset",
				Description: "Test preset",
				Resources:   []string{"skill/test"},
			},
		},
	}

	// Save the library
	if err := SaveLibrary(lib); err != nil {
		require.NoError(t, err)
	}

	// Verify file was created
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		require.Fail(t, "library.yaml was not created")
	}

	// Verify content by loading it back
	loadedLib, err := LoadLibrary(context.Background(), tmpDir)
	require.NoError(t, err)

	if loadedLib.Version != "1" {
		assert.Equal(t, 1, loadedLib.Version, "Version mismatch")
	}

	if len(loadedLib.Resources) != 1 {
		assert.Len(t, loadedLib.Resources, 1, "Resources count")
	}

	if len(loadedLib.Presets) != 1 {
		assert.Len(t, loadedLib.Presets, 1, "Presets count")
	}

	testPreset, ok := loadedLib.Presets["test-preset"]
	require.True(t, ok, "test-preset not found")
	if testPreset.Description != "Test preset" {
		assert.Equal(t, "Test preset", testPreset.Description, "Preset description mismatch")
	}
}

func TestSaveLibrary_EmptyLibrary(t *testing.T) {
	tmpDir := t.TempDir()

	lib := &Library{
		Version:   "1",
		RootPath:  tmpDir,
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	if err := SaveLibrary(lib); err != nil {
		require.NoError(t, err)
	}

	// Verify content
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	content, err := os.ReadFile(yamlPath)
	require.NoError(t, err)

	var parsed libraryYAML
	require.NoError(t, yaml.Unmarshal(content, &parsed))

	if parsed.Version != "1" {
		assert.Equal(t, 1, parsed.Version, "Version mismatch")
	}
}

func TestSaveLibrary_NoRootPath(t *testing.T) {
	lib := &Library{
		Version:  "1",
		RootPath: "",
	}

	err := SaveLibrary(lib)
	if err == nil {
		assert.Fail(t, "SaveLibrary() expected error for empty RootPath")
	}
}

func TestSaveLibrary_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "subdir", "library")

	lib := &Library{
		Version:   "1",
		RootPath:  nestedPath,
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	if err := SaveLibrary(lib); err != nil {
		require.NoError(t, err)
	}

	yamlPath := filepath.Join(nestedPath, "library.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		require.Fail(t, "library.yaml was not created in nested directory")
	}
}

func TestAddPreset(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	preset := Preset{
		Name:        "new-preset",
		Description: "A new preset",
		Resources:   []string{"skill/commit", "agent/reviewer"},
	}

	if err := AddPreset(lib, preset); err != nil {
		require.NoError(t, err)
	}

	if len(lib.Presets) != 1 {
		require.Fail(t, "Presets count != 1")
	}

	if !PresetExists(lib, "new-preset") {
		assert.Fail(t, "PresetExists() returned false for added preset")
	}

	addedPreset := lib.Presets["new-preset"]
	if addedPreset.Description != "A new preset" {
		assert.Equal(t, "A new preset", addedPreset.Description, "Description mismatch")
	}
	if len(addedPreset.Resources) != 2 {
		assert.Len(t, addedPreset.Resources, 2, "Resources count")
	}
}

func TestAddPreset_ValidationError(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	// Empty name should fail validation
	preset := Preset{
		Name:        "",
		Description: "Invalid preset",
		Resources:   []string{"skill/test"},
	}

	err := AddPreset(lib, preset)
	if err == nil {
		assert.Fail(t, "AddPreset() expected error for empty preset name")
	}
}

func TestAddPreset_EmptyResources(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	preset := Preset{
		Name:        "empty-preset",
		Description: "Preset with no resources",
		Resources:   []string{},
	}

	err := AddPreset(lib, preset)
	if err == nil {
		assert.Fail(t, "AddPreset() expected error for empty resources")
	}
}

func TestAddPreset_InvalidResourceRef(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	preset := Preset{
		Name:        "invalid-ref",
		Description: "Preset with invalid resource ref",
		Resources:   []string{"invalid-ref-format"},
	}

	err := AddPreset(lib, preset)
	if err == nil {
		assert.Fail(t, "AddPreset() expected error for invalid resource reference")
	}
}

func TestAddPreset_NilPresetsMap(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   nil, // Explicitly nil
	}

	preset := Preset{
		Name:        "test",
		Description: "Test",
		Resources:   []string{"skill/test"},
	}

	if err := AddPreset(lib, preset); err != nil {
		require.NoError(t, err)
	}

	if lib.Presets == nil {
		assert.Fail(t, "Presets map should be initialized")
	}
}

func TestPresetExists(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets: map[string]Preset{
			"existing": {Name: "existing", Description: "Exists", Resources: []string{"skill/test"}},
		},
	}

	tests := []struct {
		name     string
		preset   string
		expected bool
	}{
		{"existing preset", "existing", true},
		{"nonexistent preset", "nonexistent", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PresetExists(lib, tt.preset), "PresetExists mismatch")
		})
	}
}

func TestPresetExists_NilPresets(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   nil,
	}

	assert.False(t, PresetExists(lib, "any"), "PresetExists() should return false for nil Presets")
}

func TestPresetExists_NilLibrary(t *testing.T) {
	var lib *Library

	assert.False(t, PresetExists(lib, "any"), "PresetExists() should return false for nil Library")
}

func TestPresetExists_EmptyLibrary(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	assert.False(t, PresetExists(lib, "any"), "PresetExists() should return false for empty Presets")
}
