# Tasks — Reconcile AGENTS.md drift and dead code

Each task ends with `mise run check` passing. Tasks are grouped by the 4 logical themes (signatures, dispatch sets, build tags, dead code) and ordered so each commit is independently testable.

## 1. Reconcile signature table cells

- [ ] 1.1 In `cmd/AGENTS.md:278`, update the Foundation Units table cell for `cmdutil.Factory` / `NewFactory` from `NewFactory(io, ver, exe)` to `NewFactory(ctx, io, appVersion, executable)` to match `internal/cmdutil/factory.go:46`.
- [ ] 1.2 In `internal/library/AGENTS.md:71`, update the `FindLibrary` row in the discovery function table to `FindLibrary(flagPath, envPath string) string` to match `internal/library/discovery.go:14`.
- [ ] 1.3 In `internal/library/AGENTS.md:353`, rename the `RefreshSkipped` row to `SkipInfo` to match `internal/library/refresher.go:43` (the actual type name).
- [ ] 1.4 In `internal/library/AGENTS.md:358`, update the `RefreshError` struct definition to `type RefreshError struct { Ref string; Field string; Type string }` to match `internal/library/refresher.go:49-53`.

## 2. Reconcile FormatError dispatch set

- [ ] 2.1 In `internal/output/AGENTS.md:15`, update the `FormatError` dispatch set to: `Parse, Validation, Transform, File, Config, NotFound, PartialSuccess, Operation, Initialize` to match `internal/output/errors.go:21-50`.

## 3. Update canary string

- [ ] 3.1 In `internal/warning/canary.go:44`, replace the canary message "exit code 5 was renamed to 1 in slice 2; consult CHANGELOG for the migration timeline" with "exit code 5 was renamed to 1; see CHANGELOG.md for the migration timeline" — drops the unfindable "slice 2" reference while preserving the user-meaningful portion.
- [ ] 3.2 In `internal/warning/canary_test.go`, update any string-assertion tests to match the new canary text. Run `rg "slice 2\|slice [0-9]" internal/warning/` before editing to find all references.

## 4. Add missing build tag

- [ ] 4.1 In `internal/cmdutil/integration_test.go:1`, prepend `//go:build integration\n\n` (with a blank line) to match the format at `internal/parser/integration_test.go:1`.
- [ ] 4.2 Run `go test ./...` — confirm the file is now skipped from the default test target.
- [ ] 4.3 Run `go test -tags=integration ./...` — confirm the file still runs under the tag.

## 5. Update test/AGENTS.md integration file list

- [ ] 5.1 In `test/AGENTS.md:62-86`, add `internal/cmdutil/integration_test.go` to the integration-test file list alongside `internal/parser/integration_test.go`.

## 6. Delete unused JSON output types

- [ ] 6.1 Run `rg "PresetsJSONOutput|PresetInfoJSON|ShowResourceJSONOutput|ShowPresetJSONOutput"` — must return zero non-test callers outside the declarations.
- [ ] 6.2 In `cmd/presets.go:160`, delete the `PresetsJSONOutput` and `PresetInfoJSON` type declarations.
- [ ] 6.3 In `cmd/show.go:225`, delete the `ShowResourceJSONOutput` and `ShowPresetJSONOutput` type declarations.
- [ ] 6.4 Run `mise run build` — confirms no broken references.

## 7. Verification

- [ ] 7.1 Run `mise run lint` — must report 0 issues (no production-code edits that affect lint).
- [ ] 7.2 Run `mise run test` — all unit tests pass; `integration_test.go` is excluded.
- [ ] 7.3 Run `mise run test:integration` (or `go test -tags=integration ./...`) — both integration files run.
- [ ] 7.4 Run `mise run test:e2e` — E2E tests pass.
- [ ] 7.5 Run `rg "PresetsJSONOutput|PresetInfoJSON|ShowResourceJSONOutput|ShowPresetJSONOutput" .` — must return zero matches.
- [ ] 7.6 Run `rg "slice 2\|slice [0-9]" internal/warning/` — must return zero matches.
- [ ] 7.7 Run `openspec validate reconcile-agents-md-drift --strict` — change is coherent.
- [ ] 7.8 Manually grep each drift entry to confirm the fix:
  - `grep -n "NewFactory(" cmd/AGENTS.md internal/cmdutil/factory.go` — both should now match `NewFactory(ctx, io, appVersion, executable)`.
  - `grep -n "FindLibrary" internal/library/AGENTS.md internal/library/discovery.go` — both should now match `FindLibrary(flagPath, envPath string) string`.
  - `grep -n "RefreshError" internal/library/AGENTS.md internal/library/refresher.go` — both should now match `RefreshError struct { Ref, Field, Type }`.

## 8. Archive

- [ ] 8.1 Archive this change via `osc-archive-change reconcile-agents-md-drift`.
- [ ] 8.2 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`.
