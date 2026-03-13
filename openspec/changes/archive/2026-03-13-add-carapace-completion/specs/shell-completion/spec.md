# shell-completion Specification

## Purpose

Define a Carapace-based shell completion system providing dynamic suggestions for library resources, presets, and platforms across multiple shells.

## ADDED Requirements

### Requirement: Completion Command

The CLI SHALL provide a `completion` command that generates shell completion scripts for multiple shells.

#### Scenario: Completion command exists

- **GIVEN** germinator is installed
- **WHEN** a user runs `germinator --help`
- **THEN** the output SHALL list a "completion" subcommand

#### Scenario: Completion supports multiple shells

- **GIVEN** the completion command exists
- **WHEN** a user runs `germinator completion --help`
- **THEN** it SHALL list subcommands for: bash, zsh, fish, powershell, elvish, nushell, oil, tcsh, xonsh, cmd-clink

#### Scenario: Generate bash completion

- **GIVEN** germinator is installed
- **WHEN** a user runs `germinator completion bash`
- **THEN** it SHALL output a valid bash completion script
- **AND** the script SHALL be sourceable in bash

#### Scenario: Generate zsh completion

- **GIVEN** germinator is installed
- **WHEN** a user runs `germinator completion zsh`
- **THEN** it SHALL output a valid zsh completion script
- **AND** the script SHALL be sourceable in zsh

#### Scenario: Generate fish completion

- **GIVEN** germinator is installed
- **WHEN** a user runs `germinator completion fish`
- **THEN** it SHALL output a valid fish completion script
- **AND** the script SHALL be sourceable in fish

#### Scenario: Shell-specific instructions

- **GIVEN** a user runs `germinator completion <shell>`
- **WHEN** the output is displayed
- **THEN** it SHALL include shell-specific installation instructions in the help text

---

### Requirement: Platform Static Completions

The CLI SHALL provide static completions for the `--platform` flag on relevant commands.

#### Scenario: Complete platform flag on adapt command

- **GIVEN** a user types `germinator adapt input.yaml output.md --platform <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest: `claude-code`, `opencode`

#### Scenario: Complete platform flag on validate command

- **GIVEN** a user types `germinator validate doc.yaml --platform <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest: `claude-code`, `opencode`

#### Scenario: Complete platform flag on init command

- **GIVEN** a user types `germinator init --platform <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest: `claude-code`, `opencode`

#### Scenario: Complete platform flag on canonicalize command

- **GIVEN** a user types `germinator canonicalize input.md output.yaml --platform <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest: `claude-code`, `opencode`

---

### Requirement: Dynamic Resource Completions

The CLI SHALL provide dynamic completions for resource references from the library.

#### Scenario: Complete resources flag on init command

- **GIVEN** a library exists with resources
- **AND** a user types `germinator init --resources <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest available resources in `type/name` format
- **AND** suggestions SHALL include: `skill/commit`, `skill/merge-request`, `agent/reviewer`, etc.

#### Scenario: Complete multiple resources

- **GIVEN** a library exists with resources
- **AND** a user has already typed `germinator init --resources skill/commit,<TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest remaining available resources

#### Scenario: Resources completion with library path

- **GIVEN** a user specifies `--library /custom/path`
- **AND** a user types `germinator init --library /custom/path --resources <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL load the library from `/custom/path`
- **AND** it SHALL suggest resources from that library

---

### Requirement: Dynamic Preset Completions

The CLI SHALL provide dynamic completions for preset names from the library.

#### Scenario: Complete preset flag on init command

- **GIVEN** a library exists with presets
- **AND** a user types `germinator init --preset <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest available preset names
- **AND** suggestions SHALL include preset names like `git-workflow`, `ai-coding`, etc.

#### Scenario: Preset completion with library path

- **GIVEN** a user specifies `--library /custom/path`
- **AND** a user types `germinator init --library /custom/path --preset <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL load the library from `/custom/path`
- **AND** it SHALL suggest presets from that library

---

### Requirement: Library Show Completions

The CLI SHALL provide completions for the `library show` command argument.

#### Scenario: Complete library show argument

- **GIVEN** a library exists with resources and presets
- **AND** a user types `germinator library show <TAB>`
- **WHEN** completion is triggered
- **THEN** it SHALL suggest resource references (`skill/commit`, `agent/reviewer`, etc.)
- **AND** it SHALL suggest preset references (`preset/git-workflow`, etc.)

---

### Requirement: Completion Configuration

The CLI SHALL support configurable timeout and caching for completions.

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

---

### Requirement: Library Path Resolution for Completions

Completions SHALL resolve the library path using a priority chain.

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
- **THEN** it SHALL use `~/.config/germinator/library/` as the library path

---

### Requirement: Silent Failure on Completion Errors

Completions SHALL fail silently without error messages when library loading fails.

#### Scenario: Library not found returns empty completions

- **GIVEN** the library path does not exist
- **WHEN** a user triggers completion for `--resources`
- **THEN** no suggestions SHALL be displayed
- **AND** no error message SHALL be shown

#### Scenario: Timeout returns empty completions

- **GIVEN** library loading takes longer than the configured timeout
- **WHEN** completion is triggered
- **THEN** no suggestions SHALL be displayed
- **AND** no error message SHALL be shown

#### Scenario: Invalid library returns empty completions

- **GIVEN** the library.yaml file is malformed
- **WHEN** a user triggers completion for `--resources`
- **THEN** no suggestions SHALL be displayed
- **AND** no error message SHALL be shown

---

### Requirement: Completion Cache

Completions SHALL cache library data to improve performance.

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
