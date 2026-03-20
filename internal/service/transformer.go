// Package service provides business logic for document transformation and validation.
package service

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/parsing"
	"gitlab.com/amoconst/germinator/internal/infrastructure/serialization"
)

const (
	// PlatformClaudeCode identifies the Claude Code platform.
	PlatformClaudeCode = "claude-code"
	// PlatformOpenCode identifies the OpenCode platform.
	PlatformOpenCode = "opencode"
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
func (t *transformer) Transform(_ context.Context, req *application.TransformRequest) (*domain.TransformResult, error) {
	doc, err := parsing.LoadDocument(req.InputPath, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("loading document: %w", err)
	}

	rendered, err := serialization.RenderDocument(doc, req.Platform)
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
