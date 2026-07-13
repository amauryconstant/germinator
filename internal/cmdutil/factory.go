// Package cmdutil provides the Factory dependency-injection container
// plus shared CLI helpers: exit code mapping, output flag wiring, and
// common option types.
package cmdutil

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
)

// Factory is the only composition root in the new CLI architecture.
// Eager values (IOStreams, AppVersion, Executable, RootContext) are
// populated at construction. All other dependencies are exposed as
// lazy function fields that callers are expected to wrap in
// sync.OnceValues before assigning to the Factory.
type Factory struct {
	IOStreams   *iostreams.IOStreams
	AppVersion  string
	Executable  string
	RootContext context.Context

	rootCancel context.CancelFunc

	Config  func() (*config.Config, error)
	Library func() (*library.Library, error)

	// CompletionCache memoizes library snapshots for shell completion
	// lookups with a per-entry TTL. Populated in main.go; each Factory
	// owns its own cache so tests create a fresh cache per Factory.
	CompletionCache *CompletionCache
}

// NewFactory constructs a Factory with eager values populated. The
// signal-aware context is supplied by the caller (typically
// signal.NotifyContext in main.go) so the composition root owns the
// context lifecycle; Factory.Close cancels the same context. Lazy
// function fields are left nil and must be assigned by main.go (the
// only composition root) using OnceValuesFunc wrappers.
func NewFactory(ctx context.Context, io *iostreams.IOStreams, appVersion, executable string) *Factory {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	return &Factory{
		IOStreams:   io,
		AppVersion:  appVersion,
		Executable:  executable,
		RootContext: ctx,
		rootCancel:  cancel,
	}
}

// Close releases Factory-owned resources. Currently this cancels the
// root context so signal-driven cleanup propagates to consumers.
func (f *Factory) Close() {
	if f == nil {
		return
	}
	if f.rootCancel != nil {
		f.rootCancel()
	}
}

// configLoadForTest is a package-level seam for BuildFactory tests.
// Tests override it to inject a stub config loader; production code
// MUST NOT modify it. The variable is one mutable package-level
// binding, the documented cost of the test-injection seam.
var configLoadForTest func() (*config.Config, error) = config.Load

// BuildFactory wires the lazy Factory dependencies (Config, Library,
// CompletionCache) in a single testable function. It returns a
// fully-wired Factory plus any error from the first config load. main.go
// remains the only place that translates errors to exit codes via
// cmdutil.ExitCodeFor + os.Exit.
//
// Side effect: activates debug logging on io via IOStreams.SetDebug
// when cfg.Debug is true (the env-driven GERMINATOR_DEBUG flows through
// koanf → cfg.Debug → SetDebug, single source of truth).
func BuildFactory(ctx context.Context, io *iostreams.IOStreams, appVersion, executable string) (*Factory, error) {
	f := NewFactory(ctx, io, appVersion, executable)
	f.CompletionCache = NewCompletionCache()

	// Config is wired through OnceValuesFunc so subsequent calls from
	// completion actions return the same *Config pointer without
	// re-reading disk (per cli-cli-factory/spec.md).
	f.Config = OnceValuesFunc(configLoadForTest)

	// Eager single load: surface config errors here so BuildFactory
	// can return them. Subsequent f.Config() calls return the cached
	// pointer without re-running Load.
	cfg, err := f.Config()
	if err != nil {
		return f, err
	}

	// Activate debug logging based on the loaded config.
	io.SetDebug(cfg.Debug)

	// Library is wired here so the priority chain (flag > env > cfg >
	// XDG default) is enforced at construction time. The closure
	// captures the cached cfg pointer; later calls reuse it.
	f.Library = OnceValuesFunc(func() (*library.Library, error) {
		path := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"), cfg.Library)
		return library.LoadLibrary(f.RootContext, path)
	})

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
