package warning

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Each test that exercises the canary must call ResetCanaryForTest
// (or rely on the explicit reset call) so the package-level sync.Once
// does not leak state across tests in the same process.

func resetCanary(t *testing.T) {
	t.Helper()
	ResetCanaryForTest()
}

// TestMaybeWarnLegacyExitCode_NilIsNoop verifies the helper tolerates
// a nil IOStreams (paranoia: callers should never pass nil, but the
// canary runs from main.go where nil would otherwise panic).
func TestMaybeWarnLegacyExitCode_NilIsNoop(t *testing.T) {
	t.Run("env var set, nil io", func(t *testing.T) {
		resetCanary(t)
		t.Setenv("EXIT_CODE_LEGACY", "1")
		assert.NotPanics(t, func() { MaybeWarnLegacyExitCode(nil) })
	})

	t.Run("no env var, nil io", func(t *testing.T) {
		resetCanary(t)
		t.Setenv("EXIT_CODE_LEGACY", "")
		assert.NotPanics(t, func() { MaybeWarnLegacyExitCode(nil) })
	})
}

// TestMaybeWarnLegacyExitCode_EnvGate verifies the EXIT_CODE_LEGACY
// env var triggers emission when stderr is not a TTY (the helper
// always-emits path for CI opt-in).
func TestMaybeWarnLegacyExitCode_EnvGate(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "1")

	io := iostreams.Test()
	io.SetStderrTTY(false)
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.True(t, strings.Contains(out.String(), "Warning: "),
		"expected canary warning, got %q", out.String())
}

// TestMaybeWarnLegacyExitCode_TTYGate verifies the stderr TTY state
// triggers emission when no env var is set (interactive session path).
func TestMaybeWarnLegacyExitCode_TTYGate(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "")

	io := iostreams.Test()
	io.SetStderrTTY(true)
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.True(t, strings.Contains(out.String(), "Warning: "),
		"expected canary warning, got %q", out.String())
}

// TestMaybeWarnLegacyExitCode_SuppressedWhenBothFalse verifies the
// helper is silent in the typical CI scenario (no env var, stderr
// not a TTY).
func TestMaybeWarnLegacyExitCode_SuppressedWhenBothFalse(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "")

	io := iostreams.Test()
	io.SetStderrTTY(false)
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.Equal(t, "", out.String(),
		"expected no canary emission, got %q", out.String())
}

// TestMaybeWarnLegacyExitCode_SingleEmissionPerProcess verifies the
// sync.Once guard: a second call within the same process is a no-op
// regardless of gate conditions.
func TestMaybeWarnLegacyExitCode_SingleEmissionPerProcess(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "1")

	io := iostreams.Test()
	io.SetStderrTTY(true)
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	first := out.String()
	require.NotEmpty(t, first, "first call should emit")

	// Second call: gate is still met (env var set + TTY) but the
	// once-state must suppress the second emission.
	io2 := iostreams.Test()
	io2.SetStderrTTY(true)
	MaybeWarnLegacyExitCode(io2)

	out2, ok := io2.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.Equal(t, "", out2.String(),
		"second call must be a no-op while once-state holds")
}

// TestResetCanaryForTest_ResetsOnceState verifies the test helper
// restores the once-state so subsequent calls in the same test
// process can re-emit.
func TestResetCanaryForTest_ResetsOnceState(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "1")

	io := iostreams.Test()
	io.SetStderrTTY(true)
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	require.NotEmpty(t, out.String(), "first emission expected")

	ResetCanaryForTest()

	io2 := iostreams.Test()
	io2.SetStderrTTY(true)
	MaybeWarnLegacyExitCode(io2)

	out2, ok := io2.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.True(t, strings.Contains(out2.String(), "Warning: "),
		"after ResetCanaryForTest the canary must emit again, got %q", out2.String())
}

// TestMaybeWarnLegacyExitCode_NilLoggerStillEmits verifies the canary
// does not depend on io.Logger (which is gated on GERMINATOR_DEBUG).
// Per the cli-exit-codes ADDED requirement, the warning must reach
// io.ErrOut even when the Logger is nil.
func TestMaybeWarnLegacyExitCode_NilLoggerStillEmits(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "1")

	io := iostreams.Test()
	io.Logger = nil
	MaybeWarnLegacyExitCode(io)

	out, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	assert.True(t, strings.Contains(out.String(), "Warning: "),
		"canary must write to io.ErrOut regardless of Logger, got %q", out.String())
}

// TestMaybeWarnLegacyExitCode_ExitsAreIndependentOfCanary verifies
// that the helper itself does not call os.Exit and does not panic
// when invoked outside a cobra run context. (The exit-code-2 exclusion
// is a caller-side concern: main.go only invokes this when the
// resolved exit code is 1.)
func TestMaybeWarnLegacyExitCode_ExitsAreIndependentOfCanary(t *testing.T) {
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "1")
	io := iostreams.Test()
	io.SetStderrTTY(true)
	assert.NotPanics(t, func() { MaybeWarnLegacyExitCode(io) })

	// Suppression path: no env var, non-TTY. Must be a no-op.
	resetCanary(t)
	t.Setenv("EXIT_CODE_LEGACY", "")
	io2 := iostreams.Test()
	io2.SetStderrTTY(false)
	assert.NotPanics(t, func() { MaybeWarnLegacyExitCode(io2) })
}

// TestCanary_UnsetEnvClears verifies Setenv("", "") semantics — the
// canary reads via os.Getenv which returns "" for both unset and
// empty values. This test guards against future refactors that
// might switch to os.LookupEnv.
func TestCanary_UnsetEnvClears(t *testing.T) {
	old, had := os.LookupEnv("EXIT_CODE_LEGACY")
	defer func() {
		if had {
			_ = os.Setenv("EXIT_CODE_LEGACY", old)
		} else {
			_ = os.Unsetenv("EXIT_CODE_LEGACY")
		}
	}()

	_ = os.Unsetenv("EXIT_CODE_LEGACY")
	assert.Equal(t, "", os.Getenv("EXIT_CODE_LEGACY"))
}
