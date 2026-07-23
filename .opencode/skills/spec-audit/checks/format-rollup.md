# Check: format-rollup

**Kind**: policy
**Rule**: `policy.format-rollup`

## Purpose

Aggregate the per-spec `format_compliance` booleans across each category. Flag categories where >30% of specs fail a given rule.

This check answers: "Is the category as a whole following `rules.specs`, or has drift accumulated?"

## Detection rule

1. Read every per-spec ledger in `openspec/reviews/raw/`.
2. Group by category (the directory prefix of `spec_path`).
3. For each category and each `format_compliance` boolean (`all_shall_explicit`, `given_when_then_format`, `positive_and_negative_scenarios`, `scenarios_concrete`, `scenarios_testable`):
   - Count `true` vs `false`.
   - If `false_count / total_count > 0.30`, the category fails that rule.
4. Record one violation per (category, rule) pair that fails.

## Subjects

The set of categories that have at least 5 per-spec ledgers. (Smaller categories are too noisy for the 30% threshold; their findings live in the per-spec ledgers.)

## Severity

| Case | Severity | Kind |
|---|---|---|
| >30% of specs in a category fail a given `format_compliance` rule | NIT | format |

NIT because the per-spec ledgers already record the failure at the spec level; this check only surfaces the rollup signal.

## Output shape

One violation per (category, rule) pair:

```json
{
  "spec": "<category>",
  "kind": "format",
  "severity": "NIT",
  "quote": "<rule name>",
  "evidence": "<category>: <X> of <Y> specs failed this rule",
  "rule": "policy.format-rollup",
  "dedup_key": "format-rollup-<category>-<rule>",
  "suggested_resolution": "Review the per-spec ledgers for this category. Specs with `format_compliance.<rule> == false` are listed in the category review file."
}
```

The `spec` field holds the category name (e.g., `"cli"`) rather than a spec path, because the violation applies to the category rollup.

Compliance entries list (category, rule) pairs that pass.

## dedup_key

`format-rollup-<category>-<rule>` — one per (category, rule) pair.

## Discipline

- Skip categories with fewer than 5 per-spec ledgers.
- Round to 2 decimal places when reporting failure rate.
- Do not flag categories where all 5 rules pass — they are clean.