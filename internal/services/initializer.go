// Package services provides business logic for document transformation and validation.
package services

import (
	"context"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/library"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/internal/infrastructure/serialization"
)

// initializer implements the application.Initializer interface.
type initializer struct{}

// NewInitializer creates a new Initializer instance.
func NewInitializer() application.Initializer {
	return &initializer{}
}

// Initialize installs resources from the library to the target directory.
// It uses fail-fast error handling - stops on first error.
func (i *initializer) Initialize(_ context.Context, req *application.InitializeRequest) ([]domain.InitializeResult, error) {
	results := make([]domain.InitializeResult, 0, len(req.Refs))

	for _, ref := range req.Refs {
		result := domain.InitializeResult{Ref: ref}

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
				result.Error = domain.NewFileError(outputPath, "write", "file exists (use --force to overwrite)", nil)
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
		doc, err := parsing.LoadDocument(inputPath, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}

		// Render the document
		rendered, err := serialization.RenderDocument(doc, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			return results, err
		}

		// Create output directory
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Error = domain.NewFileError(outputPath, "mkdir", "failed to create output directory", err)
			results = append(results, result)
			return results, result.Error
		}

		// Write the file
		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
			result.Error = domain.NewFileError(outputPath, "write", "failed to write output file", err)
			results = append(results, result)
			return results, result.Error
		}

		results = append(results, result)
	}

	return results, nil
}

// Compile-time interface satisfaction check.
var _ application.Initializer = (*initializer)(nil)
