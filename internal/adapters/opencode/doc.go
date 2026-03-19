// Package opencode implements the OpenCode adapter for bidirectional conversion
// between canonical models and OpenCode-specific format.
package opencode

// The OpenCode adapter implements the Adapter interface (defined in internal/adapters/adapter.go)
// and handles all platform-specific logic for OpenCode:
//
//   - ToCanonical(): Parses OpenCode YAML files into canonical models
//   - FromCanonical(): Renders canonical models to OpenCode YAML format
//   - PermissionPolicyToPlatform(): Maps canonical PermissionPolicy to OpenCode permission object
//   - ConvertToolNameCase(): Maintains lowercase (identity operation for OpenCode)
//
// Permission Policy Mapping
//
// The adapter maps canonical PermissionPolicy enum values to OpenCode's permission object format:
//
//   restrictive    → {Edit: Ask, Bash: Ask, Read: Ask, ...}
//   balanced       → {Edit: Allow, Bash: Ask, Read: Allow, ...}
//   permissive     → {Edit: Allow, Bash: Allow, Read: Allow, ...}
//   analysis       → {Edit: Deny, Bash: Deny, Read: Allow, ...}
//   unrestricted   → {Edit: Allow, Bash: Allow, Read: Allow, ...}
//
// The permission object uses the PermissionAction enum (Allow, Ask, Deny) for type safety.
//
// Tool List Splitting
//
// OpenCode uses a boolean map for tool permissions ({tool: true|false}), while the canonical
// format uses separate tools and disallowedTools arrays. The adapter handles conversion:
//
//   - Canonical → OpenCode: tools array → {tool: true}, disallowedTools → {tool: false}
//   - OpenCode → Canonical: {tool: true} → tools array, {tool: false} → disallowedTools
//
// Tool Name Conventions
//
// OpenCode uses lowercase for tool names, matching the canonical format. The ConvertToolNameCase()
// method is an identity operation for OpenCode (lowercase → lowercase).
//
// Behavior Object Flattening
//
// OpenCode does not support nested behavior objects, so the adapter flattens canonical behavior
// fields to the top level:
//
//   - behavior.mode → mode
//   - behavior.temperature → temperature
//   - behavior.steps → maxSteps
//   - behavior.hidden → hidden
//   - behavior.prompt → prompt
//   - behavior.disabled → disable
//
// The targets.opencode section can override canonical behavior fields if needed.
//
// OpenCode-Specific Fields
//
// The adapter handles OpenCode-specific fields that differ from the canonical format:
//
//   - permission: Nested object with tool permissions (mapped from canonical PermissionPolicy)
//   - tools: Boolean map format (converted from canonical tools/disallowedTools arrays)
//   - mode: Flattened from behavior.mode
//   - disable: Flattened from behavior.disabled (name change)
//   - maxSteps: Flattened from behavior.steps (name change)
//
// Usage Example
//
//   import "gitlab.com/amoconst/germinator/internal/adapters/opencode"
//
//   adapter := opencode.New()
//
//   // Convert canonical to OpenCode
//   openCodeYAML, err := adapter.FromCanonical(canonicalAgent)
//
//   // Convert OpenCode to canonical
//   canonicalAgent, err := adapter.ToCanonical(openCodeMap)
//
//   // Map permission policy
//   permissionMap := adapter.PermissionPolicyToPlatform(canonical.PermissionPolicyBalanced)
//   // permissionMap == PermissionMap{Edit: Allow, Bash: Ask, Read: Allow, ...}
