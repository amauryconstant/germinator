package library

import (
	"os"
	"path/filepath"
)

// FindLibrary discovers the library path using priority chain:
// 1. flagPath (explicit --library flag)
// 2. envPath (GERMINATOR_LIBRARY environment variable)
// 3. DefaultLibraryPath() (~/.config/germinator/library/)
//
// Returns the first non-empty path in the priority chain.
func FindLibrary(flagPath, envPath string) string {
	if flagPath != "" {
		return flagPath
	}
	if envPath != "" {
		return envPath
	}
	return DefaultLibraryPath()
}

// DefaultLibraryPath returns the default library path.
// Uses os.UserConfigDir for XDG compliance:
// - Linux: ~/.config/germinator/library/
// - macOS: ~/Library/Application Support/germinator/library/
// - Windows: %APPDATA%/germinator/library/
func DefaultLibraryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Last resort: current directory
			return filepath.Join(".germinator", "library")
		}
		return filepath.Join(homeDir, ".germinator", "library")
	}
	return filepath.Join(configDir, "germinator", "library")
}

// LibraryExists checks if a library exists at the given path.
func LibraryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// LibraryYAMLExists checks if a library.yaml exists at the given path.
func LibraryYAMLExists(path string) bool {
	yamlPath := filepath.Join(path, "library.yaml")
	_, err := os.Stat(yamlPath)
	return err == nil
}
