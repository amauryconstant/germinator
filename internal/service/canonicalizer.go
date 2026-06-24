// Package service provides document transformation and validation services.
package service

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
)

// canonicalizer implements the application.Canonicalizer interface.
type canonicalizer struct{}

// NewCanonicalizer creates a new Canonicalizer instance.
func NewCanonicalizer() application.Canonicalizer {
	return &canonicalizer{}
}

// Canonicalize converts a platform document to canonical YAML format.
func (c *canonicalizer) Canonicalize(_ context.Context, req *application.CanonicalizeRequest) (*core.CanonicalizeResult, error) {
	doc, err := parser.ParsePlatformDocument(req.InputPath, req.Platform, req.DocType)
	if err != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return nil, core.NewValidationError("", "", "", errs[0].Error())
	}

	yamlBytes, err := renderer.MarshalCanonical(doc)
	if err != nil {
		return nil, core.NewTransformError("marshal", req.Platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(yamlBytes), 0644); err != nil { //nolint:gosec // G306: User owns output file, 0644 is standard readable permission
		return nil, core.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &core.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

// validateCanonicalDoc validates a canonical document and returns any validation errors.
func validateCanonicalDoc(doc interface{}) []error {
	switch d := doc.(type) {
	case *parser.CanonicalAgent:
		if result := core.ValidateAgent(&d.Agent); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parser.CanonicalCommand:
		if result := core.ValidateCommand(&d.Command); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parser.CanonicalSkill:
		if result := core.ValidateSkill(&d.Skill); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parser.CanonicalMemory:
		if result := core.ValidateMemory(&d.Memory); result.IsError() {
			return unwrapErrors(result.Error)
		}
	default:
		return []error{core.NewParseError("", "unknown document type", nil)}
	}
	return nil
}

// Compile-time interface satisfaction check.
var _ application.Canonicalizer = (*canonicalizer)(nil)
