//go:build e2e

package fixtures

import (
	"os"
	"path/filepath"
	"runtime"
)

// basePath is the path to the project root
var basePath string

func init() {
	// Get the path to this file's directory
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Navigate to project root (test/e2e/fixtures -> test/e2e -> test -> .)
	basePath = filepath.Join(currentDir, "..", "..", "..")
}

// ProjectRoot returns the absolute path to the project root
func ProjectRoot() string {
	absPath, _ := filepath.Abs(basePath)
	return absPath
}

// FixturesDir returns the path to the test fixtures directory
func FixturesDir() string {
	return filepath.Join(ProjectRoot(), "test", "fixtures")
}

// ValidDocument returns the path to a valid agent document fixture
func ValidDocument() string {
	return filepath.Join(FixturesDir(), "agent-valid.md")
}

// InvalidDocument returns the path to an invalid agent document fixture
func InvalidDocument() string {
	return filepath.Join(FixturesDir(), "agent-invalid.md")
}

// NonexistentFile returns a path to a file that does not exist
func NonexistentFile() string {
	return filepath.Join(ProjectRoot(), "test", "fixtures", "nonexistent-file.yaml")
}

// TempOutputFile creates a temporary file path for output testing
// The caller is responsible for cleanup
func TempOutputFile(prefix string) (string, error) {
	tmpDir := os.TempDir()
	file, err := os.CreateTemp(tmpDir, prefix+"-*.md")
	if err != nil {
		return "", err
	}
	// Close and remove the file - we just want the path
	path := file.Name()
	file.Close()
	os.Remove(path)
	return path, nil
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads the contents of a file
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
