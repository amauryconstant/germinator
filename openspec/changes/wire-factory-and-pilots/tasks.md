# Tasks — Wire Factory and migrate pilot commands

**Slice 2 of 9.** Wires `main.go` to the new architecture and migrates the two pilot commands (`adapt` + `library resources`). Establishes the `legacyBridge` shim for non-migrated commands. Deletes `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go`.

Each task ends with `mise run check` passing.

## 2.1 Wire `main.go` to the new architecture

- [ ] 2.1.1 Replace `main.go` body with:
  - `ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM); defer cancel()`
  - `io := iostreams.System()`
  - `f := &cmdutil.Factory{IOStreams: io, AppVersion: version.Version, Executable: "germinator", RootContext: ctx}`
  - Populate every lazy function field on `f` (`Config`, `Library`, `Transformer`, `Validator`, `Canonicalizer`, `Initializer`)
  - Add `os/signal` import
- [ ] 2.1.2 Add `legacyBridge` shim in `main.go`:

  ```go
  type legacyBridge struct {
      Services       *ServiceContainer
      ErrorFormatter *ErrorFormatter
      Verbosity      Verbosity
  }
  ```

  Construct it once: call the Factory functions to populate `Services` (or call `NewServiceContainer()` directly since `cmd/container.go` is still alive at this point); construct `ErrorFormatter` and `Verbosity` from the legacy types.
- [ ] 2.1.3 In the post-`Execute` block, replace the existing error handling with:
  - `output.FormatError(f.IOStreams, err)` → writes to stderr
  - `os.Exit(int(cmdutil.ExitCodeFor(err)))` → exits with the mapped code
- [ ] 2.1.4 Set `rootCmd.SetContext(ctx)` so `c.Context()` returns the signal-aware context
- [ ] 2.1.5 Add the exit-code canary deprecation warning in `cmdutil.ExitCodeFor`: on the first call per process, if `EXIT_CODE_LEGACY` env var is set OR `opts.IOStreams.IsStdoutTTY()` (no, stderr) — emit `Logger.Warn("exit code 5 was renamed to 1 (validate-v2 will remove this warning)", ...)` via the Factory's Logger. Use `sync.Once` to ensure single emission per process.
- [ ] 2.1.6 Confirm `cmd/container.go`, `cmd/command_config.go`, `cmd/error_handler.go` are NOT yet deleted (deletion happens in task 2.5.x); `legacyBridge` still references them

## 2.2 Migrate `cmd/adapt.go`

- [ ] 2.2.1 In `cmd/adapt.go`, define `adaptOptions` struct with fields: `IO *iostreams.IOStreams`, `Transformer func() (Transformer, error)`, `Ctx context.Context`, `InputPath string`, `OutputPath string`, `Platform string`
- [ ] 2.2.2 In `cmd/adapt.go`, define the `Transformer` interface (one method: `Transform(ctx, *TransformRequest) (*core.TransformResult, error)`); import the `TransformRequest` type from `internal/application/requests.go` (still alive in this change)
- [ ] 2.2.3 Implement `NewCmdAdapt(f *cmdutil.Factory, runF func(*adaptOptions) error) *cobra.Command`:
  - Build the command tree (`Use: "adapt"`, `Args: cobra.ExactArgs(2)`, etc.)
  - Add `--platform` string flag
  - In `RunE`: construct `opts`, populate from `f.IOStreams`, `f.RootContext` (via `c.Context()`), `f.Transformer`, and parsed flags
  - Call `runF(opts)` if non-nil, else `runAdapt(opts)`
- [ ] 2.2.4 Implement `runAdapt(opts *adaptOptions) error`:
  - Validate platform via `core.ValidatePlatform(opts.Platform)`
  - Call `transformer.Transform(opts.Ctx, &TransformRequest{...})`
  - Write success message to `opts.IO.Out`
  - Emit verbose progress to `opts.IO.Verbosef("transforming %s → %s", opts.InputPath, opts.OutputPath)`
- [ ] 2.2.5 Convert the `cmd/cmd_test.go` adapt test to use `iostreams.Test()` + `runF` injection:
  - `NewCmdAdapt(f, runF)` with `runF` capturing `*adaptOptions`
  - `cmd.SetArgs([]string{"adapt", "input.yaml", "output.yaml", "--platform", "claude-code"})`
  - Assert `runF` was called with correct `opts`
  - Call `runAdapt(opts)` directly with a fake Factory; assert stdout contains the expected message
- [ ] 2.2.6 Run `mise run check`; confirm `germinator adapt` produces byte-identical output to the pre-change build

## 2.3 Migrate `cmd/library/resources.go`

- [ ] 2.3.1 Split `cmd/library.go` if necessary: move the `resources` sub-command definition into `cmd/library/resources.go`; keep only the parent command in `cmd/library.go`
- [ ] 2.3.2 In `cmd/library/resources.go`, define `resourcesOptions` struct with fields: `IO *iostreams.IOStreams`, `Library func() (*library.Library, error)`, `Ctx context.Context`, `Output string`
- [ ] 2.3.3 Define the `Library` interface in `cmd/library/resources.go` (the methods actually called: `ListResources(ctx) ([]library.Resource, error)`)
- [ ] 2.3.4 Implement `NewCmdResources(f *cmdutil.Factory, runF func(*resourcesOptions) error) *cobra.Command`:
  - Call `cmdutil.AddOutputFlags(cmd, &opts.Output)`
  - Populate `opts` in `RunE` from `f.IOStreams`, `c.Context()`, `f.Library`, and the parsed `--output` flag
  - Call `runF(opts)` if non-nil, else `runResources(opts)`
- [ ] 2.3.5 Implement `runResources(opts *resourcesOptions) error`:
  - Call `lib.ListResources(opts.Ctx)`
  - If `opts.Output == "plain"`: print as plain text (current default)
  - If `opts.Output == "json"`: construct `output.JSONExporter`, call `Write(opts.IO, data)`
  - If `opts.Output == "table"`: construct `output.TableExporter`, call `Write(opts.IO, data)`
- [ ] 2.3.6 Convert the resources test in `cmd/cmd_test.go` to use `iostreams.Test()` + `runF` injection; assert plain/JSON/table output for each format
- [ ] 2.3.7 Run `mise run check`; confirm `germinator library resources`, `germinator library resources --output json`, `germinator library resources --output table` produce expected output

## 2.4 Update tests for new patterns

- [ ] 2.4.1 Convert all sub-sections of `cmd/cmd_test.go` for adapt and resources to `iostreams.Test()` + `runF` injection
- [ ] 2.4.2 Leave non-pilot sub-sections untouched (they still use `NewServiceContainer` and the legacy `CommandConfig`); they will be converted in changes 3-7
- [ ] 2.4.3 Add a smoke test in `cmd/lint_test.go` (added in change-1) that asserts no NEW `fmt.Fprintf(os.Stdout|Stderr)` or `os.Exit(` patterns in `cmd/adapt.go` or `cmd/library/resources.go`

## 2.5 Delete legacy files

- [ ] 2.5.1 Delete `cmd/container.go` (`ServiceContainer` is no longer wired; `legacyBridge` calls the underlying service constructors directly)
- [ ] 2.5.2 Delete `cmd/command_config.go` (`CommandConfig` no longer exists)
- [ ] 2.5.3 Delete `cmd/error_handler.go` (exit codes now mapped via `cmdutil.ExitCodeFor`)
- [ ] 2.5.4 Delete the legacy body of `cmd/adapt.go` (already replaced by the migrated version)
- [ ] 2.5.5 Delete the legacy body of `cmd/library/resources.go` (already replaced by the migrated version)

## 2.6 Update delta spec

- [ ] 2.6.1 Update `openspec/changes/scaffold-cli-foundation/specs/cli/exit-codes/spec.md`: mark the "exit codes collapsed" requirement as **fulfilled by this change** (with a note that the canary still warns)
- [ ] 2.6.2 Update `openspec/changes/scaffold-cli-foundation/specs/cli/framework/spec.md`: mark `CommandConfig` removal as **fulfilled by this change**
- [ ] 2.6.3 Update `openspec/changes/wire-factory-and-pilots/specs/library/library-json-output/spec.md`: add a delta requirement that `--output json` is available on `library resources`

## 2.7 Verification

- [ ] 2.7.1 Run `mise run lint` — confirm no new violations
- [ ] 2.7.2 Run `mise run test` — confirm all unit tests pass (adapt + resources tests pass; non-pilot tests still pass via legacyBridge)
- [ ] 2.7.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 2.7.4 Run `mise run test:coverage` — confirm coverage maintained for migrated packages
- [ ] 2.7.5 Run `mise run test:e2e` — confirm E2E tests pass for adapt and resources
- [ ] 2.7.6 Smoke-test every command end-to-end:
  - `germinator --help`
  - `germinator adapt` (with sample input)
  - `germinator validate --help` (non-migrated, should still work via legacyBridge)
  - `germinator canonicalize --help`
  - `germinator init --help`
  - `germinator library --help`
  - `germinator library resources`
  - `germinator library resources --output json`
  - `germinator library resources --output table`
  - `germinator config --help`
  - `germinator version`
  - `germinator completion bash | head -5`
- [ ] 2.7.7 Manually verify byte-identical output for `germinator adapt <file> <out> --platform <p>` (both platforms) compared against a pre-change build
- [ ] 2.7.8 Manually verify exit codes: `germinator --invalid-flag` returns 2; `germinator adapt nonexistent-file.yaml /tmp/out.yaml` returns 1
- [ ] 2.7.9 Update `cmd/AGENTS.md` to include `adapt` and `library resources` as canonical examples of the new pattern
- [ ] 2.7.10 Update `internal/AGENTS.md` if the section on `cmd/container.go` references needs updating
