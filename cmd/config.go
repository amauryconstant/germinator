package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carapace-sh/carapace"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/infrastructure/config"
)

// scaffoldedConfig is the default config file content with explanatory comments.
// All settings are commented out by default, requiring users to explicitly uncomment
// and configure only the settings they want to override.
const scaffoldedConfig = `# Germinator configuration
# https://github.com/anomalyco/germinator
#
# This file configures germinator's global behavior.
# All settings are optional - defaults are used if omitted.
# Settings below are commented out; uncomment and customize as needed.

# Path to your library directory containing skills, agents, commands, and presets.
# The library must contain a library.yaml index file.
# Supports ~ expansion for home directory.
# Default: ~/.local/share/germinator/library (or $XDG_DATA_HOME/germinator/library if set)
# library = "~/.local/share/germinator/library"

# Default platform when --platform is not specified.
# Options: "opencode" (default), "claude-code"
# Leave empty to require --platform on every command.
# Default: "" (none)
# platform = ""

# Shell completion configuration
[completion]

# Maximum time to wait for library loading during completion suggestions.
# Lower values = faster but may timeout on large libraries.
# Default: "500ms"
# timeout = "500ms"

# How long to cache library data for completion performance.
# Higher values = faster completions but may show stale results.
# Default: "5s"
# cache_ttl = "5s"
`

// NewConfigCommand creates the config command group with init and validate subcommands.
func NewConfigCommand(cfg *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Scaffold and validate germinator configuration files",
		Long: `Commands for working with germinator configuration files:

  germinator config init      Scaffold a new config file with documented fields
  germinator config validate   Validate an existing config file`,
	}

	cmd.AddCommand(NewConfigInitCommand(cfg))
	cmd.AddCommand(NewConfigValidateCommand(cfg))

	return cmd
}

// NewConfigInitCommand creates the config init command for scaffolding a new config file.
func NewConfigInitCommand(cfg *CommandConfig) *cobra.Command {
	var (
		outputPath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a new config file with explanatory comments",
		Long: `Create a new germinator config file at the specified location.

If no --output path is specified, the default config location is used:
  ~/.config/germinator/config.toml (or $XDG_CONFIG_HOME/germinator/config.toml)

Use --force to overwrite an existing config file.`,
		Example: `  # Create config at default location
  germinator config init

  # Create config at custom location
  germinator config init --output /path/to/config.toml

  # Overwrite existing config
  germinator config init --force`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Default to standard config path if not specified
			if outputPath == "" {
				defaultPath, err := config.GetConfigPath()
				if err != nil {
					return fmt.Errorf("determining default config path: %w", err)
				}
				outputPath = defaultPath
			}

			// Check if file already exists
			if !force {
				if _, err := os.Stat(outputPath); err == nil {
					return gerrors.NewFileError(outputPath, "creating",
						"config file already exists, use --force to overwrite", nil)
				}
			}

			// Create parent directories if needed
			dir := filepath.Dir(outputPath)
			if err := os.MkdirAll(dir, 0750); err != nil {
				return gerrors.NewFileError(dir, "creating directory",
					"failed to create parent directories", err)
			}

			// Write the scaffolded config file
			if err := os.WriteFile(outputPath, []byte(scaffoldedConfig), 0600); err != nil {
				return gerrors.NewFileError(outputPath, "writing",
					"failed to write config file", err)
			}

			VerbosePrint(cfg, "Created config file: %s", outputPath)
			_, _ = fmt.Fprintf(c.OutOrStdout(), "Successfully created config file: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: ~/.config/germinator/config.toml)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing file without prompting")

	carapace.Gen(cmd)

	return cmd
}

// NewConfigValidateCommand creates the config validate command for validating an existing config file.
func NewConfigValidateCommand(cfg *CommandConfig) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an existing config file",
		Long: `Validate a germinator configuration file.

Checks that the config file:
  - Exists at the specified path
  - Contains valid TOML syntax
  - Has valid field values (e.g., known platform names)

If no --output path is specified, the default config location is used:
  ~/.config/germinator/config.toml`,
		Example: `  # Validate config at default location
  germinator config validate

  # Validate config at custom location
  germinator config validate --output /path/to/config.toml`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			// Default to standard config path if not specified
			if outputPath == "" {
				defaultPath, err := config.GetConfigPath()
				if err != nil {
					return fmt.Errorf("determining default config path: %w", err)
				}
				outputPath = defaultPath
			}

			// Check if file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				return gerrors.NewFileError(outputPath, "validating",
					"config file not found", err)
			}

			// Load and validate the config using koanf directly
			cfgObj := config.DefaultConfig()
			k := koanf.New(".")
			if err := k.Load(file.Provider(outputPath), toml.Parser()); err != nil {
				return gerrors.NewFileError(outputPath, "parsing",
					"failed to parse config file", err)
			}
			if err := k.Unmarshal("", cfgObj); err != nil {
				return gerrors.NewParseError(outputPath, "failed to parse config", err)
			}
			if err := cfgObj.Validate(); err != nil {
				return fmt.Errorf("validating config: %w", err)
			}

			_, _ = fmt.Fprintf(c.OutOrStdout(), "Config file is valid: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Config file to validate (default: ~/.config/germinator/config.toml)")

	carapace.Gen(cmd)

	return cmd
}
