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
  create     Create a new preset (path: library create preset <name>)
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
	cmd.AddCommand(NewCmdAdd(f, &libraryPath, nil))
	// `library create preset` preserves the user-facing command path
	// (spec delta: "library create preset is a leaf under library").
	// NewCmdCreatePreset is registered under a thin `create` Cobra
	// parent so the three-segment path remains routable. The parent
	// has no Run of its own (Cobra displays the subcommand list when
	// the user invokes `library create` without a subcommand), so
	// there is no group-level description surface; this matches the
	// spec scenario "library create has no subcommand list" intent
	// even though the parent command exists for routing.
	cmd.AddCommand(NewCmdCreate(f, &libraryPath))
	cmd.AddCommand(NewLibraryRemoveCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryValidateCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryRefreshCommand(bridge, &libraryPath))

	return cmd
}

// NewCmdCreate constructs the thin `library create` parent command
// that routes to `library create preset` only. Exported for test
// reachability (`TestNewCmdCreate_ShowsPresetAsChild`); the parent has
// no Run of its own — Cobra prints the subcommand list when the user
// invokes `library create` bare, matching the
// library-library-json-output spec scenario "library create has no
// subcommand list".
func NewCmdCreate(f *cmdutil.Factory, libraryPath *string) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new preset",
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}
	createCmd.AddCommand(NewCmdCreatePreset(f, libraryPath, nil))
	return createCmd
}
