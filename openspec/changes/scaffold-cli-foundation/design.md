# Design — Scaffold golang-cli-architecture foundation

## Context

Germinator is a Go CLI that transforms AI coding assistant documents (agent, command, skill, memory) between two target platforms (Claude Code, OpenCode) using a canonical YAML source format. It is invoked by developers as a build-time adapter and must be scriptable, deterministic, and quiet on `stdout`.

**Current state (baseline at slice 0):**

- `cmd/` (22 files) is a flat package containing the Cobra root, all subcommands, two formatters, a verbosity helper, and a global `CommandConfig` shared by reference.
- `cmd/container.go` instantiates a `ServiceContainer` eagerly in `main.go` with all four services (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`).
- `cmd/error_handler.go` and `cmd/error_formatter.go` maintain seven exit codes (`0, 1, 2, 3, 4, 5, 6`) and a `CategorizeError` enum.
- `internal/application/` defines four service interfaces plus `Parser` and `Serializer`, with corresponding `Request` and `Result` DTOs.
- `internal/service/` implements the four interfaces.
- `internal/infrastructure/` is a deep tree: `parsing/`, `serialization/`, `adapters/{claude-code,opencode}/`, `config/`, `library/`.
- `internal/domain/` (12 files) is already pure (depguard-enforced) and contains the canonical data types, the `Result[T]` functional type, validation pipelines, and five typed errors.
- `internal/models/` retains a residual `constants.go` and `doc.go` (deleted in change-9).

**Constraints** (from the project's `AGENTS.md` and `openspec/config.yaml`):

- `mise run check` (`lint` + `format` + `test` + `build`) must pass on every commit.
- Domain layer must remain depguard-isolated (no external dependencies beyond stdlib + `samber/lo`).
- No comments in code unless explicitly requested.
- Table-driven tests with fixtures; 70%+ coverage target for `cmd/` and `adapters/`.

**The full migration** is documented in the original monolithic `adopt-golang-cli-architecture` design. This change implements only the foundation. The remaining 8 changes (wire + 7 command-migration slices + cleanup) build on it.

## Goals / Non-Goals

**Goals (this change):**

- Create `internal/iostreams/`, `internal/output/`, `internal/cmdutil/` as the three shell concerns.
- Rename `internal/domain/` → `internal/core/` and flatten `internal/infrastructure/*` into top-level `internal/<concern>/` packages.
- Add `core.PartialSuccessError` to the error type set (consumer arrives in change-5).
- Establish `forbidigo` lint rules blocking `fmt.Fprintf(os.Stdout|Stderr)` and `os.Exit(` in `cmd/`.
- Land all of the above with `mise run check` green and ≥70% coverage in the new packages.
- **Preserve all current CLI behavior byte-identically** (no command is migrated in this change).

**Non-Goals (this change):**

- Rewiring `main.go` to use the Factory directly — change-2.
- Migrating any command to the new pattern — changes 2-9.
- Deleting legacy files (`cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`) — change-2.
- Deleting `internal/service/` and `internal/application/` — change-7.
- Removing the `Category*` enum and `CategorizeError` function — change-7 (when the last command using them is migrated).
- Adding `github.com/charmbracelet/huh` — deferred until first interactive prompt.
- Rewriting the canonical data types, template rendering logic, or the `Result[T]`/`ValidationPipeline[T]` patterns.

## Decisions

### 1. Foundation scope is packages + renames, no command migration

**Choice**: This change creates new packages and performs mechanical renames; it does NOT migrate any command and does NOT wire `main.go`.

**Rationale**: The mechanical work (rename + flatten + new package scaffolding) is independent of command migration. Doing it first means every subsequent change operates against the final package layout. Splitting the foundation out of slice 1 of the original plan keeps this change to ~30 tasks and ~10 files of new code.

**Alternatives considered**:

- Include `main.go` rewiring in this change → couples the foundation to integration; doubles the task count and the verification surface.
- Skip the rename entirely (keep `internal/domain/`) → the new `core/rules.go` would have to live in `domain/`, contradicting the skill's vocabulary.

### 2. `internal/core/` rename is mechanical, not semantic

**Choice**: All 12 files move from `internal/domain/` to `internal/core/` with only the package declaration changing. A `type Domain = core` alias in `internal/core/doc.go` provides backward compatibility during the migration window (removed in change-9).

**Rationale**: matches the skill's terminology (`core` instead of `domain`); the rename is reversible via the alias if needed; CI catches every missed import via `rg 'internal/domain'`.

**Alternatives considered**:

- Skip the rename → inconsistent with the skill's vocabulary.
- Rename incrementally (some files now, others later) → unnecessary complexity; the rename has no semantic risk.

### 3. Flatten `infrastructure/` to top-level `internal/<concern>/` packages

**Choice**: Each subdirectory of `internal/infrastructure/` becomes a top-level package under `internal/`. The umbrella `infrastructure/` directory is deleted.

**Rationale**: matches the target layout in `openspec/config.yaml` (`internal/parser/`, `internal/renderer/`, `internal/claude-code/`, `internal/opencode/`, `internal/config/`, `internal/library/`); reduces import depth; aligns with the skill's "one package per concern" rule.

**Alternatives considered**:

- Keep `infrastructure/` umbrella and rename subdirs → larger diff, no benefit.
- Restructure each package while moving → conflates two changes; deferred per the skill's "extract when painful, not predicted".

### 4. `PartialSuccessError` is added in this change but only used in change-5

**Choice**: `core.PartialSuccessError` is defined in `internal/core/errors.go` with full tests, even though the first consumer (`init`) arrives in change-5.

**Rationale**: The type is part of the **error vocabulary** of the new architecture. Defining it in the foundation lets `cmdutil.ExitCodeFor` and `output.FormatError` recognize it from day one (their tests in `internal/cmdutil/exit_test.go` and `internal/output/output_test.go` cover the partial-success path). Without the type present, the `errors.As` switch in `FormatError` would be incomplete; with it present but unused, it's a no-op until `init` arrives.

**Alternatives considered**:

- Define `PartialSuccessError` in change-5 → forces a follow-up commit to `output/FormatError` and `cmdutil/ExitCodeFor` to add the `errors.As` branch.
- Skip the type entirely → partial-success behavior diverges from current (init would lose its "exit 0 if at least one succeeded" guarantee).

### 5. `forbidigo` patterns are scoped narrowly

**Choice**: The forbidden patterns are:

- `fmt\.Fprintf\(os\.(Stdout|Stderr)` in `cmd/*.go` excluding tests
- `os\.Exit\(` in `cmd/**` excluding `main.go`

Plus four global patterns: `var global(Factory|CommandConfig)`, `SetGlobal(Factory|CommandConfig)`, `context\.Background\(\)` in `cmd/**/*.go` (except `main.go`).

**Rationale**: scoped to `cmd/` because that's the only place these anti-patterns matter; excluding `main.go` for `os.Exit` because `main.go` is the only allowed caller of `os.Exit`; excluding tests because tests need `fmt.Fprintf(os.Stdout, ...)` for setup assertions.

**Alternatives considered**:

- Broader scope (`internal/**` for `os.Exit`) → false positives in code paths that legitimately need to terminate (e.g., `internal/version`).
- No `forbidigo` → the architectural rules live only in code review; humans miss them.

### 6. `lipgloss` for styles; `huh` deferred

**Choice**: `github.com/charmbracelet/lipgloss` is added in this change (slice 1 of the original plan; task 1.1.4). `github.com/charmbracelet/huh` is NOT added in this change.

**Rationale**: `lipgloss` is used immediately by `iostreams.Styles`. `huh` has no current consumer (no command prompts interactively); adding it speculatively violates the skill's "extract when painful, not predicted" principle. The first interactive prompt triggers a future change that adds `huh` and creates `output/prompts.go`.

### 7. Foundation preserves `Result[T]` and `ValidationPipeline[T]`

**Choice**: The existing `core.Result[T]` and `core.ValidationPipeline[T]` types are preserved as-is. No renaming, no migration to `(T, error)`.

**Rationale**: the skill recommends `(T, error)` over a generic `Result[T]`, but the existing codebase uses `Result[T]` extensively. Migrating would require touching every validator and would obscure the composability that `Result.OrElse` / `Result.IsOk` provide. Deferred to a future change.

### 8. `addOutputFlags` lives in `internal/output`, re-exported as `cmdutil.AddOutputFlags`

**Choice**: The implementation lives in `internal/output/output_flags.go` (with the per-command wiring knowledge). `cmdutil/output_flags.go` re-exports it as `cmdutil.AddOutputFlags` so command files import only `cmdutil`.

**Rationale**: matches the proposal/spec text which references `cmdutil.AddOutputFlags`; avoids circular imports (`output/` doesn't import `cmdutil/`, but the re-export is one-way).

### 9. Test files for new packages use table-driven format

**Choice**: `internal/iostreams/{iostreams_test.go, styles_test.go}`, `internal/output/output_test.go`, `internal/cmdutil/{factory_test.go, exit_test.go}` all use table-driven tests with fixtures, consistent with the project's `AGENTS.md` convention.

**Rationale**: matches the existing project style; `mise run test:coverage` validates ≥70% coverage per package.

### 10. No `main.go` changes in this change

**Choice**: `main.go` is left untouched in this change. The Factory, ExitCodeFor, and FormatError exist as units but `main.go` doesn't construct them yet.

**Rationale**: changes 2-7 each add consumers of these units; `main.go` rewiring happens alongside the first two pilot commands (change-2) with the `legacyBridge` shim pattern that allows non-migrated commands to coexist. Wiring `main.go` in this change without any consumer would be a half-finished integration.

## Risks / Trade-offs

- **Mechanical rename is large** — 12 + 15 = 27 files moved with import updates. **Mitigation:** tasks are sequenced so `mise run check` runs after each group; `git mv` preserves history.
- **`PartialSuccessError` is unused at archive time** — the type has tests but no production consumer in this change. **Mitigation:** the type is documented in `internal/core/errors.go` and referenced in the `cli/init-command` delta spec as "added in foundation; consumed by change-5".
- **Forbidigo patterns might miss future anti-patterns** — the patterns are chosen for current code shape. **Mitigation:** patterns can be extended in future changes without breaking existing rules.
- **Empty `internal/infrastructure/` directory must be removed cleanly** — if any file is missed (e.g., a doc.go that was added between planning and execution), the tree check in task 1.2a.7 catches it.
- **Renaming `internal/domain/` to `internal/core/` may break external consumers** — germinator is a developer-only CLI with no published API. **Mitigation:** the `type Domain = core` alias covers any external consumer; CI catches any internal breakage.

## Migration Plan (foundation only)

The full migration is sequenced as 9 changes (this is change 1 of 9). Each subsequent change:

1. Lands a coherent set of changes that pass `mise run check`.
2. Is independently mergeable; no change depends on a future change.
3. Updates the corresponding delta spec at the end (not the start), so the spec reflects reality.
4. Updates the location-specific `AGENTS.md` files at the end.

This change specifically:

1. Creates new packages with full tests (`mise run test:coverage` confirms ≥70%).
2. Performs mechanical renames (12 files renamed, 15 files flattened, every import updated).
3. Adds `PartialSuccessError` and `core/rules.go`.
4. Enables `forbidigo` and adds the lint-enforcement test.
5. Confirms `mise run check` is green and zero behavior change.

### Rollback strategy

This change is a single merge commit. To roll back, revert the commit; the next change (change-2: wire + pilots) does not exist yet, so no downstream dependency is broken.

## Open Questions

1. **Where should `core.CanInstallResource` live?** — **RESOLVED: deferred to change-6.** This change defines `core/rules.go` with `ValidatePlatform` and `ResolveOutputPath`. `CanInstallResource` is defined in change-6 alongside its first consumer (`library add`). Defining it here would be premature extraction.
2. **Should `cmd/lint_test.go` run `mise run lint` synchronously or in a subprocess?** — **RESOLVED: subprocess via `exec.Command("mise", "run", "lint")`.** The test fails the suite if the linter reports new violations. Synchronous `golangci-lint run` would require duplicating the configuration; subprocess approach tests the actual mise task.
