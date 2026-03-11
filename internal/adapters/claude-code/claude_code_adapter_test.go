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

func TestToCanonicalCommand(t *testing.T) {
	adapter := New()

	t.Run("command with minimal fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "command",
			"name":        "test-command",
			"description": "Test command",
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if cmd.Name != "test-command" {
			t.Errorf("cmd.Name = %q, want %q", cmd.Name, "test-command")
		}
		if cmd.Description != "Test command" {
			t.Errorf("cmd.Description = %q, want %q", cmd.Description, "Test command")
		}
	})

	t.Run("command with all fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "command",
			"name":        "test-command",
			"description": "Test command",
			"tools":       []interface{}{"Bash", "Edit"},
			"execution": map[string]interface{}{
				"context": "fork",
				"subtask": true,
				"agent":   "primary-agent",
			},
			"arguments": map[string]interface{}{
				"hint": "Enter your query",
			},
			"model": "claude-3-opus",
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if cmd.Name != "test-command" {
			t.Errorf("cmd.Name = %q, want %q", cmd.Name, "test-command")
		}
		if len(cmd.Tools) != 2 {
			t.Errorf("cmd.Tools length = %d, want 2", len(cmd.Tools))
		}
		if cmd.Tools[0] != "bash" {
			t.Errorf("cmd.Tools[0] = %q, want %q", cmd.Tools[0], "bash")
		}
		if cmd.Execution.Context != "fork" {
			t.Errorf("cmd.Execution.Context = %q, want %q", cmd.Execution.Context, "fork")
		}
		if cmd.Execution.Subtask != true {
			t.Errorf("cmd.Execution.Subtask = %v, want true", cmd.Execution.Subtask)
		}
		if cmd.Execution.Agent != "primary-agent" {
			t.Errorf("cmd.Execution.Agent = %q, want %q", cmd.Execution.Agent, "primary-agent")
		}
		if cmd.Arguments.Hint != "Enter your query" {
			t.Errorf("cmd.Arguments.Hint = %q, want %q", cmd.Arguments.Hint, "Enter your query")
		}
		if cmd.Model != "claude-3-opus" {
			t.Errorf("cmd.Model = %q, want %q", cmd.Model, "claude-3-opus")
		}
	})

	t.Run("command with empty fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "command",
			"name":        "",
			"description": "",
			"tools":       []interface{}{},
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if cmd.Name != "" {
			t.Errorf("cmd.Name = %q, want empty", cmd.Name)
		}
		if len(cmd.Tools) != 0 {
			t.Errorf("cmd.Tools length = %d, want 0", len(cmd.Tools))
		}
	})
}

func TestToCanonicalSkill(t *testing.T) {
	adapter := New()

	t.Run("skill with minimal fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "skill",
			"name":        "test-skill",
			"description": "Test skill",
		}

		_, _, skill, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if skill.Name != "test-skill" {
			t.Errorf("skill.Name = %q, want %q", skill.Name, "test-skill")
		}
		if skill.Description != "Test skill" {
			t.Errorf("skill.Description = %q, want %q", skill.Description, "Test skill")
		}
	})

	t.Run("skill with all fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "skill",
			"name":        "test-skill",
			"description": "Test skill",
			"tools":       []interface{}{"Bash", "Read"},
			"extensions": map[string]interface{}{
				"license":       "MIT",
				"compatibility": []interface{}{"claude-code", "opencode"},
				"metadata": map[string]interface{}{
					"author":  "test-author",
					"version": "1.0.0",
				},
				"hooks": map[string]interface{}{
					"preInvoke":  "setup.sh",
					"postInvoke": "cleanup.sh",
				},
			},
			"execution": map[string]interface{}{
				"context":       "fork",
				"agent":         "skill-agent",
				"userInvocable": true,
			},
			"model": "claude-3-sonnet",
		}

		_, _, skill, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if skill.Name != "test-skill" {
			t.Errorf("skill.Name = %q, want %q", skill.Name, "test-skill")
		}
		if len(skill.Tools) != 2 {
			t.Errorf("skill.Tools length = %d, want 2", len(skill.Tools))
		}
		if skill.Tools[0] != "bash" {
			t.Errorf("skill.Tools[0] = %q, want %q", skill.Tools[0], "bash")
		}
		if skill.Extensions.License != "MIT" {
			t.Errorf("skill.Extensions.License = %q, want %q", skill.Extensions.License, "MIT")
		}
		if len(skill.Extensions.Compatibility) != 2 {
			t.Errorf("skill.Extensions.Compatibility length = %d, want 2", len(skill.Extensions.Compatibility))
		}
		if skill.Extensions.Metadata["author"] != "test-author" {
			t.Errorf("skill.Extensions.Metadata[author] = %q, want %q", skill.Extensions.Metadata["author"], "test-author")
		}
		if skill.Extensions.Hooks["preInvoke"] != "setup.sh" {
			t.Errorf("skill.Extensions.Hooks[preInvoke] = %q, want %q", skill.Extensions.Hooks["preInvoke"], "setup.sh")
		}
		if skill.Execution.Context != "fork" {
			t.Errorf("skill.Execution.Context = %q, want %q", skill.Execution.Context, "fork")
		}
		if skill.Execution.Agent != "skill-agent" {
			t.Errorf("skill.Execution.Agent = %q, want %q", skill.Execution.Agent, "skill-agent")
		}
		if skill.Execution.UserInvocable != true {
			t.Errorf("skill.Execution.UserInvocable = %v, want true", skill.Execution.UserInvocable)
		}
		if skill.Model != "claude-3-sonnet" {
			t.Errorf("skill.Model = %q, want %q", skill.Model, "claude-3-sonnet")
		}
	})

	t.Run("skill with nil extensions", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "skill",
			"name":        "test-skill",
			"description": "Test skill",
			"extensions":  nil,
		}

		_, _, skill, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if skill.Name != "test-skill" {
			t.Errorf("skill.Name = %q, want %q", skill.Name, "test-skill")
		}
	})
}

func TestToCanonicalMemory(t *testing.T) {
	adapter := New()

	t.Run("memory with paths only", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "memory",
			"paths":  []interface{}{"./docs/guide.md", "./README.md"},
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 2 {
			t.Errorf("mem.Paths length = %d, want 2", len(mem.Paths))
		}
		if mem.Paths[0] != "./docs/guide.md" {
			t.Errorf("mem.Paths[0] = %q, want %q", mem.Paths[0], "./docs/guide.md")
		}
		if mem.Content != "" {
			t.Errorf("mem.Content = %q, want empty", mem.Content)
		}
	})

	t.Run("memory with content only", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":  "memory",
			"content": "This is important context for the agent.",
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if mem.Content != "This is important context for the agent." {
			t.Errorf("mem.Content = %q, want %q", mem.Content, "This is important context for the agent.")
		}
		if len(mem.Paths) != 0 {
			t.Errorf("mem.Paths length = %d, want 0", len(mem.Paths))
		}
	})

	t.Run("memory with both paths and content", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":  "memory",
			"paths":   []interface{}{"./config.yaml"},
			"content": "Additional context information.",
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 1 {
			t.Errorf("mem.Paths length = %d, want 1", len(mem.Paths))
		}
		if mem.Paths[0] != "./config.yaml" {
			t.Errorf("mem.Paths[0] = %q, want %q", mem.Paths[0], "./config.yaml")
		}
		if mem.Content != "Additional context information." {
			t.Errorf("mem.Content = %q, want %q", mem.Content, "Additional context information.")
		}
	})

	t.Run("memory with empty paths", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "memory",
			"paths":  []interface{}{},
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 0 {
			t.Errorf("mem.Paths length = %d, want 0", len(mem.Paths))
		}
	})

	t.Run("memory with missing optional fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "memory",
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 0 {
			t.Errorf("mem.Paths length = %d, want 0", len(mem.Paths))
		}
		if mem.Content != "" {
			t.Errorf("mem.Content = %q, want empty", mem.Content)
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

func TestFromCanonicalCommand(t *testing.T) {
	adapter := New()

	t.Run("minimal command", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["name"] != "test-command" {
			t.Errorf("output[name] = %q, want %q", output["name"], "test-command")
		}
		if output["description"] != "Test command" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test command")
		}
	})

	t.Run("command with tools", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Tools:       []string{"bash", "edit"},
		}

		output, err := adapter.FromCanonical("command", cmd)
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
		if tools[1] != "Edit" {
			t.Errorf("tools[1] = %q, want %q", tools[1], "Edit")
		}
	})

	t.Run("command with execution", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Execution: canonical.CommandExecution{
				Context: "fork",
				Subtask: true,
				Agent:   "primary-agent",
			},
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		execution, ok := output["execution"].(map[string]interface{})
		if !ok {
			t.Fatal("output[execution] is not map[string]interface{}")
		}
		if execution["context"] != "fork" {
			t.Errorf("execution[context] = %q, want %q", execution["context"], "fork")
		}
		if execution["subtask"] != true {
			t.Errorf("execution[subtask] = %v, want true", execution["subtask"])
		}
		if execution["agent"] != "primary-agent" {
			t.Errorf("execution[agent] = %q, want %q", execution["agent"], "primary-agent")
		}
	})

	t.Run("command with arguments", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Arguments: canonical.CommandArguments{
				Hint: "Enter query",
			},
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		arguments, ok := output["arguments"].(map[string]interface{})
		if !ok {
			t.Fatal("output[arguments] is not map[string]interface{}")
		}
		if arguments["hint"] != "Enter query" {
			t.Errorf("arguments[hint] = %q, want %q", arguments["hint"], "Enter query")
		}
	})

	t.Run("command with model", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Model:       "claude-3-opus",
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["model"] != "claude-3-opus" {
			t.Errorf("output[model] = %q, want %q", output["model"], "claude-3-opus")
		}
	})

	t.Run("command with empty fields", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "",
			Description: "",
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["name"] != "" {
			t.Errorf("output[name] = %q, want empty", output["name"])
		}
		if _, ok := output["tools"]; ok {
			t.Error("output[tools] should not be present for empty tools")
		}
	})
}

func TestFromCanonicalSkill(t *testing.T) {
	adapter := New()

	t.Run("minimal skill", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["name"] != "test-skill" {
			t.Errorf("output[name] = %q, want %q", output["name"], "test-skill")
		}
		if output["description"] != "Test skill" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test skill")
		}
	})

	t.Run("skill with tools", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
			Tools:       []string{"bash", "read"},
		}

		output, err := adapter.FromCanonical("skill", skill)
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
		if tools[1] != "Read" {
			t.Errorf("tools[1] = %q, want %q", tools[1], "Read")
		}
	})

	t.Run("skill with extensions", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
			Extensions: canonical.SkillExtensions{
				License:       "MIT",
				Compatibility: []string{"claude-code", "opencode"},
				Metadata: map[string]string{
					"author":  "test-author",
					"version": "1.0.0",
				},
				Hooks: map[string]string{
					"preInvoke": "setup.sh",
				},
			},
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		extensions, ok := output["extensions"].(map[string]interface{})
		if !ok {
			t.Fatal("output[extensions] is not map[string]interface{}")
		}
		if extensions["license"] != "MIT" {
			t.Errorf("extensions[license] = %q, want %q", extensions["license"], "MIT")
		}
		compat, ok := extensions["compatibility"].([]string)
		if !ok || len(compat) != 2 {
			t.Errorf("extensions[compatibility] = %v, want 2 items", extensions["compatibility"])
		}
		metadata, ok := extensions["metadata"].(map[string]string)
		if !ok || metadata["author"] != "test-author" {
			t.Errorf("extensions[metadata][author] = %v, want test-author", extensions["metadata"])
		}
	})

	t.Run("skill with execution", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
			Execution: canonical.SkillExecution{
				Context:       "fork",
				Agent:         "skill-agent",
				UserInvocable: true,
			},
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		execution, ok := output["execution"].(map[string]interface{})
		if !ok {
			t.Fatal("output[execution] is not map[string]interface{}")
		}
		if execution["context"] != "fork" {
			t.Errorf("execution[context] = %q, want %q", execution["context"], "fork")
		}
		if execution["agent"] != "skill-agent" {
			t.Errorf("execution[agent] = %q, want %q", execution["agent"], "skill-agent")
		}
		if execution["userInvocable"] != true {
			t.Errorf("execution[userInvocable] = %v, want true", execution["userInvocable"])
		}
	})

	t.Run("skill with model", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
			Model:       "claude-3-sonnet",
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["model"] != "claude-3-sonnet" {
			t.Errorf("output[model] = %q, want %q", output["model"], "claude-3-sonnet")
		}
	})

	t.Run("skill with empty fields", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "",
			Description: "",
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["name"] != "" {
			t.Errorf("output[name] = %q, want empty", output["name"])
		}
		if _, ok := output["tools"]; ok {
			t.Error("output[tools] should not be present for empty tools")
		}
		if _, ok := output["extensions"]; ok {
			t.Error("output[extensions] should not be present for empty extensions")
		}
	})
}

func TestFromCanonicalMemory(t *testing.T) {
	adapter := New()

	t.Run("memory with paths only", func(t *testing.T) {
		mem := &canonical.Memory{
			Paths: []string{"./docs/guide.md", "./README.md"},
		}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		paths, ok := output["paths"].([]string)
		if !ok {
			t.Fatal("output[paths] is not []string")
		}
		if len(paths) != 2 {
			t.Errorf("paths length = %d, want 2", len(paths))
		}
		if paths[0] != "./docs/guide.md" {
			t.Errorf("paths[0] = %q, want %q", paths[0], "./docs/guide.md")
		}
		if _, ok := output["content"]; ok {
			t.Error("output[content] should not be present for empty content")
		}
	})

	t.Run("memory with content only", func(t *testing.T) {
		mem := &canonical.Memory{
			Content: "This is important context.",
		}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["content"] != "This is important context." {
			t.Errorf("output[content] = %q, want %q", output["content"], "This is important context.")
		}
		if _, ok := output["paths"]; ok {
			t.Error("output[paths] should not be present for empty paths")
		}
	})

	t.Run("memory with both paths and content", func(t *testing.T) {
		mem := &canonical.Memory{
			Paths:   []string{"./config.yaml"},
			Content: "Additional context information.",
		}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		paths, ok := output["paths"].([]string)
		if !ok {
			t.Fatal("output[paths] is not []string")
		}
		if len(paths) != 1 {
			t.Errorf("paths length = %d, want 1", len(paths))
		}
		if output["content"] != "Additional context information." {
			t.Errorf("output[content] = %q, want %q", output["content"], "Additional context information.")
		}
	})

	t.Run("memory with empty fields", func(t *testing.T) {
		mem := &canonical.Memory{}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if _, ok := output["paths"]; ok {
			t.Error("output[paths] should not be present for empty paths")
		}
		if _, ok := output["content"]; ok {
			t.Error("output[content] should not be present for empty content")
		}
	})
}

func TestToCanonicalEdgeCases(t *testing.T) {
	adapter := New()

	t.Run("missing __type field", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "test",
		}

		_, _, _, _, err := adapter.ToCanonical(input)
		if err == nil {
			t.Error("ToCanonical() expected error for missing __type, got nil")
		}
	})

	t.Run("unknown document type", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "unknown",
		}

		_, _, _, _, err := adapter.ToCanonical(input)
		if err == nil {
			t.Error("ToCanonical() expected error for unknown type, got nil")
		}
	})

	t.Run("nil input values", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "command",
			"name":        nil,
			"description": nil,
			"tools":       nil,
			"execution":   nil,
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if cmd.Name != "" {
			t.Errorf("cmd.Name = %q, want empty", cmd.Name)
		}
		if cmd.Description != "" {
			t.Errorf("cmd.Description = %q, want empty", cmd.Description)
		}
	})

	t.Run("wrong type assertion", func(t *testing.T) {
		_, err := adapter.FromCanonical("command", &canonical.Agent{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}

		_, err = adapter.FromCanonical("skill", &canonical.Command{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}

		_, err = adapter.FromCanonical("memory", &canonical.Skill{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}
	})

	t.Run("unknown doc type in FromCanonical", func(t *testing.T) {
		_, err := adapter.FromCanonical("unknown", &canonical.Agent{})
		if err == nil {
			t.Error("FromCanonical() expected error for unknown type, got nil")
		}
	})
}
