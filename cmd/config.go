package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
)

// NewConfigCommand creates the config command group with init and
// validate subcommands. The Factory is forwarded to each subcommand
// constructor (slice-8 migration to flat per-command files).
func NewConfigCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Scaffold and validate germinator configuration files",
		Long: `Commands for working with germinator configuration files:

  germinator config init      Scaffold a new config file with documented fields
  germinator config validate   Validate an existing config file`,
	}

	cmd.AddCommand(NewCmdConfigInit(f, nil))
	cmd.AddCommand(NewCmdConfigValidate(f, nil))

	return cmd
}
