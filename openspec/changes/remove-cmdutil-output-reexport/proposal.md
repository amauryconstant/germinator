# Remove `cmdutil.AddOutputFlags` re-export

## Why

The re-export at `internal/cmdutil/output_flags.go:9-13` covers exactly **1** of the **~6** `output` symbols consumed by command files (`FormatError`, `NewJSONExporter`, `NewTableExporter`, `FormatResourcesList`, `DefaultOutputFormat`, `ValidOutputFormats`). Every cmd file that uses `--output` already imports `internal/output` for the other symbols, so the re-export provides no convenience — it only creates a second import path to the same function (`cmdutil.AddOutputFlags` vs `output.AddOutputFlags`) and adds a layer of indirection with no type safety.

`cmd/library_remove.go:142-150` actually **bypasses both** `cmdutil.AddOutputFlags` and `output.AddOutputFlags` because it needs `PersistentFlags()` inheritance — the re-export provides no value here either. The premise of the slice-1 rationale ("command files import only `cmdutil`") no longer holds: the package layout is `cmd → {cmdutil, output, core, library, parser, renderer, …}`, not `cmd → cmdutil → output`.

## What Changes

- **DELETE** `internal/cmdutil/output_flags.go` (13 LOC)
- **MODIFY** 7 cmd files: replace `cmdutil.AddOutputFlags` calls with `output.AddOutputFlags` calls (and add `internal/output` to the import list where missing):
  - `cmd/show.go:98`
  - `cmd/resources.go:89`
  - `cmd/library_validate.go:117`
  - `cmd/presets.go:79`
  - `cmd/library_refresh.go:121`
  - `cmd/library_init.go:125`
  - `cmd/library_add.go:212`
- **MODIFY** `cmd/library_remove.go:142-150` comment: replace the "mirrors what `cmdutil.AddOutputFlags` would register" phrase with a brief note that `output.AddOutputFlags` binds to `cmd.Flags()` (local flags) and library subcommands need `cmd.PersistentFlags()` instead, hence the inline wiring.
- **MODIFY** documentation:
  - `internal/cmdutil/AGENTS.md:18,31` — remove the `output_flags.go` row from the Files table and the `AddOutputFlags` entry from the Key Surface
  - `internal/output/AGENTS.md:5` — remove the "via the `cmdutil.AddOutputFlags` re-export so command files import only `cmdutil`" qualifier
  - `cmd/AGENTS.md:280` — remove the `cmdutil.AddOutputFlags` row from the Foundation Units table (or repoint it to `output.AddOutputFlags`)

## Capabilities

### Modified Capabilities

None. This change is a refactor of internal import paths only — no spec-level behavior changes. User-facing CLI surface, flag wiring, and output formatting are unchanged.

## Impact

### Affected code

- **Deleted (1 file):** `internal/cmdutil/output_flags.go`
- **Modified (7 files):** `cmd/show.go`, `cmd/resources.go`, `cmd/library_validate.go`, `cmd/presets.go`, `cmd/library_refresh.go`, `cmd/library_init.go`, `cmd/library_add.go`
- **Modified (1 file):** `cmd/library_remove.go` (comment update)
- **Modified (3 files):** `internal/cmdutil/AGENTS.md`, `internal/output/AGENTS.md`, `cmd/AGENTS.md`

### Affected systems

- **Import graph:** `cmd → internal/output` becomes the canonical path for `--output` flag wiring. `cmdutil` no longer re-exports `output` symbols.
- **Public API:** unchanged. `internal/cmdutil.AddOutputFlags` is an internal re-export and has no external consumers.
- **Lint baseline:** if the comment update in `cmd/library_remove.go` shifts lint output, refresh `cmd/testdata/lint_baseline.txt` (the baseline test in `cmd/lint_test.go` will catch any drift).

## Risks

- **None significant.** The change is mechanical (search/replace + AGENTS.md edits). The re-export has zero external consumers (it's an `internal/` package re-export) and the 7 call sites all already import `internal/output`. The only failure mode is a missed call site, which `mise run build` will catch immediately.
- **Lint baseline drift:** the comment update in `cmd/library_remove.go` may shift `golangci-lint` output. **Mitigation:** task `1.4` runs `mise run lint` after the swap; if output shifts, refresh `cmd/testdata/lint_baseline.txt` per the procedure in `cmd/AGENTS.md` "Lint Baseline Test" section.
