# library-library-resource-import Specification (delta)

## MODIFIED Requirements

### Requirement: Support dry-run mode

The system SHALL preview changes without modifying library in dry-run mode. The dry-run preview SHALL be written to the `Stdout io.Writer` field on `AddRequest` (populated by the cmd layer from `opts.IO.Out`). The library package SHALL NOT write to `os.Stdout` directly; the cmd layer is responsible for selecting the writer.

**Change**: ADDED the "writer field" requirement. The pre-change implementation wrote dry-run output to `os.Stdout` directly, which polluted piped consumers' stdout. The cmd layer now populates `AddRequest.Stdout` at `cmd/library_add.go:352` (the `AddResource` call site) so the dry-run output respects the cmd-side I/O discipline.

#### Scenario: Dry-run shows what would happen

- **GIVEN** a library and valid source file
- **WHEN** AddResource is called with `--dry-run`
- **THEN** no files are modified
- **AND** the expected action is described via the `Stdout io.Writer` field on `AddRequest` (gated on `opts.Stdout != nil`, not directly to `os.Stdout`)

#### Scenario: Library package does not write to os.Stdout

- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** the dry-run output SHALL be written via the `Stdout io.Writer` field on `AddRequest`
