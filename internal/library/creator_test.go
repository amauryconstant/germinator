package library

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gerrors "gitlab.com/amoconst/germinator/internal/core"
)

// Tests for CreateLibrary and defaultLibraryYAML (creator.go).
// Placed in a dedicated test file per the package's one-test-per-source
// convention (adder.go → adder_test.go, saver.go → saver_test.go, etc.).

// TestCreateLibrary covers the four observable behaviors of CreateLibrary:
// the dry-run preview path, force-overwrite of an existing library, the
// existence-check error path, and the default real-create path that writes
// the directory structure + library.yaml from defaultLibraryYAML.
func TestCreateLibrary(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, path string)
		opts      CreateOptions
		stdout    *bytes.Buffer
		wantErr   bool
		checkErr  func(t *testing.T, err error)
		wantLines []string
		wantPaths []string
	}{
		{
			name:    "dry-run",
			opts:    CreateOptions{DryRun: true},
			stdout:  &bytes.Buffer{},
			wantErr: false,
			wantLines: []string{
				"Would create library at:",
				"library.yaml",
				"skills/",
				"agents/",
				"commands/",
				"memory/",
			},
		},
		{
			name: "force-overwrite",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(path, 0o750))
				require.NoError(t, os.WriteFile(
					filepath.Join(path, "library.yaml"),
					[]byte(`version: "99"`),
					0o600,
				))
			},
			opts:    CreateOptions{Force: true},
			stdout:  &bytes.Buffer{},
			wantErr: false,
			wantLines: []string{
				`version: "1"`,
			},
			wantPaths: []string{"library.yaml", "skills", "agents", "commands", "memory"},
		},
		{
			name: "existing-library-error",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(path, 0o750))
				require.NoError(t, os.WriteFile(
					filepath.Join(path, "library.yaml"),
					[]byte(`version: "1"`),
					0o600,
				))
			},
			opts:     CreateOptions{},
			stdout:   &bytes.Buffer{},
			wantErr:  true,
			checkErr: assertFileError,
		},
		{
			name:    "default-path-resolution",
			opts:    CreateOptions{},
			stdout:  &bytes.Buffer{},
			wantErr: false,
			wantLines: []string{
				`version: "1"`,
			},
			wantPaths: []string{"library.yaml", "skills", "agents", "commands", "memory"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			libPath := filepath.Join(dir, "lib")

			if tt.setup != nil {
				tt.setup(t, libPath)
			}

			tt.opts.Path = libPath
			var stdout io.Writer = tt.stdout
			if tt.stdout == nil {
				stdout = nil
			}

			err := CreateLibrary(context.Background(), tt.opts, stdout)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
				return
			}
			require.NoError(t, err)

			if tt.opts.DryRun {
				if tt.stdout != nil {
					got := tt.stdout.String()
					for _, line := range tt.wantLines {
						if strings.Contains(got, line) {
							continue
						}
						assert.Contains(t, got, line, "dry-run stdout must contain")
					}
				}
				_, statErr := os.Stat(libPath)
				require.True(t, os.IsNotExist(statErr),
					"dry-run should not have created %s", libPath)
				return
			}

			for _, p := range tt.wantPaths {
				fp := filepath.Join(libPath, p)
				info, statErr := os.Stat(fp)
				require.NoError(t, statErr, "expected path to exist after CreateLibrary: %s", fp)
				if p != "library.yaml" {
					require.True(t, info.IsDir(), "expected %s to be a directory", fp)
				}
			}

			yamlBytes, readErr := os.ReadFile(filepath.Join(libPath, "library.yaml"))
			require.NoError(t, readErr)
			yamlStr := string(yamlBytes)
			for _, line := range tt.wantLines {
				assert.Contains(t, yamlStr, line, "library.yaml must contain")
			}
		})
	}
}

// assertFileError verifies that err is or wraps a *core.FileError.
// Used as the checkErr callback in TestCreateLibrary's table-driven
// cases that expect a typed file-error (e.g., "library already exists").
func assertFileError(t *testing.T, err error) {
	t.Helper()
	var fe *gerrors.FileError
	require.True(t, errors.As(err, &fe),
		"expected error to wrap *core.FileError, got %T (%v)", err, err)
}

// TestDefaultLibraryYAML verifies the default library.yaml content
// produced by defaultLibraryYAML() at creator.go:123. The function is
// the source of truth for the on-disk shape of a freshly-created
// library; any drift here invalidates the post-create LoadLibrary
// validation step.
func TestDefaultLibraryYAML(t *testing.T) {
	yaml := defaultLibraryYAML()

	tests := []struct {
		name     string
		contains []string
	}{
		{
			name:     "version field present",
			contains: []string{`version: "1"`},
		},
		{
			name: "empty resources map",
			contains: []string{
				"resources:",
				"skill: {}",
				"agent: {}",
				"command: {}",
				"memory: {}",
			},
		},
		{
			name:     "empty presets map",
			contains: []string{"presets: {}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, sub := range tt.contains {
				assert.Contains(t, yaml, sub,
					"defaultLibraryYAML() must contain %q", sub)
			}
		})
	}
}
