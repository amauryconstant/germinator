// Package transform provides document transformation as an I/O
// shell-package service. The package exists to satisfy the
// imperative-shell boundary defined by the
// golang-cli-architecture skill: any code that performs filesystem
// reads (parser.LoadDocument), platform rendering (renderer.RenderDocument),
// and filesystem writes (os.WriteFile) lives here at the package edge,
// not in cmd/.
//
// The Service interface, Request type, and NewService constructor are
// the canonical contract. cmd/adapt.go declares a local Transformer
// interface that is structurally identical to transform.Service (same
// parameter / return types); *transformService satisfies both via
// structural typing so cmd/ does not have to import this package's
// adapter directly.
//
// The package name `transform` mirrors the slice-7 convention that
// renamed the legacy shell package; the constructor takes the parser
// and serializer as injected dependencies so callers can construct
// them once and reuse across calls rather than rebuilding on every
// invocation.
package transform

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// Request carries the inputs for document transformation. Lifted from
// cmd/adapt.go so the shell-package owns its types per the
// golang-cli-architecture "interfaces where consumed" rule (the
// consumer, cmd/adapt.go, declares its Transformer interface using
// this type via the shared import).
type Request struct {
	InputPath  string
	OutputPath string
	Platform   string
}

// Service is the per-call contract for document transformation.
// Returns a *core.TransformResult on success; on failure returns a
// *core.FileError / *core.TransformError so cmd/cmdutil.ExitCodeFor
// maps them to exit 1 via errors.As dispatch.
type Service interface {
	Transform(ctx context.Context, req *Request) (*core.TransformResult, error)
}

// transformService is the production implementation. Holds the
// injected parser and serializer so the same Service can be reused
// across many Transform calls without rebuilding its orchestrating
// state. The struct fields are unexported so callers must use
// NewService to obtain a configured Service.
type transformService struct {
	parser     *parser.Parser
	serializer *renderer.Serializer
}

// Compile-time confirmation that *transformService satisfies the
// Service interface declared in this package.
var _ Service = (*transformService)(nil)

// NewService returns the production wiring for document
// transformation. The parser and serializer are injected so callers
// can construct them once (mirroring the slice-7 dependency-injection
// precedent at the now-archived services package) and reuse them
// across calls; tests that need to substitute a fake parser or
// renderer inject them at construction time.
func NewService(p *parser.Parser, s *renderer.Serializer) Service {
	return &transformService{
		parser:     p,
		serializer: s,
	}
}

// Transform implements Service. Composes parser.LoadDocument →
// renderer.RenderDocument → os.WriteFile as the canonical
// transform pipeline. Platform is assumed pre-validated by the
// caller (cmd/adapt.go's runAdapt validates via core.ValidatePlatform
// before resolving the Service).
//
// On error the chain is wrapped so the cmd layer can dispatch by
// type: load failures become a plain error (parser errors are already
// typed); render failures become *core.TransformError; write failures
// become *core.FileError. The single-handling rule in cmd/AGENTS.md
// keeps all error rendering centralized in main.go.
func (t *transformService) Transform(ctx context.Context, req *Request) (*core.TransformResult, error) {
	if req == nil {
		return nil, core.NewValidationError("transform", "request", "", "transform request must not be nil")
	}

	doc, err := t.parser.LoadDocument(ctx, req.InputPath, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("loading document: %w", err)
	}

	rendered, err := t.serializer.RenderDocument(ctx, doc, req.Platform)
	if err != nil {
		return nil, core.NewTransformError("render", req.Platform, "failed to render document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(rendered), 0o644); err != nil { //nolint:gosec // G306: user-owned output path; 0644 is standard readable permission
		return nil, core.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &core.TransformResult{OutputPath: req.OutputPath}, nil
}
