package claudecode

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
		{"lowercase", "bash", "Bash"},
		{"hyphenated", "write-to-file", "WriteToFile"},
		{"multiple hyphens", "read-from-github", "ReadFromGithub"},
		{"already pascal", "Bash", "Bash"},
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
		expected    string
		expectError bool
	}{
		{"restrictive", canonical.PermissionPolicyRestrictive, "default", false},
		{"balanced", canonical.PermissionPolicyBalanced, "acceptEdits", false},
		{"permissive", canonical.PermissionPolicyPermissive, "dontAsk", false},
		{"analysis", canonical.PermissionPolicyAnalysis, "plan", false},
		{"unrestricted", canonical.PermissionPolicyUnrestricted, "bypassPermissions", false},
		{"invalid", "invalid", "", true},
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
				if result != tt.expected {
					t.Errorf("PermissionPolicyToPlatform() = %q, want %q", result, tt.expected)
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
			"name":        "test-agent",
			"description": "Test agent",
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if agent.Name != "test-agent" {
			t.Errorf("agent.Name = %q, want %q", agent.Name, "test-agent")
		}
		if agent.Description != "Test agent" {
			t.Errorf("agent.Description = %q, want %q", agent.Description, "Test agent")
		}
	})

	t.Run("agent with tools", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "agent",
			"name":        "test-agent",
			"description": "Test agent",
			"tools":       []string{"Bash", "WriteToFile"},
		}

		agent, _, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(agent.Tools) != 2 {
			t.Errorf("agent.Tools length = %d, want 2", len(agent.Tools))
		}
		if agent.Tools[0] != "bash" {
			t.Errorf("agent.Tools[0] = %q, want %q", agent.Tools[0], "bash")
		}
		if agent.Tools[1] != "writetofile" {
			t.Errorf("agent.Tools[1] = %q, want %q", agent.Tools[1], "writetofile")
		}
	})

	t.Run("agent with permission policy", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":         "agent",
			"name":           "test-agent",
			"description":    "Test agent",
			"permissionMode": "acceptEdits",
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
			"name":        "test-agent",
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
			Name:        "test-agent",
			Description: "Test agent",
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["name"] != "test-agent" {
			t.Errorf("output[name] = %q, want %q", output["name"], "test-agent")
		}
		if output["description"] != "Test agent" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test agent")
		}
	})

	t.Run("agent with tools", func(t *testing.T) {
		agent := &canonical.Agent{
			Name:        "test-agent",
			Description: "Test agent",
			Tools:       []string{"bash", "writetofile"},
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		tools, ok := output["tools"].([]string)
		if !ok {
			t.Fatal("output[tools] is not []string")
		}
		if len(tools) != 2 {
			t.Errorf("tools length = %d, want 2", len(tools))
		}
		if tools[0] != "Bash" {
			t.Errorf("tools[0] = %q, want %q", tools[0], "Bash")
		}
		if tools[1] != "Writetofile" {
			t.Errorf("tools[1] = %q, want %q", tools[1], "Writetofile")
		}
	})

	t.Run("agent with permission policy", func(t *testing.T) {
		agent := &canonical.Agent{
			Name:             "test-agent",
			Description:      "Test agent",
			PermissionPolicy: canonical.PermissionPolicyBalanced,
		}

		output, err := adapter.FromCanonical("agent", agent)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["permissionMode"] != "acceptEdits" {
			t.Errorf("output[permissionMode] = %q, want %q", output["permissionMode"], "acceptEdits")
		}
	})

	t.Run("agent with behavior", func(t *testing.T) {
		agent := &canonical.Agent{
			Name:        "test-agent",
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
