# library-library-resource-import Specification (delta)

> **Behavior preserved, validation moved earlier.** The migration of `library add` adds `core.CanInstallResource` as a pre-flight ref validator. This is a behavioral addition: refs are now validated before any I/O, and malformed refs produce `*core.ValidationError` instead of a generic error.

## MODIFIED Requirements

### Requirement: Pre-flight ref validation via core.CanInstallResource

When `library add` is invoked in Mode 1 (explicit files), the command SHALL validate each input's resulting ref via `core.CanInstallResource(name)` before any I/O is performed. If validation fails, the command SHALL return `*core.ValidationError` (mapped to exit 1 by `cmdutil.ExitCodeFor` via the default-error case at `internal/cmdutil/exit.go:77`).

> New requirement added in slice-6. The existing scenarios for "Import resource to library", "Auto-detect resource type", "Auto-detect resource name", and others remain valid.

#### Scenario: Valid ref passes pre-flight validation

- **GIVEN** a `--name commit` flag and `--type skill`
- **WHEN** `germinator library add <file> --name commit --type skill` is invoked
- **THEN** `core.CanInstallResource("skill/commit")` SHALL return `nil`
- **AND** the import proceeds (file is copied, library.yaml updated)

#### Scenario: Invalid type fails pre-flight validation

- **GIVEN** a `--name commit` flag and `--type skills` (plural, not a valid type)
- **WHEN** `germinator library add <file> --name commit --type skills` is invoked
- **THEN** `core.CanInstallResource("skills/commit")` SHALL return a non-nil error
- **AND** the error SHALL be `*core.ValidationError`
- **AND** `output.FormatError` SHALL render `Error: ref type must be one of skill, agent, command, memory\n` to stderr
- **AND** no I/O is performed (no file copy, no library.yaml update)

#### Scenario: Empty name fails pre-flight validation

- **GIVEN** `--name ""` (or no `--name` flag and no frontmatter name) and `--type skill`
- **WHEN** `germinator library add <file> --name "" --type skill` is invoked
- **THEN** `core.CanInstallResource("skill/")` SHALL return a non-nil `*core.ValidationError`
- **AND** the import is aborted before I/O

#### Scenario: Malformed ref (no slash) fails pre-flight validation

- **GIVEN** `--name "commit"` and `--type ""` (no type detected)
- **WHEN** the resolved ref would be `"commit"` (no slash)
- **THEN** `core.CanInstallResource("commit")` SHALL return a non-nil `*core.ValidationError`
- **AND** the import is aborted before I/O

> **Why this is a delta, not a new capability**: `core.CanInstallResource` is a private helper called only by `runAdd` and `runCreatePreset`. Its error contract (`*core.ValidationError`) is already covered by the existing `cli-error-formatting` capability. The scenarios above describe the call-site behavior in `library add`, which belongs to the `library-library-resource-import` capability.
