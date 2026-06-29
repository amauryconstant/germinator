package cmd

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

func TestLibraryCommand_Resources(t *testing.T) {
	// Set environment variable to test fixtures
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	f := newTestFactory()
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"resources"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// runResources writes to f.IOStreams.Out (iostreams.Test buffer),
	// not to the cobra command's stdout buffer.
	outBuf, ok := f.IOStreams.Out.(*bytes.Buffer)
	if !ok {
		t.Fatal("f.IOStreams.Out is not a *bytes.Buffer")
	}
	output := outBuf.String()
	if output == "" {
		t.Error("Expected output from library resources command")
	}
}

func TestLibraryCommand_Presets(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	io := iostreams.Test()
	outBuf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"presets"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.Execute())

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library presets command")
}

func TestLibraryCommand_Show(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	io := iostreams.Test()
	outBuf, ok := io.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")

	var buf bytes.Buffer
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"show", "skill/commit"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	require.NoError(t, cmd.Execute())

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library show command")
}

func TestLibraryCommand_InvalidRef(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	f := newTestFactory()
	cmd := NewLibraryCommand(f, newTestBridge(), nil)
	cmd.SetArgs([]string{"show", "invalidformat"})

	err = cmd.Execute()
	require.Error(t, err, "invalid ref must produce an error")

	var notFound *core.NotFoundError
	require.True(t, errors.As(err, &notFound),
		"error must be a *core.NotFoundError, got %T: %v", err, err)
	assert.Equal(t, "invalidformat", notFound.Key)
}

// Legacy TestInitCommand_* tests were removed in slice 5; they used
// cmd.SetOut(&buf) to capture Cobra stdout, incompatible with the new
// pattern where output goes to opts.IO.Out via the Factory. Proper
// coverage lives in cmd/init_test.go using iostreams.Test() + runF
// injection.
