# cli-framework Specification (delta)

## ADDED Requirements

### Requirement: I/O adapter placement

Service-style I/O adapters (Transformer, Validator, Canonicalizer, Initializer, and per-resource adders) MUST live in dedicated `internal/<x>/` shell packages, not in `cmd/`. The Functional Core / Imperative Shell pattern requires that any code performing I/O (filesystem reads/writes, external tool calls, network requests) live at the package boundary (`internal/<shell>/`), not in the action layer (`cmd/`).

**Change**: NEW requirement codifying the post-extraction state. The slice-3 design rationale (`openspec/changes/archive/2026-06-26-migrate-domain-commands/design.md:38-48`) was a one-adapter argument for keeping validator logic in `cmd/`; after `extract-io-adapters` all 5 adapters live in `internal/<x>/`.

#### Scenario: Adapters live in internal/<x>/, not cmd/

- **WHEN** the codebase is searched for `transformerAdapter`, `validatorAdapter`, `canonicalizerAdapter`, `initializerAdapter`, `libraryAdapter`
- **THEN** zero matches SHALL appear in `cmd/`
- **AND** matches SHALL appear in `internal/transform/`, `internal/validate/`, `internal/canonicalize/`, `internal/install/`, and `internal/library/` (as methods on `*Library`) respectively

#### Scenario: Cmd-side interfaces remain in cmd/

- **WHEN** the codebase is searched for the `Transformer`, `Validator`, `Canonicalizer`, `Initializer`, `resourceAdder` interfaces
- **THEN** each interface SHALL be declared in its consumer's `cmd/<command>.go` file (per the skill's "interfaces where consumed" principle)
- **AND** the interfaces SHALL NOT be re-exported or duplicated in `internal/<x>/` (the cmd-side contract is the canonical declaration)

#### Scenario: NewService constructors live in shell packages

- **WHEN** the codebase is searched for `NewService` or `NewXxx` constructors of the extracted adapters
- **THEN** each constructor SHALL live in its shell package (`internal/transform/transform.go`, `internal/validate/validate.go`, `internal/canonicalize/canonicalize.go`, `internal/install/install.go`)
- **AND** each constructor SHALL return the cmd-side interface type (the package implements the interface, the consumer declares it)
- **AND** `cmd/` SHALL import the shell package only to call the constructor; the cmd file does NOT re-implement the adapter

#### Scenario: Library package methods replace the libraryAdapter

- **WHEN** the codebase is searched for `libraryAdapter`
- **THEN** zero matches SHALL appear (the type was deleted)
- **AND** `cmd/library_add.go` SHALL declare `var _ adderLibrary = (*library.Library)(nil)` as the compile-time check
- **AND** `*library.Library` SHALL expose `Add`, `BatchAddResources`, `DiscoverOrphans` methods (per slice-7 decision 6, completed by this change)

#### Scenario: New shell packages follow the internal/library convention

- **WHEN** any of the 4 new shell packages (`internal/validate/`, `internal/canonicalize/`, `internal/transform/`, `internal/install/`) is inspected
- **THEN** it SHALL have a `Service` interface, `Request`/`Result` types, and a `NewService` constructor
- **AND** it SHALL return `core.*` types (not package-local types)
- **AND** it SHALL take `ctx context.Context` as the first parameter of each public method
- **AND** it SHALL have an `AGENTS.md` following the `internal/library/AGENTS.md` template (Files table + Key Surface + skill reference)
