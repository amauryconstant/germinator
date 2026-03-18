package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/library"
)

// NewInitCommand creates init command for installing resources.
func NewInitCommand(cfg *CommandConfig) *cobra.Command {
	var (
		platform    string
		resources   string
		preset      string
		libraryPath string
		outputDir   string
		dryRun      bool
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Install resources from the library to a project",
		Long: `Install resources from the library to a target project directory.

Resources are transformed from canonical format to the target platform format
and written to platform-specific output paths.

Either --resources or --preset must be specified (mutually exclusive).

Examples:
  # Install specific resources
  germinator init --platform opencode --resources skill/commit,skill/merge-request

  # Install from a preset
  germinator init --platform opencode --preset git-workflow

  # Preview changes without writing
  germinator init --platform opencode --preset git-workflow --dry-run

  # Overwrite existing files
  germinator init --platform opencode --resources skill/commit --force`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Validate platform
			if platform == "" {
				return gerrors.NewConfigError("platform", "", "--platform flag is required").WithSuggestions(library.ValidPlatforms())
			}
			if !library.IsValidPlatform(platform) {
				return gerrors.NewConfigError("platform", platform, "unknown platform").WithSuggestions(library.ValidPlatforms())
			}

			// Validate resources or preset (mutually exclusive)
			if resources != "" && preset != "" {
				return gerrors.NewConfigError("resources/preset", "", "--resources and --preset are mutually exclusive")
			}
			if resources == "" && preset == "" {
				return gerrors.NewConfigError("resources/preset", "", "either --resources or --preset is required")
			}

			// Parse resource list
			var refs []string
			if resources != "" {
				refs = strings.Split(resources, ",")
				for i, r := range refs {
					refs[i] = strings.TrimSpace(r)
				}

				// Validate each reference format
				for _, ref := range refs {
					if err := library.ValidateRef(ref); err != nil {
						return err
					}
				}
			}

			// Find library
			envPath := os.Getenv("GERMINATOR_LIBRARY")
			libPath := library.FindLibrary(libraryPath, envPath)

			VerbosePrint(cfg, "Loading library from: %s", libPath)

			// Load library
			lib, err := library.LoadLibrary(libPath)
			if err != nil {
				return err
			}

			// Resolve preset if specified (preset resolution happens in command layer)
			if preset != "" {
				refs, err = library.ResolvePreset(lib, preset)
				if err != nil {
					return err
				}
			}

			VeryVerbosePrint(cfg, "Installing resources: %s", strings.Join(refs, ", "))

			// Initialize resources using service interface
			results, err := cfg.Services.Initializer.Initialize(context.Background(), &application.InitializeRequest{
				Library:   lib,
				Platform:  platform,
				OutputDir: outputDir,
				Refs:      refs,
				DryRun:    dryRun,
				Force:     force,
			})
			if err != nil {
				// Print any partial results
				for _, r := range results {
					if r.Error != nil {
						fmt.Fprintf(os.Stderr, "Error: %s: %v\n", r.Ref, r.Error)
					}
				}
				return err
			}

			// Output results
			if dryRun {
				_, _ = fmt.Fprint(c.OutOrStdout(), formatDryRunOutput(results))
				_, _ = fmt.Fprintln(c.OutOrStdout(), "\nDry run complete. No files were written.")
			} else {
				_, _ = fmt.Fprint(c.OutOrStdout(), formatSuccessOutput(results))
				_, _ = fmt.Fprintf(c.OutOrStdout(), "\nSuccessfully installed %d resource(s).\n", len(results))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Target platform (required: opencode, claude-code)")
	cmd.Flags().StringVar(&resources, "resources", "", "Comma-separated list of resources to install (e.g., skill/commit,skill/merge-request)")
	cmd.Flags().StringVar(&preset, "preset", "", "Preset name for bundled resources")
	cmd.Flags().StringVar(&libraryPath, "library", "", "Path to library directory (default: ~/.config/germinator/library/)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")

	_ = cmd.MarkFlagRequired("platform")

	// Add flag completions for carapace
	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform":  actionPlatforms(),
		"resources": actionResources(cmd),
		"preset":    actionPresets(cmd),
	})

	return cmd
}
