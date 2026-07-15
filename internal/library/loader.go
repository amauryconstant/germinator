package library

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
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
// It expects a library.yaml file in the directory. The provided ctx is
// checked between I/O operations; on cancellation, the partial result
// is discarded and the function returns ctx.Err() wrapped with context.
func LoadLibrary(ctx context.Context, path string) (*Library, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, gerrors.NewNotFoundError("library", path)
		}
		return nil, gerrors.NewFileError(path, "access", "failed to access library", err)
	}
	if !info.IsDir() {
		return nil, gerrors.NewFileError(path, "access", "path is not a directory", nil)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	// Read library.yaml
	yamlPath := filepath.Join(path, "library.yaml")
	yamlContent, err := os.ReadFile(yamlPath) //nolint:gosec // G304: User provides library path, must read fixed library.yaml from it
	if err != nil {
		if os.IsNotExist(err) {
			return nil, gerrors.NewNotFoundError("library.yaml", yamlPath)
		}
		return nil, gerrors.NewFileError(yamlPath, "read", "failed to read library.yaml", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	// Parse YAML
	var libYAML libraryYAML
	if err := yaml.Unmarshal(yamlContent, &libYAML); err != nil {
		return nil, gerrors.NewParseError(yamlPath, "failed to parse library.yaml", err)
	}

	// Validate version
	if libYAML.Version == "" {
		return nil, gerrors.NewConfigError("version", "", "library.yaml missing version field")
	}
	if libYAML.Version != SupportedVersion {
		return nil, gerrors.NewConfigError("version", libYAML.Version, fmt.Sprintf("unsupported library version (expected %s)", SupportedVersion))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
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
			return nil, gerrors.NewConfigError("resource-type", typ, "invalid resource type")
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
