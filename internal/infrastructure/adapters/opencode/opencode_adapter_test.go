package opencode

import (
	"testing"

	canonical "gitlab.com/amoconst/germinator/internal/domain"
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

func TestToCanonicalCommand(t *testing.T) {
	adapter := New()

	t.Run("command with minimal fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "command",
			"description": "Test command",
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if cmd.Description != "Test command" {
			t.Errorf("cmd.Description = %q, want %q", cmd.Description, "Test command")
		}
	})

	t.Run("command with all fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":        "command",
			"name":          "test-command",
			"description":   "Test command",
			"allowed-tools": []interface{}{"bash", "read"},
			"context":       "fork",
			"subtask":       true,
			"agent":         "primary",
			"argument-hint": "<file>",
			"model":         "gpt-4",
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
		if len(cmd.Tools) != 2 {
			t.Errorf("cmd.Tools length = %d, want 2", len(cmd.Tools))
		}
		if cmd.Execution.Context != "fork" {
			t.Errorf("cmd.Execution.Context = %q, want %q", cmd.Execution.Context, "fork")
		}
		if cmd.Execution.Subtask != true {
			t.Errorf("cmd.Execution.Subtask = %v, want true", cmd.Execution.Subtask)
		}
		if cmd.Execution.Agent != "primary" {
			t.Errorf("cmd.Execution.Agent = %q, want %q", cmd.Execution.Agent, "primary")
		}
		if cmd.Arguments.Hint != "<file>" {
			t.Errorf("cmd.Arguments.Hint = %q, want %q", cmd.Arguments.Hint, "<file>")
		}
		if cmd.Model != "gpt-4" {
			t.Errorf("cmd.Model = %q, want %q", cmd.Model, "gpt-4")
		}
	})
}

func TestToCanonicalSkill(t *testing.T) {
	adapter := New()

	t.Run("skill with minimal fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "skill",
			"description": "Test skill",
		}

		_, _, skill, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if skill.Description != "Test skill" {
			t.Errorf("skill.Description = %q, want %q", skill.Description, "Test skill")
		}
	})

	t.Run("skill with all fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":        "skill",
			"name":          "test-skill",
			"description":   "Test skill",
			"allowed-tools": []interface{}{"bash", "read"},
			"license":       "MIT",
			"compatibility": []interface{}{"opencode", "claude-code"},
			"metadata": map[string]interface{}{
				"version": "1.0",
				"author":  "test",
			},
			"hooks": map[string]interface{}{
				"pre":  "echo before",
				"post": "echo after",
			},
			"context":        "fork",
			"agent":          "primary",
			"user-invocable": true,
			"model":          "gpt-4",
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
		if len(skill.Tools) != 2 {
			t.Errorf("skill.Tools length = %d, want 2", len(skill.Tools))
		}
		if skill.Extensions.License != "MIT" {
			t.Errorf("skill.Extensions.License = %q, want %q", skill.Extensions.License, "MIT")
		}
		if len(skill.Extensions.Compatibility) != 2 {
			t.Errorf("skill.Extensions.Compatibility length = %d, want 2", len(skill.Extensions.Compatibility))
		}
		if skill.Extensions.Metadata["version"] != "1.0" {
			t.Errorf("skill.Extensions.Metadata[version] = %q, want %q", skill.Extensions.Metadata["version"], "1.0")
		}
		if skill.Extensions.Hooks["pre"] != "echo before" {
			t.Errorf("skill.Extensions.Hooks[pre] = %q, want %q", skill.Extensions.Hooks["pre"], "echo before")
		}
		if skill.Execution.Context != "fork" {
			t.Errorf("skill.Execution.Context = %q, want %q", skill.Execution.Context, "fork")
		}
		if skill.Execution.Agent != "primary" {
			t.Errorf("skill.Execution.Agent = %q, want %q", skill.Execution.Agent, "primary")
		}
		if skill.Execution.UserInvocable != true {
			t.Errorf("skill.Execution.UserInvocable = %v, want true", skill.Execution.UserInvocable)
		}
		if skill.Model != "gpt-4" {
			t.Errorf("skill.Model = %q, want %q", skill.Model, "gpt-4")
		}
	})
}

func TestToCanonicalMemory(t *testing.T) {
	adapter := New()

	t.Run("memory with paths only", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "memory",
			"paths":  []interface{}{"/path/to/file.md", "/another/path.md"},
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 2 {
			t.Errorf("mem.Paths length = %d, want 2", len(mem.Paths))
		}
		if mem.Paths[0] != "/path/to/file.md" {
			t.Errorf("mem.Paths[0] = %q, want %q", mem.Paths[0], "/path/to/file.md")
		}
	})

	t.Run("memory with content only", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":  "memory",
			"content": "This is memory content",
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if mem.Content != "This is memory content" {
			t.Errorf("mem.Content = %q, want %q", mem.Content, "This is memory content")
		}
	})

	t.Run("memory with both paths and content", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":  "memory",
			"paths":   []interface{}{"/path/to/file.md"},
			"content": "Additional context",
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 1 {
			t.Errorf("mem.Paths length = %d, want 1", len(mem.Paths))
		}
		if mem.Content != "Additional context" {
			t.Errorf("mem.Content = %q, want %q", mem.Content, "Additional context")
		}
	})

	t.Run("memory with empty fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":  "memory",
			"paths":   []interface{}{},
			"content": "",
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

func TestFromCanonicalCommand(t *testing.T) {
	adapter := New()

	t.Run("minimal command", func(t *testing.T) {
		cmd := &canonical.Command{
			Description: "Test command",
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["description"] != "Test command" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test command")
		}
	})

	t.Run("command with all fields", func(t *testing.T) {
		cmd := &canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Tools:       []string{"bash", "read"},
			Execution: canonical.CommandExecution{
				Context: "fork",
				Subtask: true,
				Agent:   "primary",
			},
			Model: "gpt-4",
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
		tools, ok := output["allowed-tools"].([]string)
		if !ok {
			t.Fatal("output[allowed-tools] is not []string")
		}
		if len(tools) != 2 {
			t.Errorf("tools length = %d, want 2", len(tools))
		}
		if output["context"] != "fork" {
			t.Errorf("output[context] = %q, want %q", output["context"], "fork")
		}
		if output["subtask"] != true {
			t.Errorf("output[subtask] = %v, want true", output["subtask"])
		}
		if output["agent"] != "primary" {
			t.Errorf("output[agent] = %q, want %q", output["agent"], "primary")
		}
		if output["model"] != "gpt-4" {
			t.Errorf("output[model] = %q, want %q", output["model"], "gpt-4")
		}
	})

	t.Run("command with empty tools", func(t *testing.T) {
		cmd := &canonical.Command{
			Description: "Test command",
			Tools:       []string{},
		}

		output, err := adapter.FromCanonical("command", cmd)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if _, ok := output["allowed-tools"]; ok {
			t.Error("output[allowed-tools] should be omitted for empty tools")
		}
	})
}

func TestFromCanonicalSkill(t *testing.T) {
	adapter := New()

	t.Run("minimal skill", func(t *testing.T) {
		skill := &canonical.Skill{
			Description: "Test skill",
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["description"] != "Test skill" {
			t.Errorf("output[description] = %q, want %q", output["description"], "Test skill")
		}
	})

	t.Run("skill with all fields", func(t *testing.T) {
		skill := &canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill",
			Tools:       []string{"bash", "read"},
			Extensions: canonical.SkillExtensions{
				License:       "MIT",
				Compatibility: []string{"opencode", "claude-code"},
				Metadata:      map[string]string{"version": "1.0"},
				Hooks:         map[string]string{"pre": "echo before"},
			},
			Execution: canonical.SkillExecution{
				Context:       "fork",
				Agent:         "primary",
				UserInvocable: true,
			},
			Model: "gpt-4",
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
		tools, ok := output["allowed-tools"].([]string)
		if !ok {
			t.Fatal("output[allowed-tools] is not []string")
		}
		if len(tools) != 2 {
			t.Errorf("tools length = %d, want 2", len(tools))
		}
		if output["license"] != "MIT" {
			t.Errorf("output[license] = %q, want %q", output["license"], "MIT")
		}
		compat, ok := output["compatibility"].([]string)
		if !ok {
			t.Fatal("output[compatibility] is not []string")
		}
		if len(compat) != 2 {
			t.Errorf("compatibility length = %d, want 2", len(compat))
		}
		metadata, ok := output["metadata"].(map[string]string)
		if !ok {
			t.Fatal("output[metadata] is not map[string]string")
		}
		if metadata["version"] != "1.0" {
			t.Errorf("metadata[version] = %q, want %q", metadata["version"], "1.0")
		}
		hooks, ok := output["hooks"].(map[string]string)
		if !ok {
			t.Fatal("output[hooks] is not map[string]string")
		}
		if hooks["pre"] != "echo before" {
			t.Errorf("hooks[pre] = %q, want %q", hooks["pre"], "echo before")
		}
		if output["context"] != "fork" {
			t.Errorf("output[context] = %q, want %q", output["context"], "fork")
		}
		if output["agent"] != "primary" {
			t.Errorf("output[agent] = %q, want %q", output["agent"], "primary")
		}
		if output["user-invocable"] != true {
			t.Errorf("output[user-invocable] = %v, want true", output["user-invocable"])
		}
		if output["model"] != "gpt-4" {
			t.Errorf("output[model] = %q, want %q", output["model"], "gpt-4")
		}
	})

	t.Run("skill with empty optional fields", func(t *testing.T) {
		skill := &canonical.Skill{
			Description: "Test skill",
			Tools:       []string{},
			Extensions: canonical.SkillExtensions{
				License:       "",
				Compatibility: []string{},
				Metadata:      map[string]string{},
				Hooks:         map[string]string{},
			},
			Execution: canonical.SkillExecution{
				UserInvocable: false,
			},
		}

		output, err := adapter.FromCanonical("skill", skill)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if _, ok := output["allowed-tools"]; ok {
			t.Error("output[allowed-tools] should be omitted for empty tools")
		}
		if _, ok := output["license"]; ok {
			t.Error("output[license] should be omitted for empty license")
		}
		if _, ok := output["compatibility"]; ok {
			t.Error("output[compatibility] should be omitted for empty compatibility")
		}
		if _, ok := output["metadata"]; ok {
			t.Error("output[metadata] should be omitted for empty metadata")
		}
		if _, ok := output["hooks"]; ok {
			t.Error("output[hooks] should be omitted for empty hooks")
		}
		if _, ok := output["user-invocable"]; ok {
			t.Error("output[user-invocable] should be omitted when false")
		}
	})
}

func TestFromCanonicalMemory(t *testing.T) {
	adapter := New()

	t.Run("memory with paths only", func(t *testing.T) {
		mem := &canonical.Memory{
			Paths: []string{"/path/to/file.md", "/another/path.md"},
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
		if _, ok := output["content"]; ok {
			t.Error("output[content] should be omitted when empty")
		}
	})

	t.Run("memory with content only", func(t *testing.T) {
		mem := &canonical.Memory{
			Content: "This is memory content",
		}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if output["content"] != "This is memory content" {
			t.Errorf("output[content] = %q, want %q", output["content"], "This is memory content")
		}
		if _, ok := output["paths"]; ok {
			t.Error("output[paths] should be omitted when empty")
		}
	})

	t.Run("memory with both paths and content", func(t *testing.T) {
		mem := &canonical.Memory{
			Paths:   []string{"/path/to/file.md"},
			Content: "Additional context",
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
		if output["content"] != "Additional context" {
			t.Errorf("output[content] = %q, want %q", output["content"], "Additional context")
		}
	})

	t.Run("empty memory", func(t *testing.T) {
		mem := &canonical.Memory{}

		output, err := adapter.FromCanonical("memory", mem)
		if err != nil {
			t.Fatalf("FromCanonical() error = %v", err)
		}

		if _, ok := output["paths"]; ok {
			t.Error("output[paths] should be omitted for empty paths")
		}
		if _, ok := output["content"]; ok {
			t.Error("output[content] should be omitted for empty content")
		}
	})
}

func TestToCanonicalEdgeCases(t *testing.T) {
	adapter := New()

	t.Run("missing __type field", func(t *testing.T) {
		input := map[string]interface{}{
			"description": "Test",
		}

		_, _, _, _, err := adapter.ToCanonical(input)
		if err == nil {
			t.Error("ToCanonical() expected error for missing __type, got nil")
		}
	})

	t.Run("unknown document type", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "unknown",
			"description": "Test",
		}

		_, _, _, _, err := adapter.ToCanonical(input)
		if err == nil {
			t.Error("ToCanonical() expected error for unknown type, got nil")
		}
	})

	t.Run("command with nil optional fields", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":        "command",
			"description":   "Test command",
			"allowed-tools": nil,
			"context":       nil,
		}

		_, cmd, _, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(cmd.Tools) != 0 {
			t.Errorf("cmd.Tools length = %d, want 0", len(cmd.Tools))
		}
		if cmd.Execution.Context != "" {
			t.Errorf("cmd.Execution.Context = %q, want empty", cmd.Execution.Context)
		}
	})

	t.Run("skill with invalid metadata value", func(t *testing.T) {
		input := map[string]interface{}{
			"__type":      "skill",
			"description": "Test skill",
			"metadata": map[string]interface{}{
				"version": 123,
				"valid":   "yes",
			},
		}

		_, _, skill, _, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if _, ok := skill.Extensions.Metadata["version"]; ok {
			t.Error("skill.Extensions.Metadata[version] should not exist for non-string value")
		}
		if skill.Extensions.Metadata["valid"] != "yes" {
			t.Errorf("skill.Extensions.Metadata[valid] = %q, want %q", skill.Extensions.Metadata["valid"], "yes")
		}
	})

	t.Run("memory with invalid path types", func(t *testing.T) {
		input := map[string]interface{}{
			"__type": "memory",
			"paths":  []interface{}{"/valid/path.md", 123, true},
		}

		_, _, _, mem, err := adapter.ToCanonical(input)
		if err != nil {
			t.Fatalf("ToCanonical() error = %v", err)
		}

		if len(mem.Paths) != 1 {
			t.Errorf("mem.Paths length = %d, want 1 (only valid string paths)", len(mem.Paths))
		}
	})
}

func TestFromCanonicalEdgeCases(t *testing.T) {
	adapter := New()

	t.Run("unknown document type", func(t *testing.T) {
		_, err := adapter.FromCanonical("unknown", &canonical.Agent{})
		if err == nil {
			t.Error("FromCanonical() expected error for unknown type, got nil")
		}
	})

	t.Run("wrong type for command", func(t *testing.T) {
		_, err := adapter.FromCanonical("command", &canonical.Agent{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}
	})

	t.Run("wrong type for skill", func(t *testing.T) {
		_, err := adapter.FromCanonical("skill", &canonical.Command{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}
	})

	t.Run("wrong type for memory", func(t *testing.T) {
		_, err := adapter.FromCanonical("memory", &canonical.Skill{})
		if err == nil {
			t.Error("FromCanonical() expected error for wrong type, got nil")
		}
	})

	t.Run("nil document", func(t *testing.T) {
		_, err := adapter.FromCanonical("agent", nil)
		if err == nil {
			t.Error("FromCanonical() expected error for nil document, got nil")
		}
	})
}
