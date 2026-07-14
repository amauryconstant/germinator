# library-library-batch-add Specification (delta)

## MODIFIED Requirements

### Requirement: Writer field on BatchAddOptions (forward groundwork)

The system SHALL expose a `Stdout io.Writer` field on `BatchAddOptions` (populated by the cmd layer from `opts.IO.Out`). This wires the writer discipline into batch operations so that any future batch dry-run output (success listings, failure listings, summaries) respects the cmd-side I/O discipline. The library package SHALL NOT write to `os.Stdout` directly.

**Change**: ADDED the "writer field" requirement. Pre-change behavior: `BatchAddResources` (body starts at `adder.go:558`) returned its result via `*BatchAddResult` and did not write to stdout — there is no existing `os.Stdout` call site to replace. This change is forward groundwork; future batch dry-run output (when added) will respect the cmd-side I/O discipline by writing via the writer field rather than `os.Stdout`.

#### Scenario: Library package does not write to os.Stdout in batch

- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** any future batch dry-run output SHALL be written via the `Stdout io.Writer` field on `BatchAddOptions` (gated on `opts.Stdout != nil`)

#### Scenario: Cmd layer populates the writer field

- **GIVEN** the cmd layer (`cmd/library_add.go`) calls `BatchAddResources`
- **WHEN** the call site constructs `library.BatchAddOptions{...}`
- **THEN** the literal SHALL include `Stdout: opts.IO.Out`
- **AND** tests MAY pass `nil` (no batch dry-run output is written today; the field is forward-only)
