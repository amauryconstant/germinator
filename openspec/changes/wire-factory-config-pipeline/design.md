# Design — Wire `Factory.Config` and complete the koanf config pipeline

## Context

### Current state

The codebase has a structurally complete but functionally disconnected config pipeline:

- **`internal/config/`** (the package): contains a `Config` value type (`internal/config/config.go:12-33`), a `Manager` interface (`internal/config/manager.go:15-24`), a `koanfConfigManager` implementation (`internal/config/manager.go:27-36`), `DefaultConfig()` (`config.go:36-45`), `Validate()` (`config.go:49-63`), and an XDG-aware `resolveConfigPath` (`manager.go:93-121`). The Manager's `Load()` method unmarshals from a TOML config file at `$XDG_CONFIG_HOME/germinator/config.toml` (with `~/.config/germinator/config.toml` and `./config.toml` fallbacks).
- **`internal/cmdutil/factory.go:31`**: `Factory.Config func() (*config.Config, error)` is declared as a lazy field, matching the skill's Tier 2 layout (`golang-cli-architecture/references/01-architecture.md:224`).
- **`main.go:24-38`**: `main.go` constructs the Factory and assigns `f.Library` (via `cmdutil.OnceValuesFunc`) but **never assigns `f.Config`**. The declared field is nil.

### Constraints

1. **Spec/code drift in three places**:
   - `cli-cli-factory/spec.md:37` requires `Factory.Config` to be assigned.
   - `cli-cli-factory/spec.md:74` references a non-existent `internal/config.Load()` function.
   - `cli-shell-completion/spec.md:168-170, 178-182` promise that `completion.timeout` and `completion.cache_ttl` config keys take effect — but `cmd/completions.go:103, 111` passes `nil` to the helpers.
   - `cli-shell-completion/spec.md:305, 312` literally mandates `getCompletionTimeout(nil)` (spec-internal contradiction with lines 168-170).
2. **Existing precedence contract**: `application-configuration/spec.md:122` states: *"Library path resolution follows a parallel three-tier chain within the config layer: `--library` flag > `GERMINATOR_LIBRARY` env > config file > XDG default."* The migration must preserve this order — env vars win over `Config.Library`.
3. **Test stability**: 4 E2E test files (`test/e2e/library_add_test.go:230-239`, `library_discover_test.go:215-233`, `library_refresh_test.go:156-169`, `library_remove_test.go:217-230`) test `GERMINATOR_LIBRARY` directly. They must continue to pass without modification (or with backwards-compat assertions added).
4. **Backwards compat for `iostreams.System()`**: 50+ test files construct `iostreams.Test()` or `iostreams.System()` directly without a Factory. The debug-logger behavior (`os.LookupEnv("GERMINATOR_DEBUG")` at `iostreams.go:152`) must remain functional without requiring a Factory.

### Assumption

The slice-7 design.md note "the four service interfaces were removed from the Factory in slice 7.5 — their concrete adapters are now constructed lazily inside the per-command run functions" (`main.go:30-34`) was paired with `Factory.Config` being left unassigned. The most likely explanation is that the migration focused on the Transformer/Validator/Canonicalizer/Initializer service interfaces and accidentally left `Config` in a half-wired state. **Wiring `Factory.Config` is the natural next step** and aligns with the skill's canonical Tier 2 Factory pattern.

## Goals / Non-Goals

**Goals:**

- Extend `Manager.Load()` to install a `koanf/providers/env` provider with `GERMINATOR_` prefix, completing tier 3 of the documented four-tier merge.
- Add `Config.Library`, `Config.PlatformDefault`, `Config.Debug` fields. `DefaultConfig()` seeds sensible defaults; `Validate()` covers the new fields.
- Wire `Factory.Config` in `main.go` via `cmdutil.OnceValuesFunc` wrapping a new top-level `config.Load()` wrapper.
- Add `internal/config.Load() (*Config, error)` top-level function (fixes the broken reference at `cli-cli-factory/spec.md:74`).
- Migrate `GERMINATOR_LIBRARY` reads (13+ sites) to consult `Config.Library` while preserving env-var precedence (env wins).
- Migrate `GERMINATOR_DEBUG` to flow through `Config.Debug` via a new `iostreams.IOStreams.SetDebug(bool)` method called from `main.go` after config load. `iostreams.System()` retains its env-read for backwards compat.
- Add `Config.PlatformDefault` for future use (commands that take `--platform` may opt in via a follow-up change).
- Update `cmd/completions.go` to pass a real `*config.Config` (loaded via `f.Config()`) to `resolveLibraryPath`, `getCompletionTimeout`, `getCacheTTL`.
- Update 3 spec files to reflect the new implementation reality.

**Non-Goals:**

- Migrating `NO_COLOR` (out of scope; `cli-color-policy/spec.md` mandates direct env read).
- Migrating `EXIT_CODE_LEGACY` (out of scope; `cli-exit-codes/spec.md` defines this as a boot-time feature flag).
- Adding `adrg/xdg` for cross-platform XDG resolution (Windows support is a known limitation; the spec relaxation is the recommended path).
- Adding per-command opt-in for `Config.PlatformDefault` as a default for `--platform` (the field exists; commands may wire it in follow-up changes).
- Refactoring `iostreams.System()` to take a Factory (breaking change to a public-ish constructor).

## Decisions

### 1. Env provider at the Manager layer, not per-call-site

**Choice**: Install `koanf/providers/env` inside `Manager.Load()` with the `GERMINATOR_` prefix and the `_` delimiter (koanf's default snake_case conversion). The merge order becomes: defaults → file → env. Flags remain per-command (already handled by Cobra).

**Rationale**: The skill's config pipeline diagram (`golang-cli-architecture/references/01-architecture.md:182`) shows the four tiers as a single managed pipeline, not per-call-site aggregation. Centralizing the env-var resolution in the Manager layer matches the documented contract at `application-configuration/spec.md:113-122` and avoids 13+ scattered `os.Getenv` reads.

**Alternatives considered**:
- *Per-call-site `os.Getenv` aggregation*: rejected. This is the current state and the cause of the gap; preserving it does not close the spec/code drift.
- *Separate config package for runtime* (e.g., `internal/runtime/`): rejected. Adds a new package without architectural benefit; the existing `internal/config/` package is the right home.

### 2. `Config.Library` is a config-tier default; env-var still wins

**Choice**: The call pattern becomes `library.FindLibrary(opts.Library, os.Getenv("GERMINATOR_LIBRARY"))`. `library.FindLibrary` (at `internal/library/discovery.go:32`) already implements the `--library > env > config > default` precedence via its two-string signature.

**Rationale**: This is the cleanest expression of `application-configuration/spec.md:122`'s *"parallel three-tier chain within the config layer: `--library` flag > `GERMINATOR_LIBRARY` env > config file > XDG default"* — the `opts.Library` is the config-tier default, the env var is the env-tier override, and `library.FindLibrary` already merges them. The migration only requires threading `Config.Library` into the call sites; `library.FindLibrary`'s precedence logic is unchanged.

**Alternatives considered**:
- *Add `Config.Library` to the precedence chain inside `library.FindLibrary`*: rejected. Adds a third parameter to a function that already has two; bloats the call signature.
- *Drop `GERMINATOR_LIBRARY` env var entirely*: rejected. Breaks 4 E2E test files and the documented env-precedence contract.

### 3. `iostreams.System()` keeps env-read; new `SetDebug(bool)` for explicit override

**Choice**: Two options evaluated:

- **Option A**: Replace `iostreams.System()` with `iostreams.System(*cmdutil.Factory)`. Breaking change.
- **Option B**: Keep `iostreams.System()` unchanged (still reads `os.LookupEnv("GERMINATOR_DEBUG")` at construction); add `IOStreams.SetDebug(bool)` method that `main.go` calls after loading the config.

**Chosen**: Option B.

**Rationale**: Option A breaks 50+ test files that construct `iostreams.Test()` directly. Option B preserves backwards compat: tests that use `iostreams.Test()` get the default (no-debug); production `main.go` calls `SetDebug(cfg.Debug)` after loading the config. The env-var read remains the fallback for the test code path, and the config-driven path is the production path.

**Alternatives considered**:
- *Make `iostreams.System()` read `os.LookupEnv` only, remove the config path*: rejected. This is the current state and the cause of the gap.

### 4. `internal/config.Load()` as a top-level wrapper

**Choice**: Add `func Load() (*Config, error)` at `internal/config/load.go` (or extend `manager.go`). The wrapper calls `NewConfigManager().Load()` and returns `mgr.GetConfig()`.

**Rationale**: Fixes the broken reference at `cli-cli-factory/spec.md:74`. Simplifies `main.go` (one call instead of two). Provides a stable public API for future consumers (e.g., the `config validate` command could be migrated to use it).

**Alternatives considered**:
- *Fix the spec text only (rename `Load()` → `NewConfigManager().Load()`)*: rejected. The spec describes the contract from the consumer's perspective; the consumer wants a function named `Load()`. Adding the wrapper is cheaper than rewording the spec.
- *Move `Load()` to the `Manager` interface as the primary entry point*: rejected. The interface is for testability (allows mocking); the top-level wrapper is for convenience.

### 5. Relax `adrg/xdg` requirement in `application-configuration/spec.md`

**Choice**: Remove the `adrg/xdg` mandate from `application-configuration/spec.md:147-156`. Replace with: *"The loader SHALL resolve the config path via a platform-aware helper (`resolveConfigPath`); cross-platform XDG resolution MAY be added via `adrg/xdg` in a follow-up change."*

**Rationale**: The current `resolveConfigPath` (`internal/config/manager.go:93-121`) implements `$XDG_CONFIG_HOME` env read with `$HOME/.config` fallback, which works on Unix and macOS. Windows support is a known limitation already acknowledged in the codebase. Adding `adrg/xdg` is a separate concern (dependency addition) that requires a follow-up change with its own risk assessment.

**Alternatives considered**:
- *Add `adrg/xdg` to `go.mod` and migrate `resolveConfigPath`*: rejected for this change scope. A follow-up change adds the dependency after risk assessment (license, transitive deps, Windows coverage verification).

### 6. Spec deltas for 3 specs

**Choice**: Modify 3 existing specs via delta files:

- **`cli-cli-factory/spec.md`**: replace `internal/config.Load()` reference at line 74 with the new top-level wrapper; add a scenario documenting the `Load()` wrapper contract.
- **`cli-shell-completion/spec.md`**: change lines 305, 312 (`getCompletionTimeout(nil)`) to `getCompletionTimeout(f.Config())`. The two existing "configurable" scenarios (lines 168-170, 178-182) become test-passing AND spec-consistent.
- **`application-configuration/spec.md`**: relax the `adrg/xdg` mandate (Decision 5); add a requirement documenting the env-provider layer.

**Rationale**: Each spec describes user-visible behavior (or developer-visible contract). The current spec/code drift is a maintenance liability — `cli-shell-completion/spec.md:168-170` claims configurable timeouts that don't actually work, and `cli-shell-completion/spec.md:305, 312` mandates the buggy `nil` argument. Fixing the spec/code drift is part of the change's value.

**Alternatives considered**:
- *Leave specs unchanged*: rejected. The drift is the core problem; documenting the gap would compound it.
- *Rewrite specs from scratch*: rejected. The delta approach preserves the spec history (readers can see what changed).

## Risks / Trade-offs

- **13+ call sites change for `GERMINATOR_LIBRARY` migration**: one missed site silently reverts to env-only behavior. **Mitigation**: task `2.6` runs `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)"` after migration; production code must have zero matches (test files may keep it as backwards-compat proof).
- **`Config.Debug` is an additive public API change**: `Config` is a public struct exported from `internal/`. Adding a field does not break existing consumers but is visible to anyone with field-name struct literals. **Mitigation**: `grep` confirms zero struct literals with field names in `cmd/`; the additive change is safe.
- **`iostreams.SetDebug` API addition**: backwards-compatible (additive method on existing struct). **Mitigation**: all 50+ test files continue to work without modification.
- **`adrg/xdg` spec relaxation**: Windows users may experience less-tested XDG path resolution. **Mitigation**: documented limitation; Windows support is a known gap; a follow-up change adds `adrg/xdg` if/when Windows support becomes a priority.
- **Env provider introduces key-mapping complexity**: koanf's env provider uses `_` as the default delimiter and lowercases keys. The `Config.PlatformDefault` field maps to env `GERMINATOR_PLATFORM_DEFAULT`, not `GERMINATOR_PLATFORM`. **Mitigation**: document the mapping in `internal/config/manager.go` and the `Config` field godoc; the env-provider layer's key mapping is tested in `manager_test.go`.
- **Completion timeout/TTL change may surprise users who set completion cache**: if a user authors `~/.config/germinator/config.toml` with `completion.timeout = "1s"`, completion latency increases from 500ms to 1s. **Mitigation**: `DefaultConfig()` keeps `Timeout: "500ms"` (zero-config users unaffected); documented as opt-in.

## Migration Plan

The change is applied in one PR with the following atomic phases (each commit is independently testable):

1. **Phase 1 — Extend the koanf loader** (tasks 2.1-2.4): add env provider, add fields, add `Load()` wrapper. Verify `internal/config` tests pass.
2. **Phase 2 — Wire `Factory.Config`** (tasks 2.5-2.7): assign `f.Config` in `main.go`; add `iostreams.SetDebug`; migrate `GERMINATOR_DEBUG`. Verify `mise run build`.
3. **Phase 3 — Migrate runtime reads** (tasks 2.8-2.10): update 10 cmd files to use `Config.Library`; thread `Config` into `cmd/completions.go`. Verify `mise run test` and `mise run test:e2e`.
4. **Phase 4 — Spec deltas + tests** (tasks 2.11-2.14): write delta specs; add new tests; refresh E2E for backwards-compat. Verify `openspec validate wire-factory-config-pipeline --strict`.

**Rollback strategy**: revert each phase commit independently. Phase 1 is non-breaking (additive). Phase 2's `SetDebug` method is additive (existing code unaffected). Phase 3's call-site changes are mechanical (each commit is independently testable). Phase 4 is doc-only (no behavior change).
