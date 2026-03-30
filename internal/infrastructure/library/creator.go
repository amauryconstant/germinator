// Package library provides library management for canonical resources.
package library

import (
	"fmt"
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/domain"
)

// CreateOptions contains options for library creation.
type CreateOptions struct {
	// Path is the directory path where the library will be created.
	Path string
	// DryRun preview changes without creating files or directories.
	DryRun bool
	// Force overwrites an existing library at the target path.
	Force bool
}

// CreateLibrary creates a new library directory structure at the specified path.
// It creates library.yaml and empty resource directories (skills, agents, commands, memory).
// If DryRun is true, it prints what would be created without making changes.
// If Force is false and a library already exists at Path, an error is returned.
func CreateLibrary(opts CreateOptions) error {
	// Check if library already exists
	exists := Exists(opts.Path)
	if exists && !opts.Force {
		return gerrors.NewFileError(opts.Path, "create", "library already exists at path (use --force to overwrite)", nil)
	}

	// Dry run mode - print what would be created
	if opts.DryRun {
		fmt.Printf("Would create library at: %s\n", opts.Path)
		fmt.Printf("  - %s/library.yaml\n", opts.Path)
		fmt.Printf("  - %s/skills/\n", opts.Path)
		fmt.Printf("  - %s/agents/\n", opts.Path)
		fmt.Printf("  - %s/commands/\n", opts.Path)
		fmt.Printf("  - %s/memory/\n", opts.Path)
		return nil
	}

	// Create directory structure
	dirs := []string{
		opts.Path,
		filepath.Join(opts.Path, "skills"),
		filepath.Join(opts.Path, "agents"),
		filepath.Join(opts.Path, "commands"),
		filepath.Join(opts.Path, "memory"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return gerrors.NewFileError(dir, "create", "failed to create directory", err)
		}
	}

	// Create library.yaml
	yamlContent := defaultLibraryYAML()
	yamlPath := filepath.Join(opts.Path, "library.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644); err != nil { //nolint:gosec // G302: Creating new file with standard permissions
		return gerrors.NewFileError(yamlPath, "write", "failed to write library.yaml", err)
	}

	// Validate created library by loading it
	if _, err := LoadLibrary(opts.Path); err != nil {
		// Validation failed - leave partial structure for debugging
		return fmt.Errorf("library created but validation failed: %w (partial structure left for debugging)", err)
	}

	return nil
}

// defaultLibraryYAML returns the default library.yaml content.
func defaultLibraryYAML() string {
	return `version: "1"
resources:
  skill: {}
  agent: {}
  command: {}
  memory: {}
presets: {}
`
}
