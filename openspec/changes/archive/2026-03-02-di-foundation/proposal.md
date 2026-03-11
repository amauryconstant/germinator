## Why

Germinator is evolving from a simple document transformer into a comprehensive CLI for AI coding agent configuration management. As features grow, we need a clean dependency injection pattern to manage services and keep commands testable. Currently, commands call service functions directly with no explicit dependency wiring, making testing difficult and future service expansion cumbersome.

## What Changes

- **New**: `ServiceContainer` struct in `cmd/container.go` to hold service instances
- **New**: `NewRootCommand(config *CommandConfig)` pattern replacing global `rootCmd` variable
- **New**: Command constructors `NewValidateCommand`, `NewAdaptCommand`, `NewCanonicalizeCommand`, `NewVersionCommand`
- **New**: `main.go` as composition root (extracted from `cmd/root.go`)
- **Modified**: All command files converted from `init()` pattern to constructor pattern
- **Removed**: Global command variables (`validateCmd`, `adaptCmd`, etc.)
- **Removed**: `init()` functions that register commands

## Capabilities

### New Capabilities

- `dependency-injection`: Service container pattern for managing command dependencies with composition root wiring

### Modified Capabilities

- None (this is infrastructure, existing command behavior unchanged)

## Impact

**Files Changed:**
- `cmd/container.go` (new) - ServiceContainer definition
- `cmd/root.go` - NewRootCommand constructor, remove main()
- `cmd/validate.go` - NewValidateCommand constructor
- `cmd/adapt.go` - NewAdaptCommand constructor
- `cmd/canonicalize.go` - NewCanonicalizeCommand constructor
- `cmd/version.go` - NewVersionCommand constructor
- `main.go` (new) - Composition root with service wiring

**Dependencies:**
- Builds on `cli-infrastructure` change (CommandConfig struct)
- No external dependencies

**Testability:**
- Commands can receive mock services via ServiceContainer
- Integration tests can wire test doubles at composition root
