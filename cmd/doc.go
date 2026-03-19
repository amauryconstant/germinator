// Package cmd provides CLI commands for the Germinator configuration adapter.
//
// Commands:
//
//	adapt      - Transform a document to another platform format
//	validate   - Validate a document against platform rules
//	canonicalize - Convert a platform document to canonical format
//	version    - Display version, commit, and build date
//	library    - Manage the canonical resource library
//	init       - Install resources from library to project
//	completion - Generate shell completion scripts
//
// All commands use dependency injection through CommandConfig and ServiceContainer.
package cmd
