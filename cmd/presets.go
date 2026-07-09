package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// presetsOptions holds the runtime state for a `library presets`
// invocation.
type presetsOptions struct {
	IO      *iostreams.IOStreams
	Library func() (*library.Library, error)
	Ctx     context.Context
	Output  string
}

// presetsRow is the table-exporter representation of a single preset.
// The tab struct tags drive the TableExporter column header order; the
// JSONExporter uses the `json` tags for marshaling.
type presetsRow struct {
	Name        string   `tab:"NAME"        json:"name"`
	Description string   `tab:"DESCRIPTION" json:"description,omitempty"`
	Resources   []string `tab:"RESOURCES"   json:"resources"`
}

// NewCmdPresets creates the `library presets` subcommand via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 4.
//
// libraryPath is a shared *string with the parent `library` command
// (set by `cmd.PersistentFlags().StringVar` in NewLibraryCommand).
// Passing the pointer keeps the parent's --library flag working;
// the resolved path is used to construct the Library lazy field
// per invocation.
func NewCmdPresets(f *cmdutil.Factory, libraryPath *string, runF func(*presetsOptions) error) *cobra.Command {
	opts := &presetsOptions{}
	cmd := &cobra.Command{
		Use:   "presets",
		Short: "List all presets in the library",
		Long: `List all presets in the library with their descriptions and resources.

Output formats (--output):
  plain  default; byte-identical to the pre-change plain output
  json   JSON document suitable for jq
  table  tab-aligned text table

Example:
  germinator library presets
  germinator library presets --library /path/to/library
  germinator library presets --output json
  germinator library presets --output table`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = c.Context()
			lp := ""
			if libraryPath != nil {
				lp = *libraryPath
			}
			resolved := library.FindLibrary(lp, os.Getenv("GERMINATOR_LIBRARY"))
			opts.Library = func() (*library.Library, error) {
				return library.LoadLibrary(opts.Ctx, resolved)
			}
			if runF != nil {
				return runF(opts)
			}
			return runPresets(opts)
		},
	}

	cmdutil.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// runPresets executes the presets listing logic. Plain output is
// produced by the local formatPresetsList helper to guarantee
// byte-identical output with the pre-change build.
func runPresets(opts *presetsOptions) error {
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	switch opts.Output {
	case "json":
		rows := flattenPresets(lib)
		if err := output.NewJSONExporter().Write(opts.IO, struct {
			Presets []presetsRow `json:"presets"`
		}{Presets: rows}); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		rows := flattenPresets(lib)
		if err := output.NewTableExporter().Write(opts.IO, rows); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		if _, err := fmt.Fprint(opts.IO.Out, formatPresetsList(lib)); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// flattenPresets produces a deterministic slice of presetsRow
// (sorted by name, provided by library.ListPresets) suitable for both
// JSON and table exporters.
func flattenPresets(lib *library.Library) []presetsRow {
	presets := library.ListPresets(lib)
	rows := make([]presetsRow, 0, len(presets))
	for _, p := range presets {
		resources := make([]string, len(p.Resources))
		copy(resources, p.Resources)
		rows = append(rows, presetsRow{
			Name:        p.Name,
			Description: p.Description,
			Resources:   resources,
		})
	}
	return rows
}

// formatPresetsList renders the preset list as plain text. The output
// is byte-identical to the pre-change build.
func formatPresetsList(lib *library.Library) string {
	var sb strings.Builder

	presets := library.ListPresets(lib)

	if len(presets) == 0 {
		return "No presets found.\n"
	}

	for _, preset := range presets {
		if preset.Description != "" {
			fmt.Fprintf(&sb, "%s - %s\n", preset.Name, preset.Description)
		} else {
			fmt.Fprintf(&sb, "%s\n", preset.Name)
		}

		for _, ref := range preset.Resources {
			fmt.Fprintf(&sb, "  - %s\n", ref)
		}
	}

	return sb.String()
}
