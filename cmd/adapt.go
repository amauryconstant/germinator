// Package main provides CLI for germinator tool.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	gerrors "gitlab.com/amoconst/germinator/internal/errors"
	"gitlab.com/amoconst/germinator/internal/models"
	"gitlab.com/amoconst/germinator/internal/services"
)

var platform string

var adaptCmd = &cobra.Command{
	Use:   "adapt <input> <output>",
	Short: "Transform a document to another platform",
	Long: fmt.Sprintf(`Transform a document from Germinator source format to another platform's format.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Example:
  germinator adapt agent.yaml opencode-agent.md --platform %s`, models.PlatformClaudeCode, models.PlatformOpenCode, models.PlatformOpenCode),
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := NewCommandConfig(cmd)
		inputPath := args[0]
		outputPath := args[1]

		VerbosePrint(cfg, "Transforming document...")
		VerbosePrint(cfg, "Output path: %s", outputPath)

		if platform == "" {
			HandleError(cfg, gerrors.NewConfigError("platform", "", []string{models.PlatformClaudeCode, models.PlatformOpenCode}, "--platform flag is required"))
		}

		VeryVerbosePrint(cfg, "Loading source document...")
		VeryVerbosePrint(cfg, "Rendering template for %s...", platform)
		VeryVerbosePrint(cfg, "Writing output file...")

		err := services.TransformDocument(inputPath, outputPath, platform)
		if err != nil {
			HandleError(cfg, err)
		}

		fmt.Printf("Document transformed successfully to %s\n", outputPath)
	},
}

func init() {
	adaptCmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("Target platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = adaptCmd.MarkFlagRequired("platform")
	rootCmd.AddCommand(adaptCmd)
}
