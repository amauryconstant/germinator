// Package main provides CLI for germinator tool.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amoconst/germinator/internal/services"
)

var platform string

var adaptCmd = &cobra.Command{
	Use:   "adapt <input> <output>",
	Short: "Transform a document to another platform",
	Long:  "Transform a document from source format to another platform's format.",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]

		if platform == "" {
			fmt.Fprintln(os.Stderr, "Error: --platform flag is required")
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
	adaptCmd.Flags().StringVar(&platform, "platform", "", "Target platform (required)")
	_ = adaptCmd.MarkFlagRequired("platform")
	rootCmd.AddCommand(adaptCmd)
}
