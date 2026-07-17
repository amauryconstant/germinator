**Location**: `internal/cmdutil/`
**Parent**: See `/internal/AGENTS.md` for package overview
**Skill references**: `@.opencode/skills/golang-cli-architecture/references/01-architecture.md`, `@.opencode/skills/golang-cli-architecture/references/08-completion.md`, `@.opencode/skills/golang-context/SKILL.md`

---

# Cmdutil Package

Lazy DI (`Factory`), per-Factory completion cache, exit-code mapping, and shared cmd helpers. The single composition root is `main.go` at the project root, which calls `BuildFactory` to obtain a fully-wired `*Factory`.

## Files

| File | Purpose |
|------|---------|
| `factory.go` | `Factory` struct (eager `IOStreams`/`RootContext`/`CompletionCache`; lazy `func() (T, error)` field for `Config`); `NewFactory(ctx, io)`; `BuildFactory(ctx, io)` — eagerly loads `Config` and creates the `CompletionCache`; `Close()`; `OnceValuesFunc[T]` helper; `swapConfigLoadForTest` test seam (mutex-protected package-level swap of the config loader, used via `t.Cleanup`) |
| `completion_cache.go` | `CompletionCache` type — per-Factory TTL cache for shell-completion library snapshots; `Get`/`Set`/`Reset`/`Invalidate` (concurrency-safe) |
| `exit.go` | `ExitCode` (`int`), `ExitCodeSuccess=0`/`ExitCodeError=1`/`ExitCodeUsage=2`; `ExitCodeFor(err)` via typed `errors.As` dispatch on `*pflag.{NotExist,ValueRequired,InvalidValue,InvalidSyntax}Error`, `*core.{Usage,CobraUsage}Error` (both → 2), `*core.NotFoundError` (1), `*core.PartialSuccessError` (0 if `Succeeded>0`, else 1), `*config.WriteError` (1); default `ExitCodeError` |
| `factory_test.go` | Lazy caching, cross-field caching, concurrent first-call, transient-error caching, `BuildFactory` wiring |
| `completion_cache_test.go` | Get/Set/Reset/Invalidate, TTL expiry, concurrency |
| `exit_test.go` | Table-driven: nil/pflag/Cobra/typed core errors/PartialSuccess/NotFound |
| `integration_test.go` | Cross-package: `*core.PartialSuccessError{Succeeded:3,Failed:1}` → `ExitCodeFor==0` + `FormatError` writes partial-success string |

## Key Surface

- `NewFactory(ctx, io)` — eager values only; the lazy `Config` field is nil at return; `BuildFactory` is the post-`NewFactory` constructor used by `main.go`.
- `BuildFactory(ctx, io)` — assigns `CompletionCache = NewCompletionCache()`, wraps `config.Load` in `OnceValuesFunc` for `Config`, eagerly calls `f.Config()` to surface I/O errors, and activates `io.SetDebug(cfg.Debug)`.
- `Factory.Close()` — cancels `RootContext` (caller is `main.go`)
- `Factory.CompletionCache` — `*CompletionCache`; `Invalidate()` is called by every mutating library command (`runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`) so the next completion reflects the new state
- `NewCompletionCache()` — fresh cache; populated once in `main.go`
- `ExitCodeFor(err error) ExitCode` — 0/1/2 mapping; no 3–6
- `OnceValuesFunc[T]` — wraps `fn func() (T, error)` to memoize the first call's result under a `sync.Once`
