// Package main provides CLI for germinator tool.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/models"
	"gitlab.com/amoconst/germinator/internal/services"
)

var canonicalizePlatform string
var canonicalizeDocType string

var canonicalizeCmd = &cobra.Command{
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
	Run: func(_ *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]

		if canonicalizePlatform == "" {
			fmt.Fprintf(os.Stderr, "Error: --platform flag is required (valid: %s, %s)\n", models.PlatformClaudeCode, models.PlatformOpenCode)
			os.Exit(1)
		}

		if canonicalizeDocType == "" {
			fmt.Fprintf(os.Stderr, "Error: --type flag is required (valid: agent, command, skill, memory)\n")
			os.Exit(1)
		}

		if canonicalizePlatform != models.PlatformClaudeCode && canonicalizePlatform != models.PlatformOpenCode {
			fmt.Fprintf(os.Stderr, "Error: invalid platform '%s' (valid: %s, %s)\n", canonicalizePlatform, models.PlatformClaudeCode, models.PlatformOpenCode)
			os.Exit(1)
		}

		if canonicalizeDocType != "agent" && canonicalizeDocType != "command" && canonicalizeDocType != "skill" && canonicalizeDocType != "memory" {
			fmt.Fprintf(os.Stderr, "Error: invalid document type '%s' (valid: agent, command, skill, memory)\n", canonicalizeDocType)
			os.Exit(1)
		}

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: input file not found: %s\n", inputPath)
			os.Exit(1)
		}

		err := services.CanonicalizeDocument(inputPath, outputPath, canonicalizePlatform, canonicalizeDocType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully canonicalized document to: %s\n", outputPath)
	},
}

func init() {
	canonicalizeCmd.Flags().StringVar(&canonicalizePlatform, "platform", "", fmt.Sprintf("Source platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = canonicalizeCmd.MarkFlagRequired("platform")
	canonicalizeCmd.Flags().StringVar(&canonicalizeDocType, "type", "", "Document type (required: agent, command, skill, memory)")
	_ = canonicalizeCmd.MarkFlagRequired("type")
	rootCmd.AddCommand(canonicalizeCmd)
}
