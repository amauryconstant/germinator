// Package parser provides document parsing and loading functionality.
package parser

import (
	"context"
	"errors"
	"fmt"
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
// The ctx parameter is checked at entry and forwarded to DetectType and
// ParseDocument so caller cancellation propagates through the load.
func LoadDocument(ctx context.Context, filepath, platform string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("parser: load cancelled: %w", err)
	}

	if errs := validatePlatform(platform); len(errs) > 0 {
		return nil, errs[0]
	}

	docType := DetectType(ctx, filepath)
	if docType == "" {
		return nil, core.NewParseError(filepath, "unrecognizable filename (expected: agent-*.md, *-agent.md, etc.)", nil)
	}

	doc, err := ParseDocument(ctx, filepath, docType)
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
// Forwards ctx to the package-level LoadDocument so caller cancellation
// propagates through detection and parsing.
func (p *Parser) LoadDocument(ctx context.Context, path string, platform string) (interface{}, error) {
	return LoadDocument(ctx, path, platform)
}

// DetectType detects the document type from the filename. The ctx parameter
// is checked between regex iterations so a cancelled caller terminates the
// scan promptly. Detection is regex-only (no I/O); ctx is accept-and-may-ignore
// for spec symmetry with the cli-framework I/O-adapter ctx-propagation contract.
func DetectType(ctx context.Context, filepath string) string {
	for _, p := range detectTypePatterns() {
		if err := ctx.Err(); err != nil {
			return ""
		}
		if matched, _ := regexp.MatchString(p.pattern, filepath); matched {
			return p.docType
		}
	}
	return ""
}

// detectTypePattern pairs a filename regex with the document type it signals.
// Defined as a package-level value (not a literal in DetectType) so the
// function's cognitive complexity stays low and the table is easy to audit.
type detectTypePattern struct {
	pattern string
	docType string
}

func detectTypePatterns() []detectTypePattern {
	return []detectTypePattern{
		{`agent-.*\.md$`, "agent"},
		{`.*-agent\.md$`, "agent"},
		{`agent-.*\.yaml$`, "agent"},
		{`.*-agent\.yaml$`, "agent"},
		{`command-.*\.md$`, "command"},
		{`.*-command\.md$`, "command"},
		{`command-.*\.yaml$`, "command"},
		{`.*-command\.yaml$`, "command"},
		{`memory-.*\.md$`, "memory"},
		{`.*-memory\.md$`, "memory"},
		{`memory-.*\.yaml$`, "memory"},
		{`.*-memory\.yaml$`, "memory"},
		{`skill-.*\.md$`, "skill"},
		{`.*-skill\.md$`, "skill"},
		{`skill-.*\.yaml$`, "skill"},
		{`.*-skill\.yaml$`, "skill"},
	}
}
