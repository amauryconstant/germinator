package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
)

// NewValidateCommand creates the validate command with dependency injection.
func NewValidateCommand(cfg *CommandConfig) *cobra.Command {
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

			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: filePath,
				Platform:  platform,
			})
			if err != nil {
				return err
			}

			if !result.Valid() {
				// Wrap all validation errors in a single error for centralized handling
				// The error formatter will handle displaying multiple errors
				return &ValidationResultError{Errors: result.Errors}
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
