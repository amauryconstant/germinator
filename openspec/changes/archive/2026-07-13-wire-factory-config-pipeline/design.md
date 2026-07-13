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
3. **Test stability**: 6 E2E test files (`test/e2e/library_add_test.go:230-239`, `test/e2e/library_discover_test.go:215-233`, `test/e2e/library_refresh_test.go:156-169`, `test/e2e/library_remove_test.go:217-230`, `test/e2e/library_test.go:84`, `test/e2e/init_test.go:26`) test `GERMINATOR_LIBRARY` directly. They must continue to pass without modification (or with backwards-compat assertions added).
4. **`iostreams.System()` simplification — env-read removed**: 50+ test files construct `iostreams.Test()` directly (the production-side `iostreams.System()` constructor is used by `main.go` only). The decision drops the `os.LookupEnv("GERMINATOR_DEBUG")` branch so `iostreams.System()` always returns a discard-debug `Logger`. Production activates debug via `main.go`'s `IOStreams.SetDebug(cfg.Debug)` after `config.Load()`. Test files are unaffected (`iostreams.Test()` already returns a discard logger). External code (none in-repo) that imported `iostreams.System()` and relied on env-driven debug activation would lose that activation path; they would migrate to `iostreams.System()` + `s.SetDebug(true)` — explicit over implicit, per the skill.

### Assumption

The slice-7 design.md note "the four service interfaces were removed from the Factory in slice 7.5 — their concrete adapters are now constructed lazily inside the per-command run functions" (`main.go:30-34`) was paired with `Factory.Config` being left unassigned. The most likely explanation is that the migration focused on the Transformer/Validator/Canonicalizer/Initializer service interfaces and accidentally left `Config` in a half-wired state. **Wiring `Factory.Config` is the natural next step** and aligns with the skill's canonical Tier 2 Factory pattern.

## Goals / Non-Goals

**Goals:**

- Extend `Manager.Load()` to install a `koanf/providers/env` provider with `GERMINATOR_` prefix, completing tier 3 of the documented four-tier merge.
- Rename existing `Config.Platform` → `Config.PlatformDefault` (koanf tag unchanged). Add `Config.Debug bool` (new field). `Config.Library` already exists. **`DefaultConfig()` seeds `Library: ""`** so `Config.Library == ""` is the canonical "no config-file override" signal and the call path falls through to `DefaultLibraryPath()` (XDG via `adrg/xdg`). `Validate()` covers the renamed and new fields.
- Wire `Factory.Config` in `main.go` via `cmdutil.OnceValuesFunc` wrapping a new top-level `config.Load()` wrapper.
- Add `internal/config.Load() (*Config, error)` top-level function (fixes the broken reference at `cli-cli-factory/spec.md:74`). Contract: `*Config` is always non-nil; on error the chain (`*core.FileError` / `*core.ParseError` / `*core.ConfigError` matched via `errors.As`) is the authoritative signal.
- **Extend `library.FindLibrary(flag, env)` to 3 args** (`FindLibrary(flag, env, cfg)`) encoding the spec-mandated precedence (flag > env > config > default) directly. Migrate all 11 production call sites (9 cmd files + 2 env-once closures in `library_refresh.go` / `library_remove.go`) to the 3-arg form, threading `Config.Library` through. The `--library` flag remains honored (Decision 2).
- Migrate `GERMINATOR_DEBUG` to flow through `Config.Debug` via `iostreams.IOStreams.SetDebug(bool)` called from `main.go` after config load. **Remove** the `os.LookupEnv` branch from `iostreams.System()` so debug activation is single-source-of-truth (Decision 3).
- **Adopt `adrg/xdg`** (Decision 5) by adding it to `go.mod`, migrating `internal/config/manager.go::resolveConfigPath` to `xdg.ConfigFile(...)`, and migrating `internal/library/discovery.go::DefaultLibraryPath` to `xdg.DataFile(...)`. Preserves the `./config.toml` and `./germinator/library/` working-directory fallbacks.
- Add `Config.PlatformDefault` for future use (commands that take `--platform` may opt in via a follow-up change).
- Update `cmd/completions.go` to pass a real `*config.Config` (loaded via `f.Config()`) to `resolveLibraryPath`, `getCompletionTimeout`, `getCacheTTL`.
- Update 3 spec files (`cli-cli-factory`, `cli-shell-completion`, `application-configuration`) — the third modifies the `Config field set` MODIFIED requirement to document `Library: ""` and carves out `actionPlatforms` from the two-arg action-function shape (pre-existing drift documented in the spec).

**Non-Goals:**

- Migrating `NO_COLOR` (out of scope; `cli-color-policy/spec.md` mandates direct env read).
- Migrating `EXIT_CODE_LEGACY` (out of scope; `cli-exit-codes/spec.md` defines this as a boot-time feature flag).
- Setting up a Windows CI matrix (Windows test coverage for `adrg/xdg` is deferred per user decision; tests tagged `//go:build !windows`).
- Adding per-command opt-in for `Config.PlatformDefault` as a default for `--platform` (the field exists; commands may wire it in follow-up changes).
- Refactoring `iostreams.System()` to take a Factory (breaking change to a public-ish constructor).
- Rewriting `internal/config/manager.go::Load()` to drop the bespoke `./config.toml` CWD fallback (project-local configs still win over the user's global config).

## Decisions

### 1. Env provider at the Manager layer, not per-call-site

**Choice**: Install `koanf/providers/env` inside `Manager.Load()` with the `GERMINATOR_` prefix and the `_` delimiter (koanf's default snake_case conversion). The merge order becomes: defaults → file → env. Flags remain per-command (already handled by Cobra).

**Rationale**: The skill's config pipeline diagram (`golang-cli-architecture/references/01-architecture.md:182`) shows the four tiers as a single managed pipeline, not per-call-site aggregation. Centralizing the env-var resolution in the Manager layer matches the documented contract at `application-configuration/spec.md:113-122` and avoids 11 scattered `os.Getenv` reads.

**Alternatives considered**:
- *Per-call-site `os.Getenv` aggregation*: rejected. This is the current state and the cause of the gap; preserving it does not close the spec/code drift.
- *Separate config package for runtime* (e.g., `internal/runtime/`): rejected. Adds a new package without architectural benefit; the existing `internal/config/` package is the right home.

### 2. `library.FindLibrary` extended to 3 args so the config tier slots in without dropping the flag tier

**Choice**: Change `internal/library/discovery.go::FindLibrary(flagPath, envPath string) string` → `FindLibrary(flagPath, envPath, cfgPath string) string`. The new implementation encodes the spec-mandated precedence directly:

```go
func FindLibrary(flagPath, envPath, cfgPath string) string {
    if flagPath != "" { return flagPath }   // --library flag (highest priority)
    if envPath  != "" { return envPath  }   // GERMINATOR_LIBRARY env
    if cfgPath  != "" { return cfgPath  }   // Config.Library (config-file)
    return DefaultLibraryPath()             // XDG via adrg/xdg
}
```

The call pattern at each of the 11 production sites becomes `library.FindLibrary(flagValue, os.Getenv("GERMINATOR_LIBRARY"), opts.ConfigLibraryPath)` where `opts.ConfigLibraryPath string` is a new field on each command's options struct populated from `cfg.Library` in `RunE`. The naming `ConfigLibraryPath` (not `Library`) avoids shadowing the existing `opts.Library func() (*library.Library, error)` lazy closure on the same options struct. The two env-once closures in `cmd/library_refresh.go:151` and `cmd/library_remove.go:231` are also migrated to call `cfg, _ := f.Config()` and pass `cfg.Library` as the third arg, so precedence is uniform across all 11 sites.

**Rationale**: `application-configuration/spec.md:122` mandates *"`--library` flag > `GERMINATOR_LIBRARY` env > config file > XDG default"*. The current 2-arg `FindLibrary(flag, env)` cannot encode all four tiers; shoehorning `cfg.Library` into the flag slot would silently drop the `--library` flag. Adding the third positional argument is the smallest extension that preserves the full precedence contract and the existing `cmd.Flag("library")` lookup pattern. Both env-once closures are migrated here, not kept as "capture-once-then-FindLibrary" exceptions, because the spec contract is the same for `library refresh` and `library remove` as for every other command; asymmetries are harder to defend than uniform precedence.

**Alternatives considered**:
- *Promote `cmd/completions.go:47-65::resolveLibraryPath` to `internal/library/` and route all call sites through it*: rejected. `resolveLibraryPath` reads the flag via `cmd.Flag("library")` (a Cobra-only API), which the env-once closures in `cmd/library_refresh.go` / `cmd/library_remove.go` cannot call. A 3-arg string-based signature works for both call patterns without dragging `*cobra.Command` into the library package.
- *Keep 2-arg `FindLibrary`; coalesce flag + cfg into a single string in `RunE` (`opts.ConfigLibraryPath = lp if lp != "" else cfg.Library`)*: rejected. Yields precedence `cfg > env > default` (the wrong order per the spec) and still loses the flag-vs-config distinction once flag is dropped.
- *Drop `GERMINATOR_LIBRARY` env var entirely*: rejected. Breaks 6 E2E test files and the documented env-precedence contract.

### 3. `iostreams.System()` env-read removed; `SetDebug(bool)` is the single source of truth for debug activation

**Choice**: Remove the `os.LookupEnv("GERMINATOR_DEBUG")` branch from `newDebugLogger` (currently `iostreams.go:151-156`) so `iostreams.System()` always returns a discard-debug `Logger`. Add `IOStreams.SetDebug(bool)` as the canonical activation point. `main.go` calls `io.SetDebug(cfg.Debug)` after `config.Load()` succeeds. `cfg.Debug` is itself populated by the koanf env provider (`GERMINATOR_DEBUG=1` → `cfg.Debug == true`), so the end-to-end behavior — env-var debug activation — is preserved through the new pipeline.

**Rationale**: Two paths reading the same env var (`iostreams.System()` env-read at construction + `main.go`'s `SetDebug(cfg.Debug)`) is dual-source-of-truth for observability channels, per `golang-cli-architecture` §"Two distinct channels for observability" (verbose + debug). The skill calls for explicit, deterministic activation in a single place. Removing the env-read at construction makes `iostreams.System()` pure (the env doesn't influence `Logger`), and `SetDebug(cfg.Debug)` becomes the deterministic point whose input (`cfg.Debug`) itself flows through the documented four-tier config precedence. The 50+ test files that use `iostreams.Test()` are unaffected (it already returns the discard logger); `iostreams.System()` users in this repo are the production `main.go` only (verified via `rg "iostreams\.System" cmd/ internal/`).

**Alternatives considered**:
- *Keep dual path (Option B from the prior revision)*: rejected. Two env-var reads for the same value is redundant; under future divergence they could disagree silently.
- *Replace `iostreams.System()` with `iostreams.System(*cmdutil.Factory)`*: rejected. Tests that construct `iostreams.System()` directly would have to take a Factory, which is a breaking-change surface area beyond what the change needs.

### 4. `internal/config.Load()` as a top-level wrapper in `internal/config/load.go`

**Choice**: Add `func Load() (*Config, error)` at the new file `internal/config/load.go` (not a method on `Manager`). The wrapper uses a package-level function variable for test injection (per golang-design-patterns "design for testability"):

```go
var loadFn = NewConfigManager

func Load() (*Config, error) {
    mgr := loadFn()
    if err := mgr.Load(); err != nil {
        return mgr.GetConfig(), err
    }
    return mgr.GetConfig(), nil
}
```

`loadFn` is a package-level variable (not `var loadFn = NewConfigManager().Load`) so tests can substitute a stub Manager — e.g., `loadFn = func() config.Manager { return &stubManager{...} }` — without re-running the Manager constructor. `mgr.GetConfig()` is always non-nil because `NewConfigManager()` seeds the underlying struct with `DefaultConfig()`.

**Rationale**: Fixes the broken reference at `cli-cli-factory/spec.md:74`. Simplifies `main.go` (one call from the composition root — `f.Config = cmdutil.OnceValuesFunc(config.Load); cfg, err := f.Config()` — instead of two chained calls). Provides a stable public API for future consumers (e.g., the `config validate` command could be migrated to use it). The function-variable seam lets tests inject stub managers without depending on `Manager` internals.

**Alternatives considered**:
- *Fix the spec text only (rename `Load()` → `NewConfigManager().Load()`)*: rejected. The spec describes the contract from the consumer's perspective; the consumer wants a function named `Load()`. Adding the wrapper is cheaper than rewording the spec.
- *Move `Load()` to the `Manager` interface as the primary entry point*: rejected. The interface is for testability (allows mocking); the top-level wrapper is for convenience.
- *Hard-code `NewConfigManager()` inside `Load()` (no function variable)*: rejected. Tests of `Load()` would need to touch the filesystem or stub koanf providers; a function-variable seam lets tests inject a stub Manager directly. The cost is one package-level mutable variable, which is acceptable for the test-injection purpose (documented in the code comment).

### 5. Adopt `adrg/xdg` to satisfy the existing spec contract

**Choice**: Add `github.com/adrg/xdg` to `go.mod`. Replace `internal/config/manager.go::resolveConfigPath` (lines 93-121) with a thin call to `xdg.ConfigFile("germinator/config.toml")`, preserving the working-directory `./config.toml` fallback for projects that ship their own config file alongside `germinator`. Replace `internal/library/discovery.go::DefaultLibraryPath` (lines 24-51) with `xdg.DataFile("germinator/library")`, preserving the working-directory `./germinator/library/` last-resort fallback.

**Rationale**: The source-of-truth `openspec/specs/application-configuration/spec.md:145-156` mandates `github.com/adrg/xdg`. Adopting the dependency satisfies the existing contract and delivers **runtime** cross-platform coverage (Linux, macOS, Windows) — `adrg/xdg.ConfigFile` and `adrg/xdg.DataFile` resolve correctly on Windows via `%AppData%`/`%LocalAppData%` — without the bespoke env-var ladders currently duplicated in `resolveConfigPath` and `DefaultLibraryPath`. The CWD fallbacks are not covered by `adrg/xdg` and stay as explicit project-local overrides. **Windows CI test coverage is deferred** per tasks 2.2/2.3 (tests tagged `//go:build !windows`); the runtime paths are covered by `adrg/xdg`'s own test suite.

**Alternatives considered**:
- *Keep the bespoke paths and relax the spec to drop the `adrg/xdg` mandate*: rejected. The user chose to fulfill the spec as written; Windows support is delivered in this change.
- *Adopt `adrg/xdg` for `resolveConfigPath` only, defer `DefaultLibraryPath` to a follow-up*: rejected. The spec's library-path scenario (`library.DefaultLibraryPath()` returning `adrg/xdg.DataFile(...)`) requires both paths migrated; splitting leaves the spec half-satisfied.
- *Adopt `adrg/xdg` and drop the `./config.toml` CWD fallback*: rejected. Some workflows expect the project's checked-in config to take precedence over the user's global config.

### 6. Spec deltas for 2 specs + Load() error-semantics pin

**Choice**: Modify 2 existing specs via delta files (the third spec's text does not change):

- **`cli-cli-factory/spec.md`**: replace `internal/config.Load()` reference at line 74 with the new top-level wrapper; expand the `Load()` scenario (lines 51-57) to (a) pin the never-nil `*Config` contract, (b) replace the blanket `*core.ConfigError` with the typed chain `*core.FileError` / `*core.ParseError` / `*core.ConfigError` matching the failure mode, all dispatched via `errors.As` by `output.FormatError`.
- **`cli-shell-completion/spec.md`**: change lines 305, 312 (`getCompletionTimeout(nil)`) to `getCompletionTimeout(f.Config())`. The two existing "configurable" scenarios (lines 168-170, 178-182) become test-passing AND spec-consistent.
- **`application-configuration/spec.md`** (source-of-truth text stands): the existing `adrg/xdg` requirement (lines 145-156) is the contract this change adopts (Decision 5). The delta splits the requirement set into two sections: an `## ADDED Requirements` block containing the new `Config field set` requirement (no matching requirement in the source-of-truth), and a `## MODIFIED Requirements` block containing `Configuration precedence` and `Environment variable naming` (both have matching names in the source-of-truth). The previously-planned MODIFIED-Requirement block for `XDG resolution via adrg/xdg` is dropped — the source-of-truth text stands unchanged.

**Rationale**: Each spec describes user-visible behavior (or developer-visible contract). The current spec/code drift is a maintenance liability — `cli-shell-completion/spec.md:168-170` claims configurable timeouts that don't actually work, and `cli-shell-completion/spec.md:305, 312` mandates the buggy `nil` argument. Fixing the spec/code drift is part of the change's value. The `cli-cli-factory` scenario expansion (never-nil + typed error chain) makes the `Load()` contract machine-checkable downstream: any future consumer can rely on the type shape without re-reading `internal/core/errors.go`.

**Alternatives considered**:
- *Leave specs unchanged*: rejected. The drift is the core problem; documenting the gap would compound it.
- *Rewrite specs from scratch*: rejected. The delta approach preserves the spec history (readers can see what changed).
- *Relax `adrg/xdg` in the source of truth* (prior revision of Decision 5): rejected per user decision to fulfill the spec as written.

## Risks / Trade-offs

- **11 production-code call sites change for `GERMINATOR_LIBRARY` migration**: one missed site silently reverts to env-only behavior. **Mitigation**: task 4.3 runs `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)" cmd/ internal/ --type go` after migration; the only surviving match is `cmd/completions.go:54` (the priority-chain helper inside `resolveLibraryPath` — intentional per Decision 2 and unchanged by this change). All 11 sites, including the two env-once closures in `cmd/library_refresh.go:151` and `cmd/library_remove.go:231`, are migrated to the 3-arg `FindLibrary(flag, env, cfg)` form.
- **`Config.Debug` is an additive public API change**: `Config` is a public struct exported from `internal/`. Adding a field does not break existing consumers but is visible to anyone with field-name struct literals. **Mitigation**: `rg "Debug:\s+" cmd/ internal/` returns zero hits — all `Config` initializations go through `DefaultConfig()`; the additive change is safe.
- **`Config.Platform` rename to `Config.PlatformDefault`** is a breaking change for any caller that initializes `Config{}` with the field name `Platform`. **Mitigation**: `rg "Platform:\s+" cmd/ internal/` returns zero hits in-repo; external consumers (none in-repo) should rebuild on next go-get.
- **`iostreams.SetDebug` API addition** + `os.LookupEnv` branch removal. `SetDebug` is additive (new method on existing struct). Removing the env-read means external code that imported `iostreams.System()` and relied on env-driven debug activation would lose that path — there is no such consumer in-repo (verified via `rg "iostreams\.System" --files-without-match '_test\.go'`). Production activates debug via `main.go`'s `io.SetDebug(cfg.Debug)` after `config.Load()` succeeds. Test 7.4 (`TestSystem_NoLongerReadsEnvDebug`) asserts the new behavior.
- **`adrg/xdg` dependency adoption** — three sub-risks:
  1. **Transitive-dep surprise**: pin a semver and run `go mod verify`; review `go.sum` after `go mod tidy`. The package is small and zero-dep upstream, but a downstream regression in `adrg/xdg` itself would land in this repo's CI.
  2. **Platform-specific behavior under non-standard `XDG_*` paths**: `adrg/xdg` reads `XDG_CONFIG_HOME` / `XDG_DATA_HOME` and falls back to OS-appropriate defaults. Edge cases (empty `XDG_DATA_HOME`, non-existent home) need explicit test coverage. **Mitigation**: integration tests in tasks 2.2 / 2.3 tagged `//go:build !windows` (per user decision — Windows CI deferred).
  3. **Cross-package change**: `DefaultLibraryPath` is in `internal/library/`, not `internal/config/`; the migration touches a different package than the rest of this change. Five known callers (`cmd/init.go:139`, `cmd/completions.go:64`, plus internal tests) are unaffected by signature changes; behavior preservation is verified by the existing `internal/library/discovery_test.go` suite plus the new `TestDefaultLibraryPath_AdoptsXDG`.
- **Env provider introduces key-mapping complexity**: koanf's env provider uses `_` as the default delimiter and lowercases keys after stripping the `GERMINATOR_` prefix. The `Config.PlatformDefault` field (koanf key `platform`) therefore maps to env `GERMINATOR_PLATFORM`, not `GERMINATOR_PLATFORM_DEFAULT`. **Truthiness rule for bool fields**: koanf parses bool values via its own coerce path — `1` / `t` / `T` / `true` / `TRUE` / `True` resolve to `true`; all other non-empty strings resolve to `false`; unset defaults to the struct default (per `strconv.ParseBool` semantics). **Mitigation**: document both the key mapping and the bool truthiness rule in `internal/config/manager.go` and the `Config` field godoc; the env-provider layer's key mapping is tested in `manager_test.go` (task 2.1).
- **Completion timeout/TTL change may surprise users who set completion cache**: if a user authors `~/.config/germinator/config.toml` with `completion.timeout = "1s"`, completion latency increases from 500ms to 1s. **Mitigation**: `DefaultConfig()` keeps `Timeout: "500ms"` (zero-config users unaffected); documented as opt-in.
- **Main.go `Load()` fail-fast**: surfacing `Load()` errors as exit-1 means a malformed `config.toml` (previously silently swallowed) now blocks CLI invocation. **Mitigation**: documented as a deliberate behavior change (`cli-exit-codes/spec.md` already maps `*core.*Error` to exit 1 via `cmdutil.ExitCodeFor`); users can `unset GERMINATOR_*` and remove `config.toml` to recover. Per `golang-cli-architecture` "main.go is the only composition root" + `golang-error-handling` "Returned errors MUST always be checked".

## Migration Plan

The change is applied in one PR with the following atomic phases (each commit is independently testable; task numbering refers to the actual `tasks.md` sections):

1. **Phase 1 — Adopt `adrg/xdg` + extend the koanf loader** (tasks 1.1–1.9 + 2.1–2.4): add `adrg/xdg` dep, migrate `resolveConfigPath` and `DefaultLibraryPath`, rename `Platform` → `PlatformDefault`, add `Debug`, set `Library: ""` in `DefaultConfig()`, install env provider, add `Load()` wrapper at `internal/config/load.go`, extend `FindLibrary` to 3 args, add path tests. Verify `go mod tidy && go mod verify` and `mise run test` pass.
2. **Phase 2 — Wire `Factory.Config`** (tasks 3.1–3.3): assign `f.Config` in `main.go` with fail-fast error handling (per `golang-error-handling` Rule 1); add `iostreams.SetDebug` and remove the env-read in `System()` (Decision 3). Verify `mise run build` and `mise run test`.
3. **Phase 3 — Migrate runtime reads** (tasks 4.1–4.4 + 5.1–5.3): update all 11 production call sites (9 cmd files plus the 2 env-once closures in `library_refresh.go` / `library_remove.go`) to use `Config.Library` via the 3-arg `FindLibrary(flag, env, cfg)`; thread `Config` into `cmd/completions.go`. Verify `mise run test` and `mise run test:e2e`.
4. **Phase 4 — Spec deltas + tests** (tasks 6.1–6.4 + 7.1–7.8 + 8.1–8.7): finalize the 3 modified specs (`cli-cli-factory`, `cli-shell-completion`, `application-configuration`); add new tests for completion timeout/TTL, `SetDebug`, `System()` env-read removal, `FindLibrary` 3-arg precedence, refresh/remove closure migration, and `adrg/xdg` paths; refresh E2E for backwards-compat (6 files). Verify `openspec validate wire-factory-config-pipeline --strict`.

**Rollback strategy**: revert each phase commit independently. Phase 1 is mostly additive (one rename + one dep). Phase 2's `SetDebug` method is additive; the env-read removal is behavior-preserving (env → `cfg.Debug` → `SetDebug` chain). Phase 3's call-site changes are mechanical (each commit is independently testable). Phase 4 is doc-only (no behavior change).
