package adapters

import "gitlab.com/amoconst/germinator/internal/models/canonical"

type Adapter interface {
	ToCanonical(input map[string]interface{}) (*canonical.Agent, *canonical.Command, *canonical.Skill, *canonical.Memory, error)
	FromCanonical(docType string, doc interface{}) (map[string]interface{}, error)
	PermissionPolicyToPlatform(policy canonical.PermissionPolicy) (interface{}, error)
	ConvertToolNameCase(name string) string
}
