package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/domain"
	"gitlab.com/amoconst/germinator/internal/models"
)

// NewCanonicalizeCommand creates the canonicalize command with dependency injection.
func NewCanonicalizeCommand(cfg *CommandConfig) *cobra.Command {
	var platform string
	var docType string

	cmd := &cobra.Command{
		Use:   "canonicalize <input> <output>",
		Short: "Convert a platform document to canonical format",
		Long: fmt.Sprintf(`Convert a platform document to canonical YAML format.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Supported document types:
  agent   - Agent configuration
  command - Command configuration
  skill   - Skill configuration
  memory  - Memory configuration

Example:
  germinator canonicalize agent.md canonical-agent.yaml --platform %s --type agent`, models.PlatformClaudeCode, models.PlatformOpenCode, models.PlatformOpenCode),
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			// Extract verbosity from command flags at runtime
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			inputPath := args[0]
			outputPath := args[1]

			VerbosePrint(cfg, "Canonicalizing document...")
			VerbosePrint(cfg, "Output path: %s", outputPath)

			if platform == "" {
				return gerrors.NewConfigError("platform", "", "--platform flag is required").WithSuggestions([]string{models.PlatformClaudeCode, models.PlatformOpenCode})
			}

			if docType == "" {
				return gerrors.NewConfigError("type", "", "--type flag is required").WithSuggestions([]string{"agent", "command", "skill", "memory"})
			}

			VeryVerbosePrint(cfg, "Parsing platform document...")
			VeryVerbosePrint(cfg, "Validating document...")
			VeryVerbosePrint(cfg, "Marshalling to canonical YAML...")

			_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
				InputPath:  inputPath,
				OutputPath: outputPath,
				Platform:   platform,
				DocType:    docType,
			})
			if err != nil {
				return fmt.Errorf("canonicalizing document: %w", err)
			}

			_, _ = fmt.Fprintf(c.OutOrStdout(), "Successfully canonicalized document to: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("Source platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	cmd.Flags().StringVar(&docType, "type", "", "Document type (required: agent, command, skill, memory)")
	_ = cmd.MarkFlagRequired("platform")
	_ = cmd.MarkFlagRequired("type")

	// Add flag completion for carapace
	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"platform": actionPlatforms(),
		"type":     carapace.ActionValuesDescribed("agent", "command", "skill", "memory"),
	})

	return cmd
}
