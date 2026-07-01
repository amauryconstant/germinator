package cmd

import (
	"context"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// Initializer is the command-side contract for resource installation.
// Defined in cmd/ per the target architecture ("interfaces where
// consumed" — golang-cli-architecture principle 8). Mirrors the
// slice-3 validator/canonicalizer pattern.
type Initializer interface {
	Initialize(ctx context.Context, req *InitializeRequest) ([]core.InitializeResult, error)
}

// NewInitializer returns the production Initializer backed by the
// canonical parser/serializer pair.
func NewInitializer() Initializer {
	return &initializerAdapter{
		parser:     parser.NewParser(),
		serializer: renderer.NewSerializer(),
	}
}

// initializerAdapter implements the local Initializer interface. The
// install loop resolves each ref, derives its output path, fails
// fast on existing files unless --force or --dry-run, then load →
// render → write under the matching output directory. Uses a
// Library-rooted InitializeRequest (the field is *library.Library,
// not LibraryPath string) so the loader step can rely on the
// receiver to provide RootPath.
type initializerAdapter struct {
	parser     *parser.Parser
	serializer *renderer.Serializer
}

// Compile-time confirmation that *initializerAdapter satisfies the
// cmd-side Initializer interface.
var _ Initializer = (*initializerAdapter)(nil)

// Initialize installs resources from the library to the target
// directory using partial processing — it continues on individual
// errors, collecting all per-resource results. The error return is
// reserved for transport-level failures; per-resource outcomes
// always live in result.Error, allowing callers to synthesize
// *core.PartialSuccessError.
func (i *initializerAdapter) Initialize(_ context.Context, req *InitializeRequest) ([]core.InitializeResult, error) {
	results := make([]core.InitializeResult, 0, len(req.Refs))

	for _, ref := range req.Refs {
		result := core.InitializeResult{Ref: ref}

		inputPath, err := library.ResolveResource(req.Library, ref)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		result.InputPath = inputPath

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

		if !req.DryRun && !req.Force {
			if _, err := os.Stat(outputPath); err == nil {
				result.Error = core.NewFileError(outputPath, "write", "file exists (use --force to overwrite)", nil)
				results = append(results, result)
				continue
			}
		}

		if req.DryRun {
			results = append(results, result)
			continue
		}

		doc, err := i.parser.LoadDocument(inputPath, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		rendered, err := i.serializer.RenderDocument(doc, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil { //nolint:gosec // G301: User owns output directory, 0755 is standard permission
			result.Error = core.NewFileError(outputPath, "mkdir", "failed to create output directory", err)
			results = append(results, result)
			continue
		}

		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
			result.Error = core.NewFileError(outputPath, "write", "failed to write output file", err)
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}
