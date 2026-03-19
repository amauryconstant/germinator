package claudecode

// Package claudecode implements the Claude Code adapter for bidirectional conversion
// between canonical models and Claude Code-specific format.
//
// The Claude Code adapter implements the Adapter interface (defined in internal/infrastructure/adapters/adapter.go)
// and handles all platform-specific logic for Claude Code:
//
//   - ToCanonical(): Parses Claude Code YAML files into canonical models
//   - FromCanonical(): Renders canonical models to Claude Code YAML format
//   - PermissionPolicyToPlatform(): Maps canonical PermissionPolicy to Claude Code's permissionMode enum
//   - ConvertToolNameCase(): Converts lowercase tool names to PascalCase (Claude Code convention)
//
// Permission Policy Mapping
//
// The adapter maps canonical PermissionPolicy enum values to Claude Code's permissionMode string enum:
//
//   restrictive    → "default"
//   balanced       → "acceptEdits"
//   permissive     → "dontAsk"
//   analysis       → "plan"
//   unrestricted   → "bypassPermissions"
//
// Tool Name Conventions
//
// Claude Code uses PascalCase for tool names (Bash, Read, Edit, etc.), while the canonical
// format uses lowercase. The ConvertToolNameCase() method handles bidirectional conversion:
//
//   - Canonical → Claude Code: lowercase ("bash") → PascalCase ("Bash")
//   - Claude Code → Canonical: PascalCase ("Bash") → lowercase ("bash")
//
// Claude Code-Specific Fields
//
// The adapter handles Claude Code-specific fields that are not part of the canonical format:
//
//   - skills: Array of skill references (rendered from targets.claude-code.skills)
//   - disableModelInvocation: Boolean flag to disable model invocation (from targets.claude-code)
//   - permissionMode: Claude Code's permission enum (mapped from canonical PermissionPolicy)
//
// Usage Example
//
//   import "gitlab.com/amoconst/germinator/internal/infrastructure/adapters/claude-code"
//
//   adapter := claudcode.New()
//
//   // Convert canonical to Claude Code
//   claudeCodeYAML, err := adapter.FromCanonical(canonicalAgent)
//
//   // Convert Claude Code to canonical
//   canonicalAgent, err := adapter.ToCanonical(claudeCodeMap)
//
//   // Map permission policy
//   permissionMode := adapter.PermissionPolicyToPlatform(canonical.PermissionPolicyBalanced)
//   // permissionMode == "acceptEdits"
