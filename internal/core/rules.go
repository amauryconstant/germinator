package core

import (
	"fmt"
	"strings"
)

// Platform identifiers recognized by ValidatePlatform and
// ResolveOutputPath. Any caller needing the canonical strings
// should reference these constants.
const (
	PlatformClaudeCode = "claude-code"
	PlatformOpenCode   = "opencode"
)

// ValidatePlatform returns nil if s is a recognized platform
// identifier, otherwise a *ValidationError describing the invalid value.
func ValidatePlatform(s string) error {
	switch s {
	case PlatformClaudeCode, PlatformOpenCode:
		return nil
	default:
		return NewValidationError(
			"platform",
			"platform",
			s,
			fmt.Sprintf("unknown platform %q", s),
		).WithSuggestions([]string{
			fmt.Sprintf("use %q", PlatformClaudeCode),
			fmt.Sprintf("use %q", PlatformOpenCode),
		})
	}
}

// ResolveOutputPath combines the document type, name, and platform
// into the canonical output filename. Examples:
//
//	ResolveOutputPath("skill", "commit", "claude-code") -> ".claude/skills/commit/SKILL.md"
//	ResolveOutputPath("agent", "reviewer", "opencode")  -> ".opencode/agents/reviewer.md"
//	ResolveOutputPath("command", "build", "claude-code") -> ".claude/commands/build.md"
//	ResolveOutputPath("memory", "context", "opencode") -> ".opencode/memory/context.md"
func ResolveOutputPath(docType, name, platform string) string {
	root := ".opencode"
	if platform == PlatformClaudeCode {
		root = ".claude"
	}
	switch docType {
	case "skill":
		return fmt.Sprintf("%s/skills/%s/SKILL.md", root, name)
	case "agent":
		return fmt.Sprintf("%s/agents/%s.md", root, name)
	case "command":
		return fmt.Sprintf("%s/commands/%s.md", root, name)
	case "memory":
		return fmt.Sprintf("%s/memory/%s.md", root, name)
	default:
		return strings.Join([]string{root, docType, name + ".md"}, "/")
	}
}
