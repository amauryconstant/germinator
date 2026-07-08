# Wire `Factory.Config` and complete the koanf config pipeline

## Why

`Factory.Config func() (*config.Config, error)` is declared at `internal/cmdutil/factory.go:31` and required by `openspec/specs/cli-cli-factory/spec.md:37`, but is **never assigned in `main.go`** (`main.go:24-38`). The koanf-based loader at `internal/config/manager.go` implements only the first two tiers (defaults → file) of the documented four-tier merge (defaults → file → env → flags); env vars are read directly via `os.Getenv` at 13+ call sites, and the third tier (env) is absent at the Manager layer.

Three downstream consequences:

1. **`cmd/completions.go:103, 111` pass `nil`** to `getCompletionTimeout` and `getCacheTTL`, breaking `cli-shell-completion/spec.md:168-170` and `cli-shell-completion/spec.md:178-182` which promise that `completion.timeout` and `completion.cache_ttl` in `~/.config/germinator/config.toml` will take effect.
2. **`cmd/completions.go:120, 132, 144` pass `nil`** to `resolveLibraryPath`, breaking the "Resolve library from config" scenario at `cli-shell-completion/spec.md:203-210`.
3. **`cli-cli-factory/spec.md:74`** references a non-existent `internal/config.Load()` function.

The `application-configuration/spec.md:113-122` four-tier precedence is also unmet: the loader stops at tier 2 and per-call-site `os.Getenv` reads handle the env tier ad-hoc without going through the documented pipeline.

Closing this gap requires:
- Extending the koanf `Manager.Load()` to install an env provider (tier 3).
- Adding `Config.Debug`, `Config.Library`, and `Config.PlatformDefault` fields.
- Wiring `Factory.Config` in `main.go` via `cmdutil.OnceValuesFunc`.
- Threading the loaded `Config` into the runtime call sites that currently pass `nil`.
- Migrating the 5 runtime env-var reads (`GERMINATOR_LIBRARY`, `GERMINATOR_DEBUG`, `GERMINATOR_PLATFORM`, `NO_COLOR`, `EXIT_CODE_LEGACY`) to the documented four-tier pipeline. `NO_COLOR` and `EXIT_CODE_LEGACY` remain direct (per their respective specs); the other three migrate.
- Adding the top-level `internal/config.Load()` wrapper referenced (but missing) at `cli-cli-factory/spec.md:74`.

## What Changes

### A. Extend the koanf loader

- **MODIFY** `internal/config/manager.go`: add `koanf/providers/env` with `GERMINATOR_` prefix to the `Load()` merge order. Merge order becomes: defaults → file → env → (flags handled per-command by Cobra, which already overrides per spec).
- **MODIFY** `internal/config/config.go`:
  - Add fields: `Library string` (path), `PlatformDefault string` (default platform), `Debug bool` (debug logging).
  - Update `DefaultConfig()` and `Validate()` to cover the new fields.
- **ADD** `internal/config/load.go` (or extend `manager.go`): export `func Load() (*Config, error)` as a top-level wrapper around `NewConfigManager().Load()`. Fixes the broken reference at `cli-cli-factory/spec.md:74`.

### B. Wire `Factory.Config`

- **MODIFY** `main.go`: after `f.Library = ...` (around line 38), add `f.Config = cmdutil.OnceValuesFunc(config.Load)`.
- **MODIFY** `internal/iostreams/iostreams.go:152`: read `GERMINATOR_DEBUG` from the loaded `Config.Debug` instead of `os.LookupEnv`. Two options evaluated in design.md Decision 3; **chosen**: keep `iostreams.System()` constructor unchanged (it still reads env at construction for backwards-compat with the 50+ test files that use it), and add a `SetDebug(bool)` method on `IOStreams` that `main.go` calls after loading the config.

### C. Migrate runtime env-var reads to `Config`

| Env var | Current sites | Migration |
|---|---|---|
| `GERMINATOR_LIBRARY` | 13+ `os.Getenv` calls in `main.go:36`, `cmd/show.go:87`, `cmd/resources.go:78`, `cmd/presets.go:68`, `cmd/init.go:139`, `cmd/library_add.go:251`, `cmd/library_create.go:155`, `cmd/library_refresh.go:151`, `cmd/library_remove.go:232`, `cmd/library_validate.go:156`, `cmd/completions.go:54` | Migrate to `Config.Library`. Env var remains as the precedence-winner per `application-configuration/spec.md:122` (parallel three-tier chain within the config layer). The call pattern becomes `library.FindLibrary(opts.Library, os.Getenv("GERMINATOR_LIBRARY"))` — env still wins, but the config-tier default is now reachable. |
| `GERMINATOR_DEBUG` | `internal/iostreams/iostreams.go:152` | Migrate to `Config.Debug`. Backwards-compat: `iostreams.System()` reads env at construction so the 50+ test files that use it directly keep working without changes. Production main calls `io.SetDebug(cfg.Debug)` after loading the config. |
| `GERMINATOR_PLATFORM` | Never read | Add `Config.PlatformDefault` field; wire as a default for commands that take a `--platform` flag. Today every command requires `--platform` explicitly (`cmd/AGENTS.md:46-51`); this change adds the option to default it from the config (out-of-scope commands may opt in via a follow-up change). |
| `NO_COLOR` | `internal/iostreams/styles.go:10` (constant) | **Stay direct**. The `cli-color-policy/spec.md` explicitly mandates direct env read; out of scope. |
| `EXIT_CODE_LEGACY` | `internal/warning/canary.go:40` | **Stay direct**. Boot-time feature flag, not user config; out of scope per `cli-exit-codes/spec.md`. |

### D. Thread `Config` through completion actions

- **MODIFY** `cmd/completions.go:103, 111`: replace `getCompletionTimeout(nil)` and `getCacheTTL(nil)` with `getCompletionTimeout(cfg)` and `getCacheTTL(cfg)` where `cfg` is loaded via `f.Config()`.
- **MODIFY** `cmd/completions.go:120, 132, 144`: replace `resolveLibraryPath(cmd, nil)` with `resolveLibraryPath(cmd, cfg)` for `actionResources`, `actionPresets`, `actionLibraryRefs`. The `resolveLibraryPath` helper at `cmd/completions.go:47-65` already honors `cfg.Library` at line 59 — the production callers just need to pass a non-nil `cfg`.

### E. Spec/code reconciliation

- **MODIFY** `openspec/specs/cli-cli-factory/spec.md:74`: replace the non-existent `internal/config.Load()` reference with `config.NewConfigManager().Load()` (the actual function) and add a new scenario documenting the new `Load()` top-level wrapper.
- **MODIFY** `openspec/specs/cli-shell-completion/spec.md:305, 312`: change `getCompletionTimeout(nil)` to `getCompletionTimeout(f.Config())` (now matches the implementation).
- **MODIFY** `openspec/specs/application-configuration/spec.md:147-156`: add a decision on whether to add the `adrg/xdg` dependency. **Recommended decision**: relax the spec to remove the `adrg/xdg` mandate (the current `resolveConfigPath` works correctly on Unix + macOS; Windows support is a known limitation already documented in the codebase). If Windows support becomes a priority, a follow-up change adds `adrg/xdg`.

### F. Test updates

- **MODIFY** 4 E2E test files for the `Config.Library` migration, keeping the `GERMINATOR_LIBRARY` env-var test as backwards-compat proof:
  - `test/e2e/library_add_test.go:230-239`
  - `test/e2e/library_discover_test.go:215-233`
  - `test/e2e/library_refresh_test.go:156-169`
  - `test/e2e/library_remove_test.go:217-230`
- **ADD** `TestLoadLibraryForCompletion_HonorsConfigTimeout` (~25 LOC) in `cmd/completions_test.go`: set `Config.Completion.Timeout = "2s"`, invoke `loadLibraryForCompletion`, assert the resulting context has a 2-second deadline.
- **ADD** env-provider tests in `internal/config/manager_test.go` (~50 LOC): verify that `GERMINATOR_LIBRARY`, `GERMINATOR_DEBUG`, `GERMINATOR_PLATFORM` env vars override config-file values, and that defaults still apply when neither is set.

## Capabilities

### Modified Capabilities

- **`application-configuration`** (delta) — extend the four-tier precedence to actually include tier 3 (env vars); add `Library`, `PlatformDefault`, `Debug` fields to `Config`; document the env-var migration. Relax the `adrg/xdg` mandate (Decision 5).
- **`cli-cli-factory`** (delta) — `Factory.Config` SHALL be assigned in `main.go` via `cmdutil.OnceValuesFunc` wrapping `config.Load()` (a new top-level wrapper around `NewConfigManager().Load()`).
- **`cli-shell-completion`** (delta) — `actionResources` / `actionPresets` / `actionLibraryRefs` / `loadLibraryForCompletion` SHALL call `f.Config()` and forward the `*config.Config` to `resolveLibraryPath`, `getCompletionTimeout`, and `getCacheTTL`. The two existing "configurable" scenarios (lines 168-170, 178-182) now have passing tests AND passing implementations.

## Impact

### Affected code

- **Modified (1 file):** `main.go` (+5 LOC for `f.Config` assignment)
- **Modified (3 files):** `internal/config/config.go`, `internal/config/manager.go`, `internal/config/load.go` (new file, +10 LOC for the wrapper)
- **Modified (1 file):** `internal/iostreams/iostreams.go` (+15 LOC for `SetDebug` method)
- **Modified (11 files):** 10 cmd files (`show`, `resources`, `presets`, `init`, `library_add`, `library_create`, `library_refresh`, `library_remove`, `library_validate`, `completions`) for `Config.Library` and `Config.Debug` integration
- **Modified (4 files):** 4 E2E test files for backwards-compat assertions
- **Added (2 files):** `TestLoadLibraryForCompletion_HonorsConfigTimeout` test, env-provider tests
- **Modified (3 files):** 3 spec files (`cli-cli-factory`, `cli-shell-completion`, `application-configuration`)
- **Modified (N files):** all `cmd/*.go` files that import `os` for `GERMINATOR_LIBRARY` (verify no remaining direct reads after migration)

### Affected systems

- **Config file:** `~/.config/germinator/config.toml` becomes a real runtime config source. New fields: `library = "..."`, `platform_default = "..."`, `debug = true`.
- **Env vars:** precedence order is now codified: flag > env > config file > default. `GERMINATOR_LIBRARY` still wins over `Config.Library` (per `application-configuration/spec.md:122`); `GERMINATOR_DEBUG` now flows through `Config.Debug` instead of being read directly.
- **Completion cache:** the configurable TTL and timeout finally work — `cmd/completions.go` now passes a real `*config.Config` to the helpers.
- **CLI behavior:** end-user behavior is unchanged for users who don't author a config file (all defaults remain identical). Users who author a config file gain the ability to tune completion behavior and (optionally) default the platform.

### Backward compatibility

- All existing flags continue to work (flag > env > config > default).
- All existing env vars continue to work (env > config > default).
- Missing config file falls back to defaults (no change).
- All 4 E2E test files keep their `GERMINATOR_LIBRARY` env-var test cases (they become backwards-compat proofs).

## Risks

- **Completion timeouts may shift for users with stale caches**: if a user authors `~/.config/germinator/config.toml` with `completion.timeout = "1s"` and a stale completion cache from before the migration, the next completion may wait 1s instead of 500ms. **Mitigation**: `Config.DefaultConfig()` keeps `Timeout: "500ms"` (no behavioral change unless user opts in); the migration ships with zero-config users unaffected.
- **`Config.Debug` is an additive public API change**: `Config` is a public struct exported from `internal/`. Adding a field does not break existing consumers but is visible to anyone with field-name struct literals. **Mitigation**: `grep` confirms zero struct literals with field names in `cmd/`; the additive change is safe.
- **`iostreams.System()` debug migration**: must preserve backwards compatibility with 50+ test files that construct `iostreams.Test()` directly. **Mitigation**: `iostreams.System()` retains its env-read at construction; the new `SetDebug(bool)` method is only called from `main.go` after config load.
- **`GERMINATOR_LIBRARY` migration**: 13+ call sites change; one missed site silently reverts to env-only behavior. **Mitigation**: task `2.6` runs `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)"` after the migration; zero matches expected in production code (tests may keep it as backwards-compat proof).
- **`adrg/xdg` spec relaxation**: Windows users may experience less-tested XDG path resolution. **Mitigation**: documented limitation; Windows support is a known gap already acknowledged in the codebase; a follow-up change adds `adrg/xdg` if/when Windows support becomes a priority.
