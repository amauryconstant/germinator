// Package core provides document parsing and serialization functionality.
package core

import (
	"errors"
	"regexp"

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
		errs = append(errs, gerrors.NewConfigError("platform", "", "platform is required").WithSuggestions([]string{PlatformClaudeCode, PlatformOpenCode}))
		return errs
	}

	if platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, gerrors.NewConfigError("platform", platform, "unknown platform").WithSuggestions([]string{PlatformClaudeCode, PlatformOpenCode}))
		return errs
	}

	return nil
}

// LoadDocument loads and validates a document from the given filepath.
func LoadDocument(filepath, platform string) (interface{}, error) {
	if errs := validatePlatform(platform); len(errs) > 0 {
		return nil, errs[0]
	}

	docType := DetectType(filepath)
	if docType == "" {
		return nil, gerrors.NewParseError(filepath, "unrecognizable filename (expected: agent-*.md, *-agent.md, etc.)", nil)
	}

	doc, err := ParseDocument(filepath, docType)
	if err != nil {
		var fileErr *gerrors.FileError
		if errors.As(err, &fileErr) {
			return nil, err
		}
		return nil, gerrors.NewParseError(filepath, "failed to parse document", err)
	}

	// Validation is now handled by the validation package in services layer
	// No need to call Validate() here anymore
	return doc, nil
}

// DetectType detects the document type from the filename.
func DetectType(filepath string) string {
	base := filepath

	if matched, _ := regexp.MatchString(`agent-.*\.md$`, base); matched {
		return "agent"
	}
	if matched, _ := regexp.MatchString(`.*-agent\.md$`, base); matched {
		return "agent"
	}
	if matched, _ := regexp.MatchString(`agent-.*\.yaml$`, base); matched {
		return "agent"
	}
	if matched, _ := regexp.MatchString(`.*-agent\.yaml$`, base); matched {
		return "agent"
	}

	if matched, _ := regexp.MatchString(`command-.*\.md$`, base); matched {
		return "command"
	}
	if matched, _ := regexp.MatchString(`.*-command\.md$`, base); matched {
		return "command"
	}
	if matched, _ := regexp.MatchString(`command-.*\.yaml$`, base); matched {
		return "command"
	}
	if matched, _ := regexp.MatchString(`.*-command\.yaml$`, base); matched {
		return "command"
	}

	if matched, _ := regexp.MatchString(`memory-.*\.md$`, base); matched {
		return "memory"
	}
	if matched, _ := regexp.MatchString(`.*-memory\.md$`, base); matched {
		return "memory"
	}
	if matched, _ := regexp.MatchString(`memory-.*\.yaml$`, base); matched {
		return "memory"
	}
	if matched, _ := regexp.MatchString(`.*-memory\.yaml$`, base); matched {
		return "memory"
	}

	if matched, _ := regexp.MatchString(`skill-.*\.md$`, base); matched {
		return "skill"
	}
	if matched, _ := regexp.MatchString(`.*-skill\.md$`, base); matched {
		return "skill"
	}
	if matched, _ := regexp.MatchString(`skill-.*\.yaml$`, base); matched {
		return "skill"
	}
	if matched, _ := regexp.MatchString(`.*-skill\.yaml$`, base); matched {
		return "skill"
	}

	return ""
}
