# Design — Migrate library presets and library show

## Context

After changes 2 and 3 migrate the core domain commands and `library resources`, change-4 migrates the two remaining read-only library commands (`library presets` and `library show`). These commands have no `--json` legacy flag (they previously printed plain text only); the new `--output` flag is added fresh.

## Goals / Non-Goals

**Goals:**

- `cmd/library/presets.go` and `cmd/library/show.go` follow the `NewCmdXxx(f, runF) + runXxx(opts)` pattern.
- Both commands gain `--output json|table|plain` via `cmdutil.AddOutputFlags`.
- `library show` resolves refs correctly (resource or preset).
- `cmd/library_formatters.go` is deleted; formatters live in the per-command files.
- Output remains byte-identical for the default `plain` format.

**Non-Goals:**

- Migrating mutating library commands — changes 6, 7.
- Migrating `init` — change-5 (partial-success semantics).
- Restructuring the `Library` package — deferred to a future refactor change.

## Decisions

### 1. Each command declares its minimal `Library` interface

**Choice**: `presets.go` declares a `Library` interface with just `ListPresets(ctx) ([]library.Preset, error)`. `show.go` declares a `Library` interface with just `Resolve(ctx, ref string) (*library.Resource, error)` (covering both resource and preset refs).

**Rationale**: matches the foundation's `application/command-options-pattern` capability ("Accept interfaces, return structs"); lets each command depend only on the methods it calls.

### 2. Formatters move to per-command files

**Choice**: The helpers in `cmd/library_formatters.go` (329 lines) move into the per-command files as private functions or are replaced by calls to `output.Exporter` directly.

**Rationale**: avoids a shared `cmd/library_formatters.go` file that's hard to navigate; the formatters are small (5-20 LOC each) and easy to read in context.

**Alternatives considered**:

- Move formatters to `internal/output/` → rejected; they're library-specific, not generic output helpers.
- Keep `library_formatters.go` → rejected; it's the kind of shared file that the new architecture explicitly avoids.

### 3. `library show` ref resolution stays in the command file

**Choice**: The ref resolution logic (parses `type/name` or `preset/name`) moves from `cmd/library.go` into `cmd/library/show.go` as a private helper.

**Rationale**: the resolution logic is `show`-specific; encapsulating it in the command file makes the file self-contained.

## Risks / Trade-offs

- **Ref resolution edge cases** — unusual ref formats (e.g. `agent/` with no name, or `preset/` without `preset/` prefix) could be lost in the move. **Mitigation:** existing tests in `cmd/library_test.go` cover these cases; converted to `iostreams.Test()` + `runF` injection in task 4.2.4.
- **Formatter duplication** — duplicating simple formatters across command files is acceptable; complex formatting would warrant extraction. **Mitigation:** if duplication exceeds ~50 LOC per file, refactor in a follow-up change.
- **No new dependencies** — this change doesn't add any new packages or interfaces; everything reuses the foundation.
