// Package service provides business logic for document transformation and validation.
package service

import (
	"context"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/core/opencode"
	"gitlab.com/amoconst/germinator/internal/parser"
)

// validator implements the application.Validator interface.
type validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() application.Validator {
	return &validator{}
}

// Validate validates a document and returns any validation errors.
func (v *validator) Validate(_ context.Context, req *application.ValidateRequest) (*core.ValidateResult, error) {
	if errs := validatePlatform(req.Platform); len(errs) > 0 {
		return &core.ValidateResult{Errors: errs}, nil
	}

	docType := parser.DetectType(req.InputPath)
	if docType == "" {
		return nil, core.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := parser.ParseDocument(req.InputPath, docType)
	if parseErr != nil {
		return nil, core.NewParseError(req.InputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *parser.CanonicalAgent:
		if result := core.ValidateAgent(&d.Agent); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateAgentOpenCode(&d.Agent); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *parser.CanonicalCommand:
		if result := core.ValidateCommand(&d.Command); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateCommandOpenCode(&d.Command); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	case *parser.CanonicalMemory:
		if result := core.ValidateMemory(&d.Memory); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
	case *parser.CanonicalSkill:
		if result := core.ValidateSkill(&d.Skill); result.IsError() {
			errs = append(errs, unwrapErrors(result.Error)...)
		}
		if req.Platform == PlatformOpenCode {
			if result := opencode.ValidateSkillOpenCode(&d.Skill); result.IsError() {
				errs = append(errs, unwrapErrors(result.Error)...)
			}
		}
	default:
		return nil, core.NewParseError(req.InputPath, "unknown document type", nil)
	}

	return &core.ValidateResult{Errors: errs}, nil
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
