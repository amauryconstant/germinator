# OpenSpec Specs Organization

> **Source of truth** — This file is the authoritative layout for `openspec/specs/` in this repository. It is **compatible** with the generic layout described in the `osx-concepts` skill and `osx-concepts/references/artifact-formats.md`.

> No `golang-cli-architecture` skill references apply here — this directory documents the OpenSpec artifact layout, not Go architecture.

This guide explains how specs are organized and how to decide where new specs belong.

## Structure Overview

```
openspec/specs/
├── AGENTS.md           # This file (the layout rulebook)
├── <spec-name>/
│   └── spec.md         # One spec per folder
└── ...
```

Specs live one folder deep: `openspec/specs/<spec-name>/spec.md`. The spec name is prefixed with a **category** (e.g., `cli-`, `library-`, `transformation-`) to group related specs and keep the namespace organised. There are currently 65 specs across 10 categories.

## Naming Conventions

- **Spec folders:** lowercase, hyphen-separated, **always prefixed with a category** (e.g., `cli-shell-completion`, `library-library-batch-add`).
- **Spec file:** always `spec.md`.
- **Full path:** `openspec/specs/<category>-<spec-name>/spec.md`.
- **All specs MUST live one folder deep.** Files directly under `openspec/specs/` (e.g., `openspec/specs/foo.md`) are **not permitted**. Each spec has its own folder.
- **No category subdirectories.** Specs are not grouped into `openspec/specs/<category>/` folders — categories are encoded as a name prefix only. This matches the layout the OpenSpec CLI reads natively.

### Naming caveats

- A few specs end up with a doubled prefix where the spec name already starts with the category word, e.g., `library-library-batch-add` (category `library-`, name `library-batch-add`) or `validation-validation-pipeline`. This is intentional — keeping a uniform `<category>-<spec-name>` rule is simpler than special-casing names.
- Pick the **shortest unambiguous prefix** when adding a new spec. Prefer existing category names listed below; only invent a new prefix if none of the existing categories fit.

## Category Reference

These are the active categories, with the criteria for what belongs in each. Use these criteria to pick a prefix when adding a new spec.

### `application-`

**Purpose:** Application architecture, wiring, and configuration.

**What belongs here:**

- Dependency injection patterns and containers
- Service contracts (interfaces between layers)
- Application-wide configuration loading and management
- How components are composed and wired together
- Three-concern separation (Parse / Execute / Respond)

**Code alignment:** `internal/cmdutil/`, `internal/config/`

### `cli-`

**Purpose:** Command-line interface layer.

**What belongs here:**

- CLI framework setup (Cobra, flags, parsing)
- Individual command specifications
- Output formatting for terminal
- Exit codes and error display
- User-facing messages and help text

**Code alignment:** `cmd/`

### `core-`

**Purpose:** Document I/O primitives and the purity policy for the Functional Core.

**What belongs here:**

- Loading documents from files
- Parsing document formats (YAML, TOML, etc.)
- Serializing documents back to files
- Low-level document operations
- The purity rule: `core/` may not import `os`, `net`, `exec` (depguard-enforced)

**Code alignment:** `internal/core/`

### `errors-`

**Purpose:** Error types and handling patterns (internal concern).

**What belongs here:**

- Typed error definitions (ParseError, ValidationError, etc.)
- Error construction patterns (builders, factories)
- Error formatting for internal use
- Error chaining and wrapping

**Code alignment:** `internal/errors/`

### `infrastructure-`

**Purpose:** DevOps, CI/CD, and project-level tooling.

**What belongs here:**

- CI workflow definitions
- Release and versioning processes
- Code quality tooling (linters, formatters)
- Build scripts and task runners
- Project structure conventions

**Code alignment:** Project-level (`.github/`, `mise.toml`, etc.)

### `library-`

**Purpose:** Resource management and installation.

**What belongs here:**

- Library loading and discovery
- Resource resolution by reference
- Preset expansion
- Installation of resources to target projects

**Code alignment:** `internal/library/`

### `models-`

**Purpose:** Domain models and source file formats.

**What belongs here:**

- Data structures representing document types (Agent, Skill, Command, Memory)
- Source format definitions (canonical YAML, platform-specific formats)
- Model fields and their semantics
- Serialization tags and format mappings

**Code alignment:** `internal/core/`

### `testing-`

**Purpose:** End-to-end test infrastructure.

**What belongs here:**

- E2E test suite setup and conventions
- Test fixture management
- CLI testing helpers
- Integration test scenarios

**Code alignment:** `test/`

### `transformation-`

**Purpose:** Document transformation pipeline.

**What belongs here:**

- Converting between document formats
- Platform adapters (OpenCode, Claude Code)
- Template rendering
- Permission and field transformations
- Field mapping rules between platforms

**Code alignment:** `internal/{claude-code,opencode,parser,renderer,permission}/`

### `validation-`

**Purpose:** Document validation system.

**What belongs here:**

- Validation pipeline and composition
- Individual validator functions
- Platform-specific validation rules
- Validation result types

**Code alignment:** `internal/validation/`

## Spec Catalog

The 64 current specs, grouped by category.

### `application-` (4)

- `application-configuration` — Application configuration loading (XDG + koanf)
- `application-dependency-injection` — Dependency injection via Factory
- `application-discover-orphans-batch` — Orphan discovery batch mode
- `application-three-concerns` — Three-concern separation (Parse / Execute / Respond)

### `cli-` (17)

- `cli-cli-factory` — `cmdutil.Factory` with lazy Config + Library
- `cli-color-policy` — `--color=always|never|auto` + NO_COLOR
- `cli-command-options-pattern` — Per-command `*XxxOptions` + `runF` injection
- `cli-config-commands` — `germinator config init` / `config validate`
- `cli-destructive-operations` — `confirmOrFlag` helper for destructive ops
- `cli-error-formatting` — Typed error formatting with hints
- `cli-exit-codes` — Semantic exit code conventions (0/1/2)
- `cli-flag-deprecation` — `MarkDeprecated` + removal cadence
- `cli-framework` — Cobra-based CLI framework setup
- `cli-init-command` — `germinator init` resource installation
- `cli-interactive-prompts` — Flags-first / huh fallback / TTY-gated
- `cli-iostreams` — `IOStreams` abstraction, TTY detection, Styles
- `cli-output-formats` — `--output json|table|plain` + `Exporter` interface
- `cli-self-update` — Opt-in update checking (consent + daily TTL)
- `cli-shell-completion` — Carapace shell completion
- `cli-stdin-composability` — `-` filename + no-hang-on-empty-stdin
- `cli-verbose-output` — Verbosity levels and output helpers

### `core-` (3)

- `core-document-loading` — Document loading from files
- `core-purity-policy` — `core/` may not import `os`/`net`/`exec` (depguard-enforced)
- `core-yaml-parsing` — YAML parsing primitives

### `errors-` (4)

- `errors-enhanced-errors` — Enhanced error types (private fields + builders)
- `errors-enhanced-validation-errors` — Enhanced validation errors
- `errors-operation-error` — `OperationError` (per-operation failures)
- `errors-typed-errors` — Typed error definitions (all 9 error types)

### `infrastructure-` (6)

- `infrastructure-ci-image` — CI image build
- `infrastructure-ci-workflow` — CI workflow definitions
- `infrastructure-code-quality` — Linters and formatters
- `infrastructure-mise-task-runner` — mise task definitions
- `infrastructure-release-management` — Release process
- `infrastructure-validation-scripts` — Validation scripts

### `library-` (14)

- `library-library-batch-add` — Batch resource import
- `library-library-json-output` — JSON output for library commands
- `library-library-orphan-discovery` — Orphan file discovery
- `library-library-persistence` — Atomic writes + file locking + permissions
- `library-library-preset-creation` — Preset creation
- `library-library-refresh` — Library refresh (metadata sync)
- `library-library-remove-preset` — Preset removal
- `library-library-remove-resource` — Resource removal
- `library-library-resource-import` — Single-resource import
- `library-library-scaffolding` — Library init scaffolding
- `library-library-system` — Library data model
- `library-library-validation` — Library integrity validation
- `library-partial-initialization` — Partial initialization behavior
- `library-resource-installation` — Resource installation behavior

### `models-` (2)

- `models-canonical-source-format` — Canonical source format definition
- `models-domain-models` — Domain model types

### `testing-` (3)

- `testing-e2e-canonicalize-tests` — E2E canonicalization tests
- `testing-e2e-testing` — E2E testing setup
- `testing-iostreams-injection` — `iostreams.Test()` pattern + runF injection

### `transformation-` (6)

- `transformation-document-transformation` — Document transformation pipeline
- `transformation-permission-transformation` — Permission policy mapping
- `transformation-platform-adapters` — Platform adapter implementation
- `transformation-platform-field-mappings` — Platform field mappings
- `transformation-platform-to-canonical` — Reverse transformation to canonical
- `transformation-template-rendering` — Template rendering

### `validation-` (5)

- `validation-composable-validators` — Composable validator pipeline
- `validation-opencode-platform-validation` — OpenCode-specific validation
- `validation-platform-agnostic-validation` — Cross-platform validation
- `validation-result-type` — Result type for composable errors
- `validation-validation-pipeline` — Validation pipeline orchestration
