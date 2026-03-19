package domain

// PermissionPolicy defines canonical tool permission policies.
type PermissionPolicy string

const (
	PermissionPolicyRestrictive  PermissionPolicy = "restrictive"
	PermissionPolicyBalanced     PermissionPolicy = "balanced"
	PermissionPolicyPermissive   PermissionPolicy = "permissive"
	PermissionPolicyAnalysis     PermissionPolicy = "analysis"
	PermissionPolicyUnrestricted PermissionPolicy = "unrestricted"
)

// IsValid returns true if the permission policy is a valid enum value.
func (p PermissionPolicy) IsValid() bool {
	switch p {
	case PermissionPolicyRestrictive, PermissionPolicyBalanced, PermissionPolicyPermissive,
		PermissionPolicyAnalysis, PermissionPolicyUnrestricted:
		return true
	default:
		return false
	}
}

// PlatformConfig maps platform names to platform-specific configuration.
type PlatformConfig map[string]map[string]interface{}
