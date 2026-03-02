package library

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// SupportedVersion is the only supported library.yaml version.
const SupportedVersion = "1"

// libraryYAML is the internal structure for YAML parsing.
type libraryYAML struct {
	Version   string                         `yaml:"version"`
	Resources map[string]map[string]Resource `yaml:"resources"`
	Presets   map[string]Preset              `yaml:"presets"`
}

// LoadLibrary loads a library from the given directory path.
// It expects a library.yaml file in the directory.
func LoadLibrary(path string) (*Library, error) {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("library not found: %s", path)
		}
		return nil, fmt.Errorf("failed to access library: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("library path is not a directory: %s", path)
	}

	// Read library.yaml
	yamlPath := filepath.Join(path, "library.yaml")
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("library.yaml not found: %s", yamlPath)
		}
		return nil, fmt.Errorf("failed to read library.yaml: %w", err)
	}

	// Parse YAML
	var libYAML libraryYAML
	if err := yaml.Unmarshal(yamlContent, &libYAML); err != nil {
		return nil, fmt.Errorf("failed to parse library.yaml: %w", err)
	}

	// Validate version
	if libYAML.Version == "" {
		return nil, fmt.Errorf("library.yaml missing version field")
	}
	if libYAML.Version != SupportedVersion {
		return nil, fmt.Errorf("unsupported library version: %s (expected %s)", libYAML.Version, SupportedVersion)
	}

	// Create library
	lib := &Library{
		Version:   libYAML.Version,
		RootPath:  path,
		Resources: libYAML.Resources,
		Presets:   libYAML.Presets,
	}

	// Initialize empty maps if nil
	if lib.Resources == nil {
		lib.Resources = make(map[string]map[string]Resource)
	}
	if lib.Presets == nil {
		lib.Presets = make(map[string]Preset)
	}

	// Validate resources
	for typ, resources := range lib.Resources {
		// Validate type
		resourceType := ResourceType(typ)
		if !resourceType.IsValid() {
			return nil, fmt.Errorf("invalid resource type: %s", typ)
		}

		// Validate each resource
		for name, res := range resources {
			if err := res.Validate(); err != nil {
				return nil, fmt.Errorf("invalid resource %s/%s: %w", typ, name, err)
			}
		}
	}

	// Validate presets
	for name, preset := range lib.Presets {
		// Ensure name matches key
		preset.Name = name
		lib.Presets[name] = preset

		if err := preset.Validate(); err != nil {
			return nil, fmt.Errorf("invalid preset %s: %w", name, err)
		}
	}

	return lib, nil
}
