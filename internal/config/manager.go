package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
)

// ConfigManager defines the interface for configuration management.
type ConfigManager interface {
	// Load reads and parses the configuration file.
	// Returns an error if the file exists but cannot be read or parsed.
	// Missing config file is not an error - defaults are used.
	Load() error

	// GetConfig returns the loaded configuration.
	// Returns nil if Load() has not been called.
	GetConfig() *Config
}

// koanfConfigManager implements ConfigManager using Koanf.
type koanfConfigManager struct {
	config *Config
}

// NewConfigManager creates a new ConfigManager.
func NewConfigManager() ConfigManager {
	return &koanfConfigManager{
		config: DefaultConfig(),
	}
}

// Load reads and parses the configuration file from the XDG-compliant location.
// Missing config file is not an error - defaults are used.
func (m *koanfConfigManager) Load() error {
	configPath, err := resolveConfigPath()
	if err != nil {
		return err
	}

	// If no config file found, use defaults (not an error)
	if configPath == "" {
		return nil
	}

	k := koanf.New(".")

	// Load the config file
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return gerrors.NewFileError(configPath, "reading", "failed to read config file", err)
	}

	// Unmarshal into config struct
	if err := k.Unmarshal("", m.config); err != nil {
		return gerrors.NewParseError(configPath, "failed to parse config", err)
	}

	// Validate the config
	if err := m.config.Validate(); err != nil {
		return err
	}

	// Expand paths (e.g., ~ to home directory)
	if err := m.config.ExpandPaths(); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the loaded configuration.
func (m *koanfConfigManager) GetConfig() *Config {
	return m.config
}

// resolveConfigPath returns the path to the config file.
// It checks locations in order of precedence:
// 1. $XDG_CONFIG_HOME/germinator/config.toml
// 2. $HOME/.config/germinator/config.toml
// 3. ./config.toml (current working directory)
// Returns empty string if no config file is found.
func resolveConfigPath() (string, error) {
	candidates := []string{}

	// 1. XDG_CONFIG_HOME
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		candidates = append(candidates, filepath.Join(xdgConfigHome, "germinator", "config.toml"))
	}

	// 2. HOME/.config (standard XDG fallback)
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".config", "germinator", "config.toml"))
	}

	// 3. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "config.toml"))
	}

	// Find the first existing file
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// No config file found - return empty string (use defaults)
	return "", nil
}

// GetConfigPath returns the path where the config file would be located.
// This is useful for displaying messages about config location.
// It returns the XDG path even if the file doesn't exist.
func GetConfigPath() (string, error) {
	// Prefer XDG_CONFIG_HOME if set
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "germinator", "config.toml"), nil
	}

	// Fall back to HOME/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "germinator", "config.toml"), nil
}
