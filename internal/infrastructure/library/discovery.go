package library

import (
	"os"
	"path/filepath"
)

// FindLibrary discovers the library path using priority chain:
// 1. flagPath (explicit --library flag)
// 2. envPath (GERMINATOR_LIBRARY environment variable)
// 3. DefaultLibraryPath() (XDG_DATA_HOME or platform-appropriate data directory)
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
// Follows XDG Base Directory Specification for data files:
// - $XDG_DATA_HOME/germinator/library/ if XDG_DATA_HOME is set
// - ~/.local/share/germinator/library/ on Unix-like systems (XDG default)
// - ~/Library/Application Support/germinator/library/ on macOS
// - %APPDATA%/germinator/library/ on Windows
// - ./germinator/library/ as last resort (current directory)
func DefaultLibraryPath() string {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "germinator", "library")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".germinator", "library")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(homeDir, ".local", "share", "germinator", "library")
	}

	if configDir == filepath.Join(homeDir, "Library", "Application Support") {
		return filepath.Join(configDir, "germinator", "library")
	}

	return filepath.Join(homeDir, ".local", "share", "germinator", "library")
}

// Exists checks if a library directory exists at the given path.
func Exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// YAMLExists checks if a library.yaml configuration file exists at the given library path.
func YAMLExists(path string) bool {
	yamlPath := filepath.Join(path, "library.yaml")
	_, err := os.Stat(yamlPath)
	return err == nil
}
