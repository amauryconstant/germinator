package core

import (
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/domain"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func TestRenderDocumentAgent(t *testing.T) {
	agent := &CanonicalAgent{
		Agent: domain.Agent{
			Name:             "test-agent",
			Description:      "Test agent",
			Tools:            []string{"editor", "bash"},
			Model:            "anthropic/claude-sonnet-4-20250514",
			PermissionPolicy: domain.PermissionPolicyBalanced,
			Behavior: domain.AgentBehavior{
				Mode:        "primary",
				Temperature: float64Ptr(0.7),
				Steps:       100,
			},
		},
		Content: "This is test content\nWith multiple lines",
	}

	result, err := RenderDocument(agent, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "name: test-agent") {
		t.Error("Expected name field in output")
	}
	if !strings.Contains(result, "description: Test agent") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "tools:") {
		t.Error("Expected tools section in output")
	}
	if !strings.Contains(result, "  - Editor") {
		t.Error("Expected Editor tool in output")
	}
	if !strings.Contains(result, "  - Bash") {
		t.Error("Expected Bash tool in output")
	}
	if !strings.Contains(result, "This is test content") {
		t.Error("Expected content in output")
	}
	if !strings.Contains(result, "---") {
		t.Error("Expected frontmatter delimiters in output")
	}
}

func TestRenderDocumentCommand(t *testing.T) {
	command := &CanonicalCommand{
		Command: domain.Command{
			Name:        "test-command",
			Description: "Test command",
			Tools:       []string{"bash"},
			Execution: domain.CommandExecution{
				Context: "fork",
				Subtask: true,
			},
		},
		Content: "Command content here",
	}

	result, err := RenderDocument(command, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "description: Test command") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "Command content here") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentSkill(t *testing.T) {
	skill := &CanonicalSkill{
		Skill: domain.Skill{
			Name:        "test-skill",
			Description: "Test skill description",
			Model:       "anthropic/claude-haiku-4-20250514",
			Extensions: domain.SkillExtensions{
				License:       "MIT",
				Compatibility: []string{"claude-code", "opencode"},
			},
		},
		Content: "Skill instructions",
	}

	result, err := RenderDocument(skill, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "name: test-skill") {
		t.Error("Expected name field in output")
	}
	if !strings.Contains(result, "description: Test skill description") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "Skill instructions") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentMemory(t *testing.T) {
	memory := &CanonicalMemory{
		Memory: domain.Memory{
			Paths: []string{"src/**/*.go", "README.md"},
		},
		Content: "Memory content\nWith multiple lines",
	}

	result, err := RenderDocument(memory, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "paths:") {
		t.Error("Expected paths section in output")
	}
	if !strings.Contains(result, "Memory content") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentUnknownType(t *testing.T) {
	type UnknownType struct{}

	doc := &UnknownType{}
	_, err := RenderDocument(doc, "claude-code")
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}

func TestRenderDocumentInvalidTemplate(t *testing.T) {
	agent := &CanonicalAgent{
		Agent: domain.Agent{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	_, err := RenderDocument(agent, "invalid-platform")
	if err == nil {
		t.Error("Expected error for invalid platform")
	}
}

func TestRenderDocumentMarkdownBodyPreservation(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "simple text",
			content: "Simple text content",
		},
		{
			name:    "multiline",
			content: "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "markdown formatting",
			content: "**Bold text** and *italic text*",
		},
		{
			name:    "code blocks",
			content: "```\ncode here\n```",
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &CanonicalAgent{
				Agent: domain.Agent{
					Name:        "test-agent",
					Description: "Test agent",
				},
				Content: tt.content,
			}

			result, err := RenderDocument(agent, "claude-code")
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}

			if tt.content != "" && !strings.Contains(result, tt.content) {
				t.Errorf("Expected content %q in output", tt.content)
			}
		})
	}
}

func TestRenderCanonicalAgentClaudeCode(t *testing.T) {
	tests := []struct {
		name        string
		agent       *CanonicalAgent
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "minimal agent",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:        "test-agent",
					Description: "Test agent",
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "description: Test agent") {
					t.Error("Expected description in output")
				}
			},
		},
		{
			name: "agent with tools",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:        "test-agent",
					Description: "Test agent",
					Tools:       []string{"editor", "bash", "grep"},
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "tools:") {
					t.Error("Expected tools section")
				}
				if !strings.Contains(output, "  - Editor") {
					t.Error("Expected PascalCase tool name for Claude Code")
				}
			},
		},
		{
			name: "agent with permission policy",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: domain.PermissionPolicyRestrictive,
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "permissionMode: default") {
					t.Error("Expected restrictive policy to map to default permissionMode")
				}
			},
		},
		{
			name: "agent with behavior",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:        "test-agent",
					Description: "Test agent",
					Behavior: domain.AgentBehavior{
						Mode:        "primary",
						Temperature: float64Ptr(0.5),
						Steps:       50,
					},
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "temperature: 0.5") {
					t.Error("Expected temperature in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.agent, "claude-code")
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}
			tt.checkOutput(t, result)
		})
	}
}

func TestRenderCanonicalAgentOpenCode(t *testing.T) {
	tests := []struct {
		name        string
		agent       *CanonicalAgent
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "minimal agent",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: domain.PermissionPolicyBalanced,
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "description: Test agent") {
					t.Error("Expected description in output")
				}
				if !strings.Contains(output, "permission:") {
					t.Error("Expected permission section for OpenCode")
				}
			},
		},
		{
			name: "agent with tools",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: domain.PermissionPolicyBalanced,
					Tools:            []string{"editor", "bash"},
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "tools:") {
					t.Error("Expected tools section")
				}
				if !strings.Contains(output, "  editor:") {
					t.Error("Expected lowercase tool names for OpenCode")
				}
			},
		},
		{
			name: "agent with disallowed tools",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: domain.PermissionPolicyBalanced,
					Tools:            []string{"editor", "bash"},
					DisallowedTools:  []string{"write", "computer"},
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "  write:") {
					t.Error("Expected disallowed tools in output")
				}
				if !strings.Contains(output, "  computer:") {
					t.Error("Expected disallowed tools in output")
				}
			},
		},
		{
			name: "agent with behavior",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: domain.PermissionPolicyBalanced,
					Behavior: domain.AgentBehavior{
						Mode:        "primary",
						Temperature: float64Ptr(0.7),
						Steps:       100,
					},
				},
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "mode: primary") {
					t.Error("Expected mode field in output")
				}
				if !strings.Contains(output, "temperature: 0.7") {
					t.Error("Expected temperature in output")
				}
				if !strings.Contains(output, "maxSteps: 100") {
					t.Error("Expected maxSteps field in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.agent, "opencode")
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}
			tt.checkOutput(t, result)
		})
	}
}

func TestRenderCanonicalCommand(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		command  *CanonicalCommand
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code",
			platform: "claude-code",
			command: &CanonicalCommand{
				Command: domain.Command{
					Name:        "test-command",
					Description: "Test command",
					Tools:       []string{"bash"},
				},
				Content: "echo 'hello'",
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "tools:") {
					t.Error("Expected tools for Claude Code")
				}
			},
		},
		{
			name:     "opencode",
			platform: "opencode",
			command: &CanonicalCommand{
				Command: domain.Command{
					Name:        "test-command",
					Description: "Test command",
					Tools:       []string{"bash"},
				},
				Content: "echo 'hello'",
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "description: Test command") {
					t.Error("Expected description for OpenCode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.command, tt.platform)
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestRenderCanonicalSkill(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		skill    *CanonicalSkill
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code with extensions",
			platform: "claude-code",
			skill: &CanonicalSkill{
				Skill: domain.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: domain.SkillExtensions{
						License: "MIT",
					},
				},
				Content: "Skill content",
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "license: MIT") {
					t.Error("Expected license in output")
				}
			},
		},
		{
			name:     "opencode with extensions",
			platform: "opencode",
			skill: &CanonicalSkill{
				Skill: domain.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: domain.SkillExtensions{
						License:       "MIT",
						Compatibility: []string{"claude-code", "opencode"},
					},
				},
				Content: "Skill content",
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "license:") {
					t.Error("Expected license in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.skill, tt.platform)
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestRenderCanonicalMemory(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		memory   *CanonicalMemory
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code with paths",
			platform: "claude-code",
			memory: &CanonicalMemory{
				Memory: domain.Memory{
					Paths:   []string{"README.md", "AGENTS.md"},
					Content: "Project context",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "paths:") {
					t.Error("Expected paths in output")
				}
			},
		},
		{
			name:     "opencode with paths",
			platform: "opencode",
			memory: &CanonicalMemory{
				Memory: domain.Memory{
					Paths:   []string{"README.md", "AGENTS.md"},
					Content: "Project context",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "@README.md") {
					t.Error("Expected @README.md reference for OpenCode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.memory, tt.platform)
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestGetDocType(t *testing.T) {
	tests := []struct {
		name      string
		doc       interface{}
		wantType  string
		wantError bool
	}{
		{
			name:     "CanonicalAgent",
			doc:      &CanonicalAgent{},
			wantType: "agent",
		},
		{
			name:     "CanonicalCommand",
			doc:      &CanonicalCommand{},
			wantType: "command",
		},
		{
			name:     "CanonicalSkill",
			doc:      &CanonicalSkill{},
			wantType: "skill",
		},
		{
			name:     "CanonicalMemory",
			doc:      &CanonicalMemory{},
			wantType: "memory",
		},
		{
			name:      "Unknown type",
			doc:       &struct{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docType, err := getDocType(tt.doc)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error for unknown type")
				}
				return
			}
			if err != nil {
				t.Fatalf("getDocType failed: %v", err)
			}
			if docType != tt.wantType {
				t.Errorf("getDocType() = %q, want %q", docType, tt.wantType)
			}
		})
	}
}

func TestMarshalCanonicalAgent(t *testing.T) {
	tests := []struct {
		name  string
		agent *CanonicalAgent
		check func(t *testing.T, output string)
	}{
		{
			name: "agent with all fields",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:             "test-agent",
					Description:      "Test agent description",
					Tools:            []string{"bash", "read", "grep"},
					DisallowedTools:  []string{"write", "webfetch"},
					PermissionPolicy: domain.PermissionPolicyBalanced,
					Behavior: domain.AgentBehavior{
						Mode:        "primary",
						Temperature: float64Ptr(0.7),
						Steps:       100,
						Prompt:      "Test prompt",
						Hidden:      true,
						Disabled:    false,
					},
					Model: "claude-sonnet-4-5-20250929",
					Targets: domain.PlatformConfig{
						"claude-code": map[string]interface{}{
							"skills": []string{"skill1", "skill2"},
						},
					},
					Extensions: domain.AgentExtensions{
						Hooks: map[string]string{
							"SessionStart": "hook1",
						},
					},
				},
				Content: "Agent content here",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "name: test-agent") {
					t.Error("Expected name field")
				}
				if !containsString(output, "description: Test agent description") {
					t.Error("Expected description field")
				}
				if !containsString(output, "tools:") {
					t.Error("Expected tools section")
				}
				if !containsString(output, "  - bash") {
					t.Error("Expected bash tool")
				}
				if !containsString(output, "disallowedTools:") {
					t.Error("Expected disallowedTools section")
				}
				if !containsString(output, "permissionPolicy: balanced") {
					t.Error("Expected permissionPolicy field")
				}
				if !containsString(output, "behavior:") {
					t.Error("Expected behavior section")
				}
				if !containsString(output, "mode: primary") {
					t.Error("Expected mode field")
				}
				if !containsString(output, "temperature: 0.7") {
					t.Error("Expected temperature field")
				}
				if !containsString(output, "steps: 100") {
					t.Error("Expected steps field")
				}
				if !containsString(output, "prompt: Test prompt") {
					t.Error("Expected prompt field")
				}
				if !containsString(output, "hidden: true") {
					t.Error("Expected hidden field")
				}
				if containsString(output, "disabled: true") {
					t.Error("Should not contain disabled field when false")
				}
				if !containsString(output, "model: claude-sonnet-4-5-20250929") {
					t.Error("Expected model field")
				}
				if !containsString(output, "extensions:") {
					t.Error("Expected extensions section")
				}
				if !containsString(output, "targets:") {
					t.Error("Expected targets section")
				}
				if !containsString(output, "---") {
					t.Error("Expected frontmatter delimiters")
				}
				if !containsString(output, "Agent content here") {
					t.Error("Expected content after frontmatter")
				}
			},
		},
		{
			name: "agent with minimal fields",
			agent: &CanonicalAgent{
				Agent: domain.Agent{
					Name:        "minimal-agent",
					Description: "Minimal agent",
				},
				Content: "Minimal content",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "name: minimal-agent") {
					t.Error("Expected name field")
				}
				if !containsString(output, "description: Minimal agent") {
					t.Error("Expected description field")
				}
				if containsString(output, "tools:") {
					t.Error("Should not contain tools section when empty")
				}
				if containsString(output, "permissionPolicy:") {
					t.Error("Should not contain permissionPolicy when empty")
				}
				if containsString(output, "behavior:") {
					t.Error("Should not contain behavior section when all empty")
				}
				if containsString(output, "model:") {
					t.Error("Should not contain model field when empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCanonical(tt.agent)
			if err != nil {
				t.Fatalf("MarshalCanonical() failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestMarshalCanonicalCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *CanonicalCommand
		check   func(t *testing.T, output string)
	}{
		{
			name: "command with all fields",
			command: &CanonicalCommand{
				Command: domain.Command{
					Name:        "test-command",
					Description: "Test command",
					Tools:       []string{"bash", "read"},
					Execution: domain.CommandExecution{
						Context: "fork",
						Subtask: true,
						Agent:   "general-purpose",
					},
					Arguments: domain.CommandArguments{
						Hint: "[issue-number]",
					},
					Model: "claude-haiku-4-20250514",
				},
				Content: "Command content",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "name: test-command") {
					t.Error("Expected name field")
				}
				if !containsString(output, "description: Test command") {
					t.Error("Expected description field")
				}
				if !containsString(output, "tools:") {
					t.Error("Expected tools section")
				}
				if !containsString(output, "execution:") {
					t.Error("Expected execution section")
				}
				if !containsString(output, "context: fork") {
					t.Error("Expected context field")
				}
				if !containsString(output, "subtask: true") {
					t.Error("Expected subtask field")
				}
				if !containsString(output, "agent: general-purpose") {
					t.Error("Expected agent field")
				}
				if !containsString(output, "arguments:") {
					t.Error("Expected arguments section")
				}
				if !containsString(output, "hint: [issue-number]") {
					t.Error("Expected hint field")
				}
				if !containsString(output, "model: claude-haiku-4-20250514") {
					t.Error("Expected model field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCanonical(tt.command)
			if err != nil {
				t.Fatalf("MarshalCanonical() failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestMarshalCanonicalSkill(t *testing.T) {
	tests := []struct {
		name  string
		skill *CanonicalSkill
		check func(t *testing.T, output string)
	}{
		{
			name: "skill with all fields",
			skill: &CanonicalSkill{
				Skill: domain.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Tools:       []string{"bash"},
					Extensions: domain.SkillExtensions{
						License:       "MIT",
						Compatibility: []string{"claude-code", "opencode"},
						Metadata: map[string]string{
							"author":  "test",
							"version": "1.0.0",
						},
						Hooks: map[string]string{
							"Load": "hook1",
						},
					},
					Execution: domain.SkillExecution{
						Context:       "fork",
						Agent:         "general-purpose",
						UserInvocable: true,
					},
					Model: "claude-sonnet-4-5-20250929",
				},
				Content: "Skill content",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "name: test-skill") {
					t.Error("Expected name field")
				}
				if !containsString(output, "description: Test skill") {
					t.Error("Expected description field")
				}
				if !containsString(output, "tools:") {
					t.Error("Expected tools section")
				}
				if !containsString(output, "extensions:") {
					t.Error("Expected extensions section")
				}
				if !containsString(output, "license: MIT") {
					t.Error("Expected license field")
				}
				if !containsString(output, "compatibility:") {
					t.Error("Expected compatibility field")
				}
				if !containsString(output, "metadata:") {
					t.Error("Expected metadata field")
				}
				if !containsString(output, "hooks:") {
					t.Error("Expected hooks field")
				}
				if !containsString(output, "execution:") {
					t.Error("Expected execution section")
				}
				if !containsString(output, "context: fork") {
					t.Error("Expected context field")
				}
				if !containsString(output, "agent: general-purpose") {
					t.Error("Expected agent field")
				}
				if !containsString(output, "userInvocable: true") {
					t.Error("Expected userInvocable field")
				}
				if !containsString(output, "model: claude-sonnet-4-5-20250929") {
					t.Error("Expected model field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCanonical(tt.skill)
			if err != nil {
				t.Fatalf("MarshalCanonical() failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestMarshalCanonicalMemory(t *testing.T) {
	tests := []struct {
		name   string
		memory *CanonicalMemory
		check  func(t *testing.T, output string)
	}{
		{
			name: "memory with paths only",
			memory: &CanonicalMemory{
				Memory: domain.Memory{
					Paths: []string{"src/**/*.go", "README.md"},
				},
				Content: "",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "paths:") {
					t.Error("Expected paths section")
				}
				if !containsString(output, "  - src/**/*.go") {
					t.Error("Expected path entry")
				}
				if containsString(output, "content:") {
					t.Error("Should not contain content field when empty and paths exist")
				}
				if !containsString(output, "---") {
					t.Error("Expected frontmatter delimiter")
				}
			},
		},
		{
			name: "memory with content only",
			memory: &CanonicalMemory{
				Memory: domain.Memory{
					Paths: []string{},
				},
				Content: "This is the memory content\nwith multiple lines",
			},
			check: func(t *testing.T, output string) {
				if containsString(output, "paths:") {
					t.Error("Should not contain paths section when empty")
				}
				if !containsString(output, "content: |") {
					t.Error("Expected content with pipe syntax")
				}
				if !containsString(output, "This is the memory content") {
					t.Error("Expected content text")
				}
			},
		},
		{
			name: "memory with both paths and content",
			memory: &CanonicalMemory{
				Memory: domain.Memory{
					Paths: []string{"src/**/*.go"},
				},
				Content: "Additional context",
			},
			check: func(t *testing.T, output string) {
				if !containsString(output, "paths:") {
					t.Error("Expected paths section")
				}
				if !containsString(output, "content: |") {
					t.Error("Expected content with pipe syntax")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalCanonical(tt.memory)
			if err != nil {
				t.Fatalf("MarshalCanonical() failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestMarshalCanonicalAllEmptyOptionalFields(t *testing.T) {
	agent := &CanonicalAgent{
		Agent: domain.Agent{
			Name:        "test-agent",
			Description: "Test agent",
		},
		Content: "Content",
	}

	result, err := MarshalCanonical(agent)
	if err != nil {
		t.Fatalf("MarshalCanonical() failed: %v", err)
	}

	if containsString(result, "tools:") {
		t.Error("Should not contain tools when empty")
	}
	if containsString(result, "disallowedTools:") {
		t.Error("Should not contain disallowedTools when empty")
	}
	if containsString(result, "permissionPolicy:") {
		t.Error("Should not contain permissionPolicy when empty")
	}
	if containsString(result, "behavior:") {
		t.Error("Should not contain behavior when all fields empty")
	}
	if containsString(result, "model:") {
		t.Error("Should not contain model when empty")
	}
	if containsString(result, "extensions:") {
		t.Error("Should not contain extensions when empty")
	}
	if containsString(result, "targets:") {
		t.Error("Should not contain targets when empty")
	}
}

func TestMarshalCanonicalNoAdapterFieldAccess(t *testing.T) {
	agent := &CanonicalAgent{
		Agent: domain.Agent{
			Name:        "test-agent",
			Description: "Test agent",
			Tools:       []string{"bash"},
		},
		Content: "Content",
	}

	result, err := MarshalCanonical(agent)
	if err != nil {
		t.Fatalf("MarshalCanonical() failed: %v", err)
	}

	if containsString(result, ".Adapter") {
		t.Error("Should not access Adapter field in canonical templates")
	}
	if containsString(result, "adapter.") {
		t.Error("Should not access adapter methods in canonical templates")
	}
}

func TestMarshalCanonicalUnknownType(t *testing.T) {
	type UnknownType struct{}
	doc := &UnknownType{}

	_, err := MarshalCanonical(doc)
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !containsString(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}
