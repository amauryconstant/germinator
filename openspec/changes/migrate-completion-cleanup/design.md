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

**Choice**: The package-level `var cache` in `cmd/completions.go` is replaced with a `Factory.CompletionCache` field of type `*Cache`, where `Cache` is a new type extracted within the same `cmd/completions.go` file (the file is kept in `cmd/` — see Decision 1b below). The cache is populated in `main.go` and exposes `Reset()` (for tests) and `Invalidate()` (for mutating commands) methods.

**Rationale**: package-level mutable state violates the `cli-factory` capability; per-Factory state makes tests trivially parallelizable (each test creates a new Factory with a fresh cache).

### 1b. Completion code stays in `cmd/` (no `internal/completion/` package)

**Choice**: `cmd/completions.go` is not extracted into a new `internal/completion/` package. The `Cache` type and the four action functions (`actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms`) remain in `cmd/completions.go`.

**Rationale**: per golang-cli-architecture's "extract when painful, not when predicted" trigger, extraction to a new package requires either 5+ types sharing a concern or a second consumer. Here we have one Cache type + 4 action functions used by exactly one consumer (`cmd/completion.go`). Creating `internal/completion/` would add an import path without proportional benefit. If a second consumer ever emerges, extraction is a 5-minute refactor at that point.

### 2. `Cache.Invalidate()` is explicit, called by mutating commands

**Choice**: The `*Cache` type exposes `Invalidate()` as a method on the Cache (not on Factory). All mutating library commands (`runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`) call `f.CompletionCache.Invalidate()` after a successful mutation. The TTL-based safety net remains.

**Rationale**: explicit invalidation makes the lifetime deterministic; the TTL (5 seconds, matching current behavior) catches any missed call. Putting the method on the `Cache` type (rather than `Factory.InvalidateCache()`) keeps `cmdutil.Factory` as a pure composition root — eager values + lazy `func() (T, error)` fields, no mutating methods — consistent with slices 1–8 and the `cli-factory` capability. If future caches are added (e.g., a schema cache), each cache type owns its own invalidation; Factory doesn't grow N `InvalidateX()` methods.

### 3. `internal/models/constants.go` content moves to `internal/core/platform.go`

**Choice**: The two string constants `PlatformClaudeCode = "claude-code"` and `PlatformOpenCode = "opencode"` move from `internal/models/constants.go` (7 lines total) to `internal/core/platform.go`. The `PermissionPolicy` enum and `PlatformConfig` type already live in `internal/core/platform.go` from slice 1; no other content needs to move. The depguard rule `.golangci.yml` already applies to `**/core/**` (allow stdlib + `samber/lo`), so no rule change is required.

**Rationale**: matches the project layout target; the two constants are stdlib `string` types, so depguard still applies unchanged. A separate `internal/parser/loader.go` defines the same two constants independently — that duplicate is removed in this change (loader imports from `internal/core`).

### 3b. Version source: `internal/version` package, not `Factory.AppVersion`

**Choice**: `runVersion(opts)` reads `version.Version`, `version.Commit`, `version.Date` directly from the `internal/version` package. `Factory.AppVersion` remains a short-form `string` used elsewhere (e.g., potential `--help` banner, future upgrade notifications) but is NOT the source for the `version` subcommand.

**Rationale**: the `internal/version` package is the build-time injection point for `-ldflags` (`.mise/config.toml:15`, `.goreleaser.yml:28-30`) and holds all three pieces (Version, Commit, Date). `Factory.AppVersion` is a single `string` (per slice-1 `cli-cli-factory` spec) and cannot carry the 3-piece metadata without expanding the Factory shape. Keeping the `version` subcommand's source as `internal/version` preserves the existing contract documented in `cli-framework` ("Version Command shows full info") and `testing-e2e-testing` ("Version Command E2E Tests"), and matches the current behavior noted in `cmd/cmd_test.go:128` ("not influenced by the Factory's AppVersion field").

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

- **Completion cache invalidation** — if any mutating command forgets `f.CompletionCache.Invalidate()`, stale completions persist for up to 5 seconds. **Mitigation:** explicit test in task 9.1.11; the orchestrator's PHASE2 verification runs `germinator library add` then `germinator library show <TAB>` to assert the new resource appears.
- **`internal/models/` deletion** — any reference to `gitlab.com/amoconst/germinator/internal/models` is a build error. **Mitigation:** `rg "internal/models" .` finds all references (4 files: `cmd/completions.go`, `internal/config/config.go`, `internal/config/config_test.go`, `internal/config/manager_test.go`); updated in this change. A fifth duplicate lives in `internal/parser/loader.go` (defines the constants independently) and is also cleaned up.
- **`Factory.AppVersion` vs `internal/version` divergence** — see Decision 3b. The slice-1 `cli-cli-factory` spec says `Factory.AppVersion` is "set to the build-time version", but the `version` subcommand reads from `internal/version`. **Mitigation:** Decision 3b documents this as an intentional split: `AppVersion` stays a short-form string for `--help`/future use; `internal/version` is the source for the detailed `version` subcommand. No spec change is required (the `cli-factory` scenario is about the field being populated, which it is).
- **AGENTS.md updates are many** — root + per-package + per-command. **Mitigation:** tasks are sequenced; each AGENTS.md is reviewed for accuracy before commit.
- **CHANGELOG format** — must match the project's convention (`Keep a Changelog` or similar). **Mitigation:** `osx-generate-changelog` handles formatting; the user reviews the generated entry.
- **No major version bump** — the BREAKING changes are documented but the version is not bumped to 1.0. **Mitigation:** this is a deliberate decision; the version bump is a separate change that aggregates ALL breaking changes from this migration plus any others.
