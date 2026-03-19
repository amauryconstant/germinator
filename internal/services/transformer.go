// Package services provides business logic for document transformation and validation.
package services

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/domain"
)

const (
	PlatformClaudeCode = "claude-code"
	PlatformOpenCode   = "opencode"
)

// validatePlatform checks if platform parameter is valid.
func validatePlatform(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, domain.NewConfigError("platform", "", "platform is required").WithSuggestions([]string{PlatformClaudeCode, PlatformOpenCode}))
		return errs
	}

	if platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, domain.NewConfigError("platform", platform, "unknown platform").WithSuggestions([]string{PlatformClaudeCode, PlatformOpenCode}))
		return errs
	}

	return nil
}

// transformer implements the application.Transformer interface.
type transformer struct{}

// NewTransformer creates a new Transformer instance.
func NewTransformer() application.Transformer {
	return &transformer{}
}

// Transform transforms a document to target platform format.
func (t *transformer) Transform(ctx context.Context, req *application.TransformRequest) (*domain.TransformResult, error) {
	doc, err := core.LoadDocument(req.InputPath, req.Platform)
	if err != nil {
		return nil, err
	}

	rendered, err := core.RenderDocument(doc, req.Platform)
	if err != nil {
		return nil, domain.NewTransformError("render", req.Platform, "failed to render document", err)
	}

	if err := os.WriteFile(req.OutputPath, []byte(rendered), 0644); err != nil {
		return nil, domain.NewFileError(req.OutputPath, "write", "failed to write output file", err)
	}

	return &domain.TransformResult{OutputPath: req.OutputPath}, nil
}

// Compile-time interface satisfaction check.
var _ application.Transformer = (*transformer)(nil)
