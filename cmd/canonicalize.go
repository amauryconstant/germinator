package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/application"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
)

// NewCanonicalizeCommand creates the canonicalize command with dependency injection.
func NewCanonicalizeCommand(cfg *CommandConfig) *cobra.Command {
	var canonicalizePlatform string
	var canonicalizeDocType string

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
		Run: func(c *cobra.Command, args []string) {
			// Extract verbosity from command flags at runtime
			verbosity, _ := c.Flags().GetCount("verbose")
			cfg.Verbosity = Verbosity(verbosity)

			inputPath := args[0]
			outputPath := args[1]

			VerbosePrint(cfg, "Canonicalizing document...")
			VerbosePrint(cfg, "Output path: %s", outputPath)

			if canonicalizePlatform == "" {
				HandleError(cfg, gerrors.NewConfigError("platform", "", []string{models.PlatformClaudeCode, models.PlatformOpenCode}, "--platform flag is required"))
			}

			if canonicalizeDocType == "" {
				HandleError(cfg, gerrors.NewConfigError("type", "", []string{"agent", "command", "skill", "memory"}, "--type flag is required"))
			}

			if canonicalizePlatform != models.PlatformClaudeCode && canonicalizePlatform != models.PlatformOpenCode {
				HandleError(cfg, gerrors.NewConfigError("platform", canonicalizePlatform, []string{models.PlatformClaudeCode, models.PlatformOpenCode}, "invalid platform"))
			}

			if canonicalizeDocType != "agent" && canonicalizeDocType != "command" && canonicalizeDocType != "skill" && canonicalizeDocType != "memory" {
				HandleError(cfg, gerrors.NewConfigError("type", canonicalizeDocType, []string{"agent", "command", "skill", "memory"}, "invalid document type"))
			}

			VeryVerbosePrint(cfg, "Parsing platform document...")
			VeryVerbosePrint(cfg, "Validating document...")
			VeryVerbosePrint(cfg, "Marshalling to canonical YAML...")

			_, err := cfg.Services.Canonicalizer.Canonicalize(context.Background(), &application.CanonicalizeRequest{
				InputPath:  inputPath,
				OutputPath: outputPath,
				Platform:   canonicalizePlatform,
				DocType:    canonicalizeDocType,
			})
			if err != nil {
				HandleError(cfg, err)
			}

			fmt.Printf("Successfully canonicalized document to: %s\n", outputPath)
		},
	}

	cmd.Flags().StringVar(&canonicalizePlatform, "platform", "", fmt.Sprintf("Source platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = cmd.MarkFlagRequired("platform")
	cmd.Flags().StringVar(&canonicalizeDocType, "type", "", "Document type (required: agent, command, skill, memory)")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
