package adapters

import (
	"testing"
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
