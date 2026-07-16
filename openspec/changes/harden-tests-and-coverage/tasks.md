# Tasks — Harden tests and coverage

The 17 findings from the 2026-07-08 code review ship in **one PR with 6 atomic phases** (each commit is independently testable). The D-001 hotfix (`cmd/lint_test.go:19` `t.Parallel()` removal) lands separately, before this OpenSpec change, and is documented in `proposal.md` Phase 0 only.

## 1. Phase 1 — Adapter contract + singleton + typed constants

- [ ] 1.1 Add `var _ permission.Adapter = (*Adapter)(nil)` at the **end of `internal/opencode/opencode_adapter.go`** (after all method definitions). Add a one-line comment: `// Compile-time check: *Adapter satisfies permission.Adapter. Mirror of cmd/canonicalize_test.go:20 precedent.`
- [ ] 1.2 Add `var _ permission.Adapter = (*Adapter)(nil)` at the **end of `internal/claude-code/claude_code_adapter.go`** (after all method definitions). Same comment as 1.1.
- [ ] 1.3 In `internal/opencode/opencode_adapter.go`, replace `func New() *Adapter { return &Adapter{} }` with `var OpenCode = &Adapter{}` at the package level. Update all `opencode.New()` call sites in `internal/parser/platform_parser.go` and `internal/renderer/serializer.go` to use `opencode.OpenCode` directly; update the example in `internal/opencode/doc.go`.
- [ ] 1.4 In `internal/claude-code/claude_code_adapter.go`, replace `func New() *Adapter { return &Adapter{} }` with `var ClaudeCode = &Adapter{}` at the package level. Update all `claudecode.New()` call sites in `internal/parser/platform_parser.go` and `internal/renderer/serializer.go` to use `claudecode.ClaudeCode` directly.
- [ ] 1.5 The typed `permission.Action`, `permission.Allow`, `permission.Ask`, `permission.Deny` constants are pre-existing at `internal/permission/permissions.go:38-47`. Update the permission maps in `internal/opencode/opencode_adapter.go` and `internal/claude-code/claude_code_adapter.go` (and any caller that constructs a `permission.Map` literal) to use the typed constants instead of string literals; type each map value as `permission.Action` where the underlying field allows. Verify the `PermissionPolicyMappings` lookup returns `*core.ConfigError` (not `*core.ValidationError`) for unknown policy values — change the lookup if it currently returns `*core.ValidationError`, for consistency with `internal/parser/platform_parser.go:50` (unknown platform).
- [ ] 1.5b In `internal/opencode/opencode_adapter.go::mapPermissionObjectToPolicy` (lines 486-509), extract raw `actionStr` values into typed `permission.Action` constants before comparison. The function operates on raw YAML `map[string]interface{}` (not the typed `permission.Map`), so the comparison must convert via `permission.Action(actionStr)` before equality checks. After the edit, `rg '"allow"|"ask"|"deny"' internal/opencode/opencode_adapter.go` must return zero matches.
- [ ] 1.6 In `internal/renderer/serializer.go`, replace ALL `interface{}` with `any` in signatures at lines 22, 23, 27, 34, 54, 90, 224, 243, 294, 335. Add an explicit nil-check at the top of `getDocType` (line 224) that returns `*core.TransformError` for `doc == nil`.
- [ ] 1.7 In `cmd/library_add.go::collectDiscoverFailures` (lines 614-671), delete the dual-path `c.Cause vs errors.New(c.Issue)` branch at lines 632-637. The producer-side `adder.go:835-845` already sets `Cause` correctly via `conflictErr`, so no `adder.go` change is required. After the edit, `rg 'errors\.New\(c\.Issue\)' cmd/library_add.go` must return zero matches.
- [ ] 1.8 Run `rg "opencode\.New\(\)|claudecode\.New\(\)" .` — must return zero matches (singleton migration complete).
- [ ] 1.9 Run `rg '"allow"|"ask"|"deny"' internal/opencode/ internal/claude-code/` — must return zero matches in adapter permission maps AND `mapPermissionObjectToPolicy`.
- [ ] 1.10 Run `mise run check` — must pass.

## 2. Phase 2 — Testify migration

Per `golang-stretchr-testify` "Rule": use `require` for preconditions (file writes, `LoadLibrary` errors) and `assert` for value verifications. Use `assert.Equal(t, expected, actual)` argument order. Per `golang-lint`, `testifylint` is enabled and will catch wrong argument order and assert/require misuse.

- [ ] 2.1 In `internal/library/methods_test.go`, replace raw `t.Fatalf` / `t.Errorf` / `t.Fatal` calls with `require.NoError` / `require.Error` / `assert.Equal` / `assert.True`. testify is already imported — no import changes needed.
- [ ] 2.2 In `internal/library/library_test.go`, same migration. Add the canonical import lines:
  ```go
  import (
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )
  ```
- [ ] 2.3 In `internal/library/refresher_test.go`, same migration. Add the same import lines as 2.2.
- [ ] 2.4 In `internal/library/remover_test.go`, same migration. testify is already imported.
- [ ] 2.5 In `internal/library/loader_test.go`, same migration. testify is already imported.
- [ ] 2.6 In `internal/library/saver_test.go`, same migration. Add the same import lines as 2.2.
- [ ] 2.7 In `internal/library/resolver_test.go`, same migration. testify is already imported.
- [ ] 2.8 Run `rg "t\.Fatal|t\.Error" internal/library/*_test.go` — must return zero matches.
- [ ] 2.9 Run `golangci-lint run --enable-only testifylint ./...` — no wrong-argument-order or assert/require-misuse findings.
- [ ] 2.10 Run `mise run test` — all unit tests pass.
- [ ] 2.11 Run `mise run test:race` — no race conditions.

## 3. Phase 3 — Coverage gap fixes

Per `golang-testing` Best Practice 1: table-driven tests must use named subtests. Per `golang-naming/testing.md`: subtest names are fully lowercase descriptive phrases (e.g., `"dry-run"`, `"force-overwrite"`, `"existing-library-error"`).

- [ ] 3.1a **MODIFY** `internal/library/creator_test.go`: convert the existing `TestCreateLibrary_DryRun_WritesToStdout` (single test function at lines 24-60) into a table-driven `TestCreateLibrary` with named subtests via `t.Run`:
  - `"dry-run"` (existing — assert each preview substring present in captured `bytes.Buffer`)
  - `"force-overwrite"` (create existing lib, call with `Force: true`, assert success)
  - `"existing-library-error"` (create existing lib, call without `Force`, assert `*core.OperationError`)
  - `"default-path-resolution"` (call without `Path`, assert XDG-derived path used)
- [ ] 3.1b **NEW** `TestDefaultLibraryYAML` in `internal/library/creator_test.go`: table-driven test for `defaultLibraryYAML` at `creator.go:123`. Cases:
  - `"version field present"` (assert contains `version: "1"`)
  - `"empty resources map"` (assert resources block empty)
  - `"empty presets map"` (assert presets block empty)
- [ ] 3.2 (DROPPED — `internal/library/discovery_test.go` already has comprehensive coverage: `TestFindLibrary`, `TestDefaultLibraryPath`, `TestDefaultLibraryPathXDGDataHome`, `TestResolveLibrary_FlagOverEnvOverCfgOverDefault`, `TestDefaultLibraryPath_AdoptsXDG`, `TestDefaultLibraryPath_PrefersXDGOverCWDWhenXDGExists`, `TestDefaultLibraryPath_FallsBackToCWDWhenXDGDoesNotExist`, `TestXdgReload`. No new tests needed; coverage is already in place.)
- [ ] 3.3 In `internal/library/methods_test.go`, add `TestLibrary_CreatePreset` table-driven tests for:
  - `(*Library).CreatePreset` (currently 0%): covers success, empty name, duplicate name, references validation.
  - The package-level `CreatePreset`: same scenarios.
- [ ] 3.4 **NEW** `TestGetOutputPaths` (plural — the map-returning function at `internal/library/resolver.go:185`) in `internal/library/resolver_test.go`. Note: `TestGetOutputPath` (singular) already exists at line 202; do not modify it. The new test covers:
  - All four `ResourceType`s (`skill`, `agent`, `command`, `memory`).
  - `UseSubdirectory` branch (skill uses `.opencode/skills/<name>/SKILL.md`).
  - `*core.ConfigError` paths for invalid platform / invalid type.
  - Mixed valid/invalid refs in the same call (partial-failure semantics).
- [ ] 3.5 In `internal/claude-code/claude_code_adapter_test.go`, add coverage tests for:
  - `parseAgent` (51.8% → 90%+): covers all permission modes (`default`, `acceptEdits`, `dontAsk`, `plan`, `bypassPermissions`), and both `tools` / `disallowedTools` array shapes (`[]interface{}` and `[]string`).
  - `parseTargets` (0% → 100%): covers `claude-code` target extraction.
  - `mapPermissionModeToPolicy` (28.6% → 100%): covers all 5 modes plus the unknown-mode fallback.
  - `renderAgent` (65.9% → 90%+): covers non-empty `Targets`, `PermissionPolicy`, `Behavior` (mode, temperature, steps), and `Extensions.Hooks`.
- [ ] 3.6 In `internal/opencode/opencode_adapter_test.go`, add coverage tests for:
  - `parseAgent`: covers tool boolean map splitting (true → tools array, false → disallowedTools array), behavior flattening (mode/temperature/steps/hidden/prompt/disabled), permission dual shape (`permissionMode` string + nested `permission` object).
  - `parseCommand`, `parseSkill`, `parseMemory`: cover each doc-type's shape, including kebab-case keys (`allowed-tools`, `argument-hint`, `user-invocable`).
  - `renderAgent`: covers behavior flattening (canonical `behavior.disabled` → OpenCode `disable`; canonical `behavior.steps` → OpenCode `maxSteps`).
  - `mapPermissionObjectToPolicy` (after task 1.5b): covers edit/bash deny flags and per-tool `Allow`/`Ask`/`Deny` decoding.
- [ ] 3.7 In `internal/core/results_test.go` (or extend `errors_test.go`), add tests for:
  - `(*core.ValidateResult).Valid()` (currently 0%): covers the empty-errors and non-empty-errors branches; test against real `PermissionPolicy` values (`restrictive`, `balanced`, `permissive`, `analysis`, `unrestricted`).
  - `Unwrap` chains on typed errors that actually implement `Unwrap()`: `ParseError`, `ValidationError`, `TransformError`, `FileError`, `InitializeError`, `OperationError`, `CobraUsageError`. Each test wraps a sentinel cause via `errors.Is` and asserts the cause is reachable through the chain.
  - `Unwrap` does NOT exist on `NotFoundError`, `ConfigError`, `UsageError`, `PartialSuccessError`: tests SHALL assert the wrapped error message instead of a chain (the failure mode is "no chain returned", not "wrong cause").
- [ ] 3.8 Run `mise run test:coverage` — verify the four target packages (`internal/library`, `internal/claude-code`, `internal/core`, `internal/renderer`) reach ≥ 80% as the aspirational target (≥ 70% is the floor per `config.testing`).
- [ ] 3.9 Verify all new test code uses `t.TempDir()` (auto-cleanup) and `t.Setenv()` (auto-restore) instead of `os.MkdirTemp` + manual cleanup. Run `rg 'os\.MkdirTemp|os\.Setenv' internal/library/*_test.go` — must return zero matches in NEW tests.

## 4. Phase 4 — Round-trip + E2E golden tests

The byte-equality adapter golden tests live in the **E2E tier** (`test/e2e/`, `//go:build e2e`) per design Decision 6 — they assert against renderer output which is sensitive to dependency drift, so they belong in a separate CI stage from the default suite. The round-trip tests stay in the default suite (`internal/renderer/`) because they assert semantic equality, which is deterministic. Round-trip tests live in default suite per design Decision 5; E2E byte-equality tests live in E2E tier per design Decision 6.

Per `golang-spf13-cobra` Testing section: cobra accumulates flag state across `Execute()` calls. The new E2E tests must be **Ginkgo `Describe`/`It` blocks within `package e2e_test`**, NOT standalone `Test*` functions (cobra state would leak).

- [ ] 4.1a Create `test/e2e/opencode_adapter_golden_test.go` (build tag `e2e`) with a **Ginkgo spec** in `package e2e_test`:
  ```go
  //go:build e2e

  package e2e_test

  var _ = Describe("OpenCode adapter byte-equality rendering", func() {
      It("renders permission-balanced agent byte-equally", func() {
          cli := helpers.NewGerminatorCLI(e2e.BinaryPath())
          session := cli.Run("adapt", fixturePath, outPath, "--platform", "opencode")
          cli.ShouldSucceed(session)
          cli.ShouldOutput(session, "wrote "+outPath)
          // Compare <out> against test/e2e/fixtures/opencode/agent-balanced.md
          gotBytes, err := os.ReadFile(outPath)
          Expect(err).NotTo(HaveOccurred())
          wantBytes, err := os.ReadFile(fixturePath)
          Expect(err).NotTo(HaveOccurred())
          Expect(gotBytes).To(Equal(wantBytes))
      })
  })
  ```
  Build tag: `//go:build e2e` at line 1.
- [ ] 4.2a Create `test/e2e/claude_code_adapter_golden_test.go` (build tag `e2e`) with a similar **Ginkgo spec** for Claude Code. Fixture at `test/e2e/fixtures/claude-code/agent-balanced.md`.
- [ ] 4.3a Create `test/e2e/fixtures/opencode/agent-balanced.md` — fixture for the OpenCode permission rendering test. Per `golang-testing` (Go 1.26+ test artifacts) and the project convention, fixtures live under `test/e2e/fixtures/`, not `test/e2e/testdata/`.
- [ ] 4.4a Create `test/e2e/fixtures/claude-code/agent-balanced.md` — fixture for the Claude Code permission rendering test. Renderer output for both platforms is YAML frontmatter + Markdown body (`config/templates/<platform>/agent.tmpl`), so both fixtures use the `.md` extension.
- [ ] 4.5 In `internal/renderer/serializer_test.go`, add `TestParseRenderRoundTrip` (canonical round-trip):
  - Reads `test/fixtures/canonical/agent-permission-balanced.md` (or another canonical fixture with non-default fields).
  - Parses via `parser.ParseDocument(t.Context(), inputPath, "agent")`, returning `*parser.CanonicalAgent`.
  - Marshals via `renderer.MarshalCanonical(t.Context(), doc)`.
  - Writes the marshaled string to a `t.TempDir()` file and re-parses it via `parser.ParseDocument(t.Context(), tmpPath, "agent")`.
  - Asserts the full canonical `core.Agent` field set is equal between the two parses: `Name`, `Description`, `Tools`, `DisallowedTools`, `PermissionPolicy`, `Behavior.Mode`, `Behavior.Temperature`, `Behavior.Steps`, `Behavior.Prompt`, `Behavior.Hidden`, `Behavior.Disabled`, `Model`, `Targets`, `Extensions`.

  Note: this is a **semantic** equality test (key fields compared, not bytes). It is distinct from the byte-equality golden tests in tasks 4.1a-4.4a, which live in the E2E tier. The round-trip test stays in the default suite (no build tag) because semantic equality is deterministic.
- [ ] 4.6 Add `TestParseRenderRoundTrip` subtests for `Skill`, `Command`, `Memory` doc types (mirroring the existing `Agent` subtest).
- [ ] 4.7 Add `TestPlatformRoundTrip` for `claude-code` and `opencode` that round-trips a platform fixture through canonical marshaling.
- [ ] 4.8 In `internal/canonicalize/canonicalize_golden_test.go:28-30`, replace the CWD-relative `os.Stat("../../test/fixtures/canonical")` skip with a `runtime.Caller(0)`-based fixture path resolution. Remove the `t.Skip` call so the test runs from any working directory:
  ```go
  _, thisFile, _, _ := runtime.Caller(0)
  fixturesDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "test", "fixtures")
  goldenDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "test", "golden", "canonical")
  ```
- [ ] 4.9 Run `mise run test:e2e` — the new E2E golden tests (4.1a-4.4a) pass.
- [ ] 4.10 Run `go test ./internal/renderer/...` — round-trip tests (4.5-4.7) pass in the default suite.
- [ ] 4.11 Run `go test -tags=golden ./...` — pre-existing golden tests pass after the CWD-skip fix (4.8).

## 5. Phase 5 — Test parallelization

Per `golang-testing` Best Practice 4: independent tests SHOULD use `t.Parallel()`. Per `golang-lint` (`tparallel` linter, already enabled): requires all-or-none parallelism per parent test (mixed parallelism is forbidden).

- [ ] 5.1 In `internal/library/methods_test.go`, add `t.Parallel()` to fixture-driven subtests within `t.Run(...)` blocks that don't share `t.Setenv` or `os.Chdir`. Specifically: `TestLibrary_X_CtxCancelled` subtests.
- [ ] 5.2 In `cmd/library_add_test.go`, add `t.Parallel()` ONLY to subtests inside `t.Run(...)` blocks that don't share `t.Setenv` / `os.Chdir` / `t.TempDir` with sibling subtests in the same parent test. Specifically target `TestLibrary_X_CtxCancelled` subtests at line 383+. Verify with `golangci-lint run --enable-only tparallel ./...` to confirm `tparallel` compliance.
- [ ] 5.3 Run `mise run test:race` — no race conditions from the new `t.Parallel()` calls. The existing `mise run test:race` task already covers `./...` (including `cmd/`); no widening needed.
- [ ] 5.4 Run `mise run test:coverage` — coverage unchanged (parallel subtests don't reduce coverage).

## 6. Phase 6 — Trivial folds

- [ ] 6.1 In `.golangci.yml:87`, widen the forbidigo pattern:
  ```yaml
  forbidigo:
    patterns:
      - pattern: 'var (defaultAdder|outputFormat|initOutputFormat|outputFormatRefresh|completionShells|errEmptyResources)\b'
        msg: 'package-level mutable variables are forbidden per cmd/AGENTS.md:46'
  ```
- [ ] 6.2 In `cmd/library_add.go:120`, replace `var defaultAdder resourceAdder` with a per-options field `defaultAdder` (or use a constant if the value is immutable).
- [ ] 6.3 In `cmd/library_add.go`, replace `var outputFormat` with a per-options field. In `cmd/library_init.go`, replace `var initOutputFormat` with a per-options field.
- [ ] 6.4 In `cmd/library_refresh.go`, replace `var outputFormatRefresh` with a per-options field.
- [ ] 6.5 In `cmd/completion.go`, replace `var completionShells` with a per-command constant or per-options field.
- [ ] 6.6 In `cmd/library_create.go:67`, migrate the package-level `var errEmptyResources = core.NewUsageError(...)` to an inline construction inside `runCreatePreset`. At the return site around `cmd/library_create.go:173`, return `core.NewUsageError("--resources", "must be non-empty list of refs")` directly; delete the `var errEmptyResources` declaration. The `enforce-error-discipline` change does not need to revisit this file.
- [ ] 6.7 In `cmd/lint_test.go:96` (`TestNoNewForbidigoPatterns`), replace the hard-coded `[]string{"adapt.go", "resources.go", "presets.go", "show.go"}` slice with a dynamic file list. The `go list ./cmd` invocation alone returns the package import path, not file names — use the GoFiles template instead:
  ```go
  out, _ := exec.Command("go", "list", "-f", "{{range .GoFiles}}{{.}}\n{{end}}", "./cmd").Output()
  nonTestFiles := lo.Filter(
      strings.Split(strings.TrimSpace(string(out)), "\n"),
      func(f string, _ int) bool { return !strings.HasSuffix(f, "_test.go") },
  )
  ```
  This emits one filename per line; the `lo.Filter` call (or equivalent hand-written filter) excludes test files since forbidigo runs against production code.
- [ ] 6.8 Run `golangci-lint run --verbose` to verify the widened forbidigo pattern catches exactly 6 declarations (`defaultAdder`, `outputFormat`, `initOutputFormat`, `outputFormatRefresh`, `completionShells`, `errEmptyResources`) with zero false positives.
- [ ] 6.9 Run `mise run lint` — must report 0 issues.
- [ ] 6.10 Run `mise run check` — full validation passes.

## 7. Goroutine leak detection

Per `golang-testing` Best Practice 6: packages with goroutines SHOULD use `goleak.VerifyTestMain` to detect goroutine leaks. `internal/library/adder.go:732-779` uses errgroup with `SetLimit` for concurrent orphan scanning; the new `t.Parallel()` calls in Phase 5 require leak detection.

- [ ] 7.1 Add `go.uber.org/goleak` dependency: `go get go.uber.org/goleak`. Verify `go.mod` and `go.sum` are updated.
- [ ] 7.2 Add `TestMain(m *testing.M)` to `internal/library/library_test.go`:
  ```go
  func TestMain(m *testing.M) {
      goleak.VerifyTestMain(m)
      os.Exit(m.Run())
  }
  ```
- [ ] 7.3 Verify `go test ./internal/library/...` passes with no goroutine leaks from `processScanFile` or `BatchAddResources`.
- [ ] 7.4 Document the dependency in `internal/library/AGENTS.md` (add to dependencies list if not present).

## 8. Verification

- [ ] 8.1 Run `mise run build` — no broken imports.
- [ ] 8.2 Run `mise run lint` — must report 0 issues.
- [ ] 8.3 Run `mise run test` — all unit tests pass.
- [ ] 8.4 Run `mise run test:race` — no race conditions; the existing `mise run test:race` task (already covers `./...`) catches the D-001 Cobra race and validates new `t.Parallel()` calls.
- [ ] 8.5 Run `mise run test:coverage` — every package reaches ≥ 70% per `config.testing` (aspirational target 80% for `internal/library`, `internal/claude-code`, `internal/core`, `internal/renderer`).
- [ ] 8.6 Run `mise run test:e2e` — E2E golden tests (4.1a-4.4a, byte-equality of frontmatter + body) pass; default-suite round-trip tests (4.5-4.7, semantic) pass under `go test ./internal/renderer/...`. Pre-existing E2E tests continue to pass (regression check; same task).
- [ ] 8.7 Run `rg "var (defaultAdder|outputFormat|initOutputFormat|outputFormatRefresh|completionShells|errEmptyResources)\b" cmd/` — must return zero matches (forbidigo migration complete, including the 6.6 inline refactor).
- [ ] 8.8 Run `rg "opencode\.New\(\)|claudecode\.New\(\)" .` — must return zero matches (singleton migration complete).
- [ ] 8.9 Run `openspec validate harden-tests-and-coverage --strict` — change is coherent.
