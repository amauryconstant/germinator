//go:build e2e

package fixtures

import "path/filepath"

func LibraryDir() string {
	return filepath.Join(FixturesDir(), "library")
}
