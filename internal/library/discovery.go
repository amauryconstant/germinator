package library

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
)

// FindLibrary discovers the library path using the spec-mandated
// priority chain:
//
//  1. flagPath — explicit `--library` flag (highest)
//  2. envPath — `GERMINATOR_LIBRARY` env var
//  3. cfgPath — `Config.Library` (config-file override)
//  4. DefaultLibraryPath() — XDG via `adrg/xdg.DataFile`,
//     falling back to `./germinator/library/` for project-local
//     libraries
//
// Returns the first non-empty path in the priority chain.
func FindLibrary(flagPath, envPath, cfgPath string) string {
	if flagPath != "" {
		return flagPath
	}
	if envPath != "" {
		return envPath
	}
	if cfgPath != "" {
		return cfgPath
	}
	return DefaultLibraryPath()
}

// DefaultLibraryPath returns the default library path.
//
// It first asks `adrg/xdg` for the canonical data location
// (`$XDG_DATA_HOME/germinator/library` on Unix, with platform-
// appropriate equivalents on macOS / Windows via `adrg/xdg`). When
// the resolved path does not exist on disk AND a project-local
// `./germinator/library/` does exist in the current working
// directory, the absolute project-local path is returned
// (last-resort override for projects that ship their own library
// alongside `germinator`).
//
// The function does NOT call `xdg.DataFile` directly because that
// helper attempts to create the directory on disk; we only need the
// path string. `xdg.DataHome` is computed from the env on `Reload()`
// (called below under xdgReloadMu to serialize concurrent updates).
func DefaultLibraryPath() string {
	xdgReload()
	path := filepath.Join(xdg.DataHome, "germinator", "library")
	if Exists(path) {
		return path
	}
	if cwdLib, err := filepath.Abs("./germinator/library"); err == nil && Exists(cwdLib) {
		return cwdLib
	}
	return path
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

// xdgReloadMu serializes calls to adrg/xdg.Reload, which mutates
// package-level caches in the third-party library without internal
// locking. Calling xdg.Reload from concurrent goroutines (e.g., parallel
// tests both invoking DefaultLibraryPath / resolveConfigPath after one
// mutated XDG_* via t.Setenv) races on those caches. We serialize the
// calls here so the global state is updated atomically; concurrent
// readers either see the pre-reload or post-reload value, never a
// torn update. The mutex is package-private and only held across
// xdg.Reload (cheap; no I/O).
var xdgReloadMu sync.Mutex

func xdgReload() {
	xdgReloadMu.Lock()
	defer xdgReloadMu.Unlock()
	xdg.Reload()
}
