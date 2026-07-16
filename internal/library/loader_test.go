package library

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/amoconst/germinator/internal/core"
)

func TestLoadLibrary(t *testing.T) {
	// Get absolute path to fixtures
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)

	lib, err := LoadLibrary(context.Background(), absPath)
	require.NoError(t, err)

	// Verify version
	assert.Equal(t, "1", lib.Version)

	// Verify resources exist
	assert.NotEmpty(t, lib.Resources, "Library has no resources")

	// Verify skills
	skills, ok := lib.Resources["skill"]
	require.True(t, ok, "Library has no skills")
	assert.GreaterOrEqual(t, len(skills), 2, "Expected at least 2 skills")

	// Verify specific skill
	commit, ok := skills["commit"]
	require.True(t, ok, "Library missing skill/commit")
	assert.Equal(t, "skills/skill-commit.md", commit.Path)
	assert.Equal(t, "Git commit best practices", commit.Description)

	// Verify agents
	agents, ok := lib.Resources["agent"]
	require.True(t, ok, "Library has no agents")
	_, ok = agents["reviewer"]
	assert.True(t, ok, "Library missing agent/reviewer")

	// Verify presets
	assert.NotEmpty(t, lib.Presets, "Library has no presets")

	gitWorkflow, ok := lib.Presets["git-workflow"]
	require.True(t, ok, "Library missing git-workflow preset")
	assert.Len(t, gitWorkflow.Resources, 2)
}

func TestLoadLibrary_MissingDirectory(t *testing.T) {
	_, err := LoadLibrary(context.Background(), "/nonexistent/path/to/library")
	require.Error(t, err)

	var nf *core.NotFoundError
	require.True(t, errors.As(err, &nf),
		"missing-directory failure MUST surface as *core.NotFoundError (entity 'library')")
	assert.Equal(t, "library", nf.Entity)
	assert.Equal(t, "/nonexistent/path/to/library", nf.Key)
}

func TestLoadLibrary_MissingYAML(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadLibrary(context.Background(), tmpDir)
	require.Error(t, err)

	var nf *core.NotFoundError
	require.True(t, errors.As(err, &nf),
		"missing-library.yaml failure MUST surface as *core.NotFoundError (entity 'library.yaml')")
	assert.Equal(t, "library.yaml", nf.Entity)
}

func TestLoadLibrary_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Write invalid YAML
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte("invalid: [yaml: content"), 0644))

	_, err := LoadLibrary(context.Background(), tmpDir)
	require.Error(t, err)
}

func TestLoadLibrary_MissingVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML without version
	yamlContent := `
resources:
  skill:
    test:
      path: skills/test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0644))

	_, err := LoadLibrary(context.Background(), tmpDir)
	require.Error(t, err)
}

func TestLoadLibrary_UnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML with unsupported version
	yamlContent := `
version: "2"
resources:
  skill:
    test:
      path: skills/test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0644))

	_, err := LoadLibrary(context.Background(), tmpDir)
	require.Error(t, err)
}

func TestLoadLibrary_InvalidResourceType(t *testing.T) {
	tmpDir := t.TempDir()

	// Write YAML with invalid resource type
	yamlContent := `
version: "1"
resources:
  invalid-type:
    test:
      path: test.yaml
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0644))

	_, err := LoadLibrary(context.Background(), tmpDir)
	require.Error(t, err)
}

func TestLoadLibrary_EmptyLibrary(t *testing.T) {
	tmpDir := t.TempDir()

	// Write minimal valid YAML
	yamlContent := `
version: "1"
resources: {}
presets: {}
`
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0644))

	lib, err := LoadLibrary(context.Background(), tmpDir)
	require.NoError(t, err)

	assert.Empty(t, lib.Resources)
	assert.Empty(t, lib.Presets)
}

func TestLoadLibrary_NotADirectory(t *testing.T) {
	// Create temp file (not directory)
	tmpFile := filepath.Join(t.TempDir(), "notadir")
	require.NoError(t, os.WriteFile(tmpFile, []byte("test"), 0644))

	_, err := LoadLibrary(context.Background(), tmpFile)
	require.Error(t, err)
}
