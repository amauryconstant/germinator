package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
)

// libraryResourcesOptions holds the runtime state for an
// `library resources` invocation.
type libraryResourcesOptions struct {
	IO      *iostreams.IOStreams
	Library func() (*library.Library, error)
	Ctx     context.Context
	Output  string
}

// resourcesRow is the table-exporter representation of a single
// resource. The tab struct tags drive the TableExporter column
// header order; the JSONExporter uses the `json` tags for marshaling.
// `Description` is NOT `omitempty` because the library-library-json-output
// delta spec requires a stable JSON shape of
// {"type":"...","name":"...","description":"...","path":"..."}.
type resourcesRow struct {
	Type        string `tab:"TYPE"        json:"type"`
	Name        string `tab:"NAME"        json:"name"`
	Path        string `tab:"-"           json:"path"`
	Description string `tab:"DESCRIPTION" json:"description"`
}

// NewCmdResources creates the `library resources` subcommand via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 2.
//
// libraryPath is a shared *string with the parent `library` command
// (set by `cmd.PersistentFlags().StringVar` in NewLibraryCommand).
// Passing the pointer keeps the parent's --library flag working;
// the resolved path is used to construct the Library lazy field
// per invocation (avoids caching across different --library values
// within a single process).
func NewCmdResources(f *cmdutil.Factory, libraryPath *string, runF func(*libraryResourcesOptions) error) *cobra.Command {
	opts := &libraryResourcesOptions{}
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List all resources in the library",
		Long: `List all resources in the library, grouped by type.

Resources are displayed in sections: Skills, Agents, Commands, Memory.

Output formats (--output):
  plain  default; byte-identical to the pre-change plain output
  json   JSON document suitable for jq
  table  tab-aligned text table

Example:
  germinator library resources
  germinator library resources --library /path/to/library
  germinator library resources --output json
  germinator library resources --output table`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = c.Context()
			// Resolve library path per invocation so changes to
			// --library between calls are respected (the Factory's
			// f.Library uses the env var only; the --library flag
			// is parent-persistent and lives on libraryPath).
			lp := ""
			if libraryPath != nil {
				lp = *libraryPath
			}
			resolved := library.FindLibrary(lp, os.Getenv("GERMINATOR_LIBRARY"))
			opts.Library = func() (*library.Library, error) {
				return library.LoadLibrary(resolved)
			}
			if runF != nil {
				return runF(opts)
			}
			return runResources(opts)
		},
	}

	cmdutil.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// runResources executes the resources listing logic. Plain output is
// reused from the existing formatResourcesList helper to guarantee
// byte-identical output with the pre-change build.
func runResources(opts *libraryResourcesOptions) error {
	lib, err := opts.Library()
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	switch opts.Output {
	case "json":
		rows := flattenResources(lib)
		if err := output.NewJSONExporter().Write(opts.IO, struct {
			Resources []resourcesRow `json:"resources"`
		}{Resources: rows}); err != nil {
			return fmt.Errorf("writing json output: %w", err)
		}
		return nil
	case "table":
		rows := flattenResources(lib)
		if err := output.NewTableExporter().Write(opts.IO, rows); err != nil {
			return fmt.Errorf("writing table output: %w", err)
		}
		return nil
	default:
		if _, err := fmt.Fprint(opts.IO.Out, formatResourcesList(lib)); err != nil {
			return fmt.Errorf("writing plain output: %w", err)
		}
		return nil
	}
}

// flattenResources produces a deterministic slice of resourcesRow
// (sorted by type then name) suitable for both JSON and table
// exporters. Uses library.ListResources which returns a
// map[string][]ResourceInfo.
func flattenResources(lib *library.Library) []resourcesRow {
	grouped := library.ListResources(lib)

	typeOrder := []string{
		string(library.ResourceTypeSkill),
		string(library.ResourceTypeAgent),
		string(library.ResourceTypeCommand),
		string(library.ResourceTypeMemory),
	}

	rows := make([]resourcesRow, 0)
	for _, typ := range typeOrder {
		infos, ok := grouped[typ]
		if !ok {
			continue
		}
		for _, info := range infos {
			rows = append(rows, resourcesRow{
				Type:        info.Type,
				Name:        info.Name,
				Path:        info.Path,
				Description: info.Description,
			})
		}
	}
	return rows
}
