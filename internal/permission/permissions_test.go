package permission

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/amoconst/germinator/internal/core"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "bash", "Bash"},
		{"hyphenated", "write-to-file", "WriteToFile"},
		{"multiple hyphens", "read-from-github", "ReadFromGithub"},
		{"already pascal", "Bash", "Bash"},
		{"multiple words", "read file from github", "ReadFileFromGithub"},
		{"spaces", "read from github", "ReadFromGithub"},
		{"empty", "", ""},
		{"single char", "a", "A"},
		{"underscores", "read_file", "ReadFile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToLowerCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "bash", "bash"},
		{"uppercase", "BASH", "bash"},
		{"mixed case", "WriteToFile", "writetofile"},
		{"already lowercase", "bash", "bash"},
		{"empty", "", ""},
		{"spaces", "Read File", "read file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToLowerCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToLowerCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPermissionPolicyMappings(t *testing.T) {
	tests := []struct {
		name            string
		policy          string
		hasClaudeCode   bool
		hasOpenCode     bool
		claudeCodeValue string
	}{
		{"restrictive", "restrictive", true, true, "default"},
		{"balanced", "balanced", true, true, "acceptEdits"},
		{"permissive", "permissive", true, true, "dontAsk"},
		{"analysis", "analysis", true, true, "plan"},
		{"unrestricted", "unrestricted", true, true, "bypassPermissions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, ok := PermissionPolicyMappings[tt.policy]
			if !ok {
				t.Errorf("PermissionPolicyMappings[%q] not found", tt.policy)
				return
			}

			if mapping.ClaudeCode != tt.claudeCodeValue {
				t.Errorf("ClaudeCode mapping = %q, want %q", mapping.ClaudeCode, tt.claudeCodeValue)
			}

			if mapping.OpenCode.Edit == "" && mapping.OpenCode.Bash == "" {
				t.Error("OpenCode mapping is empty")
			}
		})
	}
}

// TestValidateActionStrings is the unit-level test for the shared
// unknown-action validator. Adapter integration is covered in
// internal/opencode/opencode_adapter_test.go::TestParseAgent_UnknownPermissionActionReturnsError.
func TestValidateActionStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		perm     map[string]interface{}
		wantErr  bool
		errField string
		errValue string
	}{
		{
			name: "all known actions accepted",
			perm: map[string]interface{}{
				"edit": map[string]interface{}{"*": string(Allow)},
				"bash": map[string]interface{}{"*": string(Ask)},
				"read": map[string]interface{}{"*": string(Deny)},
			},
			wantErr: false,
		},
		{
			name: "empty map",
			perm: map[string]interface{}{},
		},
		{
			name: "nil map",
			perm: nil,
		},
		{
			name: "unknown action string returns *core.ConfigError",
			perm: map[string]interface{}{
				"edit": map[string]interface{}{"*": "denyUnlessRead"},
			},
			wantErr:  true,
			errField: "permission-action",
			errValue: "denyUnlessRead",
		},
		{
			name: "non-string action values are skipped silently",
			perm: map[string]interface{}{
				"edit": map[string]interface{}{"*": 42},
			},
			wantErr: false,
		},
		{
			name: "tool value that is not a map is skipped silently",
			perm: map[string]interface{}{
				"edit": string(Allow),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateActionStrings(tt.perm)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var cfgErr *core.ConfigError
			require.True(t, errors.As(err, &cfgErr),
				"error must unwrap to *core.ConfigError, got %T", err)
			assert.Equal(t, tt.errField, cfgErr.Field())
			assert.Equal(t, tt.errValue, cfgErr.Value())
			suggestions := cfgErr.Suggestions()
			assert.Contains(t, suggestions, string(Allow))
			assert.Contains(t, suggestions, string(Ask))
			assert.Contains(t, suggestions, string(Deny))
		})
	}
}
