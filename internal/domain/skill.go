package domain

// SkillExtensions defines skill metadata and extensions.
type SkillExtensions struct {
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility []string          `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Hooks         map[string]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// SkillExecution defines skill execution context settings.
type SkillExecution struct {
	Context       string `yaml:"context,omitempty" json:"context,omitempty"`
	Agent         string `yaml:"agent,omitempty" json:"agent,omitempty"`
	UserInvocable bool   `yaml:"userInvocable,omitempty" json:"userInvocable,omitempty"`
}

// Skill represents a skill configuration.
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
