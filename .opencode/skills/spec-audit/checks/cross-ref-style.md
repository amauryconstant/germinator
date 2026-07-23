# Check: cross-ref-style

**Kind**: policy
**Rule**: `policy.cross-ref-style`

## Purpose

Every cross-spec reference uses `file:line` form (e.g., `openspec/specs/cli-exit-codes/spec.md:35`), not a bare directory name (e.g., `cli-exit-codes/spec.md` or `see cli-exit-codes`).

## Detection rule

For every spec in `subjects_scanned`:

1. Grep the spec text for cross-spec reference patterns:
   - `(see X/spec.md)`
   - `see X/spec.md`
   - `X/spec.md` (without `:line`)
   - `see X` (bare directory name reference)
2. Exclude references that DO have `:line` form.
3. Exclude references inside code spans (backticks) that are clearly identifiers, not citations.
4. Flag the remaining references.

## Subjects

All specs under `openspec/specs/`.

## Severity

| Case | Severity | Kind |
|---|---|---|
| Bare directory name used as a cross-spec reference | MINOR | format |
| `X/spec.md` reference without `:line` | MINOR | format |
| Reference inside backticks that points to a real spec path | MINOR | format |

These are polish, not blocking — the cross-ref is correct in spirit, just imprecise. The agent should be lenient: only flag when confident the reference is meant to point to a spec but lacks `file:line`.

## Output shape

One violation per imprecise reference:

```json
{
  "spec": "openspec/specs/<spec>/spec.md",
  "kind": "format",
  "severity": "MINOR",
  "quote": "<the verbatim reference text>",
  "evidence": "openspec/specs/<spec>/spec.md:<line>",
  "rule": "policy.cross-ref-style",
  "dedup_key": "cross-ref-style-<spec>-<line>",
  "suggested_resolution": "Replace the bare reference with file:line form. Example: 'see cli-exit-codes/spec.md' becomes 'see openspec/specs/cli-exit-codes/spec.md:35' (the actual line of the cited SHALL)."
}
```

Compliance entries list specs whose cross-refs use `file:line` form.

## dedup_key

`cross-ref-style-<spec>-<line>` — one per imprecise reference.

## Discipline

- Do not flag references inside changelogs, code blocks, or backtick identifiers.
- Do not flag references that are clearly NOT spec citations (e.g., a mention of `cli-exit-codes` in a design discussion).
- When in doubt, prefer compliance over violation — this check is polish-grade.