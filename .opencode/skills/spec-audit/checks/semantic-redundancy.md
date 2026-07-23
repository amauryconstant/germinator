# Check: semantic-redundancy

**Kind**: semantic
**Rule**: `semantic.redundancy`

## Purpose

Two specs describe the same behavior — same pre-conditions, same effect — so one can be removed or merged.

This audit catches the case where two specs independently define the same behavior. The action item is deduplication: keep one, drop or merge the other.

## Detection rule

For each pair `(spec_a, spec_b)`:

1. Read both specs in full.
2. Identify any pair of SHALLs (one from each spec) that:
   - Have the same or near-same pre-conditions (same trigger, same flag, same context).
   - Produce the same or near-same effect (same behavior, same output, same error).
3. The redundancy is real if the two specs independently define the same thing — neither references the other as the canonical source.

A redundancy is NOT:

- One spec referencing the other as the source of truth (that's coherent, not redundant).
- Two specs covering the same domain but with different SHALLs (that's domain-overlap, not redundancy).
- Two specs agreeing on a policy point (that's coherence).

## Severity

Per SKILL.md "Cross-spec audit severity rules": MAJOR (the action item is to deduplicate).

## Output shape

One violation per redundant SHALL pair:

```json
{
  "spec_a": "openspec/specs/<spec_a>/spec.md",
  "spec_b": "openspec/specs/<spec_b>/spec.md",
  "kind": "redundancy",
  "severity": "MAJOR",
  "quote_a": "<verbatim SHALL from spec_a>",
  "quote_b": "<verbatim SHALL from spec_b>",
  "evidence_a": "openspec/specs/<spec_a>/spec.md:<line>",
  "evidence_b": "openspec/specs/<spec_b>/spec.md:<line>",
  "rule": "semantic.redundancy",
  "dedup_key": "redundancy-<short-topic>",
  "suggested_resolution": "Pick one spec as the canonical source of the behavior. Either (a) remove the SHALL from spec_b and add a cross-reference to spec_a, or (b) merge the two specs."
}
```

## dedup_key

Format: `redundancy-<short-topic>`.

Examples:

- `redundancy-library-resolution-precedence`
- `redundancy-cli-init-overwrite-resolution-order`
- `redundancy-exit-code-success-constants`

## Discipline

- Two specs may both happen to mention the same concept (e.g., both reference the cmdutil package). That is not redundancy. Redundancy requires two independent SHALLs with the same pre-conditions and effects.
- A spec that says "see X/spec.md for the precedence rule" is referencing, not redundant.
- Be conservative: when in doubt, prefer coherence (the specs are related but distinct) over redundancy.