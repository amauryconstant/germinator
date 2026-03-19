package cmd

// CommandConfig holds configuration and services for command execution.
type CommandConfig struct {
	Services       *ServiceContainer
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}
