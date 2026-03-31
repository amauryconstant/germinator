package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryCreateCommand creates the library create subcommand group.
func NewLibraryCreateCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources in the library",
		Long: `Create resources in the library.

Subcommands:
  preset   Create a new preset`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	cmd.AddCommand(NewCreatePresetCommand(cfg, libraryPath))

	return cmd
}

// NewCreatePresetCommand creates the preset subcommand.
func NewCreatePresetCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	var opts struct {
		resources   string
		description string
		force       bool
	}

	cmd := &cobra.Command{
		Use:   "preset <name>",
		Short: "Create a new preset",
		Long: `Create a new preset in the library.

The preset must reference resources that exist in the library.
Use --force to overwrite an existing preset.

Examples:
  germinator library create preset my-workflow --resources skill/commit,skill/pr
  germinator library create preset dev-setup --resources skill/build,agent/reviewer --description "Development setup"
  germinator library create preset old-preset --resources skill/commit --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCreatePreset(c, cfg, libraryPath, &opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.resources, "resources", "", "Comma-separated list of resource references (required)")
	cmd.Flags().StringVar(&opts.description, "description", "", "Preset description")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing preset")
	_ = cmd.MarkFlagRequired("resources")

	// Add flag completions for carapace
	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"resources": actionResources(cmd),
	})

	return cmd
}

// runCreatePreset executes the preset creation logic.
func runCreatePreset(c *cobra.Command, cfg *CommandConfig, libraryPath *string, opts *struct {
	resources   string
	description string
	force       bool
}, name string) error {
	verbosity, _ := c.Flags().GetCount("verbose")
	cfg.Verbosity = Verbosity(verbosity)

	// Validate preset name
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("preset name cannot be empty or whitespace")
	}

	// Parse resources
	resourceRefs := strings.Split(opts.resources, ",")
	for i := range resourceRefs {
		resourceRefs[i] = strings.TrimSpace(resourceRefs[i])
	}

	// Validate resources not empty
	if len(resourceRefs) == 0 || resourceRefs[0] == "" {
		return errors.New("resources list cannot be empty")
	}

	// Discover library path
	envPath := os.Getenv("GERMINATOR_LIBRARY")
	path := library.FindLibrary(*libraryPath, envPath)

	VerbosePrint(cfg, "Using library at: %s", path)

	// Load library
	lib, err := library.LoadLibrary(path)
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	// Check if preset exists
	if library.PresetExists(lib, name) && !opts.force {
		return fmt.Errorf("preset '%s' already exists (use --force to overwrite)", name)
	}

	// Validate all referenced resources exist
	for _, ref := range resourceRefs {
		typ, resName, parseErr := library.ParseRef(ref)
		if parseErr != nil {
			return fmt.Errorf("invalid resource reference '%s': %w", ref, parseErr)
		}

		typeResources, ok := lib.Resources[typ]
		if !ok {
			return fmt.Errorf("resource not found: %s (type '%s' has no resources)", ref, typ)
		}

		if _, exists := typeResources[resName]; !exists {
			return fmt.Errorf("resource not found: %s", ref)
		}
	}

	// Create preset
	preset := library.Preset{
		Name:        name,
		Description: opts.description,
		Resources:   resourceRefs,
	}

	// Add to library in-memory
	if err := library.AddPreset(lib, preset); err != nil {
		return fmt.Errorf("adding preset: %w", err)
	}

	// Save library
	if err := library.SaveLibrary(lib); err != nil {
		return fmt.Errorf("saving library: %w", err)
	}

	// Output success
	output := formatPresetOutput(lib, name)
	_, _ = fmt.Fprint(c.OutOrStdout(), output)

	return nil
}
