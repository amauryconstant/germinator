// Package permission provides permission mapping between canonical policies and platform-specific formats.
package permission

import "gitlab.com/amoconst/germinator/internal/core"

// Adapter defines the interface for platform-specific document transformation.
// Implementations convert between domain models and platform-specific formats.
type Adapter interface {
	ToCanonical(input map[string]interface{}) (*core.Agent, *core.Command, *core.Skill, *core.Memory, error)
	FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)
	PermissionPolicyToPlatform(policy core.PermissionPolicy) (interface{}, error)
	ConvertToolNameCase(name string) string
}
