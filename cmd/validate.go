package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	gerrors "gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/models"
)

// NewValidateCommand creates the validate command with dependency injection.
// Non-migrated command: reads services from bridge (transitional; converted
// to the NewCmdValidate(f, runF) pattern in slice 3).
func NewValidateCommand(_ *cmdutil.Factory, bridge *LegacyBridge) *cobra.Command {
	cfg := legacyCfgFrom(bridge)
	var platform string

	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a document file",
		Long: fmt.Sprintf(`Validate a document file and display any errors found.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Example:
  germinator validate agent.yaml --platform %s`, models.PlatformClaudeCode, models.PlatformOpenCode, models.PlatformOpenCode),
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			// Extract verbosity from command flags at runtime
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			filePath := args[0]

			VerbosePrint(cfg, "Validating file: %s", filePath)
			VerbosePrint(cfg, "Platform: %s", platform)

			if platform == "" {
				return gerrors.NewConfigError("platform", "", "--platform flag is required").WithSuggestions([]string{models.PlatformClaudeCode, models.PlatformOpenCode})
			}

			VeryVerbosePrint(cfg, "Loading document...")
			VeryVerbosePrint(cfg, "Parsing document structure...")
			VeryVerbosePrint(cfg, "Running validation...")

			result, err := bridge.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: filePath,
				Platform:  platform,
			})
			if err != nil {
				return fmt.Errorf("validating document: %w", err)
			}

			if !result.Valid() {
				// Slice-2: the legacy error_handler.go (and its
				// ValidationResultError type) is deleted. Surface the
				// first validation issue directly so output.FormatError
				// can dispatch on its concrete type via errors.As.
				if len(result.Errors) > 0 {
					return result.Errors[0]
				}
				return errors.New("validation failed")
			}

			_, _ = fmt.Fprintln(c.OutOrStdout(), "Document is valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("Target platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = cmd.MarkFlagRequired("platform")

	// Add platform flag completion
	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform": actionPlatforms(),
	})

	return cmd
}
