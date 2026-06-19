# Design — Migrate remaining library commands and delete legacy shell

## Context

After change-6 (`migrate-library-add-create`), the only remaining consumers of `internal/service/` and `internal/application/` are the four lifecycle library commands (`init`, `refresh`, `remove`, `validate`) plus the `legacyBridge` shim in `main.go`. This change migrates the last four commands and deletes the entire legacy shell in one go.

## Goals / Non-Goals

**Goals:**

- All four remaining library commands are migrated to `NewCmdXxx(f, runF) + runXxx(opts)`.
- All four commands gain `--output json|table|plain` via `cmdutil.AddOutputFlags`.
- `internal/service/`, `internal/application/`, `cmd/error_formatter.go`, `cmd/verbose.go`, and `legacyBridge` are deleted.
- `mise run check` is green; `mise run build` succeeds.

**Non-Goals:**

- Migrating `config init`, `config validate` — change-8.
- Migrating `completion`, `version` — change-9.
- Restructuring the `library` package internals — deferred to a follow-up refactor change.

## Decisions

### 1. Each library command declares its minimal `Library` interface

**Choice**: Each of the four commands declares a `Library` interface with only the methods it calls (e.g. `library init` declares `Init(ctx, *InitRequest) error`; `library refresh` declares `Refresh(ctx, *RefreshRequest) error`; etc.).

**Rationale**: matches the `application/command-options-pattern` capability; lets each command depend only on what it needs.

### 2. `library remove` is a single command with sub-command dispatch

**Choice**: `library remove` is one Cobra command (`cmd/library/remove.go`) with sub-command dispatch between `resource` and `preset`. Both sub-commands share the same `removeOptions` struct but populate different fields (`ResourceType` + `ResourceName` for resources, `PresetName` for presets).

**Rationale**: matches the legacy command shape; the sub-command names are part of the public CLI surface.

### 3. `library validate --fix` is preserved

**Choice**: The `--fix` flag on `library validate` is preserved. It triggers auto-cleanup of the `library.yaml` (e.g. removing ghost preset refs).

**Rationale**: `library validate --fix` is a maintenance feature; removing it would break CI scripts that rely on it.

### 4. Mocks deleted in this change

**Choice**: All `internal/service/*_mock_test.go` files are deleted in this change (not in earlier changes). Affected tests are converted to use `iostreams.Test()` + `runF` injection.

**Rationale**: until this change, the mocks were still needed by `cmd/cmd_test.go` sections for non-migrated commands. After this change, no command uses the mocks.

### 5. Deletion order is bottom-up

**Choice**: The deletion sequence is:
1. Migrate the four commands first (tasks 7.1-7.4)
2. Delete `internal/service/` and `internal/application/` (tasks 7.5.1, 7.5.2)
3. Delete `legacyBridge` from `main.go` (task 7.5.3)
4. Delete `cmd/error_formatter.go` and `cmd/verbose.go` (tasks 7.5.4)

**Rationale**: each step removes a consumer from the previous step; `mise run check` after each step catches any missed dependency.

## Risks / Trade-offs

- **Mass deletion** — 4 directories + 2 files in one change. **Mitigation:** `rg` checks at each step catch any missed reference; the deletion order ensures no transient broken state.
- **`library validate --fix` is auto-mutating** — it modifies `library.yaml` in place. **Mitigation:** the `--fix` flag is opt-in; without it, `library validate` is read-only.
- **`library remove` without `--force` on existing resources** — should it refuse? **Mitigation:** the existing behavior is preserved: without `--force`, `library remove` prompts (or refuses in non-TTY); with `--force`, it removes.
