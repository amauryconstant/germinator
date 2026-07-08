# Tasks — Propagate context through shell-package boundaries

Each task ends with `mise run check` passing. Tasks are grouped by phase and ordered so each commit is independently testable.

## 1. Phase 1 — cmd-side small fixes (LazyLoader + opts.Ctx)

- [ ] 1.1 In `internal/library/discovery.go`, add `func NewLazyLoader(f *cmdutil.Factory, explicitPath string) func() (*Library, error)` that:
  - Resolves the path via `FindLibrary(explicitPath, "")` on first call.
  - Loads the library via `LoadLibrary(ctx, path)` on first call.
  - Returns the cached result on subsequent calls.
  - Stores the lazy `*Library` in a `sync.OnceValues` for thread safety.
- [ ] 1.2 In `cmd/library_add.go:120` (the `addLibrary` helper), replace the `FindLibrary + LoadLibrary` body with `library.NewLazyLoader(f, opts.Library)`. Update `runAddExplicit` to call `lib.Add(ctx, req, opts.IO.Out)`.
- [ ] 1.3 In `cmd/library_create.go:140` (the `createPresetLibrary` helper), apply the same `NewLazyLoader` pattern. Update `runCreate` to use the lazy loader.
- [ ] 1.4 In `cmd/library_refresh.go:120` (the `refreshLibrary` helper), apply the same `NewLazyLoader` pattern.
- [ ] 1.5 In `cmd/library_remove.go:170` (the `removeLibrary` helper), apply the same `NewLazyLoader` pattern.
- [ ] 1.6 In `cmd/library_validate.go:120` (the `validateLibrary` helper), apply the same `NewLazyLoader` pattern.
- [ ] 1.7 In `cmd/library_init.go:120` (the `initLibrary` helper), apply the same `NewLazyLoader` pattern (this one doesn't take a path; uses the resolved path).
- [ ] 1.8 In `cmd/library_add.go:253`, replace `library.LoadLibrary(f.RootContext, resolved)` with `library.LoadLibrary(opts.Ctx, resolved)`. The `opts.Ctx` is captured in the `addLibrary` closure.
- [ ] 1.9 Run `rg "f\.RootContext" cmd/` — must return zero matches (all call sites use `opts.Ctx` or a per-call derived context).
- [ ] 1.10 Run `rg "FindLibrary.*LoadLibrary" cmd/` — must return zero matches (all helpers use `NewLazyLoader`).

## 2. Phase 2 — parser + renderer ctx

- [ ] 2.1 In `internal/parser/loader.go`, change `func LoadDocument(inputPath, platform string) (*Document, error)` to `func LoadDocument(ctx context.Context, inputPath, platform string) (*Document, error)`. Add `ctx.Err()` checks between file reads (or wrap the file open in a goroutine that respects `ctx.Done()`).
- [ ] 2.2 In `internal/parser/loader.go`, change `func DetectType(filename string) (string, error)` to `func DetectType(ctx context.Context, filename string) (string, error)`. (Lightweight; the regex loop checks `ctx.Err()` between iterations.)
- [ ] 2.3 In `internal/renderer/serializer.go`, change `func RenderDocument(doc *Document, platform string) ([]byte, error)` to `func RenderDocument(ctx context.Context, doc *Document, platform string) ([]byte, error)`. (The template cache lookup is fast; the function checks `ctx.Err()` once at entry.)
- [ ] 2.4 In `internal/transform/transform.go`, change `transformerAdapter.Transform(ctx, *TransformRequest)` to forward `ctx` to `parser.LoadDocument(ctx, ...)` and `renderer.RenderDocument(ctx, ...)`.
- [ ] 2.5 In `internal/install/install.go`, change `initializerAdapter.Initialize(ctx, *InitializeRequest)` to forward `ctx` to `parser.LoadDocument(ctx, ...)` and `renderer.RenderDocument(ctx, ...)` at every per-ref call.
- [ ] 2.6 In `internal/validate/validate.go`, change `validatorAdapter.Validate(ctx, *ValidateRequest)` to forward `ctx` to `parser.ParsePlatformDocument(ctx, ...)` and any I/O calls.
- [ ] 2.7 In `internal/canonicalize/canonicalize.go`, change `canonicalizerAdapter.Canonicalize(ctx, *CanonicalizeRequest)` to forward `ctx` to `parser.ParsePlatformDocument(ctx, ...)` and `renderer.MarshalCanonical(ctx, ...)`.
- [ ] 2.8 In `cmd/initializer.go:53`, `cmd/transformer.go:44`, `cmd/canonicalize.go:147`, `cmd/validate.go:134` (or the new shell-package files): change the parameter from `_ context.Context` to `ctx context.Context` (no-op if the file is being deleted by `extract-io-adapters`; verified by `rg "_ context\.Context" cmd/ internal/`).
- [ ] 2.9 Run `rg "_ context\.Context" cmd/ internal/` — must return zero matches.
- [ ] 2.10 Run `rg "\.LoadDocument\(|\.RenderDocument\(|\.ParsePlatformDocument\(|\.MarshalCanonical\(" internal/ cmd/` — verify every call site passes a `ctx` as the first argument.

## 3. Phase 3 — library package-level functions

- [ ] 3.1 In `internal/library/refresher.go:60`, change `func RefreshLibrary(opts RefreshOptions) error` to `func RefreshLibrary(ctx context.Context, opts RefreshOptions) error`. Remove the `context.Background()` line and the `// TODO(slice-7)` marker. Forward `ctx` to `LoadLibrary` and `SaveLibrary` calls.
- [ ] 3.2 In `internal/library/remover.go:64`, change `func RemoveResource(opts RemoveResourceOptions) error` to `func RemoveResource(ctx context.Context, opts RemoveResourceOptions) error`. Remove the `context.Background()` line and the `// TODO(slice-7)` marker.
- [ ] 3.3 In `internal/library/remover.go:129`, change `func RemovePreset(opts RemovePresetOptions) error` to `func RemovePreset(ctx context.Context, opts RemovePresetOptions) error`. Remove the `context.Background()` line and the `// TODO(slice-7)` marker.
- [ ] 3.4 In `internal/library/creator.go:71`, change `func CreateLibrary(opts CreateOptions) error` to `func CreateLibrary(ctx context.Context, opts CreateOptions, stdout io.Writer) error` (combining this change's `ctx` and `fix-library-io-discipline`'s `io.Writer`). Remove the `context.Background()` line and the `// TODO(slice-7)` marker.
- [ ] 3.5 Run `rg "TODO\(slice-7\)" internal/library/` — must return zero matches.
- [ ] 3.6 Run `rg "context\.Background" internal/library/` — must return zero matches in production code.
- [ ] 3.7 Update all test files (`internal/library/*_test.go`) that call the updated functions to pass `context.Background()` as the first argument. Test files may use `t.Context()` for new tests.

## 4. Verification

- [ ] 4.1 Run `mise run build` — no broken imports.
- [ ] 4.2 Run `mise run lint` — must report 0 issues.
- [ ] 4.3 Run `mise run test` — all unit tests pass.
- [ ] 4.4 Run `mise run test:e2e` — E2E tests pass.
- [ ] 4.5 Run `mise run test:race` — no race conditions; cancellation tests pass.
- [ ] 4.6 Run `rg "TODO\(slice-7\)|context\.Background\(\)" internal/library/` — must return zero matches.
- [ ] 4.7 Run `rg "_ context\.Context" cmd/ internal/` — must return zero matches.
- [ ] 4.8 Run `rg "f\.RootContext" cmd/` — must return zero matches.
- [ ] 4.9 Run `openspec validate propagate-context-through-shell --strict` — change is coherent.

## 5. Archive

- [ ] 5.1 Apply spec deltas via `osc-sync-specs`.
- [ ] 5.2 Archive this change via `osc-archive-change propagate-context-through-shell`.
- [ ] 5.3 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.
