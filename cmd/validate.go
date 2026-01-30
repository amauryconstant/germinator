package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/services"
)

var validatePlatform string

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a document file",
	Long: `Validate a document file and display any errors found.

Supported platforms:
  claude-code - Claude Code document format
  opencode    - OpenCode document format

Example:
  germinator validate agent.yaml --platform opencode`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		filePath := args[0]

		if validatePlatform == "" {
			fmt.Fprintln(os.Stderr, "Error: --platform flag is required (available: claude-code, opencode)")
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
	validateCmd.Flags().StringVar(&validatePlatform, "platform", "", "Target platform (required: claude-code, opencode)")
	_ = validateCmd.MarkFlagRequired("platform")
	rootCmd.AddCommand(validateCmd)
}
