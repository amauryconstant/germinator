# exit-codes Specification (delta)

> **Cross-references:** this change also modifies `errors-typed-errors`, `cli-error-formatting`, `errors-enhanced-errors`. See those delta specs.

## REMOVED Requirements

### Scenario: NotFoundError maps to ExitCodeUsage

The prior scenario *"NotFoundError maps to ExitCodeUsage"* (from `openspec/specs/cli-exit-codes/spec.md:33`) is removed. `*core.NotFoundError` no longer maps to exit code 2; the corrected mapping to `ExitCodeError` (1) is captured in the `## MODIFIED Requirements` block below.

### Scenario: Cobra usage errors via substring matching

The prior wording *"Cobra usage errors not wrapped by pflag (detected via `strings.Contains` on the error message against any of the 12 known substrings: …)"* is removed. The substring fallback is dropped in favor of typed dispatch; the new dispatch contract is documented under `## MODIFIED Requirements` below.

> **Why:** substring matching against Cobra error text is brittle to upstream wording drift; the project now uses typed dispatch against `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` (already in `internal/cmdutil/exit.go:64-69`) plus a new `*core.CobraUsageError` sentinel.

### Scenario: Exit code deprecation canary (entire requirement)

The prior `### Requirement: Exit code deprecation canary` block (at `openspec/specs/cli-exit-codes/spec.md:43-88`) and all associated scenarios (at lines 52-87) are removed. The `internal/warning.MaybeWarnLegacyExitCode` canary was added in `migrate-library-rest` (slice 7, 2026-07-01) to warn users about the exit-code 5 → 1 migration. That migration is now complete; the canary is removed in change `enforce-error-discipline` (Phase 1.7a). Without removal, the new `*core.NotFoundError → 1` mapping in this change would cause the canary to fire on every interactive lookup miss (e.g., `germinator library show ghost`), creating a worse UX than the BREAKING change itself. The entire `internal/warning` package is deleted (canary.go, canary_test.go, AGENTS.md); the canary call block is removed from `main.go:51-53`.

## MODIFIED Requirements

### Requirement: ExitCodeFor function

The `cmdutil.ExitCodeFor(err error) ExitCode` function SHALL map an error to an exit code:

- `nil` → `ExitCodeSuccess` (0)
- `*pflag.NotExistError`, `*pflag.ValueRequiredError`, `*pflag.InvalidValueError`, `*pflag.InvalidSyntaxError` (typed errors that exist in pflag `v1.0.10`) → `ExitCodeUsage` (2)
- `*core.NotFoundError` → `ExitCodeError` (1) (CORRECTED — was `ExitCodeUsage` (2) prior to change `enforce-error-discipline`; a lookup miss is a runtime state, not a user-input validation error)
- `*core.UsageError` → `ExitCodeUsage` (2) (NEW — `UsageError` introduced in change `enforce-error-discipline`; CLI flag validation maps to usage exit code)
- `*core.CobraUsageError` → `ExitCodeUsage` (2) (NEW — sentinel for Cobra arg-validation errors that the typed pflag dispatch does not cover; `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` failures and `required flag(s) "..."` errors)
- `*core.PartialSuccessError` with `Succeeded > 0` → `ExitCodeSuccess` (0)
- `*core.PartialSuccessError` with `Succeeded == 0` → `ExitCodeError` (1)
- `*config.WriteError` → `ExitCodeError` (1) (NEW — `*config.WriteError` introduced in change `enforce-error-discipline`; carries `op`, `path`, `cause`; an I/O failure during config scaffolding is an operational error, not a user-input validation error)
- All other errors → `ExitCodeError` (1)

**Change**: corrected the `NotFoundError` mapping (1 instead of 2); added `UsageError` and `CobraUsageError` mappings; replaced substring matching with `errors.As`-based typed dispatch against `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` plus the new `*core.CobraUsageError` sentinel. The prior mapping was semantically wrong per the 2026-07-08 review.

The four `*pflag.*Error` types are verified stable in pflag `v1.0.10` (`pflag/errors.go:21-149`); if a future pflag version drops or renames any of them, dispatch falls through to `ExitCodeError` (1) and the corresponding test row in `internal/cmdutil/exit_test.go` fails fast. Note: a fourth `*pflag.InvalidValueError` row is added to `exit_test.go` in Phase 1.4 (alongside the existing `NotExistError`, `ValueRequiredError`, and `InvalidSyntaxError` rows at lines 50-52) to widen coverage to all four pflag types.

#### Scenario: ExitCodeFor table-driven

- **WHEN** `cmdutil.ExitCodeFor` is called with each error type from the list above
- **THEN** it SHALL return the corresponding exit code

#### Scenario: NotFoundError returns ExitCodeError (1)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.NotFoundError`
- **THEN** it SHALL return `ExitCodeError` (1) — the prior mapping (2) was a semantic error and is corrected in change `enforce-error-discipline`. This is a BREAKING change for any script that special-cases `exit 2` for not-found scenarios; the CHANGELOG calls it out.

#### Scenario: UsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2) — `UsageError` is a CLI flag validation error and maps to the usage exit code.

#### Scenario: CobraUsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.CobraUsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2) — the sentinel wraps Cobra arg-validation errors emitted as `fmt.Errorf` strings, which historically landed on the substring fallback.

#### Scenario: Typed-error dispatch takes precedence over generic fallback

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with an error that wraps `*pflag.NotExistError` (or one of the other three `*pflag.*Error` types)
- **THEN** it SHALL return `ExitCodeUsage` (2) via the `errors.As` dispatch
- **AND** the dispatch SHALL NOT rely on substring matching (the substring fallback is dropped in this change)

## ADDED Requirements

### Requirement: UsageError exit-code mapping

`*core.UsageError` SHALL map to `ExitCodeUsage` (2) in `cmdutil.ExitCodeFor`. Consistent with the existing mapping for `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` and `*core.CobraUsageError` — CLI flag validation errors are user-input problems and produce exit code 2.

**Change**: NEW requirement. The type is introduced in change `enforce-error-discipline`.

#### Scenario: UsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.UsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

### Requirement: CobraUsageError exit-code mapping

`*core.CobraUsageError` SHALL map to `ExitCodeUsage` (2) in `cmdutil.ExitCodeFor`. Replaces the substring-prefix dispatch fallback that previously mapped errors emitted by `cobra.ExactArgs`, `cobra.MinimumNArgs`, `cobra.MaximumNArgs`, `cobra.RangeArgs`, and `cobra.MarkFlagRequired`-derived `"required flag(s) \"...\" not set"` strings.

**Change**: NEW requirement. The sentinel is introduced in change `enforce-error-discipline`.

#### Scenario: CobraUsageError returns ExitCodeUsage

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*core.CobraUsageError`
- **THEN** it SHALL return `ExitCodeUsage` (2)

### Requirement: WriteError exit-code mapping

`*config.WriteError` SHALL map to `ExitCodeError` (1) in `cmdutil.ExitCodeFor`. An I/O failure during config scaffolding (e.g., `WriteDefault` path) is an operational error, not a user-input validation error; the error wraps the underlying `*os.PathError` via `cause` for inspection.

**Change**: NEW requirement. The type is introduced in change `enforce-error-discipline` (Phase 4 task `4.1`); the dispatch case is added in Phase 1 task `1.5b`.

#### Scenario: WriteError returns ExitCodeError (1)

- **WHEN** `cmdutil.ExitCodeFor(err)` is called with `*config.WriteError`
- **THEN** it SHALL return `ExitCodeError` (1) — a config I/O failure is an operational error, not a CLI usage error.

### Requirement: Substring dispatch fallback removal

`cmdutil.ExitCodeFor` SHALL NOT use substring-prefix matching against Cobra error messages. The pre-change implementation used a 12-prefix substring-match list (full enumeration in `openspec/specs/cli-exit-codes/spec.md:33`); the post-change implementation uses typed dispatch via `errors.As` against `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error` plus the new `*core.CobraUsageError` sentinel.

**Change**: corrected dispatch. Removing the substring fallback eliminates the brittle-wording concern the 2026-07-08 review flagged (B-002).

#### Scenario: No substring matching in ExitCodeFor

- **WHEN** `cmdutil.ExitCodeFor` is inspected
- **THEN** it SHALL NOT contain a `cobraUsagePrefixes` slice
- **AND** it SHALL NOT contain a `hasCobraUsagePrefix` helper
- **AND** the dispatch SHALL rely on `errors.As` against the typed errors listed above
