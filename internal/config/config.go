// Package config provides configuration loading and management for germinator.
package config

import (
	"os"
	"path/filepath"

	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
)

// Config holds the application configuration.
type Config struct {
	// Library is the path to the library directory.
	Library string `koanf:"library"`

	// Platform is the default platform for transformations.
	// Empty string means platform must be specified via flag.
	Platform string `koanf:"platform"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Library:  "~/.config/germinator/library",
		Platform: "",
	}
}

// Validate checks that the configuration is valid.
// Returns an error if the platform value is invalid.
func (c *Config) Validate() error {
	if c.Platform == "" {
		return nil
	}

	if c.Platform != models.PlatformClaudeCode && c.Platform != models.PlatformOpenCode {
		return gerrors.NewConfigError(
			"platform",
			c.Platform,
			"unknown platform",
		).WithSuggestions([]string{models.PlatformClaudeCode, models.PlatformOpenCode})
	}

	return nil
}

// ExpandPaths expands the tilde (~) in paths to the user's home directory.
// This should be called after loading the config.
func (c *Config) ExpandPaths() error {
	expanded, err := expandTilde(c.Library)
	if err != nil {
		return err
	}
	c.Library = expanded
	return nil
}

// expandTilde expands a leading ~ in a path to the user's home directory.
func expandTilde(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", gerrors.NewConfigError("path", path, "cannot determine home directory")
		}
		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
