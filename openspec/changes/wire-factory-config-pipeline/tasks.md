# Tasks — Wire `Factory.Config` and complete the koanf config pipeline

Each task ends with `mise run check` passing.

## 1. Extend the koanf loader

- [ ] 1.1 In `internal/config/manager.go`, add `koanf/providers/env` to `Manager.Load()` with the `GERMINATOR_` prefix and `_` delimiter. Merge order: defaults → file → env. Existing tier 1 (defaults) and tier 2 (file) remain unchanged.
- [ ] 1.2 In `internal/config/config.go`, add fields to the `Config` struct: `Library string` (with `koanf:"library"` tag), `PlatformDefault string` (with `koanf:"platform"` tag), `Debug bool` (with `koanf:"debug"` tag).
- [ ] 1.3 In `internal/config/config.go`, update `DefaultConfig()` to seed the new fields (`Library: ""`, `PlatformDefault: ""`, `Debug: false`).
- [ ] 1.4 In `internal/config/config.go`, extend `Validate()` to cover the new fields: `PlatformDefault` must be empty, `claude-code`, or `opencode` (returns `*core.ConfigError` otherwise); `Debug` is always valid (bool).
- [ ] 1.5 Add `internal/config/load.go` with `func Load() (*Config, error)` — a top-level wrapper that calls `NewConfigManager().Load()` and returns `mgr.GetConfig()`. Fixes the broken reference at `cli-cli-factory/spec.md:74`.

## 2. Update internal/config tests

- [ ] 2.1 In `internal/config/manager_test.go`, add env-provider tests (~50 LOC):
  - `TestLoad_EnvOverridesFile`: with `GERMINATOR_LIBRARY=/env/lib` set and `library = "/file/lib"` in config file, `Load()` returns `Config.Library == "/env/lib"`
  - `TestLoad_EnvOverridesDefault`: with `GERMINATOR_DEBUG=1` set, `Load()` returns `Config.Debug == true`
  - `TestLoad_NoEnvNoFile`: with neither env nor file, `Load()` returns `DefaultConfig()` values
  - `TestLoad_MissingFile`: with file missing but env set, `Load()` returns env-derived values (no error)

## 3. Wire `Factory.Config`

- [ ] 3.1 In `main.go`, after `f.Library = ...` (line 35-38), add `f.Config = cmdutil.OnceValuesFunc(config.Load)`. Verify `internal/config` is imported.
- [ ] 3.2 In `internal/iostreams/iostreams.go`, add `func (s *IOStreams) SetDebug(enabled bool)` method that swaps the `Logger` field to a debug-level handler when `enabled == true` (or to a no-op handler when false). Keep `iostreams.System()` unchanged (it still reads `os.LookupEnv` at construction for backwards compat).
- [ ] 3.3 In `main.go`, after `f.Config = ...` is set, call `cfg, _ := f.Config()`; if `cfg != nil && cfg.Debug`, call `io.SetDebug(true)`. The `nil` fallback is for tests that construct `iostreams.Test()` without a Factory.

## 4. Migrate `GERMINATOR_LIBRARY` reads to `Config.Library`

- [ ] 4.1 In `main.go:36`, update `library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"))` to `library.FindLibrary(cfg.Library, os.Getenv("GERMINATOR_LIBRARY"))` (env still wins).
- [ ] 4.2 In `cmd/show.go:87`, `cmd/resources.go:78`, `cmd/presets.go:68`, `cmd/init.go:139`, `cmd/library_add.go:251`, `cmd/library_create.go:155`, `cmd/library_refresh.go:151`, `cmd/library_remove.go:232`, `cmd/library_validate.go:156`: update each `library.FindLibrary("", os.Getenv("GERMINATOR_LIBRARY"))` call to `library.FindLibrary(opts.Library, os.Getenv("GERMINATOR_LIBRARY"))` where `opts.Library` is a new field populated from `Config.Library` in each command's `RunE`. Pass `Config.Library` to each command via Factory or `f.Config()` call inside `RunE`.
- [ ] 4.3 Verify zero remaining production-code `os.Getenv("GERMINATOR_LIBRARY")` calls via `rg "os\.Getenv\(\"GERMINATOR_LIBRARY\"\)" cmd/ internal/ --type go`. Test files may keep the env-var read as backwards-compat proof.

## 5. Thread `Config` through completion actions

- [ ] 5.1 In `cmd/completions.go:103`, replace `getCompletionTimeout(nil)` with `getCompletionTimeout(cfg)` where `cfg` is loaded via `cfg, _ := f.Config()` at the top of `loadLibraryForCompletion`.
- [ ] 5.2 In `cmd/completions.go:111`, replace `getCacheTTL(nil)` with `getCacheTTL(cfg)`.
- [ ] 5.3 In `cmd/completions.go:120, 132, 144`, replace `resolveLibraryPath(cmd, nil)` with `resolveLibraryPath(cmd, cfg)` for `actionResources`, `actionPresets`, `actionLibraryRefs`. Load `cfg` via `f.Config()` once per action invocation.

## 6. Spec reconciliation

- [ ] 6.1 In `openspec/specs/cli-cli-factory/spec.md:74`, replace `internal/config.Load()` reference text with the now-real wrapper. Apply the delta spec at `openspec/changes/wire-factory-config-pipeline/specs/cli-cli-factory/spec.md` via the standard sync step (osc-sync-specs).
- [ ] 6.2 In `openspec/specs/cli-shell-completion/spec.md:305, 312`, replace `getCompletionTimeout(nil)` with `getCompletionTimeout(f.Config())`. Apply the delta at `openspec/changes/wire-factory-config-pipeline/specs/cli-shell-completion/spec.md`.
- [ ] 6.3 In `openspec/specs/application-configuration/spec.md`, apply the delta at `openspec/changes/wire-factory-config-pipeline/specs/application-configuration/spec.md` (relaxes `adrg/xdg` mandate; adds `Config field set` requirement).

## 7. Test updates

- [ ] 7.1 Add `TestLoadLibraryForCompletion_HonorsConfigTimeout` (~25 LOC) in `cmd/completions_test.go`: construct a Factory with `f.Config` returning a `*Config` containing `Completion.Timeout = "2s"`; invoke `loadLibraryForCompletion(f, libPath)`; assert the resulting context has a 2-second deadline.
- [ ] 7.2 Add `TestLoadLibraryForCompletion_HonorsCacheTTL` (similar pattern) in `cmd/completions_test.go` for the `CacheTTL` knob.
- [ ] 7.3 In `test/e2e/library_add_test.go:230-239`, `test/e2e/library_discover_test.go:215-233`, `test/e2e/library_refresh_test.go:156-169`, `test/e2e/library_remove_test.go:217-230`: keep the existing `GERMINATOR_LIBRARY` env-var E2E tests (they become backwards-compat proofs); add new E2E tests for `Config.Library` precedence (config-file value applies when env var is unset).
- [ ] 7.4 In `internal/iostreams/iostreams_test.go`, add `TestIOStreams_SetDebug` covering the new `SetDebug(bool)` method (no-debug → no-op handler; debug → debug-level handler writing to ErrOut).

## 8. Final verification

- [ ] 8.1 Run `mise run build` — confirms no broken imports.
- [ ] 8.2 Run `mise run lint` — if output shifts, refresh `cmd/testdata/lint_baseline.txt` per `cmd/AGENTS.md` "Lint Baseline Test" procedure.
- [ ] 8.3 Run `mise run test` — confirm all unit tests pass.
- [ ] 8.4 Run `mise run test:e2e` — confirm all E2E tests pass (especially the 4 E2E files with `GERMINATOR_LIBRARY` tests + new config-driven tests).
- [ ] 8.5 Run `mise run test:coverage` — confirm coverage for `internal/config/`, `cmd/completions.go`, `internal/iostreams/iostreams.go` ≥ 70%.
- [ ] 8.6 Run `openspec validate wire-factory-config-pipeline --strict` — confirm all specs and tasks are coherent.
- [ ] 8.7 Manually test the migration: edit `~/.config/germinator/config.toml` with `[completion] timeout = "2s" cache_ttl = "10s"`, run `germinator library resources`, confirm completion latency matches the configured timeout.
