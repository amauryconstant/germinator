# Germinator - OpenCode Reference

Configuration adapter transforming AI coding assistant documents between platforms.

## Architecture

```mermaid
graph LR
    subgraph CLI[cmd/ — flat layout, one file per command]
        direction TB
        AD[adapt]
        VA[validate]
        CA[canonicalize]
        IN[init]
        LIB[library_*]
        CFG[config_*]
        VER[version]
        COM[completion*]
    end

    subgraph SHELL[Imperative Shell — internal/]
        direction TB
        CU[cmdutil/]
        IO[iostreams/]
        OUT[output/]
        CONF[config/]
        LIBL[library/]
        CC[claude-code/]
        OC[opencode/]
        PAR[parser/]
        REN[renderer/]
        TRA[transform/]
        VAL[validate/]
        CAN[canonicalize/]
        INS[install/]
        VERP[version/]
    end

    subgraph CORE[Functional Core — internal/]
        COR[core/<br/>types, validation,<br/>rules, errors]
    end

    AD --> CU
    AD --> TRA
    AD --> PAR
    AD --> REN
    AD --> CC
    AD --> OC
    AD --> COR
    VA --> CU
    VA --> VAL
    VA --> PAR
    VA --> COR
    CA --> CU
    CA --> CAN
    CA --> PAR
    CA --> REN
    CA --> COR
    IN --> CU
    IN --> INS
    IN --> PAR
    IN --> REN
    IN --> COR
    LIB --> CU
    LIB --> LIBL
    CFG --> CU
    CFG --> CONF
    VER --> CU
    VER --> VERP
    COM --> CU
    COM --> LIBL
    LIBL --> COR
    CONF --> COR
    CC --> COR
    OC --> COR
    CU --> IO
    CU --> OUT
    INS --> LIBL
    COR -.->|no I/O| SHELL
```

The target architecture is **golang-cli-architecture** (Functional Core / Imperative Shell). All I/O lives in shell packages (`iostreams/`, `output/`, `cmdutil/`, `config/`, `library/`, `claude-code/`, `opencode/`, `parser/`, `renderer/`, `transform/`, `validate/`, `canonicalize/`, `install/`, `version/`); `core/` is pure logic with stdlib + samber/lo only. The service-style I/O adapters (`Transformer`, `Validator`, `Canonicalizer`, `Initializer`) live in dedicated `internal/<x>/` shell packages (per `openspec/specs/cli-framework/spec.md` "I/O adapter placement"); cmd-side interfaces remain in `cmd/` next to the consumer. `cmd/` uses a **flat layout** — one `.go` file per command, with multi-word commands as sibling files (`library.go`, `library_add.go`, `config_init.go`), not subdirectories.

## Reference Skills

This project follows the **`golang-cli-architecture`** skill as its primary reference. It is the **only** Go skill always-loaded at session start. All other Go skills are loaded **on demand** based on task intent. Per-package AGENTS.md files (see [Location-Specific Guides](#location-specific-guides)) inline-link specific `golang-cli-architecture/references/*.md` files via `@<path>` and remain authoritative for their domain.

Library selection rationale: `@.opencode/skills/golang-cli-architecture/references/14-libraries.md`

### Primary skill — always loaded

Load **`golang-cli-architecture`** (`@.opencode/skills/golang-cli-architecture/SKILL.md`)
at the start of **every** Go task. It is the source of truth for:

- Project layout (`cmd/` flat, `internal/core/` Functional Core, `internal/<x>/` Imperative Shell)
- Factory pattern with lazy function fields (replaces DI containers)
- IOStreams abstraction, exit code mapping, error formatting
- CLI testing pyramid (Core unit → Command via `runF` → Integration → E2E)
- Three-concern model (Parse / Execute / Respond)

> **Token budget.** Always-loaded skills add their description tokens to every
> session (~100 tokens per skill). `golang-cli-architecture` is the only Go skill
> with mandatory session-startup load.

### Intent-based skill loading

For each Go task, load the **primary** skill plus any **secondary** skill(s) the
intent requires. `golang-cli-architecture` is **always** a secondary (it is the
always-loaded primary above). Do not load all skills — pick by intent.

| Intent                                         | Primary                          | Also load                                                                                       |
|------------------------------------------------|----------------------------------|-------------------------------------------------------------------------------------------------|
| Add/restruct a CLI command, scaffold a command | `golang-cli-architecture`        | `golang-spf13-cobra`, `golang-naming`, `golang-code-style`                                      |
| Add a Functional Core type, validator, or rule | `golang-design-patterns`        | `golang-cli-architecture`, `golang-structs-interfaces`, `golang-samber-lo`                      |
| Add Factory wiring / lazy fn fields            | `golang-design-patterns`        | `golang-cli-architecture`                                                                       |
| Implement error wrapping, errors.Is/As, slog   | `golang-error-handling`          | `golang-cli-architecture`, `golang-safety` (nil-heavy code)                                     |
| Add a Cobra command group, ValidArgsFunction, completion | `golang-spf13-cobra`    | `golang-cli-architecture`                                                                       |
| Use samber/lo helpers in `core/`               | `golang-samber-lo`               | `golang-cli-architecture`, `golang-data-structures`                                             |
| Write unit tests with testify                  | `golang-stretchr-testify`        | `golang-cli-architecture`, `golang-testing`                                                     |
| Write integration or E2E test                  | `golang-cli-architecture`        | `golang-testing`, `golang-stretchr-testify`                                                     |
| Configure / tune golangci-lint                 | `golang-lint`                    | `golang-cli-architecture`, `golang-code-style`                                                  |
| Debug a panic, deadlock, or unexpected behavior | `golang-troubleshooting`        | `golang-cli-architecture`, `golang-safety`                                                      |
| Refactor across packages                       | `golang-design-patterns`        | `golang-cli-architecture`, `golang-naming`, `golang-code-style`                                 |
| Modernize for a new Go version                 | `golang-modernize`               | `golang-cli-architecture`, `golang-lint`                                                        |
| Write godoc for an `internal/` package         | `golang-documentation`           | `golang-cli-architecture`, `golang-naming`                                                      |
| Review formatting, style, naming               | `golang-code-style`              | `golang-cli-architecture`, `golang-naming`, `golang-lint`                                       |
| Propagate ctx / add cancellation               | `golang-context`                 | `golang-cli-architecture`                                                                       |
| Choose a new library                           | `golang-cli-architecture` (read `references/14-libraries.md`) | —                                                              |
| Look up an unfamiliar package                  | `golang-pkg-go-dev`              | `golang-cli-architecture`                                                                       |
| Navigate / refactor locally with gopls         | `golang-gopls`                   | `golang-cli-architecture`                                                                       |

> **Note on `golang-testing`:** it is partially superseded by
> `golang-cli-architecture` for the CLI testing pyramid
> (Core → Command via `runF` → Integration → E2E). It is kept for
> table-driven patterns, fuzzing, goleak, and coverage idioms.
>
> **Note on configuration:** germinator uses **`koanf`**, not `viper`.
> `golang-spf13-viper` is intentionally not listed. Library rationale:
> `@.opencode/skills/golang-cli-architecture/references/14-libraries.md`.

### Skill catalog (secondary)

| Category        | Skills                                                                                                                       |
|-----------------|------------------------------------------------------------------------------------------------------------------------------|
| Style & docs    | `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-lint`, `golang-safety`                                 |
| Architecture    | `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-data-structures`                           |
| Errors          | `golang-error-handling`                                                                                                      |
| Testing         | `golang-testing`, `golang-stretchr-testify`                                                                                  |
| Libraries       | `golang-spf13-cobra`, `golang-samber-lo`                                                                                     |
| Tools           | `golang-troubleshooting`, `golang-modernize`, `golang-gopls`, `golang-pkg-go-dev`                                             |

### Supersession — DO NOT load these `samber/cc-skills-golang` skills

The following samber skills are explicitly overridden by local skills or by
project decisions and should NOT be loaded for germinator work:

- `samber/cc-skills-golang@golang-cli` → superseded by `golang-cli-architecture`
- `samber/cc-skills-golang@golang-dependency-injection` → Factory pattern replaces DI containers
- `samber/cc-skills-golang@golang-project-layout` → CLI-specific Tier 1/2/3 layouts
- `samber/cc-skills-golang@golang-concurrency` → sequential-first for CLIs
- `samber/cc-skills-golang@golang-spf13-viper` → we use **koanf**, not viper

Domain-mismatched skills (also not loaded): `golang-database`, `golang-graphql`,
`golang-grpc`, `golang-swagger`, `golang-observability`, `golang-benchmark`,
`golang-performance`, all DI container skills (`golang-google-wire`,
`golang-uber-dig`, `golang-uber-fx`, `golang-samber-do`), and all other
`samber/*` skills not listed in the catalog above.

## Essential Commands

| Command                | Purpose                                    |
| ---------------------- | ------------------------------------------ |
| mise run build         | Build CLI to bin/germinator                |
| mise run verify         | All validation (lint, format, test, build) |
| mise run lint          | Run golangci-lint                          |
| mise run lint:fix      | Auto-fix linting issues                    |
| mise run format        | Format Go code                             |
| mise run test          | Run unit tests                             |
| mise run test:e2e      | Run E2E tests (Ginkgo v2)                  |
| mise run test:full     | Run all tests (unit + E2E)                 |
| mise run test:coverage | Run tests with coverage                    |
| mise run clean         | Clean artifacts                            |
| mise tasks             | List all tasks                             |

## Config Commands

| Command                       | Purpose                                      |
| ------------------------------ | -------------------------------------------- |
| `germinator config init`       | Scaffold a config file with documented fields |
| `germinator config validate`   | Validate an existing config file             |

**Config init flags:**
- `--output-path <path>` - Output file path (default: `~/.config/germinator/config.toml`)
- `--force` - Overwrite existing file

**Config validate flags:**
- `--output-path <path>` - Config file to validate (default: `~/.config/germinator/config.toml`)

## Library Commands

Nine subcommands for managing the canonical resource library (skills, agents, commands, memory). All support `--output plain|json|table` via `output.AddOutputFlags`; `--library` overrides `$GERMINATOR_LIBRARY` and the XDG default. See [`cmd/commands/AGENTS.md`](cmd/commands/AGENTS.md) for full flag tables, output formats, discover/batch modes, and examples.

| Command | Purpose |
|---|---|
| `library init` | Scaffold library directory structure |
| `library add` | Import a resource (`--discover`/`--batch` modes too) |
| `library create preset` | Create a preset referencing resources |
| `library resources` | List all resources grouped by type |
| `library presets` | List all presets |
| `library show <ref>` | Display resource or preset details |
| `library refresh` | Sync metadata from resource files into `library.yaml` |
| `library remove` | Remove resource (`resource <ref>`) or preset (`preset <name>`) |
| `library validate` | Check library integrity (`--fix` for auto-cleanup) |

**Library path discovery:** `--library` flag > `$GERMINATOR_LIBRARY` env > `$XDG_DATA_HOME/germinator/library/` (or `~/.local/share/germinator/library/`).
**Resource references:** format `type/name` (e.g., `skill/commit`, `agent/reviewer`); valid types are `skill`, `agent`, `command`, `memory`.
**Library init only subcommand without `--library`:** uses `--path <path>` instead.

## Release

| Command              | Purpose                                        |
| -------------------- | ---------------------------------------------- |
| mise run release     | Validate, update changelog, commit, and tag   |
| mise run release:check | Validate prerequisites (no execution)         |
| mise run release:prepare | Validate and preview operations             |
| mise run test:release | Test GoReleaser release flow (build only)     |

Workflow:
1. `mise run osx-changelog` - Generate changelog from archived OpenSpec changes
2. `mise run release:check` - Validate prerequisites
3. `mise run release:prepare <patch|minor|major>` - Preview what would happen
4. `mise run release <patch|minor|major>` - Execute release when ready

Optional: `mise run test:release` - Test goreleaser build without publishing

## Pre-Commit Hooks

Setup: `pre-commit install`
Run: `pre-commit run --all-files`
Skip: `git commit -m "msg" --no-verify`

Hooks: gofmt, govet, golangci-lint, YAML/TOML/JSON validation, file hygiene.

## OpenSpec Workflow

**Config**: `openspec/config.yaml` (spec-driven schema)

> **Spec organization** — Specs follow the flat layout described in the "Spec organization" section of [`openspec/config.yaml`](openspec/config.yaml): `openspec/specs/<category>-<spec-name>/spec.md`. Always consult the source-of-truth section in `config.yaml` before creating, syncing, or moving a spec to pick the right category prefix.

### When to Use

| Situation                       | Action                 |
| ------------------------------- | ---------------------- |
| Multi-step change (3+ tasks)    | Use OpenSpec           |
| New platform support            | Use OpenSpec           |
| Refactor / architectural change | Use OpenSpec           |
| Quick fix (1-2 lines)           | Skip OpenSpec          |
| Unclear requirements            | osc-explore first |

### Lifecycle

```mermaid
graph TB
    subgraph Exploration["Exploration"]
        E1[osc-explore]
    end

    subgraph Planning["Planning"]
        P1[osc-new-change]
        P2[osc-continue-change<br/>or osc-ff-change]
        P3[osx-review-artifacts]
        P4[osx-modify-artifacts]
    end

    subgraph Implementation["Implementation"]
        I1[osc-apply-change]
        I2[osx-review-test-compliance]
    end

    subgraph Completion["Completion"]
        C1[osc-verify-change]
        C2[osx-maintain-ai-docs]
        C3[osc-sync-specs]
        C4[osc-archive-change<br/>or bulk-archive]
        C5[osx-generate-changelog]
    end

    E1 --> P1 --> P2 --> P3 --> I1 --> C1 --> C2 --> C4 --> C5
    C2 -.->|optional| C3 --> C4

    P3 -.->|issues found| P4
    P4 -.-> P3
    I1 -.->|reality diverges| P4
    I1 -.->|test gaps| I2
    I2 -.->|implement tests| I1
    C1 -.->|with| I2
```

### Skills by Phase

| Phase              | Skill                             | Purpose                                          |
| ------------------ | --------------------------------- | ------------------------------------------------ |
| **Exploration**    | `osc-explore`                | Think through ideas                              |
| **Planning**       | `osc-new-change`             | Create change folder                             |
|                    | `osc-continue-change`        | Create one artifact                              |
|                    | `osc-ff-change`              | Create all artifacts at once                     |
|                    | `osx-review-artifacts`       | Review for quality                               |
|                    | `osx-modify-artifacts`       | Update artifacts _(also in Implementation)_      |
| **Implementation** | `osc-apply-change`           | Implement tasks                                  |
|                    | `osx-review-test-compliance` | Check spec→test alignment _(also in Completion)_ |
| **Completion**     | `osc-verify-change`          | Validate implementation                          |
|                    | `osx-maintain-ai-docs`       | Update AGENTS.md                                 |
|                    | `osc-sync-specs`             | Merge delta specs (optional)                     |
|                    | `osc-archive-change`         | Finalize single change                           |
|                    | `osc-bulk-archive-change`    | Archive multiple changes                         |
|                    | `osx-generate-changelog`     | Generate CHANGELOG.md                            |

### Project Conventions

| Rule      | Detail                                                                             |
| --------- | ---------------------------------------------------------------------------------- |
| Tests     | Unit tests alongside code, golden file tests for transformations, E2E for CLI, mocks for isolated unit testing      |
| Progress  | Check tasks.md in change folder for completion status                              |
| Artifacts | Follow openspec/config.yaml rules section                                          |
| Archive   | See openspec/changes/archive/ for examples                                         |

## Location-Specific Guides

| File                                                       | Purpose                                                      |
| ---------------------------------------------------------- | ------------------------------------------------------------ |
| [cmd/AGENTS.md](cmd/AGENTS.md)                             | CLI architecture: DI, error handling, exit codes, verbosity, lint enforcement |
| [cmd/commands/AGENTS.md](cmd/commands/AGENTS.md)           | Per-command flag tables and behavior (Library, Init, Config, Completion) |
| [internal/core/AGENTS.md](internal/core/AGENTS.md)         | Functional Core: types, validation, rules, errors (pure)     |
| [internal/iostreams/AGENTS.md](internal/iostreams/AGENTS.md) | IOStreams abstraction, TTY detection, Styles, Verbosef     |
| [internal/output/AGENTS.md](internal/output/AGENTS.md)     | Shared output: FormatError, Exporter, AddOutputFlags         |
| [internal/cmdutil/AGENTS.md](internal/cmdutil/AGENTS.md)   | Factory (lazy fn fields), ExitCode mapping, cmd helpers     |
| [internal/config/AGENTS.md](internal/config/AGENTS.md)     | Configuration loading, XDG paths, TOML parsing               |
| [internal/library/AGENTS.md](internal/library/AGENTS.md)   | Library system, resource management, preset grouping         |
| [internal/transform/AGENTS.md](internal/transform/AGENTS.md) | Document transformation shell package (`transform.NewService`) |
| [internal/validate/AGENTS.md](internal/validate/AGENTS.md) | Document validation shell package (`validate.NewService`)      |
| [internal/canonicalize/AGENTS.md](internal/canonicalize/AGENTS.md) | Document canonicalization shell package (`canonicalize.NewService`) |
| [internal/install/AGENTS.md](internal/install/AGENTS.md) | Resource installation shell package (`install.NewService`)    |
| [internal/claude-code/AGENTS.md](internal/claude-code/AGENTS.md) | Claude Code platform adapter                              |
| [internal/opencode/AGENTS.md](internal/opencode/AGENTS.md) | OpenCode platform adapter                                    |
| [internal/AGENTS.md](internal/AGENTS.md)                   | Internal package patterns                                    |
| [config/AGENTS.md](config/AGENTS.md)                       | Template patterns, permission mappings                       |
| [test/AGENTS.md](test/AGENTS.md)                           | Golden file testing, E2E testing, runF injection, fixture conventions |
| [openspec/research/AGENTS.md](openspec/research/AGENTS.md) | Platform research documentation usage                        |

**Removed**: `openspec/specs/AGENTS.md` (deleted; spec-organization rules now live in the `context` block of [`openspec/config.yaml`](openspec/config.yaml) under "Spec organization").
