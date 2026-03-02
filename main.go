package main

import (
	"os"

	"gitlab.com/amoconst/germinator/cmd"
)

func main() {
	// Composition root: wire all dependencies here
	services := cmd.NewServiceContainer()

	// Create config with services (verbosity will be extracted from command flags at runtime)
	cfg := &cmd.CommandConfig{
		Services:       services,
		ErrorFormatter: cmd.NewErrorFormatter(),
		Verbosity:      0, // Will be updated when command runs
	}

	// Build command tree and execute
	rootCmd := cmd.NewRootCommand(cfg)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
