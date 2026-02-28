package main

import (
	"fmt"
	"os"
)

type Verbosity int

func (v Verbosity) IsVerbose() bool {
	return v >= 1
}

func (v Verbosity) IsVeryVerbose() bool {
	return v >= 2
}

func VerbosePrint(cfg *CommandConfig, format string, args ...any) {
	if cfg.Verbosity.IsVerbose() {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func VeryVerbosePrint(cfg *CommandConfig, format string, args ...any) {
	if cfg.Verbosity.IsVeryVerbose() {
		fmt.Fprintf(os.Stderr, "  "+format+"\n", args...)
	}
}
