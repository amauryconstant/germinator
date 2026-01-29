// Package core provides document parsing and serialization functionality.
package core

import (
	"fmt"
	"regexp"

	"gitlab.com/amoconst/germinator/internal/models"
)

// LoadDocument loads and validates a document from the given filepath.
func LoadDocument(filepath, platform string) (interface{}, error) {
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
	case *models.Agent:
		errs = d.Validate(platform)
	case *models.Command:
		errs = d.Validate(platform)
	case *models.Memory:
		errs = d.Validate(platform)
	case *models.Skill:
		errs = d.Validate(platform)
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

	if matched, _ := regexp.MatchString(`command-.*\.md$`, base); matched {
		return "command"
	}
	if matched, _ := regexp.MatchString(`.*-command\.md$`, base); matched {
		return "command"
	}

	if matched, _ := regexp.MatchString(`memory-.*\.md$`, base); matched {
		return "memory"
	}
	if matched, _ := regexp.MatchString(`.*-memory\.md$`, base); matched {
		return "memory"
	}

	if matched, _ := regexp.MatchString(`skill-.*\.md$`, base); matched {
		return "skill"
	}
	if matched, _ := regexp.MatchString(`.*-skill\.md$`, base); matched {
		return "skill"
	}

	return ""
}
