## Context

The project's error-handling contract is documented across 4 specs:

1. **`errors-typed-errors/spec.md`** — 9 typed errors from `internal/core/errors.go`: `Parse`, `Validation`, `Transform`, `File`, `Config`, `NotFound`, `PartialSuccess`, `Operation`, `Initialize`.
2. **`cli-error-formatting/spec.md`** — `output.FormatError(io, err)` switch dispatches on each typed error and renders a user-facing message.
3. **`cli-exit-codes/spec.md`** — `cmdutil.ExitCodeFor(err)` maps typed errors to `ExitCodeSuccess` (0), `ExitCodeError` (1), or `ExitCodeUsage` (2).
4. **Single-handling rule** (per `golang-error-handling` skill and `cmd/AGENTS.md`) — "errors are either logged OR returned, NEVER both" — `main.go:46` calls `output.FormatError` on the returned error; cmd-side `runXxx` MUST NOT call `output.FormatError` itself.

The 2026-07-08 review identified 10 findings that violate this contract. The violations cluster in 5 areas:

- **Type confusion** (B-008, B-009): `FileError` and `ConfigError` used where `NotFoundError` is semantically correct in **7 sites** (`resolver.go:21,26,70`, `loader.go:36,53`, `adder.go:146`, `remover.go:83,88,142`). The lookup-vs-I/O distinction is lost.
- **Double format** (B-005, B-006): **10 inline `output.FormatError` calls** across 4 cmd files duplicate the central handler in `main.go:46`.
- **Dispatch set** (B-012): `FormatError` switch missing `InitializeError` case; also missing `UsageError` (new in this change).
- **Exit code** (B-001): `NotFoundError` mapped to `ExitCodeUsage` (2) is semantically wrong; "not found" is a runtime state.
- **Brittle dispatch** (B-002): `cmdutil.ExitCodeFor` uses substring matching against Cobra error text; upstream wording drift silently demotes exit code 2 to 1.
- **Silent swallow** (B-014): `RemoveResource` swallows `os.IsNotExist` on physical file removal.
- **Dead code** (B-013): `var _ = io.EOF` and unused `io` import in `internal/output/exporter.go`.
- **String-encoded typed error** (B-018): `errEmptyResources` encodes a Cobra-substring into an `errors.New`.
- **Wrapchain noise** (B-010): `libraryAdapter.AddResource` wrap with `fmt.Errorf("libraryAdapter.AddResource: %w", err)` when inner is already typed. Addressed in Phase 3 task `3.19`.
- **Lost chain** (B-011): `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` discards the typed-error chain at `cmd/library_add.go:546-549` and `cmd/library_add.go:703-706`.

### Constraints

1. **User-facing message changes are scoped to `UsageError`**. The new `Error: <flag>: <reason>` wording is a deliberate clean break from Cobra's "flag needs an argument" phrasing. All other error renderings must produce the same text as before. CHANGELOG `### Changed` entry required for the wording change.
2. **Public API additions are allowed** (new `UsageError` type, new `CobraUsageError` sentinel, new `BatchFailureInfo` fields, 9 new `MarshalJSON` methods on `core.*Error`). Public API removals are NOT allowed.
3. **Exit code 1 for `NotFoundError`** is a **BREAKING** change for any script that special-cases the prior exit code 2. CHANGELOG must call it out.
4. **Single-handling rule cleanup** must not leave a "two `FormatError` calls per error" intermediate state; the swap is atomic per file.

## Goals / Non-Goals

**Goals:**

- Map `NotFoundError` to `ExitCodeError` (1) with a CHANGELOG entry.
- Remove the `internal/warning` exit-code deprecation canary (`MaybeWarnLegacyExitCode`); the exit-code 5 → 1 migration it was warning about is now complete. Without removal, the new `NotFoundError → 1` mapping would cause the canary to fire on every interactive `not found` scenario.
- Remove all **10 inline `output.FormatError` calls** across **4 cmd files** (`cmd/library_add.go`, `cmd/library_validate.go`, `cmd/validate.go`, `cmd/init.go`); rely on `main.go:32` (factory-build) and `main.go:46` (post-Execute) as the two central handlers.
- Migrate `FileError` and `ConfigError` → `NotFoundError` for lookup branches in all 9 sites (`resolver.go:21,26,70`; `loader.go:36,53`; `adder.go:146`; `remover.go:83,88,142`).
- Add `case *core.InitializeError` and `case *core.UsageError` to `output.FormatError` switch.
- Drop the 12-prefix `cobraUsagePrefixes` substring fallback; rely on the four typed `*pflag.*Error` branches (already wired at `exit.go:64-69`) plus a new `*core.CobraUsageError` sentinel for arg-count and required-flag errors.
- Surface `os.IsNotExist` from `RemoveResource` as `*core.NotFoundError`.
- Delete `var _ = io.EOF` and unused `io` import.
- Introduce `*core.UsageError` (private fields + getters + immutable `WithSuggestions(...)` builder matching the existing 9 typed errors) and migrate `errEmptyResources`. `*core.CobraUsageError` is constructed via `MustNewCobraUsageError(err)` which panics on `err == nil` (a nil cause is a programmer error, not a recoverable state; the `Must` prefix telegraphs the panic to callers, matching `regexp.MustCompile` and `template.Must`).
- Extend `BatchFailureInfo` with `ErrorType string` and `Cause error` fields (additive). `ErrorType` is computed via a typed switch in `internal/library/adder.go` mapping well-known types to canonical names (`*core.NotFoundError` → `"NotFoundError"`, `*core.FileError` → `"FileError"`, `*core.ValidationError` → `"ValidationError"`, `*os.PathError` → `"PathError"`, default → `fmt.Sprintf("%T", cause)`).
- Add `MarshalJSON()` to all 9 existing `core.*Error` types returning `{"error": "<Error()>"}`. Note: any future exported struct fields on `core.*Error` types must be exposed via `MarshalJSON` to appear in JSON output — `json.Marshaler` precedence in stdlib means `MarshalJSON` wins over struct-field marshaling.
- Fold A-007 (drop `runF` from `NewLibraryCommand` and `NewCmdResources`, propagate to 5 call sites), A-009 (`config_init` → `internal/config/scaffold.go` + new `*config.WriteError` domain type), A-011 (`--type` validation via `core.ValidateDocumentType` defense-in-depth pre-flight in `runCanonicalize`; `MarkFlagRequired("type")` stays unchanged).
- Import the `InitRequest.Stdout io.Writer` field from `openspec/changes/fix-library-io-discipline/tasks.md:1.4` into `cmd/library_init.go:161-165`. Land AFTER that change ships its Phase 1.

**Non-Goals:**

- Changing the user-facing message text of any existing error EXCEPT `UsageError` (which adopts the clean-break wording).
- Removing the `libraryAdapter` (covered by `extract-io-adapters` Stage 2; cross-reference only).
- Refactoring `FormatError` to a generic dispatch table (the current `switch` is readable; only the missing `InitializeError`/`UsageError` cases are the gaps).
- Adding typed errors beyond `UsageError` and `CobraUsageError` (the 9 existing types cover all 10 review findings plus the new not-found sites; `UsageError` and `CobraUsageError` are the only gaps).

## Decisions

### 1. `NotFoundError` → `ExitCodeError` (1) (B-001)

**Choice**: Change `internal/cmdutil/exit.go:73` from `return ExitCodeUsage` to `return ExitCodeError` for `NotFoundError`. Update `exit_test.go:58` to expect `1`. Remove the `internal/warning.MaybeWarnLegacyExitCode(io)` canary entirely (delete `internal/warning/canary.go`, `canary_test.go`, `AGENTS.md`; remove the canary call from `main.go:51-53` and the `internal/warning` import).

**Rationale**: Per `references/05-errors.md`: *"The three exit codes are 0/1/2; usage=2 covers invalid flags/arguments only. 'not found' is a runtime state, not user input."* The prior mapping was incorrect; this is a semantic correction, not a feature change. The canary was emitting a one-time deprecation warning for the exit-code 5 → 1 migration (added in `migrate-library-rest` slice 7); that migration is now complete, so the canary is no longer needed. Without removal, the new `NotFoundError → 1` mapping would cause the canary to fire on every interactive lookup miss (e.g., `germinator library show ghost` would emit the deprecation warning every time), creating a worse UX than the BREAKING change itself. Removing the canary is the cleanest fix: the migration it warned about is over, the canary's only remaining trigger is a correctness change, and keeping it with an exemption would re-introduce the brittle canary-detection surface for unrelated exit-1 paths.

**Alternatives considered**:

- *Document the prior choice (2) and keep it*: rejected; the review explicitly recommends 1 as semantically correct.
- *Add a fourth exit code for "not found"*: rejected; the project's exit-code contract is 0/1/2 (per `cli-exit-codes/spec.md`).
- *Widen the canary signature to `(io, err)` and add a `NotFoundError`/`ValidationError` exemption*: rejected; the exemption logic couples the canary to the typed-error taxonomy, and the only reason an exemption is needed is because the canary's exit-code-1 trigger fires on the new `NotFoundError` mapping. Removing the canary eliminates the problem at its root.
- *Disable the canary globally*: rejected; regresses the existing `germinator`-wide deprecation surface.

### 2. Cobra string-prefix → typed-error dispatch (B-002)

**Choice**: **Drop the 12-substring `cobraUsagePrefixes` fallback entirely.** Trust the existing `errors.As(err, &pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error)` branch in `internal/cmdutil/exit.go:64-69`. Add a new `*core.CobraUsageError` sentinel that commands wrap Cobra arg-count (`MinimumNArgs`, `MaximumNArgs`, `ExactArgs`, `RangeArgs`) and required-flag errors in (currently emitted by Cobra/pflag as `fmt.Errorf` strings via `cobra/args.go:36-107` and `command.go:1198`). `ExitCodeFor` matches the sentinel and returns `ExitCodeUsage` (2).

**Rationale**: pflag's typed error classes (`NotExistError`, `ValueRequiredError`, `InvalidValueError`, `InvalidSyntaxError`) are stable across `v1.0.x` (verified in the local modcache at `pflag@v1.0.10/errors.go:21-149`). Substring matching against `"unknown flag"`, `"flag needs an argument"`, etc. was the brittle path; the typed dispatch matches the user's intent without depending on Cobra's wording. The four `*pflag.*Error` types + `*core.CobraUsageError` cover every observed exit-code-2 path in the existing test suite.

**Alternatives considered**:

- *Keep a narrow fallback (3-4 prefixes)*: rejected; the four typed `*pflag.*Error` branches plus the new sentinel fully replace the fallback. A residual "catch-all" prefix list invites the same wording-drift failure the redesign avoids.
- *Pin a specific Cobra version*: rejected; the project tracks `latest` (no version pin in `go.mod`); pinning would block routine upgrades.
- *Use `errors.As(err, &cobra.FlagError{})`*: `cobra@v1.10.2` does NOT export a `FlagError` type (modcache search confirms — only `FlagErrorFunc` and `SetFlagErrorFunc` exist). Therefore this is not a viable dispatch target.
- *Use `errors.As(err, &pflag.Error)`*: `pflag@v1.0.10` does NOT export an `Error` interface either; the four concrete `*pflag.*Error` types are the typed-error surface.

### 3. Single-handling rule atomicity (B-005, B-006)

**Choice**: Each of the **10 sites in 4 cmd files** is updated in one commit per file. The `output.FormatError` import is removed from files that no longer call it (per `goimports`).

**Rationale**: A "two `FormatError` calls per error" intermediate state would produce double output. Atomic per-file commits avoid the intermediate state and make bisection straightforward if one site has an issue.

**Alternatives considered**:

- *Single mega-commit for all 10 sites*: rejected; harder to review; harder to bisect if one site has an issue.
- *Add a `// nolint:dupe-format-error` comment and keep both*: rejected; the comment papers over a real bug.

### 4. `*core.UsageError` as a new typed error (B-018)

**Choice**: Add `UsageError` to `internal/core/errors.go` following the builder pattern of the existing 9 typed errors: private fields + `Flag()` / `Reason()` / `Suggestions()` getters + immutable `WithSuggestions([]string) *UsageError` builder that returns a new instance with the same `flag`/`reason` and the new suggestions. Suggestions are reserved for future flag-validation errors that may want to suggest valid alternatives (e.g., mis-spelled flag names). Constructor: `core.NewUsageError(flag string, reason string) *UsageError`. Add to `ExitCodeFor` dispatch with `ExitCodeUsage` (2). Add to `FormatError` switch rendering `"Error: <flag>: <reason>"`. `Error()` returns `"<flag>: <reason>"`. `Unwrap()` returns `nil` — `UsageError` is a leaf error (it does not wrap an underlying cause), and the godoc explicitly notes this so future maintainers do not add a `cause` field and break the contract.

**Rationale**: "User provided invalid input that didn't get caught by Cobra's `MarkFlagRequired` or `Args` validators" is a real category distinct from `NotFoundError` (runtime state) and `FileError` (filesystem I/O). The current `errEmptyResources = errors.New("flag needs an argument: --resources (must be non-empty list of refs)")` (declared at `cmd/library_create.go:70`) is a string-encoded typed error; making it typed surfaces the category to `ExitCodeFor` and `FormatError` for consistent rendering and exit-code mapping. The builder pattern is added now (rather than deferred) to keep the type consistent with the rest of the typed-error family — adding `WithSuggestions` later would be a non-breaking change to the constructor, but the family pattern is easier to maintain when every typed error follows the same shape. The clean-break wording is intentional: dropping the Cobra-encoded prefix phrasing lets `UsageError` carry semantic flag information (`--resources: must be non-empty list of refs`) rather than mimicking pflag's vocabulary.

**Alternatives considered**:

- *Use `core.ValidationError`*: rejected; `ValidationError` is for document validation (per `errors-enhanced-validation-errors/spec.md`); this is a CLI flag validation error.
- *Add a `flag` parameter to `core.ValidationError`*: rejected; that bloats `ValidationError` with concerns from the CLI layer.
- *Skip the `WithSuggestions` builder for now (defer to a follow-up)*: rejected; the builder adds ~6 LOC and matches the family pattern; deferring creates a one-off exception that future maintainers will need to rationalize.
- *Preserve the existing string verbatim via custom `FormatError` rendering*: rejected; the change is a clean break per the chosen strategy; CHANGELOG documents the wording change.

### 5. `BatchFailureInfo` extends with `ErrorType` and `Cause` (B-011)

**Choice**: Add two fields to `BatchFailureInfo` in `internal/library/adder.go:526`:

```go
type BatchFailureInfo struct {
    Source    string `json:"source"`                       // existing
    Error     string `json:"error"`                        // existing
    ErrorType string `json:"errorType,omitempty"`          // new; e.g., "FileError", "NotFoundError", "ParseError"
    Cause     error   `json:"cause,omitempty"`             // new; the original typed error
}
```

**Rationale**: The current `Error` field is a string-encoded representation that loses the typed-error chain. The new fields preserve the chain so downstream code can `errors.Is` / `errors.As` against the original error. Field names (`Source`, `Error`, `ErrorType`, `Cause`) match the existing JSON wire convention; the new fields carry `omitempty` so existing consumers see no breaking change.

**Alternatives considered**:

- *Replace `Error` with a typed `Err error` field*: rejected; JSON consumers expect the string. Additive fields preserve JSON compatibility.
- *Wrap `Error` in a `core.OperationError` with the string as the message*: rejected; the message is a free-form description, not an error; the type information is the structured piece.
- *Use `Path` (per the original `proposal.md` draft)*: rejected; the existing field is `Source`; renaming would break the wire format.

### 6. `os.IsNotExist` surface as `NotFoundError` (B-014)

**Choice**: In `internal/library/remover.go:104`, change the silent swallow to:

```go
if errors.Is(err, os.ErrNotExist) {
    return nil, core.NewNotFoundError("library file", path)
}
```

**Rationale**: The current code returns `nil` when the physical file doesn't exist, which is indistinguishable from a successful removal. The user (or upstream caller) may want to know "the file was already gone" vs "I removed it." `NotFoundError` surfaces this state.

**Alternatives considered**:

- *Keep silent swallow, document the behavior*: rejected; the reviewer's specific concern is observability; documenting doesn't help callers.
- *Return a distinct `AlreadyRemovedError` type*: rejected; over-engineered for a single call site; `NotFoundError` is the closest semantic match.

### 7. `*core.CobraUsageError` sentinel for arg-count and required-flag strings

**Choice**: Add a new sentinel `CobraUsageError` in `internal/core/errors.go` constructed via `MustNewCobraUsageError(err error)` which **panics on `err == nil`** (a nil cause is a programmer error, not a recoverable state). Wrap any error emitted by `cobra.ExactArgs`, `cobra.MinimumNArgs`, `cobra.MaximumNArgs`, `cobra.RangeArgs`, `cobra.MarkFlagRequired`-flavored `required flag(s)` strings with `MustNewCobraUsageError(err)`. `ExitCodeFor` matches the sentinel and returns `ExitCodeUsage` (2). The nil-guard path in the original proposal (`NewCobraUsageError(nil).Error() == "cobra usage"`) is dropped: a typed-error constructor that returns a meaningful-looking error for a missing required input is hiding a bug. The `Must*` prefix telegraphs the panic to callers (matching `regexp.MustCompile` and `template.Must`).

**Rationale**: The substring fallback handled eight distinct Cobra-emitted phrases (`"requires at least N arg(s)"`, `"accepts at most N arg(s)"`, `"required flag(s) ..."`, etc.). The typed dispatch drops substring matching entirely; the sentinel carries the underlying Cobra error so downstream code can still inspect `errors.As(err, &cobraUsage)` if needed. The sentinel is small (~10 LOC) and follows the existing typed-error pattern. The panic-on-nil contract is enforced by the `Must` naming convention; if a caller wants nil-safety, they should use a `Try*` / `New*` variant (not provided in this change — no current call site needs it). Per `golang-cli-architecture/references/05-errors.md` "Anti-Pattern: `check(err)` / `must()`" section (line 223), `must()` is appropriate for programmer-error guards (mirroring `regexp.MustCompile`/`template.Must`) but breaks down for runtime recovery paths. The panic branch here is unreachable in practice (Cobra always emits a non-nil error from arg validators and `MarkFlagRequired`), so this is a programmer-error guard, not a runtime recovery path.

**Important**: `*core.CobraUsageError` has **zero current call sites** in the existing test suite. Cobra's `Args` validators and `MarkFlagRequired` errors flow through `cmd.Execute()` to `main.go:46` directly without touching `RunE`; the new dispatch is wired (`tasks.md:1.3`) but no production code wraps a Cobra error with the sentinel today. The sentinel is **reserved for future use** — e.g., when a command adds a custom arg validator that needs typed exit-code mapping. Task `3.14` documents this no-op-for-now contract and adds unit tests for the sentinel's `Error()` / `Unwrap()` behavior including the panic-on-nil path.

**Alternatives considered**:

- *Stick to substring matching for these specific strings*: rejected; defeats the goal of dropping the brittle substring fallback.
- *Wrap the strings as `*core.ValidationError`*: rejected; `ValidationError` is for document validation and has different semantics (exit 1, not exit 2).
- *Skip the sentinel and rely on the four typed `*pflag.*Error` branches only*: rejected; the test suite (`TestExitCodeFor` at `internal/cmdutil/exit_test.go:39-66` plus task 1.4's new rows for `*core.UsageError` and `*core.CobraUsageError`) asserts that arg-count errors map to `ExitCodeUsage`; without the sentinel those become defaults to `ExitCodeError`.
- *Use `NewCobraUsageError` with a nil-guard fallback returning `"cobra usage"`*: rejected; hides a programmer error. The `Must*` prefix + panic is the correct pattern.

### 8. `MarshalJSON` on all `core.*Error` types (W8)

**Choice**: Add `MarshalJSON() ([]byte, error)` to all 9 existing typed errors (`Parse`, `Validation`, `Transform`, `File`, `Config`, `NotFound`, `Operation`, `Initialize`, `PartialSuccess`) plus the two new ones (`Usage`, `CobraUsage`). Each returns `{"error": "<Error()>"}`.

**Rationale**: Without `MarshalJSON`, stdlib's `json.Marshal(*core.NotFoundError)` returns `{}` (an error interface with no exported fields). The new `BatchFailureInfo.Cause` field would therefore render as `{}` in JSON, which contradicts the spec scenario "AND a `Cause` field with a non-nil value SHALL appear as the underlying error's `Error()` string". The `MarshalJSON` implementation aligns the JSON wire-format with the spec scenario; the `{"error": "<Error()>"}` shape is the only sensible JSON projection of a Go error interface.

**Alternatives considered**:

- *Use a separate `CauseStr string` field next to `Cause error`*: rejected; dishonest about the wire shape (`CauseStr` would coexist with `Cause`).
- *Accept the current `{}` rendering*: rejected; the spec scenario explicitly contradicts it.

## Risks / Trade-offs

- **Exit code change is a BREAKING for scripts** that special-case `exit 2` for not-found. **Mitigation**: CHANGELOG `### BREAKING` entry in Phase 6.0 (moved up from Archive so the change cannot ship without it); the `internal/warning` canary is REMOVED in Phase 1.7a — without removal, the new `NotFoundError → 1` mapping would cause the canary to fire on every interactive lookup miss.
- **Cobra typed-error dispatch requires runtime `errors.As` against `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` and the new `*core.CobraUsageError` sentinel.** pflag's typed errors are stable across `v1.0.x` (verified locally at `pflag@v1.0.10/errors.go:21-149`); `*core.CobraUsageError` is project-owned so its API stability is our responsibility. **Mitigation**: the four pflag types are in pflag's stable API; the new sentinel is documented under `errors-typed-errors`. Tasks `1.4` widens the exit-test coverage; if a future Cobra version drops one of the four typed types, the test row fails fast.
- **10-site single-handling rule cleanup** spans 4 cmd files; one missed site doubles output. **Mitigation**: task `2.11` runs `rg "output\.FormatError" cmd/ main.go` after the changes; the only remaining CALLS are `main.go:32` (factory-build), `main.go:46` (post-Execute), and the two test files (`cmd/init_test.go:185`, `cmd/show_test.go:239`) that exercise the renderer as a unit. Comment references in `cmd/AGENTS.md`, `cmd/commands/AGENTS.md`, and the `*_test.go` files are expected. Tests with double-output stderr assertions fail fast.
- **`UsageError` wording is a clean break.** The user-facing text changes from `Error: flag needs an argument: --resources (must be non-empty list of refs)` to `Error: --resources: must be non-empty list of refs`. **Mitigation**: CHANGELOG `### Changed` entry; the change is bounded to one call site (`cmd/library_create.go:203`); `cmd/library_create_test.go:152-176` is updated in task `3.12` to assert the new shape.
- **Lookup-branch migration widens to 9 sites and 6 tests.** **Mitigation**: tasks `3.1-3.9` cover all 9 sites (8 `NewFileError("…not found", nil)` + 1 `NewConfigError("preset", …, "preset not found")`); task `3.13` updates `internal/library/resolver_test.go:174-177`, `cmd/library_create_test.go:343-365` (T11 `TestRunCreatePreset_RefReferencesMissingResource`; the `ExitCodeUsage` assertion is at line 364), `cmd/show_test.go:151,179`, `cmd/library_remove_test.go:397`, `test/e2e/init_test.go:108`, `internal/cmdutil/exit_test.go:58` to assert `*core.NotFoundError` → `ExitCodeError` (1). The verification gate at task `3.15` uses the regexes `rg "NewFileError\([^)]*not found" internal/library/` and `rg "NewConfigError\("preset"[^)]*preset not found" internal/library/` to catch all 9 sites.
- **`BatchFailureInfo` field addition** is a wire-format change (additive). **Mitigation**: new fields carry `json:",omitempty"` so old consumers see no breaking change in serialized output. `ErrorType` is computed via a typed switch in `internal/library/adder.go` mapping well-known types to canonical names; non-typed causes (e.g., `*os.PathError`) get a sensible canonical name.
- **`MarshalJSON` adoption on all 9 existing `core.*Error` types** changes JSON wire-shape for any consumer that re-encodes the cause. **Mitigation**: the shape `{"error": "<Error()>"}` is the only sensible JSON projection of an error interface (stdlib's default would marshal as `{}`); the spec scenario at `specs/errors-enhanced-errors/spec.md:26-30` codifies this. Note: any future exported struct fields on `core.*Error` types must be exposed via `MarshalJSON` to appear in JSON output — `json.Marshaler` precedence in stdlib means `MarshalJSON` wins over struct-field marshaling. No E2E test asserts the old shape today (verified by `rg "BatchFailureInfo" test/`); population is bounded to the new code paths.
- **Cross-change ownership.** This change claims ownership of the `errEmptyResources` migration (task `3.12`) and the `cmd/library_init.go:161-165` `(*Library).Init` call update (task `4.5`, which sets the new `InitRequest.Stdout io.Writer` field) that are also referenced in `openspec/changes/harden-tests-and-coverage/proposal.md:59` and `openspec/changes/fix-library-io-discipline/proposal.md:60`. **Mitigation**: explicit cross-references are added to both source changes' `tasks.md` de-scoping their respective tasks; task `4.5` gates on `fix-library-io-discipline` Phase 1 shipping first (the `InitRequest` struct must include the `Stdout` field before this change sets it).

## Migration Plan

The change ships in **one PR with 7 atomic phases** (each commit is independently testable):

1. **Phase 0 — Cross-change reconciliation** (task `0.1`): re-verify every file:line in the change against the current tree; record the real locations in a setup comment so the artifact text stays auditable.
2. **Phase 1 — Exit-code semantics + dispatch set + canary REMOVAL** (tasks `1.1-1.10`): `NotFoundError` → exit 1; drop `cobraUsagePrefixes` substring fallback; add `*core.CobraUsageError` sentinel + `*core.UsageError` constructor + `MarshalJSON` on all `core.*Error` types; **delete the `internal/warning` canary entirely** (`canary.go`, `canary_test.go`, `AGENTS.md`; remove the call from `main.go:51-53`; remove the `internal/warning` import); update `exit_test.go` to reflect new expectations; **add an explicit `*config.WriteError → ExitCodeError (1)` row in `ExitCodeFor`** (per the spec ADDED Requirement at `cli-exit-codes/spec.md:91-100` — explicit row preferred over default-fallthrough for contract stability); add `case *core.InitializeError` and `case *core.UsageError` to `FormatError`; delete `var _ = io.EOF` dead code. Verify `mise run test` passes.
3. **Phase 2 — Single-handling rule cleanup** (tasks `2.1-2.12`): delete 10 inline `output.FormatError` calls in `cmd/library_add.go:355,549,693,706`, `cmd/library_validate.go:182,190`, `cmd/validate.go:123`, `cmd/init.go:218,222,258`. Lift typed causes into `BatchFailureInfo.Cause` at the two `cmd/library_add.go` sites (sites `549,706`). Verify `mise run test` and `mise run test:e2e` pass; verify no double-output in E2E test output capture.
4. **Phase 3 — Type migration + chain preservation** (tasks `3.1-3.20`): migrate 7 lookup-branch sites (`resolver.go:21,26,70`; `loader.go:36,53`; `adder.go:146`; `remover.go:83,88,142`) to `NotFoundError`; surface `os.IsNotExist` at `remover.go:104`; populate `BatchFailureInfo.ErrorType`/`Cause` at 5 sites in `adder.go`; migrate `errEmptyResources` to `*core.UsageError` at `cmd/library_create.go:70,203`; the `*core.CobraUsageError` sentinel has zero current call sites (per task `3.14`) — no production wrap sites are introduced in this change; update 6 test files to reflect new mappings; drop the redundant `libraryAdapter.AddResource: %w` wrap at the B-010 site (task `3.19`). Verify `mise run check`.
5. **Phase 4 — Trivial folds + cross-change imports** (tasks `4.1-4.7`): create `internal/config.WriteDefault`; refactor `cmd/config_init.go:144-156`; drop `runF` from `NewLibraryCommand` and `NewCmdResources`, propagate to 5 call sites; replace `MarkFlagRequired("type")` in `cmd/canonicalize.go:96` with `MatchAll(ExactArgs, OnlyValidArgs)` + pre-flight; import the `CreateLibrary` `stdout` parameter from `fix-library-io-discipline` task `1.4`; cross-reference `harden-tests-and-coverage` task `6.6`. Verify `mise run check`.
6. **Phase 5 — Verification + spec sync** (tasks `5.1-5.6`): `mise run build/lint/test/test:e2e/test:coverage` (target ≥ 70% per package); manual exit-code sanity check; `openspec validate enforce-error-discipline --strict` (must pass).
7. **Phase 6 — Archive + CHANGELOG** (tasks `6.0-6.3`): write CHANGELOG `### BREAKING` for exit-code and `### Changed` for `UsageError` wording FIRST (so the change cannot archive without it); run `osc-archive-change`; confirm `openspec list --json` shows the change archived.

**Rollback strategy**: revert each phase commit independently. Phase 0 (setup note) is harmless. Phase 1 is additive (new `UsageError` + `CobraUsageError` + `MarshalJSON` + `UsageError`/`InitializeError` dispatch cases) except for the `NotFoundError → 1` exit-code flip + deletion of the `internal/warning` canary (`canary.go`, `canary_test.go`, `AGENTS.md`; canary call block in `main.go:51-53`; `internal/warning` import) — both are the BREAKING. Phase 2's cleanup is mechanical (delete calls; revert restores them). Phase 3's type migrations swap error types; revert restores the prior types. Phase 4's folds are code-only; revert restores the prior wiring. Phase 5 (verification + spec sync) is non-runtime (verification + `osc-sync-specs`); revert is "re-run validation". Phase 6 is archive metadata and CHANGELOG; rollback is `git revert` of the CHANGELOG commit.
