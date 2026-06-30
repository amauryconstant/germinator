# library-library-orphan-discovery Specification (delta)

> **Behavioral surface change.** The pre-change code carried `name_conflict` outcomes as a string `Issue` field on the result type. This delta moves the conflict to a typed error: `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}`. The `name_conflict` scenario in the existing requirement is re-expressed in terms of the typed error.

## MODIFIED Requirements

### Requirement: Conflict detection produces typed OperationError

The `library add --discover` command SHALL detect when an orphan has the same name as an existing resource. The conflict SHALL be reported as a `*core.OperationError{Op: "register", Resource: <ref>, Cause: <origErr>}` per file, aggregated into a `*core.PartialSuccessError`, and counted toward `Failed`. The conflict SHALL be rendered to stderr via `output.FormatError` per file.

> Replaces the pre-change behavior where `ConflictInfo{Issue: "name_conflict"}` was carried as a string field on `DiscoverResult`.

#### Scenario: Detect name conflict produces OperationError

- **GIVEN** a library with existing resource `skill/commit`
- **AND** an orphan file `skills/commit.md` not in library.yaml
- **WHEN** `germinator library add --discover --force` is invoked
- **THEN** a `*core.OperationError{Op: "register", Resource: "skill/commit", Cause: <origErr>}` SHALL be produced for the file
- **AND** the OperationError SHALL be aggregated into the partial-success result with `Failed` incremented by 1
- **AND** `output.FormatError` SHALL render `Error: register: skill/commit\n` to **stderr** (`opts.IO.ErrOut`)
- **AND** the orphan SHALL NOT be registered (the file is left untouched)

#### Scenario: Name conflict counts as failure, not success

- **GIVEN** a library with 2 orphans: `skills/orphan1.md` (valid, no conflict) and `skills/orphan2.md` (conflicts with existing `skill/orphan2`)
- **WHEN** `germinator library add --discover --force --batch` is invoked
- **THEN** the partial-success aggregate SHALL have `Succeeded == 1` and `Failed == 1`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeSuccess` (0) because `Succeeded > 0`
- **AND** stdout SHALL contain the success listing for `skill/orphan1`
- **AND** stderr SHALL contain `Error: register: skill/orphan2` from the per-file FormatError render

#### Scenario: All conflicts returns exit 1

- **GIVEN** a library with 2 orphans, both conflicting with existing resources
- **WHEN** `germinator library add --discover --force --batch` is invoked
- **THEN** the partial-success aggregate SHALL have `Succeeded == 0` and `Failed == 2`
- **AND** `cmdutil.ExitCodeFor(err)` SHALL return `ExitCodeError` (1)
- **AND** stdout SHALL be empty (no data leakage on error paths)
- **AND** stderr SHALL contain the two `Error: register: ...` lines

#### Scenario: OperationError preserves wrapped cause

- **GIVEN** a name conflict where the library package returns a typed `library.ErrNameConflict` as the underlying cause
- **WHEN** the conflict is reported as `*core.OperationError{Op: "register", Resource: <ref>, Cause: library.ErrNameConflict}`
- **THEN** `errors.Is(err, library.ErrNameConflict)` SHALL be `true`
- **AND** `errors.Unwrap(err)` SHALL return the cause
- **AND** `output.FormatError` SHALL render both the typed error message and the cause on separate lines

> **Preserved scenario (unchanged)**: The existing base-spec scenarios for "Discover orphans in skills directory", "Detect orphan type from directory", "Detect orphan name from frontmatter or filename", "Report-only mode by default", "Force mode registers orphans", "Support dry-run with discover", "Discover library path", "Enhanced discover result structure" all remain valid. The delta modifies only the "Conflict detection" requirement (and the scenario immediately preceding it about `Conflicts` field).
