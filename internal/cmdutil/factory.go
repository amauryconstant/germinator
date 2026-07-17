// Package cmdutil provides the Factory dependency-injection container
// plus shared CLI helpers: exit code mapping, output flag wiring, and
// common option types.
package cmdutil

import (
	"context"
	"sync"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// Factory is the only composition root in the new CLI architecture.
// Eager values (IOStreams, RootContext) are populated at construction.
// All other dependencies are exposed as lazy function fields that
// callers are expected to wrap in sync.OnceValues before assigning
// to the Factory.
type Factory struct {
	IOStreams   *iostreams.IOStreams
	RootContext context.Context

	Config func() (*config.Config, error)

	// CompletionCache memoizes library snapshots for shell completion
	// lookups with a per-entry TTL. Populated in main.go; each Factory
	// owns its own cache so tests create a fresh cache per Factory.
	CompletionCache *CompletionCache
}

// NewFactory constructs a Factory with eager values populated. The
// caller supplies the (already signal-aware) context — typically
// signal.NotifyContext in main.go — and is therefore the sole owner
// of the context lifecycle. NewFactory does NOT re-wrap with
// signal.NotifyContext; doing so previously created a double-wrap
// that contradicted the package's own contract and made every
// cmdutil.NewFactory() call register a second signal handler
// (N4/N17 in the architecture review).
func NewFactory(ctx context.Context, io *iostreams.IOStreams) *Factory {
	return &Factory{
		IOStreams:   io,
		RootContext: ctx,
	}
}

// Close releases Factory-owned resources. Currently a nil-safe
// no-op preserved as the API surface for future Factory-owned
// resources; the root context's cancellation lifecycle is owned by
// main.go, not by the Factory.
func (f *Factory) Close() {
	_ = f
}

// configLoadForTest is a package-level seam for BuildFactory tests.
// Tests override it via swapConfigLoadForTest to inject a stub config
// loader; production code MUST NOT modify it. The variable is one
// mutable package-level binding, the documented cost of the
// test-injection seam. The mutex guards against concurrent test
// mutations if a future parallel test attempts to swap it.
var (
	configLoadForTestMu sync.RWMutex
	configLoadForTest   func() (*config.Config, error) = config.Load
)

// getConfigLoadForTest returns the currently registered config loader
// under read lock. Used by BuildFactory on the hot path.
func getConfigLoadForTest() func() (*config.Config, error) {
	configLoadForTestMu.RLock()
	defer configLoadForTestMu.RUnlock()
	return configLoadForTest
}

// swapConfigLoadForTest replaces the package-level config loader and
// returns a restore function suitable for t.Cleanup. Use in tests only;
// production code MUST NOT call this.
func swapConfigLoadForTest(fn func() (*config.Config, error)) func() {
	configLoadForTestMu.Lock()
	defer configLoadForTestMu.Unlock()
	prev := configLoadForTest
	configLoadForTest = fn
	return func() {
		configLoadForTestMu.Lock()
		defer configLoadForTestMu.Unlock()
		configLoadForTest = prev
	}
}

// BuildFactory wires the lazy Factory dependencies (Config,
// CompletionCache) in a single testable function. It returns a
// fully-wired Factory plus any error from the first config load. main.go
// remains the only place that translates errors to exit codes via
// cmdutil.ExitCodeFor + os.Exit.
//
// Side effect: activates debug logging on io via IOStreams.SetDebug
// when cfg.Debug is true (the env-driven GERMINATOR_DEBUG flows through
// koanf → cfg.Debug → SetDebug, single source of truth).
func BuildFactory(ctx context.Context, io *iostreams.IOStreams) (*Factory, error) {
	f := NewFactory(ctx, io)
	f.CompletionCache = NewCompletionCache()

	// Config is wired through OnceValuesFunc so subsequent calls from
	// completion actions return the same *Config pointer without
	// re-reading disk (per cli-cli-factory/spec.md).
	f.Config = OnceValuesFunc(getConfigLoadForTest())

	// Eager single load: surface config errors here so BuildFactory
	// can return them. Subsequent f.Config() calls return the cached
	// pointer without re-running Load.
	cfg, err := f.Config()
	if err != nil {
		return f, err
	}

	// Activate debug logging based on the loaded config.
	io.SetDebug(cfg.Debug)

	return f, nil
}

// OnceValuesFunc is a generic helper that returns a function which
// invokes fn exactly once and caches the result. Subsequent calls
// return the cached value and error.
func OnceValuesFunc[T any](fn func() (T, error)) func() (T, error) {
	var (
		once  sync.Once
		value T
		err   error
	)
	return func() (T, error) {
		once.Do(func() {
			value, err = fn()
		})
		return value, err
	}
}
