# Cross-spec audit step procedure

Use this when running **step 3** of the skill (`Cross-spec audit`). The orchestrator reads this procedure and dispatches the per-check subagents using `@prompts/policy.md` and `@prompts/semantic.md`.

## Pre-conditions

Step 3 can only begin when step 2 is complete:

- Every spec under `openspec/specs/` has a parseable JSON ledger in `openspec/reviews/raw/`.
- Every per-spec ledger has `cross_references` (raw observations) and `findings[]` (some entries have `kind: "cross-ref"`).
- `format_compliance` is populated per spec (the format-rollup check consumes it).

If any of these are missing, stop and complete step 2 first.

## Inputs

| Input | Source |
|---|---|
| Check catalog | `@checks/themes.md` |
| Per-check detection rules | `@checks/<check-id>.md` (5 policy + 4 semantic) |
| Output shape | `@schemas/ledger.md` |
| Per-spec ledgers | `openspec/reviews/raw/<spec-basename>.json` |

## Procedure

### 1. Read the check catalog

Open `@checks/themes.md`. For each entry, note the `kind` (`policy` or `semantic`) and the `Purpose`.

### 2. Policy checks (3a)

For each `kind = "policy"` entry:

1. Load `@checks/<check-id>.md`.
2. Load `@prompts/policy.md`.
3. Compose the prompt by substituting:
   - `<check_id>` — from the catalog entry
   - `<rule-name>` — the `Rule:` field at the top of the check spec
   - `<subjects>` — derived per check (most checks use all specs; some use a subset, documented in the check spec)
4. Dispatch one subagent.
5. Wait for the subagent's one-line confirmation.
6. Verify the output file `openspec/reviews/raw/check-<check-id>.json` exists and parses.

Batch ≤3 policy subagents in parallel.

Completion: every `policy` check in the catalog has a ledger in `openspec/reviews/raw/`.

### 3. Build the pair list (3b input)

```text
pairs = ∪ (per-spec.cross_references[i] ∪ per-spec.findings[j].quote
            where per-spec.findings[j].kind == "cross-ref")
         for every per-spec ledger
```

Group by the directory prefix of each spec path (the category: `cli`, `errors`, `library`, etc.).

Skip categories with zero pairs.

### 4. Semantic checks (3b)

For each category with ≥1 pair:

1. Load the four semantic check specs:
   - `@checks/semantic-contradiction.md`
   - `@checks/semantic-redundancy.md`
   - `@checks/semantic-coherence.md`
   - `@checks/semantic-domain-overlap.md`
2. Load `@prompts/semantic.md`.
3. Compose the prompt by substituting:
   - `<category>` — the category prefix
   - `<pairs>` — the pair list from step 3
   - `<per_spec_ledgers>` — paths to per-spec JSON ledgers in this category
4. Dispatch one subagent.
5. Wait for the subagent's one-line confirmation.
6. Verify the output file `openspec/reviews/raw/semantic-<category>.json` exists and parses.

Batch ≤3 semantic subagents in parallel.

Completion: every category with pairs has a ledger in `openspec/reviews/raw/`.

### 5. Quality gate (3c)

For every ledger produced in steps 2 and 4, verify:

- `audit`, `rule`, `subjects_scanned`, `compliance`, `violations`, `summary` fields are present.
- `summary.resolution_status` is one of `aligned | open | stale`.
- Every violation has `kind`, `severity`, `evidence` (or `evidence_a` + `evidence_b`), `rule`, `dedup_key`, `suggested_resolution`.
- Every violation's `kind` is in `{format, cross-ref, contradiction, redundancy, coherence, domain-overlap}`.
- For semantic checks: no violation has `spec_a == spec_b`.
- No ledger embeds a JSON Schema definition (only the `$schema` URL field is allowed).

If a ledger fails the gate, re-dispatch the subagent with the failure flagged, or fix inline if trivial.

Completion: every check ledger passes the gate.

### 6. Deduplication (3d)

For each violation across all check ledgers, key = `(audit, dedup_key)`. If two violations share a key:

- Keep the higher severity.
- Mark the duplicates with `cross_audit_duplicate_of: <canonical-dedup-key>`.

Write `openspec/reviews/dedup-map.json` with the mapping. Shape:

```json
{
  "dedup_keys": {
    "<dedup_key>": {
      "canonical_ledger": "check-<id>.json",
      "severity": "CRITICAL | MAJOR | MINOR | NIT"
    },
    ...
  },
  "duplicates": [
    { "ledger": "check-<id>.json", "violation_index": <int>, "duplicate_of": "<dedup_key>" },
    ...
  ]
}
```

Completion: `openspec/reviews/dedup-map.json` exists and is consistent with the check ledgers.

### 7. Hand off to step 4

Step 3 is complete. Step 4 re-applies the per-spec quality gate (idempotent if no re-dispatch). Step 5 then writes per-category reviews that surface check violations as their own sections.

## Batching reminder

Three concurrent subagents maximum at any point in steps 2 or 4. This limit is enforced to keep context budgets predictable.

## Failure modes

- **Subagent returns without writing the file**: re-dispatch with the file path made explicit.
- **Subagent writes malformed JSON**: do not re-dispatch; fix inline and verify against `@schemas/ledger.md`.
- **Subagent embeds a JSON Schema in the file**: reject and re-dispatch with explicit "no embedded schema" instruction.
- **Per-spec ledgers missing cross_references**: step 2 was incomplete; stop and re-run step 2 for those specs.
- **All categories have zero pairs**: skip step 4 entirely (write empty ledgers or skip); report `resolution_status: "aligned"` in the dashboard.