# shell-completion Specification (delta)

## MODIFIED Requirements

### Requirement: Completion cache lives on Factory

The completion cache SHALL be a `Factory.CompletionCache` field of type `*Cache` (defined in `cmd/completions.go`), populated in `main.go`. Each Factory instance has its own cache; constructing a new Factory starts with a fresh cache.

#### Scenario: Cache is per-Factory

- **GIVEN** two Factory instances `f1` and `f2`, each with a `CompletionCache`
- **WHEN** `f1.CompletionCache.Set("key", "value1")` is called
- **THEN** `f2.CompletionCache.Get("key")` SHALL return `nil` (each Factory has an independent cache)

#### Scenario: Cache has Reset method

- **WHEN** `f.CompletionCache.Reset()` is called
- **THEN** all cached entries SHALL be removed
- **AND** the cache SHALL be reusable (subsequent Set/Get calls work as if newly constructed)

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

The completion action functions (`actionResources`, `actionPresets`, `actionLibraryRefs`, `actionPlatforms`) SHALL take the Factory as input and use the Factory's library loader with a timeout (default 5 seconds, derived from `f.RootContext`).

#### Scenario: actionResources uses Factory.Library

- **WHEN** `actionResources(f, cmd, args, complete)` is called
- **THEN** it SHALL call `f.Library()` to obtain the library
- **AND** it SHALL use `context.WithTimeout(f.RootContext, 5*time.Second)` for the lookup

#### Scenario: Timeout returns empty completion

- **WHEN** the library load times out (5 seconds)
- **THEN** `actionResources` SHALL return an empty completion (no error)

### Requirement: TTL safety net preserved

The completion cache SHALL have a TTL (5 seconds, matching the legacy behavior) as a safety net for any missed explicit invalidation.

#### Scenario: TTL expires entries

- **GIVEN** an entry is cached at time T
- **WHEN** `Get("key")` is called at time T+6s
- **THEN** the entry SHALL be considered expired and a fresh load SHALL be triggered

> **Status:** the completion cache moves to `Factory.CompletionCache` in change-9 (`migrate-completion-cleanup`). This is the final change; after this archives, the migration is complete.
