# Check: fulfilled-footer-shape

**Kind**: policy
**Rule**: `policy.fulfilled-footer-shape`

## Purpose

Every `## Fulfilled` footer in any spec:

1. Has a `**Change:**` field naming a change archive.
2. Has a `**Date:**` field in `YYYY-MM-DD` form.
3. The cited change exists in `openspec/changes/archive/`.

This audit supplements `archive-citation-validity` (which checks every archive citation everywhere) by checking the structure of `## Fulfilled` footers specifically. The per-spec dispatch already emits a MAJOR `format` finding for the existence of any `## Fulfilled` footer (a deviation from `rules.specs`); this audit checks the SHAPE of those that do exist.

## Detection rule

For every spec in `subjects_scanned`:

1. Grep for the `## Fulfilled` header.
2. If present, parse the next ≤5 lines for `**Change:**` and `**Date:**` fields.
3. Verify the date is `YYYY-MM-DD`.
4. Verify the change archive at `openspec/changes/archive/<change-name>/` or `.md` exists.
5. Flag any spec that violates one or more of these shape rules.

## Subjects

All specs under `openspec/specs/` that contain a `## Fulfilled` footer.

## Severity

| Case | Severity | Kind |
|---|---|---|
| Missing `**Change:**` field | MAJOR | format |
| Missing `**Date:**` field | MAJOR | format |
| Date is not `YYYY-MM-DD` | MAJOR | format |
| Cited change does not exist in archive | MAJOR | format |

## Output shape

One violation per shape violation:

```json
{
  "spec": "openspec/specs/<spec>/spec.md",
  "kind": "format",
  "severity": "MAJOR",
  "quote": "## Fulfilled\n\n<the verbatim footer text>",
  "evidence": "openspec/specs/<spec>/spec.md:<footer-line>",
  "rule": "policy.fulfilled-footer-shape",
  "dedup_key": "format-fulfilled-footer-<spec>",
  "suggested_resolution": "Fix the footer to follow the documented shape: '## Fulfilled\\n\\n**Change:** <archive-name>\\n**Date:** YYYY-MM-DD'. Verify the cited change exists in openspec/changes/archive/."
}
```

Compliance entries list specs whose `## Fulfilled` footers pass all three shape checks.

## dedup_key

`format-fulfilled-footer-<spec>` — one per spec (not per shape violation within a spec; if a spec has multiple shape issues, they roll up into one violation).

## Discipline

- This audit is supplemental to per-spec dispatch's `format` finding on `## Fulfilled` existence. Do not duplicate the existence finding; only emit shape violations.
- Specs without a `## Fulfilled` footer are NOT subjects of this audit. Do not flag them.
- Use `Read` to verify the change archive exists.