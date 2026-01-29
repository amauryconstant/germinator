package core

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
//   - map[string]interface{}: OpenCode permission object structure
//   - nil: for unknown modes
func transformPermissionMode(mode string) map[string]interface{} {
	switch mode {
	case "default":
		return map[string]interface{}{
			"edit": map[string]string{"*": "ask"},
			"bash": map[string]string{"*": "ask"},
		}
	case "acceptEdits":
		return map[string]interface{}{
			"edit": map[string]string{"*": "allow"},
			"bash": map[string]string{"*": "ask"},
		}
	case "dontAsk":
		return map[string]interface{}{
			"edit": map[string]string{"*": "allow"},
			"bash": map[string]string{"*": "allow"},
		}
	case "bypassPermissions":
		return map[string]interface{}{
			"edit": map[string]string{"*": "allow"},
			"bash": map[string]string{"*": "allow"},
		}
	case "plan":
		return map[string]interface{}{
			"edit": map[string]string{"*": "deny"},
			"bash": map[string]string{"*": "deny"},
		}
	default:
		return nil
	}
}
