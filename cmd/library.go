package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
)

// NewLibraryCommand creates the library command with subcommands.
func NewLibraryCommand(f *cmdutil.Factory, bridge *LegacyBridge, runF func(*libraryResourcesOptions) error) *cobra.Command {
	var libraryPath string

	cmd := &cobra.Command{
		Use:   "library",
		Short: "Manage canonical resource library",
		Long: `Manage the canonical resource library.

The library contains skills, agents, commands, and memory resources
that can be installed to projects using the init command.

Subcommands:
  resources  List all resources in the library
  presets    List all presets in the library
  show       Display details of a resource or preset
  add        Add a resource to the library
  create     Create resources in the library
  remove     Remove a resource or preset from the library
  validate   Validate library integrity
  refresh    Sync metadata from resource files into library.yaml`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&libraryPath, "library", "", "Path to library directory (default: ~/.config/germinator/library/)")

	cmd.AddCommand(NewCmdResources(f, &libraryPath, runF))
	// NewCmdPresets and NewCmdShow take a per-command runF typed to
	// their own options struct; the parent's runF (typed for
	// *libraryResourcesOptions) cannot be passed through without a
	// type mismatch. The per-command runF is wired by main.go
	// (composition root) per command, not by this parent.
	cmd.AddCommand(NewCmdPresets(f, &libraryPath, nil))
	cmd.AddCommand(NewCmdShow(f, &libraryPath, nil))
	cmd.AddCommand(NewLibraryInitCommand(bridge))
	cmd.AddCommand(NewLibraryAddCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryCreateCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryRemoveCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryValidateCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryRefreshCommand(bridge, &libraryPath))

	return cmd
}
