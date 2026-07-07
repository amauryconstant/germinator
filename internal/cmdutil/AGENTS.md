**Location**: `internal/cmdutil/`
**Parent**: See `/internal/AGENTS.md` for package overview

---

# Cmdutil Package

Lazy DI (`Factory`), per-Factory completion cache, exit-code mapping, and shared cmd helpers. The single composition root is `main.go`, which constructs the `Factory` and assigns its lazy fields.

## Files

| File | Purpose |
|------|---------|
| `factory.go` | `Factory` struct (eager `IOStreams`/`AppVersion`/`Executable`/`RootContext`/`CompletionCache`; lazy `func() (T, error)` fields for `Config`/`Library`); `NewFactory(ctx, io, ver, exe)`; `Close()`; `OnceValuesFunc[T]` helper |
| `completion_cache.go` | `CompletionCache` type — per-Factory TTL cache for shell-completion library snapshots; `Get`/`Set`/`Reset`/`Invalidate` (concurrency-safe) |
| `exit.go` | `ExitCode` (`int`), `ExitCodeSuccess=0`/`ExitCodeError=1`/`ExitCodeUsage=2`; `ExitCodeFor(err)` via `errors.As` on pflag typed errors + Cobra string-prefix fallback + `*core.NotFoundError` (2) + `*core.PartialSuccessError` (0 if `Succeeded>0`, else 1) |
| `output_flags.go` | Re-export of `output.AddOutputFlags` as `cmdutil.AddOutputFlags` so cmd files import only `cmdutil` |
| `factory_test.go` | Lazy caching, cross-field caching, concurrent first-call, transient-error caching |
| `completion_cache_test.go` | Get/Set/Reset/Invalidate, TTL expiry, concurrency |
| `exit_test.go` | Table-driven: nil/pflag/Cobra/typed core errors/PartialSuccess/NotFound |
| `integration_test.go` | Cross-package: `*core.PartialSuccessError{Succeeded:3,Failed:1}` → `ExitCodeFor==0` + `FormatError` writes partial-success string |

## Key Surface

- `NewFactory(ctx, io, appVersion, executable)` — eager values only; lazy `Config`/`Library` fields must be assigned by `main.go` (the only composition root) using `OnceValuesFunc` (or `sync.OnceValues`) wrappers; `CompletionCache` is assigned alongside
- `Factory.Close()` — cancels `RootContext` (caller is `main.go`)
- `Factory.CompletionCache` — `*CompletionCache`; `Invalidate()` is called by every mutating library command (`runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`) so the next completion reflects the new state
- `NewCompletionCache()` — fresh cache; populated once in `main.go`
- `ExitCodeFor(err error) ExitCode` — 0/1/2 mapping; no 3–6
- `AddOutputFlags(cmd *cobra.Command, target *string)` — re-exports `output.AddOutputFlags`
