package library

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for CreateLibrary (creator.go). Placed in a dedicated test
// file per the package's one-test-per-source convention
// (adder.go → adder_test.go, saver.go → saver_test.go, etc.).
// creator.go had no test file before this; the dry-run writer path
// was previously only verified end-to-end via test/e2e/library_init_test.go.

// TestCreateLibrary_DryRun_WritesToStdout verifies that the dry-run
// preview writes to the stdout parameter (gated on stdout != nil)
// rather than to os.Stdout directly. Closes the partial coverage on
// the library-library-scaffolding spec scenario "Dry-run does not
// create files" by injecting a bytes.Buffer and asserting each
// preview line is present in the captured output.
func TestCreateLibrary_DryRun_WritesToStdout(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "writer-lib")

	var buf bytes.Buffer
	err := CreateLibrary(context.Background(), CreateOptions{
		Path:   tmpDir,
		DryRun: true,
	}, &buf)
	require.NoError(t, err)

	got := buf.String()
	wantSubstrings := []string{
		"Would create library at:",
		tmpDir,
		filepath.Join(tmpDir, "library.yaml"),
		filepath.Join(tmpDir, "skills") + "/",
		filepath.Join(tmpDir, "agents") + "/",
		filepath.Join(tmpDir, "commands") + "/",
		filepath.Join(tmpDir, "memory") + "/",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(got, s) {
			assert.Contains(t, got, s, "dry-run Stdout must contain")
		}
	}

	// Library directory and library.yaml must NOT exist after a dry-run.
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		assert.True(t, os.IsNotExist(err), "dry-run should not have created tmpDir")
	}
	yamlPath := filepath.Join(tmpDir, "library.yaml")
	if _, err := os.Stat(yamlPath); !os.IsNotExist(err) {
		assert.True(t, os.IsNotExist(err), "dry-run should not have created yamlPath")
	}
}
