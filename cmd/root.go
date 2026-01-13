package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "germinator",
	Short: "A configuration adapter for AI coding assistant documents",
	Long: `Germinator is a configuration adapter that transforms AI coding assistant 
documents (commands, memory, skills, agents) between platforms. It uses Claude Code's 
document standard as source format and adapts it for other platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
