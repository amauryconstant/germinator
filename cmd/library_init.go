package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// InitJSONOutput represents JSON output for library init.
type InitJSONOutput struct {
	Path    string `json:"path"`
	DryRun  bool   `json:"dryRun"`
	Created bool   `json:"created"`
}

// InitErrorJSON represents JSON output for library init error.
type InitErrorJSON struct {
	Error string `json:"error"`
	Path  string `json:"path"`
}

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
  germinator library init --force
  germinator library init --json`,
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
				jsonFlag, _ := c.Flags().GetBool("json")
				if jsonFlag {
					errOutput := InitErrorJSON{
						Error: err.Error(),
						Path:  path,
					}
					jsonErr, _ := json.Marshal(errOutput)
					_, _ = fmt.Fprintln(c.OutOrStderr(), string(jsonErr))
				}
				return fmt.Errorf("creating library: %w", err)
			}

			// Output success
			jsonFlag, _ := c.Flags().GetBool("json")
			if jsonFlag {
				output := InitJSONOutput{
					Path:    path,
					DryRun:  opts.dryRun,
					Created: !opts.dryRun,
				}
				jsonOutput, _ := json.Marshal(output)
				_, _ = fmt.Fprintln(c.OutOrStdout(), string(jsonOutput))
			} else {
				if opts.dryRun {
					VerbosePrint(cfg, "Dry run complete - no changes made")
				} else {
					VerbosePrint(cfg, "Library created successfully at: %s", path)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.path, "path", "", "Path to create library (default: ~/.config/germinator/library/)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without creating files")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing library at target path")

	return cmd
}
