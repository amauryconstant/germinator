package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
)

// NewValidateCommand creates the validate command with dependency injection.
func NewValidateCommand(cfg *CommandConfig) *cobra.Command {
	var validatePlatform string

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
		Run: func(c *cobra.Command, args []string) {
			// Extract verbosity from command flags at runtime
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			filePath := args[0]

			VerbosePrint(cfg, "Validating file: %s", filePath)
			VerbosePrint(cfg, "Platform: %s", validatePlatform)

			if validatePlatform == "" {
				HandleError(cfg, gerrors.NewConfigError("platform", "", []string{models.PlatformClaudeCode, models.PlatformOpenCode}, "--platform flag is required"))
			}

			VeryVerbosePrint(cfg, "Loading document...")
			VeryVerbosePrint(cfg, "Parsing document structure...")
			VeryVerbosePrint(cfg, "Running validation...")

			result, err := cfg.Services.Validator.Validate(context.Background(), &application.ValidateRequest{
				InputPath: filePath,
				Platform:  validatePlatform,
			})
			if err != nil {
				HandleError(cfg, err)
			}

			if !result.Valid() {
				for _, e := range result.Errors {
					fmt.Fprintln(os.Stderr, cfg.ErrorFormatter.Format(e))
				}
				os.Exit(int(ExitCodeUsage))
			}

			fmt.Println("Document is valid")
		},
	}

	cmd.Flags().StringVar(&validatePlatform, "platform", "", fmt.Sprintf("Target platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}
