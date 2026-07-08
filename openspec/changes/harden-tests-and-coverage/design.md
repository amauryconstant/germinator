## Context

The 2026-07-08 review identified 17 test-infra findings clustered in 4 areas. The codebase's lint baseline is clean (`mise run lint` reports 0 issues) and the test surface is large (17 E2E test files, 56+ golden files, 57 fixtures), but the infrastructure has 4 systemic issues:

1. **Race conditions in cmd tests** — `cmd/lint_test.go:19` uses `t.Parallel()` while calling `cmd.Execute()`. Cobra's package-level `OnInitialize` slice is mutated by every command registration; concurrent reads/writes trigger race detector failures. The race cascades: dozens of cmd tests that call `cmd.Execute()` race on Cobra globals.

2. **Coverage gaps in `internal/library`, `claude-code`, `renderer`, `core`** — multiple functions are at 0% coverage. The coverage rules at `config.testing` specify ≥ 70% per package, but `internal/library` is at 79.4% with 6 functions at 0%. The coverage report is below the project goal.

3. **Testify migration incomplete** — `cmd/` tests use testify (`require.NoError`, `assert.Equal`); library tests use raw `t.Fatalf` / `t.Errorf`. The mix is inconsistent and prevents test-pattern reuse.

4. **Adapter contract is unenforced** — `permission.Adapter` is declared in `internal/permission/adapter.go:8-13` but no `var _ permission.Adapter = (*Adapter)(nil)` check exists. A signature drift in either adapter is detected only at the call site, not at compile time.

### Constraints

1. **Hotfix D-001 lands first**, outside the OpenSpec change. The Blocker finding is a race-detector correctness issue; it must not wait for the larger change.
2. **Performance findings** (C-010..C-018, C-024, C-025) are deferred to a separate `perf-hardening` change per the reorg plan.
3. **Test files use `t.Parallel()` carefully** — only tests that don't share `t.Setenv` / `os.Chdir` / Cobra globals. The hotfix removes `t.Parallel()` from `cmd/lint_test.go` because the test forks `mise` processes (not because of Cobra globals; the per-test `cmd.Execute()` calls still race on the same globals).
4. **Golden test fixtures** are git-tracked and stable (per the review); the change adds new fixtures, never modifies existing ones.

## Goals / Non-Goals

**Goals:**

- Land D-001 hotfix immediately.
- Add compile-time adapter contract checks in both adapters.
- Replace per-call adapter instantiation with package-level singletons.
- Use typed `permission.Allow` / `permission.Deny` constants.
- Migrate library tests to testify.
- Add coverage for 6 functions at 0%.
- Add golden + round-trip tests for adapters.
- Add `t.Parallel()` to fixture-driven tests (where safe).
- Widen the forbidigo pattern to catch all package-level mutable vars.
- Add `interface{}` → `any` migration in `getDocType`.
- Wrap `canaryOnce` in a `sync.Mutex` for race-safe `ResetCanaryForTest`.
- Always set `c.Cause` in `checkNameConflict`.

**Non-Goals:**

- Performance optimizations (deferred to `perf-hardening`).
- Changing the `Library` coverage target.
- Adding new test frameworks.
- Refactoring existing tests beyond the testify migration.

## Decisions

### 1. Adapter singleton: `var OpenCode = &Adapter{}` (not `sync.Once`)

**Choice**: Replace `func New() *Adapter { return &Adapter{} }` with `var OpenCode = &Adapter{}` at the package level. The `Adapter` struct is stateless (no fields), so a single instance is safe to share.

**Rationale**: The stateless adapter doesn't need per-call instantiation. A package-level singleton is the simplest, most efficient pattern. The `claudecode.New()` and `opencode.New()` callers (in templates, in `createTemplateFuncMap`, etc.) update to use the singleton.

**Why this deviates from the `golang-cli-architecture` Factory pattern**: The skill prescribes "lazy function fields on `*cmdutil.Factory`" for **cross-command shared dependencies** (config loading, library loading, API clients). The `permission.Adapter` is **not** a cross-command shared dependency — it is a stateless utility struct used by the platform adapter packages (`internal/opencode/`, `internal/claude-code/`), which themselves live in the Imperative Shell. The `cmdutil.Factory` pattern addresses *orchestration* concerns (what runs before/after a command); adapter singletons address *implementation* concerns (how a platform adapter is exposed within its package). They live at different layers. The choice is deliberate and is preserved by a comment in each adapter package (`// Package-level singleton: the adapter is stateless; no DI needed.`).

**Alternatives considered**:

- *`sync.Once` initialization*: rejected; the adapter is already stateless, no init needed.
- *Pass an `*Adapter` as a parameter*: rejected; this is a refactor of every call site, not just the constructor.
- *Add `f.OpenCodeAdapter` / `f.ClaudeCodeAdapter` lazy fields to `*cmdutil.Factory`*: rejected; these are package-internal implementations of the platform adapter pattern, not cross-command dependencies. Hoisting them onto the Factory would invert the dependency direction (`internal/cmdutil` would import `internal/opencode`/`internal/claude-code`).

### 2. Testify migration: per-test-file, no per-test rewrite

**Choice**: Each test file is migrated in one commit. The migration is mechanical: `t.Fatalf("...%v", err)` → `require.NoError(t, err)`; `if got != want { t.Errorf("got %v, want %v", got, want) }` → `assert.Equal(t, want, got)`. The test logic is preserved.

**Rationale**: A wholesale rewrite is unnecessary and risks breaking test logic. The migration is a 1:1 substitution of assertion styles.

**Alternatives considered**:

- *Rewrite all library tests in table-driven form*: rejected; the table-driven pattern is for new tests; existing tests keep their current shape with testify assertions.

### 3. Adapter contract: compile-time `var _` check

**Choice**: Add `var _ permission.Adapter = (*Adapter)(nil)` to `internal/opencode/opencode_adapter.go` and `internal/claude-code/claude_code_adapter.go`. The `internal/permission` package is imported by both adapters.

**Rationale**: Compile-time checks are zero-cost and catch signature drift immediately. The check is the canonical Go idiom for interface satisfaction.

**Alternatives considered**:

- *Runtime check in a `TestAdapterContract` test*: rejected; compile-time is faster and catches drift at `go build` time.

### 4. `t.Parallel()` in `methods_test.go` and `library_add_test.go`: subtest-level

**Choice**: Add `t.Parallel()` only to subtests (within `t.Run(...)`) that don't share `t.Setenv` or `os.Chdir`. The top-level `TestLibrary_X` calls do not get `t.Parallel()` because they may share state with sibling tests in the same package.

**Rationale**: Subtest-level parallelism is the safest: each subtest gets its own goroutine but the parent test is still sequential. The race detector reports no false positives.

**Alternatives considered**:

- *Per-test file parallelism*: rejected; tests in the same file may share helpers.
- *No parallelism*: rejected; the project has 100+ tests; parallelism saves ~5s per test run.

### 5. Round-trip test pattern: `ParsePlatformDocument → MarshalCanonical → re-parse → assert equal fields`

**Choice**: Add `TestParseRenderRoundTrip` in `internal/renderer/serializer_test.go` that:
- Reads a fixture (e.g., `test/fixtures/canonical/agent.md`).
- Parses it via `parser.ParsePlatformDocument(inputPath, platform, docType)`.
- Marshals the result via `renderer.MarshalCanonical(doc)`.
- Re-parses the marshaled output.
- Asserts that key fields (`Name`, `Description`, `Mode`, etc.) are equal between the two parses.

**Rationale**: Round-trip tests catch adapter drift and tag collisions in `CanonicalAgent`-style embedding (per the review's D-013). The pattern is the standard idiom for verifying serialization correctness.

**Why this is `SEMANTIC` equality, not byte equality** (resolves the proposal ambiguity): the proposal text says "byte-equivalence" but the design says "semantic equality". The implemented behavior is **semantic** equality — the assertion walks the parsed struct fields rather than comparing raw bytes. This matches the skill's testing guidance: golden files compare *captured output* (bytes for E2E, formatted text for Command), but round-trip tests compare *semantic state* (because YAML serialization is not guaranteed byte-stable across `MarshalCanonical` versions).

**Alternatives considered**:

- *Byte-equality assertion*: rejected; YAML serialization can include field order, comments, whitespace — semantic equality is more robust.
- *Snapshot test*: rejected; golden files are for end-to-end adapter output; round-trip is for canonical marshaling.

### 5a. Golden test build tag `//go:build golden` (deviation from skill)

**Choice**: Gate the new adapter golden tests (`opencode_adapter_golden_test.go`, `claude_code_adapter_golden_test.go`) AND the round-trip test under a `//go:build golden` build tag, runnable via `go test -tags=golden ./...`. The existing golden tests in `test/golden/` remain in the default suite.

**Why this deviates from the `golang-cli-architecture` skill**: The skill (`07-testing.md`) states "Golden files are a technique within the Command and E2E tiers (capture stdout/stderr, diff against `testdata/*.golden`, refresh with `-update`), **not a separate tier**." Gating tests behind a build tag creates an implicit fourth tier.

**Why the deviation is justified for these specific tests**:
1. The new **adapter golden tests** assert **byte-equality** against `test/golden/<platform>/agent-balanced.{yaml,json}`. Byte-equality golden tests are sensitive to:
   - YAML field order emitted by `MarshalCanonical`
   - JSON key order and whitespace from `RenderDocument`
   - Sprig template whitespace
   These are NOT controlled by the test author; they are emergent properties of the renderer chain. Putting byte-equality tests in the default suite produces a CI flake whenever any renderer dependency updates (Cobra, sprig, yaml.v3).
2. The new **round-trip test** asserts **semantic** equality, but runs against **all 4 canonical types** (Agent, Skill, Command, Memory) for **both platforms** (OpenCode, Claude Code). With fixtures and assertions across the full matrix, the round-trip test takes ~3s and is gated for the same CI-flake reasons.
3. The project's existing golden tests (`test/golden/*.golden`) compare **rendered platform files** (not canonical marshaling), and use `UPDATE_GOLDEN` flag (not `-update`) for refresh — they remain in the default suite because their input is parser output (controlled) rather than renderer output (uncontrolled).
4. The CI pipeline runs `mise run test` (default) and `mise run test:golden` (with the tag) as separate stages, so the gated tests are still exercised.

**Alternatives considered**:

- *Move all golden tests under `//go:build golden`*: rejected; the existing `test/golden/*.golden` tests have stable byte output (renderer output is deterministic in those paths) and don't need gating.
- *Remove the build tag entirely*: rejected; the new adapter byte-equality tests are too sensitive to renderer dependency drift for default-suite inclusion.
- *Use `t.Skipped` based on `os.Getenv("CI")` or a flag*: rejected; build tags are checked at `go vet` time, skip env-vars are runtime checks (fail CI when CI is unset).

### 6. Forbidigo pattern: widen to catch all package-level mutable vars

**Choice**: Update `.golangci.yml:87` to:

```yaml
forbidigo:
  patterns:
    - pattern: 'var (defaultAdder|outputFormat|initOutputFormat|outputFormatRefresh|completionShells|errEmptyResources)\b'
      msg: 'package-level mutable variables are forbidden per cmd/AGENTS.md:46'
```

**Rationale**: The current pattern only catches `var global(Factory|CommandConfig)`. The 6 mutable package-level vars identified in the review slip through. Widening the pattern catches all of them at `go vet` time.

**Alternatives considered**:

- *Add a linter that requires all `var` declarations to be `var`-blocks*: rejected; too broad, would flag legitimate constants.
- *Code review*: rejected; the pattern-based check is cheaper and more consistent.

## Risks / Trade-offs

- **Hotfix D-001 slows the test suite by 2-3 seconds** (the test no longer runs in parallel with siblings). **Mitigation**: the slowness is acceptable for race-detector correctness.
- **testify migration is mechanical but spans 7 test files.** A missed `t.Fatal` is hard to detect. **Mitigation**: task 4.1 runs `rg "t\.Fatal|t\.Error" internal/library/*_test.go` after the migration; zero matches expected.
- **Adapter singleton** is a behavior change for `New()` callers. **Mitigation**: the new `OpenCode` / `ClaudeCode` package-level vars are drop-in replacements; all callers are updated in the same commit.
- **Round-trip test** uses semantic equality (key fields) rather than byte equality. **Mitigation**: a follow-up "byte-exact" test can be added in `perf-hardening` if profiling shows a need.
- **Forbidigo pattern widening** may flag legitimate `var foo` declarations if a contributor names a var with one of the forbidden prefixes. **Mitigation**: the pattern is anchored to specific var names; generic `var foo` is not flagged.

## Migration Plan

The change ships in **one PR with 6 atomic phases** (each commit is independently testable):

1. **Phase 0 — Hotfix D-001** (lands immediately, outside OpenSpec): remove `t.Parallel()` from `cmd/lint_test.go:19`. Verify `go test -race -count=1 ./cmd/...` passes.
2. **Phase 1 — Adapter contract + singleton + typed constants + canary mutex** (tasks 5.1-5.7): add `var _ permission.Adapter = (*Adapter)(nil)` to both adapters; replace `New()` with package-level singletons; use typed `permission.Allow` / `permission.Deny` constants; add `any` + nil-check to `getDocType`; wrap `canaryOnce` in mutex; always set `c.Cause` in `checkNameConflict`. Verify `mise run check`.
3. **Phase 2 — Testify migration** (tasks 5.8-5.14): migrate 7 library test files to testify. Verify `mise run test` and `mise run test:race`.
4. **Phase 3 — Coverage gap fixes** (tasks 5.15-5.22): add `creator_test.go`, `discovery_test.go`, `methods_test.go` for `CreatePreset`, `resolver_test.go` for `GetOutputPaths`; add coverage to `claude-code` adapter; add `core.Valid` / `Unwrap` tests. Verify `mise run test:coverage` reaches ≥ 70% per package.
5. **Phase 4 — Golden files + round-trip** (tasks 5.23-5.26): add `opencode_adapter_golden_test.go`, `claude_code_adapter_golden_test.go`, `TestParseRenderRoundTrip`; fix `cmd/canonicalize_golden_test.go:26` CWD-skip. Verify `go test -tags=golden ./...` passes.
6. **Phase 5 — Test parallelization** (tasks 5.27-5.28): add `t.Parallel()` to fixture-driven subtests in `methods_test.go` and `library_add_test.go`. Verify `mise run test:race` passes.
7. **Phase 6 — Trivial folds** (tasks 5.29-5.31): widen forbidigo pattern; replace package-level vars in 5 cmd files; replace hard-coded file list in `TestNoNewForbidigoPatterns`. Verify `mise run lint` and `mise run check`.

**Rollback strategy**: revert each phase commit independently. Phase 0 is a 1-line edit (easy to revert). Phases 1-6 are additive or mechanical; revert restores the prior state.
