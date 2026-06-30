package cmd

import (
	"context"
	"errors"
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
// invocation. IO, Library (lazy: loaded via createPresetLibrary), and
// Ctx come from the Factory; the rest come from parsed flags. The
// Library lazy field is func() so the Factory can cache the heavy
// work (LoadLibrary) per the slice-5 initOptions / slice-6 addOptions
// pattern.
//
// No --output field: design Decision 5 — `library create preset` does
// NOT expose --output (legacy did not have --json). The output is a
// single success line to stdout when the preset is created.
type createPresetOptions struct {
	IO          *iostreams.IOStreams
	Library     func() (*library.Library, error)
	Ctx         context.Context
	Resources   []string
	Description string
	Force       bool
}

// presetWriter is the cmd-side contract for preset creation. It is
// intentionally distinct from `Library` (which would shadow the
// library.Library struct) and from the slice-6 resourceAdder
// interface. The method signature matches the (*library.Library).CreatePreset
// method introduced in this slice.
//
// Unlike the slice-6 resourceAdder / libraryAdapter pattern (which
// wraps stateless package functions into a method-bearing wrapper),
// presetWriter is satisfied directly by *library.Library because
// CreatePreset is now a method on *Library (mirroring the slice-5
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

// errEmptyResources is returned by runCreatePreset when --resources
// is present but empty. The message is crafted to contain the
// "flag needs an argument" substring so cmdutil.ExitCodeFor's
// cobraUsagePrefixes branch maps it to ExitCodeUsage (2) per the
// spec scenario "Empty resources flag fails pre-flight validation".
//
// errEmptyResources is package-level (rather than a fmt.Errorf
// inline) to satisfy perfsprint: there are no format arguments
// and the message is stable across runs.
var errEmptyResources = errors.New("flag needs an argument: --resources (must be non-empty list of refs)")

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
named alias. The referenced resources must already be registered in
the library. Use --force to overwrite an existing preset.

Examples:
  germinator library create preset my-workflow --resources skill/commit,skill/pr
  germinator library create preset dev-setup --resources skill/build,agent/reviewer --description "Development setup"
  germinator library create preset old-preset --resources skill/commit --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := &createPresetOptions{
				IO:          f.IOStreams,
				Library:     createPresetLibrary(f, derefString(libraryPath)),
				Ctx:         c.Context(),
				Resources:   resources,
				Description: description,
				Force:       force,
			}
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

// createPresetLibrary wraps path resolution + load into a single lazy
// closure that callers populate into opts.Library. Mirrors
// cmd.initLibrary (slice-5) and cmd.addLibrary (slice-6) so the
// Factory's per-call path resolution pattern is honored.
//
//   - nil factory => nil loader (tests bypass this layer by passing
//     their own Library closure).
//   - explicitPath == "" + env unset => FindLibrary falls through to
//     the XDG default path.
//
// The Library field in createPresetOptions is typed as the canonical
// `func() (*library.Library, error)` per the task spec; the resolved
// path is captured in the closure.
func createPresetLibrary(f *cmdutil.Factory, explicitPath string) func() (*library.Library, error) {
	if f == nil {
		return nil
	}
	resolved := library.FindLibrary(explicitPath, os.Getenv("GERMINATOR_LIBRARY"))
	return func() (*library.Library, error) {
		// TODO(slice-7): replace f.RootContext with the runF ctx
		// once the Factory pattern supports per-call contexts.
		return library.LoadLibrary(f.RootContext, resolved)
	}
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
//     empty" case ("flag needs an argument" style) so
//     cmdutil.ExitCodeFor's cobraUsagePrefixes branch maps it to
//     exit 2.
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
		return errEmptyResources
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

	_, _ = fmt.Fprintf(opts.IO.Out, "Created preset: %s\n", name)
	return nil
}
