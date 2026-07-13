# Wire `Factory.Config` and complete the koanf config pipeline

## Why

`Factory.Config func() (*config.Config, error)` is declared at `internal/cmdutil/factory.go:31` and required by `openspec/specs/cli-cli-factory/spec.md:37`, but is **never assigned in `main.go`** (`main.go:24-38`). The koanf-based loader at `internal/config/manager.go` implements only the first two tiers (defaults → file) of the documented four-tier merge (defaults → file → env → flags); env vars are read directly via `os.Getenv` at 11 production call sites, and the third tier (env) is absent at the Manager layer.

Three downstream consequences:

1. **`cmd/completions.go:103, 111` pass `nil`** to `getCompletionTimeout` and `getCacheTTL`, breaking `cli-shell-completion/spec.md:168-170` and `cli-shell-completion/spec.md:178-182` which promise that `completion.timeout` and `completion.cache_ttl` in `~/.config/germinator/config.toml` will take effect.
2. **`cmd/completions.go:120, 132, 144` pass `nil`** to `resolveLibraryPath`, breaking the "Resolve library from config" scenario at `cli-shell-completion/spec.md:203-210`.
3. **`cli-cli-factory/spec.md:74`** references a non-existent `internal/config.Load()` function.

The `application-configuration/spec.md:113-122` four-tier precedence is also unmet: the loader stops at tier 2 and per-call-site `os.Getenv` reads handle the env tier ad-hoc without going through the documented pipeline.

Closing this gap requires:
- Extending the koanf `Manager.Load()` to install an env provider (tier 3).
- Renaming `Config.Platform` → `Config.PlatformDefault` (koanf tag unchanged) and adding `Config.Debug bool` (new field; `Config.Library` already exists).
- Wiring `Factory.Config` in `main.go` via `cmdutil.OnceValuesFunc` wrapping `config.Load()` — with fail-fast error handling at startup (per `golang-error-handling` Rule 1).
- Threading the loaded `Config` into the runtime call sites that currently pass `nil`.
- Migrating the runtime env-var reads (`GERMINATOR_LIBRARY`, `GERMINATOR_DEBUG`, `GERMINATOR_PLATFORM`) to the documented four-tier pipeline; `NO_COLOR` and `EXIT_CODE_LEGACY` remain direct (per their respective specs).
- Adding the top-level `internal/config.Load()` wrapper referenced (but missing) at `cli-cli-factory/spec.md:74`.
- Adopting `github.com/adrg/xdg` to satisfy the spec mandate at `application-configuration/spec.md:145-156` — migrating `internal/config/manager.go::resolveConfigPath` and `internal/library/discovery.go::DefaultLibraryPath` accordingly.

## What Changes

### A. Extend the koanf loader

- **MODIFY** `internal/config/manager.go`: add `koanf/providers/env` with `GERMINATOR_` prefix to the `Load()` merge order. Merge order becomes: defaults → file → env → (flags handled per-command by Cobra, which already overrides per spec).
- **MODIFY** `internal/config/config.go`:
  - Rename existing `Platform string` (koanf tag `platform`) to `PlatformDefault string` (same `koanf:"platform"` tag). The existing struct tag stays bound to the same config key. The renamed field is intended for future `--platform`-defaulting wiring (out of scope for this change; see Non-Goals in design.md).
  - Add field `Debug bool` (with `koanf:"debug"` tag) for debug-level structured logging, controlled by `GERMINATOR_DEBUG` env var or the `debug` config-file key.
  - `Library string` already exists at `internal/config/config.go:14` and is unchanged. The bullet list above documents the modifications only; do not add a duplicate `Library` field.
  - Update `DefaultConfig()` to seed `PlatformDefault: ""` and `Debug: false` (Library stays at the existing `"~/.config/germinator/library"` default until `DefaultLibraryPath` migration in task 1.8 takes over); extend `Validate()` to accept `PlatformDefault` (empty / `claude-code` / `opencode`) and `Debug` (always valid).
- **ADD** `internal/config/load.go` (or extend `manager.go`): export `func Load() (*Config, error)` as a top-level wrapper around `NewConfigManager().Load()`. Fixes the broken reference at `cli-cli-factory/spec.md:74`.

### B. Wire `Factory.Config` (fail-fast + single source of truth)

- **MODIFY** `main.go`: after `f.Library = ...` (around line 38), wire `Factory.Config` via the single-call pattern — `f.Config = cmdutil.OnceValuesFunc(config.Load); cfg, err := f.Config()`; on error → `output.FormatError(f.IOStreams, err); os.Exit(int(cmdutil.ExitCodeFor(err)))`. Then call `io.SetDebug(cfg.Debug)`. The single `f.Config()` invocation populates the cache; subsequent calls from completion actions return the cached `*Config`.
- **MODIFY** `internal/iostreams/iostreams.go` (lines 151-156): **remove** the `os.LookupEnv("GERMINATOR_DEBUG")` branch from `newDebugLogger`. `iostreams.System()` always returns a discard-debug `Logger`; debug activation is now single-source-of-truth via `IOStreams.SetDebug(bool)`, called from `main.go` after `config.Load()`. Behavior is preserved end-to-end: `GERMINATOR_DEBUG=1` env → koanf env provider sets `cfg.Debug = true` → `SetDebug(true)` activates the same handler the old env branch would have. Update the `System()` docstring (lines 34-37). See `design.md` Decision 3.

### C. Migrate runtime env-var reads to `Config`

| Env var | Migration |
|---|---|
| `GERMINATOR_LIBRARY` | Migrate to `Config.Library` via the **3-arg** `library.FindLibrary(flag, env, cfg)`. Precedence per `application-configuration/spec.md:122`: flag > env > config > default. |

**Current sites** (11 production `os.Getenv` calls):

`main.go:36`, `cmd/show.go:87`, `cmd/resources.go:78`, `cmd/presets.go:68`, `cmd/init.go:139`, `cmd/library_add.go:251`, `cmd/library_create.go:155`, `cmd/library_refresh.go:151`, `cmd/library_remove.go:231`, `cmd/library_validate.go:156`, plus the priority-chain helper at `cmd/completions.go:54` (stays direct by design — already 4-tier aware).

**Closure body migration**: the two env-once closures (`library_refresh.go:151`, `library_remove.go:231`) currently capture env into a closure for late use but do not thread `Config.Library` through. Decision 2's 3-arg `FindLibrary` migration closes this asymmetry.

**New options field**: `opts.ConfigLibraryPath string` populated from `cfg.Library` in `RunE`. Naming `ConfigLibraryPath` (not `opts.Library`) avoids shadowing the existing `opts.Library func() (*library.Library, error)` lazy closure on the same options struct.

**Call pattern**: `library.FindLibrary(flagValue, os.Getenv("GERMINATOR_LIBRARY"), opts.ConfigLibraryPath)`.
| `GERMINATOR_DEBUG` | `internal/iostreams/iostreams.go:152` (the env-read inside `newDebugLogger`) | **Remove** the env-read entirely (see Decision 3 + Section B). The Logger is now a constant discard-debug handler at construction. Production activates debug via `main.go`'s `io.SetDebug(cfg.Debug)` after `config.Load()` succeeds. `iostreams.Test()` is unaffected (it already returns the discard logger). The 50+ test files that use `iostreams.Test()` or `iostreams.System()` continue to compile and pass without changes — `SetDebug` is opt-in from production code only. |
| `GERMINATOR_PLATFORM` | Never read | Add `Config.PlatformDefault` field; wire as a default for commands that take a `--platform` flag. Today every command requires `--platform` explicitly (`cmd/AGENTS.md:46-51`); this change adds the option to default it from the config (out-of-scope commands may opt in via a follow-up change). |
| `NO_COLOR` | `internal/iostreams/styles.go:10` (constant) | **Stay direct**. The `cli-color-policy/spec.md` explicitly mandates direct env read; out of scope. |
| `EXIT_CODE_LEGACY` | `internal/warning/canary.go:40` | **Stay direct**. Boot-time feature flag, not user config; out of scope per `cli-exit-codes/spec.md`. |

### D. Thread `Config` through completion actions

- **MODIFY** `cmd/completions.go:103, 111`: replace `getCompletionTimeout(nil)` and `getCacheTTL(nil)` with `getCompletionTimeout(cfg)` and `getCacheTTL(cfg)` where `cfg` is loaded via `f.Config()`.
- **MODIFY** `cmd/completions.go:120, 132, 144`: replace `resolveLibraryPath(cmd, nil)` with `resolveLibraryPath(cmd, cfg)` for `actionResources`, `actionPresets`, `actionLibraryRefs`. The `resolveLibraryPath` helper at `cmd/completions.go:47-65` already honors `cfg.Library` at line 59 — the production callers just need to pass a non-nil `cfg`.

### E. Spec/code reconciliation

- **MODIFY** `openspec/specs/cli-cli-factory/spec.md:74`: replace the non-existent `internal/config.Load()` reference with `config.NewConfigManager().Load()` (the actual function) and add a new scenario documenting the new `Load()` top-level wrapper. Update the `Load()` scenario (lines 51-57) to pin the never-nil `*Config` contract and the typed-error dispatch chain (`*core.FileError` / `*core.ParseError` / `*core.ConfigError`).
- **MODIFY** `openspec/specs/cli-shell-completion/spec.md:305, 312`: change `getCompletionTimeout(nil)` to `getCompletionTimeout(f.Config())` (now matches the implementation).
- **No modification** to `openspec/specs/application-configuration/spec.md:145-156` (the `adrg/xdg` requirement). The source-of-truth spec already mandates `github.com/adrg/xdg`; this change satisfies that requirement by adopting the dependency (see Section F tasks 1.6–1.9). The corresponding MODIFIED-Requirement block in `specs/application-configuration/spec.md` (delta) is deleted — the source-of-truth text stands.

### F. Test updates

- **MODIFY** 6 E2E test files for the `Config.Library` migration, keeping the `GERMINATOR_LIBRARY` env-var test as backwards-compat proof:
  - `test/e2e/library_add_test.go:230-239`
  - `test/e2e/library_discover_test.go:215-233`
  - `test/e2e/library_refresh_test.go:156-169`
  - `test/e2e/library_remove_test.go:217-230`
  - `test/e2e/library_test.go:84` (read-only `library resources`)
  - `test/e2e/init_test.go:26` (helper used by `germinator init` tests)
- **ADD** `TestLoadLibraryForCompletion_HonorsConfigTimeout` (~25 LOC) in `cmd/completions_test.go`: set `Config.Completion.Timeout = "2s"`, assign `f.Config` via direct closure `func() (*config.Config, error) { return cfg, nil }` (not `OnceValuesFunc` — cleaner test seam per `golang-design-patterns` "design for testability"), invoke `loadLibraryForCompletion`, assert the resulting context has a 2-second deadline. Twin test `TestLoadLibraryForCompletion_HonorsCacheTTL` for the `CacheTTL` knob.
- **ADD** env-provider tests in `internal/config/manager_test.go` (~50 LOC): verify that `GERMINATOR_LIBRARY`, `GERMINATOR_DEBUG`, `GERMINATOR_PLATFORM` env vars override config-file values, and that defaults still apply when neither is set. **Truthiness rule** for bool fields: koanf parses `1` / `true` / `t` / `yes` as `true`; other non-empty strings as `false`; unset defaults to the struct default. Documented in `design.md` Risks / Trade-offs.
- **ADD** `adrg/xdg`-backed path tests in `internal/config/manager_test.go` and `internal/library/discovery_test.go` (tagged `//go:build !windows` per user decision — Windows CI deferred).

## Capabilities

### Modified Capabilities

- **`application-configuration`** (delta) — extend the four-tier precedence to actually include tier 3 (env vars); rename `Platform` → `PlatformDefault`, add `Debug` to `Config`, set `Library: ""` in `DefaultConfig()`; document the env-var migration. **Adopt `adrg/xdg`** (Decision 5) to satisfy the source-of-truth mandate at `openspec/specs/application-configuration/spec.md:145-156`; the corresponding MODIFIED-Requirement block for `adrg/xdg` in the delta is dropped, but the delta adds a `Config field set` ADDED-Requirement scenario documenting `Library: ""`. The delta splits into `## ADDED Requirements` (the new `Config field set` requirement) and `## MODIFIED Requirements` (the existing `Configuration precedence` and `Environment variable naming` requirements, whose scenarios gain explicit pointers to the now-implemented env-var tier at the `Manager.Load()` layer).
- **`cli-cli-factory`** (delta) — `Factory.Config` SHALL be assigned in `main.go` via `cmdutil.OnceValuesFunc` wrapping `config.Load()` (a new top-level wrapper around `NewConfigManager().Load()`).
- **`cli-shell-completion`** (delta) — `actionResources` / `actionPresets` / `actionLibraryRefs` / `loadLibraryForCompletion` SHALL call `f.Config()` and forward the `*config.Config` to `resolveLibraryPath`, `getCompletionTimeout`, and `getCacheTTL`. The two existing "configurable" scenarios (lines 168-170, 178-182) now have passing tests AND passing implementations. The `actionPlatforms` shape is carved out as `func(*cmdutil.Factory) carapace.Action` to document a pre-existing 1-arg signature (no Cobra command needed for static platform values).

## Impact

### Affected code

- **Modified (1 file):** `main.go` (+~10 LOC for fail-fast `Load()` + `f.Config` assignment + `io.SetDebug`)
- **Modified (4 files):** `internal/config/config.go` (rename + add fields + Validate + `Library: ""` default), `internal/config/manager.go` (env provider + `adrg/xdg` migration in `resolveConfigPath`), `internal/config/load.go` (new file, +10 LOC for the `Load()` wrapper), `internal/library/discovery.go` (`FindLibrary` extended to 3 args + `DefaultLibraryPath` migrates to `adrg/xdg.DataFile(...)`)
- **Modified (1 file):** `internal/iostreams/iostreams.go` (~10 LOC net: remove env-read branch, add `SetDebug(bool)`)
- **Modified (10 files):** 9 cmd files (`show`, `resources`, `presets`, `init`, `library_add`, `library_create`, `library_refresh`, `library_remove`, `library_validate`) for `Config.Library` integration via the 3-arg `FindLibrary(flag, env, cfg)`; plus `cmd/completions.go` for `Config`-threaded completion actions
- **Modified (6 files):** 6 E2E test files for backwards-compat assertions
- **Added:** env-provider tests, `adrg/xdg` path tests, `SetDebug` test (`TestIOStreams_SetDebug`), env-read-removal test (`TestSystem_NoLongerReadsEnvDebug`), `loadLibraryForCompletion` tests for `Timeout` and `CacheTTL`, `FindLibrary` 3-arg precedence tests (`TestResolveLibrary_FlagOverEnvOverCfgOverDefault`, `TestResolveLibrary_AllEmpty_ReturnsXDGDefault`), closure migration tests (`TestRefreshLibrary_HonorsConfigLibrary`, `TestRemoveLibrary_HonorsConfigLibrary`), env-key-mapping test (`TestLoad_EnvKeyMapping_PlatformDefault`), bool-truthiness-rule test (`TestConfig_EnvVarBoolTruthinessRule`), default-emptiness test (`TestDefaultConfig_LibraryIsEmpty`)
- **Modified (3 files):** 3 spec files (`cli-cli-factory`, `cli-shell-completion`, `application-configuration`) — `application-configuration` documents `Library: ""` in `Config field set` and `cli-shell-completion` carves out `actionPlatforms` as one-arg
- **Modified:** `go.mod`, `go.sum` (add `github.com/adrg/xdg`)
- **Modified (verification):** after migration, run `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)" cmd/ internal/ --type go`; the only surviving match is `cmd/completions.go:54` (the priority-chain helper inside `resolveLibraryPath`, unchanged by this change). All 11 production sites — including the two env-once closures — are migrated to the 3-arg `FindLibrary` form; no env-read remains outside the helper.

### Affected systems

- **Config file:** `~/.config/germinator/config.toml` becomes a real runtime config source. New fields: `library = "..."`, `platform_default = "..."`, `debug = true`.
- **Env vars:** precedence order is now codified: **flag > env > config file > default** (encoded directly in the new 3-arg `library.FindLibrary`). `GERMINATOR_LIBRARY` still wins over `Config.Library` (per `application-configuration/spec.md:122`); `GERMINATOR_DEBUG` now flows through `Config.Debug` instead of being read directly.
- **`Config.Library` semantics change**: `DefaultConfig()` now seeds `Library: ""` (was the tilde-prefix `~/.config/germinator/library`). Falling through to `DefaultLibraryPath()` yields the XDG path via `adrg/xdg.DataFile("germinator/library")`, which is the same default users would have observed via the prior hardcoded value under typical setups. Users who relied on the literal tilde-prefix string will instead observe the resolved XDG path; documented behavior change.
- **Completion cache:** the configurable TTL and timeout finally work — `cmd/completions.go` now passes a real `*config.Config` to the helpers.
- **CLI behavior:** end-user behavior is unchanged for users who don't author a config file (all defaults remain identical). Users who author a config file gain the ability to tune completion behavior and (optionally) default the platform.

### Backward compatibility

- All existing flags continue to work (flag > env > config > default).
- All existing env vars continue to work (env > config > default).
- Missing config file falls back to defaults (no change).
- All 6 E2E test files keep their `GERMINATOR_LIBRARY` env-var test cases (they become backwards-compat proofs).

## Risks

- **Completion timeouts may shift for users with stale caches**: if a user authors `~/.config/germinator/config.toml` with `completion.timeout = "1s"` and a stale completion cache from before the migration, the next completion may wait 1s instead of 500ms. **Mitigation**: `Config.DefaultConfig()` keeps `Timeout: "500ms"` (no behavioral change unless user opts in); the migration ships with zero-config users unaffected.
- **`Config.Debug` is an additive public API change**: `Config` is a public struct exported from `internal/`. Adding a field does not break existing consumers but is visible to anyone with field-name struct literals. **Mitigation**: `grep` confirms zero struct literals with field names in `cmd/`; the additive change is safe.
- **`iostreams.System()` debug channel refactor**: removing the `os.LookupEnv("GERMINATOR_DEBUG")` branch is a behavior-preserving change for production (env → `cfg.Debug` → `SetDebug` delivers the same handler) but is observable to external code that imports `iostreams.System()` and assumed env-driven debug activation. **Mitigation**: there is no such external consumer in this repo (verified via `rg "iostreams\.System" --files-without-match '_test\.go'`); `iostreams.Test()` is unaffected (it already returns the discard logger); the new `SetDebug(bool)` method is the canonical activation point and is exercised by `TestSystem_NoLongerReadsEnvDebug` (task 7.4).
- **`Config.Platform` rename to `Config.PlatformDefault`**: a breaking change for any code (in this repo or external) that initializes `Config{}` with the field name `Platform`. **Mitigation**: `rg "Platform:\s+" cmd/ internal/` returns zero hits — all `Config` initializations go through `DefaultConfig()`; the rename is safe in-repo; external consumers should rebuild (acceptable for a `internal/` package's additive evolution).
- **`GERMINATOR_LIBRARY` migration**: 11 call sites change in production code; one missed site silently reverts to env-only behavior. **Mitigation**: task 4.3 runs `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)" cmd/ internal/ --type go` after the migration; the only expected surviving match is `cmd/completions.go:54` (priority-chain helper — intentional per design Decision 2 and unchanged). All 11 production sites — including the two env-once closures in `cmd/library_refresh.go:151` and `cmd/library_remove.go:231` (now migrated to call `f.Config()` and pass `cfg.Library` to the 3-arg `FindLibrary`) — must show zero remaining direct `os.Getenv` reads at the resolution site.
- **`adrg/xdg` dependency adoption** — three sub-risks:
   1. **Transitive-dep surprise**: `adrg/xdg` brings zero upstream dependencies but does `go.sum` work; pin a specific semver. **Mitigation**: `go mod verify` and review `go.sum` after `go mod tidy`.
   2. **Platform-specific behavior under non-standard `XDG_*` paths**: `adrg/xdg`'s resolver may behave differently from the bespoke ladder under edge-case env var settings (e.g., `XDG_DATA_HOME=""`). **Mitigation**: integration tests with `XDG_DATA_HOME` and `XDG_CONFIG_HOME` set to custom values and unset (tagged `//go:build !windows`).
   3. **Cross-package change**: `DefaultLibraryPath` is in `internal/library/`, not `internal/config/`; the migration touches a different package than the rest of this change. The new 3-arg `FindLibrary(flag, env, cfg)` signature is breaking for `internal/library/` callers (the project is the only in-repo consumer). **Mitigation**: all 11 call sites are pinned in this proposal (Section C) and updated atomically with the signature change; the `library.FindLibrary` precedence is verified by new `TestResolveLibrary_FlagOverEnvOverCfgOverDefault` and `TestResolveLibrary_AllEmpty_ReturnsXDGDefault` tests. Five known callers of `DefaultLibraryPath` (`cmd/init.go:139`, `cmd/completions.go:64`, plus internal tests) are unaffected by signature changes; behavior preservation is verified by the existing `internal/library/discovery_test.go` suite plus the new `TestDefaultLibraryPath_AdoptsXDG`.
- **Main.go config-load fail-fast**: surfacing `Load()` errors as exit-1 means a malformed `config.toml` (previously silently swallowed) now blocks CLI invocation. **Mitigation**: documented as a deliberate behavior change (`cli-exit-codes/spec.md` already maps `*core.*Error` to exit 1 via `cmdutil.ExitCodeFor`); users can `unset GERMINATOR_*` and remove `config.toml` to recover.
- **Env provider introduces key-mapping complexity**: koanf's env provider uses `_` as the default delimiter and lowercases keys after stripping the configured prefix. The `Config.PlatformDefault` field (koanf key `platform`) therefore maps to env `GERMINATOR_PLATFORM`, **not** `GERMINATOR_PLATFORM_DEFAULT`. **Truthiness rule for bool fields**: koanf parses bool values per `strconv.ParseBool` semantics — `1` / `t` / `T` / `true` / `TRUE` / `True` resolve to `true`; all other non-empty strings resolve to `false`; unset defaults to the struct default. **Mitigation**: document both the key mapping and the bool truthiness rule in `internal/config/manager.go` and the `Config` field godoc; the env-provider layer's key mapping is tested by `TestLoad_EnvKeyMapping_PlatformDefault` and the bool truthiness is tested by `TestConfig_EnvVarBoolTruthinessRule` (task 2.1).
