package domain

// AgentBehavior defines agent execution behavior settings.
type AgentBehavior struct {
	Mode        string   `yaml:"mode,omitempty" json:"mode,omitempty"`
	Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	Steps       int      `yaml:"steps,omitempty" json:"steps,omitempty"`
	Prompt      string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty" json:"hidden,omitempty"`
	Disabled    bool     `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

// AgentExtensions defines agent extensions and hooks.
type AgentExtensions struct {
	Hooks map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// Agent represents an AI agent configuration.
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
