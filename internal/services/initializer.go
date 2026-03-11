// Package services provides business logic for document transformation and validation.
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/library"
)

// InitOptions contains options for the initialization process.
// Deprecated: Use application.InitializeRequest instead.
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

// initializer implements the application.Initializer interface.
type initializer struct{}

// NewInitializer creates a new Initializer instance.
func NewInitializer() application.Initializer {
	return &initializer{}
}

// Initialize installs resources from the library to the target directory.
// It uses fail-fast error handling - stops on first error.
func (i *initializer) Initialize(ctx context.Context, req *application.InitializeRequest) ([]application.InitializeResult, error) {
	results := make([]application.InitializeResult, 0, len(req.Refs))

	for _, ref := range req.Refs {
		result := application.InitializeResult{Ref: ref}

		// Resolve resource to file path
		inputPath, err := library.ResolveResource(req.Library, ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}
		result.InputPath = inputPath

		// Get output path
		typ, name, err := library.ParseRef(ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}

		outputPath, err := library.GetOutputPath(typ, name, req.Platform, req.OutputDir)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}
		result.OutputPath = outputPath

		// Check if file exists (unless force or dry-run)
		if !req.DryRun && !req.Force {
			if _, err := os.Stat(outputPath); err == nil {
				result.Error = gerrors.NewFileError(outputPath, "write", "file exists (use --force to overwrite)", nil)
				results = append(results, result)
				return results, result.Error
			}
		}

		// In dry-run mode, just record what would happen
		if req.DryRun {
			results = append(results, result)
			continue
		}

		// Load the document
		doc, err := core.LoadDocument(inputPath, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}

		// Render the document
		rendered, err := core.RenderDocument(doc, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
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

// Compile-time interface satisfaction check.
var _ application.Initializer = (*initializer)(nil)

// InitResult contains the result of initializing a single resource.
// Deprecated: Use application.InitializeResult instead.
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
