# library-library-scaffolding Specification (delta)

## MODIFIED Requirements

### Requirement: Support dry-run mode

The system SHALL preview changes without creating files when `--dry-run` is specified. The dry-run preview SHALL be written to the `stdout io.Writer` parameter passed to `CreateLibrary` (typically `opts.IO.Out` from the cmd layer). The library package SHALL NOT write to `os.Stdout` directly; the cmd layer is responsible for selecting the writer.

**Change**: ADDED the "writer parameter" requirement. The pre-change implementation wrote dry-run output to `os.Stdout` directly, which polluted piped consumers' stdout. The cmd layer now passes `opts.IO.Out` so the dry-run output respects the cmd-side I/O discipline (e.g., `--output json` writes JSON to the same writer; `--output plain` writes plain text).

#### Scenario: Dry-run does not create files

- **GIVEN** no library exists at `/tmp/my-library`
- **WHEN** `germinator library init --path /tmp/my-library --dry-run` is executed
- **THEN** no files or directories are created
- **AND** a message is printed to `opts.IO.Out` (the cmd layer's stdout writer) indicating what would be created

#### Scenario: Library package does not write to os.Stdout

- **WHEN** the codebase is searched for `fmt.Fprintln(os.Stdout` or `fmt.Fprintf(os.Stdout` in `internal/library/`
- **THEN** zero matches SHALL appear
- **AND** the dry-run output SHALL be written via the `io.Writer` parameter passed to `CreateLibrary`
