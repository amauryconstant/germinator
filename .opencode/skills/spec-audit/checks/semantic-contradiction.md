# Check: semantic-contradiction

**Kind**: semantic
**Rule**: `semantic.contradiction`

## Purpose

Two specs disagree on a shared concept.

This audit adjudicates pairs supplied by the semantic subagent (one subagent per category, judging all four semantic dimensions). The pair list is derived from per-spec `cross_references` and `findings[kind=cross-ref]`.

## Detection rule

For each pair `(spec_a, spec_b)`:

1. Read both specs in full.
2. Identify the shared concept (e.g., "where the confirmOrFlag helper lives", "what exit code NotFoundError maps to", "how library paths are resolved").
3. For each spec, extract the relevant SHALL or normative sentence.
4. Compare the two statements. They contradict if:
   - They make incompatible claims about the same observable behavior (runtime contradiction).
   - They mandate incompatible placement, ownership, naming, or ordering (policy contradiction).
5. Verify the contradiction is real, not a paraphrase artifact — read both quotes verbatim.

A contradiction is NOT:

- Two specs using different terms for the same thing (that's coherence, not contradiction).
- Two specs covering overlapping areas (that's domain-overlap).
- Two specs saying the same thing in different words (that's redundancy).

## Severity

Per SKILL.md "Cross-spec audit severity rules":

| Case | Severity |
|---|---|
| Runtime behavior disagreement (exit code, error message, file content, flag semantics) | CRITICAL |
| Policy disagreement (placement, ownership, naming, ordering) | MAJOR |

## Output shape

One violation per contradiction:

```json
{
  "spec_a": "openspec/specs/<spec_a>/spec.md",
  "spec_b": "openspec/specs/<spec_b>/spec.md",
  "kind": "contradiction",
  "severity": "CRITICAL | MAJOR",
  "quote_a": "<verbatim SHALL from spec_a>",
  "quote_b": "<verbatim SHALL from spec_b>",
  "evidence_a": "openspec/specs/<spec_a>/spec.md:<line>",
  "evidence_b": "openspec/specs/<spec_b>/spec.md:<line>",
  "rule": "semantic.contradiction",
  "dedup_key": "contradiction-<short-topic>",
  "suggested_resolution": "<how to reconcile>"
}
```

## dedup_key

Format: `contradiction-<short-topic>`.

Examples:

- `contradiction-confirmOrFlag-placement`
- `contradiction-notFoundError-exit-code`
- `contradiction-library-resolution-precedence`

The short-topic is human-readable and stable. The deduplication step (3d) keys on `(audit, dedup_key)`.

## Discipline

- Quotes MUST be verbatim from the spec text. Paraphrase is rejected at the quality gate.
- Read every cited line before quoting.
- Evidence MUST be project source only (`openspec/`, `cmd/`, `internal/`, `test/`). No external citations.
- `spec_a` MUST differ from `spec_b` (a same-spec contradiction is invalid).
- If the pair is not actually a contradiction (false positive from per-spec dispatch), skip it — do not emit a violation.
- If two specs disagree on something the user cannot observe (e.g., internal helper naming), the severity is MAJOR (policy), not CRITICAL.
- A pair can produce contradictions on multiple topics. Each topic is a separate violation with its own dedup_key.