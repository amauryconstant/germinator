package models

import (
	"errors"
	"fmt"
	"regexp"
)

// Agent represents an AI agent configuration.
type Agent struct {
	Name            string   `yaml:"name"`
	Description     string   `yaml:"description"`
	Tools           []string `yaml:"tools"`
	DisallowedTools []string `yaml:"disallowedTools"`
	Model           string   `yaml:"model"`
	PermissionMode  string   `yaml:"permissionMode"`
	Skills          []string `yaml:"skills"`
	FilePath        string   `yaml:"-"`
	Content         string   `yaml:"-"`
}

// Validate checks if the agent configuration is valid.
func (a *Agent) Validate() []error {
	var errs []error

	if a.Name == "" {
		errs = append(errs, errors.New("name is required"))
	} else {
		matched, _ := regexp.MatchString(`^[a-z-]+$`, a.Name)
		if !matched {
			errs = append(errs, fmt.Errorf("name must be lowercase letters and hyphens only (got: %s)", a.Name))
		}
	}

	if a.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}

	if a.Model != "" {
		validModels := map[string]bool{
			"sonnet":  true,
			"opus":    true,
			"haiku":   true,
			"inherit": true,
		}
		if !validModels[a.Model] {
			errs = append(errs, fmt.Errorf("model must be one of: sonnet, opus, haiku, inherit (got: %s)", a.Model))
		}
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

	return errs
}

// Command represents an AI command configuration.
type Command struct {
	Name                   string   `yaml:"-"`
	AllowedTools           []string `yaml:"allowed-tools"`
	ArgumentHint           string   `yaml:"argument-hint"`
	Context                string   `yaml:"context"`
	Agent                  string   `yaml:"agent"`
	Description            string   `yaml:"description"`
	Model                  string   `yaml:"model"`
	DisableModelInvocation bool     `yaml:"disable-model-invocation"`
	FilePath               string   `yaml:"-"`
	Content                string   `yaml:"-"`
}

// Validate checks if the command configuration is valid.
func (c *Command) Validate() []error {
	var errs []error

	if c.Context != "" && c.Context != "fork" {
		errs = append(errs, fmt.Errorf("context must be 'fork' if specified (got: %s)", c.Context))
	}

	return errs
}

// Memory represents an AI memory configuration.
type Memory struct {
	Paths    []string `yaml:"paths"`
	FilePath string   `yaml:"-"`
	Content  string   `yaml:"-"`
}

// Validate checks if the memory configuration is valid.
func (m *Memory) Validate() []error {
	return nil
}

// Skill represents an AI skill configuration.
type Skill struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	AllowedTools  []string `yaml:"allowed-tools"`
	Model         string   `yaml:"model"`
	Context       string   `yaml:"context"`
	Agent         string   `yaml:"agent"`
	UserInvocable bool     `yaml:"user-invocable"`
	FilePath      string   `yaml:"-"`
	Content       string   `yaml:"-"`
}

// Validate checks if the skill configuration is valid.
func (s *Skill) Validate() []error {
	var errs []error

	if s.Name == "" {
		errs = append(errs, errors.New("name is required"))
	} else {
		if len(s.Name) > 64 {
			errs = append(errs, fmt.Errorf("name must not exceed 64 characters (got: %d)", len(s.Name)))
		}
		matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, s.Name)
		if !matched {
			errs = append(errs, fmt.Errorf("name must be lowercase letters, numbers, and hyphens only (got: %s)", s.Name))
		}
	}

	if s.Description == "" {
		errs = append(errs, errors.New("description is required"))
	} else {
		if len(s.Description) > 1024 {
			errs = append(errs, fmt.Errorf("description must not exceed 1024 characters (got: %d)", len(s.Description)))
		}
	}

	if s.Context != "" && s.Context != "fork" {
		errs = append(errs, fmt.Errorf("context must be 'fork' if specified (got: %s)", s.Context))
	}

	return errs
}
