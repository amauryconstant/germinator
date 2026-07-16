package cmd

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/version"
)

// outString returns the captured stdout content of ios as a string.
// iostreams.Test() backs Out with *bytes.Buffer; this helper keeps
// call sites concise while preserving the type-assertion convention.
func outString(t *testing.T, ios *iostreams.IOStreams) string {
	t.Helper()
	buf, ok := ios.Out.(*bytes.Buffer)
	require.True(t, ok, "ios.Out must be a *bytes.Buffer in tests")
	return buf.String()
}

// TestRunVersion covers the formatted output of runVersion. Per design
// Decision 3b, runVersion reads from the internal/version package (set
// via -ldflags), NOT from Factory.AppVersion.
func TestRunVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		assertOut func(t *testing.T, out string)
	}{
		{
			name: "matches the documented format regex",
			assertOut: func(t *testing.T, out string) {
				t.Helper()
				// Production binaries have Version/Commit/Date set via
				// -ldflags; the test environment may leave Commit/Date
				// empty (defaults), so the strict regex cannot apply.
				if version.Commit == "" || version.Date == "" {
					t.Skip("internal/version.Commit or Date is empty (no -ldflags set); skipping strict regex")
				}
				// Format: germinator <Version> (<Commit>) <Date>
				pattern := regexp.MustCompile(`^germinator \S+ \(\S+\) \S+\n$`)
				if !pattern.MatchString(out) {
					t.Errorf("runVersion output %q does not match pattern %s", out, pattern)
				}
			},
		},
		{
			name: "contains the binary name 'germinator'",
			assertOut: func(t *testing.T, out string) {
				t.Helper()
				assert.Contains(t, out, "germinator",
					"version output should contain 'germinator'")
			},
		},
		{
			name: "contains the build-time Version from internal/version",
			assertOut: func(t *testing.T, out string) {
				t.Helper()
				if version.Version == "" {
					t.Skip("internal/version.Version is empty (no -ldflags set); skipping content check")
				}
				assert.Contains(t, out, version.Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ios := iostreams.Test()
			opts := &versionOptions{IO: ios}
			require.NoError(t, runVersion(opts))
			tt.assertOut(t, outString(t, ios))
		})
	}
}

// TestNewCmdVersion_runFRoundTrip verifies the runF injection seam:
// the fully-parsed versionOptions reaches runF without invoking runVersion.
func TestNewCmdVersion_runFRoundTrip(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "sentinel-app-version", "germinator")

	var gotOpts *versionOptions
	require.NoError(t, executeCmd(t, func() any {
		return NewCmdVersion(f, func(opts *versionOptions) error {
			gotOpts = opts
			return nil
		})
	}))

	require.NotNil(t, gotOpts)
	assert.Same(t, ios, gotOpts.IO,
		"opts.IO should be the Factory's IOStreams")
}

// TestNewCmdVersion_AppVersionIgnored pins design Decision 3b: the
// version subcommand reads from internal/version, NOT from
// Factory.AppVersion. Setting AppVersion to a sentinel must not
// influence the output.
func TestNewCmdVersion_AppVersionIgnored(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	_ = cmdutil.NewFactory(context.Background(), ios, "IGNORE-ME-SENTINEL", "germinator")

	require.NoError(t, runVersion(&versionOptions{IO: ios}))
	out := outString(t, ios)

	assert.NotContains(t, out, "IGNORE-ME-SENTINEL",
		"Factory.AppVersion sentinel must not appear in version output (Decision 3b)")
	assert.Contains(t, out, "germinator",
		"version output should always contain the binary name")
}

// TestNewCmdVersion_ExecuteExitCode0 confirms Cobra's Execute returns
// nil (exit 0) for the version subcommand end-to-end.
func TestNewCmdVersion_ExecuteExitCode0(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), ios, "test", "germinator")
	require.NoError(t, executeCmd(t, func() any {
		cmd := NewCmdVersion(f, nil)
		cmd.SetOut(&bytes.Buffer{}) // prevent help/version noise on stdout
		return cmd
	}),
		"version command should exit 0 via Cobra Execute")
}
