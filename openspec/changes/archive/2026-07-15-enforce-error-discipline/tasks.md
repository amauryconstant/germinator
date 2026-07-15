# Tasks — Enforce error-handling discipline

Tasks are ordered so each commit is independently testable. Every task ends with `mise run build && mise run lint && mise run test` passing before the next starts. File:line references in this document were verified against the tree at task `0.1`; deviations are corrected by that task.

> **Note**: Throughout this file, `gerrors` is the project alias for `internal/core/errors.go` (the `germinator/errors` import path); `core.*Error` in the prose refers to the same types via the `core` import in cmd/* code. Both names refer to the same set of typed errors (9 existing + 2 new after Phase 1).

## 0. Phase 0 — Cross-change reconciliation

> **Phase 0 reconciliation trace (verified 2026-07-15).** All file:line references below have been reconciled against the current tree. Original-to-current line shifts are recorded inline as `// shift: ±N` comments for each task. The verification commit body records this trace.

**Verified baseline (pre-Phase-1):**
- `go.mod` `go 1.25.5` — well above `1.21`; no `cloneSlice` fallback needed (task 0.1a).
- `rg "output\.FormatError" cmd/ main.go` reports 10 production CALLS in `cmd/` (`cmd/library_add.go:335,530,675,688`; `cmd/library_validate.go:154,162`; `cmd/validate.go:123`; `cmd/init.go:202,206,242`) plus 2 production CALLS in `main.go:32,46` (retained) plus 2 test CALLS (`cmd/init_test.go:186`, `cmd/show_test.go:239`) plus comment-only references. Total production+test hits: **14 calls pre-Phase 2; after Phase 2 only 4 production calls + 2 test calls remain**.
- `rg "errEmptyResources" cmd/` — declaration at `cmd/library_create.go:68` (shift: −2); single call site at `:173` (shift: −30).
- `rg --line-number 'BatchFailureInfo\{' internal/library/` — 5 population sites at `adder.go:667, 684, 702, 720, 784` (struct definition at `:541-544`; shift: +20 for population sites, +16 for struct).
- `rg --line-number 'NewFileError.*not found' internal/library/` — exactly 5 sites before migration (`resolver.go:21, 26`; `loader.go:36, 53`; `adder.go:157`); 0 after Phase 3.
- `rg --line-number 'NewConfigError.*preset not found' internal/library/` — exactly 1 site at `resolver.go:62` (shift: −8); 0 after Phase 3.
- `rg --line-number "os\.IsNotExist" internal/library/remover.go` — silent-swallow site at `:103` (shift: −1).
- `rg --line-number "os\.ErrNotExist" internal/library/` — 0 sites today; Phase 3.10 introduces the `errors.Is(err, os.ErrNotExist)` form.

- [x] 0.1 Re-verify every file:line in this `tasks.md` against the current tree. For each site, run the listed `rg` and `sed -n` checks; record discrepancies in this commit's body and update the affected task descriptions before proceeding. **The rebase extends beyond `tasks.md`** — update `proposal.md` Affected-code table and `specs/*/spec.md` `When … internal/library/adder.go:nnn` style site references to match the verified line numbers; the verification commit body MUST record each original-to-current line shift as a regression trace. Specifically:
  - `rg "output\.FormatError" cmd/ main.go` — must show 10 production calls in `cmd/` (deleted in Phase 2) + 2 production calls in `main.go` (lines 32, 46, retained) + 2 test calls in `cmd/` (retained) = 14 production/test hits before Phase 2; after Phase 2 only the 4 retained production calls + 2 test calls + comment references remain.
  - `rg "errEmptyResources" cmd/` — must show a single declaration at `cmd/library_create.go:68` and a single call at `:173`.
  - `rg --line-number 'BatchFailureInfo\{' internal/library/` — must show 5 population sites at `adder.go:667, 684, 702, 720, 784`.
  - `rg --line-number 'NewFileError.*not found' internal/library/` — must show exactly 5 sites (`resolver.go:21,26`; `loader.go:36,53`; `adder.go:157`) before migration; 0 after.
  - `rg --line-number 'NewConfigError.*preset not found' internal/library/` — must show exactly 1 site (`resolver.go:62`) before migration; 0 after.
  - `rg --line-number "os\.IsNotExist" internal/library/remover.go` — must show the silent-swallow site at `:103`.
- [x] 0.1a Verify `go.mod` Go version is `>= 1.21` (for `slices` stdlib import added in task `1.7`). **Verified: `go 1.25.5`** — above `1.21`, no `cloneSlice` fallback needed. Record `1.25.5` in the commit body.
- [x] 0.1b Audit `cmd/*.go` lookup chains for direct `lib.Resources[type][name]` or `lib.Presets[name]` dereferences. Run `rg "lib\.Resources\[|lib\.Presets\[" cmd/` and report every production-file match (test-file matches are expected and OK because they assert fixture state, not implement production lookups). For each production match, add a Phase 3 sub-task that migrates the lookup to `internal/library/resolver.go` and surfaces `*core.NotFoundError` from the resolver instead of the cmd layer.
  - **Production sites requiring migration** (Phase 3.20): `cmd/show.go:138, 142, 172` — confirmed direct dereferences of `lib.Resources[typ]` / `resources[name]` / `lib.Presets[presetName]`. Migrate to new `internal/library/resolver.go` helpers `ResolveResourceEntry(lib, ref) (*Resource, error)` and `ResolvePresetEntry(lib, name) (*Preset, error)` returning `*core.NotFoundError` on miss (existing `ResolveResource(lib, ref) (string, error)` returns the file path and keeps its signature).
  - **Production sites already correct — DO NOT migrate**: `cmd/library_remove.go:273, 277, 344` — the surrounding code returns `*core.NotFoundError` on miss (lines 275, 279, 346). The dereferences are direct but the typed-error dispatch is already correct; migrating them is scope creep.

## 1. Phase 1 — Exit-code semantics + dispatch set + canary exemption

- [x] 1.1 In `internal/cmdutil/exit.go:73`, change `*core.NotFoundError` mapping from `ExitCodeUsage` (2) to `ExitCodeError` (1).
- [x] 1.2 In `internal/cmdutil/exit.go:24-37`, delete the `cobraUsagePrefixes` slice (12 entries) and the `hasCobraUsagePrefix` function. The existing typed branch (`errors.As(err, &pflag.NotExistError)` etc. at lines 64-69) handles the four pflag-emitted types directly. Remove the corresponding call to `hasCobraUsagePrefix` at line 70. **Fallback**: if code review rejects the `*core.CobraUsageError` introduction in task 1.3, fall back to keeping a 4-prefix substring list — `requires at least`, `requires exactly`, `accepts at most`, `required flag(s)` — and skip tasks 1.3 and 3.14. The fallback's test rows replace the `*core.CobraUsageError` row in task 1.4 with four substring cases.
- [x] 1.3 In `internal/cmdutil/exit.go`, add a new typed branch for `*core.CobraUsageError`: `if errors.As(err, &cobraUsage) { return ExitCodeUsage }`. Update the function comment to enumerate the new dispatch set.
- [x] 1.4 In `internal/cmdutil/exit_test.go:58`, flip the `*core.NotFoundError` expectation from `ExitCodeUsage` to `ExitCodeError`. **Delete the 5 cobra-substring rows at lines 45-49** (these test the substring-fallback dispatch that task 1.2 removes; they are dead code after Phase 1). Add table rows for `*core.UsageError` (expects `ExitCodeUsage`) and `*core.CobraUsageError` (expects `ExitCodeUsage`). Update the package comment if it still references the substring fallback.
- [x] 1.5 In `internal/output/errors.go:21-50`, add `case *core.InitializeError` rendering `"Error: " + e.Error() + "\n"` (delegates to `InitializeError.Error()` which colon-joins `<ref>: output: <outputPath>: <cause>` segments per `internal/core/errors.go`). Add `case *core.UsageError` rendering `"Error: " + e.Flag() + ": " + e.Reason() + "\n"`. Reuse the existing `writeErrOut` helper.
- [x] 1.5b In `internal/output/errors.go`, add `case *config.WriteError` rendering `"Error: " + e.Op() + " " + e.Path() + ": " + e.Error() + "\n"` (reuses `writeErrOut`). The new case sits after `UsageError` and before the generic-error fallback; adds `internal/config` to the file imports if not present. Add an explicit `if errors.As(err, &writeErr) { return ExitCodeError }` branch in `internal/cmdutil/exit.go` between the `PartialSuccess` branch and the default. Add `internal/config` to the file imports. The explicit row matches the spec's ADDED Requirement at `cli-exit-codes/spec.md:91-100` (which mandates `*config.WriteError → ExitCodeError (1)` as a SHALL contract); the explicit row is preferred over relying on default-fallthrough — a future default-branch change (e.g., adding `Logger.Error`) cannot silently break the contract. Add a new test row to `internal/cmdutil/exit_test.go`: `name: "config WriteError", err: &config.WriteError{...}, want: ExitCodeError`.
- [x] 1.6 In `internal/output/exporter.go`, delete the `var _ = io.EOF` line and the `io` import on line 7 (verify zero other uses via `rg "io\." internal/output/exporter.go`).
- [x] 1.7 In `internal/core/errors.go`, add the `UsageError` and `CobraUsageError` types following the existing builder pattern:
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
- [x] 1.7a Remove the `internal/warning` canary. Delete `internal/warning/canary.go` (the `MaybeWarnLegacyExitCode` function and `canaryOnce` var), `internal/warning/canary_test.go`, and `internal/warning/AGENTS.md`. If `internal/warning/` is now empty, delete the directory. The exit-code 5 → 1 migration the canary was added to support is now complete; without removal, the new `NotFoundError → 1` mapping in task 1.1 would cause the canary to fire on every interactive lookup miss.
- [x] 1.8 Run `mise run test`, `mise run lint`, `mise run build` — all clean. (Tasks 1.8 and 1.9 from the original proposal are removed: the `internal/warning` canary is deleted entirely in task 1.7a, not widened with an exemption.)

## 2. Phase 2 — Single-handling rule cleanup

> **Phase 0 line-shift trace (Phase 2):**
> | Task | Original | Verified | Shift |
> |------|----------|----------|-------|
> | 2.1 | `cmd/library_add.go:355` | `:335` | −20 |
> | 2.2 | `:549` (FormatError) | `:530` (with `opErr :=` at `:527`) | −19 |
> | 2.3 | `:693` | `:675` | −18 |
> | 2.4 | `:706` (FormatError) | `:688` (with `opErr :=` at `:685`) | −18 |
> | 2.5 | `cmd/library_validate.go:182` | `:154` | −28 |
> | 2.6 | `:190` | `:162` | −28 |
> | 2.7 | `cmd/validate.go:123` | `:123` | match |
> | 2.8 | `cmd/init.go:218` | `:202` | −16 |
> | 2.9 | `:222` | `:206` | −16 |
> | 2.10 | `:258` | `:242` | −16 |

- [x] 2.1 In `cmd/library_add.go:335`, delete the inline `output.FormatError(opts.IO, opErr)` call (Mode 1 explicit error path). `cmd/library_add.go` already has a per-file `output.FormatError` chain in the run loop; the delete is a no-op for behavior, only removes the duplicate. Verify the lines above (`opErr := core.NewOperationError(...)` at `:333` and `initErrs = append(...)` at `:334`) still wrap the error and thread it into `initErrs`.
- [x] 2.2 In `cmd/library_add.go:530`, delete the inline `output.FormatError(opts.IO, opErr)` call (`runAddBatchFiles` failure path) AND change the line above from `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` (at `:527`) to `opErr := core.NewOperationError("add", f.Source, nil)` with `opErr.Cause` lifted from the typed cause (the typed error must come from `processBatchAddFile`, not from `f.Error` as a string). Adjust `adder.go` if needed to surface typed errors per task `3.11`.
- [x] 2.3 In `cmd/library_add.go:675`, delete the inline `output.FormatError(opts.IO, opErr)` call (`collectDiscoverFailures` conflict path). Verify the surrounding `if isPlainOutput(opts.Output)` guard (`:674-676`) remains intact.
- [x] 2.4 In `cmd/library_add.go:688`, delete the inline `output.FormatError(opts.IO, opErr)` call (`collectDiscoverFailures` batch failure path). Same wrap-and-lift pattern as 2.2 — `opErr := core.NewOperationError("add", f.Source, errors.New(f.Error))` is at `:685`.
- [x] 2.5 In `cmd/library_validate.go:154`, delete the inline `output.FormatError(opts.IO, err)` call. The wrapping `fmt.Errorf("loading library: %w", err)` at `:155` must remain.
- [x] 2.6 In `cmd/library_validate.go:162`, delete the inline `output.FormatError(opts.IO, err)` call. The wrapping `fmt.Errorf("validating library: %w", err)` at `:163` must remain.
- [x] 2.7 In `cmd/validate.go:123`, delete the inline `output.FormatError(opts.IO, e)` from inside the `for _, e := range result.Errors` loop. The terminal `return result.Errors[0]` at `:125` must remain so `ExitCodeFor` still maps the first validation error correctly.
- [x] 2.8 In `cmd/init.go:202`, delete the inline `output.FormatError(opts.IO, partialErr)` call from the all-failure branch (`succeeded == 0`).
- [x] 2.9 In `cmd/init.go:206`, delete the inline `output.FormatError(opts.IO, partialErr)` call from the partial-success branch (`succeeded > 0 && failed > 0`).
- [x] 2.10 In `cmd/init.go:242`, delete the inline `output.FormatError(opts.IO, core.NewInitializeError(r.Ref, r.InputPath, r.OutputPath, r.Error))` from `renderResults`. The terminal `fmt.Fprintf(opts.IO.Out, "Initialized %d resource(s).", s)` at `:255` (and the `, %d failed.` line at `:257`) must remain.
- [x] 2.11 Run `rg "output\.FormatError" cmd/ main.go` — must return the 2 retained production calls (`main.go:32` factory-build handler, `main.go:46` post-Execute handler) and 2 retained test-file calls (`cmd/init_test.go:186`, `cmd/show_test.go:239`), plus comment-only references in `cmd/AGENTS.md`, `cmd/commands/AGENTS.md`, and various `cmd/*_test.go` files. Any additional production CALL sites in `cmd/` are missed deletes from tasks `2.1-2.10`.
- [x] 2.12 Run `mise run test`, `mise run test:e2e` — no double-output in any captured stderr.

## 3. Phase 3 — Type migration + chain preservation

> **Phase 0 line-shift trace (Phase 3):**
> | Task | Original | Verified | Shift |
> |------|----------|----------|-------|
> | 3.1-3.2 | `internal/library/resolver.go:21, 26` | `:21, :26` | match |
> | 3.3 | `:70` | `:62` | −8 |
> | 3.4-3.5 | `loader.go:36, :53` | `:36, :53` | match |
> | 3.6 | `adder.go:146` | `:157` | +11 |
> | 3.7-3.9 | `remover.go:83, :88, :142` | `:82, :87, :140` | −1, −1, −2 |
> | 3.10 | `remover.go:104` (`os.IsNotExist` swallow) | `:103` | −1 |
> | 3.11 | `adder.go:525-529` (struct) | `:541-544` | +16 |
> | 3.11 | popul. `:647-651, 664-668, 682-686, 700-704, 764-768` | `:667, :684, :702, :720, :784` | +20 |
> | 3.12 | `cmd/library_create.go:70` | `:68` | −2 |
> | 3.12 | call `:203` | `:173` | −30 |
> | 3.13 | `resolver_test.go:174-177` | `:146-148` | −28 |
> | 3.13 | `cmd/library_create_test.go:343-365` | `:342-365` | match (range) |
> | 3.13 | `cmd/show_test.go:151, 179` | `:148,151 / :176,179` | match (off-by-3) |
> | 3.13 | `cmd/library_remove_test.go:397` | `:397` | match |
> | 3.13 | `test/e2e/init_test.go:108, 117` | `:108, :117` | match |
> | 3.20 | `cmd/show.go:138, 142, 172` | `:138, :142, :172` | match |

- [x] 3.1 In `internal/library/resolver.go:21`, replace `gerrors.NewFileError(ref, "resolve", "resource not found", nil)` with `gerrors.NewNotFoundError("resource", ref)`.
- [x] 3.2 In `internal/library/resolver.go:26`, same replacement.
- [x] 3.3 In `internal/library/resolver.go:62`, replace `gerrors.NewConfigError("preset", name, "preset not found")` with `gerrors.NewNotFoundError("preset", name)`.
- [x] 3.4 In `internal/library/loader.go:36`, replace `gerrors.NewFileError(path, "access", "library not found", nil)` with `gerrors.NewNotFoundError("library", path)`.
- [x] 3.5 In `internal/library/loader.go:53`, replace `gerrors.NewFileError(yamlPath, "read", "library.yaml not found", nil)` with `gerrors.NewNotFoundError("library.yaml", yamlPath)`.
- [x] 3.6 In `internal/library/adder.go:157`, replace `gerrors.NewFileError(source, "access", "source file not found", nil)` with `gerrors.NewNotFoundError("source file", source)`.
- [x] 3.7 In `internal/library/remover.go:82`, replace `gerrors.NewFileError(opts.LibraryPath, "access", fmt.Sprintf("resource %s not found", opts.Ref), nil)` with `gerrors.NewNotFoundError("library ref", opts.Ref)`.
- [x] 3.8 In `internal/library/remover.go:87`, same pattern (second occurrence of the `NewFileError` lookup, `nameExists` branch).
- [x] 3.9 In `internal/library/remover.go:140`, replace `gerrors.NewFileError(opts.LibraryPath, "access", fmt.Sprintf("preset %s not found", opts.Name), nil)` with `gerrors.NewNotFoundError("preset", opts.Name)`.
- [x] 3.10 In `internal/library/remover.go:103`, replace the silent `os.IsNotExist` swallow: `if errors.Is(err, os.ErrNotExist) { return nil, core.NewNotFoundError("library file", path) }`. Add the `os` and `errors` imports if missing. **Behavior change**: idempotent removal becomes non-idempotent (per design Decision 6). CHANGELOG `### Changed` entry required.
- [x] 3.11 In `internal/library/adder.go:541-544`, extend `BatchFailureInfo` (absorbed into Phase 2 by user decision — typed-cause lift landed alongside the Phase 2 single-handling-rule cleanup):
  ```go
  type BatchFailureInfo struct {
      Source    string `json:"source"`
      Error     string `json:"error"`
      ErrorType string `json:"errorType,omitempty"`
      Cause     error   `json:"cause,omitempty"`
  }
  ```
  Update every `BatchFailureInfo{...}` literal at `adder.go:667, 684, 702, 720, 784` to populate `ErrorType` and `Cause`. Compute `ErrorType` via a typed switch in a small helper in `adder.go`:
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
      case *gerrors.UsageError:
          return "UsageError"
      case *gerrors.CobraUsageError:
          return "CobraUsageError"
      case *os.PathError:
          return "PathError"
      default:
          return fmt.Sprintf("%T", cause)
      }
  }
  ```
  Plumb the typed cause through from `processBatchAddFile`'s callers so `Cause` is set even when `Error` is the stringified fallback. Population sites MUST convert non-typed causes (e.g., `*os.PathError` returned from a filesystem call) to `*core.FileError` before assigning to `f.Cause`; the stdlib default for non-typed errors is `{}` per `json.Marshaler` precedence rules, which defeats the typed-error-chain preservation contract documented in `errors-enhanced-errors/spec.md` at the "Cause MUST be a typed error" scenario.
- [x] 3.12 In `cmd/library_create.go:68`, migrate `errEmptyResources = errors.New("flag needs an argument: --resources (must be non-empty list of refs)")` to `errEmptyResources = core.NewUsageError("--resources", "must be non-empty list of refs")`. The single call site at `cmd/library_create.go:173` returns `errEmptyResources` directly — verify the new `*core.UsageError` flows through `ExitCodeFor` to `ExitCodeUsage (2)` in `cmd/library_create_test.go:152-176`.
- [x] 3.13 Update the following test files to assert `*core.NotFoundError` → `ExitCodeError (1)`:
  - `internal/library/resolver_test.go:146-148` — currently `errors.As(&cfgErr)` against `*gerrors.ConfigError`; change to `errors.As(&nf)` against `*gerrors.NotFoundError`.
  - `cmd/library_create_test.go:342-365` (T11 `TestRunCreatePreset_RefReferencesMissingResource`; the `ExitCodeUsage` assertion is at `:365`) — same swap.
  - `cmd/show_test.go:148-151, 176-179` — same (off-by-3 from original reference).
  - `cmd/library_remove_test.go:391-397` — same.
  - `test/e2e/init_test.go:108-117` — `Describe("init fails for nonexistent preset", It("should fail with exit code 2 for invalid preset name (NotFoundError → ExitCodeUsage)"))` — rename the `It` description to `should fail with exit code 1 for invalid preset name (NotFoundError → ExitCodeError)` and update the inline `ShouldFailWithExit(session, 2)` to `ShouldFailWithExit(session, 1)`. The slice-5 §5.0.1 inline comment (`:113-116`) must also be updated to reflect the corrected exit-code mapping.
  - `internal/cmdutil/exit_test.go:58` — already updated in task `1.4`, but verify no other row still expects `ExitCodeUsage` for `NotFoundError`.
- [x] 3.14 Verify no Cobra arg-validation errors need re-wrapping today. `cobra.ExactArgs`/`MinimumNArgs`/`MaximumNArgs`/`RangeArgs` validators and `MarkFlagRequired` errors flow through `cmd.Execute()` to `main.go:46` directly without touching `RunE`; `*core.CobraUsageError` has zero current call sites and is reserved for future typed dispatch (e.g., custom command code that needs typed exit-code mapping). The `ExitCodeFor` dispatch contract for `*core.CobraUsageError` is in place per task `1.3`; no code-side changes are required for the existing test suite. Add a unit test in `internal/core/errors_test.go` asserting `MustNewCobraUsageError(nil)` panics (programmer-error guard), `MustNewCobraUsageError(errors.New("x")).Error() == "x"`, and `errors.Unwrap(MustNewCobraUsageError(errors.New("x"))) == errors.New("x")`.
- [x] 3.15 Verification regex — `rg "NewFileError\([^)]*not found|NewConfigError\("preset"[^)]*preset not found" internal/library/` must return zero matches. Run `mise run check` before task `3.16`.
- [x] 3.16 Run `mise run test`, `mise run test:e2e`, `mise run test:coverage`.
- [x] 3.17 In `cmd/init.go:188-191`, drop the redundant `NewNotFoundError("preset", opts.Preset)` re-wrap. After task `3.3` lands, `lib.ResolvePreset` returns `*core.NotFoundError` directly; simplify to `if rerr != nil { return rerr }` (errors.As still resolves to `*core.NotFoundError` for exit-code mapping). Add a regression test asserting `cmd/init.go`'s not-found-preset error path returns `*core.NotFoundError` and `ExitCodeFor` returns `ExitCodeError` (1).
- [x] 3.18 Update `cmd/init_test.go` assertions to match the new `*core.UsageError` constructor (from tasks 1.7 and 3.12). Audit and rewrite any test that asserts against the prior `errEmptyResources` cobra-substring wording. Verification: `rg "flag needs an argument" cmd/` must return zero matches after Phase 3 lands.
- [x] 3.19 Drop the redundant `fmt.Errorf("libraryAdapter.AddResource: %w", err)` wrapchain noise identified in the 2026-07-08 review (B-010). When the inner error is already a `*core.*Error` typed error, the wrap prefix is meaningless — `errors.As` traverses the typed-error chain without it. Locate the production site (currently `cmd/library_add.go:90` in the `libraryAdapter.AddResource` method), replace `fmt.Errorf("libraryAdapter.AddResource: %w", err)` with the direct `err` return, and update affected tests. **Note**: `libraryAdapter.DiscoverOrphans` (`cmd/library_add.go:100`) and `libraryAdapter.BatchAddResources` (`:110`) carry the same redundant wrap pattern and SHOULD be similarly migrated for consistency. Verification: `rg "libraryAdapter\.(AddResource|DiscoverOrphans|BatchAddResources).*%w" cmd/` must return zero matches after this task lands.
- [x] 3.20 **Migrate unconditionally per Phase 0 decisions.** Add new `internal/library/resolver.go` helpers:
  - `ResolveResourceEntry(lib *Library, ref string) (*Resource, error)` — returns the `Resource` struct (not just the file path) with `*core.NotFoundError` on miss. Replaces `cmd/show.go:138, 142` direct dereferences.
  - `ResolvePresetEntry(lib *Library, name string) (*Preset, error)` — returns the `Preset` struct with `*core.NotFoundError` on miss. Replaces `cmd/show.go:172` direct dereference.

  Update `cmd/show.go:renderResource` to call `ResolveResourceEntry` and `cmd/show.go:renderPreset` to call `ResolvePresetEntry`. Existing `ResolveResource(lib, ref) (string, error)` and `(*Library).ResolvePreset(ctx, name) ([]string, error)` keep their current signatures. Update `cmd/show_test.go:148-151, 176-179` (already uses `*core.NotFoundError` assertions; verify no wording change needed). Do NOT introduce new lookup logic at the cmd layer; encapsulate in the resolver per the design's "Migrate lookup branches to internal/library" Direction (Decision #4). Confirmed-out-of-scope matches (already correct): `cmd/library_remove.go:273, 277, 344` — the surrounding code returns `*core.NotFoundError` on miss; do not migrate.

## 4. Phase 4 — Trivial folds + cross-change imports

- [x] 4.1 Create `internal/config/scaffold.go` with `func WriteDefault(path string, force bool) error`. Implementation: `if !force { os.Stat } else { skip }` → `os.MkdirAll(dir, 0o750)` → `os.WriteFile(path, []byte(defaultTOML), 0o600)`. Return `*config.WriteError` (new domain type in `internal/config/errors.go`) on I/O failure — NOT `*core.FileError`, because `internal/config/` is the Imperative Shell layer that owns its own I/O errors per `golang-cli-architecture/references/05-errors.md` ("errors that originate from external I/O are defined near their origin, not in the core"). Add `*config.WriteError` with private fields (`op`, `path`, `cause`), `Error()`, `Unwrap()`, and `Op()`/`Path()`/`Cause()` accessors. Add unit tests for `WriteDefault` and `WriteError` using `t.TempDir()`. The `case *config.WriteError` arm of `output.FormatError` ships with the type introduction in Phase 1 task `1.5b` (same PR; the only consumer is `cmd/config_init.go` via `WriteDefault`).
- [x] 4.2 In `cmd/config_init.go:144-159`, replace the inline `os.Stat` / `os.MkdirAll(0o750)` / `os.WriteFile(0o600)` block with `return config.WriteDefault(path, opts.Force)`. The function signature for `WriteDefault` must match the existing `Path/Force` semantics; no behavior change for `--force` or default path resolution. Update the `core.NewFileError` returns at `cmd/config_init.go:145, 152, 157` to also return `*config.WriteError` paths through the new helper (the helper unwraps and rewraps the underlying `*os.PathError`).
- [x] 4.3 **DEFERRED** — Dropping the `runF` parameter from `cmd/library.go:11` (`NewLibraryCommand`) and `cmd/resources.go:48` (`NewCmdResources`) is deferred to a follow-up change. Rationale: the canonical `golang-cli-architecture` Factory pattern treats `runF` as the test-injection seam; until a concrete audit confirms the seam is unused across all current and planned test files for these two constructors (currently this change does not exercise that audit), the removal risks closing a future-test affordance. No edits to `cmd/library.go`, `cmd/resources.go`, `cmd/library_test.go`, `cmd/resources_test.go`, or `cmd/root.go:35` are made by this change.
- [x] 4.4a In `internal/core/rules.go`, add `ValidateDocumentType(docType string) error` that returns `*core.ValidationError` for inputs not in `validResourceTypes`. Mirror `CanInstallResource`'s suggestion-builder pattern: `NewValidationError("canonicalize", "type", docType, "type must be one of skill, agent, command, memory").WithSuggestions([]string{"use one of: skill, agent, command, memory"})`. Export `validResourceTypes` as a public slice `ValidResourceTypes() []string` (or keep it unexported and have the helper own the lookup). Add unit tests in `internal/core/rules_test.go` mirroring the `TestCanInstallResource` table-driven cases. The only current caller of `core.ValidateDocumentType` is `cmd/canonicalize.go`; the helper is added with the expectation that future commands will need it.
- [x] 4.4 Keep `_ = cmd.MarkFlagRequired("type")` at `cmd/canonicalize.go:96` unchanged. **Do NOT** add `Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs)` with `ValidArgs: []string{"agent", "command", "skill", "memory"}` — `cobra.ValidArgs` is for positional argument validation, and the positional args of `canonicalize` are `<input> <output>` file paths, so `ValidArgs: ["agent", ...]` would reject every legitimate file path. The original proposal's design was a misuse of `cobra.ValidArgs`. In `runCanonicalize`, add a `core.ValidateDocumentType(opts.DocType)` defense-in-depth pre-flight that returns `*core.ValidationError` for unknown types (catches the case where `--type` is provided but has an unknown value, e.g., `"skills"` plural, `"bot"`, `""`; do NOT use `CanInstallResource` — it validates `"type/name"` shape and would reject every valid bare type). Update `cmd/canonicalize_test.go` to assert the new validation error for unknown types. Flag-value completion for `--type` is already wired at `cmd/canonicalize.go:98-100` via `carapace.Gen(cmd).FlagCompletion(...)` and is unchanged.
- [x] 4.5 **NO-OP per Phase 0 decision** — `internal/library/InitRequest` does NOT have a `Stdout io.Writer` field; the field was assumed by the original proposal but never landed. `library.Init(ctx, req, stdout)` already takes `stdout` as a separate parameter (see `internal/library/creator.go:104`), and `cmd/library_init.go:162-166` already passes `opts.IO.Out` correctly. **No code change required.** Recorded here for traceability — the proposal/design assumption did not match reality. If `library.InitRequest.Stdout` is added in a future change, this task can be revisited.
- [x] 4.6 Cross-references — update `openspec/changes/harden-tests-and-coverage/tasks.md:6.6` to mark the `errEmptyResources` migration as deferred (this change owns it). **(Verified in Phase 3: cross-reference at `harden-tests-and-coverage/tasks.md:130` already in place from the Phase 3.12 implementation; no edit required.)**
- [x] 4.7 Run `mise run check`.

## 5. Phase 5 — Verification + spec sync

- [x] 5.1 `mise run build` — no broken imports.
- [x] 5.2 `mise run lint` — must report 0 issues (refresh `cmd/testdata/lint_baseline.txt` only if intentional new violations appear).
- [x] 5.3 `mise run test` — all unit tests pass; new `UsageError` (with builder), `CobraUsageError` (with panic-on-nil guard), `MarshalJSON`, `*config.WriteError`, and `BatchFailureInfo.ErrorType`/`Cause` paths are covered. Per-package coverage targets:
  - `internal/core` ≥ 85% (new `UsageError`, `CobraUsageError`, `MarshalJSON` on 11 types, `MustNewCobraUsageError` panic path).
  - `internal/config` ≥ 80% (new `WriteError` type + `WriteDefault` helper, both new package surface).
  - `internal/output` ≥ 90% (three new dispatch arms — `*core.InitializeError`, `*core.UsageError`, `*config.WriteError` — plus the explicit `WriteError` case wiring).
  - Overall cross-package coverage ≥ 70%.

  The pre-existing `TestExitCodeFor` row `core PartialSuccessError S==0` (asserting `ExitCodeError (1)`) MUST remain green after Phase 3 — `*core.PartialSuccessError{Succeeded: 0, Failed: N}` is the failure-aggregate type for `cmd init --resources <missing>` (see `test/e2e/init_test.go:94-105`); verify the test row covers the `Succeeded == 0` case explicitly.
- [x] 5.4 `mise run test:e2e` — E2E tests pass; verify no double-output in captured stderr.
- [x] 5.5 Manual: run `germinator library show nonexistent-ref` and verify exit code is `1` (was `2`); verify NO deprecation canary warning on stderr (the canary was removed in Phase 1.7a); verify the user-visible text is `Error: not found: nonexistent-ref`.
- [x] 5.5a Verify `cmd/show.go`'s lookup chain (after Phase 3) routes through `internal/library/resolver.go` or its loader and returns `*core.NotFoundError` for missing refs. If `cmd/show.go` uses a local lookup (e.g., direct `lib.Resources[type][name]` access), add a migration sub-task to surface the missing-key as `*core.NotFoundError` so task `5.5`'s exit-code assertion holds. Verified: `cmd/show.go:79` uses `cobra.ExactArgs(1)` and `cmd/show.go` calls `library.LoadLibrary` then dereferences `lib.Resources[type][name]` — add a `cmd/show.go` migration if the dereference happens in `cmd/` rather than `internal/library/`.
- [x] 5.6 `openspec validate enforce-error-discipline --strict` — change is coherent.
