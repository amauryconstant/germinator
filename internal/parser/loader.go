// Package parser provides document parsing and loading functionality.
package parser

import (
	"errors"
	"regexp"

	"gitlab.com/amoconst/germinator/internal/core"
)

// validatePlatform checks if platform parameter is valid.
func validatePlatform(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, core.NewConfigError("platform", "", "platform is required").WithSuggestions([]string{core.PlatformClaudeCode, core.PlatformOpenCode}))
		return errs
	}

	if platform != core.PlatformClaudeCode && platform != core.PlatformOpenCode {
		errs = append(errs, core.NewConfigError("platform", platform, "unknown platform").WithSuggestions([]string{core.PlatformClaudeCode, core.PlatformOpenCode}))
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
		return nil, core.NewParseError(filepath, "unrecognizable filename (expected: agent-*.md, *-agent.md, etc.)", nil)
	}

	doc, err := ParseDocument(filepath, docType)
	if err != nil {
		var fileErr *core.FileError
		if errors.As(err, &fileErr) {
			return nil, err
		}
		return nil, core.NewParseError(filepath, "failed to parse document", err)
	}

	// Validation is now handled by the validation package in services layer
	// No need to call Validate() here anymore
	return doc, nil
}

// Parser is the concrete parser type. NewParser returns *Parser; the
// (*Parser).LoadDocument method delegates to the package-level LoadDocument
// function so callers can choose between functional and method-style usage.
type Parser struct{}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// LoadDocument loads and parses a document from the given path.
func (p *Parser) LoadDocument(path string, platform string) (interface{}, error) {
	return LoadDocument(path, platform)
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
