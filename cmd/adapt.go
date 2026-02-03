// Package main provides CLI for germinator tool.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	Run: func(_ *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]

		if platform == "" {
			fmt.Fprintf(os.Stderr, "Error: --platform flag is required (available: %s, %s)\n", models.PlatformClaudeCode, models.PlatformOpenCode)
			os.Exit(1)
		}

		err := services.TransformDocument(inputPath, outputPath, platform)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Document transformed successfully to %s\n", outputPath)
	},
}

func init() {
	adaptCmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("Target platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = adaptCmd.MarkFlagRequired("platform")
	rootCmd.AddCommand(adaptCmd)
}
