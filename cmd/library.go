package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryCommand creates the library command with subcommands.
func NewLibraryCommand(cfg *CommandConfig) *cobra.Command {
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
  create     Create resources in the library`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&libraryPath, "library", "", "Path to library directory (default: ~/.config/germinator/library/)")

	cmd.AddCommand(NewLibraryResourcesCommand(cfg, &libraryPath))
	cmd.AddCommand(NewLibraryPresetsCommand(cfg, &libraryPath))
	cmd.AddCommand(NewLibraryShowCommand(cfg, &libraryPath))
	cmd.AddCommand(NewLibraryInitCommand(cfg))
	cmd.AddCommand(NewLibraryAddCommand(cfg, &libraryPath))
	cmd.AddCommand(NewLibraryCreateCommand(cfg, &libraryPath))

	return cmd
}

// NewLibraryResourcesCommand creates the library resources subcommand.
func NewLibraryResourcesCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "resources",
		Short: "List all resources in the library",
		Long: `List all resources in the library, grouped by type.

Resources are displayed in sections: Skills, Agents, Commands, Memory.

Example:
  germinator library resources
  germinator library resources --library /path/to/library`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Find library
			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", path)

			// Load library
			lib, err := library.LoadLibrary(path)
			if err != nil {
				return fmt.Errorf("loading library: %w", err)
			}

			// List resources
			output := formatResourcesList(lib)
			_, _ = fmt.Fprint(c.OutOrStdout(), output)
			return nil
		},
	}
}

// NewLibraryPresetsCommand creates the library presets subcommand.
func NewLibraryPresetsCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "presets",
		Short: "List all presets in the library",
		Long: `List all presets in the library with their descriptions and resources.

Example:
  germinator library presets
  germinator library presets --library /path/to/library`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Find library
			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", path)

			// Load library
			lib, err := library.LoadLibrary(path)
			if err != nil {
				return fmt.Errorf("loading library: %w", err)
			}

			// List presets
			output := formatPresetsList(lib)
			_, _ = fmt.Fprint(c.OutOrStdout(), output)
			return nil
		},
	}
}

// NewLibraryShowCommand creates the library show subcommand.
func NewLibraryShowCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <ref>",
		Short: "Display details of a resource or preset",
		Long: `Display details of a resource or preset.

For resources, use the format: type/name (e.g., skill/commit)
For presets, use the preset/ prefix (e.g., preset/git-workflow)

Examples:
  germinator library show skill/commit
  germinator library show preset/git-workflow
  germinator library show agent/reviewer --library /path/to/library`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			ref := args[0]

			// Find library
			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", path)

			// Load library
			lib, err := library.LoadLibrary(path)
			if err != nil {
				return fmt.Errorf("loading library: %w", err)
			}

			// Check if it's a preset reference
			if len(ref) > 7 && ref[:7] == "preset/" {
				presetName := ref[7:]
				output, err := formatPresetDetails(lib, presetName)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(c.OutOrStdout(), output)
				return nil
			}

			// Validate resource reference format
			if _, _, err := library.ParseRef(ref); err != nil {
				return fmt.Errorf("invalid reference format: %s (expected type/name or preset/name)", ref)
			}

			// Show resource details
			output, err := formatResourceDetails(lib, ref)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(c.OutOrStdout(), output)
			return nil
		},
	}

	// Add positional completion for 'library show <ref>'
	carapace.Gen(cmd).PositionalCompletion(
		actionLibraryRefs(cmd),
	)

	return cmd
}
