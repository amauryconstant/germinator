// Package main provides the germinator CLI application for transforming
// AI coding assistant configuration documents between platforms.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/amoconst/germinator/cmd"
	"gitlab.com/amoconst/germinator/internal/application"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
	"gitlab.com/amoconst/germinator/internal/parser"
	"gitlab.com/amoconst/germinator/internal/renderer"
	"gitlab.com/amoconst/germinator/internal/service"
	"gitlab.com/amoconst/germinator/internal/version"
	"gitlab.com/amoconst/germinator/internal/warning"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	io := iostreams.System()
	f := cmdutil.NewFactory(ctx, io, version.Version, "germinator")
	defer f.Close()

	// Populate lazy function fields. cmdutil.OnceValuesFunc[T] is the
	// generic helper equivalent to sync.OnceValues; main.go is the
	// only composition root allowed to wire these per the
	// golang-cli-architecture skill.
	p := parser.NewParser()
	s := renderer.NewSerializer()

	f.Transformer = cmdutil.OnceValuesFunc(func() (application.Transformer, error) {
		return service.NewTransformer(p, s), nil
	})
	f.Validator = cmdutil.OnceValuesFunc(func() (application.Validator, error) {
		return cmd.NewValidator(), nil
	})
	f.Canonicalizer = cmdutil.OnceValuesFunc(func() (application.Canonicalizer, error) {
		return cmd.NewCanonicalizer(), nil
	})
	f.Initializer = cmdutil.OnceValuesFunc(func() (application.Initializer, error) {
		return service.NewInitializer(p, s), nil
	})
	f.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		path := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"))
		return library.LoadLibrary(path)
	})

	// LegacyBridge keeps non-migrated commands (init, library
	// sub-commands other than resources, config, completion, version)
	// wired until slice 7 deletes them. Validator/Canonicalizer now
	// come from the cmd-side adapters (slice-3); the service
	// implementations in internal/service/{validator,canonicalizer}.go
	// were deleted. Service implementations are constructed directly
	// via cmd.New* / service.New*; no indirection through the deleted
	// cmd/container.go.
	bridge := &cmd.LegacyBridge{
		Services: &cmd.LegacyServices{
			Transformer:   service.NewTransformer(p, s),
			Validator:     cmd.NewValidator(),
			Canonicalizer: cmd.NewCanonicalizer(),
			Initializer:   service.NewInitializer(p, s),
		},
		ErrorFormatter: cmd.NewErrorFormatter(),
		Verbosity:      0,
	}

	rootCmd := cmd.NewRootCommand(f, bridge)
	rootCmd.SetContext(ctx)
	if err := rootCmd.Execute(); err != nil {
		output.FormatError(f.IOStreams, err)
		// Per the cli-exit-codes ADDED requirement, the canary fires
		// only when the resolved exit code is 1 (ExitCodeError).
		// Exit code 2 (ExitCodeUsage, from Cobra/pflag usage errors)
		// MUST NOT trigger the canary.
		if cmdutil.ExitCodeFor(err) == cmdutil.ExitCodeError {
			warning.MaybeWarnLegacyExitCode(f.IOStreams)
		}
		os.Exit(int(cmdutil.ExitCodeFor(err)))
	}
}
