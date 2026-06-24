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

	"gitlab.com/amoconst/germinator/internal/application"
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

	Config        func() (*config.Config, error)
	Library       func() (*library.Library, error)
	Transformer   func() (application.Transformer, error)
	Validator     func() (application.Validator, error)
	Canonicalizer func() (application.Canonicalizer, error)
	Initializer   func() (application.Initializer, error)
}

// NewFactory constructs a Factory with eager values populated. The
// lazy function fields are left nil and must be assigned by main.go
// (the only composition root) using sync.OnceValues wrappers.
func NewFactory(io *iostreams.IOStreams, appVersion, executable string) *Factory {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
