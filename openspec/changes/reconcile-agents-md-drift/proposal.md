## Why

The 2026-07-08 code review (`docs/reviews/2026-07-08-code-review.md`) identified **8 drift entries (6 in AGENTS.md, 1 canary-string inaccuracy, 1 missing build tag)**. The drift includes: the `NewFactory` signature, the `FindLibrary` signature, the `RefreshError` and `SkipInfo` struct shapes, the `FormatError` dispatch set, the integration-test file list in `test/AGENTS.md`, the canary "slice 2" reference unfindable in `CHANGELOG.md`, and a missing `//go:build integration` directive on `internal/cmdutil/integration_test.go`. The drift creates an onboarding hazard: new contributors reading AGENTS.md form wrong mental models of the code. Additionally, four deprecated JSON output types (`PresetsJSONOutput`, `PresetInfoJSON`, `ShowResourceJSONOutput`, `ShowPresetJSONOutput`) are dead code from earlier slices and should be removed.

This change reconciles documentation and build tags with the actual codebase. It is **documentation-only** at the production-code level — no user-facing behavior changes. It adds one new infrastructure spec, `infrastructure-documentation-accuracy` (see Capabilities), and modifies no existing spec.

## What Changes

- **MODIFY** `cmd/AGENTS.md:278` — fix Foundation Units table cell for `NewFactory` to include `ctx` parameter.
- **MODIFY** `internal/library/AGENTS.md:71` — fix `FindLibrary` signature description.
- **MODIFY** `internal/library/AGENTS.md:353` — rename `RefreshSkipped` to `SkipInfo` to match `internal/library/refresher.go:43`.
- **MODIFY** `internal/library/AGENTS.md:358` — fix `RefreshError` struct fields (`Ref`, `Field`, `Type`) to match `internal/library/refresher.go:49-53`.
- **MODIFY** `internal/output/AGENTS.md:15` — expand `FormatError` dispatch set to include `NotFoundError` and `OperationError`.
- **MODIFY** `internal/warning/canary.go:44` — replace the "slice 2" reference with a CHANGELOG entry pointer.
- **MODIFY** `internal/cmdutil/integration_test.go:1` — add `//go:build integration` directive to match `test/AGENTS.md` contract.
- **MODIFY** `test/AGENTS.md:62-86` — list both `internal/parser/integration_test.go` and `internal/cmdutil/integration_test.go`.
- **DELETE** deprecated JSON output types: `PresetsJSONOutput`, `PresetInfoJSON`, `ShowResourceJSONOutput`, `ShowPresetJSONOutput` from `cmd/presets.go:160` and `cmd/show.go:225` (unused by any caller in `./cmd ./internal ./test` per `rg -e 'PresetsJSONOutput' -e 'PresetInfoJSON' -e 'ShowResourceJSONOutput' -e 'ShowPresetJSONOutput' ./cmd ./internal ./test`).

## Capabilities

### New Capabilities

- **`infrastructure-documentation-accuracy`** — establishes the post-reconciliation contract that AGENTS.md files throughout the project MUST accurately reflect the code they document, and that dead code SHALL be deleted rather than annotated. Captured at `specs/infrastructure-documentation-accuracy/spec.md` (2 requirements, 6 scenarios).

### Modified Capabilities

None. The `cli-factory`, `library-library-refresh`, and `library-library-validation` specs continue to describe the same observable behavior; the AGENTS.md drift is internal documentation accuracy only.

## Impact

### Affected code

| File | Change | LOC impact |
|---|---|---|
| `cmd/AGENTS.md:278` | Doc table edit | 0 (table cell text) |
| `internal/library/AGENTS.md:71,353,358` | Doc text edits | 0 (table cells) |
| `internal/output/AGENTS.md:15` | Doc text edit | 0 (table cell) |
| `internal/warning/canary.go:44` | Canary string edit | 0 (constant change) |
| `internal/cmdutil/integration_test.go:1` | Build tag added | +1 (directive line) |
| `test/AGENTS.md:62-86` | Doc text edit | 0 (list update) |
| `cmd/presets.go:160` | Delete dead type | -3 to -8 |
| `cmd/show.go:225` | Delete dead type | -3 to -8 |

### Affected systems

- **No CLI behavior change.** All edits are documentation or dead-code removal.
- **No public API change.** The deleted JSON output types are unexported to external consumers (per `rg` verification: zero non-test callers).
- **Build:** adding `//go:build integration` to `internal/cmdutil/integration_test.go` means `go test ./...` (without the `integration` tag) will skip this file. The `mise` `test:integration` task must invoke the tag explicitly. This matches the existing pattern at `internal/parser/integration_test.go:1`.
- **Lint baseline:** expected unchanged (no production-code edits that affect `golangci-lint` output).

## Risks

- **Build tag migration:** removing `internal/cmdutil/integration_test.go` from the default `go test ./...` set may hide test failures during normal development. **Mitigation:** task 7.3 verifies `mise run test:integration` runs both integration files under the tag; task 7.2 confirms `mise run test` (without the tag) excludes `internal/cmdutil/integration_test.go` from the default target.
- **Deprecation ripple:** deleting the 4 unused JSON types may surface hidden callers in downstream branches (uncommon, but possible). **Mitigation:** task 6.1 runs `rg -e 'PresetsJSONOutput' -e 'PresetInfoJSON' -e 'ShowResourceJSONOutput' -e 'ShowPresetJSONOutput'` before deletion; zero matches expected outside the declarations. Task 7.5 re-runs the same rg after deletion and expects zero matches.
- **CHANGELOG cross-reference:** the canary string in `internal/warning/canary.go:44` is user-visible (printed on stderr when a user triggers the legacy exit code path). Changing the text is observable. **Mitigation:** the new text retains the user-meaningful portion ("exit code 5 was renamed to 1") and only replaces the "slice 2" navigation hint with a CHANGELOG reference; task 3.2 verifies that no test asserts on the literal "slice 2" substring (the canary test only checks for `"Warning: "`, so no test edit is expected).
