## Why

The 2026-07-08 code review identified **8 context-propagation findings** (C-001..C-004, BCD-009, C-007, C-008, D-009) that span the cmd layer and the library package. The root cause is the slice-7 design (per `archive/2026-07-01-migrate-library-rest/design.md:69-92`), which added methods on `*library.Library` for `Refresh`, `RemoveResource`, `Validate`, `Fix` but left `context.Background()` hard-coded in:

- `internal/library/refresher.go:58` — `RefreshLibrary` package-level function
- `internal/library/remover.go:62` — `RemoveResource` package-level function
- `internal/library/remover.go:127` — `RemovePreset` package-level function
- `internal/library/creator.go:29` — `CreateLibrary` package-level function

Each of these has a `// TODO(slice-7): replace with caller context` marker. Additionally, 4 cmd-layer adapters (`cmd/initializer.go:53`, `cmd/transformer.go:44`, `cmd/canonicalize.go:147`, `cmd/validate.go:134`) accept `ctx context.Context` but discard it (`_ context.Context`); the adapters call `parser.LoadDocument` and `renderer.RenderDocument` which are blocking I/O. The `extract-io-adapters` change relocates these adapters to `internal/{install,transform,validate,canonicalize}/` but does not thread `ctx` through.

The 6 helper closures in `cmd/library_*.go` and `cmd/init.go` (`addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`, `initLibrary`) wrap `FindLibrary + LoadLibrary` and capture `f.RootContext` rather than the per-call `c.Context()` — defeating the per-command cancellation that `ctx` is supposed to provide. The `f.Library` field's eager wiring at `internal/cmdutil/factory.go:134-137` repeats this bug; the struct field itself is preserved per `cli-cli-factory/spec.md`.

This change propagates `ctx` through the shell-package boundaries so cancellation signals from the cmd layer reach the parser and renderer call sites. It is a **production-code refactor** with spec deltas because the `cli-framework` and library specs promise ctx-aware behavior that the code does not yet deliver.

## Dependencies

- **`fix-library-io-discipline`** — absorbed by this change. The current `CreateLibrary` signature on `git HEAD` is `func CreateLibrary(opts CreateOptions) error`; this change introduces the `stdout io.Writer` parameter AND the ctx-first reorder in one step, replacing what would have been two sequential changes. The end-state signature is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` per Go convention (design.md Decision 1).
- **`extract-io-adapters`** — NOT applied. The original Phase 2 tasks 2.6-2.9 (which reference `internal/{transform,install,validate,canonicalize}/`) are deferred until that change ships. The cmd-side adapter underscore-binding cleanup (`cmd/initializer.go:53`, etc.) remains in scope because those files exist on HEAD.

## What Changes

Changes are grouped by package to mirror the dependency graph: `cmd/` wires the per-command context into the shell; `internal/cmdutil/` removes the eager wiring; `internal/library/` accepts ctx in package-level functions; `internal/parser/` and `internal/renderer/` forward ctx into blocking I/O.

### `internal/library/` (7 changes)

- **MODIFY** `internal/library/refresher.go:58` — change `func RefreshLibrary(opts RefreshOptions) (*RefreshResult, error)` to `func RefreshLibrary(ctx context.Context, opts RefreshOptions) (*RefreshResult, error)`. Remove the `context.Background()` line and the `// TODO(slice-7)` marker.
- **MODIFY** `internal/library/remover.go:62` — change `func RemoveResource(opts RemoveResourceOptions) (*RemoveResourceOutput, error)` to `func RemoveResource(ctx context.Context, opts RemoveResourceOptions) (*RemoveResourceOutput, error)`. Remove the `context.Background()` line (line 64) and the `// TODO(slice-7)` marker (line 63).
- **MODIFY** `internal/library/remover.go:127` — change `func RemovePreset(opts RemovePresetOptions) (*RemovePresetOutput, error)` to `func RemovePreset(ctx context.Context, opts RemovePresetOptions) (*RemovePresetOutput, error)`. Remove the `context.Background()` line (line 129) and the `// TODO(slice-7)` marker (line 128).
- **MODIFY** `internal/library/creator.go:29` — current signature `func CreateLibrary(opts CreateOptions) error` becomes `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` (absorbs the `stdout` parameter that `fix-library-io-discipline` would have introduced, plus the ctx-first reorder per Go convention / design.md Decision 1). Remove the `context.Background()` line (line 71) and the `// TODO(slice-7)` marker (line 70).
- **MODIFY** `internal/library/resolver.go:67` — change `func (lib *Library) ResolvePreset(_ context.Context, name string) ([]string, error)` to `func (lib *Library) ResolvePreset(ctx context.Context, name string) ([]string, error)`. Per design.md Decision 1 footnote: `ctx` is accept-and-may-ignore (current body is a pure in-memory map lookup; no `os.ReadFile` exists at `resolver.go:67-73`). If the resolution path is extended to perform I/O in the future, `ctx` is in place to forward to it. Removes the underscore binding.
- **MODIFY** `internal/library/resolver.go:54-56` — delete the package-level `func ResolvePreset(lib *Library, name string) ([]string, error)`. The shim synthesizes `context.TODO()` mid-request-path (violates `golang-context` rule #8 — never create a new `context.Background()` mid-request), and the synthesis is invisible to the `_ context.Context` rg verification (it uses `context.TODO()`, not `_ context.Context`). All callers migrate to `(*Library).ResolvePreset(ctx, name)`. The only callers are in-package (unqualified) in `internal/library/resolver_test.go`: lines 125 and 138 (`TestResolvePreset`, `TestResolvePreset_NotFound`) and lines 209 and 218 (`TestResolvePreset_PackageShimDelegatesToMethod`); all three shim-only tests are deleted by task 3.4b (the method-form tests at lines 144-197 already cover the same behavior). No external `cmd/` callers exist (verified via `rg "\bResolvePreset\(lib," cmd/ internal/library/` — the unqualified pattern is required because in-package calls omit the `library.` qualifier).
- **MODIFY** `internal/library/adder.go` — no signature change. `AddResource`, `BatchAddResources`, and `DiscoverOrphans` already accept `ctx context.Context` as the first parameter (per slice 7, see `internal/library/adder.go:48,558,785`); this change enforces forward propagation through the entire call chain (no synthesis of `context.Background()` or `context.TODO()` at any downstream call site).

### `internal/cmdutil/` (1 change)

- **MODIFY** `internal/cmdutil/factory.go:134-137` — remove the `f.Library = OnceValuesFunc(...)` eager wiring (which captures `f.RootContext` at `BuildFactory` time, defeating per-command cancellation). The `Library` lazy function field is **preserved on the `Factory` struct** as required by `cli-cli-factory/spec.md:25,30-32,105,125-126` (which mandates exactly `Config` and `Library` as exported `func() (T, error)` fields); the field is left nil so each `RunE` builds its own `opts.Library` closure from `c.Context()`. Update `cmdutil/factory_test.go` to drop the `f.Library` lazy-caching tests, and update `cmd/AGENTS.md` Foundation Units table to clarify the field is preserved-but-unwired.

### `internal/parser/` (2 changes)

- **MODIFY** `internal/parser/loader.go:29` — `LoadDocument(filepath, platform string) (interface{}, error)` becomes `LoadDocument(ctx context.Context, filepath, platform string) (interface{}, error)`. The function checks `ctx.Err()` between file reads.
- **MODIFY** `internal/parser/platform_parser.go:14` — `ParsePlatformDocument(path, platform, docType string) (interface{}, error)` becomes `ParsePlatformDocument(ctx context.Context, path, platform, docType string) (interface{}, error)`. File read uses the forwarded ctx.
- **MODIFY** `internal/parser/loader.go:69` — `DetectType(filepath string) string` becomes `DetectType(ctx context.Context, filepath string) string`. The regex loop checks `ctx.Err()` between iterations. (Returns `string`, not `error`; ctx is accept-and-may-ignore per design.md Decision 1 footnote.)

### `internal/renderer/` (2 changes)

- **MODIFY** `internal/renderer/serializer.go:30` — `RenderDocument(doc interface{}, platform string) (string, error)` becomes `RenderDocument(ctx context.Context, doc interface{}, platform string) (string, error)`. The function checks `ctx.Err()` once at entry.
- **MODIFY** `internal/renderer/serializer.go:232` — `MarshalCanonical(doc interface{}) (string, error)` becomes `MarshalCanonical(ctx context.Context, doc interface{}) (string, error)`. YAML serialization checks ctx before writing.

### `cmd/` (5 changes)

- **MODIFY** `cmd/initializer.go:53`, `cmd/transformer.go:44`, `cmd/canonicalize.go:147`, `cmd/validate.go:134` — change the parameter from `_ context.Context` to `ctx context.Context`. (These cmd-side adapter files exist on HEAD; if `extract-io-adapters` deletes them before this change runs, tasks 2.6-2.9 in `tasks.md` are no-ops. The phase-2 cmd-side rename remains in scope regardless.)
- **MODIFY** `cmd/{validate,init,canonicalize,adapt}_test.go` — rename `_ context.Context` to `ctx context.Context` in the 4 fake adapter implementations (zero behavior change; satisfies the spec scenario rg).
- **MODIFY** `cmd/library_add.go`, `cmd/library_create.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go`, `cmd/init.go` — delete the 6 per-command helpers (`addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`, `initLibrary`). Each `RunE`:
  - Captures `cfg, cfgErr := f.Config()` once at entry (after existing flag validation, before the closure is constructed; `cfgErr` is handled by the existing error path).
  - Resolves `path := library.FindLibrary(explicitPath, os.Getenv("GERMINATOR_LIBRARY"), cfg.Library)` once per invocation.
  - Builds `opts.Library = cmdutil.OnceValuesFunc(func() (*library.Library, error) { return library.LoadLibrary(c.Context(), path) })` inline, capturing `c.Context()` (per design.md Decision 4; tasks.md 1.14; preserves per-RunE caching — without the wrapper, any `RunE` that calls `opts.Library()` multiple times re-loads each time).
- **MODIFY** `cmd/AGENTS.md` — update the Factory.Library row in the Foundation Units table to clarify the field is **preserved** (signature unchanged per `cli-cli-factory/spec.md:25,30-32,105,125-126`) but `BuildFactory` no longer wires it; each `RunE` builds its own closure from `c.Context()` (see design.md Decision 6 and the new `cli-cli-factory` delta under §Capabilities below).
- **MODIFY** `cmd/show.go`, `cmd/resources.go`, `cmd/presets.go` — these 3 read-only commands already use the inline `opts.Library` closure pattern with `opts.Ctx` set from `c.Context()` at lines `show.go:93-96`, `resources.go:84-87`, `presets.go:74-77`. Production change is limited to wrapping each closure in `cmdutil.OnceValuesFunc` per task 1.14 (unifies all 9 sites on the same idiom; preserves per-RunE caching).

### Scope expansion: parser + renderer helpers for spec symmetry

The `cli-framework` requirement at `specs/cli-framework/spec.md:30` mandates that **every** public method on the new shell adapters SHALL have `ctx` as a named, non-underscore parameter. Three additional parser/renderer helpers that the original `What Changes` list did not enumerate also flow through this change so the spec scenario holds:

- `internal/parser/loader.go:69` — `DetectType(filepath string) string` becomes `DetectType(ctx context.Context, filepath string) string`. The regex loop checks `ctx.Err()` between iterations. (Returns `string`, not `error`, because detection is regex-only; ctx is checked but not forwarded to any I/O.)
- `internal/parser/platform_parser.go:14` — `ParsePlatformDocument(path, platform, docType string) (interface{}, error)` becomes `ParsePlatformDocument(ctx context.Context, path, platform, docType string) (interface{}, error)`. File read uses the forwarded ctx.
- `internal/renderer/serializer.go:232` — `MarshalCanonical(doc interface{}) (string, error)` becomes `MarshalCanonical(ctx context.Context, doc interface{}) (string, error)`. YAML serialization checks ctx before writing.

These three additions correspond to tasks.md 2.2, 2.3, 2.5.

## Capabilities

### New Capabilities

- **`library-library-resolution`** — `ResolvePreset` SHALL accept `ctx context.Context` as the first parameter and SHALL forward it to any I/O performed during resolution. (Currently has `_ context.Context` underscore binding; this change enforces the same ctx-propagation contract as the other `*Library` methods. Resolution is currently a pure in-memory map lookup, so `ctx` is accepted and may be ignored per the accept-and-may-ignore pattern ADDED by this change's `cli-framework` delta — see requirement `I/O adapter ctx propagation` in `specs/cli-framework/spec.md` within this change.)

### Modified Capabilities

- **`cli-framework`** — add explicit requirement that I/O adapter `Service` methods SHALL take `ctx context.Context` as the first parameter and forward it to parser/renderer call sites. This formalizes the promise in `extract-io-adapters/specs/cli-framework/spec.md:42`.
- **`cli-cli-factory`** — allow `Factory.Library` to be nil-by-default (the field is preserved per the existing `cli-cli-factory/spec.md:25,30-32,105,125-126` mandate, but `BuildFactory` no longer wires it). The "SHALL be set" wording at `cli-cli-factory/spec.md:31` is relaxed to "SHALL be exposed with signature `func() (*library.Library, error)` and SHALL be assignable by callers." The contract test `TestFactory_OnlyConfigAndLibraryAreLazyFields` continues to pass because the field exists with the correct signature; only the `BuildFactory` wiring is removed (per design.md Decision 6; tasks.md 1.1-1.2, 4.10; new `specs/cli-cli-factory/spec.md` delta).
- **`library-library-refresh`** — `RefreshLibrary` SHALL accept `ctx` as the first parameter; the package-level function delegates to a `*Library.Refresh` method that uses the caller's `ctx`.
- **`library-library-remove-resource`** — `RemoveResource` SHALL accept `ctx` as the first parameter.
- **`library-library-remove-preset`** — `RemovePreset` SHALL accept `ctx` as the first parameter.
- **`library-library-scaffolding`** — `CreateLibrary` SHALL accept `ctx` as the first parameter. The combined signature with `fix-library-io-discipline` is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` — ctx first per Go convention (see design.md Decision 1). This change reorders from the `fix-library-io-discipline` shipping signature `(opts, stdout)`.
- **`library-library-resource-import`** — `AddResource` SHALL forward the caller's `ctx` to all I/O. (The package-level function already accepts ctx per slice 7; this change clarifies that the underlying `*Library.Add` method — when introduced by `extract-io-adapters` Stage 2 — and the `libraryAdapter` in `cmd/library_add.go:82` MUST use it without synthesizing `context.Background()`.)
- **`library-library-batch-add`** — `BatchAddResources` SHALL forward the caller's `ctx` to all I/O. (Same shape as `AddResource`.)
- **`library-library-orphan-discovery`** — `DiscoverOrphans` SHALL honor caller-supplied cancellation by checking `ctx.Err()` at every top-level directory scan and between every per-file walker entry. (Package-level function already accepts ctx per slice 7; the sequential per-directory pattern is implemented on HEAD at `adder.go:803,819,834,863` and is preserved by this change per design.md Decision 3.)

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `internal/library/refresher.go:58` | Add `ctx` param | +3 / -2 |
| `internal/library/remover.go:62,127` | Add `ctx` param | +6 / -4 |
| `internal/library/creator.go:29` | Reorder to `(ctx, opts, stdout)` | +3 / -2 |
| `internal/library/resolver.go:67` | Rename `_` → `ctx` | +1 / -1 |
| `internal/library/resolver.go:54-56` | Delete package-level `ResolvePreset(lib, name)` shim; migrate callers | -3 / +3 (test call-site updates) |
| `internal/parser/loader.go:29` (`LoadDocument`) | Add `ctx` param | +5 / -2 |
| `internal/parser/loader.go:69` (`DetectType`) | Add `ctx` param | +3 / -1 |
| `internal/parser/platform_parser.go:14` | Add `ctx` param | +5 / -2 |
| `internal/renderer/serializer.go:30` (`RenderDocument`) | Add `ctx` param | +5 / -2 |
| `internal/renderer/serializer.go:232` (`MarshalCanonical`) | Add `ctx` param | +3 / -1 |
| `internal/library/adder.go:48,558,785` | No signature change; verify forwarded | +0 |
| `cmd/initializer.go`, `cmd/transformer.go`, `cmd/canonicalize.go`, `cmd/validate.go` | Rename `_ context.Context` → `ctx` | +4 / -4 |
| `cmd/{validate,init,canonicalize,adapt}_test.go` | Rename `_` → `ctx` in fakes | +4 / -4 |
| `cmd/library_add.go`, `cmd/library_create.go`, `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_validate.go`, `cmd/init.go` | Delete 6 helpers; inline `cmdutil.OnceValuesFunc`-wrapped closure in `RunE` | -180 / +57 |
| `internal/cmdutil/factory.go:134-137` | Remove `f.Library` eager wiring (field preserved) | -10 |
| `internal/cmdutil/factory.go` (struct) | `Library` field preserved (mandated by `cli-cli-factory/spec.md`) | 0 |
| `internal/cmdutil/factory_test.go` | Drop `f.Library` lazy-caching tests | -20 |
| `cmd/AGENTS.md` | Update Factory.Library row (field preserved, wiring removed) | 0 |
| `cmd/show.go`, `cmd/resources.go`, `cmd/presets.go` | No production change (verify) | 0 |
| `internal/library/AGENTS.md` | Audit and update every public signature (CreateLibrary, RefreshLibrary, RemoveResource, RemovePreset, ResolvePreset, slice-7 `(*Library)` methods) | +10 / -5 |
| `specs/cli-cli-factory/spec.md` (NEW delta) | Add delta spec relaxing "SHALL be set" to "SHALL be exposed and assignable" | +20 / 0 |

_Shell-package ctx propagation (4 files in `internal/{transform,install,validate,canonicalize}/`) is owned by `extract-io-adapters` and is out of scope for this change._

### Affected systems

- **Cancellation latency:** the cmd layer's `ctx` is now respected by parser and renderer. A user pressing Ctrl-C during `germinator validate` or `germinator adapt` sees the operation terminate within one file-read time, not at the cmd-side post-call return.
- **Test surface:** every test that calls `AddResource`, `RemoveResource`, `RefreshLibrary`, `CreateLibrary`, `LoadDocument`, `RenderDocument`, `ResolvePreset` must be updated to pass a `ctx`. The test pattern becomes `ctx := context.Background()` (or `t.Context()` for new tests) as the first argument.
- **Public API:** all the affected functions are in `internal/` packages, so signature changes are acceptable. The `cli-framework` spec is updated to formalize the contract.
- **Factory surface:** `f.Library` eager wiring is removed; the struct field is preserved per `cli-cli-factory/spec.md` and left nil by `BuildFactory`. `cmdutil/factory_test.go` lazy-caching tests for `f.Library` are removed (the wired closure is no longer invoked by any caller).
- **Lint baseline:** expected unchanged (no production-code patterns that affect `golangci-lint`).

## Risks

- **Signature break is widespread** — 4 cmd-layer adapters + 4 library package-level functions + 3 parser/renderer helpers + `ResolvePreset`. The migration is mechanical but spans 12+ files. (The 4 shell packages from `extract-io-adapters` are out of scope per §Dependencies.) **Mitigation**: tasks are ordered so cmd-side changes ship first, then parser/renderer helpers, then library package-level functions. Each commit is independently testable against HEAD; `mise run build` catches missed call sites.
- **Test fixture churn** — every test that calls an updated function needs a `ctx` argument. The migration touches ~50+ test files. **Mitigation**: design.md Decision 2 evaluates the `t.Context()` pattern (Go 1.24+) for new tests; existing tests use `context.Background()` for the minimum churn.
- **`opts.Ctx` vs `f.RootContext`** — see design.md Decision 5; tasks.md 1.9, 4.8. The per-command `c.Context()` is the source of truth for non-completion paths; `cmd/completions.go:106,122` legitimately retain `f.RootContext` for completion cache lookups and `context.WithTimeout` parents that must outlive a single `RunE`. **Mitigation**: each closure captures `c.Context()` at construction time; verify with `mise run test` that no behavior change occurs.
- **Cancellation pattern in `DiscoverOrphans`** — the sequential `ctx.Err()` checks at every directory scan and walker entry satisfy the rewritten spec scenario without an errgroup refactor. The current implementation at `internal/library/adder.go:803,819,834,863` already provides bounded cancellation via `ctx.Err()` checks; this change formalizes the contract. **Mitigation**: design.md Decision 3 documents the pattern; errgroup refactor is a follow-up change if I/O latency ever becomes user-perceptible.
- **`CreateLibrary` signature absorbs `fix-library-io-discipline`** — the `stdout io.Writer` parameter that change would have introduced is folded into this change's task 3.4 alongside the ctx-first reorder. The combined signature is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` in one step rather than two sequential changes. **Mitigation**: the work is a single-line signature change plus call-site updates; `cmd/library_init.go:161` already passes `opts.Ctx` so the cmd-side wiring needs no edit.
- **`f.RootContext` retained in `cmd/completions.go:122`** — completion actions use `context.WithTimeout(f.RootContext, ...)` to bound lookup latency. Cobra's `c.Context()` is cancelled when `RunE` returns, which would defeat the completion-timeout parent. The `f.RootContext` use here is intentional and documented at `cmd/completions.go:90-97`. **Mitigation**: task 1.8 narrows the verification rg to `cmd/(library_*.go|init.go)` to exclude this file.
- **`f.RootContext` retained in `internal/cmdutil/factory.go` for completion** — `loadLibraryForCompletion` (called from `cmd/completions.go:106`) legitimately uses `f.RootContext` because the completion cache lookup outlives a single `RunE`. The `f.RootContext` field is preserved on Factory for this purpose. **Mitigation**: the eager `f.Library` wiring is removed (no other consumer); `f.RootContext` itself remains a Factory field.
- **`DetectType` and `ResolvePreset` ctx parameters are unusual** — both accept `ctx` for spec symmetry with the `cli-framework` requirement, not strictly for cancellation safety. `DetectType` is regex-only (no I/O); `ResolvePreset` is an in-memory map lookup (no `os.ReadFile` despite the spec scenario's prior wording). **Mitigation**: design.md Decision 1 footnote documents the accept-and-may-ignore pattern; `ResolvePreset` is rewritten to match (see `specs/library-library-resolution/spec.md`). If either path later acquires I/O, the `ctx` parameter is already in place to forward to it.
- **`cli-cli-factory/spec.md:31` "SHALL be set" wording** — the existing `cli-cli-factory/spec.md:31` mandates `Factory.Library func() (*library.Library, error)` SHALL be set; this change removes the `BuildFactory` wiring and leaves the field nil-by-default. **Mitigation**: a delta spec is added under `specs/cli-cli-factory/spec.md` that relaxes the wording to "SHALL be exposed with signature `func() (*library.Library, error)` and SHALL be assignable by callers"; `TestFactory_OnlyConfigAndLibraryAreLazyFields` continues to pass because the field exists with the correct signature.
- **Package-level `ResolvePreset` shim uses `context.TODO()` mid-request-path** — `internal/library/resolver.go:54-56` synthesizes `context.TODO()` in a code path reachable from cmd adapters, violating `golang-context` rule #7 (never pass `nil`; use `context.TODO()` only as a placeholder) and rule #8 (never create a new `context.Background()` mid-request); see `golang-context/references/cancellation.md`. The synthesis is invisible to the proposal's `_ context.Context` rg verification. **Mitigation**: task 3.4b deletes the shim and migrates all callers to `(*Library).ResolvePreset(ctx, name)`; the rg in task 1.12 catches `context.TODO()` in production code under `cmd/` and `internal/`.
- **Spec scenario style alignment** — the change's delta specs adopt the dash-prefixed uppercase style (`- **WHEN** / - **THEN** / - **GIVEN** / - **AND**`) consistently across all 10 deltas. The project's existing specs use two co-existing styles: dash-prefixed uppercase (the dominant style, used by every library parent and by `cli-cli-factory`) and no-dash mixed-case (used by some `cli-framework` and `infrastructure-*` scenarios). The `cli-framework` parent is itself mixed-style (both dash-prefixed and no-dash scenarios appear within it); its delta uses dash-prefixed for the new requirement, consistent with one of the parent's two conventions and with the project-dominant style. **Mitigation**: task 4.13 audits all 10 delta specs with a positive `rg -c "^- \*\*WHEN\*\*"` check (each delta must have ≥1 dash-prefixed scenario). The library deltas and `cli-cli-factory` delta match their parent's dash-prefixed uppercase style exactly.

## Goals / Non-Goals

**Goals:**

- Retire all 4 `TODO(slice-7)` markers in `internal/library/{creator,refresher,remover}.go`.
- Update `internal/library/AGENTS.md` to document every public signature change — `CreateLibrary`, `RefreshLibrary`, `RemoveResource`, `RemovePreset`, `ResolvePreset` (method only), and the slice-7 `(*Library)` method set — verified against post-change code via `rg "^func " internal/library/*.go` (see tasks.md 4.12). The new `CreateLibrary` signature is `(ctx context.Context, opts CreateOptions, stdout io.Writer) error`.
- Thread `ctx` through the 4 cmd-layer adapters. (The 4 new shell packages from `extract-io-adapters` are that change's responsibility per §Dependencies.)
- Replace `f.RootContext` with the captured per-command context in the 6 helper sites listed under What Changes, and wrap each `RunE`'s inline `opts.Library` closure in `cmdutil.OnceValuesFunc` across all 9 sites (per design.md Decision 4; preserves per-RunE caching).
- Drop the 6 per-command lazy-loader helpers and centralize via the inline `opts.Library` closure pattern (see design.md Decision 6).
- Remove the `f.Library` eager wiring from `internal/cmdutil/factory.go:134-137` (preserving the struct field per `cli-cli-factory/spec.md` and adding the delta spec to relax the "SHALL be set" wording).
- Update `ResolvePreset` to accept ctx (removes the `_ context.Context` underscore binding).
- Delete the package-level `ResolvePreset(lib, name)` shim at `internal/library/resolver.go:54-56`; migrate all callers to `(*Library).ResolvePreset(ctx, name)` (eliminates the `context.TODO()` mid-request-path synthesis).

**Non-Goals:**

- Refactoring parser/renderer internals beyond adding the `ctx` parameter (the change adds entry-level checks and threads ctx into the inner `os.ReadFile`; it does not refactor per-IO cancellation deeper).
- Changing the `Factory.RootContext` semantics (it remains the signal-aware root context; `opts.Ctx` is the per-call derivative).
- Removing the `*Library` method set introduced by slice 7 (the methods are preserved; the new shell packages from `extract-io-adapters` are the canonical consumers).
- Changing the 3 read-only commands (`show.go`, `resources.go`, `presets.go`) — they already conform to the inline-closure pattern.
- Removing `f.RootContext` from `Factory` — it remains for `loadLibraryForCompletion` and direct completion-time use.