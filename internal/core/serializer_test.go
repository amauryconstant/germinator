package core

import (
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models/canonical"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func TestRenderDocumentAgent(t *testing.T) {
	agent := &CanonicalAgent{
		Agent: canonical.Agent{
			Name:             "test-agent",
			Description:      "Test agent",
			Tools:            []string{"editor", "bash"},
			Model:            "anthropic/claude-sonnet-4-20250514",
			PermissionPolicy: canonical.PermissionPolicyBalanced,
			Behavior: canonical.AgentBehavior{
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
		Command: canonical.Command{
			Name:        "test-command",
			Description: "Test command",
			Tools:       []string{"bash"},
			Execution: canonical.CommandExecution{
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
		Skill: canonical.Skill{
			Name:        "test-skill",
			Description: "Test skill description",
			Model:       "anthropic/claude-haiku-4-20250514",
			Extensions: canonical.SkillExtensions{
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
		Memory: canonical.Memory{
			Paths:   []string{"src/**/*.go", "README.md"},
			Content: "Memory content\nWith multiple lines",
		},
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
		Agent: canonical.Agent{
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
				Agent: canonical.Agent{
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
				Agent: canonical.Agent{
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
				Agent: canonical.Agent{
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
				Agent: canonical.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: canonical.PermissionPolicyRestrictive,
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
				Agent: canonical.Agent{
					Name:        "test-agent",
					Description: "Test agent",
					Behavior: canonical.AgentBehavior{
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
				Agent: canonical.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: canonical.PermissionPolicyBalanced,
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
				Agent: canonical.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: canonical.PermissionPolicyBalanced,
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
				Agent: canonical.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: canonical.PermissionPolicyBalanced,
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
				Agent: canonical.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: canonical.PermissionPolicyBalanced,
					Behavior: canonical.AgentBehavior{
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
				Command: canonical.Command{
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
				Command: canonical.Command{
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
				Skill: canonical.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: canonical.SkillExtensions{
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
				Skill: canonical.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: canonical.SkillExtensions{
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
				Memory: canonical.Memory{
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
				Memory: canonical.Memory{
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
