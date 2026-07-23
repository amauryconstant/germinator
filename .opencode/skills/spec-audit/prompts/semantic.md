# Semantic check subagent prompt

Use this template for the cross-spec semantic checks. One subagent per category, judging every pair on all four dimensions (contradiction, redundancy, coherence, domain-overlap). The pair list is the input.

---

## Prompt template

```
You are running cross-spec semantic checks on the germinator OpenSpec set for one category.

INPUT
- category: <category> (e.g. "cli", "errors", "library")
- pairs: <list of pairs and their per-spec ledger citations>
- per_spec_ledgers: <paths to the relevant per-spec JSON ledgers in openspec/reviews/raw/>
- check_specs:
  - @checks/semantic-contradiction.md
  - @checks/semantic-redundancy.md
  - @checks/semantic-coherence.md
  - @checks/semantic-domain-overlap.md

OUTPUT
- Write JSON to: openspec/reviews/raw/semantic-<category>.json
- Return ONLY: "Wrote openspec/reviews/raw/semantic-<category>.json — N pairs judged, M violations (K contradiction, L redundancy, P coherence, Q domain-overlap)."
- Do NOT include the JSON in your reply.

PROCESS
1. Load all four semantic check specs (`@checks/semantic-contradiction.md`, `semantic-redundancy.md`, `semantic-coherence.md`, `semantic-domain-overlap.md`). Internalize each detection rule and severity rule.
2. Load every per-spec ledger in per_spec_ledgers. Skim cross_references and findings[kind=cross-ref] for context.
3. For each pair:
   a. Read both specs in full.
   b. Judge the pair on each of the four dimensions.
   c. For each dimension where the pair produces a violation, record one violation with the appropriate kind, severity, quote_a, quote_b, evidence_a, evidence_b, dedup_key, suggested_resolution.
   d. If the pair is clean on all four dimensions, record nothing for that pair.
4. Dedup: if two violations in your ledger share a (kind, dedup_key) tuple, keep the higher severity.
5. Compute resolution_status:
   - "aligned" — no violations.
   - "open" — at least one violation.
   - "stale" — at least one spec in subjects_scanned contains a "## Fulfilled" footer.
6. Build subjects_scanned: the union of spec paths appearing in any pair.
7. Write the JSON file conforming exactly to the check ledger schema in @schemas/ledger.md.
8. Return one-line confirmation.

JUDGMENT DISCIPLINE
- A pair can produce violations of multiple kinds (e.g., contradictory AND redundant). Record each as a separate violation with a distinct dedup_key.
- A violation's spec_a and spec_b MUST be different paths.
- Quotes must be verbatim SHALL text, not paraphrase.
- Read every cited line before quoting it.
- Evidence is project source only: openspec/, cmd/, internal/, test/.
- Severity MUST follow SKILL.md "Cross-spec audit severity rules". Do not invent severities.
- dedup_key format: `<kind>-<short-topic>` (e.g., "contradiction-confirmOrFlag-placement", "redundancy-library-resolution-precedence").

EMPTY LEDGER
- If after judging every pair you find no violations, write a ledger with empty compliance and empty violations arrays and resolution_status = "aligned". Do NOT invent compliance entries.

NO DUPLICATES
- Do NOT emit the same (kind, dedup_key) twice within your ledger.
- Do NOT embed a JSON Schema definition in the output file. Only a "$schema" URL field is allowed.

TOOLS
- Use Read, Grep, Glob, Write.
- Do NOT modify any source file outside openspec/reviews/.
```

## Placeholder substitution

| Placeholder | Source |
|---|---|
| `<category>` | A category prefix in `openspec/specs/` that has at least one pair |
| `<pairs>` | The per-spec `cross_references` and `findings[kind=cross-ref]` entries from every per-spec ledger, restricted to spec pairs within `<category>` |
| `<per_spec_ledgers>` | Paths to per-spec ledgers in `openspec/reviews/raw/` for specs in `<category>` |

## How this differs from the per-spec prompt

The per-spec agent reads one spec and judges its alignment with code. The semantic agent reads two specs and judges their relationship with each other on four semantic dimensions. The semantic agent never reads code; the per-spec agent never compares two specs.

## When this template is wrong

- For per-spec dispatch, use `@prompts/per-spec.md`.
- For deterministic cross-spec checks (format, naming, citations), use `@prompts/policy.md`.