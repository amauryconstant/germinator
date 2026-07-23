---
name: spec-audit
description: "Audit OpenSpec specifications against implementation and across the spec set."
disable-model-invocation: true
license: MIT
---

# spec-audit

> Audit OpenSpec specifications against implementation, and across the spec set for non-contradictory, non-redundant, coherent, and clearly-scoped specifications.

## Purpose

Produce an evidence-grounded audit of every spec under `openspec/specs/`. Every requirement is verified against the codebase; every finding is graded by severity. The output is a **ledger** — a JSON record that an orchestrator can replay, count, and roll up.

Two passes run on the spec set:

1. **Per-spec dispatch** — every spec is dispatched individually for code alignment, format compliance, and clarity.
2. **Cross-spec audit** — the spec set is audited as a whole for: policy compliance (naming, archive citations, format), contradiction, redundancy, coherence, and domain overlap.

The skill never edits specs or code. All edits are advisory text in the review files.

## Vocabulary

- **audit** — the skill itself; also step 3 (Cross-spec audit) of the process. NOT a single verification unit — see `check`. The audit produces audit ledgers (`check-<id>.json`, `semantic-<category>.json`).
- **check** — one verification unit within the cross-spec audit. 5 policy checks (deterministic) and 4 semantic check dimensions (judgment-heavy). Each check has a spec under `@checks/<id>.md`.
- **per-spec dispatch** — a subagent run for one spec; produces `<spec-basename>.json` in `openspec/reviews/raw/`.
- **cross-spec audit** — a subagent run for the spec set; produces `check-<id>.json` (policy) or `semantic-<category>.json` (semantic).
- **pair** — two specs compared by a semantic check. A pair produces 0+ violations.
- **compliance** — the positive side of an audit ledger. A subject that follows the audited rule. Distinct from the per-spec `format_compliance` field, which is unrelated.
- **violation** — a finding of a subject breaking an audited rule. Has `kind`, `severity`, `rule`, `dedup_key`.
- **dedup_key** — a stable identifier on every violation. Enables deduplication (step 3d).
- **resolution_status** — derived top-level field on every audit ledger: `aligned` (no violations), `open` (unresolved violations exist), `stale` (subjects_scanned includes a `## Fulfilled`-footer spec).
- **finding** — per-spec ledger output. Each per-spec ledger has a `findings[]` array; entries have `kind` ∈ {drift, format, missing-test, unclear, cross-ref}.
- **verdict** — per-SHALL determination in per-spec dispatch: `satisfied` / `contradicted` / `unverifiable`. No verdict without a citation.
- **alignment** — top-level field on per-spec ledger: how the spec's SHALLs relate to code (`aligned` / `partial` / `divergent` / `ahead`). Derived from verdict distribution.
- **deviation** — anything in a spec that is not covered by `rules.specs` in `@openspec/config.yaml`. Recorded as `format` findings.

## Severity ladder

| Verdict | When |
|---|---|
| **CRITICAL** | A `SHALL` the code contradicts; cross-spec contradiction on a runtime behavior |
| **MAJOR** | Missing test for a current scenario; `SHALL` contradicted by current code's precedent; cross-spec policy contradiction; cross-spec redundancy or domain-overlap requiring a split; format deviation (incl. `## Fulfilled` footers) |
| **MINOR** | Placeholder in scenario; `SHALL`/`MAY` mix; missing cross-ref; minor format issue; coherence polish; advisory domain-overlap |
| **NIT** | Wording, formatting, ordering |

## Cross-spec audit severity rules

Deterministic. The agent applies the same rule every time given the same input.

| Kind | Severity | Trigger (observable from spec text) |
|---|---|---|
| `contradiction` runtime | **CRITICAL** | Two specs disagree on a runtime behavior the user observes (exit code, error message text, file content, flag semantics). |
| `contradiction` policy | **MAJOR** | Two specs disagree on placement, ownership, naming, or ordering. |
| `redundancy` | **MAJOR** | Two specs have SHALLs whose pre-conditions and effects are identical or near-identical (one should be removed or merged). |
| `coherence` blocking | **MAJOR** | Reading spec A produces an unambiguous interpretation that contradicts spec B. |
| `coherence` polish | **MINOR** | Reading spec A and B together creates ambiguity but no contradiction. |
| `domain-overlap` split | **MAJOR** | Two specs cover the same concept area and both contain SHALLs on it. |
| `domain-overlap` advisory | **MINOR** | Two specs cover the same area but only one has SHALLs. |

## Alignment (top-level, derived, per-spec)

`alignment` is computed from the `code_alignment[].status` distribution:

| Value | Rule |
|---|---|
| `aligned` | All SHALLs `satisfied` |
| `partial` | Mix of `satisfied` + (`contradicted` or `unverifiable`) |
| `divergent` | Majority `contradicted` |
| `ahead` | Majority `unverifiable` (spec describes future work) |

The per-spec subagent records `alignment` at the end of its dispatch (step 7 of `@prompts/per-spec.md`). The orchestrator can recompute it from `code_alignment` if needed.

## Deviations as findings

Any element in a spec that is not covered by `rules.specs` in `@openspec/config.yaml` is a deviation and emits a `format` finding. Today the only systematic deviation in this project is the `## Fulfilled` footer (5 specs use it). The per-spec subagent emits one `format` finding per spec with this footer; see the example in `@schemas/ledger.md`.

Other deviations are surfaced the same way — the auditor does not invent status fields for them.

## Process

### 1. Inventory specs

For every spec under `openspec/specs/<category>-*/spec.md`, record `path` and `lines`. Read `@openspec/config.yaml` to extract the alignment hint for the spec's category prefix.

Completion: every spec has a path, line count, and category alignment.

### 2. Per-spec dispatch (parallel, ≤3 at a time)

For each spec, launch one subagent with `@prompts/per-spec.md`. The subagent writes the JSON ledger and returns a one-line confirmation.

In step 4 of the per-spec prompt ("Cross-ref check"), the subagent records every cross-spec pair in `cross_references[]` (raw observation) and emits structured `findings[].kind = "cross-ref"` for any adjudication it can make inline. Both forms feed step 3.

Completion: every spec has a parseable JSON in `openspec/reviews/raw/` with an `alignment` field and a `cross_references` array.

### 3. Cross-spec audit

The orchestrator delegates this step to `@prompts/cross-spec.md`, which loads `@checks/themes.md`, dispatches subagents using `@prompts/policy.md` and `@prompts/semantic.md`, runs the quality gate, and writes the deduplication mapping. Read that procedure to execute this step.

The four sub-steps below summarize what the procedure does:

#### 3a. Policy checks

Read `@checks/themes.md` for the check index. For each check where `kind = "policy"`:

- Load `@checks/<check-id>.md` (full spec).
- Load `@prompts/policy.md`.
- Dispatch one subagent.

Batch ≤3 in parallel.

Completion: every policy check has a parseable ledger in `openspec/reviews/raw/check-<id>.json`.

#### 3b. Semantic checks

Build the pair list:

```text
pairs = ∪ (per-spec.cross_references[i] ∪ per-spec.findings[j].quote
            where per-spec.findings[j].kind == "cross-ref")
         for every per-spec ledger
```

Group pairs by category (the directory prefix of each spec path). For each category:

- Load the four semantic check specs: `@checks/semantic-contradiction.md`, `semantic-redundancy.md`, `semantic-coherence.md`, `semantic-domain-overlap.md`.
- Load `@prompts/semantic.md`.
- Pass the pair list and the relevant per-spec ledgers as input.
- Dispatch one subagent.

The semantic subagent judges each pair on all four dimensions (contradiction, redundancy, coherence, domain-overlap). One subagent per category, not per check.

Batch ≤3 in parallel.

Completion: every category with at least one pair has a parseable ledger in `openspec/reviews/raw/semantic-<category>.json`.

#### 3c. Quality gate

Reject any ledger (per-spec or check) that:

- lacks an `alignment` field (per-spec only) or `resolution_status` field (check only)
- cites a `file:line` that does not exist when read
- cites an external source (Cobra docs, skill references, URLs) as evidence
- has malformed JSON
- embeds a JSON Schema definition (only the `$schema` URL field is allowed)

For check ledgers, additionally reject:

- any violation whose `spec_a == spec_b` (semantic checks)
- any violation missing `audit`, `rule`, or `dedup_key`
- any `kind` not in `{format, cross-ref, contradiction, redundancy, coherence, domain-overlap}`

Re-dispatch with the issue flagged, or rewrite inline if trivial.

Completion: every ledger passes the gate.

#### 3d. Deduplication

For each violation across all check ledgers, key = `(audit, dedup_key)`. If two violations share a key, keep the higher severity and mark the duplicates with `cross_audit_duplicate_of: <key>`.

Write `openspec/reviews/dedup-map.json` with the mapping. The category-review step reads this file.

Completion: `dedup-map.json` exists and is consistent with the check ledgers.

### 4. Per-spec gate

Re-apply step 3c's per-spec gates after step 3 (check findings do not change per-spec ledgers, but per-spec ledgers may have been re-written during re-dispatch). This step is idempotent if no re-dispatch occurred.

### 5. Category synthesis

Write `openspec/reviews/<category>-review.md` per category with three sections:

1. **Per-spec rollup** — alignment rollup, severity rollup, top-3 findings per spec.
2. **Policy check violations** — every violation from `check-<id>.json` whose `subjects_scanned` includes a spec in this category, post-dedup.
3. **Semantic check findings** — every violation from `semantic-<category>.json`, post-dedup.

Completion: one review file per category, with all three sections populated.

### 6. Dashboard

Write `openspec/reviews/REVIEW.md`: totals table, per-category rollup, CRITICAL+MAJOR only.

Completion: dashboard exists and is consistent with per-category files.

## Branches

- **pilot** — first run on a new codebase, 3 specs only, stop after step 2 for calibration.
- **standard** — full inventory; run all 6 steps.
- **category** — user specifies one category; steps 1, 2, 4, 5 only.

## Discipline

- Read every cited line before quoting it.
- Evidence is project source only: `cmd/`, `internal/`, `test/`, `.github/`, `mise.toml`, `go.mod`, `openspec/`.
- Verify every `file:line` exists before including it (use `Read` with the line range).
- Paths are relative to repo root.
- One verdict per `SHALL`; one severity per violation.
- One `dedup_key` per violation; never re-use a key across distinct findings.
- `rules.specs` in `@openspec/config.yaml` is the source of truth for spec format. Deviations are findings, not accommodations.
- Check ledgers never embed a JSON Schema definition; only the `$schema` URL field is allowed.

## References

- `@prompts/per-spec.md` — per-spec dispatch subagent prompt template.
- `@prompts/policy.md` — policy check subagent prompt template.
- `@prompts/semantic.md` — semantic check subagent prompt template (judges all 4 dimensions per category).
- `@prompts/cross-spec.md` — orchestrator-side procedure for step 3 (load this when running the cross-spec audit step).
- `@checks/themes.md` — check index (read this first; 5 lines per check).
- `@checks/<check-id>.md` — full spec per check (5 policy + 4 semantic files, all flat in `checks/`).
- `@schemas/ledger.md` — canonical JSON shapes for per-spec and check ledgers (schemas, field annotations, examples).
- `@openspec/config.yaml` — category alignment hints, rules.specs, rules.testing, project conventions.