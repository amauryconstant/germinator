package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultTOMLByteIdentical pins the byte content of DefaultTOML
// against the golden file at cmd/testdata/config_init_default.golden.
// Any drift between the const and the golden baseline fails the test
// (cmd/config_init_test.go T3 also asserts this from the cmd side).
func TestDefaultTOMLByteIdentical(t *testing.T) {
	t.Parallel()

	got := []byte(DefaultTOML)
	want, err := os.ReadFile("../../cmd/testdata/config_init_default.golden")
	require.NoError(t, err, "golden file must be readable")

	require.Equal(t, len(want), len(got),
		"DefaultTOML byte length must match the golden baseline")
	assert.Equal(t, want, got,
		"DefaultTOML must be byte-identical to the golden baseline")
}

func TestWriteDefault(t *testing.T) {
	t.Parallel()

	t.Run("creates file at custom path", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.toml")

		require.NoError(t, WriteDefault(path, false))

		info, err := os.Stat(path)
		require.NoError(t, err, "config file must exist")
		assert.False(t, info.IsDir())

		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(got), "# Germinator configuration")
		assert.Contains(t, string(got), "[completion]")
	})

	t.Run("creates parent directories", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "subdir", "nested", "config.toml")

		require.NoError(t, WriteDefault(path, false))

		info, err := os.Stat(path)
		require.NoError(t, err, "config file must exist at nested path")
		assert.False(t, info.IsDir())

		di, err := os.Stat(filepath.Join(tmp, "subdir"))
		require.NoError(t, err, "parent directory must exist")
		assert.True(t, di.IsDir())
	})

	t.Run("file permissions are 0600", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.toml")

		require.NoError(t, WriteDefault(path, false))

		fi, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(),
			"config file must be written with 0600 permissions")
	})

	t.Run("parent directories are 0750-or-narrower (umask-respecting)", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "subdir", "nested", "config.toml")

		require.NoError(t, WriteDefault(path, false))

		di, err := os.Stat(filepath.Join(tmp, "subdir"))
		require.NoError(t, err)
		assert.LessOrEqual(t, di.Mode().Perm(), os.FileMode(0o750),
			"parent directory must be no more permissive than 0750 (umask may narrow it)")
	})

	t.Run("force=false and existing file returns WriteError with 'already exists' message", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.toml")
		require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))

		err := WriteDefault(path, false)
		require.Error(t, err)

		var writeErr *WriteError
		require.True(t, errors.As(err, &writeErr),
			"existing-file failure must surface as *config.WriteError")
		assert.Equal(t, "create", writeErr.Op())
		assert.Equal(t, path, writeErr.Path())
		assert.Equal(t, "config file already exists (use --force to overwrite)", writeErr.Message())
		assert.Nil(t, writeErr.Cause(),
			"'already exists' is a precondition check, not an I/O failure; cause must be nil")

		// File must be untouched after the failed write attempt.
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, []byte("existing"), got,
			"file content must not be modified on the rejected-overwrite path")
	})

	t.Run("force=true overwrites existing file", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.toml")
		require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))

		require.NoError(t, WriteDefault(path, true))

		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.True(t, len(got) > 0, "file must not be empty after force-overwrite")
		assert.Contains(t, string(got), "# Germinator configuration",
			"--force must replace content with the scaffolded default")
	})

	t.Run("happy path is idempotent when called twice with force", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.toml")

		require.NoError(t, WriteDefault(path, false))
		first, err := os.ReadFile(path)
		require.NoError(t, err)

		require.NoError(t, WriteDefault(path, true),
			"second call with force=true must succeed regardless of file state")
		second, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, first, second,
			"two consecutive WriteDefault calls with the same content must produce identical bytes")
	})
}
