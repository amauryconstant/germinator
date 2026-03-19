package domain

// Memory represents memory/context configuration.
type Memory struct {
	Paths    []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	Content  string   `yaml:"content,omitempty" json:"content,omitempty"`
	FilePath string   `yaml:"-" json:"-"`
}
