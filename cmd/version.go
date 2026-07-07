package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/version"
)

// versionOptions holds the runtime state for a `version` invocation.
// Per design Decision 3b, runVersion reads from the internal/version
// package (set via -ldflags at build time), NOT from Factory.AppVersion
// — AppVersion remains a short-form string used elsewhere; the version
// subcommand is the authoritative detailed view.
type versionOptions struct {
	IO  *iostreams.IOStreams
	Ctx context.Context
}

// NewCmdVersion creates the version subcommand via the canonical
// NewCmdXxx(f, runF) pattern. The format is "germinator <Version> (<Commit>) <Date>\n"
// matching the cli-framework spec ("Version Command shows full info")
// and the testing-e2e-testing spec ("Version Command E2E Tests").
func NewCmdVersion(f *cmdutil.Factory, runF func(*versionOptions) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version of germinator",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &versionOptions{
				IO:  f.IOStreams,
				Ctx: c.Context(),
			}
			if runF != nil {
				return runF(opts)
			}
			return runVersion(opts)
		},
	}
	return cmd
}

// runVersion writes the formatted version string to opts.IO.Out.
func runVersion(opts *versionOptions) error {
	_, _ = fmt.Fprintf(opts.IO.Out, "germinator %s (%s) %s\n",
		version.Version, version.Commit, version.Date)
	return nil
}
