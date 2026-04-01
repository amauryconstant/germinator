package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryRefreshCommand creates the library refresh subcommand.
func NewLibraryRefreshCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	var opts struct {
		dryRun bool
		force  bool
		json   bool
	}

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
  germinator library refresh --json`,
		RunE: func(c *cobra.Command, _ []string) error {
			return runLibraryRefresh(c, cfg, libraryPath, &opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without modifying library.yaml")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Skip resources with conflicts")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")

	return cmd
}

// runLibraryRefresh executes the library refresh logic.
func runLibraryRefresh(c *cobra.Command, cfg *CommandConfig, libraryPath *string, opts *struct {
	dryRun bool
	force  bool
	json   bool
}) error {
	verbosity, _ := c.Flags().GetCount("verbose")
	cfg.Verbosity = Verbosity(verbosity)

	// Discover library path
	envPath := os.Getenv("GERMINATOR_LIBRARY")
	path := library.FindLibrary(*libraryPath, envPath)

	VerbosePrint(cfg, "Using library at: %s", path)

	// Refresh the library
	result, err := library.RefreshLibrary(library.RefreshOptions{
		LibraryPath: path,
		DryRun:      opts.dryRun,
		Force:       opts.force,
	})
	if err != nil {
		return fmt.Errorf("refreshing library: %w", err)
	}

	// Output results
	if opts.json {
		return outputRefreshJSON(c, result)
	}

	outputRefresh(c, result, opts.dryRun)

	return nil
}

// outputRefresh outputs human-readable refresh results.
func outputRefresh(c *cobra.Command, result *library.RefreshResult, dryRun bool) {
	if dryRun {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "Dry-run: no changes made")
	}

	if len(result.Refreshed) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nRefreshed:")
		for _, r := range result.Refreshed {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  [refreshed] %s: %s (%s → %s)\n", r.Ref, r.Field, r.Old, r.New)
		}
	}

	if len(result.Skipped) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nSkipped:")
		for _, s := range result.Skipped {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  [skipped] %s: %s\n", s.Ref, s.Reason)
		}
	}

	if len(result.Errors) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nErrors:")
		for _, e := range result.Errors {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  [error] %s: %s (%s)\n", e.Ref, e.Type, e.Field)
		}
	}
}

// outputRefreshJSON outputs JSON-formatted refresh results.
func outputRefreshJSON(c *cobra.Command, result *library.RefreshResult) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "{\n")
	fmt.Fprintf(&sb, "  \"refreshed\": %d,\n", len(result.Refreshed))
	fmt.Fprintf(&sb, "  \"skipped\": %d,\n", len(result.Skipped))
	fmt.Fprintf(&sb, "  \"errors\": %d,\n", len(result.Errors))
	sb.WriteString("  \"details\": {\n")

	// Refreshed items
	sb.WriteString("    \"refreshed\": [")
	for i, r := range result.Refreshed {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "\n      {\"ref\": %q, \"field\": %q, \"old\": %q, \"new\": %q}",
			r.Ref, r.Field, r.Old, r.New)
	}
	sb.WriteString("\n    ],\n")

	// Skipped items
	sb.WriteString("    \"skipped\": [")
	for i, s := range result.Skipped {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "\n      {\"ref\": %q, \"reason\": %q}",
			s.Ref, s.Reason)
	}
	sb.WriteString("\n    ],\n")

	// Error items
	sb.WriteString("    \"errors\": [")
	for i, e := range result.Errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "\n      {\"ref\": %q, \"type\": %q, \"field\": %q}",
			e.Ref, e.Type, e.Field)
	}
	sb.WriteString("\n    ]\n")

	sb.WriteString("  }\n")
	sb.WriteString("}\n")

	_, err := c.OutOrStdout().Write([]byte(sb.String()))
	if err != nil {
		return fmt.Errorf("writing JSON output: %w", err)
	}
	return nil
}
