# Tasks — Wire Factory and migrate pilot commands

**Slice 2 of 9.** Wires `main.go` to the new architecture and migrates the two pilot commands (`adapt` + `library resources`). Establishes the `LegacyBridge` shim for non-migrated commands. Deletes `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`.

Each task ends with `mise run check` passing.

## 2.1 Wire `main.go` to the new architecture

- [ ] 2.1.1 Replace `main.go` body with:
  - `ctx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` (the `cancel` is captured by `NewFactory` via the wrapped context — do not discard it; `defer f.Close()` cancels the same context)
  - `io := iostreams.System()`
  - `f := cmdutil.NewFactory(ctx, io, version.Version, "germinator")` (eager values per `internal/cmdutil/AGENTS.md`; signature is now `NewFactory(ctx, io, version, executable)` — the refactor is part of this change so `NewFactory` no longer auto-creates its own context)
  - `defer f.Close()` to cancel `RootContext` on exit (per `internal/cmdutil/AGENTS.md` `Factory.Close()` lifecycle)
  - Add `os/signal` import
- [ ] 2.1.2 Populate every lazy function field on `f` using `cmdutil.OnceValuesFunc[T]` wrappers (per `internal/cmdutil/AGENTS.md` line 23; `OnceValuesFunc` is the foundation's generic helper, equivalent to `sync.OnceValues` but with project-wide consistency):

  ```go
  f.Transformer = cmdutil.OnceValuesFunc(func() (application.Transformer, error) {
      return application.NewTransformer(...)
  })
  // ... repeat for Config, Library, Validator, Canonicalizer, Initializer
  ```

  Each lazy field uses `cmdutil.OnceValuesFunc[T]` so that multiple call sites within one command invocation share a single result (avoids re-reading config, re-parsing, etc.).
- [ ] 2.1.3a Declare the `cmd.LegacyBridge` type in a new file `cmd/legacy_bridge.go` (per design Decision 7):

  ```go
  package cmd

  type LegacyBridge struct {
      Services       *LegacyServices
      ErrorFormatter *ErrorFormatter
      Verbosity      Verbosity
  }

  type LegacyServices struct {
      Transformer   application.Transformer
      Validator     application.Validator
      Canonicalizer application.Canonicalizer
      Initializer   application.Initializer
  }
  ```

  `LegacyBridge` is exported (uppercase) so it can cross the package boundary from `main.go` to `cmd/`. Update `cmd.NewRootCommand` signature to `NewRootCommand(f *cmdutil.Factory, bridge *LegacyBridge) *cobra.Command`. Non-migrated commands take `bridge` as a second parameter (transitional; slice 7 deletes `LegacyBridge` and the second parameter). For task 2.1.3a, `bridge.Services` may be nil; non-migrated commands must handle a nil `Services` field until task 2.1.3b populates it.

- [ ] 2.1.3b (AFTER task 2.5.1 deletes `cmd/container.go`) Populate `bridge.Services` in `main.go` by calling each underlying service constructor directly:

  ```go
  bridge := &cmd.LegacyBridge{
      Services: &cmd.LegacyServices{
          Transformer:   application.NewTransformer(parser, renderer),
          Validator:     application.NewValidator(),
          Canonicalizer: application.NewCanonicalizer(),
          Initializer:   application.NewInitializer(parser, renderer),
      },
      ErrorFormatter: cmd.NewErrorFormatter(),
      Verbosity:      0,
  }
  ```

  The `main.go` import of `internal/application` is temporary and removed in slice 7.
- [ ] 2.1.4 In the post-`Execute` block, replace the existing error handling with:
  - `output.FormatError(f.IOStreams, err)` → writes to stderr
  - `warning.MaybeWarnLegacyExitCode(f.IOStreams)` → emits deprecation warning once per process if gate conditions are met (per design Decisions 6 and 8)
  - `os.Exit(int(cmdutil.ExitCodeFor(err)))` → exits with the mapped code; `cmdutil.ExitCodeFor` remains a pure function (no logger parameter)
- [ ] 2.1.5 After constructing `rootCmd` via `cmd.NewRootCommand(f, bridge)` (the new signature with the Factory and LegacyBridge parameters, defined in task 2.1.3a), call `rootCmd.SetContext(ctx)` so `c.Context()` returns the signal-aware context.
- [ ] 2.1.6 Create `internal/warning/canary.go` with the canary helper:

  ```go
  package warning

  import (
      "os"
      "sync"

      "gitlab.com/amoconst/germinator/internal/iostreams"
  )

  var canaryOnce sync.Once

  func MaybeWarnLegacyExitCode(io *iostreams.IOStreams) {
      if io == nil {
          return
      }
      if os.Getenv("EXIT_CODE_LEGACY") == "" && !io.IsStderrTTY() {
          return
      }
      canaryOnce.Do(func() {
          io.Warnf("exit code 5 was renamed to 1 in slice 2; consult CHANGELOG for the migration timeline")
      })
  }

  func ResetCanaryForTest() {
      canaryOnce = sync.Once{}
  }
  ```

  `Warnf` writes to `io.ErrOut` (the user-facing stderr channel) via `Styles.Warning`. The helper does NOT depend on `io.Logger` (which is gated on `GERMINATOR_DEBUG` and would be a no-op in production). Export `ResetCanaryForTest()` for unit-test cleanup via `t.Cleanup`. Add `internal/warning/canary_test.go` covering: (a) single emission per process; (b) env-var gate (`EXIT_CODE_LEGACY=1` → emits); (c) TTY gate (`io.IsStderrTTY()==true` → emits); (d) suppression when both gates are false; (e) warning emitted even when `io.Logger` is nil; (f) `ResetCanaryForTest()` resets state between sub-tests.

## 2.2 Migrate `cmd/adapt.go`

- [ ] 2.2.1 In `cmd/adapt.go`, define `adaptOptions` struct with fields: `IO *iostreams.IOStreams`, `Transformer func() (Transformer, error)`, `Ctx context.Context`, `InputPath string`, `OutputPath string`, `Platform string`
- [ ] 2.2.2 In `cmd/adapt.go`, define the `Transformer` interface inline per the `golang-cli-architecture` skill principle 8 ("interfaces where consumed") and `internal/AGENTS.md` line 50 (interfaces defined in cmd files for the target architecture):

  ```go
  // Transformer is the local command-side contract for document transformation.
  // Defined in cmd/ per the target architecture; will move to
  // internal/core/contracts.go in change-7 when internal/application/ is deleted.
  type Transformer interface {
      Transform(ctx context.Context, req *TransformRequest) (*core.TransformResult, error)
  }
  ```

  Import `TransformRequest` from `internal/application/requests.go` (still alive in this change).
- [ ] 2.2.3 Implement `NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command`:
  - Build the command tree (`Use: "adapt"`, `Args: cobra.ExactArgs(2)`, etc.)
  - Add `--platform` string flag
  - In `RunE`: construct `opts`, populate from `f.IOStreams`, `c.Context()` (signal-aware), `f.Transformer`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runAdapt(opts)`
- [ ] 2.2.4 Implement `runAdapt(opts *adaptOptions) error`:
  - Validate platform via `core.ValidatePlatform(opts.Platform)`; return the typed error if invalid
  - Resolve the transformer: `transformer, err := opts.Transformer(); if err != nil { return err }`
  - Call `transformer.Transform(opts.Ctx, &TransformRequest{InputPath: opts.InputPath, OutputPath: opts.OutputPath, Platform: opts.Platform})`
  - Write success message to `opts.IO.Out` (stdout, per skill principle 5): `fmt.Fprintf(opts.IO.Out, "wrote %s\n", opts.OutputPath)`
  - Emit verbose progress to `opts.IO.Verbosef("transforming %s → %s", opts.InputPath, opts.OutputPath)` (verbose writes to stderr via `IOStreams.Verbosef`)
- [ ] 2.2.5 Convert the `cmd/cmd_test.go` adapt test to use `iostreams.Test()` + `runF` injection:
  - `NewCmdAdapt(f, runF)` with `runF` capturing `*adaptOptions`
  - `cmd.SetArgs([]string{"adapt", "input.yaml", "output.yaml", "--platform", "claude-code"})`
  - Assert `runF` was called with correct `opts`
  - Call `runAdapt(opts)` directly with a fake `Transformer`; assert stdout contains the expected message and stderr is empty (verifies stream discipline)
- [ ] 2.2.6 Run `mise run check`; confirm `germinator adapt` produces byte-identical output to the pre-change build

## 2.3 Migrate `cmd/library/resources.go`

- [ ] 2.3.1 Split `cmd/library.go`:
  - Move the `resources` sub-command definition into `cmd/library/resources.go`; keep only the parent command in `cmd/library.go`. The parent stays as `NewCmdLibrary(f, bridge, runF)` (note the new `bridge` parameter for the transitional `LegacyBridge`) with the sub-commands attached.
  - **Remove** `cmd.PersistentFlags().Bool("json", false, "Output as JSON")` from the parent command (currently `cmd/library.go:39`). The persistent `--json` flag is REMOVED per the delta spec `library-library-json-output` REMOVED requirement "Library parent command accepts --json flag". After this removal, `germinator library resources --json` (and any other library sub-command's `--json`) will fail with `unknown flag --json` and exit code 2.
- [ ] 2.3.2 In `cmd/library/resources.go`, define `resourcesOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
- [ ] 2.3.3 Note: `library.ListResources` is a **package function** in `internal/library/lister.go:17`, not a method on a `Library` interface. **Do not declare a `Library` interface** in this file; use the concrete `*library.Library` returned from `opts.Library()` and call `library.ListResources(lib)` directly. The returned `map[string][]ResourceInfo` is then flattened into a slice of structs (with `tab:"HEADER"` struct tags for the table exporter) before being passed to the exporters.
- [ ] 2.3.4 Implement `NewCmdResources(f *cmdutil.Factory, runF func(*resourcesOptions) error) *cobra.Command`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)` (per `internal/output/AGENTS.md`, valid values are `["json", "table", "plain"]` with `DefaultOutputFormat = "plain"`)
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and the parsed `--output` flag
  - Call `runF(opts)` if non-nil, else `runResources(opts)`
- [ ] 2.3.5 Implement `runResources(opts *resourcesOptions) error`:
  - Resolve the library: `lib, err := opts.Library(); if err != nil { return err }`
  - Call `grouped := library.ListResources(lib)`; flatten into a `[]resourcesRow` (a local struct with `tab:"HEADER"` tags: `Type string \`tab:"TYPE"\``, `Name string \`tab:"NAME"\``, `Description string \`tab:"DESCRIPTION"\``, `Platform string \`tab:"PLATFORM"\``)
  - Dispatch on `opts.Output`:
    - `"json"`: `return output.NewJSONExporter().Write(opts.IO, rows)` — writes to `opts.IO.Out` (stdout, per skill principle 5)
    - `"table"`: `return output.NewTableExporter().Write(opts.IO, rows)` — writes to `opts.IO.Out` (stdout)
    - `"plain"` (default): reuse the existing `formatResourcesList(lib)` from `cmd/library_formatters.go:15-57` (the grouped "Skills:\n  ...\n\nAgents:\n  ...\n" output). Print via `_, err = fmt.Fprint(opts.IO.Out, formatResourcesList(lib))`. Plain output MUST remain byte-identical to the pre-change build per the delta spec scenario "Plain is the default".
  - All three formats write to stdout; verbose progress (none in this command) goes to stderr via `opts.IO.Verbosef`
- [ ] 2.3.6 Convert the resources test in `cmd/cmd_test.go` to use `iostreams.Test()` + `runF` injection; assert plain/JSON/table output for each format. Also assert stream discipline: stdout contains the data, stderr is empty (no verbose leakage into stdout).
- [ ] 2.3.7 Run `mise run check`; confirm `germinator library resources`, `germinator library resources --output json`, `germinator library resources --output table` produce expected output. Also confirm `germinator library resources --json` returns exit code 2 (old `--json` flag rejected per spec scenario).

## 2.4 Update tests for new patterns

- [ ] 2.4.1 The adapt test rewrite is owned by task 2.2.5; the resources test rewrite is owned by task 2.3.6. After both run, audit `cmd/cmd_test.go` for any remaining references to deleted legacy types (`ServiceContainer`, `CommandConfig`, `CategorizeError`); remove them as part of the section's rewrite.

- [ ] 2.4.2a Add a minimal legacy test-only adapter in `cmd/legacy_test_helpers_test.go` (file suffix `_test.go` limits it to test builds). The adapter re-exports `newTestConfig() *cmdutil.Factory` for non-pilot tests that still need a Factory instance, plus shim helpers (`NewServiceContainer()`, `*CommandConfig`, `cmd.HandleCLIError`) that satisfy non-pilot test sections until slices 3-7 convert them. Tag the file with `//nolint:paralleltest` and a TODO pointing to slice 7 for removal.

- [ ] 2.4.2b Leave non-pilot sub-sections untouched (they still use `NewServiceContainer` and the legacy `CommandConfig` via the test-only adapter from 2.4.2a); they will be converted in changes 3-7. The adapter is deleted in change-7.
- [ ] 2.4.3 Verify `cmd/lint_test.go` exists. If it does NOT exist, create it with a smoke test that asserts no NEW `fmt.Fprintf(os.Stdout|Stderr)` or `os.Exit(` patterns in `cmd/adapt.go` or `cmd/library/resources.go`. If it already exists, add the smoke test to it directly.
- [ ] 2.4.4 Add unit tests in `internal/warning/canary_test.go` (paired with the canary helper from task 2.1.6). Cover: (a) single emission per process (call `MaybeWarnLegacyExitCode` twice; assert the second call is a no-op); (b) `EXIT_CODE_LEGACY=1` env var triggers emission; (c) `io.IsStderrTTY()==true` triggers emission; (d) both gates false suppresses emission; (e) warning is emitted to `io.ErrOut` even when `io.Logger` is nil (the canary does not depend on the Logger); (f) `ResetCanaryForTest()` resets state between sub-tests.
- [ ] 2.4.5 Add golden file tests in `test/e2e/library_resources_test.go` (build tag `e2e`) for each output format. Use byte-level comparison against fixtures in `test/e2e/fixtures/library-resources/{plain,json,table}/`. The fixtures are generated by capturing pre-change `germinator library resources --output X` output and committed alongside the test.

## 2.5 Delete legacy files

- [ ] 2.5.1 Delete `cmd/container.go` (`ServiceContainer` is no longer wired; `LegacyBridge.Services` is constructed in `main.go` by calling each underlying service constructor directly per design Decision 7 — see task 2.1.3b)
- [ ] 2.5.2 Delete `cmd/command_config.go` (`CommandConfig` no longer exists)
- [ ] 2.5.3 Delete `cmd/error_handler.go` (exit codes now mapped via `cmdutil.ExitCodeFor`)
- [ ] 2.5.4 Delete the legacy body of `cmd/adapt.go` (already replaced by the migrated version)
- [ ] 2.5.5 Delete the legacy body of `cmd/library/resources.go` (already replaced by the migrated version)

## 2.6 Update delta spec

- [ ] 2.6.1 Verify the delta spec at `openspec/changes/wire-factory-and-pilots/specs/library-library-json-output/spec.md` (renamed per Phase 0 of the correction plan). Confirm:
  - `## REMOVED Requirements` section lists all 7 obsolete `--json` requirements with `**Reason:**` blockquotes
  - `## MODIFIED Requirements` section's "library resources supports --output flag" requirement uses `## MODIFIED Requirements` not `## ADDED Requirements` (the requirement is a modification of the base spec's JSON output requirements, not a brand-new capability)
  - All four scenarios are present: Plain is the default; JSON output via `--output json`; Table output via `--output table`; Old `--json` flag is rejected
  - The stream contract blockquote at the top of the spec is present
- [ ] 2.6.2 Create the new delta spec at `openspec/changes/wire-factory-and-pilots/specs/cli-exit-codes/spec.md` (already drafted) with `## ADDED Requirements` → `### Requirement: Exit code deprecation canary`. Confirm:
  - All six scenarios are present: interactive session, CI suppression, explicit `EXIT_CODE_LEGACY`, single-emission, nil-safety, exit-code-2 exclusion
  - The requirement SHALL/MUST language is consistent with the base spec (`openspec/specs/cli-exit-codes/spec.md`)
  - The helper signature `MaybeWarnLegacyExitCode(io *iostreams.IOStreams)` is referenced
  - `cmdutil.ExitCodeFor` purity is preserved (no logger parameter, no side effects)
- [ ] 2.6.3 Verify `cmdutil.NewFactory`, `cmdutil.Factory.Close()`, `cmdutil.ExitCodeFor`, and `cmdutil.AddOutputFlags` match the requirements in `openspec/specs/cli-factory/spec.md`, `cli-cli-factory/spec.md`, and `cli-output-formats/spec.md`. No file edits required; this is a documentation cross-check.
- [ ] 2.6.4 Run `openspec show wire-factory-and-pilots --json` and confirm both delta specs appear:
  - `library-library-json-output` (renamed folder) — with 7 REMOVED requirements and 1 MODIFIED requirement
  - `cli-exit-codes` — with 1 ADDED requirement (`Exit code deprecation canary`) and 6 scenarios

## 2.7 Verification

- [ ] 2.7.1 Run `mise run lint` — confirm no new violations (including `forbidigo` checks for `fmt.Fprintf(os.Stdout|Stderr)`, `os.Exit(`, `var global(Factory|CommandConfig)` outside `main.go` and tests)
- [ ] 2.7.2 Run `mise run test` — confirm all unit tests pass (adapt + resources tests pass; non-pilot tests still pass via LegacyBridge; new `internal/warning/canary_test.go` covers all gate conditions)
- [ ] 2.7.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 2.7.4 Run `mise run test:coverage` — confirm coverage maintained for migrated packages; new `internal/warning/` package has ≥70% coverage
- [ ] 2.7.5 Run `mise run test:e2e` — confirm E2E tests pass for adapt and resources (golden file tests from task 2.4.5 included)
- [ ] 2.7.6 Smoke-test every command end-to-end:
  - `germinator --help`
  - `germinator adapt` (with sample input)
  - `germinator validate --help` (non-migrated, should still work via LegacyBridge)
  - `germinator canonicalize --help`
  - `germinator init --help`
  - `germinator library --help`
  - `germinator library resources`
  - `germinator library resources --output json`
  - `germinator library resources --output table`
  - `germinator library resources --json` (verify exit code 2 per spec scenario "Old --json flag is rejected")
  - `germinator config --help`
  - `germinator version`
  - `germinator completion bash | head -5`
  - Canary emission (exit code 1 path, TTY state controlled by env var since `IsStderrTTY()` reflects actual fd state and is hard to mock from a shell):
    - `EXIT_CODE_LEGACY=1 germinator adapt /nonexistent.yaml /tmp/out.yaml` — emits canary warning to stderr once, exit code 1
    - `germinator adapt /nonexistent.yaml /tmp/out.yaml` (without `EXIT_CODE_LEGACY`, in non-TTY pipeline) — no warning, exit code 1
    - For interactive TTY verification: use `script(1)` to allocate a pty: `script -q -c "EXIT_CODE_LEGACY=1 ./bin/germinator adapt ..." /dev/null` (Linux only; macOS `script` differs).
  - Canary exclusion (exit code 2 path):
    - `germinator --invalid-flag` — exit code 2, no canary warning regardless of `EXIT_CODE_LEGACY`
- [ ] 2.7.7 Byte-identical output for `germinator adapt <file> <out> --platform <p>` is verified by the golden file test from task 2.4.5 against a pre-change build fixture. Manual sanity check that the golden test passes is a backstop, not the primary verification.
- [ ] 2.7.8 Verify exit codes via the new unit tests in `internal/warning/canary_test.go` (covers gate conditions) and the existing `internal/cmdutil/exit_test.go` (covers the mapping). Manual confirmation: `germinator --invalid-flag` returns 2; `germinator adapt nonexistent-file.yaml /tmp/out.yaml` returns 1; `EXIT_CODE_LEGACY=1 germinator adapt nonexistent-file.yaml /tmp/out.yaml` emits the canary warning to stderr once.
- [ ] 2.7.9 Update `cmd/AGENTS.md` to include `adapt` and `library resources` as canonical examples of the new pattern, and `internal/warning/AGENTS.md` for the canary helper. Defer broader doc updates to `osx-maintain-ai-docs` after archive.
