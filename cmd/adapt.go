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

// NewAdaptCommand creates the adapt command with dependency injection.
func NewAdaptCommand(cfg *CommandConfig) *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "adapt <input> <output>",
		Short: "Transform a document to another platform",
		Long: fmt.Sprintf(`Transform a document from Germinator source format to another platform's format.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Example:
  germinator adapt agent.yaml opencode-agent.md --platform %s`, models.PlatformClaudeCode, models.PlatformOpenCode, models.PlatformOpenCode),
		Args: cobra.ExactArgs(2),
		Run: func(c *cobra.Command, args []string) {
			// Extract verbosity from command flags at runtime
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			inputPath := args[0]
			outputPath := args[1]

			VerbosePrint(cfg, "Transforming document...")
			VerbosePrint(cfg, "Output path: %s", outputPath)

			if platform == "" {
				HandleError(cfg, gerrors.NewConfigError("platform", "", "--platform flag is required").WithSuggestions([]string{models.PlatformClaudeCode, models.PlatformOpenCode}))
			}

			VeryVerbosePrint(cfg, "Loading source document...")
			VeryVerbosePrint(cfg, "Rendering template for %s...", platform)
			VeryVerbosePrint(cfg, "Writing output file...")

			_, err := cfg.Services.Transformer.Transform(context.Background(), &application.TransformRequest{
				InputPath:  inputPath,
				OutputPath: outputPath,
				Platform:   platform,
			})
			if err != nil {
				HandleError(cfg, err)
			}

			_, _ = fmt.Fprintf(c.OutOrStdout(), "Document transformed successfully to %s\n", outputPath)
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
