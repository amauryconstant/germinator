# Tasks — Scaffold golang-cli-architecture foundation

**Slice 1 of 9.** Lands only the foundation packages, renames, and lint rules. No command is migrated; `main.go` is untouched; legacy files are not deleted. Subsequent changes (2–9) build on this.

Each task ends with `mise run check` (or a narrower check like `mise run lint`) passing.

## 1.1 Land the new scaffolding packages

- [ ] 1.1.1 Create `internal/iostreams/iostreams.go` with `IOStreams` struct (`In`, `Out`, `ErrOut`, `Verbose bool`, `Logger *slog.Logger`, `Styles Styles`), `System()` constructor (uses `golang.org/x/term` for TTY detection), `Test()` constructor (buffer-backed for unit tests), `IsStdoutTTY()`, `IsInteractive()`, `Verbosef(format string, args ...any)`, `SetStdoutTTY(bool)`. `System()` must gate the `Logger` on the `GERMINATOR_DEBUG` env var: a no-op handler (writes to `io.Discard`) when unset, and a debug-level structured handler writing to `ErrOut` when set to any non-empty value (see `specs/cli/iostreams/spec.md` "Structured Logger" requirement)
- [ ] 1.1.2 Create `internal/iostreams/styles.go` with `Styles` struct (`Error`, `Success`, `Warning`, `Dim`, `Bold`) using `github.com/charmbracelet/lipgloss`; respect `NO_COLOR` env var
- [ ] 1.1.3 Create `internal/iostreams/iostreams_test.go` (table-driven) covering `System()` TTY detection, `Test()` buffer-backed mode, `Verbosef` formatting, `NO_COLOR` respect
- [ ] 1.1.4 Create `internal/iostreams/styles_test.go` (table-driven) covering each `Styles` method's output in TTY vs non-TTY mode
- [ ] 1.1.5 Add `github.com/charmbracelet/lipgloss` to `go.mod` (`go get github.com/charmbracelet/lipgloss@latest`)
- [ ] 1.1.6 Add `golang.org/x/term@latest` to `go.mod` and add the direct import
- [ ] 1.1.7 Add direct import of `github.com/spf13/pflag` to `go.mod` (run `go mod tidy`); required by `cmdutil.ExitCodeFor` for typed `*pflag.NotExistError` / `*pflag.ValueRequiredError` / `*pflag.InvalidValueError` / `*pflag.InvalidSyntaxError` detection (these typed errors all exist in pflag v1.0.10; see `design.md` Decision 5 for the Cobra string-prefix fallback)
- [ ] 1.1.8 Create `internal/output/errors.go` with `FormatError(io *iostreams.IOStreams, err error)` and the `errors.As` switch for `core.ParseError`, `core.ValidationError`, `core.TransformError`, `core.FileError`, `core.ConfigError`, `core.PartialSuccessError`
- [ ] 1.1.9 Create `internal/output/exporter.go` with `Exporter` interface (`Write(io *iostreams.IOStreams, data any) error`), `JSONExporter` (uses `encoding/json`, 2-space indent, trailing newline), `TableExporter` (uses `text/tabwriter`)
- [ ] 1.1.10 Create `internal/output/output_flags.go` with `AddOutputFlags(cmd *cobra.Command, output *string)` (default `"plain"`; valid values hardcoded as `json`, `table`, `plain`; completion wired for those three values via `cobra.RegisterFlagCompletionFunc`)
- [ ] 1.1.11 Create `internal/output/output_test.go` (table-driven) covering `FormatError` dispatch for each typed error, both exporters (`JSONExporter` round-trips through `encoding/json`; `TableExporter` aligns columns)
- [ ] 1.1.12 Create `internal/cmdutil/factory.go` with `Factory` struct fields: `IOStreams *iostreams.IOStreams` (eager), `AppVersion string` (eager), `Executable string` (eager), `RootContext context.Context` (eager), plus lazy `func() (T, error)` fields: `Config`, `Library`, `Transformer`, `Validator`, `Canonicalizer`, `Initializer`. Each lazy field uses `sync.OnceValues` (Go 1.21+) for caching
- [ ] 1.1.13 Create `internal/cmdutil/exit.go` with `ExitCode` type (`int`), `ExitCodeSuccess = 0`, `ExitCodeError = 1`, `ExitCodeUsage = 2` constants, and `ExitCodeFor(err error) ExitCode` (usage detection via `errors.As` on `*pflag.NotExistError` / `*pflag.ValueRequiredError` / `*pflag.InvalidValueError` / `*pflag.InvalidSyntaxError` typed errors, plus a `strings.HasPrefix(err.Error(), "unknown flag")` / `"flag needs an argument"` / `"invalid argument"` fallback for Cobra's own usage errors that pflag doesn't wrap; returns 0 for `*core.PartialSuccessError` when `Succeeded > 0`, 1 otherwise)
- [ ] 1.1.14 Create `internal/cmdutil/output_flags.go` re-exporting `output.AddOutputFlags` as `cmdutil.AddOutputFlags` so command files import only `cmdutil`
- [ ] 1.1.15 Create `internal/cmdutil/factory_test.go` covering: (a) lazy-field caching (second call returns same instance), (b) two callers in one command invocation share the cached `*library.Library`, (c) concurrent first call invokes the function exactly once (use `-race`), (d) a transient factory error is cached and re-returned on subsequent calls, (e) cross-dependency caching (`Initializer` chains through `Library` and `Transformer`, which chain through `Config`; counter asserts `f.Config()` is invoked exactly once)
- [ ] 1.1.16 Create `internal/cmdutil/exit_test.go` (table-driven) covering: `nil → 0`, Cobra usage errors (string-prefix match) → 2, `*pflag.NotExistError` → 2, `*pflag.ValueRequiredError` → 2, `*pflag.InvalidValueError` → 2, `*pflag.InvalidSyntaxError` → 2, `*core.ValidationError` → 1, `*core.ParseError` → 1, `*core.TransformError` → 1, `*core.FileError` → 1, `*core.ConfigError` → 1, generic errors → 1, `*core.PartialSuccessError` with `Succeeded > 0 → 0`, `*core.PartialSuccessError` with `Succeeded == 0 → 1`

## 1.2 Rename `internal/domain/` to `internal/core/`

- [ ] 1.2.1 Move all files from `internal/domain/` to `internal/core/` (`agent.go`, `command.go`, `skill.go`, `memory.go`, `platform.go`, `errors.go`, `validation.go`, `pipeline.go`, `result.go`, `results.go`, `doc.go`, plus the `opencode/` sub-directory)
- [ ] 1.2.2 Update package declarations from `package domain` to `package core` in every moved file
- [ ] 1.2.3 Update the depguard rule in `.golangci.yml`: change `files: ["**/domain/**"]` to `files: ["**/core/**"]`; rename the rule from `domain` to `core-isolation`; allow `$gostd, github.com/samber/lo`; deny all `github.com/*` packages. Also update `linters.settings.wrapcheck.ignorePackageSig` from `internal/domain` to `internal/core`
- [ ] 1.2.4 Update every import of `gitlab.com/amoconst/germinator/internal/domain` to `gitlab.com/amoconst/germinator/internal/core` across the whole tree; run `rg 'internal/domain' .` to verify zero matches outside `openspec/` history (final group check happens in task 1.3.7)
- [ ] 1.2.5 Add `type Domain = core` (Go type alias) in `internal/core/doc.go` as a temporary compatibility shim for any external consumer; remove in change-9
- [ ] 1.2.6 Run `mise run check`; confirm zero issues and no behavior change

## 1.3 Flatten `internal/infrastructure/`

- [ ] 1.3.1 Move `internal/infrastructure/parsing/*.go` to `internal/parser/*.go`; update package decl to `package parser`; update all imports
- [ ] 1.3.2 Move `internal/infrastructure/serialization/*.go` to `internal/renderer/*.go`; update package decl to `package renderer`; update all imports
- [ ] 1.3.3 Move `internal/infrastructure/adapters/claude-code/*.go` to `internal/claude-code/*.go`; update package decl; update all imports
- [ ] 1.3.4 Move `internal/infrastructure/adapters/opencode/*.go` to `internal/opencode/*.go`; update package decl; update all imports
- [ ] 1.3.5 Move `internal/infrastructure/config/*.go` to `internal/config/*.go`; update package decl; update all imports
- [ ] 1.3.6 Move `internal/infrastructure/library/*.go` to `internal/library/*.go`; update package decl; update all imports
- [ ] 1.3.7 Run `mise run check` after task 1.3.6; then run `find internal/infrastructure -type f` to verify zero files remain
- [ ] 1.3.8 Remove the now-empty `internal/infrastructure/` directory tree (`rm -rf internal/infrastructure/`)

## 1.4 Add `internal/core/rules.go`

- [ ] 1.4.1 Create `internal/core/rules.go` with `ValidatePlatform(s string) error` (returns `*core.ValidationError` if `s` is not `"claude-code"` or `"opencode"`) and `ResolveOutputPath(docType, name, platform string) string` (pure function combining the three into the canonical output filename)
- [ ] 1.4.2 Create `internal/core/rules_test.go` (table-driven) covering valid + invalid platform strings, output path resolution for each `(docType, name, platform)` triple
- [ ] 1.4.3 Confirm `internal/core/rules.go` imports only stdlib (no `internal/library/` or any I/O package) — depguard enforces this

## 1.5 Add `core.InitializeError` and `core.PartialSuccessError`

- [ ] 1.5.1 In `internal/core/errors.go`, add `InitializeError` and `PartialSuccessError`, both following the existing builder pattern (lowercase fields, `WithSuggestions`/`WithContext` immutable builders, `Error()`/`Unwrap()` methods, getter methods, constructor `NewXxxError`):

  ```go
  type InitializeError struct {
      ref         string
      inputPath   string
      outputPath  string
      cause       error
      suggestions []string
      context     string
  }

  func NewInitializeError(ref, inputPath, outputPath string, cause error) *InitializeError
  func (e *InitializeError) WithSuggestions(suggestions []string) *InitializeError
  func (e *InitializeError) WithContext(context string) *InitializeError
  func (e *InitializeError) Ref() string
  func (e *InitializeError) InputPath() string
  func (e *InitializeError) OutputPath() string
  func (e *InitializeError) Cause() error
  func (e *InitializeError) Suggestions() []string
  func (e *InitializeError) Context() string
  func (e *InitializeError) Error() string
  func (e *InitializeError) Unwrap() error

  type PartialSuccessError struct {
      succeeded int
      failed    int
      errors    []InitializeError
  }

  func NewPartialSuccessError(succeeded, failed int, errs []InitializeError) *PartialSuccessError
  func (e *PartialSuccessError) Succeeded() int
  func (e *PartialSuccessError) Failed() int
  func (e *PartialSuccessError) Errors() []InitializeError
  func (e *PartialSuccessError) Error() string
  ```

- [ ] 1.5.2 Add unit tests in `internal/core/errors_test.go` covering: `InitializeError.Error()` format, `InitializeError.Unwrap()` returns the cause, `WithSuggestions`/`WithContext` return new instances (immutability), `NewPartialSuccessError` constructor sets fields correctly, `errors.As(err, &ps)` works, `errors.As(err, &ie)` works
- [ ] 1.5.3 Confirm `core.PartialSuccessError` is recognized by `cmdutil.ExitCodeFor` and `output.FormatError` (test coverage in tasks 1.1.15 and 1.1.11 respectively)
- [ ] 1.5.4 Add a cross-package integration test in `internal/output/output_test.go` (or a new `internal/cmdutil/integration_test.go`) that constructs a `*core.PartialSuccessError{Succeeded: 3, Failed: 1, ...}`, asserts `cmdutil.ExitCodeFor(err) == ExitCodeSuccess` (per Decision 12), and asserts `output.FormatError(io, err)` writes the expected partial-success string to `io.ErrOut`. This guards against the three packages drifting out of sync.

## 1.6 Enable `forbidigo` lint rules

- [ ] 1.6.1 In `.golangci.yml`, enable `forbidigo` linter with the following patterns:
  - `fmt\.Fprintf\(os\.(Stdout|Stderr)` in `cmd/*.go` excluding `cmd/**/*_test.go`
  - `os\.Exit\(` in `cmd/**` excluding `main.go`
  - `var global(Factory|CommandConfig)` in `cmd/**`
  - `SetGlobal(Factory|CommandConfig)` in `cmd/**`
  - `context\.Background\(\)` in `cmd/**/*.go` (except `main.go`)
- [ ] 1.6.2 Enable `nolintlint` linter with `require-explanation: true, require-specific: true`
- [ ] 1.6.3 Run `rg 'nolint:' cmd/ internal/ --type go` and verify every existing `//nolint:` directive has both an explanation and a specific linter name; add where missing (otherwise `nolintlint` will fail the suite on day one)
- [ ] 1.6.4 Run `mise run lint` to verify the rules apply; existing `cmd/*.go` may have violations that are deferred to their respective command-migration changes
- [ ] 1.6.5 Create `cmd/testdata/` directory and `cmd/testdata/lint_baseline.txt` containing the initial `mise run lint` output (capture both streams: `mise run lint > cmd/testdata/lint_baseline.txt 2>&1`), then commit the file. Create `cmd/lint_test.go` that runs `exec.Command("mise", "run", "lint")` and asserts no NEW violations exist beyond the baseline. The test diffs current output against the baseline and fails on any non-baseline entry
- [ ] 1.6.6 Confirm `mise run lint` is green overall

## 1.7 Update spec delta files (this change's slice)

> **Note on paths:** delta specs live in flat folders (`specs/<name>/spec.md`), as required by the `openspec` CLI's delta discovery. The category-based layout (`specs/<category>/<name>/spec.md`) is a *sync-time* concern handled when deltas land in `openspec/specs/` — see the proposal's "Capabilities" note for the capability→category mapping. Tasks below reference the on-disk flat paths.

- [ ] 1.7.1 Update `openspec/changes/scaffold-cli-foundation/specs/dependency-injection/spec.md` to add a `## REMOVED Requirements` section for `ServiceContainer` with `**Reason:**` and `**Migration:**` fields; reword the existing MODIFIED requirement text to describe only the new Factory-based DI mechanism
- [ ] 1.7.2 Update `openspec/changes/scaffold-cli-foundation/specs/exit-codes/spec.md` to add `**Reason:**` and `**Migration:**` fields to the existing `## REMOVED Requirements` block for `CategorizeError`; mark the 0/1/2 collapse as **in-progress** (the `cmdutil.ExitCodeFor` exists with tests, but `main.go` doesn't use it yet — that's change-2)
- [ ] 1.7.3 Update `openspec/changes/scaffold-cli-foundation/specs/framework/spec.md` to add a `## REMOVED Requirements` section for `CommandConfig` with `**Reason:**` and `**Migration:**` fields; reword the existing MODIFIED requirement text to describe only the Factory-based constructor signature
- [ ] 1.7.4 Update `openspec/changes/scaffold-cli-foundation/specs/verbose-output/spec.md` to add a `## REMOVED Requirements` section for `Verbosity` and `VerbosePrint`/`VeryVerbosePrint` with `**Reason:**` and `**Migration:**` fields; reword the existing MODIFIED requirement text to describe only the new `opts.IO.Verbosef` mechanism
- [ ] 1.7.5 Update `openspec/changes/scaffold-cli-foundation/specs/error-formatting/spec.md` to add a `## REMOVED Requirements` section for `ErrorFormatter` with `**Reason:**` and `**Migration:**` fields; reword the existing MODIFIED requirement text to describe only the new `output.FormatError` dispatch

## 1.8 Verification

- [ ] 1.8.1 Run `mise run lint` — confirm no NEW violations introduced by this change
- [ ] 1.8.2 Run `mise run test` — confirm all unit tests pass (including new tests for `iostreams/`, `output/`, `cmdutil/`, `core/rules.go`, `core/errors.go`)
- [ ] 1.8.3 Run `mise run build` — confirm `bin/germinator` builds
- [ ] 1.8.4 Run `mise run test:coverage` — confirm `internal/iostreams/`, `internal/output/`, `internal/cmdutil/`, `internal/core/` coverage ≥ 70%
- [ ] 1.8.5 Run `mise run check` — full validation passes
- [ ] 1.8.6 Smoke-test every existing command: `germinator --help`, `germinator adapt --help`, `germinator validate --help`, `germinator canonicalize --help`, `germinator init --help`, `germinator library --help`, `germinator config --help`, `germinator version`, `germinator completion bash | head -5` — confirm byte-identical output to pre-change behavior
- [ ] 1.8.7 Run `openspec validate scaffold-cli-foundation --strict` and confirm all specs and tasks are coherent

> **Note on AGENTS.md:** per project convention (see root `AGENTS.md` documentation workflow and the `osx-maintain-ai-docs` skill), AGENTS.md updates are NOT tracked as numbered tasks in this checklist. They are handled by the documentation-maintenance phase between verify-change and archive-change. Documentation work for this change includes:
>
> - updating `internal/AGENTS.md` to reflect the `domain → core` rename and the flattened `infrastructure/` packages
> - adding per-package `AGENTS.md` files for the new packages (`internal/iostreams/AGENTS.md`, `internal/output/AGENTS.md`, `internal/cmdutil/AGENTS.md`)
> - updating `cmd/AGENTS.md` to document the new `forbidigo` patterns and the `cmd/lint_test.go` enforcement test
