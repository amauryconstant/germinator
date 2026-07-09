# infrastructure-documentation-accuracy Specification (delta)

## ADDED Requirements

### Requirement: AGENTS.md claims match code

AGENTS.md files throughout the project MUST accurately reflect the code they document. When a function signature, struct shape, dispatch set, or integration-test file list changes, the corresponding AGENTS.md entry MUST be updated in the same change. Drift is a CI-detectable defect.

**Change**: NEW requirement codifying the post-reconciliation state. Before this change, the codebase had 8 known AGENTS.md drift entries (per the 2026-07-08 code review); this requirement establishes that the post-change state is the contract, not the pre-change drift.

#### Scenario: Function signature table cells match code

- **WHEN** an AGENTS.md file lists a function signature in a table cell (e.g., `internal/library/AGENTS.md:71` lists `FindLibrary(flagPath, envPath string) string`)
- **THEN** the actual function signature in the code MUST match the table cell
- **AND** a code change that alters the signature SHALL update the table cell in the same commit

#### Scenario: Struct shape table cells match code

- **WHEN** an AGENTS.md file lists a struct definition in a table cell (e.g., `internal/library/AGENTS.md:353,358` list `SkipInfo` and `RefreshError`)
- **THEN** the actual struct definition in the code MUST match the table cell
- **AND** a code change that adds, removes, or renames a struct field SHALL update the table cell in the same commit

#### Scenario: FormatError dispatch set is complete in AGENTS.md

- **WHEN** `internal/output/AGENTS.md:15` documents the set of error types handled by `output.FormatError`
- **THEN** the documented set MUST include every error type present in the `switch { case errors.As(...) }` block at `internal/output/errors.go:21-50`
- **AND** adding a new `case` to the switch SHALL update the AGENTS.md set in the same commit

#### Scenario: Integration-test file list matches reality

- **WHEN** `test/AGENTS.md:62-86` documents the set of `*_integration_test.go` files
- **THEN** every file matching the pattern in the repository MUST be listed
- **AND** every listed file MUST exist with the `//go:build integration` directive at line 1
- **AND** adding a new `*_integration_test.go` file SHALL update the list in the same commit

#### Scenario: User-visible strings reference discoverable docs

- **WHEN** the codebase emits a user-visible string that references project documentation (e.g., the canary at `internal/warning/canary.go:44` post-change mentions "CHANGELOG"; the pre-change text references "slice 2", which is the inaccuracy this requirement fixes)
- **THEN** the referenced document MUST exist at the named location
- **AND** the string MUST be navigable from the user's perspective (a CHANGELOG entry, an openspec change archive, or a release note)

### Requirement: Dead code is removed, not annotated

Types, functions, and variables that have zero non-test callers SHALL be deleted rather than annotated with `// Deprecated:` comments.

**Change**: NEW requirement codifying the project's existing convention. The 4 dead JSON output types deleted in this change (`PresetsJSONOutput`, `PresetInfoJSON`, `ShowResourceJSONOutput`, `ShowPresetJSONOutput`) are examples of the pattern: delete, do not annotate.

#### Scenario: Unused types are deleted

- **WHEN** a type has zero non-test callers (verified by `rg`)
- **THEN** the type SHALL be deleted
- **AND** the type SHALL NOT be marked `// Deprecated:` or moved to a `legacy/` package
- **AND** any test fixtures that reference the type SHALL be updated to use the canonical replacement
