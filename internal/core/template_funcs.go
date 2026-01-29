package core

import (
	"fmt"
	"strings"
)

// transformPermissionMode transforms Claude Code's permissionMode enum to OpenCode's permission object format.
//
// This function provides a basic approximation because the two permission systems are fundamentally different:
// - Claude Code uses an enum (default, acceptEdits, dontAsk, bypassPermissions, plan)
// - OpenCode uses nested objects with command keys (e.g., {"bash": {"*": "ask"}, "edit": {"*": "allow"}})
//
// The transformation preserves semantic intent:
//   - default: ask for both edit and bash tools
//   - acceptEdits: allow edit, ask for bash
//   - dontAsk: allow both (don't prompt user)
//   - bypassPermissions: allow both (override restrictions)
//   - plan: deny both (restricted mode for analysis)
//
// Parameters:
//   - mode: Claude Code permissionMode enum value
//
// Returns:
//   - string: YAML-formatted permission object structure with proper indentation
//   - empty string: for unknown modes
func transformPermissionMode(mode string) string {
	var perms map[string]map[string]string

	switch mode {
	case "default":
		perms = map[string]map[string]string{
			"edit": {"*": "ask"},
			"bash": {"*": "ask"},
		}
	case "acceptEdits":
		perms = map[string]map[string]string{
			"edit": {"*": "allow"},
			"bash": {"*": "ask"},
		}
	case "dontAsk":
		perms = map[string]map[string]string{
			"edit": {"*": "allow"},
			"bash": {"*": "allow"},
		}
	case "bypassPermissions":
		perms = map[string]map[string]string{
			"edit": {"*": "allow"},
			"bash": {"*": "allow"},
		}
	case "plan":
		perms = map[string]map[string]string{
			"edit": {"*": "deny"},
			"bash": {"*": "deny"},
		}
	default:
		return ""
	}

	var builder strings.Builder
	for tool, rules := range perms {
		for pattern, permission := range rules {
			builder.WriteString(fmt.Sprintf("  %s:\n", tool))
			builder.WriteString(fmt.Sprintf("    %s: %s\n", pattern, permission))
		}
	}
	return builder.String()
}
