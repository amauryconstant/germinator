## Why

The 2026-07-08 code review identified 17 test-infra findings across the testing pyramid, clustered in four areas: a Cobra race in `cmd/lint_test.go` (D-001), coverage gaps in `internal/library` and the platform adapters (D-010..D-026), inconsistent use of testify vs raw `t.Fatal` calls (D-004), and an unenforced `permission.Adapter` contract with `New()` per-call allocation instead of a singleton (D-003, D-016, D-017, D-029). The hotfix D-001 lands separately, before this change; the remaining 16 findings ship together as one mega-change because they share a coherent "test infrastructure" theme (mid-state review, dependency tree, and race-detector posture). Full finding inventory (D-001..D-030) is captured in the design's Context section.

This change is a production-test refactor with spec deltas; the D-001 hotfix lands ahead of it via a separate commit.

## What Changes

### Phase 0 — Hotfix D-001 (lands immediately, outside OpenSpec)

- **MODIFY** `cmd/lint_test.go:19` — remove `t.Parallel()` from `TestLintBaseline` and any other test that calls `cmd.Execute()`.

### Phase 1 — Adapter contract + singleton + typed constants

- **MODIFY** `internal/opencode/opencode_adapter.go` — add `var _ permission.Adapter = (*Adapter)(nil)` compile-time check.
- **MODIFY** `internal/claude-code/claude_code_adapter.go` — add `var _ permission.Adapter = (*Adapter)(nil)` compile-time check.
- **MODIFY** `internal/opencode/opencode_adapter.go` — replace `func New() *Adapter { return &Adapter{} }` with `var OpenCode = &Adapter{}` package-level singleton.
- **MODIFY** `internal/claude-code/claude_code_adapter.go` — replace `func New() *Adapter { return &Adapter{} }` with `var ClaudeCode = &Adapter{}` package-level singleton.
- **MODIFY** `internal/permission/permissions.go` and adapter maps — replace raw string literals with typed `permission.Allow` / `permission.Deny` constants.
- **MODIFY** `internal/renderer/serializer.go` — replace ALL `interface{}` with `any` in signatures at lines 22, 23, 27, 34, 54, 90, 224, 243, 294, 335. Add explicit nil-check at top of `getDocType` (line 224) returning `*core.TransformError` for `doc == nil`.
- **MODIFY** `internal/opencode/opencode_adapter.go::mapPermissionObjectToPolicy` (lines 486-509) — extract raw `actionStr` values into typed `permission.Action` constants before comparison. The function operates on raw YAML `map[string]interface{}` (not the typed `permission.Map`), so the comparison must convert via `permission.Action(actionStr)` before equality checks.
- **MODIFY** `cmd/library_add.go:632-637` — delete the dual-path `c.Cause vs errors.New(c.Issue)` branch in `collectDiscoverFailures`. The producer-side `adder.go:835-845` already sets `Cause` correctly via `conflictErr`, so no `adder.go` change is required.

### Phase 2 — testify migration

- **MODIFY** `internal/library/methods_test.go` — migrate raw `t.Fatalf` / `t.Errorf` / `t.Fatal` to `require.NoError` / `require.Error` / `assert.Equal`.
- **MODIFY** `internal/library/library_test.go`, `internal/library/refresher_test.go`, `internal/library/remover_test.go`, `internal/library/loader_test.go`, `internal/library/saver_test.go`, `internal/library/resolver_test.go` — same migration.

### Phase 3 — Coverage gap fixes

- **MODIFY** `internal/library/creator_test.go` — convert the existing `TestCreateLibrary_DryRun_WritesToStdout` (60 lines) into a table-driven `TestCreateLibrary` with named subtests (`dry-run`, `force-overwrite`, `existing-library-error`, `default-path-resolution`). **NEW** `TestDefaultLibraryYAML` for the 0% `defaultLibraryYAML` function at `creator.go:123`. Per `golang-naming/testing.md`: subtest names are fully lowercase descriptive phrases.
- **(DROPPED)** `internal/library/discovery_test.go` already has comprehensive coverage (`TestFindLibrary`, `TestDefaultLibraryPath`, `TestDefaultLibraryPathXDGDataHome`, `TestResolveLibrary_FlagOverEnvOverCfgOverDefault`, `TestDefaultLibraryPath_AdoptsXDG`, `TestDefaultLibraryPath_PrefersXDGOverCWDWhenXDGExists`, `TestDefaultLibraryPath_FallsBackToCWDWhenXDGDoesNotExist`, `TestXdgReload`). No new tests needed.
- **MODIFY** `internal/library/methods_test.go` — add `TestLibrary_CreatePreset` table-driven tests for `(*Library).CreatePreset` (currently 0%) and the package-level `CreatePreset`. Same scenarios for both forms.
- **NEW** `TestGetOutputPaths` (plural — the map-returning function at `internal/library/resolver.go:185`) in `internal/library/resolver_test.go`. Note: `TestGetOutputPath` (singular) already exists at line 202; do not modify it. The new test covers all four `ResourceType`s (`skill`, `agent`, `command`, `memory`), the `UseSubdirectory` branch, and the `*core.ConfigError` paths.
- **MODIFY** `internal/claude-code/claude_code_adapter_test.go` — add coverage for `parseAgent` (51.8%), `parseTargets` (0%), `mapPermissionModeToPolicy` (28.6%), `renderAgent` (65.9%).
- **NEW** `internal/opencode/opencode_adapter_test.go` — coverage for `parseAgent` (boolean tool map splitting, behavior flattening, permission dual shape), `parseCommand`, `parseSkill`, `parseMemory` (kebab-case keys), `renderAgent` (canonical-to-platform field renaming), and `mapPermissionObjectToPolicy` (after Phase 1 task 1.5b).
- **MODIFY** `internal/core/results_test.go` (or extend `errors_test.go`) — add tests for `(*core.ValidateResult).Valid()` (`internal/core/results.go:17`, 0%) and `Unwrap` chains on typed errors.

### Phase 4 — Golden files + round-trip

- **NEW** `test/e2e/opencode_adapter_golden_test.go` (build tag `e2e`) — **Ginkgo spec** (not standalone `Test*`) in `package e2e_test` per `golang-spf13-cobra` Testing section (cobra accumulates flag state across `Execute()` calls). Placed in the E2E tier per design Decision 6 (sensitive to renderer dependency drift). Compares rendered output against `test/e2e/fixtures/opencode/agent-balanced.md`.
- **NEW** `test/e2e/claude_code_adapter_golden_test.go` (build tag `e2e`) — same Ginkgo pattern for Claude Code; fixture at `test/e2e/fixtures/claude-code/agent-balanced.md`. Both fixtures are `.md` because the renderer emits YAML frontmatter wrapped around a Markdown body for both platforms (`config/templates/<platform>/agent.tmpl`).
- **NEW** `test/e2e/fixtures/opencode/agent-balanced.md` — fixture for the OpenCode byte-equality test (placed under `test/e2e/fixtures/`, not `test/e2e/testdata/`, per existing E2E convention in `test/AGENTS.md`).
- **NEW** `test/e2e/fixtures/claude-code/agent-balanced.md` — fixture for the Claude Code byte-equality test.
- **MODIFY** `internal/renderer/serializer_test.go` — add `TestParseRenderRoundTrip` (canonical round-trip) and `TestPlatformRoundTrip` (platform round-trip). The canonical test reads a fixture, parses via `parser.ParseDocument`, re-emits via `MarshalCanonical`, re-parses the marshal output, and asserts **semantic** equality of the canonical fields. The platform test reads a platform fixture, parses via `parser.ParsePlatformDocument`, marshals via `MarshalCanonical`, re-parses, and asserts the same canonical fields. Both are semantic (not byte) equality — see design Decision 5 for the rationale (YAML serialization is not byte-stable across renderer dependency versions).
- **MODIFY** `internal/canonicalize/canonicalize_golden_test.go:28-30` — replace the CWD-relative `os.Stat("../../test/fixtures/canonical")` skip with `runtime.Caller(0)`-based fixture path resolution; remove the `t.Skip` call so the test runs from any working directory. (Note: the proposal previously referenced `cmd/canonicalize_golden_test.go`, which does not exist — the golden file lives at `internal/canonicalize/canonicalize_golden_test.go` after the `extract-io-adapters` change.)

### Phase 5 — Test parallelization

- **MODIFY** `internal/library/methods_test.go` — add `t.Parallel()` to fixture-driven subtests that don't share `t.Setenv` or `os.Chdir` (e.g., `TestLibrary_X_CtxCancelled`).
- **MODIFY** `cmd/library_add_test.go` — add `t.Parallel()` to fixture-driven subtests at line 383+.

### Phase 6 — Trivial folds (per the reorg plan)

- **FOLD** A-008 — widen the forbidigo pattern at `.golangci.yml:87` to catch all package-level mutable vars (`var defaultAdder`, `var outputFormat`, `var initOutputFormat`, `var outputFormatRefresh`, `var completionShells`, `var errEmptyResources`). Replace with per-options fields or per-command constants.
- **FOLD** A-015 — replace the hard-coded file list in `TestNoNewForbidigoPatterns` (`cmd/lint_test.go:96`) with `go list ./cmd | grep -v _test.go`.
- **FOLD** D-015 — `interface{}` → `any` across all `internal/renderer/serializer.go` signatures (already in Phase 1).
- **FOLD** D-021 — `core.Valid` / `Unwrap` coverage (already in Phase 3).
- **FOLD** D-022 — `t.Parallel()` migration (already in Phase 5).
- **FOLD** D-025 — `internal/canonicalize/canonicalize_golden_test.go` CWD-skip (already in Phase 4).
- **FOLD** D-029 — **DROPPED** (canary race-safety task). The `internal/warning/canary.go` file was deleted by the archived `enforce-error-discipline` change; no race-prone `sync.Once` survives in `cmd/` or `internal/iostreams/`.
- **FOLD** D-030 — `t.Parallel()` in `cmd/library_add_test.go` (already in Phase 5).

### Phase 7 — Goroutine leak detection

Per `golang-testing` Best Practice 6: packages with goroutines SHOULD use `goleak.VerifyTestMain`. `internal/library/adder.go:732-779` uses errgroup with `SetLimit` for concurrent orphan scanning; the new `t.Parallel()` calls in Phase 5 require leak detection.

- **MODIFY** `go.mod` — add `go.uber.org/goleak` dependency: `go get go.uber.org/goleak`.
- **MODIFY** `internal/library/library_test.go` — add `TestMain(m *testing.M)` with `goleak.VerifyTestMain(m)` before `os.Exit(m.Run())`.
- **MODIFY** `internal/library/AGENTS.md` — document the `goleak` dependency in the package's dependency list.

## Capabilities

### New Capabilities

- **`testing-round-trip-tests`** — define the parse→render round-trip contract for semantic equality testing (canonical + platform round-trips); codify the test pattern as a project convention (lives in default suite, no build tag).
- **`testing-adapter-golden-tests`** — define the byte-equality adapter golden test contract; codify the E2E pattern as a project convention (lives in E2E tier under `//go:build e2e`, fixtures are `.md` because the renderer emits frontmatter + body).

### Modified Capabilities

- **`testing-e2e-testing`** — adds 2 NEW requirements: (a) `t.Parallel()` safety in cmd tests (race-condition hazard from `t.Parallel()` + `cmd.Execute()` interactions), and (b) `t.Context()` adoption for new tests. No existing requirement is modified or removed.
- **`transformation-platform-adapters`** — adds 2 new requirements: (a) adapter contract guard (`var _ permission.Adapter = (*Adapter)(nil)` SHALL be present in both adapter files), and (b) adapter singleton (`var OpenCode = &Adapter{}` / `var ClaudeCode = &Adapter{}`; `New()` SHALL return a singleton, not a per-call `&Adapter{}`). No existing requirement is modified or removed.
- **`transformation-permission-transformation`** — adds 1 new requirement: typed `permission.Action` constants (`permission.Allow`, `permission.Ask`, `permission.Deny`) are used in adapter permission maps instead of raw string literals; unknown action values SHALL be rejected at runtime via `*core.ConfigError`. The constants are pre-existing at `internal/permission/permissions.go:38-47`; this change enforces their use-site. No existing requirement is modified or removed.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `cmd/lint_test.go:19` | Hotfix `t.Parallel()` | -1 (hotfix) |
| `cmd/lint_test.go:96` | Dynamic file list | -5 / +3 |
| `internal/opencode/opencode_adapter.go` | Singleton + contract check + typed `Action` in `mapPermissionObjectToPolicy` | -3 / +8 |
| `internal/claude-code/claude_code_adapter.go` | Singleton + contract check | -3 / +3 |
| `internal/permission/permissions.go` | Typed constants (lookup returns `*core.ConfigError`) | -10 / +12 |
| `internal/renderer/serializer.go` | `interface{}` → `any` across all signatures + nil-check | -5 / +10 |
| `cmd/library_add.go:632-637` | Delete dual-path branch | -2 / +0 |
| `internal/library/methods_test.go` | testify migration + CreatePreset table-driven tests | -20 / +60 |
| `internal/library/library_test.go` | testify migration | -3 / +5 |
| `internal/library/refresher_test.go` | testify migration | -25 / +35 |
| `internal/library/remover_test.go` | testify migration | -10 / +15 |
| `internal/library/loader_test.go` | testify migration | -10 / +15 |
| `internal/library/saver_test.go` | testify migration | -15 / +20 |
| `internal/library/resolver_test.go` | testify migration + `TestGetOutputPaths` (plural) | -5 / +60 |
| `internal/library/creator_test.go` (modify) | Table-driven `TestCreateLibrary` + `TestDefaultLibraryYAML` | -5 / +90 |
| `internal/claude-code/claude_code_adapter_test.go` | Coverage | -0 / +200 |
| `internal/opencode/opencode_adapter_test.go` (new) | Coverage | +200 |
| `internal/core/results_test.go` | Valid/Unwrap tests | -0 / +30 |
| `test/e2e/opencode_adapter_golden_test.go` (new) | E2E golden Ginkgo spec | +100 |
| `test/e2e/claude_code_adapter_golden_test.go` (new) | E2E golden Ginkgo spec | +100 |
| `test/e2e/fixtures/opencode/agent-balanced.md` (new) | Fixture | +30 |
| `test/e2e/fixtures/claude-code/agent-balanced.md` (new) | Fixture | +30 |
| `internal/renderer/serializer_test.go` | Round-trip tests | +80 |
| `internal/canonicalize/canonicalize_golden_test.go:28-30` | `runtime.Caller(0)` fixture | -3 / +5 |
| `internal/library/library_test.go` | `TestMain` + goleak | +5 |
| `internal/library/methods_test.go` | `t.Parallel()` | +5 |
| `cmd/library_add_test.go` | `t.Parallel()` | +10 |
| `.golangci.yml:87` | Widen forbidigo pattern | +1 / -1 |
| `cmd/library_add.go:120` (and 5 sibling files) | Remove package-level vars | -5 / +5 |
| `go.mod` / `go.sum` | Add `go.uber.org/goleak` | +2 |
| **Total** | | **~+850 LOC tests, ~-50 LOC prod** |

### Affected systems

- **Coverage:** all packages reach ≥ 70% per `config.testing`, with an 80% aspirational target for `internal/library`, `internal/renderer`, `internal/claude-code`, and `internal/core`.
- **Test stability:** race detector (`go test -race -count=1 ./...`) passes after the hotfix and the `t.Parallel()` migrations. The existing `mise run test:race` task already covers the whole tree (including `cmd/`); no widening needed.
- **CI:** the `mise run test:race` task is validated by Phase 8.4 (already a soft requirement per `config.testing`).
- **Public API:** no production-code API changes (only test internals and adapter internals change).

### Change scope rationale

The 6 phases ship as a single OpenSpec change (and a single PR) for the following reasons:

1. **Shared review theme**: All 17 findings came from one code review (2026-07-08) and share the "test infrastructure" theme. Splitting them creates artificial boundaries between related changes — e.g., Phase 2 (testify migration) and Phase 3 (coverage gap fixes) are interdependent: the newly-migrated testify assertions are easier to write for new coverage tests. Without the migration, Phase 3's new tests would have to be written in the old `t.Fatalf` style and then re-migrated.
2. **Mid-state safety**: Phase 0 (D-001 hotfix) lands separately, before this OpenSpec change. After that, intermediate states (Phase 1 done, Phase 2 not done) are coherent — each commit is independently testable. The order is locked because later phases depend on earlier ones (e.g., Phase 6 forbidigo widening requires the package-level vars from earlier phases to be removed first; otherwise the new pattern flags them).
3. **Review burden**: A single PR is reviewed once. Splitting into 3 changes would require 3 PR reviews, 3 CI runs, 3 OpenSpec archives. The ~850 LOC of test additions is large but trivially reviewable (table-driven tests with explicit scenarios, named subtests per the testing skill).
4. **Risk mitigation**: The design includes rollback per phase (see `design.md` Migration Plan). Each commit is revertable independently — Phase 0/1/2 are low-risk hygiene; Phase 3/4 add coverage; Phase 5/6 are mechanical.

## Risks

- **Hotfix D-001 is a behavioral change for the race detector.** Removing `t.Parallel()` from `cmd/lint_test.go:19` adds 2-3 seconds to the test runtime (no longer parallel with sibling tests). **Mitigation**: the slowness is acceptable for race-detector correctness; `mise run test` (without `-race`) is unaffected.
- **testify migration is mechanical but spans 7 test files.** A missed `t.Fatal` is hard to detect. **Mitigation**: task 2.8 runs `rg "t\.Fatal|t\.Error" internal/library/*_test.go` after the migration; zero matches expected outside of test-only assertions.
- **New test files** (creator_test.go convert+extend, opencode_adapter_test.go new, resolver_test.go `TestGetOutputPaths` plural, and the new E2E golden Ginkgo specs under `test/e2e/`) add ~850 LOC. Coverage increase is real but the tests must actually exercise the code, not just call constructors. **Mitigation**: design Decision 4 requires table-driven tests with concrete inputs and expected outputs (matching the `library-library-batch-add` precedent).
- **Adapter singleton** (Phase 1) is a behavior change for the `New()` callers. **Mitigation**: the new `OpenCode` / `ClaudeCode` package-level vars are drop-in replacements; all `claudecode.New()` / `opencode.New()` call sites are updated in the same commit.
