package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// refreshOptions holds the runtime state for a `library refresh` invocation.
// IO, Library (lazy: loaded via refreshLibrary in RunE), and Ctx
// come from the Factory; the rest come from parsed flags. The Library
// lazy field is func() so the Factory can cache the heavy work
// (LoadLibrary) per the slice-5/6 addOptions / createPresetOptions
// pattern.
type refreshOptions struct {
	IO              *iostreams.IOStreams
	Library         func() (*library.Library, error)
	Ctx             context.Context
	DryRun          bool
	Force           bool
	Output          string
	CompletionCache *cmdutil.CompletionCache
}

// refresherLibrary is the cmd-side contract for refresh operations.
// It is intentionally distinct from Library (which would shadow the
// library.Library struct). The method signature matches the
// (*library.Library).Refresh method introduced in slice 7.0 stage A
// (mirroring the slice-6 (*Library).CreatePreset precedent at
// internal/library/creator.go:176).
//
// Unlike the slice-6 resourceAdder / libraryAdapter pattern (which
// wraps stateless package functions into a method-bearing wrapper),
// refresherLibrary is satisfied directly by *library.Library because
// Refresh is now a method on *Library. This removes the need for a
// stateless adapter and keeps the cmd layer free of indirection.
type refresherLibrary interface {
	Refresh(ctx context.Context, req *library.RefreshRequest) (*library.RefreshResult, error)
}

// Compile-time confirmation that *library.Library satisfies the
// refresherLibrary contract. If either side changes (interface or
// (*Library).Refresh method), the build fails immediately.
// *library.Library is the live receiver used by runRefresh, so no
// suppression directive is required.
var _ refresherLibrary = (*library.Library)(nil)

// NewCmdRefresh creates the `library refresh` command via the canonical
// NewCmdXxx(f, libraryPath, runF) pattern. Migrated in slice 7.
//
// `libraryPath` is the parent's shared `--library` pointer so the
// parent's flag value is honored (same shape as the slice-6
// NewCmdAdd / NewCmdCreatePreset signatures). The pointer is read
// in RunE via derefString so changes to the parent's flag value
// between invocations are reflected on the next call.
//
// RunE populates opts from f.IOStreams, the lazy Library, c.Context(),
// and parsed flags, then dispatches to runF (test injection point)
// or runRefresh (production).
//
// Flags:
//
//	--dry-run   preview changes without modifying library.yaml
//	--force     skip resources with conflicts (name mismatch, malformed)
//	--output    json|table|plain (default: plain)
//
// The --library flag is registered on the parent `library` command as
// a PersistentFlag and is inherited transparently via the
// libraryPath *string parameter (mirrors the slice-6 NewCmdAdd
// pattern at cmd/library_add.go:136). It does NOT need to be
// re-registered here.
func NewCmdRefresh(f *cmdutil.Factory, libraryPath *string, runF func(*refreshOptions) error) *cobra.Command {
	var (
		dryRun bool
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh library metadata from resource files",
		Long: `Sync metadata from registered resource files into library.yaml.

This command updates descriptions when they differ from frontmatter,
updates paths when files are renamed (if frontmatter name matches),
and detects conflicts.

Examples:
  germinator library refresh
  germinator library refresh --dry-run
  germinator library refresh --force
  germinator library refresh --output json
  germinator library refresh --output table`,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &refreshOptions{
				IO:              f.IOStreams,
				Library:         refreshLibrary(f, derefString(libraryPath)),
				Ctx:             c.Context(),
				DryRun:          dryRun,
				Force:           force,
				Output:          outputFormatRefresh,
				CompletionCache: f.CompletionCache,
			}

			if runF != nil {
				return runF(opts)
			}
			return runRefresh(opts)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying library.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Skip resources with conflicts")
	output.AddOutputFlags(cmd, &outputFormatRefresh)

	return cmd
}

// outputFormatRefresh is the package-level string used by
// AddOutputFlags. Must be a package-level variable (not a stack-local)
// because Cobra binds the flag via StringVar into its backing
// storage; the runF closure captures &outputFormatRefresh when
// opts.Output is set in RunE.
var outputFormatRefresh string

// refreshLibrary wraps path resolution + load into a single lazy
// closure that callers populate into opts.Library. Mirrors
// cmd.addLibrary (slice 6) and cmd.createPresetLibrary (slice 6).
//
//   - nil factory => nil loader (tests bypass this layer by passing
//     their own Library closure).
//   - explicitPath == "" + env unset => FindLibrary falls through to
//     the XDG default path.
//
// The Library field in refreshOptions is typed as the canonical
// `func() (*library.Library, error)` per the task spec; the resolved
// path is captured in the closure per call. f.RootContext is the
// signal-aware context owned by the Factory.
func refreshLibrary(f *cmdutil.Factory, explicitPath string) func() (*library.Library, error) {
	if f == nil {
		return nil
	}
	return func() (*library.Library, error) {
		envPath := os.Getenv("GERMINATOR_LIBRARY")
		resolved := library.FindLibrary(explicitPath, envPath, "")
		return library.LoadLibrary(f.RootContext, resolved)
	}
}

// refreshChangedRow is the per-resource shape for refreshed entries.
// The dual json/tab tags mirror the slice-6 discoverRow pattern so the
// same struct drives both --output json and --output table.
type refreshChangedRow struct {
	Ref   string `json:"ref" tab:"REF"`
	Field string `json:"field" tab:"FIELD"`
	Old   string `json:"old" tab:"OLD"`
	New   string `json:"new" tab:"NEW"`
}

// refreshUnchangedRow carries the per-resource shape for the
// Unchanged section (Decision 7): a ref + the file's RFC3339 mtime
// (empty when not determinable).
type refreshUnchangedRow struct {
	Ref        string `json:"ref" tab:"REF"`
	LastSynced string `json:"lastSynced,omitempty" tab:"LAST_SYNCED"`
}

// refreshSkippedRow carries the per-resource shape for the Skipped
// section (a ref + reason like "missing_file" or "name_mismatch").
type refreshSkippedRow struct {
	Ref    string `json:"ref" tab:"REF"`
	Reason string `json:"reason" tab:"REASON"`
}

// refreshErrorRow carries the per-resource shape for the Errors
// section (a ref + the error type + the offending field).
type refreshErrorRow struct {
	Ref   string `json:"ref" tab:"REF"`
	Type  string `json:"type" tab:"TYPE"`
	Field string `json:"field" tab:"FIELD"`
}

// refreshJSONPayload is the net-new shape for --output json per
// design Decision 7 (slice 7). All four sections are always present
// in the payload (even when empty) so consumers can rely on a
// stable shape regardless of which scan outcomes occurred.
type refreshJSONPayload struct {
	Refreshed []refreshChangedRow   `json:"refreshed"`
	Unchanged []refreshUnchangedRow `json:"unchanged"`
	Skipped   []refreshSkippedRow   `json:"skipped"`
	Errors    []refreshErrorRow     `json:"errors"`
}

// buildRefreshJSONPayload materializes the JSON payload from the
// library's *RefreshResult, reshaping each section into its
// dedicated row type so the JSON field names are predictable
// (lowercase, per the slice-6 discoverRow pattern).
func buildRefreshJSONPayload(result *library.RefreshResult) refreshJSONPayload {
	refreshed := make([]refreshChangedRow, 0, len(result.Refreshed))
	for _, r := range result.Refreshed {
		refreshed = append(refreshed, refreshChangedRow{
			Ref:   r.Ref,
			Field: r.Field,
			Old:   r.Old,
			New:   r.New,
		})
	}
	unchanged := make([]refreshUnchangedRow, 0, len(result.Unchanged))
	for _, u := range result.Unchanged {
		unchanged = append(unchanged, refreshUnchangedRow{
			Ref:        u.Ref,
			LastSynced: u.LastSynced,
		})
	}
	skipped := make([]refreshSkippedRow, 0, len(result.Skipped))
	for _, s := range result.Skipped {
		skipped = append(skipped, refreshSkippedRow{
			Ref:    s.Ref,
			Reason: s.Reason,
		})
	}
	errors := make([]refreshErrorRow, 0, len(result.Errors))
	for _, e := range result.Errors {
		errors = append(errors, refreshErrorRow{
			Ref:   e.Ref,
			Type:  e.Type,
			Field: e.Field,
		})
	}
	return refreshJSONPayload{
		Refreshed: refreshed,
		Unchanged: unchanged,
		Skipped:   skipped,
		Errors:    errors,
	}
}

// buildRefreshTableRows flattens Refreshed entries into table rows.
// Only the Refreshed entries fit the (REF, FIELD, OLD, NEW) column
// shape; Skipped / Unchanged / Errors are rendered in plain output
// only (the spec scenario "Table output" accepts this omission: the
// table is a flat per-change view, not an aggregate report). The
// refreshChangedRow struct drives both JSON and table via its dual
// json/tab tags.
func buildRefreshTableRows(result *library.RefreshResult) []refreshChangedRow {
	rows := make([]refreshChangedRow, 0, len(result.Refreshed))
	for _, r := range result.Refreshed {
		rows = append(rows, refreshChangedRow{
			Ref:   r.Ref,
			Field: r.Field,
			Old:   r.Old,
			New:   r.New,
		})
	}
	return rows
}

// runRefresh executes the refresh logic. It is the production wiring
// for NewCmdRefresh's runF parameter.
//
// Flow:
//  1. Resolve the lazy library once via opts.Library().
//  2. Call lib.Refresh(opts.Ctx, &RefreshRequest{DryRun, Force}).
//  3. Dispatch on opts.Output for the result rendering.
//
// Errors from lib.Refresh are wrapped with %w and returned; main.go's
// centralized error handler renders them via output.FormatError.
// Per-resource errors from the scan are surfaced inline in the plain
// "Errors:" section (NOT returned as a typed error) so the human-
// readable report stays a single artifact on stdout.
func runRefresh(opts *refreshOptions) error {
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	opts.IO.Verbosef("refreshing library at %s", lib.RootPath)

	result, err := lib.Refresh(opts.Ctx, &library.RefreshRequest{
		DryRun: opts.DryRun,
		Force:  opts.Force,
	})
	if err != nil {
		return fmt.Errorf("refreshing library: %w", err)
	}

	if !opts.DryRun && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	switch opts.Output {
	case "json":
		if werr := output.NewJSONExporter().Write(opts.IO, buildRefreshJSONPayload(result)); werr != nil {
			return fmt.Errorf("json output: %w", werr)
		}
		return nil
	case "table":
		if werr := output.NewTableExporter().Write(opts.IO, buildRefreshTableRows(result)); werr != nil {
			return fmt.Errorf("table output: %w", werr)
		}
		return nil
	default:
		return renderRefreshPlain(opts, result)
	}
}

// renderRefreshPlain emits the per-section plain output. The section
// order matches design Decision 7 (slice 7): Refreshed, Unchanged
// (NEW), Skipped, Errors. Each section is suppressed when its
// underlying slice is empty so an all-clean refresh emits a minimal
// "Dry-run: ..." line on the dry-run path or nothing otherwise.
//
// If opts.DryRun is true, a "Dry-run: no changes made" line is
// prepended so users immediately see that no mutation occurred. The
// line is emitted even when there are Refreshed entries — those
// represent changes that WOULD have been made in non-dry-run mode.
//
// Write errors are intentionally discarded (the _, _ = idiom from
// the slice-6 plain-output helpers in cmd/library_add.go); the
// underlying io.Writer is opts.IO.Out which the shell layer owns,
// so a transient write failure on stdout cannot be retried here.
// Breaking renderRefreshPlain into a small dispatcher + four
// per-section renderers keeps gocognit under the 30 threshold.
func renderRefreshPlain(opts *refreshOptions, result *library.RefreshResult) error {
	if opts.DryRun {
		_, _ = fmt.Fprintln(opts.IO.Out, "Dry-run: no changes made")
	}
	renderRefreshedSection(opts.IO.Out, result.Refreshed)
	renderUnchangedSection(opts.IO.Out, result.Unchanged)
	renderSkippedSection(opts.IO.Out, result.Skipped)
	renderErrorsSection(opts.IO.Out, result.Errors)
	return nil
}

// renderRefreshedSection emits the "Refreshed:" section header and
// one "[refreshed] ref: field (old → new)" line per entry. Suppressed
// when no entries exist.
func renderRefreshedSection(w io.Writer, items []library.RefreshChange) {
	if len(items) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w, "\nRefreshed:")
	for _, r := range items {
		_, _ = fmt.Fprintf(w, "  [refreshed] %s: %s (%s → %s)\n",
			r.Ref, r.Field, r.Old, r.New)
	}
}

// renderUnchangedSection emits the "Unchanged:" section header and
// one "[unchanged] ref (lastSynced)" line per entry. LastSynced
// carries the file's mtime as RFC3339 when available; the empty
// string is rendered as "(no mtime)" so the trailing slot is always
// present and predictable for downstream tooling.
func renderUnchangedSection(w io.Writer, items []library.RefreshUnchanged) {
	if len(items) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w, "\nUnchanged:")
	for _, u := range items {
		last := u.LastSynced
		if last == "" {
			last = "no mtime"
		}
		_, _ = fmt.Fprintf(w, "  [unchanged] %s (%s)\n", u.Ref, last)
	}
}

// renderSkippedSection emits the "Skipped:" section header and one
// "[skipped] ref: reason" line per entry. Suppressed when no entries
// exist; the slice is populated by refresher.go's recordNameMismatch
// helper when a name mismatch is detected.
func renderSkippedSection(w io.Writer, items []library.SkipInfo) {
	if len(items) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w, "\nSkipped:")
	for _, s := range items {
		_, _ = fmt.Fprintf(w, "  [skipped] %s: %s\n", s.Ref, s.Reason)
	}
}

// renderErrorsSection emits the "Errors:" section header and one
// "[error] ref: type (field)" line per entry. Suppressed when no
// entries exist; the slice is populated by refresher.go's
// recordNameMismatch and isMalformedFrontmatter paths.
func renderErrorsSection(w io.Writer, items []library.RefreshError) {
	if len(items) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w, "\nErrors:")
	for _, e := range items {
		_, _ = fmt.Fprintf(w, "  [error] %s: %s (%s)\n", e.Ref, e.Type, e.Field)
	}
}

// NewLibraryRefreshCommand was the legacy bridge shim used by
// cmd/library.go while slice 7 landed. It was deleted in task 7.5.6
// once cmd/library.go was rewired to call NewCmdRefresh directly with
// the parent Factory.
