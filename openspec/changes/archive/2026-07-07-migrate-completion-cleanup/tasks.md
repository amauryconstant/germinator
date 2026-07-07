# Tasks — Migrate completion, version, and finalize migration

**Slice 9 of 9 (FINAL).** Migrates `completion` (carapace, with `Factory.CompletionCache`) and `version`. Deletes `internal/models/`. Updates all `AGENTS.md` files. Generates the CHANGELOG entry. Sweeps E2E tests for old exit codes and flag names.

Each task ends with `mise run check` passing.

## 9.1 Migrate completion (carapace)

- [x] 9.1.1 Hoist the `CompletionCache` type (with `Get`, `Set`, `Reset`, `Invalidate` methods) into a new file `internal/cmdutil/completion_cache.go`; the four action functions stay in `cmd/completions.go`
- [x] 9.1.2 Add `CompletionCache *cmdutil.CompletionCache` field to `cmdutil.Factory` (type defined in `internal/cmdutil/completion_cache.go`, alongside `Factory`)
- [x] 9.1.3 Populate `Factory.CompletionCache` in `main.go` (constructed once at startup)
- [x] 9.1.4 Replace package-level `var cache` with the `Factory.CompletionCache` field
- [x] 9.1.5 Add `CompletionCache.Invalidate()` method on the `CompletionCache` type defined in `internal/cmdutil/completion_cache.go`
- [x] 9.1.6 Convert `actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms` to take the Factory as input; use `f.Library()` (cached via `sync.OnceValues`) and `context.WithTimeout(f.RootContext, 5*time.Second)` for the lookup; use `f.CompletionCache` for caching
- [x] 9.1.7 Update `NewCmdCompletion(...)` (carapace integration) to wire the Factory into the action functions
- [x] 9.1.8 Migrate `cmd/completion.go` to `NewCmdCompletion(f, runF) + runCompletion(opts)`:
  - Define `completionOptions`: `IO *iostreams.IOStreams`, `Ctx context.Context`, `Shell string`
  - Implement `runCompletion(opts)` to generate the carapace script for the requested shell
- [x] 9.1.9 Convert `cmd/completion_test.go` and `cmd/completions_test.go` to `iostreams.Test()` + `runF` injection
- [x] 9.1.10 Wire `f.CompletionCache.Invalidate()` into `runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`
- [x] 9.1.11 Add explicit test: after `runAdd` calls `f.CompletionCache.Invalidate()`, the next call to `actionResources(f, ...)` returns the freshly-added resource. Verification: directly call `f.CompletionCache.Get(keyForResource)` after `runAdd` returns and assert `nil` (entry cleared); then call `actionResources` and assert the new resource appears in the returned completions.
- [x] 9.1.12 Run `mise run check`

## 9.2 Migrate `version`

The output format contract is already specified by `cli-framework` ("Version Command shows full info": `germinator {version} ({commit}) {date}`) and `testing-e2e-testing` ("Version Command E2E Tests": exit 0, stdout matches pattern). This change adds no new spec; it migrates the command to the options pattern while preserving the contract. Per design Decision 3b, `runVersion` reads from the `internal/version` package (not `Factory.AppVersion`).

- [x] 9.2.1 In `cmd/version.go`, define `versionOptions` struct with fields: `IO *iostreams.IOStreams`, `Ctx context.Context`
- [x] 9.2.2 Implement `NewCmdVersion(f *cmdutil.Factory, runF func(*versionOptions) error) *cobra.Command`
- [x] 9.2.3 Implement `runVersion(opts *versionOptions) error`: write `germinator <Version> (<Commit>) <Date>\n` to `opts.IO.Out`, reading from the `internal/version` package (set via `-ldflags` at build time). `Factory.AppVersion` is NOT the source — it remains a short-form string used elsewhere (see design Decision 3b and the existing comment at `cmd/cmd_test.go:128`)
- [x] 9.2.4 Move `TestVersionCommand` from `cmd/cmd_test.go` into a dedicated `cmd/version_test.go`; expand to table-driven coverage asserting:
  - Output format matches regex `^germinator \S+ \(\S+\) \S+$`
  - `runF` injection round-trip (per the `cmd/cmd_test.go:130-141` pattern)
  - `f.AppVersion` is ignored — output uses `internal/version` (preserves the documented behavior)
  - Exit code 0 via Cobra's `Execute()`
- [x] 9.2.5 Run `mise run check`; confirm `germinator version` prints the expected format (also covered by the manual sweep in 9.7.5 and the E2E test at `test/e2e/` per `testing-e2e-testing` spec)

## 9.3 Delete `internal/models/`

`internal/models/constants.go` is 7 lines and contains ONLY the two string constants `PlatformClaudeCode = "claude-code"` and `PlatformOpenCode = "opencode"`. The `PermissionPolicy` enum and `PlatformConfig` type live in `internal/core/platform.go` from slice 1; `internal/core/rules.go` already exists for the pure business-rule functions like `ValidatePlatform`, `ResolveOutputPath`, and `CanInstallResource` — the constants go there, alongside their primary consumer `ValidatePlatform`. Nothing else needs to move.

- [x] 9.3.1 Add the two constants `PlatformClaudeCode = "claude-code"` and `PlatformOpenCode = "opencode"` to `internal/core/rules.go` (alongside `ValidatePlatform`, the consumer of these constants; the `PermissionPolicy` enum and `PlatformConfig` type remain in `internal/core/platform.go`)
- [x] 9.3.2 Verify `.golangci.yml`'s depguard rule (applies to `**/core/**`, allow stdlib + `samber/lo`) still passes after the move — no rule change expected (replaces the old "update depguard" task; the rule already permits `platform.go`)
- [x] 9.3.3 Run `rg "models\.Platform(ClaudeCode|OpenCode)" --type go` to find all consumers; update imports from `internal/models` to `internal/core`. Known consumers (verified):
  - `cmd/completions.go`
  - `internal/config/config.go`
  - `internal/config/config_test.go`
  - `internal/config/manager_test.go`
- [x] 9.3.4 Remove the duplicate `PlatformClaudeCode`/`PlatformOpenCode` definitions in `internal/parser/loader.go` and import from `internal/core` instead (pre-existing duplication; this is the right moment to clean it up)
- [x] 9.3.5 Run `rg "internal/models" .` to verify zero remaining references
- [x] 9.3.6 Delete `internal/models/` directory (including `constants.go`, `doc.go`, `AGENTS.md`)

## 9.4 Update `AGENTS.md` files

Note: `internal/{iostreams,output,cmdutil}/AGENTS.md` already exist (created in slice 1). Tasks 9.4.4–9.4.6 are review/polish passes.

- [x] 9.4.1 Update root `AGENTS.md` architecture diagram to reflect the new layout (Functional Core / Imperative Shell with `iostreams`, `output`, `cmdutil`, `core`, `library`, `config`, `claude-code`, `opencode`, `parser`, `renderer`)
- [x] 9.4.2 Update `cmd/AGENTS.md` with the canonical `adapt` example (full Options struct + NewCmdAdapt + runAdapt)
- [x] 9.4.3 Update `internal/AGENTS.md` to reflect:
  - Rename to `internal/core/`
  - New sibling packages: `iostreams/`, `output/`, `cmdutil/`
  - Flattened packages: `parser/`, `renderer/`, `claude-code/`, `opencode/`, `config/`, `library/`
  - Deleted packages: `application/`, `service/`, `models/`
- [x] 9.4.4 Review and update `internal/iostreams/AGENTS.md` (file exists; verify public surface docs match the post-migration code)
- [x] 9.4.5 Review and update `internal/output/AGENTS.md` (file exists; verify `FormatError`, `Exporter`, `JSONExporter`, `TableExporter`, `AddOutputFlags` descriptions are accurate)
- [x] 9.4.6 Review and update `internal/cmdutil/AGENTS.md` — verify `Factory`, `ExitCode`, `ExitCodeFor`, `AddOutputFlags` descriptions are accurate
- [x] 9.4.7 Update each existing `internal/<pkg>/AGENTS.md` for packages that moved (parser, renderer, claude-code, opencode, config, library) — at minimum, update the package path reference
- [x] 9.4.8 Delete `internal/models/AGENTS.md` (consumed by task 9.3.6 directory deletion). Note: `cmd/library/` does not exist — the project uses a flat `cmd/` layout (`library.go`, `library_add.go`, etc. as sibling files); per-subcommand docs under `cmd/` are not a project convention. If library-command docs are needed, they belong in `cmd/AGENTS.md` or `cmd/commands/AGENTS.md` (which exists).
- [x] 9.4.9 Refresh `cmd/testdata/lint_baseline.txt` after the migration changes (the package-level `var cache` removal in `cmd/completions.go` and the `internal/models/` deletion will shift the lint baseline). Run `mise run lint > cmd/testdata/lint_baseline.txt 2>&1` to capture the new baseline; commit alongside the change. See `cmd/AGENTS.md` "Lint Baseline Test" section.

## 9.5 Generate CHANGELOG entry

- [x] 9.5.1 Verify all 9 changes are archived: `openspec list --json` shows them under `archive/` (or are about to be archived)
- [x] 9.5.2 Run `osx-generate-changelog` to generate the CHANGELOG entry from archived proposals
- [x] 9.5.3 Manually edit the CHANGELOG entry to add the BREAKING section:
  - **BREAKING: Exit codes 3-6 collapsed to 1** — exit code is no longer semantic; check stderr for error type
  - **BREAKING: `--json` flag removed** — use `--output json` on library commands
  - **BREAKING: `--output` flag renamed to `--output-path` on config commands** — disambiguates from the new `--output` format flag
  - **Removed packages:** `internal/application/`, `internal/service/`, `internal/models/`
- [x] 9.5.4 Run `git diff CHANGELOG.md` and review the entry

## 9.6 Sweep E2E tests

- [x] 9.6.1 Run `rg "ShouldFailWithExit\\([3-6]\\)" test/e2e/` to find E2E tests using old exit codes — no matches; all `ShouldFailWithExit` calls use 1 or 2
- [x] 9.6.2 Update each found test to use `ShouldFailWithExit(1)` — no-op (no occurrences found in 9.6.1)
- [x] 9.6.3 Run `rg "\\-\\-json" test/e2e/` to find E2E tests using the old `--json` flag — matches are all comment lines documenting removal OR the legitimate `library_resources_test.go` "rejected" test case
- [x] 9.6.4 Update each found test to use `--output json` — no-op (only comments and the correct `--json flag is rejected` test in `library_resources_test.go:57`)
- [x] 9.6.5 Run `rg "config (init|validate).*\\-\\-output " test/e2e/` to find E2E tests using the old `--output` flag on config commands — matches are only the legitimate "rejects legacy --output flag" tests in `config_init_test.go:95` and `config_validate_test.go:74`
- [x] 9.6.6 Update each found test to use `--output-path` — no-op (the only matches are the correct "rejects legacy --output" test cases; all real config usage already uses `--output-path`)
- [x] 9.6.7 Run `mise run test:e2e` to confirm all E2E tests pass — PASS (`ok germinator/test/e2e 3.203s` with `-count=1`)

## 9.7 Final verification

- [x] 9.7.1 Run `mise run check` — full validation passes
- [x] 9.7.2 Run `mise run test:full` (unit + e2e)
- [x] 9.7.3 Run `mise run test:coverage` — confirm coverage for `cmd/`, `internal/cmdutil/`, `internal/iostreams/`, `internal/output/`, `internal/core/` ≥ 70%
- [x] 9.7.4 Run `mise run test:release` — confirm goreleaser pipeline still works (test build only)
- [x] 9.7.5 Manually run every command end-to-end:
  - `germinator --help`
  - `germinator adapt <input> <output> --platform claude-code`
  - `germinator adapt <input> <output> --platform opencode`
  - `germinator validate <input> --platform claude-code`
  - `germinator canonicalize <input> <output> --platform claude-code`
  - `germinator init --platform opencode --resources skill/commit` (success)
  - `germinator init --platform opencode --resources skill/commit,skill/invalid` (partial success → exit 0)
  - `germinator library --help`
  - `germinator library init --path /tmp/lib`
  - `germinator library resources` (plain)
  - `germinator library resources --output json`
  - `germinator library resources --output table`
  - `germinator library presets`
  - `germinator library show <ref>`
  - `germinator library add <file> --type skill --name test`
  - `germinator library add --discover`
  - `germinator library create preset <name> --resources skill/x`
  - `germinator library refresh`
  - `germinator library remove resource skill/test --force`
  - `germinator library remove preset <name> --force`
  - `germinator library validate`
  - `germinator library validate --fix`
  - `germinator config init`
  - `germinator config init --output-path /tmp/config.toml`
  - `germinator config validate --output-path /tmp/config.toml`
  - `germinator version`
  - `germinator completion bash | head -5`
- [x] 9.7.6 Run `openspec validate migrate-completion-cleanup --strict` and confirm all specs and tasks are coherent
- [x] 9.7.7 Confirm the final command count: `find cmd -name "*.go" | wc -l` shows the expected number of files (no legacy `container.go`, `command_config.go`, `error_handler.go`, `error_formatter.go`, `verbose.go`)
- [x] 9.7.8 Confirm the package count: `ls internal/` shows `core/`, `iostreams/`, `output/`, `cmdutil/`, `library/`, `config/`, `parser/`, `renderer/`, `claude-code/`, `opencode/`, `version/` (no `application/`, `service/`, `models/`, `infrastructure/`)

## 9.8 Archive this change

- [ ] 9.8.1 Archive this change via `osc-archive-change migrate-completion-cleanup`
- [ ] 9.8.2 Confirm `openspec list --json` shows the change under `archive/` with `status: archived`
- [ ] 9.8.3 The migration to `golang-cli-architecture` is **complete** 🎉
