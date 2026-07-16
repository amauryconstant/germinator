## Context

The 2026-07-08 review identified 17 test-infra findings clustered in 4 areas. The codebase's lint baseline is clean (`mise run lint` reports 0 issues) and the test surface is large (17 E2E test files, 56+ golden files, 57 fixtures), but the infrastructure has 4 systemic issues:

1. **Race conditions in cmd tests** — `cmd/lint_test.go:19` uses `t.Parallel()` while calling `cmd.Execute()`. Cobra's package-level `OnInitialize` slice is mutated by every command registration; concurrent reads/writes trigger race detector failures. The race cascades: dozens of cmd tests that call `cmd.Execute()` race on Cobra globals.

2. **Coverage gaps in `internal/library`, `claude-code`, `renderer`, `core`** — multiple functions are at 0% coverage. The coverage rules at `config.testing` specify ≥ 70% per package, but `internal/library` is at 79.4% with 6 functions at 0%. The coverage report is below the project goal.

3. **Testify migration incomplete** — `cmd/` tests use testify (`require.NoError`, `assert.Equal`); library tests use raw `t.Fatalf` / `t.Errorf`. The mix is inconsistent and prevents test-pattern reuse.

4. **Adapter contract is unenforced** — `permission.Adapter` is declared in `internal/permission/adapter.go:8-13` but no `var _ permission.Adapter = (*Adapter)(nil)` check exists. A signature drift in either adapter is detected only at the call site, not at compile time.

### Constraints

1. **Hotfix D-001 lands first**, outside the OpenSpec change. The Blocker finding is a race-detector correctness issue; it must not wait for the larger change.
2. **Performance findings** (C-010..C-018, C-024, C-025) are deferred to a separate `perf-hardening` change per the reorg plan.
3. **Test files use `t.Parallel()` carefully** — only tests that don't share `t.Setenv` / `os.Chdir` / Cobra globals. The hotfix removes `t.Parallel()` from `cmd/lint_test.go` because the test forks `mise` processes *and* the per-test `cmd.Execute()` calls race on Cobra's package-level `OnInitialize` slice; both triggers compound the race.
4. **Golden test fixtures** are git-tracked and stable (per the review); the change adds new fixtures, never modifies existing ones.
5. **`errEmptyResources` migration is in this change**, not deferred to `enforce-error-discipline`. The widened forbidigo pattern requires the migration to land concurrently; the constructor call is small and fits Phase 6.

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
- The canary race-safety task (D-029): obsoleted by the `enforce-error-discipline` change which deleted `internal/warning/canary.go`. No race-prone `sync.Once` survives in `cmd/` or `internal/iostreams/`.

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

### 5. Round-trip test pattern: canonical-only round-trip in default suite, platform round-trip as a follow-up

**Choice**: Add **two** round-trip tests in `internal/renderer/serializer_test.go`:

1. **`TestParseRenderRoundTrip` (canonical round-trip)**:
   - Reads a canonical fixture (e.g., `test/fixtures/canonical/agent-permission-balanced.md`).
   - Parses it via `parser.ParseDocument(t.Context(), inputPath, "agent")`, returning `*parser.CanonicalAgent`.
   - Marshals via `renderer.MarshalCanonical(t.Context(), doc)`.
   - Writes the marshaled string to a `t.TempDir()` file and re-parses it via `parser.ParseDocument(t.Context(), tmpPath, "agent")`.
   - Asserts the full canonical `core.Agent` field set is equal between the two parses.

2. **`TestPlatformRoundTrip` (platform round-trip)**:
   - Reads a platform fixture (e.g., `test/fixtures/claude-code/agent-permission-balanced.md`).
   - Parses via `parser.ParsePlatformDocument(t.Context(), inputPath, "claude-code", "agent")`.
   - Marshals via `renderer.MarshalCanonical(t.Context(), doc)`.
   - Re-parses the canonical marshal output via `parser.ParseDocument`.
   - Asserts the same canonical fields.

Both tests live in the default suite (no build tag) because semantic equality is deterministic.

**Rationale**: Round-trip tests catch adapter drift and tag collisions in canonical struct embedding (per the review's D-013). The canonical test validates `MarshalCanonical` correctness; the platform test additionally exercises the forward adapter path (`ParsePlatformDocument`) on real fixtures and confirms that the platform→canonical conversion preserves all canonical fields.

**Why semantic equality, not byte equality**: Same as the byte-vs-semantic rationale in `golang-cli-architecture/references/07-testing.md` §Golden Files — YAML serialization is not guaranteed byte-stable across `MarshalCanonical` versions (Cobra, sprig, yaml.v3 dependency drift). The skill's testing guidance distinguishes "captured output" (byte-for-byte golden files in the E2E tier) from "semantic state" (round-trip parses in the default suite).

**Alternatives considered**:

- *Single canonical-only round-trip*: rejected; the platform test catches adapter-specific drift that the canonical-only path cannot see (a `FromCanonical` field omission is not visible without exercising the platform parser first).
- *Byte-equality assertion*: rejected; YAML serialization can include field order, comments, whitespace — semantic equality is more robust.
- *Snapshot test*: rejected; golden files are for end-to-end adapter output; round-trip is for canonical marshaling.

Round-trip tests live in default suite per Decision 5; E2E byte-equality tests live in E2E tier per Decision 6.

### 6. Byte-equality golden tests live in E2E tier (`//go:build e2e`) with `.md` fixtures

**Choice**: The new byte-equality adapter golden tests (`opencode_adapter_golden_test.go`, `claude_code_adapter_golden_test.go`) live in `test/e2e/` with `//go:build e2e`. They use the existing E2E infrastructure (`gexec.Build` or equivalent binary-build helper, golden-file comparison) and run via `mise run test:e2e`. Fixtures are stored at `test/e2e/testdata/<platform>/agent-balanced.md` — **both** fixtures use the `.md` extension because the renderer emits YAML frontmatter wrapped around a Markdown body for both Claude Code and OpenCode (`config/templates/<platform>/agent.tmpl`). The round-trip tests (`TestParseRenderRoundTrip`, `TestPlatformRoundTrip`) stay in the default suite at `internal/renderer/serializer_test.go` because they assert semantic equality (deterministic), not byte equality.

**Why this aligns with the skill**: The skill (`07-testing.md`) states "Golden files are a technique within the Command and E2E tiers, **not a separate tier**." Byte-equality golden tests exercise the full adapter chain end-to-end, so they belong in the E2E tier alongside the existing binary tests. Earlier considered placing them under a custom `//go:build golden` tag — rejected because that creates an implicit fourth tier, contradicting the skill.

Per `golang-spf13-cobra` Testing section: cobra accumulates flag state across `Execute()` calls. The new E2E tests must be **Ginkgo `Describe`/`It` blocks within `package e2e_test`**, NOT standalone `Test*` functions (cobra state would leak).

**Note on existing golden-tagged tests**: The pre-change golden tests at `internal/canonicalize/canonicalize_golden_test.go` (build tag `golden`) cover the *canonicalize* service's byte output and run via `mise run test:golden`. Those tests assert against parser output (stable) and are deliberately separate from the new adapter tests, which assert against renderer output (sensitive to renderer dependency drift). The two tiers coexist; the new adapter tests use the `e2e` build tag deliberately to keep byte-sensitive checks in a single, gated CI stage.

**Rationale for E2E placement over default suite**: Byte-equality golden tests are sensitive to dependency drift:
- YAML frontmatter field order emitted by `MarshalCanonical`
- Sprig template whitespace inside the body

These are emergent properties of the renderer chain (Cobra, sprig, yaml.v3), not controlled by the test author. Running them under `mise run test:e2e` (separate CI stage) prevents flakes in `mise run test` while still exercising the tests regularly.

**Alternatives considered**:

- *Custom `//go:build golden` tag for byte-equality tests*: rejected; creates an implicit fourth tier contrary to the skill and would conflate them with the canonicalize command's existing `golden`-tagged tests.
- *Remove the build tag entirely and put byte-equality tests in default suite*: rejected; the new adapter byte-equality tests are too sensitive to renderer dependency drift for default-suite inclusion.
- *Use `t.Skipped` based on `os.Getenv("CI")` or a flag*: rejected; build tags are checked at `go vet` time, skip env-vars are runtime checks (fail CI when CI is unset).

### 7. Forbidigo pattern: widen to catch all package-level mutable vars

**Choice**: Update `.golangci.yml:87` to:

```yaml
forbidigo:
  patterns:
    - pattern: 'var (defaultAdder|outputFormat|initOutputFormat|outputFormatRefresh|completionShells|errEmptyResources)\b'
      msg: 'package-level mutable variables are forbidden per cmd/AGENTS.md:46'
```

**Rationale**: The current pattern only catches `var global(Factory|CommandConfig)`. The 6 mutable package-level vars identified in the review slip through. Widening the pattern catches all of them at `go vet` time. The `errEmptyResources` migration (inlining the `core.NewUsageError(...)` construction inside `runCreatePreset`) lands in this change rather than waiting for the deferred `enforce-error-discipline` task 3.12, so the widened pattern applies from day one.

**Alternatives considered**:

- *Add a linter that requires all `var` declarations to be `var`-blocks*: rejected; too broad, would flag legitimate constants.
- *Code review*: rejected; the pattern-based check is cheaper and more consistent.

## Risks / Trade-offs

- **Hotfix D-001 slows the test suite by 2-3 seconds** (the test no longer runs in parallel with siblings). **Mitigation**: the slowness is acceptable for race-detector correctness.
- **testify migration is mechanical but spans 7 test files.** A missed `t.Fatal` is hard to detect. **Mitigation**: task 2.8 runs `rg "t\.Fatal|t\.Error" internal/library/*_test.go` after the migration; zero matches expected.
- **Adapter singleton** is an internal API change (the `New()` constructor is deleted; no external package imports the adapters). **Mitigation**: the new `OpenCode` / `ClaudeCode` package-level vars are drop-in replacements; all internal callers (`internal/parser`, `internal/renderer`, `internal/opencode/doc.go`) update in the same commit.
- **Round-trip tests** use semantic equality (key fields) rather than byte equality. **Mitigation**: a follow-up change can add byte-equality coverage if a specific drift is observed (the round-trip pattern already catches field-level drift, so byte equality is rarely needed).
- **Forbidigo pattern widening** may flag legitimate `var foo` declarations if a contributor names a var with one of the forbidden prefixes. **Mitigation**: the pattern is anchored to specific var names; generic `var foo` is not flagged.
- **`mise run test:race`** already covers the whole tree (`./...`, including `cmd/`). The task description in `.mise/config.toml` even references this change. The widening is already done; this change only validates it. **Mitigation**: Phase 8.4 runs the existing task; if any non-D-001 race is exposed, it is logged and triaged before merge.

## Migration Plan

The change ships in **one PR with 6 atomic phases** (each commit is independently testable):

1. **Phase 1 — Adapter contract + singleton + typed constants + canary mutex** (tasks 1.1-1.11): add `var _ permission.Adapter = (*Adapter)(nil)` to both adapters; replace `New()` with package-level singletons; use typed `permission.Allow` / `permission.Deny` constants; add `any` + nil-check to `getDocType`; replace `canaryOnce` with `sync.OnceFunc`; always set `c.Cause` in `checkNameConflict`. Verify `mise run check`.
2. **Phase 2 — Testify migration** (tasks 2.1-2.10): migrate 7 library test files to testify. Verify `mise run test` and `mise run test:race`.
3. **Phase 3 — Coverage gap fixes** (tasks 3.1-3.7): add `creator_test.go`, `discovery_test.go`, `methods_test.go` for `CreatePreset`, `resolver_test.go` for `GetOutputPaths`; add coverage to `claude-code` adapter; add `core.Valid` / `Unwrap` tests. Verify `mise run test:coverage` reaches ≥ 70% per package.
4. **Phase 4 — Round-trip + E2E golden tests** (tasks 4.1a-4.4a, 4.5, 4.6a-4.6b, 4.7, 4.8a, 4.9): byte-equality golden tests move to `test/e2e/` under `//go:build e2e` (per design Decision 6); round-trip test stays in default suite. Verify `mise run test:e2e` and `go test ./internal/renderer/...`.
5. **Phase 5 — Test parallelization** (tasks 5.1-5.4): add `t.Parallel()` to fixture-driven subtests in `methods_test.go` and `library_add_test.go`. Verify `mise run test:race` passes.
6. **Phase 6 — Trivial folds** (tasks 6.1-6.9): widen forbidigo pattern; replace package-level vars in 5 cmd files; replace hard-coded file list in `TestNoNewForbidigoPatterns`. Verify `mise run lint` and `mise run check`.

The D-001 hotfix (`cmd/lint_test.go:19` `t.Parallel()` removal) lands separately, before this OpenSpec change, and is documented in `proposal.md` Phase 0 only — it is not part of this change's atomic phase sequence.

**Rollback strategy**: revert each phase commit independently. Phases 1-6 are additive or mechanical; revert restores the prior state.
