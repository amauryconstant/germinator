package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
)

// NewRootCommand creates the root command with all subcommands registered.
// f is the Factory (composition root for the new architecture) and bridge
// is the LegacyBridge shim that keeps non-migrated commands working until
// slice 7 deletes them.
func NewRootCommand(f *cmdutil.Factory, bridge *LegacyBridge) *cobra.Command {
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

	cmd.AddCommand(NewValidateCommand(f, bridge))
	cmd.AddCommand(NewCmdAdapt(f, nil))
	cmd.AddCommand(NewCanonicalizeCommand(f, bridge))
	cmd.AddCommand(NewVersionCommand(f, bridge))
	cmd.AddCommand(NewLibraryCommand(f, bridge, nil))
	cmd.AddCommand(NewInitCommand(f, bridge))
	cmd.AddCommand(NewCompletionCommand(f, bridge))
	cmd.AddCommand(NewConfigCommand(f, bridge))

	// Initialize carapace for enhanced completions
	carapace.Gen(cmd)

	return cmd
}
