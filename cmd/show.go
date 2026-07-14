package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// presetPrefix is the canonical prefix used to disambiguate preset
// refs from resource refs (which use the `type/name` shape).
const presetPrefix = "preset/"

// showOptions holds the runtime state for a `library show <ref>`
// invocation.
type showOptions struct {
	IO                *iostreams.IOStreams
	Library           func() (*library.Library, error)
	Ctx               context.Context
	ConfigLibraryPath string
	Ref               string
	Output            string
}

// showResourceRow is the table-exporter representation of a single
// resource as displayed by `library show`. The tab struct tags drive
// the TableExporter column header order; the JSONExporter uses the
// `json` tags for marshaling.
type showResourceRow struct {
	Ref         string `tab:"REF"         json:"ref"`
	Path        string `tab:"-"           json:"path"`
	Description string `tab:"DESCRIPTION" json:"description,omitempty"`
}

// showPresetRow is the table-exporter representation of a single
// preset as displayed by `library show preset/<name>`.
type showPresetRow struct {
	Name        string   `tab:"NAME"        json:"name"`
	Description string   `tab:"DESCRIPTION" json:"description,omitempty"`
	Resources   []string `tab:"RESOURCES"   json:"resources"`
}

// NewCmdShow creates the `library show <ref>` subcommand via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 4.
//
// libraryPath is a shared *string with the parent `library` command
// (set by `cmd.PersistentFlags().StringVar` in NewLibraryCommand).
// Passing the pointer keeps the parent's --library flag working.
func NewCmdShow(f *cmdutil.Factory, libraryPath *string, runF func(*showOptions) error) *cobra.Command {
	opts := &showOptions{}
	cmd := &cobra.Command{
		Use:   "show <ref>",
		Short: "Display details of a resource or preset",
		Long: `Display details of a resource or preset.

For resources, use the format: type/name (e.g., skill/commit)
For presets, use the preset/ prefix (e.g., preset/git-workflow)

Output formats (--output):
  plain  default; byte-identical to the pre-change plain output
  json   JSON document suitable for jq
  table  tab-aligned text table

Examples:
  germinator library show skill/commit
  germinator library show preset/git-workflow
  germinator library show agent/reviewer --library /path/to/library
  germinator library show skill/commit --output json
  germinator library show preset/git-workflow --output table`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = c.Context()
			opts.Ref = args[0]
			lp := ""
			if libraryPath != nil {
				lp = *libraryPath
			}
			if f.Config != nil {
				if cfg, cfgErr := f.Config(); cfgErr == nil && cfg != nil {
					opts.ConfigLibraryPath = cfg.Library
				}
			}
			resolved := library.FindLibrary(lp, os.Getenv("GERMINATOR_LIBRARY"), opts.ConfigLibraryPath)
			opts.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) {
				return library.LoadLibrary(opts.Ctx, resolved)
			})
			if runF != nil {
				return runF(opts)
			}
			return runShow(opts)
		},
	}

	output.AddOutputFlags(cmd, &opts.Output)

	carapace.Gen(cmd).PositionalCompletion(actionLibraryRefs(f, cmd))

	return cmd
}

// runShow executes the show command: parses the ref, resolves the
// resource or preset, and renders the result in the requested format.
func runShow(opts *showOptions) error {
	if opts.IO != nil && opts.IO.Logger != nil {
		opts.IO.Logger.Debug("showing library entry", "ref", opts.Ref)
	}
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	if strings.HasPrefix(opts.Ref, presetPrefix) {
		presetName := strings.TrimPrefix(opts.Ref, presetPrefix)
		return renderPreset(opts, lib, presetName)
	}

	return renderResource(opts, lib, opts.Ref)
}

// renderResource resolves a resource ref against the library and
// renders it. On miss returns *core.NotFoundError.
func renderResource(opts *showOptions, lib *library.Library, ref string) error {
	typ, name, err := library.ParseRef(ref)
	if err != nil {
		return core.NewNotFoundError("library ref", opts.Ref)
	}

	resources, ok := lib.Resources[typ]
	if !ok {
		return core.NewNotFoundError("library ref", opts.Ref)
	}
	res, ok := resources[name]
	if !ok {
		return core.NewNotFoundError("library ref", opts.Ref)
	}

	switch opts.Output {
	case "json":
		payload := showResourceRow{Ref: ref, Path: res.Path, Description: res.Description}
		if err := output.NewJSONExporter().Write(opts.IO, payload); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		row := showResourceRow{Ref: ref, Path: res.Path, Description: res.Description}
		if err := output.NewTableExporter().Write(opts.IO, []showResourceRow{row}); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		out := formatResourceDetails(ref, res)
		if _, err := fmt.Fprint(opts.IO.Out, out); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// renderPreset resolves a preset name against the library and renders
// it. On miss returns *core.NotFoundError.
func renderPreset(opts *showOptions, lib *library.Library, presetName string) error {
	preset, ok := lib.Presets[presetName]
	if !ok {
		return core.NewNotFoundError("library ref", opts.Ref)
	}

	switch opts.Output {
	case "json":
		payload := showPresetRow{
			Name:        preset.Name,
			Description: preset.Description,
			Resources:   preset.Resources,
		}
		if err := output.NewJSONExporter().Write(opts.IO, payload); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		row := showPresetRow{
			Name:        preset.Name,
			Description: preset.Description,
			Resources:   preset.Resources,
		}
		if err := output.NewTableExporter().Write(opts.IO, []showPresetRow{row}); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		out := formatPresetDetails(presetName, preset)
		if _, err := fmt.Fprint(opts.IO.Out, out); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// formatResourceDetails renders a resource's details as plain text.
// Output is byte-identical to the pre-change build.
func formatResourceDetails(ref string, res library.Resource) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Reference: %s\n", ref)
	fmt.Fprintf(&sb, "Path: %s\n", res.Path)
	if res.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", res.Description)
	}
	return sb.String()
}

// formatPresetDetails renders a preset's details as plain text.
// Output is byte-identical to the pre-change build.
func formatPresetDetails(name string, preset library.Preset) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Preset: %s\n", name)
	if preset.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", preset.Description)
	}
	sb.WriteString("Resources:\n")
	for _, ref := range preset.Resources {
		fmt.Fprintf(&sb, "  - %s\n", ref)
	}
	return sb.String()
}
