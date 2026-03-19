// Package services provides document transformation and validation services.
package services

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/internal/infrastructure/serialization"
)

// canonicalizer implements the application.Canonicalizer interface.
type canonicalizer struct{}

// NewCanonicalizer creates a new Canonicalizer instance.
func NewCanonicalizer() application.Canonicalizer {
	return &canonicalizer{}
}

// Canonicalize converts a platform document to canonical YAML format.
func (c *canonicalizer) Canonicalize(_ context.Context, req *application.CanonicalizeRequest) (*domain.CanonicalizeResult, error) {
	doc, err := parsing.ParsePlatformDocument(req.InputPath, req.Platform, req.DocType)
	if err != nil {
		return nil, domain.NewParseError(req.InputPath, "failed to parse platform document", err)
	}

	if errs := validateCanonicalDoc(doc); len(errs) > 0 {
		return nil, domain.NewValidationError("", "", "", errs[0].Error())
	}

	yamlBytes, err := serialization.MarshalCanonical(doc)
	if err != nil {
		return nil, domain.NewTransformError("marshal", req.Platform, "failed to marshal canonical document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(yamlBytes), 0644); err != nil {
		return nil, domain.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &domain.CanonicalizeResult{OutputPath: req.OutputPath}, nil
}

// validateCanonicalDoc validates a canonical document and returns any validation errors.
func validateCanonicalDoc(doc interface{}) []error {
	switch d := doc.(type) {
	case *parsing.CanonicalAgent:
		if result := domain.ValidateAgent(&d.Agent); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parsing.CanonicalCommand:
		if result := domain.ValidateCommand(&d.Command); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parsing.CanonicalSkill:
		if result := domain.ValidateSkill(&d.Skill); result.IsError() {
			return unwrapErrors(result.Error)
		}
	case *parsing.CanonicalMemory:
		if result := domain.ValidateMemory(&d.Memory); result.IsError() {
			return unwrapErrors(result.Error)
		}
	default:
		return []error{domain.NewParseError("", "unknown document type", nil)}
	}
	return nil
}

// Compile-time interface satisfaction check.
var _ application.Canonicalizer = (*canonicalizer)(nil)
