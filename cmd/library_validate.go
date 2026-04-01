package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// NewLibraryValidateCommand creates the library validate subcommand.
func NewLibraryValidateCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	var opts struct {
		fix  bool
		json bool
	}

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

Examples:
  germinator library validate
  germinator library validate --library /path/to/library
  germinator library validate --json
  germinator library validate --fix`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runLibraryValidate(c, cfg, libraryPath, &opts)
		},
	}

	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Auto-clean library.yaml (removes missing entries and ghost preset refs)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")

	return cmd
}

// runLibraryValidate executes the library validate logic.
func runLibraryValidate(c *cobra.Command, cfg *CommandConfig, libraryPath *string, opts *struct {
	fix  bool
	json bool
}) error {
	verbosity, _ := c.Flags().GetCount("verbose")
	cfg.Verbosity = Verbosity(verbosity)

	// Discover library path
	envPath := os.Getenv("GERMINATOR_LIBRARY")
	path := library.FindLibrary(*libraryPath, envPath)

	VerbosePrint(cfg, "Loading library from: %s", path)

	// Load library
	lib, err := library.LoadLibrary(path)
	if err != nil {
		return fmt.Errorf("loading library: %w", err)
	}

	// Validate library
	result, err := library.ValidateLibrary(lib)
	if err != nil {
		return fmt.Errorf("validating library: %w", err)
	}

	// Apply fixes if requested
	if opts.fix && !result.Valid {
		fixResult, err := library.FixLibrary(lib)
		if err != nil {
			return fmt.Errorf("fixing library: %w", err)
		}
		result.FixApplied = true
		result.FixResult = fixResult
	}

	// Output results
	if opts.json {
		return outputJSON(c, result)
	}
	return outputHuman(c, cfg, result, opts.fix)
}

// ValidationOutput is the JSON output structure for validation results.
type ValidationOutput struct {
	Valid        bool              `json:"valid"`
	ErrorCount   int               `json:"errorCount"`
	WarningCount int               `json:"warningCount"`
	Issues       []ValidationIssue `json:"issues"`
}

// ValidationIssue represents a single issue in JSON output.
type ValidationIssue struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Ref      string `json:"ref,omitempty"`
	Path     string `json:"path,omitempty"`
	InPreset string `json:"inPreset,omitempty"`
	Message  string `json:"message,omitempty"`
}

// outputJSON outputs validation results as JSON.
func outputJSON(c *cobra.Command, result *library.ValidationResult) error {
	output := ValidationOutput{
		Valid:        result.Valid,
		ErrorCount:   result.ErrorCount,
		WarningCount: result.WarningCount,
		Issues:       make([]ValidationIssue, 0, len(result.Issues)),
	}

	for _, issue := range result.Issues {
		output.Issues = append(output.Issues, ValidationIssue{
			Type:     string(issue.Type),
			Severity: string(issue.Severity),
			Ref:      issue.Ref,
			Path:     issue.Path,
			InPreset: issue.InPreset,
			Message:  issue.Message,
		})
	}

	// Sort issues for deterministic output
	sort.Slice(output.Issues, func(i, j int) bool {
		if output.Issues[i].Severity != output.Issues[j].Severity {
			return output.Issues[i].Severity == "error"
		}
		return output.Issues[i].Type < output.Issues[j].Type
	})

	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// outputHuman outputs validation results in human-readable format.
func outputHuman(c *cobra.Command, _ *CommandConfig, result *library.ValidationResult, fix bool) error {
	var sb strings.Builder

	outputHeader(&sb, result)
	outputIssues(&sb, result)
	outputFooter(&sb, result, fix)

	_, err := fmt.Fprint(c.OutOrStdout(), sb.String())
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

func outputHeader(sb *strings.Builder, result *library.ValidationResult) {
	if result.Valid && len(result.Issues) == 0 {
		fmt.Fprintln(sb, "✓ Library is valid")
		fmt.Fprintf(sb, "  errors: 0, warnings: 0\n")
	} else if result.Valid {
		fmt.Fprintln(sb, "✓ Library is valid (warnings only)")
		fmt.Fprintf(sb, "  errors: 0, warnings: %d\n", result.WarningCount)
		fmt.Fprintln(sb)
	} else {
		fmt.Fprintln(sb, "✗ Library has issues")
		fmt.Fprintf(sb, "  errors: %d, warnings: %d\n", result.ErrorCount, result.WarningCount)
		fmt.Fprintln(sb)
	}
}

func outputIssues(sb *strings.Builder, result *library.ValidationResult) {
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
					formatIssueLine(sb, issue)
				}
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Fprintln(sb, "Warnings:")
		for _, issue := range warnings {
			formatIssueLine(sb, issue)
		}
	}
}

func formatIssueLine(sb *strings.Builder, issue library.Issue) {
	fmt.Fprintf(sb, "  [%s] %s", formatIssueType(issue.Type), formatIssueRefOrPath(issue))
	if issue.InPreset != "" {
		fmt.Fprintf(sb, " (in preset %q)", issue.InPreset)
	}
	fmt.Fprintln(sb)
	if issue.Message != "" {
		fmt.Fprintf(sb, "    %s\n", issue.Message)
	}
}

func outputFooter(sb *strings.Builder, result *library.ValidationResult, fix bool) {
	fmt.Fprintln(sb)
	if !result.Valid {
		if fix {
			fmt.Fprintln(sb, "Fix applied to library.yaml")
		} else {
			fmt.Fprintln(sb, "Hint: Run with --fix to auto-clean library.yaml")
		}
	}
	fmt.Fprintln(sb, "Hint: Run with --json for machine-readable output")
}

// formatIssueType returns a human-readable issue type.
func formatIssueType(t library.IssueType) string {
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

// formatIssueRefOrPath returns the ref or path for an issue.
func formatIssueRefOrPath(issue library.Issue) string {
	if issue.Ref != "" {
		return issue.Ref
	}
	return issue.Path
}
