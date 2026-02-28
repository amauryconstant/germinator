package main

import (
	"github.com/spf13/cobra"
)

type CommandConfig struct {
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}

func NewCommandConfig(cmd *cobra.Command) *CommandConfig {
	verbosity, _ := cmd.Flags().GetCount("verbose")
	return &CommandConfig{
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      Verbosity(verbosity),
	}
}
