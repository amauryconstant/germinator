package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// scaffoldedConfig is the default config file content with
// explanatory comments. All settings are commented out by default,
// requiring users to explicitly uncomment and configure only the
// settings they want to override.
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

// configInitOptions holds the runtime state for a `config init`
// invocation. IO and Ctx come from the Factory; OutputPath and Force
// come from parsed flags.
type configInitOptions struct {
	IO         *iostreams.IOStreams
	Ctx        context.Context
	OutputPath string
	Force      bool
}

// NewCmdConfigInit creates the `config init` command via the canonical
// NewCmdXxx(f, runF) pattern. Migrated in slice 8.
//
// Flags:
//
//	--output-path  (optional) target file path; defaults to config.GetConfigPath()
//	--force        (optional) overwrite an existing config file
//
// The legacy --output / -o short flag is dropped with the rename to
// --output-path (the --output flag now belongs to the cli-output-formats
// capability); invoking --output now yields a Cobra usage error (exit 2).
// The -f short alias for --force is also dropped for consistency with
// the slice-7 library_init migration.
func NewCmdConfigInit(f *cmdutil.Factory, runF func(*configInitOptions) error) *cobra.Command {
	var (
		outputPath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a new config file with explanatory comments",
		Long: `Create a new germinator config file at the specified location.

If no --output-path is specified, the default config location is used:
  ~/.config/germinator/config.toml (or $XDG_CONFIG_HOME/germinator/config.toml)

Use --force to overwrite an existing config file.`,
		Example: `  # Create config at default location
  germinator config init

  # Create config at custom location
  germinator config init --output-path /path/to/config.toml

  # Overwrite existing config
  germinator config init --force`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &configInitOptions{
				IO:         f.IOStreams,
				Ctx:        c.Context(),
				OutputPath: outputPath,
				Force:      force,
			}
			if runF != nil {
				return runF(opts)
			}
			return runConfigInit(opts)
		},
	}

	cmd.Flags().StringVar(&outputPath, "output-path", "",
		"Output file path (default: ~/.config/germinator/config.toml)")
	cmd.Flags().BoolVar(&force, "force", false,
		"Overwrite existing file")

	return cmd
}

// runConfigInit executes the config creation logic. It is the
// production wiring for NewCmdConfigInit's runF parameter.
//
// Path resolution: opts.OutputPath is honored when set; otherwise the
// XDG-resolved default (config.GetConfigPath) is used. Parent
// directories are created (0750) if missing, and the file is written
// with 0600 permissions.
//
// Errors are wrapped via core.NewFileError so output.FormatError (in
// main.go) renders them once through the typed-error dispatcher.
func runConfigInit(opts *configInitOptions) error {
	path := opts.OutputPath
	if path == "" {
		defaultPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("determining default config path: %w", err)
		}
		path = defaultPath
	}

	if !opts.Force {
		if _, err := os.Stat(path); err == nil {
			return core.NewFileError(path, "create",
				"config file already exists (use --force to overwrite)", nil)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return core.NewFileError(dir, "create directory",
			"failed to create parent directories", err)
	}

	if err := os.WriteFile(path, []byte(scaffoldedConfig), 0o600); err != nil {
		return core.NewFileError(path, "write",
			"failed to write config file", err)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "Successfully created config file: %s\n", path)
	return nil
}
