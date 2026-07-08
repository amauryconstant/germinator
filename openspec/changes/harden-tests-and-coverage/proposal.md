## Why

The 2026-07-08 code review identified **17 test-infra findings** (D-001, D-003, D-004, D-007, D-008 / D-024, D-010..D-014, D-015..D-026, D-029, D-030) that span the testing pyramid. The findings cluster in 4 areas:

1. **Blocker race condition** (D-001) — `cmd/lint_test.go:19` uses `t.Parallel()` while calling `cmd.Execute()` 8 times. Combined with `t.Parallel()` in dozens of cmd tests, the race detector flags concurrent reads/writes on Cobra's package-level `OnInitialize` slice.
2. **Test coverage gaps** (D-010, D-011, D-012, D-018, D-019, D-020, D-021, D-026) — `internal/library` is 79.4% (target 80%); six functions at 0% coverage; `internal/renderer` is 79%; `claude-code` adapter `parseTargets` 0% / `mapPermissionModeToPolicy` 28.6%.
3. **Testify migration** (D-004) — `internal/library/methods_test.go` uses raw `t.Fatalf` / `t.Errorf` / `t.Fatal` instead of testify's `require` / `assert`. Inconsistent with `cmd/` tests.
4. **Adapter contract** (D-003, D-016, D-017, D-029) — `permission.Adapter` is declared in `internal/permission/adapter.go` but no adapter satisfies it via compile-time check. Adapter `New()` returns stateless `&Adapter{}` per call instead of a package-level singleton.

The review's hotfix recommendation is to land D-001 immediately. The remaining 16 findings are bundled into this change because they share the "test infrastructure" theme and benefit from a coherent testing-pyramid pass.

This change is a **production-test refactor** with spec deltas. The hotfix D-001 lands first as a one-line edit (no spec change needed); the remainder ships as a single mega-change.

## What Changes

### Phase 0 — Hotfix D-001 (lands immediately, outside OpenSpec)

- **MODIFY** `cmd/lint_test.go:19` — remove `t.Parallel()` from `TestLintBaseline` and any other test that calls `cmd.Execute()`.

### Phase 1 — Adapter contract + singleton + typed constants

- **MODIFY** `internal/opencode/opencode_adapter.go` — add `var _ permission.Adapter = (*Adapter)(nil)` compile-time check.
- **MODIFY** `internal/claude-code/claude_code_adapter.go` — add `var _ permission.Adapter = (*Adapter)(nil)` compile-time check.
- **MODIFY** `internal/opencode/opencode.go` — replace `func New() *Adapter { return &Adapter{} }` with `var OpenCode = &Adapter{}` package-level singleton.
- **MODIFY** `internal/claude-code/claude_code.go` — replace `func New() *Adapter { return &Adapter{} }` with `var ClaudeCode = &Adapter{}` package-level singleton.
- **MODIFY** `internal/permission/permissions.go` and adapter maps — replace raw string literals with typed `permission.Allow` / `permission.Deny` constants.
- **MODIFY** `internal/renderer/serializer.go:216` — change `func getDocType(doc interface{})` to `func getDocType(doc any)` and add explicit nil-check.
- **MODIFY** `internal/warning/canary.go:51` — wrap `canaryOnce` in a `sync.Mutex` to make `ResetCanaryForTest` race-safe.
- **MODIFY** `internal/library/adder.go:948` — always set `c.Cause` in `checkNameConflict` (even when synthesized from `c.Issue`); drop the dual-path `c.Cause vs errors.New(c.Issue)` branch in `collectDiscoverFailures`.

### Phase 2 — testify migration

- **MODIFY** `internal/library/methods_test.go` — migrate raw `t.Fatalf` / `t.Errorf` / `t.Fatal` to `require.NoError` / `require.Error` / `assert.Equal`.
- **MODIFY** `internal/library/library_test.go`, `internal/library/refresher_test.go`, `internal/library/remover_test.go`, `internal/library/loader_test.go`, `internal/library/saver_test.go`, `internal/library/resolver_test.go` — same migration.

### Phase 3 — Coverage gap fixes

- **NEW** `internal/library/creator_test.go` — table-driven tests for `CreateLibrary` (currently 43.5%) and `defaultLibraryYAML` (0%).
- **NEW** `internal/library/discovery_test.go` — table-driven tests for `DefaultLibraryPath` (18.2%) and `FindLibrary` priority resolution.
- **NEW** `internal/library/methods_test.go` — table-driven tests for `CreatePreset` (both forms at 0%): `(*Library).CreatePreset` and the package-level `CreatePreset`.
- **NEW** `internal/library/resolver_test.go` — table-driven test for `GetOutputPaths` (0%).
- **MODIFY** `internal/claude-code/claude_code_adapter_test.go` — add coverage for `parseAgent` (51.8%), `parseTargets` (0%), `mapPermissionModeToPolicy` (28.6%), `renderAgent` (65.9%).
- **MODIFY** `internal/core/results.go` and `errors.go` — add tests for `core.Valid` (0%), `Unwrap` chains (0%).

### Phase 4 — Golden files + round-trip

- **NEW** `internal/opencode/opencode_adapter_golden_test.go` (build tag `golden`) — fixture-parse → adapter-render → fixture-golden-compare for at least one agent with `permissionPolicy=balanced`.
- **NEW** `internal/claude-code/claude_code_adapter_golden_test.go` (build tag `golden`) — same pattern for Claude Code.
- **MODIFY** `internal/renderer/serializer_test.go` — add `TestParseRenderRoundTrip` that reads a fixture, parses it through `parser.ParsePlatformDocument`, re-emits via `MarshalCanonical`, and asserts byte-equivalence of the canonical form.
- **MODIFY** `cmd/canonicalize_golden_test.go:26` — compute fixture path via `runtime.Caller(0)` instead of the CWD-relative `../test/fixtures/canonical`.

### Phase 5 — Test parallelization

- **MODIFY** `internal/library/methods_test.go` — add `t.Parallel()` to fixture-driven subtests that don't share `t.Setenv` or `os.Chdir` (e.g., `TestLibrary_X_CtxCancelled`).
- **MODIFY** `cmd/library_add_test.go` — add `t.Parallel()` to fixture-driven subtests at line 383+.

### Phase 6 — Trivial folds (per the reorg plan)

- **FOLD** A-008 — widen the forbidigo pattern at `.golangci.yml:87` to catch all package-level mutable vars (`var defaultAdder`, `var outputFormat`, `var initOutputFormat`, `var outputFormatRefresh`, `var completionShells`, `var errEmptyResources`). Replace with per-options fields or per-command constants.
- **FOLD** A-015 — replace the hard-coded file list in `TestNoNewForbidigoPatterns` (`cmd/lint_test.go:96`) with `go list ./cmd | grep -v _test.go`.
- **FOLD** D-015 — `interface{}` → `any` (already in Phase 1).
- **FOLD** D-021 — `core.Valid` / `Unwrap` coverage (already in Phase 3).
- **FOLD** D-022 — `t.Parallel()` migration (already in Phase 5).
- **FOLD** D-025 — `cmd/canonicalize_golden_test.go` CWD-skip (already in Phase 4).
- **FOLD** D-030 — `t.Parallel()` in `cmd/library_add_test.go` (already in Phase 5).

## Capabilities

### New Capabilities

- **`testing-round-trip-tests`** — define the parse→render round-trip contract for adapter golden tests; codify the test pattern as a project convention.

### Modified Capabilities

- **`testing-e2e-testing`** — add `t.Parallel()` patterns for fixture-driven tests; document the race-condition hazard from `t.Parallel()` + `cmd.Execute()` interactions.
- **`testing-iostreams-injection`** — add the parse→render round-trip pattern as a standard test idiom.
- **`transformation-platform-adapters`** — add the compile-time adapter contract requirement: `var _ permission.Adapter = (*Adapter)(nil)` SHALL be present in both adapter files; `New()` SHALL return a singleton (not a per-call `&Adapter{}`).
- **`transformation-permission-transformation`** — add the typed `permission.Allow` / `permission.Deny` constant requirement; raw string literals SHALL NOT appear in adapter permission maps.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `cmd/lint_test.go:19` | Hotfix `t.Parallel()` | -1 (hotfix) |
| `cmd/lint_test.go:96` | Dynamic file list | -5 / +3 |
| `internal/opencode/opencode.go` | Singleton + contract check | -3 / +3 |
| `internal/claude-code/claude_code.go` | Singleton + contract check | -3 / +3 |
| `internal/permission/permissions.go` | Typed constants | -10 / +10 |
| `internal/renderer/serializer.go:216` | `any` + nil-check | -1 / +2 |
| `internal/warning/canary.go:51` | Mutex wrap | -1 / +3 |
| `internal/library/adder.go:948` | `c.Cause` always set | -2 / +2 |
| `internal/library/methods_test.go` | testify migration | -20 / +30 |
| `internal/library/*_test.go` (6 files) | testify migration | -30 / +45 |
| `internal/library/creator_test.go` (new) | New tests | +200 |
| `internal/library/discovery_test.go` (new) | New tests | +80 |
| `internal/library/resolver_test.go` | GetOutputPaths test | +30 |
| `internal/claude-code/claude_code_adapter_test.go` | Coverage | +150 |
| `internal/core/results_test.go` | Valid/Unwrap tests | +30 |
| `internal/opencode/opencode_adapter_golden_test.go` (new) | Golden test | +100 |
| `internal/claude-code/claude_code_adapter_golden_test.go` (new) | Golden test | +100 |
| `internal/renderer/serializer_test.go` | Round-trip test | +60 |
| `cmd/canonicalize_golden_test.go:26` | `runtime.Caller(0)` fixture | -3 / +5 |
| `internal/library/methods_test.go` | `t.Parallel()` | +5 |
| `cmd/library_add_test.go` | `t.Parallel()` | +10 |
| `.golangci.yml:87` | Widen forbidigo pattern | +1 / -1 |
| `cmd/library_add.go:120` (and 5 sibling files) | Remove package-level vars | -5 / +5 |
| **Total** | | **~+700 LOC tests, ~-30 LOC prod** |

### Affected systems

- **Coverage:** all packages reach ≥ 70% per `config.testing`. `internal/library` reaches 85%+. `internal/renderer` reaches 80%+. `claude-code` adapter reaches 80%+.
- **Test stability:** race detector (`go test -race -count=1 ./...`) passes after the hotfix and the `t.Parallel()` migrations.
- **CI:** the `mise run test:race` task becomes a hard requirement (already a soft requirement per `config.testing`).
- **Public API:** no production-code API changes (only test internals and adapter internals change).

## Risks

- **Hotfix D-001 is a behavioral change for the race detector.** Removing `t.Parallel()` from `cmd/lint_test.go:19` adds 2-3 seconds to the test runtime (no longer parallel with sibling tests). **Mitigation**: the slowness is acceptable for race-detector correctness; `mise run test` (without `-race`) is unaffected.
- **testify migration is mechanical but spans 7 test files.** A missed `t.Fatal` is hard to detect. **Mitigation**: task 4.1 runs `rg "t\.Fatal|t\.Error" internal/library/*_test.go` after the migration; zero matches expected outside of test-only assertions.
- **New test files** (creator_test.go, discovery_test.go, golden files) add ~700 LOC. Coverage increase is real but the tests must actually exercise the code, not just call constructors. **Mitigation**: design Decision 4 requires table-driven tests with concrete inputs and expected outputs (matching the `library-library-batch-add` precedent).
- **Adapter singleton** (Phase 1) is a behavior change for the `New()` callers. **Mitigation**: the new `OpenCode` / `ClaudeCode` package-level vars are drop-in replacements; all `claudecode.New()` / `opencode.New()` call sites are updated in the same commit.

## Goals / Non-Goals (folded into design)

**Goals:**

- Land D-001 hotfix immediately (separate from this change).
- Add compile-time adapter contract checks.
- Replace per-call adapter instantiation with package-level singletons.
- Use typed `permission.Allow` / `permission.Deny` constants.
- Migrate library tests to testify.
- Add coverage for 6 functions at 0%.
- Add golden + round-trip tests for adapters.
- Add `t.Parallel()` to fixture-driven tests.
- Widen the forbidigo pattern to catch all package-level mutable vars.

**Non-Goals:**

- Performance optimizations (C-010..C-018, C-024, C-025) — deferred to a separate `perf-hardening` change.
- Changing the `Library` coverage target from 70% to 80% (the 80% target is aspirational; 70% is the testing rule).
- Adding new test frameworks (the project uses testify + Ginkgo v2; the change stays within those).
