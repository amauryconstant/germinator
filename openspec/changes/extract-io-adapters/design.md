# Design — Extract I/O adapters to `internal/<x>/` shell packages

## Context

The current state has five I/O adapters living in `cmd/`:

- **`transformerAdapter`** at `cmd/transformer.go:33-36` — composes `parser.NewParser()` + `renderer.NewSerializer()`; implements the cmd-side `Transformer` interface (`cmd/adapt.go:19-21`).
- **`validatorAdapter`** at `cmd/validate.go:214` (zero-size type) — delegates to `validateDocument` (L130-183) and `unwrapErrors` (L184-197). Implements the cmd-side `Validator` interface (`cmd/validate.go:21`).
- **`canonicalizerAdapter`** at `cmd/canonicalize.go:232` (zero-size type) — delegates to `canonicalizeDocument` (L158-177), `validateCanonicalDoc` (L182-207), `unwrapCanonicalErrors` (L209-220). Implements the cmd-side `Canonicalizer` interface (`cmd/canonicalize.go:22`).
- **`initializerAdapter`** at `cmd/initializer.go:38-41` — 126 lines of per-ref orchestration (load → render → write). Implements the cmd-side `Initializer` interface (`cmd/initializer.go:18-20`).
- **`libraryAdapter`** at `cmd/library_add.go:83` (zero-state) — wraps the package-level `library.AddResource`, `library.DiscoverOrphans`, `library.BatchAddResources` functions to expose them as interface methods. Implements the cmd-side `adderLibrary` interface (`cmd/library_add.go:70`; renamed from `resourceAdder` as part of this change).

The cmd-side **interfaces** are correctly co-located with their consumers per the skill's "interfaces where consumed" principle (`SKILL.md:1316`). The deviation is the **concrete production adapters** living in `cmd/` rather than in `internal/<x>/`.

### Constraints

1. **`runF` injection must remain the test seam.** Each command constructor takes `runF func(*XxxOptions) error`; tests substitute a stub to bypass the real body. After extraction, the `runXxx` body still calls `xxx.NewService(...)` to construct the adapter; the injection seam is unaffected.
2. **Per-options lazy function fields stay.** `adaptOptions.Transformer`, `validateOptions.Validator`, etc. are typed as `func() (Transformer, error)` so tests can inject fakes. The lazy field pattern is preserved; only the production constructor changes.
3. **No DI container.** The slice-7 design removed `Factory.Transformer`/`Validator`/etc. lazy fields (`main.go:30-34` comment). This change does **not** re-introduce them; the per-options field is the only injection point.
4. **`internal/library.Library` methods precedent.** Slice 7 established that mutating library operations live as methods on `*library.Library` (`openspec/changes/archive/2026-07-01-migrate-library-rest/design.md:69-92`). The `libraryAdapter` is debt because `Add`, `BatchAdd`, `DiscoverOrphans` weren't migrated. This change completes the migration.
5. **`depguard` only gates `internal/core/**`.** Adding 4 new shell packages does not require any `depguard` rule changes (the existing rule allows stdlib + `samber/lo` + `gitlab.com/amoconst/germinator/internal/core` for core; other internal packages follow the general import rules).
6. **Lint baseline test.** `cmd/lint_test.go` runs `mise run lint` and diffs against `cmd/testdata/lint_baseline.txt`. New shell packages follow existing conventions; baseline is expected to remain unchanged.

### Assumption

The skill's "When to Extract" guidance (`SKILL.md:1051-1061`) is a **heuristic** — *"Extract when you have 5+ types or functions sharing a clear concern; An I/O boundary you want to test independently; A distinct external dependency."* — not a strict rule. The conservative interpretation (3+ commands sharing) is satisfied by other shell packages (e.g., `internal/library` is consumed by 9 commands). The conservative-extraction principle matches the project's apparent direction (slice 7 deleted `internal/service/` and `internal/application/` as anti-patterns; this change fills the gap with per-concern packages).

## Goals / Non-Goals

**Goals:**

- Extract the 5 I/O adapters into 4 dedicated `internal/<x>/` shell packages (`validate`, `canonicalize`, `transform`, `install`) plus 3 methods on `*library.Library` (`Add`, `BatchAddResources`, `DiscoverOrphans`).
- Each new package follows the existing shell-package convention (per `internal/library/AGENTS.md`): constructor `NewService(...)`, returns `core.*` types, takes `ctx context.Context`, AGENTS.md inline-linking the skill reference.
- Retire the explicit `libraryAdapter` debt (Stage 2).
- Preserve the `runF` injection seam and the per-options lazy function field pattern.
- Preserve all CLI behavior (no user-visible change).

**Non-Goals:**

- Re-introducing `Factory.Transformer`/`Validator`/`Canonicalizer`/`Initializer` lazy fields. The slice-7 deletion is preserved; the per-options field remains the only injection point.
- Refactoring `internal/parser` or `internal/renderer`. The new shell packages consume them; their internals are unchanged.
- Adding `internal/<x>/<x>_test.go` golden file tests. Stage 1 considers moving `cmd/canonicalize_golden_test.go`; other stages inherit existing test files (the cmd-side tests use `runF` injection and continue to pass).
- Removing the cmd-side `Transformer`/`Validator`/`Canonicalizer`/`Initializer`/`adderLibrary` interfaces. They stay per "interfaces where consumed".
- Changing the dependency direction between `internal/library` and other packages. `library` continues to be a leaf; the new shell packages depend on `library` but not vice versa.

## Decisions

### 1. Package naming: `internal/install/` (not `internal/init/`)

**Choice**: The Stage 3 package for the install/initialize logic is named `internal/install/`, not `internal/init/`.

**Rationale**: `init` is reserved as a Go package name initializer; Go tooling often treats `init` packages specially (e.g., `go test ./init/...` is flagged in some linters). `install` avoids the collision while preserving the semantic match (`core.InitializeResult`, `Initializer` interface, etc.).

**Alternatives considered**:
- *`internal/initialize/`*: works but is verbose; matches Go's `package` convention less naturally.
- *`internal/initializer/`*: similar.
- *Keep `cmd/initializer.go` as the home*: rejected; this is the deviation we're fixing.

### 2. Shell packages follow the `internal/library` convention

**Choice**: Each new shell package mirrors the structure of `internal/library/`:

```text
internal/validate/
├── AGENTS.md           # Files table + Key Surface + skill reference
├── validate.go         # Service interface, Request/Result types, adapter, NewService constructor
└── validate_test.go    # Table-driven tests with t.TempDir() fixtures
```

**Rationale**: Consistency across shell packages makes the codebase navigable. The `internal/library/AGENTS.md` template (Files table at line 19-37, Key Surface at line 24-31) is the reference for the AGENTS.md structure.

**Alternatives considered**:
- *Single-file packages without AGENTS.md*: rejected; project convention mandates AGENTS.md per package (see `internal/AGENTS.md` structure diagram).

### 3. `libraryAdapter` retirement via `*Library` methods

**Choice**: Stage 2 converts `library.AddResource`, `library.BatchAddResources`, `library.DiscoverOrphans` from package-level functions to methods on `*library.Library`. Existing package-level functions delegate to the methods (slice-7 decision 6 precedent).

**Rationale**: The slice-7 design rationale explicitly committed to this pattern for `Refresh`, `RemoveResource`, `Validate`, `Fix`. The remaining three (`Add`, `BatchAdd`, `DiscoverOrphans`) complete the migration. The adapter shim becomes unnecessary because `*library.Library` directly satisfies the cmd-side `adderLibrary` interface (via `var _ adderLibrary = (*library.Library)(nil)` compile-time check).

**Alternatives considered**:
- *Keep `libraryAdapter`, add a `ResourceAdder` interface to `internal/library`*: rejected; this re-creates the "parallel type" anti-pattern that slice 7 deleted (per `openspec/changes/archive/2026-07-01-migrate-library-rest/design.md:92`).
- *Move only `Add` to a method, keep `BatchAdd`/`DiscoverOrphans` as functions*: rejected; partial migration leaves an inconsistent API surface.

### 4. Import direction: shell packages depend on `internal/library`, not vice versa

**Choice**: The new shell packages (`internal/install/`, etc.) depend on `internal/library/` for resource resolution, library loading, and library validation. `internal/library/` does **not** import any of the new shell packages.

**Rationale**: `internal/library` is a leaf shell package consumed by multiple top-level packages (`cmd/`, the new `internal/install/`, etc.). Reversing the dependency would create a cycle. The depguard rule does not enforce this direction, but it follows the general "downstream packages depend on upstream" principle.

**Alternatives considered**:
- *Co-locate `install` logic with `library`*: rejected; the install logic is a different concern (per-ref orchestration) from library I/O (load/save/validate).

### 5. Preserve `runF` injection seam; per-options lazy field pattern unchanged

**Choice**: The extraction does not change how `runXxx` is called or tested. The cmd-side options struct's lazy field (`adaptOptions.Transformer func() (Transformer, error)`) is preserved. The only change is the production wiring inside `runXxx`:

```go
// Before (cmd/adapt.go:104-107)
resolve := opts.Transformer
if resolve == nil {
    resolve = func() (Transformer, error) { return NewTransformer(), nil }
}

// After (cmd/adapt.go:104-107)
resolve := opts.Transformer
if resolve == nil {
    resolve = func() (Transformer, error) {
        return transform.NewService(parser.NewParser(), renderer.NewSerializer()), nil
    }
}
```

**Rationale**: The injection seam is the project's primary testability mechanism (per `cmd/AGENTS.md:7`). Refactoring the production constructor without touching the seam preserves all existing tests; the new shell-package tests are additive.

**Alternatives considered**:
- *Move the lazy field to `Factory`*: rejected; slice 7 deleted these factory fields; this change does not re-introduce them.
- *Use `MustNewService` constructors that panic on error*: rejected; the existing pattern returns `(Xxx, error)` and handles the error in `runXxx`.

### 6. AGENTS.md updates mirror existing package templates

**Choice**: Each new shell package's AGENTS.md follows the `internal/library/AGENTS.md` template:

```markdown
**Location**: `internal/validate/`
**Parent**: See [`/internal/AGENTS.md`](../AGENTS.md) for package overview
**Skill reference**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`

---

# Validate Package

[Description of what the package does and why it exists.]

## Files

| File | Purpose |
|---|---|
| `validate.go` | `Service` interface, `Request`/`Result` types, `validatorAdapter`, `NewService` constructor |

## Key Surface

- `NewService() Service` — returns the production wiring
- `Service.Validate(ctx, *Request) (*Result, error)` — the per-call contract

## Why this package exists

[Lifted from `cmd/validate.go` per the `extract-io-adapters` change. ...]
```

**Rationale**: Consistency makes the codebase navigable and discoverable. Future contributors can pattern-match new shell packages against the existing ones.

**Alternatives considered**:
- *Skip AGENTS.md for the new packages*: rejected; the project mandates AGENTS.md per package.

### 7. Single mega-change with 4 internal stages

**Choice**: All 3 stages ship as a single OpenSpec change (`extract-io-adapters`) for archival cohesion. The internal stages are documented as numbered sections in `tasks.md` (1.0, 2.0, 3.0).

**Rationale**: The stages share a single architectural theme ("extract cmd-resident adapters to internal/<x>/"). Splitting into 4 separate changes adds 4× the OpenSpec ceremony for a coherent refactor. The user approved this approach: *"One mega-change per priority tier (Recommended)"*.

**Alternatives considered**:
- *Four separate OpenSpec changes*: rejected per user direction; would add review overhead without proportional benefit since the stages share architectural rationale.

## Risks / Trade-offs

- **Stage 2 method conversion: package-level functions must continue to work.** Existing tests (`internal/library/adder_test.go`, `cmd/library_add_test.go`) may use `library.AddResource` as a function value or pass it as a callback. **Mitigation**: Stage 2 task 2.2 keeps the package-level functions as thin wrappers (`func AddResource(ctx, req) error { return defaultLibrary.Add(ctx, &req) }`) so all existing call sites compile unchanged. The new methods are the canonical implementation; the package functions are the convenience layer.
- **Stage 3 import cycles.** If `internal/install` imports `internal/library` for `ResolveResource` / `ParseRef` / `GetOutputPath`, and `internal/library` ever needs `internal/install` for any reason, a cycle forms. **Mitigation**: design Decision 4 codifies the dependency direction; `depguard` does not enforce this directly, but the next lint run will fail with a clear error if a regression occurs. Stage 3 task 3.11 verifies `go build ./...` succeeds.
- **Golden file tests in `cmd/canonicalize_golden_test.go`** may depend on cmd-side state (the `NewCmdCanonicalize` constructor) for fixture setup. **Mitigation**: Stage 1 task 1.9 moves the golden test to `internal/canonicalize/canonicalize_golden_test.go` and uses `canonicalize.NewService()` directly; existing fixtures are byte-identical and move with the test.
- **`libraryAdapter` docstring at `cmd/library_add.go:60-63` becomes stale** after Stage 2 (the adapter no longer exists). **Mitigation**: Stage 2 task 2.8 deletes the docstring along with the adapter.
- **The 4 new shell packages add 4 new AGENTS.md files** that must follow the project convention. **Mitigation**: each follows the `internal/library/AGENTS.md` template as the starting point; ~30 lines; handled as part of the separate doc phase.
- **`internal/AGENTS.md` package dependency diagram** must include the 4 new packages. **Mitigation**: handled as part of the separate doc phase; the existing diagram (`internal/AGENTS.md:14-28`) is a simple text-block update.
- **Test coverage may dip briefly** during the refactor. **Mitigation**: each stage ends with `mise run check`; the new shell-package tests are additive (existing cmd tests still cover the command layer via `runF` injection). Final verification task 4.7 confirms coverage for the new packages ≥ 70%.
- **`gocyclo` / `gocognit` thresholds** (set to 25 and 30 respectively in `.golangci.yml`) may flag the larger shell-package method bodies. **Mitigation**: each method follows the existing `cmd/initializer.go:Initialize` pattern, which currently passes lint; if a method exceeds the threshold, extract a helper into the same package (no impact on the public API).

## Migration Plan

The change ships in **3 sequential stages**, each producing independently-mergeable commits:

1. **Stage 1** — Extract validator + canonicalizer (~3 hours). New packages: `internal/validate/`, `internal/canonicalize/`. No public API change. End state: `cmd/validate.go` and `cmd/canonicalize.go` are ~70 lines shorter.
2. **Stage 2** — Convert library adders to `*library.Library` methods (~2 hours). `libraryAdapter` deleted; `*library.Library` directly satisfies the cmd-side `adderLibrary` interface. End state: `cmd/library_add.go` loses 60 LOC of adapter code.
3. **Stage 3** — Extract transformer + initializer (~4 hours). New packages: `internal/transform/`, `internal/install/`. End state: `cmd/transformer.go` and `cmd/initializer.go` are deleted; `cmd/adapt.go` and `cmd/init.go` import the new packages.

**Rollback strategy**: revert each stage commit independently. Stages 1, 3 are additive (new packages); Stage 2 is a method addition + adapter deletion (revert by restoring the adapter).

**Sequencing rationale**: Stage 1 first because the slice-3 rationale applies most directly to validator/canonicalizer (smallest, simplest). Stage 2 second because it retires the explicit `libraryAdapter` debt documented at `cmd/library_add.go:60-63`. Stage 3 last because it touches the largest volume of code (transformer + initializer).
