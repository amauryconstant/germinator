package cmd

import (
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

// testExecMu serialises carapace-touching operations (command
// construction AND execution) across parallel tests.
//
// carapace-sh/carapace@v1.11.1 keeps completion state in a
// package-level variable that cobra's preRun hook reads/writes
// without internal locking (carapace storage.go:59,66). Concurrent
// carapace.Gen(...).FlagCompletion(...) calls during command
// construction race on that state, and concurrent cmd.Execute()
// calls race again on the cobra preRun bridge. We hold the mutex
// across BOTH the constructor callback and any cobra method that
// touches that state (Execute, Help, Commands, …) so the entire
// construct-then-interact sequence is serialised.
//
// This mirrors the configLoadForTestMu pattern at
// internal/cmdutil/factory.go:75 (sync.RWMutex around a test-only
// global). The wall-clock cost of serialising N parallel tests'
// cobra interaction is negligible: each carapace operation is
// sub-millisecond.
//
// Tests that bypass both construction and execution (e.g. runXxx
// invoked directly via runF injection) do not need this helper.
var testExecMu sync.Mutex

// executeCmd is the test-only entry point that serialises cobra
// command construction + execution through testExecMu. Use in place
// of `cmd := NewCmdX(...); cmd.Execute()` from parallel tests that
// build any carapace-registered command.
//
// The ctor callback runs under the lock, so carapace's shared state
// is not touched concurrently by parallel test goroutines. The ctor
// return type is `any` (not `*cobra.Command`) so callers do not
// need to import github.com/spf13/cobra just to construct the
// closure — Go's type inference lets `func() any { return NewCmdX(f, ...) }`
// satisfy the parameter without naming cobra in the test file.
func executeCmd(t *testing.T, ctor func() any, args ...string) error {
	t.Helper()
	testExecMu.Lock()
	defer testExecMu.Unlock()
	cmd := ctor().(*cobra.Command)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// withCmd runs ctor under testExecMu and passes the resulting
// *cobra.Command to fn. Use for tests that construct a command and
// then immediately interact with it (Help, Commands, etc.) — both
// steps need the lock because carapace's package-level state is
// touched by construction AND by cobra methods that walk the flag
// tree.
func withCmd(t *testing.T, ctor func() any, fn func(*cobra.Command)) {
	t.Helper()
	testExecMu.Lock()
	defer testExecMu.Unlock()
	cmd := ctor().(*cobra.Command)
	fn(cmd)
}
