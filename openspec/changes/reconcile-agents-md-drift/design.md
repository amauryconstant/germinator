## Context

The 2026-07-08 code review identified 8 drift entries (6 in AGENTS.md, 1 canary-string inaccuracy, 1 missing build tag) that misrepresent the code. These are scattered across 4 AGENTS.md files plus one canary constant and one missing build tag. The drift is concentrated in three areas:

1. **Signature drift in tables**: `cmd/AGENTS.md:278` and `internal/library/AGENTS.md:71,353,358` list outdated function/struct shapes. The first was changed during the 2026-07-01 `migrate-library-rest` change (`openspec/changes/archive/2026-07-01-migrate-library-rest`); the latter three changed during subsequent library work without AGENTS.md updates.
2. **Dispatch-set drift**: `internal/output/AGENTS.md:15` lists a partial `FormatError` dispatch set. New error types `NotFoundError` and `OperationError` were added to the `FormatError` switch (`internal/output/errors.go:31-50`) without AGENTS.md updates. (`InitializeError` exists in `internal/core` but is intentionally not a top-level dispatch case — it is a nested type inside `PartialSuccessError`.)
3. **Cross-doc inconsistency**: `test/AGENTS.md` lists one integration-test file but `internal/cmdutil/integration_test.go` also exists, and the latter is missing the `//go:build integration` tag that the doc promises.

Additionally, four JSON output types in `cmd/presets.go` and `cmd/show.go` are dead code from earlier slices — no callers (per `rg` verification). They are folded into this change because they share the same code-hygiene theme.

The canary string at `internal/warning/canary.go:44` references "slice 2" — a project-internal migration marker that does not appear in `CHANGELOG.md` (verified by `rg 'slice [0-9]' CHANGELOG.md` returning zero matches). Users who see the canary cannot navigate from it to the relevant changelog entry.

### Constraints

1. **Adds one new spec**: this change introduces one new infrastructure spec, `infrastructure-documentation-accuracy` (see `specs/infrastructure-documentation-accuracy/spec.md`). The three existing library/cli specs (`cli-factory`, `library-library-refresh`, `library-library-validation`) are not modified.
2. **Build tag semantics**: the project convention is that `*_integration_test.go` files use `//go:build integration` and are excluded from `go test ./...` (per `test/AGENTS.md`). `internal/parser/integration_test.go:1` already follows this; the cmdutil file does not.
3. **Canary is user-visible**: changing the string in `internal/warning/canary.go:44` is observable on stderr. The change must preserve the user-meaningful portion.
4. **No CLI behavior change**: every other edit is documentation or dead-code removal; no user-facing change.

## Goals / Non-Goals

**Goals:**

- Reconcile 8 AGENTS.md entries with the current code.
- Add the missing `//go:build integration` directive to `internal/cmdutil/integration_test.go`.
- Update `test/AGENTS.md` to list both integration-test files.
- Replace the "slice 2" reference in the canary with a CHANGELOG navigation hint.
- Delete 4 unused JSON output types from `cmd/presets.go` and `cmd/show.go`.

**Non-Goals:**

- Refactoring any production code.
- Changing the semantics or exit-code mapping of any error type.
- Adding new error types or AGENTS.md sections beyond the corrections.
- Restructuring the canary beyond the string text.

## Decisions

### 1. Update each AGENTS.md table cell in place

**Choice**: Edit the 8 drift entries in their existing tables/cells. Do not reorganize the tables.

**Rationale**: AGENTS.md files are referenced by line numbers from skill files and from AGENTS.md files in subpackages. Reorganizing the tables would invalidate `file:line` references in the project's skill-anchored documentation. Editing in place is the minimum-scope change that achieves accuracy.

**Alternatives considered**:

- *Reorganize the tables to alphabetical order*: rejected; would break line-anchored references and add churn unrelated to drift.
- *Move each drift entry to a "see code" pointer*: rejected; defeats the purpose of the table.

### 2. Replace "slice 2" with a CHANGELOG-based navigation hint

**Choice**: Change `internal/warning/canary.go:44` from `"exit code 5 was renamed to 1 in slice 2; consult CHANGELOG for the migration timeline"` to `"exit code 5 was renamed to 1; see CHANGELOG.md for the migration timeline"`.

**Rationale**: The "slice 2" marker is internal migration nomenclature; users navigating from the canary do not have access to it. The CHANGELOG already documents the exit-code change. Removing the "slice 2" portion preserves the user-meaningful warning while removing the unfindable reference.

**Alternatives considered**:

- *Add a `*->` link to the CHANGELOG anchor*: rejected; the canary is written to stderr (not a terminal) and a URL would not be helpful.
- *Reference a specific changelog entry by openspec change ID*: rejected; the canary fires on a deprecation that predates the openspec change-ID scheme; the CHANGELOG entry is the authoritative link.

### 3. Add `//go:build integration` to the cmdutil test file

**Choice**: Add `//go:build integration` as the first line of `internal/cmdutil/integration_test.go`, matching the format at `internal/parser/integration_test.go:1`.

**Rationale**: This is the project's documented convention. Adding the tag excludes the file from `go test ./...` (the default `mise run test` target) and includes it under `go test -tags=integration` (the `mise run test:integration` target). The file content does not need to change.

**Alternatives considered**:

- *Rename the file to `integration_cmdutil_test.go`*: rejected; file naming is a Go convention and `*_integration_test.go` already implies the build tag (per `golang-testing` skill).
- *Add the file to a separate `test/cmdutil` directory*: rejected; the file belongs in the package it tests (per `cmd/AGENTS.md` "I/O in cmd/ forbidden" boundary, the test must live next to the code).

### 4. Delete the 4 unused JSON output types

**Choice**: Delete `PresetsJSONOutput` (cmd/presets.go:160), `PresetInfoJSON` (cmd/presets.go:163), `ShowResourceJSONOutput` (cmd/show.go:225), `ShowPresetJSONOutput` (cmd/show.go:228). Verified zero non-test callers via `rg` before deletion.

**Rationale**: The types were added for the `germinator library presets` and `germinator library show` JSON output paths. Subsequent refactors replaced the per-type JSON projections with the generic `output.NewJSONExporter` (`internal/output/exporter.go:21`). The old types are dead.

**Alternatives considered**:

- *Move them to `internal/output/` for future use*: rejected; dead code in a different package is still dead code. The generic `Exporter` interface is the canonical replacement.
- *Keep them and add a deprecation comment*: rejected; the project convention is to delete unused code, not annotate it.

## Risks / Trade-offs

- **Build tag migration hides test failures**: adding `//go:build integration` to `internal/cmdutil/integration_test.go` removes it from `go test ./...`. **Mitigation**: task 7.3 runs `mise run test:integration` after the change to confirm the file still runs; task 7.2 runs `mise run test` to confirm the file is skipped from the default target. CI must invoke both targets.
- **Deprecation ripple**: deleting the 4 JSON types may surface a hidden caller in a downstream branch. **Mitigation**: task 6.1 runs `rg -e 'PresetsJSONOutput' -e 'PresetInfoJSON' -e 'ShowResourceJSONOutput' -e 'ShowPresetJSONOutput' ./cmd ./internal ./test` before deletion; zero matches expected outside the declarations.
- **CHANGELOG cross-reference accuracy**: the canary text changes from "slice 2" to "CHANGELOG.md" navigation. **Mitigation**: the test at `internal/warning/canary_test.go` must be updated to match the new string; the assertion is on the substring `"exit code 5 was renamed to 1"` which is preserved.
- **AGENTS.md edit drift in future**: the same drift pattern (signature change without doc update) will recur unless the contributing process is improved. **Mitigation**: not in scope for this change; a follow-up infra change can add a CI check that verifies AGENTS.md claims against code.

## Migration Plan

This change is **atomic** — there is no migration sequence because no production code changes behavior. Implementation order within the change:

1. Land the AGENTS.md edits first (no test impact).
2. Land the canary string change + the canary test update (one PR with both).
3. Land the build tag addition + the `test/AGENTS.md` update (one PR — they are coupled).
4. Land the JSON type deletions (one PR; runs `rg` verification pre-merge).

Each step ends with `mise run check` passing. The change is reversible: each step is a single commit that can be reverted independently.

**Rollback strategy**: revert the change's commits in reverse order. The AGENTS.md edits are doc-only and safe to revert. The canary string change is safe to revert. The build tag addition is safe to revert (the file is still integration-tagged; the default `go test` would run it again). The JSON type deletions are safe to revert (the types are restored verbatim).
