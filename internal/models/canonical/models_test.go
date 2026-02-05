package canonical

import (
	"strings"
	"testing"
)

func TestPermissionPolicyIsValid(t *testing.T) {
	tests := []struct {
		name     string
		policy   PermissionPolicy
		expected bool
	}{
		{"restrictive is valid", PermissionPolicyRestrictive, true},
		{"balanced is valid", PermissionPolicyBalanced, true},
		{"permissive is valid", PermissionPolicyPermissive, true},
		{"analysis is valid", PermissionPolicyAnalysis, true},
		{"unrestricted is valid", PermissionPolicyUnrestricted, true},
		{"invalid policy is not valid", "invalid-policy", false},
		{"empty string is not valid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.IsValid()
			if got != tt.expected {
				t.Errorf("PermissionPolicy.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAgentBehaviorValidate(t *testing.T) {
	tests := []struct {
		name       string
		behavior   AgentBehavior
		errorCount int
	}{
		{
			name: "valid behavior",
			behavior: AgentBehavior{
				Mode:        "primary",
				Temperature: ptr(0.5),
				Steps:       10,
				Prompt:      "You are helpful",
				Hidden:      false,
				Disabled:    false,
			},
			errorCount: 0,
		},
		{
			name: "invalid mode",
			behavior: AgentBehavior{
				Mode: "invalid-mode",
			},
			errorCount: 1,
		},
		{
			name: "temperature too low",
			behavior: AgentBehavior{
				Temperature: ptr(-0.1),
			},
			errorCount: 1,
		},
		{
			name: "temperature too high",
			behavior: AgentBehavior{
				Temperature: ptr(1.1),
			},
			errorCount: 1,
		},
		{
			name: "steps negative",
			behavior: AgentBehavior{
				Steps: -1,
			},
			errorCount: 1,
		},
		{
			name: "valid temperature boundary low",
			behavior: AgentBehavior{
				Temperature: ptr(0.0),
			},
			errorCount: 0,
		},
		{
			name: "valid temperature boundary high",
			behavior: AgentBehavior{
				Temperature: ptr(1.0),
			},
			errorCount: 0,
		},
		{
			name: "valid mode subagent",
			behavior: AgentBehavior{
				Mode: "subagent",
			},
			errorCount: 0,
		},
		{
			name: "valid mode all",
			behavior: AgentBehavior{
				Mode: "all",
			},
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.behavior.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("AgentBehavior.Validate() error count = %d, want %d", len(errs), tt.errorCount)
			}
		})
	}
}

func TestAgentValidate(t *testing.T) {
	tests := []struct {
		name       string
		agent      Agent
		errorCount int
	}{
		{
			name: "valid agent",
			agent: Agent{
				Name:             "test-agent",
				Description:      "A test agent",
				PermissionPolicy: PermissionPolicyBalanced,
			},
			errorCount: 0,
		},
		{
			name: "missing name",
			agent: Agent{
				Description: "A test agent",
			},
			errorCount: 1,
		},
		{
			name: "missing description",
			agent: Agent{
				Name: "test-agent",
			},
			errorCount: 1,
		},
		{
			name: "invalid name with uppercase",
			agent: Agent{
				Name:        "Test-Agent",
				Description: "A test agent",
			},
			errorCount: 1,
		},
		{
			name: "invalid name with consecutive hyphens",
			agent: Agent{
				Name:        "test--agent",
				Description: "A test agent",
			},
			errorCount: 1,
		},
		{
			name: "invalid name starting with hyphen",
			agent: Agent{
				Name:        "-test-agent",
				Description: "A test agent",
			},
			errorCount: 1,
		},
		{
			name: "invalid name ending with hyphen",
			agent: Agent{
				Name:        "test-agent-",
				Description: "A test agent",
			},
			errorCount: 1,
		},
		{
			name: "invalid permission policy",
			agent: Agent{
				Name:             "test-agent",
				Description:      "A test agent",
				PermissionPolicy: "invalid-policy",
			},
			errorCount: 1,
		},
		{
			name: "valid agent with behavior",
			agent: Agent{
				Name:        "test-agent",
				Description: "A test agent",
				Behavior: AgentBehavior{
					Mode:        "primary",
					Temperature: ptr(0.5),
					Steps:       10,
				},
			},
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.agent.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("Agent.Validate() error count = %d, want %d. Errors: %v", len(errs), tt.errorCount, errs)
			}
		})
	}
}

func TestCommandExecutionValidate(t *testing.T) {
	tests := []struct {
		name       string
		execution  CommandExecution
		errorCount int
	}{
		{
			name: "valid execution with fork",
			execution: CommandExecution{
				Context: "fork",
			},
			errorCount: 0,
		},
		{
			name:       "valid execution without context",
			execution:  CommandExecution{},
			errorCount: 0,
		},
		{
			name: "invalid context",
			execution: CommandExecution{
				Context: "invalid-context",
			},
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.execution.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("CommandExecution.Validate() error count = %d, want %d", len(errs), tt.errorCount)
			}
		})
	}
}

func TestCommandValidate(t *testing.T) {
	tests := []struct {
		name       string
		command    Command
		errorCount int
	}{
		{
			name: "valid command",
			command: Command{
				Name:        "test-command",
				Description: "A test command",
			},
			errorCount: 0,
		},
		{
			name: "missing name",
			command: Command{
				Description: "A test command",
			},
			errorCount: 1,
		},
		{
			name: "missing description",
			command: Command{
				Name: "test-command",
			},
			errorCount: 1,
		},
		{
			name: "valid command with execution",
			command: Command{
				Name:        "test-command",
				Description: "A test command",
				Execution: CommandExecution{
					Context: "fork",
				},
			},
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.command.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("Command.Validate() error count = %d, want %d", len(errs), tt.errorCount)
			}
		})
	}
}

func TestMemoryValidate(t *testing.T) {
	tests := []struct {
		name       string
		memory     Memory
		errorCount int
	}{
		{
			name: "valid memory with paths",
			memory: Memory{
				Paths: []string{"README.md", "CONTRIBUTING.md"},
			},
			errorCount: 0,
		},
		{
			name: "valid memory with content",
			memory: Memory{
				Content: "This is memory content",
			},
			errorCount: 0,
		},
		{
			name: "valid memory with both",
			memory: Memory{
				Paths:   []string{"README.md"},
				Content: "This is memory content",
			},
			errorCount: 0,
		},
		{
			name:       "invalid memory without paths or content",
			memory:     Memory{},
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.memory.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("Memory.Validate() error count = %d, want %d", len(errs), tt.errorCount)
			}
		})
	}
}

func TestSkillExecutionValidate(t *testing.T) {
	tests := []struct {
		name       string
		execution  SkillExecution
		errorCount int
	}{
		{
			name: "valid execution with fork",
			execution: SkillExecution{
				Context: "fork",
			},
			errorCount: 0,
		},
		{
			name:       "valid execution without context",
			execution:  SkillExecution{},
			errorCount: 0,
		},
		{
			name: "invalid context",
			execution: SkillExecution{
				Context: "invalid-context",
			},
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.execution.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("SkillExecution.Validate() error count = %d, want %d", len(errs), tt.errorCount)
			}
		})
	}
}

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		name       string
		skill      Skill
		errorCount int
	}{
		{
			name: "valid skill",
			skill: Skill{
				Name:        "test-skill",
				Description: "A test skill",
			},
			errorCount: 0,
		},
		{
			name: "missing name",
			skill: Skill{
				Description: "A test skill",
			},
			errorCount: 1,
		},
		{
			name: "missing description",
			skill: Skill{
				Name: "test-skill",
			},
			errorCount: 1,
		},
		{
			name: "name too short",
			skill: Skill{
				Name:        "",
				Description: "A test skill",
			},
			errorCount: 1,
		},
		{
			name: "name too long",
			skill: Skill{
				Name:        strings.Repeat("a", 65),
				Description: "A test skill",
			},
			errorCount: 1,
		},
		{
			name: "description too short",
			skill: Skill{
				Name:        "test-skill",
				Description: "",
			},
			errorCount: 1,
		},
		{
			name: "description too long",
			skill: Skill{
				Name:        "test-skill",
				Description: "a" + string(make([]byte, 1024)),
			},
			errorCount: 1,
		},
		{
			name: "invalid name with uppercase",
			skill: Skill{
				Name:        "Test-Skill",
				Description: "A test skill",
			},
			errorCount: 1,
		},
		{
			name: "valid skill with execution",
			skill: Skill{
				Name:        "test-skill",
				Description: "A test skill",
				Execution: SkillExecution{
					Context: "fork",
				},
			},
			errorCount: 0,
		},
		{
			name: "valid skill with extensions",
			skill: Skill{
				Name:        "test-skill",
				Description: "A test skill",
				Extensions: SkillExtensions{
					License: "MIT",
				},
			},
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.skill.Validate()
			if len(errs) != tt.errorCount {
				t.Errorf("Skill.Validate() error count = %d, want %d. Errors: %v", len(errs), tt.errorCount, errs)
			}
		})
	}
}

func TestAgentExtensionsValidate(t *testing.T) {
	extensions := AgentExtensions{
		Hooks: map[string]string{
			"pre-run":  "echo 'Starting'",
			"post-run": "echo 'Done'",
		},
	}
	errs := extensions.Validate()
	if len(errs) != 0 {
		t.Errorf("AgentExtensions.Validate() returned %d errors, want 0", len(errs))
	}
}

func TestSkillExtensionsValidate(t *testing.T) {
	extensions := SkillExtensions{
		License:       "MIT",
		Compatibility: []string{"claude-code", "opencode"},
		Metadata: map[string]string{
			"version": "1.0.0",
			"author":  "Test Author",
		},
		Hooks: map[string]string{
			"pre-review": "run-linters",
		},
	}
	errs := extensions.Validate()
	if len(errs) != 0 {
		t.Errorf("SkillExtensions.Validate() returned %d errors, want 0", len(errs))
	}
}

func ptr(f float64) *float64 {
	return &f
}
