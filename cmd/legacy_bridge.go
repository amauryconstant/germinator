package cmd

import "gitlab.com/amoconst/germinator/internal/application"

// LegacyBridge is the transitional shim that keeps non-migrated
// commands working between slice 2 (this change) and slice 7
// (cleanup-and-finalize). It is constructed in main.go and passed to
// cmd.NewRootCommand alongside the Factory. Non-migrated commands
// (validate, canonicalize, init, library sub-commands other than
// resources, config, completion, version) read from bridge.Services
// and bridge.ErrorFormatter during the migration window.
//
// The bridge is intentionally a value-object exposed from the cmd
// package so it can cross the package boundary from main.go (which
// lives in the composition root) to cmd/ (which owns the command
// tree). All fields are public; nil-services handling is documented
// per consumer. bridge.Services may be nil until task 2.1.3b runs
// (after cmd/container.go is deleted in slice 2).
type LegacyBridge struct {
	Services       *LegacyServices
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}

// LegacyServices holds the legacy application-service implementations
// used by non-migrated commands. Constructed by main.go via direct
// calls to application.New* and service.New* constructors (no
// indirection through the deleted cmd/container.go).
type LegacyServices struct {
	Transformer   application.Transformer
	Validator     application.Validator
	Canonicalizer application.Canonicalizer
	Initializer   application.Initializer
}

// CommandConfig holds the per-process configuration shared between
// non-migrated commands during the slice-2..slice-7 migration window:
// verbosity and the legacy error formatter. Service references are
// no longer wired through CommandConfig; non-migrated commands read
// services directly from the LegacyBridge. CommandConfig is
// deleted in slice 7 alongside the LegacyBridge.
//
// The struct lives in legacy_bridge.go (per slice-2 design: a single
// file owns the migration-window types).
type CommandConfig struct {
	ErrorFormatter *ErrorFormatter
	Verbosity      Verbosity
}

// legacyCfgFrom builds the local CommandConfig used by non-migrated
// command RunE closures. The Verbosity and ErrorFormatter come from
// the LegacyBridge; Service references are read directly from
// bridge.Services (the deleted ServiceContainer type is no longer
// plumbed through CommandConfig). A nil bridge yields a CommandConfig
// with a fresh NewErrorFormatter() so callers never see nil.
func legacyCfgFrom(bridge *LegacyBridge) *CommandConfig {
	cfg := &CommandConfig{}
	if bridge != nil {
		cfg.Verbosity = bridge.Verbosity
		cfg.ErrorFormatter = bridge.ErrorFormatter
	}
	if cfg.ErrorFormatter == nil {
		cfg.ErrorFormatter = NewErrorFormatter()
	}
	return cfg
}
