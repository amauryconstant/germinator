package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// InitializeRequest is the application-side request type, re-exported
// for callers in this file. Imported from internal/application/requests.go.
type InitializeRequest = application.InitializeRequest

// initOptions holds the runtime state for an `init` invocation. IO,
// Ctx, Library, and Initializer come from the Factory; the rest come
// from parsed flags. Library and Initializer are lazy closures so the
// Factory can cache the heavy work (LoadLibrary, NewInitializer) per
// the golang-cli-architecture skill's functional-core principle.
type initOptions struct {
	IO          *iostreams.IOStreams
	Library     func() (*library.Library, error)
	Initializer func() (application.Initializer, error)
	Ctx         context.Context
	LibraryPath string
	Platform    string
	OutputDir   string
	Refs        []string
	Preset      string
	DryRun      bool
	Force       bool
}

// NewCmdInit creates the `init` command via the canonical
// NewCmdXxx(f, runF) pattern. Migrated in slice 5.
//
// RunE populates opts from f.IOStreams, f.Library, f.Initializer,
// c.Context(), and the parsed flags, then dispatches to runF or
// runInit. test injection: tests pass a non-nil runF to capture
// options without invoking runInit.
func NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command {
	var (
		platform    string
		resources   []string
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

Resources are transformed from canonical format to the target platform
format and written to platform-specific output paths.

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
			opts := &initOptions{
				IO:          f.IOStreams,
				Library:     initLibrary(f, libraryPath),
				Initializer: initInitializer(f),
				Ctx:         c.Context(),
				LibraryPath: libraryPath,
				Platform:    platform,
				OutputDir:   outputDir,
				Refs:        resources,
				Preset:      preset,
				DryRun:      dryRun,
				Force:       force,
			}
			if runF != nil {
				return runF(opts)
			}
			return runInit(opts)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Target platform (required: opencode, claude-code)")
	cmd.Flags().StringSliceVar(&resources, "resources", nil, "Comma-separated list of resources to install (e.g., skill/commit,skill/merge-request)")
	cmd.Flags().StringVar(&preset, "preset", "", "Preset name for bundled resources")
	cmd.Flags().StringVar(&libraryPath, "library", "", "Path to library directory (default: ~/.config/germinator/library/)")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory (default: current directory)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")

	_ = cmd.MarkFlagRequired("platform")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform":  actionPlatforms(),
		"resources": actionResources(cmd),
		"preset":    actionPresets(cmd),
	})

	return cmd
}

// initLibrary wraps the Factory's lazy Library field with per-call
// path resolution so --library is honored on each invocation (the
// Factory's own Library uses env-only resolution; the --library flag
// is parented to the command).
func initLibrary(f *cmdutil.Factory, explicitPath string) func() (*library.Library, error) {
	if f == nil {
		return nil
	}
	return func() (*library.Library, error) {
		resolved := library.FindLibrary(explicitPath, os.Getenv("GERMINATOR_LIBRARY"))
		return library.LoadLibrary(resolved)
	}
}

// initInitializer wraps the Factory's lazy application.Initializer
// field. Nil Factory or unset Initializer returns nil so a missing
// initializer surfaces as an error from runInit rather than a nil
// dereference.
func initInitializer(f *cmdutil.Factory) func() (application.Initializer, error) {
	if f == nil || f.Initializer == nil {
		return nil
	}
	return f.Initializer
}

// runInit executes the init logic against the resolved options. It
// is the production wiring for NewCmdInit's runF parameter.
//
// Validation order (matches proposal.md decision matrix):
//  1. Refs XOR Preset (mutex per base spec).
//  2. Platform validated via core.ValidatePlatform.
//  3. If Preset != "", expand via (*Library).ResolvePreset; on miss
//     wrap as *core.NotFoundError so ExitCodeFor returns 2.
//  4. Build InitializeRequest; invoke f.Initializer().
//  5. Count successes/failures from the result slice.
//  6. Render per-resource status; return nil or *core.PartialSuccessError.
func runInit(opts *initOptions) error {
	hasRefs := len(opts.Refs) > 0
	hasPreset := opts.Preset != ""
	if hasRefs && hasPreset {
		return core.NewValidationError("init", "resources/preset", "", "--resources and --preset are mutually exclusive")
	}
	if !hasRefs && !hasPreset {
		return core.NewValidationError("init", "resources/preset", "", "either --resources or --preset is required")
	}

	if err := core.ValidatePlatform(opts.Platform); err != nil {
		return fmt.Errorf("validating platform: %w", err)
	}

	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	refs := opts.Refs
	if hasPreset {
		expanded, rerr := lib.ResolvePreset(opts.Ctx, opts.Preset)
		if rerr != nil {
			return core.NewNotFoundError("preset", opts.Preset)
		}
		refs = expanded
	}

	opts.IO.Verbosef("installing resources: %s", strings.Join(refs, ", "))

	initializer, err := opts.Initializer()
	if err != nil {
		return fmt.Errorf("resolving initializer: %w", err)
	}

	results, err := initializer.Initialize(opts.Ctx, &InitializeRequest{
		Library:   lib,
		Platform:  opts.Platform,
		OutputDir: opts.OutputDir,
		Refs:      refs,
		DryRun:    opts.DryRun,
		Force:     opts.Force,
	})
	if err != nil {
		return fmt.Errorf("initializing resources: %w", err)
	}

	succeeded, failed, initErrs := classifyResults(results)

	renderResults(opts, results)

	switch {
	case failed == 0:
		return nil
	case succeeded == 0:
		partialErr := core.NewPartialSuccessError(0, failed, initErrs)
		output.FormatError(opts.IO, partialErr)
		return partialErr
	default:
		partialErr := core.NewPartialSuccessError(succeeded, failed, initErrs)
		output.FormatError(opts.IO, partialErr)
		return partialErr
	}
}

// classifyResults counts successes/failures from a results slice and
// materializes the matching []core.InitializeError list for the
// partial-success aggregate. Success is derived from Error == nil;
// the core.InitializeResult type does not carry an explicit Succeeded
// field (design §3 implicitly defines it this way).
func classifyResults(results []core.InitializeResult) (succeeded, failed int, errs []core.InitializeError) {
	for _, r := range results {
		if r.Error != nil {
			failed++
			errs = append(errs, *core.NewInitializeError(r.Ref, r.InputPath, r.OutputPath, r.Error))
			continue
		}
		succeeded++
	}
	return succeeded, failed, errs
}

// renderResults writes per-resource status to IO: successes to Out,
// failures to ErrOut via output.FormatError. The overall command exit
// code is determined by runInit's error return, not by what is
// rendered here.
func renderResults(opts *initOptions, results []core.InitializeResult) {
	for _, r := range results {
		if r.Error == nil {
			if opts.DryRun {
				_, _ = fmt.Fprintf(opts.IO.Out, "Would write: %s\n  from: %s\n", r.OutputPath, r.InputPath)
				continue
			}
			_, _ = fmt.Fprintf(opts.IO.Out, "Installed: %s -> %s\n", r.Ref, r.OutputPath)
			continue
		}
		output.FormatError(opts.IO, core.NewInitializeError(r.Ref, r.InputPath, r.OutputPath, r.Error))
	}
	if opts.DryRun && len(results) > 0 {
		_, _ = fmt.Fprintln(opts.IO.Out, "Dry run complete. No files were written.")
	}
	s, f := 0, 0
	for _, r := range results {
		if r.Error == nil {
			s++
			continue
		}
		f++
	}
	_, _ = fmt.Fprintf(opts.IO.Out, "Initialized %d resource(s).", s)
	if f > 0 {
		_, _ = fmt.Fprintf(opts.IO.Out, ", %d failed.", f)
	}
	_, _ = fmt.Fprintln(opts.IO.Out)
}
