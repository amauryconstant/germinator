package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
)

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
		return fmt.Errorf("adding resource: %w", err)
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
}) error {
	result, err := library.DiscoverOrphans(library.DiscoverOptions{
		LibraryPath: path,
		DryRun:      opts.dryRun,
		Force:       opts.force,
	})
	if err != nil {
		return fmt.Errorf("discovering orphans: %w", err)
	}

	// Output results
	if opts.dryRun {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "Dry-run: no changes made")
	}

	if len(result.Orphans) > 0 {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "\nOrphaned resources:")
		for _, orphan := range result.Orphans {
			_, _ = fmt.Fprintf(c.OutOrStdout(), "  %s/%s - %s\n", orphan.Type, orphan.Name, orphan.Description)
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
