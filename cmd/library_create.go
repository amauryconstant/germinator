package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// createPresetOptions holds the runtime state for a `library create preset`
// invocation. IO, Library (lazy: built inline in RunE via
// cmdutil.OnceValuesFunc), and Ctx come from the Factory; the rest
// come from parsed flags.
//
// No --output field: design Decision 5 — `library create preset` does
// NOT expose --output (legacy did not have --json). The output is a
// single success line to stdout when the preset is created.
type createPresetOptions struct {
	IO              *iostreams.IOStreams
	Library         func() (*library.Library, error)
	Ctx             context.Context
	Resources       []string
	Description     string
	Force           bool
	CompletionCache *cmdutil.CompletionCache
}

// presetWriter is the cmd-side contract for preset creation. It is
// intentionally distinct from `Library` (which would shadow the
// library.Library struct). The method signature matches the
// (*library.Library).CreatePreset method introduced in slice 6.
//
// presetWriter is satisfied directly by *library.Library because
// CreatePreset is a method on *Library (mirroring the slice-5
// ResolvePreset dual form). This removes the need for a stateless
// adapter and keeps the cmd layer free of indirection.
type presetWriter interface {
	CreatePreset(ctx context.Context, req *library.CreatePresetRequest) error
}

// Compile-time confirmation that *library.Library satisfies the
// presetWriter contract. If either side changes (interface or
// (*Library).CreatePreset method), the build fails immediately.
// *library.Library is the live receiver used by runCreatePreset, so
// no suppression directive is required.
var _ presetWriter = (*library.Library)(nil)

// NewCmdCreatePreset creates the `library create preset` command via
// the canonical NewCmdXxx(f, libraryPath, runF) pattern. Migrated in
// slice 6.
//
// `libraryPath` is the parent's shared `--library` pointer so the
// parent's flag value is honored (same shape as slice-3 resources /
// slice-4 presets / slice-6 add commands).
//
// RunE populates opts from f.IOStreams, the lazy Library, c.Context(),
// and parsed flags, then dispatches to runF (test injection point)
// or runCreatePreset (production).
//
// Flags:
//
//	--resources   (required) comma-separated "type/name" refs
//	--description (optional) human-readable description
//	--force       (optional) overwrite an existing preset
//
// No --output flag (design Decision 5).
func NewCmdCreatePreset(f *cmdutil.Factory, libraryPath *string, runF func(*createPresetOptions) error) *cobra.Command {
	var (
		resources   []string
		description string
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "preset <name>",
		Short: "Create a new preset",
		Long: `Create a new preset in the library.

A preset bundles one or more resource references ("type/name") under a
named alias. Referenced resources must already be registered.

Examples:
  germinator library create preset my-workflow --resources skill/commit,skill/pr
  germinator library create preset dev-setup --resources skill/build,agent/reviewer --description "Development setup"
  germinator library create preset old-preset --resources skill/commit --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := &createPresetOptions{
				IO:              f.IOStreams,
				Ctx:             c.Context(),
				Resources:       resources,
				Description:     description,
				Force:           force,
				CompletionCache: f.CompletionCache,
			}
			var cfgPath string
			if f.Config != nil {
				if cfg, cfgErr := f.Config(); cfgErr == nil && cfg != nil {
					cfgPath = cfg.Library
				}
			}
			resolved := library.FindLibrary(derefString(libraryPath), os.Getenv("GERMINATOR_LIBRARY"), cfgPath)
			opts.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) {
				return library.LoadLibrary(c.Context(), resolved)
			})
			if runF != nil {
				return runF(opts)
			}
			return runCreatePreset(opts, args[0])
		},
	}

	cmd.Flags().StringSliceVar(&resources, "resources", nil, "Comma-separated list of resource references (e.g., skill/commit,agent/reviewer)")
	cmd.Flags().StringVar(&description, "description", "", "Preset description")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing preset")

	_ = cmd.MarkFlagRequired("resources")

	return cmd
}

// runCreatePreset executes the preset creation logic. It is the
// production wiring for NewCmdCreatePreset's runF parameter.
//
// Validation order (matches proposal.md / spec.md):
//  1. presetName trim + non-empty check (returns *core.ValidationError
//     so cmdutil.ExitCodeFor maps it to exit 1).
//  2. opts.Resources non-empty check. Cobra's MarkFlagRequired emits
//     a "required flag(s) \"resources\" not set" message when
//     --resources is absent; this branch handles the "flag present but
//     empty" case via an inline core.NewUsageError, a *core.UsageError
//     mapped to ExitCodeUsage (2) by the typed-error dispatch in
//     cmdutil.ExitCodeFor (Phase 3.12 migration; the prior
//     cobraUsagePrefixes substring fallback was dropped in Phase 1;
//     Phase 6 inline construction removed the package-level
//     errEmptyResources var).
//  3. For each ref, core.CanInstallResource pre-flight check. The
//     first malformed ref short-circuits the loop with a
//     *core.ValidationError (exit 1 via default-error case).
//  4. Lazy load the library.
//  5. Delegate to lib.CreatePreset for the authoritative validation
//     + mutation + persistence cycle.
//
// On success, writes "Created preset: <name>" to opts.IO.Out (matches
// the legacy human output contract).
func runCreatePreset(opts *createPresetOptions, presetName string) error {
	name := strings.TrimSpace(presetName)
	if name == "" {
		return core.NewValidationError("library create preset", "name", presetName, "preset name cannot be empty or whitespace")
	}

	if len(opts.Resources) == 0 || opts.Resources[0] == "" {
		return core.NewUsageError("--resources", "must be non-empty list of refs")
	}

	for _, ref := range opts.Resources {
		ref = strings.TrimSpace(ref)
		if err := core.CanInstallResource(ref); err != nil {
			return fmt.Errorf("validating ref %q: %w", ref, err)
		}
	}

	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	if err := lib.CreatePreset(opts.Ctx, &library.CreatePresetRequest{
		Name:        name,
		Description: opts.Description,
		Resources:   opts.Resources,
		Force:       opts.Force,
	}); err != nil {
		return fmt.Errorf("creating preset: %w", err)
	}

	if opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "Created preset: %s\n", name)
	return nil
}
