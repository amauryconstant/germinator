# Tasks — Wire `Factory.Config` and complete the koanf config pipeline

Each task ends with `mise run check` passing.

## 1. Extend the koanf loader

- [x] 1.1 In `internal/config/manager.go`, add `koanf/providers/env` to `Manager.Load()` with the `GERMINATOR_` prefix and `_` delimiter. Merge order: defaults → file → env. Existing tier 1 (defaults) and tier 2 (file) remain unchanged. Also add a compile-time interface check `var _ Manager = (*koanfConfigManager)(nil)` next to the `koanfConfigManager` struct definition (per golang-structs-interfaces "Compile-Time Interface Check" — catches interface drift at build time, costs nothing at runtime).
- [x] 1.2 In `internal/config/config.go`, **rename** the existing `Platform string` field (with `koanf:"platform"` tag, currently at `config.go:18`) to `PlatformDefault string` (same `koanf:"platform"` tag). Add a new `Debug bool` field (with `koanf:"debug"` tag). `Library string` already exists — do not add a duplicate. **Mandatory pre-read**: `internal/AGENTS.md` and `internal/config/AGENTS.md` for the struct-tag conventions. Note: `internal/AGENTS.md` currently mentions `AppConfig` (a pre-existing doc drift); the actual type is `Config`. The struct-tag conventions and field descriptions in the doc are accurate; only the type name is wrong. Out of scope for this change.
- [x] 1.3 In `internal/config/config.go`, update `DefaultConfig()` to seed the new fields (`Library: ""`, `PlatformDefault: ""`, `Debug: false`). Setting `Library: ""` makes `cfg.Library == ""` the canonical "no config-file override" signal — falling through to `DefaultLibraryPath()` (XDG via `adrg/xdg`, see task 1.8) yields the default path. The prior tilde-prefix hardcode (`"~/.config/germinator/library"`) is removed.
- [x] 1.4 In `internal/config/config.go`, extend `Validate()` to cover the renamed/new fields: `PlatformDefault` must be empty, `claude-code`, or `opencode` (returns `*core.ConfigError` otherwise); `Debug` is always valid (bool). `Library` is always valid (empty falls through to XDG default at resolution time).
- [x] 1.5 Add `internal/config/load.go` with `func Load() (*Config, error)` — a top-level wrapper using a package-level function variable for test injection (per golang-design-patterns "design for testability"):
  ```go
  // loadFn is a package-level seam for testing. Tests override it to inject a
  // stub Manager without re-running NewConfigManager(). Production code MUST
  // NOT modify it. The variable is one mutable package-level binding, which
  // is the documented cost of the test-injection seam (see design.md
  // Decision 4 alternatives).
  var loadFn = NewConfigManager

  func Load() (*Config, error) {
      mgr := loadFn()
      if err := mgr.Load(); err != nil {
          return mgr.GetConfig(), err
      }
      return mgr.GetConfig(), nil
  }
  ```
  Tests can swap `loadFn` to inject a stub Manager. Fixes the broken reference at `cli-cli-factory/spec.md:74`. Per the contract pinned in `cli-cli-factory/spec.md` (lines 51-57), `*Config` is always non-nil because `mgr.GetConfig()` returns the `DefaultConfig()`-seeded struct initialized in `NewConfigManager()`. On error it holds the same `DefaultConfig()` values (the koanf unmarshal happens after defaults seeding), and the error chain is the authoritative signal.
- [x] 1.6 Add `github.com/adrg/xdg` as a dependency: `go get github.com/adrg/xdg`, then `go mod tidy`. Verify `go.sum` integrity with `go mod verify`. The dependency is a direct dependency (no `// indirect` comment in `go.mod`) since it is imported directly by `internal/config/manager.go` and `internal/library/discovery.go`.
- [x] 1.7 In `internal/config/manager.go`, replace `resolveConfigPath` (lines 93-121) with a thin wrapper: `path, err := xdg.ConfigFile("germinator/config.toml")`; on `err == nil`, return `path`. If `adrg/xdg` returns an empty path or an error, fall back to the current `./config.toml` working-directory check (preserve current CWD behavior for projects that ship their own config). Import as `import "github.com/adrg/xdg"` (no alias — the default package name `xdg` does not collide with `internal/config` package locals, and golang-naming prefers no alias unless collision requires one).
- [x] 1.8 In `internal/library/discovery.go`, replace the body of `DefaultLibraryPath` (lines 31-51) with `xdg.DataFile("germinator/library")`. Preserve the existing working-directory `./germinator/library/` last-resort fallback for project-local libraries (`if !pathExists(path) { return "./germinator/library" }` or similar). The five known callers (`cmd/init.go:139`, `cmd/completions.go:64`, plus internal tests) are unaffected by signature change.
- [x] 1.9 In `internal/library/discovery.go`, extend `FindLibrary` from 2 args to 3 args to encode the spec-mandated 4-tier precedence directly. Change signature: `func FindLibrary(flagPath, envPath, cfgPath string) string` and implement:
  ```go
  if flagPath != "" { return flagPath }   // 1. --library flag (highest)
  if envPath  != "" { return envPath  }   // 2. GERMINATOR_LIBRARY env
  if cfgPath  != "" { return cfgPath  }   // 3. Config.Library
  return DefaultLibraryPath()             // 4. XDG via adrg/xdg
  ```
  Update the godoc to reflect the new precedence. **All 11 production call sites are migrated atomically with the signature change** (see task 4.2 and 4.3). Update unit tests in `internal/library/discovery_test.go` for the 3-arg signature.

## 2. Update internal/config tests

- [x] 2.1 In `internal/config/manager_test.go`, add env-provider tests (~50 LOC). **Test discipline**: Tests in this section that mutate process env (via `t.Setenv(name, value)` — Go 1.17+, natively supported by the Go 1.25.5 module) are sequential by default per golang-testing Rule 4 + golang-safety Rule 4 (process-env mutation is not parallel-safe). **`TestDefaultConfig_LibraryIsEmpty` is an exception**: it touches no env state and SHOULD use `t.Parallel()` for speed. All table-driven tests in this section MUST use named subtests via `t.Run(tt.name, ...)` per golang-testing Rule 1.
  - `TestLoad_EnvOverridesFile`: with `GERMINATOR_LIBRARY=/env/lib` set and `library = "/file/lib"` in config file, `Load()` returns `Config.Library == "/env/lib"`
  - `TestLoad_EnvOverridesDefault`: with `GERMINATOR_DEBUG=1` set, `Load()` returns `Config.Debug == true`. **Truthiness rule documented in `design.md` Risk section**: koanf env provider parses boolean fields via `strconv.ParseBool` semantics — values `1` / `t` / `T` / `true` / `TRUE` / `True` resolve to `true`; all other non-empty strings resolve to `false`; unset defaults to the struct default. Test that `GERMINATOR_DEBUG=0` (non-empty but `false`-parsing) yields `Config.Debug == false`, pinning the rule.
  - `TestLoad_NoEnvNoFile`: with neither env nor file, `Load()` returns `DefaultConfig()` values (including `Library: ""`)
  - `TestLoad_MissingFile`: with file missing but env set, `Load()` returns env-derived values (no error)
  - `TestLoad_EnvKeyMapping_PlatformDefault`: with `GERMINATOR_PLATFORM=opencode` set (NOT `GERMINATOR_PLATFORM_DEFAULT`), `Load()` returns `Config.PlatformDefault == "opencode"`. Pins the lowercase-after-prefix-stripping key mapping documented in `design.md` Risk section.
  - `TestConfig_EnvVarBoolTruthinessRule` (~30 LOC): table-driven test of `1`/`t`/`T`/`true`/`TRUE`/`True` → `true`; `0`/`f`/`F`/`false`/`""` (unset) → `false`; unparseable values like `no`/`garbage` → `*core.ParseError` (koanf calls `strconv.ParseBool`, which errors on unknown strings; the bool field is left at its struct default but the error chain is the authoritative signal). Documents the koanf bool parsing rule at test level so future koanf upgrades can detect regressions.
  - `TestDefaultConfig_LibraryIsEmpty`: `config.DefaultConfig().Library == ""`. Pins the `Library: ""` shape change from task 1.3.
- [x] 2.2 In `internal/config/manager_test.go`, add `adrg/xdg`-backed path-resolution tests:
  - `TestResolveConfigPath_HonorsXDGConfigHome`: with `XDG_CONFIG_HOME=/xdg/cfg` set and no `HOME`, `resolveConfigPath` returns `/xdg/cfg/germinator/config.toml`. Tag `//go:build !windows` per user decision (Windows CI skipped for this change). **Build tag clarification**: `//go:build !windows` is a cross-platform skip per user decision (Windows CI deferred). This is NOT a `//go:build integration` tag — these remain unit tests, just excluded from Windows CI.
  - `TestResolveConfigPath_FallsBackToCWD`: with neither env set, returns `./config.toml` when it exists in the CWD.
- [x] 2.3 In `internal/library/discovery_test.go`, add `TestDefaultLibraryPath_AdoptsXDG`: with `XDG_DATA_HOME=/xdg/lib` set, `DefaultLibraryPath()` returns `/xdg/lib/germinator/library`. Tag `//go:build !windows` (cross-platform skip; not an integration tag — see clarification in task 2.2).
- [x] 2.4 In `internal/library/discovery_test.go`, update `FindLibrary` unit tests for the 3-arg signature (per task 1.9):
  - `TestResolveLibrary_FlagOverEnvOverCfgOverDefault`: table-driven, 5 rows with `name` fields using lowercase descriptive phrases per `golang-naming` subtest convention (e.g., `"flag wins over env"`, `"env wins over cfg"`, `"cfg wins over default"`, `"all empty returns xdg default"`).
  - `TestResolveLibrary_AllEmpty_ReturnsXDGDefault`: `FindLibrary("", "", "")` returns the result of `DefaultLibraryPath()` (mocked).
  - Update existing `TestFindLibrary_*` tests if any reference the 2-arg signature.

## 3. Wire `Factory.Config`

- [x] 3.1 In `main.go`, single-call pattern: `f.Config = cmdutil.OnceValuesFunc(config.Load); cfg, err := f.Config()`; on error → `output.FormatError(f.IOStreams, err); os.Exit(int(cmdutil.ExitCodeFor(err)))`. Then call `io.SetDebug(cfg.Debug)` and assign the lazy library closure. The single `f.Config()` invocation populates the cache — subsequent calls from completion actions return the cached `*Config` (per golang-cli-architecture "Cached Lazy Initialization"). Verify `internal/config` is imported.
- [x] 3.2 In `internal/iostreams/iostreams.go`, **remove** the `os.LookupEnv("GERMINATOR_DEBUG")` branch from `newDebugLogger` (lines 151-156) so it always returns a discard-debug handler. Add `func (s *IOStreams) SetDebug(enabled bool)` method that swaps the `Logger` field to a debug-level `slog.NewTextHandler(s.ErrOut, &slog.HandlerOptions{Level: slog.LevelDebug})` when `enabled == true`, or to a discard handler when `false`. Update the `System()` docstring (lines 34-37) to reflect that debug activation is no longer env-driven.
- [x] 3.3 (verification) `mise run build && mise run test`. Per `golang-error-handling` rule 1 (returned errors MUST be checked) and `golang-cli-architecture` "main.go is the only composition root". The single `f.Config()` call populates the cache and feeds `SetDebug`; subsequent `f.Config()` calls from completion actions return the cached `*Config`.

## 4. Migrate `GERMINATOR_LIBRARY` reads to `Config.Library`

- [x] 4.1 In `main.go:36`, update `library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"))` to `library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"), cfg.Library)` (per task 1.9's 3-arg signature — flag empty, env wins, config is fallback tier).
- [x] 4.2 Update all 11 production call sites to the 3-arg form `library.FindLibrary(flagValue, os.Getenv("GERMINATOR_LIBRARY"), opts.ConfigLibraryPath)`:
  - `cmd/show.go:87`, `cmd/resources.go:78`, `cmd/presets.go:68`, `cmd/init.go:139`, `cmd/library_add.go:251`, `cmd/library_create.go:155`, `cmd/library_refresh.go:146-155` (the `refreshLibrary` closure body — see task 4.4 for the safe nil-check pattern), `cmd/library_remove.go:226-235` (the `removeLibrary` closure body — see task 4.4 for the safe nil-check pattern), `cmd/library_validate.go:156` — pass `opts.ConfigLibraryPath` (a **new** `string` field on each command's options struct populated from `cfg.Library` in `RunE`). **Naming rationale**: the existing `opts.Library func() (*library.Library, error)` lazy closure is a different concern; using `ConfigLibraryPath` avoids shadowing the closure name. For the 9 non-closure sites, the `cfg.Library` value is sourced via `opts.ConfigLibraryPath` which is populated in `RunE` from `cfg.Library` after `f.Config()` is called once and checked for errors. For the two closure bodies (`refreshLibrary`, `removeLibrary`), the `cfg.Library` value is sourced inside the closure using the explicit nil-safe pattern from task 4.4 (NOT `cfg, _ := f.Config()` — see task 4.4 for rationale). Per the shaped plan, both closures are migrated (not kept as design exceptions). **Implementation note**: the actual implementation uses the closure-with-`f.Config()` pattern from task 4.4 uniformly across all 6 helpers (`initLibrary`, `addLibrary`, `createPresetLibrary`, `refreshLibrary`, `removeLibrary`, `validateLibrary`) — the helper signatures stay 2-arg (`f, explicitPath`), which preserves test call sites and avoids the `ConfigLibraryPath` field churn on the helper-shaped options structs. Only the 3 direct call sites (`show`, `resources`, `presets`) get a new `ConfigLibraryPath` field on the options struct.
- [x] 4.3 Verify zero remaining production-code `os.Getenv("GERMINATOR_LIBRARY")` calls outside `cmd/completions.go:54` via `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)" cmd/ internal/ --type go`. The only expected surviving match is `cmd/completions.go:54` (the priority-chain helper inside `resolveLibraryPath`, unchanged by this change). Test files may keep the env-var read as backwards-compat proof.
- [x] 4.4 Update `cmd/library_refresh.go`, `cmd/library_remove.go`, `cmd/library_add.go`, `cmd/library_create.go`, `cmd/library_validate.go`, `cmd/init.go` helper closures to use the explicit nil-safe pattern (per golang-safety "design for zero values"). The pattern guards **both** `f.Config == nil` (function field unset) and the typed-nil `cfg` case:
  ```go
  var cfgPath string
  if f.Config != nil {
      if cfg, cfgErr := f.Config(); cfgErr == nil && cfg != nil {
          cfgPath = cfg.Library
      }
  }
  resolved := library.FindLibrary(explicitPath, envPath, cfgPath)
  ```
  Production paths always wire `f.Config` (main.go:3.1); the nil-checks are defense-in-depth for test paths (e.g., `cmd/init_test.go:585 TestInitLibrary_HonorsExplicitPath` constructs a Factory via `cmdutil.NewFactory(...)` without setting `f.Config` — without the outer `f.Config != nil` guard, the test panics with a nil-function-value dereference). New tests in tasks 7.6 + 7.7 cover the wired-`f.Config` happy path; test 7.8 covers the `f.Config == nil` graceful fallback. `

## 5. Thread `Config` through completion actions

- [x] 5.1 In `cmd/completions.go:103`, replace `getCompletionTimeout(nil)` with `getCompletionTimeout(cfg)` where `cfg` is loaded via the explicit nil-safe pattern from task 4.4 at the top of `loadLibraryForCompletion`:
  ```go
  var cfg *config.Config
  if c, cfgErr := f.Config(); cfgErr == nil && c != nil {
      cfg = c
  } else if cfgErr != nil {
      // Graceful fallback to defaults when config load fails. Per
      // golang-error-handling Rule 7 (errors MUST be either logged OR returned),
      // the error is logged at debug level (not silently dropped) so failure is
      // observable in `--verbose` runs. Production main.go fail-fast (task 3.1)
      // catches load errors before this path is reached in practice.
      f.IOStreams.Logger.Debug("config load failed; using defaults", "error", cfgErr)
  }
  loadCtx, cancel := context.WithTimeout(f.RootContext, getCompletionTimeout(cfg))
  ```
  The nil-check is the same shape as task 4.4 — completion actions must follow the same defensive pattern (per golang-error-handling Rule 1: "Returned errors MUST always be checked"). If `f.Config()` fails or returns a typed-nil, `cfg` is `nil` and the helpers fall through to their default timeouts/TTLs (`500ms` / `5s`).
- [x] 5.2 In `cmd/completions.go:111`, replace `getCacheTTL(nil)` with `getCacheTTL(cfg)` using the same nil-safe pattern (and debug-log fallback) as task 5.1.
- [x] 5.3 In `cmd/completions.go:120, 132, 144`, replace `resolveLibraryPath(cmd, nil)` with `resolveLibraryPath(cmd, cfg)` for `actionResources`, `actionPresets`, `actionLibraryRefs`. Load `cfg` once per action invocation using the same nil-safe pattern (and debug-log fallback) from task 5.1. `resolveLibraryPath` already handles a nil `cfg` by falling through to `DefaultLibraryPath()` (XDG via `adrg/xdg`), so the behavior is identical when config is unavailable.

## 6. Spec reconciliation

- [ ] 6.1 In `openspec/specs/cli-cli-factory/spec.md:74`, replace `internal/config.Load()` reference text with the now-real wrapper. Apply the delta spec at `openspec/changes/wire-factory-config-pipeline/specs/cli-cli-factory/spec.md` via the standard sync step (osc-sync-specs). Update the `Load()` scenario (lines 51-57) to: (a) pin the never-nil `*Config` contract, (b) replace the blanket `*core.ConfigError` with the typed chain `*core.FileError` / `*core.ParseError` / `*core.ConfigError` matching the failure mode (dispatched by `output.FormatError` via `errors.As`). Add a sentence that the wrapper has the same precedence contract as `NewConfigManager().Load()` (defaults → file → env).
- [ ] 6.2 In `openspec/specs/cli-shell-completion/spec.md:305, 312`, replace `getCompletionTimeout(nil)` with `getCompletionTimeout(f.Config())`. Apply the delta at `openspec/changes/wire-factory-config-pipeline/specs/cli-shell-completion/spec.md`. Also carve out `actionPlatforms` as `func(*cmdutil.Factory) carapace.Action` (1-arg, no Cobra command needed for static platform values) — documents a pre-existing 1-arg signature drift; no production-code change required.
- [ ] 6.3 In `openspec/specs/application-configuration/spec.md` (delta), verify the `Config field set` ADDED-Requirement (lines 5-28) documents `Library: ""` correctly in the `DefaultConfig seeds all fields` scenario at lines 16-20. The scenario already pins `Library: empty` and the bullet list documents the field set; no further delta edit needed. No source-of-truth edit needed (the source text uses generic field descriptions).
- [ ] 6.4 Re-run `openspec validate wire-factory-config-pipeline --strict` and confirm all 3 spec deltas + design + proposal + tasks remain coherent.

## 7. Test updates

- [x] 7.1 Add `TestLoadLibraryForCompletion_HonorsConfigTimeout` (~25 LOC) in `cmd/completions_test.go`: build a fresh `*cobra.Command` per test invocation (per golang-spf13-cobra "Re-create the command tree per test" to avoid cobra flag-state bleed); construct a Factory with `f.Config = func() (*config.Config, error) { return cfg, nil }` (direct closure, **not** `OnceValuesFunc` — clearer test seam per golang-design-patterns "design for testability"); set `cfg.Completion.Timeout = "2s"`; invoke `loadLibraryForCompletion(f, libPath)`; assert the resulting context has a 2-second deadline. Wrap the test body in `synctest.Test(t, func(t *testing.T) { ... })` (Go 1.25+) for deterministic time-based assertions. Depends on tasks 5.1-5.2 being applied first.
- [x] 7.2 Add `TestLoadLibraryForCompletion_HonorsCacheTTL` (similar pattern) in `cmd/completions_test.go` for the `CacheTTL` knob. Same `synctest.Test` wrapper as 7.1; fresh `*cobra.Command` per test invocation.
- [x] 7.3 In `test/e2e/library_add_test.go:230-239`, `test/e2e/library_discover_test.go:215-233`, `test/e2e/library_refresh_test.go:156-169`, `test/e2e/library_remove_test.go:217-230`, plus the additional `test/e2e/library_test.go:84` (read-only `library resources`) and `test/e2e/init_test.go:26` (helper used by all init tests): keep the existing `GERMINATOR_LIBRARY` env-var E2E tests (they become backwards-compat proofs); add new E2E tests for `Config.Library` precedence (config-file value applies when env var is unset). Add a new `Describe("using config file library setting")` block in `test/e2e/library_test.go` that writes `~/.config/germinator/config.toml` with `library = "..."` and asserts library resolution.
- [x] 7.4 In `internal/iostreams/iostreams_test.go`, add `TestIOStreams_SetDebug` covering the new `SetDebug(bool)` method (no-debug → no-op handler; debug → debug-level handler writing to ErrOut). Add `TestSystem_NoLongerReadsEnvDebug` asserting that `iostreams.System()` returns a discard `Logger` even when `GERMINATOR_DEBUG=1` is set in the process env (verifies task 3.2's env-read removal).
- [x] 7.5 In `cmd/completions_test.go`, add `TestResolveLibraryPath_PrefersCfgOverEnv` covering `cmd/completions.go:54-65` precedence (flag > env > `cfg.Library` > default) — ensures the intentional env-read at line 54 remains as designed. Build a fresh `*cobra.Command` per test invocation (per golang-spf13-cobra convention).
- [x] 7.6 In `cmd/library_refresh_test.go`, add `TestRefreshLibrary_HonorsConfigLibrary`: construct a Factory with `f.Config = func() (*config.Config, error) { return cfg, nil }`, call `refreshLibrary(f, "")` (no flag, no env), assert the resolved path is `cfg.Library`. Pins task 4.4's closure migration.
- [x] 7.7 In `cmd/library_remove_test.go`, add `TestRemoveLibrary_HonorsConfigLibrary`: same shape as 7.6 for the `removeLibrary` closure. Pins task 4.4's closure migration.
- [x] 7.8 In `cmd/library_refresh_test.go` and `cmd/library_remove_test.go`, add coverage that `f == nil` returns `nil` (existing nil-guard) **and** that `f.Config == nil` falls through to the env-only path (the closure should not panic, per the explicit nil-check pattern from task 4.4).

## 8. Final verification

- [x] 8.1 Run `mise run build` — confirms no broken imports.
- [x] 8.2 Run `mise run lint` — if output shifts, refresh `cmd/testdata/lint_baseline.txt` per `cmd/AGENTS.md` "Lint Baseline Test" procedure.
- [x] 8.3 Run `mise run test` — confirm all unit tests pass.
- [x] 8.4 Run `mise run test -- -race && mise run test:e2e -- -race` — confirm all unit + E2E tests pass with race detection. Required because `cmdutil.OnceValuesFunc` uses `sync.Once` and completion actions may be invoked concurrently (per golang-testing Rule 10). Without `-race`, race conditions on the cache initialization may not surface.
- [x] 8.5 Run `mise run test:coverage` — confirm coverage for `internal/config/`, `internal/iostreams/iostreams.go`, `internal/library/discovery.go` ≥ 70%. (Shell-package coverage for `cmd/completions.go` is tracked separately via the `cmd/lint_test.go` baseline; the 70% threshold applies to internal packages only.)
- [x] 8.6 Run `openspec validate wire-factory-config-pipeline --strict` — confirm all specs and tasks are coherent.
- [x] 8.7 Manually test the migration: edit `~/.config/germinator/config.toml` with `[completion] timeout = "2s" cache_ttl = "10s"`, run `germinator library resources`, confirm completion latency matches the configured timeout.

**Parallelization note**: Tests in section 2 are sequential (env mutations); tests in section 7 may use `t.Parallel()` per-subtest where the test is read-only and doesn't mutate shared state (per golang-testing Rule 4).

## 9. Post-implementation amendments (verification follow-up)

The implementation was reviewed via `osc-verify-change` after section 8 completed. Phase A (`osc-sync-specs`) was deferred; this section documents the code/doc amendments applied during the build.

- [x] 9.1 Fix `resolveConfigPath` docstring (`internal/config/manager.go:124-140`) — the original comment claimed the function does NOT call `xdg.Reload()`, but the implementation does (via the mutex-protected `xdgReload()` helper). Rewrote the comment to accurately describe the implementation.
- [x] 9.2 Add `sync.RWMutex` to `loadFn` (`internal/config/load.go`) — introduced `getLoadFn()` and `swapLoadFn(fn) func()` helpers; `Load()` now reads under RLock. Tests use `t.Cleanup(swapLoadFn(...))` for safe mutation + restore.
- [x] 9.3 Add `sync.RWMutex` to `configLoadForTest` (`internal/cmdutil/factory.go`) — same pattern: `getConfigLoadForTest()` and `swapConfigLoadForTest(fn) func()`. `BuildFactory` now uses the mutex-protected reader. All 8 wiring tests in `factory_wiring_test.go` migrated to `swapConfigLoadForTest`.
- [x] 9.4 Create `internal/paths` package — new `paths.ExpandHome(path string) (string, error)` consolidates tilde expansion. Coverage 87.5%.
- [x] 9.5 Migrate `internal/config/config.go::ExpandPaths` to call `paths.ExpandHome` — the local `expandTilde` helper was removed. The error from `paths.ExpandHome` is wrapped as a `*core.ConfigError` (matching the original behavior). The duplicate `TestExpandTilde` in `config_test.go` was replaced with a `t.Skip` pointer to the canonical coverage in `internal/paths/expand_test.go`.
- [x] 9.6 Migrate `cmd/completions.go::resolveLibraryPath` to call `paths.ExpandHome` — the local `expandTildeInPath` was removed. A tiny `expandTildeForCompletion` wrapper preserves the legacy silent-fallback behavior (returns original path on `os.UserHomeDir` failure). `cmd/completions_test.go::TestResolveLibraryPath_TildeExpansion` updated to call the new helper.
- [x] 9.7 Refresh AGENTS.md doc drifts (4 files):
  - `internal/AGENTS.md` — replaced `AppConfig struct with toml/koanf tags` with the current `Config` struct description (Library, PlatformDefault, Debug, Completion; koanf env provider; Library="" as the canonical "no override" signal).
  - `internal/config/AGENTS.md` — rewrote the Public Surface section with the new Config field table, env-var mapping, Load() wrapper contract, and updated path resolution description.
  - `internal/library/AGENTS.md` — updated FindLibrary signature to 3-arg; documented the 4-tier precedence (flag > env > cfg > default).
  - `cmd/AGENTS.md` — replaced the `GERMINATOR_DEBUG=1 enables...` line with the canonical SetDebug(cfg.Debug) activation flow via koanf env provider.
- [x] 9.8 Rename `TestSetDebug` → `TestIOStreams_SetDebug` (`internal/iostreams/iostreams_test.go:159`).
- [x] 9.9 Rename `TestRefreshLibrary_NilConfigFallsThrough` → `TestRefreshLibrary_FConfigIsNilFallsBack` (`cmd/library_refresh_test.go:153`).
- [x] 9.10 Rename `TestRemoveLibrary_NilConfigFallsThrough` → `TestRemoveLibrary_FConfigIsNilFallsBack` (`cmd/library_remove_test.go:728`).
- [x] 9.11 Add `TestResolveLibraryPath_PrefersCfgOverEnv` (`cmd/completions_test.go`) — two-subtest case covering both precedence directions: env wins when set; cfg consulted when env unset.
- [x] 9.12 Migrate `TestLoadLibraryForCompletion_HonorsConfigTimeout` and `_HonorsCacheTTL` to `synctest.Test` (Go 1.25+) — synthetic time replaces the pre-cancelled-context + `CompletionCache.WithClock` fake clock approach. The Timeout test uses a cancelled parent context (which propagates to the wrapped loadCtx); the CacheTTL test uses `time.Sleep(10s + 1ms)` + `synctest.Wait()` to deterministically evict the cache entry. Imports `testing/synctest`.
- [x] 9.13 Re-run `mise run lint && mise run test -- -race && mise run test:coverage && openspec validate wire-factory-config-pipeline --strict` — all green.

### Coverage results after amendments

| Package | Before | After | Threshold |
|---------|--------|-------|-----------|
| `internal/config` | 85.5% | **87.5%** | ≥ 70% |
| `internal/iostreams` | 93.1% | 93.1% | ≥ 70% |
| `internal/library` | 80.4% | 80.4% | ≥ 70% |
| `internal/cmdutil` | 98.4% | **98.6%** | ≥ 70% |
| `internal/paths` (new) | n/a | **87.5%** | ≥ 70% |
| `cmd` | 83.3% | **83.5%** | (lint baseline) |

### Phase A (deferred)

Spec deltas for `cli-cli-factory`, `cli-shell-completion`, and `application-configuration` remain unsynced. To complete the archive path:

```bash
openspec sync-specs wire-factory-config-pipeline
openspec validate wire-factory-config-pipeline --strict  # re-validate after sync
```
