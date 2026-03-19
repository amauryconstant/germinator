// Package adapters provides platform-specific adapters for transforming between
// canonical document models and platform formats (Claude Code, OpenCode).
package adapters

import "gitlab.com/amoconst/germinator/internal/domain"

// Adapter defines the interface for platform-specific document transformation.
// Implementations convert between domain models and platform-specific formats.
type Adapter interface {
	// ToCanonical parses platform-specific input into domain document structs.
	// Returns one of Agent, Command, Skill, or Memory based on detected document type.
	ToCanonical(input map[string]interface{}) (*domain.Agent, *domain.Command, *domain.Skill, *domain.Memory, error)

	// FromCanonical converts a domain document struct to a platform-specific map
	// suitable for template rendering.
	FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)

	// PermissionPolicyToPlatform transforms a domain PermissionPolicy enum
	// to platform-specific representation (string for Claude Code, object for OpenCode).
	PermissionPolicyToPlatform(policy domain.PermissionPolicy) (interface{}, error)

	// ConvertToolNameCase transforms tool names to platform's expected casing:
	// PascalCase for Claude Code, lowercase for OpenCode.
	ConvertToolNameCase(name string) string
}
