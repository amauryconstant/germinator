# library-library-resolution Specification (delta)

## ADDED Requirements

### Requirement: ResolvePreset accepts ctx

The `library.(*Library).ResolvePreset` method SHALL accept `ctx context.Context` as the first parameter. The method SHALL forward `ctx` to any I/O performed during resolution. Resolution is currently a pure in-memory map lookup (`lib.Presets[name]` at `internal/library/resolver.go:67-73`); the `ctx` parameter is accepted for spec symmetry with the `cli-framework` requirement and is forwarded whenever the resolution path is extended to perform I/O. If no I/O is performed, `ctx` MAY be ignored per the `cli-framework/spec.md` accept-and-may-ignore pattern.

**Change**: ADDED the `ctx` parameter requirement. The pre-change `ResolvePreset` discarded `ctx` via the `_ context.Context` underscore binding at `internal/library/resolver.go:67`. This change renames the parameter to `ctx context.Context` for spec symmetry. Earlier wording referenced an `os.ReadFile` call that did not exist in the implementation; that reference is removed and replaced with the accept-and-may-ignore pattern documented in `cli-framework/spec.md` (the canonical example of the pattern). The package-level `ResolvePreset(lib, name)` shim at `internal/library/resolver.go:54-56` (which synthesized `context.TODO()` mid-request-path, violating `golang-context` best practice) is removed by task 3.4b; callers migrate to `(*Library).ResolvePreset(ctx, name)`.

#### Scenario: ResolvePreset signature

- **WHEN** `(*library.Library).ResolvePreset(ctx, name)` is called
- **THEN** the method SHALL return `([]string, error)`
- **AND** the method SHALL have `ctx context.Context` as a named, non-underscore first parameter
- **AND** if the resolution path is extended to perform I/O in the future, that I/O SHALL respect `ctx` cancellation (forwarded via the `ctx` parameter)

#### Scenario: Cancellation during resolution

- **GIVEN** a `ctx` that is cancelled mid-resolution
- **WHEN** `(*library.Library).ResolvePreset(ctx, name)` is called
- **THEN** the method SHALL return within bounded time
- **AND** if any future I/O is added to the resolution path, the returned error SHALL wrap `context.Canceled` via `%w`

#### Scenario: Zero underscore bindings in production code

- **WHEN** the codebase is searched for `_ context\.Context` in `internal/library/resolver.go`
- **THEN** zero matches SHALL appear (the underscore binding is replaced with `ctx`)

#### Scenario: Package-level ResolvePreset shim is removed

- **WHEN** the codebase is searched for the package-level declaration `func ResolvePreset(lib \*Library` in `internal/library/` AND for the unqualified call form `ResolvePreset(lib,` in `cmd/` and `internal/`
- **THEN** zero matches SHALL appear for both (the shim at `internal/library/resolver.go:54-56` is deleted by task 3.4b; all callers use the method form `lib.ResolvePreset(ctx, name)`). The qualified pattern `library\.ResolvePreset\(` is insufficient on its own — in-package callers omit the `library.` qualifier, so it returns zero matches even when the shim's callers have not been migrated; the unqualified `ResolvePreset(lib,` form is required to catch them.
