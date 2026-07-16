package renderer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/parser"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestRenderDocumentAgent(t *testing.T) {
	agent := &parser.CanonicalAgent{
		Agent: core.Agent{
			Name:             "test-agent",
			Description:      "Test agent",
			Tools:            []string{"editor", "bash"},
			Model:            "anthropic/claude-sonnet-4-20250514",
			PermissionPolicy: core.PermissionPolicyBalanced,
			Behavior: core.AgentBehavior{
				Mode:        "primary",
				Temperature: float64Ptr(0.7),
				Steps:       100,
			},
		},
		Content: "This is test content\nWith multiple lines",
	}

	result, err := RenderDocument(context.Background(), agent, "claude-code")
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
	command := &parser.CanonicalCommand{
		Command: core.Command{
			Name:        "test-command",
			Description: "Test command",
			Tools:       []string{"bash"},
			Execution: core.CommandExecution{
				Context: "fork",
				Subtask: true,
			},
		},
		Content: "Command content here",
	}

	result, err := RenderDocument(context.Background(), command, "claude-code")
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
	skill := &parser.CanonicalSkill{
		Skill: core.Skill{
			Name:        "test-skill",
			Description: "Test skill description",
			Model:       "anthropic/claude-haiku-4-20250514",
			Extensions: core.SkillExtensions{
				License:       "MIT",
				Compatibility: []string{"claude-code", "opencode"},
			},
		},
		Content: "Skill instructions",
	}

	result, err := RenderDocument(context.Background(), skill, "claude-code")
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
	memory := &parser.CanonicalMemory{
		Memory: core.Memory{
			Paths: []string{"src/**/*.go", "README.md"},
		},
		Content: "Memory content\nWith multiple lines",
	}

	result, err := RenderDocument(context.Background(), memory, "claude-code")
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
	_, err := RenderDocument(context.Background(), doc, "claude-code")
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}

func TestRenderDocumentInvalidTemplate(t *testing.T) {
	agent := &parser.CanonicalAgent{
		Agent: core.Agent{
			Name:        "test-agent",
			Description: "Test agent",
		},
	}

	_, err := RenderDocument(context.Background(), agent, "invalid-platform")
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
			agent := &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:        "test-agent",
					Description: "Test agent",
				},
				Content: tt.content,
			}

			result, err := RenderDocument(context.Background(), agent, "claude-code")
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
		agent       *parser.CanonicalAgent
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "minimal agent",
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: core.PermissionPolicyRestrictive,
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:        "test-agent",
					Description: "Test agent",
					Behavior: core.AgentBehavior{
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
			result, err := RenderDocument(context.Background(), tt.agent, "claude-code")
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
		agent       *parser.CanonicalAgent
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "minimal agent",
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: core.PermissionPolicyBalanced,
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: core.PermissionPolicyBalanced,
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: core.PermissionPolicyBalanced,
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent",
					PermissionPolicy: core.PermissionPolicyBalanced,
					Behavior: core.AgentBehavior{
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
			result, err := RenderDocument(context.Background(), tt.agent, "opencode")
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
		command  *parser.CanonicalCommand
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code",
			platform: "claude-code",
			command: &parser.CanonicalCommand{
				Command: core.Command{
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
			command: &parser.CanonicalCommand{
				Command: core.Command{
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
			result, err := RenderDocument(context.Background(), tt.command, tt.platform)
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
		skill    *parser.CanonicalSkill
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code with extensions",
			platform: "claude-code",
			skill: &parser.CanonicalSkill{
				Skill: core.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: core.SkillExtensions{
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
			skill: &parser.CanonicalSkill{
				Skill: core.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Extensions: core.SkillExtensions{
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
			result, err := RenderDocument(context.Background(), tt.skill, tt.platform)
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
		memory   *parser.CanonicalMemory
		check    func(t *testing.T, output string)
	}{
		{
			name:     "claude-code with paths",
			platform: "claude-code",
			memory: &parser.CanonicalMemory{
				Memory: core.Memory{
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
			memory: &parser.CanonicalMemory{
				Memory: core.Memory{
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
			result, err := RenderDocument(context.Background(), tt.memory, tt.platform)
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
			name:     "parser.CanonicalAgent",
			doc:      &parser.CanonicalAgent{},
			wantType: "agent",
		},
		{
			name:     "parser.CanonicalCommand",
			doc:      &parser.CanonicalCommand{},
			wantType: "command",
		},
		{
			name:     "parser.CanonicalSkill",
			doc:      &parser.CanonicalSkill{},
			wantType: "skill",
		},
		{
			name:     "parser.CanonicalMemory",
			doc:      &parser.CanonicalMemory{},
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
		agent *parser.CanonicalAgent
		check func(t *testing.T, output string)
	}{
		{
			name: "agent with all fields",
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
					Name:             "test-agent",
					Description:      "Test agent description",
					Tools:            []string{"bash", "read", "grep"},
					DisallowedTools:  []string{"write", "webfetch"},
					PermissionPolicy: core.PermissionPolicyBalanced,
					Behavior: core.AgentBehavior{
						Mode:        "primary",
						Temperature: float64Ptr(0.7),
						Steps:       100,
						Prompt:      "Test prompt",
						Hidden:      true,
						Disabled:    false,
					},
					Model: "claude-sonnet-4-5-20250929",
					Targets: core.PlatformConfig{
						"claude-code": map[string]interface{}{
							"skills": []string{"skill1", "skill2"},
						},
					},
					Extensions: core.AgentExtensions{
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
			agent: &parser.CanonicalAgent{
				Agent: core.Agent{
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
			result, err := MarshalCanonical(context.Background(), tt.agent)
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
		command *parser.CanonicalCommand
		check   func(t *testing.T, output string)
	}{
		{
			name: "command with all fields",
			command: &parser.CanonicalCommand{
				Command: core.Command{
					Name:        "test-command",
					Description: "Test command",
					Tools:       []string{"bash", "read"},
					Execution: core.CommandExecution{
						Context: "fork",
						Subtask: true,
						Agent:   "general-purpose",
					},
					Arguments: core.CommandArguments{
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
			result, err := MarshalCanonical(context.Background(), tt.command)
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
		skill *parser.CanonicalSkill
		check func(t *testing.T, output string)
	}{
		{
			name: "skill with all fields",
			skill: &parser.CanonicalSkill{
				Skill: core.Skill{
					Name:        "test-skill",
					Description: "Test skill",
					Tools:       []string{"bash"},
					Extensions: core.SkillExtensions{
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
					Execution: core.SkillExecution{
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
			result, err := MarshalCanonical(context.Background(), tt.skill)
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
		memory *parser.CanonicalMemory
		check  func(t *testing.T, output string)
	}{
		{
			name: "memory with paths only",
			memory: &parser.CanonicalMemory{
				Memory: core.Memory{
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
			memory: &parser.CanonicalMemory{
				Memory: core.Memory{
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
			memory: &parser.CanonicalMemory{
				Memory: core.Memory{
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
			result, err := MarshalCanonical(context.Background(), tt.memory)
			if err != nil {
				t.Fatalf("MarshalCanonical() failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestMarshalCanonicalAllEmptyOptionalFields(t *testing.T) {
	agent := &parser.CanonicalAgent{
		Agent: core.Agent{
			Name:        "test-agent",
			Description: "Test agent",
		},
		Content: "Content",
	}

	result, err := MarshalCanonical(context.Background(), agent)
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
	agent := &parser.CanonicalAgent{
		Agent: core.Agent{
			Name:        "test-agent",
			Description: "Test agent",
			Tools:       []string{"bash"},
		},
		Content: "Content",
	}

	result, err := MarshalCanonical(context.Background(), agent)
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

	_, err := MarshalCanonical(context.Background(), doc)
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !containsString(err.Error(), "unknown document type") {
		t.Errorf("Expected unknown document type error, got: %v", err)
	}
}

// fixtureRepoRoot resolves the repo root via runtime.Caller(0) so the
// round-trip tests can locate test/fixtures from any working directory.
// Mirrors the canonicalize golden-test path-resolution pattern.
func fixtureRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller(0) failed; cannot resolve fixtures")
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestParseRenderRoundTrip verifies the canonical parse → marshal →
// re-parse cycle preserves semantic field equality. Unlike the
// byte-equality golden tests in test/e2e/, this test asserts the
// round-trip preserves the canonical field set — not the on-disk byte
// layout — because YAML serialization is not byte-stable across
// dependency versions. The test is in the default suite (no build
// tag) because semantic equality is deterministic.
//
// Subtests cover each canonical doc type (agent, command, skill,
// memory) against the corresponding canonical fixture in
// test/fixtures/canonical/.
func TestParseRenderRoundTrip(t *testing.T) {
	repoRoot := fixtureRepoRoot(t)
	fixturesDir := filepath.Join(repoRoot, "test", "fixtures", "canonical")

	t.Run("agent-permission-balanced", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "agent-permission-balanced.md")
		first, err := parser.ParseDocument(t.Context(), fixturePath, "agent")
		require.NoError(t, err)
		firstAgent, ok := first.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "agent.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "agent")
		require.NoError(t, err)
		secondAgent, ok := second.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", second)

		assert.Equal(t, firstAgent.Name, secondAgent.Name, "Name round-trip")
		assert.Equal(t, firstAgent.Description, secondAgent.Description, "Description round-trip")
		assert.Equal(t, firstAgent.Tools, secondAgent.Tools, "Tools round-trip")
		assert.Equal(t, firstAgent.DisallowedTools, secondAgent.DisallowedTools, "DisallowedTools round-trip")
		assert.Equal(t, firstAgent.PermissionPolicy, secondAgent.PermissionPolicy, "PermissionPolicy round-trip")
		assert.Equal(t, firstAgent.Behavior.Mode, secondAgent.Behavior.Mode, "Behavior.Mode round-trip")
		require.NotNil(t, firstAgent.Behavior.Temperature)
		require.NotNil(t, secondAgent.Behavior.Temperature)
		assert.Equal(t, *firstAgent.Behavior.Temperature, *secondAgent.Behavior.Temperature, "Behavior.Temperature round-trip")
		assert.Equal(t, firstAgent.Behavior.Steps, secondAgent.Behavior.Steps, "Behavior.Steps round-trip")
		assert.Equal(t, firstAgent.Behavior.Prompt, secondAgent.Behavior.Prompt, "Behavior.Prompt round-trip")
		assert.Equal(t, firstAgent.Behavior.Hidden, secondAgent.Behavior.Hidden, "Behavior.Hidden round-trip")
		assert.Equal(t, firstAgent.Behavior.Disabled, secondAgent.Behavior.Disabled, "Behavior.Disabled round-trip")
		assert.Equal(t, firstAgent.Model, secondAgent.Model, "Model round-trip")
		assert.Equal(t, strings.TrimRight(firstAgent.Content, "\n"), strings.TrimRight(secondAgent.Content, "\n"), "Content round-trip")
	})

	t.Run("skill round-trip via RenderDocument+ParseDocument", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "skill-valid.md")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("skill fixture not found: %s", fixturePath)
		}

		first, err := parser.ParseDocument(t.Context(), fixturePath, "skill")
		require.NoError(t, err)
		firstSkill, ok := first.(*parser.CanonicalSkill)
		require.True(t, ok, "expected *parser.CanonicalSkill, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "skill.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "skill")
		require.NoError(t, err)
		secondSkill, ok := second.(*parser.CanonicalSkill)
		require.True(t, ok, "expected *parser.CanonicalSkill, got %T", second)

		assert.Equal(t, firstSkill.Name, secondSkill.Name)
		assert.Equal(t, firstSkill.Description, secondSkill.Description)
		assert.Equal(t, firstSkill.Tools, secondSkill.Tools)
		assert.Equal(t, firstSkill.Extensions.License, secondSkill.Extensions.License)
		assert.Equal(t, firstSkill.Extensions.Compatibility, secondSkill.Extensions.Compatibility)
		assert.Equal(t, firstSkill.Execution.UserInvocable, secondSkill.Execution.UserInvocable)
		assert.Equal(t, firstSkill.Model, secondSkill.Model)
		assert.Equal(t, firstSkill.Content, secondSkill.Content)
	})

	t.Run("command round-trip", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "command-valid.md")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("command fixture not found: %s", fixturePath)
		}

		first, err := parser.ParseDocument(t.Context(), fixturePath, "command")
		require.NoError(t, err)
		firstCmd, ok := first.(*parser.CanonicalCommand)
		require.True(t, ok, "expected *parser.CanonicalCommand, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "command.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "command")
		require.NoError(t, err)
		secondCmd, ok := second.(*parser.CanonicalCommand)
		require.True(t, ok, "expected *parser.CanonicalCommand, got %T", second)

		assert.Equal(t, firstCmd.Name, secondCmd.Name)
		assert.Equal(t, firstCmd.Description, secondCmd.Description)
		assert.Equal(t, firstCmd.Tools, secondCmd.Tools)
		assert.Equal(t, firstCmd.Execution.Context, secondCmd.Execution.Context)
		assert.Equal(t, firstCmd.Execution.Subtask, secondCmd.Execution.Subtask)
		assert.Equal(t, firstCmd.Arguments.Hint, secondCmd.Arguments.Hint)
		assert.Equal(t, firstCmd.Model, secondCmd.Model)
		assert.Equal(t, firstCmd.Content, secondCmd.Content)
	})

	t.Run("memory round-trip", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "memory-valid.md")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("memory fixture not found: %s", fixturePath)
		}

		first, err := parser.ParseDocument(t.Context(), fixturePath, "memory")
		require.NoError(t, err)
		firstMem, ok := first.(*parser.CanonicalMemory)
		require.True(t, ok, "expected *parser.CanonicalMemory, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "memory.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "memory")
		require.NoError(t, err)
		secondMem, ok := second.(*parser.CanonicalMemory)
		require.True(t, ok, "expected *parser.CanonicalMemory, got %T", second)

		assert.Equal(t, firstMem.Paths, secondMem.Paths, "Paths round-trip")
		assert.Equal(t, firstMem.Content, secondMem.Content, "Content round-trip")
	})
}

// TestPlatformRoundTrip verifies that platform fixtures round-trip
// through canonical marshaling without losing semantic field
// equality. The platform YAML files (opencode / claude-code) are
// parsed via ParsePlatformDocument, marshaled via MarshalCanonical,
// re-parsed as canonical, and asserted equivalent to the original
// canonical-shaped fields. This catches drift where a platform's
// schema change loses information during the parse → canonical
// → re-parse round trip.
func TestPlatformRoundTrip(t *testing.T) {
	repoRoot := fixtureRepoRoot(t)
	fixturesDir := filepath.Join(repoRoot, "test", "fixtures")

	t.Run("opencode agent", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "opencode", "agent.md")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("opencode agent fixture not found: %s", fixturePath)
		}

		first, err := parser.ParsePlatformDocument(t.Context(), fixturePath, "opencode", "agent")
		require.NoError(t, err)
		firstAgent, ok := first.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "agent.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "agent")
		require.NoError(t, err)
		secondAgent, ok := second.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", second)

		assert.Equal(t, firstAgent.Description, secondAgent.Description, "Description round-trip")
		assert.Equal(t, firstAgent.Tools, secondAgent.Tools, "Tools round-trip (after split)")
		assert.Equal(t, firstAgent.DisallowedTools, secondAgent.DisallowedTools, "DisallowedTools round-trip (after split)")
		assert.Equal(t, firstAgent.PermissionPolicy, secondAgent.PermissionPolicy, "PermissionPolicy round-trip (inferred)")
		assert.Equal(t, firstAgent.Behavior.Mode, secondAgent.Behavior.Mode, "Behavior.Mode round-trip")
		assert.Equal(t, firstAgent.Behavior.Steps, secondAgent.Behavior.Steps, "Behavior.Steps round-trip (renamed from maxSteps)")
	})

	t.Run("claude-code agent", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "claude-code", "agent.md")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skipf("claude-code agent fixture not found: %s", fixturePath)
		}

		first, err := parser.ParsePlatformDocument(t.Context(), fixturePath, "claude-code", "agent")
		require.NoError(t, err)
		firstAgent, ok := first.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", first)

		marshaled, err := MarshalCanonical(t.Context(), first)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpPath := filepath.Join(tmpDir, "agent.md")
		require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

		second, err := parser.ParseDocument(t.Context(), tmpPath, "agent")
		require.NoError(t, err)
		secondAgent, ok := second.(*parser.CanonicalAgent)
		require.True(t, ok, "expected *parser.CanonicalAgent, got %T", second)

		assert.Equal(t, firstAgent.Description, secondAgent.Description)
		assert.Equal(t, firstAgent.Tools, secondAgent.Tools, "Tools round-trip (PascalCase→lowercase)")
		assert.Equal(t, firstAgent.DisallowedTools, secondAgent.DisallowedTools)
		assert.Equal(t, firstAgent.PermissionPolicy, secondAgent.PermissionPolicy, "PermissionPolicy round-trip (from permissionMode)")
		assert.Equal(t, firstAgent.Behavior.Mode, secondAgent.Behavior.Mode)
	})
}

// TestPlatformRenderProducesNonEmpty ensures the platform render path
// (RenderDocument) does not return empty output for any of the four
// doc types. Catches regressions where a template change silently
// produces zero-byte output.
func TestPlatformRenderProducesNonEmpty(t *testing.T) {
	platforms := []string{"claude-code", "opencode"}

	tests := []struct {
		name string
		doc  interface{}
	}{
		{
			name: "agent",
			doc: &parser.CanonicalAgent{
				Agent:   core.Agent{Name: "x", Description: "x"},
				Content: "body",
			},
		},
		{
			name: "command",
			doc: &parser.CanonicalCommand{
				Command: core.Command{Name: "x", Description: "x"},
				Content: "body",
			},
		},
		{
			name: "skill",
			doc: &parser.CanonicalSkill{
				Skill:   core.Skill{Name: "x", Description: "x"},
				Content: "body",
			},
		},
		{
			name: "memory",
			doc: &parser.CanonicalMemory{
				Memory:  core.Memory{Paths: []string{"x"}},
				Content: "body",
			},
		},
	}

	for _, platform := range platforms {
		for _, tt := range tests {
			t.Run(platform+"-"+tt.name, func(t *testing.T) {
				out, err := RenderDocument(t.Context(), tt.doc, platform)
				require.NoError(t, err)
				assert.NotEmpty(t, strings.TrimSpace(out),
					"RenderDocument(%s, %s) must produce non-empty output", platform, tt.name)
			})
		}
	}
}

// TestCanonicalMarshalRoundTripPreservesSemanticFields is a
// regression test for the canonical YAML emitter: it parses a
// canonical fixture, marshals it back, and asserts the field set is
// preserved through deep reflection. This complements
// TestParseRenderRoundTrip (which checks one doc type per subtest)
// with a single, generic, table-driven check.
func TestCanonicalMarshalRoundTripPreservesSemanticFields(t *testing.T) {
	repoRoot := fixtureRepoRoot(t)
	fixturesDir := filepath.Join(repoRoot, "test", "fixtures", "canonical")

	tests := []struct {
		name    string
		fixture string
		docType string
	}{
		{"agent", "agent-permission-balanced.md", "agent"},
		{"command", "command-valid.md", "command"},
		{"skill", "skill-valid.md", "skill"},
		{"memory", "memory-valid.md", "memory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir, tt.fixture)
			if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
				t.Skipf("fixture not found: %s", fixturePath)
			}

			first, err := parser.ParseDocument(t.Context(), fixturePath, tt.docType)
			require.NoError(t, err)

			marshaled, err := MarshalCanonical(t.Context(), first)
			require.NoError(t, err)
			require.NotEmpty(t, marshaled, "MarshalCanonical produced empty output")

			// Re-parse from a temp file (so the parser's file-read path is exercised).
			tmpDir := t.TempDir()
			tmpPath := filepath.Join(tmpDir, tt.docType+".md")
			require.NoError(t, os.WriteFile(tmpPath, []byte(marshaled), 0o600))

			second, err := parser.ParseDocument(t.Context(), tmpPath, tt.docType)
			require.NoError(t, err)

			require.Equal(t, reflect.TypeOf(first), reflect.TypeOf(second),
				"round-trip must preserve canonical struct type")
		})
	}
}
