// Package adapters provides platform-specific adapters for transforming between
// canonical document models and platform formats (Claude Code, OpenCode).
package adapters

import "gitlab.com/amoconst/germinator/internal/models/canonical"

// Adapter defines the interface for platform-specific document transformation.
// Implementations convert between canonical models and platform-specific formats.
type Adapter interface {
	// ToCanonical parses platform-specific input into canonical document structs.
	// Returns one of Agent, Command, Skill, or Memory based on detected document type.
	ToCanonical(input map[string]interface{}) (*canonical.Agent, *canonical.Command, *canonical.Skill, *canonical.Memory, error)

	// FromCanonical converts a canonical document struct to a platform-specific map
	// suitable for template rendering.
	FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)

	// PermissionPolicyToPlatform transforms a canonical PermissionPolicy enum
	// to the platform-specific representation (string for Claude Code, object for OpenCode).
	PermissionPolicyToPlatform(policy canonical.PermissionPolicy) (interface{}, error)

	// ConvertToolNameCase transforms tool names to the platform's expected casing:
	// PascalCase for Claude Code, lowercase for OpenCode.
	ConvertToolNameCase(name string) string
}
