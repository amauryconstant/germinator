# OpenSpec Specs Organization

This guide explains how specs are organized and how to decide where new specs belong.

## Structure Overview

```
openspec/specs/
├── application/    # Architecture and wiring
├── cli/            # User interface
├── core/           # Document I/O
├── errors/         # Error types
├── infrastructure/ # DevOps tooling
├── library/        # Resource management
├── models/         # Data structures
├── testing/        # E2E tests
├── transformation/ # Format conversion
└── validation/     # Validation rules
```

## Naming Conventions

- **Category folders:** lowercase, hyphen-separated (e.g., `validation/`)
- **Spec folders:** lowercase, hyphen-separated (e.g., `composable-validators/`)
- **Spec file:** always `spec.md`
- **Full path:** `openspec/specs/<category>/<spec-name>/spec.md`

## Category Reference

### `application/`

**Purpose:** Application architecture, wiring, and configuration.

**What belongs here:**

- Dependency injection patterns and containers
- Service contracts (interfaces between layers)
- Application-wide configuration loading and management
- How components are composed and wired together

**Questions to ask:**

- Does this spec describe how parts of the application connect?
- Is this about the "glue" between layers?
- Does it define interfaces that services implement?

**Code alignment:** `internal/application/`, `internal/config/`

---

### `cli/`

**Purpose:** Command-line interface layer.

**What belongs here:**

- CLI framework setup (Cobra, flags, parsing)
- Individual command specifications
- Output formatting for terminal
- Exit codes and error display
- User-facing messages and help text

**Questions to ask:**

- Does this spec describe what users type and see?
- Is this about command parsing or flag handling?
- Does it affect how errors are shown to users?

**Code alignment:** `cmd/`

---

### `core/`

**Purpose:** Document I/O primitives.

**What belongs here:**

- Loading documents from files
- Parsing document formats (YAML, TOML, etc.)
- Serializing documents back to files
- Low-level document operations

**Questions to ask:**

- Does this spec describe reading or writing files?
- Is this about parsing syntax or format?
- Does it operate at the document level, not the model level?

**Code alignment:** `internal/core/`

---

### `errors/`

**Purpose:** Error types and handling patterns (internal concern).

**What belongs here:**

- Typed error definitions (ParseError, ValidationError, etc.)
- Error construction patterns (builders, factories)
- Error formatting for internal use
- Error chaining and wrapping

**Questions to ask:**

- Is this spec defining a new error type?
- Does it describe how errors are constructed or enriched?
- Is this about error structure, not error display?

**Code alignment:** `internal/errors/`

---

### `infrastructure/`

**Purpose:** DevOps, CI/CD, and project-level tooling.

**What belongs here:**

- CI workflow definitions
- Release and versioning processes
- Code quality tooling (linters, formatters)
- Build scripts and task runners
- Project structure conventions

**Questions to ask:**

- Does this spec affect how the project is built or released?
- Is this about developer tooling, not runtime behavior?
- Would a DevOps engineer care about this?

**Code alignment:** Project-level (`.github/`, `mise.toml`, etc.)

---

### `library/`

**Purpose:** Resource management and installation.

**What belongs here:**

- Library loading and discovery
- Resource resolution by reference
- Preset expansion
- Installation of resources to target projects

**Questions to ask:**

- Does this spec describe finding or loading resources?
- Is this about the library of templates/skills/agents?
- Does it involve installing something into a user's project?

**Code alignment:** `internal/library/`

---

### `models/`

**Purpose:** Domain models and source file formats.

**What belongs here:**

- Data structures representing document types (Agent, Skill, Command, Memory)
- Source format definitions (canonical YAML, platform-specific formats)
- Model fields and their semantics
- Serialization tags and format mappings

**Questions to ask:**

- Does this spec define what a document type looks like?
- Is this about the structure of data, not how it's processed?
- Does it describe fields, types, and their meanings?

**Code alignment:** `internal/models/`

---

### `testing/`

**Purpose:** End-to-end test infrastructure.

**What belongs here:**

- E2E test suite setup and conventions
- Test fixture management
- CLI testing helpers
- Integration test scenarios

**Questions to ask:**

- Does this spec describe how to test the CLI?
- Is this about test infrastructure, not unit tests?
- Does it involve running the actual binary?

**Code alignment:** `test/`

---

### `transformation/`

**Purpose:** Document transformation pipeline.

**What belongs here:**

- Converting between document formats
- Platform adapters (OpenCode, Claude Code)
- Template rendering
- Permission and field transformations
- Field mapping rules between platforms

**Questions to ask:**

- Does this spec describe converting one format to another?
- Is this about the transformation logic, not validation?
- Does it involve templates or adapters?

**Code alignment:** `internal/services/`

---

### `validation/`

**Purpose:** Document validation system.

**What belongs here:**

- Validation pipeline and composition
- Individual validator functions
- Platform-specific validation rules
- Validation result types

**Questions to ask:**

- Does this spec describe checking if something is valid?
- Is this about validation rules or error collection?
- Does it involve the Result[T] type or validators?

**Code alignment:** `internal/validation/`

---
