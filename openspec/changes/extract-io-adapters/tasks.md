# Tasks — Extract I/O adapters to `internal/<x>/` shell packages

The change ships in **3 sequential stages**. Each stage ends with `mise run check` passing. Stage 1 first (lowest-risk slice-3 rationale applies), Stage 2 second (retires explicit `libraryAdapter` debt), Stage 3 last (largest volume). Documentation (AGENTS.md updates, spec sync) and archive are handled separately.

## 1.0 Stage 1 — Extract validator + canonicalizer

- [ ] 1.1 Create `internal/validate/validate.go` (~90 LOC). Define `Service` interface, `Request` and `Result` types, lift `validateDocument` (L130-183) and `unwrapErrors` (L184-197) from `cmd/validate.go`. Add `NewService()` constructor returning the production wiring.
- [ ] 1.2 Create `internal/validate/AGENTS.md` (~30 LOC). Follow `internal/library/AGENTS.md` template: Files table, Key Surface, skill reference.
- [ ] 1.3 Create `internal/validate/validate_test.go`. Table-driven tests with `t.TempDir()` fixtures for happy-path and error cases.
- [ ] 1.4 Create `internal/canonicalize/canonicalize.go` (~85 LOC). Define `Service` interface, `Request` and `Result` types, lift `canonicalizeDocument` (L158-177), `validateCanonicalDoc` (L182-207), `unwrapCanonicalErrors` (L209-220) from `cmd/canonicalize.go`. Add `NewService()` constructor.
- [ ] 1.5 Create `internal/canonicalize/AGENTS.md`.
- [ ] 1.6 Create `internal/canonicalize/canonicalize_test.go`.
- [ ] 1.7 In `cmd/validate.go`, delete `validateDocument`, `unwrapErrors`, `validatorAdapter` (lines 130-218). Import `internal/validate`. Update `runValidate` to construct via `validate.NewService()`. The cmd-side `Validator` interface (`cmd/validate.go:21`) stays.
- [ ] 1.8 In `cmd/canonicalize.go`, delete `canonicalizeDocument` (L158-177), `validateCanonicalDoc` (L182-207), `unwrapCanonicalErrors` (L209-220), `canonicalizerAdapter` (L232-240). Import `internal/canonicalize`. Update `runCanonicalize` to construct via `canonicalize.NewService()`. The cmd-side `Canonicalizer` interface stays.
- [ ] 1.9 Move `cmd/canonicalize_golden_test.go` to `internal/canonicalize/canonicalize_golden_test.go`. Update fixture paths and the test entry point to call `canonicalize.NewService().Canonicalize(...)` directly instead of through the cmd constructor. Fixtures are byte-identical.
- [ ] 1.10 Run `rg "validatorAdapter|canonicalizerAdapter" cmd/` — must return zero matches.
- [ ] 1.11 Run `mise run check` — full validation passes.

## 2.0 Stage 2 — Convert library adders to `*library.Library` methods

- [ ] 2.1 In `internal/library/library.go`, add methods on `*Library`: `Add(ctx context.Context, req *AddRequest) error`, `BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error)`, `DiscoverOrphans(ctx context.Context, opts *DiscoverOptions) (*DiscoverResult, error)`.
- [ ] 2.2 In `internal/library/adder.go`, move the bodies of `library.AddResource`, `library.BatchAddResources`, `library.DiscoverOrphans` into the new methods on `*Library`. Existing package-level functions delegate to the methods (slice-7 decision 6 precedent).
- [ ] 2.3 In `internal/library/discovery.go`, similarly convert `DiscoverOrphans` to a method (if not already covered by task 2.1; consolidate if duplicated).
- [ ] 2.4 In `cmd/library_add.go`, delete `resourceAdder` interface (L70), `libraryAdapter` (L83-120), `defaultAdder` (L118), and the compile-time check (L113). Replace with `var _ adderLibrary = (*library.Library)(nil)`. Update `runAddExplicit` (L308), `runAddBatchFiles` (L497), `runAddDiscover` (L583), and any other callers to use `lib.Add`, `lib.BatchAddResources`, `lib.DiscoverOrphans` directly.
- [ ] 2.5 Update `cmd/library_add.go` calls: anywhere that previously called `adder.AddResource(ctx, req)` becomes `lib.Add(ctx, req)`; same pattern for `BatchAddResources` and `DiscoverOrphans`.
- [ ] 2.6 Delete the stale docstring at `cmd/library_add.go:60-63` (along with the deleted adapter).
- [ ] 2.7 In `cmd/library_add_test.go`, update test code that references the deleted `libraryAdapter` / `resourceAdder` types:
  - [ ] 2.7a Delete `TestResourceAdderInterfaceSatisfied` (T12, L566-575) — the `var _ adderLibrary = (*library.Library)(nil)` compile-time check at task 2.4 already proves the contract.
  - [ ] 2.7b Delete the `captureAdder` type (L539-564) — replaced by real `*library.Library` instances per project convention (`internal/AGENTS.md` "When to Mock vs Use Real Implementations").
  - [ ] 2.7c Rewrite `TestRunAdd_PopulatesStdoutOnAddRequest` (T18b, L444-489) and `TestRunAddBatchFiles_PopulatesStdoutOnBatchAddOptions` (T18c, L494-532) to:
        - Build a real `*library.Library` via `library.LoadLibrary(ctx, libDir)` (mirrors `makeTestLibrary`).
        - Inject a captured `*library.AddRequest` / `library.BatchAddOptions` via a thin test-local stub that records the call and short-circuits the real body (avoids mutating library.yaml).
        - Keep the existing `Stdout`-assertion checks identical (the assertions on `captured.Stdout != nil` and `captured.Stdout == ios.Out` are the test's purpose).
  - [ ] 2.7d Update doc comments at L426-443, L491-493, L534-538 to reference `(*library.Library)` instead of `resourceAdder` / `defaultAdder`.
- [ ] 2.8 Sweep stale doc comments referencing the soon-to-be-deleted `resourceAdder` / `libraryAdapter` / `defaultAdder` in `cmd/library_create.go:37,41`, `cmd/library_remove.go:53`, `cmd/library_refresh.go:37`, and `cmd/library_add.go:24-69`. Rewrite to describe the post-extraction state: `cmd/library_add.go` directly uses `*library.Library` methods (`lib.Add`, `lib.BatchAddResources`, `lib.DiscoverOrphans`); no adapter shim.
- [ ] 2.9 Run `rg "libraryAdapter" .` — must return zero matches.
- [ ] 2.10 Run `rg "var _ resourceAdder"` — must return zero matches (the interface is renamed to `adderLibrary`).
- [ ] 2.11 Run `mise run check` — full validation passes.

## 3.0 Stage 3 — Extract transformer + initializer

- [ ] 3.1 Create `internal/transform/transform.go` (~60 LOC). Define `Service` interface, `Request` and `Result` types, lift `transformerAdapter` body from `cmd/transformer.go:33-60`. Constructor `NewService(p *parser.Parser, s *renderer.Serializer) Service` takes the parser + renderer as dependencies (constructed once, not on every call).
- [ ] 3.2 Create `internal/transform/AGENTS.md`.
- [ ] 3.3 Create `internal/transform/transform_test.go`. Table-driven tests with `t.TempDir()` fixtures; assert load → render → write pipeline correctness.
- [ ] 3.4 Create `internal/install/install.go` (~126 LOC). Define `Service` interface, lift `initializerAdapter` body from `cmd/initializer.go:38-125`. Constructor `install.NewService(p *parser.Parser, s *renderer.Serializer) Service`. Package name `install` chosen to avoid collision with Go's reserved `init` identifier (per design Decision 1).
- [ ] 3.5 Create `internal/install/AGENTS.md`.
- [ ] 3.6 Create `internal/install/install_test.go`. Per-ref scenarios: load → render → write; dry-run short-circuit; existing-file check (force vs no-force); ctx cancellation; partial-success aggregation.
- [ ] 3.7 In `cmd/adapt.go`, update `runAdapt` (lines 97-122): replace `func() (Transformer, error) { return NewTransformer(), nil }` with `func() (Transformer, error) { return transform.NewService(parser.NewParser(), renderer.NewSerializer()), nil }`. Import `internal/transform`. Delete `cmd/transformer.go` (the entire file).
- [ ] 3.8 In `cmd/init.go`, declare the `Initializer` interface (currently at `cmd/initializer.go:18-20`); place it adjacent to the existing `InitializeRequest` struct at `cmd/init.go:24-31`. Update `runInit` (L152-209) to call `install.NewService(parser.NewParser(), renderer.NewSerializer())` instead of `NewInitializer()`. Delete `cmd/initializer.go` (entire file). The doc comment at `cmd/init.go:149` and `cmd/init_test.go` references are updated by task 3.9.
- [ ] 3.9 Sweep stale doc comments referencing the soon-to-be-deleted `NewTransformer`, `NewInitializer`, `transformerAdapter`, `initializerAdapter` in:
  - `main.go:39`
  - `cmd/adapt.go:36,93` (doc comments; L106 is the actual call site, updated by task 3.7)
  - `cmd/adapt_test.go:118,241,271`
  - `cmd/init.go:149` (doc comment in runInit body; L185 call site is updated by task 3.8)
  - `cmd/init_test.go:22-34,607`
  - `cmd/cmd_test.go:19,50,314`
  Rewrite to describe the post-extraction state: `transform.NewService(...)` for adapt, `install.NewService(...)` for init; the cmd-side `Transformer` and `Initializer` interfaces remain as compile-time contract assertions.
- [ ] 3.10 Run `rg "transformerAdapter|initializerAdapter|NewTransformer|NewInitializer" cmd/` — must return zero matches.
- [ ] 3.11 Verify `go build ./...` succeeds — confirms no import cycles (per design Decision 4, shell packages depend on `internal/library` but not vice versa).
- [ ] 3.12 Run `mise run check` — full validation passes.

## 4.0 Final verification

- [ ] 4.1 Run `rg "transformerAdapter|validatorAdapter|canonicalizerAdapter|initializerAdapter|libraryAdapter" cmd/ internal/` — must return zero matches (all 5 adapters extracted).
- [ ] 4.2 Run `rg "var _ resourceAdder"` — must return zero matches (replaced with `var _ adderLibrary = (*library.Library)(nil)`).
- [ ] 4.3 Run `go build ./...` — confirms no import cycles and no broken refs.
- [ ] 4.4 Run `mise run lint` — if output shifts (e.g., from new shell packages following existing conventions), refresh `cmd/testdata/lint_baseline.txt` per the procedure in `cmd/AGENTS.md` "Lint Baseline Test" section.
- [ ] 4.5 Run `mise run test` — confirm all unit tests pass (including the new shell-package tests).
- [ ] 4.6 Run `mise run test:e2e` — confirm all E2E tests pass (especially the `library add` tests that exercise `Add`, `BatchAddResources`, `DiscoverOrphans`).
- [ ] 4.7 Run `mise run test:coverage` — confirm coverage for the 4 new shell packages ≥ 70% (per the coverage threshold in `mise.toml`). The new packages are additive; existing cmd/ and internal/library/ coverage is not expected to change.
- [ ] 4.8 Run `openspec validate extract-io-adapters --strict` — confirm all specs and tasks are coherent.
- [ ] 4.9 Manually exercise every command end-to-end to verify CLI behavior is unchanged:
  - `germinator adapt <input> <output> --platform claude-code`
  - `germinator adapt <input> <output> --platform opencode`
  - `germinator validate <input> --platform claude-code`
  - `germinator canonicalize <input> <output> --platform claude-code --type agent`
  - `germinator init --platform opencode --resources skill/commit`
  - `germinator library add <file> --type skill --name test`
  - `germinator library add --discover --batch --force`
