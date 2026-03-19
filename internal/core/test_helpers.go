package core

import (
	"os"
	"path/filepath"
	"testing"
)

func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	testPath := filepath.Join(wd, "..", "..", "test")
	if _, err := os.Stat(testPath); err == nil {
		return filepath.Abs(filepath.Join(wd, "..", ".."))
	}

	altTestPath := filepath.Join(wd, "..", "..", "..", "test")
	if _, err := os.Stat(altTestPath); err == nil {
		return filepath.Abs(filepath.Join(wd, "..", "..", ".."))
	}

	return "", os.ErrNotExist
}

func getFixturesDir(t *testing.T) string {
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}
	return filepath.Join(root, "test", "fixtures")
}
