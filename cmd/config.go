package cmd

type CommandConfig struct {
	Services       *ServiceContainer
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}
