package cmd

import (
	"github.com/spf13/cobra"
)

// CommandConfig holds configuration and services for command execution.
type CommandConfig struct {
	Services       *ServiceContainer
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}

// NewCommandConfig creates a CommandConfig with services and runtime verbosity.
func NewCommandConfig(cmd *cobra.Command, services *ServiceContainer) *CommandConfig {
	verbosity, _ := cmd.Flags().GetCount("verbose")
	return &CommandConfig{
		Services:       services,
		ErrorFormatter: NewErrorFormatter(),
		Verbosity:      Verbosity(verbosity),
	}
}
