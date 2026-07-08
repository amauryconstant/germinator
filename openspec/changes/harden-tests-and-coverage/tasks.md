# Tasks — Harden tests and coverage

**Hotfix D-001 lands first**, outside this OpenSpec change. The remaining tasks ship in **one PR with 6 atomic phases** (each commit is independently testable).

## 0. Hotfix (outside OpenSpec)

- [ ] 0.1 In `cmd/lint_test.go:19`, remove `t.Parallel()` from `TestLintBaseline` (the test forks `mise` processes and races on Cobra globals).
- [ ] 0.2 Run `go test -race -count=1 ./cmd/...` — must pass without race conditions.
- [ ] 0.3 Run `mise run test:e2e` — must pass.

## 1. Phase 1 — Adapter contract + singleton + typed constants + canary mutex

- [ ] 1.1 In `internal/opencode/opencode_adapter.go`, add `var _ permission.Adapter = (*Adapter)(nil)` at the bottom of the file.
- [ ] 1.2 In `internal/claude-code/claude_code_adapter.go`, add `var _ permission.Adapter = (*Adapter)(nil)` at the bottom of the file.
- [ ] 1.3 In `internal/opencode/opencode.go`, replace `func New() *Adapter { return &Adapter{} }` with `var OpenCode = &Adapter{}` at the package level. Update all `opencode.New()` call sites to use `opencode.OpenCode` directly.
- [ ] 1.4 In `internal/claude-code/claude_code.go`, replace `func New() *Adapter { return &Adapter{} }` with `var ClaudeCode = &Adapter{}` at the package level. Update all `claudecode.New()` call sites to use `claudecode.ClaudeCode` directly.
- [ ] 1.5 In `internal/permission/permissions.go`, define `type Action string` and constants `Allow Action = "allow"`, `Ask Action = "ask"`, `Deny Action = "deny"`. Update the permission maps in `internal/opencode/opencode_adapter.go` and `internal/claude-code/claude_code_adapter.go` to use the typed constants.
- [ ] 1.6 In `internal/renderer/serializer.go:216`, change `func getDocType(doc interface{})` to `func getDocType(doc any)`. Add explicit nil-check at the top of the function.
- [ ] 1.7 In `internal/warning/canary.go:51`, wrap `canaryOnce` in a `sync.Mutex`:
  ```go
  var (
      canaryOnceMu sync.Mutex
      canaryOnce   = &sync.Once{}
  )
  func ResetCanaryForTest() {
      canaryOnceMu.Lock()
      defer canaryOnceMu.Unlock()
      canaryOnce = &sync.Once{}
  }
  ```
  Update `MaybeWarnLegacyExitCode` to acquire the mutex when calling `canaryOnce.Do(...)`.
- [ ] 1.8 In `internal/library/adder.go:948`, always set `c.Cause` in `checkNameConflict`:
  ```go
  func (lib *Library) checkNameConflict(...) *ConflictInfo {
      // ...
      return &ConflictInfo{
          Type: orphan.Type,
          Name: orphan.Name,
          Path: orphan.Path,
          Issue: fmt.Sprintf("%s/%s: name conflict", orphan.Type, orphan.Name),
          Cause: ErrNameConflict,  // always set, even if synthesized
      }
  }
  ```
  Update `collectDiscoverFailures` to use `c.Cause` (drop the dual-path `c.Cause vs errors.New(c.Issue)` branch).
- [ ] 1.9 Run `rg "opencode\.New\(\)|claudecode\.New\(\)" .` — must return zero matches (singleton migration complete).
- [ ] 1.10 Run `rg "\"allow\"|\"ask\"|\"deny\"" internal/opencode/ internal/claude-code/` — must return zero matches in the adapter permission maps.
- [ ] 1.11 Run `mise run check` — must pass.

## 2. Phase 2 — Testify migration

- [ ] 2.1 In `internal/library/methods_test.go`, replace raw `t.Fatalf` / `t.Errorf` / `t.Fatal` with `require.NoError` / `require.Error` / `assert.Equal` / `assert.True`. Add `import "github.com/stretchr/testify/{assert,require}"`.
- [ ] 2.2 In `internal/library/library_test.go`, same migration.
- [ ] 2.3 In `internal/library/refresher_test.go`, same migration.
- [ ] 2.4 In `internal/library/remover_test.go`, same migration.
- [ ] 2.5 In `internal/library/loader_test.go`, same migration.
- [ ] 2.6 In `internal/library/saver_test.go`, same migration.
- [ ] 2.7 In `internal/library/resolver_test.go`, same migration.
- [ ] 2.8 Run `rg "t\.Fatal|t\.Error" internal/library/*_test.go` — must return zero matches.
- [ ] 2.9 Run `mise run test` — all unit tests pass.
- [ ] 2.10 Run `mise run test:race` — no race conditions.

## 3. Phase 3 — Coverage gap fixes

- [ ] 3.1 Create `internal/library/creator_test.go` with table-driven tests for:
  - `CreateLibrary` (43.5% → 90%+): covers dry-run, force-overwrite, existing-library error, default-path resolution.
  - `defaultLibraryYAML` (0% → 100%): covers version field, empty resources/presets.
- [ ] 3.2 Create `internal/library/discovery_test.go` with table-driven tests for:
  - `DefaultLibraryPath` (18.2% → 100%): covers XDG_DATA_HOME set, XDG_DATA_HOME unset, macOS, Windows, env-var override.
  - `FindLibrary` priority resolution: covers explicit flag, env var, default.
- [ ] 3.3 In `internal/library/methods_test.go`, add `TestLibrary_CreatePreset` table-driven tests for:
  - `(*Library).CreatePreset` (0% → 90%+): covers success, empty name, duplicate name, references validation.
  - The package-level `CreatePreset` (0% → 90%+): same scenarios.
- [ ] 3.4 In `internal/library/resolver_test.go`, add `TestGetOutputPaths` (0% → 100%): covers canonical format output, platform-specific format output, missing-format fallback.
- [ ] 3.5 In `internal/claude-code/claude_code_adapter_test.go`, add coverage tests for:
  - `parseAgent` (51.8% → 90%+): covers all permission modes (`default`, `acceptEdits`, `dontAsk`, `plan`, `bypassPermissions`).
  - `parseTargets` (0% → 100%): covers `claude-code` target extraction.
  - `mapPermissionModeToPolicy` (28.6% → 100%): covers all 5 modes plus the unknown-mode fallback.
  - `renderAgent` (65.9% → 90%+): covers non-empty `Targets`, `PermissionPolicy`, `Behavior` (mode, temperature, steps).
- [ ] 3.6 In `internal/core/results_test.go` (or extend `errors_test.go`), add tests for:
  - `core.Valid` (0% → 100%): covers valid and invalid `PermissionPolicy` values.
  - `Unwrap` chains on typed errors (0% → 100%): covers `ParseError`, `ValidationError`, `TransformError`, `FileError`, `ConfigError`, `NotFoundError`, `OperationError`, `InitializeError`, `PartialSuccessError`.
- [ ] 3.7 Run `mise run test:coverage` — verify `internal/library` ≥ 80%, `claude-code` ≥ 80%, `core` ≥ 80%, `renderer` ≥ 80%.

## 4. Phase 4 — Golden files + round-trip

- [ ] 4.1 Create `internal/opencode/opencode_adapter_golden_test.go` (build tag `golden`) with:
  - `TestOpenCodeAdapter_GoldenPermissionRendering`: reads a canonical agent fixture with `permissionPolicy=balanced`; renders via `OpenCode.RenderDocument`; compares the output to `test/golden/opencode/agent-balanced.yaml` byte-for-byte.
  - Build tag: `//go:build golden` at line 1.
- [ ] 4.2 Create `internal/claude-code/claude_code_adapter_golden_test.go` (build tag `golden`) with:
  - `TestClaudeCodeAdapter_GoldenPermissionRendering`: same pattern for Claude Code.
- [ ] 4.3 Create `test/golden/opencode/agent-balanced.yaml` — fixture for the OpenCode permission rendering test.
- [ ] 4.4 Create `test/golden/claude-code/agent-balanced.json` — fixture for the Claude Code permission rendering test.
- [ ] 4.5 In `internal/renderer/serializer_test.go`, add `TestParseRenderRoundTrip`:
  - Reads `test/fixtures/canonical/agent.md` (or similar).
  - Parses via `parser.ParsePlatformDocument(inputPath, "claude-code", "agent")`.
  - Marshals via `renderer.MarshalCanonical(doc)`.
  - Re-parses the marshaled output.
  - Asserts `Name`, `Description`, `Mode`, `Temperature`, `Steps`, `Hidden`, `Disabled`, `Tools`, `PermissionPolicy` are equal between the two parses.
- [ ] 4.6 Add similar round-trip tests for `Skill`, `Command`, `Memory` doc types.
- [ ] 4.7 In `cmd/canonicalize_golden_test.go:26`, compute the fixture path via `runtime.Caller(0)`:
  ```go
  _, thisFile, _, _ := runtime.Caller(0)
  fixtureDir := filepath.Join(filepath.Dir(thisFile), "..", "test", "fixtures", "canonical")
  ```
  This eliminates the CWD-skip; the test runs from any working directory.
- [ ] 4.8 Run `go test -tags=golden ./...` — all golden tests pass.
- [ ] 4.9 Run `go test ./internal/renderer/...` — round-trip tests pass.

## 5. Phase 5 — Test parallelization

- [ ] 5.1 In `internal/library/methods_test.go`, add `t.Parallel()` to fixture-driven subtests within `t.Run(...)` blocks that don't share `t.Setenv` or `os.Chdir`. Specifically: `TestLibrary_X_CtxCancelled` subtests.
- [ ] 5.2 In `cmd/library_add_test.go`, add `t.Parallel()` to fixture-driven subtests at line 383+ that don't share `t.Setenv` / `os.Chdir` with sibling tests.
- [ ] 5.3 Run `mise run test:race` — no race conditions from the new `t.Parallel()` calls.
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
- [ ] 6.3 In `cmd/library_add.go`, replace `var outputFormat` and `var initOutputFormat` with per-options fields.
- [ ] 6.4 In `cmd/library_refresh.go`, replace `var outputFormatRefresh` with a per-options field.
- [ ] 6.5 In `cmd/completion.go`, replace `var completionShells` with a per-command constant or per-options field.
- [ ] 6.6 In `cmd/library_add.go:82`, replace `var errEmptyResources` with a per-options field or a typed error constructor.
- [ ] 6.7 In `cmd/lint_test.go:96` (`TestNoNewForbidigoPatterns`), replace the hard-coded `[]string{"adapt.go", "resources.go", "presets.go", "show.go"}` slice with `out, _ := exec.Command("go", "list", "./cmd").Output(); strings.Split(string(out), "\n")` filtered to exclude `_test.go` files.
- [ ] 6.8 Run `mise run lint` — must report 0 issues (the widened forbidigo pattern may flag the package-level vars; the migration in 6.2-6.6 must precede the lint check).
- [ ] 6.9 Run `mise run check` — full validation passes.

## 7. Verification

- [ ] 7.1 Run `mise run build` — no broken imports.
- [ ] 7.2 Run `mise run lint` — must report 0 issues.
- [ ] 7.3 Run `mise run test` — all unit tests pass.
- [ ] 7.4 Run `mise run test:race` — no race conditions; hotfix D-001 and the new `t.Parallel()` calls do not race.
- [ ] 7.5 Run `mise run test:coverage` — every package reaches ≥ 70% (target 80% for `internal/library`).
- [ ] 7.6 Run `go test -tags=golden ./...` — golden tests pass.
- [ ] 7.7 Run `mise run test:e2e` — E2E tests pass.
- [ ] 7.8 Run `rg "var (defaultAdder|outputFormat|initOutputFormat|outputFormatRefresh|completionShells|errEmptyResources)\b" cmd/` — must return zero matches (forbidigo migration complete).
- [ ] 7.9 Run `rg "opencode\.New\(\)|claudecode\.New\(\)" .` — must return zero matches (singleton migration complete).
- [ ] 7.10 Run `openspec validate harden-tests-and-coverage --strict` — change is coherent.

## 8. Archive

- [ ] 8.1 Apply spec deltas via `osc-sync-specs`.
- [ ] 8.2 Archive this change via `osc-archive-change harden-tests-and-coverage`.
- [ ] 8.3 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.
