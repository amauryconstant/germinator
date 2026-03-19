// Package services provides business logic for document transformation and validation.
package services

import (
	"context"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/domain/opencode"
)

// validator implements the application.Validator interface.
type validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() application.Validator {
	return &validator{}
}

// Validate validates a document and returns any validation errors.
func (v *validator) Validate(_ context.Context, req *application.ValidateRequest) (*domain.ValidateResult, error) {
	if errs := validatePlatform(req.Platform); len(errs) > 0 {
		return &domain.ValidateResult{Errors: errs}, nil
	}

	docType := core.DetectType(req.InputPath)
	if docType == "" {
		return nil, domain.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := core.ParseDocument(req.InputPath, docType)
	if parseErr != nil {
		return nil, domain.NewParseError(req.InputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *core.CanonicalAgent:
		if result := domain.ValidateAgent(&d.Agent); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateAgentOpenCode(&d.Agent); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *core.CanonicalCommand:
		if result := domain.ValidateCommand(&d.Command); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateCommandOpenCode(&d.Command); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *core.CanonicalMemory:
		if result := domain.ValidateMemory(&d.Memory); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
	case *core.CanonicalSkill:
		if result := domain.ValidateSkill(&d.Skill); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateSkillOpenCode(&d.Skill); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	default:
		return nil, domain.NewParseError(req.InputPath, "unknown document type", nil)
	}

	return &domain.ValidateResult{Errors: errs}, nil
}

// unwrapErrors unwraps a joined error into individual errors.
// If the error is not a joined error, returns a slice with just that error.
func unwrapErrors(err error) []error {
	if err == nil {
		return nil
	}

	// Try to unwrap as joined error using the Unwrap() []error interface
	type multipleUnwrapper interface {
		Unwrap() []error
	}

	if unwrapper, ok := err.(multipleUnwrapper); ok {
		return unwrapper.Unwrap()
	}

	// Not a joined error, return as single-element slice
	return []error{err}
}

// Compile-time interface satisfaction check.
var _ application.Validator = (*validator)(nil)
