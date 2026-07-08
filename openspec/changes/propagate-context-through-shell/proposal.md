## Why

The 2026-07-08 code review identified **8 context-propagation findings** (C-001..C-004, BCD-009, C-007, C-008, D-009) that span the cmd layer and the library package. The root cause is the slice-7 design (per `archive/2026-07-01-migrate-library-rest/design.md:69-92`), which added methods on `*library.Library` for `Refresh`, `RemoveResource`, `Validate`, `Fix` but left `context.Background()` hard-coded in:

- `internal/library/refresher.go:60` — `RefreshLibrary` package-level function
- `internal/library/remover.go:64` — `RemoveResource` package-level function
- `internal/library/remover.go:129` — `RemovePreset` package-level function
- `internal/library/creator.go:71` — `CreateLibrary` package-level function

Each of these has a `// TODO(slice-7): replace with caller context` marker. Additionally, 4 cmd-layer adapters (`cmd/initializer.go:53`, `cmd/transformer.go:44`, `cmd/canonicalize.go:147`, `cmd/validate.go:134`) accept `ctx context.Context` but discard it (`_ context.Context`); the adapters call `parser.LoadDocument` and `renderer.RenderDocument` which are blocking I/O. The `extract-io-adapters` change relocates these adapters to `internal/{install,transform,validate,canonicalize}/` but does not thread `ctx` through.

The `cmd/library_add.go:253` line uses `f.RootContext` (the Factory-level context) instead of `opts.Ctx` (the per-command context), defeating the per-command cancellation that `ctx` is supposed to provide.

This change propagates `ctx` through the shell-package boundaries so cancellation signals from the cmd layer reach the parser and renderer call sites. It is a **production-code refactor** with spec deltas because the `cli-framework` and library specs promise ctx-aware behavior that the code does not yet deliver.

## What Changes

- **MODIFY** `internal/library/refresher.go:60` — change `RefreshLibrary(opts RefreshOptions) error` to `RefreshLibrary(ctx context.Context, opts RefreshOptions) error`. Remove the `context.Background()` line and the `// TODO(slice-7)` marker.
- **MODIFY** `internal/library/remover.go:64` — change `RemoveResource(opts RemoveResourceOptions) error` to `RemoveResource(ctx context.Context, opts RemoveResourceOptions) error`. Same.
- **MODIFY** `internal/library/remover.go:129` — change `RemovePreset(opts RemovePresetOptions) error` to `RemovePreset(ctx context.Context, opts RemovePresetOptions) error`. Same.
- **MODIFY** `internal/library/creator.go:71` — change `CreateLibrary(opts CreateOptions) error` to `CreateLibrary(ctx context.Context, opts CreateOptions) error`. Same. (Note: this signature also changes in `fix-library-io-discipline` to add `stdout io.Writer`; the combined change is `CreateLibrary(ctx, opts, stdout)`.)
- **MODIFY** `internal/library/adder.go` (in the methods on `*Library` introduced by `extract-io-adapters` Stage 2): `Add`, `BatchAddResources`, `DiscoverOrphans` methods MUST accept and use the receiver's stored context. Specifically, the methods take a `ctx` parameter even though the receiver has `lib.RootPath` — the context is the per-call context, not the receiver's.
- **MODIFY** `internal/parser/loader.go:69` — `LoadDocument(inputPath, platform) (*Document, error)` becomes `LoadDocument(ctx context.Context, inputPath, platform) (*Document, error)`. The function checks `ctx.Err()` between file reads.
- **MODIFY** `internal/renderer/serializer.go:59` — `RenderDocument(doc, platform) ([]byte, error)` becomes `RenderDocument(ctx context.Context, doc, platform) ([]byte, error)`. (The template cache lookup is fast and doesn't need ctx checks; the document write is done by the cmd layer.)
- **MODIFY** `internal/transform/transform.go`, `internal/install/install.go`, `internal/validate/validate.go`, `internal/canonicalize/canonicalize.go` (the 4 new shell packages from `extract-io-adapters`): each `Service` method takes `ctx context.Context` as the first parameter and forwards it to `parser.LoadDocument` and `renderer.RenderDocument`.
- **MODIFY** `cmd/library_add.go:253` — replace `library.LoadLibrary(f.RootContext, resolved)` with `library.LoadLibrary(opts.Ctx, resolved)`. The `opts.Ctx` is captured in the `addLibrary` closure.
- **FOLD** A-012 — promote the 5 per-command `FindLibrary + LoadLibrary` helpers (`addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`, `initLibrary`) to a single `library.NewLazyLoader(f *Factory, explicitPath string) func() (*Library, error)`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- **`cli-framework`** — add explicit requirement that I/O adapter `Service` methods SHALL take `ctx context.Context` as the first parameter and forward it to parser/renderer call sites. This formalizes the promise in `extract-io-adapters/specs/cli-framework/spec.md:42`.
- **`library-library-refresh`** — `RefreshLibrary` SHALL accept `ctx` as the first parameter; the package-level function delegates to a `*Library.Refresh` method that uses the caller's `ctx`.
- **`library-library-remove-resource`** — `RemoveResource` SHALL accept `ctx` as the first parameter.
- **`library-library-remove-preset`** — `RemovePreset` SHALL accept `ctx` as the first parameter.
- **`library-library-scaffolding`** — `CreateLibrary` SHALL accept `ctx` as the first parameter.
- **`library-library-resource-import`** — `AddResource` SHALL accept `ctx` as the first parameter (already does, but the underlying method on `*Library` must use it).
- **`library-library-batch-add`** — `BatchAddResources` SHALL accept `ctx` as the first parameter.
- **`library-library-orphan-discovery`** — `DiscoverOrphans` SHALL accept `ctx` as the first parameter (already does, but the errgroup path added in `fix-library-io-discipline` must use it).

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `internal/library/refresher.go:60` | Add `ctx` param | +3 / -2 |
| `internal/library/remover.go:64,129` | Add `ctx` param | +6 / -4 |
| `internal/library/creator.go:71` | Add `ctx` param | +3 / -2 |
| `internal/parser/loader.go:69` | Add `ctx` param | +5 / -2 |
| `internal/renderer/serializer.go:59` | Add `ctx` param | +5 / -2 |
| `internal/transform/transform.go` | Forward `ctx` to parser/renderer | +4 |
| `internal/install/install.go` | Forward `ctx` to parser/renderer | +4 |
| `internal/validate/validate.go` | Forward `ctx` to parser/renderer | +4 |
| `internal/canonicalize/canonicalize.go` | Forward `ctx` to parser/renderer | +4 |
| `cmd/library_add.go:253` | `f.RootContext` → `opts.Ctx` | -1 / +1 |
| `cmd/library_add.go` and 4 sibling files | Use `library.NewLazyLoader` | -10 / +3 |
| `cmd/library_add.go:64-120` (extract-io-adapters Stage 2) | Call `lib.Add(ctx, req)` | (covered by extract-io-adapters) |

### Affected systems

- **Cancellation latency:** the cmd layer's `ctx` is now respected by parser and renderer. A user pressing Ctrl-C during `germinator validate` or `germinator adapt` sees the operation terminate within one file-read time, not at the cmd-side post-call return.
- **Test surface:** every test that calls `AddResource`, `RemoveResource`, `RefreshLibrary`, `CreateLibrary`, `LoadDocument`, `RenderDocument` must be updated to pass a `ctx`. The test pattern becomes `ctx := context.Background()` (or `t.Context()` for new tests) as the first argument.
- **Public API:** all the affected functions are in `internal/` packages, so signature changes are acceptable. The `cli-framework` spec is updated to formalize the contract.
- **Lint baseline:** expected unchanged (no production-code patterns that affect `golangci-lint`).

## Risks

- **Signature break is widespread** — 4 cmd-layer adapters + 4 shell packages + 4 library package-level functions + parser + renderer. The migration is mechanical but spans 14+ files. **Mitigation**: tasks are ordered so cmd-side changes (the small ones) ship first, then shell packages, then library package-level functions. Each commit is independently testable; `mise run build` catches missed call sites.
- **Test fixture churn** — every test that calls an updated function needs a `ctx` argument. The migration touches ~50+ test files. **Mitigation**: design Decision 2 evaluates the `t.Context()` pattern (Go 1.24+) for new tests; existing tests use `context.Background()` for the minimum churn.
- **`opts.Ctx` vs `f.RootContext`** — the `cmd/library_add.go:253` fix changes the context source. The `f.RootContext` is the signal-aware root context (cancelled on SIGINT/SIGTERM); `opts.Ctx` is the per-command context. For most commands, `opts.Ctx` is `f.RootContext` (set in `RunE`); for commands that have per-call deadlines (e.g., completion), `opts.Ctx` may be derived. **Mitigation**: the change is a one-line swap; verify with `mise run test` that no behavior change occurs.
- **errgroup context in `DiscoverOrphans`** — the `errgroup.WithContext(ctx)` from `fix-library-io-discipline` derives a child context that is cancelled when any goroutine returns an error OR when the parent `ctx` is cancelled. The `DiscoverOrphans` function must propagate the child `ctx` to its callers, not the parent. **Mitigation**: design Decision 3 evaluates the errgroup return value; the function returns `ctx.Err()` if the errgroup fails.

## Goals / Non-Goals (folded into design)

**Goals:**

- Retire all 4 `TODO(slice-7)` markers in `internal/library/{creator,refresher,remover}.go`.
- Thread `ctx` through the 4 cmd-layer adapters and the 4 new shell packages from `extract-io-adapters`.
- Replace `f.RootContext` with `opts.Ctx` in `cmd/library_add.go:253`.
- Promote the 5 per-command lazy-loader helpers to `library.NewLazyLoader`.

**Non-Goals:**

- Refactoring parser/renderer internals beyond adding the `ctx` parameter (the ctx check is between file reads, not at every I/O call).
- Changing the `Factory.RootContext` semantics (it remains the signal-aware root context; `opts.Ctx` is the per-call derivative).
- Removing the `*Library` method set introduced by slice 7 (the methods are preserved; the new shell packages from `extract-io-adapters` are the canonical consumers).
