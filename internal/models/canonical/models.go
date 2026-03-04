package canonical

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

type AgentExtensions struct {
	Hooks map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
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

type CommandExecution struct {
	Context string `yaml:"context,omitempty" json:"context,omitempty"`
	Subtask bool   `yaml:"subtask,omitempty" json:"subtask,omitempty"`
	Agent   string `yaml:"agent,omitempty" json:"agent,omitempty"`
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

type Memory struct {
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Content  string   `yaml:"content,omitempty" json:"content,omitempty"`
	FilePath string   `yaml:"-" json:"-"`
}

type SkillExtensions struct {
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Hooks         map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type SkillExecution struct {
	Context       string `yaml:"context,omitempty" json:"context,omitempty"`
	Agent         string `yaml:"agent,omitempty" json:"agent,omitempty"`
	UserInvocable bool   `yaml:"userInvocable,omitempty" json:"userInvocable,omitempty"`
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
