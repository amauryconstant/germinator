## Context

The project's error-handling contract is documented across 4 specs:

1. **`errors-typed-errors/spec.md`** — 9 typed errors from `internal/core/errors.go`: `Parse`, `Validation`, `Transform`, `File`, `Config`, `NotFound`, `PartialSuccess`, `Operation`, `Initialize`.
2. **`cli-error-formatting/spec.md`** — `output.FormatError(io, err)` switch dispatches on each typed error and renders a user-facing message.
3. **`cli-exit-codes/spec.md`** — `cmdutil.ExitCodeFor(err)` maps typed errors to `ExitCodeSuccess` (0), `ExitCodeError` (1), or `ExitCodeUsage` (2).
4. **Single-handling rule** (per `golang-error-handling` skill and `cmd/AGENTS.md`) — "errors are either logged OR returned, NEVER both" — `main.go:43` calls `output.FormatError` on the returned error; cmd-side `runXxx` MUST NOT call `output.FormatError` itself.

The 2026-07-08 review identified 10 findings that violate this contract. The violations cluster in 3 areas:

- **Type confusion** (B-008, B-009): `FileError` and `ConfigError` used where `NotFoundError` is semantically correct. The lookup-vs-I-O distinction is lost.
- **Double format** (B-005, B-006): 7 inline `output.FormatError` calls in `cmd/` duplicate the central handler in `main.go`.
- **Dispatch set** (B-012): `FormatError` switch missing `InitializeError` case.
- **Exit code** (B-001): `NotFoundError` mapped to `ExitCodeUsage` (2) is semantically wrong; "not found" is a runtime state.
- **Brittle dispatch** (B-002): `cmdutil.ExitCodeFor` uses substring matching against Cobra error text; upstream wording drift silently demotes exit code 2 to 1.
- **Silent swallow** (B-014): `RemoveResource` swallows `os.IsNotExist` on physical file removal.
- **Dead code** (B-013): `var _ = io.EOF` and unused `io` import in `internal/output/exporter.go`.
- **String-encoded typed error** (B-018): `errEmptyResources` encodes a Cobra-substring into an `errors.New`.
- **Wrapchain noise** (B-010): `libraryAdapter.AddResource` wrap with `fmt.Errorf("libraryAdapter.AddResource: %w", err)` when inner is already typed.
- **Lost chain** (B-011): `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` discards the typed-error chain.

### Constraints

1. **No user-facing message changes** (except the "slice 2" → CHANGELOG navigation; that's in the doc-reconciliation change). Error rendering must produce the same text as before.
2. **Public API additions are allowed** (new `UsageError` type, new `BatchFailureInfo` fields). Public API removals are NOT allowed.
3. **Exit code 1 for `NotFoundError`** is a **BREAKING** change for any script that special-cases the prior exit code 2. CHANGELOG must call it out.
4. **Single-handling rule cleanup** must not leave a "two `FormatError` calls per error" intermediate state; the swap is atomic per file.

## Goals / Non-Goals

**Goals:**

- Map `NotFoundError` to `ExitCodeError` (1) with a CHANGELOG entry.
- Remove all 7 inline `output.FormatError` calls in `cmd/`; rely on `main.go`'s central handler.
- Migrate `FileError` → `NotFoundError` for lookup branches in `internal/library/remover.go` and `resolver.go`.
- Add `case *core.InitializeError` to `output.FormatError` switch.
- Replace brittle Cobra string-prefix matching with `errors.As`-based dispatch.
- Surface `os.IsNotExist` from `RemoveResource` as `*core.NotFoundError`.
- Delete `var _ = io.EOF` and unused `io` import.
- Introduce `*core.UsageError` and migrate `errEmptyResources`.
- Add `ErrorType` / `ErrorCause` fields to `BatchFailureInfo`.
- Drop `libraryAdapter.AddResource` wrap noise (covered by `extract-io-adapters` Stage 2 — cross-reference).
- Fold A-007 (dead `runF` param), A-009 (`config_init` → `internal/config/scaffold.go`), A-011 (front-load `--type` validation).

**Non-Goals:**

- Changing the user-facing message text of any existing error.
- Removing the `libraryAdapter` (covered by `extract-io-adapters` Stage 2).
- Refactoring `FormatError` to a generic dispatch table (the current `switch` is readable; only the missing `InitializeError` case is the bug).
- Adding typed errors beyond `UsageError` (the 9 existing types cover all 10 review findings; `UsageError` is the only gap).

## Decisions

### 1. `NotFoundError` → `ExitCodeError` (1) (B-001)

**Choice**: Change `internal/cmdutil/exit.go:73` from `return ExitCodeUsage` to `return ExitCodeError` for `NotFoundError`. Update `exit_test.go:58` to expect `1`.

**Rationale**: Per `references/05-errors.md`: "The three exit codes are 0/1/2; usage=2 covers invalid flags/arguments only. 'not found' is a runtime state, not user input." The prior mapping was incorrect; this is a semantic correction, not a feature change.

**Alternatives considered**:

- *Document the prior choice (2) and keep it*: rejected; the review explicitly recommends 1 as semantically correct.
- *Add a fourth exit code for "not found"*: rejected; the project's exit-code contract is 0/1/2 (per `cli-exit-codes/spec.md`).

### 2. Cobra string-prefix → typed-error dispatch (B-002)

**Choice**: Replace `cobraUsagePrefixes` substring matching in `internal/cmdutil/exit.go:24` with `errors.As(err, &pflag.Error)` and `errors.As(err, &cobra.FlagError{})` dispatch. For errors that don't match a typed error, fall back to a documented prefix list.

**Rationale**: spf13/cobra exports typed errors (`cobra.FlagError`, `pflag.Error`) that wrap the underlying flag errors. Typed dispatch is robust to wording drift; substring matching breaks on every Cobra version bump.

**Alternatives considered**:

- *Pin a specific Cobra version*: rejected; the project tracks `latest` (no version pin in `go.mod`); pinning would block routine upgrades.
- *Keep substring matching and document the prefixes as a contract*: rejected; the reviewer's specific concern is upstream wording drift; a documented contract doesn't help when Cobra changes the strings.

### 3. Single-handling rule atomicity (B-005, B-006)

**Choice**: Each of the 7 sites is updated in one commit. The `output.FormatError` import is removed from files that no longer call it (per `goimports`).

**Rationale**: A "two `FormatError` calls per error" intermediate state would produce double output. Atomic per-file commits avoid the intermediate state.

**Alternatives considered**:

- *Single mega-commit for all 7 sites*: rejected; harder to review; harder to bisect if one site has an issue.
- *Add a `// nolint:dupe-format-error` comment and keep both*: rejected; the comment papers over a real bug.

### 4. `*core.UsageError` as a new typed error (B-018)

**Choice**: Add `UsageError` to `internal/core/errors.go` following the existing builder pattern. Constructor: `core.NewUsageError(flag string, reason string) *UsageError`. Add to `ExitCodeFor` dispatch with `ExitCodeUsage` (2). Add to `FormatError` switch rendering `"Error: <flag>: <reason>"`.

**Rationale**: "User provided invalid input that didn't get caught by flag parsing" is a real category distinct from `NotFoundError` (which is a runtime state) and `FileError` (which is filesystem I/O). The current `errEmptyResources = errors.New("flag needs an argument: --resources ...")` is a string-encoded typed error; making it typed surfaces the category to `ExitCodeFor` and `FormatError` for consistent rendering and exit-code mapping.

**Alternatives considered**:

- *Use `core.ValidationError`*: rejected; `ValidationError` is for document validation (per `errors-enhanced-validation-errors/spec.md`); this is a CLI flag validation error, not a document validation error.
- *Add a `flag` parameter to `core.ValidationError`*: rejected; that bloats `ValidationError` with concerns from the CLI layer.

### 5. `BatchFailureInfo` extends with `ErrorType` and `ErrorCause` (B-011)

**Choice**: Add two fields to `BatchFailureInfo` in `internal/library/adder.go`:

```go
type BatchFailureInfo struct {
    Path      string
    Error     string  // existing
    ErrorType string  // new; e.g., "FileError", "NotFoundError", "ParseError"
    Cause     error   // new; the original typed error (omitempty in JSON)
}
```

**Rationale**: The current `Error` field is a string-encoded representation that loses the typed-error chain. The new fields preserve the chain so downstream code can `errors.Is` / `errors.As` against the original error.

**Alternatives considered**:

- *Replace `Error` with a typed `Err error` field*: rejected; JSON consumers expect the string. Additive fields preserve JSON compatibility.
- *Wrap `Error` in a `core.OperationError` with the string as the message*: rejected; the message is a free-form description, not an error; the type information is the structured piece.

### 6. `os.IsNotExist` surface as `NotFoundError` (B-014)

**Choice**: In `internal/library/remover.go:103`, change the silent swallow to:

```go
if errors.Is(err, os.ErrNotExist) {
    return nil, core.NewNotFoundError("library file", path)
}
```

**Rationale**: The current code returns `nil` when the physical file doesn't exist, which is indistinguishable from a successful removal. The user (or upstream caller) may want to know "the file was already gone" vs "I removed it." `NotFoundError` surfaces this state.

**Alternatives considered**:

- *Keep silent swallow, document the behavior*: rejected; the reviewer's specific concern is observability; documenting doesn't help callers.
- *Return a distinct `AlreadyRemovedError` type*: rejected; over-engineered for a single call site; `NotFoundError` is the closest semantic match.

## Risks / Trade-offs

- **Exit code change is a BREAKING for scripts** that special-case `exit 2` for not-found. **Mitigation**: CHANGELOG entry; the change is a correctness fix, not a feature toggle; the review explicitly recommends the change.
- **Cobra typed-error dispatch requires runtime `errors.As` against `pflag.Error` and `cobra.FlagError`.** If Cobra does not export these types in a future version, dispatch fails silently. **Mitigation**: design Decision 2's fallback to a documented prefix list handles the case where typed dispatch fails; the prefix list is narrower than the current one (only the 3-4 most common Cobra error prefixes remain).
- **7-site single-handling rule cleanup** spans multiple files; one missed site doubles output. **Mitigation**: task `5.1.7` runs `rg "output\.FormatError" cmd/` after the changes; only the `main.go` central handler and the new `cmd/library_init.go` (if it uses `output.FormatError` for a non-error path) should remain.
- **`UsageError` introduction** is a new public type. **Mitigation**: the type follows the existing `core.Error` builder pattern; godoc explains the purpose; the constructor is exported. Test coverage ≥ 70% per `config.testing`.
- **`BatchFailureInfo` field addition** is a wire-format change (additive). **Mitigation**: new fields have `omitempty` JSON tags so old consumers don't see a breaking change in serialized output.

## Migration Plan

The change ships in **one PR with 4 atomic phases** (each commit is independently testable):

1. **Phase 1 — Exit code + dispatch set** (tasks 5.2, 5.5, 5.6, 5.7): change `NotFoundError` → 1; add `InitializeError` case; delete dead code; introduce `UsageError` typed error. Verify `mise run test` passes.
2. **Phase 2 — Single-handling rule cleanup** (tasks 5.1): remove 7 inline `FormatError` calls. Verify `mise run test` and `mise run test:e2e` pass; verify no double-output in E2E test output capture.
3. **Phase 3 — Type migration + error chain** (tasks 5.3, 5.4, 5.8, 5.9): `FileError` → `NotFoundError` in `remover.go`; `ConfigError` → `NotFoundError` in `resolver.go`; surface `os.IsNotExist`; replace Cobra string-prefix with typed dispatch; extend `BatchFailureInfo`. Verify `mise run check`.
4. **Phase 4 — Trivial folds + docs** (tasks 5.10, 5.11, 5.12, 5.13, 5.14): extract `config_init` to `internal/config/scaffold.go`; front-load `--type` validation; drop dead `runF` param; spec deltas applied via `osc-sync-specs`. Verify `openspec validate enforce-error-discipline --strict`.

**Rollback strategy**: revert each phase commit independently. Phase 1 is additive (new `UsageError` type, new dispatch case) except for the exit code change (which is the BREAKING). Phase 2's cleanup is mechanical (delete calls; revert restores them). Phase 3's type migrations swap error types; revert restores the prior types. Phase 4 is doc-only.
