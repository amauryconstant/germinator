// Package library provides library management for canonical resources.
package library

import (
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

// SaveLibrary persists the library to its RootPath as library.yaml.
// It marshals the entire library structure and writes it to disk.
func SaveLibrary(lib *Library) error {
	if lib.RootPath == "" {
		return gerrors.NewFileError("", "write", "library has no root path set", nil)
	}

	// Ensure directory exists
	if err := os.MkdirAll(lib.RootPath, 0o750); err != nil {
		return gerrors.NewFileError(lib.RootPath, "create", "failed to create library directory", err)
	}

	// Marshal library to YAML
	data, err := yaml.Marshal(lib)
	if err != nil {
		return gerrors.NewFileError(lib.RootPath, "marshal", "failed to marshal library to YAML", err)
	}

	// Write to library.yaml
	yamlPath := filepath.Join(lib.RootPath, "library.yaml")
	if err := os.WriteFile(yamlPath, data, 0o600); err != nil {
		return gerrors.NewFileError(yamlPath, "write", "failed to write library.yaml", err)
	}

	return nil
}

// AddPreset adds a preset to the library in-memory.
// It validates the preset before adding and ensures the Presets map is initialized.
func AddPreset(lib *Library, preset Preset) error {
	if lib.Presets == nil {
		lib.Presets = make(map[string]Preset)
	}

	// Validate preset before adding
	if err := preset.Validate(); err != nil {
		return err
	}

	// Add preset using the name as the key
	lib.Presets[preset.Name] = preset

	return nil
}

// PresetExists checks if a preset with the given name exists in the library.
func PresetExists(lib *Library, name string) bool {
	if lib == nil || lib.Presets == nil {
		return false
	}
	_, exists := lib.Presets[name]
	return exists
}
