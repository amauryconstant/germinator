package domain

import (
	"testing"
)

func TestPermissionPolicyIsValid(t *testing.T) {
	tests := []struct {
		name     string
		policy   PermissionPolicy
		expected bool
	}{
		{"restrictive is valid", PermissionPolicyRestrictive, true},
		{"balanced is valid", PermissionPolicyBalanced, true},
		{"permissive is valid", PermissionPolicyPermissive, true},
		{"analysis is valid", PermissionPolicyAnalysis, true},
		{"unrestricted is valid", PermissionPolicyUnrestricted, true},
		{"invalid policy is not valid", "invalid-policy", false},
		{"empty string is not valid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.IsValid()
			if got != tt.expected {
				t.Errorf("PermissionPolicy.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
