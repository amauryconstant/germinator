package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/models"
	"gitlab.com/amoconst/germinator/internal/services"
)

var validatePlatform string

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a document file",
	Long: fmt.Sprintf(`Validate a document file and display any errors found.

Supported platforms:
  %s - Claude Code document format
  %s    - OpenCode document format

Example:
  germinator validate agent.yaml --platform %s`, models.PlatformClaudeCode, models.PlatformOpenCode, models.PlatformOpenCode),
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		filePath := args[0]

		if validatePlatform == "" {
			fmt.Fprintf(os.Stderr, "Error: --platform flag is required (available: %s, %s)\n", models.PlatformClaudeCode, models.PlatformOpenCode)
			os.Exit(1)
		}

		errs, err := services.ValidateDocument(filePath, validatePlatform)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "%v\n", e)
			}
			os.Exit(1)
		}

		fmt.Println("Document is valid")
	},
}

func init() {
	validateCmd.Flags().StringVar(&validatePlatform, "platform", "", fmt.Sprintf("Target platform (required: %s, %s)", models.PlatformClaudeCode, models.PlatformOpenCode))
	_ = validateCmd.MarkFlagRequired("platform")
	rootCmd.AddCommand(validateCmd)
}
