# Migrate library add and library create

## Why

The mutating library commands (`library add` and `library create`) have more complex behavior than the read-only commands: `library add` supports three modes (explicit files, `--discover` scan, `--discover --batch --force` continuous), and `library create` builds a preset from a list of refs. Migrating them after `init` (change-5) lets us reuse the `core.PartialSuccessError` pattern and adds `core.CanInstallResource` (a pure rule function in `internal/core/rules.go`).

## What Changes

### Add `core.CanInstallResource` to `internal/core/rules.go`

- **ADD** `core.CanInstallResource(ref string) error` to `internal/core/rules.go`:
  - Parses `ref` using `strings.Cut(ref, "/")` (Go 1.18+)
  - Validates that `type` is one of `skill`, `agent`, `command`, `memory` (using `slices.Contains`)
  - Validates that `name` is a non-empty identifier
  - Returns `*core.ValidationError` on failure
  - **String-only** (does NOT import `internal/library/` — depguard enforces this)

### Migrate `cmd/library/add.go`

- **MIGRATE** `cmd/library/add.go`:
  - Declare `addOptions`: `IO`, `Library`, `InputPaths []string`, `Name string`, `Description string`, `Type string`, `Platform string`, `Discover bool`, `Batch bool`, `Force bool`, `DryRun bool`, `Output string`
  - Declare the `Library` interface with methods called by all three modes
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` (legacy `--json` replaced by `--output json`)
  - Implement `NewCmdAdd(f, runF)` and `runAdd(opts)`:
    - Mode 1 (explicit files): validate each `InputPath` with `core.CanInstallResource(name)`, add to library
    - Mode 2 (`--discover`): scan directories for orphan files; for each, validate ref; collect successes/failures
    - Mode 3 (`--discover --batch --force`): continuous processing; on per-file failure, skip and continue
  - On partial success (some added, some failed): return `*core.PartialSuccessError` (exit 0 via `cmdutil.ExitCodeFor`)
  - On context cancellation during batch processing: return wrapped `ctx.Err()` after collecting partial results
  - **Thread `opts.Ctx` into every call** to `library.DiscoverOrphans`, `library.BatchAddResources`, `library.LoadLibrary`

### Migrate `cmd/library/create.go`

- **MIGRATE** `cmd/library/create.go`:
  - Declare `createPresetOptions`: `IO`, `Library`, `Resources []string`, `Description string`, `Force bool`
  - Implement `NewCmdCreatePreset(f, runF)` and `runCreatePreset(opts)`
  - **No `--output` flag** (legacy implementation didn't support `--json`; matches the `output-formats` capability's "only commands that previously had `--json` get `--output`")

### Delete legacy files

- **DELETE** `internal/service/adder.go` if present
- **DELETE** `internal/service/creator.go` if present

## Capabilities

### Modified

- **`library/library-json-output`** — `--output` flag is now available on `library add` (the legacy `--json` flag, if any, is replaced). `library create` does NOT get `--output` (it didn't previously have `--json`).

### Added (extends `internal/core/rules.go` from change-1)

- `core.CanInstallResource(ref string) error` — string-only ref validation

## Out of scope (deferred)

- Migrating remaining library commands (`library init`, `library refresh`, `library remove`, `library validate`) — change-7
- Migrating config / completion / version — changes 8, 9

## Impact

### Affected code

- **Modified (1 file):** `internal/core/rules.go` (add `CanInstallResource`)
- **Modified (1 file):** `internal/core/rules_test.go` (add test cases for `CanInstallResource`)
- **Rewritten (1 file):** `cmd/library/add.go`
- **Rewritten (1 file):** `cmd/library/create.go`
- **Modified (1 file):** `internal/library/` (add `ctx` parameter to `DiscoverOrphans`, `BatchAddResources`, `LoadLibrary`)
- **Modified (1 file):** `internal/library/discovery.go` (or similar) to support cancellation
- **Modified (1 file):** `internal/library/adder.go` (or similar) to support cancellation
- **Deleted (1-2 files):** `internal/service/adder.go`, `internal/service/creator.go` if present
- **Modified (1 file):** `cmd/library/add_test.go` (converted to `iostreams.Test()` + `runF` injection)
- **Modified (1 file):** `cmd/library/create_test.go` (converted similarly)

### Affected systems

- **CLI behavior:** `--output` flag added to `library add`; `library add` and `library create` output format may shift slightly (per-resource status lines instead of flat list, when partial success)

## Risks

- **Three modes in `library add`** — explicit, discover, batch — increase the migration surface. **Mitigation:** each mode is tested independently in tasks 6.1.5; batch mode's continuous behavior is the most complex and gets explicit attention.
- **Context cancellation handling in batch mode** — when the user hits Ctrl-C during a long batch, partial successes must be reported. **Mitigation:** task 6.1.4 explicitly handles this; `*core.PartialSuccessError` collects partial successes; `cmdutil.ExitCodeFor` returns 0 if any succeeded.
- **`CanInstallResource` must be string-only** — it can't import `internal/library/` (depguard). **Mitigation:** the function parses and validates the ref syntactically; the actual library lookup happens in `runAdd` after validation passes.
- **`internal/library/` API changes** — adding `ctx` to `DiscoverOrphans`, `BatchAddResources`, `LoadLibrary` is a breaking change to a package that may have other callers. **Mitigation:** the only callers in this codebase are `cmd/library_add.go` (now `cmd/library/add.go`); `mise run check` catches any missed caller.
