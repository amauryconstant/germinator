// Package library provides library management for canonical resources.
package library

import (
	"fmt"
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

// RemoveResourceOptions contains options for removing a resource from the library.
type RemoveResourceOptions struct {
	Ref         string
	LibraryPath string
	JSON        bool
}

// RemovePresetOptions contains options for removing a preset from the library.
type RemovePresetOptions struct {
	Name        string
	LibraryPath string
	JSON        bool
}

// RemoveResourceOutput represents the JSON output for a successful resource removal.
type RemoveResourceOutput struct {
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Name         string `json:"name"`
	FileDeleted  string `json:"fileDeleted"`
	LibraryPath  string `json:"libraryPath"`
}

// RemovePresetOutput represents the JSON output for a successful preset removal.
type RemovePresetOutput struct {
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	ResourcesRemoved []string `json:"resourcesRemoved"`
}

// RemoveResourceError represents a structured error when --json is used.
type RemoveResourceError struct {
	Error        string `json:"error"`
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Name         string `json:"name"`
}

// RemovePresetError represents a structured error when --json is used.
type RemovePresetError struct {
	Error string `json:"error"`
	Type  string `json:"type"`
	Name  string `json:"name"`
}

// RemoveResource removes a resource from the library.
// It deletes both the physical file and the YAML entry.
func RemoveResource(opts RemoveResourceOptions) (*RemoveResourceOutput, error) {
	typ, name, err := ParseRef(opts.Ref)
	if err != nil {
		return nil, fmt.Errorf("parsing reference: %w", err)
	}

	resourceType := ResourceType(typ)
	if !resourceType.IsValid() {
		return nil, gerrors.NewConfigError("type", typ, "invalid resource type")
	}

	lib, err := LoadLibrary(opts.LibraryPath)
	if err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	typeMap, typeExists := lib.Resources[typ]
	if !typeExists {
		return nil, gerrors.NewFileError(opts.LibraryPath, "access",
			fmt.Sprintf("resource %s not found", opts.Ref), nil)
	}
	resource, nameExists := typeMap[name]
	if !nameExists {
		return nil, gerrors.NewFileError(opts.LibraryPath, "access",
			fmt.Sprintf("resource %s not found", opts.Ref), nil)
	}

	for presetName, preset := range lib.Presets {
		for _, resRef := range preset.Resources {
			if resRef == opts.Ref {
				return nil, gerrors.NewFileError(opts.LibraryPath, "remove",
					fmt.Sprintf("cannot remove resource %s: it is referenced by preset %s (remove preset first)", opts.Ref, presetName), nil)
			}
		}
	}

	physicalPath := filepath.Join(opts.LibraryPath, resource.Path)

	if err := os.Remove(physicalPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, gerrors.NewFileError(physicalPath, "remove", "failed to delete resource file", err)
		}
	}

	if err := removeResourceFromLibrary(opts.LibraryPath, typ, name); err != nil {
		return nil, fmt.Errorf("updating library.yaml: %w", err)
	}

	if _, err := LoadLibrary(opts.LibraryPath); err != nil {
		return nil, fmt.Errorf("validating updated library: %w", err)
	}

	return &RemoveResourceOutput{
		Type:         "resource",
		ResourceType: typ,
		Name:         name,
		FileDeleted:  physicalPath,
		LibraryPath:  opts.LibraryPath,
	}, nil
}

// RemovePreset removes a preset from the library.
func RemovePreset(opts RemovePresetOptions) (*RemovePresetOutput, error) {
	if opts.Name == "" {
		return nil, gerrors.NewValidationError("", "name", "", "preset name is required")
	}

	lib, err := LoadLibrary(opts.LibraryPath)
	if err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}

	preset, exists := lib.Presets[opts.Name]
	if !exists {
		return nil, gerrors.NewFileError(opts.LibraryPath, "access",
			fmt.Sprintf("preset %s not found", opts.Name), nil)
	}

	resourcesRemoved := make([]string, len(preset.Resources))
	copy(resourcesRemoved, preset.Resources)

	if err := removePresetFromLibrary(opts.LibraryPath, opts.Name); err != nil {
		return nil, fmt.Errorf("updating library.yaml: %w", err)
	}

	if _, err := LoadLibrary(opts.LibraryPath); err != nil {
		return nil, fmt.Errorf("validating updated library: %w", err)
	}

	return &RemovePresetOutput{
		Type:             "preset",
		Name:             opts.Name,
		ResourcesRemoved: resourcesRemoved,
	}, nil
}

func removeResourceFromLibrary(libraryPath, typ, name string) error {
	yamlPath := filepath.Join(libraryPath, "library.yaml")

	content, err := os.ReadFile(yamlPath) //nolint:gosec // G304: User provides library path, must read fixed library.yaml
	if err != nil {
		return gerrors.NewFileError(yamlPath, "read", "failed to read library.yaml", err)
	}

	var lib libraryYAML
	if err := yaml.Unmarshal(content, &lib); err != nil {
		return gerrors.NewParseError(yamlPath, "failed to parse library.yaml", err)
	}

	if lib.Resources != nil && lib.Resources[typ] != nil {
		delete(lib.Resources[typ], name)
		if len(lib.Resources[typ]) == 0 {
			delete(lib.Resources, typ)
		}
	}

	output, err := yaml.Marshal(lib)
	if err != nil {
		return gerrors.NewParseError(yamlPath, "failed to marshal library.yaml", err)
	}

	tmpPath := yamlPath + ".tmp"
	if err := os.WriteFile(tmpPath, output, 0o600); err != nil {
		return gerrors.NewFileError(tmpPath, "write", "failed to write library.yaml", err)
	}
	if err := os.Rename(tmpPath, yamlPath); err != nil {
		return gerrors.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
	}

	return nil
}

func removePresetFromLibrary(libraryPath, name string) error {
	yamlPath := filepath.Join(libraryPath, "library.yaml")

	content, err := os.ReadFile(yamlPath) //nolint:gosec // G304: User provides library path, must read fixed library.yaml
	if err != nil {
		return gerrors.NewFileError(yamlPath, "read", "failed to read library.yaml", err)
	}

	var lib libraryYAML
	if err := yaml.Unmarshal(content, &lib); err != nil {
		return gerrors.NewParseError(yamlPath, "failed to parse library.yaml", err)
	}

	if lib.Presets != nil {
		delete(lib.Presets, name)
	}

	output, err := yaml.Marshal(lib)
	if err != nil {
		return gerrors.NewParseError(yamlPath, "failed to marshal library.yaml", err)
	}

	tmpPath := yamlPath + ".tmp"
	if err := os.WriteFile(tmpPath, output, 0o600); err != nil {
		return gerrors.NewFileError(tmpPath, "write", "failed to write library.yaml", err)
	}
	if err := os.Rename(tmpPath, yamlPath); err != nil {
		return gerrors.NewFileError(yamlPath, "rename", "failed to update library.yaml", err)
	}

	return nil
}
