package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// libraryInitOptions holds the runtime state for a `library init`
// invocation. IO and Ctx come from the Factory; the rest come from
// parsed flags.
//
// No Library lazy loader field per design Decision 6: `init` creates
// a fresh library, so there is no pre-existing *library.Library to
// receive a method call. runLibraryInit invokes library.Init
// (the package-level function in internal/library/creator.go:93)
// directly without an interface or adapter shim, matching the
// dual-form pattern of CreatePreset / (*Library).CreatePreset at
// internal/library/creator.go:127.
type libraryInitOptions struct {
	IO              *iostreams.IOStreams
	Ctx             context.Context
	Path            string
	Force           bool
	DryRun          bool
	Output          string
	CompletionCache *cmdutil.CompletionCache
}

// initRow is the single-row table/JSON payload for `library init`
// results. The `tab:"..."` tags drive the TableExporter column
// headers; the `json:"..."` tags drive JSON output. Only the three
// fields named in the spec (path, dryRun, created) are exposed; no
// other metadata is included.
type initRow struct {
	Path    string `json:"path"    tab:"PATH"`
	DryRun  bool   `json:"dryRun"  tab:"DRYRUN"`
	Created bool   `json:"created" tab:"CREATED"`
}

// initOutputFormat is the package-level string used by AddOutputFlags.
// It must be a package-level variable (not a stack-local) because
// Cobra binds the flag via StringVar into its backing storage; the
// runF closure captures &initOutputFormat when opts.Output is set
// in RunE (same pattern as the slice-6 library_add.go outputFormat).
var initOutputFormat string

// NewCmdLibraryInit creates the `library init` command via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 7.
//
// No `libraryPath` parameter per design Decision 6: `init` does not
// consult the parent's --library flag (the operation creates a fresh
// library, not a mutation against an existing one). The --path
// flag's default is the XDG-resolved library path computed at
// run-time via library.DefaultLibraryPath().
//
// RunE populates opts from f.IOStreams, c.Context(), and parsed
// flags, then dispatches to runF (test injection point) or
// runLibraryInit (production).
//
// Flags:
//
//	--path     (optional) target path; defaults to library.DefaultLibraryPath()
//	--force    (optional) overwrite existing library
//	--dry-run  (optional) preview without creating files
//	--output   (optional) json|table|plain (default: plain)
//
// The legacy --json flag is replaced by --output json; invoking
// --json now triggers a Cobra usage error mapped to ExitCodeUsage.
func NewCmdLibraryInit(f *cmdutil.Factory, runF func(*libraryInitOptions) error) *cobra.Command {
	var (
		path   string
		force  bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new library directory structure",
		Long: `Create a new library directory structure at the specified path.

Creates a library.yaml file and empty resource directories (skills, agents,
commands, memory). The created library is validated by loading it to ensure
structural correctness.

By default, creates at ` + library.DefaultLibraryPath() + ` unless --path is
specified. Returns an error if a library already exists at the target path
unless --force is used.

Examples:
  germinator library init
  germinator library init --path /tmp/my-library
  germinator library init --dry-run
  germinator library init --force
  germinator library init --output json`,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &libraryInitOptions{
				IO:              f.IOStreams,
				Ctx:             c.Context(),
				Path:            path,
				Force:           force,
				DryRun:          dryRun,
				Output:          initOutputFormat,
				CompletionCache: f.CompletionCache,
			}
			if runF != nil {
				return runF(opts)
			}
			return runLibraryInit(opts)
		},
	}

	cmd.Flags().StringVar(&path, "path", "",
		"Path to create library (default: "+library.DefaultLibraryPath()+")")
	cmd.Flags().BoolVar(&force, "force", false,
		"Overwrite existing library at target path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Preview changes without creating files")
	output.AddOutputFlags(cmd, &initOutputFormat)

	return cmd
}

// runLibraryInit executes the library creation logic. It is the
// production wiring for NewCmdLibraryInit's runF parameter.
//
// Path resolution: opts.Path is honored when set; otherwise the
// XDG-resolved default path (library.DefaultLibraryPath) is used.
// The resolved path is what library.Init (and therefore
// CreateLibrary) will create on success.
//
// Errors from library.Init are wrapped with %w so cmdutil.ExitCodeFor
// can map typed core errors to exit codes (FileError → 1 via the
// default-error branch).
//
// Output dispatch:
//
//	"json":  single-row initRow via output.NewJSONExporter
//	"table": single-row initRow via output.NewTableExporter
//	default: confirmation message on opts.IO.Out (stdout)
//
// For dry-run, the underlying library.CreateLibrary has already
// printed the "Would create library at..." block via the writer
// passed in InitRequest.Stdout (cmd layer sets it to opts.IO.Out,
// so the preview lands on the same writer as the confirmation line
// below). runLibraryInit does not double-print; the cmd-layer
// "Dry run complete" confirmation is the only line added in
// dry-run plain mode.
func runLibraryInit(opts *libraryInitOptions) error {
	path := opts.Path
	if path == "" {
		path = library.DefaultLibraryPath()
	}

	opts.IO.Verbosef("Creating library at: %s", path)

	if err := library.Init(opts.Ctx, &library.InitRequest{
		Path:   path,
		Force:  opts.Force,
		DryRun: opts.DryRun,
		Stdout: opts.IO.Out,
	}); err != nil {
		return fmt.Errorf("creating library: %w", err)
	}

	if !opts.DryRun && opts.CompletionCache != nil {
		opts.CompletionCache.Invalidate()
	}

	row := initRow{Path: path, DryRun: opts.DryRun, Created: !opts.DryRun}

	switch opts.Output {
	case "json":
		if werr := output.NewJSONExporter().Write(opts.IO, row); werr != nil {
			return fmt.Errorf("json output: %w", werr)
		}
	case "table":
		if werr := output.NewTableExporter().Write(opts.IO, []initRow{row}); werr != nil {
			return fmt.Errorf("table output: %w", werr)
		}
	default:
		if opts.DryRun {
			_, _ = fmt.Fprintln(opts.IO.Out, "Dry run complete - no changes made")
		} else {
			_, _ = fmt.Fprintf(opts.IO.Out, "Library created successfully at: %s\n", path)
		}
	}
	return nil
}
