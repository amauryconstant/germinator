// Package service provides business logic for document transformation and validation.
package service

import (
	"context"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
)

// initializer implements the application.Initializer interface.
type initializer struct {
	parser     application.Parser
	serializer application.Serializer
}

// NewInitializer creates a new Initializer instance.
func NewInitializer(parser application.Parser, serializer application.Serializer) application.Initializer {
	return &initializer{
		parser:     parser,
		serializer: serializer,
	}
}

// Initialize installs resources from the library to the target directory.
// It uses partial processing - continues on individual errors, collecting all results.
// Returns error only if ALL resources fail; returns nil if at least one succeeds.
func (i *initializer) Initialize(_ context.Context, req *application.InitializeRequest) ([]core.InitializeResult, error) {
	results := make([]core.InitializeResult, 0, len(req.Refs))

	for _, ref := range req.Refs {
		result := core.InitializeResult{Ref: ref}

		// Resolve resource to file path
		inputPath, err := library.ResolveResource(req.Library, ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		result.InputPath = inputPath

		// Get output path
		typ, name, err := library.ParseRef(ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		outputPath, err := library.GetOutputPath(typ, name, req.Platform, req.OutputDir)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		result.OutputPath = outputPath

		// Check if file exists (unless force or dry-run)
		if !req.DryRun && !req.Force {
			if _, err := os.Stat(outputPath); err == nil {
				result.Error = core.NewFileError(outputPath, "write", "file exists (use --force to overwrite)", nil)
				results = append(results, result)
				continue
			}
		}

		// In dry-run mode, just record what would happen
		if req.DryRun {
			results = append(results, result)
			continue
		}

		// Load the document
		doc, err := i.parser.LoadDocument(inputPath, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		// Render the document
		rendered, err := i.serializer.RenderDocument(doc, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		// Create output directory
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil { //nolint:gosec // G301: User owns output directory, 0755 is standard permission
			result.Error = core.NewFileError(outputPath, "mkdir", "failed to create output directory", err)
			results = append(results, result)
			continue
		}

		// Write the file
		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
			result.Error = core.NewFileError(outputPath, "write", "failed to write output file", err)
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	// Per the slice-5 contract: error is reserved for transport-level
	// failures. Per-resource outcomes (success or per-resource error)
	// always live in result.Error; callers inspect the slice to detect
	// partial success / all-failed / all-success and synthesize the
	// appropriate *core.PartialSuccessError (or nil).
	return results, nil
}

// Compile-time interface satisfaction check.
var _ application.Initializer = (*initializer)(nil)
