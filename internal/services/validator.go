// Package services provides business logic for document transformation and validation.
package services

import (
	"context"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// validator implements the application.Validator interface.
type validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() application.Validator {
	return &validator{}
}

// Validate validates a document and returns any validation errors.
func (v *validator) Validate(ctx context.Context, req *application.ValidateRequest) (*application.ValidateResult, error) {
	if errs := validatePlatform(req.Platform); len(errs) > 0 {
		return &application.ValidateResult{Errors: errs}, nil
	}

	docType := core.DetectType(req.InputPath)
	if docType == "" {
		return nil, gerrors.NewParseError(req.InputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := core.ParseDocument(req.InputPath, docType)
	if parseErr != nil {
		return nil, gerrors.NewParseError(req.InputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *core.CanonicalAgent:
		errs = d.Validate()
		if req.Platform == PlatformOpenCode {
			errs = append(errs, validateOpenCodeAgent(*d)...)
		}
	case *core.CanonicalCommand:
		errs = d.Validate()
	case *core.CanonicalMemory:
		errs = d.Validate()
	case *core.CanonicalSkill:
		errs = d.Validate()
	default:
		return nil, gerrors.NewParseError(req.InputPath, "unknown document type", nil)
	}

	return &application.ValidateResult{Errors: errs}, nil
}

// validateOpenCodeAgent performs OpenCode-specific validation on an agent.
// Note: Temperature and mode validation are already in AgentBehavior.Validate()
func validateOpenCodeAgent(agent core.CanonicalAgent) []error {
	// No OpenCode-specific validation needed beyond what's already in AgentBehavior.Validate()
	return nil
}

// Compile-time interface satisfaction check.
var _ application.Validator = (*validator)(nil)
