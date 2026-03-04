// Package services provides document transformation and validation services.
package services

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// canonicalizer implements the application.Canonicalizer interface.
type canonicalizer struct{}

// NewCanonicalizer creates a new Canonicalizer instance.
func NewCanonicalizer() application.Canonicalizer {
	return &canonicalizer{}
}

// Canonicalize converts a platform document to canonical YAML format.
func (c *canonicalizer) Canonicalize(ctx context.Context, req *application.CanonicalizeRequest) (*application.CanonicalizeResult, error) {
	doc, err := core.ParsePlatformDocument(req.InputPath, req.Platform, req.DocType)
	if err != nil {
		return nil, gerrors.NewParseError(req.InputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return nil, gerrors.NewValidationError("", "", "", errs[0].Error())
	}

	yamlBytes, err := core.MarshalCanonical(doc)
	if err != nil {
		return nil, gerrors.NewTransformError("marshal", req.Platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(yamlBytes), 0644); err != nil {
		return nil, gerrors.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &application.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

// validateCanonicalDoc validates a canonical document and returns any validation errors.
func validateCanonicalDoc(doc interface{}) []error {
	switch d := doc.(type) {
	case *core.CanonicalAgent:
		return d.Validate()
	case *core.CanonicalCommand:
		return d.Validate()
	case *core.CanonicalSkill:
		return d.Validate()
	case *core.CanonicalMemory:
		return d.Validate()
	default:
		return []error{gerrors.NewParseError("", "unknown document type", nil)}
	}
}

// Compile-time interface satisfaction check.
var _ application.Canonicalizer = (*canonicalizer)(nil)
