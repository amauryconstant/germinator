package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

// AddResourceJSONOutput represents JSON output for successful resource add.
type AddResourceJSONOutput struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	LibraryPath string `json:"libraryPath"`
}

// AddResourceErrorJSON represents JSON output for failed resource add.
type AddResourceErrorJSON struct {
	Error string `json:"error"`
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
}

// DiscoverJSONOutput represents JSON output for orphan discovery.
type DiscoverJSONOutput struct {
	Orphans     []OrphanInfoJSON    `json:"orphans,omitempty"`
	Added       []AddSuccessJSON    `json:"added,omitempty"`
	Conflicts   []ConflictInfoJSON  `json:"conflicts,omitempty"`
	Summary     DiscoverSummaryJSON `json:"summary,omitempty"`
	DryRun      bool                `json:"dryRun"`
	LibraryPath string              `json:"libraryPath"`
}

// OrphanInfoJSON represents an orphaned resource in JSON output.
type OrphanInfoJSON struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	Issue string `json:"issue,omitempty"`
}

// AddSuccessJSON represents a successfully added orphan in JSON output.
type AddSuccessJSON struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// ConflictInfoJSON represents a conflict during orphan discovery.
type ConflictInfoJSON struct {
	Orphan OrphanInfoJSON `json:"orphan"`
	Issue  string         `json:"issue"`
}

// DiscoverSummaryJSON represents discovery statistics in JSON output.
type DiscoverSummaryJSON struct {
	TotalScanned int `json:"totalScanned"`
	TotalOrphans int `json:"totalOrphans"`
	TotalAdded   int `json:"totalAdded"`
	TotalSkipped int `json:"totalSkipped"`
	TotalFailed  int `json:"totalFailed"`
}

// NewLibraryAddCommand creates the library add subcommand.
func NewLibraryAddCommand(cfg *CommandConfig, libraryPath *string) *cobra.Command {
	var opts struct {
		name        string
		description string
		resType     string
		platform    string
		force       bool
		dryRun      bool
		discover    bool
		batch       bool
	}

	cmd := &cobra.Command{
		Use:   "add [source]",
		Short: "Add a resource to the library",
		Long: `Add a resource from a source file to the library.

The source can be a canonical document or a platform-specific document.
Type, name, and description are auto-detected if not provided.

Alternatively, use --discover to find orphaned resource files not in library.yaml.

Examples:
  germinator library add skill-commit.md
  germinator library add agent-reviewer.md --type agent
  germinator library add code-reviewer.md --platform opencode
  germinator library add skill-commit.md --dry-run
  germinator library add --discover
  germinator library add --discover --force`,
		Args: func(_ *cobra.Command, args []string) error {
			// If --discover is set, no source is required
			if opts.discover {
				return nil
			}
			// Otherwise, exactly one argument required
			if len(args) != 1 {
				return errors.New("requires a source file argument (or use --discover to find orphans)")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			return runLibraryAdd(c, cfg, libraryPath, &opts, args)
		},
	}

	cmd.Flags().StringVar(&opts.name, "name", "", "Resource name")
	cmd.Flags().StringVar(&opts.description, "description", "", "Resource description")
	cmd.Flags().StringVar(&opts.resType, "type", "", "Resource type (skill, agent, command, memory)")
	cmd.Flags().StringVar(&opts.platform, "platform", "", "Source platform (opencode, claude-code)")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing resource")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Preview changes without adding")
	cmd.Flags().BoolVar(&opts.discover, "discover", false, "Discover orphaned resource files not in library.yaml")
	cmd.Flags().BoolVar(&opts.batch, "batch", false, "Batch mode: process all orphans continuously (use with --discover --force)")

	return cmd
}

// runLibraryAdd executes the library add logic.
func runLibraryAdd(c *cobra.Command, cfg *CommandConfig, libraryPath *string, opts *struct {
	name        string
	description string
	resType     string
	platform    string
	force       bool
	dryRun      bool
	discover    bool
	batch       bool
}, args []string) error {
	verbosity, _ := c.Flags().GetCount("verbose")
	cfg.Verbosity = Verbosity(verbosity)

	// Discover library path
	envPath := os.Getenv("GERMINATOR_LIBRARY")
	path := library.FindLibrary(*libraryPath, envPath)

	VerbosePrint(cfg, "Using library at: %s", path)

	// Handle discover mode
	if opts.discover {
		return runLibraryDiscover(c, cfg, path, opts)
	}

	// Normal add mode - args[0] is the source
	source := args[0]

	// Detect resource type
	resType := detectResourceType(source, opts.resType)

	// Detect name
	name := detectResourceName(source, opts.name)

	// Detect description
	description := detectResourceDescription(source, opts.description)

	// Detect platform
	platform := detectResourcePlatform(source, opts.platform)

	// Canonicalize if needed (platform document)
	canonicalSource := source
	if platform != "" && !library.IsCanonicalFormat(source, resType) {
		VerbosePrint(cfg, "Canonicalizing %s document from %s platform", resType, platform)
		canonicalPath, err := canonicalizeToTemp(cfg, source, platform, resType)
		if err != nil {
			return err
		}
		canonicalSource = canonicalPath
	}

	// Add to library
	err := library.AddResource(library.AddOptions{
		Source:      canonicalSource,
		Name:        name,
		Description: description,
		Type:        resType,
		LibraryPath: path,
		DryRun:      opts.dryRun,
		Force:       opts.force,
	})
	if err != nil {
		jsonFlag, _ := c.Flags().GetBool("json")
		if jsonFlag {
			errOutput := AddResourceErrorJSON{
				Error: err.Error(),
				Type:  resType,
				Name:  name,
			}
			jsonErr, _ := json.Marshal(errOutput)
			_, _ = fmt.Fprintln(c.OutOrStderr(), string(jsonErr))
		}
		return fmt.Errorf("adding resource: %w", err)
	}

	// Output success
	jsonFlag, _ := c.Flags().GetBool("json")
	if jsonFlag {
		// Derive the path from library structure
		resourcePath := fmt.Sprintf("%ss/%s.md", resType, name)
		output := AddResourceJSONOutput{
			Type:        resType,
			Name:        name,
			Path:        resourcePath,
			LibraryPath: path,
		}
		jsonOutput, _ := json.Marshal(output)
		_, _ = fmt.Fprintln(c.OutOrStdout(), string(jsonOutput))
	} else {
		fmt.Printf("Added resource: %s/%s\n", resType, name)
	}

	return nil
}

// runLibraryDiscover executes the orphan discovery logic.
func runLibraryDiscover(c *cobra.Command, _ *CommandConfig, path string, opts *struct {
	name        string
	description string
	resType     string
	platform    string
	force       bool
	dryRun      bool
	discover    bool
	batch       bool
}) error {
	result, err := library.DiscoverOrphans(library.DiscoverOptions{
		LibraryPath: path,
		DryRun:      opts.dryRun,
		Force:       opts.force,
		Batch:       opts.batch,
	})
	if err != nil {
		return fmt.Errorf("discovering orphans: %w", err)
	}

	// Check for JSON output
	jsonFlag, _ := c.Flags().GetBool("json")
	if jsonFlag {
		return outputDiscoverJSON(c, result, path, opts.dryRun)
	}

	// Output results (human-readable)
	if opts.dryRun {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "Dry-run: no changes made")
	}

	if len(result.Orphans) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nOrphaned resources:")
		for _, orphan := range result.Orphans {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s/%s (%s)\n", orphan.Type, orphan.Name, orphan.Path)
		}
	}

	if len(result.Added) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nRegistered:")
		for _, added := range result.Added {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s/%s\n", added.Type, added.Name)
		}
	}

	if len(result.Conflicts) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nConflicts:")
		for _, conflict := range result.Conflicts {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s/%s: %s\n", conflict.Orphan.Type, conflict.Orphan.Name, conflict.Issue)
		}
	}

	// Output summary
	_, _ = fmt.Fprintf(c.OutOrStdout(), "\nSummary: scanned=%d, orphans=%d, added=%d, skipped=%d, failed=%d\n",
		result.Summary.TotalScanned, result.Summary.TotalOrphans,
		result.Summary.TotalAdded, result.Summary.TotalSkipped, result.Summary.TotalFailed)

	return nil
}

// outputDiscoverJSON outputs discovery results as JSON.
func outputDiscoverJSON(c *cobra.Command, result *library.DiscoverResult, path string, dryRun bool) error {
	output := DiscoverJSONOutput{
		Orphans:   make([]OrphanInfoJSON, 0, len(result.Orphans)),
		Added:     make([]AddSuccessJSON, 0, len(result.Added)),
		Conflicts: make([]ConflictInfoJSON, 0, len(result.Conflicts)),
		Summary: DiscoverSummaryJSON{
			TotalScanned: result.Summary.TotalScanned,
			TotalOrphans: result.Summary.TotalOrphans,
			TotalAdded:   result.Summary.TotalAdded,
			TotalSkipped: result.Summary.TotalSkipped,
			TotalFailed:  result.Summary.TotalFailed,
		},
		DryRun:      dryRun,
		LibraryPath: path,
	}

	for _, orphan := range result.Orphans {
		output.Orphans = append(output.Orphans, OrphanInfoJSON{
			Type:  orphan.Type,
			Name:  orphan.Name,
			Path:  orphan.Path,
			Issue: orphan.Issue,
		})
	}

	for _, added := range result.Added {
		output.Added = append(output.Added, AddSuccessJSON{
			Type: added.Type,
			Name: added.Name,
			Path: added.Path,
		})
	}

	for _, conflict := range result.Conflicts {
		output.Conflicts = append(output.Conflicts, ConflictInfoJSON{
			Orphan: OrphanInfoJSON{
				Type:  conflict.Orphan.Type,
				Name:  conflict.Orphan.Name,
				Path:  conflict.Orphan.Path,
				Issue: conflict.Orphan.Issue,
			},
			Issue: conflict.Issue,
		})
	}

	encoder := json.NewEncoder(c.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}

// detectResourceType determines the resource type from flag, filename, or frontmatter.
func detectResourceType(source, flag string) string {
	if flag != "" {
		return flag
	}
	if detected := library.DetectTypeFromFilename(source); detected != "" {
		return detected
	}
	if frontmatterType, _ := detectFrontmatterField(source, "type"); frontmatterType != "" {
		return frontmatterType
	}
	return ""
}

// detectResourceName determines the resource name from flag, frontmatter, or filename.
func detectResourceName(source, flag string) string {
	if flag != "" {
		return flag
	}
	if frontmatterName, _ := detectFrontmatterField(source, "name"); frontmatterName != "" {
		return frontmatterName
	}
	return ""
}

// detectResourceDescription determines the resource description from flag or frontmatter.
func detectResourceDescription(source, flag string) string {
	if flag != "" {
		return flag
	}
	if frontmatterDesc, _ := detectFrontmatterField(source, "description"); frontmatterDesc != "" {
		return frontmatterDesc
	}
	return ""
}

// detectResourcePlatform determines the platform from flag, frontmatter, or filename.
func detectResourcePlatform(source, flag string) string {
	if flag != "" {
		return flag
	}
	if frontmatterPlatform, _ := detectFrontmatterField(source, "platform"); frontmatterPlatform != "" {
		return frontmatterPlatform
	}
	return ""
}

// detectFrontmatterField extracts a field from YAML frontmatter.
func detectFrontmatterField(source, field string) (string, error) {
	content, err := os.ReadFile(source) //nolint:gosec // G304: User provides source path, intentionally reading user documents
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", source, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", errors.New("no frontmatter")
	}

	var yamlLines []string
	foundEnd := false

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			foundEnd = true
			break
		}
		yamlLines = append(yamlLines, lines[i])
	}

	if !foundEnd {
		return "", errors.New("no frontmatter end")
	}

	yamlContent := strings.Join(yamlLines, "\n")

	// Simple YAML parsing for a single field
	for _, line := range strings.Split(yamlContent, "\n") {
		if strings.HasPrefix(line, field+":") {
			value := strings.TrimSpace(strings.TrimPrefix(line, field+":"))
			value = strings.Trim(value, "\"")
			return value, nil
		}
	}

	return "", errors.New("field not found")
}

// canonicalizeToTemp converts a platform document to canonical format in a temp file.
func canonicalizeToTemp(cfg *CommandConfig, source, platform, docType string) (string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "germinator-canonical-*."+filepath.Ext(source))
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	// Use the canonicalizer service
	ctx := context.Background()
	req := &application.CanonicalizeRequest{
		InputPath:  source,
		OutputPath: tmpPath,
		Platform:   platform,
		DocType:    docType,
	}

	canonicalizer := cfg.Services.Canonicalizer
	result, err := canonicalizer.Canonicalize(ctx, req)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("canonicalizing: %w", err)
	}

	return result.OutputPath, nil
}
