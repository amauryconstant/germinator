package models

import (
	"errors"
	"fmt"
	"regexp"
)

// Agent represents an AI agent configuration.
type Agent struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	Tools           []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	DisallowedTools []string `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`

	Mode        string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxSteps    int      `yaml:"maxSteps,omitempty" json:"maxSteps,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty" json:"hidden,omitempty"`
	Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Disable     bool     `yaml:"disable,omitempty" json:"disable,omitempty"`

	PermissionMode string   `yaml:"permissionMode,omitempty" json:"permissionMode,omitempty"`
	Skills         []string `yaml:"skills,omitempty" json:"skills,omitempty"`

	Model string `yaml:"model,omitempty" json:"model,omitempty"`
}

// Validate checks if agent configuration is valid.
func (a *Agent) Validate(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, errors.New("platform is required (available: claude-code, opencode)"))
	}

	if platform != "" && platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: claude-code, opencode)", platform))
	}

	if a.Name == "" {
		errs = append(errs, errors.New("name is required"))
	} else {
		matched, err := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, a.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to validate name regex: %w", err))
		} else if !matched {
			errs = append(errs, fmt.Errorf("name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$ (got: %s)", a.Name))
		}
	}

	if a.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}

	if a.PermissionMode != "" {
		validModes := map[string]bool{
			"default":           true,
			"acceptEdits":       true,
			"dontAsk":           true,
			"bypassPermissions": true,
			"plan":              true,
		}
		if !validModes[a.PermissionMode] {
			errs = append(errs, fmt.Errorf("permissionMode must be one of: default, acceptEdits, dontAsk, bypassPermissions, plan (got: %s)", a.PermissionMode))
		}
	}

	if platform == PlatformOpenCode {
		errs = append(errs, ValidateOpenCodeAgent(a)...)
	}

	return errs
}

// Command represents an AI command configuration.
type Command struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
	DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

	Subtask bool `yaml:"subtask,omitempty" json:"subtask,omitempty"`

	ArgumentHint           string `yaml:"argument-hint,omitempty" json:"argument-hint,omitempty"`
	Context                string `yaml:"context,omitempty" json:"context,omitempty"`
	Agent                  string `yaml:"agent,omitempty" json:"agent,omitempty"`
	DisableModelInvocation bool   `yaml:"disable-model-invocation,omitempty" json:"disable-model-invocation,omitempty"`

	Model string `yaml:"model,omitempty" json:"model,omitempty"`
}

// Validate checks if command configuration is valid.
func (c *Command) Validate(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, errors.New("platform is required (available: claude-code, opencode)"))
	}

	if platform != "" && platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: claude-code, opencode)", platform))
	}

	if c.Context != "" && c.Context != "fork" {
		errs = append(errs, fmt.Errorf("context must be 'fork' if specified (got: %s)", c.Context))
	}

	if platform == PlatformOpenCode {
		errs = append(errs, ValidateOpenCodeCommand(c)...)
	}

	return errs
}

// Memory represents an AI memory configuration.
type Memory struct {
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Content  string   `yaml:"content,omitempty" json:"content,omitempty"`
	FilePath string   `yaml:"-" json:"-"`
}

// Validate checks if memory configuration is valid.
func (m *Memory) Validate(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, errors.New("platform is required (available: claude-code, opencode)"))
	}

	if platform != "" && platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: claude-code, opencode)", platform))
	}

	if len(m.Paths) == 0 && m.Content == "" {
		errs = append(errs, errors.New("paths or content is required"))
	}

	if platform == PlatformOpenCode {
		errs = append(errs, ValidateOpenCodeMemory(m)...)
	}

	return errs
}

// Skill represents an AI skill configuration.
type Skill struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	AllowedTools    []string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
	DisallowedTools []string `yaml:"disallowed-tools,omitempty" json:"disallowed-tools,omitempty"`

	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Hooks         map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`

	Model         string `yaml:"model,omitempty" json:"model,omitempty"`
	Context       string `yaml:"context,omitempty" json:"context,omitempty"`
	Agent         string `yaml:"agent,omitempty" json:"agent,omitempty"`
	UserInvocable bool   `yaml:"user-invocable,omitempty" json:"user-invocable,omitempty"`
}

// Validate checks if skill configuration is valid.
func (s *Skill) Validate(platform string) []error {
	var errs []error

	if platform == "" {
		errs = append(errs, errors.New("platform is required (available: claude-code, opencode)"))
	}

	if platform != "" && platform != PlatformClaudeCode && platform != PlatformOpenCode {
		errs = append(errs, fmt.Errorf("unknown platform: %s (available: claude-code, opencode)", platform))
	}

	if s.Name == "" {
		errs = append(errs, errors.New("name is required"))
	} else {
		if len(s.Name) < 1 || len(s.Name) > 64 {
			errs = append(errs, fmt.Errorf("name must be 1-64 characters (got: %d)", len(s.Name)))
		}
		matched, err := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, s.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to validate name regex: %w", err))
		} else if !matched {
			errs = append(errs, fmt.Errorf("name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$ (got: %s)", s.Name))
		}
	}

	if s.Description == "" {
		errs = append(errs, errors.New("description is required"))
	} else {
		if len(s.Description) < 1 || len(s.Description) > 1024 {
			errs = append(errs, fmt.Errorf("description must be 1-1024 characters (got: %d)", len(s.Description)))
		}
	}

	if s.Context != "" && s.Context != "fork" {
		errs = append(errs, fmt.Errorf("context must be 'fork' if specified (got: %s)", s.Context))
	}

	if platform == PlatformOpenCode {
		errs = append(errs, ValidateOpenCodeSkill(s)...)
	}

	return errs
}

// ValidateOpenCodeAgent validates OpenCode-specific Agent constraints.
func ValidateOpenCodeAgent(agent *Agent) []error {
	var errs []error

	if agent.Mode != "" && agent.Mode != "primary" && agent.Mode != "subagent" && agent.Mode != "all" {
		errs = append(errs, fmt.Errorf("invalid mode: %s (valid values: primary, subagent, all)", agent.Mode))
	}

	if agent.Temperature != nil && (*agent.Temperature < 0.0 || *agent.Temperature > 1.0) {
		errs = append(errs, fmt.Errorf("temperature must be between 0.0 and 1.0, got %f", *agent.Temperature))
	}

	if agent.MaxSteps != 0 && agent.MaxSteps < 1 {
		errs = append(errs, fmt.Errorf("maxSteps must be >= 1, got %d", agent.MaxSteps))
	}

	return errs
}

// ValidateOpenCodeCommand validates OpenCode-specific Command constraints.
func ValidateOpenCodeCommand(cmd *Command) []error {
	var errs []error

	if cmd.Content == "" {
		errs = append(errs, fmt.Errorf("template (content) is required"))
	}

	return errs
}

// ValidateOpenCodeSkill validates OpenCode-specific Skill constraints.
func ValidateOpenCodeSkill(skill *Skill) []error {
	var errs []error

	skillNameRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !skillNameRegex.MatchString(skill.Name) {
		errs = append(errs, fmt.Errorf("skill name must match pattern ^[a-z0-9]+(-[a-z0-9]+)*$, got %s", skill.Name))
	}

	if skill.Content == "" {
		errs = append(errs, fmt.Errorf("content is required"))
	}

	return errs
}

// ValidateOpenCodeMemory validates OpenCode-specific Memory constraints.
func ValidateOpenCodeMemory(mem *Memory) []error {
	var errs []error

	if len(mem.Paths) == 0 && mem.Content == "" {
		errs = append(errs, fmt.Errorf("paths or content is required"))
	}

	return errs
}
