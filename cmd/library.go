package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/library"
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
	cmd.AddCommand(NewLibraryPresetsCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryShowCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryInitCommand(bridge))
	cmd.AddCommand(NewLibraryAddCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryCreateCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryRemoveCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryValidateCommand(bridge, &libraryPath))
	cmd.AddCommand(NewLibraryRefreshCommand(bridge, &libraryPath))

	return cmd
}

// NewLibraryPresetsCommand creates the library presets subcommand.
func NewLibraryPresetsCommand(bridge *LegacyBridge, libraryPath *string) *cobra.Command {
	cfg := legacyCfgFrom(bridge)
	return &cobra.Command{
		Use:   "presets",
		Short: "List all presets in the library",
		Long: `List all presets in the library with their descriptions and resources.

Example:
  germinator library presets
  germinator library presets --library /path/to/library
  germinator library presets --json`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", path)

			lib, err := library.LoadLibrary(path)
			if err != nil {
				return fmt.Errorf("loading library: %w", err)
			}

			jsonFlag, _ := c.Flags().GetBool("json")
			if jsonFlag {
				return outputPresetsJSON(c, lib)
			}

			output := formatPresetsList(lib)
			_, _ = fmt.Fprint(c.OutOrStdout(), output)
			return nil
		},
	}
}

// NewLibraryShowCommand creates the library show subcommand.
func NewLibraryShowCommand(bridge *LegacyBridge, libraryPath *string) *cobra.Command {
	cfg := legacyCfgFrom(bridge)
	cmd := &cobra.Command{
		Use:   "show <ref>",
		Short: "Display details of a resource or preset",
		Long: `Display details of a resource or preset.

For resources, use the format: type/name (e.g., skill/commit)
For presets, use the preset/ prefix (e.g., preset/git-workflow)

Examples:
  germinator library show skill/commit
  germinator library show preset/git-workflow
  germinator library show agent/reviewer --library /path/to/library
  germinator library show skill/commit --json
  germinator library show preset/git-workflow --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			ref := args[0]

			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", path)

			lib, err := library.LoadLibrary(path)
			if err != nil {
				return fmt.Errorf("loading library: %w", err)
			}

			jsonFlag, _ := c.Flags().GetBool("json")

			if len(ref) > 7 && ref[:7] == "preset/" {
				presetName := ref[7:]
				if jsonFlag {
					return outputShowPresetJSON(c, lib, presetName)
				}
				output, err := formatPresetDetails(lib, presetName)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(c.OutOrStdout(), output)
				return nil
			}

			if _, _, err := library.ParseRef(ref); err != nil {
				return fmt.Errorf("invalid reference format: %s (expected type/name or preset/name)", ref)
			}

			if jsonFlag {
				return outputShowResourceJSON(c, lib, ref)
			}
			output, err := formatResourceDetails(lib, ref)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(c.OutOrStdout(), output)
			return nil
		},
	}

	carapace.Gen(cmd).PositionalCompletion(
		actionLibraryRefs(cmd),
	)

	return cmd
}
