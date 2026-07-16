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
// path string. `xdg.DataHome` is read from `currentXDGDataHome()`
// which holds xdgReloadMu across both the Reload and the read so
// concurrent callers see either pre-reload or post-reload state
// atomically (the third-party xdg package does not lock its own
// caches).
func DefaultLibraryPath() string {
	home := currentXDGDataHome()
	path := filepath.Join(home, "germinator", "library")
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
// locking. Calling xdg.Reload from concurrent goroutines (e.g.,
// parallel tests both invoking DefaultLibraryPath after one mutated
// XDG_* via t.Setenv) races on those caches. We serialize the
// calls AND the subsequent reads of xdg's package-level fields
// (xdg.DataHome / xdg.ConfigHome / xdg.ConfigFile return values)
// under the same mutex so concurrent callers see either pre-reload
// or post-reload state atomically; holding the mutex across the
// reload only would leave a window where another goroutine can
// Reload and produce a torn read. The mutex is package-private and
// only held across the cheap (no-I/O) reload + read.
var xdgReloadMu sync.Mutex

// currentXDGDataHome returns xdg.DataHome after a serialized Reload,
// so concurrent callers see either the pre-reload or post-reload
// value atomically. The mutex is held across both the Reload write
// and the DataHome read.
func currentXDGDataHome() string {
	xdgReloadMu.Lock()
	defer xdgReloadMu.Unlock()
	xdg.Reload()
	return xdg.DataHome
}
