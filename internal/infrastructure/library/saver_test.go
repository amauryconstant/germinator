package library

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

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
		t.Fatalf("SaveLibrary() error = %v", err)
	}

	// Verify file was created
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("library.yaml was not created")
	}

	// Verify content by loading it back
	loadedLib, err := LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}

	if loadedLib.Version != "1" {
		t.Errorf("Version = %v, want 1", loadedLib.Version)
	}

	if len(loadedLib.Resources) != 1 {
		t.Errorf("Resources count = %d, want 1", len(loadedLib.Resources))
	}

	if len(loadedLib.Presets) != 1 {
		t.Errorf("Presets count = %d, want 1", len(loadedLib.Presets))
	}

	testPreset, ok := loadedLib.Presets["test-preset"]
	if !ok {
		t.Fatal("test-preset not found")
	}
	if testPreset.Description != "Test preset" {
		t.Errorf("Preset description = %v, want 'Test preset'", testPreset.Description)
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
		t.Fatalf("SaveLibrary() error = %v", err)
	}

	// Verify content
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Failed to read library.yaml: %v", err)
	}

	var parsed libraryYAML
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if parsed.Version != "1" {
		t.Errorf("Version = %v, want 1", parsed.Version)
	}
}

func TestSaveLibrary_NoRootPath(t *testing.T) {
	lib := &Library{
		Version:  "1",
		RootPath: "",
	}

	err := SaveLibrary(lib)
	if err == nil {
		t.Error("SaveLibrary() expected error for empty RootPath")
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
		t.Fatalf("SaveLibrary() error = %v", err)
	}

	yamlPath := filepath.Join(nestedPath, "library.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatal("library.yaml was not created in nested directory")
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
		t.Fatalf("AddPreset() error = %v", err)
	}

	if len(lib.Presets) != 1 {
		t.Fatalf("Presets count = %d, want 1", len(lib.Presets))
	}

	if !PresetExists(lib, "new-preset") {
		t.Error("PresetExists() returned false for added preset")
	}

	addedPreset := lib.Presets["new-preset"]
	if addedPreset.Description != "A new preset" {
		t.Errorf("Description = %v, want 'A new preset'", addedPreset.Description)
	}
	if len(addedPreset.Resources) != 2 {
		t.Errorf("Resources count = %d, want 2", len(addedPreset.Resources))
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
		t.Error("AddPreset() expected error for empty preset name")
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
		t.Error("AddPreset() expected error for empty resources")
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
		t.Error("AddPreset() expected error for invalid resource reference")
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
		t.Fatalf("AddPreset() error = %v", err)
	}

	if lib.Presets == nil {
		t.Error("Presets map should be initialized")
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
			if got := PresetExists(lib, tt.preset); got != tt.expected {
				t.Errorf("PresetExists(%q) = %v, want %v", tt.preset, got, tt.expected)
			}
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

	if PresetExists(lib, "any") {
		t.Error("PresetExists() should return false for nil Presets")
	}
}

func TestPresetExists_NilLibrary(t *testing.T) {
	var lib *Library

	if PresetExists(lib, "any") {
		t.Error("PresetExists() should return false for nil Library")
	}
}

func TestPresetExists_EmptyLibrary(t *testing.T) {
	lib := &Library{
		Version:   "1",
		RootPath:  "/tmp/test",
		Resources: make(map[string]map[string]Resource),
		Presets:   make(map[string]Preset),
	}

	if PresetExists(lib, "any") {
		t.Error("PresetExists() should return false for empty Presets")
	}
}
