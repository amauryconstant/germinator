# Check: naming-convention-compliance

**Kind**: policy
**Rule**: `policy.naming-convention-compliance`

## Purpose

Every directory under `openspec/specs/` matches the doubled-prefix convention documented at `openspec/config.yaml:132-135`: `<category>-<name>` (e.g., `cli-cli-factory`, `errors-typed-errors`).

## Detection rule

For every directory `openspec/specs/<dir>/`:

1. Split `<dir>` on `-`. The first segment is the category.
2. Verify the remaining segments form a valid kebab-case name (lowercase letters, digits, and dashes only).
3. Flag directories that don't match `<category>-<name>` shape (missing the doubled prefix, or have an unexpected structure).

## Subjects

All directories under `openspec/specs/` that contain a `spec.md`.

## Severity

| Case | Severity | Kind |
|---|---|---|
| Directory name does not match `<category>-<name>` shape | MAJOR | format |
| Directory name uses underscores or other non-kebab-case characters | MAJOR | format |
| Directory name is a single segment with no doubled prefix | MAJOR | format |

## Output shape

One violation per misnamed directory:

```json
{
  "spec": "openspec/specs/<misnamed-dir>/spec.md",
  "kind": "format",
  "severity": "MAJOR",
  "quote": "<the directory name>",
  "evidence": "openspec/specs/<misnamed-dir>/",
  "rule": "policy.naming-convention-compliance",
  "dedup_key": "format-naming-convention-<misnamed-dir>",
  "suggested_resolution": "Rename the directory to follow <category>-<name> shape. Verify no other specs, change archives, or code references the old name."
}
```

Compliance entries list correctly-named directories.

## dedup_key

`format-naming-convention-<dir-name>` — one per misnamed directory.

## Example (no violations)

If all directories are correctly named, the ledger has `violations: []` and `resolution_status: "aligned"`.

## Discipline

- Do not flag directories that match the shape — they are compliance.
- Do not propose alternate names; just flag the violation and let the orchestrator suggest a fix.
- Verify the directory exists before flagging it (use `Glob`).