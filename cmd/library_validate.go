package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// libraryValidateOptions holds the runtime state for a
// `library validate` invocation. IO, Library (lazy: loaded via
// validateLibrary), and Ctx come from the Factory; the rest come from
// parsed flags. The Library lazy field is func() so the Factory can
// cache the heavy work (LoadLibrary) per call, matching the slice-6
// libraryPath pattern at cmd/library_create.go:149-159.
type libraryValidateOptions struct {
	IO              *iostreams.IOStreams
	Library         func() (*library.Library, error)
	Ctx             context.Context
	Fix             bool
	Output          string
	CompletionCache *cmdutil.CompletionCache
}

// validatorLibrary is the cmd-side contract for library validation.
// It is intentionally distinct from `Library` (which would shadow
// the library.Library struct) and from the slice-6 presetWriter
// interface. The method signature matches the
// (*library.Library).Validate method introduced in slice-7 7.0.2
// which internally invokes (*library.Library).Fix when req.Fix is
// true so the cmd-side code only needs to call Validate.
type validatorLibrary interface {
	Validate(ctx context.Context, req *library.ValidateRequest) (*library.ValidationResult, error)
}

// Compile-time confirmation that *library.Library satisfies the
// validatorLibrary contract. If either side changes (interface or
// (*Library).Validate method), the build fails immediately.
// *library.Library is the live receiver used by runLibraryValidate,
// so no suppression directive is required.
var _ validatorLibrary = (*library.Library)(nil)

// NewCmdLibraryValidate creates the `library validate` subcommand
// via the canonical NewCmdXxx(f, runF) pattern. Migrated in slice 7.
//
// runF is the test-injection seam; production wires it to
// runLibraryValidate, tests substitute a stub that captures the
// fully-populated *libraryValidateOptions. The constructor receives
// only the Factory — there is no `libraryPath *string` parameter
// because the parent's shared `--library` persistent flag is read
// directly from the cobra tree inside RunE (PersistentFlags are
// merged into Flags() at parse time, so children transparently
// inherit the parent's value via c.Flags().Lookup("library")).
//
// Flags:
//
//	--fix     (optional) auto-clean library.yaml after validation
//	--output  (optional) plain (default), json, or table
//
// Examples:
//
//	germinator library validate
//	germinator library validate --fix
//	germinator library validate --output json
//	germinator library validate --fix --output json
//	germinator library validate --fix --output table
func NewCmdLibraryValidate(f *cmdutil.Factory, runF func(*libraryValidateOptions) error) *cobra.Command {
	var (
		fix          bool
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate library integrity",
		Long: `Validate library.yaml metadata against the filesystem.

Checks for four issue types:
  - missing-file: entry in library.yaml but file doesn't exist
  - ghost-resource: preset references non-existent resource
  - orphan: file exists but isn't registered in library.yaml
  - malformed-frontmatter: resource file has invalid YAML frontmatter

Use --fix to auto-clean library.yaml (removes missing entries, strips ghost refs).
Only modifies library.yaml - never deletes actual files.

Output formats (--output):
  plain  default; human-readable summary grouped by severity
  json   machine-readable report with optional fix section
  table  tab-aligned rows (severity, type, ref, message)`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &libraryValidateOptions{
				IO:              f.IOStreams,
				Library:         validateLibrary(f, resolveLibraryFlag(c)),
				Ctx:             c.Context(),
				Fix:             fix,
				Output:          outputFormat,
				CompletionCache: f.CompletionCache,
			}
			if runF != nil {
				return runF(opts)
			}
			return runLibraryValidate(opts)
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Auto-clean library.yaml (removes missing entries and ghost preset refs)")
	cmdutil.AddOutputFlags(cmd, &outputFormat)

	return cmd
}

// resolveLibraryFlag reads the parent's --library persistent flag
// value from the cobra tree. PersistentFlags are merged into Flags()
// at parse time, so children transparently inherit the parent's
// value via c.Flags().Lookup("library"). Returns "" when the flag
// is not registered (e.g., when NewCmdLibraryValidate is invoked
// directly in tests without the parent cmd registered).
func resolveLibraryFlag(c *cobra.Command) string {
	if c == nil {
		return ""
	}
	pf := c.Flags().Lookup("library")
	if pf == nil {
		return ""
	}
	return pf.Value.String()
}

// validateLibrary wraps path resolution + load into a single lazy
// closure that callers populate into opts.Library. Mirrors
// cmd.createPresetLibrary (slice-6) and cmd.addLibrary (slice-6) so
// the Factory's per-call path resolution pattern is honored.
//
//   - nil factory => nil loader (tests bypass this layer by passing
//     their own Library closure).
//   - explicitPath == "" + env unset => FindLibrary falls through to
//     the XDG default path.
//
// The Library field in libraryValidateOptions is typed as the
// canonical `func() (*library.Library, error)` per the task spec;
// the resolved path is captured in the closure.
func validateLibrary(f *cmdutil.Factory, explicitPath string) func() (*library.Library, error) {
	if f == nil {
		return nil
	}
	resolved := library.FindLibrary(explicitPath, os.Getenv("GERMINATOR_LIBRARY"))
	return func() (*library.Library, error) {
		return library.LoadLibrary(f.RootContext, resolved)
	}
}

// runLibraryValidate executes the validation logic. Dispatches on
// opts.Output to plain / JSON / table. Errors from the
// validatorLibrary call are surfaced via output.FormatError (to
// stderr) and wrapped for exit-code mapping; per-issue findings
// stay data and flow through the chosen output format.
func runLibraryValidate(opts *libraryValidateOptions) error {
	lib, err := opts.Library()
	if err != nil {
		output.FormatError(opts.IO, err)
		return fmt.Errorf("loading library: %w", err)
	}

	opts.IO.Verbosef("validating library at %s (fix=%t)", lib.RootPath, opts.Fix)

	result, err := lib.Validate(opts.Ctx, &library.ValidateRequest{Fix: opts.Fix})
	if err != nil {
		output.FormatError(opts.IO, err)
		return fmt.Errorf("validating library: %w", err)
	}

	// Invalidate the completion cache only when --fix was used and the
	// validator reports a non-nil FixResult (meaning the fix path ran
	// and may have rewritten library.yaml). Plain validation is
	// read-only and leaves the cache untouched.
	if opts.Fix && result.FixResult != nil && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	switch opts.Output {
	case outputJSON:
		return writeValidateJSON(opts.IO, result)
	case outputTable:
		return writeValidateTable(opts.IO, result)
	default:
		return writeValidatePlain(opts.IO, result, opts.Fix)
	}
}

// outputFormat sentinel constants — avoid magic-string drift with
// output.AddOutputFlags / output.DefaultOutputFormat. Mirrors the
// pattern at cmd/library_add.go:789-793 (isPlainOutput).
const (
	outputJSON  = "json"
	outputTable = "table"
)

// validateIssueJSON is the JSON projection of a single library.Issue.
// Mirrors the legacy ValidationOutput shape (cmd pre-rewrite) for
// backwards compatibility with downstream tooling that consumes
// `germinator library validate --json` output.
type validateIssueJSON struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Ref      string `json:"ref,omitempty"`
	Path     string `json:"path,omitempty"`
	InPreset string `json:"inPreset,omitempty"`
	Message  string `json:"message,omitempty"`
}

// fixJSONSection is the JSON projection of *library.FixResult.
// Returned under the top-level "fix" field when opts.Fix is true;
// omitted (json:",omitempty") otherwise so default validation runs
// stay identical to the pre-fix shape.
type fixJSONSection struct {
	RemovedEntries []string `json:"removedEntries"`
	StrippedRefs   []string `json:"strippedRefs"`
}

// validateJSONPayload is the net-new JSON shape for
// `library validate --output json`. Diverges from the legacy
// ValidationOutput only by the optional Fix field (omitempty) so
// pre-fix tooling keeps working unchanged.
type validateJSONPayload struct {
	Valid        bool                `json:"valid"`
	ErrorCount   int                 `json:"errorCount"`
	WarningCount int                 `json:"warningCount"`
	Issues       []validateIssueJSON `json:"issues"`
	Fix          *fixJSONSection     `json:"fix,omitempty"`
}

// writeValidateJSON materializes the JSON payload and writes it via
// the JSONExporter. The Issues slice is sorted (errors before
// warnings, then by type) to keep the byte output deterministic so
// golden tests and downstream parsers stay stable.
func writeValidateJSON(io *iostreams.IOStreams, result *library.ValidationResult) error {
	payload := validateJSONPayload{
		Valid:        result.Valid,
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		Issues:       make([]validateIssueJSON, 0, len(result.Issues)),
	}

	for _, issue := range result.Issues {
		payload.Issues = append(payload.Issues, validateIssueJSON{
			Type:     string(issue.Type),
			Severity: string(issue.Severity),
			Ref:      issue.Ref,
			Path:     issue.Path,
			InPreset: issue.InPreset,
			Message:  issue.Message,
		})
	}

	sort.Slice(payload.Issues, func(i, j int) bool {
		if payload.Issues[i].Severity != payload.Issues[j].Severity {
			return payload.Issues[i].Severity == string(library.SeverityError)
		}
		return payload.Issues[i].Type < payload.Issues[j].Type
	})

	// Populate the fix section only when --fix was actually used
	// (lib.Validate internally calls lib.Fix and merges the result
	// into result.FixResult; for non-fix runs FixResult stays nil).
	if result.FixResult != nil {
		payload.Fix = &fixJSONSection{
			RemovedEntries: result.FixResult.RemovedEntries,
			StrippedRefs:   result.FixResult.StrippedRefs,
		}
	}

	if err := output.NewJSONExporter().Write(io, payload); err != nil {
		return fmt.Errorf("writing JSON output: %w", err)
	}
	return nil
}

// validateRow is the table-exporter projection of a single
// library.Issue. Columns: severity, type, ref, message. The Ref
// field falls back to Path when Ref is empty so the ref column is
// always populated.
type validateRow struct {
	Severity string `tab:"SEVERITY"`
	Type     string `tab:"TYPE"`
	Ref      string `tab:"REF"`
	Message  string `tab:"MESSAGE"`
}

// fixActionRow is the table-exporter projection of a single fix
// action (removed entry or stripped ref). Columns: action, ref. The
// action column names the category ("removed" vs "stripped") so
// the table reads as an action log.
type fixActionRow struct {
	Action string `tab:"ACTION"`
	Ref    string `tab:"REF"`
}

// writeValidateTable dispatches between the two table shapes per
// whether --fix was used:
//
//   - issues-only: severity/type/ref/message (always available)
//   - --fix:       action/ref (renders only the fix actions; the
//     plain output above already covers the issues summary)
//
// The TableExporter requires a homogeneous slice, so the two shapes
// are rendered with separate Write calls — never a single mixed
// slice (the reflect-driven exporter cannot reconcile the two
// column definitions inside a single call).
func writeValidateTable(io *iostreams.IOStreams, result *library.ValidationResult) error {
	// Fix-action table takes priority when --fix was actually used
	// AND removed entries or stripped refs are present; the issues
	// table supplies the always-available detail rows.
	fixRows := buildFixActionRows(result.FixResult)
	if len(fixRows) > 0 {
		if err := output.NewTableExporter().Write(io, fixRows); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	}

	issueRows := buildIssueRows(result)
	if len(issueRows) == 0 {
		return nil
	}
	if err := output.NewTableExporter().Write(io, issueRows); err != nil {
		return fmt.Errorf("writing table output: %w", err)
	}
	return nil
}

// buildIssueRows projects result.Issues into validateRow slice
// suitable for the TableExporter. The Ref field falls back to Path
// when Ref is empty so the column always has something to display.
func buildIssueRows(result *library.ValidationResult) []validateRow {
	rows := make([]validateRow, 0, len(result.Issues))
	for _, issue := range result.Issues {
		ref := issue.Ref
		if ref == "" {
			ref = issue.Path
		}
		rows = append(rows, validateRow{
			Severity: string(issue.Severity),
			Type:     string(issue.Type),
			Ref:      ref,
			Message:  issue.Message,
		})
	}
	return rows
}

// buildFixActionRows projects a *library.FixResult into the
// (action, ref) table rows, ordered: removed-entries first, then
// stripped-refs. Each row gets the appropriate Action label.
func buildFixActionRows(fix *library.FixResult) []fixActionRow {
	if fix == nil {
		return nil
	}
	rows := make([]fixActionRow, 0, len(fix.RemovedEntries)+len(fix.StrippedRefs))
	for _, ref := range fix.RemovedEntries {
		rows = append(rows, fixActionRow{Action: "removed", Ref: ref})
	}
	for _, ref := range fix.StrippedRefs {
		rows = append(rows, fixActionRow{Action: "stripped", Ref: ref})
	}
	return rows
}

// writeValidatePlain renders the legacy human-readable output:
// header (valid/invalid + counts) + issues (grouped by type /
// severity) + footer (fix hint). Writes entirely to opts.IO.Out so
// plain output stays scriptable via "germinator library validate |
// grep".
func writeValidatePlain(io *iostreams.IOStreams, result *library.ValidationResult, fix bool) error {
	var sb strings.Builder

	validatePlainHeader(&sb, result)
	validatePlainIssues(&sb, result)
	validatePlainFooter(&sb, result, fix)

	if _, err := io.Out.Write([]byte(sb.String())); err != nil {
		return fmt.Errorf("writing plain output: %w", err)
	}
	return nil
}

func validatePlainHeader(sb *strings.Builder, result *library.ValidationResult) {
	if result.Valid && len(result.Issues) == 0 {
		fmt.Fprintln(sb, "\u2713 Library is valid")
		fmt.Fprintf(sb, "  errors: 0, warnings: 0\n")
		return
	}
	if result.Valid {
		fmt.Fprintln(sb, "\u2713 Library is valid (warnings only)")
		fmt.Fprintf(sb, "  errors: 0, warnings: %d\n", result.WarningCount)
		fmt.Fprintln(sb)
		return
	}
	fmt.Fprintln(sb, "\u2717 Library has issues")
	fmt.Fprintf(sb, "  errors: %d, warnings: %d\n", result.ErrorCount, result.WarningCount)
	fmt.Fprintln(sb)
}

func validatePlainIssues(sb *strings.Builder, result *library.ValidationResult) {
	errorsByType := make(map[library.IssueType][]library.Issue)
	warnings := make([]library.Issue, 0)

	for _, issue := range result.Issues {
		if issue.Severity == library.SeverityError {
			errorsByType[issue.Type] = append(errorsByType[issue.Type], issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	if len(errorsByType) > 0 {
		fmt.Fprintln(sb, "Errors:")
		for _, issueType := range []library.IssueType{
			library.IssueTypeMissingFile,
			library.IssueTypeGhostResource,
			library.IssueTypeMalformedFrontmatter,
		} {
			if issues, ok := errorsByType[issueType]; ok && len(issues) > 0 {
				for _, issue := range issues {
					validatePlainFormatIssue(sb, issue)
				}
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Fprintln(sb, "Warnings:")
		for _, issue := range warnings {
			validatePlainFormatIssue(sb, issue)
		}
	}
}

func validatePlainFormatIssue(sb *strings.Builder, issue library.Issue) {
	fmt.Fprintf(sb, "  [%s] %s", validatePlainIssueType(issue.Type), validatePlainIssueRefOrPath(issue))
	if issue.InPreset != "" {
		fmt.Fprintf(sb, " (in preset %q)", issue.InPreset)
	}
	fmt.Fprintln(sb)
	if issue.Message != "" {
		fmt.Fprintf(sb, "    %s\n", issue.Message)
	}
}

func validatePlainFooter(sb *strings.Builder, result *library.ValidationResult, fix bool) {
	fmt.Fprintln(sb)
	if !result.Valid {
		if fix {
			fmt.Fprintln(sb, "Fix applied to library.yaml")
		} else {
			fmt.Fprintln(sb, "Hint: Run with --fix to auto-clean library.yaml")
		}
	} else if fix {
		fmt.Fprintln(sb, "no fixes needed")
	}
	fmt.Fprintln(sb, "Hint: Run with --json for machine-readable output")
}

func validatePlainIssueType(t library.IssueType) string {
	switch t {
	case library.IssueTypeMissingFile:
		return "missing"
	case library.IssueTypeGhostResource:
		return "ghost"
	case library.IssueTypeOrphan:
		return "orphan"
	case library.IssueTypeMalformedFrontmatter:
		return "malformed"
	default:
		return string(t)
	}
}

func validatePlainIssueRefOrPath(issue library.Issue) string {
	if issue.Ref != "" {
		return issue.Ref
	}
	return issue.Path
}
