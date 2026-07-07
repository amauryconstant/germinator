# Migrate completion, version, and finalize migration

## Why

This is the **final change** in the migration sequence. It migrates the last two shell commands (`completion` for carapace shell completion; `version` for build metadata), deletes the residual `internal/models/` package, updates all `AGENTS.md` files to reflect the new architecture, and generates the CHANGELOG entry documenting the BREAKING changes. After this change archives, the migration to `golang-cli-architecture` is complete.

## What Changes

### Migrate completion (carapace)

- **MIGRATE** `cmd/completion.go` and `cmd/completions.go`:
  - **KEEP** `cmd/completions.go` in `cmd/` package (single consumer: `cmd/completion.go`; per golang-cli-architecture "extract when painful, not predicted"). The Cache TYPE itself is hoisted to `internal/cmdutil/completion_cache.go` so it sits next to `Factory` (see design Decision 1b)
  - **ADD** `Factory.CompletionCache *cmdutil.CompletionCache` field populated in `main.go` so tests can reset it
  - **REPLACE** package-level `var cache` with the `Factory.CompletionCache` field; expose `Reset()` and `Invalidate()` methods on the `CompletionCache` type
  - **CONVERT** `actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms` to take the Factory as input and use the Factory's library loader (with timeout) and the Factory's cache
  - **MIGRATE** `cmd/completion.go` to `NewCmdCompletion(f, runF) + runCompletion(opts)`:
    - `completionOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `Shell string` (Ctx added for symmetry with `versionOptions` and the golang-context skill's "all I/O accepts ctx" rule)
    - No new behavior; preserves carapace integration

### Migrate `version`

- **MIGRATE** `cmd/version.go`:
  - Declare `versionOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`
  - Implement `NewCmdVersion(f *cmdutil.Factory, runF func(*versionOptions) error) *cobra.Command`
  - Implement `runVersion(opts *versionOptions) error`: write `germinator <Version> (<Commit>) <Date>\n` to `opts.IO.Out`, reading from the `internal/version` package (injected via `-ldflags` at build time). `Factory.AppVersion` is NOT the source — it remains a short-form string used elsewhere; the `version` subcommand is the authoritative detailed view.
  - The output format contract is already specified by `cli-framework` ("Version Command shows full info") and `testing-e2e-testing` ("Version Command E2E Tests"); this change adds no new spec.
  - Move `TestVersionCommand` from `cmd/cmd_test.go` into a dedicated `cmd/version_test.go` with table-driven coverage

### Delete `internal/models/`

- **MOVE** the two string constants `PlatformClaudeCode` and `PlatformOpenCode` from `internal/models/constants.go` to `internal/core/rules.go` (alongside `ValidatePlatform`, the consumer of these constants; the `PermissionPolicy` enum and `PlatformConfig` type live in `internal/core/platform.go` from slice 1; nothing else needs to move)
- **DELETE** `internal/models/` directory
- **VERIFY** the depguard rule `.golangci.yml` (applies to `**/core/**`, stdlib only) still passes after the move — no rule change expected
- **UPDATE** all consumers (see task 9.3.3) including `internal/parser/loader.go`, which defines the same constants independently

### Update documentation

- **UPDATE** root `AGENTS.md` architecture diagram
- **UPDATE** `cmd/AGENTS.md` with the canonical `adapt` example
- **UPDATE** `internal/AGENTS.md` to reflect rename to `internal/core/` and new sibling packages
- **VERIFY and UPDATE** `internal/{iostreams,output,cmdutil}/AGENTS.md` (these files already exist from earlier slices; the work here is review/polish, not creation)
- **UPDATE** `internal/library/AGENTS.md`, `internal/parser/AGENTS.md`, etc. for packages that moved (note: `cmd/library/` does not exist — the project uses a flat `cmd/` layout with sibling files like `library.go`, `library_add.go`, etc.; per-subcommand docs under `cmd/` are not yet a project convention)

### Generate CHANGELOG

- **ADD** `CHANGELOG.md` entry (via `osx-generate-changelog`) with the two BREAKING CLI changes:
  - Exit codes 3–6 collapsed to 1
  - `--json` flag renamed to `--output json` on library commands
  - `--output` flag renamed to `--output-path` on config commands

### Final sweep

- **UPDATE** `test/e2e/` for old exit codes (3-6) → 1
- **UPDATE** `test/e2e/` for old `--json` flag → `--output json`
- **UPDATE** `test/e2e/` for old `--output` flag on config commands → `--output-path`

## Capabilities

### Modified

- **`shell-completion`** (delta) — completion cache moves to `Factory.CompletionCache`; explicit `f.CompletionCache.Invalidate()` is called by mutating commands

## Out of scope (none — this is the final change)

## Impact

### Affected code

- **Migrated (2 files):** `cmd/completion.go`, `cmd/version.go`
- **Refactored (2 files):** `cmd/completions.go` (kept in `cmd/`; consumes `Factory.CompletionCache` via the new `*cmdutil.CompletionCache` field) and `internal/cmdutil/completion_cache.go` (new file, hosts the `CompletionCache` type with `Get`/`Set`/`Reset`/`Invalidate`)
- **Added (1 file):** `cmd/version_test.go` (moves `TestVersionCommand` out of `cmd/cmd_test.go` into a dedicated file with table-driven coverage)
- **Modified (1 file):** `main.go` (populate `Factory.CompletionCache` field; mutating commands call `f.CompletionCache.Invalidate()`)
- **Modified (1 file):** `internal/core/rules.go` (the two `Platform*` constants move in from `internal/models/constants.go`)
- **Modified (1 file):** `internal/parser/loader.go` (drop its duplicate `PlatformClaudeCode`/`PlatformOpenCode` definitions; import from `internal/core`)
- **Modified (4 files):** `internal/config/config.go`, `internal/config/config_test.go`, `internal/config/manager_test.go`, `cmd/completions.go` (update `models.Platform*` references to `core.Platform*`)
- **Deleted (1 directory):** `internal/models/`
- **Modified (1 file):** `cmd/completion_test.go` (converted to new pattern)
- **Modified (1 file):** `cmd/cmd_test.go` (`TestVersionCommand` moved out; remaining tests unchanged)
- **Modified (N files):** all `AGENTS.md` files (review/polish existing files)
- **Modified (1 file):** `cmd/testdata/lint_baseline.txt` (refreshed after the `var cache` removal and `internal/models/` deletion — see new task 9.4.9)
- **Added (1 file):** `CHANGELOG.md` (BREAKING entry)
- **Modified (multiple files):** `test/e2e/` exit codes and flag renames

### Affected systems

- **CLI behavior:** completion and version commands now follow the new pattern (no externally observable behavior change)
- **Shell completion cache:** invalidated by mutating library commands (improvement; previously relied on TTL only)
- **BREAKING changes:** CHANGELOG entry documents the cumulative breaking changes from changes 2, 6, 7, 8

## Risks

- **Completion cache invalidation** — if `f.CompletionCache.Invalidate()` is missed in any mutating command, stale completions persist until TTL. **Mitigation:** explicit test in task 9.1.11; the orchestrator's verification will catch this.
- **`internal/models/constants.go` may have external consumers** — the constants are used across the codebase. **Mitigation:** `rg "PlatformClaudeCode|PlatformOpenCode" .` finds every consumer; update imports in the same change.
- **AGENTS.md updates are tedious** — many files to update. **Mitigation:** tasks 9.6.x are mechanical; review is per-file.
- **E2E test sweep** — many test files to update. **Mitigation:** `rg "ShouldFailWithExit\\([3-6]\\)" test/e2e/` finds them all; bulk update with `sed` (or hand-edited if patterns vary).
- **CHANGELOG generation** — `osx-generate-changelog` reads archived change proposals; if any earlier change wasn't archived properly, the CHANGELOG may miss entries. **Mitigation:** task 9.5.1 verifies all 9 changes are archived before generating the CHANGELOG.
