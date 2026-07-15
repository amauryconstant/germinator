package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

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
// XDG-resolved default (config.GetConfigPath) is used. The actual
// file scaffolding (parent-directory creation + write with 0600
// permissions + force-precondition check) is delegated to
// config.WriteDefault, which returns *config.WriteError for I/O
// failures and the "already exists" precondition violation. The
// cmdutil.ExitCodeFor dispatch maps *WriteError to ExitCodeError (1).
func runConfigInit(opts *configInitOptions) error {
	path := opts.OutputPath
	if path == "" {
		defaultPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("determining default config path: %w", err)
		}
		path = defaultPath
	}

	if err := config.WriteDefault(path, opts.Force); err != nil {
		return err //nolint:wrapcheck // typed *config.WriteError chain traversal (Phase 4.2)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "Successfully created config file: %s\n", path)
	return nil
}
