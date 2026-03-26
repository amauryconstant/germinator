# Germinator - OpenCode Reference

Configuration adapter transforming AI coding assistant documents between platforms.

## Architecture

```mermaid
graph LR
    subgraph CLI[CLI Layer]
        V[validate]
        A[adapt]
        C[canonicalize]
        VER[version]
        L[library]
        I[init]
    end

    subgraph APP[Application Layer]
        IT[Transformer]
        IV[Validator]
        IC[Canonicalizer]
        II[Initializer]
    end

    subgraph SVC[Services Layer]
        ST[transformer]
        SV[validator]
        SC[canonicalizer]
        IN[initializer]
    end

    subgraph LIB[Library Layer]
        LL[LoadLibrary]
        LR[ResolveResource]
        LS[ListResources]
        LD[FindLibrary]
    end

    subgraph CORE[Core Layer]
        LDc[LoadDocument]
        P[ParseDocument]
        R[RenderDocument]
        MP[ParsePlatformDocument]
        MC[MarshalCanonical]
    end

    subgraph DOMAIN[Domain Layer]
        DM[Models]
        DE[Errors]
        DV[Validation]
        DR[Results]
    end

    subgraph CFG[Config Layer]
        CM[ConfigManager]
    end

    subgraph ADP[Adapters Layer]
        ACC[claude-code]
        AOC[opencode]
    end

    subgraph TPL[Templates Layer]
        TC[canonical]
        TCC[claude-code]
        TOC[opencode]
    end

    V --> IV
    A --> IT
    C --> IC
    I --> II
    L --> LL
    L --> LS
    IT --> ST
    IV --> SV
    IC --> SC
    II --> IN
    ST --> LDc
    SV --> LDc
    SV --> DV
    DV --> DM
    SC --> MP
    IN --> LDc
    IN --> LR
    LR --> LL
    LL --> LD
    LDc --> P
    R --> TC
    R --> TCC
    R --> TOC
    MC --> TC
    P --> DM
    MP --> ACC
    MP --> AOC
    ACC --> DM
    AOC --> DM
    CM -.->|library path| LDc
```

## Essential Commands

| Command                | Purpose                                    |
| ---------------------- | ------------------------------------------ |
| mise run build         | Build CLI to bin/germinator                |
| mise run check         | All validation (lint, format, test, build) |
| mise run lint          | Run golangci-lint                          |
| mise run lint:fix      | Auto-fix linting issues                    |
| mise run format        | Format Go code                             |
| mise run test          | Run unit tests                             |
| mise run test:e2e      | Run E2E tests (Ginkgo v2)                  |
| mise run test:full     | Run all tests (unit + E2E)                 |
| mise run test:coverage | Run tests with coverage                    |
| mise run clean         | Clean artifacts                            |
| mise tasks             | List all tasks                             |

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

### When to Use

| Situation                       | Action                 |
| ------------------------------- | ---------------------- |
| Multi-step change (3+ tasks)    | Use OpenSpec           |
| New platform support            | Use OpenSpec           |
| Refactor / architectural change | Use OpenSpec           |
| Quick fix (1-2 lines)           | Skip OpenSpec          |
| Unclear requirements            | openspec-explore first |

### Lifecycle

```mermaid
graph TB
    subgraph Exploration["Exploration"]
        E1[openspec-explore]
    end

    subgraph Planning["Planning"]
        P1[openspec-new-change]
        P2[openspec-continue-change<br/>or openspec-ff-change]
        P3[openspec-review-artifacts]
        P4[openspec-modify-artifacts]
    end

    subgraph Implementation["Implementation"]
        I1[openspec-apply-change]
        I2[openspec-review-test-compliance]
    end

    subgraph Completion["Completion"]
        C1[openspec-verify-change]
        C2[openspec-maintain-ai-docs]
        C3[openspec-sync-specs]
        C4[openspec-archive-change<br/>or bulk-archive]
        C5[openspec-generate-changelog]
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
| **Exploration**    | `openspec-explore`                | Think through ideas                              |
| **Planning**       | `openspec-new-change`             | Create change folder                             |
|                    | `openspec-continue-change`        | Create one artifact                              |
|                    | `openspec-ff-change`              | Create all artifacts at once                     |
|                    | `openspec-review-artifacts`       | Review for quality                               |
|                    | `openspec-modify-artifacts`       | Update artifacts _(also in Implementation)_      |
| **Implementation** | `openspec-apply-change`           | Implement tasks                                  |
|                    | `openspec-review-test-compliance` | Check spec→test alignment _(also in Completion)_ |
| **Completion**     | `openspec-verify-change`          | Validate implementation                          |
|                    | `openspec-maintain-ai-docs`       | Update AGENTS.md                                 |
|                    | `openspec-sync-specs`             | Merge delta specs (optional)                     |
|                    | `openspec-archive-change`         | Finalize single change                           |
|                    | `openspec-bulk-archive-change`    | Archive multiple changes                         |
|                    | `openspec-generate-changelog`     | Generate CHANGELOG.md                            |

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
| [cmd/AGENTS.md](cmd/AGENTS.md)                             | CLI commands, Cobra patterns, command specs                  |
| [internal/application/AGENTS.md](internal/application/AGENTS.md) | Service interfaces, request types for DI (results moved to domain) |
| [internal/domain/AGENTS.md](internal/domain/AGENTS.md)     | Domain types, errors, validation, results (consolidated layer) |
| [internal/service/AGENTS.md](internal/service/AGENTS.md)   | Service implementations (Transformer, Validator, etc.)        |
| [internal/infrastructure/parsing/AGENTS.md](internal/infrastructure/parsing/AGENTS.md) | Document loading, parsing, platform detection |
| [internal/infrastructure/serialization/AGENTS.md](internal/infrastructure/serialization/AGENTS.md) | Serialization, template functions |
| [internal/infrastructure/adapters/AGENTS.md](internal/infrastructure/adapters/AGENTS.md) | Platform adapters (Claude Code, OpenCode) |
| [internal/infrastructure/config/AGENTS.md](internal/infrastructure/config/AGENTS.md) | Configuration loading, XDG paths, TOML parsing |
| [internal/infrastructure/library/AGENTS.md](internal/infrastructure/library/AGENTS.md) | Library system, resource management, preset grouping |
| [internal/AGENTS.md](internal/AGENTS.md)                   | Internal package patterns                                    |
| [config/AGENTS.md](config/AGENTS.md)                       | Template patterns, permission mappings                       |
| [test/AGENTS.md](test/AGENTS.md)                           | Golden file testing, E2E testing, mock infrastructure, fixture conventions        |
| [openspec/research/AGENTS.md](openspec/research/AGENTS.md) | Platform research documentation usage                        |
