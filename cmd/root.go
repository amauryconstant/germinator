package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
)

// NewRootCommand creates the root command with all subcommands
// registered. f is the Factory (composition root for the new
// architecture). All subcommand constructors take f directly.
//
// PersistentPreRunE wires the parsed --verbose flag into
// f.IOStreams.Verbose so every subcommand's Verbosef call honors it.
// Per references/09-logging.md and the cli-verbose-output spec, the
// verbose channel (-v / --verbose, progress disclosure to stderr) is
// independent of the debug channel (--debug / GERMINATOR_DEBUG →
// cfg.Debug → io.SetDebug → slog). Both channels are populated by
// distinct sources and must remain independent.
func NewRootCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "germinator",
		Short: "A configuration adapter for AI coding assistant documents",
		Long: `Germinator is a configuration adapter that transforms AI coding assistant
documents (commands, memory, skills, agents) between platforms. It uses a canonical
Germinator source format and adapts it for target platforms like Claude Code and OpenCode.`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
		// main.go owns all error/usage rendering via output.FormatError;
		// suppress Cobra's built-in copy so users see a single message.
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			if f != nil && f.IOStreams != nil {
				f.IOStreams.Verbose, _ = c.Flags().GetBool("verbose")
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolP("verbose", "v", false, "Increase output verbosity (-v)")

	cmd.AddCommand(NewCmdValidate(f, nil))
	cmd.AddCommand(NewCmdAdapt(f, nil))
	cmd.AddCommand(NewCmdCanonicalize(f, nil))
	cmd.AddCommand(NewCmdVersion(f, nil))
	cmd.AddCommand(NewLibraryCommand(f, nil))
	cmd.AddCommand(NewCmdInit(f, nil))
	cmd.AddCommand(NewCmdCompletion(f, nil))
	cmd.AddCommand(NewConfigCommand(f))

	// Initialize carapace for enhanced completions
	carapace.Gen(cmd)

	return cmd
}
