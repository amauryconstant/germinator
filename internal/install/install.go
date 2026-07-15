// Package install provides resource installation as an I/O shell-package
// service. The package exists to satisfy the imperative-shell boundary
// defined by the golang-cli-architecture skill: any code that performs
// filesystem reads (parser.LoadDocument), platform rendering
// (renderer.RenderDocument), and filesystem writes (os.WriteFile + mkdir)
// for a batch of library refs lives here at the package edge, not in
// cmd/.
//
// The Service interface, Request type, and NewService constructor are
// the canonical contract. cmd/init.go declares a local Initializer
// interface that is structurally identical to install.Service (same
// parameter / return types); *installService satisfies both via
// structural typing so cmd/ does not have to import this package's
// adapter directly.
//
// The package name `install` was chosen over `init` to avoid collision
// with Go's reserved `init` identifier (per design Decision 1). The
// constructor takes the parser and serializer as injected dependencies
// so callers can construct them once and reuse across calls.
package install

import (
	"context"
	"os"
	"path/filepath"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// Request carries the inputs for resource installation. Lifted from
// cmd/init.go so the shell-package owns its types per the
// golang-cli-architecture "interfaces where consumed" rule (the
// consumer, cmd/init.go, declares its Initializer interface using
// this type via the shared import).
//
// Library carries a fully-loaded *library.Library (its RootPath feeds
// the loader steps); OutputDir is the base directory used by
// library.GetOutputPath to derive the per-resource output path.
type Request struct {
	Library   *library.Library
	Platform  string
	OutputDir string
	Refs      []string
	DryRun    bool
	Force     bool
}

// Service is the per-call contract for resource installation.
// Returns a slice of *core.InitializeResult (one per ref); per-resource
// errors live in result.Error so callers can synthesize a
// *core.PartialSuccessError. The error return is reserved for
// transport-level failures.
type Service interface {
	Initialize(ctx context.Context, req *Request) ([]core.InitializeResult, error)
}

// installService is the production implementation. Holds the
// injected parser and serializer so the same Service can be reused
// across many Initialize calls without rebuilding its orchestrating
// state. The struct fields are unexported so callers must use
// NewService to obtain a configured Service.
type installService struct {
	parser     *parser.Parser
	serializer *renderer.Serializer
}

// Compile-time confirmation that *installService satisfies the
// Service interface declared in this package.
var _ Service = (*installService)(nil)

// NewService returns the production wiring for resource installation.
// The parser and serializer are injected so callers can construct
// them once (mirroring the slice-7 dependency-injection precedent at
// the now-archived services package) and reuse them across calls;
// tests that need to substitute a fake parser or renderer inject them
// at construction time.
func NewService(p *parser.Parser, s *renderer.Serializer) Service {
	return &installService{
		parser:     p,
		serializer: s,
	}
}

// Initialize implements Service. Resolves each ref against the
// supplied library, derives its output path, fails fast on existing
// files unless --force or --dry-run, then runs the canonical
// load → render → write pipeline under the matching output directory.
//
// Per-ref errors are recorded in result.Error and the loop continues
// so the partial-success aggregate is consistent. The error return
// is reserved for transport-level failures; per-resource outcomes
// always live in result.Error, allowing callers to synthesize
// *core.PartialSuccessError.
func (i *installService) Initialize(ctx context.Context, req *Request) ([]core.InitializeResult, error) {
	if req == nil {
		return nil, core.NewValidationError("install", "request", "", "install request must not be nil")
	}
	if req.Library == nil || req.Library.RootPath == "" {
		return nil, core.NewValidationError("install", "library", "",
			"library is not loaded (RootPath is empty)")
	}

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

		doc, err := i.parser.LoadDocument(ctx, inputPath, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		rendered, err := i.serializer.RenderDocument(ctx, doc, req.Platform)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}

		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0o755); err != nil { //nolint:gosec // G301: user-owned output directory; 0755 is standard permission
			result.Error = core.NewFileError(outputPath, "mkdir", "failed to create output directory", err)
			results = append(results, result)
			continue
		}

		if err := os.WriteFile(outputPath, []byte(rendered), 0o644); err != nil { //nolint:gosec // G306: user-owned output file; 0644 is standard readable permission
			result.Error = core.NewFileError(outputPath, "write", "failed to write output file", err)
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}
