# library-library-batch-add Specification (delta)

## MODIFIED Requirements

### Requirement: Dry-run in batch mode

The system SHALL preview changes without modifying library when `--dry-run` is set in batch mode. The dry-run preview SHALL be written to the `stdout io.Writer` parameter passed to `BatchAddResources` (typically `opts.IO.Out` from the cmd layer). The library package SHALL NOT write to `os.Stdout` directly.

**Change**: ADDED the "writer parameter" requirement.

#### Scenario: Batch dry-run shows additions

- **GIVEN** `--batch` and `--dry-run` flags are set with valid sources
- **WHEN** `library add --batch --dry-run skill-a.md skill-b.md` is called
- **THEN** no files are modified
- **AND** summary shows what would be added (via the `io.Writer` parameter)

#### Scenario: Batch dry-run shows failures

- **GIVEN** `--batch` and `--dry-run` flags are set with one invalid source
- **WHEN** `library add --batch --dry-run valid.md invalid.md` is called
- **THEN** no files are modified
- **AND** the invalid file is shown as would-fail (via the `io.Writer` parameter)
