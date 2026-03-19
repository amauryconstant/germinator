package domain

// CommandExecution defines command execution context settings.
type CommandExecution struct {
	Context string `yaml:"context,omitempty" json:"context,omitempty"`
	Subtask bool   `yaml:"subtask,omitempty" json:"subtask,omitempty"`
	Agent   string `yaml:"agent,omitempty" json:"agent,omitempty"`
}

// CommandArguments defines command argument hints.
type CommandArguments struct {
	Hint string `yaml:"hint,omitempty" json:"hint,omitempty"`
}

// Command represents a command configuration.
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
