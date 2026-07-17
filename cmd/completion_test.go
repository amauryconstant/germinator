package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// TestNewCmdCompletion_RegistersAllShells asserts the parent
// `completion` command registers exactly one subcommand per entry in
// the supported shells list returned by shells().
func TestNewCmdCompletion_RegistersAllShells(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	withCmd(t, func() any { return NewCmdCompletion(f, nil) }, func(cmd *cobra.Command) {
		names := map[string]bool{}
		for _, sub := range cmd.Commands() {
			names[sub.Name()] = true
		}
		for _, shell := range shells() {
			assert.True(t, names[shell], "expected subcommand %q to be registered", shell)
		}
		assert.Len(t, cmd.Commands(), len(shells()),
			"completion command should register exactly one subcommand per shell")
	})
}

// TestNewCmdCompletion_RunFInjection verifies the runF seam captures
// the per-shell options without invoking carapace snippet generation.
func TestNewCmdCompletion_RunFInjection(t *testing.T) {
	t.Parallel()

	captured := make(chan *completionOptions, 1)
	runF := func(opts *completionOptions) error { //nolint:unparam // runF is a test callback; success is the only meaningful return
		captured <- opts
		return nil
	}

	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdCompletion(f, runF)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd
	}, "bash"))
	close(captured)
	opts := <-captured
	require.NotNil(t, opts, "runF must be invoked")
	assert.Equal(t, "bash", opts.Shell)
	assert.Equal(t, ios, opts.IO, "opts.IO must be wired from f.IOStreams")
	assert.NotNil(t, opts.Ctx, "opts.Ctx must be wired from c.Context()")
}

// TestRunCompletion_WritesSnippet verifies runCompletion writes a
// non-empty snippet to opts.IO.Out for each supported shell.
func TestRunCompletion_WritesSnippet(t *testing.T) {
	t.Parallel()

	for _, shell := range shells() {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			ios := iostreams.Test()
			f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
			// runCompletion reads the per-shell leaf's Root() and
			// calls carapace.Gen on it; both root construction and
			// snippet generation mutate carapace's package-level
			// state, so serialise through testExecMu.
			var (
				root *cobra.Command
				err  error
			)
			testExecMu.Lock()
			root = NewRootCommand(f)
			sub := findShellLeaf(t, root, shell)
			opts := &completionOptions{
				IO:    ios,
				Ctx:   context.Background(),
				Shell: shell,
			}
			err = runCompletion(sub, opts)
			testExecMu.Unlock()
			require.NoError(t, err)
			out, ok := ios.Out.(*bytes.Buffer)
			require.True(t, ok, "iostreams.Test must return *bytes.Buffer-backed Out")
			assert.NotEmpty(t, out.String(), "snippet for %s must be non-empty", shell)
		})
	}
}

// findShellLeaf walks the root command tree to locate the per-shell
// `completion <shell>` leaf command. Fails the test if missing.
func findShellLeaf(t *testing.T, root *cobra.Command, shell string) *cobra.Command {
	t.Helper()
	var leaf *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == "completion" {
			for _, c := range sub.Commands() {
				if c.Name() == shell {
					leaf = c
					break
				}
			}
		}
	}
	require.NotNil(t, leaf, "expected completion %q subcommand to be registered", shell)
	return leaf
}
