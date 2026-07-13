# shell-completion Specification (delta)

## MODIFIED Requirements

### Requirement: Completion actions take Factory as input

The completion action functions (`actionResources`, `actionPresets`, `actionLibraryRefs`) SHALL be implemented as `func(*cmdutil.Factory, *cobra.Command) carapace.Action` (the Factory and the Cobra command). They SHALL wrap `f.RootContext` with `context.WithTimeout(f.RootContext, getCompletionTimeout(f.Config()))` for each lookup, consult `f.CompletionCache.Get(libPath)` first, and on cache miss load the library directly via `library.LoadLibrary(loadCtx, libPath)` rather than `f.Library()`. The bypass of `f.Library()` is intentional: `f.Library` is `sync.OnceValues`-cached and would permanently pin the first error; completion lookups must always reflect current state. `actionPlatforms` SHALL be implemented as `func(*cmdutil.Factory) carapace.Action` (the Factory only) since it returns static platform values and does not need the Cobra command. The default timeout (`cmd/completions.go:20`) is `500ms`; configurable via `completion.timeout` in the config file.

**Change**: replace `getCompletionTimeout(nil)` with `getCompletionTimeout(f.Config())` at lines 305 and 312. The `*config.Config` is loaded lazily via the Factory's `Config` field (per `cli-cli-factory`).

#### Scenario: actionResources loads library with timeout and bypasses f.Library

- **WHEN** `actionResources(f, cmd)` returns an Action that runs
- **THEN** the Action SHALL consult `f.CompletionCache.Get(libPath)` first; on hit it returns the cached library
- **AND** on cache miss it SHALL call `library.LoadLibrary(loadCtx, libPath)` directly (NOT `f.Library()`)
- **AND** it SHALL use `context.WithTimeout(f.RootContext, getCompletionTimeout(f.Config()))` (default `500ms`, configurable via `completion.timeout`) as `loadCtx`

#### Scenario: Timeout returns empty completion

- **WHEN** the library load times out (default `500ms`, or whatever `completion.timeout` resolves to)
- **THEN** `actionResources` SHALL return an empty completion (no error)

### Requirement: Library Path Resolution for Completions

Completions SHALL resolve the library path using a priority chain.

**Change**: Updated the "Resolve library from default" scenario to reflect the XDG-resolved data path (`$XDG_DATA_HOME/germinator/library/` via `adrg/xdg.DataFile`). The source-of-truth scenario at lines 188-216 of `openspec/specs/cli-shell-completion/spec.md` referenced `~/.config/germinator/library/` — the config dir, not the data dir — pre-existing source-of-truth drift corrected by this delta. The other existing scenarios become testable after this change lands: `cmd/completions.go:120, 132, 144` previously passed `nil` to `resolveLibraryPath`; after this change they pass `f.Config()` so the "Resolve library from config" scenario (lines 203-210) actually executes.

#### Scenario: Resolve library from flag

- **GIVEN** a user has typed `--library /custom/path` before triggering completion
- **WHEN** completion loads the library
- **THEN** it SHALL use `/custom/path` as the library path

#### Scenario: Resolve library from environment

- **GIVEN** `GERMINATOR_LIBRARY=/env/path` is set
- **AND** no `--library` flag was typed
- **WHEN** completion loads the library
- **THEN** it SHALL use `/env/path` as the library path

#### Scenario: Resolve library from config

- **GIVEN** config file specifies `library = "/config/path"`
- **AND** no `--library` flag was typed
- **AND** `GERMINATOR_LIBRARY` is not set
- **WHEN** completion loads the library
- **THEN** it SHALL use `/config/path` as the library path

#### Scenario: Resolve library from default

- **GIVEN** no `--library` flag, no env var, and no config
- **WHEN** completion loads the library
- **THEN** it SHALL use the XDG-resolved data path (`$XDG_DATA_HOME/germinator/library/` if `XDG_DATA_HOME` is set, falling back to `~/.local/share/germinator/library/` on Unix, or the platform-appropriate `%LocalAppData%` path on Windows) via `adrg/xdg.DataFile("germinator/library")`

### Requirement: Completion Configuration

The CLI SHALL support configurable timeout and caching for completions.

**Change**: NO scenario text change. The two "Configurable" scenarios (lines 168-170, 178-182) become test-passing after this change — `cmd/completions.go:103, 111` previously passed `nil` to the helpers; after this change they pass `f.Config()`.

#### Scenario: Default completion timeout

- **GIVEN** no completion config is specified
- **WHEN** a completion action loads the library
- **THEN** it SHALL timeout after 500ms

#### Scenario: Configurable completion timeout

- **GIVEN** config file specifies `completion.timeout = "1s"`
- **WHEN** a completion action loads the library
- **THEN** it SHALL timeout after 1 second

#### Scenario: Default cache TTL

- **GIVEN** no completion config is specified
- **WHEN** library data is cached for completions
- **THEN** the cache SHALL expire after 5 seconds

#### Scenario: Configurable cache TTL

- **GIVEN** config file specifies `completion.cache_ttl = "10s"`
- **WHEN** library data is cached for completions
- **THEN** the cache SHALL expire after 10 seconds
