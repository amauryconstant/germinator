# Tasks — Enforce error-handling discipline

Tasks are ordered so each commit is independently testable. Every task ends with `mise run build && mise run lint && mise run test` passing before the next starts. File:line references in this document were verified against the tree at task `0.1`; deviations are corrected by that task.

> **Note**: Throughout this file, `gerrors` is the project alias for `internal/core/errors.go` (the `germinator/errors` import path); `core.*Error` in the prose refers to the same types via the `core` import in cmd/* code. Both names refer to the same set of typed errors (9 existing + 2 new after Phase 1).

## 0. Phase 0 — Cross-change reconciliation

- [ ] 0.1 Re-verify every file:line in this `tasks.md` against the current tree. For each site, run the listed `rg` and `sed -n` checks; record discrepancies in this commit's body and update the affected task descriptions before proceeding. **The rebase extends beyond `tasks.md`** — update `proposal.md` Affected-code table and `specs/*/spec.md` `When … internal/library/adder.go:nnn` style site references to match the verified line numbers; the verification commit body MUST record each original-to-current line shift as a regression trace. Specifically:
  - `rg "output\.FormatError" cmd/ main.go` — must show 10 production calls in `cmd/` (deleted in Phase 2) + 2 production calls in `main.go` (lines 32, 46, retained) + 2 test calls in `cmd/` (retained) = 14 production/test hits before Phase 2; after Phase 2 only the 4 retained production calls + 2 test calls + comment references remain.
  - `rg "errEmptyResources" cmd/` — must show a single declaration at `cmd/library_create.go:nn` and a single call at `:nn`.
  - `rg --line-number 'BatchFailureInfo\{' internal/library/` — must show 5 population sites.
  - `rg --line-number 'NewFileError.*not found' internal/library/` — must show exactly 5 sites (resolver.go:21,26; loader.go:36,53; adder.go:nnn) before migration; 0 after.
  - `rg --line-number 'NewConfigError.*preset not found' internal/library/` — must show exactly 1 site (resolver.go:nnn) before migration; 0 after.
  - `rg --line-number "os\.IsNotExist" internal/library/remover.go` — must show the silent-swallow site at line `nnn`.
- [ ] 0.1a Verify `go.mod` Go version is `>= 1.21` (for `slices` stdlib import added in task `1.7`). If the project minimum is `< 1.21`, switch the defensive copies in `UsageError.Suggestions()` and `UsageError.WithSuggestions(s)` to a manual `append([]string(nil), s...)` expression or `cloneSlice(s) = append(make([]string, 0, len(s)), s...)`. Record the actual version in the commit body for Phase 1 traceability.
- [ ] 0.1b Audit `cmd/*.go` lookup chains for direct `lib.Resources[type][name]` or `lib.Presets[name]` dereferences. Run `rg "lib\.Resources\[|lib\.Presets\[" cmd/` and report every production-file match (test-file matches are expected and OK because they assert fixture state, not implement production lookups). For each production match, add a Phase 3 sub-task that migrates the lookup to `internal/library/resolver.go` and surfaces `*core.NotFoundError` from the resolver instead of the cmd layer. Known suspected matches (verified): `cmd/show.go:138,142,172`. Confirmed-but-already-correct matches (do NOT migrate — already use `*core.NotFoundError`): `cmd/library_remove.go:287,292,296,363`.

## 1. Phase 1 — Exit-code semantics + dispatch set + canary exemption

- [ ] 1.1 In `internal/cmdutil/exit.go:73`, change `*core.NotFoundError` mapping from `ExitCodeUsage` (2) to `ExitCodeError` (1).
- [ ] 1.2 In `internal/cmdutil/exit.go:24-37`, delete the `cobraUsagePrefixes` slice (12 entries) and the `hasCobraUsagePrefix` function. The existing typed branch (`errors.As(err, &pflag.NotExistError)` etc. at lines 64-69) handles the four pflag-emitted types directly. Remove the corresponding call to `hasCobraUsagePrefix` at line 70. **Fallback**: if code review rejects the `*core.CobraUsageError` introduction in task 1.3, fall back to keeping a 4-prefix substring list — `requires at least`, `requires exactly`, `accepts at most`, `required flag(s)` — and skip tasks 1.3 and 3.14. The fallback's test rows replace the `*core.CobraUsageError` row in task 1.4 with four substring cases.
- [ ] 1.3 In `internal/cmdutil/exit.go`, add a new typed branch for `*core.CobraUsageError`: `if errors.As(err, &cobraUsage) { return ExitCodeUsage }`. Update the function comment to enumerate the new dispatch set.
- [ ] 1.4 In `internal/cmdutil/exit_test.go:58`, flip the `*core.NotFoundError` expectation from `ExitCodeUsage` to `ExitCodeError`. **Delete the 5 cobra-substring rows at lines 45-49** (these test the substring-fallback dispatch that task 1.2 removes; they are dead code after Phase 1). Add table rows for `*core.UsageError` (expects `ExitCodeUsage`) and `*core.CobraUsageError` (expects `ExitCodeUsage`). Update the package comment if it still references the substring fallback.
- [ ] 1.5 In `internal/output/errors.go:21-50`, add `case *core.InitializeError` rendering `"Error: " + e.Error() + "\n"` (delegates to `InitializeError.Error()` which colon-joins `<ref>: output: <outputPath>: <cause>` segments per `internal/core/errors.go`). Add `case *core.UsageError` rendering `"Error: " + e.Flag() + ": " + e.Reason() + "\n"`. Reuse the existing `writeErrOut` helper.
- [ ] 1.5b In `internal/output/errors.go`, add `case *config.WriteError` rendering `"Error: " + e.Op() + " " + e.Path() + ": " + e.Error() + "\n"` (reuses `writeErrOut`). The new case sits after `UsageError` and before the generic-error fallback; adds `internal/config` to the file imports if not present. Add an explicit `if errors.As(err, &writeErr) { return ExitCodeError }` branch in `internal/cmdutil/exit.go` between the `PartialSuccess` branch and the default. Add `internal/config` to the file imports. The explicit row matches the spec's ADDED Requirement at `cli-exit-codes/spec.md:91-100` (which mandates `*config.WriteError → ExitCodeError (1)` as a SHALL contract); the explicit row is preferred over relying on default-fallthrough — a future default-branch change (e.g., adding `Logger.Error`) cannot silently break the contract. Add a new test row to `internal/cmdutil/exit_test.go`: `name: "config WriteError", err: &config.WriteError{...}, want: ExitCodeError`.
- [ ] 1.6 In `internal/output/exporter.go`, delete the `var _ = io.EOF` line and the `io` import on line 7 (verify zero other uses via `rg "io\." internal/output/exporter.go`).
- [ ] 1.7 In `internal/core/errors.go`, add the `UsageError` and `CobraUsageError` types following the existing builder pattern:
  ```go
  type UsageError struct {
      flag, reason string
      suggestions  []string
  }
  func NewUsageError(flag, reason string) *UsageError { ... }
  func (e *UsageError) Flag() string             { return e.flag }
  func (e *UsageError) Reason() string           { return e.reason }
  // Suggestions returns a defensive copy via slices.Clone (Go 1.21+) so
  // callers cannot mutate the receiver's internal slice.
  func (e *UsageError) Suggestions() []string     { return slices.Clone(e.suggestions) }
  // WithSuggestions returns a NEW *UsageError with the same flag/reason and
  // a freshly-allocated suggestions slice (immutable builder). The input
  // slice is defensive-copied to prevent caller mutation.
  func (e *UsageError) WithSuggestions(s []string) *UsageError {
      return &UsageError{flag: e.flag, reason: e.reason, suggestions: slices.Clone(s)}
  }
  func (e *UsageError) Error() string  { return e.flag + ": " + e.reason }
  func (e *UsageError) Unwrap() error  { return nil } // UsageError is a leaf error; godoc must note this

  type CobraUsageError struct{ err error }
  func MustNewCobraUsageError(err error) *CobraUsageError {
      if err == nil { panic("MustNewCobraUsageError: cause is required (programmer error)") }
      return &CobraUsageError{err: err}
  }
  func (e *CobraUsageError) Error() string { return e.err.Error() }
  func (e *CobraUsageError) Unwrap() error { return e.err }
  ```
  Then add `MarshalJSON()` to all 11 typed errors (the 9 existing plus the 2 new). Each returns `{"error": "<Error()>"}`:
  ```go
  func (e *ParseError) MarshalJSON() ([]byte, error)         { return json.Marshal(struct{ Error string }{Error: e.Error()}) }
  func (e *ValidationError) MarshalJSON() ([]byte, error)    { ... }
  // ... repeat for all 11 types
  ```
  Add `encoding/json` and `slices` (Go 1.21+) to the file imports. The godoc for `*UsageError` MUST state: "the `flag` and `reason` parameters MUST be lowercase with no trailing punctuation (Go error-string convention per `golang-error-handling` rule 3 — `references/error-creation.md:32`)".

  Add compile-time interface checks near the type declarations to catch method-signature typos at build time (per `golang-design-patterns` rule 19):
  ```go
  // 11 error interface checks (every typed error satisfies `error`).
  var (
      _ error = (*ParseError)(nil)
      _ error = (*ValidationError)(nil)
      _ error = (*TransformError)(nil)
      _ error = (*FileError)(nil)
      _ error = (*ConfigError)(nil)
      _ error = (*NotFoundError)(nil)
      _ error = (*OperationError)(nil)
      _ error = (*InitializeError)(nil)
      _ error = (*PartialSuccessError)(nil)
      _ error = (*UsageError)(nil)
      _ error = (*CobraUsageError)(nil)
  )

  // 11 json.Marshaler interface checks (every typed error satisfies MarshalJSON).
  var (
      _ json.Marshaler = (*ParseError)(nil)
      _ json.Marshaler = (*ValidationError)(nil)
      _ json.Marshaler = (*TransformError)(nil)
      _ json.Marshaler = (*FileError)(nil)
      _ json.Marshaler = (*ConfigError)(nil)
      _ json.Marshaler = (*NotFoundError)(nil)
      _ json.Marshaler = (*OperationError)(nil)
      _ json.Marshaler = (*InitializeError)(nil)
      _ json.Marshaler = (*PartialSuccessError)(nil)
      _ json.Marshaler = (*UsageError)(nil)
      _ json.Marshaler = (*CobraUsageError)(nil)
  )
  ```
  **Note**: any future exported struct fields on `core.*Error` types must be exposed via `MarshalJSON` to appear in JSON output — `json.Marshaler` precedence in stdlib means `MarshalJSON` wins over struct-field marshaling.
- [ ] 1.7a Remove the `internal/warning` canary. Delete `internal/warning/canary.go` (the `MaybeWarnLegacyExitCode` function and `canaryOnce` var), `internal/warning/canary_test.go`, and `internal/warning/AGENTS.md`. If `internal/warning/` is now empty, delete the directory. The exit-code 5 → 1 migration the canary was added to support is now complete; without removal, the new `NotFoundError → 1` mapping in task 1.1 would cause the canary to fire on every interactive lookup miss.
- [ ] 1.8 Run `mise run test`, `mise run lint`, `mise run build` — all clean. (Tasks 1.8 and 1.9 from the original proposal are removed: the `internal/warning` canary is deleted entirely in task 1.7a, not widened with an exemption.)

## 2. Phase 2 — Single-handling rule cleanup

- [ ] 2.1 In `cmd/library_add.go:355`, delete the inline `output.FormatError(opts.IO, opErr)` call. `cmd/library_add.go` already has a per-file `output.FormatError` chain in the run loop; the delete is a no-op for behavior, only removes the duplicate. Verify the line above still wraps the error in `core.NewOperationError` and threads it into `initErrs`.
- [ ] 2.2 In `cmd/library_add.go:549`, delete the inline `output.FormatError(opts.IO, opErr)` call AND change the line above from `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` to `opErr := core.NewOperationError("add", f.Source, nil)` with `opErr.Cause` lifted from the typed cause (the typed error must come from `processBatchAddFile`, not from `f.Error` as a string). Adjust `adder.go` if needed to surface typed errors per task `3.11`.
- [ ] 2.3 In `cmd/library_add.go:693`, delete the inline `output.FormatError(opts.IO, opErr)` call. Verify the surrounding `if isPlainOutput(opts.Output)` guard remains intact.
- [ ] 2.4 In `cmd/library_add.go:706`, delete the inline `output.FormatError(opts.IO, opErr)` call. Same wrap-and-lift pattern as 2.2.
- [ ] 2.5 In `cmd/library_validate.go:182`, delete the inline `output.FormatError(opts.IO, err)` call. The wrapping `fmt.Errorf("loading library: %w", err)` at line 183 must remain.
- [ ] 2.6 In `cmd/library_validate.go:190`, delete the inline `output.FormatError(opts.IO, err)` call. The wrapping `fmt.Errorf("validating library: %w", err)` at line 191 must remain.
- [ ] 2.7 In `cmd/validate.go:123`, delete the inline `output.FormatError(opts.IO, e)` from inside the `for _, e := range result.Errors` loop. The terminal `return result.Errors[0]` at line 125 must remain so `ExitCodeFor` still maps the first validation error correctly.
- [ ] 2.8 In `cmd/init.go:218`, delete the inline `output.FormatError(opts.IO, partialErr)` call from the all-failure branch.
- [ ] 2.9 In `cmd/init.go:222`, delete the inline `output.FormatError(opts.IO, partialErr)` call from the partial-success branch.
- [ ] 2.10 In `cmd/init.go:258`, delete the inline `output.FormatError(opts.IO, core.NewInitializeError(r.Ref, r.InputPath, r.OutputPath, r.Error))` from `renderResults`. The terminal `fmt.Fprintf(opts.IO.Out, "Initialized %d resource(s).", s)` at lines 271-275 must remain.
- [ ] 2.11 Run `rg "output\.FormatError" cmd/ main.go` — must return the 2 retained production calls (`main.go:32` factory-build handler, `main.go:46` post-Execute handler) and 2 retained test-file calls (`cmd/init_test.go:185`, `cmd/show_test.go:239`), plus comment-only references in `cmd/AGENTS.md`, `cmd/commands/AGENTS.md`, and various `cmd/*_test.go` files. Any additional production CALL sites in `cmd/` are missed deletes from tasks `2.1-2.10`.
- [ ] 2.12 Run `mise run test`, `mise run test:e2e` — no double-output in any captured stderr.

## 3. Phase 3 — Type migration + chain preservation

- [ ] 3.1 In `internal/library/resolver.go:21`, replace `gerrors.NewFileError(ref, "resolve", "resource not found", nil)` with `gerrors.NewNotFoundError("resource", ref)`.
- [ ] 3.2 In `internal/library/resolver.go:26`, same replacement.
- [ ] 3.3 In `internal/library/resolver.go:70`, replace `gerrors.NewConfigError("preset", name, "preset not found")` with `gerrors.NewNotFoundError("preset", name)`.
- [ ] 3.4 In `internal/library/loader.go:36`, replace `gerrors.NewFileError(path, "access", "library not found", nil)` with `gerrors.NewNotFoundError("library", path)`.
- [ ] 3.5 In `internal/library/loader.go:53`, replace `gerrors.NewFileError(yamlPath, "read", "library.yaml not found", nil)` with `gerrors.NewNotFoundError("library.yaml", yamlPath)`.
- [ ] 3.6 In `internal/library/adder.go:146`, replace `gerrors.NewFileError(source, "access", "source file not found", nil)` with `gerrors.NewNotFoundError("source file", source)`.
- [ ] 3.7 In `internal/library/remover.go:83`, replace `gerrors.NewFileError(opts.LibraryPath, "access", fmt.Sprintf("resource %s not found", opts.Ref), nil)` with `gerrors.NewNotFoundError("library ref", opts.Ref)`.
- [ ] 3.8 In `internal/library/remover.go:88`, same pattern (second occurrence of the `NewFileError` lookup).
- [ ] 3.9 In `internal/library/remover.go:142`, replace `gerrors.NewFileError(opts.LibraryPath, "access", fmt.Sprintf("preset %s not found", opts.Name), nil)` with `gerrors.NewNotFoundError("preset", opts.Name)`.
- [ ] 3.10 In `internal/library/remover.go:104`, replace the silent `os.IsNotExist` swallow: `if errors.Is(err, os.ErrNotExist) { return nil, core.NewNotFoundError("library file", path) }`. Add the `os` and `errors` imports if missing.
- [ ] 3.11 In `internal/library/adder.go:526-529`, extend `BatchFailureInfo`:
  ```go
  type BatchFailureInfo struct {
      Source    string `json:"source"`
      Error     string `json:"error"`
      ErrorType string `json:"errorType,omitempty"`
      Cause     error   `json:"cause,omitempty"`
  }
  ```
  Update every `BatchFailureInfo{...}` literal at `adder.go:647-651, 664-668, 682-686, 700-704, 764-768` to populate `ErrorType` and `Cause`. Compute `ErrorType` via a typed switch in a small helper in `adder.go`:
  ```go
  func errorTypeName(cause error) string {
      switch cause.(type) {
      case nil:
          return ""
      case *gerrors.NotFoundError:
          return "NotFoundError"
      case *gerrors.FileError:
          return "FileError"
      case *gerrors.ValidationError:
          return "ValidationError"
      case *gerrors.ParseError:
          return "ParseError"
      case *gerrors.ConfigError:
          return "ConfigError"
      case *gerrors.OperationError:
          return "OperationError"
      case *gerrors.InitializeError:
          return "InitializeError"
      case *gerrors.PartialSuccessError:
          return "PartialSuccessError"
      case *os.PathError:
          return "PathError"
      default:
          return fmt.Sprintf("%T", cause)
      }
  }
  ```
  Plumb the typed cause through from `processBatchAddFile`'s callers so `Cause` is set even when `Error` is the stringified fallback. Population sites MUST convert non-typed causes (e.g., `*os.PathError` returned from a filesystem call) to `*core.FileError` before assigning to `f.Cause`; the stdlib default for non-typed errors is `{}` per `json.Marshaler` precedence rules, which defeats the typed-error-chain preservation contract documented in `errors-enhanced-errors/spec.md` at the "Cause MUST be a typed error" scenario.
- [ ] 3.12 In `cmd/library_create.go:70`, migrate `errEmptyResources = errors.New("flag needs an argument: --resources (must be non-empty list of refs)")` to `errEmptyResources = core.NewUsageError("--resources", "must be non-empty list of refs")`. The single call site at `cmd/library_create.go:203` returns `errEmptyResources` directly — verify the new `*core.UsageError` flows through `ExitCodeFor` to `ExitCodeUsage (2)` in `cmd/library_create_test.go:152-176`.
- [ ] 3.13 Update the following test files to assert `*core.NotFoundError` → `ExitCodeError (1)`:
  - `internal/library/resolver_test.go:174-177` — currently `errors.As(&cfgErr)` against `*gerrors.ConfigError`; change to `errors.As(&nf)` against `*gerrors.NotFoundError`.
  - `cmd/library_create_test.go:343-365` (T11 `TestRunCreatePreset_RefReferencesMissingResource`; the `ExitCodeUsage` assertion is at line 364) — same swap.
  - `cmd/show_test.go:151,179` — same.
  - `cmd/library_remove_test.go:397` — same.
  - `test/e2e/init_test.go:108` — `Describe("init fails for nonexistent preset", It("should fail with exit code 2 for invalid preset name (NotFoundError → ExitCodeUsage)"))` — rename the `It` description to `should fail with exit code 1 for invalid preset name (NotFoundError → ExitCodeError)` and update the inline `ShouldFailWithExit(session, 2)` to `ShouldFailWithExit(session, 1)`. The slice-5 §5.0.1 inline comment (lines 113-116) must also be updated to reflect the corrected exit-code mapping.
  - `internal/cmdutil/exit_test.go:58` — already updated in task `1.4`, but verify no other row still expects `ExitCodeUsage` for `NotFoundError`.
- [ ] 3.14 Verify no Cobra arg-validation errors need re-wrapping today. `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` validators and `MarkFlagRequired` errors flow through `cmd.Execute()` to `main.go:46` directly without touching `RunE`; `*core.CobraUsageError` has zero current call sites and is reserved for future typed dispatch (e.g., custom command code that needs typed exit-code mapping). The `ExitCodeFor` dispatch contract for `*core.CobraUsageError` is in place per task `1.3`; no code-side changes are required for the existing test suite. Add a unit test in `internal/core/errors_test.go` asserting `MustNewCobraUsageError(nil)` panics (programmer-error guard), `MustNewCobraUsageError(errors.New("x")).Error() == "x"`, and `errors.Unwrap(MustNewCobraUsageError(errors.New("x"))) == errors.New("x")`.
- [ ] 3.15 Verification regex — `rg "NewFileError\([^)]*not found|NewConfigError\("preset"[^)]*preset not found" internal/library/` must return zero matches. Run `mise run check` before task `3.16`.
- [ ] 3.16 Run `mise run test`, `mise run test:e2e`, `mise run test:coverage`.
- [ ] 3.17 In `cmd/init.go:188-191`, drop the redundant `NewNotFoundError("preset", opts.Preset)` re-wrap. After task `3.3` lands, `lib.ResolvePreset` returns `*core.NotFoundError` directly; simplify to `if rerr != nil { return rerr }` (errors.As still resolves to `*core.NotFoundError` for exit-code mapping). Add a regression test asserting `cmd/init.go`'s not-found-preset error path returns `*core.NotFoundError` and `ExitCodeFor` returns `ExitCodeError` (1).
- [ ] 3.18 Update `cmd/init_test.go` assertions to match the new `*core.UsageError` constructor (from tasks 1.7 and 3.12). Audit and rewrite any test that asserts against the prior `errEmptyResources` cobra-substring wording. Verification: `rg "flag needs an argument" cmd/` must return zero matches after Phase 3 lands.
- [ ] 3.19 Drop the redundant `fmt.Errorf("libraryAdapter.AddResource: %w", err)` wrapchain noise identified in the 2026-07-08 review (B-010). When the inner error is already a `*core.*Error` typed error, the wrap prefix is meaningless — `errors.As` traverses the typed-error chain without it. Locate the production site, replace `fmt.Errorf("libraryAdapter.AddResource: %w", err)` with the direct `err` return, and update affected tests. Verification: `rg "libraryAdapter\.AddResource.*%w" internal/library/` must return zero matches after this task lands.
- [ ] 3.20 *(conditional — add only if Phase 0 task `0.1b` confirms outstanding cmd-side lookups)* Migrate `cmd/show.go:renderResource` (lookups at lines 138, 142) and `cmd/show.go:renderPreset` (lookup at line 172) to call new `internal/library/resolver.go` helpers `ResolveResource(lib, ref) (*library.Resource, error)` and `ResolvePreset(lib, name) (*library.Preset, error)` that return `*core.NotFoundError` on miss. Update `cmd/show_test.go:151,179` (already uses `*core.NotFoundError` assertions; verify no wording change needed). Do NOT introduce new lookup logic at the cmd layer; encapsulate in the resolver per the design's "Migrate lookup branches to internal/library" Direction (Decision #4). Confirmed-out-of-scope matches (already migrated in slice 7.3 per `cmd/library_remove.go` AGENTS.md): `cmd/library_remove.go:287,292,296,363`. If `0.1b` reports no other outstanding cmd-side lookups beyond `cmd/show.go`, add this single task; if `0.1b` reports additional sites, add `3.21`, `3.22`, etc., one per site.

## 4. Phase 4 — Trivial folds + cross-change imports

- [ ] 4.1 Create `internal/config/scaffold.go` with `func WriteDefault(path string, force bool) error`. Implementation: `if !force { os.Stat } else { skip }` → `os.MkdirAll(dir, 0o750)` → `os.WriteFile(path, []byte(defaultTOML), 0o600)`. Return `*config.WriteError` (new domain type in `internal/config/errors.go`) on I/O failure — NOT `*core.FileError`, because `internal/config/` is the Imperative Shell layer that owns its own I/O errors per `golang-cli-architecture/references/05-errors.md` ("errors that originate from external I/O are defined near their origin, not in the core"). Add `*config.WriteError` with private fields (`op`, `path`, `cause`), `Error()`, `Unwrap()`, and `Op()`/`Path()`/`Cause()` accessors. Add unit tests for `WriteDefault` and `WriteError` using `t.TempDir()`. The `case *config.WriteError` arm of `output.FormatError` ships with the type introduction in Phase 1 task `1.5b` (same PR; the only consumer is `cmd/config_init.go` via `WriteDefault`).
- [ ] 4.2 In `cmd/config_init.go:144-159`, replace the inline `os.Stat` / `os.MkdirAll(0o750)` / `os.WriteFile(0o600)` block with `return config.WriteDefault(path, opts.Force)`. The function signature for `WriteDefault` must match the existing `Path/Force` semantics; no behavior change for `--force` or default path resolution. Update the `core.NewFileError` returns at `cmd/config_init.go:145, 152, 157` to also return `*config.WriteError` paths through the new helper (the helper unwraps and rewraps the underlying `*os.PathError`).
- [ ] 4.3 **DEFERRED** — Dropping the `runF` parameter from `cmd/library.go:11` (`NewLibraryCommand`) and `cmd/resources.go:48` (`NewCmdResources`) is deferred to a follow-up change. Rationale: the canonical `golang-cli-architecture` Factory pattern treats `runF` as the test-injection seam; until a concrete audit confirms the seam is unused across all current and planned test files for these two constructors (currently this change does not exercise that audit), the removal risks closing a future-test affordance. No edits to `cmd/library.go`, `cmd/resources.go`, `cmd/library_test.go`, `cmd/resources_test.go`, or `cmd/root.go:35` are made by this change.
- [ ] 4.4a In `internal/core/rules.go`, add `ValidateDocumentType(docType string) error` that returns `*core.ValidationError` for inputs not in `validResourceTypes`. Mirror `CanInstallResource`'s suggestion-builder pattern: `NewValidationError("canonicalize", "type", docType, "type must be one of skill, agent, command, memory").WithSuggestions([]string{"use one of: skill, agent, command, memory"})`. Export `validResourceTypes` as a public slice `ValidResourceTypes() []string` (or keep it unexported and have the helper own the lookup). Add unit tests in `internal/core/rules_test.go` mirroring the `TestCanInstallResource` table-driven cases. The only current caller of `core.ValidateDocumentType` is `cmd/canonicalize.go`; the helper is added with the expectation that future commands will need it.
- [ ] 4.4 Keep `_ = cmd.MarkFlagRequired("type")` at `cmd/canonicalize.go:96` unchanged. **Do NOT** add `Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs)` with `ValidArgs: []string{"agent", "command", "skill", "memory"}` — `cobra.ValidArgs` is for positional argument validation, and the positional args of `canonicalize` are `<input> <output>` file paths, so `ValidArgs: ["agent", ...]` would reject every legitimate file path. The original proposal's design was a misuse of `cobra.ValidArgs`. In `runCanonicalize`, add a `core.ValidateDocumentType(opts.DocType)` defense-in-depth pre-flight that returns `*core.ValidationError` for unknown types (catches the case where `--type` is provided but has an unknown value, e.g., `"skills"` plural, `"bot"`, `""`; do NOT use `CanInstallResource` — it validates `"type/name"` shape and would reject every valid bare type). Update `cmd/canonicalize_test.go` to assert the new validation error for unknown types. Flag-value completion for `--type` is already wired at `cmd/canonicalize.go:98-100` via `carapace.Gen(cmd).FlagCompletion(...)` and is unchanged.
- [ ] 4.5 Update `cmd/library_init.go:161-165` to set the `InitRequest.Stdout io.Writer` field on the `library.InitRequest` literal. The field is now available on the struct (the `fix-library-io-discipline` change has shipped). Pass `opts.IO.Out` to preserve the current stderr-as-OutForCobra semantics. The field is additive (existing call sites that pass `nil` for `Stdout` fall through to `os.Stdout`).
- [ ] 4.6 Cross-references — update `openspec/changes/harden-tests-and-coverage/tasks.md:6.6` to mark the `errEmptyResources` migration as deferred (this change owns it).
- [ ] 4.7 Run `mise run check`.

## 5. Phase 5 — Verification + spec sync

- [ ] 5.1 `mise run build` — no broken imports.
- [ ] 5.2 `mise run lint` — must report 0 issues (refresh `cmd/testdata/lint_baseline.txt` only if intentional new violations appear).
- [ ] 5.3 `mise run test` — all unit tests pass; new `UsageError` (with builder), `CobraUsageError` (with panic-on-nil guard), `MarshalJSON`, `*config.WriteError`, and `BatchFailureInfo.ErrorType`/`Cause` paths are covered. Per-package coverage targets:
  - `internal/core` ≥ 85% (new `UsageError`, `CobraUsageError`, `MarshalJSON` on 11 types, `MustNewCobraUsageError` panic path).
  - `internal/config` ≥ 80% (new `WriteError` type + `WriteDefault` helper, both new package surface).
  - `internal/output` ≥ 90% (three new dispatch arms — `*core.InitializeError`, `*core.UsageError`, `*config.WriteError` — plus the explicit `WriteError` case wiring).
  - Overall cross-package coverage ≥ 70%.

  The pre-existing `TestExitCodeFor` row `core PartialSuccessError S==0` (asserting `ExitCodeError (1)`) MUST remain green after Phase 3 — `*core.PartialSuccessError{Succeeded: 0, Failed: N}` is the failure-aggregate type for `cmd init --resources <missing>` (see `test/e2e/init_test.go:94-105`); verify the test row covers the `Succeeded == 0` case explicitly.
- [ ] 5.4 `mise run test:e2e` — E2E tests pass; verify no double-output in captured stderr.
- [ ] 5.5 Manual: run `germinator library show nonexistent-ref` and verify exit code is `1` (was `2`); verify NO deprecation canary warning on stderr (the canary was removed in Phase 1.7a); verify the user-visible text is `Error: not found: nonexistent-ref`.
- [ ] 5.5a Verify `cmd/show.go`'s lookup chain (after Phase 3) routes through `internal/library/resolver.go` or its loader and returns `*core.NotFoundError` for missing refs. If `cmd/show.go` uses a local lookup (e.g., direct `lib.Resources[type][name]` access), add a migration sub-task to surface the missing-key as `*core.NotFoundError` so task `5.5`'s exit-code assertion holds. Verified: `cmd/show.go:79` uses `cobra.ExactArgs(1)` and `cmd/show.go` calls `library.LoadLibrary` then dereferences `lib.Resources[type][name]` — add a `cmd/show.go` migration if the dereference happens in `cmd/` rather than `internal/library/`.
- [ ] 5.6 `openspec validate enforce-error-discipline --strict` — change is coherent.
