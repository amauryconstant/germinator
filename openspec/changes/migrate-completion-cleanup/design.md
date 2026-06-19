# Design — Migrate completion, version, and finalize migration

## Context

This is the final change in the migration sequence. After change-8 (`migrate-config-commands`), only three things remain: the `completion` and `version` shell commands, the residual `internal/models/` package, and the documentation. This change migrates the last two commands, deletes the residual package, updates all `AGENTS.md` files, and generates the CHANGELOG entry.

## Goals / Non-Goals

**Goals:**

- `cmd/completion.go`, `cmd/completions.go`, and `cmd/version.go` follow the new pattern.
- The completion cache lives on `Factory.CompletionCache` (per-Factory, testable, with `Reset()` method).
- `internal/models/` is deleted; its constants move to `internal/core/platform.go`.
- All `AGENTS.md` files reflect the new architecture.
- CHANGELOG entry documents the BREAKING changes from the entire migration.
- E2E test sweep updates old exit codes and flag names.
- The orchestrator's pre-flight + verification confirms the migration is complete.

**Non-Goals:**

- Restructuring any package internals (already deferred to follow-up changes).
- Adding `huh` for interactive prompts (no current consumer).
- Restructuring the `library` package (deferred to `refactor-library-package` follow-up).
- Bumping the major version (1.0) — deferred to a separate change that aggregates ALL the BREAKING changes from this migration.

## Decisions

### 1. Completion cache moves from package-level to Factory

**Choice**: The package-level `var cache` in `cmd/completions.go` is replaced with a `Factory.CompletionCache` field of type `*completion.Cache`. The cache is populated in `main.go` and has a `Reset()` method for tests.

**Rationale**: package-level mutable state violates the `cli-factory` capability; per-Factory state makes tests trivially parallelizable (each test creates a new Factory with a fresh cache).

### 2. `Factory.InvalidateCache()` is explicit, called by mutating commands

**Choice**: All mutating library commands (`runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`) call `f.InvalidateCache()` after a successful mutation. The TTL-based safety net remains.

**Rationale**: explicit invalidation makes the lifetime deterministic; the TTL (5 seconds, matching current behavior) catches any missed call.

### 3. `internal/models/constants.go` moves to `internal/core/platform.go`

**Choice**: The constants `PlatformClaudeCode`, `PlatformOpenCode`, document-type constants, and permission-mode enums move from `internal/models/constants.go` to `internal/core/platform.go`. The depguard rule for `internal/core/**` is updated to allow this file (still stdlib only).

**Rationale**: matches the project layout target; constants are stdlib types (`string`); depguard still applies.

### 4. CHANGELOG entry aggregates all BREAKING changes

**Choice**: A single CHANGELOG entry (under a new "BREAKING" heading) documents:
- Exit codes 3–6 collapsed to 1 (from change-2)
- `--json` flag replaced by `--output json` on library commands (from changes 2, 4, 6, 7)
- `--output` flag on config commands renamed to `--output-path` (from change-8)
- `internal/service/`, `internal/application/`, `internal/models/` directories deleted (from changes 7, 9)

**Rationale**: a single CHANGELOG entry gives consumers one place to learn about the breaking changes.

### 5. E2E test sweep is mechanical

**Choice**: All E2E tests using old exit codes (3-6) and old flag names (`--json`, `--output` on config) are updated in this change. The patterns are well-defined enough for `rg` to find them all.

**Rationale**: bulk update is faster than per-file review; the orchestrator's verification (PHASE2) catches any missed update.

## Risks / Trade-offs

- **Completion cache invalidation** — if any mutating command forgets `f.InvalidateCache()`, stale completions persist for up to 5 seconds. **Mitigation:** explicit test in task 9.1.5; the orchestrator's PHASE2 verification runs `germinator library add` then `germinator library show <TAB>` to assert the new resource appears.
- **`internal/models/` deletion** — any reference to `gitlab.com/amoconst/germinator/internal/models` is a build error. **Mitigation:** `rg "internal/models" .` finds all references; updated in this change.
- **AGENTS.md updates are many** — root + per-package + per-command. **Mitigation:** tasks are sequenced; each AGENTS.md is reviewed for accuracy before commit.
- **CHANGELOG format** — must match the project's convention (`Keep a Changelog` or similar). **Mitigation:** `osx-generate-changelog` handles formatting; the user reviews the generated entry.
- **No major version bump** — the BREAKING changes are documented but the version is not bumped to 1.0. **Mitigation:** this is a deliberate decision; the version bump is a separate change that aggregates ALL breaking changes from this migration plus any others.
