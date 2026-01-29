package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/services"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a document file",
	Long:  "Validate a document file and display any errors found.",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		filePath := args[0]

		errs, err := services.ValidateDocument(filePath, "claude-code")
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
	rootCmd.AddCommand(validateCmd)
}
