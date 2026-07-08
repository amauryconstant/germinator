## Why

The 2026-07-08 code review identified **10 error-handling findings** (B-001..B-014, B-018) that violate the project's documented error-handling contract. The single-handling rule — "errors are either logged OR returned, NEVER both" — is broken in 7 production sites; the typed-error model is inconsistently applied (`FileError` used where `NotFoundError` is semantically correct; `ConfigError` used where `NotFoundError` is correct); the `FormatError` dispatch set is missing `InitializeError`; the `NotFoundError` exit code mapping is semantically wrong (mapped to `ExitCodeUsage` (2) instead of `ExitCodeError` (1)); the `internal/output/exporter.go:172` `var _ = io.EOF` dead-code suppression anchors an unused `io` import; the `errEmptyResources` constructor encodes a Cobra-substring into an `errors.New`; and the `RemoveResource` call path silently swallows `os.IsNotExist` on physical file removal.

This change enforces the error-handling contract end-to-end. It is a **production-code refactor** with spec deltas because several findings require semantic changes (exit code mapping, new typed error, dispatch set expansion).

## What Changes

- **MODIFY** `internal/cmdutil/exit.go:73` — map `NotFoundError` to `ExitCodeError` (1) instead of `ExitCodeUsage` (2). Update `internal/cmdutil/exit_test.go:58` to match.
- **MODIFY** `internal/cmdutil/exit.go:24` — replace brittle Cobra string-prefix matching with `errors.As` dispatch against `*cobra.FlagError` / `*flag.Error` (or document the limitation explicitly if typed errors are not available).
- **DELETE** 6 inline `output.FormatError(opts.IO, err)` calls in `cmd/library_remove.go:170,178`, `cmd/library_add.go:343,537,681,694`, `cmd/init.go:205/209`. Rely on `main.go`'s central `output.FormatError` handler.
- **MODIFY** `internal/library/remover.go:83` — replace `gerrors.NewFileError("access", "resource not found", nil)` with `gerrors.NewNotFoundError("library ref", opts.Ref)`.
- **MODIFY** `internal/library/resolver.go:67` — replace `gerrors.NewConfigError("preset", name, "preset not found")` with `gerrors.NewNotFoundError("preset", name)`. Remove the cmd-side translation in `cmd/init.go:178`.
- **MODIFY** `internal/output/errors.go:21-50` — add `case *core.InitializeError` to the `FormatError` switch. Render `"Error: initialize failed: <ref>"`.
- **DELETE** `var _ = io.EOF` line and unused `io` import in `internal/output/exporter.go:172,7`.
- **MODIFY** `internal/library/remover.go:103` — surface `os.IsNotExist` errors as `*core.NotFoundError` instead of silent swallow.
- **MODIFY** `cmd/library_add.go:534` — add `ErrorType` / `ErrorCause` fields to `BatchFailureInfo` so the typed-error chain is preserved.
- **NEW** typed `core.UsageError` in `internal/core/errors.go`. Migrate `errEmptyResources = errors.New("flag needs an argument: --resources ...")` in `cmd/library_add.go:82` to use the new typed error.
- **MODIFY** `cmd/library_init.go:144` — extract `os.Stat` / `os.MkdirAll(0o750)` / `os.WriteFile(0o600)` to a new `internal/config/scaffold.go` `WriteDefault(path string, force bool) error` helper. (Trivial fold: A-009.)
- **MODIFY** `cmd/canonicalize.go:94` — front-load `--type` validation by calling `core.ValidateDocumentType` before the `default` branch in `validateCanonicalDoc`. (Trivial fold: A-011.)
- **MODIFY** `cmd/library.go:11` — drop the dead `runF` parameter from `NewLibraryCommand`. (Trivial fold: A-007.)

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- **`errors-typed-errors`** — add `UsageError` typed error; clarify `NotFoundError` vs `FileError` boundaries (lookup branches use `NotFoundError`; filesystem I/O failures use `FileError`).
- **`errors-enhanced-errors`** — extend `BatchFailureInfo` with `ErrorType` and `ErrorCause` fields to preserve the typed-error chain.
- **`cli-error-formatting`** — extend `FormatError` dispatch set to include `InitializeError`; document the rendering format.
- **`cli-exit-codes`** — `NotFoundError` SHALL map to `ExitCodeError` (1), not `ExitCodeUsage` (2). "Not found" is a runtime state, not a user-input validation error.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `internal/cmdutil/exit.go:73` | Exit code mapping | -1 / +1 |
| `internal/cmdutil/exit_test.go:58` | Test update | 0 (literal) |
| `internal/cmdutil/exit.go:24` | Cobra dispatch | -10 / +20 |
| `cmd/library_remove.go:170,178` | Delete `FormatError` | -2 |
| `cmd/library_add.go:343,537,681,694` | Delete `FormatError` | -4 |
| `cmd/library_add.go:534` | Extend `BatchFailureInfo` | +2 fields |
| `cmd/init.go:205,209` | Delete `FormatError` | -2 |
| `cmd/init.go:178` | Delete cross-package translation | -3 |
| `cmd/library_add.go:82` | Migrate to `UsageError` | -1 / +2 |
| `internal/library/remover.go:83` | Error type swap | -2 / +1 |
| `internal/library/remover.go:103` | Surface `os.IsNotExist` | -3 / +4 |
| `internal/library/resolver.go:67` | Error type swap | -1 |
| `internal/output/errors.go:21-50` | Add `InitializeError` case | +5 |
| `internal/output/exporter.go:172,7` | Delete dead code | -2 |
| `internal/core/errors.go` | New `UsageError` type | +30 |
| `internal/config/scaffold.go` (new) | `WriteDefault` helper | +20 |
| `cmd/library_init.go:144` | Use new helper | -8 |
| `cmd/canonicalize.go:94` | Front-load validation | +1 |
| `cmd/library.go:11` | Drop dead param | -1 / -1 |

### Affected systems

- **CLI behavior:** users will see `NotFoundError` produce exit code 1 (was 2). This is a deliberate semantic correction; script authors using `exit 2` as a "not found" signal must update to `exit 1`. CHANGELOG entry required.
- **Error messages:** "preset not found" error now comes from `NotFoundError` instead of `ConfigError`; the cmd-side translation is removed. The rendered text is identical (the cmd-side wrapper was just re-typing).
- **Public API:** adds `*core.UsageError` (new type). `BatchFailureInfo` gains 2 fields (additive). No breaking changes.
- **Test surface:** `exit_test.go:58` requires the literal `ExitCodeError` update. `cmd/init_test.go` requires updates to the new `UsageError` constructor.

## Risks

- **Exit code change is user-visible.** `NotFoundError` → 1 (was 2) breaks scripts that special-case `exit 2` for not-found. **Mitigation:** document in CHANGELOG as a **BREAKING** entry; the prior mapping was semantically wrong (per the review). The change is a correctness fix, not a feature toggle.
- **Single-handling rule cleanup is mechanical but spans 7 sites.** A missed `FormatError` call doubles output. **Mitigation:** task `5.1.7` runs `rg "output\.FormatError" cmd/` and verifies the remaining calls are all in the `main.go` central handler.
- **Typed error migration in `BatchFailureInfo`** is additive (new fields); old consumers that don't read them are unaffected. **Mitigation:** task `5.11.1` verifies the new fields have `omitempty` JSON tags so JSON consumers don't see a breaking change in serialized output.
- **`UsageError` introduction** is a new public type. **Mitigation:** the type follows the existing `core.Error` builder pattern; godoc explains the purpose (user-input validation that should map to exit code 2). The constructor is exported.
- **Cobra string-prefix replacement** is a non-trivial refactor. **Mitigation:** design Decision 2 evaluates both typed-error dispatch and a documented-prefix-list approach; the chosen path is documented in `design.md`.
