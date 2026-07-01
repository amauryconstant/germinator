// Package cmd provides CLI commands for the Germinator configuration adapter.
//
// Commands:
//
//	adapt         - Transform a document to another platform format
//	validate      - Validate a document against platform rules
//	canonicalize  - Convert a platform document to canonical format
//	version       - Display version, commit, and build date
//	library       - Manage the canonical resource library
//	init          - Install resources from library to project
//	completion    - Generate shell completion scripts
//
// All commands receive a *cmdutil.Factory via NewCmdXxx(f, runF);
// lazy closures on the Factory are the only dependency-injection
// surface in this package (slice 7).
package cmd
