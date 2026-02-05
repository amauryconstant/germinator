package canonical

import (
	"errors"
	"fmt"
	"regexp"
)

type PermissionPolicy string

const (
	PermissionPolicyRestrictive  PermissionPolicy = "restrictive"
	PermissionPolicyBalanced     PermissionPolicy = "balanced"
	PermissionPolicyPermissive   PermissionPolicy = "permissive"
	PermissionPolicyAnalysis     PermissionPolicy = "analysis"
	PermissionPolicyUnrestricted PermissionPolicy = "unrestricted"
)

func (p PermissionPolicy) IsValid() bool {
	switch p {
	case PermissionPolicyRestrictive, PermissionPolicyBalanced, PermissionPolicyPermissive,
		PermissionPolicyAnalysis, PermissionPolicyUnrestricted:
		return true
	default:
		return false
	}
}

type AgentBehavior struct {
	Mode        string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	Steps       int      `yaml:"steps,omitempty" json:"steps,omitempty"`
	Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty" json:"hidden,omitempty"`
	Disabled    bool     `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

func (b *AgentBehavior) Validate() []error {
	var errs []error

	if b.Mode != "" {
		validModes := map[string]bool{
			"primary":  true,
			"subagent": true,
			"all":      true,
		}
		if !validModes[b.Mode] {
			errs = append(errs, fmt.Errorf("behavior.mode must be one of: primary, subagent, all (got: %s)", b.Mode))
		}
	}

	if b.Temperature != nil {
		if *b.Temperature < 0.0 || *b.Temperature > 1.0 {
			errs = append(errs, fmt.Errorf("behavior.temperature must be between 0.0 and 1.0 (got: %f)", *b.Temperature))
		}
	}

	if b.Steps < 0 {
		errs = append(errs, fmt.Errorf("behavior.steps must be >= 0 (got: %d)", b.Steps))
	}

	return errs
}

type AgentExtensions struct {
	Hooks map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

func (e *AgentExtensions) Validate() []error {
	return nil
}

type PlatformConfig map[string]map[string]interface{}

type Agent struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	Tools            []string         `yaml:"tools,omitempty" json:"tools,omitempty"`
	DisallowedTools  []string         `yaml:"disallowedTools,omitempty" json:"disallowedTools,omitempty"`
	PermissionPolicy PermissionPolicy `yaml:"permissionPolicy,omitempty" json:"permissionPolicy,omitempty"`
	Behavior         AgentBehavior    `yaml:"behavior,omitempty" json:"behavior,omitempty"`
	Targets          PlatformConfig   `yaml:"targets,omitempty" json:"targets,omitempty"`
	Extensions       AgentExtensions  `yaml:"extensions,omitempty" json:"extensions,omitempty"`

	Model string `yaml:"model,omitempty" json:"model,omitempty"`
}

func (a *Agent) Validate() []error {
	var errs []error

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

	if a.PermissionPolicy != "" && !a.PermissionPolicy.IsValid() {
		errs = append(errs, fmt.Errorf("permissionPolicy must be one of: restrictive, balanced, permissive, analysis, unrestricted (got: %s)", a.PermissionPolicy))
	}

	errs = append(errs, a.Behavior.Validate()...)
	errs = append(errs, a.Extensions.Validate()...)

	return errs
}

type CommandExecution struct {
	Context string `yaml:"context,omitempty" json:"context,omitempty"`
	Subtask bool   `yaml:"subtask,omitempty" json:"subtask,omitempty"`
	Agent   string `yaml:"agent,omitempty" json:"agent,omitempty"`
}

func (e *CommandExecution) Validate() []error {
	var errs []error

	if e.Context != "" && e.Context != "fork" {
		errs = append(errs, fmt.Errorf("execution.context must be 'fork' if specified (got: %s)", e.Context))
	}

	return errs
}

type CommandArguments struct {
	Hint string `yaml:"hint,omitempty" json:"hint,omitempty"`
}

type Command struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	Tools     []string         `yaml:"tools,omitempty" json:"tools,omitempty"`
	Execution CommandExecution `yaml:"execution,omitempty" json:"execution,omitempty"`
	Arguments CommandArguments `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Targets   PlatformConfig   `yaml:"targets,omitempty" json:"targets,omitempty"`

	Model string `yaml:"model,omitempty" json:"model,omitempty"`
}

func (c *Command) Validate() []error {
	var errs []error

	if c.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if c.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}

	errs = append(errs, c.Execution.Validate()...)

	return errs
}

type Memory struct {
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Content  string   `yaml:"content,omitempty" json:"content,omitempty"`
	FilePath string   `yaml:"-" json:"-"`
}

func (m *Memory) Validate() []error {
	var errs []error

	if len(m.Paths) == 0 && m.Content == "" {
		errs = append(errs, errors.New("paths or content is required"))
	}

	return errs
}

type SkillExtensions struct {
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Hooks         map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

func (e *SkillExtensions) Validate() []error {
	return nil
}

type SkillExecution struct {
	Context       string `yaml:"context,omitempty" json:"context,omitempty"`
	Agent         string `yaml:"agent,omitempty" json:"agent,omitempty"`
	UserInvocable bool   `yaml:"userInvocable,omitempty" json:"userInvocable,omitempty"`
}

func (e *SkillExecution) Validate() []error {
	var errs []error

	if e.Context != "" && e.Context != "fork" {
		errs = append(errs, fmt.Errorf("execution.context must be 'fork' if specified (got: %s)", e.Context))
	}

	return errs
}

type Skill struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Content     string `yaml:"-" json:"-"`
	FilePath    string `yaml:"-" json:"-"`

	Tools      []string        `yaml:"tools,omitempty" json:"tools,omitempty"`
	Extensions SkillExtensions `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Execution  SkillExecution  `yaml:"execution,omitempty" json:"execution,omitempty"`
	Targets    PlatformConfig  `yaml:"targets,omitempty" json:"targets,omitempty"`

	Model string `yaml:"model,omitempty" json:"model,omitempty"`
}

func (s *Skill) Validate() []error {
	var errs []error

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

	errs = append(errs, s.Extensions.Validate()...)
	errs = append(errs, s.Execution.Validate()...)

	return errs
}
