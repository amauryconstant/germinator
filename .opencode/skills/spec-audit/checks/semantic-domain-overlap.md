# Check: semantic-domain-overlap

**Kind**: semantic
**Rule**: `semantic.domain-overlap`

## Purpose

Two specs cover the same conceptual area. The action item is to either split the domain (each spec owns a clear sub-domain) or merge the specs.

This audit answers: "Are these two specs trying to do the same job?"

## Detection rule

For each pair `(spec_a, spec_b)`:

1. Read both specs in full.
2. Determine the conceptual area each spec covers (one sentence: "this spec is about X").
3. If the two areas overlap (one is a subset, superset, or near-equivalent of the other), record a domain-overlap finding.
4. Decide whether the overlap requires a split (both specs have SHALLs on the same concept) or is just advisory (one spec has SHALLs, the other is contextual).

Domain overlap is NOT:

- Two specs in the same category (categories are organizational, not domain boundaries).
- Two specs sharing a dependency (e.g., both reference the same error type).
- Two specs that are part of the same change (those are deliberate sub-specs of one spec).

## Severity

Per SKILL.md "Cross-spec audit severity rules":

| Case | Severity |
|---|---|
| Both specs have SHALLs on the same concept area (split needed) | MAJOR |
| Only one spec has SHALLs; the other is contextual (advisory) | MINOR |

The split-vs-advisory distinction hinges on whether maintainers would have to decide which spec owns the concept. If yes, MAJOR. If the contextual spec clearly defers, MINOR.

## Output shape

One violation per overlapping pair:

```json
{
  "spec_a": "openspec/specs/<spec_a>/spec.md",
  "spec_b": "openspec/specs/<spec_b>/spec.md",
  "kind": "domain-overlap",
  "severity": "MAJOR | MINOR",
  "quote_a": "<representative SHALL from spec_a>",
  "quote_b": "<representative SHALL from spec_b>",
  "evidence_a": "openspec/specs/<spec_a>/spec.md:<line>",
  "evidence_b": "openspec/specs/<spec_b>/spec.md:<line>",
  "rule": "semantic.domain-overlap",
  "dedup_key": "domain-overlap-<short-topic>",
  "suggested_resolution": "<how to split or merge>"
}
```

## dedup_key

Format: `domain-overlap-<short-topic>`.

Examples:

- `domain-overlap-library-resource-import-vs-add`
- `domain-overlap-cli-init-vs-library-init`
- `domain-overlap-exit-codes-vs-error-formatting`

## Discipline

- Categories organize specs but don't define domains. Two specs in the same category can have clear, separate domains.
- Be precise about WHY the domains overlap. The suggested_resolution should name the sub-domain each spec should own.
- If the two specs clearly partition their domain (each owns a distinct sub-area), skip it — that's good design.