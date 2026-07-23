# Check: archive-citation-validity

**Kind**: policy
**Rule**: `policy.archive-citation-validity`

## Purpose

Every reference to `openspec/changes/archive/<name>` in any spec resolves to an existing file or directory in `openspec/changes/archive/`.

## Detection rule

For every spec in `subjects_scanned`:

1. Grep the spec text for the pattern `openspec/changes/archive/<name>`.
2. For each match, extract `<name>` (the change identifier).
3. Verify the file `openspec/changes/archive/<name>/` or `openspec/changes/archive/<name>.md` exists.
4. If not, record a violation.

## Subjects

All specs under `openspec/specs/`.

## Severity

| Case | Severity | Kind |
|---|---|---|
| Cited archive change does not exist | MAJOR | cross-ref |
| Cited archive change exists but is `## Fulfilled`-stamped (the spec is referring to its own archive in a way that creates a loop) | MAJOR | cross-ref |

The audit does not flag legitimate `## Fulfilled` footers — those are handled by the `fulfilled-footer-shape` audit.

## Output shape

One violation per broken citation:

```json
{
  "spec": "openspec/specs/<spec>/spec.md",
  "kind": "cross-ref",
  "severity": "MAJOR",
  "quote": "<the verbatim text containing the broken citation>",
  "evidence": "openspec/specs/<spec>/spec.md:<line>",
  "rule": "policy.archive-citation-validity",
  "dedup_key": "cross-ref-archive-<name>",
  "suggested_resolution": "Either (a) the cited change exists but the path is wrong — fix the path, or (b) the change does not exist — replace the citation with the actual archive name (search openspec/changes/archive/ for similar names)."
}
```

Compliance entries list citations that resolve.

## dedup_key

`cross-ref-archive-<change-name>` — one per broken citation.

## Discipline

- Use `Grep` with `path: "openspec/specs/"` and a precise regex to avoid false positives.
- Verify each citation with `Glob` or `Read` before flagging.
- The dedup_key uses the change identifier from the citation, not the spec name. Different specs citing the same broken archive share a dedup_key (the deduplication step 3d consolidates them).