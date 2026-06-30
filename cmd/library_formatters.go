package cmd

// TODO(slice-7): delete this file once slice-3/slice-6 consumers migrate
// off the remaining helpers (formatResourcesList, FormatBatchAddSummary).

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gitlab.com/amoconst/germinator/internal/library"
)

func formatResourcesList(lib *library.Library) string {
	var sb strings.Builder

	resources := library.ListResources(lib)

	typeOrder := []string{
		string(library.ResourceTypeSkill),
		string(library.ResourceTypeAgent),
		string(library.ResourceTypeCommand),
		string(library.ResourceTypeMemory),
	}

	hasContent := false
	for _, typ := range typeOrder {
		infos, ok := resources[typ]
		if !ok || len(infos) == 0 {
			continue
		}

		if hasContent {
			sb.WriteString("\n")
		}
		hasContent = true

		header := cases.Title(language.English).String(typ) + "s"
		sb.WriteString(header + ":\n")

		for _, info := range infos {
			ref := library.FormatRef(info.Type, info.Name)
			if info.Description != "" {
				fmt.Fprintf(&sb, "  %s - %s\n", ref, info.Description)
			} else {
				fmt.Fprintf(&sb, "  %s\n", ref)
			}
		}
	}

	if !hasContent {
		return "No resources found.\n"
	}

	return sb.String()
}

// FormatBatchAddSummary formats and outputs the batch add summary.
func FormatBatchAddSummary(c *cobra.Command, result *library.BatchAddResult) {
	if result == nil {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "No resources processed.")
		return
	}

	// Output each category if non-empty
	if len(result.Added) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nAdded:")
		for _, added := range result.Added {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s\n", added.Ref)
		}
	}

	if len(result.Skipped) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nSkipped:")
		for _, skipped := range result.Skipped {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s (%s)\n", skipped.Source, skipped.Issue)
		}
	}

	if len(result.Failed) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nFailed:")
		for _, failed := range result.Failed {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s: %s\n", failed.Source, failed.Error)
		}
	}

	// Output summary
	_, _ = fmt.Fprintf(c.OutOrStdout(), "\nAdded %d, skipped %d, failed %d\n",
		result.Summary.Added, result.Summary.Skipped, result.Summary.Failed)
}
