package domain

// PermissionPolicy defines canonical tool permission policies.
type PermissionPolicy string

const (
	// PermissionPolicyRestrictive requires explicit approval for all tool use.
	PermissionPolicyRestrictive PermissionPolicy = "restrictive"
	// PermissionPolicyBalanced requires approval for sensitive tools only.
	PermissionPolicyBalanced PermissionPolicy = "balanced"
	// PermissionPolicyPermissive allows most tools without approval.
	PermissionPolicyPermissive PermissionPolicy = "permissive"
	// PermissionPolicyAnalysis allows read-only tools without approval.
	PermissionPolicyAnalysis PermissionPolicy = "analysis"
	// PermissionPolicyUnrestricted allows all tools without approval.
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
