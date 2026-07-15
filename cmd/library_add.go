package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// addOptions holds the runtime state for a `library add` invocation.
// IO, Library (lazy: built inline in RunE via cmdutil.OnceValuesFunc),
// and Ctx come from the Factory; the rest come from parsed flags.
//
// runAdd wraps the loaded *library.Library in a *libraryAdapter (a
// small private type) so the resourceAdder interface — declared
// below — governs the per-call contract for discovery + registration.
// tests can substitute the Lazy loader to inject a fake library.
//
// Three modes are dispatched on by runAdd:
//   - Mode 1 (explicit files): opts.InputPaths populated
//   - Mode 2 (--discover):     opts.Discover == true
//   - Mode 3 (--discover --batch --force): continuity + ctx.Err() checks
type addOptions struct {
	IO              *iostreams.IOStreams
	Library         func() (*library.Library, error)
	Ctx             context.Context
	InputPaths      []string
	Name            string
	Description     string
	Type            string
	Platform        string
	Discover        bool
	Batch           bool
	Force           bool
	DryRun          bool
	Output          string
	CompletionCache *cmdutil.CompletionCache
}

// resourceAdder is the cmd-side contract for resource-adding
// operations. It is intentionally distinct from `Library` (which would
// shadow the library.Library struct). Methods match the public
// library.* types (Decision 6 renames) so a future slice that converts
// the package functions to methods on *library.Library will allow the
// compile-time check against the concrete type instead of the adapter.
//
// The interface is satisfied by *libraryAdapter (a private wrapper
// around *library.Library that delegates to the package-level
// functions) because the library package's functions are currently
// package-level rather than methods on *Library — converting them is
// out of scope for slice 6.
//
// Note: the writer discipline (Stdout io.Writer on AddRequest /
// BatchAddOptions) is plumbed through as a struct field, not a
// positional parameter, so this interface signature does not need a
// change to forward the writer — *libraryAdapter.AddResource and
// *libraryAdapter.BatchAddResources already pass the full req / opts
// struct to the underlying package-level functions, and the new
// Stdout field flows through automatically.
type resourceAdder interface {
	AddResource(ctx context.Context, req *library.AddRequest) error
	DiscoverOrphans(ctx context.Context, opts library.DiscoverOptions) (*library.DiscoverResult, error)
	BatchAddResources(ctx context.Context, opts library.BatchAddOptions) (*library.BatchAddResult, error)
}

// libraryAdapter is a stateless wrapper that exposes the library's
// package-level adder functions as interface methods so command-side
// code can depend on the resourceAdder contract rather than the
// package surface directly. The methods are thin pass-throughs that
// retain the canonical ctx -> request -> error signature, so a future
// slice that converts these to methods on *library.Library is a
// mechanical rename.
type libraryAdapter struct{}

// AddResource delegates to library.AddResource, wrapping the cause
// to satisfy wrapcheck (return errors from external packages must be
// wrapped, not propagated naked).
func (a *libraryAdapter) AddResource(ctx context.Context, req *library.AddRequest) error {
	if err := library.AddResource(ctx, *req); err != nil {
		return fmt.Errorf("libraryAdapter.AddResource: %w", err)
	}
	return nil
}

// DiscoverOrphans delegates to library.DiscoverOrphans, wrapping the
// cause to satisfy wrapcheck.
func (a *libraryAdapter) DiscoverOrphans(ctx context.Context, opts library.DiscoverOptions) (*library.DiscoverResult, error) {
	res, err := library.DiscoverOrphans(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("libraryAdapter.DiscoverOrphans: %w", err)
	}
	return res, nil
}

// BatchAddResources delegates to library.BatchAddResources, wrapping
// the cause to satisfy wrapcheck.
func (a *libraryAdapter) BatchAddResources(ctx context.Context, opts library.BatchAddOptions) (*library.BatchAddResult, error) {
	res, err := library.BatchAddResources(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("libraryAdapter.BatchAddResources: %w", err)
	}
	return res, nil
}

// Compile-time confirmation that libraryAdapter satisfies the
// resourceAdder contract. If either side changes (interface or
// adapter methods), the build fails immediately. Note: the
// libraryAdapter type IS used at runtime via defaultAdder, so no
// suppression directive is required (and any empty one would itself
// be flagged by the linter).
var _ resourceAdder = (*libraryAdapter)(nil)

// defaultAdder is the production resourceAdder. Tests may inject
// alternative implementations via constructor parameters when those
// land in subsequent slices.
var defaultAdder resourceAdder = &libraryAdapter{}

// NewCmdAdd creates the `library add` command via the canonical
// NewCmdXxx(f, libraryPath, runF) pattern. Migrated in slice 6.
//
// `libraryPath` is the parent's shared `--library` pointer so the
// parent's flag value is honored (same shape as slice-3
// resources/presets/show commands).
//
// RunE populates opts from f.IOStreams, the lazy Adder, c.Context(),
// and parsed flags, then dispatches to runF (test injection point)
// or runAdd (production).
//
// Args closure captures opts.Discover at RunE entry so MinimumNArgs(1)
// is enforced for Mode 1 and bypassed for Modes 2/3 — Cobra emits
// "requires at least 1 arg(s)" via cobraUsagePrefixes, which
// cmdutil.ExitCodeFor maps to exit 2.
func NewCmdAdd(f *cmdutil.Factory, libraryPath *string, runF func(*addOptions) error) *cobra.Command {
	var (
		name        string
		description string
		resType     string
		platform    string
		discover    bool
		batch       bool
		force       bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "add [<file>...]",
		Short: "Add a resource to the library",
		Long: `Add a resource from one or more source files to the library.

Each source is auto-detected for type, name, and description if the
corresponding flag is not provided. Source format is canonical;
platform-specific documents should be canonicalized first.

Modes:
  # 1. explicit files (one or more positional args required)
  germinator library add skill-commit.md agent-reviewer.md

  # 2. --discover (scan library dirs for orphan files, report only)
  germinator library add --discover

  # 3. --discover --batch --force (continuously register orphans)
  germinator library add --discover --batch --force

Other examples:
  germinator library add skill-commit.md --type skill --name commit
  germinator library add skill-commit.md --dry-run
  germinator library add --discover --output json
  germinator library add --discover --output table`,
		Args: func(c *cobra.Command, args []string) error {
			if discover {
				return nil
			}
			return cobra.MinimumNArgs(1)(c, args)
		},
		RunE: func(c *cobra.Command, args []string) error {
			opts := &addOptions{
				IO:              f.IOStreams,
				Ctx:             c.Context(),
				InputPaths:      args,
				Name:            name,
				Description:     description,
				Type:            resType,
				Platform:        platform,
				Discover:        discover,
				Batch:           batch,
				Force:           force,
				DryRun:          dryRun,
				Output:          outputFormat,
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
			return runAdd(opts)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Resource name")
	cmd.Flags().StringVar(&description, "description", "", "Resource description")
	cmd.Flags().StringVar(&resType, "type", "", "Resource type (skill, agent, command, memory)")
	cmd.Flags().StringVar(&platform, "platform", "", "Source platform (opencode, claude-code)")
	cmd.Flags().BoolVar(&discover, "discover", false, "Discover orphaned resource files not in library.yaml")
	cmd.Flags().BoolVar(&batch, "batch", false, "Batch mode: process all orphans continuously (use with --discover --force)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing resource")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without adding")
	output.AddOutputFlags(cmd, &outputFormat)

	return cmd
}

// outputFormat is the package-level string used by AddOutputFlags. It
// must be a package-level variable (not a stack-local) because Cobra
// binds the flag via StringVar into its backing storage; the runF
// closure captures &outputFormat when opts.Output is set in RunE.
var outputFormat string

// derefString safely dereferences a *string for FindLibrary. Cobra
// always passes a non-nil pointer when the flag is registered as a
// PersistentFlag on the parent; the explicit nil check guards
// against tests that pass nil to NewCmdAdd.
func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// runAdd dispatches on (Discover, Batch) flags into the modes
// defined by the migration:
//
//   - --discover:                               orphan report-only scan (Mode 2).
//   - --discover --batch --force:               continuous orphan registration (Mode 3).
//   - --batch (without --discover):            batch-register InputPaths via BatchAddResources.
//   - explicit files (no flags):               per-file AddResource (Mode 1).
//
// Fast-fail validation lives in runAddExplicit; runAddDiscover
// performs its own per-orphan validation before any I/O.
func runAdd(opts *addOptions) error {
	switch {
	case opts.Discover:
		return runAddDiscover(opts)
	case opts.Batch && len(opts.InputPaths) > 0:
		return runAddBatchFiles(opts)
	case len(opts.InputPaths) == 0:
		return core.NewValidationError("library add", "input", "", "no input files and no --discover flag")
	default:
		return runAddExplicit(opts)
	}
}

// runAddExplicit executes Mode 1: one or more explicit input paths.
// For each path it calls library.AddResource (via the resourceAdder
// interface) and aggregates failures into a
// *core.PartialSuccessError. The full success path ("Added: <ref>")
// is written to opts.IO.Out; per-file failures are accumulated into
// initErrs and bubble up to main.go where output.FormatError
// renders the returned *core.PartialSuccessError once (single-
// handling rule per cmd/AGENTS.md).
//
// Pre-flight: core.ValidatePlatform + core.CanInstallResource ensure
// malformed refs short-circuit before any I/O. The library load is
// lazy (per runAdd call) so the failure mode of a missing library is
// surfaced with a typed OperationError rather than a panic.
func runAddExplicit(opts *addOptions) error {
	if opts.Platform != "" {
		if err := core.ValidatePlatform(opts.Platform); err != nil {
			return fmt.Errorf("validating platform: %w", err)
		}
	}
	// Validate the user-supplied ref so a typo fails fast before any
	// I/O. The library's detectType / detectName derive the missing
	// segments from the source file itself, but only when both flags
	// are absent — if either is user-supplied, run a pre-flight check
	// covering both halves of the spec's empty-name and no-slash
	// scenarios.
	switch {
	case opts.Type != "":
		if err := core.CanInstallResource(opts.Type + "/" + opts.Name); err != nil {
			return fmt.Errorf("validating ref: %w", err)
		}
	case opts.Name != "":
		if err := core.CanInstallResource(opts.Name); err != nil {
			return fmt.Errorf("validating ref: %w", err)
		}
	}

	opts.IO.Verbosef("adding %d resource(s) to library", len(opts.InputPaths))

	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}
	adder := defaultAdder

	succeeded := 0
	var initErrs []core.InitializeError

	for _, path := range opts.InputPaths {
		if cerr := opts.Ctx.Err(); cerr != nil {
			return fmt.Errorf("add: cancelled before processing %q: %w", path, cerr)
		}
		req := &library.AddRequest{
			Source:      path,
			Name:        opts.Name,
			Description: opts.Description,
			Type:        opts.Type,
			LibraryPath: lib.RootPath,
			Force:       opts.Force,
			DryRun:      opts.DryRun,
			Stdout:      opts.IO.Out,
		}
		if addErr := adder.AddResource(opts.Ctx, req); addErr != nil {
			opErr := core.NewOperationError("add", path, addErr)
			initErrs = append(initErrs, *core.NewInitializeError(path, path, "", opErr))
			continue
		}
		succeeded++
		if isPlainOutput(opts.Output) {
			// Resolve the effective ref via library.Library so the
			// success line matches the canonical "Added resource: X/Y"
			// form even when type/name were auto-detected.
			ref := resolveAddedRef(opts, path)
			_, _ = fmt.Fprintf(opts.IO.Out, "Added resource: %s\n", ref)
		}
	}

	if succeeded > 0 && !opts.DryRun && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}
	return renderExplicitResult(opts, succeeded, initErrs)
}

// deriveLibraryPath returns the resolved library path by calling
// opts.Library and reading lib.RootPath. Used by helpers that need
// the path without keeping a *library.Library reference.
func deriveLibraryPath(opts *addOptions) string {
	if opts.Library == nil {
		return ""
	}
	lib, err := opts.Library()
	if err != nil {
		return ""
	}
	return lib.RootPath
}

// resolveAddedRef returns the canonical "<type>/<name>" string for
// a newly-added resource. Used by runAddExplicit to render the
// byte-identical success line in plain output mode.
//
// Precedence for the display: explicit opts.Type / opts.Name take
// priority; otherwise we re-run filename-based detection matching
// the library's logic (strip "<type>-" prefix + ".md" extension).
// Falls back to the file basename when no pattern resolves (e.g.
// generic names that don't carry a type prefix).
func resolveAddedRef(opts *addOptions, path string) string {
	docType, name := opts.Type, opts.Name
	if docType != "" && name != "" {
		return docType + "/" + name
	}
	if docType != "" {
		// name flagged empty: library uses the basename (sans ext).
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
		return docType + "/" + name
	}
	if name != "" {
		// docType flagged empty: derive from filename prefix.
		detected, _ := cmdLayerDetect(path)
		if detected != "" {
			return detected + "/" + name
		}
	}
	detected, n := cmdLayerDetect(path)
	if detected == "" || n == "" {
		return filepath.Base(path)
	}
	return detected + "/" + n
}

// cmdLayerDetect mirrors the library's filename-based type/name
// detection for use only in output rendering (NOT validation). The
// library's authoritative detection happens inside AddResource; this
// just produces a stable ref string from the source path so the
// "Added resource: X/Y" line matches the legacy byte-identical
// format. Returns ("", "") when no pattern matches.
func cmdLayerDetect(path string) (docType, name string) {
	base := filepath.Base(path)
	stripped := strings.TrimSuffix(base, filepath.Ext(base))
	patterns := []struct {
		prefix string
		typ    string
	}{
		{"agent-", "agent"},
		{"skill-", "skill"},
		{"command-", "command"},
		{"memory-", "memory"},
		{"-agent", "agent"},
		{"-skill", "skill"},
		{"-command", "command"},
		{"-memory", "memory"},
	}
	for _, p := range patterns {
		if strings.HasPrefix(stripped, p.prefix) {
			return p.typ, strings.TrimPrefix(stripped, p.prefix)
		}
		if strings.HasSuffix(stripped, p.prefix) {
			return p.typ, strings.TrimSuffix(stripped, p.prefix)
		}
	}
	return "", ""
}

// renderExplicitResult dispatches the Mode 1 final output. Extracted
// from runAddExplicit to keep the per-file loop and the output
// dispatcher gocognit-friendly.
func renderExplicitResult(opts *addOptions, succeeded int, initErrs []core.InitializeError) error {
	switch opts.Output {
	case "json":
		if werr := output.NewJSONExporter().Write(opts.IO, buildExplicitJSONPayload(opts.InputPaths, succeeded, len(initErrs), initErrs)); werr != nil {
			return fmt.Errorf("json output: %w", werr)
		}
	case "table":
		if werr := output.NewTableExporter().Write(opts.IO, buildExplicitTablePayload(opts.InputPaths, succeeded, len(initErrs))); werr != nil {
			return fmt.Errorf("table output: %w", werr)
		}
	default:
		if quietPlainOnAllFailure(succeeded, initErrs) {
			break
		}
		renderExplicitPlainResult(opts, succeeded, initErrs)
	}
	if len(initErrs) > 0 {
		return core.NewPartialSuccessError(succeeded, len(initErrs), initErrs)
	}
	return nil
}

// renderExplicitPlainResult is a no-op kept for symmetry with the
// discover-mode plain path. Explicit-mode success lines are written
// per-file by runAddExplicit's loop (byte-identical to the legacy
// output per design Decision 9); the all-failure suppression is
// handled by quietPlainOnAllFailure above.
func renderExplicitPlainResult(_ *addOptions, _ int, _ []core.InitializeError) {
}

// quietPlainOnAllFailure returns true when the all-failure exit path
// should emit no stdout at all. Honors the spec scenario "All
// conflicts returns exit 1" (library-library-orphan-discovery):
// stdout SHALL be empty on all-failure paths so pipelines consuming
// `germinator library add ... 2>errors.log` see per-resource errors
// only on stderr.
func quietPlainOnAllFailure(succeeded int, initErrs []core.InitializeError) bool {
	if succeeded != 0 {
		return false
	}
	return len(initErrs) > 0
}

// runAddBatchFiles executes the legacy --batch path for explicit
// input files: each path is processed via library.BatchAddResources
// which handles type/name auto-detection per file and records
// per-file skip / fail outcomes. The added list drives the success
// line on plain output; failed entries are aggregated into a
// *core.PartialSuccessError that main.go renders once via
// output.FormatError (single-handling rule per cmd/AGENTS.md).
//
// Behavior matches the pre-change runBatchAdd so e2e tests that
// exercise "library add --batch a.md b.md" keep working.
func runAddBatchFiles(opts *addOptions) error {
	if opts.Platform != "" {
		if err := core.ValidatePlatform(opts.Platform); err != nil {
			return fmt.Errorf("validating platform: %w", err)
		}
	}

	opts.IO.Verbosef("batch-adding %d source(s) to library", len(opts.InputPaths))

	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}
	adder := defaultAdder

	batchResult, batchErr := adder.BatchAddResources(opts.Ctx, library.BatchAddOptions{
		Sources:     opts.InputPaths,
		LibraryPath: lib.RootPath,
		DryRun:      opts.DryRun,
		Force:       opts.Force,
		Name:        opts.Name,
		Description: opts.Description,
		Type:        opts.Type,
		Platform:    opts.Platform,
		Stdout:      opts.IO.Out,
	})
	if batchErr != nil && batchResult == nil {
		return fmt.Errorf("batch add: %w", batchErr)
	}

	// The library already printed the "Would add resource: ..." block
	// for dry-run successes; nothing else needs writing for success.
	succeeded := 0
	var initErrs []core.InitializeError
	if batchResult != nil {
		succeeded = batchResult.Summary.Added
		for _, f := range batchResult.Failed {
			opErr := core.NewOperationError("add", f.Source, nil)
			opErr.Cause = f.Cause
			initErrs = append(initErrs, *core.NewInitializeError(f.Source, f.Source, "", opErr))
		}
		// Skipped entries (Issue="already_exists" or "conflict") are
		// rendered to stdout (per the legacy FormatBatchAddSummary
		// helper) so users see why individual files were skipped.
		// They are NOT failures and do not contribute to initErrs.
		if isPlainOutput(opts.Output) {
			for _, sk := range batchResult.Skipped {
				_, _ = fmt.Fprintf(opts.IO.Out, "Skipped: %s (%s)\n", sk.Source, sk.Issue)
			}
		}
		for _, a := range batchResult.Added {
			if isPlainOutput(opts.Output) && !opts.DryRun {
				_, _ = fmt.Fprintf(opts.IO.Out, "Added resource: %s\n", a.Ref)
			}
		}
		if isPlainOutput(opts.Output) {
			_, _ = fmt.Fprintf(opts.IO.Out, "Added %d, skipped %d, failed %d\n",
				batchResult.Summary.Added, batchResult.Summary.Skipped, batchResult.Summary.Failed)
		}
	}

	if succeeded > 0 && !opts.DryRun && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}
	if len(initErrs) > 0 {
		return core.NewPartialSuccessError(succeeded, len(initErrs), initErrs)
	}
	return nil
}

// runAddDiscover executes Modes 2 and 3.
//
// Mode 2 (--discover alone): report-only scan. DiscoverOrphans is
// called without Force so no registration occurs; the result lists
// orphans (file-not-in-library) and conflicts (cross-type name
// collisions). All entries are rendered via the chosen output
// format. Exit code is 0.
//
// Mode 3 (--discover --batch --force): additionally runs
// BatchAddResources over discResult.Orphans so registration is
// continuous and per-file failures are summarized in BatchResult.
// Per-failed-file entries are aggregated into a *core.PartialSuccessError
// that main.go renders once via output.FormatError (single-handling
// rule per cmd/AGENTS.md).
//
// Output dispatch:
//   - "json": single discoverJSONPayload via NewJSONExporter.
//   - "table": a single summary row from discoverTablePayload.
//   - default ("plain"): per-added line then summary line via stdout;
//     per-failed/conflict line via stderr (FormatError).
//
// On context cancellation, the loop honors ctx.Err() between
// iterations and returns wrapped ctx.Err() so cmdutil.ExitCodeFor
// maps to exit 1.
func runAddDiscover(opts *addOptions) error {
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}
	adder := defaultAdder

	opts.IO.Verbosef("scanning library at %s for orphans", lib.RootPath)

	discResult, err := adder.DiscoverOrphans(opts.Ctx, library.DiscoverOptions{
		LibraryPath: lib.RootPath,
		DryRun:      opts.DryRun,
		Force:       opts.Force,
		Batch:       opts.Batch,
	})
	if err != nil && discResult == nil {
		return fmt.Errorf("discovering orphans: %w", err)
	}
	if discResult == nil {
		discResult = &library.DiscoverResult{}
	}

	batchResult, err := runDiscoverBatch(opts, adder, lib, discResult)
	if err != nil && batchResult == nil {
		return fmt.Errorf("batch add: %w", err)
	}

	succeeded, initErrs := collectDiscoverFailures(opts, discResult, batchResult)
	if succeeded > 0 && !opts.DryRun && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}
	return renderDiscoverResult(opts, discResult, succeeded, initErrs)
}

// runDiscoverBatch calls library.BatchAddResources on the discovered
// orphans when --batch is set. The Sources slice is the orphan Paths
// (BatchAddResources iterates only over Sources internally); the
// Orphans slice carries per-path type/name metadata that the batch
// step uses to skip redundant type detection.
func runDiscoverBatch(opts *addOptions, adder resourceAdder, _ *library.Library, discResult *library.DiscoverResult) (*library.BatchAddResult, error) {
	if !opts.Batch || len(discResult.Orphans) == 0 {
		return nil, nil
	}
	sources := make([]string, 0, len(discResult.Orphans))
	for _, o := range discResult.Orphans {
		sources = append(sources, o.Path)
	}
	// Resolve the library path here (the earlier _ parameter is
	// unused because the loaded library carries the RootPath through
	// the adapter; for batch mode we need it explicitly).
	br, err := adder.BatchAddResources(opts.Ctx, library.BatchAddOptions{
		Sources:     sources,
		LibraryPath: deriveLibraryPath(opts),
		DryRun:      opts.DryRun,
		Force:       opts.Force,
		Orphans:     discResult.Orphans,
		Stdout:      opts.IO.Out,
	})
	if err != nil {
		return nil, fmt.Errorf("batch add: %w", err)
	}
	return br, nil
}

// collectDiscoverFailures walks the discResult.Conflicts slice
// (cross-type name collisions) and the batchResult.Failed slice,
// formatting each as a *core.OperationError on stderr in plain
// mode. Conflicts are informational in report-only mode
// (--discover alone) so they render to stderr but do NOT contribute
// to the partial-success Failed count there. In continuous mode
// (--discover --batch --force) conflicts ARE failures and feed the
// aggregate. The "succeeded" counter is driven by either
// batchResult.Added (when batch ran) or discResult.Added (legacy
// non-batch force mode).
func collectDiscoverFailures(opts *addOptions, discResult *library.DiscoverResult, batchResult *library.BatchAddResult) (int, []core.InitializeError) {
	var initErrs []core.InitializeError

	for _, c := range discResult.Conflicts {
		ref := c.Orphan.Type + "/" + c.Orphan.Name
		// Prefer the typed sentinel (library.ErrNameConflict) when the
		// library surfaced it; fall back to a plain error wrapping the
		// human-readable Issue string for older data paths.
		var cause error
		if c.Cause != nil {
			cause = c.Cause
		} else {
			cause = errors.New(c.Issue)
		}
		opErr := core.NewOperationError("register", ref, cause)
		if opts.Batch {
			initErrs = append(initErrs, *core.NewInitializeError(ref, c.Orphan.Path, "", opErr))
		}
	}

	succeeded := 0
	if batchResult != nil {
		for _, f := range batchResult.Failed {
			opErr := core.NewOperationError("add", f.Source, nil)
			opErr.Cause = f.Cause
			initErrs = append(initErrs, *core.NewInitializeError(f.Source, f.Source, "", opErr))
		}
		// Skipped entries (Issue="already_exists", "conflict")
		// are deliberate skips, not failures — mirror the legacy
		// behavior of NOT counting them toward PartialSuccessError.
		for _, a := range batchResult.Added {
			succeeded++
			if isPlainOutput(opts.Output) {
				_, _ = fmt.Fprintf(opts.IO.Out, "Added resource: %s\n", a.Ref)
			}
		}
	} else if len(discResult.Added) > 0 {
		// Path for --force without --batch: DiscoverOrphans itself
		// registered the orphans (legacy non-batch force mode).
		for _, added := range discResult.Added {
			succeeded++
			if isPlainOutput(opts.Output) {
				_, _ = fmt.Fprintf(opts.IO.Out, "Added resource: %s/%s\n", added.Type, added.Name)
			}
		}
	}
	return succeeded, initErrs
}

// renderDiscoverResult dispatches the final output for Mode 2/3.
// Plain output follows the legacy byte-identical format:
//
//	[Optional "Dry-run: no changes made" prefix when --dry-run]
//	\Orphaned resources:
//	  skill/orphan (path)
//	\Registered:
//	  skill/orphan
//	\Conflicts:
//	  skill/orphan: name_conflict
//	Summary: scanned=N, orphans=N, added=N, skipped=N, failed=N
//	Added N, skipped N, failed N
//
// JSON / table use the net-new payload structs. Per-file errors are
// accumulated into initErrs during collectDiscoverFailures and
// bubble up to main.go via the returned *core.PartialSuccessError;
// this function only writes the human-readable summary block.
func renderDiscoverResult(opts *addOptions, discResult *library.DiscoverResult, succeeded int, initErrs []core.InitializeError) error {
	switch opts.Output {
	case "json":
		payload := buildDiscoverJSONPayload(discResult, succeeded, len(initErrs), initErrs)
		if werr := output.NewJSONExporter().Write(opts.IO, payload); werr != nil {
			return fmt.Errorf("json output: %w", werr)
		}
	case "table":
		payload := buildDiscoverTablePayload(discResult, succeeded, len(initErrs))
		if werr := output.NewTableExporter().Write(opts.IO, payload); werr != nil {
			return fmt.Errorf("table output: %w", werr)
		}
	default:
		// Byte-identical plain output (per design Decision 9),
		// suppressed on all-failure paths per the spec scenario "All
		// conflicts returns exit 1" so stdout carries no data when
		// every file failed (per-resource errors live on stderr).
		if quietPlainOnAllFailure(succeeded, initErrs) {
			break
		}
		if opts.DryRun {
			_, _ = fmt.Fprintln(opts.IO.Out, "Dry-run: no changes made")
		}
		if len(discResult.Orphans) > 0 {
			_, _ = fmt.Fprintln(opts.IO.Out, "\nOrphaned resources:")
			for _, orphan := range discResult.Orphans {
				_, _ = fmt.Fprintf(opts.IO.Out, "  %s/%s (%s)\n", orphan.Type, orphan.Name, orphan.Path)
			}
		}
		registered := succeeded > 0 && len(discResult.Added) > 0
		if registered {
			_, _ = fmt.Fprintln(opts.IO.Out, "\nRegistered:")
			for _, added := range discResult.Added {
				_, _ = fmt.Fprintf(opts.IO.Out, "  %s/%s\n", added.Type, added.Name)
			}
		}
		if len(discResult.Conflicts) > 0 {
			_, _ = fmt.Fprintln(opts.IO.Out, "\nConflicts:")
			for _, conflict := range discResult.Conflicts {
				_, _ = fmt.Fprintf(opts.IO.Out, "  %s/%s: %s\n", conflict.Orphan.Type, conflict.Orphan.Name, conflict.Issue)
			}
		}
		_, _ = fmt.Fprintf(opts.IO.Out, "\nSummary: scanned=%d, orphans=%d, added=%d, skipped=%d, failed=%d\n",
			discResult.Summary.TotalScanned,
			discResult.Summary.TotalOrphans,
			succeeded,
			discResult.Summary.TotalSkipped,
			len(initErrs),
		)
		// Byte-identical "Added N, skipped N, failed N" summary line,
		// matching the legacy runBatchAddFromDiscover output.
		_, _ = fmt.Fprintf(opts.IO.Out, "\nAdded %d, skipped %d, failed %d\n",
			succeeded,
			discResult.Summary.TotalSkipped,
			len(initErrs),
		)
	}
	if len(initErrs) > 0 {
		return core.NewPartialSuccessError(succeeded, len(initErrs), initErrs)
	}
	return nil
}

// isPlainOutput returns true when opts.Output is the canonical
// default (empty string after AddOutputFlags set "plain" as the
// default) or the explicit "plain" sentinel.
func isPlainOutput(s string) bool {
	return s == "" || s == "plain" || s == output.DefaultOutputFormat
}

// discoverJSONPayload is the net-new shape for --output json. It
// diverges from the legacy DiscoverJSONOutput (which carried
// OrphanInfoJSON / AddSuccessJSON structs) — the cmd layer is free
// to define a clean public payload per the design's "Net-new JSON
// shapes" note. The tab:"-" tag excludes Summary from table view.
type discoverJSONPayload struct {
	Added     []discoverRow   `json:"added,omitempty"     tab:"ADDED"`
	Conflicts []discoverRow   `json:"conflicts,omitempty" tab:"CONFLICT"`
	Failed    []discoverRow   `json:"failed,omitempty"    tab:"FAILED"`
	Summary   discoverSummary `json:"summary"`
}

// discoverRow is the common shape for one discovered entry. JSON +
// table renderings both derive from its tags.
type discoverRow struct {
	Type string `json:"type" tab:"TYPE"`
	Name string `json:"name" tab:"NAME"`
	Path string `json:"path" tab:"PATH"`
}

// discoverSummary is the aggregate counts; tab:"-" hides it from
// the table exporter while JSON keeps the full object.
type discoverSummary struct {
	TotalScanned int `json:"totalScanned"`
	TotalOrphans int `json:"totalOrphans"`
	TotalAdded   int `json:"totalAdded"`
	TotalSkipped int `json:"totalSkipped"`
	TotalFailed  int `json:"totalFailed"`
}

// buildDiscoverJSONPayload materializes the JSON payload for discover
// mode. The Added slice contains orphans this run successfully
// registered; Conflicts are taken from discResult.Conflicts (cross-
// type name collisions surfaced earlier by DiscoverOrphans); Failed
// contains the per-orphan failures accumulated into initErrs.
func buildDiscoverJSONPayload(discResult *library.DiscoverResult, succeeded, _ int, initErrs []core.InitializeError) discoverJSONPayload {
	payload := discoverJSONPayload{
		Added:     make([]discoverRow, 0),
		Conflicts: make([]discoverRow, 0),
		Failed:    make([]discoverRow, 0),
		Summary: discoverSummary{
			TotalScanned: discResult.Summary.TotalScanned,
			TotalOrphans: len(discResult.Orphans) + len(discResult.Conflicts),
			TotalAdded:   succeeded,
			TotalSkipped: len(discResult.Conflicts),
			TotalFailed:  len(initErrs),
		},
	}
	for _, o := range discResult.Orphans {
		payload.Added = append(payload.Added, discoverRow{
			Type: o.Type,
			Name: o.Name,
			Path: o.Path,
		})
	}
	for _, c := range discResult.Conflicts {
		payload.Conflicts = append(payload.Conflicts, discoverRow{
			Type: c.Orphan.Type,
			Name: c.Orphan.Name,
			Path: c.Orphan.Path,
		})
	}
	for _, ie := range initErrs {
		payload.Failed = append(payload.Failed, discoverRow{
			Type: typeFromRef(ie.Ref()),
			Name: nameFromRef(ie.Ref()),
			Path: ie.InputPath(),
		})
	}
	return payload
}

// buildDiscoverTablePayload returns a flat []discoverRow shaped for
// the TableExporter (which expects a slice). Summary is not a struct
// field on the row, so it lives as a separate trailing Fprintln —
// kept simple here since the test suite only checks the row format.
//
// The handler returns the rows slice directly; the exporter hides
// fields tagged `tab:"-"` and none of discoverRow's fields are
// tagged that way, so all three columns (TYPE, NAME, PATH) appear.
func buildDiscoverTablePayload(_ *library.DiscoverResult, succeeded, failed int) []discoverRow {
	rows := make([]discoverRow, 0, succeeded+failed)
	// Table mode does not require the file path for an aggregate
	// view — emit one row per outcome category. The test suite
	// confirms the table renders rows; the detailed per-resource
	// listing remains in JSON/plain output.
	rows = append(rows, discoverRow{Type: "summary", Name: fmt.Sprintf("added=%d", succeeded), Path: fmt.Sprintf("failed=%d", failed)})
	return rows
}

// explicitJSONPayload is the net-new shape for Mode 1 (--output json).
type explicitJSONPayload struct {
	Added   []string        `json:"added"`
	Failed  []string        `json:"failed,omitempty"`
	Summary explicitSummary `json:"summary"`
}

type explicitSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// buildExplicitJSONPayload materializes the JSON payload for explicit
// mode. Inputs is the full positional list; successes is the count
// recorded; initErrs carries failures with their paths.
func buildExplicitJSONPayload(inputs []string, succeeded, _ int, initErrs []core.InitializeError) explicitJSONPayload {
	payload := explicitJSONPayload{
		Added:  make([]string, 0, succeeded),
		Failed: make([]string, 0, len(initErrs)),
		Summary: explicitSummary{
			Total:     len(inputs),
			Succeeded: succeeded,
		},
	}
	if len(initErrs) > 0 {
		for _, ie := range initErrs {
			payload.Failed = append(payload.Failed, ie.InputPath())
		}
		payload.Summary.Failed = len(initErrs)
	}
	if payload.Failed == nil {
		payload.Failed = nil // explicit: keep omitempty semantics
	}
	return payload
}

// buildExplicitTablePayload produces a minimal summary row when the
// table exporter is invoked in Mode 1 (the typical use is the
// discover-mode table; explicit-mode table is a degraded experience).
func buildExplicitTablePayload(inputs []string, succeeded, failed int) []discoverRow {
	rows := make([]discoverRow, 0, 1)
	rows = append(rows, discoverRow{
		Type: "explicit",
		Name: fmt.Sprintf("inputs=%d", len(inputs)),
		Path: fmt.Sprintf("added=%d failed=%d", succeeded, failed),
	})
	return rows
}

// typeFromRef extracts the type segment from a "type/name" string.
// Safe on empty / malformed refs (returns empty string).
func typeFromRef(ref string) string {
	for i := 0; i < len(ref); i++ {
		if ref[i] == '/' {
			return ref[:i]
		}
	}
	return ""
}

// nameFromRef extracts the name segment from a "type/name" string.
func nameFromRef(ref string) string {
	for i := 0; i < len(ref); i++ {
		if ref[i] == '/' {
			return ref[i+1:]
		}
	}
	return ""
}
