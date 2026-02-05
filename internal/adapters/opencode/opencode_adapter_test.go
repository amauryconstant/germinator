package opencode

import (
	"testing"

	canonical "gitlab.com/amoconst/germinator/internal/models/canonical"
)

func TestConvertToolNameCase(t *testing.T) {
	adapter := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "bash", "bash"},
		{"already lowercase", "read", "read"},
		{"hyphenated", "write-to-file", "write-to-file"},
		{"multiple hyphens", "read-from-github", "read-from-github"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.ConvertToolNameCase(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToolNameCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPermissionPolicyToPlatform(t *testing.T) {
	adapter := New()

	tests := []struct {
		name        string
		policy      canonical.PermissionPolicy
		expectError bool
	}{
		{"restrictive", canonical.PermissionPolicyRestrictive, false},
		{"balanced", canonical.PermissionPolicyBalanced, false},
		{"permissive", canonical.PermissionPolicyPermissive, false},
		{"analysis", canonical.PermissionPolicyAnalysis, false},
		{"unrestricted", canonical.PermissionPolicyUnrestricted, false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.PermissionPolicyToPlatform(tt.policy)
			if tt.expectError {
				if err == nil {
					t.Error("PermissionPolicyToPlatform() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("PermissionPolicyToPlatform() unexpected error: %v", err)
				}
				if result == nil {
					t.Error("PermissionPolicyToPlatform() returned nil map")
				}
			}
		})
	}
}

func TestToCanonical(t *testing.T) {
	adapter := New()

	t.Run("agent with minimal fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "agent",
			"description": "Test agent",
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if agent.Description != "Test agent" {
			t.Errorf("agent.Description = %q, want %q", agent.Description, "Test agent")
		}
	})

	t.Run("agent with tools map", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "agent",
			"description": "Test agent",
			"tools": map[string]interface{}{
				"bash":          true,
				"write-to-file": true,
				"read":          false,
			},
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(agent.Tools) != 2 {
			t.Errorf("agent.Tools length = %d, want 2", len(agent.Tools))
		}
		if len(agent.DisallowedTools) != 1 {
			t.Errorf("agent.DisallowedTools length = %d, want 1", len(agent.DisallowedTools))
		}
	})

	t.Run("agent with permission object", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "agent",
			"description": "Test agent",
			"permission": map[string]interface{}{
				"edit": map[string]interface{}{"*": "allow"},
				"bash": map[string]interface{}{"*": "ask"},
			},
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if agent.PermissionPolicy != canonical.PermissionPolicyBalanced {
			t.Errorf("agent.PermissionPolicy = %q, want %q", agent.PermissionPolicy, canonical.PermissionPolicyBalanced)
		}
	})

	t.Run("agent with behavior", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "agent",
			"description": "Test agent",
			"mode":        "primary",
			"temperature": 0.5,
			"maxSteps":    10,
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if agent.Behavior.Mode != "primary" {
			t.Errorf("agent.Behavior.Mode = %q, want %q", agent.Behavior.Mode, "primary")
		}
		if agent.Behavior.Temperature == nil {
			t.Error("agent.Behavior.Temperature is nil, want 0.5")
		} else if *agent.Behavior.Temperature != 0.5 {
			t.Errorf("agent.Behavior.Temperature = %f, want 0.5", *agent.Behavior.Temperature)
		}
		if agent.Behavior.Steps != 10 {
			t.Errorf("agent.Behavior.Steps = %d, want 10", agent.Behavior.Steps)
		}
	})
}

func TestFromCanonical(t *testing.T) {
	adapter := New()

	t.Run("minimal agent", func(t *testing.T) {
		agent := &canonical.Agent{
			Description: "Test agent",
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["description"] != "Test agent" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test agent")
		}
	})

	t.Run("agent with tools arrays", func(t *testing.T) {
		agent := &canonical.Agent{
			Description:     "Test agent",
			Tools:           []string{"bash", "write-to-file"},
			DisallowedTools: []string{"read"},
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		tools, ok := output["tools"].(map[string]bool)
		if !ok {
			t.Fatal("output[tools] is not map[string]bool")
		}
		if tools["bash"] != true {
			t.Errorf("tools[bash] = %v, want true", tools["bash"])
		}
		if tools["write-to-file"] != true {
			t.Errorf("tools[write-to-file] = %v, want true", tools["write-to-file"])
		}
		if tools["read"] != false {
			t.Errorf("tools[read] = %v, want false", tools["read"])
		}
	})

	t.Run("agent with permission policy", func(t *testing.T) {
		agent := &canonical.Agent{
			Description:      "Test agent",
			PermissionPolicy: canonical.PermissionPolicyBalanced,
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		permission, ok := output["permission"]
		if !ok {
			t.Error("output[permission] not found")
		}
		if permission == nil {
			t.Error("output[permission] is nil")
		}
	})

	t.Run("agent with behavior", func(t *testing.T) {
		agent := &canonical.Agent{
			Description: "Test agent",
			Behavior: canonical.AgentBehavior{
				Mode:     "primary",
				Steps:    10,
				Hidden:   true,
				Disabled: false,
			},
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["mode"] != "primary" {
			t.Errorf("output[mode] = %q, want %q", output["mode"], "primary")
		}
		if output["maxSteps"] != 10 {
			t.Errorf("output[maxSteps] = %d, want 10", output["maxSteps"])
		}
		if output["hidden"] != true {
			t.Errorf("output[hidden] = %v, want true", output["hidden"])
		}
		if output["disable"] != nil {
			t.Errorf("output[disable] should be omitted, got %v", output["disable"])
		}
	})
}
