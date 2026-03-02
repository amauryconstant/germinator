package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command with all subcommands registered.
func NewRootCommand(cfg *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "germinator",
		Short: "A configuration adapter for AI coding assistant documents",
		Long: `Germinator is a configuration adapter that transforms AI coding assistant
documents (commands, memory, skills, agents) between platforms. It uses a canonical
Germinator source format and adapts it for target platforms like Claude Code and OpenCode.`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (use -v or -vv)")

	// Register subcommands
	cmd.AddCommand(NewValidateCommand(cfg))
	cmd.AddCommand(NewAdaptCommand(cfg))
	cmd.AddCommand(NewCanonicalizeCommand(cfg))
	cmd.AddCommand(NewVersionCommand(cfg))
	cmd.AddCommand(NewLibraryCommand(cfg))
	cmd.AddCommand(NewInitCommand(cfg))

	return cmd
}
