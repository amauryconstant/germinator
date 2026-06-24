**Location**: `internal/cmdutil/`
**Parent**: See `/internal/AGENTS.md` for package overview

---

# Cmdutil Package

Lazy DI (`Factory`), exit code mapping, and shared cmd helpers. Foundation units exist in slice 1; `main.go` is not yet rewired — that happens in slice 2.

## Files

| File | Purpose |
|------|---------|
| `factory.go` | `Factory` struct (eager `IOStreams`/`AppVersion`/`Executable`/`RootContext`; lazy `func() (T, error)` fields for `Config`/`Library`/`Transformer`/`Validator`/`Canonicalizer`/`Initializer`); `NewFactory(io, ver, exe)`; `OnceValuesFunc[T]` helper |
| `exit.go` | `ExitCode` (`int`), `ExitCodeSuccess=0`/`ExitCodeError=1`/`ExitCodeUsage=2`; `ExitCodeFor(err)` via `errors.As` on pflag typed errors + Cobra string-prefix fallback + `*core.PartialSuccessError` (0 if `Succeeded>0`, else 1) |
| `output_flags.go` | Re-export of `output.AddOutputFlags` as `cmdutil.AddOutputFlags` so cmd files import only `cmdutil` |
| `factory_test.go` | Lazy caching, cross-field caching, concurrent first-call, transient-error caching |
| `exit_test.go` | Table-driven: nil/pflag/Cobra/typed core errors/PartialSuccess |
| `integration_test.go` | Cross-package: `*core.PartialSuccessError{Succeeded:3,Failed:1}` → `ExitCodeFor==0` + `FormatError` writes partial-success string |

## Key Surface

- `NewFactory(io, appVersion, executable)` — eager values only; lazy fields must be assigned by `main.go` (the only composition root) using `sync.OnceValues` wrappers
- `Factory.Close()` — cancels `RootContext` (caller is `main.go`)
- `ExitCodeFor(err error) ExitCode` — 0/1/2 mapping; no 3-6
- `AddOutputFlags(cmd *cobra.Command, target *string)` — re-exports `output.AddOutputFlags`
