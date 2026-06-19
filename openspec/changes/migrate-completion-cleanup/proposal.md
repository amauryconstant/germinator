# Migrate completion, version, and finalize migration

## Why

This is the **final change** in the migration sequence. It migrates the last two shell commands (`completion` for carapace shell completion; `version` for build metadata), deletes the residual `internal/models/` package, updates all `AGENTS.md` files to reflect the new architecture, and generates the CHANGELOG entry documenting the BREAKING changes. After this change archives, the migration to `golang-cli-architecture` is complete.

## What Changes

### Migrate completion (carapace)

- **MIGRATE** `cmd/completion.go` and `cmd/completions.go`:
  - **MOVE** `cmd/completions.go` to `internal/completion/` package (or keep in `cmd/`); the cache lives on a `Factory.CompletionCache` field
  - **ADD** `Factory.CompletionCache` field (a `*completion.Cache` instance) populated in `main.go` so tests can reset it
  - **REPLACE** package-level `var cache` with the struct field; expose `Reset()` for tests
  - **ADD** `Factory.InvalidateCache()` method called by all mutating library commands after a successful mutation
  - **CONVERT** `actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms` to take the Factory as input and use the Factory's library loader (with timeout) and the Factory's cache
  - **MIGRATE** `cmd/completion.go` to `NewCmdCompletion(f, runF) + runCompletion(opts)`:
    - `completionOptions`: `IO *iostreams.IOStreams`, `Shell string`
    - No new behavior; preserves carapace integration

### Migrate `version`

- **MIGRATE** `cmd/version.go`:
  - Declare `versionOptions`: `IO *iostreams.IOStreams`
  - Implement `NewCmdVersion(f *runF func(*versionOptions) error) *cobra.Command`
  - Implement `runVersion(opts *versionOptions) error`: print version to `opts.IO.Out`

### Delete `internal/models/`

- **MOVE** `internal/models/constants.go` content to `internal/core/platform.go` (constants `PlatformClaudeCode`, `PlatformOpenCode`, document-type constants, permission-mode enums)
- **UPDATE** depguard rule for `internal/core/**` to allow `platform.go` (still stdlib only)
- **DELETE** `internal/models/` directory

### Update documentation

- **UPDATE** root `AGENTS.md` architecture diagram
- **UPDATE** `cmd/AGENTS.md` with the canonical `adapt` example
- **UPDATE** `internal/AGENTS.md` to reflect rename to `internal/core/` and new sibling packages
- **ADD** `internal/{iostreams,output,cmdutil}/AGENTS.md`
- **UPDATE** `cmd/library/AGENTS.md` (if it exists; create if not)
- **UPDATE** `internal/library/AGENTS.md`, `internal/parser/AGENTS.md`, etc. for packages that moved

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

- **`shell-completion`** (delta) — completion cache moves to `Factory.CompletionCache`; explicit `Factory.InvalidateCache()` is called by mutating commands
- **`config-commands`** — `--output` → `--output-path` rename (BREAKING) is documented in the CHANGELOG

## Out of scope (none — this is the final change)

## Impact

### Affected code

- **Migrated (2 files):** `cmd/completion.go`, `cmd/version.go`
- **Migrated (1 file):** `cmd/completions.go` → `internal/completion/completion.go` (or stays in `cmd/`)
- **Modified (1 file):** `main.go` (add `Factory.CompletionCache` field, call `f.InvalidateCache()` from mutating commands)
- **Modified (1 file):** `internal/core/platform.go` (constants moved from `internal/models/constants.go`)
- **Deleted (1 directory):** `internal/models/`
- **Modified (1 file):** `cmd/completion_test.go` (converted to new pattern)
- **Modified (1 file):** `cmd/cmd_test.go` (version test converted)
- **Modified (N files):** all `AGENTS.md` files
- **Modified (1 file):** `CHANGELOG.md` (BREAKING entry)
- **Modified (multiple files):** `test/e2e/` exit codes and flag renames

### Affected systems

- **CLI behavior:** completion and version commands now follow the new pattern (no externally observable behavior change)
- **Shell completion cache:** invalidated by mutating library commands (improvement; previously relied on TTL only)
- **BREAKING changes:** CHANGELOG entry documents the cumulative breaking changes from changes 2, 6, 7, 8

## Risks

- **Completion cache invalidation** — if `Factory.InvalidateCache()` is missed in any mutating command, stale completions persist until TTL. **Mitigation:** explicit test in task 9.1.5; the orchestrator's verification will catch this.
- **`internal/models/constants.go` may have external consumers** — the constants are used across the codebase. **Mitigation:** `rg "PlatformClaudeCode|PlatformOpenCode" .` finds every consumer; update imports in the same change.
- **AGENTS.md updates are tedious** — many files to update. **Mitigation:** tasks 9.6.x are mechanical; review is per-file.
- **E2E test sweep** — many test files to update. **Mitigation:** `rg "ShouldFailWithExit\\([3-6]\\)" test/e2e/` finds them all; bulk update with `sed` (or hand-edited if patterns vary).
- **CHANGELOG generation** — `osx-generate-changelog` reads archived change proposals; if any earlier change wasn't archived properly, the CHANGELOG may miss entries. **Mitigation:** task 9.5.1 verifies all 9 changes are archived before generating the CHANGELOG.
