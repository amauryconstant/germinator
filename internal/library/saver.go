package library

// Package library provides library management for canonical resources.

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	gerrors "gitlab.com/amoconst/germinator/internal/core"
	yaml "gopkg.in/yaml.v3"
)

// atomicWriteFile writes data to path with perm atomically via the
// write-temp-then-rename pattern, falling back to copy+remove on
// syscall.EXDEV (cross-filesystem rename). This is the single source
// of truth for library.yaml atomic writes; AddResource, RemoveResource,
// RemovePreset, and SaveLibrary all delegate here.
//
// renameFunc is a test-only seam; production callers (and the default)
// call os.Rename. When the seam returns syscall.EXDEV, the helper
// transparently falls back to copy+remove so cross-filesystem atomic
// writes succeed where plain rename would fail.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error { //nolint:unparam // perm parameter is part of the helper's public shape; future call sites may pass non-0o600 modes.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return gerrors.NewFileError(tmpPath, "write", "failed to write temp", err)
	}
	rename := defaultRenameFunc
	if renameFunc != nil {
		rename = renameFunc
	}
	if err := rename(tmpPath, path); err != nil {
		if errors.Is(err, syscall.EXDEV) {
			return atomicWriteFileCrossFS(tmpPath, path, perm)
		}
		return gerrors.NewFileError(path, "rename", "failed to update file", err)
	}
	return nil
}

// atomicWriteFileCrossFS is the EXDEV fallback for atomicWriteFile.
// It copies the freshly-written temp file to the target path and
// removes the temp. The new file is fully written before the old temp
// is removed, preserving the atomic-or-fail contract at the user-
// observable level.
func atomicWriteFileCrossFS(tmpPath, path string, perm os.FileMode) error {
	in, err := os.Open(tmpPath) //nolint:gosec // G304: temp file path is internally controlled
	if err != nil {
		return gerrors.NewFileError(path, "rename", "failed to open temp for copy", err)
	}
	defer in.Close()                                                        //nolint:errcheck // close on read-only file
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm) //nolint:gosec // G304: target path is the library.yaml the helper was asked to write
	if err != nil {
		return gerrors.NewFileError(path, "rename", "failed to open target for copy", err)
	}
	defer out.Close() //nolint:errcheck // close best-effort
	if _, err := io.Copy(out, in); err != nil {
		return gerrors.NewFileError(path, "rename", "failed to copy across filesystems", err)
	}
	if err := out.Sync(); err != nil {
		return gerrors.NewFileError(path, "rename", "failed to sync target", err)
	}
	if err := os.Remove(tmpPath); err != nil {
		return gerrors.NewFileError(path, "rename", "failed to remove temp", err)
	}
	return nil
}

// defaultRenameFunc is the production rename; tests override via renameFunc.
// The error is returned unwrapped so atomicWriteFile can errors.Is
// against syscall.EXDEV directly; callers wrap as needed.
func defaultRenameFunc(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath) //nolint:wrapcheck // raw error required for errors.Is(err, syscall.EXDEV) detection in atomicWriteFile
}

// renameFunc is a test-only seam that overrides defaultRenameFunc in
// atomicWriteFile. nil falls back to defaultRenameFunc. Set only via
// setRenameFuncForTest.
var renameFunc func(oldPath, newPath string) error

// SaveLibrary persists the library to its RootPath as library.yaml.
// It marshals the entire library structure and writes it to disk via
// the atomicWriteFile helper (atomic temp+rename on the same
// filesystem; copy+remove fallback across filesystems).
func SaveLibrary(lib *Library) error {
	if lib.RootPath == "" {
		return gerrors.NewFileError("", "write", "library has no root path set", nil)
	}

	// Ensure directory exists.
	// Unix permission bits (0o750) are no-ops on Windows; Windows
	// support is out of scope.
	if err := os.MkdirAll(lib.RootPath, 0o750); err != nil {
		return gerrors.NewFileError(lib.RootPath, "create", "failed to create library directory", err)
	}

	// Marshal library to YAML
	data, err := yaml.Marshal(lib)
	if err != nil {
		return gerrors.NewFileError(lib.RootPath, "marshal", "failed to marshal library to YAML", err)
	}

	// Write to library.yaml atomically (was previously non-atomic
	// direct os.WriteFile; converted per fix-library-io-discipline to
	// gain torn-write safety via temp+rename).
	yamlPath := filepath.Join(lib.RootPath, "library.yaml")
	if err := atomicWriteFile(yamlPath, data, 0o600); err != nil {
		return err
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
