package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/version"
)

// NewVersionCommand creates the version command with dependency injection.
func NewVersionCommand(_ *CommandConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of germinator",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
		},
	}
}
