# Tasks — Reconcile AGENTS.md drift and dead code

Each task ends with `mise run check` passing. Tasks are grouped by the 4 logical themes (signatures, dispatch sets, build tags, dead code) and ordered so each commit is independently testable.

## 1. Reconcile signature table cells

- [x] 1.1 In `cmd/AGENTS.md:278`, update the Foundation Units table cell for `cmdutil.Factory` / `NewFactory` from `NewFactory(io, ver, exe)` to `NewFactory(ctx, io, appVersion, executable)` to match `internal/cmdutil/factory.go:46`. Context-first ordering follows the `golang-cli-architecture` skill; parameter names match the actual declaration.
- [x] 1.2 In `internal/library/AGENTS.md:71`, update the `FindLibrary` row in the discovery function table to `FindLibrary(flagPath, envPath string) string` to match `internal/library/discovery.go:14`.
- [x] 1.3 In `internal/library/AGENTS.md:353`, rename the `RefreshSkipped` row to `SkipInfo` to match `internal/library/refresher.go:43` (the actual type name).
- [x] 1.4 In `internal/library/AGENTS.md:358`, update the `RefreshError` struct definition to `type RefreshError struct { Ref string; Field string; Type string }` to match `internal/library/refresher.go:49-53`.

## 2. Reconcile FormatError dispatch set

- [x] 2.1 In `internal/output/AGENTS.md:15`, update the `FormatError` dispatch set to: `Parse, Validation, Transform, File, Config, NotFound, PartialSuccess, Operation` to match `internal/output/errors.go:21-50`.

## 3. Update canary string

- [x] 3.1 In `internal/warning/canary.go:44`, replace the canary message "exit code 5 was renamed to 1 in slice 2; consult CHANGELOG for the migration timeline" with "exit code 5 was renamed to 1; see CHANGELOG.md for the migration timeline" — drops the unfindable "slice 2" reference while preserving the user-meaningful portion.
- [x] 3.2 Verify no test asserts on the literal "slice 2" substring (the canary test only checks for `"Warning: "`, so no test edit is expected). Run `rg -e 'slice (2|[0-9])' internal/warning/` before editing to find all references.

## 4. Add missing build tag

- [x] 4.1 In `internal/cmdutil/integration_test.go:1`, prepend the `//go:build integration` directive line followed by a blank line, matching the format at `internal/parser/integration_test.go:1`.

## 5. Update test/AGENTS.md integration file list

- [x] 5.1 In `test/AGENTS.md:62-86`, add `internal/cmdutil/integration_test.go` to the integration-test file list alongside `internal/parser/integration_test.go`.

## 6. Delete unused JSON output types

- [x] 6.1 Run `rg -e 'PresetsJSONOutput' -e 'PresetInfoJSON' -e 'ShowResourceJSONOutput' -e 'ShowPresetJSONOutput' ./cmd ./internal ./test` — must return exactly 8 matches, the 4 type declarations and the 4 type-name references from other struct fields at `cmd/presets.go:160-168` and `cmd/show.go:225-234`; zero other lines.
- [x] 6.2 In `cmd/presets.go:160-168`, delete the `PresetsJSONOutput` (L160) and `PresetInfoJSON` (L163) type declarations.
- [x] 6.3 In `cmd/show.go:225-234`, delete the `ShowResourceJSONOutput` (L225) and `ShowPresetJSONOutput` (L228) type declarations.
- [x] 6.4 Run `mise run build` — confirms no broken references.

## 7. Verification

- [x] 7.1 Run `mise run lint` — must report 0 issues (no production-code edits that affect lint).
- [x] 7.2 Run `mise run test` — all unit tests pass; `internal/cmdutil/integration_test.go` is excluded from the default target.
- [x] 7.3 Run `mise run test:integration` (or `go test -tags=integration ./...`) — both `internal/parser/integration_test.go` and `internal/cmdutil/integration_test.go` run under the tag.
- [x] 7.4 Run `mise run test:e2e` — E2E tests pass.
- [x] 7.5 Run `rg -e 'PresetsJSONOutput' -e 'PresetInfoJSON' -e 'ShowResourceJSONOutput' -e 'ShowPresetJSONOutput' ./cmd ./internal ./test` — must return zero matches.
- [x] 7.6 Run `rg -e 'slice (2|[0-9])' internal/warning/` — must return zero matches.
- [x] 7.7 Run `openspec validate reconcile-agents-md-drift --strict` — change is coherent.
- [x] 7.8 Manually grep each drift entry to confirm the fix:
  - `rg -n 'NewFactory\(' cmd/AGENTS.md internal/cmdutil/factory.go` — both should now match `NewFactory(ctx, io, appVersion, executable)`.
  - `rg -n 'FindLibrary' internal/library/AGENTS.md internal/library/discovery.go` — both should now match `FindLibrary(flagPath, envPath string) string`.
  - `rg -n 'RefreshError' internal/library/AGENTS.md internal/library/refresher.go` — both should now match `RefreshError struct { Ref, Field, Type }`.

## 8. Archive

- [x] 8.1 Archive this change via `osc-archive-change reconcile-agents-md-drift`.
- [x] 8.2 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.
