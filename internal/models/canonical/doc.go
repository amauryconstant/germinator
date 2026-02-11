package canonical

// Package canonical defines domain-driven document models for AI coding assistant configuration.
//
// These models represent a platform-agnostic canonical format that expresses configuration
// intent independent of platform-specific details. The canonical format serves as the
// single source of truth for transformations to target platforms (Claude Code, OpenCode).
//
// Document Types
//
// The canonical format supports four document types:
//
//   - Agent: AI assistant agents with tool permissions, execution behavior, and model configuration
//   - Command: Custom tools/commands with execution context and argument hints
//   - Memory: Project context via file references and narrative content
//   - Skill: Reusable agent capabilities with metadata and hooks
//
// Key Design Principles
//
// Permission Policy Enum
//
//   The canonical format uses a domain-driven PermissionPolicy enum (restrictive, balanced,
//   permissive, analysis, unrestricted) to express security posture without referencing
//   platform-specific terminology. Platform adapters map these policies to platform-specific
//   values (e.g., restrictive → Claude Code's "default").
//
// Platform-Agnostic Field Names
//
//   Canonical models use intent-driven field names rather than platform-specific ones:
//     - permissionPolicy (not permissionMode or mode)
//     - behavior object (grouping mode, temperature, maxSteps, prompt, hidden, disabled)
//     - tools/disallowedTools arrays (split lists for clarity)
//     - targets section (platform-specific overrides)
//
// Targets Section
//
//   The targets section contains platform-specific configuration overrides:
//     - targets.claude-code: Claude Code-specific fields (skills array, disableModelInvocation)
//     - targets.opencode: OpenCode-specific overrides (can override behavior object fields)
//   This separation allows the canonical format to remain platform-agnostic while supporting
//   platform-specific configuration when needed.
//
// Simple Model Strings
//
//   Models are provided as full provider IDs (e.g., "anthropic/claude-sonnet-4-20250514")
//   without alias resolution. This gives users full control and avoids normalization complexity.
//   Platform adapters pass model strings to output as-is after any required case conversion.
//
// Validation
//
// Each canonical struct implements a Validate() method that returns []error for multiple
// validation issues. Validation focuses on:
//   - Required fields (name, description)
//   - Domain constraints (name regex patterns, permissionPolicy enum values)
//   - Structure validation (tool array format, behavior object fields)
//   Platform-specific validation is handled by adapters, not canonical models.
//
// Usage Example
//
//   import "gitlab.com/amoconst/germinator/internal/models/canonical"
//
//   agent := &canonical.Agent{
//       Name:            "code-reviewer",
//       Description:      "Reviews code for patterns and improvements",
//       PermissionPolicy:  canonical.PermissionPolicyBalanced,
//       Tools:           []string{"bash", "grep", "read", "edit"},
//       Behavior: canonical.AgentBehavior{
//           Mode:        "primary",
//           Temperature: pointerTo(0.3),
//           MaxSteps:    25,
//       },
//   }
//
//   if errs := agent.Validate(); len(errs) > 0 {
//       // Handle validation errors
//   }
