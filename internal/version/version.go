// Package version provides version information for the Germinator CLI.
// Version strings are populated at build time via ldflags.
package version

var (
	// Version is the current semantic version of the application.
	Version = "dev"
	// Commit is the git commit hash from which the binary was built.
	Commit = ""
	// Date is the timestamp when the binary was built.
	Date = ""
)
