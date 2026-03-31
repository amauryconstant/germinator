package cmd

import (
	"bytes"
	"testing"

	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

func TestCreatePresetCommand_Success(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	// Create a temporary library
	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add a resource to the library first
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources == nil {
		lib.Resources = make(map[string]map[string]library.Resource)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["commit"] = library.Resource{
		Path:        "skills/commit.md",
		Description: "Git commit skill",
	}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test-preset", "--resources", "skill/commit", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify preset was created
	loadedLib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}

	if !library.PresetExists(loadedLib, "test-preset") {
		t.Error("Preset was not created")
	}

	preset, ok := loadedLib.Presets["test-preset"]
	if !ok {
		t.Fatal("Preset not found in library")
	}
	if preset.Description != "" {
		t.Errorf("Expected empty description, got %s", preset.Description)
	}
	if len(preset.Resources) != 1 || preset.Resources[0] != "skill/commit" {
		t.Errorf("Expected resources [skill/commit], got %v", preset.Resources)
	}
}

func TestCreatePresetCommand_WithDescription(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add resources
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["test"] = library.Resource{Path: "skills/test.md"}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "my-preset", "--resources", "skill/test", "--description", "My test preset", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify preset
	loadedLib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}

	preset := loadedLib.Presets["my-preset"]
	if preset.Description != "My test preset" {
		t.Errorf("Expected description 'My test preset', got %s", preset.Description)
	}
}

func TestCreatePresetCommand_AlreadyExistsError(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add resources and existing preset
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["test"] = library.Resource{Path: "skills/test.md"}
	lib.Presets["existing"] = library.Preset{Name: "existing", Resources: []string{"skill/test"}}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "existing", "--resources", "skill/test", "--library", tmpDir})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error when preset already exists")
	}
}

func TestCreatePresetCommand_ForceOverwrite(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add resources and existing preset
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["test"] = library.Resource{Path: "skills/test.md"}
	if lib.Resources["agent"] == nil {
		lib.Resources["agent"] = make(map[string]library.Resource)
	}
	lib.Resources["agent"]["new"] = library.Resource{Path: "agents/new.md"}
	lib.Presets["existing"] = library.Preset{Name: "existing", Description: "Old description", Resources: []string{"skill/test"}}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "existing", "--resources", "agent/new", "--description", "New description", "--force", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify preset was overwritten
	loadedLib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}

	preset := loadedLib.Presets["existing"]
	if preset.Description != "New description" {
		t.Errorf("Expected description 'New description', got %s", preset.Description)
	}
	if len(preset.Resources) != 1 || preset.Resources[0] != "agent/new" {
		t.Errorf("Expected resources [agent/new], got %v", preset.Resources)
	}
}

func TestCreatePresetCommand_ResourceNotFound(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test", "--resources", "skill/nonexistent", "--library", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when resource doesn't exist")
	}
}

func TestCreatePresetCommand_InvalidResourceFormat(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test", "--resources", "invalid-format", "--library", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid resource format")
	}
}

func TestCreatePresetCommand_MultipleResources(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add resources
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	if lib.Resources["agent"] == nil {
		lib.Resources["agent"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["commit"] = library.Resource{Path: "skills/commit.md"}
	lib.Resources["agent"]["reviewer"] = library.Resource{Path: "agents/reviewer.md"}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "multi", "--resources", "skill/commit,agent/reviewer", "--library", tmpDir})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify preset
	loadedLib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}

	preset := loadedLib.Presets["multi"]
	if len(preset.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(preset.Resources))
	}
}

func TestCreatePresetCommand_MissingResourcesFlag(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test", "--library", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when --resources flag is missing")
	}
}

func TestCreatePresetCommand_WhitespaceName(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "   ", "--resources", "skill/test", "--library", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for whitespace-only name")
	}
}

func TestCreatePresetCommand_ResourceTypeNotFound(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	tmpDir := t.TempDir()
	if err := library.CreateLibrary(library.CreateOptions{Path: tmpDir, Force: true}); err != nil {
		t.Fatalf("Failed to create test library: %v", err)
	}

	// Add a skill resource but not an agent
	lib, err := library.LoadLibrary(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load library: %v", err)
	}
	if lib.Resources["skill"] == nil {
		lib.Resources["skill"] = make(map[string]library.Resource)
	}
	lib.Resources["skill"]["test"] = library.Resource{Path: "skills/test.md"}
	if err := library.SaveLibrary(lib); err != nil {
		t.Fatalf("Failed to save library: %v", err)
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test", "--resources", "agent/nonexistent", "--library", tmpDir})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error when resource type doesn't exist in library")
	}
}

func TestCreatePresetCommand_LibraryNotFound(t *testing.T) {
	cfg := &CommandConfig{
		Services:       NewServiceContainer(),
		ErrorFormatter: NewErrorFormatter(),
	}

	cmd := NewLibraryCommand(cfg)
	cmd.SetArgs([]string{"create", "preset", "test", "--resources", "skill/test", "--library", "/nonexistent/path"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when library doesn't exist")
	}
}
