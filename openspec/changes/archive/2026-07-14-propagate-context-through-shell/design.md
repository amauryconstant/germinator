## Context

The 2026-07-08 review identified 8 findings all stemming from the same root cause: the slice-7 design (`openspec/changes/archive/2026-07-01-migrate-library-rest/design.md:69-92`) added methods on `*library.Library` for `Refresh`, `RemoveResource`, `Validate`, `Fix` but left `context.Background()` hard-coded in the package-level functions that delegate to those methods. The TODO markers say "replace with caller context (c.Context() in runF wiring)" — a debt that has not been paid.

Simultaneously, the cmd-layer adapters (`initializer.go`, `transformer.go`, `canonicalize.go`, `validate.go`) accept `ctx context.Context` as a method parameter but discard it via the `_ context.Context` underscore binding. The `extract-io-adapters` change relocates these adapters to `internal/{install,transform,validate,canonicalize}/` shell packages but does not thread `ctx` through; the new `extract-io-adapters/specs/cli-framework/spec.md:42` promises:

> *"it SHALL take `ctx context.Context` as the first parameter of each public method"*

This change fulfills that promise.

The 6 helper closures in `cmd/library_*.go` and `cmd/init.go` (`addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`, `initLibrary`) wrap `FindLibrary + LoadLibrary` and capture `f.RootContext` rather than the per-call `c.Context()`. The same pattern repeats inline in 3 read-only commands (`cmd/show.go:93-96`, `cmd/resources.go:84-87`, `cmd/presets.go:74-77`) — those use `opts.Ctx` correctly, but the duplication is real. The `f.Library` field's eager wiring at `internal/cmdutil/factory.go:134-137` is a 7th site that captures `f.RootContext`. This change removes the eager wiring while preserving the struct field per `cli-cli-factory/spec.md:25,30-32,105,125-126`.

The `internal/library/resolver.go:67` `ResolvePreset` method also uses `_ context.Context` (one of the rare legitimate use cases for an underscore binding, since the method is path-based). After this change, `ResolvePreset` accepts `ctx` to satisfy the cli-framework spec symmetry requirement — the file-read I/O now respects cancellation.

Additionally, the package-level `ResolvePreset(lib, name)` shim at `internal/library/resolver.go:54-56` synthesizes `context.TODO()` mid-request-path. The shim exists to support non-migrated callers during the slice-5 → slice-7 transition window; that window is closed, but the shim was left in place. Per `golang-context` best-practice #8 ("never create a new `context.Background()` in the middle of a request path"), synthesizing any context mid-call — `context.Background()` or `context.TODO()` — is a violation when the caller has a live ctx in scope. This change deletes the shim (Decision 7) and migrates callers to `(*Library).ResolvePreset(ctx, name)`.

### Constraints

1. **No public API break for external consumers.** The library package is `internal/`; signature changes are acceptable.
2. **`extract-io-adapters` Stage 2** is in flight: it converts `AddResource` / `BatchAddResources` / `DiscoverOrphans` from package-level functions to methods on `*Library`. The `ctx` parameter must be present in both the package-level function and the `*Library` method.
3. **`fix-library-io-discipline`** — absorbed by this change. The current `CreateLibrary` signature on `git HEAD` is `func CreateLibrary(opts CreateOptions) error`; this change introduces BOTH the `io.Writer` parameter (what `fix-library-io-discipline` would have added) AND the ctx-first reorder in a single step. The end-state signature is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error`. `fix-library-io-discipline` is no longer a prerequisite and is folded into task 3.4.
4. **Test surface churn** — every test calling an updated function needs a `ctx` argument. The minimum-churn approach is to use `context.Background()` in test files that don't test cancellation; new tests use `t.Context()` (Go 1.24+).
5. **Test fakes** — 4 fake adapter implementations (`fakeValidator`, `fakeInitializer`, `fakeCanonicalizer`, `fakeTransformer`) in `cmd/*_test.go` use `_ context.Context` to satisfy the interface. They have zero behavior dependency on ctx. This change renames them to `ctx` (zero behavior change) so the cli-framework spec scenario's `rg "_ context\.Context"` returns zero matches.

## Goals / Non-Goals

**Goals:**

- Retire all 4 `TODO(slice-7)` markers in `internal/library/{creator,refresher,remover}.go`.
- Thread `ctx` through the 4 cmd-layer adapters and the 4 new shell packages from `extract-io-adapters`.
- Replace `f.RootContext` with the captured per-command context in the 6 helper sites (and remove the helpers entirely).
- Drop the eager `f.Library` wiring from `internal/cmdutil/factory.go:134-137` (preserving the struct field per `cli-cli-factory/spec.md`).
- Add `ctx` parameter to `parser.LoadDocument` / `DetectType` / `ParsePlatformDocument` and `renderer.RenderDocument` / `MarshalCanonical`.
- Update `ResolvePreset` to accept ctx (spec symmetry).

**Non-Goals:**

- Refactoring parser/renderer internals beyond adding the `ctx` parameter.
- Changing `Factory.RootContext` semantics (the field remains for completion-time use).
- Removing the `*Library` method set introduced by slice 7.
- Changing the 3 read-only commands (`show.go`, `resources.go`, `presets.go`) — they already conform.

## Decisions

### 1. `ctx` as first parameter (Go convention)

**Choice**: `ctx` is added as the first parameter to all updated functions, following Go convention (per `golang-context` skill: "all I/O functions take ctx as first parameter").

**Rationale**: Standard Go convention. Matches the existing pattern at the cmd-side: `runXxx(opts *XxxOptions) error` calls `library.Foo(opts.Ctx, ...)`. Avoids churn at the cmd layer.

**Footnote**: Two functions accept `ctx` for spec symmetry rather than for active cancellation: `DetectType` (returns `string`, regex-only) and `ResolvePreset` (in-memory map lookup, no `os.ReadFile` despite earlier spec wording). For both, `ctx` is accepted and may be ignored when no I/O is performed; the parameter is in place for the moment either path acquires I/O. This accept-and-may-ignore pattern is ADDED by this change's `cli-framework` delta (requirement `I/O adapter ctx propagation`, see `openspec/changes/propagate-context-through-shell/specs/cli-framework/spec.md`): "If a method does not need cancellation, the `ctx` SHALL still be accepted (and may be ignored), so the call site is uniform across the codebase." The pattern lands in the parent spec when the delta syncs.

Per `golang-context` rule #7 (never pass `nil`; use `context.TODO()` only as a placeholder) and rule #8 (never create a new `context.Background()` mid-request); see `golang-context/references/cancellation.md`. The change's rg verifications (tasks 1.12, 1.13, 3.5, 3.6, 4.6) enforce both rules in production code.

**Alternatives considered**:

- *Add `ctx` as last parameter*: rejected; non-idiomatic; `go vet` would flag it.
- *Add a `Ctx` field to a request struct*: rejected; bloats the request struct for what is a per-call concern, not a per-request concern.

### 2. Test fixture pattern: `context.Background()` for existing, `t.Context()` for new

**Choice**: Existing test files use `context.Background()` as the minimum-churn approach. New tests added in this change use `t.Context()` (Go 1.24+).

**Rationale**: `t.Context()` is the modern pattern (auto-cancelled when the test ends), but migrating all existing tests is out of scope. New tests for ctx-cancellation behavior use `t.Context()` to make the test self-cleaning.

**Alternatives considered**:

- *Migrate all tests to `t.Context()`*: rejected; out of scope (covered by `harden-tests-and-coverage`).
- *Add a new `internal/testutil` package with `NewTestContext(t)` helper*: rejected; over-engineered for a single use case.

### 3. Sequential `ctx.Err()` checks in `DiscoverOrphans` (not errgroup)

**Choice**: `DiscoverOrphans` retains its existing sequential implementation: per-directory `ctx.Err()` checks at the top of each subdirectory scan plus per-file `ctx.Err()` checks inside the recursive walker. No `errgroup.WithContext` refactor is performed in this change.

**Rationale**: The sequential pattern is exactly what `golang-context/references/cancellation.md` recommends for CPU-bound work ("For CPU-bound work, periodically check `ctx.Err()`"). The scan touches 4 directories (`skills`, `agents`, `commands`, `memory`) with per-file walker entries; the latency improvement from parallelization is sub-millisecond and not user-perceptible, so per `golang-cli-architecture/references/06-concurrency.md` ("sequential-first; reach for `errgroup` only when I/O latency is measurably user-perceptible"), the sequential form is the right choice. The spec scenario is rewritten to reflect this pattern (see `specs/library-library-orphan-discovery/spec.md`).

**Alternatives considered**:

- *Refactor to `errgroup.WithContext(ctx)` with `SetLimit(4)`*: rejected; sub-millisecond latency delta, increased concurrency-test surface, and a more complex cancellation contract for no measurable benefit. If a future workload makes directory scans slow (e.g., reading file metadata over NFS), the errgroup refactor is a follow-up change.
- *Distinguish between cancellation and other errors*: rejected; `errors.Is(err, context.Canceled)` is the standard Go pattern; the existing `ctx.Err()` wrap with `fmt.Errorf("%w", ctx.Err())` already preserves the sentinel.

### 4. Inline closure pattern in `RunE` wrapped in `cmdutil.OnceValuesFunc`

**Choice**: Each `RunE` constructs `opts.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) { return library.LoadLibrary(c.Context(), path) })`. The existing `cmdutil.OnceValuesFunc[T]` wrapper preserves per-RunE caching (multiple `opts.Library()` calls in the same `RunE` invocation share the loaded library) while still capturing `c.Context()` at construction time. This matches the prior `f.Library = OnceValuesFunc(...)` caching contract at the cmd layer; only the cache scope shrinks from per-Factory to per-RunE. No new helper is introduced.

**Rationale**: The pattern already exists at `cmd/show.go:93-96`, `cmd/resources.go:84-87`, `cmd/presets.go:74-77` and works correctly with `opts.Ctx`. A `NewLazyLoader` helper would be a thin wrapper with no value-add. Capturing `c.Context()` at `RunE` construction time gives the per-command context the closure needs. Wrapping in `cmdutil.OnceValuesFunc` is required to preserve per-RunE caching — without it, any `RunE` that calls `opts.Library()` multiple times (e.g., `germinator library add` iterating over multiple resource refs) would re-load the library each call, which is a perf regression compared to the prior `f.Library = OnceValuesFunc(...)` wiring. The wrapper is already used throughout the codebase (per `golang-cli-architecture`'s "Cached Lazy Initialization" reference: *"calling `cfg.Config()` twice in one command invocation loads the config file twice. For expensive operations (disk reads, network calls), add `sync.Once` inside the factory function"*). This pattern matches `golang-cli-architecture`'s `CreateOptions.Client func()` lazy field pattern with caching (a closure captured at construction time, cached per-RunE; per design.md Decision 4; tasks.md 1.3-1.8, 1.14).

The `path` is computed once per `RunE` invocation via `library.FindLibrary(explicitPath, envPath, cfgPath)`. `cfgPath` comes from `cfg, cfgErr := f.Config()` called once at `RunE` entry (after the existing `runXxx` flag validation, before the `opts.Library` closure is constructed). If `cfgErr != nil`, the existing `runXxx` error path returns before the closure is built — the closure is unreachable in the error path. The closure captures both `c` (for `c.Context()`) and `path` (for the resolved location). The 3-arg `FindLibrary` encodes the spec-mandated precedence directly: flag > env > config-file > XDG default. This matches the existing `BuildFactory` wiring at `internal/cmdutil/factory.go:131-137`, with the only difference being that the wiring moves from `BuildFactory` time (capturing `f.RootContext`) to `RunE` time (capturing `c.Context()`).

**Alternatives considered**:

- *Add `library.NewLazyLoader(f, ctx, explicitPath) func() (*Library, error)`*: rejected; duplicates the existing `opts.Library` inline pattern with no behavioral benefit. Same churn at 9 sites, no reduction in LOC.
- *Reuse `f.Library` lazy field*: rejected; the field is preserved but left nil (mandated by `cli-cli-factory/spec.md`); the eager wiring is removed instead (Decision 6). The wiring captures `f.RootContext` which is the bug being fixed; the field itself is required by the spec.
- *Add a new `internal/lazy` package*: rejected; over-engineered.

### 5. `f.RootContext` vs `opts.Ctx` precedence

**Choice**: `opts.Ctx` (or `c.Context()` captured at construction time) is used in all cmd-side call sites that currently use `f.RootContext` **except** `cmd/completions.go:122`, which retains `f.RootContext` because it parents a `context.WithTimeout` for completion actions (cobra's `c.Context()` is cancelled when `RunE` returns, defeating the completion-timeout purpose). The Factory's `RootContext` is the signal-aware root; the per-command `Ctx` may derive from it (via `context.WithTimeout` for completion actions, or `context.WithCancel` for the runF pattern). `loadLibraryForCompletion` (`cmd/completions.go:106`) also legitimately uses `f.RootContext` because the completion cache lookup outlives a single `RunE`.

**Rationale**: The per-command `Ctx` is what the cmd-side wiring sets up. Using `f.RootContext` directly skips this layer in the 6 helper closures and the `f.Library` field. The completion path is the legitimate exception because the timeout parent must outlive `RunE`.

**Side effect**: `f.RootContext` is preserved on `Factory` for the completion path, but the eager `f.Library = OnceValuesFunc(...)` wiring is removed (Decision 6).

**Alternatives considered**:

- *Always use `f.RootContext` and derive `opts.Ctx` from it*: rejected; the per-command `Ctx` is the source of truth for cancellation in non-completion paths.
- *Remove `f.RootContext`*: rejected; the Factory field is used by `loadLibraryForCompletion` and `context.WithTimeout` parents in completion actions, both of which legitimately outlive a single `RunE`.

### 6. Drop the 6 per-command helpers and the `f.Library` eager wiring (keep the struct field)

**Choice**: Delete the 6 per-command `FindLibrary + LoadLibrary` helper functions from `cmd/`. Each `RunE` constructs the inline `opts.Library` closure directly per Decision 4. Additionally, remove the eager `f.Library = OnceValuesFunc(...)` wiring from `internal/cmdutil/factory.go:134-137` — but **preserve the `Library` lazy function field on the `Factory` struct**, as required by `cli-cli-factory/spec.md:25,30-32,105,125-126` (which mandates exactly `Config` and `Library` as exported `func() (T, error)` fields, enforced by `TestFactory_OnlyConfigAndLibraryAreLazyFields`). The field is left nil after `BuildFactory` returns; **callers MAY assign `f.Library` if they need a per-Factory cached loader**, but production code does not. The 3 read-only commands (`cmd/show.go`, `cmd/resources.go`, `cmd/presets.go`) already use the inline-closure pattern — verified by task 1.10, no production change required. The wording relaxation at `cli-cli-factory/spec.md:31` (the "SHALL be set" mandate) is documented in the new `specs/cli-cli-factory/spec.md` delta: the field is exposed with the correct signature, and `TestFactory_OnlyConfigAndLibraryAreLazyFields` continues to pass because only the BuildFactory wiring is removed.

**Rationale**: The 6 helpers are thin wrappers around the same `FindLibrary + LoadLibrary` sequence with verbose slice-7 comment blocks documenting obsolete history. Centralizing via the inline closure pattern (Decision 4) eliminates ~180 lines of helper code while keeping the same per-RunE lifecycle. After this change, all 9 sites use the same idiom.

The `f.Library` eager wiring at `internal/cmdutil/factory.go:134-137` captures `f.RootContext` at `BuildFactory` time, defeating per-command cancellation (the bug being fixed). Removing the wiring fixes the bug without violating `cli-cli-factory/spec.md`. The field itself is preserved (and required) by the spec; the wiring is the bug, not the field's existence.

Per-RunE caching is preserved by wrapping the inline closure in `cmdutil.OnceValuesFunc` (per Decision 4), so the only lost behavior is sharing the library instance across multiple `RunE` invocations within one Factory instance — which never occurs because the Factory is constructed once per CLI invocation.

**Side effects**:

- `cmd/AGENTS.md` Foundation Units table updates the Factory.Library row to clarify the field is preserved-but-unwired by this change.
- `cmdutil/factory_test.go` lazy-caching tests for `f.Library` are removed (the cached closure is no longer used by any caller).
- The `f.RootContext` Factory field is **preserved** for completion-time use (loadLibraryForCompletion + completion-timeout parents at `cmd/completions.go:106,122`).
- A new delta spec is added at `specs/cli-cli-factory/spec.md` to relax the "SHALL be set" wording to "SHALL be exposed with signature `func() (*library.Library, error)` and SHALL be assignable by callers" — preserving the field-presence contract while reflecting the new nil-by-default state.

**Alternatives considered**:

- *Keep the helpers as thin wrappers around the inline closure*: rejected; no value-add. The wrappers would just forward 3 parameters.
- *Inline the closure in each `RunE` without removing the helpers*: rejected; the helpers remain as dead code (unused after this change).
- *Remove the `Library` field from `Factory` entirely*: rejected; `cli-cli-factory/spec.md` mandates the field. A delta spec to remove the requirement would grow the change's surface unnecessarily.
- *Keep `f.Library` and rewire it to capture ctx lazily*: rejected; complicates the lazy-fn semantics (the closure would need to be re-built on every `c.Context()` change, which is impossible since `c` is not in scope at `BuildFactory` time). The simpler invariant (the field is preserved but nil by default; each RunE builds its own closure from `c.Context()`) wins.

### 7. Delete the package-level `ResolvePreset(lib, name)` shim (no dual-form)

**Choice**: Delete the package-level `func ResolvePreset(lib *Library, name string) ([]string, error)` at `internal/library/resolver.go:54-56` rather than threading `ctx` through it. All callers migrate to `(*Library).ResolvePreset(ctx, name)`.

**Rationale**: The shim's only purpose was to support non-migrated callers during the slice-5 → slice-7 migration window. That window is closed — the slice-7 method form `(*Library).ResolvePreset(ctx, name)` is the canonical form, and the pre-change spec scenario for "ResolvePreset signature" already cites the method form. Keeping a dual-form package function that synthesizes `context.TODO()` mid-request-path violates `golang-context` best practices #7 and #8, and the synthesis is invisible to the proposal's `_ context.Context` rg verification (which only catches the underscore binding pattern). The cleanest fix is to delete the shim and force callers to adopt the method form; no new option, no functional-options dance, just one canonical signature.

**Alternatives considered**:
- *Thread `ctx` through the shim*: rejected; bloats the public API with a function form that exists only for legacy callers. Once those callers are migrated, the shim is dead code.
- *Add a `func ResolvePresetWithContext(ctx, lib, name)` instead*: rejected; creates a third form that nobody calls. Same end-state (one canonical method form) with more surface area to maintain.
- *Keep the shim with `context.Background()`*: rejected; replaces a `context.TODO()` violation with a `context.Background()` violation. Both are `golang-context` best-practice violations; the change's verification rg (task 4.6) catches `context.Background()` but not `context.TODO()`, so this would silently regress the goal.

## Risks / Trade-offs

- **Signature break is widespread** — 4 cmd-layer adapters + 4 library package-level functions + 3 parser/renderer helpers + `ResolvePreset`. The migration is mechanical but spans 12+ files. (The original plan also referenced 4 shell packages from `extract-io-adapters`; those tasks are deferred until that change ships — see §Dependencies.) **Mitigation**: tasks are ordered so cmd-side changes (the small ones) ship first, then parser/renderer helpers, then library package-level functions. Each commit is independently testable against HEAD; `mise run build` catches missed call sites.
- **Test fixture churn** — every test that calls an updated function needs a `ctx` argument. The migration touches ~50+ test files. **Mitigation**: design.md Decision 2 uses `context.Background()` for existing tests as the minimum-churn approach.
- **`opts.Ctx` vs `f.RootContext`** — the 6 inline closures change the context source. For most commands, `opts.Ctx` is `f.RootContext`; for commands with per-call deadlines, `opts.Ctx` may be derived. **Mitigation**: each closure captures `c.Context()` at construction time; verify with `mise run test` that no behavior change occurs.
- **Cancellation pattern in `DiscoverOrphans`** — the sequential `ctx.Err()` checks at every directory scan and walker entry satisfy the rewritten spec scenario without an errgroup refactor. **Mitigation**: design.md Decision 3 documents the sequential pattern; `internal/library/adder.go:803,819,834,863` already implements it on HEAD.
- **`f.Library` wiring removal impacts `cmdutil/factory_test.go`** — lazy-caching tests for `f.Library` no longer apply (no caller invokes the wired closure after this change). **Mitigation**: task 1.1 explicitly removes those tests; the `Factory.Library` struct field is preserved (mandated by `cli-cli-factory/spec.md`) but left nil by `BuildFactory`.
- **`ResolvePreset` ctx threading** — adding ctx to `ResolvePreset` is for spec symmetry with the cli-framework requirement, not strictly for I/O cancellation safety. The function is currently a pure in-memory map lookup (`lib.Presets[name]`); the previous spec wording that referenced `os.ReadFile` was incorrect. **Mitigation**: design.md Decision 1 footnote documents the accept-and-may-ignore pattern (applied to both `ResolvePreset` and `DetectType`); the spec scenario is rewritten to match the actual implementation.
- **Package-level `ResolvePreset(lib, name)` shim uses `context.TODO()`** — the shim at `internal/library/resolver.go:54-56` is reachable from cmd adapters (via `library.ResolvePreset(lib, name)` calls) and synthesizes `context.TODO()` mid-request-path, violating `golang-context` best practices (#7: never pass `nil`, use `context.TODO()` only as a placeholder; #8: never create a new `context.Background()` mid-request). **Mitigation**: task 3.4b deletes the shim and migrates all callers; the rg in task 1.12 catches any future `context.TODO()` synthesis under `cmd/` or `internal/`.
- **Spec scenario style alignment** — the change's delta specs adopt the dash-prefixed uppercase style (`- **WHEN** / - **THEN** / - **GIVEN** / - **AND**`) consistently across all 10 deltas. The project's existing specs use two co-existing styles: dash-prefixed uppercase (the dominant style, used by every library parent and by `cli-cli-factory`) and no-dash mixed-case (used by some `cli-framework` and `infrastructure-*` scenarios). The `cli-framework` parent is itself mixed-style — both dash-prefixed and no-dash scenarios appear within it — so its delta uses dash-prefixed for the new requirement, consistent with one of the parent's two conventions and with the project-dominant style. **Mitigation**: task 4.13 audits all 10 delta specs with a positive `rg -c "^- \*\*WHEN\*\*"` check (each delta must have ≥1 dash-prefixed scenario). The library deltas and the `cli-cli-factory` delta match their parent's dash-prefixed uppercase style exactly.

## Migration Plan

The change ships in **one PR with 3 atomic phases** — see `tasks.md` sections 1, 2, 3 for the executable task breakdown. Each phase is independently testable against HEAD.

1. **Phase 1 — cmd-side inline closure + Factory.Library wiring removal** (tasks.md 1.x): drop the 6 per-command helpers; introduce the inline `opts.Library` closure pattern wrapped in `cmdutil.OnceValuesFunc` (per Decision 4; preserves per-RunE caching); remove the eager `f.Library = OnceValuesFunc(...)` wiring from `internal/cmdutil/factory.go:134-137` (the `Library` struct field is preserved but left nil per `cli-cli-factory/spec.md`); update `cmdutil/factory_test.go`; update `cmd/AGENTS.md`. Replace `f.RootContext` → captured `c.Context()` in the 6 call sites; wrap the 3 read-only commands (`show.go`, `resources.go`, `presets.go`) to unify all 9 sites on the same idiom (task 1.14).
2. **Phase 2 — parser + renderer ctx + ResolvePreset + test fakes** (tasks.md 2.x): add `ctx` to `parser.LoadDocument` / `DetectType` / `ParsePlatformDocument` and `renderer.RenderDocument` / `MarshalCanonical`; rename `_ context.Context` → `ctx` in `cmd/{initializer,transformer,canonicalize,validate}.go` and 4 test files; update `ResolvePreset` to accept ctx (accept-and-may-ignore pattern). **Tasks 2.6-2.9 from the original plan are deferred** until `extract-io-adapters` lands.
3. **Phase 3 — library package-level functions** (tasks.md 3.x): add `ctx` to `RefreshLibrary`, `RemoveResource`, `RemovePreset`; in `internal/library/creator.go:29`, change `CreateLibrary(opts)` to `CreateLibrary(ctx, opts, stdout io.Writer)` (absorbs both the stdout change from `fix-library-io-discipline` AND the ctx-first reorder in one step). Remove `TODO(slice-7)` markers.

**Rollback strategy**: revert each phase commit independently. Phase 1 restores the 6 helpers and the `f.Library` eager wiring (the struct field stays regardless). Phase 2 restores the prior parser/renderer signatures and the underscore bindings. Phase 3 restores the `context.Background()` calls and the `TODO(slice-7)` markers.