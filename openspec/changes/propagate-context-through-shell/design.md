## Context

The 2026-07-08 review identified 8 findings all stemming from the same root cause: the slice-7 design (`openspec/changes/archive/2026-07-01-migrate-library-rest/design.md:69-92`) added methods on `*library.Library` for `Refresh`, `RemoveResource`, `Validate`, `Fix` but left `context.Background()` hard-coded in the package-level functions that delegate to those methods. The TODO markers say "replace with caller context (c.Context() in runF wiring)" — a debt that has not been paid.

Simultaneously, the cmd-layer adapters (`initializer.go`, `transformer.go`, `canonicalize.go`, `validate.go`) accept `ctx context.Context` as a method parameter but discard it via the `_ context.Context` underscore binding. The `extract-io-adapters` change relocates these adapters to `internal/{install,transform,validate,canonicalize}/` shell packages but does not thread `ctx` through; the new `extract-io-adapters/specs/cli-framework/spec.md:42` promises:

> *"it SHALL take `ctx context.Context` as the first parameter of each public method"*

This change fulfills that promise.

The `cmd/library_add.go:253` line uses `f.RootContext` (the Factory-level signal-aware context) instead of `opts.Ctx` (the per-command context). The per-command context is what `runF` injection wires up; using `f.RootContext` skips the per-call deadline that the per-command `Ctx` field may impose. This is a single-line fix.

The review's A-012 finding identifies 5 per-command helpers (`addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`, `initLibrary`) that all repeat the same `FindLibrary + LoadLibrary` pattern. Folding this into a `library.NewLazyLoader` helper is a small refactor that reduces churn and prepares for the new `*Library` method signatures.

### Constraints

1. **No public API break for external consumers.** The library package is `internal/`; signature changes are acceptable.
2. **`extract-io-adapters` Stage 2** is in flight: it converts `AddResource` / `BatchAddResources` / `DiscoverOrphans` from package-level functions to methods on `*Library`. The `ctx` parameter must be present in both the package-level function and the `*Library` method.
3. **`fix-library-io-discipline`** is in flight: it adds an `io.Writer` parameter to `CreateLibrary`. The combined `CreateLibrary` signature is `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error`.
4. **Test surface churn** — every test calling an updated function needs a `ctx` argument. The minimum-churn approach is to use `context.Background()` in test files that don't test cancellation; new tests use `t.Context()` (Go 1.24+).

## Goals / Non-Goals

**Goals:**

- Retire all 4 `TODO(slice-7)` markers in `internal/library/{creator,refresher,remover}.go`.
- Thread `ctx` through the 4 cmd-layer adapters and the 4 new shell packages from `extract-io-adapters`.
- Replace `f.RootContext` with `opts.Ctx` in `cmd/library_add.go:253`.
- Add `ctx` parameter to `parser.LoadDocument` and `renderer.RenderDocument` (the downstream consumers).
- Promote the 5 per-command lazy-loader helpers to `library.NewLazyLoader`.

**Non-Goals:**

- Refactoring parser/renderer internals beyond adding the `ctx` parameter.
- Changing `Factory.RootContext` semantics.
- Removing the `*Library` method set introduced by slice 7.

## Decisions

### 1. `ctx` as first parameter (Go convention)

**Choice**: `ctx` is added as the first parameter to all updated functions, following Go convention (per `golang-context` skill: "all I/O functions take ctx as first parameter").

**Rationale**: Standard Go convention. Matches the existing pattern at the cmd-side: `runXxx(opts *XxxOptions) error` calls `library.Foo(opts.Ctx, ...)`. Avoids churn at the cmd layer.

**Alternatives considered**:

- *Add `ctx` as last parameter*: rejected; non-idiomatic; `go vet` would flag it.
- *Add a `Ctx` field to a request struct*: rejected; bloats the request struct for what is a per-call concern, not a per-request concern.

### 2. Test fixture pattern: `context.Background()` for existing, `t.Context()` for new

**Choice**: Existing test files use `context.Background()` as the minimum-churn approach. New tests added in this change use `t.Context()` (Go 1.24+).

**Rationale**: `t.Context()` is the modern pattern (auto-cancelled when the test ends), but migrating all existing tests is out of scope. New tests for ctx-cancellation behavior use `t.Context()` to make the test self-cleaning.

**Alternatives considered**:

- *Migrate all tests to `t.Context()`*: rejected; out of scope (covered by `harden-tests-and-coverage`).
- *Add a new `internal/testutil` package with `NewTestContext(t)` helper*: rejected; over-engineered for a single use case.

### 3. errgroup context propagation in `DiscoverOrphans`

**Choice**: `DiscoverOrphans` returns the `errgroup.Wait()` error directly. The function does NOT wrap `ctx.Err()`; the errgroup already does. The caller receives the wrapped error.

**Rationale**: The errgroup's `Wait()` returns the first non-nil error from any goroutine, or `nil` if all succeed. If the parent `ctx` is cancelled, the errgroup returns `ctx.Err()`. No additional wrapping is needed.

**Alternatives considered**:

- *Wrap `ctx.Err()` in a `*core.OperationError`*: rejected; the error chain is already preserved by the errgroup; wrapping loses information.
- *Distinguish between cancellation and other errors*: rejected; `errors.Is(err, context.Canceled)` is the standard Go pattern.

### 4. `library.NewLazyLoader` placement: in `internal/library/discovery.go`

**Choice**: `NewLazyLoader` is added to `internal/library/discovery.go` (next to `FindLibrary`). It returns a `func() (*Library, error)` that captures the `f *cmdutil.Factory` and the explicit path. On first call, it resolves the path via `FindLibrary(f.Library, "")` and loads the library; subsequent calls return the cached result.

**Rationale**: The lazy loader is a thin wrapper around `FindLibrary` + `LoadLibrary`. Co-locating it with `FindLibrary` keeps the discovery concern in one file. The signature is small and well-typed.

**Alternatives considered**:

- *Add to `cmdutil.Factory` as `f.LibraryLoader func() (*library.Library, error)`*: rejected; slice 7 removed all such lazy fields; the pattern is intentionally not reintroduced.
- *Add a new `internal/lazy` package*: rejected; over-engineered.

### 5. `f.RootContext` vs `opts.Ctx` precedence

**Choice**: `opts.Ctx` is used in all cmd-side call sites that currently use `f.RootContext`. The Factory's `RootContext` is the signal-aware root; the per-command `Ctx` may derive from it (via `context.WithTimeout` for completion actions, or `context.WithCancel` for the runF pattern).

**Rationale**: The per-command `Ctx` is what the cmd-side wiring sets up. Using `f.RootContext` directly skips this layer. The fix is a one-line swap.

**Alternatives considered**:

- *Always use `f.RootContext` and derive `opts.Ctx` from it*: rejected; the per-command `Ctx` is the source of truth for cancellation.
- *Remove `f.RootContext`*: rejected; the Factory field is used by completion and other cmd paths that don't have a per-command `Ctx`.

## Risks / Trade-offs

- **Signature break is widespread** — 4 cmd-layer adapters + 4 shell packages + 4 library package-level functions + parser + renderer. The migration is mechanical but spans 14+ files. **Mitigation**: tasks are ordered so cmd-side changes (the small ones) ship first, then shell packages, then library package-level functions. Each commit is independently testable; `mise run build` catches missed call sites.
- **Test fixture churn** — every test that calls an updated function needs a `ctx` argument. The migration touches ~50+ test files. **Mitigation**: design Decision 2 uses `context.Background()` for existing tests as the minimum-churn approach.
- **`opts.Ctx` vs `f.RootContext`** — the `cmd/library_add.go:253` fix changes the context source. For most commands, `opts.Ctx` is `f.RootContext`; for commands with per-call deadlines, `opts.Ctx` may be derived. **Mitigation**: the change is a one-line swap; verify with `mise run test` that no behavior change occurs.
- **errgroup context in `DiscoverOrphans`** — the `errgroup.WithContext(ctx)` from `fix-library-io-discipline` derives a child context. **Mitigation**: design Decision 3 returns the errgroup error directly; no additional wrapping.

## Migration Plan

The change ships in **one PR with 3 atomic phases** (each commit is independently testable):

1. **Phase 1 — cmd-side small fixes** (tasks 4.1, 4.2, 4.3): `f.RootContext` → `opts.Ctx` in `cmd/library_add.go:253`; add `library.NewLazyLoader`; refactor the 5 per-command helpers. Verify `mise run test`.
2. **Phase 2 — parser + renderer ctx** (tasks 4.4, 4.5): add `ctx` to `parser.LoadDocument` and `renderer.RenderDocument`; thread `ctx` through the 4 new shell packages from `extract-io-adapters`. Verify `mise run build` and `mise run test`.
3. **Phase 3 — library package-level functions** (tasks 4.6, 4.7, 4.8, 4.9): add `ctx` to `RefreshLibrary`, `RemoveResource`, `RemovePreset`, `CreateLibrary`; remove `TODO(slice-7)` markers. Verify `mise run test` and `mise run test:e2e`.

**Rollback strategy**: revert each phase commit independently. Phase 1 is a small refactor (revert restores `f.RootContext` and the 5 helpers). Phase 2 is signature changes (revert restores the prior signatures). Phase 3 is signature changes (revert restores the `context.Background()` calls and the `TODO(slice-7)` markers).
