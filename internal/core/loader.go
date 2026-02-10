// Package core provides document parsing and serialization functionality.
package core

import (
	"fmt"
	"regexp"
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

// LoadDocument loads and validates a document from the given filepath.
func LoadDocument(filepath, platform string) (interface{}, error) {
	if errs := validatePlatform(platform); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errs)
	}

	docType := DetectType(filepath)
	if docType == "" {
		return nil, fmt.Errorf("unrecognizable filename: %s (expected: agent-*.md, *-agent.md, etc.)", filepath)
	}

	doc, err := ParseDocument(filepath, docType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	var errs []error
	switch d := doc.(type) {
	case *CanonicalAgent:
		errs = d.Validate()
	case *CanonicalCommand:
		errs = d.Validate()
	case *CanonicalMemory:
		errs = d.Validate()
	case *CanonicalSkill:
		errs = d.Validate()
	}

	if len(errs) > 0 {
		return doc, fmt.Errorf("validation failed: %v", errs)
	}

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
