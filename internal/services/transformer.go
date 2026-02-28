// Package services provides business logic for document transformation and validation.
package services

import (
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

const (
	PlatformClaudeCode = "claude-code"
	PlatformOpenCode   = "opencode"
)

// validatePlatform checks if platform parameter is valid.
func validatePlatform(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, gerrors.NewConfigError("platform", "", []string{PlatformClaudeCode, PlatformOpenCode}, "platform is required"))
		return errs
	}

	if platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, gerrors.NewConfigError("platform", platform, []string{PlatformClaudeCode, PlatformOpenCode}, "unknown platform"))
		return errs
	}

	return nil
}

// TransformDocument transforms a document to target platform format.
func TransformDocument(inputPath, outputPath, platform string) error {
	doc, err := core.LoadDocument(inputPath, platform)
	if err != nil {
		return err
	}

	rendered, err := core.RenderDocument(doc, platform)
	if err != nil {
		return gerrors.NewTransformError("render", platform, "failed to render document", err)
	}

	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return gerrors.NewFileError(outputPath, "write", "failed to write output file", err)
	}

	return nil
}

// ValidateDocument validates a document and returns any validation errors.
func ValidateDocument(inputPath, platform string) ([]error, error) {
	if errs := validatePlatform(platform); len(errs) > 0 {
		return errs, nil
	}

	docType := core.DetectType(inputPath)
	if docType == "" {
		return nil, gerrors.NewParseError(inputPath, "unrecognizable filename", nil)
	}

	doc, parseErr := core.ParseDocument(inputPath, docType)
	if parseErr != nil {
		return nil, gerrors.NewParseError(inputPath, "failed to parse document", parseErr)
	}

	var errs []error

	switch d := doc.(type) {
	case *core.CanonicalAgent:
		errs = d.Validate()
		if platform == PlatformOpenCode {
			errs = append(errs, validateOpenCodeAgent(*d)...)
		}
	case *core.CanonicalCommand:
		errs = d.Validate()
	case *core.CanonicalMemory:
		errs = d.Validate()
	case *core.CanonicalSkill:
		errs = d.Validate()
	default:
		return nil, gerrors.NewParseError(inputPath, "unknown document type", nil)
	}

	return errs, nil
}

// validateOpenCodeAgent performs OpenCode-specific validation on an agent.
// Note: Temperature and mode validation are already in AgentBehavior.Validate()
func validateOpenCodeAgent(agent core.CanonicalAgent) []error {
	// No OpenCode-specific validation needed beyond what's already in AgentBehavior.Validate()
	return nil
}
