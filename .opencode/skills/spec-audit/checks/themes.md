# Check index

Cross-spec checks. Read this file first; for each check, follow the reference link to the full spec.

Two kinds:

- **policy** — deterministic, mechanical. One subagent per check.
- **semantic** — judgment-heavy, pairwise. One subagent per category (runs all 4 semantic check dimensions on the pairs in that category).

Per-spec dispatch produces the pair list. Checks adjudicate it.

---

## Policy checks (5)

### naming-convention-compliance
- **Kind**: policy
- **Purpose**: Every `openspec/specs/<dir>/spec.md` directory matches the doubled-prefix rule (`openspec/config.yaml:132-135`).
- **Spec**: `@checks/naming-convention-compliance.md`

### archive-citation-validity
- **Kind**: policy
- **Purpose**: Every `openspec/changes/archive/<name>` reference in any spec resolves to an existing file.
- **Spec**: `@checks/archive-citation-validity.md`

### cross-ref-style
- **Kind**: policy
- **Purpose**: Every cross-spec reference uses `file:line` form, not a bare directory name.
- **Spec**: `@checks/cross-ref-style.md`

### fulfilled-footer-shape
- **Kind**: policy
- **Purpose**: Every `## Fulfilled` footer has Change and Date fields, and the referenced change exists in `openspec/changes/archive/`.
- **Spec**: `@checks/fulfilled-footer-shape.md`

### format-rollup
- **Kind**: policy
- **Purpose**: Aggregate per-spec `format_compliance` booleans per category per rule; flag categories where >30% of specs fail a rule.
- **Spec**: `@checks/format-rollup.md`

---

## Semantic checks (4)

Run as one subagent per category via `@prompts/semantic.md`. The subagent judges each pair on all 4 dimensions.

### semantic-contradiction
- **Kind**: semantic
- **Purpose**: Two specs disagree on a shared concept. Severity per SKILL.md "Cross-spec audit severity rules".
- **Spec**: `@checks/semantic-contradiction.md`

### semantic-redundancy
- **Kind**: semantic
- **Purpose**: Two specs describe the same behavior (same pre-conditions and effects). Action item is to merge or remove one.
- **Spec**: `@checks/semantic-redundancy.md`

### semantic-coherence
- **Kind**: semantic
- **Purpose**: Two specs read together are ambiguous, use conflicting terminology, or have undefined shared symbols.
- **Spec**: `@checks/semantic-coherence.md`

### semantic-domain-overlap
- **Kind**: semantic
- **Purpose**: Two specs cover the same conceptual area.
- **Spec**: `@checks/semantic-domain-overlap.md`

---

## Editing this index

To add a policy check: add an entry here + write `checks/<id>.md`.

To add a semantic check dimension: add an entry here + write `checks/<id>.md` + reference it from `prompts/semantic.md`.

To retire a check: delete its entry here and its spec file. Existing check ledgers stay in `openspec/reviews/raw/` until manually cleaned.