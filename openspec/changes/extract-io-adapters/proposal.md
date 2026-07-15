# Extract I/O adapters to `internal/<x>/` shell packages

## Why

Five cross-command I/O adapters currently live in `cmd/` rather than in dedicated `internal/<x>/` shell packages:

| Adapter | File | LOC | Consumer |
|---|---|---|---|
| `validatorAdapter` | `cmd/validate.go:210-218` (+ helpers at 130-198) | ~90 | `cmd/validate` |
| `canonicalizerAdapter` | `cmd/canonicalize.go:232-240` (+ helpers at 158-220) | ~85 | `cmd/canonicalize` |
| `transformerAdapter` | `cmd/transformer.go:33-60` | 60 | `cmd/adapt` |
| `initializerAdapter` | `cmd/initializer.go:38-125` | 126 | `cmd/init` |
| `libraryAdapter` | `cmd/library_add.go:64-120` | ~60 | `cmd/library add` |

The `golang-cli-architecture` skill's decision trigger (`SKILL.md:1236-1250`) states: *"3+ commands sharing the same I/O adapter → Extract adapter to its own package."* This trigger does not strictly fire (each adapter has exactly one consumer). However, the secondary "When to Extract" guidance (`SKILL.md:1051-1061`) lists two additional criteria — *"an I/O boundary you want to test independently"* and *"a distinct external dependency (a specific tool, API, or library)"* — which clearly fire for all five adapters. The body of `cmd/transformer.go:44-60` does filesystem I/O (`os.WriteFile`) and composes two existing shell packages (`internal/parser` + `internal/renderer`); `cmd/initializer.go:53-125` does filesystem I/O across 73 lines of orchestration; the other three follow the same pattern.

The slice-3 design rationale (`openspec/changes/archive/2026-06-26-migrate-domain-commands/design.md:38-48`) justified the co-location for **one** 99-line validator. The codebase has since accumulated **five** adapters of varying complexity. The cumulative weight now exceeds the original rationale:

1. `libraryAdapter` is **explicitly documented as temporary debt** at `cmd/library_add.go:60-63`: *"the library package's functions are currently package-level rather than methods on `*Library` — converting them is out of scope for slice 6."*
2. The slice-7 design (`openspec/changes/archive/2026-07-01-migrate-library-rest/design.md:69-92`) added methods on `*library.Library` for `Refresh`, `RemoveResource`, `Validate`, `Fix` — but stopped short of `AddResource`, `BatchAddResources`, `DiscoverOrphans`, leaving `libraryAdapter` as the sole remaining bridge.
3. The skill's standard Tier 2 layout (`golang-cli-architecture/references/01-architecture.md:152-187`) explicitly diagrams `internal/<x>/` packages as the standard home for adapters. The five outliers are the **sole** deviations in the codebase.

Extracting these adapters aligns germinator with the rest of its `internal/<x>/` structure and retires the explicit `libraryAdapter` debt.

## What Changes

The change is structured as **three internal stages**, all under this single OpenSpec change for archival purposes. Each stage is independently mergeable and revertable. Documentation (AGENTS.md updates, spec sync) and archive are handled separately.

### Stage 1 — Extract validator + canonicalizer

- **NEW** `internal/validate/validate.go` (~90 LOC + AGENTS.md): move `validateDocument` and `unwrapErrors` from `cmd/validate.go:130-198`. Define `Service` interface, `Request`/`Result` types, and `transformerAdapter`-style implementation.
- **NEW** `internal/canonicalize/canonicalize.go` (~85 LOC + AGENTS.md): move `canonicalizeDocument` (L158-177), `validateCanonicalDoc` (L182-207), `unwrapCanonicalErrors` (L209-220) from `cmd/canonicalize.go`.
- **MODIFY** `cmd/validate.go`: delete `validateDocument`, `unwrapErrors`, `validatorAdapter` (lines 130-218). Import `internal/validate`. The cmd-side `Validator` interface at `cmd/validate.go:21` stays (per "interfaces where consumed").
- **MODIFY** `cmd/canonicalize.go`: delete `canonicalizeDocument`, `validateCanonicalDoc`, `unwrapCanonicalErrors`, `canonicalizerAdapter` (lines 158-240). Import `internal/canonicalize`. The cmd-side `Canonicalizer` interface at `cmd/canonicalize.go:22` stays.

### Stage 2 — Convert library adders to `*library.Library` methods

- **MODIFY** `internal/library/library.go`: add methods on `*Library`:
  - `Add(ctx context.Context, req *AddRequest) error`
  - `BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error)`
  - `DiscoverOrphans(ctx context.Context, opts *DiscoverOptions) (*DiscoverResult, error)`
  - Existing package-level functions (`library.AddResource`, `library.BatchAddResources`, `library.DiscoverOrphans`) delegate to the methods (slice-7 decision 6 precedent — keeps public API stable).
- **MODIFY** `internal/library/adder.go`: rewrite as methods on `*Library` (per slice-7 decision 6).
- **MODIFY** `internal/library/discovery.go`: rewrite `DiscoverOrphans` as a method on `*Library`.
- **MODIFY** `cmd/library_add.go`: delete `resourceAdder` interface (L70), `libraryAdapter` (L83-120), `defaultAdder` (L118), and the compile-time check (L113). Replace `var _ resourceAdder = (*libraryAdapter)(nil)` with `var _ adderLibrary = (*library.Library)(nil)`. Update `runAdd*` bodies to use `lib.Add`, `lib.BatchAddResources`, `lib.DiscoverOrphans` directly.

### Stage 3 — Extract transformer + initializer

- **NEW** `internal/transform/transform.go` (~60 LOC + AGENTS.md): move the entire `cmd/transformer.go` content. Define `Service` interface, `Request`/`Result` types, `transformerAdapter` (renamed from `cmd.transformerAdapter`).
- **NEW** `internal/install/install.go` (~126 LOC + AGENTS.md): move the entire `cmd/initializer.go` content. Define `Service` interface, `InitializeRequest`/`InitializeResult` types (re-export of `core.InitializeResult`), `initializerAdapter`. **Note**: package name `install` chosen to avoid collision with Go's reserved `init` identifier; the cmd-side `Initializer` interface stays.
- **MODIFY** `cmd/adapt.go`: import `internal/transform`. The `Transformer` interface at `cmd/adapt.go:19-21` stays. Update `runAdapt` to construct via `transform.NewService(parser.NewParser(), renderer.NewSerializer())`.
- **MODIFY** `cmd/init.go`: import `internal/install`. The `Initializer` interface at `cmd/initializer.go:18-20` (kept in `cmd/init.go` after `cmd/initializer.go` deletion) stays. Update `runInit` to construct via `install.NewService(parser.NewParser(), renderer.NewSerializer())`.
- **DELETE** `cmd/transformer.go`, `cmd/initializer.go` (entire files).

## Capabilities

### Modified Capabilities

- **`cli-framework`** (delta) — Service-style I/O adapters (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`, and per-resource adders) MUST live in dedicated `internal/<x>/` shell packages, not in `cmd/`. Commands depend on cmd-side `xxx` interfaces declared at the call site; production wiring lives in the shell package and is reached via `xxx.NewService(...)` constructors.

## Impact

### Affected code

| Stage | New files | Modified files | Deleted files | LOC moved |
|---|---|---|---|---|
| 1 | `internal/validate/validate.go` + AGENTS.md, `internal/canonicalize/canonicalize.go` + AGENTS.md | `cmd/validate.go`, `cmd/canonicalize.go` | — | -175 |
| 2 | — | `internal/library/library.go`, `internal/library/adder.go`, `internal/library/discovery.go`, `cmd/library_add.go` | `cmd/library_add.go:64-120` (delete adapter + interface) | -60 / +100 |
| 3 | `internal/transform/transform.go` + AGENTS.md, `internal/install/install.go` + AGENTS.md | `cmd/adapt.go`, `cmd/init.go` | `cmd/transformer.go`, `cmd/initializer.go` | -186 / +186 |
| **Net** | **+8 files** (4 .go + 4 AGENTS.md) | **8 files** | **3 files** | **-421 cmd / +461 internal** |

### Affected systems

- **cmd/ size:** shrinks by ~421 LOC. Each command file becomes shorter and more focused on the parse/execute/respond concerns (no inline adapter body).
- **internal/ structure:** gains 4 new shell packages (`validate`, `canonicalize`, `transform`, `install`), each following the standard shell-package convention (constructor → `NewService(...)`, returns `core.*` types, takes `ctx`).
- **`*library.Library`:** gains 3 new methods (`Add`, `BatchAddResources`, `DiscoverOrphans`), completing the slice-7 forward path. The `libraryAdapter` bridge is retired.
- **CLI behavior:** unchanged. End users see no difference; the `germinator adapt`, `validate`, `canonicalize`, `init`, `library add` commands produce the same outputs as before.
- **Test surface:** the `runF` injection seam is preserved. The new shell packages get table-driven unit tests with `t.TempDir()` fixtures (per the existing shell-package convention).
- **Lint baseline:** likely unchanged (the new packages follow existing shell-package conventions; `depguard` only gates `internal/core/**`).

### Backward compatibility

- **No public API break.** All `internal/<x>/` packages are new; the existing `cmd/` interfaces (`Transformer`, `Validator`, etc.) stay in place. The `library.AddResource`, `library.BatchAddResources`, `library.DiscoverOrphans` package functions keep their existing signatures and delegate to the new `*Library` methods.
- **No CLI behavior change.** All existing flags, args, output formats, and exit codes preserved.
- **No migration required for end users.** This is a refactor of internal code organization only.

## Risks

- **Stage 2 method conversion breaks callers that pass package-level functions as values.** `library.AddResource` may be used as a top-level function value in test files. **Mitigation**: verify via `rg "^\s*library\.(AddResource|BatchAddResources|DiscoverOrphans)" internal/library/adder_test.go cmd/library_add_test.go` before implementing; the package-level functions continue to exist as thin wrappers around the methods so existing callers keep working.
- **Stage 3 import cycles if `internal/install` imports `internal/library` AND `internal/library` ever imports `internal/install`.** Currently, `internal/library` does not import `internal/install` (the dependency flows one way: `install` → `library` for resource resolution). **Mitigation**: design Decision 4 enforces the import direction; `depguard` will catch a regression on the next lint run.
- **Golden file tests in `cmd/canonicalize_golden_test.go` may need relocation** if the golden file fixtures depend on cmd-side state. **Mitigation**: Stage 1 task 1.9 moves the golden test to `internal/canonicalize/canonicalize_golden_test.go` (per the shell-package convention); the existing fixtures are byte-identical and move with the test.
- **The 4 new shell packages add 4 new `AGENTS.md` files** that must be kept in sync with the project conventions. **Mitigation**: each AGENTS.md follows the `internal/library/AGENTS.md` template as a starting point; ~30 lines following the established Files + Key Surface structure; handled as part of the separate doc phase.
- **`internal/AGENTS.md` package dependency diagram must be updated** to include the 4 new packages. **Mitigation**: handled as part of the separate doc phase; the diagram is verified by `depguard` (which already passes for the existing packages).
- **The `libraryAdapter` docstring at `cmd/library_add.go:60-63` becomes stale** after Stage 2. **Mitigation**: Stage 2 task 2.8 deletes the docstring along with the adapter.
