# Remove `cmdutil.AddOutputFlags` re-export

## Why

The re-export at `internal/cmdutil/output_flags.go:9-13` covers exactly **1** of the **7** `output` symbols consumed by command files (`FormatError`, `NewJSONExporter`, `NewTableExporter`, `FormatResourcesList`, `DefaultOutputFormat`, `ValidOutputFormats`, and `AddOutputFlags` itself). Every cmd file that uses `--output` already imports `internal/output` for the other symbols, so the re-export provides no convenience — it only creates a second import path to the same function (`cmdutil.AddOutputFlags` vs `output.AddOutputFlags`) and adds a layer of indirection with no type safety.

`cmd/library_remove.go:142-150` actually **bypasses both** `cmdutil.AddOutputFlags` and `output.AddOutputFlags` because it needs `PersistentFlags()` inheritance — the re-export provides no value here either. The premise of the slice-1 rationale ("command files import only `cmdutil`") no longer holds: the package layout is `cmd → {cmdutil, output, core, library, parser, renderer, …}`, not `cmd → cmdutil → output`.

The corresponding base spec (`openspec/specs/cli-output-formats/spec.md`) references `cmdutil.AddOutputFlags` in three places (the `AddOutputFlags helper` requirement, the `AddOutputFlags is opt-in per command` requirement, and the capability purpose statement). After the rehome, those references must point to `output.AddOutputFlags`. A new requirement is added documenting the `PersistentFlags()` limitation so future contributors do not re-introduce the re-export.

## What Changes

- **DELETE** `internal/cmdutil/output_flags.go` (~6 LOC; 13 lines including comments and blanks)
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
  - `internal/output/AGENTS.md:9` — remove the "via the `cmdutil.AddOutputFlags` re-export so command files import only `cmdutil`" qualifier
  - `cmd/AGENTS.md:280` — remove the `cmdutil.AddOutputFlags` row from the Foundation Units table (or repoint it to `output.AddOutputFlags`)
- **MODIFY** spec: `openspec/changes/remove-cmdutil-output-reexport/specs/cli-output-formats/spec.md` — `MODIFIED Requirements`: rehome `AddOutputFlags helper` body and rehome `AddOutputFlags is opt-in per command` body (both `cmdutil.AddOutputFlags` → `output.AddOutputFlags`); `ADDED Requirements`: document the PersistentFlags wiring contract for parent commands.
- **MODIFY** base spec capability purpose statement (manual edit during sync): `openspec/specs/cli-output-formats/spec.md:5` — update the parenthetical reference from `cmdutil.AddOutputFlags` to `output.AddOutputFlags` (out of OpenSpec delta scope; tracked here so it is not forgotten).

## Capabilities

### Modified Capabilities

- `cli-output-formats` — rehome `AddOutputFlags` reference in the `AddOutputFlags helper` requirement; rehome the same reference in the `AddOutputFlags is opt-in per command` requirement; add a new requirement codifying the `PersistentFlags()` wiring contract for parent commands.

## Impact

### Affected code

- **Deleted (1 file):** `internal/cmdutil/output_flags.go`
- **Modified (7 files):** `cmd/show.go`, `cmd/resources.go`, `cmd/library_validate.go`, `cmd/presets.go`, `cmd/library_refresh.go`, `cmd/library_init.go`, `cmd/library_add.go`
- **Modified (1 file):** `cmd/library_remove.go` (comment update)
- **Modified (3 files):** `internal/cmdutil/AGENTS.md`, `internal/output/AGENTS.md`, `cmd/AGENTS.md`

### Affected systems

- **Import graph:** `cmd → internal/output` becomes the canonical path for `--output` flag wiring. `cmdutil` no longer re-exports `output` symbols.
- **Public API:** unchanged. `internal/cmdutil.AddOutputFlags` is an internal re-export and has no external consumers.
- **Spec contract:** the new PersistentFlags requirement codifies a limitation of `output.AddOutputFlags` (local `Flags()` binding only). Parent commands (e.g., `library remove`) must wire `--output` via `cmd.PersistentFlags()` manually; future contributors must not extract a helper that abstracts over the two flag-set bindings.
- **Lint baseline:** if the comment update in `cmd/library_remove.go` shifts lint output, refresh `cmd/testdata/lint_baseline.txt` (the baseline test in `cmd/lint_test.go` will catch any drift).

## Risks

- **None significant.** The change is mechanical (search/replace + AGENTS.md edits). The re-export has zero external consumers (it's an `internal/` package re-export) and the 7 call sites all already import `internal/output`. The only failure mode is a missed call site, which `mise run build` will catch immediately.
- **Lint baseline drift:** the comment update in `cmd/library_remove.go` may shift `golangci-lint` output. **Mitigation:** task `5.3` runs `mise run lint` after the swap; if output shifts, refresh `cmd/testdata/lint_baseline.txt` per the procedure in `cmd/AGENTS.md` "Lint Baseline Test" section.
