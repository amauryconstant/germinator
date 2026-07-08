# Design — Remove `cmdutil.AddOutputFlags` re-export

## Context

The current state has a half-applied re-export at `internal/cmdutil/output_flags.go:9-13`:

```go
// internal/cmdutil/output_flags.go
package cmdutil

import "gitlab.com/amoconst/germinator/internal/output"

func AddOutputFlags(cmd *cobra.Command, target *string) {
    output.AddOutputFlags(cmd, target)
}
```

This was introduced in slice 1 (`openspec/changes/archive/2026-06-24-scaffold-cli-foundation/design.md:126-134`) with the rationale:

> *"The implementation lives in `internal/output/output_flags.go` (with the per-command wiring knowledge). `cmdutil/output_flags.go` re-exports it as `cmdutil.AddOutputFlags` so command files import only `cmdutil`."*

The premise — that cmd files would need only `AddOutputFlags` from `output` — no longer holds. Today, the cmd files that use `--output` consume **6** symbols from `internal/output` (verified via grep):

| Symbol | Number of cmd files importing |
|---|---|
| `FormatError` | 7 |
| `NewJSONExporter` | 5 |
| `NewTableExporter` | 5 |
| `FormatResourcesList` | 3 |
| `DefaultOutputFormat` | 2 |
| `ValidOutputFormats` | 1 |
| `AddOutputFlags` (now via re-export) | 7 |

Every cmd file that uses the re-export **already imports `internal/output`** for the other symbols. The re-export provides no convenience, only an indirection layer.

Furthermore, `cmd/library_remove.go:142-150` cannot use either `cmdutil.AddOutputFlags` or `output.AddOutputFlags` because library sub-commands need `PersistentFlags()` (inherited by descendants) rather than `Flags()` (the `output.AddOutputFlags` implementation uses `cmd.Flags()`). The author re-implemented the wiring inline. This is the strongest evidence the re-export's premise no longer holds: a sibling command couldn't use it.

Constraints:
- **No spec deltas:** this change has no spec-level impact. Existing `cli-output-formats` and `cli-shell-completion` specs describe user-facing behavior that is unchanged.
- **Lint baseline test:** `cmd/lint_test.go` runs `mise run lint` and diffs against `cmd/testdata/lint_baseline.txt`. Any shift in lint output requires a baseline refresh.
- **No public API break:** `internal/cmdutil.AddOutputFlags` is an internal re-export with zero external consumers.

## Goals / Non-Goals

**Goals:**

- Delete `internal/cmdutil/output_flags.go` (the re-export wrapper).
- Replace all 7 in-tree call sites with `output.AddOutputFlags`.
- Update the 3 affected AGENTS.md files so future contributors find the canonical path.
- Preserve all existing CLI behavior (flag registration, output formats, exit codes).

**Non-Goals:**

- Refactoring `internal/output.AddOutputFlags` itself (e.g., to accept a flag-set selector). Out of scope; `cmd/library_remove.go` keeps its inline wiring with an updated comment.
- Removing `internal/output` re-exports in general. The other 5 symbols (`FormatError`, etc.) have single-package owners and need no consolidation.
- Touching the `cli-output-formats` or `cli-shell-completion` specs. No requirement changes.

## Decisions

### 1. Delete the re-export rather than expand it

**Choice**: Delete `internal/cmdutil/output_flags.go` rather than expanding it to re-export all 6 `output` symbols.

**Rationale**: A complete re-export of `output` symbols into `cmdutil` would invert the dependency (cmdutil → output for types, but cmd → cmdutil for everything else, creating two import paths to the same functions). The clean architecture is the existing one: `cmd → output` directly. The re-export was a leftover from slice 1's "command files import only `cmdutil`" assumption, which the rest of the codebase has outgrown.

**Alternatives considered**:
- *Expand the re-export* to cover all `output` symbols: rejected because it adds a complete parallel API surface with no type safety; the same drift problem recurs.
- *Keep `cmdutil.AddOutputFlags` but rename it to `cmdutil.RegisterOutputFlags`* (an indirection that documents the limitation): rejected because it preserves the duplicate import path; the goal is to remove it.

### 2. Use `output.AddOutputFlags` directly in all 7 cmd files

**Choice**: All 7 call sites call `output.AddOutputFlags` directly. Every file already imports `internal/output`, so no new imports are needed.

**Rationale**: One canonical import path per symbol is the Go convention. The current 7 sites mix `cmdutil.AddOutputFlags` (re-export) with `output.FormatError` (canonical), which is inconsistent. After the swap, every `output.*` symbol is reached via the `output` import path.

**Alternatives considered**:
- *Introduce a `cmdutil.AddOutputFlagsPersistent(cmd)` variant for `library_remove.go`*: rejected; the inline wiring is 8 lines and is the only place that needs it. Adding a public function for one call site is over-engineering.

### 3. Update `cmd/library_remove.go:142-150` comment

**Choice**: Replace the comment "mirrors what `cmdutil.AddOutputFlags` would register" with a brief note explaining why `library_remove` inlines the wiring (`PersistentFlags()` requirement, not local `Flags()`).

**Rationale**: Without the re-export, the old comment becomes confusing (it referenced a symbol that's gone). The new comment documents the limitation of `output.AddOutputFlags` (local flags only) so future contributors don't try to "fix" the inline wiring by extracting a helper.

**Alternatives considered**:
- *Leave the comment unchanged*: rejected; it would reference a non-existent `cmdutil.AddOutputFlags` and confuse future readers.

### 4. No spec changes

**Choice**: Leave `cli-output-formats/spec.md` and `cli-shell-completion/spec.md` unchanged. No delta files.

**Rationale**: The change is internal refactoring. The user-facing behavior — `germinator library resources --output json` works as before — is unchanged. Specs describe behavior, not implementation; this is an implementation detail.

**Alternatives considered**:
- *Add a spec note about the import path*: rejected; OpenSpec specs describe WHAT the system does, not HOW the code is organized.

## Risks / Trade-offs

- **Missed call site**: a future contributor adds `cmdutil.AddOutputFlags` in a new file. **Mitigation**: the function is deleted, so the build fails immediately (`undefined: cmdutil.AddOutputFlags`). The error message names the canonical replacement (`output.AddOutputFlags`).
- **Lint baseline drift**: the comment update in `cmd/library_remove.go` may shift lint output. **Mitigation**: task `1.4` runs `mise run lint` after the swap and refreshes `cmd/testdata/lint_baseline.txt` if output shifts (per `cmd/AGENTS.md` "Lint Baseline Test" procedure).
- **Test breakage from import swap**: `cmd/show_test.go` and `cmd/resources_test.go` import `output.FormatResourcesList`; the swap doesn't touch those imports, but verify `mise run test` passes post-change. **Mitigation**: task `1.5` runs the full test suite.
- **Two import paths remain temporarily**: between the swap and the lint baseline refresh, a developer could commit a half-applied change. **Mitigation**: tasks `1.2-1.3` are atomic — each call site is updated in one commit with no intermediate state merged.
