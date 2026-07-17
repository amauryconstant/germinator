package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// TestRootCommand_VerboseFlagWiresIOStreams verifies that the -v / --verbose
// persistent flag is wired into Factory.IOStreams.Verbose via
// PersistentPreRunE. Per the cli-verbose-output spec, -v must enable progress
// disclosure (Verbosef) to stderr; before this wiring, the flag was declared
// but never read, making every production Verbosef call a silent no-op.
func TestRootCommand_VerboseFlagWiresIOStreams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantVerbose bool
	}{
		{
			name:        "no flag leaves Verbose false",
			args:        []string{"version"},
			wantVerbose: false,
		},
		{
			name:        "-v sets Verbose true",
			args:        []string{"-v", "version"},
			wantVerbose: true,
		},
		{
			name:        "--verbose (long form) sets Verbose true",
			args:        []string{"--verbose", "version"},
			wantVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios := iostreams.Test()
			f := cmdutil.NewFactory(context.Background(), ios)

			require.NoError(t, executeCmd(t, func() any {
				return NewRootCommand(f)
			}, tt.args...))

			assert.Equal(t, tt.wantVerbose, ios.Verbose,
				"Factory.IOStreams.Verbose should match the parsed --verbose flag")
		})
	}
}

// TestRootCommand_VerbosefHonorsFlag verifies the end-to-end flow: once -v
// is set via PersistentPreRunE, a subsequent Verbosef call on the same
// IOStreams instance writes the formatted message to ErrOut. This proves the
// wiring is not just a no-op assignment but actually flips the Verbosef gate.
func TestRootCommand_VerbosefHonorsFlag(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios)

	require.NoError(t, executeCmd(t, func() any {
		return NewRootCommand(f)
	}, "-v", "version"))

	require.True(t, ios.Verbose, "Verbose must be true after -v wires through")

	ios.Verbosef("progress: %d items", 7)

	errBuf, ok := ios.ErrOut.(*bytes.Buffer)
	require.True(t, ok, "ios.ErrOut must be a *bytes.Buffer in tests")
	assert.Contains(t, errBuf.String(), "progress: 7 items",
		"Verbosef must write to ErrOut when Verbose is true")
}

// TestRootCommand_VerbosefSilentWithoutFlag is the negative path: without
// -v, Verbosef must NOT write to ErrOut. Locks down the gate semantics.
func TestRootCommand_VerbosefSilentWithoutFlag(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios)

	require.NoError(t, executeCmd(t, func() any {
		return NewRootCommand(f)
	}, "version"))

	require.False(t, ios.Verbose, "Verbose must be false when -v is absent")

	ios.Verbosef("this should not appear")

	errBuf, ok := ios.ErrOut.(*bytes.Buffer)
	require.True(t, ok, "ios.ErrOut must be a *bytes.Buffer in tests")
	assert.Empty(t, errBuf.String(),
		"Verbosef must be a no-op when Verbose is false")
}
