# Migrate init command with partial-success semantics

## Why

The `init` command has unique semantics among the germinator commands: it processes N resources (a preset expansion or explicit refs), and partial success is "success" (exit 0) if at least one resource succeeded. The legacy `init` returns `[]InitializeResult` and exits 0 if any succeeded. The new pattern uses `error` returns, so partial success needs a sentinel error type (`core.PartialSuccessError`, defined in change-1) recognized by `cmdutil.ExitCodeFor` (returns 0 for `Succeeded > 0`).

## What Changes

### Wire `core.PartialSuccessError` into the error pipeline

- **UPDATE** `cmdutil.ExitCodeFor` (already in change-1) to:
  - Return `ExitCodeSuccess` (0) when `err` is `*core.PartialSuccessError` and `Succeeded > 0`
  - Return `ExitCodeError` (1) when `err` is `*core.PartialSuccessError` and `Succeeded == 0`
- **UPDATE** `output.FormatError` (already in change-1) to format `*core.PartialSuccessError` as `partial success: N succeeded, M failed` followed by per-resource error lines
- **ADD** tests for partial-success formatting and exit-code mapping (already in change-1's `cmdutil.ExitCodeFor` tests; this change adds integration tests for the full flow)

### Migrate `cmd/init.go`

- **MIGRATE** `cmd/init.go`:
  - Declare `initOptions`: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Initializer func() (Initializer, error)`, `Ctx context.Context`, `LibraryPath string`, `Platform string`, `OutputDir string`, `Refs []string`, `Preset string`, `DryRun bool`, `Force bool`
  - Declare the `Initializer` interface in the same file (one method: `Initialize(ctx, *InitializeRequest) ([]core.InitializeResult, error)`)
  - Implement `NewCmdInit(f *cmdutil.Factory, runF func(*initOptions) error) *cobra.Command`
  - Implement `runInit(opts *initOptions) error`:
    - If `opts.Preset` is set, resolve the preset and expand it to a list of refs
    - Call `f.Initializer().Initialize(opts.Ctx, req)` where `req` contains the refs and platform
    - Collect per-resource results
    - If at least one succeeded AND at least one failed: return `core.NewPartialSuccessError(succeeded, failed, errors)`
    - If all succeeded: return nil (exit 0)
    - If all failed: return the first error (or a `*core.PartialSuccessError` with `Succeeded == 0`, which maps to exit 1)

### Update tests

- **CONVERT** `cmd/init_test.go` (or `internal/service/initializer_test.go`) to `iostreams.Test()` + `runF` injection
- **ADD** explicit test cases for:
  - All-success: `germinator init --platform opencode --resources skill/commit` → exit 0
  - Partial-success: `--resources skill/commit,skill/invalid` → exit 0 with formatted "partial success: 1 succeeded, 1 failed"
  - All-failed: `--resources skill/invalid1,skill/invalid2` → exit 1

## Capabilities

### Modified

- **`cli/init-command`** — The `init` command adopts the `command-options-pattern` shape (`NewCmdInit(f, runF)` + `initOptions` + `runInit`). Per-resource errors are rendered via `output.FormatError`. Partial-success is signalled by returning `*core.PartialSuccessError`.
- **`library/partial-initialization`** — The `Initializer.Initialize` contract changes from "nil error on partial success, non-nil on full failure" to "always return `*core.PartialSuccessError` when at least one resource was processed" so the caller can distinguish partial vs full failure by inspecting `Succeeded == 0`.

## Out of scope (deferred)

- Migrating `library add`, `library create` — change-6
- Migrating remaining library commands — change-7
- Migrating config / completion / version — changes 8, 9

## Impact

### Affected code

- **Rewritten (1 file):** `cmd/init.go`
- **Modified (1 file):** `cmd/init_test.go` (or moved from `internal/service/initializer_test.go`)
- **Modified (1 file):** `internal/service/initializer.go` (the `Initialize` method's signature and return contract; the file itself is deleted in change-7)
- **No new dependencies** — uses `core.PartialSuccessError` from change-1

### Affected systems

- **CLI behavior:** `init` output format changes slightly (per-resource status lines instead of a flat list); exit code semantics preserved (0 if any succeeded, 1 if all failed)

## Risks

- **Partial-success exit code preservation is critical** — consumers rely on exit 0 when at least one resource installs. **Mitigation:** task 5.3 explicitly tests this case; the canary mechanism from change-2 (exit code deprecation warning) doesn't apply here.
- **Preset expansion complexity** — resolving a preset name to a list of refs and then processing each ref may fail at the resolution step or at the per-ref step. **Mitigation:** the resolution step uses `core.NewPartialSuccessError` if any refs in the preset fail to resolve; per-ref processing uses the same error type.
- **`internal/service/initializer.go` is partially migrated** — the file is modified to return `[]core.InitializeResult` properly but isn't deleted (used by `library add` in change-6 and `library init` in change-7). **Mitigation:** the file stays alive until change-7.
