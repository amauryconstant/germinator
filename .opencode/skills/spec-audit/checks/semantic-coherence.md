# Check: semantic-coherence

**Kind**: semantic
**Rule**: `semantic.coherence`

## Purpose

Two specs read together are ambiguous, use conflicting terminology, or have undefined shared symbols.

This audit catches the case where neither spec contradicts the other in isolation, but together they leave the reader with multiple possible interpretations.

## Detection rule

For each pair `(spec_a, spec_b)`:

1. Read both specs in full.
2. Identify any case where the two specs together create one of:
   - **Terminology drift**: the same word means different things in each spec (e.g., "library" the subcommand vs. "library" the package).
   - **Undefined shared symbol**: a symbol (function, type, error) used by both specs is defined in neither (one spec assumes the other defines it).
   - **Ambiguous reference**: spec A references something that spec B defines, but the reference doesn't pin down which one (e.g., "the validation error" when there are three types).
   - **Conflicting ordering**: spec A says "first X then Y"; spec B says "first Y then X" — neither contradicts the other directly, but together they leave the order unclear.

A coherence issue is NOT:

- A direct contradiction (use the contradiction audit).
- Two specs covering the same domain (use domain-overlap).
- Two specs with the same SHALL (use redundancy).

## Severity

Per SKILL.md "Cross-spec audit severity rules":

| Case | Severity |
|---|---|
| Reading the pair produces an unambiguous interpretation that contradicts one of them (blocking) | MAJOR |
| Reading the pair creates ambiguity but no contradiction (polish) | MINOR |

The MAJOR / MINOR distinction hinges on whether the ambiguity blocks correct implementation. If a maintainer implementing the pair would write code that violates at least one spec, MAJOR. If the ambiguity is cosmetic (multiple terms for the same thing), MINOR.

## Output shape

One violation per coherence issue:

```json
{
  "spec_a": "openspec/specs/<spec_a>/spec.md",
  "spec_b": "openspec/specs/<spec_b>/spec.md",
  "kind": "coherence",
  "severity": "MAJOR | MINOR",
  "quote_a": "<the problematic text from spec_a>",
  "quote_b": "<the problematic text from spec_b>",
  "evidence_a": "openspec/specs/<spec_a>/spec.md:<line>",
  "evidence_b": "openspec/specs/<spec_b>/spec.md:<line>",
  "rule": "semantic.coherence",
  "dedup_key": "coherence-<short-topic>",
  "suggested_resolution": "<how to disambiguate>"
}
```

## dedup_key

Format: `coherence-<short-topic>`.

Examples:

- `coherence-prompt-terminology`
- `coherence-error-type-naming`
- `coherence-library-vs-library-subcommand`

## Discipline

- Be precise about WHY the pair is incoherent. The suggested_resolution should be actionable.
- Do not flag every terminology difference — only those that produce ambiguity in the combined reading.
- If the pair is unambiguously coherent (the reader can resolve any apparent conflict), skip it.