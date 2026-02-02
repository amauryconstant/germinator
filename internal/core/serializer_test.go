package core

import (
	"strings"
	"testing"

	"gitlab.com/amoconst/germinator/internal/models"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func TestRenderDocumentAgent(t *testing.T) {
	agent := &models.Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Tools:       []string{"editor", "bash"},
		Model:       "sonnet",
		Content:     "This is test content\nWith multiple lines",
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
	if !strings.Contains(result, "  - editor") {
		t.Error("Expected editor tool in output")
	}
	if !strings.Contains(result, "This is test content") {
		t.Error("Expected content in output")
	}
	if !strings.Contains(result, "---") {
		t.Error("Expected frontmatter delimiters in output")
	}
}

func TestRenderDocumentCommand(t *testing.T) {
	command := &models.Command{
		Name:         "test-command",
		Description:  "Test command",
		AllowedTools: []string{"bash"},
		Content:      "Command content here",
	}

	result, err := RenderDocument(command, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "description: Test command") {
		t.Error("Expected description field in output")
	}
	if !strings.Contains(result, "allowed-tools:") {
		t.Error("Expected allowed-tools section in output")
	}
	if !strings.Contains(result, "Command content here") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentSkill(t *testing.T) {
	skill := &models.Skill{
		Name:        "test-skill",
		Description: "Test skill description",
		Model:       "haiku",
		Content:     "Skill instructions",
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
	memory := &models.Memory{
		Paths:   []string{"src/**/*.go", "README.md"},
		Content: "Memory content\nWith multiple lines",
	}

	result, err := RenderDocument(memory, "claude-code")
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	if !strings.Contains(result, "paths:") {
		t.Error("Expected paths section in output")
	}
	if !strings.Contains(result, "  - src/**/*.go") {
		t.Error("Expected path in output")
	}
	if !strings.Contains(result, "Memory content") {
		t.Error("Expected content in output")
	}
}

func TestRenderDocumentUnknownType(t *testing.T) {
	unknownDoc := "not a document"

	_, err := RenderDocument(unknownDoc, "claude-code")
	if err == nil {
		t.Error("Expected error for unknown document type")
	}
	if !strings.Contains(err.Error(), "unknown document type") {
		t.Errorf("Expected 'unknown document type' error, got: %v", err)
	}
}

func TestRenderDocumentInvalidTemplate(t *testing.T) {
	agent := &models.Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Content:     "Content",
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
		{"simple text", "Simple text content"},
		{"multiline", "Line 1\nLine 2\nLine 3"},
		{"markdown formatting", "# Header\n\nSome **bold** text"},
		{"code blocks", "```go\nfunc test() {}\n```"},
		{"empty content", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &models.Agent{
				Name:        "test-agent",
				Description: "Test",
				Content:     tt.content,
			}

			result, err := RenderDocument(agent, "claude-code")
			if err != nil {
				t.Fatalf("RenderDocument failed: %v", err)
			}

			if !strings.Contains(result, tt.content) {
				t.Errorf("Expected content %q not found in output", tt.content)
			}

			parts := strings.Split(result, "---")
			if len(parts) < 3 {
				t.Error("Expected frontmatter with --- delimiters")
			}

			body := strings.TrimSpace(parts[len(parts)-1])
			if body != tt.content && tt.content != "" {
				t.Errorf("Expected body %q, got %q", tt.content, body)
			}
		})
	}
}

func TestRenderOpenCodeAgent(t *testing.T) {
	tests := []struct {
		name    string
		agent   *models.Agent
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "minimal agent",
			agent: &models.Agent{
				Name:           "code-reviewer",
				Description:    "Reviews code for quality and best practices",
				Tools:          []string{"read", "grep", "glob"},
				Model:          "anthropic/claude-sonnet-4-20250514",
				PermissionMode: "default",
				Content:        "You are a code reviewer ensuring high standards of code quality and security.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "name: code-reviewer") {
					t.Error("Name should not be in frontmatter (derived from filename)")
				}
				if !strings.Contains(result, "description: Reviews code for quality and best practices") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "read: true") {
					t.Error("Expected read tool in output")
				}
				if !strings.Contains(result, "grep: true") {
					t.Error("Expected grep tool in output")
				}
				if !strings.Contains(result, "glob: true") {
					t.Error("Expected glob tool in output")
				}
				if !strings.Contains(result, "permission:") {
					t.Error("Expected permission section in output")
				}
				if !strings.Contains(result, "edit:") {
					t.Error("Expected edit permission in output")
				}
				if !strings.Contains(result, "bash:") {
					t.Error("Expected bash permission in output")
				}
				if strings.Contains(result, "mode:") {
					t.Error("Expected mode to be omitted when empty")
				}
				if !strings.Contains(result, "model: anthropic/claude-sonnet-4-20250514") {
					t.Error("Expected model field in output")
				}
				if !strings.Contains(result, "You are a code reviewer") {
					t.Error("Expected content in output")
				}
			},
		},
		{
			name: "full agent with all OpenCode fields",
			agent: &models.Agent{
				Name:            "code-reviewer-full",
				Description:     "A comprehensive code reviewer with all configuration options",
				Tools:           []string{"read", "write", "bash", "grep", "edit"},
				DisallowedTools: []string{"dangerous-command"},
				Model:           "anthropic/claude-sonnet-4-20250514",
				PermissionMode:  "acceptEdits",
				Mode:            "primary",
				Temperature:     float64Ptr(0.1),
				MaxSteps:        50,
				Hidden:          false,
				Prompt:          "You are an expert code reviewer.",
				Disable:         false,
				Content:         "You are an expert code reviewer specializing in security, performance, and best practices.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if strings.Contains(result, "name: code-reviewer-full") {
					t.Error("Name should not be in frontmatter (derived from filename)")
				}
				if !strings.Contains(result, "read: true") {
					t.Error("Expected read tool in output")
				}
				if !strings.Contains(result, "write: true") {
					t.Error("Expected write tool in output")
				}
				if !strings.Contains(result, "bash: true") {
					t.Error("Expected bash tool in output")
				}
				if !strings.Contains(result, "dangerous-command: false") {
					t.Error("Expected disallowed tool in output")
				}
				if !strings.Contains(result, "permission:") {
					t.Error("Expected permission section in output")
				}
				if !strings.Contains(result, "mode: primary") {
					t.Error("Expected mode field in output")
				}
				if !strings.Contains(result, "temperature: 0.1") {
					t.Error("Expected temperature field in output")
				}
				if !strings.Contains(result, "maxSteps: 50") {
					t.Error("Expected maxSteps field in output")
				}
				if !strings.Contains(result, "prompt:") {
					t.Errorf("Expected prompt field in output")
				}
				if strings.Contains(result, "hidden:") {
					t.Error("Expected hidden field to be omitted when false")
				}
				if strings.Contains(result, "disable:") {
					t.Error("Expected disable field to be omitted when false")
				}
			},
		},
		{
			name: "mixed tools (allowed and disallowed)",
			agent: &models.Agent{
				Name:            "code-reviewer-mixed",
				Description:     "Code reviewer with both allowed and disallowed tools",
				Tools:           []string{"read", "write", "bash"},
				DisallowedTools: []string{"dangerous-command", "system-config"},
				Model:           "anthropic/claude-sonnet-4-20250514",
				PermissionMode:  "dontAsk",
				Content:         "You are a code reviewer with restricted tool access.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "read: true") {
					t.Error("Expected read tool in output")
				}
				if !strings.Contains(result, "write: true") {
					t.Error("Expected write tool in output")
				}
				if !strings.Contains(result, "bash: true") {
					t.Error("Expected bash tool in output")
				}
				if !strings.Contains(result, "dangerous-command: false") {
					t.Error("Expected disallowed tool 'dangerous-command' in output")
				}
				if !strings.Contains(result, "system-config: false") {
					t.Error("Expected disallowed tool 'system-config' in output")
				}
			},
		},
		{
			name: "all permission modes",
			agent: &models.Agent{
				Name:           "test-agent",
				Description:    "Test agent",
				PermissionMode: "plan",
				Content:        "Test content",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "permission:") {
					t.Error("Expected permission section in output")
				}
				if !strings.Contains(result, "edit:") {
					t.Error("Expected edit permission in output")
				}
				if !strings.Contains(result, "bash:") {
					t.Error("Expected bash permission in output")
				}
			},
		},
		{
			name: "agent mode empty (omitted from output)",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Content:     "Test content",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "mode:") {
					t.Error("Expected mode to be omitted when empty")
				}
			},
		},
		{
			name: "agent mode explicit primary",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Mode:        "primary",
				Content:     "Test content",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "mode: primary") {
					t.Error("Expected mode to be 'primary'")
				}
			},
		},
		{
			name: "temperature nil (omitted from output)",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Content:     "Test content",
				Temperature: nil,
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "temperature:") {
					t.Error("Expected temperature field to be omitted when nil")
				}
			},
		},
		{
			name: "temperature 0.0 (rendered in output)",
			agent: &models.Agent{
				Name:        "test-agent",
				Description: "Test agent",
				Content:     "Test content",
				Temperature: float64Ptr(0.0),
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "temperature: 0") {
					t.Error("Expected temperature field to be rendered when set to 0.0")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.agent, "opencode")
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestRenderOpenCodeCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *models.Command
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "minimal command",
			command: &models.Command{
				Name:        "run-tests",
				Description: "Run all tests with coverage",
				Content:     "Run the full test suite with coverage reporting:\n\n```bash\nmise run test:coverage\n```\n\nThis will execute all tests and generate a coverage report.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if strings.Contains(result, "name: run-tests") {
					t.Error("Name should not be in frontmatter (derived from filename)")
				}
				if !strings.Contains(result, "description: Run all tests with coverage") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "Run the full test suite with coverage reporting") {
					t.Error("Expected content in output")
				}
				if strings.Contains(result, "agent:") {
					t.Error("Expected agent field to be omitted when empty")
				}
				if strings.Contains(result, "model:") {
					t.Error("Expected model field to be omitted when empty")
				}
				if strings.Contains(result, "subtask:") {
					t.Error("Expected subtask field to be omitted when false")
				}
			},
		},
		{
			name: "command with $ARGUMENTS placeholder",
			command: &models.Command{
				Name:        "search",
				Description: "Search code using ripgrep",
				Content:     "Search the codebase for the specified pattern:\n\n```bash\nrg $ARGUMENTS\n```\n\nYou can also use regular expressions.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "description: Search code using ripgrep") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "$ARGUMENTS") {
					t.Error("Expected $ARGUMENTS placeholder to be preserved in content")
				}
				if !strings.Contains(result, "```bash") {
					t.Error("Expected code block in content")
				}
			},
		},
		{
			name: "full command with all optional fields",
			command: &models.Command{
				Name:        "build",
				Description: "Build the project with specified configuration",
				Content:     "Build the project with the given configuration:\n\n```bash\nmise run build --target $1\n```\n\nThis will compile the project and generate the specified output.",
				Agent:       "build-agent",
				Model:       "anthropic/claude-sonnet-4-20250514",
				Subtask:     true,
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "description: Build the project with specified configuration") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "agent: build-agent") {
					t.Error("Expected agent field in output")
				}
				if !strings.Contains(result, "model: anthropic/claude-sonnet-4-20250514") {
					t.Error("Expected model field in output")
				}
				if !strings.Contains(result, "subtask: true") {
					t.Error("Expected subtask field in output")
				}
				if strings.Contains(result, "allowed-tools:") {
					t.Error("Expected allowed-tools to be omitted (Claude Code only)")
				}
				if strings.Contains(result, "argument-hint:") {
					t.Error("Expected argument-hint to be omitted (Claude Code only)")
				}
				if strings.Contains(result, "context:") {
					t.Error("Expected context to be omitted (Claude Code only)")
				}
				if strings.Contains(result, "disable-model-invocation:") {
					t.Error("Expected disable-model-invocation to be omitted (Claude Code only)")
				}
				if !strings.Contains(result, "$1") {
					t.Error("Expected $1 placeholder to be preserved in content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.command, "opencode")
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestRenderOpenCodeSkill(t *testing.T) {
	tests := []struct {
		name    string
		skill   *models.Skill
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "minimal skill",
			skill: &models.Skill{
				Name:        "git-release",
				Description: "Create consistent releases and changelogs",
				Content:     "## What I do\n- Draft release notes from merged PRs\n- Propose a version bump\n\n## When to use me\nUse this when you are preparing a tagged release.",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if !strings.Contains(result, "name: git-release") {
					t.Error("Expected name field in output")
				}
				if !strings.Contains(result, "description: Create consistent releases and changelogs") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "## What I do") {
					t.Error("Expected content in output")
				}
				if strings.Contains(result, "license:") {
					t.Error("Expected license field to be omitted when empty")
				}
				if strings.Contains(result, "compatibility:") {
					t.Error("Expected compatibility field to be omitted when empty")
				}
				if strings.Contains(result, "metadata:") {
					t.Error("Expected metadata field to be omitted when empty")
				}
				if strings.Contains(result, "hooks:") {
					t.Error("Expected hooks field to be omitted when empty")
				}
			},
		},
		{
			name: "skill with all OpenCode fields",
			skill: &models.Skill{
				Name:          "code-analyzer",
				Description:   "Analyze code quality and identify patterns",
				License:       "MIT",
				Compatibility: []string{"claude-code", "opencode"},
				Metadata: map[string]string{
					"version":    "1.0.0",
					"maintainer": "ops-team",
				},
				Hooks: map[string]string{
					"pre-run":  "validate-environment",
					"post-run": "cleanup-temp-files",
				},
				Content: "## What I do\nAnalyze code for:\n- Quality issues\n- Performance bottlenecks\n- Security vulnerabilities",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if !strings.Contains(result, "name: code-analyzer") {
					t.Error("Expected name field in output")
				}
				if !strings.Contains(result, "description: Analyze code quality and identify patterns") {
					t.Error("Expected description field in output")
				}
				if !strings.Contains(result, "license: MIT") {
					t.Error("Expected license field in output")
				}
				if !strings.Contains(result, "compatibility:") {
					t.Error("Expected compatibility section in output")
				}
				if !strings.Contains(result, "- claude-code") {
					t.Error("Expected claude-code in compatibility list")
				}
				if !strings.Contains(result, "- opencode") {
					t.Error("Expected opencode in compatibility list")
				}
				if !strings.Contains(result, "metadata:") {
					t.Error("Expected metadata section in output")
				}
				if !strings.Contains(result, "version: \"1.0.0\"") {
					t.Error("Expected version in metadata")
				}
				if !strings.Contains(result, "maintainer: \"ops-team\"") {
					t.Error("Expected maintainer in metadata")
				}
				if !strings.Contains(result, "hooks:") {
					t.Error("Expected hooks section in output")
				}
				if !strings.Contains(result, "pre-run: \"validate-environment\"") {
					t.Error("Expected pre-run hook in output")
				}
				if !strings.Contains(result, "post-run: \"cleanup-temp-files\"") {
					t.Error("Expected post-run hook in output")
				}
				if strings.Contains(result, "allowed-tools:") {
					t.Error("Expected allowed-tools to be omitted (Claude Code only)")
				}
				if strings.Contains(result, "user-invocable:") {
					t.Error("Expected user-invocable to be omitted (Claude Code only)")
				}
				if !strings.Contains(result, "## What I do") {
					t.Error("Expected content in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.skill, "opencode")
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestRenderOpenCodeMemory(t *testing.T) {
	tests := []struct {
		name    string
		memory  *models.Memory
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "paths-only memory",
			memory: &models.Memory{
				Paths: []string{
					"README.md",
					"AGENTS.md",
					"IMPLEMENTATION_PLAN.md",
				},
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if strings.Contains(result, "---") {
					t.Error("Expected no YAML frontmatter")
				}
				if !strings.Contains(result, "@README.md") {
					t.Error("Expected @README.md in output")
				}
				if !strings.Contains(result, "@AGENTS.md") {
					t.Error("Expected @AGENTS.md in output")
				}
				if !strings.Contains(result, "@IMPLEMENTATION_PLAN.md") {
					t.Error("Expected @IMPLEMENTATION_PLAN.md in output")
				}
			},
		},
		{
			name: "content-only memory",
			memory: &models.Memory{
				Content: "## Project Context\n\nThis is a configuration adapter for AI coding assistant documents.\n\n## Architecture\n\n- Go-based CLI tool\n- Template-based serialization\n- Platform-agnostic models",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if strings.Contains(result, "---") {
					t.Error("Expected no YAML frontmatter")
				}
				if !strings.Contains(result, "## Project Context") {
					t.Error("Expected project context in output")
				}
				if !strings.Contains(result, "This is a configuration adapter") {
					t.Error("Expected content in output")
				}
				if !strings.Contains(result, "## Architecture") {
					t.Error("Expected architecture section in output")
				}
				if strings.Contains(result, "@") {
					t.Error("Expected no @ file references when paths is empty")
				}
			},
		},
		{
			name: "memory with both paths and content",
			memory: &models.Memory{
				Paths: []string{
					"config/mise.toml",
					".mise/config.toml",
				},
				Content: "## Project Rules\n\nFollow these conventions:\n- Use mise task runner for all commands\n- Maintain Go standard layout\n- Test before committing",
			},
			wantErr: false,
			check: func(t *testing.T, result string) {
				t.Logf("Generated output:\n%s\n", result)
				if strings.Contains(result, "---") {
					t.Error("Expected no YAML frontmatter")
				}
				if !strings.Contains(result, "@config/mise.toml") {
					t.Error("Expected @config/mise.toml in output")
				}
				if !strings.Contains(result, "@.mise/config.toml") {
					t.Error("Expected @.mise/config.toml in output")
				}
				if !strings.Contains(result, "## Project Rules") {
					t.Error("Expected content in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderDocument(tt.memory, "opencode")
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestGetDocType(t *testing.T) {
	tests := []struct {
		name     string
		doc      interface{}
		expected string
		wantErr  bool
	}{
		{"agent", &models.Agent{}, "agent", false},
		{"command", &models.Command{}, "command", false},
		{"skill", &models.Skill{}, "skill", false},
		{"memory", &models.Memory{}, "memory", false},
		{"unknown", "string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDocType(tt.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDocType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getDocType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
