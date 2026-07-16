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
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	ios := iostreams.Test()
	outBuf, ok := ios.Out.(*bytes.Buffer)
	require.True(t, ok, "ios.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	require.NoError(t, executeCmd(t, func() any {
		return NewLibraryCommand(f, nil)
	}, "resources"))

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library resources command")
}

func TestLibraryCommand_Presets(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	ios := iostreams.Test()
	outBuf, ok := ios.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")

	var buf bytes.Buffer
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewLibraryCommand(f, nil)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		return cmd
	}, "presets"))

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library presets command")
}

func TestLibraryCommand_Show(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	ios := iostreams.Test()
	outBuf, ok := ios.Out.(*bytes.Buffer)
	require.True(t, ok, "io.Out must be a *bytes.Buffer")

	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")

	var buf bytes.Buffer
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewLibraryCommand(f, nil)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		return cmd
	}, "show", "skill/commit"))

	assert.NotEmpty(t, outBuf.String(),
		"Expected output from library show command")
}

func TestLibraryCommand_InvalidRef(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "library")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)
	t.Setenv("GERMINATOR_LIBRARY", absPath)

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")

	err = executeCmd(t, func() any {
		return NewLibraryCommand(f, nil)
	}, "show", "invalidformat")
	require.Error(t, err, "invalid ref must produce an error")

	// Phase 3.20: an unparseable ref (no "/" separator) now surfaces
	// as *core.ConfigError — the ref is malformed, not "missing".
	// NotFoundError is reserved for runtime lookup misses.
	var cfgErr *core.ConfigError
	require.True(t, errors.As(err, &cfgErr),
		"error must be a *core.ConfigError (parse error), got %T: %v", err, err)
	assert.Equal(t, "reference", cfgErr.Field())
}

// Legacy TestInitCommand_* tests were removed in slice 5; they used
// cmd.SetOut(&buf) to capture Cobra stdout, incompatible with the new
// pattern where output goes to opts.IO.Out via the Factory. Proper
// coverage lives in cmd/init_test.go using iostreams.Test() + runF
// injection.
