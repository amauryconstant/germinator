package cmd

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// NewTransformer returns the production Transformer backed by the
// canonical parser/serializer pair. The cmd-side Transformer interface
// stays local (defined in cmd/adapt.go) — this constructor materializes
// the concrete adapter that satisfies it.
//
// Per the slice-3 design decision for validator/canonicalizer, the
// adapter is co-located with the command rather than in a new
// internal package; no service-locator indirection is involved.
func NewTransformer() Transformer {
	return &transformerAdapter{
		parser:     parser.NewParser(),
		serializer: renderer.NewSerializer(),
	}
}

// transformerAdapter implements the local Transformer interface. It
// composes parser.Parser (load document) and renderer.Serializer
// (render to platform format) in the canonical read → render → write
// pipeline. The local TransformRequest type replaces the previous
// cross-package request type alias.
type transformerAdapter struct {
	parser     *parser.Parser
	serializer *renderer.Serializer
}

// Compile-time confirmation that *transformerAdapter satisfies the
// cmd-side Transformer interface.
var _ Transformer = (*transformerAdapter)(nil)

// Transform loads the source document, renders it for the target
// platform, and writes the rendered bytes to the output path.
func (t *transformerAdapter) Transform(_ context.Context, req *TransformRequest) (*core.TransformResult, error) {
	doc, err := t.parser.LoadDocument(req.InputPath, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("loading document: %w", err)
	}

	rendered, err := t.serializer.RenderDocument(doc, req.Platform)
	if err != nil {
		return nil, core.NewTransformError("render", req.Platform, "failed to render document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(rendered), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
		return nil, core.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &core.TransformResult{OutputPath: req.OutputPath}, nil
}
