# Tasks — Remove `cmdutil.AddOutputFlags` re-export

Each task ends with `mise run check` passing.

## 1. Swap call sites

- [x] 1.1 In `cmd/show.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.2 In `cmd/resources.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.3 In `cmd/library_validate.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.4 In `cmd/presets.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.5 In `cmd/library_refresh.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.6 In `cmd/library_init.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [x] 1.7 In `cmd/library_add.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.

## 2. Update library_remove comment

- [x] 2.1 In `cmd/library_remove.go:142-150`, replace the comment referencing `cmdutil.AddOutputFlags` with a brief note explaining that `output.AddOutputFlags` binds to `cmd.Flags()` (local flags) and `library remove` is a parent command needing `cmd.PersistentFlags()` (inherited by `resource` and `preset` sub-commands), hence the inline wiring.

## 3. Delete the re-export

- [x] 3.1 Delete `internal/cmdutil/output_flags.go`.

## 4. Update documentation

- [x] 4.1 In `internal/cmdutil/AGENTS.md`, remove the `output_flags.go` row from the Files table (line ~18) and the `AddOutputFlags` entry from the Key Surface (line ~31).
- [x] 4.2 In `internal/output/AGENTS.md:9`, remove the "via the `cmdutil.AddOutputFlags` re-export so command files import only `cmdutil`" qualifier from the description.
- [x] 4.3 In `cmd/AGENTS.md:280`, delete the `cmdutil.AddOutputFlags` row from the Foundation Units table. The `output.AddOutputFlags` row on line 277 already documents the canonical path.
- [x] 4.4 In `openspec/specs/cli-output-formats/spec.md:5`, update the capability purpose statement — change the parenthetical reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags`. (Out of OpenSpec delta scope; tracked here as a manual edit during sync so it is not forgotten.)
- [x] 4.5 In `AGENTS.md:204` (root), change "via `cmdutil.AddOutputFlags`" to "via `output.AddOutputFlags`" so the global `--output` flag description points at the canonical helper.
- [x] 4.6 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-validation/spec.md` rehoming the `library validate supports --output flag` requirement body (`cmdutil.AddOutputFlags` → `output.AddOutputFlags`).
- [x] 4.7 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-scaffolding/spec.md` rehoming the `library init supports --output flag` requirement body.
- [x] 4.8 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-remove-resource/spec.md` rehoming the `library remove resource supports --output flag` requirement body.
- [x] 4.9 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-remove-preset/spec.md` rehoming the `library remove preset supports --output flag` requirement body (including the inherited-from-parent note about `cmd.PersistentFlags()` wiring).
- [x] 4.10 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-refresh/spec.md` rehoming the `library refresh supports --output flag` requirement body.
- [x] 4.11 Add `MODIFIED Requirements` delta at `openspec/changes/remove-cmdutil-output-reexport/specs/library-library-json-output/spec.md` rehoming the three `library {presets,add,show} supports --output flag` requirement bodies. (Plus a manual sync edit at line 5 of the base spec for the purpose statement reference to `cmdutil.AddOutputFlags`.)

## 5. Verification

- [x] 5.1 Run `rg "cmdutil\.AddOutputFlags" .` — must return zero matches (the function is deleted). Verified: matches remain only in `openspec/changes/archive/**` (historical, by design), `openspec/changes/remove-cmdutil-output-reexport/**` (the change's own artifacts and delta files — change notes and a scenario deliberately name the deleted symbol to describe the rehome), and `openspec/specs/**` base specs (13 matches across 5 spec files: cli-output-formats, library-library-validation, library-library-scaffolding, library-library-remove-resource, library-library-remove-preset, library-library-refresh, library-library-json-output — all covered by `MODIFIED Requirements` deltas in this change and applied by `osc-archive-change` at sync time, per OpenSpec's delta workflow).
- [x] 5.2 Run `mise run build` — confirms no broken imports. **PASS** (`Finished in 561.7ms`).
- [x] 5.3 Run `mise run lint` — `0 issues.` No drift from baseline; the comment update at §2 did not shift lint output. Lint baseline test (`cmd/lint_test.go::TestLintBaseline`) re-runs lint 8× and passes against `cmd/testdata/lint_baseline.txt` without refresh.
- [x] 5.4 Run `mise run test` — all unit/integration/golden tests pass across all 17 packages.
- [x] 5.5 Run `mise run test:e2e` — `test/e2e` passes (`3.588s`); `--output json|table|plain` scenarios for `library {resources,presets,show,add,refresh,remove,validate}` are green.
- [x] 5.6 Run `mise run check` — full validation (build + lint + unit + integration + golden + E2E) passes (`Finished in 12.29s`).
