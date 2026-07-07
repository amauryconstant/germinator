# shell-completion Specification (delta)

## MODIFIED Requirements

### Requirement: Completion Cache

The CLI SHALL cache library data to improve shell-completion performance, with the cache living on `Factory.CompletionCache` (a `*cmdutil.CompletionCache` field defined in `internal/cmdutil/completion_cache.go`). Each Factory instance has its own cache; constructing a new Factory starts with a fresh cache. The cache SHALL have a TTL (5 seconds, matching the legacy behavior) as a safety net for any missed explicit invalidation.

#### Scenario: Cache hit returns cached data

- **GIVEN** library data was loaded within the cache TTL
- **WHEN** a user triggers completion again
- **THEN** it SHALL return cached data without reloading the library

#### Scenario: Cache miss reloads library

- **GIVEN** library data was loaded but cache TTL has expired
- **WHEN** a user triggers completion
- **THEN** it SHALL reload the library from disk

#### Scenario: Cache is process-scoped

- **GIVEN** library data was cached in a previous shell process
- **WHEN** a new shell process triggers completion
- **THEN** it SHALL NOT have access to the previous process's cache
- **AND** it SHALL reload the library

#### Scenario: Cache is per-Factory

- **GIVEN** two Factory instances `f1` and `f2`, each with a `CompletionCache`
- **WHEN** `f1.CompletionCache.Set("key", "value1")` is called
- **THEN** `f2.CompletionCache.Get("key")` SHALL return `nil` (each Factory has an independent cache)

#### Scenario: Cache has Reset method

- **WHEN** `f.CompletionCache.Reset()` is called
- **THEN** all cached entries SHALL be removed
- **AND** the cache SHALL be reusable (subsequent Set/Get calls work as if newly constructed)

#### Scenario: TTL expires entries

- **GIVEN** an entry is cached at time T
- **WHEN** `Get("key")` is called at time T+6s
- **THEN** the entry SHALL be considered expired and a fresh load SHALL be triggered

## ADDED Requirements

### Requirement: Cache.Invalidate explicit invalidation

When a mutating library command completes successfully, the completion cache SHALL be invalidated so that subsequent completion calls reflect the new state. Mutating commands (`runAdd`, `runRemove`, `runCreate`, `runLibraryInit`, `runRefresh`, `runLibraryValidate`) SHALL call `f.CompletionCache.Invalidate()` before returning nil.

#### Scenario: Invalidate after runAdd

- **WHEN** `runAdd(opts)` successfully adds a resource to the library
- **THEN** the function SHALL call `f.CompletionCache.Invalidate()` before returning nil

#### Scenario: Fresh resource appears in completion

- **GIVEN** a successful `runAdd` that added a resource named `skill/test`
- **WHEN** `germinator library show <TAB>` is invoked immediately after
- **THEN** `skill/test` SHALL appear in the completion list
- **AND** the user SHALL NOT have to wait for the cache TTL to expire

### Requirement: Completion actions take Factory as input

The completion action functions (`actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms`) SHALL be implemented as `func(*cmdutil.Factory, *cobra.Command) carapace.Action` (the Factory and the Cobra command). They SHALL wrap `f.RootContext` with `context.WithTimeout(f.RootContext, 5*time.Second)` for each lookup, consult `f.CompletionCache.Get(libPath)` first, and on cache miss load the library directly via `library.LoadLibrary(loadCtx, libPath)` rather than `f.Library()`. The bypass of `f.Library()` is intentional: `f.Library` is `sync.OnceValues`-cached and would permanently pin the first error; completion lookups must always reflect current state.

#### Scenario: actionResources loads library with timeout and bypasses f.Library

- **WHEN** `actionResources(f, cmd)` returns an Action that runs
- **THEN** the Action SHALL consult `f.CompletionCache.Get(libPath)` first; on hit it returns the cached library
- **AND** on cache miss it SHALL call `library.LoadLibrary(loadCtx, libPath)` directly (NOT `f.Library()`)
- **AND** it SHALL use `context.WithTimeout(f.RootContext, 5*time.Second)` as `loadCtx`

#### Scenario: Timeout returns empty completion

- **WHEN** the library load times out (5 seconds)
- **THEN** `actionResources` SHALL return an empty completion (no error)

> **Status:** the completion cache moves to `Factory.CompletionCache` in change-9 (`migrate-completion-cleanup`). This is the final change; after this archives, the migration is complete.
