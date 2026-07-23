# Policy check subagent prompt

Use this template for every check where `kind = "policy"`. The subagent runs deterministic checks against the spec set and emits violations.

---

## Prompt template

```
You are running a cross-spec policy check on the germinator OpenSpec set.

INPUT
- check_id: <check_id>
- check_rule: <rule-name>
- subjects: <list of spec paths in openspec/specs/...>
- check_spec: @checks/<check_id>.md (read this — it defines the detection rule, severity rule, and dedup_key scheme for this check)

OUTPUT
- Write JSON to: openspec/reviews/raw/check-<check_id>.json
- Return ONLY: "Wrote openspec/reviews/raw/check-<check_id>.json — N compliance, M violations."
- Do NOT include the JSON in your reply.

PROCESS
1. Read @checks/<check_id>.md in full. Internalize the detection rule and severity rule.
2. For each subject in subjects:
   a. Read the spec.
   b. Apply the detection rule.
   c. If the spec violates the rule, record one violation; if it follows the rule, record one compliance entry.
3. Compute resolution_status:
   - "aligned" — no violations.
   - "open" — at least one violation.
   - "stale" — at least one subject in subjects contains a "## Fulfilled" footer (the spec was archived).
4. Write the JSON file conforming exactly to the check ledger schema in @schemas/ledger.md.
5. Return one-line confirmation.

JSON SHAPE

Read @schemas/ledger.md for the check ledger schema. Required fields: audit, rule, subjects_scanned, compliance, violations, summary. The violations[].kind for policy checks is always "format" or "cross-ref" (per the check spec).

DISCIPLINE
- Read every cited line before quoting it.
- Evidence is project source only: openspec/.
- No external citations.
- Paths relative to repo root.
- Verify every file:line exists before including it.
- One dedup_key per violation; never re-use across distinct findings.
- Do NOT embed a JSON Schema definition in the output file. Only a "$schema" URL field is allowed.

TOOLS
- Use Read, Grep, Glob, Write.
- Do NOT modify any source file outside openspec/reviews/.
```

## Placeholder substitution

| Placeholder | Source |
|---|---|
| `<check_id>` | From `@checks/themes.md` |
| `<rule-name>` | From `@checks/<check_id>.md` (the rule field) |
| `<subjects>` | All spec paths under `openspec/specs/` (or the subset relevant to this check, per the check spec) |

## When this template is wrong

- For per-spec dispatch, use `@prompts/per-spec.md`.
- For cross-spec semantic checks (judgment on pairs), use `@prompts/semantic.md`.