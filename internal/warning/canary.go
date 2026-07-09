// Package warning provides user-facing warning helpers that emit to
// stderr independent of verbosity and the debug Logger.
//
// The package exists to keep side effects out of pure helpers (per
// the golang-cli-architecture skill: exits and side effects are
// imperative). Pure logic stays in core/cmdutil; this package owns
// the emission of one-shot warnings (currently the exit-code
// deprecation canary emitted from main.go on exit code 1).
package warning

import (
	"os"
	"sync"

	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// canaryOnce guards single-emission semantics for the exit-code
// deprecation warning across the lifetime of one process.
var canaryOnce sync.Once

// MaybeWarnLegacyExitCode emits a one-time deprecation warning to
// io.ErrOut when the legacy exit-code gate fires:
//
//	EXIT_CODE_LEGACY env var set OR stderr is a TTY.
//
// The warning is written via io.Warnf (Styles.Warning) and is
// independent of io.Logger (which is gated on GERMINATOR_DEBUG).
// A nil io is a no-op. Subsequent calls are no-ops within the same
// process; ResetCanaryForTest resets the once-state for unit tests.
//
// Invoked from main.go immediately before os.Exit(int(cmdutil.ExitCodeFor(err)))
// when the resolved exit code is 1 (the mapped "general error" code).
// Exit code 2 (ExitCodeUsage, from Cobra/pflag usage errors) MUST NOT
// trigger the canary per the cli-exit-codes ADDED requirement.
func MaybeWarnLegacyExitCode(io *iostreams.IOStreams) {
	if io == nil {
		return
	}
	if os.Getenv("EXIT_CODE_LEGACY") == "" && !io.IsStderrTTY() {
		return
	}
	canaryOnce.Do(func() {
		io.Warnf("exit code 5 was renamed to 1; see CHANGELOG.md for the migration timeline")
	})
}

// ResetCanaryForTest resets the package-level sync.Once so unit tests
// can exercise the canary across multiple sub-tests. Not for
// production use.
func ResetCanaryForTest() {
	canaryOnce = sync.Once{}
}
