# Per-spec dispatch subagent prompt

Use this template when dispatching one subagent per spec. Replace the placeholders in angle brackets before dispatching. The subagent writes the JSON ledger directly to disk.

---

## Prompt template

```
You are auditing a single OpenSpec spec for the germinator CLI project (Go, Cobra-based configuration adapter).

INPUT
- spec_path: <spec_path>
- category_alignment: read from @openspec/config.yaml — find the entry whose prefix matches the spec's category. The full alignment map is at lines 122-172 of that file.

OUTPUT
- Write JSON to: openspec/reviews/raw/<spec-basename>.json
  where <spec-basename> is the spec's directory name (e.g. cli-cli-factory).
- Return ONLY: "Wrote openspec/reviews/raw/<spec-basename>.json — N findings (X CRITICAL, Y MAJOR, Z MINOR, W NIT)."
- Do NOT include the JSON in your reply.

PROCESS
1. Read the spec end-to-end. Inventory:
   - number of `### Requirement:` blocks
   - number of `#### Scenario:` blocks
   - number of `SHALL` statements
   - total line count
   - presence of a "## Fulfilled" footer (record in summary.fulfilled; if present, emit the format finding in step 8)
2. Spec-format check (5 booleans + issues list) against rules.specs in @openspec/config.yaml:
   - all_shall_explicit
   - given_when_then_format
   - positive_and_negative_scenarios
   - scenarios_concrete
   - scenarios_testable
3. For each SHALL:
   a. Map to candidate code locations via the category alignment.
   b. Read the cited lines BEFORE quoting them.
   c. Status: satisfied / contradicted / unverifiable.
   d. Find test coverage in `*_test.go`, `test/golden/`, `test/e2e/`.
   e. Score the corresponding scenario concrete-ness (CONCRETE / PARTIAL / ABSTRACT).
4. Cross-ref check (file:line existence, sibling specs, historical artifacts like BuildFactory/ServiceContainer/CommandConfig/cobraUsagePrefixes). Record every cross-spec pair as a free-form string in `cross_references[]`; emit structured `findings[].kind = "cross-ref"` for any contradiction you can adjudicate inline. The semantic check subagent (step 3b in `SKILL.md`) consumes `cross_references` and the cross-ref findings as the pair list.
5. Clarity check (SHALL/SHOULD/MAY discipline, placeholders, AND clauses).
6. Apply the severity ladder (see SKILL.md):
   - CRITICAL: a SHALL the code contradicts; cross-spec contradiction on a runtime behavior
   - MAJOR: missing test for a current scenario; SHALL contradicted by current code's precedent; cross-spec policy contradiction; format deviation (incl. ## Fulfilled footer)
   - MINOR: placeholder in scenario; SHALL/MAY mix; missing cross-ref; minor format issue
   - NIT: wording, formatting, ordering
7. Compute top-level alignment from the code_alignment[].status distribution:
   - all satisfied → "aligned"
   - mix → "partial"
   - majority contradicted → "divergent"
   - majority unverifiable → "ahead"
8. If summary.fulfilled is true, emit one MAJOR format finding for the ## Fulfilled footer (see SKILL.md "Deviations as findings" for the canonical shape).
9. Write the JSON file.
10. Return one-line confirmation.

JSON SHAPE

Read `@schemas/ledger.md` for the full schema, field annotations, alignment rules, concrete-ness rules, and a filled example. Produce a JSON object that conforms exactly to the per-spec ledger schema.

DISCIPLINE
- Read every cited line before quoting it.
- Evidence is project source only: cmd/, internal/, test/, .github/, mise.toml, go.mod, openspec/.
- No external citations (Cobra docs, skill references, URLs).
- Paths relative to repo root.
- rules.specs in @openspec/config.yaml is authoritative for spec format; deviations are format findings.
- Verify every file:line exists before including it.

TOOLS
- Use Read, Grep, Glob, Write.
- Do NOT modify any source file outside openspec/reviews/.
```

## Placeholder substitution

| Placeholder | Source |
|---|---|
| `<spec_path>` | Inventory from step 1 of the parent skill |
| `<spec-basename>` | Directory name (e.g., `cli-cli-factory`) |
| `<category_alignment>` | Read from `@openspec/config.yaml` — the entry matching the spec's category prefix |

## When this template is wrong

- For **policy checks** (cross-spec deterministic checks like naming, archive citations), use `@prompts/policy.md`.
- For **semantic checks** (cross-spec judgment on contradiction, redundancy, coherence, domain-overlap), use `@prompts/semantic.md`.
- For **specs with `## Fulfilled` footer**, the dispatch runs the full review (not drift-only) and emits one MAJOR `format` finding for the footer itself; the rest of the review treats the spec as a normal current source of truth.