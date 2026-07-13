// Package main provides the germinator CLI application for transforming
// AI coding assistant configuration documents between platforms.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/amoconst/germinator/cmd"
	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/config"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/library"
	"gitlab.com/amoconst/germinator/internal/output"
	"gitlab.com/amoconst/germinator/internal/version"
	"gitlab.com/amoconst/germinator/internal/warning"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	io := iostreams.System()
	f := cmdutil.NewFactory(ctx, io, version.Version, "germinator")
	defer f.Close()

	f.CompletionCache = cmdutil.NewCompletionCache()

	// Config is wired here (the only composition root). OnceValuesFunc
	// caches the result so subsequent calls from completion actions
	// return the same *Config pointer without re-reading the disk.
	f.Config = cmdutil.OnceValuesFunc(config.Load)
	cfg, err := f.Config()
	if err != nil {
		output.FormatError(f.IOStreams, err)
		os.Exit(int(cmdutil.ExitCodeFor(err)))
	}

	// Activate debug logging based on the loaded config (GERMINATOR_DEBUG
	// flows through koanf env provider -> cfg.Debug -> SetDebug).
	io.SetDebug(cfg.Debug)

	// The four service interfaces (Transformer/Validator/Canonicalizer/
	// Initializer) were removed from the Factory in slice 7.5 — their
	// concrete adapters are now constructed lazily inside the per-command
	// run functions (cmd.NewTransformer, cmd.NewValidator, etc.), so
	// main.go has nothing to wire for those.
	f.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) {
		path := library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"), cfg.Library)
		return library.LoadLibrary(f.RootContext, path)
	})

	rootCmd := cmd.NewRootCommand(f)
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
