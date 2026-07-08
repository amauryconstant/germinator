# Tasks — Extract I/O adapters to `internal/<x>/` shell packages

The change ships in **4 sequential stages**. Each stage ends with `mise run check` passing. Stage 1 first (lowest-risk slice-3 rationale applies), Stage 2 second (retires explicit `libraryAdapter` debt), Stage 3 last (largest volume), Stage 4 doc sweep.

## 1.0 Stage 1 — Extract validator + canonicalizer

- [ ] 1.1 Create `internal/validate/validate.go` (~90 LOC). Define `Service` interface, `Request` and `Result` types, lift `validateDocument` and `unwrapErrors` from `cmd/validate.go:132-202`. Add `NewService()` constructor returning the production wiring.
- [ ] 1.2 Create `internal/validate/AGENTS.md` (~30 LOC). Follow `internal/library/AGENTS.md` template: Files table, Key Surface, skill reference.
- [ ] 1.3 Create `internal/validate/validate_test.go`. Table-driven tests with `t.TempDir()` fixtures for happy-path and error cases.
- [ ] 1.4 Create `internal/canonicalize/canonicalize.go` (~85 LOC). Define `Service` interface, `Request` and `Result` types, lift `canonicalizeDocument`, `validateCanonicalDoc`, `unwrapCanonicalErrors` from `cmd/canonicalize.go:144-209`. Add `NewService()` constructor.
- [ ] 1.5 Create `internal/canonicalize/AGENTS.md`.
- [ ] 1.6 Create `internal/canonicalize/canonicalize_test.go`.
- [ ] 1.7 In `cmd/validate.go`, delete `validateDocument`, `unwrapErrors`, `validatorAdapter` (lines 132-222). Import `internal/validate`. Update `runValidate` to construct via `validate.NewService()`. The cmd-side `Validator` interface (`cmd/validate.go:22-24`) stays.
- [ ] 1.8 In `cmd/canonicalize.go`, delete `canonicalizeDocument`, `validateCanonicalDoc`, `unwrapCanonicalErrors`, `canonicalizerAdapter` (lines 144-229). Import `internal/canonicalize`. Update `runCanonicalize` to construct via `canonicalize.NewService()`. The cmd-side `Canonicalizer` interface stays.
- [ ] 1.9 Move `cmd/canonicalize_golden_test.go` to `internal/canonicalize/canonicalize_golden_test.go`. Update fixture paths and the test entry point to call `canonicalize.NewService().Canonicalize(...)` directly instead of through the cmd constructor. Fixtures are byte-identical.
- [ ] 1.10 Run `rg "validatorAdapter|canonicalizerAdapter" cmd/` — must return zero matches.
- [ ] 1.11 Run `mise run check` — full validation passes.

## 2.0 Stage 2 — Convert library adders to `*library.Library` methods

- [ ] 2.1 In `internal/library/library.go`, add methods on `*Library`: `Add(ctx context.Context, req *AddRequest) error`, `BatchAddResources(ctx context.Context, opts *BatchAddOptions) (*BatchAddResult, error)`, `DiscoverOrphans(ctx context.Context, opts *DiscoverOptions) (*DiscoverResult, error)`.
- [ ] 2.2 In `internal/library/adder.go`, move the bodies of `library.AddResource`, `library.BatchAddResources`, `library.DiscoverOrphans` into the new methods on `*Library`. Existing package-level functions delegate to the methods (slice-7 decision 6 precedent).
- [ ] 2.3 In `internal/library/discovery.go`, similarly convert `DiscoverOrphans` to a method (if not already covered by task 2.1; consolidate if duplicated).
- [ ] 2.4 In `cmd/library_add.go`, delete `resourceAdder` interface, `libraryAdapter`, `defaultAdder`, and the compile-time check at lines 64-120. Replace with `var _ adderLibrary = (*library.Library)(nil)`. Update `runAddExplicit`, `runAddBatchFiles`, `runAddDiscover`, and any other callers to use `lib.Add`, `lib.BatchAddResources`, `lib.DiscoverOrphans` directly.
- [ ] 2.5 Update `cmd/library_add.go` calls: anywhere that previously called `adder.AddResource(ctx, req)` becomes `lib.Add(ctx, req)`; same pattern for `BatchAddResources` and `DiscoverOrphans`.
- [ ] 2.6 Delete the stale docstring at `cmd/library_add.go:60-63` (along with the deleted adapter).
- [ ] 2.7 In `internal/library/adder_test.go` and `internal/library/discovery_test.go`, update test cases to call the methods via a constructed `*library.Library` instead of the package-level functions (or keep the package-level function calls — they're still public and delegate to the methods).
- [ ] 2.8 Run `rg "libraryAdapter" .` — must return zero matches.
- [ ] 2.9 Run `rg "var _ resourceAdder"` — must return zero matches (the interface is renamed to `adderLibrary`).
- [ ] 2.10 Run `mise run check` — full validation passes.

## 3.0 Stage 3 — Extract transformer + initializer

- [ ] 3.1 Create `internal/transform/transform.go` (~60 LOC). Define `Service` interface, `Request` and `Result` types, lift `transformerAdapter` body from `cmd/transformer.go:33-60`. Constructor `NewService(p *parser.Parser, s *renderer.Serializer) Service` takes the parser + renderer as dependencies (constructed once, not on every call).
- [ ] 3.2 Create `internal/transform/AGENTS.md`.
- [ ] 3.3 Create `internal/transform/transform_test.go`. Table-driven tests with `t.TempDir()` fixtures; assert load → render → write pipeline correctness.
- [ ] 3.4 Create `internal/install/install.go` (~126 LOC). Define `Service` interface, lift `initializerAdapter` body from `cmd/initializer.go:38-125`. Constructor `install.NewService(p *parser.Parser, s *renderer.Serializer) Service`. Package name `install` chosen to avoid collision with Go's reserved `init` identifier (per design Decision 1).
- [ ] 3.5 Create `internal/install/AGENTS.md`.
- [ ] 3.6 Create `internal/install/install_test.go`. Per-ref scenarios: load → render → write; dry-run short-circuit; existing-file check (force vs no-force); ctx cancellation; partial-success aggregation.
- [ ] 3.7 In `cmd/adapt.go`, update `runAdapt` (lines 97-122): replace `func() (Transformer, error) { return NewTransformer(), nil }` with `func() (Transformer, error) { return transform.NewService(parser.NewParser(), renderer.NewSerializer()), nil }`. Import `internal/transform`. Delete `cmd/transformer.go` (the entire file).
- [ ] 3.8 In `cmd/init.go`, move the `Initializer` interface declaration and `InitializeRequest` type from `cmd/initializer.go:18-32` into `cmd/init.go`. Update `runInit` (lines 155-213): replace `NewInitializer()` with `install.NewService(parser.NewParser(), renderer.NewSerializer())`. Import `internal/install`. Delete `cmd/initializer.go`.
- [ ] 3.9 Run `rg "transformerAdapter|initializerAdapter|NewTransformer|NewInitializer" cmd/` — must return zero matches.
- [ ] 3.10 Verify `go build ./...` succeeds — confirms no import cycles (per design Decision 4, shell packages depend on `internal/library` but not vice versa).
- [ ] 3.11 Run `mise run check` — full validation passes.

## 4.0 Stage 4 — Document the convention

- [ ] 4.1 In `internal/AGENTS.md`, add 4 new packages to the package list (lines 14-28): `validate/`, `canonicalize/`, `transform/`, `install/`. Update the package dependency diagram to reflect: `cmd/ → {validate, canonicalize, transform, install} → {core, parser, renderer, library}`.
- [ ] 4.2 In `internal/AGENTS.md`, update the `internal/{claude-code,opencode}/` bullet (line 78) to clarify: pure platform-specific validation rules (e.g. OpenCode mode/temperature enums) live in `internal/core/<platform>/` subpackages; I/O-bound transformation logic stays in `internal/<platform>/`.
- [ ] 4.3 In `cmd/AGENTS.md`, update the canonical `adapt` example to construct via `transform.NewService(...)` instead of the deleted `cmd.NewTransformer()`. Update the Foundation Units table to reflect the new import surface (`internal/validate`, `internal/canonicalize`, `internal/transform`, `internal/install`).
- [ ] 4.4 In `cmd/commands/AGENTS.md`, update per-command reference tables that mention `cmd/transformer.go` or `cmd/initializer.go` (now deleted) to point to the new shell packages.
- [ ] 4.5 Apply the spec delta at `openspec/changes/extract-io-adapters/specs/cli-framework/spec.md` via `osc-sync-specs` (this codifies the new "I/O adapter placement" requirement).

## 5.0 Final verification

- [ ] 5.1 Run `rg "transformerAdapter|validatorAdapter|canonicalizerAdapter|initializerAdapter|libraryAdapter" cmd/ internal/` — must return zero matches (all 5 adapters extracted).
- [ ] 5.2 Run `rg "var _ resourceAdder"` — must return zero matches (replaced with `var _ adderLibrary = (*library.Library)(nil)`).
- [ ] 5.3 Run `go build ./...` — confirms no import cycles and no broken refs.
- [ ] 5.4 Run `mise run lint` — if output shifts (e.g., from the AGENTS.md updates or new package conventions), refresh `cmd/testdata/lint_baseline.txt` per the procedure in `cmd/AGENTS.md` "Lint Baseline Test" section.
- [ ] 5.5 Run `mise run test` — confirm all unit tests pass (including the new shell-package tests).
- [ ] 5.6 Run `mise run test:e2e` — confirm all E2E tests pass (especially the `library add` tests that exercise `Add`, `BatchAddResources`, `DiscoverOrphans`).
- [ ] 5.7 Run `mise run test:coverage` — confirm coverage for the 4 new shell packages ≥ 70% (per `config.testing`).
- [ ] 5.8 Run `openspec validate extract-io-adapters --strict` — confirm all specs and tasks are coherent.
- [ ] 5.9 Manually exercise every command end-to-end to verify CLI behavior is unchanged:
  - `germinator adapt <input> <output> --platform claude-code`
  - `germinator adapt <input> <output> --platform opencode`
  - `germinator validate <input> --platform claude-code`
  - `germinator canonicalize <input> <output> --platform claude-code --type agent`
  - `germinator init --platform opencode --resources skill/commit`
  - `germinator library add <file> --type skill --name test`
  - `germinator library add --discover --batch --force`
- [ ] 5.10 Run `git status` — verify only the expected files are modified (no accidental edits).

## 6.0 Archive

- [ ] 6.1 Archive this change via `osc-archive-change extract-io-adapters`.
- [ ] 6.2 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.
