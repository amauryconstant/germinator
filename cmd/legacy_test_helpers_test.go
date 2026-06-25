package cmd

// Legacy test-only helpers. The file suffix _test.go restricts these
// to test builds; nothing in this file is reachable from production
// binaries.
//
// TODO(slice-7): delete this file when the remaining non-pilot
// commands (validate, canonicalize, init, library presets/show/add/
// create/remove/validate/refresh, config, completion, version) are
// migrated to the NewCmdXxx(f, runF) pattern.
//
//nolint:paralleltest // helpers are package-globals shared across sub-tests.

import (
	"context"
	"os"

	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
	"gitlab.com/amoconst/germinator/internal/service"
)

// newTestFactory builds a Factory wired with a populated LegacyBridge
// containing real service implementations. Suitable for tests that
// need a Factory instance without manually constructing the bridge.
//
// The returned Factory's lazy fields are populated with OnceValuesFunc
// wrappers; tests that exercise the same service from multiple call
// sites share a single instance. The Library lazy field loads from
// the GERMINATOR_LIBRARY env var (matching main.go's production path).
func newTestFactory() *cmdutil.Factory {
	io := iostreams.Test()
	f := cmdutil.NewFactory(context.Background(), io, "test", "germinator")
	p := parser.NewParser()
	s := renderer.NewSerializer()
	f.Transformer = cmdutil.OnceValuesFunc(func() (application.Transformer, error) {
		return service.NewTransformer(p, s), nil
	})
	f.Validator = cmdutil.OnceValuesFunc(func() (application.Validator, error) {
		return service.NewValidator(), nil
	})
	f.Canonicalizer = cmdutil.OnceValuesFunc(func() (application.Canonicalizer, error) {
		return service.NewCanonicalizer(), nil
	})
	f.Initializer = cmdutil.OnceValuesFunc(func() (application.Initializer, error) {
		return service.NewInitializer(p, s), nil
	})
	f.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		path := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"))
		return library.LoadLibrary(path)
	})
	return f
}

// newTestBridge returns a LegacyBridge with all four services
// populated by direct calls to the application/service constructors.
// Tests that exercise non-migrated commands can pass this to
// NewRootCommand and the per-command constructors.
func newTestBridge() *LegacyBridge {
	p := parser.NewParser()
	s := renderer.NewSerializer()
	return &LegacyBridge{
		Services: &LegacyServices{
			Transformer:   service.NewTransformer(p, s),
			Validator:     service.NewValidator(),
			Canonicalizer: service.NewCanonicalizer(),
			Initializer:   service.NewInitializer(p, s),
		},
		ErrorFormatter: NewErrorFormatter(),
	}
}

// newTestConfig is the legacy helper retained for non-pilot tests
// that still reference *CommandConfig directly. It returns a
// CommandConfig with a fresh ErrorFormatter; services are not
// populated (non-pilot RunE bodies read bridge.Services directly,
// not cfg.Services).
func newTestConfig() *CommandConfig {
	return &CommandConfig{ErrorFormatter: NewErrorFormatter()}
}
