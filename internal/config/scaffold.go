package config

import (
	"errors"
	"os"
	"path/filepath"
)

// DefaultTOML is the default config file content written by
// `config init`. All settings are commented out by default, requiring
// users to explicitly uncomment and configure only the settings they
// want to override. The byte content is pinned by the golden-file
// test at cmd/testdata/config_init_default.golden — any drift is
// caught by `mise run test`.
const DefaultTOML = `# Germinator configuration
# https://github.com/anomalyco/germinator
#
# This file configures germinator's global behavior.
# All settings are optional - defaults are used if omitted.
# Settings below are commented out; uncomment and customize as needed.

# Path to your library directory containing skills, agents, commands, and presets.
# The library must contain a library.yaml index file.
# Supports ~ expansion for home directory.
# Default: ~/.local/share/germinator/library (or $XDG_DATA_HOME/germinator/library if set)
# library = "~/.local/share/germinator/library"

# Default platform when --platform is not specified.
# Options: "opencode" (default), "claude-code"
# Leave empty to require --platform on every command.
# Default: "" (none)
# platform = ""

# Shell completion configuration
[completion]

# Maximum time to wait for library loading during completion suggestions.
# Lower values = faster but may timeout on large libraries.
# Default: "500ms"
# timeout = "500ms"

# How long to cache library data for completion performance.
# Higher values = faster completions but may show stale results.
# Default: "5s"
# cache_ttl = "5s"
`

// WriteDefault scaffolds a default germinator config file at path,
// creating any missing parent directories and respecting the force
// flag for overwriting an existing file. It returns *WriteError for
// every I/O failure (per the project's typed-error contract for the
// Imperative Shell layer).
//
// Behavior:
//   - When force is false and path already exists, WriteDefault
//     returns *WriteError{Op: "create", Message: "config file already
//     exists (use --force to overwrite)"} without touching the file.
//   - When force is true, an existing file is overwritten without
//     prompting.
//   - Parent directories are created with permission 0750 (umask may
//     narrow it further); the config file is written with permission
//     0600.
//   - On any underlying *os.PathError from os.Stat, os.MkdirAll, or
//     os.WriteFile, the original error is wrapped in *WriteError via
//     its cause chain so callers can errors.Is / errors.As inspect it.
func WriteDefault(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return NewWriteErrorWithMessage(
				"create", path,
				"config file already exists (use --force to overwrite)",
				nil,
			)
		} else if !errors.Is(err, os.ErrNotExist) {
			return NewWriteError("stat", path, err)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return NewWriteError("mkdir", dir, err)
	}

	if err := os.WriteFile(path, []byte(DefaultTOML), 0o600); err != nil {
		return NewWriteError("write", path, err)
	}

	return nil
}
