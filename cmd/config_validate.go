package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// configValidateOptions holds the runtime state for a `config validate`
// invocation. IO and Ctx come from the Factory; OutputPath comes from
// the parsed --output-path flag (resolved to the XDG default if empty
// inside runConfigValidate).
type configValidateOptions struct {
	IO         *iostreams.IOStreams
	Ctx        context.Context
	OutputPath string
}

// NewCmdConfigValidate creates the `config validate` command via the
// canonical NewCmdXxx(f, runF) pattern. Migrated in slice 8.
//
// Flags:
//
//	--output-path  (optional) config file to validate; defaults to config.GetConfigPath()
//
// No --output format flag: this command produces text output (a single
// success line or a returned typed error), not structured data. The
// legacy --output / -o short flag is dropped with the rename to
// --output-path; invoking --output now yields a Cobra usage error.
func NewCmdConfigValidate(f *cmdutil.Factory, runF func(*configValidateOptions) error) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an existing config file",
		Long: `Validate a germinator configuration file.

Checks that the config file:
  - Exists at the specified path
  - Contains valid TOML syntax
  - Has valid field values (e.g., known platform names)

If no --output-path is specified, the default config location is used:
  ~/.config/germinator/config.toml`,
		Example: `  # Validate config at default location
  germinator config validate

  # Validate config at custom location
  germinator config validate --output-path /path/to/config.toml`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &configValidateOptions{
				IO:         f.IOStreams,
				Ctx:        c.Context(),
				OutputPath: outputPath,
			}
			if runF != nil {
				return runF(opts)
			}
			return runConfigValidate(opts)
		},
	}

	cmd.Flags().StringVar(&outputPath, "output-path", "",
		"Config file to validate (default: ~/.config/germinator/config.toml)")

	return cmd
}

// runConfigValidate executes the validation logic. It reads the config
// file, parses it with koanf, and delegates field-level validation to
// config.Validate (platform-only today). Per the single-handling rule,
// errors are RETURNED — main.go renders them once via
// output.FormatError; this function never calls output.FormatError
// itself.
func runConfigValidate(opts *configValidateOptions) error {
	if opts.OutputPath == "" {
		defaultPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("determining default config path: %w", err)
		}
		opts.OutputPath = defaultPath
	}

	if _, err := os.Stat(opts.OutputPath); os.IsNotExist(err) {
		return core.NewFileError(opts.OutputPath, "read",
			"config file not found", err)
	}

	cfgObj := config.DefaultConfig()
	k := koanf.New(".")
	if err := k.Load(file.Provider(opts.OutputPath), toml.Parser()); err != nil {
		return core.NewFileError(opts.OutputPath, "parse",
			"failed to parse config file", err)
	}
	if err := k.Unmarshal("", cfgObj); err != nil {
		return core.NewParseError(opts.OutputPath, "failed to parse config", err)
	}
	if err := cfgObj.Validate(); err != nil {
		return fmt.Errorf("validating config: %w", err)
	}

	_, _ = fmt.Fprintf(opts.IO.Out, "Config file is valid: %s\n", opts.OutputPath)
	return nil
}
