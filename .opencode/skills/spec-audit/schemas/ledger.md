# Ledger schemas

Canonical JSON shapes for `spec-audit` ledgers. Three prompt templates reference this file via `@schemas/ledger.md`:

- `@prompts/per-spec.md` — produces **per-spec** ledgers (`<spec-basename>.json`).
- `@prompts/policy.md` — produces **policy check** ledgers (`check-<check-id>.json`).
- `@prompts/semantic.md` — produces **semantic check** ledgers (`semantic-<category>.json`).

Subagents `Read` this file directly to get the exact shape; the orchestrator reads ledgers for rollup.

The `$schema` field on every ledger is a URL only — no embedded JSON Schema definitions are allowed.

---

## Per-spec ledger

### JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["spec_path", "alignment", "summary", "format_compliance", "code_alignment", "findings"],
  "properties": {
    "spec_path": { "type": "string", "description": "Absolute or repo-relative path to the spec" },
    "alignment": {
      "enum": ["aligned", "partial", "divergent", "ahead"],
      "description": "Derived from code_alignment[].status distribution. See SKILL.md 'Alignment'."
    },
    "summary": {
      "type": "object",
      "required": ["requirements_count", "scenarios_count", "shalls_count", "lines", "fulfilled"],
      "properties": {
        "requirements_count": { "type": "integer", "minimum": 0 },
        "scenarios_count":    { "type": "integer", "minimum": 0 },
        "shalls_count":       { "type": "integer", "minimum": 0 },
        "lines":              { "type": "integer", "minimum": 0 },
        "fulfilled":          { "type": "boolean", "description": "Whether a ## Fulfilled footer exists" }
      }
    },
    "format_compliance": {
      "type": "object",
      "required": ["all_shall_explicit", "given_when_then_format", "positive_and_negative_scenarios", "scenarios_concrete", "scenarios_testable", "issues"],
      "properties": {
        "all_shall_explicit":              { "type": "boolean" },
        "given_when_then_format":          { "type": "boolean" },
        "positive_and_negative_scenarios": { "type": "boolean" },
        "scenarios_concrete":              { "type": "boolean" },
        "scenarios_testable":              { "type": "boolean" },
        "issues":                          { "type": "array", "items": { "type": "string" } }
      }
    },
    "code_alignment": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["shall_quote", "status", "evidence", "test_coverage", "scenario_concrete"],
        "properties": {
          "shall_quote":       { "type": "string", "description": "Verbatim SHALL statement from the spec" },
          "status":            { "enum": ["satisfied", "contradicted", "unverifiable"] },
          "evidence":          { "type": "string", "description": "repo-relative path:line, e.g. cmd/root.go:42" },
          "test_coverage":     { "type": "string", "description": "'TestName @ path:line' or 'missing'" },
          "scenario_concrete": { "enum": ["CONCRETE", "PARTIAL", "ABSTRACT"] }
        }
      }
    },
    "cross_references": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Raw cross-spec observations. Each entry is a free-form string naming the related spec and the kind of relationship. These are the pair list for semantic checks; the semantic subagent judges each one."
    },
    "clarity_issues":   { "type": "array", "items": { "type": "string" } },
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["severity", "kind", "quote", "evidence", "suggested_resolution", "rule"],
        "properties": {
          "severity":             { "enum": ["CRITICAL", "MAJOR", "MINOR", "NIT"] },
          "kind":                 { "enum": ["drift", "format", "missing-test", "unclear", "cross-ref"] },
          "quote":                { "type": "string" },
          "evidence":             { "type": "string", "description": "repo-relative path:line" },
          "suggested_resolution": { "type": "string" },
          "rule":                 { "type": "string", "description": "e.g. rule.specs.scenarios-concrete or context" }
        }
      }
    }
  }
}
```

### Field-by-field (per-spec)

| Field | Type | Notes |
|---|---|---|
| `spec_path` | string | The exact path the subagent dispatched. One ledger per spec. |
| `alignment` | enum | Top-level signal. Derived from `code_alignment[].status` distribution. See SKILL.md. |
| `summary.requirements_count` | int | Number of `### Requirement:` blocks |
| `summary.scenarios_count` | int | Number of `#### Scenario:` blocks |
| `summary.shalls_count` | int | Number of `SHALL` statements (verbatim count from grep) |
| `summary.lines` | int | Total spec line count |
| `summary.fulfilled` | bool | Whether a `## Fulfilled` footer exists. **If true, the subagent also emits one MAJOR `format` finding** for the footer itself. |
| `format_compliance.*` | bool | Five per-`rules.specs` booleans. `false` = violation. |
| `format_compliance.issues` | string[] | Specific format violations, e.g. "scenario 3 mixes MAY and SHALL" |
| `code_alignment[].shall_quote` | string | Verbatim SHALL text — never paraphrase |
| `code_alignment[].status` | enum | The verdict for this SHALL. |
| `code_alignment[].evidence` | string | `path:line` of the cited code. Must resolve when read. |
| `code_alignment[].test_coverage` | string | `TestName @ path:line` or `missing` |
| `code_alignment[].scenario_concrete` | enum | See "Concrete-ness" below. |
| `cross_references` | string[] | Raw cross-spec observations. The pair list for semantic checks. |
| `clarity_issues` | string[] | SHALL/MAY discipline, placeholders, AND clauses |
| `findings[].severity` | enum | See SKILL.md "Severity ladder". |
| `findings[].kind` | enum | drift · format · missing-test · unclear · cross-ref |
| `findings[].evidence` | string | `path:line`. Same discipline as `code_alignment`. |

### Verdicts (per SHALL)

| Verdict | Meaning |
|---|---|
| `satisfied` | Code implements the SHALL exactly |
| `contradicted` | Code does something different from the SHALL |
| `unverifiable` | No code path exercises the SHALL; the implementation work is the action item |

### Concrete-ness (per scenario)

| Score | Meaning |
|---|---|
| `CONCRETE` | Exact flag, exact path, exact error string |
| `PARTIAL` | Known-shape placeholder (`<resource>`) but anchored |
| `ABSTRACT` | Fully hypothetical (`--old-name`, `oldcmd`, `<message>`) |

### Example — per-spec with `## Fulfilled` footer

```json
{
  "spec_path": "openspec/specs/cli-exit-codes/spec.md",
  "alignment": "aligned",
  "summary": {
    "requirements_count": 9,
    "scenarios_count": 11,
    "shalls_count": 0,
    "lines": 111,
    "fulfilled": true
  },
  "format_compliance": {
    "all_shall_explicit": true,
    "given_when_then_format": true,
    "positive_and_negative_scenarios": true,
    "scenarios_concrete": true,
    "scenarios_testable": true,
    "issues": []
  },
  "code_alignment": [],
  "cross_references": [
    "errors-typed-errors/spec.md:3 cross-references this spec as a parallel delta spec.",
    "cli-error-formatting/spec.md:3 cross-references this spec as a parallel delta spec."
  ],
  "findings": [
    {
      "severity": "MAJOR",
      "kind": "format",
      "quote": "## Fulfilled\n\n**Change:** migrate-library-rest (slice 7 of 9)\n**Date:** 2026-07-01",
      "evidence": "openspec/specs/cli-exit-codes/spec.md:108-110",
      "suggested_resolution": "The footer is a deviation from rules.specs in @openspec/config.yaml. OpenSpec's archival mechanism is `openspec/changes/archive/` (this project has 55 archived changes there). Remove the footer; the change archive is the source of truth. If the project wants to keep this convention, add it to rules.specs.",
      "rule": "rule.specs.*"
    }
  ]
}
```

---

## Check ledger (policy and semantic)

Shared schema for both `check-<check-id>.json` (policy) and `semantic-<category>.json` (semantic). Fields differ in which are populated.

### JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["audit", "rule", "subjects_scanned", "compliance", "violations", "summary"],
  "properties": {
    "audit":             { "type": "string", "description": "Check identifier (matches filename)" },
    "rule":              { "type": "string", "description": "Dotted rule name, e.g. policy.naming-convention-compliance" },
    "subjects_scanned":  { "type": "array",  "items": { "type": "string" }, "description": "Spec paths examined" },
    "compliance": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["spec", "evidence"],
        "properties": {
          "spec":     { "type": "string", "description": "Spec path that follows the rule" },
          "evidence": { "type": "string", "description": "spec-path:line citation" }
        }
      }
    },
    "violations": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["kind", "severity", "evidence", "rule", "dedup_key", "suggested_resolution"],
        "properties": {
          "spec_a":               { "type": "string", "description": "Spec A path. Required for semantic checks; optional for policy checks." },
          "spec_b":               { "type": "string", "description": "Spec B path. Required for semantic checks; optional for policy checks." },
          "kind":                 { "enum": ["format", "cross-ref", "contradiction", "redundancy", "coherence", "domain-overlap"] },
          "severity":             { "enum": ["CRITICAL", "MAJOR", "MINOR", "NIT"] },
          "quote_a":              { "type": "string", "description": "Verbatim SHALL from spec_a (semantic only)" },
          "quote_b":              { "type": "string", "description": "Verbatim SHALL from spec_b (semantic only)" },
          "evidence":             { "type": "string", "description": "For single-spec checks: the spec-path:line of the violation. For pairwise checks: the same — pick the most informative of the two." },
          "evidence_a":           { "type": "string", "description": "spec-a-path:line (semantic only)" },
          "evidence_b":           { "type": "string", "description": "spec-b-path:line (semantic only)" },
          "rule":                 { "type": "string", "description": "Dotted rule name, e.g. policy.naming-convention-compliance or semantic.contradiction" },
          "dedup_key":            { "type": "string", "description": "Stable identifier for cross-check deduplication. Format: <kind>-<short-topic>." },
          "suggested_resolution": { "type": "string" },
          "cross_audit_duplicate_of": { "type": "string", "description": "Set by the deduplication step (3d). The dedup_key of the canonical violation this one duplicates." }
        }
      }
    },
    "summary": {
      "type": "object",
      "required": ["specs_scanned", "violations_count", "resolution_status"],
      "properties": {
        "specs_scanned":      { "type": "integer", "minimum": 0 },
        "violations_count":   { "type": "integer", "minimum": 0 },
        "resolution_status":  { "enum": ["aligned", "open", "stale"], "description": "aligned = no violations; open = violations exist; stale = subjects_scanned includes a Fulfilled-footer spec" }
      }
    }
  }
}
```

### Field-by-field (check)

| Field | Type | Notes |
|---|---|---|
| `audit` | string | Check identifier; matches the filename (`check-<id>.json` or `semantic-<category>.json`). |
| `rule` | string | Dotted rule name. `policy.<id>` for policy checks; `semantic.<kind>` for semantic checks. |
| `subjects_scanned` | string[] | Spec paths examined. The union of paths in `compliance` and `violations`. |
| `compliance[].spec` | string | Spec path that follows the rule. |
| `compliance[].evidence` | string | `path:line` citation. |
| `violations[].spec_a` / `spec_b` | string | Two specs in conflict. Required for semantic checks; optional for policy checks. `spec_a` MUST differ from `spec_b`. |
| `violations[].kind` | enum | One of `format`, `cross-ref`, `contradiction`, `redundancy`, `coherence`, `domain-overlap`. |
| `violations[].severity` | enum | Determined by SKILL.md "Cross-spec audit severity rules" or the check spec. |
| `violations[].quote_a` / `quote_b` | string | Verbatim SHALL text. Required for semantic checks; not used for policy checks. |
| `violations[].evidence` | string | For single-spec checks: the spec-path:line of the violation. For pairwise checks: the most informative of the two. |
| `violations[].evidence_a` / `evidence_b` | string | spec-path:line for each side of a pair (semantic only). |
| `violations[].rule` | string | Dotted rule name. |
| `violations[].dedup_key` | string | Stable identifier for cross-check deduplication. Format: `<kind>-<short-topic>`. |
| `violations[].suggested_resolution` | string | How to reconcile or fix the violation. |
| `violations[].cross_audit_duplicate_of` | string | Set by step 3d (deduplication). The dedup_key of the canonical violation this one duplicates. |
| `summary.resolution_status` | enum | `aligned` (no violations), `open` (violations exist), `stale` (subjects_scanned includes a Fulfilled-footer spec). |

### Severity rules (check violations)

Cross-spec severity rules live in SKILL.md "Cross-spec audit severity rules". They are deterministic. The agent applies them mechanically.

### Example — policy check (no violations)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "audit": "naming-convention-compliance",
  "rule": "policy.naming-convention-compliance",
  "subjects_scanned": [
    "openspec/specs/cli-cli-factory/spec.md",
    "openspec/specs/cli-exit-codes/spec.md"
  ],
  "compliance": [
    { "spec": "openspec/specs/cli-cli-factory/spec.md", "evidence": "openspec/specs/cli-cli-factory/spec.md:1" },
    { "spec": "openspec/specs/cli-exit-codes/spec.md",    "evidence": "openspec/specs/cli-exit-codes/spec.md:1" }
  ],
  "violations": [],
  "summary": {
    "specs_scanned": 2,
    "violations_count": 0,
    "resolution_status": "aligned"
  }
}
```

### Example — policy check (with violations)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "audit": "archive-citation-validity",
  "rule": "policy.archive-citation-validity",
  "subjects_scanned": [
    "openspec/specs/cli-exit-codes/spec.md",
    "openspec/specs/application-dependency-injection/spec.md"
  ],
  "compliance": [
    { "spec": "openspec/specs/application-dependency-injection/spec.md", "evidence": "openspec/specs/application-dependency-injection/spec.md:46-49 (cites migrate-library-rest, exists in archive)" }
  ],
  "violations": [
    {
      "spec": "openspec/specs/cli-exit-codes/spec.md",
      "kind": "cross-ref",
      "severity": "MAJOR",
      "quote": "**Change:** migrate-library-rest (slice 7 of 9)",
      "evidence": "openspec/specs/cli-exit-codes/spec.md:109",
      "rule": "policy.archive-citation-validity",
      "dedup_key": "cross-ref-migrate-library-rest-citation",
      "suggested_resolution": "Verify that openspec/changes/archive/migrate-library-rest exists. If it does not, replace the citation with the actual change archive name."
    }
  ],
  "summary": {
    "specs_scanned": 2,
    "violations_count": 1,
    "resolution_status": "open"
  }
}
```

### Example — semantic check (with pairwise contradiction)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "audit": "semantic-cli",
  "rule": "semantic.contradiction",
  "subjects_scanned": [
    "openspec/specs/cli-destructive-operations/spec.md",
    "openspec/specs/cli-interactive-prompts/spec.md"
  ],
  "compliance": [],
  "violations": [
    {
      "spec_a": "openspec/specs/cli-destructive-operations/spec.md",
      "spec_b": "openspec/specs/cli-interactive-prompts/spec.md",
      "kind": "contradiction",
      "severity": "MAJOR",
      "quote_a": "The `cmdutil` package SHALL provide a `confirmOrFlag(streams *iostreams.IOStreams, force bool, message string) (bool, error)` helper",
      "quote_b": "Shared prompt helpers SHALL live in `internal/output/prompt.go` (or co-located with the prompt user).",
      "evidence_a": "openspec/specs/cli-destructive-operations/spec.md:11",
      "evidence_b": "openspec/specs/cli-interactive-prompts/spec.md:65",
      "rule": "semantic.contradiction",
      "dedup_key": "contradiction-confirmOrFlag-placement",
      "suggested_resolution": "Pick one owning package. Either move both helpers to internal/output/prompt.go (matching cli-interactive-prompts) or to cmdutil (matching cli-destructive-operations), and update the other spec to match. The two specs cannot both mandate different owners for the same concept."
    }
  ],
  "summary": {
    "specs_scanned": 2,
    "violations_count": 1,
    "resolution_status": "open"
  }
}
```

---

## Shared rules

### Severity rubric

The severity ladders live in `SKILL.md` ("Severity ladder" for per-spec, "Cross-spec audit severity rules" for check). Key cases:

| Verdict on a SHALL | Severity (typical) |
|---|---|
| `contradicted` on an implemented feature | CRITICAL |
| `unverifiable` on an unimplemented feature | MAJOR |
| missing test for a current scenario | MAJOR |
| cross-spec contradiction on runtime | CRITICAL |
| cross-spec contradiction on policy | MAJOR |
| cross-spec redundancy | MAJOR |
| cross-spec coherence (blocking) | MAJOR |
| cross-spec coherence (polish) | MINOR |
| cross-spec domain-overlap (split) | MAJOR |
| cross-spec domain-overlap (advisory) | MINOR |
| `## Fulfilled` footer present | MAJOR (format deviation) |
| placeholder in scenario | MINOR |
| `SHALL`/`MAY` mix | MINOR |
| wording, formatting | NIT |

### Evidence discipline (all ledgers)

- Every `evidence` and `test_coverage` path is **repo-relative** (`cmd/root.go:42`, never `/home/...`).
- Paths must **resolve when read** — the quality gate at SKILL.md step 3c rejects ledgers with broken paths.
- No external citations (Cobra docs, skill references, URLs) are allowed as evidence.
- Quotes are **verbatim** — paraphrase is rejected at the quality gate.

### Concrete-ness scoring (per-spec only)

The `scenario_concrete` field is per-SHALL, not per-requirement. A requirement with 3 scenarios can have 3 different concrete-ness scores.

### dedup_key format

Format: `<kind>-<short-topic>`.

Examples:

- `contradiction-confirmOrFlag-placement`
- `redundancy-library-resolution-precedence`
- `coherence-prompt-terminology`
- `domain-overlap-library-resource-import-vs-add`
- `cross-ref-migrate-library-rest-citation`
- `format-naming-convention-cli-factory`

The short-topic is human-readable and stable across runs. The deduplication step (3d) keys on `(audit, dedup_key)`.

### File naming

| Ledger | Path |
|---|---|
| Per-spec | `openspec/reviews/raw/<spec-basename>.json` |
| Policy check | `openspec/reviews/raw/check-<check-id>.json` |
| Semantic check | `openspec/reviews/raw/semantic-<category>.json` |
| Dedup mapping | `openspec/reviews/dedup-map.json` |

`<spec-basename>` is the spec's directory name (e.g., `cli-cli-factory`). `<check-id>` is the check identifier from `themes.md`. `<category>` is a directory prefix in `openspec/specs/` (e.g., `cli`, `errors`, `library`).