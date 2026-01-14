package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version of germinator",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("germinator %s (%s) %s\n", version.Version, version.Commit, version.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
