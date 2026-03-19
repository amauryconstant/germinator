package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/version"
)

// NewVersionCommand creates a new version command.
func NewVersionCommand(_ *CommandConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of germinator",
		Run: func(c *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
		},
	}
}
