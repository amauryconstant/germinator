package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryRemoveCommand creates the library remove subcommand.
func NewLibraryRemoveCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a resource or preset from the library",
		Long: `Remove a resource or preset from the library.

Subcommands:
  resource  Remove a resource from the library
  preset    Remove a preset from the library`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	cmd.AddCommand(NewLibraryRemoveResourceCommand(cfg, libraryPath))
	cmd.AddCommand(NewLibraryRemovePresetCommand(cfg, libraryPath))

	return cmd
}

// NewLibraryRemoveResourceCommand creates the library remove resource subcommand.
func NewLibraryRemoveResourceCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource <ref>",
		Short: "Remove a resource from the library",
		Long: `Remove a resource from the library.

Deletes both the physical file and YAML entry.
Errors if any preset references the resource.

Examples:
  germinator library remove resource skill/commit
  germinator library remove resource skill/commit --library /path/to/library
  germinator library remove resource skill/commit --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			ref := args[0]

			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Using library at: %s", path)

			result, err := library.RemoveResource(library.RemoveResourceOptions{
				Ref:         ref,
				LibraryPath: path,
			})
			if err != nil {
				jsonFlag, _ := c.Flags().GetBool("json")
				if jsonFlag {
					errOutput := library.RemoveResourceError{
						Error:        err.Error(),
						Type:         "resource",
						ResourceType: "",
						Name:         ref,
					}
					jsonErr, _ := json.Marshal(errOutput)
					_, _ = fmt.Fprintln(c.OutOrStderr(), string(jsonErr))
				}
				return fmt.Errorf("removing resource: %w", err)
			}

			jsonFlag, _ := c.Flags().GetBool("json")
			if jsonFlag {
				jsonOutput, _ := json.Marshal(result)
				_, _ = fmt.Fprintln(c.OutOrStdout(), string(jsonOutput))
			} else {
				fmt.Printf("Removed resource: %s\n", ref)
			}

			return nil
		},
	}

	return cmd
}

// NewLibraryRemovePresetCommand creates the library remove preset subcommand.
func NewLibraryRemovePresetCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset <name>",
		Short: "Remove a preset from the library",
		Long: `Remove a preset from the library.

Removes only the YAML entry (no physical file to delete).

Examples:
  germinator library remove preset git-workflow
  germinator library remove preset git-workflow --library /path/to/library
  germinator library remove preset git-workflow --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			name := args[0]

			envPath := os.Getenv("GERMINATOR_LIBRARY")
			path := library.FindLibrary(*libraryPath, envPath)

			VerbosePrint(cfg, "Using library at: %s", path)

			result, err := library.RemovePreset(library.RemovePresetOptions{
				Name:        name,
				LibraryPath: path,
			})
			if err != nil {
				jsonFlag, _ := c.Flags().GetBool("json")
				if jsonFlag {
					errOutput := library.RemovePresetError{
						Error: err.Error(),
						Type:  "preset",
						Name:  name,
					}
					jsonErr, _ := json.Marshal(errOutput)
					_, _ = fmt.Fprintln(c.OutOrStderr(), string(jsonErr))
				}
				return fmt.Errorf("removing preset: %w", err)
			}

			jsonFlag, _ := c.Flags().GetBool("json")
			if jsonFlag {
				jsonOutput, _ := json.Marshal(result)
				_, _ = fmt.Fprintln(c.OutOrStdout(), string(jsonOutput))
			} else {
				fmt.Printf("Removed preset: %s\n", name)
			}

			return nil
		},
	}

	return cmd
}
