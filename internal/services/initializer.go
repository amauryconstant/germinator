// Package services provides business logic for document transformation and validation.
package services

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/library"
)

// InitOptions contains options for the initialization process.
type InitOptions struct {
	// Library is the loaded library.
	Library *library.Library
	// Platform is the target platform (opencode or claude-code).
	Platform string
	// OutputDir is the base output directory.
	OutputDir string
	// DryRun indicates whether to preview changes without writing.
	DryRun bool
	// Force indicates whether to overwrite existing files.
	Force bool
}

// InitResult contains the result of initializing a single resource.
type InitResult struct {
	// Ref is the resource reference (e.g., "skill/commit").
	Ref string
	// InputPath is the source file path.
	InputPath string
	// OutputPath is the destination file path.
	OutputPath string
	// Error is any error that occurred during initialization.
	Error error
}

// InitializeResources installs resources from the library to the target directory.
// It uses fail-fast error handling - stops on first error.
func InitializeResources(opts InitOptions, refs []string) ([]InitResult, error) {
	results := make([]InitResult, 0, len(refs))

	for _, ref := range refs {
		result := InitResult{Ref: ref}

		// Resolve resource to file path
		inputPath, err := library.ResolveResource(opts.Library, ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, fmt.Errorf("failed to resolve %s: %w", ref, err)
		}
		result.InputPath = inputPath

		// Get output path
		typ, name, err := library.ParseRef(ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}

		outputPath, err := library.GetOutputPath(typ, name, opts.Platform, opts.OutputDir)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, fmt.Errorf("failed to get output path for %s: %w", ref, err)
		}
		result.OutputPath = outputPath

		// Check if file exists (unless force or dry-run)
		if !opts.DryRun && !opts.Force {
			if _, err := os.Stat(outputPath); err == nil {
				result.Error = fmt.Errorf("file exists: %s (use --force to overwrite)", outputPath)
				results = append(results, result)
				return results, result.Error
			}
		}

		// In dry-run mode, just record what would happen
		if opts.DryRun {
			results = append(results, result)
			continue
		}

		// Load the document
		doc, err := core.LoadDocument(inputPath, opts.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, fmt.Errorf("failed to load %s: %w", inputPath, err)
		}

		// Render the document
		rendered, err := core.RenderDocument(doc, opts.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, fmt.Errorf("failed to render %s: %w", ref, err)
		}

		// Create output directory
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Error = gerrors.NewFileError(outputPath, "mkdir", "failed to create output directory", err)
			results = append(results, result)
			return results, result.Error
		}

		// Write the file
		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
			result.Error = gerrors.NewFileError(outputPath, "write", "failed to write output file", err)
			results = append(results, result)
			return results, result.Error
		}

		results = append(results, result)
	}

	return results, nil
}

// InitializeFromPreset installs all resources from a preset.
func InitializeFromPreset(opts InitOptions, presetName string) ([]InitResult, error) {
	refs, err := library.ResolvePreset(opts.Library, presetName)
	if err != nil {
		return nil, err
	}

	return InitializeResources(opts, refs)
}

// FormatDryRunOutput formats the results for dry-run display.
func FormatDryRunOutput(results []InitResult) string {
	var output string
	for _, result := range results {
		output += fmt.Sprintf("Would write: %s\n  from: %s\n", result.OutputPath, result.InputPath)
	}
	return output
}

// FormatSuccessOutput formats the results for success display.
func FormatSuccessOutput(results []InitResult) string {
	var output string
	for _, result := range results {
		output += fmt.Sprintf("Installed: %s -> %s\n", result.Ref, result.OutputPath)
	}
	return output
}
