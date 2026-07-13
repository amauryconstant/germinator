package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// removeOptions holds the runtime state for a `library remove` invocation.
// IO, Library (lazy: loaded via removeLibrary), and Ctx come from the
// Factory; the rest come from parsed flags + positional args.
//
// `Ref` and `PresetName` are mutually exclusive: the resource sub-command
// populates `Ref` (e.g., "skill/commit"), the preset sub-command
// populates `PresetName` (e.g., "git-workflow"). `runRemove` dispatches on
// `PresetName != ""` to choose the remove path. The legacy CLI surface
// `germinator library remove resource <ref>` and
// `germinator library remove preset <name>` is preserved (no
// `--type` / `--name` flag substitution).
//
// `Library` is a lazy function field so the Factory can cache the
// heavy work (LoadLibrary) per the slice-5/6/7 pattern.
type removeOptions struct {
	IO              *iostreams.IOStreams
	Library         func() (*library.Library, error)
	Ctx             context.Context
	Ref             string
	PresetName      string
	Force           bool
	Output          string
	CompletionCache *cmdutil.CompletionCache
}

// removerLibrary is the cmd-side contract for removal operations. It
// is intentionally distinct from `Library` (which would shadow the
// library.Library struct). The method signatures match the
// (*library.Library) methods introduced in slice 7.0 — both return
// only `error` and rely on the cmd layer to capture the output data
// (file path / resources list) from the loaded library before the
// mutation so the JSON / table payloads can be rendered after the
// call returns.
//
// Unlike the slice-6 resourceAdder / libraryAdapter pattern, no
// stateless adapter shim is needed: RemoveResource and RemovePreset
// are real methods on *Library (introduced in 7.0), so
// *library.Library satisfies removerLibrary directly.
type removerLibrary interface {
	RemoveResource(ctx context.Context, req *library.RemoveResourceRequest) error
	RemovePreset(ctx context.Context, req *library.RemovePresetRequest) error
}

// Compile-time confirmation that *library.Library satisfies the
// removerLibrary contract. If either side changes (interface or
// (*Library) method), the build fails immediately. *library.Library
// is the live receiver used by runRemove, so no suppression directive
// is required.
var _ removerLibrary = (*library.Library)(nil)

// removeResourceRow is the table-exporter representation of a single
// resource removal. The tab struct tags drive the TableExporter
// column header order; the JSONExporter uses the `json` tags for
// marshaling. The spec scenario "Table output" requires columns
// "ref, action" with the removed resource appearing as a single row.
type removeResourceRow struct {
	Ref    string `tab:"REF"    json:"ref"`
	Action string `tab:"ACTION" json:"action"`
}

// removePresetRow is the table-exporter representation of a single
// preset removal. Columns are "name, action" per the spec scenario
// "Table output" for presets.
type removePresetRow struct {
	Name   string `tab:"NAME"   json:"name"`
	Action string `tab:"ACTION" json:"action"`
}

// NewCmdRemove creates the `library remove` command tree (parent +
// two sub-commands: `resource` and `preset`) via the canonical
// NewCmdXxx(f, runF) pattern. Migrated in slice 7.3.
//
// The parent itself has NO RunE — invoking `library remove` with no
// sub-command prints the help text (Cobra's default for a command
// with sub-commands and no Run). The two sub-commands each have
// their own RunE that populates `opts` and dispatches to runF (test
// injection point) or runRemove (production).
//
// `--force` and `--output` are bound as PersistentFlags on the
// parent so both sub-commands inherit them without re-declaration.
// The spec scenario "JSON output" requires
// `germinator library remove resource skill/commit --output json`
// to work, which is only possible with PersistentFlags (cobra
// rejects sub-command-local flags that aren't declared on the
// sub-command).
//
// libraryPath is the parent's shared `--library` pointer so the
// parent's persistent flag value is honored (same shape as slice-3
// resources / slice-4 presets / slice-6 add / slice-7 refresh). The
// pointer is read in RunE via derefString so changes between
// invocations are reflected on the next call.
//
// runF lets tests capture the fully-parsed *removeOptions without
// invoking runRemove. The production wiring in main.go passes nil
// so the production body runs.
func NewCmdRemove(f *cmdutil.Factory, libraryPath *string, runF func(*removeOptions) error) *cobra.Command {
	opts := &removeOptions{}

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a resource or preset from the library",
		Long: `Remove a resource or preset from the library.

Subcommands:
  resource  Remove a resource from the library (deletes file + YAML entry)
  preset    Remove a preset from the library (YAML entry only)`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	// Persistent flags inherited by both sub-commands. --force is a
	// no-op for RemovePreset (no physical file to force) but is
	// accepted on the parent so both sub-commands share the same
	// struct; see the constraint note in the task spec.
	cmd.PersistentFlags().BoolVar(&opts.Force, "force", false,
		"Skip confirmation prompts and remove unconditionally")

	// --output on the parent as a PersistentFlag so both
	// sub-commands see it. The spec's "JSON output" scenario
	// (`library remove resource skill/commit --output json`) requires
	// the flag to be visible after the sub-command token, which is
	// only possible with PersistentFlags. `output.AddOutputFlags`
	// binds to local `cmd.Flags()`, so we wire the flag inline
	// here to use `cmd.PersistentFlags()`.
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o",
		output.DefaultOutputFormat, "Output format: json, table, plain")
	_ = cmd.RegisterFlagCompletionFunc("output",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return output.ValidOutputFormats, cobra.ShellCompDirectiveNoFileComp
		})

	// Sub-command: resource. Args: cobra.ExactArgs(1) — the
	// positional <ref> is required; no flag substitution.
	resourceCmd := &cobra.Command{
		Use:   "resource <ref>",
		Short: "Remove a resource from the library",
		Long: `Remove a resource from the library.

Deletes both the physical file and YAML entry.
Errors if any preset references the resource.

Examples:
  germinator library remove resource skill/commit
  germinator library remove resource skill/commit --force
  germinator library remove resource skill/commit --output json
  germinator library remove resource skill/commit --output table`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = c.Context()
			opts.Ref = args[0]
			opts.Library = removeLibrary(f, derefString(libraryPath))
			opts.CompletionCache = f.CompletionCache
			if runF != nil {
				return runF(opts)
			}
			return runRemove(opts)
		},
	}
	cmd.AddCommand(resourceCmd)

	// Sub-command: preset. Args: cobra.ExactArgs(1) — the
	// positional <name> is required.
	presetCmd := &cobra.Command{
		Use:   "preset <name>",
		Short: "Remove a preset from the library",
		Long: `Remove a preset from the library.

Removes only the YAML entry (no physical file to delete).

Examples:
  germinator library remove preset git-workflow
  germinator library remove preset git-workflow --force
  germinator library remove preset git-workflow --output json
  germinator library remove preset git-workflow --output table`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = c.Context()
			opts.PresetName = args[0]
			opts.Library = removeLibrary(f, derefString(libraryPath))
			opts.CompletionCache = f.CompletionCache
			if runF != nil {
				return runF(opts)
			}
			return runRemove(opts)
		},
	}
	cmd.AddCommand(presetCmd)

	return cmd
}

// removeLibrary wraps path resolution + load into a single lazy
// closure that callers populate into opts.Library. Mirrors
// cmd.refreshLibrary (slice 7) so the Factory's per-call path
// resolution pattern is honored.
//
//   - nil factory => nil loader (tests bypass this layer by passing
//     their own Library closure).
//   - explicitPath == "" + env unset => FindLibrary falls through to
//     the XDG default path.
//
// The Library field in removeOptions is typed as the canonical
// `func() (*library.Library, error)` per the task spec; the resolved
// path is captured in the closure per call.
func removeLibrary(f *cmdutil.Factory, explicitPath string) func() (*library.Library, error) {
	if f == nil {
		return nil
	}
	return func() (*library.Library, error) {
		envPath := os.Getenv("GERMINATOR_LIBRARY")
		resolved := library.FindLibrary(explicitPath, envPath)
		return library.LoadLibrary(f.RootContext, resolved)
	}
}

// runRemove dispatches on `opts.PresetName` to choose the removal
// path. Preset name wins over ref when both are populated (defensive;
// the RunE closures only populate one per sub-command).
func runRemove(opts *removeOptions) error {
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	if opts.PresetName != "" {
		return runRemovePreset(opts, lib)
	}
	return runRemoveResource(opts, lib)
}

// runRemoveResource removes a single resource by ref. The
// pre-validation pass captures the file path (BEFORE the mutation
// deletes the resource) so the JSON / table output can report
// `fileDeleted` and the path-keyed column. After the method call,
// the loaded library's in-memory map is unchanged (the method saves
// to library.yaml but does not mutate the receiver), so the
// captured data is still valid for output rendering.
//
// Error mapping:
//   - `library.ParseRef` failure → *core.ConfigError (mapped to exit 1)
//   - invalid resource type → *core.NotFoundError (mapped to exit 2
//     via cmdutil.ExitCodeFor's NotFoundError branch)
//   - resource not registered → *core.NotFoundError
//   - preset references resource → wrapped *core.FileError
//   - method call failure → wrapped error
func runRemoveResource(opts *removeOptions, lib *library.Library) error {
	typ, name, err := library.ParseRef(opts.Ref)
	if err != nil {
		return fmt.Errorf("parsing reference: %w", err)
	}

	resourceType := library.ResourceType(typ)
	if !resourceType.IsValid() {
		return core.NewNotFoundError("library ref", opts.Ref)
	}

	typeMap, ok := lib.Resources[typ]
	if !ok {
		return core.NewNotFoundError("library ref", opts.Ref)
	}
	res, ok := typeMap[name]
	if !ok {
		return core.NewNotFoundError("library ref", opts.Ref)
	}
	fileDeleted := filepath.Join(lib.RootPath, res.Path)

	opts.IO.Verbosef("removing resource %s from %s", opts.Ref, lib.RootPath)

	if err := lib.RemoveResource(opts.Ctx, &library.RemoveResourceRequest{
		Ref:   opts.Ref,
		Force: opts.Force,
	}); err != nil {
		return fmt.Errorf("removing resource: %w", err)
	}

	if opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	payload := &library.RemoveResourceOutput{
		Type:         "resource",
		ResourceType: typ,
		Name:         name,
		FileDeleted:  fileDeleted,
		LibraryPath:  lib.RootPath,
	}

	return renderRemoveResource(opts, payload)
}

// renderRemoveResource dispatches the final output for resource
// removal. Plain output is "Removed resource: <ref>\n" — byte-
// identical to the legacy pre-change build. JSON and table use the
// net-new struct shapes defined in internal/library/remover.go and
// the local removeResourceRow type respectively.
func renderRemoveResource(opts *removeOptions, payload *library.RemoveResourceOutput) error {
	switch opts.Output {
	case "json":
		if err := output.NewJSONExporter().Write(opts.IO, payload); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		row := removeResourceRow{Ref: opts.Ref, Action: "removed"}
		if err := output.NewTableExporter().Write(opts.IO, []removeResourceRow{row}); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		if _, err := fmt.Fprintf(opts.IO.Out, "Removed resource: %s\n", opts.Ref); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// runRemovePreset removes a single preset by name. The pre-validation
// pass captures the resources list (BEFORE the mutation drops the
// preset) so the JSON output can report `resourcesRemoved`. The
// `--force` flag is accepted on the parent (shared struct) but is
// a no-op for preset removal (no physical file to force) — see the
// task spec's constraint note.
//
// Error mapping:
//   - preset not found → *core.NotFoundError (exit 2)
//   - method call failure → wrapped error
func runRemovePreset(opts *removeOptions, lib *library.Library) error {
	preset, ok := lib.Presets[opts.PresetName]
	if !ok {
		return core.NewNotFoundError("preset", opts.PresetName)
	}
	resourcesRemoved := append([]string{}, preset.Resources...)

	opts.IO.Verbosef("removing preset %s from %s", opts.PresetName, lib.RootPath)

	if err := lib.RemovePreset(opts.Ctx, &library.RemovePresetRequest{
		Name:  opts.PresetName,
		Force: opts.Force,
	}); err != nil {
		return fmt.Errorf("removing preset: %w", err)
	}

	if opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	payload := &library.RemovePresetOutput{
		Type:             "preset",
		Name:             opts.PresetName,
		ResourcesRemoved: resourcesRemoved,
	}

	return renderRemovePreset(opts, payload)
}

// renderRemovePreset dispatches the final output for preset removal.
// Plain output is "Removed preset: <name>\n" — byte-identical to
// the legacy pre-change build. JSON and table use the net-new
// struct shapes.
func renderRemovePreset(opts *removeOptions, payload *library.RemovePresetOutput) error {
	switch opts.Output {
	case "json":
		if err := output.NewJSONExporter().Write(opts.IO, payload); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		row := removePresetRow{Name: opts.PresetName, Action: "removed"}
		if err := output.NewTableExporter().Write(opts.IO, []removePresetRow{row}); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		if _, err := fmt.Fprintf(opts.IO.Out, "Removed preset: %s\n", opts.PresetName); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// NewLibraryRemoveCommand was the legacy bridge shim used by
// cmd/library.go while slice 7 landed. It was deleted in task 7.5.6
// once cmd/library.go was rewired to call NewCmdRemove directly with
// the parent Factory.
