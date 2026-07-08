# library-library-resource-import Specification (delta)

## MODIFIED Requirements

### Requirement: Support dry-run mode

The system SHALL preview changes without modifying library in dry-run mode. The dry-run preview SHALL be written to the `stdout io.Writer` parameter passed to `AddResource` (typically `opts.IO.Out` from the cmd layer). The library package SHALL NOT write to `os.Stdout` directly; the cmd layer is responsible for selecting the writer.

**Change**: ADDED the "writer parameter" requirement. The pre-change implementation wrote dry-run output to `os.Stdout` directly, which polluted piped consumers' stdout.

#### Scenario: Dry-run shows what would happen

- **GIVEN** a library and valid source file
- **WHEN** AddResource is called with `--dry-run`
- **THEN** no files are modified
- **AND** the expected action is described via the `io.Writer` parameter (not directly to `os.Stdout`)

#### Scenario: Library package does not write to os.Stdout

- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** the dry-run output SHALL be written via the `io.Writer` parameter passed to `AddResource`
