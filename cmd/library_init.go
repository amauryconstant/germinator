package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryInitCommand creates the library init subcommand.
func NewLibraryInitCommand(cfg *CommandConfig) *cobra.Command {
	var opts struct {
		path   string
		dryRun bool
		force  bool
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new library directory structure",
		Long: `Create a new library directory structure at the specified path.

Creates a library.yaml file and empty resource directories (skills, agents, commands, memory).
The created library is validated by loading it to ensure structural correctness.

By default, creates at ~/.config/germinator/library/ unless --path is specified.
Returns an error if a library already exists at the target path unless --force is used.

Examples:
  germinator library init
  germinator library init --path /tmp/my-library
  germinator library init --dry-run
  germinator library init --force`,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Default path if not specified
			path := opts.path
			if path == "" {
				path = library.DefaultLibraryPath()
			}

			VerbosePrint(cfg, "Creating library at: %s", path)

			// Create library
			err := library.CreateLibrary(library.CreateOptions{
				Path:   path,
				DryRun: opts.dryRun,
				Force:  opts.force,
			})
			if err != nil {
				return fmt.Errorf("creating library: %w", err)
			}

			if opts.dryRun {
				VerbosePrint(cfg, "Dry run complete - no changes made")
			} else {
				VerbosePrint(cfg, "Library created successfully at: %s", path)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.path, "path", "", "Path to create library (default: ~/.config/germinator/library/)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without creating files")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing library at target path")

	return cmd
}
