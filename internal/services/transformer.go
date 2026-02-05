// Package services provides business logic for document transformation and validation.
package services

import (
	"fmt"
	"os"

	"gitlab.com/amoconst/germinator/internal/core"
)

const (
	PlatformClaudeCode = "claude-code"
	PlatformOpenCode   = "opencode"
)

// validatePlatform checks if platform parameter is valid.
func validatePlatform(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, fmt.Errorf("platform is required (available: %s, %s)", PlatformClaudeCode, PlatformOpenCode))
		return errs
	}

	if platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: %s, %s)", platform, PlatformClaudeCode, PlatformOpenCode))
		return errs
	}

	return nil
}

// TransformDocument transforms a document to target platform format.
func TransformDocument(inputPath, outputPath, platform string) error {
	doc, err := core.LoadDocument(inputPath, platform)
	if err != nil {
		return fmt.Errorf("failed to load document: %w", err)
	}

	rendered, err := core.RenderDocument(doc, platform)
	if err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
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
		return nil, fmt.Errorf("unrecognizable filename: %s", inputPath)
	}

	doc, parseErr := core.ParseDocument(inputPath, docType)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse document: %w", parseErr)
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
		return nil, fmt.Errorf("unknown document type: %T", d)
	}

	return errs, nil
}

// validateOpenCodeAgent performs OpenCode-specific validation on an agent.
// Note: Temperature and mode validation are already in AgentBehavior.Validate()
func validateOpenCodeAgent(agent core.CanonicalAgent) []error {
	// No OpenCode-specific validation needed beyond what's already in AgentBehavior.Validate()
	return nil
}
