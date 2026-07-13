// Package paths provides shared filesystem path helpers used by both
// internal/config (library path expansion) and cmd/completions
// (flag-value expansion). Centralizing these helpers avoids drift
// between the two call sites that previously maintained independent
// (and slightly different) implementations of `~/` expansion.
package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ExpandHome expands a leading "~/" in path to the current user's
// home directory.
//
// Empty paths are returned unchanged. Non-tilde paths and a bare
// "~" (without a trailing slash) are returned as-is — only the
// "~/" prefix is expanded. This matches the legacy expandTilde
// behavior in internal/config (pre-amendment), where bare "~"
// was returned unchanged.
//
// An error is returned only when the path starts with "~/" but the
// home directory cannot be determined (os.UserHomeDir failure). The
// underlying error is wrapped via errors.Join so callers can inspect
// it via errors.Is.
//
// Callers that want a silent fallback (e.g., completion resolution
// where a broken HOME should not break the user experience) should
// catch the error and use the original path:
func ExpandHome(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if len(path) < 2 || path[:2] != "~/" {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Join(
			fmt.Errorf("cannot determine home directory for path %q", path),
			err,
		)
	}
	return filepath.Join(homeDir, path[2:]), nil
}
