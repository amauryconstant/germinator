# Tasks — Remove `cmdutil.AddOutputFlags` re-export

Each task ends with `mise run check` passing.

## 1. Swap call sites

- [ ] 1.1 In `cmd/show.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.2 In `cmd/resources.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.3 In `cmd/library_validate.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.4 In `cmd/presets.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.5 In `cmd/library_refresh.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.6 In `cmd/library_init.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.
- [ ] 1.7 In `cmd/library_add.go`, change `cmdutil.AddOutputFlags(cmd, &outputFormat)` to `output.AddOutputFlags(cmd, &outputFormat)`. Verify `internal/output` is already imported.

## 2. Update library_remove comment

- [ ] 2.1 In `cmd/library_remove.go:142-150`, replace the comment referencing `cmdutil.AddOutputFlags` with a brief note explaining that `output.AddOutputFlags` binds to `cmd.Flags()` (local flags) and `library remove` is a parent command needing `cmd.PersistentFlags()` (inherited by `resource` and `preset` sub-commands), hence the inline wiring.

## 3. Delete the re-export

- [ ] 3.1 Delete `internal/cmdutil/output_flags.go`.

## 4. Update documentation

- [ ] 4.1 In `internal/cmdutil/AGENTS.md`, remove the `output_flags.go` row from the Files table (line ~18) and the `AddOutputFlags` entry from the Key Surface (line ~31).
- [ ] 4.2 In `internal/output/AGENTS.md:5`, remove the "via the `cmdutil.AddOutputFlags` re-export so command files import only `cmdutil`" qualifier from the description.
- [ ] 4.3 In `cmd/AGENTS.md:280`, update the Foundation Units table — either repoint the row from `cmdutil.AddOutputFlags` to `output.AddOutputFlags` (and the package from `cmdutil` to `output`), or remove the row entirely if `output.AddOutputFlags` is already mentioned elsewhere in the docs.

## 5. Verification

- [ ] 5.1 Run `rg "cmdutil\.AddOutputFlags" .` — must return zero matches (the function is deleted).
- [ ] 5.2 Run `mise run build` — confirms no broken imports.
- [ ] 5.3 Run `mise run lint` — if output shifts (e.g., the comment update in `cmd/library_remove.go`), refresh `cmd/testdata/lint_baseline.txt` via `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` per the procedure in `cmd/AGENTS.md` "Lint Baseline Test" section.
- [ ] 5.4 Run `mise run test` — confirm all unit tests pass.
- [ ] 5.5 Run `mise run test:e2e` — confirm all E2E tests pass (especially the `--output json|table|plain` scenarios for `library resources`, `library presets`, `library show`, `library add`, `library refresh`, `library remove`, `library validate`).
- [ ] 5.6 Run `mise run check` — full validation passes.
