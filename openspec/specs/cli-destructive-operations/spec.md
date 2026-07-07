# cli-destructive-operations Specification

## Purpose

Define the shared `confirmOrFlag` helper for destructive operations: TTY-gated confirmation prompts that can be bypassed by a `--force` flag, ensuring every interactive command is fully scriptable.

## Requirements

### Requirement: confirmOrFlag helper

The `cmdutil` package SHALL provide a `confirmOrFlag(streams *iostreams.IOStreams, force bool, message string) (bool, error)` helper that resolves a destructive-operation confirmation via three-tier precedence:

1. `force == true` → return `(true, nil)` immediately
2. `streams.IsInteractive()` → print `message` to `ErrOut`; return `(confirmed, nil)` based on user input
3. Neither (`force == false` AND stdin not a TTY) → return an error explaining `--force` is required

#### Scenario: --force bypasses the prompt

- **GIVEN** an `IOStreams` with `IsInteractive() == true`
- **WHEN** `confirmOrFlag(io, true, "delete foo?")` is called
- **THEN** it SHALL return `(true, nil)` without printing the prompt

#### Scenario: Interactive confirmation

- **GIVEN** an `IOStreams` with `IsInteractive() == true`
- **WHEN** `confirmOrFlag(io, false, "delete foo?")` is called
- **AND** the user types `y` and presses Enter
- **THEN** it SHALL return `(true, nil)`

#### Scenario: Interactive decline

- **GIVEN** an `IOStreams` with `IsInteractive() == true`
- **WHEN** `confirmOrFlag(io, false, "delete foo?")` is called
- **AND** the user types `n` and presses Enter
- **THEN** it SHALL return `(false, nil)`
- **AND** the caller SHALL treat the operation as aborted (no side effect)

#### Scenario: Non-interactive without --force

- **GIVEN** an `IOStreams` with `IsInteractive() == false`
- **WHEN** `confirmOrFlag(io, false, "delete foo?")` is called
- **THEN** it SHALL return an error of the form: `<message> (use --force to skip confirmation)`
- **AND** the error SHALL be typed (e.g., `*core.ValidationError` or `*core.FileError`) so `cmdutil.ExitCodeFor` maps it correctly

### Requirement: Library init uses confirmOrFlag

`germinator library init` SHALL use `confirmOrFlag` when overwriting an existing library directory.

#### Scenario: library init --force overwrites

- **GIVEN** a library directory already exists at the target path
- **WHEN** `germinator library init --force --path /existing` is invoked
- **THEN** the existing directory SHALL be overwritten (no prompt)

#### Scenario: library init prompts on TTY

- **GIVEN** a library directory already exists at the target path
- **AND** stdin AND stdout are TTYs
- **WHEN** `germinator library init --path /existing` is invoked without `--force`
- **THEN** the CLI SHALL print `Overwrite existing library at /existing?` to `ErrOut`
- **AND** the overwrite SHALL only proceed on user confirmation

### Requirement: Init uses confirmOrFlag for existing files

`germinator init` SHALL use `confirmOrFlag` per-resource when `--force` is not set and an existing output file would be overwritten (see `cli-init-command/spec.md` for the per-file resolution order).

#### Scenario: init prompts per file when TTY

- **GIVEN** the user runs `germinator init --platform opencode --resources skill/foo,skill/bar --force=false`
- **AND** `skill/foo.md` already exists at the target output path
- **AND** `skill/bar.md` does not exist
- **AND** stdin AND stdout are TTYs
- **WHEN** `runInit` processes the two resources
- **THEN** the CLI SHALL prompt before overwriting `skill/foo.md`
- **AND** `skill/bar.md` SHALL be written without prompting (no existing file)
- **AND** partial-success semantics SHALL apply across the two resources

### Requirement: Undo hints after destructive ops

After a destructive operation succeeds, the CLI SHALL print an "Undo:" hint on `ErrOut` when an undo path exists (e.g., a snapshot, a git reflog entry, a CLI inverse command).

#### Scenario: Library remove prints undo hint

- **GIVEN** a resource `skill/foo` is removed via `germinator library remove resource skill/foo`
- **WHEN** the remove succeeds
- **THEN** the CLI SHALL print `Deleted skill/foo. To restore, re-add from <source>` on `ErrOut`

#### Scenario: Library init overwrite prints undo hint

- **GIVEN** `germinator library init --force` overwrote an existing library
- **WHEN** the operation succeeds
- **THEN** the CLI SHALL print `Overwrote <path>. No automatic undo — restore from backup.`
