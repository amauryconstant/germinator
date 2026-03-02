package cmd

import (
	"fmt"
	"os"
)

// Verbosity represents the verbosity level for command output.
type Verbosity int

// IsVerbose returns true if verbosity level is 1 or higher.
func (v Verbosity) IsVerbose() bool {
	return v >= 1
}

// IsVeryVerbose returns true if verbosity level is 2 or higher.
func (v Verbosity) IsVeryVerbose() bool {
	return v >= 2
}

// VerbosePrint prints a message to stderr if verbosity is enabled.
func VerbosePrint(cfg *CommandConfig, format string, args ...any) {
	if cfg.Verbosity.IsVerbose() {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// VeryVerbosePrint prints a detailed message to stderr if very verbose mode is enabled.
func VeryVerbosePrint(cfg *CommandConfig, format string, args ...any) {
	if cfg.Verbosity.IsVeryVerbose() {
		fmt.Fprintf(os.Stderr, "  "+format+"\n", args...)
	}
}
